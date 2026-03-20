-- Migration 037: Seed default spin_tiers and wheel_prizes
-- Both tables were created by base schema migrations but had no data,
-- causing /api/v1/spins/tier-progress → 500 and /api/v1/spin/play → 400.
--
-- Tier model (cumulative daily recharge → daily spin cap):
--   Bronze:   ₦1,000  – ₦2,499   → 1 spin/day   (100,000  – 249,999  kobo)
--   Silver:   ₦2,500  – ₦4,999   → 2 spins/day  (250,000  – 499,999  kobo)
--   Gold:     ₦5,000  – ₦9,999   → 3 spins/day  (500,000  – 999,999  kobo)
--   Platinum: ₦10,000+            → 5 spins/day  (1,000,000+ kobo, no upper cap)
--
-- All amounts are in KOBO (1 NGN = 100 kobo), so ₦1,000 = 100,000 kobo.
--
-- ON CONFLICT DO UPDATE is used for spin_tiers so that every deploy re-applies
-- the canonical ranges. This means:
--   - Fresh installs get the correct values immediately.
--   - Existing deployments with wrong values (e.g. Platinum capped at ₦99,999)
--     get corrected automatically on the next restart — no separate patch migration needed.
--   - Admin-customised tiers are intentionally overwritten back to defaults here;
--     if you want tiers to be admin-only, change this back to DO NOTHING.

-- ── Spin Tiers ────────────────────────────────────────────────────────────────
INSERT INTO spin_tiers (
    id, tier_name, tier_display_name,
    min_daily_amount, max_daily_amount, spins_per_day,
    tier_color, tier_icon, tier_badge, description,
    sort_order, is_active, created_at, updated_at
) VALUES
    (gen_random_uuid(), 'bronze',   'Bronze',
     100000,  249999, 1,
     '#cd7f32', '🥉', 'Bronze',
     'Recharge ₦1,000–₦2,499 to earn 1 spin per day',
     1, true, NOW(), NOW()),

    (gen_random_uuid(), 'silver',   'Silver',
     250000,  499999, 2,
     '#c0c0c0', '🥈', 'Silver',
     'Recharge ₦2,500–₦4,999 to earn 2 spins per day',
     2, true, NOW(), NOW()),

    (gen_random_uuid(), 'gold',     'Gold',
     500000,  999999, 3,
     '#ffd700', '🥇', 'Gold',
     'Recharge ₦5,000–₦9,999 to earn 3 spins per day',
     3, true, NOW(), NOW()),

    (gen_random_uuid(), 'platinum', 'Platinum',
     1000000, 999999999999, 5,
     '#e5e4e2', '💎', 'Platinum',
     'Recharge ₦10,000+ to earn 5 spins per day',
     4, true, NOW(), NOW())

ON CONFLICT (tier_name) DO UPDATE SET
    min_daily_amount  = EXCLUDED.min_daily_amount,
    max_daily_amount  = EXCLUDED.max_daily_amount,
    spins_per_day     = EXCLUDED.spins_per_day,
    tier_display_name = EXCLUDED.tier_display_name,
    description       = EXCLUDED.description,
    sort_order        = EXCLUDED.sort_order,
    is_active         = EXCLUDED.is_active,
    updated_at        = NOW();

-- ── Wheel Prizes ──────────────────────────────────────────────────────────────
-- prize_value in KOBO. Probabilities sum to 100.
-- ON CONFLICT DO NOTHING so admin-configured prize changes are preserved.
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
