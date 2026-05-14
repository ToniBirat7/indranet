-- 002_hosts.sql — Host machine listings

CREATE TABLE IF NOT EXISTS hosts (
    id                  TEXT PRIMARY KEY DEFAULT 'hst_' || encode(gen_random_bytes(12), 'hex'),
    user_id             TEXT NOT NULL REFERENCES users(id),
    display_name        TEXT NOT NULL,
    gpu_model           TEXT NOT NULL,
    vram_gb             INTEGER NOT NULL,
    cpu_model           TEXT NOT NULL,
    ram_gb              INTEGER NOT NULL,
    os                  TEXT NOT NULL,
    price_per_hour_cents BIGINT NOT NULL,  -- Price in USD cents (e.g., 250 = $2.50/hr)
    tags                TEXT[] NOT NULL DEFAULT '{}',
    online              BOOLEAN NOT NULL DEFAULT FALSE,
    agent_token_hash    TEXT,              -- Hashed JWT for agent authentication
    stripe_account_id   TEXT,             -- Stripe Connect Express account ID
    payouts_enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    total_sessions      INTEGER NOT NULL DEFAULT 0,
    rating_sum          INTEGER NOT NULL DEFAULT 0,
    rating_count        INTEGER NOT NULL DEFAULT 0,
    machine_fingerprint TEXT,             -- Hardware fingerprint hash
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hosts_user_id ON hosts(user_id);
CREATE INDEX IF NOT EXISTS idx_hosts_online ON hosts(online) WHERE online = TRUE;
CREATE INDEX IF NOT EXISTS idx_hosts_price ON hosts(price_per_hour_cents);
