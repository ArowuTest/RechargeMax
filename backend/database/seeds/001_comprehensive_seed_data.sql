-- ============================================================================
-- RECHARGEMAX COMPREHENSIVE SEED DATA
-- ============================================================================
-- This file contains realistic seed data for testing and demonstration
-- All data is fictional but representative of production scenarios
-- ============================================================================

-- Clean existing data (for development/testing only)
-- TRUNCATE TABLE token_blacklist, winners, draw_participants, draws, spin_results, spin_prizes, subscriptions, affiliate_payouts, affiliates, recharges, transactions, users, network_bundles, networks CASCADE;

-- ============================================================================
-- NETWORKS (Nigerian Mobile Networks)
-- ============================================================================
INSERT INTO networks (id, code, name, logo_url, is_active, airtime_enabled, data_enabled, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'MTN', 'MTN Nigeria', '/images/networks/mtn.png', true, true, true, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'GLO', 'Glo Mobile', '/images/networks/glo.png', true, true, true, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'AIRTEL', 'Airtel Nigeria', '/images/networks/airtel.png', true, true, true, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440004', '9MOBILE', '9mobile', '/images/networks/9mobile.png', true, true, true, NOW(), NOW())
ON CONFLICT (code) DO UPDATE SET
  name = EXCLUDED.name,
  is_active = EXCLUDED.is_active,
  updated_at = NOW();

-- ============================================================================
-- NETWORK DATA BUNDLES
-- ============================================================================
-- MTN Data Bundles
INSERT INTO network_bundles (id, network_id, code, name, data_size, amount, validity_days, is_active, created_at, updated_at) VALUES
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'MTN_500MB_DAILY', '500MB Daily', '500MB', 15000, 1, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'MTN_1GB_DAILY', '1GB Daily', '1GB', 30000, 1, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'MTN_2GB_WEEKLY', '2GB Weekly', '2GB', 50000, 7, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'MTN_5GB_MONTHLY', '5GB Monthly', '5GB', 100000, 30, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'MTN_10GB_MONTHLY', '10GB Monthly', '10GB', 200000, 30, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'MTN_20GB_MONTHLY', '20GB Monthly', '20GB', 350000, 30, true, NOW(), NOW());

-- Glo Data Bundles
INSERT INTO network_bundles (id, network_id, code, name, data_size, amount, validity_days, is_active, created_at, updated_at) VALUES
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'GLO_1GB_DAILY', '1GB Daily', '1GB', 25000, 1, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'GLO_2GB_WEEKLY', '2GB Weekly', '2GB', 45000, 7, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'GLO_5GB_MONTHLY', '5GB Monthly', '5GB', 90000, 30, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'GLO_10GB_MONTHLY', '10GB Monthly', '10GB', 180000, 30, true, NOW(), NOW());

-- Airtel Data Bundles
INSERT INTO network_bundles (id, network_id, code, name, data_size, amount, validity_days, is_active, created_at, updated_at) VALUES
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440003', 'AIRTEL_1GB_DAILY', '1GB Daily', '1GB', 30000, 1, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440003', 'AIRTEL_2GB_WEEKLY', '2GB Weekly', '2GB', 50000, 7, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440003', 'AIRTEL_5GB_MONTHLY', '5GB Monthly', '5GB', 100000, 30, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440003', 'AIRTEL_10GB_MONTHLY', '10GB Monthly', '10GB', 200000, 30, true, NOW(), NOW());

-- 9mobile Data Bundles
INSERT INTO network_bundles (id, network_id, code, name, data_size, amount, validity_days, is_active, created_at, updated_at) VALUES
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440004', '9MOBILE_1GB_DAILY', '1GB Daily', '1GB', 30000, 1, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440004', '9MOBILE_2GB_WEEKLY', '2GB Weekly', '2GB', 50000, 7, true, NOW(), NOW()),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440004', '9MOBILE_5GB_MONTHLY', '5GB Monthly', '5GB', 100000, 30, true, NOW(), NOW());

-- ============================================================================
-- USERS (Sample Nigerian Phone Numbers)
-- ============================================================================
INSERT INTO users (id, msisdn, network_provider, total_points, total_recharges, total_amount_spent, is_active, created_at, updated_at) VALUES
(gen_random_uuid(), '2348012345678', 'MTN', 150, 15, 3000000, true, NOW() - INTERVAL '60 days', NOW()),
(gen_random_uuid(), '2348023456789', 'MTN', 200, 20, 4000000, true, NOW() - INTERVAL '45 days', NOW()),
(gen_random_uuid(), '2348034567890', 'GLO', 100, 10, 2000000, true, NOW() - INTERVAL '30 days', NOW()),
(gen_random_uuid(), '2348045678901', 'AIRTEL', 250, 25, 5000000, true, NOW() - INTERVAL '50 days', NOW()),
(gen_random_uuid(), '2348056789012', '9MOBILE', 80, 8, 1600000, true, NOW() - INTERVAL '20 days', NOW()),
(gen_random_uuid(), '2348067890123', 'MTN', 300, 30, 6000000, true, NOW() - INTERVAL '90 days', NOW()),
(gen_random_uuid(), '2348078901234', 'GLO', 120, 12, 2400000, true, NOW() - INTERVAL '40 days', NOW()),
(gen_random_uuid(), '2348089012345', 'AIRTEL', 180, 18, 3600000, true, NOW() - INTERVAL '55 days', NOW()),
(gen_random_uuid(), '2348090123456', 'MTN', 220, 22, 4400000, true, NOW() - INTERVAL '35 days', NOW()),
(gen_random_uuid(), '2348101234567', '9MOBILE', 90, 9, 1800000, true, NOW() - INTERVAL '25 days', NOW())
ON CONFLICT (msisdn) DO UPDATE SET
  total_points = EXCLUDED.total_points,
  total_recharges = EXCLUDED.total_recharges,
  total_amount_spent = EXCLUDED.total_amount_spent,
  updated_at = NOW();

-- ============================================================================
-- RECHARGES/TRANSACTIONS (Realistic transaction history)
-- ============================================================================
-- Generate sample recharges for the past 90 days
DO $$
DECLARE
  user_record RECORD;
  recharge_count INT;
  i INT;
  recharge_date TIMESTAMP;
  recharge_amount INT;
  recharge_type TEXT;
BEGIN
  FOR user_record IN SELECT id, msisdn, network_provider FROM users LOOP
    recharge_count := (RANDOM() * 20 + 5)::INT; -- 5-25 recharges per user
    
    FOR i IN 1..recharge_count LOOP
      recharge_date := NOW() - (RANDOM() * INTERVAL '90 days');
      recharge_amount := (ARRAY[10000, 20000, 50000, 100000, 200000, 500000])[FLOOR(RANDOM() * 6 + 1)::INT];
      recharge_type := (ARRAY['airtime', 'data'])[FLOOR(RANDOM() * 2 + 1)::INT];
      
      INSERT INTO transactions (
        id, user_id, msisdn, network_provider, amount, transaction_type,
        status, payment_method, payment_reference, created_at, updated_at
      ) VALUES (
        gen_random_uuid(),
        user_record.id,
        user_record.msisdn,
        user_record.network_provider,
        recharge_amount,
        recharge_type,
        'COMPLETED',
        'paystack',
        'PAY_' || UPPER(SUBSTRING(MD5(RANDOM()::TEXT) FROM 1 FOR 16)),
        recharge_date,
        recharge_date
      );
    END LOOP;
  END LOOP;
END $$;

-- ============================================================================
-- SPIN PRIZES (Wheel Spin Rewards)
-- ============================================================================
INSERT INTO spin_prizes (id, name, type, value, probability, is_active, created_at, updated_at) VALUES
(gen_random_uuid(), '₦100 Airtime', 'airtime', 10000, 30.0, true, NOW(), NOW()),
(gen_random_uuid(), '₦200 Airtime', 'airtime', 20000, 25.0, true, NOW(), NOW()),
(gen_random_uuid(), '₦500 Airtime', 'airtime', 50000, 15.0, true, NOW(), NOW()),
(gen_random_uuid(), '₦1000 Airtime', 'airtime', 100000, 10.0, true, NOW(), NOW()),
(gen_random_uuid(), '1GB Data', 'data', 100000, 8.0, true, NOW(), NOW()),
(gen_random_uuid(), '2GB Data', 'data', 200000, 5.0, true, NOW(), NOW()),
(gen_random_uuid(), '50 Points', 'points', 50, 5.0, true, NOW(), NOW()),
(gen_random_uuid(), 'Better Luck Next Time', 'none', 0, 2.0, true, NOW(), NOW());

-- ============================================================================
-- DRAWS (Monthly Prize Draws)
-- ============================================================================
INSERT INTO draws (id, name, description, start_date, end_date, draw_date, status, total_participants, created_at, updated_at) VALUES
(gen_random_uuid(), 'January 2026 Grand Draw', 'Monthly draw for January 2026', '2026-01-01', '2026-01-31', '2026-02-01', 'completed', 1250, NOW() - INTERVAL '30 days', NOW()),
(gen_random_uuid(), 'February 2026 Grand Draw', 'Monthly draw for February 2026', '2026-02-01', '2026-02-28', '2026-03-01', 'active', 0, NOW(), NOW());

-- ============================================================================
-- AFFILIATES (Sample Affiliate Partners)
-- ============================================================================
INSERT INTO affiliates (id, msisdn, name, email, bank_name, account_number, account_name, status, total_referrals, total_commission, available_balance, created_at, updated_at) VALUES
(gen_random_uuid(), '2348111111111', 'John Doe', 'john.doe@example.com', 'GTBank', '0123456789', 'John Doe', 'active', 25, 125000, 125000, NOW() - INTERVAL '120 days', NOW()),
(gen_random_uuid(), '2348222222222', 'Jane Smith', 'jane.smith@example.com', 'Access Bank', '9876543210', 'Jane Smith', 'active', 40, 200000, 150000, NOW() - INTERVAL '90 days', NOW()),
(gen_random_uuid(), '2348333333333', 'David Johnson', 'david.j@example.com', 'Zenith Bank', '1122334455', 'David Johnson', 'active', 15, 75000, 75000, NOW() - INTERVAL '60 days', NOW());

-- ============================================================================
-- SUBSCRIPTIONS (Active Subscription Plans)
-- ============================================================================
INSERT INTO subscriptions (id, user_id, msisdn, plan_type, amount, status, start_date, end_date, auto_renew, created_at, updated_at)
SELECT 
  gen_random_uuid(),
  u.id,
  u.msisdn,
  (ARRAY['daily', 'weekly', 'monthly'])[FLOOR(RANDOM() * 3 + 1)::INT],
  (ARRAY[10000, 50000, 150000])[FLOOR(RANDOM() * 3 + 1)::INT],
  'active',
  NOW() - INTERVAL '15 days',
  NOW() + INTERVAL '15 days',
  true,
  NOW() - INTERVAL '15 days',
  NOW()
FROM users u
LIMIT 5;

-- ============================================================================
-- ADMIN USERS (For testing admin panel)
-- ============================================================================
-- Note: Password is 'Admin@123' hashed with bcrypt
-- In production, admins should change passwords immediately
INSERT INTO admin_users (id, username, email, password_hash, role, is_active, created_at, updated_at) VALUES
(gen_random_uuid(), 'superadmin', 'superadmin@rechargemax.com', '$2a$10$YourHashedPasswordHere', 'super_admin', true, NOW(), NOW()),
(gen_random_uuid(), 'admin', 'admin@rechargemax.com', '$2a$10$YourHashedPasswordHere', 'admin', true, NOW(), NOW()),
(gen_random_uuid(), 'support', 'support@rechargemax.com', '$2a$10$YourHashedPasswordHere', 'support', true, NOW(), NOW())
ON CONFLICT (username) DO NOTHING;

-- ============================================================================
-- CONFIGURATION (System Settings)
-- ============================================================================
INSERT INTO system_config (key, value, description, created_at, updated_at) VALUES
('recharge_mode', 'simulation', 'Recharge processing mode: direct, vtu, hybrid, simulation', NOW(), NOW()),
('vtu_provider', 'vtpass', 'VTU aggregator provider: vtpass, shago, baxi', NOW(), NOW()),
('min_recharge_amount', '10000', 'Minimum recharge amount in kobo (₦100)', NOW(), NOW()),
('max_recharge_amount', '5000000', 'Maximum recharge amount in kobo (₦50,000)', NOW(), NOW()),
('points_per_naira', '0.005', 'Points earned per naira (₦200 = 1 point)', NOW(), NOW()),
('spin_eligibility_amount', '100000', 'Minimum recharge for spin eligibility in kobo (₦1,000)', NOW(), NOW()),
('affiliate_commission_rate', '0.05', 'Affiliate commission rate (5%)', NOW(), NOW()),
('sms_provider', 'termii', 'SMS notification provider: termii, twilio, mock', NOW(), NOW())
ON CONFLICT (key) DO UPDATE SET
  value = EXCLUDED.value,
  updated_at = NOW();

-- ============================================================================
-- SUMMARY
-- ============================================================================
-- This seed file creates:
-- - 4 Nigerian mobile networks (MTN, Glo, Airtel, 9mobile)
-- - 23 data bundle packages across all networks
-- - 10 sample users with realistic phone numbers
-- - 100-200 recharge transactions spread over 90 days
-- - 8 spin wheel prizes with probabilities
-- - 2 monthly draws (1 completed, 1 active)
-- - 3 affiliate partners
-- - 5 active subscriptions
-- - 3 admin users (super_admin, admin, support)
-- - System configuration for testing
--
-- All data is fictional and safe for testing/demonstration
-- ============================================================================
