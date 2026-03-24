-- Migration 051: Deduplicate prize_categories and add unique constraint
-- 
-- Problem: prize_categories has no unique constraint on (prize_template_id, category_name).
-- The seed in 020_prize_tier_system.sql uses ON CONFLICT DO NOTHING, but without a
-- unique constraint it never conflicts — so every server restart re-inserts duplicates.
--
-- Fix:
--   Step 1: Delete duplicates — keep only the row with the earliest created_at per (template, name).
--   Step 2: Add unique constraint so ON CONFLICT DO NOTHING works correctly on future restarts.

-- Step 1: Delete duplicates, keeping earliest per (prize_template_id, category_name)
DELETE FROM prize_categories
WHERE id NOT IN (
  SELECT DISTINCT ON (template_id, category_name) id
  FROM prize_categories
  ORDER BY template_id, category_name, created_at ASC
);

-- Step 2: Add unique constraint to prevent duplicates on future re-seeds
-- DB column is 'template_id' (mapped to prize_template_id in Go entity)
ALTER TABLE prize_categories
  ADD CONSTRAINT uq_prize_categories_template_name
  UNIQUE (template_id, category_name);
