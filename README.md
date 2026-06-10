# kvault

**Self-hosted knowledge management system.** Keep notes, web pages, and PDF documents in one place — with automatic tagging and full-text search across all content.

Monorepo: [`backend/`](./backend) — REST API (Go), [`frontend/`](./frontend) — web UI (Vue 3). The whole stack is orchestrated by `docker-compose.yml` at the root.

---

## Features

- **Heterogeneous sources.** Markdown text notes, web pages by URL (with text, metadata, and preview extraction), PDF documents.
- **Full-text search.** Searches titles, note text, web page content, and text extracted from PDFs. Supports prefix search and relevance ranking.
- **Automatic tagging.** Tags are derived from the entry's content; the stopword list (RU/EN) is configurable.
- **Tag and stopword management.** Manual assignment, renaming, filtering by tags.
- **Trash bin.** Deletion with restore support and permanent purge.
- **File storage.** S3-compatible storage (Garage) with access via temporary presigned URLs.
- **Asynchronous processing.** Text extraction from PDFs and web pages runs in a background worker.

---

## Documentation

| Document                                   | Purpose                                                                  |
| ------------------------------------------ | ------------------------------------------------------------------------ |
| [DOCUMENTATION.md](./DOCUMENTATION.md)     | User guide: search, filters, tags, stopwords, common workflows.          |
| [HOSTING.md](./HOSTING.md)                 | Complete deployment guide: `.env`, domain, reverse proxy, HTTPS.         |
| [SECURITY.md](./SECURITY.md)               | Security model and data protection recommendations.                      |
| [CONTRIBUTING.md](./CONTRIBUTING.md)       | How to contribute.                                                       |
| [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md) | Contributor code of conduct.                                             |

---

## Quick start

Requires **Docker** and **Docker Compose**.

```bash
mkdir kvault && cd kvault

curl -O https://raw.githubusercontent.com/qvarkk/kvault/main/docker-compose.yml
curl -O https://raw.githubusercontent.com/qvarkk/kvault/main/setup.sh
sh setup.sh        # creates .env and generates secrets

docker compose up -d
```

Prebuilt images are pulled from GitHub Container Registry. The version is set by `KVAULT_VERSION` in `.env` (`latest` — latest release, `v1.0.0` — specific version). Garage storage is initialized automatically (the one-shot `garage-init` service creates the layout, access key, and bucket from the values in `.env`).

The frontend opens at `http://<server-address>` (port `80` by default).

> Domain setup, reverse proxy and HTTPS, plus a detailed description of the Garage step — in **[HOSTING.md](./HOSTING.md)**.

---

## Security

> [!WARNING]
> kvault is designed for **personal self-hosting in a trusted environment**. There is no data encryption; the server administrator can see all users' content. API keys are issued per device and expire after inactivity (30 days by default). **Do not expose the service to the open internet** — use a VPN or firewall. Details: [SECURITY.md](./SECURITY.md).

---

## Development

```bash
sh setup.sh            # creates .env and generates secrets

# Backend (commands run from backend/)
cd backend
make docker-up-infra   # start PostgreSQL, Redis, Garage
make migrate-up        # apply migrations
make run-api           # run the API
make run-worker        # run the background worker

# Frontend (in a separate terminal)
cd frontend
cp .env.example .env   # local API address for the dev server
npm install
npm run dev            # Vite dev server
```

The interactive API reference (Swagger) is available at `/swagger`. More details in [CONTRIBUTING.md](./CONTRIBUTING.md).

---

## Stack

**Backend:** Go · Gin · PostgreSQL · Redis · Garage (S3) · Asynq · sqlx + squirrel · Zap
**Frontend:** Vue 3 · TypeScript · Vite · Tailwind CSS

---

## License

[MIT](./LICENSE)
