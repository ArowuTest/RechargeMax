-- Migration 051: Deduplicate prize_categories and add unique constraint
-- 
-- Problem: prize_categories has no unique constraint on (prize_template_id, category_name),
-- causing duplicate rows to accumulate on every re-seed / deployment.
-- Solution: 
--   1. Delete all duplicates, keeping only the earliest record per (template, name).
--   2. Add a unique constraint to prevent future duplicates.

-- Step 1: Delete duplicates — keep only the row with the earliest created_at per (template, name)
DELETE FROM prize_categories
WHERE id NOT IN (
  SELECT DISTINCT ON (prize_template_id, category_name) id
  FROM prize_categories
  ORDER BY prize_template_id, category_name, created_at ASC
);

-- Step 2: Add unique constraint to prevent this happening again
ALTER TABLE prize_categories
  ADD CONSTRAINT uq_prize_categories_template_name
  UNIQUE (prize_template_id, category_name);
