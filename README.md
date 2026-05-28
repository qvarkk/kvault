# kvault

**Самостоятельно хостируемая система управления знаниями.** Храните заметки, веб-страницы и PDF-документы в одном месте — с автоматической тегизацией и полнотекстовым поиском по всему содержимому.

Этот репозиторий содержит **бэкенд** (REST API). Веб-интерфейс — в отдельном репозитории [kvault-frontend](https://gitverse.ru/qvarkk/kvault-frontend).

---

## Возможности

- **Разнородные источники.** Текстовые заметки в Markdown, веб-страницы по ссылке (с извлечением текста, метаданных и превью), PDF-документы.
- **Полнотекстовый поиск.** Поиск по заголовкам, тексту заметок, содержимому веб-страниц и тексту из PDF. Поддержка префиксного поиска и ранжирования по релевантности.
- **Автоматическая тегизация.** Теги подбираются по содержимому записи; список стоп-слов (RU/EN) настраивается.
- **Управление тегами и стоп-словами.** Ручная привязка, переименование, фильтрация по тегам.
- **Корзина.** Удаление с возможностью восстановления и безвозвратной очистки.
- **Хранение файлов.** S3-совместимое хранилище (Garage) с доступом по временным presigned-ссылкам.
- **Асинхронная обработка.** Извлечение текста из PDF и веб-страниц выполняется фоновым воркером.

---

## Документация

| Документ | Назначение |
| --- | --- |
| [DOCUMENTATION.md](./DOCUMENTATION.md) | Руководство пользователя: поиск, фильтры, теги, стоп-слова, рабочие сценарии. |
| [HOSTING.md](./HOSTING.md) | Полное руководство по развёртыванию: `.env`, домен, реверс-прокси, HTTPS. |
| [SECURITY.md](./SECURITY.md) | Модель безопасности и рекомендации по защите данных. |
| [CONTRIBUTING.md](./CONTRIBUTING.md) | Как участвовать в разработке бэкенда. |
| [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md) | Кодекс поведения участников. |

---

## Быстрый старт

Требуется **Docker** и **Docker Compose**.

```bash
mkdir kvault && cd kvault

# Compose-файл и пример конфига
curl -O https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/docker-compose.yml
curl -O https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/.env.example
mv .env.example .env

# Конфиги Redis и Garage
mkdir -p docker/redis docker/garage
curl -o docker/redis/redis.conf    https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/docker/redis/redis.conf
curl -o docker/redis/entrypoint.sh https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/docker/redis/entrypoint.sh
curl -o docker/garage/garage.toml  https://gitverse.ru/api/repos/qvarkk/kvault/raw/branch/main/docker/garage/garage.toml

# Перед запуском смените в .env как минимум DB_PASSWORD и REDIS_PASSWORD

docker compose up -d --build
```

Образы собираются из исходников — внешний реестр не нужен. Версию задаёт `KVAULT_VERSION` в `.env` (git-тег/ветка, по умолчанию `main`).

**Инициализация хранилища Garage (обязательно — без неё не работает загрузка файлов):**

```bash
alias garage="docker exec kvault_garage /garage"

# Создать layout (подставьте <node-id> из вывода garage status)
garage status
garage layout assign -z dc1 -c 1G <node-id>
garage layout apply --version 1

# Создать ключ (сохраните Secret Key — позже недоступен) и бакет
garage key create kvault-key
garage bucket create kvault-bucket
garage bucket allow --read --write --owner kvault-bucket --key kvault-key

# Впишите Access Key ID и Secret Key в .env (AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY), затем:
docker compose up -d
```

Фронтенд откроется на `http://<адрес-сервера>` (порт `80` по умолчанию).

> Настройка домена, реверс-прокси и HTTPS, а также подробное описание шага с Garage — в **[HOSTING.md](./HOSTING.md)**.

---

## Безопасность

> [!WARNING]
> kvault рассчитан на **личный самостоятельный хостинг в доверенном окружении**. Шифрования данных нет, администратор сервера видит содержимое всех пользователей, API-ключи постоянны. **Не выставляйте сервис в открытый интернет** — используйте VPN или файрвол. Подробнее: [SECURITY.md](./SECURITY.md).

---

## Разработка

```bash
cp .env.example .env

make docker-up-infra   # поднять PostgreSQL, Redis, Garage
make migrate-up        # применить миграции
make run-api           # запустить API
make run-worker        # запустить фоновый воркер
```

Интерактивный справочник API (Swagger) доступен по адресу `/swagger`. Подробнее — в [CONTRIBUTING.md](./CONTRIBUTING.md).

---

## Стек

Go · Gin · PostgreSQL · Redis · Garage (S3) · Asynq · sqlx + squirrel · Zap

---

## Лицензия

[MIT](./LICENSE)
