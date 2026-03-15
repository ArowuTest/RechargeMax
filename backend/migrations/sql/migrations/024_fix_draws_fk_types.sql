-- Migration 042: Ensure FK constraints on draws table are correctly typed (UUID).
-- draw_type_id and prize_template_id are already UUID in this schema version.
-- This migration is idempotent and safe to run.

-- Drop existing FK constraints if any (ignore errors if they don't exist)
ALTER TABLE draws DROP CONSTRAINT IF EXISTS draws_draw_type_id_fkey;
ALTER TABLE draws DROP CONSTRAINT IF EXISTS draws_prize_template_id_fkey;

-- Add prize_template_id column if it doesn't exist
ALTER TABLE draws ADD COLUMN IF NOT EXISTS prize_template_id UUID;

-- Re-add FK constraints (UUID → UUID, no casting needed)
ALTER TABLE draws
    ADD CONSTRAINT draws_draw_type_id_fkey
    FOREIGN KEY (draw_type_id) REFERENCES draw_types(id) ON DELETE SET NULL
    NOT VALID;  -- NOT VALID so it doesn't scan existing rows

ALTER TABLE draws
    ADD CONSTRAINT draws_prize_template_id_fkey
    FOREIGN KEY (prize_template_id) REFERENCES prize_templates(id) ON DELETE SET NULL
    NOT VALID;
