-- Migration 043: Drop user_id+subscription_date unique constraint
-- ─────────────────────────────────────────────────────────────────────────────
-- PROBLEM: daily_subscriptions has UNIQUE(user_id, subscription_date) from the
-- original base schema (14_daily_subscriptions.sql). This constraint means a
-- logged-in user can only create ONE subscription per calendar day.
-- Migration 041 already tries to drop it, but it runs statement-by-statement and
-- the embed runner's error suppression ("already exists" / "does not exist") means
-- a failed DROP in 041 is silently swallowed. This migration guarantees removal.
--
-- subscription_date is a DATE column, so time.Time values from Go are truncated
-- to just the calendar date. A retry on the same day by the same user gets:
--   ERROR 23505: duplicate key value violates unique constraint
--   "daily_subscriptions_user_id_subscription_date_key"
-- The service catches ANY 23505 and returns 409 "Duplicate subscription code"
-- which confuses the user — they see "please retry" but every retry also fails.
-- ─────────────────────────────────────────────────────────────────────────────

-- 1. Drop the problematic unique constraint (idempotent — IF EXISTS is safe)
ALTER TABLE public.daily_subscriptions
    DROP CONSTRAINT IF EXISTS daily_subscriptions_user_id_subscription_date_key;

-- 2. Also drop any uniqueness index that may have been created separately
DROP INDEX IF EXISTS public.daily_subscriptions_user_id_subscription_date_key;
DROP INDEX IF EXISTS public.idx_daily_subscriptions_user_subscription_date;

-- 3. Ensure subscription_code column is wide enough for UUID-based codes
--    (SUB + 4 digits + 8 UUID hex = 15 chars; was VARCHAR(20) which is fine,
--    but entity has size:50 — align the column to avoid future truncation)
ALTER TABLE public.daily_subscriptions
    ALTER COLUMN subscription_code TYPE VARCHAR(60);

-- 4. Keep a non-unique index on (user_id, subscription_date) for query performance
--    (we still query by these fields, just no longer enforce uniqueness)
CREATE INDEX IF NOT EXISTS idx_daily_subscriptions_user_date
    ON public.daily_subscriptions (user_id, subscription_date);
