-- Migration 047: Add missing updated_at column to daily_subscriptions
-- ─────────────────────────────────────────────────────────────────────────────
-- ROOT CAUSE OF 500 on POST /api/v1/subscription/create:
--
-- The DailySubscription entity has:
--   UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
--
-- GORM automatically sets updated_at on every INSERT and UPDATE.
-- But daily_subscriptions was created by 14_daily_subscriptions.sql which
-- only has: id, user_id, msisdn, subscription_date, amount, draw_entries_earned,
-- points_earned, payment_reference, status, is_paid, customer_email,
-- customer_name, created_at, subscription_code
--
-- No updated_at column. The DO $$ block in migration 041 added 16 new columns
-- but missed updated_at. Every INSERT attempt fails with:
--   ERROR: column "updated_at" of relation "daily_subscriptions" does not exist
-- This is a non-AppError → RespondWithError returns 500.
-- ─────────────────────────────────────────────────────────────────────────────

ALTER TABLE public.daily_subscriptions
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Backfill existing rows so the column is not null on reads
UPDATE public.daily_subscriptions
    SET updated_at = created_at
    WHERE updated_at IS NULL;

-- Create update trigger so updated_at stays current (same pattern as other tables)
CREATE OR REPLACE FUNCTION update_daily_subscriptions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_daily_subscriptions_updated_at ON public.daily_subscriptions;

CREATE TRIGGER trg_daily_subscriptions_updated_at
    BEFORE UPDATE ON public.daily_subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_daily_subscriptions_updated_at();
