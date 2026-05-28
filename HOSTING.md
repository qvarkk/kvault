# Самостоятельный хостинг kvault

Это руководство описывает, как развернуть **kvault** на собственном сервере: от первого запуска до доступа по вашему домену через реверс-прокси с HTTPS.

kvault распространяется как набор Docker-образов и запускается одной командой через Docker Compose. Весь стек (фронтенд, API, фоновый воркер, PostgreSQL, Redis и S3-хранилище Garage) поднимается вместе.

---

## Содержание

- [Что вам понадобится](#что-вам-понадобится)
- [Шаг 1. Получение файлов](#шаг-1-получение-файлов)
- [Шаг 2. Настройка .env](#шаг-2-настройка-env)
- [Шаг 3. Первый запуск](#шаг-3-первый-запуск)
- [Шаг 4. Настройка хранилища Garage (S3)](#шаг-4-настройка-хранилища-garage-s3)
- [Доступ по своему домену](#доступ-по-своему-домену)
- [Реверс-прокси на поддомен (с HTTPS)](#реверс-прокси-на-поддомен-с-https)
- [Загрузка файлов и presigned-ссылки](#загрузка-файлов-и-presigned-ссылки)
- [Справочник переменных .env](#справочник-переменных-env)
- [Обновление и обслуживание](#обновление-и-обслуживание)
- [Решение частых проблем](#решение-частых-проблем)

---

## Что вам понадобится

- Сервер (VPS или физическая машина) с **Docker** и **Docker Compose**.
- Открытые наружу порты (минимум — порт фронтенда, по умолчанию `80`).
- *Опционально, но рекомендуется:* доменное имя и реверс-прокси (nginx, Caddy, Traefik) для HTTPS.

---

## Шаг 1. Получение файлов

Создайте отдельный каталог под развёртывание и скачайте в него compose-файл, пример конфигурации и вспомогательные конфиги Redis и Garage:

```bash
mkdir kvault && cd kvault

# Compose-файл и пример конфига
curl -O https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/docker-compose.yml
curl -O https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/.env.example
mv .env.example .env

# Конфиги Redis и Garage
mkdir -p docker/redis docker/garage
curl -o docker/redis/redis.conf      https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/docker/redis/redis.conf
curl -o docker/redis/entrypoint.sh   https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/docker/redis/entrypoint.sh
curl -o docker/garage/garage.toml    https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/docker/garage/garage.toml
```

> **Реестр образов.** По умолчанию используется `gitverse.ru/qvarkk` — он стабильно доступен в России. Если предпочитаете GitHub Container Registry, задайте в `.env` переменную `REGISTRY=ghcr.io/qvarkk` (из РФ доступ к ghcr может быть нестабилен).

---

## Шаг 2. Настройка .env

Откройте `.env` и **обязательно** измените перед запуском в продакшене:

| Переменная | Зачем менять |
| --- | --- |
| `DB_PASSWORD` | Пароль базы данных — задайте надёжный. |
| `REDIS_PASSWORD` | Пароль Redis — задайте надёжный. |
| `API_CORS_ORIGINS` | Публичный адрес, по которому будет открываться kvault (см. ниже). |
| `AWS_PUBLIC_ENDPOINT_URL` | Публичный адрес хранилища файлов (см. [presigned-ссылки](#загрузка-файлов-и-presigned-ссылки)). |

Ключи доступа к хранилищу (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`) вы заполните позже — на [шаге 4](#шаг-4-настройка-хранилища-garage-s3), после генерации их в Garage.

Полное описание всех переменных — в [справочнике ниже](#справочник-переменных-env).

---

## Шаг 3. Первый запуск

```bash
docker compose pull
docker compose up -d
```

Поднимутся все сервисы. Контейнер `migrate` один раз применит миграции базы данных и завершится — это нормально.

Проверить состояние:

```bash
docker compose ps
docker compose logs -f        # логи всех сервисов
```

После запуска фронтенд доступен по адресу `http://<адрес-сервера>` (порт `80` по умолчанию).

> На этом этапе вход в систему уже работает, но **загрузка файлов ещё не будет работать**, пока вы не настроите хранилище Garage на следующем шаге.

---

## Шаг 4. Настройка хранилища Garage (S3)

Garage — это S3-совместимое хранилище для загружаемых файлов. При первом запуске нужно один раз инициализировать кластер и сгенерировать ключи доступа.

```bash
# Удобный алиас
alias garage="docker exec kvault_garage /garage"

# 1. Узнать ID узла
garage status

# 2. Создать layout кластера (подставьте <node-id> из вывода выше).
#    Флаг -c задаёт ёмкость узла.
garage layout assign -z dc1 -c 1G <node-id>
garage layout apply --version 1

# 3. Создать ключ доступа.
#    ВАЖНО: сохраните Secret Key сразу — позже его посмотреть нельзя.
garage key create kvault-key

# 4. Создать бакет и выдать ключу права на него
garage bucket create kvault-bucket
garage bucket allow --read --write --owner kvault-bucket --key kvault-key
```

Подробнее о layout: <https://garagehq.deuxfleurs.fr/documentation/quick-start/#creating-a-cluster-layout>

Перенесите выданные **Access Key ID** и **Secret Key** в `.env`:

```bash
AWS_ACCESS_KEY_ID="<сгенерированный Access Key ID>"
AWS_SECRET_ACCESS_KEY="<сгенерированный Secret Key>"
AWS_S3_BUCKET="kvault-bucket"
```

Перезапустите, чтобы применить новые значения:

```bash
docker compose up -d
```

Теперь загрузка и просмотр файлов работают.

---

## Доступ по своему домену

Как именно настраивать домен, зависит от того, открываете ли вы kvault напрямую или через реверс-прокси.

### Вариант А. Прямой доступ (без прокси)

1. Направьте A-запись домена на IP сервера.
2. В `.env` укажите домен в списке разрешённых источников:
   ```bash
   API_CORS_ORIGINS=http://kvault.example.com
   ```
3. Перезапустите: `docker compose up -d`.

kvault откроется по `http://kvault.example.com`. Подходит для теста, но **без HTTPS** — для боевого использования предпочтителен следующий вариант.

### Вариант Б. Через реверс-прокси (рекомендуется)

Этот вариант даёт HTTPS и аккуратный доступ по поддомену — см. следующий раздел.

> **Как фронтенд находит API.** Внутри развёртывания фронтенд сам проксирует запросы `/api/` на сервис API по внутренней сети Docker. Поэтому браузеру и вашему реверс-прокси достаточно «знать» **только адрес фронтенда** — API отдельно публиковать наружу не нужно.

---

## Реверс-прокси на поддомен (с HTTPS)

Сценарий: kvault должен открываться по `https://kvault.example.com`, а сам контейнер фронтенда слушает локальный порт на сервере.

**1. Освободите порт `80` для прокси** и переведите фронтенд на другой порт хоста — в `.env`:

```bash
FRONTEND_PORT=8081
API_CORS_ORIGINS=https://kvault.example.com
```

Перезапустите: `docker compose up -d`. Теперь контейнер фронтенда доступен локально на `http://localhost:8081`.

**2. Настройте реверс-прокси.** Пример для nginx (`/etc/nginx/sites-available/kvault`):

```nginx
server {
    listen 80;
    server_name kvault.example.com;

    # Файлы могут быть крупными — поднимаем лимит тела запроса
    client_max_body_size 100M;

    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**3. Добавьте HTTPS.** Проще всего через Let's Encrypt:

```bash
sudo certbot --nginx -d kvault.example.com
```

Certbot сам пропишет TLS-сертификат и редирект с HTTP на HTTPS.

> **Не забудьте про `API_CORS_ORIGINS`.** Значение должно **точно** совпадать со схемой и хостом, по которым открывается сайт. Если перешли на HTTPS — укажите `https://kvault.example.com`. Несовпадение приведёт к ошибкам CORS в браузере.

<details>
<summary>Альтернатива: Caddy (HTTPS из коробки)</summary>

`Caddyfile`:

```caddy
kvault.example.com {
    reverse_proxy localhost:8081
    request_body {
        max_size 100MB
    }
}
```

Caddy получит и автоматически продлит TLS-сертификат самостоятельно.

</details>

---

## Загрузка файлов и presigned-ссылки

Скачивание и просмотр файлов работают через **presigned-ссылки** на хранилище Garage: API формирует временную ссылку, а **браузер пользователя** идёт по ней напрямую в хранилище. Поэтому адрес хранилища в такой ссылке должен быть **доступен из браузера**, а не только внутри Docker.

За это отвечает `AWS_PUBLIC_ENDPOINT_URL`. Укажите в нём **публичный** адрес сервера и порт Garage:

```bash
# Порт должен совпадать с GARAGE_S3_PORT (по умолчанию 3900)
AWS_PUBLIC_ENDPOINT_URL=http://kvault.example.com:3900
```

При этом:

- `AWS_ENDPOINT_URL` (внутренний адрес, `http://garage:3900`) **менять не нужно** — по нему API общается с Garage внутри сети Docker.
- Порт `GARAGE_S3_PORT` (по умолчанию `3900`) должен быть **открыт наружу** на сервере, чтобы браузеры могли скачивать файлы.

> **Хотите спрятать порт `3900` за HTTPS?** Заведите для Garage отдельный поддомен (например, `s3.example.com`), проксируйте его на `localhost:3900` тем же способом, что и фронтенд, и тогда укажите `AWS_PUBLIC_ENDPOINT_URL=https://s3.example.com`. Не забудьте поднять `client_max_body_size`/лимит тела запроса в конфиге прокси.

---

## Справочник переменных .env

### Общие

| Переменная | По умолчанию | Описание |
| --- | --- | --- |
| `REGISTRY` | `gitverse.ru/qvarkk` | Реестр образов: `gitverse.ru/qvarkk` или `ghcr.io/qvarkk`. |
| `DEBUG` | `false` | Режим отладки — в продакшене держите `false`. |
| `API_PORT` | `6767` | Порт API **внутри** сети Docker. Наружу обычно не публикуется. |
| `FRONTEND_PORT` | `80` | Порт фронтенда **на хосте**. Поменяйте при работе за реверс-прокси. |
| `GARAGE_S3_PORT` | `3900` | Порт Garage S3 на хосте. Должен совпадать с портом в `AWS_PUBLIC_ENDPOINT_URL`. |
| `API_CORS_ORIGINS` | `http://localhost` | Публичный адрес фронтенда. Несколько — через запятую: `http://a.com,https://a.com`. |

### База данных (PostgreSQL)

| Переменная | По умолчанию | Описание |
| --- | --- | --- |
| `DB_HOST` | `pg` | Хост БД (имя сервиса в Docker). |
| `DB_PORT` | `5432` | Порт БД. |
| `DB_DATABASE` | `kvault` | Имя базы. |
| `DB_USERNAME` | `postgres` | Пользователь. |
| `DB_PASSWORD` | `postgres` | **Смените на надёжный пароль.** |

### Redis

| Переменная | По умолчанию | Описание |
| --- | --- | --- |
| `REDIS_HOST` | `redis` | Хост Redis. |
| `REDIS_PORT` | `6379` | Порт Redis. |
| `REDIS_USER` | `redis` | Пользователь. |
| `REDIS_PASSWORD` | `redis` | **Смените на надёжный пароль.** |
| `REDIS_QUEUE_DB` | `0` | Номер БД Redis под очередь фоновых задач. |
| `REDIS_CACHE_DB` | `1` | Номер БД Redis под кэш. |

### Кэш

| Переменная | По умолчанию | Описание |
| --- | --- | --- |
| `CACHE_ENABLED` | `true` | Включить кэширование. |
| `CACHE_ITEMS_TTL` | `5m` | Время жизни кэша заметок. |
| `CACHE_FILES_TTL` | `10m` | Время жизни кэша файлов. |
| `CACHE_TAGS_TTL` | `15m` | Время жизни кэша тегов. |
| `CACHE_STOPWORDS_TTL` | `30m` | Время жизни кэша стоп-слов. |

### Хранилище (Garage / S3)

| Переменная | По умолчанию | Описание |
| --- | --- | --- |
| `AWS_ACCESS_KEY_ID` | — | Access Key из Garage (см. [шаг 4](#шаг-4-настройка-хранилища-garage-s3)). |
| `AWS_SECRET_ACCESS_KEY` | — | Secret Key из Garage. |
| `AWS_REGION` | `garage` | Регион — с Garage не менять. |
| `AWS_ENDPOINT_URL` | `http://garage:3900` | **Внутренний** адрес S3 — не менять при использовании Garage. |
| `AWS_S3_BUCKET` | `kvault-bucket` | Имя бакета. |
| `AWS_URL_EXPIRATION` | `60s` | Срок жизни ссылки на скачивание. |
| `AWS_VIEW_URL_EXPIRATION` | `24h` | Срок жизни ссылки на просмотр. |
| `AWS_PUBLIC_ENDPOINT_URL` | `http://localhost:3900` | **Публичный** адрес S3 — встраивается в ссылки на файлы для браузера. |

### Воркер

| Переменная | По умолчанию | Описание |
| --- | --- | --- |
| `WORKER_CONCURRENT_TASKS` | `10` | Число одновременно обрабатываемых фоновых задач (извлечение текста из PDF и веб-страниц). |

---

## Обновление и обслуживание

**Обновить до свежих образов:**

```bash
docker compose pull
docker compose up -d
```

Новые миграции БД применятся автоматически при старте (контейнер `migrate`).

**Резервное копирование.** Данные хранятся в именованных Docker-томах:

- `pg_data` — база данных (заметки, теги, стоп-слова, пользователи);
- `garage_meta`, `garage_data` — загруженные файлы.

Для бэкапа сохраняйте эти тома (и сам файл `.env` с ключами доступа). Дамп базы:

```bash
docker exec kvault_pg pg_dump -U postgres kvault > kvault-backup.sql
```

**Остановить / запустить:**

```bash
docker compose down      # остановить (тома сохраняются)
docker compose up -d     # запустить снова
```

---

## Решение частых проблем

**В браузере ошибки CORS, интерфейс не загружает данные.**
`API_CORS_ORIGINS` не совпадает с адресом сайта. Проверьте схему (`http`/`https`) и хост — они должны совпадать с тем, что в адресной строке. После правки `.env` выполните `docker compose up -d`.

**Файлы загружаются, но не скачиваются / не открываются.**
Неверный `AWS_PUBLIC_ENDPOINT_URL` или закрыт порт Garage. Адрес должен быть доступен **из браузера**, а порт `GARAGE_S3_PORT` — открыт наружу. См. [presigned-ссылки](#загрузка-файлов-и-presigned-ссылки).

**При загрузке файла — ошибка о размере / `413`.**
Поднимите лимит тела запроса в реверс-прокси (`client_max_body_size` для nginx).

**Загрузка файла проходит, но из него не извлекается текст.**
Убедитесь, что запущен контейнер `kvault_worker` (`docker compose ps`) — извлечение текста из PDF выполняет фоновый воркер. Из PDF без текстового слоя (сканы-картинки) текст извлечь нельзя.

**Файлы вообще не загружаются после установки.**
Скорее всего, не пройден [шаг 4](#шаг-4-настройка-хранилища-garage-s3): не инициализирован Garage или не заданы ключи `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`.
