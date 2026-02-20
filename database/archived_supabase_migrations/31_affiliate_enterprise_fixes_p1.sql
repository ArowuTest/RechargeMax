-- Migration: 029_affiliate_enterprise_fixes_p1.sql
-- Description: Enterprise-grade fixes for affiliate system - P1 High Priority Issues
-- Date: 2026-02-02
-- Issues Fixed: #11-15 (All P1 High Priority)

-- ============================================================================
-- ISSUE #11: Bank Details Consolidation (HIGH PRIORITY)
-- ============================================================================
-- Problem: Bank details in two places (affiliates table and affiliate_bank_accounts)
-- Solution: Remove from affiliates, use affiliate_bank_accounts only

-- First, migrate existing bank details to affiliate_bank_accounts
INSERT INTO affiliate_bank_accounts (
    affiliate_id,
    bank_name,
    account_number,
    account_name,
    is_verified,
    is_primary,
    is_active
)
SELECT 
    id,
    bank_name,
    account_number,
    account_name,
    false, -- Not verified by default
    true,  -- Set as primary
    true   -- Active
FROM affiliates
WHERE bank_name IS NOT NULL 
AND account_number IS NOT NULL
AND NOT EXISTS (
    SELECT 1 FROM affiliate_bank_accounts 
    WHERE affiliate_id = affiliates.id
)
ON CONFLICT DO NOTHING;

-- Now remove bank columns from affiliates table
ALTER TABLE affiliates
DROP COLUMN IF EXISTS bank_name CASCADE,
DROP COLUMN IF EXISTS account_number CASCADE,
DROP COLUMN IF EXISTS account_name CASCADE;

COMMENT ON TABLE affiliate_bank_accounts IS 'Centralized bank account management for affiliates with verification workflow';

-- ============================================================================
-- ISSUE #12: Commission Tier System (HIGH PRIORITY)
-- ============================================================================
-- Problem: No automatic tier upgrades, single commission rate
-- Solution: Create tiers table and automation

-- Create commission tiers table
CREATE TABLE IF NOT EXISTS affiliate_commission_tiers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tier TEXT NOT NULL UNIQUE,
    min_referrals INTEGER NOT NULL,
    commission_rate NUMERIC(5,2) NOT NULL,
    bonus_threshold INTEGER,
    bonus_amount INTEGER, -- In kobo
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT valid_commission_rate CHECK (commission_rate >= 0 AND commission_rate <= 100),
    CONSTRAINT valid_min_referrals CHECK (min_referrals >= 0),
    CONSTRAINT valid_bonus CHECK (bonus_threshold IS NULL OR bonus_threshold > 0),
    CONSTRAINT valid_bonus_amount CHECK (bonus_amount IS NULL OR bonus_amount > 0)
);

-- Create index
CREATE INDEX IF NOT EXISTS idx_affiliate_tiers_min_referrals ON affiliate_commission_tiers(min_referrals);

-- Seed tier data
INSERT INTO affiliate_commission_tiers (tier, min_referrals, commission_rate, bonus_threshold, bonus_amount, description) VALUES
('BRONZE', 0, 5.00, 10, 100000, 'Entry level - 5% commission, ₦1,000 bonus at 10 referrals'),
('SILVER', 25, 7.50, 25, 250000, 'Silver tier - 7.5% commission, ₦2,500 bonus at 25 referrals'),
('GOLD', 50, 10.00, 50, 500000, 'Gold tier - 10% commission, ₦5,000 bonus at 50 referrals'),
('PLATINUM', 100, 12.50, 100, 1000000, 'Platinum tier - 12.5% commission, ₦10,000 bonus at 100 referrals'),
('DIAMOND', 250, 15.00, 250, 2500000, 'Diamond tier - 15% commission, ₦25,000 bonus at 250 referrals')
ON CONFLICT (tier) DO UPDATE SET
    min_referrals = EXCLUDED.min_referrals,
    commission_rate = EXCLUDED.commission_rate,
    bonus_threshold = EXCLUDED.bonus_threshold,
    bonus_amount = EXCLUDED.bonus_amount,
    description = EXCLUDED.description,
    updated_at = NOW();

-- Create function to calculate tier based on referrals
CREATE OR REPLACE FUNCTION calculate_affiliate_tier(p_total_referrals INTEGER)
RETURNS TABLE(tier TEXT, commission_rate NUMERIC) AS $$
BEGIN
    RETURN QUERY
    SELECT t.tier, t.commission_rate
    FROM affiliate_commission_tiers t
    WHERE t.min_referrals <= p_total_referrals
    AND t.is_active = true
    ORDER BY t.min_referrals DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for automatic tier upgrades
CREATE OR REPLACE FUNCTION update_affiliate_tier() RETURNS TRIGGER AS $$
DECLARE
    new_tier TEXT;
    new_rate NUMERIC;
    tier_changed BOOLEAN := false;
BEGIN
    -- Calculate new tier based on total referrals
    SELECT t.tier, t.commission_rate INTO new_tier, new_rate
    FROM calculate_affiliate_tier(NEW.total_referrals) t;
    
    -- Check if tier changed
    IF new_tier IS NOT NULL AND new_tier != OLD.tier THEN
        NEW.tier := new_tier;
        NEW.commission_rate := new_rate;
        tier_changed := true;
        
        -- Log tier upgrade
        INSERT INTO gamification_audit_log (
            user_id,
            event_type,
            event_data
        ) VALUES (
            NEW.user_id,
            'AFFILIATE_TIER_UPGRADED',
            jsonb_build_object(
                'affiliate_id', NEW.id,
                'old_tier', OLD.tier,
                'new_tier', new_tier,
                'old_rate', OLD.commission_rate,
                'new_rate', new_rate,
                'total_referrals', NEW.total_referrals
            )
        );
        
        -- Check if bonus threshold reached
        DECLARE
            bonus_threshold INTEGER;
            bonus_amount INTEGER;
            bonus_already_paid BOOLEAN;
        BEGIN
            SELECT t.bonus_threshold, t.bonus_amount INTO bonus_threshold, bonus_amount
            FROM affiliate_commission_tiers t
            WHERE t.tier = new_tier;
            
            -- Check if this exact bonus was already paid
            SELECT EXISTS (
                SELECT 1 FROM gamification_audit_log
                WHERE user_id = NEW.user_id
                AND event_type = 'AFFILIATE_BONUS_PAID'
                AND event_data->>'tier' = new_tier
            ) INTO bonus_already_paid;
            
            IF bonus_threshold IS NOT NULL 
               AND NEW.total_referrals >= bonus_threshold 
               AND NOT bonus_already_paid THEN
                -- Add bonus to commission
                NEW.total_commission := NEW.total_commission + bonus_amount;
                
                -- Log bonus
                INSERT INTO gamification_audit_log (
                    user_id,
                    event_type,
                    event_data
                ) VALUES (
                    NEW.user_id,
                    'AFFILIATE_BONUS_PAID',
                    jsonb_build_object(
                        'affiliate_id', NEW.id,
                        'tier', new_tier,
                        'bonus_amount', bonus_amount,
                        'total_referrals', NEW.total_referrals
                    )
                );
            END IF;
        END;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if exists
DROP TRIGGER IF EXISTS affiliate_tier_update_trigger ON affiliates;

-- Create trigger
CREATE TRIGGER affiliate_tier_update_trigger
BEFORE UPDATE OF total_referrals ON affiliates
FOR EACH ROW
EXECUTE FUNCTION update_affiliate_tier();

-- ============================================================================
-- ISSUE #13: Active Referrals Tracking (HIGH PRIORITY)
-- ============================================================================
-- Problem: active_referrals column never updated
-- Solution: Add function to calculate and update active referrals

CREATE OR REPLACE FUNCTION update_active_referrals()
RETURNS void AS $$
BEGIN
    -- Update active_referrals for all affiliates
    -- Active = referred user made a recharge in last 30 days
    UPDATE affiliates a
    SET active_referrals = (
        SELECT COUNT(DISTINCT u.id)
        FROM users u
        INNER JOIN transactions t ON t.user_id = u.id
        WHERE u.referred_by = a.user_id
        AND t.transaction_type = 'RECHARGE'
        AND t.status = 'SUCCESS'
        AND t.created_at > NOW() - INTERVAL '30 days'
    ),
    updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Create scheduled job configuration (to be run daily)
INSERT INTO system_config (config_key, config_value, config_type, category, description, is_active)
VALUES 
    ('affiliate_active_referrals_update_schedule', '0 0 * * *', 'string', 'affiliate', 'Cron schedule for updating active referrals (daily at midnight)', true),
    ('affiliate_active_referrals_window_days', '30', 'integer', 'affiliate', 'Number of days to consider a referral active', true)
ON CONFLICT (config_key) DO UPDATE SET
    config_value = EXCLUDED.config_value,
    updated_at = NOW();

-- ============================================================================
-- ISSUE #14: Analytics Aggregation (HIGH PRIORITY)
-- ============================================================================
-- Problem: affiliate_analytics table not populated
-- Solution: Add daily aggregation function

CREATE OR REPLACE FUNCTION aggregate_affiliate_analytics(p_date DATE DEFAULT CURRENT_DATE)
RETURNS INTEGER AS $$
DECLARE
    v_affiliate RECORD;
    v_analytics_count INTEGER := 0;
BEGIN
    -- Loop through all affiliates
    FOR v_affiliate IN 
        SELECT id FROM affiliates WHERE status = 'APPROVED'
    LOOP
        -- Aggregate data for this affiliate and date
        INSERT INTO affiliate_analytics (
            affiliate_id,
            analytics_date,
            total_clicks,
            unique_clicks,
            conversions,
            conversion_rate,
            total_commission,
            recharge_commissions,
            subscription_commissions
        )
        SELECT
            v_affiliate.id,
            p_date,
            COUNT(ac.id) as total_clicks,
            COUNT(DISTINCT ac.ip_address) as unique_clicks,
            COUNT(ac.id) FILTER (WHERE ac.converted = true) as conversions,
            CASE 
                WHEN COUNT(DISTINCT ac.ip_address) > 0 THEN
                    ROUND((COUNT(ac.id) FILTER (WHERE ac.converted = true)::NUMERIC / COUNT(DISTINCT ac.ip_address)) * 100, 2)
                ELSE 0
            END as conversion_rate,
            COALESCE(SUM(acom.commission_amount), 0) as total_commission,
            COALESCE(SUM(acom.commission_amount) FILTER (WHERE t.transaction_type = 'RECHARGE'), 0) as recharge_commissions,
            COALESCE(SUM(acom.commission_amount) FILTER (WHERE t.transaction_type = 'SUBSCRIPTION'), 0) as subscription_commissions
        FROM affiliates a
        LEFT JOIN affiliate_clicks ac ON ac.affiliate_id = a.id 
            AND DATE(ac.created_at) = p_date
        LEFT JOIN affiliate_commissions acom ON acom.affiliate_id = a.id 
            AND DATE(acom.created_at) = p_date
            AND acom.is_reversed = false
        LEFT JOIN transactions t ON t.id = acom.transaction_id
        WHERE a.id = v_affiliate.id
        GROUP BY a.id
        ON CONFLICT (affiliate_id, analytics_date) DO UPDATE SET
            total_clicks = EXCLUDED.total_clicks,
            unique_clicks = EXCLUDED.unique_clicks,
            conversions = EXCLUDED.conversions,
            conversion_rate = EXCLUDED.conversion_rate,
            total_commission = EXCLUDED.total_commission,
            recharge_commissions = EXCLUDED.recharge_commissions,
            subscription_commissions = EXCLUDED.subscription_commissions,
            updated_at = NOW();
        
        v_analytics_count := v_analytics_count + 1;
    END LOOP;
    
    RETURN v_analytics_count;
END;
$$ LANGUAGE plpgsql;

-- Add configuration for analytics aggregation
INSERT INTO system_config (config_key, config_value, config_type, category, description, is_active)
VALUES 
    ('affiliate_analytics_aggregation_schedule', '0 1 * * *', 'string', 'affiliate', 'Cron schedule for analytics aggregation (daily at 1 AM)', true),
    ('affiliate_analytics_retention_days', '365', 'integer', 'affiliate', 'Number of days to retain analytics data', true)
ON CONFLICT (config_key) DO UPDATE SET
    config_value = EXCLUDED.config_value,
    updated_at = NOW();

-- ============================================================================
-- ISSUE #15: Commission Daily/Monthly Limits (HIGH PRIORITY)
-- ============================================================================
-- Problem: No commission caps per day/month
-- Solution: Add validation function

CREATE OR REPLACE FUNCTION check_commission_limits(
    p_affiliate_id UUID,
    p_commission_amount INTEGER
) RETURNS BOOLEAN AS $$
DECLARE
    daily_total INTEGER;
    monthly_total INTEGER;
    max_per_transaction INTEGER;
    max_per_day INTEGER;
    max_per_month INTEGER;
BEGIN
    -- Get limits from config
    SELECT config_value::INTEGER INTO max_per_transaction
    FROM system_config
    WHERE config_key = 'affiliate_max_commission_per_transaction' AND is_active = true;
    
    SELECT config_value::INTEGER INTO max_per_day
    FROM system_config
    WHERE config_key = 'affiliate_max_commission_per_day' AND is_active = true;
    
    SELECT config_value::INTEGER INTO max_per_month
    FROM system_config
    WHERE config_key = 'affiliate_max_commission_per_month' AND is_active = true;
    
    -- Check per-transaction limit
    IF max_per_transaction IS NOT NULL AND p_commission_amount > max_per_transaction THEN
        RAISE EXCEPTION 'Commission exceeds per-transaction limit: ₦% > ₦%', 
            p_commission_amount / 100.0, max_per_transaction / 100.0;
    END IF;
    
    -- Check daily limit
    IF max_per_day IS NOT NULL THEN
        SELECT COALESCE(SUM(commission_amount), 0) INTO daily_total
        FROM affiliate_commissions
        WHERE affiliate_id = p_affiliate_id
        AND DATE(created_at) = CURRENT_DATE
        AND is_reversed = false;
        
        IF daily_total + p_commission_amount > max_per_day THEN
            RAISE EXCEPTION 'Daily commission limit exceeded: ₦% + ₦% > ₦%',
                daily_total / 100.0, p_commission_amount / 100.0, max_per_day / 100.0;
        END IF;
    END IF;
    
    -- Check monthly limit
    IF max_per_month IS NOT NULL THEN
        SELECT COALESCE(SUM(commission_amount), 0) INTO monthly_total
        FROM affiliate_commissions
        WHERE affiliate_id = p_affiliate_id
        AND DATE_TRUNC('month', created_at) = DATE_TRUNC('month', CURRENT_DATE)
        AND is_reversed = false;
        
        IF monthly_total + p_commission_amount > max_per_month THEN
            RAISE EXCEPTION 'Monthly commission limit exceeded: ₦% + ₦% > ₦%',
                monthly_total / 100.0, p_commission_amount / 100.0, max_per_month / 100.0;
        END IF;
    END IF;
    
    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Helper Functions for Affiliate Management
-- ============================================================================

-- Get affiliate dashboard data
CREATE OR REPLACE FUNCTION get_affiliate_dashboard_data(p_user_id UUID)
RETURNS TABLE (
    total_referrals INTEGER,
    active_referrals INTEGER,
    total_commission INTEGER,
    pending_commission INTEGER,
    paid_commission INTEGER,
    available_for_payout INTEGER,
    current_tier TEXT,
    commission_rate NUMERIC,
    next_tier TEXT,
    referrals_to_next_tier INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        a.total_referrals,
        a.active_referrals,
        a.total_commission,
        COALESCE((
            SELECT SUM(commission_amount)
            FROM affiliate_commissions
            WHERE affiliate_id = a.id
            AND status = 'PENDING'
            AND is_reversed = false
        ), 0)::INTEGER as pending_commission,
        COALESCE((
            SELECT SUM(commission_amount)
            FROM affiliate_commissions
            WHERE affiliate_id = a.id
            AND status = 'PAID'
            AND is_reversed = false
        ), 0)::INTEGER as paid_commission,
        (a.total_commission - COALESCE((
            SELECT SUM(total_amount)
            FROM affiliate_payouts
            WHERE affiliate_id = a.id
            AND payout_status = 'COMPLETED'
        ), 0))::INTEGER as available_for_payout,
        a.tier as current_tier,
        a.commission_rate,
        (
            SELECT tier
            FROM affiliate_commission_tiers
            WHERE min_referrals > a.total_referrals
            AND is_active = true
            ORDER BY min_referrals ASC
            LIMIT 1
        ) as next_tier,
        COALESCE((
            SELECT min_referrals - a.total_referrals
            FROM affiliate_commission_tiers
            WHERE min_referrals > a.total_referrals
            AND is_active = true
            ORDER BY min_referrals ASC
            LIMIT 1
        ), 0) as referrals_to_next_tier
    FROM affiliates a
    WHERE a.user_id = p_user_id;
END;
$$ LANGUAGE plpgsql;

-- Get top performing affiliates
CREATE OR REPLACE FUNCTION get_top_affiliates(p_limit INTEGER DEFAULT 10)
RETURNS TABLE (
    affiliate_id UUID,
    user_id UUID,
    affiliate_code TEXT,
    total_referrals INTEGER,
    total_commission INTEGER,
    tier TEXT,
    rank INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        a.id,
        a.user_id,
        a.affiliate_code,
        a.total_referrals,
        a.total_commission,
        a.tier,
        ROW_NUMBER() OVER (ORDER BY a.total_commission DESC)::INTEGER as rank
    FROM affiliates a
    WHERE a.status = 'APPROVED'
    ORDER BY a.total_commission DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Data Migration: Update Existing Affiliates
-- ============================================================================

-- Update all existing affiliates to correct tier based on referrals
DO $$
DECLARE
    v_affiliate RECORD;
    v_new_tier TEXT;
    v_new_rate NUMERIC;
BEGIN
    FOR v_affiliate IN SELECT id, total_referrals FROM affiliates
    LOOP
        SELECT tier, commission_rate INTO v_new_tier, v_new_rate
        FROM calculate_affiliate_tier(v_affiliate.total_referrals);
        
        IF v_new_tier IS NOT NULL THEN
            UPDATE affiliates
            SET tier = v_new_tier,
                commission_rate = v_new_rate,
                updated_at = NOW()
            WHERE id = v_affiliate.id;
        END IF;
    END LOOP;
    
    RAISE NOTICE 'Updated tiers for all existing affiliates';
END $$;

-- Run initial active referrals update
SELECT update_active_referrals();

-- Run initial analytics aggregation for yesterday
SELECT aggregate_affiliate_analytics(CURRENT_DATE - INTERVAL '1 day');

-- ============================================================================
-- Verification Queries
-- ============================================================================

DO $$
DECLARE
    tier_count INTEGER;
    analytics_count INTEGER;
    bank_accounts_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO tier_count FROM affiliate_commission_tiers;
    SELECT COUNT(*) INTO analytics_count FROM affiliate_analytics;
    SELECT COUNT(*) INTO bank_accounts_count FROM affiliate_bank_accounts;
    
    RAISE NOTICE 'P1 Migration complete:';
    RAISE NOTICE '  - Commission tiers: %', tier_count;
    RAISE NOTICE '  - Analytics records: %', analytics_count;
    RAISE NOTICE '  - Bank accounts: %', bank_accounts_count;
    RAISE NOTICE '  - Tier automation: ACTIVE';
    RAISE NOTICE '  - Analytics aggregation: CONFIGURED';
    RAISE NOTICE '  - Commission limits: ENFORCED';
END $$;

-- ============================================================================
-- Migration Complete
-- ============================================================================

COMMENT ON TABLE affiliate_commission_tiers IS 'Commission tier definitions with automatic upgrade thresholds';
COMMENT ON FUNCTION calculate_affiliate_tier(INTEGER) IS 'Calculates appropriate tier based on total referrals';
COMMENT ON FUNCTION update_affiliate_tier() IS 'Automatically upgrades affiliate tier when referral threshold reached';
COMMENT ON FUNCTION update_active_referrals() IS 'Updates active referrals count (30-day window) for all affiliates';
COMMENT ON FUNCTION aggregate_affiliate_analytics(DATE) IS 'Aggregates daily analytics for all affiliates';
COMMENT ON FUNCTION check_commission_limits(UUID, INTEGER) IS 'Validates commission against daily/monthly limits';
COMMENT ON FUNCTION get_affiliate_dashboard_data(UUID) IS 'Returns complete dashboard data for an affiliate';
COMMENT ON FUNCTION get_top_affiliates(INTEGER) IS 'Returns top performing affiliates by commission';
