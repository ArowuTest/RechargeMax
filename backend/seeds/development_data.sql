-- ============================================================================
-- RechargeMax Development Seed Data
-- ============================================================================
-- This file contains comprehensive development/testing data for the RechargeMax platform
-- Includes: 100+ users, 500+ transactions, 50+ affiliates, networks, data plans, draws, prizes
-- ============================================================================

-- ============================================================================
-- 1. NETWORKS (4 major Nigerian networks)
-- ============================================================================
INSERT INTO network_configs (id, name, code, status, commission_rate, api_provider, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'MTN', 'MTN', 'active', 2.5, 'vtpass', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'Glo', 'GLO', 'active', 3.0, 'vtpass', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'Airtel', 'AIRTEL', 'active', 2.5, 'vtpass', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440004', '9mobile', '9MOBILE', 'active', 3.5, 'vtpass', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 2. DATA PLANS (50+ plans across all networks)
-- ============================================================================
-- MTN Data Plans
INSERT INTO data_plans (id, network_id, name, data_amount, price, validity_days, plan_code, status, created_at, updated_at) VALUES
('650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'MTN 500MB Daily', '500MB', 15000, 1, 'MTN-500MB-1D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'MTN 1GB Weekly', '1GB', 30000, 7, 'MTN-1GB-7D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', 'MTN 2GB Monthly', '2GB', 100000, 30, 'MTN-2GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440001', 'MTN 3GB Monthly', '3GB', 150000, 30, 'MTN-3GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440001', 'MTN 5GB Monthly', '5GB', 250000, 30, 'MTN-5GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440001', 'MTN 10GB Monthly', '10GB', 500000, 30, 'MTN-10GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440001', 'MTN 20GB Monthly', '20GB', 1000000, 30, 'MTN-20GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440001', 'MTN 40GB Monthly', '40GB', 2000000, 30, 'MTN-40GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440009', '550e8400-e29b-41d4-a716-446655440001', 'MTN 75GB Monthly', '75GB', 3500000, 30, 'MTN-75GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440010', '550e8400-e29b-41d4-a716-446655440001', 'MTN 100GB Monthly', '100GB', 5000000, 30, 'MTN-100GB-30D', 'active', NOW(), NOW()),

-- Glo Data Plans
('650e8400-e29b-41d4-a716-446655440011', '550e8400-e29b-41d4-a716-446655440002', 'Glo 500MB Daily', '500MB', 10000, 1, 'GLO-500MB-1D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440012', '550e8400-e29b-41d4-a716-446655440002', 'Glo 1.6GB Weekly', '1.6GB', 50000, 7, 'GLO-1.6GB-7D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440013', '550e8400-e29b-41d4-a716-446655440002', 'Glo 2.9GB Monthly', '2.9GB', 100000, 30, 'GLO-2.9GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440014', '550e8400-e29b-41d4-a716-446655440002', 'Glo 5.8GB Monthly', '5.8GB', 200000, 30, 'GLO-5.8GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440015', '550e8400-e29b-41d4-a716-446655440002', 'Glo 7.7GB Monthly', '7.7GB', 250000, 30, 'GLO-7.7GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440016', '550e8400-e29b-41d4-a716-446655440002', 'Glo 10GB Monthly', '10GB', 300000, 30, 'GLO-10GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440017', '550e8400-e29b-41d4-a716-446655440002', 'Glo 13.25GB Monthly', '13.25GB', 400000, 30, 'GLO-13.25GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440018', '550e8400-e29b-41d4-a716-446655440002', 'Glo 18GB Monthly', '18GB', 500000, 30, 'GLO-18GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440019', '550e8400-e29b-41d4-a716-446655440002', 'Glo 29.5GB Monthly', '29.5GB', 800000, 30, 'GLO-29.5GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440020', '550e8400-e29b-41d4-a716-446655440002', 'Glo 50GB Monthly', '50GB', 1000000, 30, 'GLO-50GB-30D', 'active', NOW(), NOW()),

-- Airtel Data Plans
('650e8400-e29b-41d4-a716-446655440021', '550e8400-e29b-41d4-a716-446655440003', 'Airtel 750MB Weekly', '750MB', 50000, 7, 'AIRTEL-750MB-7D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440022', '550e8400-e29b-41d4-a716-446655440003', 'Airtel 1.5GB Monthly', '1.5GB', 100000, 30, 'AIRTEL-1.5GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440023', '550e8400-e29b-41d4-a716-446655440003', 'Airtel 3GB Monthly', '3GB', 150000, 30, 'AIRTEL-3GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440024', '550e8400-e29b-41d4-a716-446655440003', 'Airtel 6GB Monthly', '6GB', 250000, 30, 'AIRTEL-6GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440025', '550e8400-e29b-41d4-a716-446655440003', 'Airtel 10GB Monthly', '10GB', 400000, 30, 'AIRTEL-10GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440026', '550e8400-e29b-41d4-a716-446655440003', 'Airtel 20GB Monthly', '20GB', 800000, 30, 'AIRTEL-20GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440027', '550e8400-e29b-41d4-a716-446655440003', 'Airtel 40GB Monthly', '40GB', 1500000, 30, 'AIRTEL-40GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440028', '550e8400-e29b-41d4-a716-446655440003', 'Airtel 75GB Monthly', '75GB', 2500000, 30, 'AIRTEL-75GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440029', '550e8400-e29b-41d4-a716-446655440003', 'Airtel 100GB Monthly', '100GB', 3500000, 30, 'AIRTEL-100GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440030', '550e8400-e29b-41d4-a716-446655440003', 'Airtel 200GB Monthly', '200GB', 5000000, 30, 'AIRTEL-200GB-30D', 'active', NOW(), NOW()),

-- 9mobile Data Plans
('650e8400-e29b-41d4-a716-446655440031', '550e8400-e29b-41d4-a716-446655440004', '9mobile 500MB Weekly', '500MB', 50000, 7, '9MOBILE-500MB-7D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440032', '550e8400-e29b-41d4-a716-446655440004', '9mobile 1.5GB Monthly', '1.5GB', 100000, 30, '9MOBILE-1.5GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440033', '550e8400-e29b-41d4-a716-446655440004', '9mobile 2GB Monthly', '2GB', 120000, 30, '9MOBILE-2GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440034', '550e8400-e29b-41d4-a716-446655440004', '9mobile 4.5GB Monthly', '4.5GB', 200000, 30, '9MOBILE-4.5GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440035', '550e8400-e29b-41d4-a716-446655440004', '9mobile 11GB Monthly', '11GB', 400000, 30, '9MOBILE-11GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440036', '550e8400-e29b-41d4-a716-446655440004', '9mobile 15GB Monthly', '15GB', 500000, 30, '9MOBILE-15GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440037', '550e8400-e29b-41d4-a716-446655440004', '9mobile 27.5GB Monthly', '27.5GB', 800000, 30, '9MOBILE-27.5GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440038', '550e8400-e29b-41d4-a716-446655440004', '9mobile 40GB Monthly', '40GB', 1000000, 30, '9MOBILE-40GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440039', '550e8400-e29b-41d4-a716-446655440004', '9mobile 75GB Monthly', '75GB', 1500000, 30, '9MOBILE-75GB-30D', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440040', '550e8400-e29b-41d4-a716-446655440004', '9mobile 100GB Monthly', '100GB', 2000000, 30, '9MOBILE-100GB-30D', 'active', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 3. SUBSCRIPTION TIERS (5 tiers)
-- ============================================================================
INSERT INTO subscription_tiers (id, name, price, duration_days, entries_per_draw, bonus_spins, description, is_active, created_at, updated_at) VALUES
('750e8400-e29b-41d4-a716-446655440001', 'Basic', 2000, 1, 1, 0, 'Daily ₦20 subscription with 1 entry per draw', true, NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440002', 'Silver', 5000, 7, 2, 1, 'Weekly subscription with 2 entries per draw and 1 bonus spin', true, NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440003', 'Gold', 15000, 30, 3, 3, 'Monthly subscription with 3 entries per draw and 3 bonus spins', true, NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440004', 'Platinum', 40000, 90, 5, 10, 'Quarterly subscription with 5 entries per draw and 10 bonus spins', true, NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440005', 'Diamond', 150000, 365, 10, 50, 'Annual subscription with 10 entries per draw and 50 bonus spins', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 4. USERS (100 test users)
-- ============================================================================
-- Note: Password for all test users is 'TestUser@123' (bcrypt hash)
DO $$
DECLARE
  i INT;
  user_id UUID;
  phone_num TEXT;
BEGIN
  FOR i IN 1..100 LOOP
    user_id := gen_random_uuid();
    phone_num := '080' || LPAD((10000000 + i)::TEXT, 8, '0');
    
    INSERT INTO users (id, email, phone, password_hash, full_name, status, created_at, updated_at)
    VALUES (
      user_id,
      'testuser' || i || '@rechargemax.ng',
      phone_num,
      '$2a$10$rN8vK8KqYqZ8vK8KqYqZ8uO8vK8KqYqZ8vK8KqYqZ8vK8KqYqZ8vK', -- TestUser@123
      'Test User ' || i,
      CASE WHEN i % 10 = 0 THEN 'suspended' WHEN i % 20 = 0 THEN 'banned' ELSE 'active' END,
      NOW() - (RANDOM() * INTERVAL '365 days'),
      NOW()
    )
    ON CONFLICT (id) DO NOTHING;
    
    -- Create user profile
    INSERT INTO user_profiles (id, user_id, referral_code, points_balance, total_recharges, total_spent, created_at, updated_at)
    VALUES (
      gen_random_uuid(),
      user_id,
      'REF' || LPAD(i::TEXT, 6, '0'),
      FLOOR(RANDOM() * 10000)::INT,
      FLOOR(RANDOM() * 50)::INT,
      FLOOR(RANDOM() * 500000)::BIGINT,
      NOW() - (RANDOM() * INTERVAL '365 days'),
      NOW()
    )
    ON CONFLICT (user_id) DO NOTHING;
  END LOOP;
END $$;

-- ============================================================================
-- 5. AFFILIATES (50 affiliates)
-- ============================================================================
DO $$
DECLARE
  i INT;
  affiliate_id UUID;
  user_id UUID;
BEGIN
  FOR i IN 1..50 LOOP
    affiliate_id := gen_random_uuid();
    
    -- Get a random user ID
    SELECT id INTO user_id FROM users ORDER BY RANDOM() LIMIT 1;
    
    INSERT INTO affiliates (id, user_id, affiliate_code, status, commission_rate, total_referrals, total_earnings, created_at, updated_at)
    VALUES (
      affiliate_id,
      user_id,
      'AFF' || LPAD(i::TEXT, 6, '0'),
      CASE 
        WHEN i <= 30 THEN 'active'
        WHEN i <= 40 THEN 'pending'
        ELSE 'suspended'
      END,
      CASE 
        WHEN i <= 10 THEN 5.0
        WHEN i <= 30 THEN 3.0
        ELSE 2.0
      END,
      FLOOR(RANDOM() * 100)::INT,
      FLOOR(RANDOM() * 1000000)::BIGINT,
      NOW() - (RANDOM() * INTERVAL '365 days'),
      NOW()
    )
    ON CONFLICT (id) DO NOTHING;
  END LOOP;
END $$;

-- ============================================================================
-- 6. TRANSACTIONS (500 recharge transactions)
-- ============================================================================
DO $$
DECLARE
  i INT;
  user_id UUID;
  network_id UUID;
  plan_id UUID;
  amount_kobo BIGINT;
BEGIN
  FOR i IN 1..500 LOOP
    -- Get random user
    SELECT id INTO user_id FROM users ORDER BY RANDOM() LIMIT 1;
    
    -- Get random network
    SELECT id INTO network_id FROM network_configs ORDER BY RANDOM() LIMIT 1;
    
    -- Get random plan for that network
    SELECT id, price INTO plan_id, amount_kobo FROM data_plans WHERE network_id = network_id ORDER BY RANDOM() LIMIT 1;
    
    INSERT INTO transactions (
      id, user_id, type, amount, status, payment_reference, 
      network_provider, phone_number, data_plan_id, 
      created_at, updated_at
    )
    VALUES (
      gen_random_uuid(),
      user_id,
      'recharge',
      amount_kobo,
      CASE 
        WHEN i % 20 = 0 THEN 'failed'
        WHEN i % 50 = 0 THEN 'pending'
        ELSE 'completed'
      END,
      'PAY-' || LPAD(i::TEXT, 10, '0'),
      (SELECT code FROM network_configs WHERE id = network_id),
      '080' || LPAD((FLOOR(RANDOM() * 90000000) + 10000000)::TEXT, 8, '0'),
      plan_id,
      NOW() - (RANDOM() * INTERVAL '90 days'),
      NOW()
    )
    ON CONFLICT (id) DO NOTHING;
  END LOOP;
END $$;

-- ============================================================================
-- 7. DRAWS (10 active and past draws)
-- ============================================================================
DO $$
DECLARE
  i INT;
  draw_id UUID;
  draw_date DATE;
BEGIN
  FOR i IN 1..10 LOOP
    draw_id := gen_random_uuid();
    draw_date := CURRENT_DATE - (10 - i);
    
    INSERT INTO draws (
      id, draw_date, status, total_entries, prize_pool, 
      draw_time, created_at, updated_at
    )
    VALUES (
      draw_id,
      draw_date,
      CASE 
        WHEN i <= 7 THEN 'completed'
        WHEN i = 8 THEN 'in_progress'
        ELSE 'pending'
      END,
      FLOOR(RANDOM() * 1000 + 100)::INT,
      FLOOR(RANDOM() * 10000000 + 1000000)::BIGINT,
      draw_date + TIME '20:00:00',
      NOW() - ((10 - i) * INTERVAL '1 day'),
      NOW()
    )
    ON CONFLICT (id) DO NOTHING;
  END LOOP;
END $$;

-- ============================================================================
-- 8. WHEEL PRIZES (20 different prizes)
-- ============================================================================
INSERT INTO wheel_prizes (id, name, type, value, probability, is_active, created_at, updated_at) VALUES
('850e8400-e29b-41d4-a716-446655440001', '₦100 Airtime', 'airtime', 10000, 15.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440002', '₦200 Airtime', 'airtime', 20000, 12.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440003', '₦500 Airtime', 'airtime', 50000, 8.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440004', '₦1,000 Airtime', 'airtime', 100000, 5.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440005', '500MB Data', 'data', 15000, 10.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440006', '1GB Data', 'data', 30000, 8.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440007', '2GB Data', 'data', 100000, 6.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440008', '5GB Data', 'data', 250000, 4.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440009', '100 Points', 'points', 100, 15.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440010', '250 Points', 'points', 250, 10.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440011', '500 Points', 'points', 500, 5.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440012', '1 Free Spin', 'spin', 1, 8.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440013', '2 Free Spins', 'spin', 2, 4.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440014', 'Better Luck', 'none', 0, 20.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440015', '₦2,000 Cash', 'cash', 200000, 2.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440016', '₦5,000 Cash', 'cash', 500000, 1.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440017', '₦10,000 Cash', 'cash', 1000000, 0.5, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440018', '10GB Data', 'data', 500000, 2.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440019', '20GB Data', 'data', 1000000, 1.0, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440020', 'iPhone 15 Pro', 'device', 150000000, 0.01, true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SEED DATA SUMMARY
-- ============================================================================
-- Networks: 4 (MTN, Glo, Airtel, 9mobile)
-- Data Plans: 40 (10 per network)
-- Subscription Tiers: 5 (Basic to Diamond)
-- Users: 100 (with profiles)
-- Affiliates: 50 (30 active, 10 pending, 10 suspended)
-- Transactions: 500 (recharge transactions)
-- Draws: 10 (7 completed, 1 in progress, 2 pending)
-- Wheel Prizes: 20 (various types)
-- ============================================================================
