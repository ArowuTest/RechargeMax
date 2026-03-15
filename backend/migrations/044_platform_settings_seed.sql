-- Migration 044: Seed all required platform_settings keys
-- Ensures affiliate_commission_rate and all business-critical settings are present.

INSERT INTO platform_settings (setting_key, setting_value, description, is_public) VALUES
    -- Affiliate
    ('affiliate.commission_rate',     '1.0',   'Affiliate commission rate as percentage (e.g. 1.0 = 1%)', false),
    ('affiliate.min_payout_amount',   '5000',  'Minimum affiliate payout in kobo (5000 = ₦50)', false),
    ('affiliate.auto_release_days',   '7',     'Days after transaction before commission is auto-released', false),
    
    -- Points
    ('points.naira_per_point',        '200',   '₦200 spend = 1 point (amount in kobo = 20000)', true),
    ('points.min_recharge_kobo',      '5000',  'Minimum recharge in kobo to earn points (₦50)', true),
    ('points.draw_entries_per_point', '1',     'Number of draw entries awarded per point earned', true),
    
    -- Draw / claim window
    ('draw.claim_window_days',        '30',    'Number of days a winner has to claim a prize', true),
    ('draw.max_entries_per_msisdn',   '1000',  'Maximum draw entries per MSISDN per draw', false),
    
    -- Spin wheel
    ('spin.min_recharge_kobo',        '100000','Minimum recharge in kobo to earn a spin (₦1000)', true),
    ('spin.daily_spin_limit',         '3',     'Maximum spins per user per day', true),
    
    -- Daily subscription
    ('subscription.price_kobo',       '2000',  'Daily subscription price in kobo (₦20)', true),
    ('subscription.draw_entries',     '1',     'Draw entries per daily subscription', true),
    ('subscription.points_per_sub',   '0',     'Bonus points per daily subscription', true),
    
    -- USSD
    ('ussd.draw_entries_per_200_naira', '1',   'Draw entries per ₦200 recharged via USSD', true),
    ('ussd.points_per_200_naira',     '1',     'Points per ₦200 recharged via USSD (20000 kobo)', true),
    
    -- Loyalty tiers
    ('loyalty.bronze_min_points',     '0',     'Minimum points for Bronze tier', true),
    ('loyalty.silver_min_points',     '500',   'Minimum points for Silver tier', true),
    ('loyalty.gold_min_points',       '2000',  'Minimum points for Gold tier', true),
    ('loyalty.platinum_min_points',   '5000',  'Minimum points for Platinum tier', true),
    ('loyalty.silver_multiplier',     '1.25',  'Draw entry multiplier for Silver tier', true),
    ('loyalty.gold_multiplier',       '1.5',   'Draw entry multiplier for Gold tier', true),
    ('loyalty.platinum_multiplier',   '2.0',   'Draw entry multiplier for Platinum tier', true),
    
    -- Network detection
    ('network.hlr_enabled',           'true',  'Use HLR API for network detection (falls back to prefix)', false),
    ('network.hlr_timeout_seconds',   '5',     'HLR API timeout in seconds', false)
ON CONFLICT (setting_key) DO UPDATE 
    SET description = EXCLUDED.description,
        updated_at  = now()
    -- Only update description, not value (preserve admin overrides)
;
