-- Migration 040: Fix remaining corrupt prize_value in wheel_prizes + spin_results
-- 
-- Context:
--   Migration 037 seeded wheel_prizes with ON CONFLICT DO NOTHING.
--   If wheel_prizes already existed (with corrupt values from pre-037 admin setup),
--   the correct values were never applied — the insert was silently skipped.
--
--   Migration 039 fixed values > 100,000,000 kobo but missed "moderately corrupt"
--   rows where the value was wrong but below that threshold
--   (e.g. prize_value = 2,000,000 for a ₦200 prize that should be 20,000 kobo).
--
--   Observed symptoms:
--     ₦200 Airtime showing ₦20,000  → prize_value = 2,000,000 (100x too large)
--     ₦200 Cash   showing ₦20,000  → prize_value = 2,000,000 (100x too large)
--     ₦100 Airtime showing ₦1M     → prize_value = 100,000,000 (boundary case, = cap)
--
-- Fix strategy:
--   For each canonical prize name, force the correct kobo value on ANY row
--   where the current value is wrong (i.e. not equal to the expected value).
--   Then backfill spin_results via the prize_id FK.
--
-- This migration is idempotent — re-running it on already-correct data is a no-op.

-- ── Step 1: Fix wheel_prizes to canonical kobo values ────────────────────────

UPDATE wheel_prizes SET prize_value = 10000,  updated_at = NOW()
  WHERE prize_name = '₦100 Airtime'  AND prize_value <> 10000;

UPDATE wheel_prizes SET prize_value = 20000,  updated_at = NOW()
  WHERE prize_name = '₦200 Airtime'  AND prize_value <> 20000;

UPDATE wheel_prizes SET prize_value = 50000,  updated_at = NOW()
  WHERE prize_name = '500MB Data'    AND prize_value <> 50000;

UPDATE wheel_prizes SET prize_value = 100000, updated_at = NOW()
  WHERE prize_name = '1GB Data'      AND prize_value <> 100000;

UPDATE wheel_prizes SET prize_value = 10000,  updated_at = NOW()
  WHERE prize_name = '₦100 Cash'    AND prize_value <> 10000;

UPDATE wheel_prizes SET prize_value = 20000,  updated_at = NOW()
  WHERE prize_name = '₦200 Cash'    AND prize_value <> 20000;

UPDATE wheel_prizes SET prize_value = 50000,  updated_at = NOW()
  WHERE prize_name = '₦500 Cash'    AND prize_value <> 50000;

UPDATE wheel_prizes SET prize_value = 100000, updated_at = NOW()
  WHERE prize_name = '₦1000 Cash'   AND prize_value <> 100000;

-- ── Step 2: Backfill spin_results from the (now correct) wheel_prizes ─────────
-- Any spin_results row whose prize_id points to a wheel_prizes row will inherit
-- the corrected value. This fixes all historical spins regardless of when they
-- were played.

UPDATE spin_results sr
SET    prize_value = wp.prize_value,
       updated_at  = NOW()
FROM   wheel_prizes wp
WHERE  sr.prize_id = wp.id
  AND  sr.prize_value <> wp.prize_value;

-- ── Step 3: Fix orphan spin_results (prize_id IS NULL) by prize_name ─────────
-- Some spins may have been written before a valid prize_id was set, or the FK
-- was not populated. For those, use the prize_name to find the correct value.

UPDATE spin_results sr
SET    prize_value = wp.prize_value,
       updated_at  = NOW()
FROM   wheel_prizes wp
WHERE  sr.prize_id IS NULL
  AND  sr.prize_name = wp.prize_name
  AND  sr.prize_value <> wp.prize_value;

-- Verify (informational — check logs after migration)
DO $$
DECLARE
  wp_corrupt INT;
  sr_corrupt INT;
BEGIN
  SELECT COUNT(*) INTO wp_corrupt FROM wheel_prizes
    WHERE (prize_name = '₦100 Airtime'  AND prize_value <> 10000)
       OR (prize_name = '₦200 Airtime'  AND prize_value <> 20000)
       OR (prize_name = '500MB Data'    AND prize_value <> 50000)
       OR (prize_name = '1GB Data'      AND prize_value <> 100000)
       OR (prize_name = '₦100 Cash'     AND prize_value <> 10000)
       OR (prize_name = '₦200 Cash'     AND prize_value <> 20000)
       OR (prize_name = '₦500 Cash'     AND prize_value <> 50000)
       OR (prize_name = '₦1000 Cash'    AND prize_value <> 100000);

  SELECT COUNT(*) INTO sr_corrupt FROM spin_results sr
    JOIN wheel_prizes wp ON sr.prize_id = wp.id
    WHERE sr.prize_value <> wp.prize_value;

  RAISE NOTICE '040 verification: wheel_prizes still-corrupt=%, spin_results still-mismatched=%',
    wp_corrupt, sr_corrupt;
END $$;
