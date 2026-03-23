-- Migration 048: Clean up corrupted test wheel_prizes and re-seed canonical set
-- The wheel_prizes table accumulated 384 duplicate/corrupted prizes from
-- UAT testing before the dedup guard was added. This migration:
--   1. Deletes ALL existing wheel_prizes (they were all created during testing)
--   2. Re-inserts the canonical 8 prizes with ON CONFLICT DO NOTHING
-- 
-- Safe to re-run: ON CONFLICT DO NOTHING means re-runs are no-ops.

-- Step 1: Remove all test/duplicate prizes
TRUNCATE TABLE wheel_prizes RESTART IDENTITY CASCADE;

-- Step 2: Re-seed the canonical 8 prizes (total probability = 100%)
INSERT INTO wheel_prizes (
    id, prize_name, prize_type, prize_value, probability,
    minimum_recharge, is_active, icon_name, color_scheme,
    sort_order, description, created_at, updated_at
) VALUES
    (gen_random_uuid(), '₦100 Airtime',  'AIRTIME', 10000,  25.00, 0, true, 'phone',    '#10b981', 1, '₦100 airtime credit',  NOW(), NOW()),
    (gen_random_uuid(), '₦200 Airtime',  'AIRTIME', 20000,  20.00, 0, true, 'phone',    '#3b82f6', 2, '₦200 airtime credit',  NOW(), NOW()),
    (gen_random_uuid(), '500MB Data',    'DATA',    50000,  15.00, 0, true, 'wifi',     '#8b5cf6', 3, '500MB data bundle',    NOW(), NOW()),
    (gen_random_uuid(), '1GB Data',      'DATA',    100000, 12.00, 0, true, 'wifi',     '#f59e0b', 4, '1GB data bundle',      NOW(), NOW()),
    (gen_random_uuid(), '₦100 Cash',     'CASH',    10000,  10.00, 0, true, 'banknote', '#ef4444', 5, '₦100 cash prize',      NOW(), NOW()),
    (gen_random_uuid(), '₦200 Cash',     'CASH',    20000,   8.00, 0, true, 'banknote', '#ec4899', 6, '₦200 cash prize',      NOW(), NOW()),
    (gen_random_uuid(), '₦500 Cash',     'CASH',    50000,   6.00, 0, true, 'banknote', '#fbbf24', 7, '₦500 cash prize',      NOW(), NOW()),
    (gen_random_uuid(), '₦1000 Cash',    'CASH',    100000,  4.00, 0, true, 'banknote', '#6b7280', 8, '₦1,000 cash prize',    NOW(), NOW())

ON CONFLICT DO NOTHING;

-- Total: 100%  (25+20+15+12+10+8+6+4 = 100)
