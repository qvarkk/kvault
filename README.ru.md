# kvault

> [English](README.md)

Self-hosted система управления знаниями. Хранение и поиск знаний из разнородных источников — текстовые заметки, PDF-файлы, веб-страницы по URL — с автоматической тегизацией и полнотекстовым поиском.

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
curl -O https://raw.githubusercontent.com/qvarkk/kvault/main/docker-compose.yml
curl -O https://raw.githubusercontent.com/qvarkk/kvault/main/.env.example
mv .env.example .env

# 2. Скачать необходимые конфиги
mkdir -p docker/redis docker/garage
curl -o docker/redis/redis.conf https://raw.githubusercontent.com/qvarkk/kvault/main/docker/redis/redis.conf
curl -o docker/redis/entrypoint.sh https://raw.githubusercontent.com/qvarkk/kvault/main/docker/redis/entrypoint.sh
curl -o docker/garage/garage.toml https://raw.githubusercontent.com/qvarkk/kvault/main/docker/garage/garage.toml

# 3. Отредактировать .env — заполнить перед запуском:
#    DB_PASSWORD, REDIS_PASSWORD, AWS_PUBLIC_ENDPOINT_URL, API_CORS_ORIGINS
#
#    AWS_ACCESS_KEY_ID и AWS_SECRET_ACCESS_KEY берутся из Garage —
#    сначала запустите стек, сгенерируйте их (см. "Настройка Garage" ниже),
#    затем добавьте в .env и перезапустите.

# 4. Запустить
docker compose pull
docker compose up -d
```

Фронтенд доступен по адресу `http://ваш-сервер`. Порт по умолчанию — 80, настраивается через `FRONTEND_PORT`.

### Основные параметры конфигурации

| Переменная                                    | Описание                                                                                                 |
| --------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| `REGISTRY`                                    | Реестр образов: `ghcr.io/qvarkk` (GitHub) или `registry.gitlab.com/qvarkk` (GitLab, стабильнее в России) |
| `API_CORS_ORIGINS`                            | Публичный URL фронтенда, например `http://myserver.com`. При локальном деплое можно оставить по умолчанию |
| `AWS_PUBLIC_ENDPOINT_URL`                     | Публичный URL S3-хранилища, например `http://myserver.com:3900` — встраивается в ссылки на файлы. При локальном деплое можно оставить по умолчанию |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` | Credentials S3-хранилища — для Garage генерируются после первого запуска (см. ниже)                      |
| `DB_PASSWORD`                                 | Пароль PostgreSQL — используйте надёжный                                                                 |
| `REDIS_PASSWORD`                              | Пароль Redis — используйте надёжный                                                                      |
| `FRONTEND_PORT`                               | Порт фронтенда на хосте, по умолчанию `80`. Измените при использовании реверс-прокси.                   |
| `GARAGE_S3_PORT`                              | Порт Garage S3 API на хосте, по умолчанию `3900`. Должен совпадать с портом в `AWS_PUBLIC_ENDPOINT_URL`. |
| `DEBUG`                                       | Установите `false` в продакшене                                                                          |

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

### Настройка Garage (S3)

Garage — S3-совместимое хранилище. При первом запуске нужно сгенерировать учётные данные:

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
garage key create kvault-key
# → скопировать Access Key ID и Secret Key в .env

# Создать бакет
garage bucket create kvault-bucket
garage bucket allow --read --write --owner kvault-bucket --key kvault-key
```

После обновления `.env` — перезапустить:

```bash
docker compose up -d
```

Затем применить CORS-политику — нужна для просмотра файлов в браузере. Укажите ключи и URL из `AWS_PUBLIC_ENDPOINT_URL` (по умолчанию `http://localhost:3900`):

```bash
docker run --rm --network host \
  -e AWS_ACCESS_KEY_ID=<ваш-access-key-id> \
  -e AWS_SECRET_ACCESS_KEY=<ваш-secret-key> \
  amazon/aws-cli s3api put-bucket-cors \
    --endpoint-url <aws-public-endpoint-url> \
    --bucket kvault-bucket \
    --cors-configuration '{"CORSRules":[{"AllowedOrigins":["*"],"AllowedMethods":["GET"],"AllowedHeaders":["*"]}]}'
```

Перезапуск после этого не нужен.

## Разработка

```bash
# Скопировать и заполнить конфиг
cp .env.example .env

# Запустить инфраструктуру
make docker-up

# Подключиться к API и воркеру
make run-api
make run-worker

# Применить миграции
make migrate-up
```

## Стек

Go · Gin · PostgreSQL · Redis · Garage (S3) · Asynq · Vue 3 · Vite · shadcn
