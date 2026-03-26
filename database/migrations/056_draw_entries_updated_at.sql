-- Migration 056: Add updated_at column to draw_entries
-- The trigger_update_draw_entries() function sets NEW.updated_at = NOW()
-- but the draw_entries table was created without this column, causing
-- every INSERT/UPDATE to fail with "record new has no field updated_at"

ALTER TABLE public.draw_entries
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Backfill existing rows
UPDATE public.draw_entries SET updated_at = created_at WHERE updated_at IS NULL;
