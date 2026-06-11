#!/usr/bin/env bash
# First-time setup: generates lockfiles (if missing) and starts the dev stack.
# Requires only Docker — no local Go, Node, or pnpm needed.
set -euo pipefail

REPO="$(git rev-parse --show-toplevel)"
cd "$REPO"

echo "=== IndraNet Dev Environment Setup (Docker-only) ==="

command -v docker >/dev/null 2>&1 || { echo "ERROR: Docker is not installed"; exit 1; }
command -v git    >/dev/null 2>&1 || { echo "ERROR: git is not installed"; exit 1; }

echo "✓ Docker found"

# Create .env from example if missing
if [ ! -f .env ]; then
  cp .env.example .env
  echo "✓ Created .env from .env.example — edit with your secrets before starting"
fi

# Generate lockfiles if they don't exist (uses Docker — no local tooling)
MISSING_LOCKFILES=0
[ ! -f packages/backend/go.sum ] && MISSING_LOCKFILES=1
[ ! -f pnpm-lock.yaml ]          && MISSING_LOCKFILES=1

if [ "$MISSING_LOCKFILES" -eq 1 ]; then
  echo "Generating lockfiles via Docker (this runs once)..."
  bash scripts/gen-lockfiles.sh
fi

echo ""
echo "=== Setup complete! ==="
echo ""
echo "Start the dev stack:"
echo "  bash scripts/dev.sh"
echo ""
echo "Or start infra only:"
echo "  docker compose up -d postgres redis nats minio"
