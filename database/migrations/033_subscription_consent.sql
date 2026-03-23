-- 033_subscription_consent.sql
-- Adds explicit recurring-charge consent fields to daily_subscriptions.
-- These are captured at the moment the user clicks "Pay & Subscribe" and
-- stored server-side so they form a legally defensible audit trail.
--
-- Fields recorded per consent event:
--   consent_given_at  – UTC timestamp the user ticked the checkbox
--   consent_ip        – client IP address (X-Forwarded-For or RemoteAddr)
--   consent_user_agent– browser user-agent string
--   consent_amount    – exact daily amount (NGN) the user was shown and agreed to
--   consent_entries   – exact entry count the user was shown and agreed to
--   consent_text      – the full consent sentence that was displayed (snapshot)

ALTER TABLE daily_subscriptions
    ADD COLUMN IF NOT EXISTS consent_given_at   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS consent_ip          TEXT,
    ADD COLUMN IF NOT EXISTS consent_user_agent  TEXT,
    ADD COLUMN IF NOT EXISTS consent_amount_ngn  NUMERIC(12,2),
    ADD COLUMN IF NOT EXISTS consent_entries     INTEGER,
    ADD COLUMN IF NOT EXISTS consent_text        TEXT;

-- Index for compliance queries ("show all subscriptions where consent was given")
CREATE INDEX IF NOT EXISTS idx_daily_subscriptions_consent_given_at
    ON daily_subscriptions (consent_given_at)
    WHERE consent_given_at IS NOT NULL;

COMMENT ON COLUMN daily_subscriptions.consent_given_at  IS 'UTC timestamp when user explicitly accepted recurring charge authorisation';
COMMENT ON COLUMN daily_subscriptions.consent_ip        IS 'Client IP address at time of consent (for chargeback defence)';
COMMENT ON COLUMN daily_subscriptions.consent_user_agent IS 'Browser user-agent at time of consent';
COMMENT ON COLUMN daily_subscriptions.consent_amount_ngn IS 'Daily NGN amount the user consented to (snapshot — immutable after creation)';
COMMENT ON COLUMN daily_subscriptions.consent_entries   IS 'Entry count the user consented to (snapshot — immutable after creation)';
COMMENT ON COLUMN daily_subscriptions.consent_text      IS 'Full authorisation text shown to and accepted by the user';
