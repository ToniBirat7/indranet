#!/usr/bin/env bash
# Starts the full dev environment using Docker Compose only. No local tools needed.
set -euo pipefail

echo "Starting IndraNet dev environment..."

# Infra + app (hot-reload backend via air)
docker compose --profile dev up -d

echo ""
echo "Services:"
echo "  Web frontend:         http://localhost:3000"
echo "  Backend (hot reload): http://localhost:8080"
echo "  PostgreSQL:           localhost:5432"
echo "  Redis:                localhost:6379"
echo "  NATS:                 localhost:4222"
echo "  MinIO:                http://localhost:9001"
echo "  pgAdmin:              http://localhost:5050"
echo ""
echo "Logs (backend):  docker compose logs -f backend-dev"
echo "Logs (web):      docker compose logs -f web-dev"
echo "Stop:            docker compose --profile dev down"
