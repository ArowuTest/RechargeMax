-- Comprehensive initial data for all tables
-- Created: 2026-01-30 14:00 UTC
-- ============================================================================

-- ============================================================================
-- NETWORK CONFIGURATIONS
-- ============================================================================

INSERT INTO public.network_configs (
    id, network_name, network_code, is_active, airtime_enabled, data_enabled, 
    commission_rate, minimum_amount, maximum_amount, sort_order
) VALUES 
    (
        '11111111-1111-1111-1111-111111111111',
        'MTN Nigeria',
        'MTN',
        true,
        true,
        true,
        2.50,
        50.00,
        50000.00,
        1
    ),
    (
        '22222222-2222-2222-2222-222222222222',
        'Airtel Nigeria',
        'AIRTEL',
        true,
        true,
        true,
        2.75,
        50.00,
        50000.00,
        2
    ),
    (
        '33333333-3333-3333-3333-333333333333',
        'Globacom Limited',
        'GLO',
        true,
        true,
        true,
        3.00,
        50.00,
        50000.00,
        3
    ),
    (
        '44444444-4444-4444-4444-444444444444',
        '9mobile Nigeria',
        'NINE_MOBILE',
        true,
        true,
        true,
        3.25,
        50.00,
        50000.00,
        4
    )
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- DATA PLANS
-- ============================================================================

-- MTN Data Plans
INSERT INTO public.data_plans (
    id, network_id, plan_name, data_amount, price, validity_days, plan_code, is_active, sort_order
) VALUES 
    (uuid_generate_v4(), '11111111-1111-1111-1111-111111111111', 'MTN 100MB Daily', '100MB', 100.00, 1, 'MTN_100MB_1D', true, 1),
    (uuid_generate_v4(), '11111111-1111-1111-1111-111111111111', 'MTN 200MB Daily', '200MB', 200.00, 1, 'MTN_200MB_1D', true, 2),
    (uuid_generate_v4(), '11111111-1111-1111-1111-111111111111', 'MTN 500MB Weekly', '500MB', 500.00, 7, 'MTN_500MB_7D', true, 3),
    (uuid_generate_v4(), '11111111-1111-1111-1111-111111111111', 'MTN 1GB Monthly', '1GB', 1000.00, 30, 'MTN_1GB_30D', true, 4),
    (uuid_generate_v4(), '11111111-1111-1111-1111-111111111111', 'MTN 2GB Monthly', '2GB', 2000.00, 30, 'MTN_2GB_30D', true, 5),
    (uuid_generate_v4(), '11111111-1111-1111-1111-111111111111', 'MTN 3GB Monthly', '3GB', 3000.00, 30, 'MTN_3GB_30D', true, 6),
    (uuid_generate_v4(), '11111111-1111-1111-1111-111111111111', 'MTN 5GB Monthly', '5GB', 5000.00, 30, 'MTN_5GB_30D', true, 7),
    (uuid_generate_v4(), '11111111-1111-1111-1111-111111111111', 'MTN 10GB Monthly', '10GB', 10000.00, 30, 'MTN_10GB_30D', true, 8);

-- Airtel Data Plans
INSERT INTO public.data_plans (
    id, network_id, plan_name, data_amount, price, validity_days, plan_code, is_active, sort_order
) VALUES 
    (uuid_generate_v4(), '22222222-2222-2222-2222-222222222222', 'Airtel 100MB Daily', '100MB', 100.00, 1, 'AIRTEL_100MB_1D', true, 1),
    (uuid_generate_v4(), '22222222-2222-2222-2222-222222222222', 'Airtel 300MB Daily', '300MB', 300.00, 1, 'AIRTEL_300MB_1D', true, 2),
    (uuid_generate_v4(), '22222222-2222-2222-2222-222222222222', 'Airtel 500MB Weekly', '500MB', 500.00, 7, 'AIRTEL_500MB_7D', true, 3),
    (uuid_generate_v4(), '22222222-2222-2222-2222-222222222222', 'Airtel 1GB Monthly', '1GB', 1000.00, 30, 'AIRTEL_1GB_30D', true, 4),
    (uuid_generate_v4(), '22222222-2222-2222-2222-222222222222', 'Airtel 2GB Monthly', '2GB', 2000.00, 30, 'AIRTEL_2GB_30D', true, 5),
    (uuid_generate_v4(), '22222222-2222-2222-2222-222222222222', 'Airtel 4GB Monthly', '4GB', 4000.00, 30, 'AIRTEL_4GB_30D', true, 6),
    (uuid_generate_v4(), '22222222-2222-2222-2222-222222222222', 'Airtel 6GB Monthly', '6GB', 6000.00, 30, 'AIRTEL_6GB_30D', true, 7),
    (uuid_generate_v4(), '22222222-2222-2222-2222-222222222222', 'Airtel 11GB Monthly', '11GB', 11000.00, 30, 'AIRTEL_11GB_30D', true, 8);

-- Glo Data Plans
INSERT INTO public.data_plans (
    id, network_id, plan_name, data_amount, price, validity_days, plan_code, is_active, sort_order
) VALUES 
    (uuid_generate_v4(), '33333333-3333-3333-3333-333333333333', 'Glo 200MB Daily', '200MB', 200.00, 1, 'GLO_200MB_1D', true, 1),
    (uuid_generate_v4(), '33333333-3333-3333-3333-333333333333', 'Glo 500MB Weekly', '500MB', 500.00, 7, 'GLO_500MB_7D', true, 2),
    (uuid_generate_v4(), '33333333-3333-3333-3333-333333333333', 'Glo 1.35GB Monthly', '1.35GB', 1000.00, 30, 'GLO_1350MB_30D', true, 3),
    (uuid_generate_v4(), '33333333-3333-3333-3333-333333333333', 'Glo 2.9GB Monthly', '2.9GB', 2000.00, 30, 'GLO_2900MB_30D', true, 4),
    (uuid_generate_v4(), '33333333-3333-3333-3333-333333333333', 'Glo 5.8GB Monthly', '5.8GB', 4000.00, 30, 'GLO_5800MB_30D', true, 5),
    (uuid_generate_v4(), '33333333-3333-3333-3333-333333333333', 'Glo 7.7GB Monthly', '7.7GB', 5000.00, 30, 'GLO_7700MB_30D', true, 6),
    (uuid_generate_v4(), '33333333-3333-3333-3333-333333333333', 'Glo 10GB Monthly', '10GB', 8000.00, 30, 'GLO_10GB_30D', true, 7),
    (uuid_generate_v4(), '33333333-3333-3333-3333-333333333333', 'Glo 13.25GB Monthly', '13.25GB', 10000.00, 30, 'GLO_13250MB_30D', true, 8);

-- 9mobile Data Plans
INSERT INTO public.data_plans (
    id, network_id, plan_name, data_amount, price, validity_days, plan_code, is_active, sort_order
) VALUES 
    (uuid_generate_v4(), '44444444-4444-4444-4444-444444444444', '9mobile 150MB Daily', '150MB', 150.00, 1, '9MOBILE_150MB_1D', true, 1),
    (uuid_generate_v4(), '44444444-4444-4444-4444-444444444444', '9mobile 650MB Weekly', '650MB', 500.00, 7, '9MOBILE_650MB_7D', true, 2),
    (uuid_generate_v4(), '44444444-4444-4444-4444-444444444444', '9mobile 1.5GB Monthly', '1.5GB', 1000.00, 30, '9MOBILE_1500MB_30D', true, 3),
    (uuid_generate_v4(), '44444444-4444-4444-4444-444444444444', '9mobile 2GB Monthly', '2GB', 2000.00, 30, '9MOBILE_2GB_30D', true, 4),
    (uuid_generate_v4(), '44444444-4444-4444-4444-444444444444', '9mobile 4.5GB Monthly', '4.5GB', 4000.00, 30, '9MOBILE_4500MB_30D', true, 5),
    (uuid_generate_v4(), '44444444-4444-4444-4444-444444444444', '9mobile 11GB Monthly', '11GB', 8000.00, 30, '9MOBILE_11GB_30D', true, 6),
    (uuid_generate_v4(), '44444444-4444-4444-4444-444444444444', '9mobile 15GB Monthly', '15GB', 10000.00, 30, '9MOBILE_15GB_30D', true, 7),
    (uuid_generate_v4(), '44444444-4444-4444-4444-444444444444', '9mobile 27.5GB Monthly', '27.5GB', 15000.00, 30, '9MOBILE_27500MB_30D', true, 8);

-- ============================================================================
-- WHEEL PRIZES CONFIGURATION
-- ============================================================================

INSERT INTO public.wheel_prizes (
    id, prize_name, prize_type, prize_value, probability, minimum_recharge, 
    is_active, icon_name, color_scheme, sort_order
) VALUES 
    (uuid_generate_v4(), '₦50 Cash Prize', 'CASH', 50.00, 25.0, 1000.00, true, 'dollar-sign', 'green', 1),
    (uuid_generate_v4(), '₦100 Cash Prize', 'CASH', 100.00, 20.0, 1000.00, true, 'dollar-sign', 'green', 2),
    (uuid_generate_v4(), '₦200 Cash Prize', 'CASH', 200.00, 15.0, 2000.00, true, 'dollar-sign', 'blue', 3),
    (uuid_generate_v4(), '₦500 Cash Prize', 'CASH', 500.00, 10.0, 5000.00, true, 'dollar-sign', 'purple', 4),
    (uuid_generate_v4(), '₦1000 Cash Prize', 'CASH', 1000.00, 5.0, 10000.00, true, 'dollar-sign', 'yellow', 5),
    (uuid_generate_v4(), '₦100 Airtime', 'AIRTIME', 100.00, 15.0, 1000.00, true, 'phone', 'blue', 6),
    (uuid_generate_v4(), '₦200 Airtime', 'AIRTIME', 200.00, 8.0, 2000.00, true, 'phone', 'indigo', 7),
    (uuid_generate_v4(), '500MB Data', 'DATA', 500.00, 1.5, 1000.00, true, 'wifi', 'orange', 8),
    (uuid_generate_v4(), '1GB Data', 'DATA', 1000.00, 0.5, 5000.00, true, 'wifi', 'red', 9);

-- ============================================================================
-- ADMIN USERS
-- ============================================================================

-- Create super admin user (password: SuperAdmin123!)
INSERT INTO public.admin_users (
    id, email, password_hash, full_name, role, permissions, is_active
) VALUES (
    uuid_generate_v4(),
    'admin@rechargemax.ng',
    crypt('SuperAdmin123!', gen_salt('bf')),
    'Super Administrator',
    'SUPER_ADMIN',
    '["view_analytics", "manage_users", "manage_transactions", "manage_networks", "manage_prizes", "manage_affiliates", "manage_settings", "manage_admins", "view_monitoring", "manage_draws"]'::jsonb,
    true
);

-- Create regular admin user (password: Admin123!)
INSERT INTO public.admin_users (
    id, email, password_hash, full_name, role, permissions, is_active
) VALUES (
    uuid_generate_v4(),
    'support@rechargemax.ng',
    crypt('Admin123!', gen_salt('bf')),
    'Support Administrator',
    'ADMIN',
    '["view_analytics", "manage_users", "manage_transactions", "manage_affiliates"]'::jsonb,
    true
);

-- ============================================================================
-- DAILY SUBSCRIPTION CONFIGURATION
-- ============================================================================

INSERT INTO public.daily_subscription_config (
    id, amount, draw_entries_earned, is_paid, description
) VALUES (
    uuid_generate_v4(),
    30.00,
    1,
    true,
    'Daily subscription for guaranteed draw entries - ₦30 per day for 1 draw entry'
);

-- ============================================================================
-- PLATFORM SETTINGS
-- ============================================================================

INSERT INTO public.platform_settings (
    setting_key, setting_value, description, is_public
) VALUES 
    ('platform_name', '"RechargeMax"', 'Platform name displayed to users', true),
    ('platform_tagline', '"Your Ultimate Mobile Recharge Platform"', 'Platform tagline', true),
    ('support_email', '"support@rechargemax.ng"', 'Support email address', true),
    ('support_phone', '"+234-800-RECHARGE"', 'Support phone number', true),
    ('minimum_recharge_amount', '50', 'Minimum recharge amount in Naira', true),
    ('maximum_recharge_amount', '50000', 'Maximum recharge amount in Naira', true),
    ('spin_wheel_minimum', '1000', 'Minimum recharge amount to unlock spin wheel', true),
    ('points_per_naira', '1', 'Points earned per Naira spent', true),
    ('draw_entries_per_200_points', '1', 'Draw entries earned per 200 points', true),
    ('maintenance_mode', 'false', 'Enable/disable maintenance mode', false),
    ('registration_enabled', 'true', 'Enable/disable user registration', false),
    ('guest_recharge_enabled', 'true', 'Enable/disable guest recharges', false),
    ('affiliate_program_enabled', 'true', 'Enable/disable affiliate program', false),
    ('daily_subscription_enabled', 'true', 'Enable/disable daily subscriptions', false),
    ('spin_wheel_enabled', 'true', 'Enable/disable spin wheel feature', false),
    ('draw_system_enabled', 'true', 'Enable/disable draw system', false),
    ('max_daily_transactions_per_user', '10', 'Maximum daily transactions per user', false),
    ('max_daily_amount_per_user', '100000', 'Maximum daily amount per user in Naira', false),
    ('commission_payout_threshold', '5000', 'Minimum commission amount for payout', false),
    ('prize_claim_expiry_days', '30', 'Days before unclaimed prizes expire', false);

-- ============================================================================
-- SAMPLE DRAWS
-- ============================================================================

-- Create a sample daily draw
INSERT INTO public.draws (
    id, name, type, status, start_time, end_time, prize_pool, 
    total_entries, winners_count, entry_cost
) VALUES (
    uuid_generate_v4(),
    'Daily Cash Draw - ' || TO_CHAR(CURRENT_DATE, 'DD Mon YYYY'),
    'DAILY',
    'ACTIVE',
    CURRENT_DATE + INTERVAL '00:00:00',
    CURRENT_DATE + INTERVAL '23:59:59',
    50000.00,
    0,
    5,
    0.00
);

-- Create a sample weekly draw
INSERT INTO public.draws (
    id, name, type, status, start_time, end_time, prize_pool, 
    total_entries, winners_count, entry_cost
) VALUES (
    uuid_generate_v4(),
    'Weekly Mega Draw - Week ' || TO_CHAR(CURRENT_DATE, 'WW/YYYY'),
    'WEEKLY',
    'UPCOMING',
    DATE_TRUNC('week', CURRENT_DATE) + INTERVAL '7 days',
    DATE_TRUNC('week', CURRENT_DATE) + INTERVAL '13 days 23:59:59',
    500000.00,
    0,
    10,
    100.00
);

-- ============================================================================
-- SAMPLE USERS (for testing)
-- ============================================================================

-- Note: These are sample users for testing. In production, users are created through registration
INSERT INTO public.users (
    id, msisdn, full_name, email, phone_verified, loyalty_tier, 
    total_points, referral_code, is_active
) VALUES 
    (
        uuid_generate_v4(),
        '2348012345678',
        'John Doe',
        'john.doe@example.com',
        true,
        'BRONZE',
        1500,
        'JOHN2026',
        true
    ),
    (
        uuid_generate_v4(),
        '2348087654321',
        'Jane Smith',
        'jane.smith@example.com',
        true,
        'SILVER',
        5000,
        'JANE2026',
        true
    ),
    (
        uuid_generate_v4(),
        '2347012345678',
        'Mike Johnson',
        'mike.johnson@example.com',
        true,
        'GOLD',
        12000,
        'MIKE2026',
        true
    );

-- ============================================================================
-- SAMPLE TRANSACTIONS (for testing)
-- ============================================================================

-- Sample successful transactions
INSERT INTO public.transactions (
    id, user_id, msisdn, network_provider, recharge_type, amount, 
    payment_method, payment_reference, status, points_earned, draw_entries,
    customer_email, customer_name, created_at, completed_at
) VALUES 
    (
        uuid_generate_v4(),
        (SELECT id FROM public.users WHERE msisdn = '2348012345678'),
        '2348012345678',
        'MTN',
        'AIRTIME',
        1000.00,
        'CARD',
        'TXN_' || EXTRACT(EPOCH FROM NOW())::TEXT,
        'SUCCESS',
        1000,
        5,
        'john.doe@example.com',
        'John Doe',
        NOW() - INTERVAL '2 hours',
        NOW() - INTERVAL '2 hours' + INTERVAL '30 seconds'
    ),
    (
        uuid_generate_v4(),
        (SELECT id FROM public.users WHERE msisdn = '2348087654321'),
        '2348087654321',
        'AIRTEL',
        'DATA',
        2000.00,
        'CARD',
        'TXN_' || (EXTRACT(EPOCH FROM NOW()) + 1)::TEXT,
        'SUCCESS',
        2000,
        10,
        'jane.smith@example.com',
        'Jane Smith',
        NOW() - INTERVAL '1 hour',
        NOW() - INTERVAL '1 hour' + INTERVAL '45 seconds'
    );

-- ============================================================================
-- SAMPLE SPIN RESULTS (for testing)
-- ============================================================================

-- Sample spin results
INSERT INTO public.spin_results (
    id, user_id, transaction_id, msisdn, prize_id, prize_name, 
    prize_type, prize_value, claim_status, created_at
) VALUES 
    (
        uuid_generate_v4(),
        (SELECT id FROM public.users WHERE msisdn = '2348012345678'),
        (SELECT id FROM public.transactions WHERE msisdn = '2348012345678' LIMIT 1),
        '2348012345678',
        (SELECT id FROM public.wheel_prizes WHERE prize_name = '₦100 Cash Prize' LIMIT 1),
        '₦100 Cash Prize',
        'CASH',
        100.00,
        'PENDING',
        NOW() - INTERVAL '2 hours'
    );

-- ============================================================================
-- SAMPLE DAILY SUBSCRIPTIONS (for testing)
-- ============================================================================

-- Sample daily subscriptions
INSERT INTO public.daily_subscriptions (
    id, user_id, msisdn, subscription_date, amount, draw_entries_earned, 
    points_earned, payment_reference, status, is_paid, customer_email, customer_name
) VALUES 
    (
        uuid_generate_v4(),
        (SELECT id FROM public.users WHERE msisdn = '2348087654321'),
        '2348087654321',
        CURRENT_DATE,
        30.00,
        1,
        30,
        'SUB_' || EXTRACT(EPOCH FROM NOW())::TEXT,
        'active',
        true,
        'jane.smith@example.com',
        'Jane Smith'
    );

-- ============================================================================
-- SAMPLE AFFILIATES (for testing)
-- ============================================================================

-- Sample affiliate
INSERT INTO public.affiliates (
    id, user_id, affiliate_code, status, tier, commission_rate, 
    total_referrals, active_referrals, total_commission, approved_at
) VALUES 
    (
        uuid_generate_v4(),
        (SELECT id FROM public.users WHERE msisdn = '2347012345678'),
        'MIKE2026AFF',
        'APPROVED',
        'SILVER',
        7.50,
        15,
        12,
        2500.00,
        NOW() - INTERVAL '30 days'
    );

-- ============================================================================
-- UPDATE SEQUENCES AND FINAL SETUP
-- ============================================================================

-- Update user statistics based on sample data
UPDATE public.users SET 
    total_recharge_amount = (
        SELECT COALESCE(SUM(amount), 0) 
        FROM public.transactions 
        WHERE user_id = users.id AND status = 'SUCCESS'
    ),
    total_transactions = (
        SELECT COUNT(*) 
        FROM public.transactions 
        WHERE user_id = users.id AND status = 'SUCCESS'
    ),
    last_recharge_date = (
        SELECT MAX(completed_at) 
        FROM public.transactions 
        WHERE user_id = users.id AND status = 'SUCCESS'
    );

-- Add sample draw entries for users with subscriptions
INSERT INTO public.draw_entries (
    id, draw_id, user_id, msisdn, entries_count, source_type, source_subscription_id
) 
SELECT 
    uuid_generate_v4(),
    (SELECT id FROM public.draws WHERE type = 'DAILY' AND status = 'ACTIVE' LIMIT 1),
    ds.user_id,
    ds.msisdn,
    ds.draw_entries_earned,
    'SUBSCRIPTION',
    ds.id
FROM public.daily_subscriptions ds
WHERE ds.status = 'active';

-- Update draw total entries
UPDATE public.draws SET 
    total_entries = (
        SELECT COALESCE(SUM(entries_count), 0) 
        FROM public.draw_entries 
        WHERE draw_id = draws.id
    );


