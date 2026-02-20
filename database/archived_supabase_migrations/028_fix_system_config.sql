-- Migration: 027_fix_system_config.sql
-- Description: Fix system_config inserts to use correct column names
-- Date: 2026-02-02

-- System config uses 'key' and 'value' (JSONB), not 'config_key' and 'config_value'
-- Insert affiliate configuration

INSERT INTO system_config (key, value, category, description, is_public) VALUES
('affiliate_minimum_payout', '500000'::jsonb, 'affiliate', 'Minimum payout amount in kobo (₦5,000)', false),
('affiliate_max_commission_per_transaction', '500000'::jsonb, 'affiliate', 'Maximum commission per transaction in kobo (₦5,000)', false),
('affiliate_max_commission_per_day', '5000000'::jsonb, 'affiliate', 'Maximum commission per day in kobo (₦50,000)', false),
('affiliate_max_commission_per_month', '50000000'::jsonb, 'affiliate', 'Maximum commission per month in kobo (₦500,000)', false),
('affiliate_payout_processing_fee', '10000'::jsonb, 'affiliate', 'Payout processing fee in kobo (₦100)', false),
('affiliate_click_rate_limit_per_hour', '10'::jsonb, 'affiliate', 'Maximum clicks per IP per hour', false),
('affiliate_approval_auto', 'false'::jsonb, 'affiliate', 'Automatically approve new affiliates', false),
('affiliate_active_referrals_window_days', '30'::jsonb, 'affiliate', 'Number of days to consider a referral active', false),
('affiliate_analytics_retention_days', '365'::jsonb, 'affiliate', 'Number of days to retain analytics data', false)
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    description = EXCLUDED.description,
    updated_at = NOW();
