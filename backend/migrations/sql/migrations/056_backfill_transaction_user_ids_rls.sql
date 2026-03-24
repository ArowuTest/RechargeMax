-- Migration 056: Backfill NULL user_id in transactions (RLS-safe retry of 055)
-- Migration 055 ran but updated 0 rows because Row Level Security blocked the
-- UPDATE under the service connection. This migration uses SET LOCAL row_security = off
-- to bypass RLS for the session, then performs the same join-based backfill.

SET LOCAL row_security = off;

UPDATE transactions t
SET    user_id = u.id
FROM   users u
WHERE  t.msisdn = u.msisdn
  AND  t.user_id IS NULL;

SET LOCAL row_security = on;
