-- Subscription Tiers Seed Data
-- Date: February 15, 2026
-- Purpose: Seed subscription tiers for daily draw auto-entry

-- Clear existing data
TRUNCATE TABLE subscription_tiers RESTART IDENTITY CASCADE;

-- Insert subscription tiers
INSERT INTO subscription_tiers (
    name,
    description,
    entries,
    is_active,
    sort_order,
    created_at,
    updated_at
) VALUES
-- Basic Tier (1 entry per day)
(
    'Basic',
    'Perfect for casual players. Get automatic entry into daily draws. ₦50/day, ₦300/week, ₦1,000/month',
    1,
    true,
    1,
    NOW(),
    NOW()
),

-- Silver Tier (3 entries per day)
(
    'Silver',
    'Enhanced chances with multiple entries and bonus spins. ₦100/day, ₦600/week, ₦2,000/month',
    3,
    true,
    2,
    NOW(),
    NOW()
),

-- Gold Tier (5 entries per day)
(
    'Gold',
    'Premium tier with maximum entries, bonus spins, and priority support. ₦200/day, ₦1,200/week, ₦4,000/month',
    5,
    true,
    3,
    NOW(),
    NOW()
),

-- Platinum Tier (10 entries per day)
(
    'Platinum',
    'VIP tier with unlimited entries, maximum bonus spins, priority support, and exclusive prizes. ₦500/day, ₦3,000/week, ₦10,000/month',
    10,
    true,
    4,
    NOW(),
    NOW()
);

-- Verify insertion
SELECT 
    name,
    description,
    entries,
    is_active,
    sort_order
FROM subscription_tiers
ORDER BY sort_order;
