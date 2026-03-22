-- Migration 045: Widen legacy amount columns to prevent numeric overflow
-- ─────────────────────────────────────────────────────────────────────────────
-- PROBLEM: daily_subscriptions.amount is numeric(5,2) from the original schema.
-- numeric(5,2) holds at most 999.99. When subscription_service.go inserts
-- dailyAmountKobo (2000 for ₦20, 20000 for ₦200 etc.) PostgreSQL raises:
--   ERROR 22003: numeric field overflow
-- causing a 500 Internal Server Error on every subscription attempt.
--
-- FIX: Widen the column to numeric(12,2) which holds up to 9,999,999,999.99
-- This covers both naira values (20.00 – 9,999.99) and kobo values (2000 – 9,999,999)
-- with plenty of headroom for future premium tiers.
-- ─────────────────────────────────────────────────────────────────────────────

ALTER TABLE public.daily_subscriptions
    ALTER COLUMN amount TYPE numeric(12,2);

-- Also remove the positive_amount CHECK so 0.00 amounts (free tiers) are allowed
-- in future without another migration.
ALTER TABLE public.daily_subscriptions
    DROP CONSTRAINT IF EXISTS positive_amount;

ALTER TABLE public.daily_subscriptions
    ADD CONSTRAINT positive_amount CHECK (amount >= 0);
