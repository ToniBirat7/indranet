# IndraNet API Specification

Base URL: `https://api.indranet.io/v1` (production), `http://localhost:8080/v1` (dev)

All requests require `Content-Type: application/json` unless noted.
All authenticated endpoints require `Authorization: Bearer <jwt>`.

---

## Health

### GET /health
Returns service health status. No authentication required.

**Response 200:**
```json
{
  "status": "ok",
  "postgres": "ok",
  "redis": "ok",
  "version": "0.1.0"
}
```

---

## Authentication

### POST /auth/register
Register a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "minimum8chars",
  "name": "Alice"
}
```

**Response 201:**
```json
{
  "user_id": "usr_01H...",
  "token": "<jwt>"
}
```

### POST /auth/login
**Request:** `{ "email": "...", "password": "..." }`
**Response 200:** `{ "token": "<jwt>", "expires_at": "2025-06-01T00:00:00Z" }`

---

## Hosts

### POST /hosts/register
Register this machine as a host. Authenticated (host user role).

**Request:**
```json
{
  "display_name": "Alice's Gaming Rig",
  "gpu_model": "NVIDIA RTX 4080",
  "vram_gb": 16,
  "cpu_model": "Intel Core i9-14900K",
  "ram_gb": 64,
  "os": "Windows 11 Pro",
  "price_per_hour_usd": "2.50",
  "tags": ["gaming", "4k", "high-fps"]
}
```

**Response 201:**
```json
{
  "host_id": "hst_01H...",
  "agent_token": "<agent-jwt>"
}
```
The `agent_token` is passed to the C++ agent on startup.

### GET /hosts
Browse available hosts. No authentication required.

**Query params:** `gpu_min_vram_gb`, `price_max_per_hour`, `tags`, `page`, `limit`

**Response 200:**
```json
{
  "hosts": [
    {
      "host_id": "hst_01H...",
      "display_name": "Alice's Gaming Rig",
      "gpu_model": "NVIDIA RTX 4080",
      "vram_gb": 16,
      "price_per_hour_usd": "2.50",
      "online": true,
      "rating": 4.8
    }
  ],
  "total": 42,
  "page": 1
}
```

### GET /hosts/:id
Get host details.

---

## Sessions

### POST /sessions
Create a new session (initiates payment flow). Authenticated.

**Request:**
```json
{
  "host_id": "hst_01H...",
  "duration_minutes": 60
}
```

**Response 201:**
```json
{
  "session_id": "ses_01H...",
  "state": "CREATED",
  "checkout_url": "https://checkout.stripe.com/pay/...",
  "expires_at": "2025-05-14T15:30:00Z"
}
```

### PUT /sessions/:id/start
Called by host agent after sandbox is ready. Agent-authenticated (agent JWT).

**Response 200:** `{ "state": "ACTIVE" }`

### PUT /sessions/:id/heartbeat
Called by host agent every 30s to report liveness. Agent-authenticated.

**Response 200:** `{ "action": "continue" }` or `{ "action": "kill" }`

### DELETE /sessions/:id
End a session (user-initiated). Authenticated.

**Response 200:** `{ "state": "ENDING" }`

### GET /sessions/:id
Get session status. Authenticated.

**Response 200:**
```json
{
  "session_id": "ses_01H...",
  "state": "ACTIVE",
  "host_id": "hst_01H...",
  "started_at": "2025-05-14T14:00:00Z",
  "duration_seconds": 3600,
  "balance_remaining_minutes": 23,
  "signaling_url": "wss://api.indranet.io/v1/signal/ses_01H..."
}
```

---

## WebSocket Signaling

### WS /signal/:session_id
WebRTC signaling endpoint. Authenticated via `?token=<jwt>` query param.

**Messages (JSON):**

Client → Server:
```json
{ "type": "offer", "sdp": "v=0\r\n..." }
{ "type": "ice_candidate", "candidate": {...} }
```

Server → Client:
```json
{ "type": "answer", "sdp": "v=0\r\n..." }
{ "type": "ice_candidate", "candidate": {...} }
{ "type": "session_ready" }
{ "type": "session_warning", "minutes_remaining": 5 }
{ "type": "session_kill" }
```

---

## Webhooks

### POST /webhooks/stripe
Stripe event delivery. No JWT auth; verified via `Stripe-Signature` header (HMAC-SHA256 with `STRIPE_WEBHOOK_SECRET`).

Handled events:
- `checkout.session.completed` → session AUTHORIZED
- `payment_intent.payment_failed` → session FAILED
- `account.updated` → host Connect account status update
