-- 005_perf_indexes.sql — Partial and composite indexes for hot billing/agent/listing paths

-- Billing sweep every 2 min queries WHERE state = 'ENDING'
CREATE INDEX IF NOT EXISTS idx_sessions_ending
    ON sessions(state) WHERE state = 'ENDING';

-- Agent pending poll + awaitAgentReady + sweep query WHERE state = 'AUTHORIZED'
CREATE INDEX IF NOT EXISTS idx_sessions_authorized
    ON sessions(state) WHERE state = 'AUTHORIZED';

-- Sweep for abandoned CREATED sessions (WHERE state = 'CREATED')
CREATE INDEX IF NOT EXISTS idx_sessions_created_state
    ON sessions(state) WHERE state = 'CREATED';

-- ListSessions pagination: (user_id, created_at DESC) covers ORDER BY + WHERE user_id = $1
CREATE INDEX IF NOT EXISTS idx_sessions_user_created_desc
    ON sessions(user_id, created_at DESC);
