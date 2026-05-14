# Billing Model

## Core Model
Per-minute billing with pre-funded balance.

1. User tops up balance ($10, $25, $50, $100)
2. User starts session → balance locked for minimum 15 minutes
3. Backend billing engine deducts per-minute rate every 60 seconds
4. Session ends when user disconnects or balance = 0
5. Unused balance stays in wallet for future sessions
6. Refund policy: unused balance refundable within 30 days

## Rate Calculation
```
session_cost_per_minute = host_price_per_hour / 60
billing_tick_amount = session_cost_per_minute * 1
```

## Minimum Session
To avoid micro-transactions, require a minimum 15-minute pre-authorization on session start.
The balance is reserved (not yet charged) until the session ends.

## Balance Warning
At `balance_remaining < 5 * session_cost_per_minute`:
- Send `session_warning` WebSocket event to user client
- Client shows "5 minutes remaining" banner with top-up link

## Balance Exhaustion
At `balance_remaining <= 0`:
- Backend billing engine emits kill event via NATS
- Backend marks session as ENDING
- Next heartbeat response to host agent includes `action: kill`
- Host agent begins session teardown

## Platform Economics
```
user_pays:       $2.50/hr (example)
platform_takes:  $0.50/hr (20%)
host_receives:   $2.00/hr (80%)
stripe_fees:     ~$0.10/hr (varies)
host_net:        ~$1.90/hr
```
