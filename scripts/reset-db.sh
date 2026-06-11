#!/usr/bin/env bash
# Reset the local dev database. Migrations run automatically on next backend start.
set -euo pipefail

echo "WARNING: This will DROP all IndraNet data."
read -r -p "Are you sure? [y/N] " confirm

if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
  echo "Aborted."
  exit 0
fi

DB_NAME="${POSTGRES_DB:-indranet}"
DB_USER="${POSTGRES_USER:-indranet}"

echo "Dropping and recreating database $DB_NAME..."
docker compose exec -T postgres psql -U "$DB_USER" -c "DROP DATABASE IF EXISTS $DB_NAME;"
docker compose exec -T postgres psql -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;"

echo "Done. Restart the backend (docker compose --profile dev restart backend-dev) to run migrations."
