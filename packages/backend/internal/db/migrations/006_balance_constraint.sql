-- 006_balance_constraint.sql — Non-negative balance enforcement at DB level

-- NOT VALID skips scanning existing rows (safe for zero-downtime deploys).
-- New inserts and updates are checked immediately; existing rows assumed valid.
-- Run VALIDATE CONSTRAINT in a separate step if you need full historical enforcement.
ALTER TABLE users ADD CONSTRAINT IF NOT EXISTS check_balance_non_negative
    CHECK (balance_cents >= 0) NOT VALID;
