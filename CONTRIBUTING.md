# Участие в разработке kvault

Спасибо за интерес к kvault. Монорепозиторий: бэкенд (REST API на Go) — в [`backend/`](./backend), фронтенд (Vue 3) — в [`frontend/`](./frontend).

## Стек

**Бэкенд:** Go · Gin · PostgreSQL · Redis · Garage (S3) · Asynq
**Фронтенд:** Vue 3 · TypeScript · Vite · Tailwind CSS

## Локальное окружение

```bash
# Скопировать и заполнить конфиг (в корне репозитория)
cp .env.example .env

# Команды make выполняются из backend/
cd backend

# Поднять инфраструктуру (PostgreSQL, Redis, Garage)
make docker-up-infra

# Применить миграции
make migrate-up

# Запустить API и фоновый воркер (в отдельных терминалах)
make run-api
make run-worker
```

Фронтенд (в отдельном терминале):

```bash
cd frontend
npm install
npm run dev    # dev-сервер Vite
```

## Полезные команды

```bash
make build-api        # собрать API в bin/kvault_api
make build-worker     # собрать воркер в bin/kvault_worker
make migrate-up       # применить миграции
make migrate-down     # откатить одну миграцию
make swagger          # перегенерировать Swagger-документацию
make tidy             # go mod tidy
```

## Как внести вклад

1. Создайте ветку от `main`.
2. Внесите изменения; держите коммиты сфокусированными, пишите осмысленные сообщения.
3. Убедитесь, что проект собирается (`make build-api`, `make build-worker`).
4. При изменении аннотаций обработчиков обновите Swagger: `make swagger`.
5. Откройте pull request с описанием сути и причины изменений.

## Правила по коду

- **Запросы к БД** стройте через `squirrel` — не через форматирование строк (риск SQL-инъекций).
- Соблюдайте слоистую архитектуру: `domain → repositories → services → handlers → routes`.
- Ошибки прокидывайте через типизированные ошибки слоёв (`repositories/errors.go`, `services/errors.go`); HTTP-ответы об ошибках формирует middleware.
- Изменения схемы БД оформляйте новой миграцией в `migrations/` (формат golang-migrate), не правьте существующие миграции.

## Сообщения об ошибках

Для багов и предложений создавайте issue с описанием, шагами воспроизведения и ожидаемым поведением. Уязвимости — не в публичные issue, см. [SECURITY.md](./SECURITY.md).

Участвуя в проекте, вы соглашаетесь соблюдать [Кодекс поведения](./CODE_OF_CONDUCT.md).
