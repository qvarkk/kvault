#!/bin/bash
set -euo pipefail

echo "[init] running terraform init..."

cd /app/terraform-state
cp /app/terraform/main.tf .

tflocal init -input=false
tflocal apply -auto-approve

echo "[init] terraform done"