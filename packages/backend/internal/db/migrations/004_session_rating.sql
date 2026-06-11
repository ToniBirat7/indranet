-- 004_session_rating.sql — Per-session user rating (1–5 stars)

ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS rating INTEGER CHECK (rating BETWEEN 1 AND 5);

CREATE INDEX IF NOT EXISTS idx_sessions_host_rating ON sessions(host_id) WHERE rating IS NOT NULL;
