-- Add user_code column if not yet created (029 adds it; guard here for ordering)
ALTER TABLE users ADD COLUMN IF NOT EXISTS user_code VARCHAR(20);

-- Migration: Fix referral code unique constraint to allow NULL values
-- Date: 2026-02-19
-- Purpose: Allow multiple users with NULL referral codes while maintaining uniqueness for non-NULL values

-- Drop the existing unique constraint (it's a constraint, not just an index)
ALTER TABLE users DROP CONSTRAINT IF EXISTS idx_users_referral_code;

-- Drop the index if it exists separately
DROP INDEX IF EXISTS idx_users_referral_code;

-- Create a partial unique index that only applies to non-NULL referral codes
-- This allows multiple NULL values while ensuring uniqueness for actual referral codes
CREATE UNIQUE INDEX idx_users_referral_code ON users (referral_code) WHERE referral_code IS NOT NULL;

-- Add comment for documentation
COMMENT ON INDEX idx_users_referral_code IS 'Unique constraint on referral_code, allowing multiple NULL values';

-- Also fix user_code unique constraint to allow NULL values
ALTER TABLE users DROP CONSTRAINT IF EXISTS idx_users_user_code;
DROP INDEX IF EXISTS idx_users_user_code;
CREATE UNIQUE INDEX idx_users_user_code ON users (user_code) WHERE user_code IS NOT NULL AND user_code != '';

COMMENT ON INDEX idx_users_user_code IS 'Unique constraint on user_code, allowing multiple NULL or empty values';
