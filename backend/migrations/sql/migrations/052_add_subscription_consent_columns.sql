-- Migration 052: Add consent audit columns to daily_subscriptions
-- ─────────────────────────────────────────────────────────────────────────────
-- ROOT CAUSE OF 500 on POST /api/v1/subscriptions/daily:
--
--   ERROR: column "consent_given_at" of relation "daily_subscriptions"
--          does not exist (SQLSTATE 42703)
--
-- The DailySubscription entity has 6 consent fields added in commit ec4315e:
--   consent_given_at, consent_ip, consent_user_agent,
--   consent_amount_ngn, consent_entries, consent_text
--
-- The migration for these columns was mistakenly placed in database/migrations/
-- which is NOT embedded by the migration runner (which only reads
-- migrations/sql/*.sql and migrations/sql/migrations/*.sql).
-- So every INSERT into daily_subscriptions fails with column-does-not-exist.
-- ─────────────────────────────────────────────────────────────────────────────

ALTER TABLE public.daily_subscriptions
    ADD COLUMN IF NOT EXISTS consent_given_at   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS consent_ip          TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS consent_user_agent  TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS consent_amount_ngn  NUMERIC(12,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consent_entries     INTEGER     NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consent_text        TEXT        NOT NULL DEFAULT '';

-- Index for compliance queries
CREATE INDEX IF NOT EXISTS idx_daily_subscriptions_consent_given_at
    ON public.daily_subscriptions (consent_given_at)
    WHERE consent_given_at IS NOT NULL;
