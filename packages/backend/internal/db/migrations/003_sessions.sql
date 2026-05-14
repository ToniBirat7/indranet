-- 003_sessions.sql — Session lifecycle and billing records

CREATE TYPE session_state AS ENUM (
    'CREATED',
    'AUTHORIZED',
    'ACTIVE',
    'ENDING',
    'ENDED',
    'FAILED'
);

CREATE TABLE IF NOT EXISTS sessions (
    id                  TEXT PRIMARY KEY DEFAULT 'ses_' || encode(gen_random_bytes(12), 'hex'),
    user_id             TEXT NOT NULL REFERENCES users(id),
    host_id             TEXT NOT NULL REFERENCES hosts(id),
    state               session_state NOT NULL DEFAULT 'CREATED',
    rate_per_minute_cents BIGINT NOT NULL, -- Locked-in rate at session start
    pre_auth_minutes    INTEGER NOT NULL DEFAULT 15,
    total_charged_cents BIGINT NOT NULL DEFAULT 0,
    stripe_checkout_id  TEXT,
    stripe_payment_intent_id TEXT,
    started_at          TIMESTAMPTZ,
    ended_at            TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Billing tick records (one row per minute of active session)
CREATE TABLE IF NOT EXISTS billing_ticks (
    id          BIGSERIAL PRIMARY KEY,
    session_id  TEXT NOT NULL REFERENCES sessions(id),
    amount_cents BIGINT NOT NULL,
    ticked_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_host_id ON sessions(host_id);
CREATE INDEX IF NOT EXISTS idx_sessions_state ON sessions(state);
CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(state) WHERE state = 'ACTIVE';
CREATE INDEX IF NOT EXISTS idx_billing_ticks_session ON billing_ticks(session_id);
