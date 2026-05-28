# kvault

Self-hosted система управления знаниями. Хранение и поиск знаний из разнородных источников - текстовые заметки, PDF-файлы, веб-страницы по URL - с автоматической тегизацией и полнотекстовым поиском.

## Возможности

- Аутентификация по API-ключу
- Добавление записей: текстовые заметки, PDF-документы, веб-страницы по URL (в разработке)
- Автоматическая и ручная тегизация
- Полнотекстовый поиск, фильтрация по тегам, пагинация
- Корзина / восстановление / безвозвратное удаление
- Просмотр файлов через presigned URL

## Самохостинг

### Требования

- Docker + Docker Compose

### Быстрый старт

```bash
# 1. Скачать compose-файл и пример конфига
curl -O https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/docker-compose.yml
curl -O https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/.env.example
mv .env.example .env

# 2. Скачать необходимые конфиги
mkdir -p docker/redis docker/garage
curl -o docker/redis/redis.conf https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/docker/redis/redis.conf
curl -o docker/redis/entrypoint.sh https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/docker/redis/entrypoint.sh
curl -o docker/garage/garage.toml https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/docker/garage/garage.toml

# 3. Отредактировать .env - перед запуском рекомендуется изменить:
#    DB_PASSWORD, REDIS_PASSWORD

# 4. Запустить
docker compose pull
docker compose up -d

# 5. Выполнить настройку GarageHQ (см. далее)
```

Фронтенд доступен по адресу `http://ваш-сервер`. Порт по умолчанию - 80, настраивается через `FRONTEND_PORT`.

### Настройка Garage (S3)

Garage - S3-совместимое хранилище. При первом запуске нужно сгенерировать учётные данные:

```bash
# Алиас для удобства
alias garage="docker exec kvault_garage /garage"

# Получить ID узла
garage status

# Создать layout (заменить <node-id> на вывод выше)
# Флаг -c задаёт ёмкость узла
# Подробнее: https://garagehq.deuxfleurs.fr/documentation/quick-start/#creating-a-cluster-layout
garage layout assign -z dc1 -c 1G <node-id>
garage layout apply --version 1

# Создать credentials
# Сохраните секретный ключ, он будет недоступен позже
garage key create kvault-key
# → скопируйте Access Key ID и Secret Key в .env

# Создать бакет
garage bucket create kvault-bucket
garage bucket allow --read --write --owner kvault-bucket --key kvault-key
```

После обновления `.env` - перезапустить:

```bash
docker compose up -d
```

### Параметры конфигурации

**Общие**

| Переменная         | По умолчанию         | Описание                                                                                             |
| ------------------ | -------------------- | ---------------------------------------------------------------------------------------------------- |
| `REGISTRY`         | `gitverse.ru/qvarkk` | Реестр образов: `gitverse.ru/qvarkk` или `ghcr.io/qvarkk` (к ghrc могут быть проблемы доступа из РФ) |
| `DEBUG`            | `false`              | Режим отладки - установите `false` при развертке                                                     |
| `API_PORT`         | `6767`               | Порт API-сервера внутри контейнера. Не влияет на работу                                              |
| `FRONTEND_PORT`    | `80`                 | Порт фронтенда на хосте. Измените при использовании реверс-прокси                                    |
| `GARAGE_S3_PORT`   | `3900`               | Порт Garage S3 API на хосте. Должен совпадать с портом в `AWS_PUBLIC_ENDPOINT_URL`                   |
| `API_CORS_ORIGINS` | `http://localhost`   | Публичный URL фронтенда. Например: `http://myserver.com`                                             |

**База данных (PostgreSQL)**

| Переменная    | По умолчанию | Описание                      |
| ------------- | ------------ | ----------------------------- |
| `DB_HOST`     | `pg`         | Хост PostgreSQL               |
| `DB_PORT`     | `5432`       | Порт PostgreSQL               |
| `DB_DATABASE` | `kvault`     | Имя базы данных               |
| `DB_USERNAME` | `postgres`   | Имя пользователя              |
| `DB_PASSWORD` | -            | Пароль - используйте надёжный |

**Redis**

| Переменная       | По умолчанию | Описание                         |
| ---------------- | ------------ | -------------------------------- |
| `REDIS_HOST`     | `redis`      | Хост Redis                       |
| `REDIS_PORT`     | `6379`       | Порт Redis                       |
| `REDIS_USER`     | `redis`      | Имя пользователя                 |
| `REDIS_PASSWORD` | -            | Пароль - используйте надёжный    |
| `REDIS_QUEUE_DB` | `0`          | Номер БД Redis для очереди задач |
| `REDIS_CACHE_DB` | `1`          | Номер БД Redis для кэша          |

**Кэш**

| Переменная            | По умолчанию | Описание             |
| --------------------- | ------------ | -------------------- |
| `CACHE_ENABLED`       | `true`       | Включить кэширование |
| `CACHE_ITEMS_TTL`     | `5m`         | TTL кэша записей     |
| `CACHE_FILES_TTL`     | `10m`        | TTL кэша файлов      |
| `CACHE_TAGS_TTL`      | `15m`        | TTL кэша тегов       |
| `CACHE_STOPWORDS_TTL` | `30m`        | TTL кэша стоп-слов   |

**S3-хранилище (Garage)**
<sub><sup>Устанавливайте значения портов (цифры после ":") в данном разделе в соответствии со значением, указанным в параметре `GARAGE_S3_PORT`</sup></sub>

| Переменная                | По умолчанию            | Описание                                                                                |
| ------------------------- | ----------------------- | --------------------------------------------------------------------------------------- |
| `AWS_ACCESS_KEY_ID`       | -                       | Access Key - генерируется в Garage (см. «Настройка Garage»)                             |
| `AWS_SECRET_ACCESS_KEY`   | -                       | Secret Key - генерируется в Garage                                                      |
| `AWS_REGION`              | `garage`                | Регион - не менять при использовании Garage                                             |
| `AWS_ENDPOINT_URL`        | `http://garage:3900`    | Внутренний URL S3 - не изменять URL при использовании Garage                            |
| `AWS_S3_BUCKET`           | `kvault-bucket`         | Имя бакета                                                                              |
| `AWS_URL_EXPIRATION`      | `60s`                   | Срок действия presigned URL для скачивания                                              |
| `AWS_VIEW_URL_EXPIRATION` | `24h`                   | Срок действия presigned URL для просмотра                                               |
| `AWS_PUBLIC_ENDPOINT_URL` | `http://localhost:3900` | Публичный URL S3 - встраивается в ссылки на файлы. Например: `http://myserver.com:3900` |

**Воркер**

| Переменная                | По умолчанию | Описание                      |
| ------------------------- | ------------ | ----------------------------- |
| `WORKER_CONCURRENT_TASKS` | `10`         | Количество параллельных задач |

### Запуск за реверс-прокси

Установите `FRONTEND_PORT` на свободный порт, затем настройте реверс-прокси:

```bash
# .env
FRONTEND_PORT=8081
API_CORS_ORIGINS=https://kvault.yourdomain.com
```

Пример конфига nginx:

```nginx
server {
    listen 80;
    server_name kvault.yourdomain.com;

    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Разработка

```bash
# Скопировать и заполнить конфиг
cp .env.example .env

# Запустить инфраструктуру (pg, redis, garage)
make docker-up-infra

# Применить миграции
make migrate-up

# Запустить API и воркер
make run-api
make run-worker
```

## Стек

Go · Gin · PostgreSQL · Redis · Garage (S3) · Asynq · Vue 3 · Vite · shadcn
