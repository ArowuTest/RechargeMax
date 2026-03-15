-- ============================================================================
-- RECHARGEMAX PRODUCTION SIMULATION SEED DATA (V2 - 100% Schema Aligned)
-- ============================================================================
-- This file creates realistic production data perfectly aligned with actual schema
-- Target: 1,000 users, 5,000+ transactions for realistic testing and demos
-- Purpose: Demo, testing, performance validation
-- ============================================================================

\echo '========================================'
\echo 'RECHARGEMAX PRODUCTION SEED DATA V2'
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
    ELSE '807' || LPAD((1000000 + (i * 4567) % 9000000)::TEXT, 7, '0')
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
  msisdn,
  balance,
  pending_balance,
  total_earned,
  total_withdrawn,
  min_payout_amount,
  is_active,
  is_suspended,
  created_at,
  updated_at
)
SELECT 
  gen_random_uuid(),
  id,
  msisdn,
  0,  -- Start with zero balance
  0,
  0,
  0,
  100000,  -- ₦1000 minimum payout
  true,
  false,
  created_at,
  created_at
FROM users;

\echo '✓ Created 1,000 wallets'

-- ============================================================================
-- PART 3: TRANSACTIONS & VTU TRANSACTIONS (5,000+ realistic transactions)
-- ============================================================================

\echo 'Creating transactions and VTU transactions...'

-- First create parent transactions
WITH inserted_transactions AS (
  INSERT INTO transactions (
    id,
    user_id,
    msisdn,
    network_provider,
    recharge_type,
    amount,
    payment_method,
    payment_reference,
    status,
    points_earned,
    draw_entries,
    created_at,
    updated_at
  )
  SELECT 
    gen_random_uuid(),
    u.id,
    u.msisdn,
    (ARRAY['MTN', 'GLO', 'AIRTEL', 'NINE_MOBILE'])[((i-1) % 4) + 1],
    CASE WHEN i % 100 < 60 THEN 'AIRTIME' ELSE 'DATA' END,
    -- Amount in naira (not kobo for transactions table)
    CASE (i % 10)
      WHEN 0 THEN (500 + RANDOM() * 500)::NUMERIC(10,2)
      WHEN 1 THEN (1000 + RANDOM() * 1000)::NUMERIC(10,2)
      WHEN 2 THEN (2000 + RANDOM() * 3000)::NUMERIC(10,2)
      WHEN 3 THEN (5000 + RANDOM() * 5000)::NUMERIC(10,2)
      WHEN 4 THEN (10000 + RANDOM() * 10000)::NUMERIC(10,2)
      ELSE (1500 + RANDOM() * 3500)::NUMERIC(10,2)
    END,
    'WALLET',
    'TXN-' || TO_CHAR(NOW(), 'YYYYMMDD') || '-' || LPAD((ROW_NUMBER() OVER ())::TEXT, 10, '0'),
    CASE 
      WHEN i % 100 < 90 THEN 'SUCCESS'
      WHEN i % 100 < 95 THEN 'PENDING'
      ELSE 'FAILED'
    END,
    CASE WHEN i % 100 < 90 THEN (10 + RANDOM() * 50)::INT ELSE 0 END,
    CASE WHEN i % 100 < 90 THEN 1 ELSE 0 END,
    u.created_at + (RANDOM() * (NOW() - u.created_at)),
    u.created_at + (RANDOM() * (NOW() - u.created_at)) + INTERVAL '10 seconds'
  FROM users u
  CROSS JOIN generate_series(1, 5) AS i
  RETURNING id, user_id, msisdn, network_provider, recharge_type, amount, payment_reference, status, created_at
)
-- Then create corresponding VTU transactions
INSERT INTO vtu_transactions (
  id,
  transaction_reference,
  parent_transaction_id,
  user_id,
  phone_number,
  network_provider,
  recharge_type,
  amount,
  provider_used,
  provider_transaction_id,
  provider_reference,
  provider_response,
  provider_status,
  status,
  retry_count,
  max_retries,
  is_reconciled,
  created_at,
  processing_started_at,
  completed_at
)
SELECT 
  gen_random_uuid(),
  t.payment_reference,
  t.id,  -- parent_transaction_id
  t.user_id,
  t.msisdn,
  t.network_provider,
  t.recharge_type,
  (t.amount * 100)::BIGINT,  -- Convert to kobo
  'simulation',
  CASE WHEN t.status = 'SUCCESS' THEN 'SIM-' || SUBSTRING(t.payment_reference FROM 14) ELSE NULL END,
  CASE WHEN t.status = 'SUCCESS' THEN 'VTU-' || SUBSTRING(t.payment_reference FROM 14) ELSE NULL END,
  CASE WHEN t.status = 'SUCCESS' THEN '{"status": "success", "message": "Recharge successful"}'::jsonb ELSE NULL END,
  CASE WHEN t.status = 'SUCCESS' THEN 'success' ELSE 'failed' END,
  CASE WHEN t.status = 'SUCCESS' THEN 'COMPLETED' ELSE t.status END,  -- Map SUCCESS to COMPLETED for VTU
  0,
  3,
  t.status = 'SUCCESS',
  t.created_at,
  t.created_at + INTERVAL '2 seconds',
  CASE WHEN t.status = 'SUCCESS' THEN t.created_at + INTERVAL '10 seconds' ELSE NULL END
FROM inserted_transactions t;

\echo '✓ Created 5,000 transactions and VTU transactions'

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
  icon_name,
  color_scheme,
  is_active,
  sort_order,
  description,
  created_at,
  updated_at
) VALUES
  (gen_random_uuid(), 'Better Luck Next Time', 'POINTS', 0, 40.0, 'sad', 'gray', true, 1, 'Try again next time!', NOW(), NOW()),
  (gen_random_uuid(), '₦100 Airtime', 'AIRTIME', 10000, 25.0, 'phone', 'blue', true, 2, 'Win ₦100 airtime', NOW(), NOW()),
  (gen_random_uuid(), '₦200 Airtime', 'AIRTIME', 20000, 15.0, 'phone', 'green', true, 3, 'Win ₦200 airtime', NOW(), NOW()),
  (gen_random_uuid(), '₦500 Airtime', 'AIRTIME', 50000, 10.0, 'money', 'yellow', true, 4, 'Win ₦500 airtime', NOW(), NOW()),
  (gen_random_uuid(), '₦1000 Airtime', 'AIRTIME', 100000, 5.0, 'diamond', 'purple', true, 5, 'Win ₦1000 airtime', NOW(), NOW()),
  (gen_random_uuid(), '100 Bonus Points', 'POINTS', 100, 3.0, 'star', 'orange', true, 6, 'Win 100 bonus points', NOW(), NOW()),
  (gen_random_uuid(), 'iPhone 15 Pro', 'CASH', 150000000, 0.5, 'phone', 'red', true, 7, 'Win an iPhone 15 Pro!', NOW(), NOW()),
  (gen_random_uuid(), '₦2000 Airtime', 'AIRTIME', 200000, 1.0, 'trophy', 'gold', true, 8, 'Win ₦2000 airtime', NOW(), NOW());

\echo '✓ Created 8 wheel prizes'

-- ============================================================================
-- PART 5: SPIN RESULTS (For transactions >= ₦1000)
-- ============================================================================

\echo 'Creating spin results for eligible transactions...'

-- Create spin results for all eligible transactions
WITH eligible_transactions AS (
  SELECT 
    parent_transaction_id,
    user_id,
    phone_number,
    completed_at,
    ROW_NUMBER() OVER (ORDER BY completed_at) as rn
  FROM vtu_transactions
  WHERE status = 'COMPLETED' 
    AND amount >= 100000
    AND completed_at IS NOT NULL
    AND parent_transaction_id IS NOT NULL
),
prizes_with_weights AS (
  SELECT 
    id,
    prize_name,
    prize_type,
    prize_value,
    probability,
    ROW_NUMBER() OVER (ORDER BY id) as prize_num
  FROM wheel_prizes
  WHERE is_active = true
)
INSERT INTO spin_results (
  id,
  user_id,
  transaction_id,
  msisdn,
  prize_id,
  prize_name,
  prize_type,
  prize_value,
  claim_status,
  claimed_at,
  created_at
)
SELECT 
  gen_random_uuid(),
  et.user_id,
  et.parent_transaction_id,  -- Use parent transaction ID
  et.phone_number,
  p.id,
  p.prize_name,
  p.prize_type,
  p.prize_value,
  'CLAIMED',
  et.completed_at + INTERVAL '30 seconds',
  et.completed_at + INTERVAL '30 seconds'
FROM eligible_transactions et
CROSS JOIN LATERAL (
  SELECT id, prize_name, prize_type, prize_value
  FROM prizes_with_weights
  ORDER BY RANDOM()
  LIMIT 1
) p;

\echo '✓ Created spin results for eligible transactions'

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
  active_referrals,
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
  (3 + RANDOM() * 15)::INT,
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
