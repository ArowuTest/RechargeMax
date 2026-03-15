-- Migration 016: Normalize all MSISDN data to international format (234XXXXXXXXXX)
-- Rewritten as safe DO blocks to handle tables that may not exist yet in migration order.

DO $$ BEGIN
  UPDATE users SET msisdn = '234' || SUBSTRING(msisdn FROM 2) WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;
EXCEPTION WHEN undefined_column THEN NULL;
END $$;

DO $$ BEGIN
  UPDATE otp_verifications SET msisdn = '234' || SUBSTRING(msisdn FROM 2) WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;
EXCEPTION WHEN undefined_table OR undefined_column THEN NULL;
END $$;

DO $$ BEGIN
  UPDATE transactions SET msisdn = '234' || SUBSTRING(msisdn FROM 2) WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;
EXCEPTION WHEN undefined_table OR undefined_column THEN NULL;
END $$;

DO $$ BEGIN
  UPDATE daily_subscriptions SET msisdn = '234' || SUBSTRING(msisdn FROM 2) WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;
EXCEPTION WHEN undefined_table OR undefined_column THEN NULL;
END $$;

DO $$ BEGIN
  UPDATE spin_results SET msisdn = '234' || SUBSTRING(msisdn FROM 2) WHERE msisdn LIKE '0%' AND LENGTH(msisdn) = 11;
EXCEPTION WHEN undefined_table OR undefined_column THEN NULL;
END $$;

-- Add check constraints
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_msisdn_format;
ALTER TABLE users ADD CONSTRAINT chk_users_msisdn_format CHECK (msisdn ~ '^234[7-9][0-1][0-9]{8}$');

ALTER TABLE otp_verifications DROP CONSTRAINT IF EXISTS chk_otp_msisdn_format;
ALTER TABLE otp_verifications ADD CONSTRAINT chk_otp_msisdn_format CHECK (msisdn ~ '^234[7-9][0-1][0-9]{8}$');

CREATE INDEX IF NOT EXISTS idx_otp_verifications_msisdn ON otp_verifications(msisdn);
CREATE INDEX IF NOT EXISTS idx_transactions_msisdn2 ON transactions(msisdn);
CREATE INDEX IF NOT EXISTS idx_spin_results_msisdn2 ON spin_results(msisdn);
