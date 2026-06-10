# Contributing to kvault

Thanks for your interest in kvault. Monorepo: the backend (REST API in Go) lives in [`backend/`](./backend), the frontend (Vue 3) in [`frontend/`](./frontend).

## Stack

**Backend:** Go · Gin · PostgreSQL · Redis · Garage (S3) · Asynq
**Frontend:** Vue 3 · TypeScript · Vite · Tailwind CSS

## Local environment

```bash
# Create .env with generated secrets (at the repo root)
sh setup.sh

# make commands run from backend/
cd backend

# Start the infrastructure (PostgreSQL, Redis, Garage)
make docker-up-infra

# Apply migrations
make migrate-up

# Run the API and the background worker (in separate terminals)
make run-api
make run-worker
```

Frontend (in a separate terminal):

```bash
cd frontend
cp .env.example .env   # local API address for the dev server
npm install
npm run dev            # Vite dev server
```

## Useful commands

```bash
make build-api        # build the API into bin/kvault_api
make build-worker     # build the worker into bin/kvault_worker
make migrate-up       # apply migrations
make migrate-down     # roll back one migration
make swagger          # regenerate Swagger documentation
make tidy             # go mod tidy
```

## How to contribute

1. Create a branch off `main`.
2. Make your changes; keep commits focused and write meaningful messages.
3. Make sure the project builds (`make build-api`, `make build-worker`).
4. If you change handler annotations, update Swagger: `make swagger`.
5. Open a pull request describing what changed and why.

## Code guidelines

- **Build DB queries** with `squirrel` — not with string formatting (SQL injection risk).
- Follow the layered architecture: `domain → repositories → services → handlers → routes`.
- Propagate errors through the typed layer errors (`repositories/errors.go`, `services/errors.go`); HTTP error responses are produced by middleware.
- Ship DB schema changes as a new migration in `migrations/` (golang-migrate format); never edit existing migrations.

## Reporting issues

For bugs and feature requests, open an issue with a description, reproduction steps, and expected behavior. Vulnerabilities — not in public issues, see [SECURITY.md](./SECURITY.md).

By participating in this project you agree to abide by the [Code of Conduct](./CODE_OF_CONDUCT.md).
