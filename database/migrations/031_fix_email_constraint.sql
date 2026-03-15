-- Fix email check constraint to allow empty string
-- This allows users to be created without specifying email

-- Drop existing constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS valid_email;

-- Add new constraint that allows empty string, NULL, or valid email format
ALTER TABLE users ADD CONSTRAINT valid_email 
  CHECK (email = '' OR email IS NULL OR email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');

COMMENT ON CONSTRAINT valid_email ON users IS 'Allows empty string, NULL, or valid email format';
