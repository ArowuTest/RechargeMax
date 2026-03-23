-- Migration 050: Add missing columns required by mutation handlers
-- 
-- 1. points_adjustments.admin_id
--    Go entity uses `admin_id` but the original schema only has `created_by`.
--    Add admin_id as an alias column (populated from created_by context).
--
-- 2. draws.total_winners
--    Go entity maps TotalWinners → total_winners but the original schema
--    only has winners_count. Add total_winners column.

-- ── points_adjustments: add admin_id ─────────────────────────────────────────
ALTER TABLE points_adjustments
    ADD COLUMN IF NOT EXISTS admin_id UUID;

-- Backfill from created_by (they represent the same thing)
UPDATE points_adjustments
    SET admin_id = created_by
    WHERE admin_id IS NULL AND created_by IS NOT NULL;

-- ── draws: add total_winners ──────────────────────────────────────────────────
ALTER TABLE draws
    ADD COLUMN IF NOT EXISTS total_winners INTEGER NOT NULL DEFAULT 0;

-- Backfill from winners_count
UPDATE draws
    SET total_winners = winners_count
    WHERE total_winners = 0 AND winners_count > 0;

-- ── Fix foreign key: created_by → admin_users (not users) ────────────────────
-- The original schema had: created_by FK → users.id
-- But created_by is populated from admin_id (which is an admin_users.id).
-- Drop the wrong FK and either remove the constraint or point it to admin_users.

ALTER TABLE points_adjustments
    DROP CONSTRAINT IF EXISTS fk_points_adjustments_admin;

-- Make created_by nullable so it doesn't block inserts when admin ID
-- doesn't exist in users table (admin_users is a separate table).
ALTER TABLE points_adjustments
    ALTER COLUMN created_by DROP NOT NULL;
