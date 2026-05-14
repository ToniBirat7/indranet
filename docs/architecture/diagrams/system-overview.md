# System Overview Diagram

See `docs/architecture/overview.md` for the full annotated ASCII diagram.

## Component Boundaries

```
[User Browser/App] ←──── WebRTC P2P ────→ [Host Agent]
        │                                       │
        │ HTTPS                                 │ WebSocket
        ▼                                       ▼
[Go Backend API] ←──────────────────────→ [Go Backend API]
        │                                       
        ├── PostgreSQL (persistent state)       
        ├── Redis (session state cache)         
        └── NATS (event bus)                    
```

## Trust Boundaries

```
UNTRUSTED ZONE          │  PLATFORM ZONE       │  HOST ZONE
─────────────────────────┼──────────────────────┼──────────────────
User browser            │  Go backend          │  Host machine OS
User desktop client     │  PostgreSQL          │  C++ agent daemon
                        │  Redis               │  Windows Sandbox
                        │  NATS                │    └─ Guest app
                        │  Stripe webhooks     │
```

Every crossing of a trust boundary requires authentication:
- User → Backend: JWT bearer token
- Backend → Host Agent: Session token (verified on WebSocket connect)
- Stripe → Backend: Webhook signing secret (HMAC-SHA256)
- Host Agent → Sandbox: Named pipe (localhost-only, no auth needed)
