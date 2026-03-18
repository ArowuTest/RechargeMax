-- Migration 037: Seed default spin_tiers and wheel_prizes
-- Both tables were created by base schema migrations but had no data,
-- causing /api/v1/spins/tier-progress → 500 and /api/v1/spin/play → 400.
-- Uses ON CONFLICT DO NOTHING so existing admin configuration is never overwritten.

-- ── Spin Tiers ────────────────────────────────────────────────────────────────
-- Amounts are in KOBO (1 NGN = 100 kobo), so ₦1,000 = 100,000 kobo.
INSERT INTO spin_tiers (
    id, tier_name, tier_display_name,
    min_daily_amount, max_daily_amount, spins_per_day,
    tier_color, tier_icon, tier_badge, description,
    sort_order, is_active, created_at, updated_at
) VALUES
    (gen_random_uuid(), 'bronze',   'Bronze',
     100000,  249999, 1, '#cd7f32', '🥉', 'Bronze',
     'Recharge ₦1,000–₦2,499 to earn 1 spin per day',
     1, true, NOW(), NOW()),

    (gen_random_uuid(), 'silver',   'Silver',
     250000,  499999, 2, '#c0c0c0', '🥈', 'Silver',
     'Recharge ₦2,500–₦4,999 to earn 2 spins per day',
     2, true, NOW(), NOW()),

    (gen_random_uuid(), 'gold',     'Gold',
     500000,  999999, 3, '#ffd700', '🥇', 'Gold',
     'Recharge ₦5,000–₦9,999 to earn 3 spins per day',
     3, true, NOW(), NOW()),

    (gen_random_uuid(), 'platinum', 'Platinum',
     1000000, 9999999, 5, '#e5e4e2', '💎', 'Platinum',
     'Recharge ₦10,000+ to earn 5 spins per day',
     4, true, NOW(), NOW())

ON CONFLICT (tier_name) DO NOTHING;

-- ── Wheel Prizes ──────────────────────────────────────────────────────────────
-- prize_value in KOBO. Probabilities must sum to 100.
INSERT INTO wheel_prizes (
    id, prize_name, prize_type, prize_value, probability,
    minimum_recharge, is_active, icon_name, color_scheme,
    sort_order, description, created_at, updated_at
) VALUES
    (gen_random_uuid(), '₦100 Airtime',  'AIRTIME', 10000,  25.00, 0, true, 'phone',   '#10b981', 1, '₦100 airtime credit',   NOW(), NOW()),
    (gen_random_uuid(), '₦200 Airtime',  'AIRTIME', 20000,  20.00, 0, true, 'phone',   '#3b82f6', 2, '₦200 airtime credit',   NOW(), NOW()),
    (gen_random_uuid(), '500MB Data',    'DATA',    50000,  15.00, 0, true, 'wifi',    '#8b5cf6', 3, '500MB data bundle',     NOW(), NOW()),
    (gen_random_uuid(), '1GB Data',      'DATA',    100000, 12.00, 0, true, 'wifi',    '#f59e0b', 4, '1GB data bundle',       NOW(), NOW()),
    (gen_random_uuid(), '₦100 Cash',     'CASH',    10000,  10.00, 0, true, 'banknote','#ef4444', 5, '₦100 cash prize',       NOW(), NOW()),
    (gen_random_uuid(), '₦200 Cash',     'CASH',    20000,  8.00,  0, true, 'banknote','#ec4899', 6, '₦200 cash prize',       NOW(), NOW()),
    (gen_random_uuid(), '₦500 Cash',     'CASH',    50000,  6.00,  0, true, 'banknote','#fbbf24', 7, '₦500 cash prize',       NOW(), NOW()),
    (gen_random_uuid(), '₦1000 Cash',    'CASH',    100000, 4.00,  0, true, 'banknote','#6b7280', 8, '₦1,000 cash prize',     NOW(), NOW())

ON CONFLICT DO NOTHING;
