-- Migration 038: Create token_blacklist table
-- Supports instant JWT token revocation on admin logout (SEC-003).
CREATE TABLE IF NOT EXISTS public.token_blacklist (
    id          UUID        NOT NULL DEFAULT uuid_generate_v4(),
    token       TEXT        NOT NULL,
    admin_id    UUID        NOT NULL,
    reason      VARCHAR(255) NOT NULL DEFAULT 'logout',
    expires_at  TIMESTAMP   NOT NULL,
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT token_blacklist_pkey PRIMARY KEY (id),
    CONSTRAINT token_blacklist_token_unique UNIQUE (token)
);

CREATE INDEX IF NOT EXISTS idx_token_blacklist_token       ON public.token_blacklist (token);
CREATE INDEX IF NOT EXISTS idx_token_blacklist_expires_at  ON public.token_blacklist (expires_at);
CREATE INDEX IF NOT EXISTS idx_token_blacklist_admin_id    ON public.token_blacklist (admin_id);

-- Clean up expired tokens automatically (requires pg_cron; safe to ignore if not available)
-- Expired rows are harmless — IsBlacklisted always filters by expires_at > NOW().

-- Seed default prize fulfillment configuration (safe defaults for staging)
-- MANUAL mode ensures prizes don't get lost when VTPass is not yet configured.
INSERT INTO prize_fulfillment_config (
    prize_type, fulfillment_mode, auto_retry_enabled, max_retry_attempts,
    retry_delay_seconds, fallback_to_manual, fallback_notification_enabled,
    provision_timeout_seconds, is_active, created_by
) VALUES
    ('AIRTIME', 'MANUAL', false, 0, 0, true, true, 60, true, 'SYSTEM_SEED'),
    ('DATA',    'MANUAL', false, 0, 0, true, true, 60, true, 'SYSTEM_SEED'),
    ('CASH',    'MANUAL', false, 0, 0, true, true, 60, true, 'SYSTEM_SEED'),
    ('POINTS',  'MANUAL', false, 0, 0, true, true, 60, true, 'SYSTEM_SEED')
ON CONFLICT (prize_type) DO NOTHING;
