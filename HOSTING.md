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
- [Доступ через Tailscale (VPN)](#доступ-через-tailscale-vpn)
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
- _Опционально, но рекомендуется:_ доменное имя и реверс-прокси (nginx, Caddy, Traefik) для HTTPS.

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

> **Образы собираются из исходников.** kvault не зависит от внешнего реестра образов — при первом запуске Docker сам скачивает исходный код из репозитория и собирает образы локально. Какую версию собирать, задаёт переменная `KVAULT_VERSION` в `.env` (любой git-тег, ветка или коммит; по умолчанию `main` — последняя версия). Подробнее — в разделе [Версии и обновление](#обновление-и-обслуживание).

---

## Шаг 2. Настройка .env

Откройте `.env` и **обязательно** измените перед запуском в продакшене:

| Переменная                | Зачем менять                                                                                    |
| ------------------------- | ----------------------------------------------------------------------------------------------- |
| `DB_PASSWORD`             | Пароль базы данных — задайте надёжный.                                                          |
| `REDIS_PASSWORD`          | Пароль Redis — задайте надёжный.                                                                |
| `API_CORS_ORIGINS`        | Публичный адрес, по которому будет открываться kvault (см. ниже).                               |
| `AWS_PUBLIC_ENDPOINT_URL` | Публичный адрес хранилища файлов (см. [presigned-ссылки](#загрузка-файлов-и-presigned-ссылки)). |

Надёжные пароли удобно генерировать через `openssl` — по 24 случайных байта (48 hex-символов):

```bash
openssl rand -hex 24
```

Запустите команду отдельно для `DB_PASSWORD` и `REDIS_PASSWORD` и впишите полученные значения в `.env`.

Ключи доступа к хранилищу (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`) вы заполните позже — на [шаге 4](#шаг-4-настройка-хранилища-garage-s3), после генерации их в Garage.

Полное описание всех переменных — в [справочнике ниже](#справочник-переменных-env).

---

## Шаг 3. Первый запуск

```bash
docker compose up -d --build
```

Флаг `--build` собирает образы из исходников (Docker скачает код из репозитория). Первая сборка занимает несколько минут — дальше образы кэшируются.

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

## Доступ через Tailscale (VPN)

Если kvault крутится на домашней машине или ноутбуке без публичного IP и домена, самый простой и безопасный вариант — открыть его **только внутри tailnet** через [Tailscale](https://tailscale.com). Tailscale поднимает приватную WireGuard-сеть: сервис видят лишь ваши доверенные устройства, наружу ничего не торчит. Это рекомендуемый способ для личного хостинга и конкретная реализация требования из [SECURITY.md](./SECURITY.md) «не выставлять сервис в открытый интернет».

Установите Tailscale на сервер **и** на устройства, с которых будете заходить, и авторизуйтесь:

```bash
# на сервере
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up

# узнать tailnet-адрес сервера (вида 100.x.y.z)
tailscale ip -4
```

Дальше — два варианта.

### Вариант 1. Обычный HTTP по тайлнету (просто)

Привяжите публикуемые порты к **tailnet-адресу** сервера. Тогда контейнеры слушают только на интерфейсе Tailscale — в локальной сети и наружу они недоступны:

```bash
# .env — подставьте свой tailnet-адрес из `tailscale ip -4` (например 100.72.84.64)
FRONTEND_PORT=100.72.84.64:8000
GARAGE_S3_PORT=100.72.84.64:3900
API_CORS_ORIGINS=http://100.72.84.64:8000
AWS_PUBLIC_ENDPOINT_URL=http://100.72.84.64:3900
```

Перезапустите: `docker compose up -d`.

Заходите с любого устройства в тайлнете по `http://100.72.84.64:8000`. Трафик внутри tailnet шифруется WireGuard, поэтому обычного HTTP достаточно; реверс-прокси и сертификаты не нужны. И фронтенд, и S3 работают по HTTP на одном адресе — проблемы [mixed content](#s3-по-https) не возникает.

### Вариант 2. HTTPS через Tailscale Serve

Если хочется аккуратный адрес `https://<имя>.<tailnet>.ts.net` с валидным сертификатом, используйте [Tailscale Serve](https://tailscale.com/kb/1242/tailscale-serve). В админке tailnet включите **MagicDNS** и **HTTPS Certificates**.

Привяжите порты контейнеров к localhost — наружу их отдаёт уже Tailscale Serve:

```bash
# .env — порты только на localhost
FRONTEND_PORT=127.0.0.1:8000
GARAGE_S3_PORT=127.0.0.1:3900
API_CORS_ORIGINS=https://<имя>.<tailnet>.ts.net
AWS_PUBLIC_ENDPOINT_URL=https://<имя>.<tailnet>.ts.net:3900
```

Перезапустите (`docker compose up -d`) и поднимите Serve для фронтенда и S3:

```bash
tailscale serve --bg --https=443  http://127.0.0.1:8000   # фронтенд
tailscale serve --bg --https=3900 http://127.0.0.1:3900   # Garage S3
```

S3 публикуется на собственном HTTPS-порту `3900`, поэтому presigned-ссылки остаются валидными без переписывания пути. И фронтенд, и хранилище работают по HTTPS — [mixed content](#s3-по-https) не возникает. Проверить активные маршруты: `tailscale serve status`.

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

> **При переходе фронтенда на HTTPS, хранилище S3 тоже должно быть переведено на HTTPS для предотвращения mixed content.** См. пример ниже.

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

### S3 по HTTPS

> **Почему это вообще нужно.** Если фронтенд открывается по **HTTPS**, а `AWS_PUBLIC_ENDPOINT_URL` указывает на **HTTP** (`http://194.0.2.10:3900`), браузер заблокирует переход по presigned-ссылке как **mixed content** — файлы перестанут скачиваться и загружаться.

Самый простой способ перевести S3 на HTTPS без отдельного поддомена — проксировать Garage на **префиксе пути** того же домена, на котором уже работает фронтенд. Браузер ходит на `https://kvault.example.com/<бакет>/...`, nginx терминирует TLS и проксирует на локальный Garage.

**1. В `.env`** укажите публичный адрес без порта (имя бакета попадёт в путь автоматически):

```bash
AWS_S3_BUCKET=kvault-bucket
AWS_PUBLIC_ENDPOINT_URL=https://kvault.example.com
```

**2. В конфиге nginx** добавьте `location` для бакета в тот же `server`-блок на `443`, где проксируется фронтенд:

```nginx
# HTTP → HTTPS
server {
    # Автоматически сгенерировано certbot
    listen 80;
    server_name kvault.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name kvault.example.com;

    # Автоматически сгенерировано certbot
    ssl_certificate     /etc/letsencrypt/live/kvault.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/kvault.example.com/privkey.pem;

    # Файлы могут быть крупными — поднимаем лимит тела запроса
    client_max_body_size 100M;

    location / {
        proxy_pass http://localhost:8081;          # контейнер фронтенда
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /kvault-bucket/ {
        proxy_pass http://localhost:3900;           # локальный Garage S3
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Обязательно для chunked-загрузок в S3:
        # иначе nginx буферизует тело и Garage отвергает запрос
        proxy_buffering off;
        proxy_request_buffering off;
    }
}
```

При этом:

- Путь в `location` (`/kvault-bucket/`) должен **совпадать с именем бакета** `AWS_S3_BUCKET` — presigned-ссылка имеет вид `https://kvault.example.com/kvault-bucket/<ключ>?...`.
- Порт `3900` **не нужно** открывать наружу: браузер ходит на `443`, а Garage слушает только локально.
- `proxy_buffering off` / `proxy_request_buffering off` критичны для загрузки файлов (chunked upload).

Готовый пример конфига — в [`deploy/nginx.example.conf`](deploy/nginx.example.conf).

---

## Справочник переменных .env

### Общие

| Переменная             | По умолчанию       | Описание                                                                                                                                                                                                             |
| ---------------------- | ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `KVAULT_VERSION`       | `main`             | Версия для сборки: git-тег, ветка или коммит (напр. `v0.1.0`).                                                                                                                                                       |
| `KVAULT_REPO`          | gitverse           | Репозиторий бэкенда. Переопределяйте, только если зеркалите код (напр. на GitHub).                                                                                                                                   |
| `KVAULT_FRONTEND_REPO` | gitverse           | Репозиторий фронтенда. Аналогично.                                                                                                                                                                                   |
| `DEBUG`                | `false`            | Режим отладки — в продакшене держите `false`.                                                                                                                                                                        |
| `API_PORT`             | `6767`             | Порт API **внутри** сети Docker. Наружу обычно не публикуется.                                                                                                                                                       |
| `FRONTEND_PORT`        | `80`               | Порт фронтенда **на хосте**. Поменяйте при работе за реверс-прокси. Можно указать в форме `IP:порт`, чтобы слушать только на одном интерфейсе — напр. tailnet-адресе (см. [Tailscale](#доступ-через-tailscale-vpn)). |
| `GARAGE_S3_PORT`       | `3900`             | Порт Garage S3 на хосте. Должен совпадать с портом в `AWS_PUBLIC_ENDPOINT_URL`. Тоже поддерживает форму `IP:порт`.                                                                                                   |
| `API_CORS_ORIGINS`     | `http://localhost` | Публичный адрес фронтенда. Несколько — через запятую: `http://a.com,https://a.com`.                                                                                                                                  |

### База данных (PostgreSQL)

| Переменная    | По умолчанию | Описание                        |
| ------------- | ------------ | ------------------------------- |
| `DB_HOST`     | `pg`         | Хост БД (имя сервиса в Docker). |
| `DB_PORT`     | `5432`       | Порт БД.                        |
| `DB_DATABASE` | `kvault`     | Имя базы.                       |
| `DB_USERNAME` | `postgres`   | Пользователь.                   |
| `DB_PASSWORD` | `postgres`   | **Смените на надёжный пароль.** |

### Redis

| Переменная       | По умолчанию | Описание                                  |
| ---------------- | ------------ | ----------------------------------------- |
| `REDIS_HOST`     | `redis`      | Хост Redis.                               |
| `REDIS_PORT`     | `6379`       | Порт Redis.                               |
| `REDIS_USER`     | `redis`      | Пользователь.                             |
| `REDIS_PASSWORD` | `redis`      | **Смените на надёжный пароль.**           |
| `REDIS_QUEUE_DB` | `0`          | Номер БД Redis под очередь фоновых задач. |
| `REDIS_CACHE_DB` | `1`          | Номер БД Redis под кэш.                   |

### Кэш

| Переменная            | По умолчанию | Описание                    |
| --------------------- | ------------ | --------------------------- |
| `CACHE_ENABLED`       | `true`       | Включить кэширование.       |
| `CACHE_ITEMS_TTL`     | `5m`         | Время жизни кэша заметок.   |
| `CACHE_FILES_TTL`     | `10m`        | Время жизни кэша файлов.    |
| `CACHE_TAGS_TTL`      | `15m`        | Время жизни кэша тегов.     |
| `CACHE_STOPWORDS_TTL` | `30m`        | Время жизни кэша стоп-слов. |

### Хранилище (Garage / S3)

| Переменная                | По умолчанию            | Описание                                                                  |
| ------------------------- | ----------------------- | ------------------------------------------------------------------------- |
| `AWS_ACCESS_KEY_ID`       | —                       | Access Key из Garage (см. [шаг 4](#шаг-4-настройка-хранилища-garage-s3)). |
| `AWS_SECRET_ACCESS_KEY`   | —                       | Secret Key из Garage.                                                     |
| `AWS_REGION`              | `garage`                | Регион — с Garage не менять.                                              |
| `AWS_ENDPOINT_URL`        | `http://garage:3900`    | **Внутренний** адрес S3 — не менять при использовании Garage.             |
| `AWS_S3_BUCKET`           | `kvault-bucket`         | Имя бакета.                                                               |
| `AWS_URL_EXPIRATION`      | `60s`                   | Срок жизни ссылки на скачивание.                                          |
| `AWS_VIEW_URL_EXPIRATION` | `24h`                   | Срок жизни ссылки на просмотр.                                            |
| `AWS_PUBLIC_ENDPOINT_URL` | `http://localhost:3900` | **Публичный** адрес S3 — встраивается в ссылки на файлы для браузера.     |

### Воркер

| Переменная                | По умолчанию | Описание                                                                                  |
| ------------------------- | ------------ | ----------------------------------------------------------------------------------------- |
| `WORKER_CONCURRENT_TASKS` | `10`         | Число одновременно обрабатываемых фоновых задач (извлечение текста из PDF и веб-страниц). |
| `WORKER_MAX_RETRIES`      | `3`          | Максимум повторных попыток упавшей фоновой задачи.                                        |
| `WORKER_RETRY_TIMEOUT`    | `5m`         | Дедлайн одной попытки; при превышении задача повторяется.                                 |

### Аутентификация

| Переменная         | По умолчанию | Описание                                                                                             |
| ------------------ | ------------ | ---------------------------------------------------------------------------------------------------- |
| `AUTH_API_KEY_TTL` | `720h`       | Срок жизни API-ключа. Окно скользящее: ключ истекает через это время после последнего использования. |

---

## Обновление и обслуживание

**Выбор версии.** Версию задаёт `KVAULT_VERSION` в `.env`. Для воспроизводимого развёртывания фиксируйте конкретный тег:

```bash
# .env
KVAULT_VERSION=v0.1.0
```

Значение `main` (по умолчанию) собирает последнее состояние кода. Список версий — в разделе Releases репозитория.

**Обновить / пересобрать:**

```bash
docker compose up -d --build
```

Команда заново скачивает исходники нужной версии и пересобирает образы. Новые миграции БД применятся автоматически при старте (контейнер `migrate`).

> Откат к прежней версии: верните прежний `KVAULT_VERSION` и снова выполните `docker compose up -d --build`.

**Резервное копирование.** Данные лежат в двух местах: база PostgreSQL (заметки, теги, стоп-слова, пользователи) и хранилище Garage (загруженные файлы). Всегда сохраняйте также сам файл `.env` — в нём пароли и ключи доступа.

База PostgreSQL бэкапится «на лету», останавливать её не нужно:

```bash
docker exec kvault_pg pg_dump -U postgres kvault > kvault-backup.sql
```

Файлы Garage — в зависимости от того, как смонтировано хранилище.

> **Останавливайте Garage перед копированием.** Каталоги `meta`/`data` — это активная база; копия «на ходу» может оказаться несогласованной. `docker compose stop garage` → копирование → `docker compose start garage`.

**Именованные тома (по умолчанию).** Garage хранит данные в томах `garage_meta` и `garage_data`. Упаковать их в архив:

```bash
docker compose stop garage
docker run --rm -v garage_meta:/m -v garage_data:/d -v "$PWD:/backup" \
  alpine tar czf /backup/garage.tar.gz -C / m d
docker compose start garage
```

**Bind-mount каталоги (удобнее для бэкапов).** Если хранить файлы Garage в обычных каталогах рядом с `docker-compose.yml`, бэкап сводится к копированию папки. Замените в `docker-compose.yml` тома Garage на bind-mount:

```yaml
garage:
  volumes:
    - "./docker/garage/garage.toml:/etc/garage.toml"
    - ./data/garage/meta:/var/lib/garage/meta
    - ./data/garage/data:/var/lib/garage/data
```

и уберите `garage_meta` / `garage_data` из верхнего блока `volumes:`. Если разворачивание уже работает на томах, один раз перенесите данные: остановите Garage и скопируйте содержимое старых томов в `./data/garage/`. Дальше бэкап — обычный архив каталога:

```bash
docker compose stop garage
tar czf garage-backup.tar.gz ./data/garage
docker compose start garage
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
