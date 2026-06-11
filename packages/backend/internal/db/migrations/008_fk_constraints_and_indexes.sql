-- 008_fk_constraints_and_indexes.sql
-- Explicit ON DELETE semantics for all FKs + unique Stripe ID indexes +
-- composite indexes for frequent (host_id, state) and (user_id, state) filters.
--
-- ON DELETE policy:
--   hosts.user_id    → CASCADE  (removing user removes their listings)
--   sessions.user_id → RESTRICT (financial history — cannot delete user with sessions)
--   sessions.host_id → RESTRICT (cannot delete host with session records)
--   billing_ticks.session_id → RESTRICT (revenue records must outlive sessions)
--   wallet_topups.user_id    → RESTRICT (financial records, not safe to cascade)

DO $$
BEGIN
  -- hosts.user_id: implicit NO ACTION → CASCADE
  IF EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_name = 'hosts_user_id_fkey' AND table_name = 'hosts'
  ) THEN
    ALTER TABLE hosts DROP CONSTRAINT hosts_user_id_fkey;
  END IF;
  ALTER TABLE hosts ADD CONSTRAINT hosts_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

  -- sessions.user_id: → RESTRICT
  IF EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_name = 'sessions_user_id_fkey' AND table_name = 'sessions'
  ) THEN
    ALTER TABLE sessions DROP CONSTRAINT sessions_user_id_fkey;
  END IF;
  ALTER TABLE sessions ADD CONSTRAINT sessions_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

  -- sessions.host_id: → RESTRICT
  IF EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_name = 'sessions_host_id_fkey' AND table_name = 'sessions'
  ) THEN
    ALTER TABLE sessions DROP CONSTRAINT sessions_host_id_fkey;
  END IF;
  ALTER TABLE sessions ADD CONSTRAINT sessions_host_id_fkey
    FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE RESTRICT;

  -- billing_ticks.session_id: → RESTRICT
  IF EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_name = 'billing_ticks_session_id_fkey' AND table_name = 'billing_ticks'
  ) THEN
    ALTER TABLE billing_ticks DROP CONSTRAINT billing_ticks_session_id_fkey;
  END IF;
  ALTER TABLE billing_ticks ADD CONSTRAINT billing_ticks_session_id_fkey
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE RESTRICT;

  -- wallet_topups.user_id: → RESTRICT
  IF EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_name = 'wallet_topups_user_id_fkey' AND table_name = 'wallet_topups'
  ) THEN
    ALTER TABLE wallet_topups DROP CONSTRAINT wallet_topups_user_id_fkey;
  END IF;
  ALTER TABLE wallet_topups ADD CONSTRAINT wallet_topups_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
END $$;

-- Unique partial indexes: prevent two accounts sharing a Stripe ID
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_stripe_customer_id
    ON users(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_hosts_stripe_account_id
    ON hosts(stripe_account_id) WHERE stripe_account_id IS NOT NULL;

-- Composite indexes for session queries that filter on (host_id, state) and (user_id, state)
CREATE INDEX IF NOT EXISTS idx_sessions_host_state
    ON sessions(host_id, state);

CREATE INDEX IF NOT EXISTS idx_sessions_user_state
    ON sessions(user_id, state);
