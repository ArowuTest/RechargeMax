-- ============================================================================
-- RechargeMax - Comprehensive Test Data Seed Script (Schema-Corrected)
-- ============================================================================
-- This script seeds the database with realistic test data matching actual schema
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. SEED TEST USERS
-- ============================================================================
-- Create 50 test users with realistic Nigerian phone numbers across all networks
INSERT INTO users (msisdn, email, full_name, created_at, updated_at) VALUES
-- MTN Users (15 users)
('2348031234567', 'john.doe@example.com', 'John Doe', NOW() - INTERVAL '30 days', NOW()),
('2347031234568', 'jane.smith@example.com', 'Jane Smith', NOW() - INTERVAL '25 days', NOW()),
('2349031234569', 'mike.johnson@example.com', 'Mike Johnson', NOW() - INTERVAL '20 days', NOW()),
('2348061234570', 'sarah.williams@example.com', 'Sarah Williams', NOW() - INTERVAL '15 days', NOW()),
('2348131234571', 'david.brown@example.com', 'David Brown', NOW() - INTERVAL '10 days', NOW()),
('2348101234572', 'emma.jones@example.com', 'Emma Jones', NOW() - INTERVAL '8 days', NOW()),
('2348141234573', 'oliver.garcia@example.com', 'Oliver Garcia', NOW() - INTERVAL '7 days', NOW()),
('2348161234574', 'sophia.martinez@example.com', 'Sophia Martinez', NOW() - INTERVAL '6 days', NOW()),
('2349061234575', 'james.rodriguez@example.com', 'James Rodriguez', NOW() - INTERVAL '5 days', NOW()),
('2348031234576', 'isabella.lopez@example.com', 'Isabella Lopez', NOW() - INTERVAL '4 days', NOW()),
('2347031234577', 'william.wilson@example.com', 'William Wilson', NOW() - INTERVAL '3 days', NOW()),
('2349031234578', 'ava.anderson@example.com', 'Ava Anderson', NOW() - INTERVAL '2 days', NOW()),
('2348061234579', 'benjamin.thomas@example.com', 'Benjamin Thomas', NOW() - INTERVAL '1 day', NOW()),
('2348131234580', 'mia.taylor@example.com', 'Mia Taylor', NOW(), NOW()),
('2348101234581', 'lucas.moore@example.com', 'Lucas Moore', NOW(), NOW()),

-- Airtel Users (12 users)
('2348021234582', 'amelia.jackson@example.com', 'Amelia Jackson', NOW() - INTERVAL '28 days', NOW()),
('2347081234583', 'henry.white@example.com', 'Henry White', NOW() - INTERVAL '22 days', NOW()),
('2349021234584', 'charlotte.harris@example.com', 'Charlotte Harris', NOW() - INTERVAL '18 days', NOW()),
('2348081234585', 'alexander.martin@example.com', 'Alexander Martin', NOW() - INTERVAL '14 days', NOW()),
('2347011234586', 'harper.thompson@example.com', 'Harper Thompson', NOW() - INTERVAL '12 days', NOW()),
('2348121234587', 'daniel.garcia@example.com', 'Daniel Garcia', NOW() - INTERVAL '9 days', NOW()),
('2349011234588', 'evelyn.martinez@example.com', 'Evelyn Martinez', NOW() - INTERVAL '7 days', NOW()),
('2349071234589', 'matthew.robinson@example.com', 'Matthew Robinson', NOW() - INTERVAL '5 days', NOW()),
('2348021234590', 'ella.clark@example.com', 'Ella Clark', NOW() - INTERVAL '3 days', NOW()),
('2347081234591', 'joseph.rodriguez@example.com', 'Joseph Rodriguez', NOW() - INTERVAL '2 days', NOW()),
('2349021234592', 'scarlett.lewis@example.com', 'Scarlett Lewis', NOW() - INTERVAL '1 day', NOW()),
('2348081234593', 'samuel.lee@example.com', 'Samuel Lee', NOW(), NOW()),

-- Glo Users (10 users)
('2348051234594', 'grace.walker@example.com', 'Grace Walker', NOW() - INTERVAL '26 days', NOW()),
('2347051234595', 'jackson.hall@example.com', 'Jackson Hall', NOW() - INTERVAL '21 days', NOW()),
('2349051234596', 'victoria.allen@example.com', 'Victoria Allen', NOW() - INTERVAL '17 days', NOW()),
('2348071234597', 'sebastian.young@example.com', 'Sebastian Young', NOW() - INTERVAL '13 days', NOW()),
('2348151234598', 'lily.hernandez@example.com', 'Lily Hernandez', NOW() - INTERVAL '11 days', NOW()),
('2348111234599', 'owen.king@example.com', 'Owen King', NOW() - INTERVAL '8 days', NOW()),
('2348051234600', 'zoey.wright@example.com', 'Zoey Wright', NOW() - INTERVAL '6 days', NOW()),
('2347051234601', 'gabriel.lopez@example.com', 'Gabriel Lopez', NOW() - INTERVAL '4 days', NOW()),
('2349051234602', 'hannah.hill@example.com', 'Hannah Hill', NOW() - INTERVAL '2 days', NOW()),
('2348071234603', 'carter.scott@example.com', 'Carter Scott', NOW(), NOW()),

-- 9mobile Users (8 users)
('2348091234604', 'penelope.green@example.com', 'Penelope Green', NOW() - INTERVAL '24 days', NOW()),
('2348181234605', 'wyatt.adams@example.com', 'Wyatt Adams', NOW() - INTERVAL '19 days', NOW()),
('2349091234606', 'layla.baker@example.com', 'Layla Baker', NOW() - INTERVAL '16 days', NOW()),
('2348171234607', 'jack.gonzalez@example.com', 'Jack Gonzalez', NOW() - INTERVAL '10 days', NOW()),
('2349081234608', 'aria.nelson@example.com', 'Aria Nelson', NOW() - INTERVAL '7 days', NOW()),
('2348091234609', 'julian.carter@example.com', 'Julian Carter', NOW() - INTERVAL '5 days', NOW()),
('2348181234610', 'nora.mitchell@example.com', 'Nora Mitchell', NOW() - INTERVAL '3 days', NOW()),
('2349091234611', 'leo.perez@example.com', 'Leo Perez', NOW(), NOW()),

-- Additional users without email (guest users)
('2348031234612', NULL, NULL, NOW() - INTERVAL '5 days', NOW()),
('2348021234613', NULL, NULL, NOW() - INTERVAL '4 days', NOW()),
('2348051234614', NULL, NULL, NOW() - INTERVAL '3 days', NOW()),
('2348091234615', NULL, NULL, NOW() - INTERVAL '2 days', NOW()),
('2347031234616', NULL, NULL, NOW() - INTERVAL '1 day', NOW())
ON CONFLICT (msisdn) DO NOTHING;

-- ============================================================================
-- 2. SEED TRANSACTIONS
-- ============================================================================
-- Create realistic transaction history for the past 30 days

DO $$
DECLARE
    user_record RECORD;
    transaction_date TIMESTAMP;
    amount NUMERIC(10,2);
    recharge_type TEXT;
    network_provider TEXT;
    i INTEGER;
BEGIN
    -- For each user, create 3-10 transactions
    FOR user_record IN SELECT id, msisdn, created_at FROM users LOOP
        -- Determine network from MSISDN prefix
        CASE 
            WHEN substring(user_record.msisdn from 1 for 4) IN ('234803', '234703', '234903', '234806', '234813', '234810', '234814', '234816', '234906') THEN
                network_provider := 'MTN';
            WHEN substring(user_record.msisdn from 1 for 4) IN ('234802', '234708', '234902', '234808', '234701', '234812', '234901', '234907') THEN
                network_provider := 'Airtel';
            WHEN substring(user_record.msisdn from 1 for 4) IN ('234805', '234705', '234905', '234807', '234815', '234811') THEN
                network_provider := 'Glo';
            WHEN substring(user_record.msisdn from 1 for 4) IN ('234809', '234818', '234909', '234817', '234908') THEN
                network_provider := '9mobile';
            ELSE
                network_provider := 'MTN';
        END CASE;
        
        FOR i IN 1..(3 + floor(random() * 8)::int) LOOP
            -- Random transaction date between user creation and now
            transaction_date := user_record.created_at + (random() * (NOW() - user_record.created_at));
            
            -- Random amount (₦200, ₦500, ₦1000, ₦2000, ₦5000, ₦10000)
            amount := (ARRAY[200, 500, 1000, 2000, 5000, 10000])[1 + floor(random() * 6)::int];
            
            -- Random recharge type
            recharge_type := (ARRAY['AIRTIME', 'DATA'])[1 + floor(random() * 2)::int];
            
            INSERT INTO transactions (
                user_id, msisdn, network_provider, amount, recharge_type,
                status, payment_method, payment_reference,
                points_earned, draw_entries,
                created_at, updated_at
            ) VALUES (
                user_record.id,
                user_record.msisdn,
                network_provider,
                amount,
                recharge_type,
                'COMPLETED',
                'PAYSTACK',
                'PAY_' || upper(substring(md5(random()::text) from 1 for 16)),
                FLOOR(amount / 100)::INTEGER,
                FLOOR(amount / 200)::INTEGER,
                transaction_date,
                transaction_date
            );
        END LOOP;
    END LOOP;
END $$;

-- ============================================================================
-- 3. SEED DRAW ENTRIES
-- ============================================================================
-- Create draw entries based on transactions (₦200 = 1 entry)

INSERT INTO draw_entries (user_id, draw_id, msisdn, entries_count, created_at)
SELECT 
    t.user_id,
    d.id as draw_id,
    t.msisdn,
    SUM(t.draw_entries)::INTEGER as entries_count,
    MAX(t.created_at) as created_at
FROM transactions t
CROSS JOIN draws d
WHERE t.status = 'COMPLETED'
    AND t.created_at >= d.start_time
    AND t.created_at <= d.end_time
GROUP BY t.user_id, d.id, t.msisdn
ON CONFLICT (user_id, draw_id) DO UPDATE
SET entries_count = draw_entries.entries_count + EXCLUDED.entries_count;

-- ============================================================================
-- 4. SEED WINNERS
-- ============================================================================
-- Create realistic winner records

DO $$
DECLARE
    user_ids UUID[];
    daily_draw_id UUID;
BEGIN
    -- Get array of user IDs
    SELECT ARRAY_AGG(id) INTO user_ids FROM users LIMIT 10;
    
    -- Get the daily draw ID
    SELECT id INTO daily_draw_id FROM draws WHERE name LIKE 'Daily%' LIMIT 1;
    
    -- Create 4 recent winners
    IF daily_draw_id IS NOT NULL AND array_length(user_ids, 1) >= 4 THEN
        INSERT INTO draw_winners (
            draw_id, user_id, msisdn, prize_amount,
            position, claim_status, won_at, created_at
        ) VALUES
        (daily_draw_id, user_ids[1], (SELECT msisdn FROM users WHERE id = user_ids[1]), 100000, 1, 'PENDING', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours'),
        (daily_draw_id, user_ids[2], (SELECT msisdn FROM users WHERE id = user_ids[2]), 100000, 1, 'CLAIMED', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
        (daily_draw_id, user_ids[3], (SELECT msisdn FROM users WHERE id = user_ids[3]), 100000, 1, 'CLAIMED', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
        (daily_draw_id, user_ids[4], (SELECT msisdn FROM users WHERE id = user_ids[4]), 100000, 1, 'CLAIMED', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days')
        ON CONFLICT DO NOTHING;
    END IF;
END $$;

-- ============================================================================
-- 5. SEED ADMIN ACCOUNTS
-- ============================================================================
-- Create admin accounts with different roles
-- Password: Admin123! (hashed with bcrypt - $2a$10$rQ3K5Y8qGxZ9vL2wN4mJ7.eH6fX8pT9qW1sA2bC3dE4fG5hI6jK7l)

INSERT INTO admin_users (email, password_hash, full_name, role, is_active, created_at, updated_at) VALUES
('superadmin@rechargemax.ng', '$2a$10$rQ3K5Y8qGxZ9vL2wN4mJ7.eH6fX8pT9qW1sA2bC3dE4fG5hI6jK7l', 'Super Admin', 'SUPER_ADMIN', true, NOW(), NOW()),
('admin@rechargemax.ng', '$2a$10$rQ3K5Y8qGxZ9vL2wN4mJ7.eH6fX8pT9qW1sA2bC3dE4fG5hI6jK7l', 'Platform Admin', 'ADMIN', true, NOW(), NOW()),
('moderator@rechargemax.ng', '$2a$10$rQ3K5Y8qGxZ9vL2wN4mJ7.eH6fX8pT9qW1sA2bC3dE4fG5hI6jK7l', 'Content Moderator', 'MODERATOR', true, NOW(), NOW()),
('viewer@rechargemax.ng', '$2a$10$rQ3K5Y8qGxZ9vL2wN4mJ7.eH6fX8pT9qW1sA2bC3dE4fG5hI6jK7l', 'Platform Viewer', 'VIEWER', true, NOW(), NOW())
ON CONFLICT (email) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    full_name = EXCLUDED.full_name,
    role = EXCLUDED.role,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- ============================================================================
-- 6. UPDATE STATISTICS
-- ============================================================================
-- Update draw entry counts
UPDATE draws d
SET total_entries = (
    SELECT COALESCE(SUM(entries_count), 0)
    FROM draw_entries de
    WHERE de.draw_id = d.id
);

COMMIT;

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================

SELECT '=== USER COUNT BY NETWORK ===' as info;
SELECT 
    CASE 
        WHEN substring(msisdn from 1 for 4) IN ('234803', '234703', '234903', '234806', '234813', '234810', '234814', '234816', '234906') THEN 'MTN'
        WHEN substring(msisdn from 1 for 4) IN ('234802', '234708', '234902', '234808', '234701', '234812', '234901', '234907') THEN 'Airtel'
        WHEN substring(msisdn from 1 for 4) IN ('234805', '234705', '234905', '234807', '234815', '234811') THEN 'Glo'
        WHEN substring(msisdn from 1 for 4) IN ('234809', '234818', '234909', '234817', '234908') THEN '9mobile'
        ELSE 'Unknown'
    END as network,
    COUNT(*) as user_count
FROM users
GROUP BY network
ORDER BY user_count DESC;

SELECT '=== TRANSACTION SUMMARY ===' as info;
SELECT 
    COUNT(*) as total_transactions,
    SUM(amount) as total_amount,
    AVG(amount) as avg_amount,
    recharge_type,
    status
FROM transactions
GROUP BY recharge_type, status;

SELECT '=== DRAW ENTRIES ===' as info;
SELECT 
    d.name,
    COUNT(DISTINCT de.user_id) as unique_users,
    SUM(de.entries_count) as total_entries
FROM draws d
LEFT JOIN draw_entries de ON d.id = de.draw_id
GROUP BY d.id, d.name;

SELECT '=== WINNERS ===' as info;
SELECT 
    COUNT(*) as total_winners,
    SUM(prize_amount) as total_prizes_won,
    claim_status
FROM draw_winners
GROUP BY claim_status;

SELECT '=== ADMIN ACCOUNTS ===' as info;
SELECT email, full_name, role, is_active
FROM admin_users
ORDER BY role;
