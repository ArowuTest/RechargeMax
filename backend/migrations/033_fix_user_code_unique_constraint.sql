-- Fix user_code unique constraint to allow multiple empty strings
-- Drop the partial unique index and recreate without the WHERE clause
-- This allows multiple users with empty user_code

DROP INDEX IF EXISTS idx_users_user_code;

-- Don't create any unique constraint on user_code for now
-- Users can have empty user_code until they're assigned one

COMMENT ON COLUMN users.user_code IS 'Optional unique user code for identification (can be empty)';
