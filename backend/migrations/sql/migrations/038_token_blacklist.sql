-- Migration 038: Create token_blacklist table + seed prize_fulfillment_config
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

-- ── prize_fulfillment_config seed ─────────────────────────────────────────────
-- prize_fulfillment_config.id is a serial backed by prize_fulfillment_config_id_seq.
-- The sequence may not exist yet on a clean deployment (base schema file 31 uses
-- ALTER COLUMN … SET DEFAULT nextval(…) which references the sequence but does not
-- CREATE it). We create it here idempotently before the INSERT so the default fires.

CREATE SEQUENCE IF NOT EXISTS public.prize_fulfillment_config_id_seq
    START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;

ALTER TABLE public.prize_fulfillment_config
    ALTER COLUMN id SET DEFAULT nextval('public.prize_fulfillment_config_id_seq'::regclass);

-- Seed default prize fulfillment configuration (safe defaults for staging).
-- MANUAL mode ensures prizes don't get lost when VTPass is not yet configured.
-- ON CONFLICT DO UPDATE keeps any admin-changed mode intact except for the
-- safety-critical defaults (retry settings, fallback flags).
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
