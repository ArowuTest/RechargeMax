-- Add user_code column to users table
-- This provides a unique, human-readable identifier for each user

ALTER TABLE users ADD COLUMN IF NOT EXISTS user_code VARCHAR(20);

-- Create unique index on user_code
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_user_code ON users(user_code) WHERE user_code IS NOT NULL;

-- Add comment
COMMENT ON COLUMN users.user_code IS 'Unique human-readable user code for identification';
