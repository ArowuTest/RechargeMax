-- ============================================================================
-- MIGRATION: Standardize all amounts to kobo (BIGINT)
-- Date: 2026-02-01
-- Purpose: Fix critical amount storage inconsistency
-- 
-- PROBLEM: Database stores Naira (DECIMAL), code expects kobo (int64)
-- SOLUTION: Convert all amount columns to BIGINT storing kobo
-- 
-- 1 Naira = 100 kobo
-- Example: ₦1,000 = 100,000 kobo
-- ============================================================================

BEGIN;

-- ============================================================================
-- STEP 1: Add new columns with _kobo suffix
-- ============================================================================

ALTER TABLE transactions_2026_01_30_14_00 
ADD COLUMN amount_kobo BIGINT;

ALTER TABLE vtu_transactions 
ADD COLUMN amount_kobo BIGINT;

ALTER TABLE wheel_prizes_2026_01_30_14_00
ADD COLUMN prize_value_kobo BIGINT,
ADD COLUMN minimum_recharge_kobo BIGINT;

ALTER TABLE draw_winners_2026_01_30_14_00
ADD COLUMN prize_amount_kobo BIGINT;

-- ============================================================================
-- STEP 2: Convert existing data (Naira * 100 = kobo)
-- ============================================================================

-- Transactions: Convert amount from Naira to kobo
UPDATE transactions_2026_01_30_14_00 
SET amount_kobo = CAST(ROUND(amount * 100) AS BIGINT)
WHERE amount IS NOT NULL;

-- VTU Transactions: Convert amount from Naira to kobo
UPDATE vtu_transactions 
SET amount_kobo = CAST(ROUND(amount * 100) AS BIGINT)
WHERE amount IS NOT NULL;

-- Wheel Prizes: Convert prize_value and minimum_recharge from Naira to kobo
UPDATE wheel_prizes_2026_01_30_14_00 
SET prize_value_kobo = CAST(ROUND(prize_value * 100) AS BIGINT),
    minimum_recharge_kobo = CAST(ROUND(minimum_recharge * 100) AS BIGINT)
WHERE prize_value IS NOT NULL;

-- Draw Winners: Convert prize_amount from Naira to kobo
UPDATE draw_winners_2026_01_30_14_00
SET prize_amount_kobo = CAST(ROUND(prize_amount * 100) AS BIGINT)
WHERE prize_amount IS NOT NULL;

-- ============================================================================
-- STEP 3: Verify conversion (should return 0 for all)
-- ============================================================================

DO $$
DECLARE
    null_count INTEGER;
BEGIN
    -- Check transactions
    SELECT COUNT(*) INTO null_count 
    FROM transactions_2026_01_30_14_00 
    WHERE amount IS NOT NULL AND amount_kobo IS NULL;
    
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Conversion failed: % transactions have NULL amount_kobo', null_count;
    END IF;
    
    -- Check vtu_transactions
    SELECT COUNT(*) INTO null_count 
    FROM vtu_transactions 
    WHERE amount IS NOT NULL AND amount_kobo IS NULL;
    
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Conversion failed: % vtu_transactions have NULL amount_kobo', null_count;
    END IF;
    
    -- Check wheel_prizes
    SELECT COUNT(*) INTO null_count 
    FROM wheel_prizes_2026_01_30_14_00 
    WHERE prize_value IS NOT NULL AND prize_value_kobo IS NULL;
    
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Conversion failed: % wheel_prizes have NULL prize_value_kobo', null_count;
    END IF;
    
    -- Check draw_winners
    SELECT COUNT(*) INTO null_count 
    FROM draw_winners_2026_01_30_14_00 
    WHERE prize_amount IS NOT NULL AND prize_amount_kobo IS NULL;
    
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Conversion failed: % draw_winners have NULL prize_amount_kobo', null_count;
    END IF;
    
    RAISE NOTICE 'Verification passed: All amounts converted successfully';
END $$;

-- ============================================================================
-- STEP 4: Make new columns NOT NULL (with defaults for future inserts)
-- ============================================================================

ALTER TABLE transactions_2026_01_30_14_00 
ALTER COLUMN amount_kobo SET NOT NULL;

ALTER TABLE vtu_transactions 
ALTER COLUMN amount_kobo SET NOT NULL;

ALTER TABLE wheel_prizes_2026_01_30_14_00
ALTER COLUMN prize_value_kobo SET NOT NULL,
ALTER COLUMN minimum_recharge_kobo SET NOT NULL,
ALTER COLUMN minimum_recharge_kobo SET DEFAULT 0;

ALTER TABLE draw_winners_2026_01_30_14_00
ALTER COLUMN prize_amount_kobo SET NOT NULL;

-- ============================================================================
-- STEP 5: Rename old columns (keep as backup for safety)
-- ============================================================================

ALTER TABLE transactions_2026_01_30_14_00 
RENAME COLUMN amount TO amount_naira_deprecated;

ALTER TABLE vtu_transactions 
RENAME COLUMN amount TO amount_naira_deprecated;

ALTER TABLE wheel_prizes_2026_01_30_14_00
RENAME COLUMN prize_value TO prize_value_naira_deprecated,
RENAME COLUMN minimum_recharge TO minimum_recharge_naira_deprecated;

ALTER TABLE draw_winners_2026_01_30_14_00
RENAME COLUMN prize_amount TO prize_amount_naira_deprecated;

-- ============================================================================
-- STEP 6: Rename new columns to standard names
-- ============================================================================

ALTER TABLE transactions_2026_01_30_14_00 
RENAME COLUMN amount_kobo TO amount;

ALTER TABLE vtu_transactions 
RENAME COLUMN amount_kobo TO amount;

ALTER TABLE wheel_prizes_2026_01_30_14_00
RENAME COLUMN prize_value_kobo TO prize_value,
RENAME COLUMN minimum_recharge_kobo TO minimum_recharge;

ALTER TABLE draw_winners_2026_01_30_14_00
RENAME COLUMN prize_amount_kobo TO prize_amount;

-- ============================================================================
-- STEP 7: Update constraints
-- ============================================================================

-- Drop old constraint and add new one for positive amounts
ALTER TABLE transactions_2026_01_30_14_00 
DROP CONSTRAINT IF EXISTS positive_amount,
ADD CONSTRAINT positive_amount CHECK (amount > 0);

-- Add constraint for wheel prizes
ALTER TABLE wheel_prizes_2026_01_30_14_00
DROP CONSTRAINT IF EXISTS positive_prize_value,
ADD CONSTRAINT positive_prize_value CHECK (prize_value >= 0);

-- ============================================================================
-- STEP 8: Add helpful comments
-- ============================================================================

COMMENT ON COLUMN transactions_2026_01_30_14_00.amount IS 
'Amount in kobo (1 Naira = 100 kobo). Example: ₦1,000 = 100000 kobo';

COMMENT ON COLUMN vtu_transactions.amount IS 
'Amount in kobo (1 Naira = 100 kobo). Example: ₦1,000 = 100000 kobo';

COMMENT ON COLUMN wheel_prizes_2026_01_30_14_00.prize_value IS 
'Prize value in kobo (1 Naira = 100 kobo). Example: ₦50 = 5000 kobo';

COMMENT ON COLUMN wheel_prizes_2026_01_30_14_00.minimum_recharge IS 
'Minimum recharge in kobo (1 Naira = 100 kobo). Example: ₦1,000 = 100000 kobo';

COMMENT ON COLUMN draw_winners_2026_01_30_14_00.prize_amount IS 
'Prize amount in kobo (1 Naira = 100 kobo). Example: ₦10,000 = 1000000 kobo';

-- ============================================================================
-- STEP 9: Log migration completion
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE '✅ Amount standardization migration completed successfully';
    RAISE NOTICE '   All amounts now stored as BIGINT in kobo';
    RAISE NOTICE '   Old Naira columns renamed with _deprecated suffix';
    RAISE NOTICE '   To drop deprecated columns after verification, run:';
    RAISE NOTICE '   ALTER TABLE transactions_2026_01_30_14_00 DROP COLUMN amount_naira_deprecated;';
END $$;

COMMIT;

-- ============================================================================
-- ROLLBACK SCRIPT (if needed - run manually)
-- ============================================================================
-- 
-- BEGIN;
-- 
-- -- Rename current columns back to _kobo
-- ALTER TABLE transactions_2026_01_30_14_00 RENAME COLUMN amount TO amount_kobo;
-- ALTER TABLE vtu_transactions RENAME COLUMN amount TO amount_kobo;
-- ALTER TABLE wheel_prizes_2026_01_30_14_00 
--     RENAME COLUMN prize_value TO prize_value_kobo,
--     RENAME COLUMN minimum_recharge TO minimum_recharge_kobo;
-- ALTER TABLE draw_winners_2026_01_30_14_00 RENAME COLUMN prize_amount TO prize_amount_kobo;
-- 
-- -- Restore old columns
-- ALTER TABLE transactions_2026_01_30_14_00 RENAME COLUMN amount_naira_deprecated TO amount;
-- ALTER TABLE vtu_transactions RENAME COLUMN amount_naira_deprecated TO amount;
-- ALTER TABLE wheel_prizes_2026_01_30_14_00 
--     RENAME COLUMN prize_value_naira_deprecated TO prize_value,
--     RENAME COLUMN minimum_recharge_naira_deprecated TO minimum_recharge;
-- ALTER TABLE draw_winners_2026_01_30_14_00 RENAME COLUMN prize_amount_naira_deprecated TO prize_amount;
-- 
-- -- Drop kobo columns
-- ALTER TABLE transactions_2026_01_30_14_00 DROP COLUMN amount_kobo;
-- ALTER TABLE vtu_transactions DROP COLUMN amount_kobo;
-- ALTER TABLE wheel_prizes_2026_01_30_14_00 
--     DROP COLUMN prize_value_kobo,
--     DROP COLUMN minimum_recharge_kobo;
-- ALTER TABLE draw_winners_2026_01_30_14_00 DROP COLUMN prize_amount_kobo;
-- 
-- COMMIT;
