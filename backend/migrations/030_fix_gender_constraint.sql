-- Fix gender check constraint to allow empty string
-- This allows users to be created without specifying gender

-- Drop existing constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_gender_check;

-- Add new constraint that allows empty string or NULL
ALTER TABLE users ADD CONSTRAINT users_gender_check 
  CHECK (gender IN ('MALE', 'FEMALE', 'OTHER', '') OR gender IS NULL);

COMMENT ON CONSTRAINT users_gender_check ON users IS 'Allows MALE, FEMALE, OTHER, empty string, or NULL';
