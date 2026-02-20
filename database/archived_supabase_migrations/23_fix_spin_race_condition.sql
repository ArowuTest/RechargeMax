-- ============================================================================
-- MIGRATION: Fix Spin Race Condition
-- Date: 2026-02-01
-- Purpose: Prevent users from spinning multiple times by adding unique constraint
-- ============================================================================

BEGIN;

-- Add unique constraint to prevent duplicate spins for same user on same day
-- This prevents race conditions where multiple concurrent requests could create duplicate spins
ALTER TABLE wheel_spins 
ADD CONSTRAINT unique_user_spin_per_day 
UNIQUE (user_id, DATE(created_at));

-- Add index for performance on spin eligibility queries
CREATE INDEX IF NOT EXISTS idx_wheel_spins_user_date 
ON wheel_spins(user_id, DATE(created_at));

-- Add index for user_id lookups
CREATE INDEX IF NOT EXISTS idx_wheel_spins_user_id 
ON wheel_spins(user_id);

-- Add comment explaining the constraint
COMMENT ON CONSTRAINT unique_user_spin_per_day ON wheel_spins IS 
'Prevents race conditions by ensuring only one spin per user per day at database level';

COMMIT;

-- Rollback script (if needed):
-- BEGIN;
-- DROP INDEX IF EXISTS idx_wheel_spins_user_id;
-- DROP INDEX IF EXISTS idx_wheel_spins_user_date;
-- ALTER TABLE wheel_spins DROP CONSTRAINT IF EXISTS unique_user_spin_per_day;
-- COMMIT;
