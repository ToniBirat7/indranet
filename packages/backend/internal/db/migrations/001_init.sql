-- 001_init.sql — Base tables: users

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id           TEXT PRIMARY KEY DEFAULT 'usr_' || encode(gen_random_bytes(12), 'hex'),
    email        TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    name         TEXT NOT NULL,
    role         TEXT NOT NULL DEFAULT 'user',  -- user | host | admin
    balance_cents BIGINT NOT NULL DEFAULT 0,    -- pre-funded wallet in USD cents
    stripe_customer_id TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
