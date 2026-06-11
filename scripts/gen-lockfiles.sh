#!/usr/bin/env bash
# Generates all lockfiles (go.sum, pnpm-lock.yaml) using Docker.
# No local Go, Node, or pnpm required.
set -euo pipefail

REPO=$(git rev-parse --show-toplevel)
cd "$REPO"

echo "[1/4] backend go.sum..."
docker run --rm \
  -v "$REPO/packages/backend:/workspace" \
  -w /workspace \
  golang:1.22-alpine \
  sh -c "go mod download && go mod tidy"

echo "[2/4] PoC 02 go.sum..."
docker run --rm \
  -v "$REPO/poc/02-webrtc-stream/server:/workspace" \
  -w /workspace \
  golang:1.22-alpine \
  sh -c '[ -f go.mod ] || (go mod init poc-02-signal && go get github.com/gorilla/websocket@v1.5.1)
         go mod tidy'

echo "[3/4] PoC 04 go.sum..."
docker run --rm \
  -v "$REPO/poc/04-payment-gate/server:/workspace" \
  -w /workspace \
  golang:1.22-alpine \
  sh -c '[ -f go.mod ] || (go mod init poc-04-payment \
           && go get github.com/golang-jwt/jwt/v5@v5.2.1 \
           && go get github.com/stripe/stripe-go/v76@v76.25.0)
         go mod tidy'

echo "[4/4] pnpm-lock.yaml..."
docker run --rm \
  -v "$REPO:/workspace" \
  -w /workspace \
  node:20-alpine \
  sh -c "corepack enable && corepack prepare pnpm@9 --activate && pnpm install"

echo "Done. Commit: packages/backend/go.sum, poc/*/server/go.{mod,sum}, pnpm-lock.yaml"
