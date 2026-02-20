-- Migration: Normalize all MSISDN data to international format (234XXXXXXXXXX)
-- Date: February 19, 2026
-- Purpose: Ensure consistent phone number format across all tables

BEGIN;

-- 1. Normalize otp_verifications table
UPDATE otp_verifications
SET msisdn = '234' || SUBSTRING(msisdn FROM 2)
WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;

-- 2. Normalize recharge_transactions table (recipient_msisdn)
UPDATE recharge_transactions
SET recipient_msisdn = '234' || SUBSTRING(recipient_msisdn FROM 2)
WHERE recipient_msisdn LIKE '0%' AND LENGTH(recipient_msisdn) = 11;

-- 3. Normalize daily_draw_subscriptions table
UPDATE daily_draw_subscriptions
SET msisdn = '234' || SUBSTRING(msisdn FROM 2)
WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;

-- 4. Normalize daily_draw_winners table
UPDATE daily_draw_winners
SET msisdn = '234' || SUBSTRING(msisdn FROM 2)
WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;

-- 5. Normalize affiliates table
UPDATE affiliates
SET msisdn = '234' || SUBSTRING(msisdn FROM 2)
WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;

-- 6. Normalize commissions table
UPDATE commissions
SET msisdn = '234' || SUBSTRING(msisdn FROM 2)
WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;

-- 7. Normalize wallet_transactions table
UPDATE wallet_transactions
SET msisdn = '234' || SUBSTRING(msisdn FROM 2)
WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;

-- 8. Normalize notifications table
UPDATE notifications
SET msisdn = '234' || SUBSTRING(msisdn FROM 2)
WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;

-- 9. Normalize spin_results table
UPDATE spin_results
SET msisdn = '234' || SUBSTRING(msisdn FROM 2)
WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;

-- 10. Add check constraints to enforce format
ALTER TABLE users
DROP CONSTRAINT IF EXISTS chk_users_msisdn_format;

ALTER TABLE users
ADD CONSTRAINT chk_users_msisdn_format
CHECK (msisdn ~ '^234[7-9][0-1][0-9]{8}$');

ALTER TABLE otp_verifications
DROP CONSTRAINT IF EXISTS chk_otp_msisdn_format;

ALTER TABLE otp_verifications
ADD CONSTRAINT chk_otp_msisdn_format
CHECK (msisdn ~ '^234[7-9][0-1][0-9]{8}$');

-- 11. Add indexes for performance (if not already exist)
CREATE INDEX IF NOT EXISTS idx_otp_verifications_msisdn ON otp_verifications(msisdn);
CREATE INDEX IF NOT EXISTS idx_recharge_transactions_recipient_msisdn ON recharge_transactions(recipient_msisdn);
CREATE INDEX IF NOT EXISTS idx_daily_draw_subscriptions_msisdn ON daily_draw_subscriptions(msisdn);
CREATE INDEX IF NOT EXISTS idx_daily_draw_winners_msisdn ON daily_draw_winners(msisdn);
CREATE INDEX IF NOT EXISTS idx_affiliates_msisdn ON affiliates(msisdn);
CREATE INDEX IF NOT EXISTS idx_commissions_msisdn ON commissions(msisdn);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_msisdn ON wallet_transactions(msisdn);
CREATE INDEX IF NOT EXISTS idx_notifications_msisdn ON notifications(msisdn);
CREATE INDEX IF NOT EXISTS idx_spin_results_msisdn ON spin_results(msisdn);

-- 12. Verify migration (count remaining local format numbers)
DO $$
DECLARE
    local_count INTEGER;
BEGIN
    SELECT 
        (SELECT COUNT(*) FROM users WHERE msisdn LIKE '0%') +
        (SELECT COUNT(*) FROM otp_verifications WHERE msisdn LIKE '0%') +
        (SELECT COUNT(*) FROM recharge_transactions WHERE recipient_msisdn LIKE '0%') +
        (SELECT COUNT(*) FROM daily_draw_subscriptions WHERE msisdn LIKE '0%') +
        (SELECT COUNT(*) FROM daily_draw_winners WHERE msisdn LIKE '0%') +
        (SELECT COUNT(*) FROM affiliates WHERE msisdn LIKE '0%') +
        (SELECT COUNT(*) FROM commissions WHERE msisdn LIKE '0%') +
        (SELECT COUNT(*) FROM wallet_transactions WHERE msisdn LIKE '0%') +
        (SELECT COUNT(*) FROM notifications WHERE msisdn LIKE '0%') +
        (SELECT COUNT(*) FROM spin_results WHERE msisdn LIKE '0%')
    INTO local_count;
    
    IF local_count > 0 THEN
        RAISE EXCEPTION 'Migration incomplete: % records still in local format', local_count;
    ELSE
        RAISE NOTICE 'Migration successful: All MSISDN normalized to international format';
    END IF;
END $$;

COMMIT;
