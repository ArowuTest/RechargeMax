-- Add spin_code column to spin_results table
-- This column stores a unique code for each spin result

ALTER TABLE spin_results 
  ADD COLUMN IF NOT EXISTS spin_code VARCHAR(30) UNIQUE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_spin_results_spin_code 
  ON spin_results(spin_code);

COMMENT ON COLUMN spin_results.spin_code IS 'Unique code for the spin result (e.g., SPIN_1234_1234567890)';
