-- Migration 055: Backfill NULL user_id in transactions from msisdn
-- All existing transactions have user_id = NULL because ProcessSuccessfulPayment
-- only linked user_id in the guest-create path, not the existing-user path.
-- This one-time fix joins transactions to users on msisdn to populate user_id.
--
-- Row Level Security is temporarily disabled for this session-level UPDATE so
-- the service-role connection can write without hitting policy restrictions.

SET LOCAL row_security = off;

UPDATE transactions t
SET    user_id = u.id
FROM   users u
WHERE  t.msisdn = u.msisdn
  AND  t.user_id IS NULL;

SET LOCAL row_security = on;
