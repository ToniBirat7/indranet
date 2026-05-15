#!/usr/bin/env bash
set -euo pipefail

echo "=== IndraNet Dev Environment Setup ==="

# Check dependencies
command -v go >/dev/null 2>&1 || { echo "ERROR: Go is not installed"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "ERROR: Node.js is not installed"; exit 1; }
command -v pnpm >/dev/null 2>&1 || { echo "ERROR: pnpm is not installed (npm i -g pnpm)"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "ERROR: Docker is not installed"; exit 1; }

echo "✓ Dependencies found"

# Copy env file
if [ ! -f .env ]; then
  cp .env.example .env
  echo "✓ Created .env from .env.example — edit with your secrets"
fi

# Install JS/TS dependencies
pnpm install
echo "✓ pnpm packages installed"

# Install Go dependencies
cd packages/backend && go mod download && cd ../..
echo "✓ Go modules downloaded"

# Start local services
docker compose up -d
echo "✓ docker-compose services started (postgres, redis, nats, minio)"

# Run DB migrations
echo "Waiting for postgres to be ready..."
sleep 3
cd packages/backend && make migrate && cd ../..
echo "✓ Database migrations applied"

echo ""
echo "=== Setup complete! ==="
echo "  Backend:  cd packages/backend && make dev"
echo "  Web:      cd packages/web && pnpm dev"
echo "  Client:   cd packages/client && pnpm dev"
