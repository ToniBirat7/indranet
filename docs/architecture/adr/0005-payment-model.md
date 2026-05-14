# ADR-0005: Stripe Connect for Marketplace Payments

**Status:** Accepted  
**Date:** 2025-05-14  
**Author:** @architect

## Context

IndraNet is a two-sided marketplace: users pay for compute time, hosts earn revenue. The platform takes a commission. This requires:
- Users can pay with credit card or stored balance
- Hosts receive payouts automatically
- Platform collects its commission without manual reconciliation
- Stripe's KYC/AML compliance handles host identity verification
- Disputes and refunds can be managed through a familiar interface

Three payment integration patterns exist:
1. **Stripe Connect** (marketplace platform)
2. **Direct Stripe** (platform collects, manually pays hosts)
3. **Crypto** (permissionless, no KYC)

## Decision

Use **Stripe Connect with Express accounts** for host payouts.

Under this model:
- Users pay IndraNet (platform account) for sessions
- IndraNet's billing engine tracks per-session earnings per host
- Stripe Connect routes 80% of session revenue to host's Express account
- Platform retains 20% as commission
- Stripe handles KYC for hosts during Express account onboarding
- Hosts can receive payouts on weekly or monthly schedules

Session billing model:
- Users pre-fund a balance (USD) via Stripe Checkout
- Billing engine deducts the per-minute rate every 60 seconds
- When balance reaches zero, session is terminated
- Minimum session pre-authorization: 10 minutes at listed rate

## Consequences

**Positive:**
- Stripe's compliance infrastructure handles international payouts, tax forms (1099), and AML checks
- Express account onboarding is well-documented and fast (5-10 minutes for hosts)
- Stripe's webhook system reliably notifies the backend of payment events
- Test mode allows full integration testing without real money

**Negative:**
- Stripe Connect fees are ~2.9% + $0.30 per transaction plus a platform fee
- Hosts in some countries are not supported by Stripe Express accounts
- Per-minute billing requires real-time balance tracking (Redis) to kill sessions promptly

## Alternatives Considered

**Crypto payments (USDC/ETH):** Permissionless, no KYC friction for hosts in restricted countries. But: no chargebacks (good and bad), unfamiliar for mainstream users, gas fees on Ethereum, regulatory uncertainty. Deferred to Phase 3 as an optional alternative.

**Direct Stripe (no Connect):** Platform collects all revenue, manually pays hosts via bank transfer. Doesn't scale, creates withholding tax complexity. Rejected.

**PayPal:** Worse developer experience, higher fraud rates, PayPal can freeze accounts arbitrarily. Rejected.
