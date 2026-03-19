-- Migration 039: Fix corrupt prize_value in wheel_prizes table
-- Pre-seed wheel_prizes rows had astronomically wrong prize_value (e.g. 2e13 kobo).
-- This migration corrects them based on prize_name patterns.
-- Safe to re-run (uses ON CONFLICT / WHERE clauses to only touch corrupt rows).

-- Threshold: any prize_value > ₦1,000,000 (100,000,000 kobo) is considered corrupt
-- We UPDATE based on prize_name matching the seeded canonical names from migration 037.

UPDATE wheel_prizes SET prize_value = 10000,  updated_at = NOW()
  WHERE prize_name = '₦100 Airtime'  AND prize_value > 100000000;

UPDATE wheel_prizes SET prize_value = 20000,  updated_at = NOW()
  WHERE prize_name = '₦200 Airtime'  AND prize_value > 100000000;

UPDATE wheel_prizes SET prize_value = 50000,  updated_at = NOW()
  WHERE prize_name = '500MB Data'     AND prize_value > 100000000;

UPDATE wheel_prizes SET prize_value = 100000, updated_at = NOW()
  WHERE prize_name = '1GB Data'       AND prize_value > 100000000;

UPDATE wheel_prizes SET prize_value = 10000,  updated_at = NOW()
  WHERE prize_name = '₦100 Cash'     AND prize_value > 100000000;

UPDATE wheel_prizes SET prize_value = 20000,  updated_at = NOW()
  WHERE prize_name = '₦200 Cash'     AND prize_value > 100000000;

UPDATE wheel_prizes SET prize_value = 50000,  updated_at = NOW()
  WHERE prize_name = '₦500 Cash'     AND prize_value > 100000000;

UPDATE wheel_prizes SET prize_value = 100000, updated_at = NOW()
  WHERE prize_name = '₦1000 Cash'    AND prize_value > 100000000;

-- Also fix the corresponding spin_results rows so the copied prize_value is correct too
-- (these are the existing "won" prizes before migration 037)
UPDATE spin_results sr
SET    prize_value = wp.prize_value
FROM   wheel_prizes wp
WHERE  sr.prize_id = wp.id
  AND  sr.prize_value > 100000000;
