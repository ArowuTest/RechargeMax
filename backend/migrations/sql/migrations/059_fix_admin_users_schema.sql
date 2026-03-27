-- Migration 059: Fix admin_users schema gaps
-- 1) Add last_login column (entity has both last_login and last_login_at)
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS last_login TIMESTAMPTZ;

-- 2) Expand role CHECK constraint to include VIEWER and SUPPORT
ALTER TABLE admin_users DROP CONSTRAINT IF EXISTS admin_users_role_check;
ALTER TABLE admin_users ADD CONSTRAINT admin_users_role_check
  CHECK (role = ANY (ARRAY[
    'SUPER_ADMIN'::text,
    'ADMIN'::text,
    'MODERATOR'::text,
    'SUPPORT'::text,
    'VIEWER'::text
  ]));
