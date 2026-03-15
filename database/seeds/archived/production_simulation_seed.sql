-- ============================================================================
-- RECHARGEMAX PRODUCTION SIMULATION SEED DATA
-- ============================================================================
-- This file creates a realistic production environment with comprehensive data
-- Target: 10,000 users, 50,000+ transactions, full gamification data
-- Purpose: Demo, testing, performance validation, staff training
-- ============================================================================

-- Clean existing transactional data (keep reference data)
TRUNCATE TABLE 
  wheel_spins,
  vtu_transactions,
  affiliate_commissions,
  affiliate_referrals,
  daily_subscriptions,
  wallet_transactions,
  wallets,
  users
CASCADE;

-- ============================================================================
-- PART 1: USERS (10,000 realistic Nigerian users)
-- ============================================================================

-- Helper function to generate realistic Nigerian phone numbers
CREATE OR REPLACE FUNCTION generate_nigerian_phone(network_code TEXT, index_num INT) 
RETURNS TEXT AS $$
DECLARE
  prefix TEXT;
  suffix TEXT;
BEGIN
  -- Network prefixes (MTN, GLO, AIRTEL, 9MOBILE)
  prefix := CASE network_code
    WHEN 'MTN' THEN (ARRAY['0803', '0806', '0810', '0813', '0814', '0816', '0903', '0906'])[((index_num % 8) + 1)]
    WHEN 'GLO' THEN (ARRAY['0805', '0807', '0811', '0815', '0905'])[((index_num % 5) + 1)]
    WHEN 'AIRTEL' THEN (ARRAY['0802', '0808', '0812', '0901', '0902'])[((index_num % 5) + 1)]
    WHEN '9MOBILE' THEN (ARRAY['0809', '0817', '0818', '0908', '0909'])[((index_num % 5) + 1)]
    ELSE '0803'
  END;
  
  -- Generate unique 7-digit suffix
  suffix := LPAD((1000000 + index_num)::TEXT, 7, '0');
  
  RETURN prefix || suffix;
END;
$$ LANGUAGE plpgsql;

-- Generate 10,000 users with realistic distribution
DO $$
DECLARE
  i INT;
  network_code TEXT;
  phone_number TEXT;
  user_id UUID;
  registration_date TIMESTAMP;
  total_recharges INT;
  total_amount INT;
  total_points INT;
  loyalty_tier TEXT;
BEGIN
  FOR i IN 1..10000 LOOP
    -- Network distribution: MTN 40%, GLO 25%, AIRTEL 25%, 9MOBILE 10%
    network_code := CASE 
      WHEN i % 100 < 40 THEN 'MTN'
      WHEN i % 100 < 65 THEN 'GLO'
      WHEN i % 100 < 90 THEN 'AIRTEL'
      ELSE '9MOBILE'
    END;
    
    phone_number := generate_nigerian_phone(network_code, i);
    user_id := gen_random_uuid();
    
    -- Registration dates spread over past 180 days (6 months)
    registration_date := NOW() - (RANDOM() * INTERVAL '180 days');
    
    -- Realistic user activity patterns
    -- 20% power users, 50% regular users, 30% light users
    IF i % 100 < 20 THEN
      -- Power users: 20-50 recharges, ₦50k-₦200k spent
      total_recharges := (20 + RANDOM() * 30)::INT;
      total_amount := (5000000 + RANDOM() * 15000000)::INT; -- 50k-200k Naira in kobo
      total_points := total_amount / 20000; -- ₦200 = 1 point
      loyalty_tier := CASE 
        WHEN total_amount >= 10000000 THEN 'platinum'
        WHEN total_amount >= 5000000 THEN 'gold'
        ELSE 'silver'
      END;
    ELSIF i % 100 < 70 THEN
      -- Regular users: 5-20 recharges, ₦10k-₦50k spent
      total_recharges := (5 + RANDOM() * 15)::INT;
      total_amount := (1000000 + RANDOM() * 4000000)::INT; -- 10k-50k Naira in kobo
      total_points := total_amount / 20000;
      loyalty_tier := CASE 
        WHEN total_amount >= 3000000 THEN 'silver'
        ELSE 'bronze'
      END;
    ELSE
      -- Light users: 1-5 recharges, ₦1k-₦10k spent
      total_recharges := (1 + RANDOM() * 4)::INT;
      total_amount := (100000 + RANDOM() * 900000)::INT; -- 1k-10k Naira in kobo
      total_points := total_amount / 20000;
      loyalty_tier := 'bronze';
    END IF;
    
    INSERT INTO users (
      id,
      msisdn,
      email,
      first_name,
      last_name,
      gender,
      date_of_birth,
      network_provider,
      total_points,
      total_recharge_amount,
      loyalty_tier,
      referral_code,
      is_active,
      is_verified,
      created_at,
      updated_at
    ) VALUES (
      user_id,
      phone_number,
      'user' || i || '@rechargemax.ng',
      'User',
      'Test' || i,
      (ARRAY['male', 'female', ''])[FLOOR(RANDOM() * 3 + 1)::INT],
      NOW() - (RANDOM() * INTERVAL '18250 days' + INTERVAL '6570 days'), -- Age 18-68
      network_code,
      total_points,
      total_amount,
      loyalty_tier,
      CASE WHEN i % 10 = 0 THEN 'REF' || LPAD(i::TEXT, 6, '0') ELSE '' END, -- 10% have referral codes
      true,
      i % 100 < 80, -- 80% verified
      registration_date,
      registration_date + (RANDOM() * INTERVAL '30 days')
    );
    
    -- Create wallet for each user
    INSERT INTO wallets (
      id,
      user_id,
      balance,
      currency,
      created_at,
      updated_at
    ) VALUES (
      gen_random_uuid(),
      user_id,
      0, -- Start with zero balance
      'NGN',
      registration_date,
      registration_date
    );
    
    -- Progress indicator
    IF i % 1000 = 0 THEN
      RAISE NOTICE 'Created % users...', i;
    END IF;
  END LOOP;
  
  RAISE NOTICE 'Successfully created 10,000 users!';
END $$;

-- Drop helper function
DROP FUNCTION IF EXISTS generate_nigerian_phone(TEXT, INT);

-- ============================================================================
-- VERIFICATION & STATISTICS
-- ============================================================================

-- Show user distribution by network
SELECT 
  network_provider,
  COUNT(*) as user_count,
  ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM users), 2) as percentage
FROM users
GROUP BY network_provider
ORDER BY user_count DESC;

-- Show user distribution by loyalty tier
SELECT 
  loyalty_tier,
  COUNT(*) as user_count,
  ROUND(AVG(total_recharge_amount) / 100.0, 2) as avg_spent_naira,
  ROUND(AVG(total_points), 0) as avg_points
FROM users
GROUP BY loyalty_tier
ORDER BY 
  CASE loyalty_tier
    WHEN 'platinum' THEN 1
    WHEN 'gold' THEN 2
    WHEN 'silver' THEN 3
    WHEN 'bronze' THEN 4
  END;

RAISE NOTICE 'Part 1 Complete: 10,000 users created with realistic profiles';


-- ============================================================================
-- PART 2: RECHARGE TRANSACTIONS (50,000+ realistic transactions)
-- ============================================================================

DO $$
DECLARE
  user_record RECORD;
  transaction_count INT;
  i INT;
  transaction_date TIMESTAMP;
  recharge_amount INT;
  recharge_type TEXT;
  recharge_status TEXT;
  payment_ref TEXT;
  transaction_id UUID;
BEGIN
  RAISE NOTICE 'Starting transaction generation for 10,000 users...';
  
  FOR user_record IN 
    SELECT id, msisdn, network_provider, total_recharge_amount, created_at 
    FROM users 
    ORDER BY created_at
  LOOP
    -- Calculate number of transactions based on total amount
    -- Average transaction: ₦10,000 (1,000,000 kobo)
    transaction_count := GREATEST(1, (user_record.total_recharge_amount / 1000000)::INT);
    
    FOR i IN 1..transaction_count LOOP
      transaction_id := gen_random_uuid();
      
      -- Transaction dates between user registration and now
      transaction_date := user_record.created_at + (RANDOM() * (NOW() - user_record.created_at));
      
      -- Realistic recharge amounts (in kobo)
      -- Distribution: 30% small (₦500-₦2k), 50% medium (₦2k-₦10k), 20% large (₦10k-₦50k)
      IF i % 100 < 30 THEN
        recharge_amount := (50000 + RANDOM() * 150000)::INT; -- ₦500-₦2,000
      ELSIF i % 100 < 80 THEN
        recharge_amount := (200000 + RANDOM() * 800000)::INT; -- ₦2,000-₦10,000
      ELSE
        recharge_amount := (1000000 + RANDOM() * 4000000)::INT; -- ₦10,000-₦50,000
      END IF;
      
      -- Recharge type distribution: 60% airtime, 40% data
      recharge_type := CASE WHEN i % 100 < 60 THEN 'airtime' ELSE 'data' END;
      
      -- Status distribution: 95% completed, 3% pending, 2% failed
      recharge_status := CASE 
        WHEN i % 100 < 95 THEN 'COMPLETED'
        WHEN i % 100 < 98 THEN 'pending'
        ELSE 'failed'
      END;
      
      -- Generate unique payment reference
      payment_ref := 'PAY-' || TO_CHAR(transaction_date, 'YYYYMMDD') || '-' || 
                     LPAD((i * 1000 + EXTRACT(EPOCH FROM transaction_date)::INT % 1000)::TEXT, 8, '0');
      
      INSERT INTO vtu_transactions (
        id,
        user_id,
        msisdn,
        network_provider,
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
      ) VALUES (
        transaction_id,
        user_record.id,
        user_record.msisdn,
        user_record.network_provider,
        recharge_amount,
        recharge_type,
        recharge_status,
        'paystack',
        payment_ref,
        'paystack',
        CASE WHEN recharge_status = 'COMPLETED' THEN 'VTU-' || payment_ref ELSE NULL END,
        CASE WHEN recharge_status = 'COMPLETED' THEN '{"status": "success", "message": "Recharge successful"}' ELSE NULL END,
        transaction_date,
        transaction_date + INTERVAL '5 seconds',
        CASE WHEN recharge_status = 'COMPLETED' THEN transaction_date + INTERVAL '10 seconds' ELSE NULL END
      );
      
    END LOOP;
    
  END LOOP;
  
  RAISE NOTICE 'Transaction generation complete!';
END $$;

-- ============================================================================
-- TRANSACTION STATISTICS
-- ============================================================================

-- Total transactions
SELECT 
  'Total Transactions' as metric,
  COUNT(*) as count,
  ROUND(SUM(amount) / 100.0, 2) as total_naira,
  ROUND(AVG(amount) / 100.0, 2) as avg_naira
FROM vtu_transactions;

-- Transactions by status
SELECT 
  status,
  COUNT(*) as count,
  ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM vtu_transactions), 2) as percentage,
  ROUND(SUM(amount) / 100.0, 2) as total_naira
FROM vtu_transactions
GROUP BY status
ORDER BY count DESC;

-- Transactions by type
SELECT 
  recharge_type,
  COUNT(*) as count,
  ROUND(AVG(amount) / 100.0, 2) as avg_amount_naira
FROM vtu_transactions
GROUP BY recharge_type;

-- Transactions by network
SELECT 
  network_provider,
  COUNT(*) as count,
  ROUND(SUM(amount) / 100.0, 2) as total_naira
FROM vtu_transactions
GROUP BY network_provider
ORDER BY count DESC;

-- Daily transaction volume (last 30 days)
SELECT 
  DATE(created_at) as transaction_date,
  COUNT(*) as transactions,
  ROUND(SUM(amount) / 100.0, 2) as total_naira
FROM vtu_transactions
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY transaction_date DESC
LIMIT 30;

RAISE NOTICE 'Part 2 Complete: 50,000+ recharge transactions created';


-- ============================================================================
-- PART 3: GAMIFICATION - WHEEL PRIZES & SPIN RESULTS
-- ============================================================================

-- First, ensure wheel prizes exist (these should already be in the database from migrations)
-- If not, create them here

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
  (gen_random_uuid(), '500 Bonus Points', 'points', 500, 1.5, '🌟', '#gold', true, 500, 0, NOW(), NOW()),
  (gen_random_uuid(), 'iPhone 15 Pro', 'physical', 150000000, 0.3, '📱', '#red', true, 10, 0, NOW(), NOW()),
  (gen_random_uuid(), 'Samsung Galaxy S24', 'physical', 120000000, 0.2, '📱', '#red', true, 10, 0, NOW(), NOW())
ON CONFLICT (prize_name) DO UPDATE SET
  probability = EXCLUDED.probability,
  is_active = EXCLUDED.is_active,
  updated_at = NOW();

-- Generate wheel spin results for eligible users (recharges >= ₦1000)
DO $$
DECLARE
  transaction_record RECORD;
  prize_record RECORD;
  random_val FLOAT;
  cumulative_prob FLOAT;
  selected_prize_id UUID;
  spin_id UUID;
BEGIN
  RAISE NOTICE 'Generating wheel spin results for eligible transactions...';
  
  -- Get all completed transactions >= ₦1000 (100,000 kobo)
  FOR transaction_record IN 
    SELECT id, user_id, msisdn, amount, created_at, completed_at
    FROM vtu_transactions
    WHERE status = 'COMPLETED' 
      AND amount >= 100000
    ORDER BY created_at
  LOOP
    -- Select a prize based on probability
    random_val := RANDOM() * 100;
    cumulative_prob := 0;
    selected_prize_id := NULL;
    
    FOR prize_record IN 
      SELECT id, probability 
      FROM wheel_prizes 
      WHERE is_active = true 
      ORDER BY probability DESC
    LOOP
      cumulative_prob := cumulative_prob + prize_record.probability;
      IF random_val <= cumulative_prob THEN
        selected_prize_id := prize_record.id;
        EXIT;
      END IF;
    END LOOP;
    
    -- If no prize selected (shouldn't happen), select "Better Luck Next Time"
    IF selected_prize_id IS NULL THEN
      SELECT id INTO selected_prize_id 
      FROM wheel_prizes 
      WHERE prize_name = 'Better Luck Next Time' 
      LIMIT 1;
    END IF;
    
    -- Create spin result
    spin_id := gen_random_uuid();
    
    INSERT INTO wheel_spins (
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
      spin_id,
      transaction_record.user_id,
      transaction_record.id,
      selected_prize_id,
      wp.prize_name,
      wp.prize_type,
      wp.prize_value,
      'claimed',
      transaction_record.completed_at + INTERVAL '30 seconds',
      transaction_record.completed_at + INTERVAL '30 seconds',
      transaction_record.completed_at + INTERVAL '30 seconds'
    FROM wheel_prizes wp
    WHERE wp.id = selected_prize_id;
    
    -- Update prize claimed count
    UPDATE wheel_prizes
    SET total_claimed = total_claimed + 1
    WHERE id = selected_prize_id;
    
  END LOOP;
  
  RAISE NOTICE 'Wheel spin generation complete!';
END $$;

-- ============================================================================
-- GAMIFICATION STATISTICS
-- ============================================================================

-- Total spins
SELECT 
  'Total Spins' as metric,
  COUNT(*) as count
FROM wheel_spins;

-- Prize distribution
SELECT 
  prize_name,
  prize_type,
  COUNT(*) as times_won,
  ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM wheel_spins), 2) as percentage,
  ROUND(SUM(prize_value) / 100.0, 2) as total_value_naira
FROM wheel_spins
GROUP BY prize_name, prize_type
ORDER BY times_won DESC;

-- Top winners
SELECT 
  u.msisdn,
  COUNT(ws.id) as total_spins,
  COUNT(CASE WHEN ws.prize_type != 'none' THEN 1 END) as prizes_won,
  ROUND(SUM(ws.prize_value) / 100.0, 2) as total_prize_value_naira
FROM users u
JOIN wheel_spins ws ON u.id = ws.user_id
GROUP BY u.msisdn
ORDER BY total_prize_value_naira DESC
LIMIT 20;

RAISE NOTICE 'Part 3 Complete: Gamification data (wheel spins and prizes) created';


-- ============================================================================
-- PART 4: AFFILIATE NETWORK (500 affiliates with referral networks)
-- ============================================================================

DO $$
DECLARE
  user_record RECORD;
  affiliate_count INT := 0;
  target_affiliates INT := 500;
  affiliate_id UUID;
  referral_count INT;
  referred_user RECORD;
  commission_amount INT;
  total_commission INT;
BEGIN
  RAISE NOTICE 'Creating affiliate network...';
  
  -- Select 500 users to become affiliates (top users by activity)
  FOR user_record IN 
    SELECT id, msisdn, network_provider, total_recharge_amount, created_at
    FROM users
    WHERE referral_code != ''
    ORDER BY total_recharge_amount DESC
    LIMIT target_affiliates
  LOOP
    affiliate_id := gen_random_uuid();
    affiliate_count := affiliate_count + 1;
    
    -- Calculate total commission (5% of referrals' spending)
    total_commission := 0;
    
    -- Create affiliate record
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
    ) VALUES (
      affiliate_id,
      user_record.id,
      user_record.referral_code,
      'active',
      0, -- Will be updated below
      0, -- Will be updated below
      5.0, -- 5% commission rate
      (ARRAY['GTBank', 'Access Bank', 'First Bank', 'UBA', 'Zenith Bank'])[FLOOR(RANDOM() * 5 + 1)::INT],
      LPAD((1000000000 + RANDOM() * 999999999)::BIGINT::TEXT, 10, '0'),
      'Affiliate ' || affiliate_count,
      user_record.created_at,
      user_record.created_at + INTERVAL '1 day',
      user_record.created_at + INTERVAL '1 day'
    );
    
    -- Create 5-20 referrals for each affiliate
    referral_count := (5 + RANDOM() * 15)::INT;
    
    FOR referred_user IN 
      SELECT id, msisdn, total_recharge_amount, created_at
      FROM users
      WHERE id != user_record.id
        AND referral_code = ''
        AND created_at > user_record.created_at
      ORDER BY RANDOM()
      LIMIT referral_count
    LOOP
      -- Update referred user with referral code
      UPDATE users
      SET referral_code = user_record.referral_code
      WHERE id = referred_user.id;
      
      -- Create referral record
      INSERT INTO affiliate_referrals (
        id,
        affiliate_id,
        referred_user_id,
        referred_msisdn,
        referral_code,
        status,
        first_recharge_at,
        created_at,
        updated_at
      ) VALUES (
        gen_random_uuid(),
        affiliate_id,
        referred_user.id,
        referred_user.msisdn,
        user_record.referral_code,
        'active',
        referred_user.created_at + INTERVAL '1 hour',
        referred_user.created_at,
        referred_user.created_at
      );
      
      -- Calculate commission (5% of referred user's total spending)
      commission_amount := (referred_user.total_recharge_amount * 0.05)::INT;
      total_commission := total_commission + commission_amount;
      
      -- Create commission record for each recharge by referred user
      FOR i IN 1..(referred_user.total_recharge_amount / 1000000)::INT LOOP
        INSERT INTO affiliate_commissions (
          id,
          affiliate_id,
          user_id,
          transaction_id,
          commission_amount,
          commission_rate,
          transaction_amount,
          status,
          created_at,
          updated_at,
          paid_at
        )
        SELECT 
          gen_random_uuid(),
          affiliate_id,
          referred_user.id,
          vt.id,
          (vt.amount * 0.05)::INT,
          5.0,
          vt.amount,
          'paid',
          vt.created_at,
          vt.created_at + INTERVAL '7 days',
          vt.created_at + INTERVAL '7 days'
        FROM vtu_transactions vt
        WHERE vt.user_id = referred_user.id
          AND vt.status = 'COMPLETED'
        ORDER BY vt.created_at
        LIMIT 1
        OFFSET i - 1;
      END LOOP;
      
    END LOOP;
    
    -- Update affiliate with totals
    UPDATE affiliates
    SET 
      total_referrals = referral_count,
      total_commission = total_commission
    WHERE id = affiliate_id;
    
    IF affiliate_count % 100 = 0 THEN
      RAISE NOTICE 'Created % affiliates...', affiliate_count;
    END IF;
    
  END LOOP;
  
  RAISE NOTICE 'Affiliate network creation complete!';
END $$;

-- ============================================================================
-- AFFILIATE STATISTICS
-- ============================================================================

-- Total affiliates
SELECT 
  'Total Affiliates' as metric,
  COUNT(*) as count,
  SUM(total_referrals) as total_referrals,
  ROUND(SUM(total_commission) / 100.0, 2) as total_commission_naira
FROM affiliates;

-- Affiliate performance tiers
SELECT 
  CASE 
    WHEN total_referrals >= 15 THEN 'Top Performer (15+)'
    WHEN total_referrals >= 10 THEN 'High Performer (10-14)'
    WHEN total_referrals >= 5 THEN 'Average Performer (5-9)'
    ELSE 'New Affiliate (< 5)'
  END as performance_tier,
  COUNT(*) as affiliate_count,
  ROUND(AVG(total_referrals), 1) as avg_referrals,
  ROUND(AVG(total_commission) / 100.0, 2) as avg_commission_naira
FROM affiliates
GROUP BY performance_tier
ORDER BY avg_referrals DESC;

-- Top 20 affiliates by commission
SELECT 
  a.affiliate_code,
  u.msisdn,
  a.total_referrals,
  ROUND(a.total_commission / 100.0, 2) as commission_naira,
  a.status
FROM affiliates a
JOIN users u ON a.user_id = u.id
ORDER BY a.total_commission DESC
LIMIT 20;

RAISE NOTICE 'Part 4 Complete: Affiliate network with 500 affiliates created';


-- ============================================================================
-- PART 5: DAILY SUBSCRIPTIONS (1000 active subscribers)
-- ============================================================================

DO $$
DECLARE
  user_record RECORD;
  subscription_count INT := 0;
  target_subscriptions INT := 1000;
  subscription_id UUID;
  subscription_start DATE;
  subscription_status TEXT;
  days_active INT;
BEGIN
  RAISE NOTICE 'Creating daily subscriptions...';
  
  -- Select 1000 users for daily subscriptions (regular and power users)
  FOR user_record IN 
    SELECT id, msisdn, network_provider, total_recharge_amount, created_at
    FROM users
    WHERE total_recharge_amount >= 1000000 -- Users who spent at least ₦10,000
    ORDER BY RANDOM()
    LIMIT target_subscriptions
  LOOP
    subscription_id := gen_random_uuid();
    subscription_count := subscription_count + 1;
    
    -- Subscription start date (between user registration and now)
    subscription_start := (user_record.created_at + (RANDOM() * (NOW() - user_record.created_at)))::DATE;
    
    -- Days since subscription started
    days_active := (NOW()::DATE - subscription_start)::INT;
    
    -- Status distribution: 80% active, 15% cancelled, 5% expired
    subscription_status := CASE 
      WHEN subscription_count % 100 < 80 THEN 'active'
      WHEN subscription_count % 100 < 95 THEN 'cancelled'
      ELSE 'expired'
    END;
    
    INSERT INTO daily_subscriptions (
      id,
      user_id,
      msisdn,
      network_provider,
      subscription_amount,
      status,
      start_date,
      end_date,
      next_billing_date,
      total_days_subscribed,
      payment_method,
      payment_reference,
      created_at,
      updated_at,
      cancelled_at
    ) VALUES (
      subscription_id,
      user_record.id,
      user_record.msisdn,
      user_record.network_provider,
      20000, -- ₦200 per day
      subscription_status,
      subscription_start,
      CASE 
        WHEN subscription_status = 'active' THEN NULL
        ELSE subscription_start + (days_active || ' days')::INTERVAL
      END,
      CASE 
        WHEN subscription_status = 'active' THEN NOW()::DATE + 1
        ELSE NULL
      END,
      days_active,
      'paystack',
      'SUB-' || TO_CHAR(subscription_start, 'YYYYMMDD') || '-' || LPAD(subscription_count::TEXT, 6, '0'),
      subscription_start::TIMESTAMP,
      NOW(),
      CASE 
        WHEN subscription_status = 'cancelled' THEN subscription_start + (days_active || ' days')::INTERVAL
        ELSE NULL
      END
    );
    
    IF subscription_count % 200 = 0 THEN
      RAISE NOTICE 'Created % subscriptions...', subscription_count;
    END IF;
    
  END LOOP;
  
  RAISE NOTICE 'Daily subscription creation complete!';
END $$;

-- ============================================================================
-- SUBSCRIPTION STATISTICS
-- ============================================================================

-- Total subscriptions
SELECT 
  'Total Subscriptions' as metric,
  COUNT(*) as count,
  SUM(total_days_subscribed) as total_days,
  ROUND(SUM(total_days_subscribed * subscription_amount) / 100.0, 2) as total_revenue_naira
FROM daily_subscriptions;

-- Subscriptions by status
SELECT 
  status,
  COUNT(*) as count,
  ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM daily_subscriptions), 2) as percentage,
  ROUND(AVG(total_days_subscribed), 1) as avg_days_subscribed
FROM daily_subscriptions
GROUP BY status
ORDER BY count DESC;

-- Subscription revenue by network
SELECT 
  network_provider,
  COUNT(*) as subscribers,
  ROUND(SUM(total_days_subscribed * subscription_amount) / 100.0, 2) as total_revenue_naira
FROM daily_subscriptions
GROUP BY network_provider
ORDER BY total_revenue_naira DESC;

RAISE NOTICE 'Part 5 Complete: 1000 daily subscriptions created';

-- ============================================================================
-- FINAL COMPREHENSIVE STATISTICS
-- ============================================================================

RAISE NOTICE '========================================';
RAISE NOTICE 'PRODUCTION SIMULATION DATA - SUMMARY';
RAISE NOTICE '========================================';

-- Overall platform statistics
DO $$
DECLARE
  total_users INT;
  total_transactions INT;
  total_revenue NUMERIC;
  total_spins INT;
  total_affiliates INT;
  total_subscriptions INT;
BEGIN
  SELECT COUNT(*) INTO total_users FROM users;
  SELECT COUNT(*) INTO total_transactions FROM vtu_transactions WHERE status = 'COMPLETED';
  SELECT ROUND(SUM(amount) / 100.0, 2) INTO total_revenue FROM vtu_transactions WHERE status = 'COMPLETED';
  SELECT COUNT(*) INTO total_spins FROM wheel_spins;
  SELECT COUNT(*) INTO total_affiliates FROM affiliates;
  SELECT COUNT(*) INTO total_subscriptions FROM daily_subscriptions WHERE status = 'active';
  
  RAISE NOTICE 'Total Users: %', total_users;
  RAISE NOTICE 'Total Completed Transactions: %', total_transactions;
  RAISE NOTICE 'Total Revenue: ₦%', total_revenue;
  RAISE NOTICE 'Total Wheel Spins: %', total_spins;
  RAISE NOTICE 'Total Active Affiliates: %', total_affiliates;
  RAISE NOTICE 'Total Active Subscriptions: %', total_subscriptions;
  RAISE NOTICE '========================================';
END $$;

-- User engagement metrics
SELECT 
  'User Engagement' as category,
  ROUND(AVG(total_recharge_amount) / 100.0, 2) as avg_lifetime_value_naira,
  ROUND(STDDEV(total_recharge_amount) / 100.0, 2) as stddev_ltv_naira,
  MAX(total_points) as max_points,
  ROUND(AVG(total_points), 0) as avg_points
FROM users;

-- Transaction velocity (last 30 days)
SELECT 
  'Last 30 Days' as period,
  COUNT(*) as transactions,
  ROUND(SUM(amount) / 100.0, 2) as revenue_naira,
  ROUND(AVG(amount) / 100.0, 2) as avg_transaction_naira
FROM vtu_transactions
WHERE created_at >= NOW() - INTERVAL '30 days'
  AND status = 'COMPLETED';

-- Gamification engagement
SELECT 
  'Gamification' as category,
  COUNT(DISTINCT user_id) as unique_spinners,
  COUNT(*) as total_spins,
  ROUND(COUNT(*) * 1.0 / COUNT(DISTINCT user_id), 2) as avg_spins_per_user,
  ROUND(SUM(CASE WHEN prize_type != 'none' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as win_rate_percentage
FROM wheel_spins;

RAISE NOTICE '========================================';
RAISE NOTICE 'SEED DATA GENERATION COMPLETE!';
RAISE NOTICE 'Platform is ready for testing and demo';
RAISE NOTICE '========================================';
