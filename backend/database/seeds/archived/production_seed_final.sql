-- ============================================================================
-- RECHARGEMAX PRODUCTION SIMULATION SEED DATA (Final - Schema Aligned)
-- ============================================================================
-- This file creates realistic production data perfectly aligned with actual schema
-- Target: 1,000 users, 5,000+ transactions for realistic testing and demos
-- Purpose: Demo, testing, performance validation
-- ============================================================================

\echo '========================================'
\echo 'RECHARGEMAX PRODUCTION SEED DATA'
\echo '========================================'

-- Clean existing data (except system tables)
\echo 'Cleaning existing data...'

TRUNCATE TABLE 
  spin_results,
  vtu_transactions,
  affiliate_commissions,
  affiliate_payouts,
  daily_subscriptions,
  wallet_transactions,
  wallets,
  affiliates,
  users
CASCADE;

\echo 'Existing data cleaned'

-- ============================================================================
-- PART 1: USERS (1,000 realistic users)
-- ============================================================================

\echo 'Creating 1,000 users with realistic profiles...'

INSERT INTO users (
  id,
  msisdn,
  full_name,
  email,
  gender,
  date_of_birth,
  state,
  city,
  total_points,
  total_recharge_amount,
  total_transactions,
  loyalty_tier,
  referral_code,
  is_active,
  is_verified,
  phone_verified,
  kyc_status,
  created_at,
  updated_at
)
SELECT 
  gen_random_uuid(),
  -- Generate realistic Nigerian phone numbers (234 format)
  '234' || CASE (i % 4)
    WHEN 0 THEN '803' || LPAD((1000000 + (i * 1234) % 9000000)::TEXT, 7, '0')
    WHEN 1 THEN '805' || LPAD((1000000 + (i * 2345) % 9000000)::TEXT, 7, '0')
    WHEN 2 THEN '802' || LPAD((1000000 + (i * 3456) % 9000000)::TEXT, 7, '0')
    ELSE '809' || LPAD((1000000 + (i * 4567) % 9000000)::TEXT, 7, '0')
  END,
  -- Realistic Nigerian names
  (ARRAY['Adebayo Johnson', 'Chioma Okafor', 'Emeka Nwankwo', 'Fatima Abubakar', 'Ibrahim Yusuf',
         'Ngozi Eze', 'Oluwaseun Adeleke', 'Blessing Okoro', 'Chinedu Obi', 'Amina Mohammed',
         'Tunde Bakare', 'Aisha Ibrahim', 'Kunle Adeyemi', 'Zainab Hassan', 'Segun Ogunleye'])[((i-1) % 15) + 1] || ' ' || i,
  'user' || i || '@rechargemax.ng',
  CASE (i % 3) WHEN 0 THEN 'MALE' WHEN 1 THEN 'FEMALE' ELSE '' END,
  (NOW() - (RANDOM() * INTERVAL '18250 days' + INTERVAL '6570 days'))::DATE,
  (ARRAY['Lagos', 'Abuja', 'Kano', 'Port Harcourt', 'Ibadan', 'Kaduna', 'Enugu'])[((i-1) % 7) + 1],
  (ARRAY['Ikeja', 'Lekki', 'Surulere', 'Yaba', 'Victoria Island', 'Ikoyi', 'Ajah'])[((i-1) % 7) + 1],
  -- Points based on user tier
  CASE 
    WHEN i % 100 < 5 THEN (500 + RANDOM() * 1500)::INT  -- Platinum
    WHEN i % 100 < 20 THEN (200 + RANDOM() * 300)::INT  -- Gold
    WHEN i % 100 < 50 THEN (50 + RANDOM() * 150)::INT   -- Silver
    ELSE (10 + RANDOM() * 40)::INT                       -- Bronze
  END,
  -- Total recharge amount (in kobo) - determines loyalty tier
  CASE 
    WHEN i % 100 < 5 THEN (5000000 + RANDOM() * 15000000)::BIGINT  -- ₦50k-200k
    WHEN i % 100 < 20 THEN (1000000 + RANDOM() * 4000000)::BIGINT  -- ₦10k-50k
    WHEN i % 100 < 50 THEN (500000 + RANDOM() * 500000)::BIGINT    -- ₦5k-10k
    ELSE (100000 + RANDOM() * 400000)::BIGINT                       -- ₦1k-5k
  END,
  -- Total transactions
  CASE 
    WHEN i % 100 < 5 THEN (50 + RANDOM() * 150)::INT
    WHEN i % 100 < 20 THEN (20 + RANDOM() * 30)::INT
    WHEN i % 100 < 50 THEN (5 + RANDOM() * 15)::INT
    ELSE (1 + RANDOM() * 4)::INT
  END,
  CASE 
    WHEN i % 100 < 5 THEN 'PLATINUM'
    WHEN i % 100 < 20 THEN 'GOLD'
    WHEN i % 100 < 50 THEN 'SILVER'
    ELSE 'BRONZE'
  END,
  -- Generate referral codes for 10% of users
  CASE WHEN i % 10 = 0 THEN 'REF' || LPAD(i::TEXT, 6, '0') ELSE '' END,
  true,
  i % 100 < 80,  -- 80% verified
  i % 100 < 90,  -- 90% phone verified
  CASE WHEN i % 100 < 70 THEN 'VERIFIED' WHEN i % 100 < 90 THEN 'PENDING' ELSE 'REJECTED' END,
  NOW() - (RANDOM() * INTERVAL '180 days'),
  NOW() - (RANDOM() * INTERVAL '30 days')
FROM generate_series(1, 1000) AS i;

\echo '✓ Created 1,000 users'

-- ============================================================================
-- PART 2: WALLETS (One per user)
-- ============================================================================

\echo 'Creating wallets for all users...'

INSERT INTO wallets (
  id,
  user_id,
  balance,
  created_at,
  updated_at
)
SELECT 
  gen_random_uuid(),
  id,
  0,  -- Start with zero balance
  created_at,
  created_at
FROM users;

\echo '✓ Created 1,000 wallets'

\echo 'Part 1 Complete: Users and Wallets created'
\echo '========================================'

-- ============================================================================
-- PART 3: VTU TRANSACTIONS (5,000+ realistic transactions)
-- ============================================================================

\echo 'Creating VTU transactions...'

INSERT INTO vtu_transactions (
  id,
  user_id,
  amount,
  recharge_type,
  network,
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
  -- Realistic recharge amounts (in kobo)
  CASE (i % 10)
    WHEN 0 THEN (50000 + RANDOM() * 50000)::BIGINT    -- ₦500-1000
    WHEN 1 THEN (100000 + RANDOM() * 100000)::BIGINT  -- ₦1000-2000
    WHEN 2 THEN (200000 + RANDOM() * 300000)::BIGINT  -- ₦2000-5000
    WHEN 3 THEN (500000 + RANDOM() * 500000)::BIGINT  -- ₦5000-10000
    WHEN 4 THEN (1000000 + RANDOM() * 1000000)::BIGINT -- ₦10000-20000
    ELSE (150000 + RANDOM() * 350000)::BIGINT          -- ₦1500-5000
  END,
  CASE WHEN i % 100 < 60 THEN 'airtime' ELSE 'data' END,
  (ARRAY['MTN', 'GLO', 'AIRTEL', '9MOBILE'])[((i-1) % 4) + 1],
  CASE 
    WHEN i % 100 < 90 THEN 'COMPLETED'
    WHEN i % 100 < 95 THEN 'pending'
    ELSE 'failed'
  END,
  'paystack',
  'PAY-' || TO_CHAR(NOW(), 'YYYYMMDD') || '-' || LPAD((u_row * 100 + i)::TEXT, 10, '0'),
  'paystack',
  CASE WHEN i % 100 < 90 THEN 'VTU-' || (u_row * 100 + i)::TEXT ELSE NULL END,
  CASE WHEN i % 100 < 90 THEN '{"status": "success", "message": "Recharge successful"}' ELSE NULL END,
  u.created_at + (RANDOM() * (NOW() - u.created_at)),
  u.created_at + (RANDOM() * (NOW() - u.created_at)) + INTERVAL '5 seconds',
  CASE WHEN i % 100 < 90 THEN u.created_at + (RANDOM() * (NOW() - u.created_at)) + INTERVAL '10 seconds' ELSE NULL END
FROM users u
CROSS JOIN generate_series(1, 5) AS i
CROSS JOIN LATERAL (SELECT ROW_NUMBER() OVER () as u_row FROM users WHERE id = u.id) AS row_data;

\echo '✓ Created 5,000+ VTU transactions'

-- ============================================================================
-- PART 4: WHEEL PRIZES (Ensure prizes exist)
-- ============================================================================

\echo 'Creating wheel prizes...'

-- Delete existing prizes first
DELETE FROM wheel_prizes;

INSERT INTO wheel_prizes (
  id,
  prize_name,
  prize_type,
  prize_value,
  probability,
  color,
  is_active,
  total_available,
  total_claimed,
  created_at,
  updated_at
) VALUES
  (gen_random_uuid(), 'Better Luck Next Time', 'none', 0, 40.0, '#gray', true, 999999, 0, NOW(), NOW()),
  (gen_random_uuid(), '₦100 Airtime', 'airtime', 10000, 25.0, '#blue', true, 10000, 0, NOW(), NOW()),
  (gen_random_uuid(), '₦200 Airtime', 'airtime', 20000, 15.0, '#green', true, 5000, 0, NOW(), NOW()),
  (gen_random_uuid(), '₦500 Airtime', 'airtime', 50000, 10.0, '#yellow', true, 2000, 0, NOW(), NOW()),
  (gen_random_uuid(), '₦1000 Airtime', 'airtime', 100000, 5.0, '#purple', true, 1000, 0, NOW(), NOW()),
  (gen_random_uuid(), '100 Bonus Points', 'points', 100, 3.0, '#orange', true, 5000, 0, NOW(), NOW()),
  (gen_random_uuid(), 'iPhone 15 Pro', 'physical', 150000000, 0.5, '#red', true, 10, 0, NOW(), NOW()),
  (gen_random_uuid(), '₦2000 Airtime', 'airtime', 200000, 1.0, '#gold', true, 500, 0, NOW(), NOW());

\echo '✓ Created 8 wheel prizes'

-- ============================================================================
-- PART 5: SPIN RESULTS (For transactions >= ₦1000)
-- ============================================================================

\echo 'Creating spin results for eligible transactions...'

INSERT INTO spin_results (
  id,
  user_id,
  transaction_id,
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
  ORDER BY 
    CASE 
      WHEN probability >= 40 THEN RANDOM() * 0.4
      WHEN probability >= 25 THEN RANDOM() * 0.25 + 0.4
      WHEN probability >= 15 THEN RANDOM() * 0.15 + 0.65
      WHEN probability >= 10 THEN RANDOM() * 0.10 + 0.80
      ELSE RANDOM() * 0.10 + 0.90
    END
  LIMIT 1
) wp
WHERE vt.status = 'COMPLETED' 
  AND vt.amount >= 100000  -- ₦1000 or more
  AND vt.completed_at IS NOT NULL;

\echo '✓ Created spin results for eligible transactions'

\echo 'Part 2 Complete: Transactions and Gamification created'
\echo '========================================'

-- ============================================================================
-- PART 6: AFFILIATES (100 active affiliates)
-- ============================================================================

\echo 'Creating affiliate network...'

INSERT INTO affiliates (
  id,
  user_id,
  affiliate_code,
  tier,
  status,
  commission_rate,
  total_referrals,
  total_earnings,
  total_commission,
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
  CASE 
    WHEN loyalty_tier = 'PLATINUM' THEN 'PLATINUM'
    WHEN loyalty_tier = 'GOLD' THEN 'GOLD'
    WHEN loyalty_tier = 'SILVER' THEN 'SILVER'
    ELSE 'BRONZE'
  END,
  'APPROVED',  -- Uppercase as per constraint
  5.0,
  (5 + RANDOM() * 20)::INT,
  (0 + RANDOM() * 500000)::INT,
  (50000 + RANDOM() * 450000)::NUMERIC(12,2),
  (ARRAY['GTBank', 'Access Bank', 'First Bank', 'UBA', 'Zenith Bank', 'Fidelity Bank', 'Sterling Bank'])[FLOOR(RANDOM() * 7 + 1)::INT],
  LPAD((1000000000 + RANDOM() * 8999999999)::BIGINT::TEXT, 10, '0'),
  full_name,
  created_at,
  created_at + INTERVAL '1 day',
  created_at + INTERVAL '1 day'
FROM users
WHERE referral_code != '' AND referral_code IS NOT NULL
LIMIT 100;

\echo '✓ Created 100 active affiliates'

-- ============================================================================
-- PART 7: DAILY SUBSCRIPTIONS (200 subscribers)
-- ============================================================================

\echo 'Creating daily subscriptions...'

INSERT INTO daily_subscriptions (
  id,
  user_id,
  msisdn,
  subscription_date,
  amount,
  draw_entries_earned,
  points_earned,
  payment_reference,
  status,
  is_paid,
  customer_email,
  customer_name,
  created_at
)
SELECT 
  gen_random_uuid(),
  u.id,
  u.msisdn,
  (u.created_at + (RANDOM() * INTERVAL '90 days'))::DATE,
  20.00,  -- ₦20 daily subscription
  1,
  10,
  'SUB-' || TO_CHAR(NOW(), 'YYYYMMDD') || '-' || LPAD(row_num::TEXT, 6, '0'),
  CASE 
    WHEN row_num % 100 < 75 THEN 'active'
    WHEN row_num % 100 < 90 THEN 'cancelled'
    ELSE 'expired'
  END,
  true,
  u.email,
  u.full_name,
  u.created_at + (RANDOM() * INTERVAL '90 days')
FROM (
  SELECT 
    u.*,
    ROW_NUMBER() OVER (ORDER BY u.id) as row_num
  FROM users u
  WHERE u.total_recharge_amount >= 1000000  -- Users who have recharged at least ₦10k
  LIMIT 200
) u;

\echo '✓ Created 200 daily subscriptions'

\echo 'Part 3 Complete: Affiliates and Subscriptions created'
\echo '========================================'

-- ============================================================================
-- FINAL STATISTICS & VERIFICATION
-- ============================================================================

\echo ''
\echo '========================================'
\echo 'PRODUCTION SIMULATION DATA - COMPLETE!'
\echo '========================================'
\echo ''

DO $$
DECLARE
  user_count INT;
  transaction_count INT;
  completed_transactions INT;
  spin_count INT;
  affiliate_count INT;
  subscription_count INT;
  total_revenue BIGINT;
BEGIN
  SELECT COUNT(*) INTO user_count FROM users;
  SELECT COUNT(*) INTO transaction_count FROM vtu_transactions;
  SELECT COUNT(*) INTO completed_transactions FROM vtu_transactions WHERE status = 'COMPLETED';
  SELECT COUNT(*) INTO spin_count FROM spin_results;
  SELECT COUNT(*) INTO affiliate_count FROM affiliates;
  SELECT COUNT(*) INTO subscription_count FROM daily_subscriptions WHERE status = 'active';
  SELECT COALESCE(SUM(amount), 0) INTO total_revenue FROM vtu_transactions WHERE status = 'COMPLETED';
  
  RAISE NOTICE 'Total Users: %', user_count;
  RAISE NOTICE 'Total Transactions: %', transaction_count;
  RAISE NOTICE 'Completed Transactions: %', completed_transactions;
  RAISE NOTICE 'Total Spin Results: %', spin_count;
  RAISE NOTICE 'Total Affiliates: %', affiliate_count;
  RAISE NOTICE 'Active Subscriptions: %', subscription_count;
  RAISE NOTICE 'Total Revenue: ₦%', (total_revenue / 100.0)::NUMERIC(12,2);
  RAISE NOTICE '';
  RAISE NOTICE 'Platform ready for testing and demo!';
END $$;

\echo ''
\echo '========================================'
\echo 'SEED DATA LOADED SUCCESSFULLY'
\echo '========================================'
\echo ''
\echo 'Next Steps:'
\echo '1. Restart backend server'
\echo '2. Test authentication flow'
\echo '3. Test recharge and spin wheel'
\echo '4. Test affiliate dashboard'
\echo '5. Test admin portal'
\echo ''
