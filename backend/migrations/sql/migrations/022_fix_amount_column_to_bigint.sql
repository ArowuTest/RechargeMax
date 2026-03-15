-- Migration: Fix amount column type mismatch
-- Issue: Column is numeric(10,2) but Go code expects bigint
-- Solution: Convert to bigint and store amounts in kobo

-- Note: Existing data is already in kobo (frontend was multiplying by 100)
-- So we just need to change the column type, no data conversion needed

BEGIN;

-- Change column type to bigint
-- The USING clause converts numeric to bigint
ALTER TABLE transactions 
ALTER COLUMN amount TYPE bigint USING (amount::bigint);

-- Add comment for clarity
COMMENT ON COLUMN transactions.amount IS 'Amount in kobo (1 Naira = 100 kobo)';

COMMIT;
