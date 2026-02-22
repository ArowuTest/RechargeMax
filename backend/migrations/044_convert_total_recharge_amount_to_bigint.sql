-- Convert total_recharge_amount from numeric(12,2) to bigint (kobo)
-- This ensures consistency with the rest of the system where all amounts are in kobo

-- Convert existing values: numeric(12,2) in Naira → bigint in kobo
-- Example: 1000.50 → 100050 (multiply by 100)
ALTER TABLE users 
  ALTER COLUMN total_recharge_amount TYPE BIGINT 
  USING (total_recharge_amount * 100)::BIGINT;

-- Set default to 0
ALTER TABLE users 
  ALTER COLUMN total_recharge_amount SET DEFAULT 0;

COMMENT ON COLUMN users.total_recharge_amount IS 'Total recharge amount in kobo (₦1 = 100 kobo)';
