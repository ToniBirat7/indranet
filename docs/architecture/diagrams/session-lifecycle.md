# Session Lifecycle Diagram

## State Machine

```
                    ┌─────────┐
                    │ CREATED │
                    └────┬────┘
                         │ Stripe payment completed
                         │ (webhook: checkout.session.completed)
                         ▼
                  ┌────────────┐
                  │ AUTHORIZED │
                  └─────┬──────┘
                        │ Host agent acknowledged
                        │ (WebSocket: session_ready)
                        ▼
                   ┌─────────┐
                   │  ACTIVE │ ←──── Billing ticks every 60s
                   └────┬────┘
          ┌─────────────┼───────────────┐
          │             │               │
          │ User        │ Balance       │ Host crash /
          │ disconnects │ exhausted     │ agent timeout
          ▼             ▼               ▼
       ┌───────┐    ┌─────────┐    ┌────────┐
       │ENDING │    │ ENDING  │    │ FAILED │
       └───┬───┘    └────┬────┘    └────────┘
           └──────┬───────┘
                  │ Sandbox destroyed confirmed
                  │ Final billing calculated
                  ▼
             ┌────────┐
             │ ENDED  │
             └────────┘
                  │
                  ▼
         Host payout queued (Stripe Connect)
```

## State Transition Rules

| From | To | Trigger | Actor |
|------|----|---------|-------|
| CREATED | AUTHORIZED | `checkout.session.completed` Stripe webhook | Stripe → Backend |
| CREATED | FAILED | Payment timeout (30 min) | Backend cron |
| AUTHORIZED | ACTIVE | `session_ready` WebSocket event | Host Agent → Backend |
| AUTHORIZED | FAILED | Host agent timeout (5 min) | Backend timer |
| ACTIVE | ENDING | User sends `session_end` | User Client → Backend |
| ACTIVE | ENDING | Billing engine: balance = 0 | Backend Billing Engine |
| ACTIVE | FAILED | Host agent heartbeat timeout (90s) | Backend watchdog |
| ENDING | ENDED | Host agent confirms `sandbox_destroyed` | Host Agent → Backend |
| ENDING | FAILED | Sandbox teardown timeout (60s) | Backend timer |

## Session Events (NATS topics)

```
session.created        → billing engine subscribes to start tracking
session.authorized     → host agent subscribes to get notified
session.active         → user client subscribes to unlock viewer
session.warning        → user client subscribes to show low balance alert
session.ending         → host agent subscribes to begin teardown
session.ended          → billing engine subscribes to finalize charges
session.failed         → admin alerting subscribes
```
