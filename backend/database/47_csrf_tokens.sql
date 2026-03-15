-- Migration 047: CSRF tokens table (INFRA-001)
-- Replaces the in-memory CSRF store so tokens survive restarts
-- and remain consistent across multiple API instances.
CREATE TABLE IF NOT EXISTS csrf_tokens (
    token       TEXT        PRIMARY KEY,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index to speed up TTL-based cleanup
CREATE INDEX IF NOT EXISTS idx_csrf_tokens_expires_at ON csrf_tokens (expires_at);

-- Helper function for manual or pg_cron cleanup
CREATE OR REPLACE FUNCTION cleanup_expired_csrf_tokens() RETURNS void AS $$
BEGIN
    DELETE FROM csrf_tokens WHERE expires_at <= NOW();
END;
$$ LANGUAGE plpgsql;
