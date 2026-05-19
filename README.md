# kvault

> [Русский](README.ru.md)

Self-hosted knowledge management system. Store and search knowledge from heterogeneous sources — text notes, PDFs, web URLs — with automatic tagging and full-text search.

## Features

- Authentication via API key
- Add entries: text notes, PDF files, web URLs (TBD)
- Autotagging and manual tags
- Full-text search, tag filtering, pagination
- Trash / restore / permanent deletion
- File viewing with presigned URLs

## Self-hosting

### Prerequisites

- Docker + Docker Compose

### Quick start

```bash
# 1. Get the compose file and example config
curl -O https://raw.githubusercontent.com/qvarkk/kvault/main/docker-compose.yml
curl -O https://raw.githubusercontent.com/qvarkk/kvault/main/.env.example
mv .env.example .env

# 2. Also grab the required config files
mkdir -p docker/redis docker/garage
curl -o docker/redis/redis.conf https://raw.githubusercontent.com/qvarkk/kvault/main/docker/redis/redis.conf
curl -o docker/redis/entrypoint.sh https://raw.githubusercontent.com/qvarkk/kvault/main/docker/redis/entrypoint.sh
curl -o docker/garage/garage.toml https://raw.githubusercontent.com/qvarkk/kvault/main/docker/garage/garage.toml

# 3. Edit .env - set these before starting:
#    DB_PASSWORD
#    REDIS_PASSWORD
#    AWS_PUBLIC_ENDPOINT_URL
#    API_CORS_ORIGINS
#
#    AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY come from Garage -
#    start the app first, generate them (see "Garage setup" below),
#    then add them to .env and restart

# 4. Start
docker compose pull
docker compose up -d
```

Frontend available at `http://your-server` (or `http://localhost`). Default port is 80, configurable via `FRONTEND_PORT`.

### Garage (S3) setup

Garage is the S3-compatible storage backend. On first run, generate credentials:

```bash
# Set up alias for convenience
alias garage="docker exec kvault_garage /garage"

# Get the node ID
garage status

# Create layout (replace <node-id> with output above)
# Modify -c flag to change capacity of the node
# Reference: https://garagehq.deuxfleurs.fr/documentation/quick-start/#creating-a-cluster-layout
garage layout assign -z dc1 -c 1G <node-id>
garage layout apply --version 1

# Create credentials
garage key create kvault-key
# → copy Access Key ID and Secret Key into .env

# Create bucket
garage bucket create kvault-bucket
garage bucket allow --read --write --owner kvault-bucket --key kvault-key
```

After updating `.env` with the credentials, restart:

```bash
docker compose up -d
```

Then apply the CORS policy with your access keys and URL (http://localhost:3900 by default). It's required for file viewing in browser:

```bash
docker run --rm --network host \
  -e AWS_ACCESS_KEY_ID=<your-access-key-id> \
  -e AWS_SECRET_ACCESS_KEY=<your-secret-key> \
  amazon/aws-cli s3api put-bucket-cors \
    --endpoint-url <aws-public-endpoint-url> \
    --bucket kvault-bucket \
    --cors-configuration '{"CORSRules":[{"AllowedOrigins":["*"],"AllowedMethods":["GET"],"AllowedHeaders":["*"]}]}'
```

After that, no restart needed.

### Key configuration

| Variable                                      | Description                                                                                                                |
| --------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `REGISTRY`                                    | Image registry: `ghcr.io/qvarkk` (GitHub) or `registry.gitlab.com/qvarkk` (GitLab)                                         |
| `API_CORS_ORIGINS`                            | Your frontend's public URL, e.g. `http://myserver.com`. Leave this as default for local deploy                             |
| `AWS_PUBLIC_ENDPOINT_URL`                     | Public URL of S3 storage, e.g. `http://myserver.com:3900` — embedded in file links. Also leave as default for local deploy |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` | S3 compatible storage credentials — for garage generate after first run (see below)                                        |
| `DB_PASSWORD`                                 | PostgreSQL password — use a strong one                                                                                     |
| `REDIS_PASSWORD`                              | Redis password — use a strong one                                                                                          |
| `FRONTEND_PORT`                               | Host port for the frontend, default `80`. Change if running behind a reverse proxy                                         |
| `GARAGE_S3_PORT`                              | Host port for Garage S3 API, default `3900`. Must match the port in `AWS_PUBLIC_ENDPOINT_URL`                              |
| `DEBUG`                                       | Set to `false` in production                                                                                               |

### Running behind reverse proxy

Set `FRONTEND_PORT` to a free port, then point your reverse proxy to it:

```bash
# .env
FRONTEND_PORT=8081
API_CORS_ORIGINS=https://kvault.yourdomain.com
```

Example nginx vhost:

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

## Development

```bash
# Copy and fill config
cp .env.example .env

# Start infrastructure
make docker-up

# Attach to API and worker
make run-api
make run-worker

# Apply migrations
make migrate-up
```

## Stack

Go · Gin · PostgreSQL · Redis · Garage (S3) · Asynq · Vue 3 · Vite · shadcn
