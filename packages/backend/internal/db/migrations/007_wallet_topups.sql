-- 007_wallet_topups.sql — Idempotency log for Stripe wallet top-up webhooks
-- Stripe guarantees at-least-once delivery; duplicate checkout.session.completed
-- events for wallet_topup must not credit the wallet twice.
-- PRIMARY KEY on stripe_checkout_id makes the INSERT fail on duplicates.

CREATE TABLE IF NOT EXISTS wallet_topups (
    stripe_checkout_id TEXT PRIMARY KEY,
    user_id            TEXT NOT NULL REFERENCES users(id),
    amount_cents       BIGINT NOT NULL,
    credited_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
