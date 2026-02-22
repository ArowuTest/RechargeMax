-- Convert prize_value from numeric(10,2) to bigint (storing in kobo/cents)
-- This matches the rest of the application which stores all amounts in kobo

-- Update wheel_prizes table
ALTER TABLE wheel_prizes 
  ALTER COLUMN prize_value TYPE BIGINT USING (prize_value * 100)::BIGINT;

-- Update spin_results table  
ALTER TABLE spin_results
  ALTER COLUMN prize_value TYPE BIGINT USING (prize_value * 100)::BIGINT;

COMMENT ON COLUMN wheel_prizes.prize_value IS 'Prize value in kobo (₦1 = 100 kobo)';
COMMENT ON COLUMN spin_results.prize_value IS 'Prize value in kobo (₦1 = 100 kobo)';
