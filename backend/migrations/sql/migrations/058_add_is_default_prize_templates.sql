-- Migration 058: Add is_default column to prize_templates if missing
ALTER TABLE prize_templates ADD COLUMN IF NOT EXISTS is_default BOOLEAN NOT NULL DEFAULT false;

-- Also ensure prize_categories column names are consistent
-- (category_name / winners_count / runner_ups_count already defined in entity)
