# Backend — Go REST API + WebSocket Signaling

## Overview
The IndraNet backend is a stateless Go service providing:
- REST API for host registration, session management, user accounts, and payments
- WebSocket signaling hub for WebRTC offer/answer/ICE exchange
- Billing engine (per-minute tick loop) for active sessions
- Stripe webhook processing for payment events

## Structure
```
cmd/server/main.go              Entry point
internal/api/                   HTTP handlers
internal/signaling/             WebSocket signaling hub
internal/models/                Data models
internal/db/                    Database layer (pgx + redis)
internal/billing/               Session billing engine
internal/config/                Configuration from env vars
tests/                          Integration tests
```

## Running

```bash
# Start local services first
docker-compose up -d

# Run with hot reload (requires air)
make dev

# Or run directly
go run ./cmd/server
```

## Testing
```bash
make test
# or
go test -v -race ./...
```

## Environment Variables
See `../../.env.example` for all required env vars.

## API Documentation
See `../../docs/specs/api-spec.md`.
