# Payment Flow Diagram

## User Payment Flow

```
User selects host
       │
       ▼
POST /sessions  ──→  Backend creates session (CREATED)
       │              Returns: { session_id, checkout_url }
       │
       ▼
User redirected to Stripe Checkout
       │
       ▼
User completes payment
       │
       ▼
Stripe sends webhook: checkout.session.completed
       │
       ▼
Backend verifies webhook signature (HMAC-SHA256)
       │
       ▼
Backend marks session AUTHORIZED
       │
       ├──→ Notify host agent via WebSocket
       │
       ▼
Host agent creates sandbox, initializes stream
       │
       ▼
Session becomes ACTIVE
       │
       ▼
       ═══════ Per-minute billing loop ═══════
       │
       ▼
Billing engine ticks every 60s:
  balance -= (host_rate_per_minute * 1)
  if balance < 5min_threshold → emit WARNING
  if balance <= 0 → emit KILL
       │
       ▼
Session ends (user, balance, or crash)
       │
       ▼
Final billing calculated:
  platform_revenue = total_charged * 0.20
  host_revenue     = total_charged * 0.80
       │
       ▼
Stripe Connect transfer to host Express account
(Payout on next scheduled payout date)
```

## Stripe Objects Used

| Object | Purpose |
|--------|---------|
| `Customer` | Represents a user; stores payment methods |
| `PaymentIntent` | One per session; pre-authorizes the session cost |
| `CheckoutSession` | Hosted payment page for session initiation |
| `Account` (Express) | One per host; receives payouts |
| `Transfer` | Platform → host account after session ends |
| `Webhook` | Payment events → backend session state machine |

## Escrow Model

IndraNet does not hold user funds indefinitely. The model:
1. User adds funds to balance via Stripe Checkout (top-up)
2. Balance tracked in IndraNet's database (not a Stripe wallet)
3. As sessions are consumed, backend records charges
4. Host payouts batched daily via Stripe Connect transfers
5. Refund policy: unused balance refundable within 30 days
