# Stripe Connect Notes

## Question
How does Stripe Connect Express work for IndraNet's marketplace model?

## Account Types
- **Standard**: Host creates their own Stripe account, grants IndraNet access. More host control, more setup friction.
- **Express**: IndraNet creates the Stripe account on host's behalf. Stripe handles KYC UI. Recommended for marketplaces.
- **Custom**: IndraNet fully controls the UI. Maximum customization, maximum compliance burden.

**Decision:** Use Express accounts for Phase 1.

## Onboarding Flow
1. Host clicks "Connect Stripe" in dashboard
2. Backend calls `stripe.accounts.create({ type: 'express' })`
3. Backend calls `stripe.accountLinks.create(...)` to get onboarding URL
4. Host completes Stripe's KYC form (ID, bank account, tax info)
5. Stripe webhook: `account.updated` with `details_submitted: true`
6. Backend marks host as payouts_enabled

## Transfer Flow (after session ends)
```go
// Transfer 80% of session revenue to host
stripe.transfers.create({
  amount: sessionRevenueCents * 0.80,
  currency: "usd",
  destination: host.stripeAccountId,
  transfer_group: sessionId,
})
```

## Webhooks Required
- `checkout.session.completed` → session AUTHORIZED
- `payment_intent.payment_failed` → session FAILED, notify user
- `account.updated` → update host payout status
- `transfer.created` → log payout to host

## Fee Structure (approximate, 2025)
- Stripe processing: 2.9% + $0.30 per transaction
- Stripe Connect: +0.25% per transfer to connected account
- Platform commission: 20% of session revenue (IndraNet's cut)
- Host receives: ~77% of session revenue after all fees

## TODO
- Test full onboarding flow in Stripe test mode
- Verify webhook signature verification in Go (`stripe.ConstructEvent`)
- Test refund flow for abandoned sessions
