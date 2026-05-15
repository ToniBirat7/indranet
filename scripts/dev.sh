#!/usr/bin/env bash
set -euo pipefail

# Start all local dev services in parallel
echo "Starting IndraNet dev environment..."

# Ensure Docker services are running
docker compose up -d

# Run backend, web in parallel using background jobs
(cd packages/backend && make dev) &
BACKEND_PID=$!

(cd packages/web && pnpm dev) &
WEB_PID=$!

echo "Backend PID: $BACKEND_PID"
echo "Web PID: $WEB_PID"
echo "Press Ctrl+C to stop all services"

trap "kill $BACKEND_PID $WEB_PID 2>/dev/null; exit 0" INT TERM

wait
