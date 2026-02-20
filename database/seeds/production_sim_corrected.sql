-- ============================================================================
-- RECHARGEMAX PRODUCTION SIMULATION SEED DATA (Schema-Corrected)
-- ============================================================================
-- This file creates realistic production data matching the actual database schema
-- Target: 1,000 users, 5,000+ transactions (scaled down for faster loading)
-- Purpose: Demo, testing, performance validation
-- ============================================================================

\echo 'Starting production simulation data generation...'

-- Clean existing data
TRUNCATE TABLE 
  spin_results,
  vtu_transactions,
  affiliate_commissions,
  affiliate_referrals,
  daily_subscriptions,
  wallet_transactions,
  wallets,
  affiliates,
  users
CASCADE;

\echo 'Cleaned existing data'

-- ============================================================================
-- PART 1: USERS (1,000 realistic users)
-- ============================================================================

\echo 'Creating 1,000 users...'

INSERT INTO users (
  id,
  msisdn,
  full_name,
  email,
  gender,
  date_of_birth,
  total_points,
  total_recharge_amount,
  loyalty_tier,
  referral_code,
  is_active,
  is_verified,
  created_at,
  updated_at
)
SELECT 
  gen_random_uuid(),
  -- Generate realistic Nigerian phone numbers
  CASE (i % 4)
    WHEN 0 THEN '0803' || LPAD((i * 1234567)::TEXT, 7, '0')
    WHEN 1 THEN '0805' || LPAD((i * 2345678)::TEXT, 7, '0')
    WHEN 2 THEN '0802' || LPAD((i * 3456789)::TEXT, 7, '0')
    ELSE '0809' || LPAD((i * 4567890)::TEXT, 7, '0')
  END,
  'User ' || i,
  'user' || i || '@rechargemax.ng',
  CASE (i % 3) WHEN 0 THEN 'MALE' WHEN 1 THEN 'FEMALE' ELSE '' END,
  NOW() - (RANDOM() * INTERVAL '18250 days' + INTERVAL '6570 days'),
  -- Points based on user tier
  CASE 
    WHEN i % 100 < 20 THEN (200 + RANDOM() * 800)::INT
    WHEN i % 100 < 70 THEN (50 + RANDOM() * 150)::INT
    ELSE (10 + RANDOM() * 40)::INT
  END,
  -- Total recharge amount (in kobo)
  CASE 
    WHEN i % 100 < 20 THEN (5000000 + RANDOM() * 15000000)::INT
    WHEN i % 100 < 70 THEN (1000000 + RANDOM() * 4000000)::INT
    ELSE (100000 + RANDOM() * 900000)::INT
  END,
  CASE 
    WHEN i % 100 < 5 THEN 'PLATINUM'
    WHEN i % 100 < 20 THEN 'GOLD'
    WHEN i % 100 < 50 THEN 'SILVER'
    ELSE 'BRONZE'
  END,
  CASE WHEN i % 10 = 0 THEN 'REF' || LPAD(i::TEXT, 6, '0') ELSE '' END,
  true,
  i % 100 < 80,
  NOW() - (RANDOM() * INTERVAL '180 days'),
  NOW() - (RANDOM() * INTERVAL '30 days')
FROM generate_series(1, 1000) AS i;

\echo 'Created 1,000 users'

-- Create wallets for all users
INSERT INTO wallets (
  id,
  user_id,
  balance,
  currency,
  created_at,
  updated_at
)
SELECT 
  gen_random_uuid(),
  id,
  0,
  'NGN',
  created_at,
  created_at
FROM users;

\echo 'Created wallets for all users'

-- ============================================================================
-- PART 2: RECHARGE TRANSACTIONS (5,000+ transactions)
-- ============================================================================

\echo 'Creating recharge transactions...'

INSERT INTO vtu_transactions (
  id,
  user_id,
  msisdn,
  amount,
  recharge_type,
  status,
  payment_method,
  payment_reference,
  payment_gateway,
  provider_reference,
  provider_response,
  created_at,
  updated_at,
  completed_at
)
SELECT 
  gen_random_uuid(),
  u.id,
  u.msisdn,
  -- Realistic recharge amounts (in kobo)
  CASE (i % 10)
    WHEN 0 THEN (50000 + RANDOM() * 150000)::INT
    WHEN 1 THEN (200000 + RANDOM() * 300000)::INT
    WHEN 2 THEN (500000 + RANDOM() * 500000)::INT
    WHEN 3 THEN (1000000 + RANDOM() * 1000000)::INT
    WHEN 4 THEN (2000000 + RANDOM() * 3000000)::INT
    ELSE (100000 + RANDOM() * 400000)::INT
  END,
  CASE WHEN i % 100 < 60 THEN 'airtime' ELSE 'data' END,
  CASE WHEN i % 100 < 95 THEN 'COMPLETED' WHEN i % 100 < 98 THEN 'pending' ELSE 'failed' END,
  'paystack',
  'PAY-' || TO_CHAR(NOW(), 'YYYYMMDD') || '-' || LPAD(i::TEXT, 8, '0'),
  'paystack',
  CASE WHEN i % 100 < 95 THEN 'VTU-' || i::TEXT ELSE NULL END,
  CASE WHEN i % 100 < 95 THEN '{"status": "success"}' ELSE NULL END,
  u.created_at + (RANDOM() * (NOW() - u.created_at)),
  u.created_at + (RANDOM() * (NOW() - u.created_at)) + INTERVAL '5 seconds',
  CASE WHEN i % 100 < 95 THEN u.created_at + (RANDOM() * (NOW() - u.created_at)) + INTERVAL '10 seconds' ELSE NULL END
FROM users u
CROSS JOIN generate_series(1, 5) AS i;

\echo 'Created 5,000+ recharge transactions'

-- ============================================================================
-- PART 3: WHEEL PRIZES & SPIN RESULTS
-- ============================================================================

\echo 'Creating wheel prizes...'

-- Ensure wheel prizes exist
INSERT INTO wheel_prizes (
  id,
  prize_name,
  prize_type,
  prize_value,
  probability,
  icon,
  color,
  is_active,
  total_available,
  total_claimed,
  created_at,
  updated_at
) VALUES
  (gen_random_uuid(), 'Better Luck Next Time', 'none', 0, 40.0, '😢', '#gray', true, 999999, 0, NOW(), NOW()),
  (gen_random_uuid(), '₦100 Airtime', 'airtime', 10000, 25.0, '📱', '#blue', true, 10000, 0, NOW(), NOW()),
  (gen_random_uuid(), '₦200 Airtime', 'airtime', 20000, 15.0, '📱', '#green', true, 5000, 0, NOW(), NOW()),
  (gen_random_uuid(), '₦500 Airtime', 'airtime', 50000, 10.0, '💰', '#yellow', true, 2000, 0, NOW(), NOW()),
  (gen_random_uuid(), '₦1000 Airtime', 'airtime', 100000, 5.0, '💎', '#purple', true, 1000, 0, NOW(), NOW()),
  (gen_random_uuid(), '100 Bonus Points', 'points', 100, 3.0, '⭐', '#orange', true, 5000, 0, NOW(), NOW()),
  (gen_random_uuid(), 'iPhone 15 Pro', 'physical', 150000000, 0.5, '📱', '#red', true, 10, 0, NOW(), NOW())
ON CONFLICT (prize_name) DO UPDATE SET
  probability = EXCLUDED.probability,
  updated_at = NOW();

\echo 'Created wheel prizes'

\echo 'Creating spin results...'

-- Create spin results for eligible transactions (>= ₦1000)
INSERT INTO spin_results (
  id,
  user_id,
  recharge_id,
  prize_id,
  prize_name,
  prize_type,
  prize_value,
  status,
  spun_at,
  created_at,
  updated_at
)
SELECT 
  gen_random_uuid(),
  vt.user_id,
  vt.id,
  wp.id,
  wp.prize_name,
  wp.prize_type,
  wp.prize_value,
  'claimed',
  vt.completed_at + INTERVAL '30 seconds',
  vt.completed_at + INTERVAL '30 seconds',
  vt.completed_at + INTERVAL '30 seconds'
FROM vtu_transactions vt
CROSS JOIN LATERAL (
  SELECT id, prize_name, prize_type, prize_value
  FROM wheel_prizes
  WHERE is_active = true
  ORDER BY RANDOM()
  LIMIT 1
) wp
WHERE vt.status = 'COMPLETED' 
  AND vt.amount >= 100000;

\echo 'Created spin results for eligible transactions'

-- ============================================================================
-- PART 4: AFFILIATES (100 affiliates)
-- ============================================================================

\echo 'Creating affiliate network...'

-- Create affiliates from users with referral codes
INSERT INTO affiliates (
  id,
  user_id,
  affiliate_code,
  status,
  total_referrals,
  total_commission,
  commission_rate,
  bank_name,
  account_number,
  account_name,
  created_at,
  updated_at,
  approved_at
)
SELECT 
  gen_random_uuid(),
  id,
  referral_code,
  'active',
  (5 + RANDOM() * 15)::INT,
  (100000 + RANDOM() * 500000)::INT,
  5.0,
  (ARRAY['GTBank', 'Access Bank', 'First Bank', 'UBA', 'Zenith Bank'])[FLOOR(RANDOM() * 5 + 1)::INT],
  LPAD((1000000000 + RANDOM() * 999999999)::BIGINT::TEXT, 10, '0'),
  full_name,
  created_at,
  created_at + INTERVAL '1 day',
  created_at + INTERVAL '1 day'
FROM users
WHERE referral_code != ''
LIMIT 100;

\echo 'Created 100 affiliates'

-- ============================================================================
-- PART 5: DAILY SUBSCRIPTIONS (200 subscribers)
-- ============================================================================

\echo 'Creating daily subscriptions...'

INSERT INTO daily_subscriptions (
  id,
  user_id,
  msisdn,
  subscription_amount,
  status,
  start_date,
  end_date,
  next_billing_date,
  payment_method,
  payment_reference,
  created_at,
  updated_at
)
SELECT 
  gen_random_uuid(),
  id,
  msisdn,
  20000,
  CASE WHEN i % 100 < 80 THEN 'active' WHEN i % 100 < 95 THEN 'cancelled' ELSE 'expired' END,
  (created_at + (RANDOM() * INTERVAL '90 days'))::DATE,
  CASE WHEN i % 100 >= 80 THEN (created_at + (RANDOM() * INTERVAL '120 days'))::DATE ELSE NULL END,
  CASE WHEN i % 100 < 80 THEN (NOW() + INTERVAL '1 day')::DATE ELSE NULL END,
  'paystack',
  'SUB-' || TO_CHAR(NOW(), 'YYYYMMDD') || '-' || LPAD(i::TEXT, 6, '0'),
  created_at,
  NOW()
FROM users
CROSS JOIN generate_series(1, 1) AS i
WHERE total_recharge_amount >= 1000000
LIMIT 200;

\echo 'Created 200 daily subscriptions'

-- ============================================================================
-- FINAL STATISTICS
-- ============================================================================

\echo '========================================'
\echo 'PRODUCTION SIMULATION DATA - COMPLETE'
\echo '========================================'

SELECT 'Total Users' as metric, COUNT(*) as count FROM users;
SELECT 'Total Transactions' as metric, COUNT(*) as count FROM vtu_transactions WHERE status = 'COMPLETED';
SELECT 'Total Spin Results' as metric, COUNT(*) as count FROM spin_results;
SELECT 'Total Affiliates' as metric, COUNT(*) as count FROM affiliates;
SELECT 'Total Active Subscriptions' as metric, COUNT(*) as count FROM daily_subscriptions WHERE status = 'active';

\echo '========================================'
\echo 'Platform ready for testing and demo!'
\echo '========================================'
