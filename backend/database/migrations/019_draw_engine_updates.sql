-- Draw Engine Schema Updates
-- Remove entry_cost column (business model change - no paid entry)
ALTER TABLE draws DROP COLUMN IF EXISTS entry_cost;

-- Add runner_ups_count column
ALTER TABLE draws ADD COLUMN IF NOT EXISTS runner_ups_count INT DEFAULT 1;

-- Add runner-up fields to draw_winners
ALTER TABLE draw_winners ADD COLUMN IF NOT EXISTS is_runner_up BOOLEAN DEFAULT FALSE;
ALTER TABLE draw_winners ADD COLUMN IF NOT EXISTS is_forfeited BOOLEAN DEFAULT FALSE;
ALTER TABLE draw_winners ADD COLUMN IF NOT EXISTS promoted_from UUID NULL;

-- Create index for runner-up queries
CREATE INDEX IF NOT EXISTS idx_draw_winners_runner_up ON draw_winners(draw_id, is_runner_up);
CREATE INDEX IF NOT EXISTS idx_draw_winners_forfeited ON draw_winners(draw_id, is_forfeited);
