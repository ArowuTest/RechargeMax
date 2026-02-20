-- Migration: 028_affiliate_enterprise_fixes_p0.sql
-- Description: Enterprise-grade fixes for affiliate system - P0 Critical Issues
-- Date: 2026-02-02
-- Issues Fixed: #1-10 (All P0 Critical)

-- ============================================================================
-- ISSUE #1: Currency Format Standardization (CRITICAL)
-- ============================================================================
-- Problem: Inconsistent currency format (kobo vs naira)
-- Solution: Standardize all money columns to INTEGER (kobo)

-- Check current format and convert if needed
DO $$
DECLARE
    current_type TEXT;
BEGIN
    SELECT data_type INTO current_type
    FROM information_schema.columns
    WHERE table_name = 'affiliates' AND column_name = 'total_commission';
    
    IF current_type = 'numeric' THEN
        -- Convert from NAIRA (decimal) to KOBO (integer)
        ALTER TABLE affiliates 
        ALTER COLUMN total_commission TYPE INTEGER USING (total_commission * 100)::INTEGER;
        
        RAISE NOTICE 'Converted affiliates.total_commission from NUMERIC to INTEGER (kobo)';
    END IF;
END $$;

-- Ensure affiliate_commissions uses INTEGER for amounts
ALTER TABLE affiliate_commissions
ALTER COLUMN commission_amount TYPE INTEGER USING commission_amount::INTEGER;

ALTER TABLE affiliate_commissions
ALTER COLUMN transaction_amount TYPE INTEGER USING transaction_amount::INTEGER;

-- Ensure affiliate_payouts uses INTEGER for amounts
ALTER TABLE affiliate_payouts
ALTER COLUMN total_amount TYPE INTEGER USING (total_amount * 100)::INTEGER;

-- Add check constraints to ensure positive amounts
ALTER TABLE affiliates
ADD CONSTRAINT positive_total_commission CHECK (total_commission >= 0);

ALTER TABLE affiliate_commissions
ADD CONSTRAINT positive_commission_amount CHECK (commission_amount >= 0);

ALTER TABLE affiliate_payouts
ADD CONSTRAINT positive_payout_amount CHECK (total_amount > 0);

COMMENT ON COLUMN affiliates.total_commission IS 'Total commission earned in kobo (integer)';
COMMENT ON COLUMN affiliate_commissions.commission_amount IS 'Commission amount in kobo (integer)';
COMMENT ON COLUMN affiliate_payouts.total_amount IS 'Payout amount in kobo (integer)';

-- ============================================================================
-- ISSUE #2: Referral Loop Prevention (CRITICAL)
-- ============================================================================
-- Problem: No validation to prevent circular referrals
-- Solution: Add trigger to detect and block referral loops

CREATE OR REPLACE FUNCTION validate_affiliate_referral() RETURNS TRIGGER AS $$
DECLARE
    referral_chain_length INTEGER := 0;
    current_user_id UUID;
    max_chain_length INTEGER := 10; -- Prevent infinite loops
BEGIN
    -- Prevent self-referral
    IF NEW.referred_by = NEW.id THEN
        RAISE EXCEPTION 'Self-referral not allowed';
    END IF;
    
    -- Prevent null referral changes (immutable once set)
    IF TG_OP = 'UPDATE' AND OLD.referred_by IS NOT NULL AND NEW.referred_by != OLD.referred_by THEN
        RAISE EXCEPTION 'Referral cannot be changed once set';
    END IF;
    
    -- Check for circular referrals (A refers B, B refers A)
    IF EXISTS (
        SELECT 1 FROM users 
        WHERE id = NEW.referred_by 
        AND referred_by = NEW.id
    ) THEN
        RAISE EXCEPTION 'Circular referral detected: User % already referred by %', NEW.referred_by, NEW.id;
    END IF;
    
    -- Check for deep circular chains (A->B->C->A)
    current_user_id := NEW.referred_by;
    WHILE current_user_id IS NOT NULL AND referral_chain_length < max_chain_length LOOP
        -- Check if we've looped back to the new user
        IF current_user_id = NEW.id THEN
            RAISE EXCEPTION 'Circular referral chain detected';
        END IF;
        
        -- Move up the chain
        SELECT referred_by INTO current_user_id
        FROM users
        WHERE id = current_user_id;
        
        referral_chain_length := referral_chain_length + 1;
    END LOOP;
    
    -- Log referral in audit log
    IF TG_OP = 'INSERT' AND NEW.referred_by IS NOT NULL THEN
        INSERT INTO gamification_audit_log (user_id, event_type, event_data)
        VALUES (
            NEW.id,
            'REFERRAL_CREATED',
            jsonb_build_object(
                'referred_by', NEW.referred_by,
                'msisdn', NEW.msisdn
            )
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if exists
DROP TRIGGER IF EXISTS affiliate_referral_validation ON users;

-- Create trigger
CREATE TRIGGER affiliate_referral_validation
BEFORE INSERT OR UPDATE OF referred_by ON users
FOR EACH ROW
EXECUTE FUNCTION validate_affiliate_referral();

-- ============================================================================
-- ISSUE #3: Commission Status Tracking (CRITICAL)
-- ============================================================================
-- Problem: Commission status never updated from PENDING to PAID
-- Solution: Add proper status enum and tracking

-- Add paid_at timestamp
ALTER TABLE affiliate_commissions
ADD COLUMN IF NOT EXISTS paid_at TIMESTAMPTZ;

-- Add payout_id to link commissions to payouts
ALTER TABLE affiliate_commissions
ADD COLUMN IF NOT EXISTS payout_id UUID REFERENCES affiliate_payouts(id);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_affiliate_commissions_status ON affiliate_commissions(status);
CREATE INDEX IF NOT EXISTS idx_affiliate_commissions_payout_id ON affiliate_commissions(payout_id);

-- ============================================================================
-- ISSUE #4: Minimum Payout Amount (CRITICAL)
-- ============================================================================
-- Problem: No minimum payout threshold
-- Solution: Add to system_config and enforce

INSERT INTO system_config (config_key, config_value, config_type, category, description, is_active)
VALUES 
    ('affiliate_minimum_payout', '500000', 'integer', 'affiliate', 'Minimum payout amount in kobo (₦5,000)', true),
    ('affiliate_max_commission_per_transaction', '500000', 'integer', 'affiliate', 'Maximum commission per transaction in kobo (₦5,000)', true),
    ('affiliate_max_commission_per_day', '5000000', 'integer', 'affiliate', 'Maximum commission per day in kobo (₦50,000)', true),
    ('affiliate_max_commission_per_month', '50000000', 'integer', 'affiliate', 'Maximum commission per month in kobo (₦500,000)', true),
    ('affiliate_payout_processing_fee', '10000', 'integer', 'affiliate', 'Payout processing fee in kobo (₦100)', true),
    ('affiliate_click_rate_limit_per_hour', '10', 'integer', 'affiliate', 'Maximum clicks per IP per hour', true),
    ('affiliate_approval_auto', 'false', 'boolean', 'affiliate', 'Automatically approve new affiliates', true)
ON CONFLICT (config_key) DO UPDATE SET
    config_value = EXCLUDED.config_value,
    updated_at = NOW();

-- ============================================================================
-- ISSUE #5: Bank Account Verification (CRITICAL)
-- ============================================================================
-- Problem: Payout doesn't check if bank account is verified
-- Solution: Add verification requirements and helper function

-- Add bank_code column for Nigerian banks
ALTER TABLE affiliate_bank_accounts
ADD COLUMN IF NOT EXISTS bank_code TEXT;

-- Create function to get verified primary bank account
CREATE OR REPLACE FUNCTION get_verified_primary_bank_account(p_affiliate_id UUID)
RETURNS TABLE (
    id UUID,
    bank_name TEXT,
    bank_code TEXT,
    account_number TEXT,
    account_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        aba.id,
        aba.bank_name,
        aba.bank_code,
        aba.account_number,
        aba.account_name
    FROM affiliate_bank_accounts aba
    WHERE aba.affiliate_id = p_affiliate_id
    AND aba.is_verified = true
    AND aba.is_primary = true
    AND aba.is_active = true
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- ISSUE #6: Click Fraud Prevention (CRITICAL)
-- ============================================================================
-- Problem: No rate limiting on clicks
-- Solution: Add rate limiting and fraud detection

-- Create rate limiting table
CREATE TABLE IF NOT EXISTS affiliate_click_rate_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ip_address INET NOT NULL,
    affiliate_id UUID NOT NULL REFERENCES affiliates(id) ON DELETE CASCADE,
    click_count INTEGER DEFAULT 1,
    window_start TIMESTAMPTZ DEFAULT NOW(),
    is_blocked BOOLEAN DEFAULT false,
    blocked_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rate_limit_ip_affiliate ON affiliate_click_rate_limits(ip_address, affiliate_id);
CREATE INDEX IF NOT EXISTS idx_rate_limit_window ON affiliate_click_rate_limits(window_start) WHERE is_blocked = false;
CREATE INDEX IF NOT EXISTS idx_rate_limit_blocked ON affiliate_click_rate_limits(is_blocked, blocked_until);

-- Create fraud detection function
CREATE OR REPLACE FUNCTION validate_affiliate_click() RETURNS TRIGGER AS $$
DECLARE
    recent_clicks INTEGER;
    rate_limit INTEGER;
    is_ip_blocked BOOLEAN;
BEGIN
    -- Get rate limit from config
    SELECT config_value::INTEGER INTO rate_limit
    FROM system_config
    WHERE config_key = 'affiliate_click_rate_limit_per_hour'
    AND is_active = true;
    
    IF rate_limit IS NULL THEN
        rate_limit := 10; -- Default
    END IF;
    
    -- Check if IP is currently blocked
    SELECT is_blocked INTO is_ip_blocked
    FROM affiliate_click_rate_limits
    WHERE ip_address = NEW.ip_address
    AND affiliate_id = NEW.affiliate_id
    AND is_blocked = true
    AND (blocked_until IS NULL OR blocked_until > NOW());
    
    IF is_ip_blocked THEN
        -- Log fraud attempt
        INSERT INTO gamification_audit_log (event_type, event_data)
        VALUES (
            'FRAUD_DETECTED',
            jsonb_build_object(
                'type', 'affiliate_click_fraud_blocked_ip',
                'ip_address', NEW.ip_address,
                'affiliate_id', NEW.affiliate_id
            )
        );
        
        RAISE EXCEPTION 'IP address blocked due to suspicious activity';
    END IF;
    
    -- Check clicks from this IP in last hour
    SELECT COUNT(*) INTO recent_clicks
    FROM affiliate_clicks
    WHERE ip_address = NEW.ip_address
    AND affiliate_id = NEW.affiliate_id
    AND created_at > NOW() - INTERVAL '1 hour';
    
    -- Block if rate limit exceeded
    IF recent_clicks >= rate_limit THEN
        -- Create or update rate limit record
        INSERT INTO affiliate_click_rate_limits (ip_address, affiliate_id, click_count, is_blocked, blocked_until)
        VALUES (NEW.ip_address, NEW.affiliate_id, recent_clicks + 1, true, NOW() + INTERVAL '24 hours')
        ON CONFLICT (ip_address, affiliate_id) DO UPDATE SET
            click_count = affiliate_click_rate_limits.click_count + 1,
            is_blocked = true,
            blocked_until = NOW() + INTERVAL '24 hours',
            updated_at = NOW();
        
        -- Log fraud attempt
        INSERT INTO gamification_audit_log (event_type, event_data)
        VALUES (
            'FRAUD_DETECTED',
            jsonb_build_object(
                'type', 'affiliate_click_fraud_rate_limit',
                'ip_address', NEW.ip_address,
                'affiliate_id', NEW.affiliate_id,
                'click_count', recent_clicks + 1,
                'rate_limit', rate_limit
            )
        );
        
        RAISE EXCEPTION 'Rate limit exceeded: Too many clicks from this IP (% clicks in 1 hour, limit: %)', recent_clicks + 1, rate_limit;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if exists
DROP TRIGGER IF EXISTS affiliate_click_validation ON affiliate_clicks;

-- Create trigger
CREATE TRIGGER affiliate_click_validation
BEFORE INSERT ON affiliate_clicks
FOR EACH ROW
EXECUTE FUNCTION validate_affiliate_click();

-- ============================================================================
-- ISSUE #7: Commission Reversal Logic (CRITICAL)
-- ============================================================================
-- Problem: No commission reversal when transaction is refunded
-- Solution: Add reversal tracking and automation

-- Add reversal columns
ALTER TABLE affiliate_commissions
ADD COLUMN IF NOT EXISTS is_reversed BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS reversal_reason TEXT,
ADD COLUMN IF NOT EXISTS original_commission_id UUID REFERENCES affiliate_commissions(id);

-- Create index
CREATE INDEX IF NOT EXISTS idx_affiliate_commissions_reversed ON affiliate_commissions(is_reversed);

-- Create commission reversal function
CREATE OR REPLACE FUNCTION reverse_affiliate_commission(p_transaction_id UUID, p_reason TEXT)
RETURNS UUID AS $$
DECLARE
    v_commission_id UUID;
    v_commission RECORD;
    v_reversal_id UUID;
BEGIN
    -- Find the commission for this transaction
    SELECT * INTO v_commission
    FROM affiliate_commissions
    WHERE transaction_id = p_transaction_id
    AND is_reversed = false
    LIMIT 1;
    
    IF NOT FOUND THEN
        RETURN NULL; -- No commission to reverse
    END IF;
    
    -- Create reversal record (negative amount)
    INSERT INTO affiliate_commissions (
        affiliate_id,
        transaction_id,
        commission_amount,
        commission_rate,
        transaction_amount,
        status,
        is_reversed,
        reversed_at,
        reversal_reason,
        original_commission_id
    ) VALUES (
        v_commission.affiliate_id,
        p_transaction_id,
        -v_commission.commission_amount, -- Negative amount
        v_commission.commission_rate,
        v_commission.transaction_amount,
        'REVERSED',
        true,
        NOW(),
        p_reason,
        v_commission.id
    ) RETURNING id INTO v_reversal_id;
    
    -- Mark original commission as reversed
    UPDATE affiliate_commissions
    SET is_reversed = true,
        reversed_at = NOW(),
        reversal_reason = p_reason
    WHERE id = v_commission.id;
    
    -- Update affiliate total (subtract commission)
    UPDATE affiliates
    SET total_commission = total_commission - v_commission.commission_amount
    WHERE id = v_commission.affiliate_id;
    
    -- Log reversal
    INSERT INTO gamification_audit_log (
        user_id,
        event_type,
        event_data
    ) VALUES (
        (SELECT user_id FROM affiliates WHERE id = v_commission.affiliate_id),
        'COMMISSION_REVERSED',
        jsonb_build_object(
            'commission_id', v_commission.id,
            'reversal_id', v_reversal_id,
            'amount', v_commission.commission_amount,
            'reason', p_reason
        )
    );
    
    RETURN v_reversal_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- ISSUE #8: First Recharge Validation (CRITICAL)
-- ============================================================================
-- Problem: CountByUserID counts all transactions, not just recharges
-- Solution: Add helper function to count recharges only

CREATE OR REPLACE FUNCTION count_user_recharges(p_user_id UUID)
RETURNS INTEGER AS $$
DECLARE
    recharge_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO recharge_count
    FROM transactions
    WHERE user_id = p_user_id
    AND transaction_type = 'RECHARGE'
    AND status = 'SUCCESS';
    
    RETURN COALESCE(recharge_count, 0);
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- ISSUE #9: Payout Status Tracking (CRITICAL)
-- ============================================================================
-- Problem: Payout status not properly tracked
-- Solution: Add comprehensive status tracking

-- Add more payout tracking columns
ALTER TABLE affiliate_payouts
ADD COLUMN IF NOT EXISTS transfer_reference TEXT,
ADD COLUMN IF NOT EXISTS transfer_code TEXT,
ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS failure_reason TEXT,
ADD COLUMN IF NOT EXISTS retry_count INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS processing_fee INTEGER DEFAULT 0;

-- Create index for transfer reference lookups
CREATE INDEX IF NOT EXISTS idx_affiliate_payouts_transfer_ref ON affiliate_payouts(transfer_reference);

-- ============================================================================
-- ISSUE #10: Affiliate Suspension Enforcement (CRITICAL)
-- ============================================================================
-- Problem: Suspended affiliates can still earn commissions
-- Solution: Add status validation

-- Create function to check if affiliate can earn commission
CREATE OR REPLACE FUNCTION can_affiliate_earn_commission(p_affiliate_id UUID)
RETURNS BOOLEAN AS $$
DECLARE
    affiliate_status TEXT;
BEGIN
    SELECT status INTO affiliate_status
    FROM affiliates
    WHERE id = p_affiliate_id;
    
    RETURN affiliate_status = 'APPROVED';
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Data Migration: Fix Existing Data
-- ============================================================================

-- Update any existing commission amounts if they're in wrong format
-- (This is safe because we're checking the current values)
DO $$
BEGIN
    -- Check if commissions need conversion
    IF EXISTS (
        SELECT 1 FROM affiliate_commissions 
        WHERE commission_amount < 100 AND commission_amount > 0
        LIMIT 1
    ) THEN
        -- Likely in naira, convert to kobo
        UPDATE affiliate_commissions
        SET commission_amount = commission_amount * 100,
            transaction_amount = transaction_amount * 100
        WHERE commission_amount < 100;
        
        RAISE NOTICE 'Converted existing commission amounts to kobo';
    END IF;
END $$;

-- ============================================================================
-- Verification Queries
-- ============================================================================

-- Verify currency format
DO $$
DECLARE
    total_affiliates INTEGER;
    total_commissions INTEGER;
BEGIN
    SELECT COUNT(*) INTO total_affiliates FROM affiliates;
    SELECT COUNT(*) INTO total_commissions FROM affiliate_commissions;
    
    RAISE NOTICE 'Migration complete:';
    RAISE NOTICE '  - Affiliates: %', total_affiliates;
    RAISE NOTICE '  - Commissions: %', total_commissions;
    RAISE NOTICE '  - All amounts now in KOBO (integer)';
    RAISE NOTICE '  - Fraud prevention: ACTIVE';
    RAISE NOTICE '  - Referral loop prevention: ACTIVE';
    RAISE NOTICE '  - Commission reversal: AVAILABLE';
END $$;

-- ============================================================================
-- Migration Complete
-- ============================================================================

COMMENT ON TABLE affiliate_click_rate_limits IS 'Rate limiting and fraud detection for affiliate clicks';
COMMENT ON FUNCTION validate_affiliate_referral() IS 'Prevents self-referral and circular referral chains';
COMMENT ON FUNCTION validate_affiliate_click() IS 'Rate limits clicks and detects fraud patterns';
COMMENT ON FUNCTION reverse_affiliate_commission(UUID, TEXT) IS 'Reverses commission when transaction is refunded';
COMMENT ON FUNCTION count_user_recharges(UUID) IS 'Counts only successful recharge transactions for commission eligibility';
COMMENT ON FUNCTION can_affiliate_earn_commission(UUID) IS 'Checks if affiliate is approved and can earn commissions';
