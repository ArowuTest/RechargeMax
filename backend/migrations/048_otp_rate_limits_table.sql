-- Migration 048: PostgreSQL-backed OTP rate limiting table
-- Replaces the in-memory rate limiter so limits survive restarts and work
-- correctly across multiple server instances / horizontal scaling.

CREATE TABLE IF NOT EXISTS otp_rate_limits (
    id          BIGSERIAL PRIMARY KEY,
    key         VARCHAR(64)  NOT NULL,   -- 'msisdn:<msisdn>' or 'ip:<ip>'
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for fast per-key range lookups (sliding window)
CREATE INDEX IF NOT EXISTS idx_otp_rate_limits_key_time
    ON otp_rate_limits (key, requested_at);

-- Auto-cleanup: entries are pruned by the application-level cleanup job.
-- A simple index on requested_at allows efficient range deletes.
CREATE INDEX IF NOT EXISTS idx_otp_rate_limits_time
    ON otp_rate_limits (requested_at);
