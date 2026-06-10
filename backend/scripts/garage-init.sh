#!/bin/sh
# One-shot Garage initialization: cluster layout, S3 access key, bucket.
# Runs against the garage service over RPC; needs the garage config at
# /etc/garage.toml and the garage metadata volume mounted (the CLI reads
# the node key from there). Idempotent: safe to run on every compose up.
set -e

: "${AWS_ACCESS_KEY_ID:?AWS_ACCESS_KEY_ID is not defined}"
: "${AWS_SECRET_ACCESS_KEY:?AWS_SECRET_ACCESS_KEY is not defined}"
: "${AWS_S3_BUCKET:?AWS_S3_BUCKET is not defined}"

i=0
until garage status >/dev/null 2>&1; do
  i=$((i + 1))
  if [ "$i" -gt 30 ]; then
    echo "garage-init: garage is unreachable after 30s, giving up" >&2
    exit 1
  fi
  sleep 1
done

if garage status | grep -q "NO ROLE ASSIGNED"; then
  NODE_ID=$(garage node id -q | cut -d@ -f1)
  garage layout assign -z dc1 -c "${GARAGE_NODE_CAPACITY:-1G}" "$NODE_ID"
  garage layout apply --version 1
fi

garage key info "$AWS_ACCESS_KEY_ID" >/dev/null 2>&1 ||
  garage key import --yes -n kvault-key "$AWS_ACCESS_KEY_ID" "$AWS_SECRET_ACCESS_KEY"

garage bucket info "$AWS_S3_BUCKET" >/dev/null 2>&1 ||
  garage bucket create "$AWS_S3_BUCKET"

garage bucket allow --read --write --owner "$AWS_S3_BUCKET" --key "$AWS_ACCESS_KEY_ID" >/dev/null

echo "garage-init: done"
