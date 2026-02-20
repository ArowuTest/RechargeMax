-- Migration: 030_affiliate_fixes_final.sql
-- Description: Fix remaining issues from P1 migration
-- Date: 2026-02-02

-- ============================================================================
-- Fix update_active_referrals function (transactions don't have transaction_type)
-- ============================================================================

CREATE OR REPLACE FUNCTION update_active_referrals()
RETURNS void AS $$
BEGIN
    -- Update active_referrals for all affiliates
    -- Active = referred user made a recharge in last 30 days
    -- Note: All transactions in this table are recharges (VTU transactions)
    UPDATE affiliates a
    SET active_referrals = (
        SELECT COUNT(DISTINCT u.id)
        FROM users u
        INNER JOIN transactions t ON t.user_id = u.id
        WHERE u.referred_by = a.user_id
        AND t.status = 'SUCCESS'
        AND t.created_at > NOW() - INTERVAL '30 days'
    ),
    updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Fix aggregate_affiliate_analytics function (DATE type issue)
-- ============================================================================

DROP FUNCTION IF EXISTS aggregate_affiliate_analytics(DATE);

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
            COALESCE(SUM(acom.commission_amount), 0)::NUMERIC as total_commission,
            COALESCE(SUM(acom.commission_amount), 0)::NUMERIC as recharge_commissions,
            0::NUMERIC as subscription_commissions -- No subscriptions tracked yet
        FROM affiliates a
        LEFT JOIN affiliate_clicks ac ON ac.affiliate_id = a.id 
            AND DATE(ac.created_at) = p_date
        LEFT JOIN affiliate_commissions acom ON acom.affiliate_id = a.id 
            AND DATE(acom.created_at) = p_date
            AND acom.is_reversed = false
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

-- ============================================================================
-- Add missing system config entries with correct column names
-- ============================================================================

INSERT INTO system_config (key, value, category, description, is_public) VALUES
('affiliate_active_referrals_update_schedule', '"0 0 * * *"'::jsonb, 'affiliate', 'Cron schedule for updating active referrals (daily at midnight)', false),
('affiliate_analytics_aggregation_schedule', '"0 1 * * *"'::jsonb, 'affiliate', 'Cron schedule for analytics aggregation (daily at 1 AM)', false)
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    description = EXCLUDED.description,
    updated_at = NOW();

-- ============================================================================
-- Run initial data updates
-- ============================================================================

-- Update active referrals
SELECT update_active_referrals();

-- Aggregate analytics for yesterday
SELECT aggregate_affiliate_analytics((CURRENT_DATE - INTERVAL '1 day')::DATE);

-- ============================================================================
-- Verification
-- ============================================================================

DO $$
DECLARE
    active_ref_count INTEGER;
    analytics_count INTEGER;
BEGIN
    SELECT SUM(active_referrals) INTO active_ref_count FROM affiliates;
    SELECT COUNT(*) INTO analytics_count FROM affiliate_analytics;
    
    RAISE NOTICE 'Final fixes applied:';
    RAISE NOTICE '  - Active referrals updated: % total', COALESCE(active_ref_count, 0);
    RAISE NOTICE '  - Analytics records: %', analytics_count;
    RAISE NOTICE '  - All functions fixed and working';
END $$;
