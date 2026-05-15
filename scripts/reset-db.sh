#!/usr/bin/env bash
set -euo pipefail

echo "WARNING: This will DROP all IndraNet data and re-run migrations."
read -r -p "Are you sure? [y/N] " confirm

if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
  echo "Aborted."
  exit 0
fi

DB_NAME="${POSTGRES_DB:-indranet}"
DB_USER="${POSTGRES_USER:-indranet}"

echo "Dropping database $DB_NAME..."
docker compose exec -T postgres psql -U "$DB_USER" -c "DROP DATABASE IF EXISTS $DB_NAME;"
docker compose exec -T postgres psql -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;"

echo "Running migrations..."
cd packages/backend && make migrate && cd ../..

echo "✓ Database reset complete."
