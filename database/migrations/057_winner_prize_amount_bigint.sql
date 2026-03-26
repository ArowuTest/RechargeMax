-- Migration 057: Cast draw_winners.prize_amount from numeric(10,2) to bigint
-- The Winner entity uses *int64 for prize_amount but the DB column is numeric(10,2).
-- GORM fails to scan numeric into *int64, causing all winner queries to 500.
-- Amounts are stored in kobo (integer), so numeric precision is not needed.

ALTER TABLE public.draw_winners
  ALTER COLUMN prize_amount TYPE bigint USING prize_amount::bigint;
