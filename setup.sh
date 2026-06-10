#!/bin/sh
# Created .env for deployment: copies .env.example and
# fills in secrets (DB, Redis, Garage RPC, S3 key).
# Existing .env does not get overriden
set -eu

ENV_EXAMPLE_URL="https://raw.githubusercontent.com/qvarkk/kvault/main/.env.example"

if [ -f .env ]; then
  echo "setup: .env already exists, not overriding" >&2
  exit 1
fi

if [ ! -f .env.example ]; then
  echo "setup: downloading .env.example"
  curl -fsSO "$ENV_EXAMPLE_URL"
fi

rand_hex() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex "$1"
  else
    head -c "$1" /dev/urandom | od -An -tx1 | tr -d ' \n'
  fi
}

cp .env.example .env

set_var() {
  sed "s|^$1=.*|$1=\"$2\"|" .env > .env.tmp && mv .env.tmp .env
}

set_var DB_PASSWORD           "$(rand_hex 24)"
set_var REDIS_PASSWORD        "$(rand_hex 24)"
set_var GARAGE_RPC_SECRET     "$(rand_hex 32)"
set_var AWS_ACCESS_KEY_ID     "GK$(rand_hex 12)"
set_var AWS_SECRET_ACCESS_KEY "$(rand_hex 32)"

chmod 600 .env

echo "setup: .env was created, secrets generated"
echo "setup: to access from anything other than http://localhost change .env"
echo "       API_CORS_ORIGINS and AWS_PUBLIC_ENDPOINT_URL (look HOSTING.md)"
echo "setup: next step - docker compose up -d"
