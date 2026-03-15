-- Migration: Normalize MSISDN in existing tables only
-- Date: February 19, 2026
-- Purpose: Normalize phone numbers to international format (234XXXXXXXXXX)

BEGIN;

-- 1. Normalize otp_verifications table
UPDATE otp_verifications
SET msisdn = '234' || SUBSTRING(msisdn FROM 2)
WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;

-- 2. Normalize otps table (if it has msisdn column)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'otps' AND column_name = 'msisdn'
    ) THEN
        UPDATE otps
        SET msisdn = '234' || SUBSTRING(msisdn FROM 2)
        WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;
    END IF;
END $$;

-- 3. Normalize spin_results table
UPDATE spin_results
SET msisdn = '234' || SUBSTRING(msisdn FROM 2)
WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;

-- 4. Add check constraints to enforce format (drop if exists first)
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

-- 5. Add indexes for performance (if not already exist)
CREATE INDEX IF NOT EXISTS idx_users_msisdn ON users(msisdn);
CREATE INDEX IF NOT EXISTS idx_otp_verifications_msisdn ON otp_verifications(msisdn);
CREATE INDEX IF NOT EXISTS idx_spin_results_msisdn ON spin_results(msisdn);

-- 6. Verify migration
DO $$
DECLARE
    local_count INTEGER;
BEGIN
    SELECT 
        (SELECT COUNT(*) FROM users WHERE msisdn LIKE '0%') +
        (SELECT COUNT(*) FROM otp_verifications WHERE msisdn LIKE '0%') +
        (SELECT COUNT(*) FROM spin_results WHERE msisdn LIKE '0%')
    INTO local_count;
    
    IF local_count > 0 THEN
        RAISE WARNING 'Migration incomplete: % records still in local format', local_count;
    ELSE
        RAISE NOTICE 'Migration successful: All MSISDN normalized to international format';
    END IF;
END $$;

COMMIT;
