-- ============================================================================
-- RechargeMax Draw & Prize Configuration Seed
-- Migrated from backend/seeds/prize_tier_seed.sql
-- Corrected to match actual schema (UUID PKs, correct column names)
-- ============================================================================

-- ----------------------------------------------------------------------------
-- DRAW TYPES
-- ----------------------------------------------------------------------------

INSERT INTO public.draw_types (id, name, description, is_active, created_at, updated_at)
VALUES
  ('a1000000-0000-0000-0000-000000000001', 'Daily',   'Daily draw with smaller prize pool',        true, NOW(), NOW()),
  ('a1000000-0000-0000-0000-000000000002', 'Weekly',  'Weekly draw with mega prize pool',           true, NOW(), NOW()),
  ('a1000000-0000-0000-0000-000000000003', 'Special', 'Special event draw with custom prizes',      true, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- ----------------------------------------------------------------------------
-- PRIZE TEMPLATES
-- (draw_type_id references draw_types above; no is_default column in schema)
-- ----------------------------------------------------------------------------

INSERT INTO public.prize_templates (id, name, draw_type_id, description, is_active, created_at, updated_at)
VALUES
  ('b1000000-0000-0000-0000-000000000001', 'Daily Standard Template',
   'a1000000-0000-0000-0000-000000000001', 'Standard daily draw with 3 prize tiers',        true, NOW(), NOW()),
  ('b1000000-0000-0000-0000-000000000002', 'Daily Premium Template',
   'a1000000-0000-0000-0000-000000000001', 'Premium daily draw with 4 prize tiers',         true, NOW(), NOW()),
  ('b1000000-0000-0000-0000-000000000003', 'Weekly Mega Template',
   'a1000000-0000-0000-0000-000000000002', 'Mega weekly draw with 5 prize tiers',           true, NOW(), NOW()),
  ('b1000000-0000-0000-0000-000000000004', 'Weekly Standard Template',
   'a1000000-0000-0000-0000-000000000002', 'Standard weekly draw with 4 prize tiers',       true, NOW(), NOW()),
  ('b1000000-0000-0000-0000-000000000005', 'Special Event Template',
   'a1000000-0000-0000-0000-000000000003', 'Special event draw with custom prize structure', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ----------------------------------------------------------------------------
-- PRIZE CATEGORIES
-- (template_id references prize_templates; winners_count/runner_ups_count per schema)
-- ----------------------------------------------------------------------------

-- Daily Standard Template categories
INSERT INTO public.prize_categories (id, template_id, category_name, prize_amount, winners_count, runner_ups_count, display_order, created_at, updated_at)
VALUES
  ('c1000000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'Jackpot',     500000,  1,  2, 1, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000002', 'b1000000-0000-0000-0000-000000000001', 'First Prize', 200000,  3,  5, 2, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000003', 'b1000000-0000-0000-0000-000000000001', 'Second Prize',100000,  5, 10, 3, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Daily Premium Template categories
INSERT INTO public.prize_categories (id, template_id, category_name, prize_amount, winners_count, runner_ups_count, display_order, created_at, updated_at)
VALUES
  ('c1000000-0000-0000-0000-000000000004', 'b1000000-0000-0000-0000-000000000002', 'Grand Prize',  1000000, 1,  3, 1, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000005', 'b1000000-0000-0000-0000-000000000002', 'First Prize',   300000, 2,  4, 2, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000006', 'b1000000-0000-0000-0000-000000000002', 'Second Prize',  150000, 5,  8, 3, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000007', 'b1000000-0000-0000-0000-000000000002', 'Third Prize',    50000,10, 15, 4, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Weekly Mega Template categories
INSERT INTO public.prize_categories (id, template_id, category_name, prize_amount, winners_count, runner_ups_count, display_order, created_at, updated_at)
VALUES
  ('c1000000-0000-0000-0000-000000000008', 'b1000000-0000-0000-0000-000000000003', 'Mega Jackpot', 10000000, 1,  3, 1, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000009', 'b1000000-0000-0000-0000-000000000003', 'First Prize',   5000000, 2,  5, 2, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000010', 'b1000000-0000-0000-0000-000000000003', 'Second Prize',  2000000, 5, 10, 3, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000011', 'b1000000-0000-0000-0000-000000000003', 'Third Prize',   1000000,10, 20, 4, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000012', 'b1000000-0000-0000-0000-000000000003', 'Fourth Prize',   500000,20, 30, 5, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Weekly Standard Template categories
INSERT INTO public.prize_categories (id, template_id, category_name, prize_amount, winners_count, runner_ups_count, display_order, created_at, updated_at)
VALUES
  ('c1000000-0000-0000-0000-000000000013', 'b1000000-0000-0000-0000-000000000004', 'Jackpot',     5000000, 1,  2, 1, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000014', 'b1000000-0000-0000-0000-000000000004', 'First Prize', 2000000, 3,  5, 2, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000015', 'b1000000-0000-0000-0000-000000000004', 'Second Prize',1000000, 5, 10, 3, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000016', 'b1000000-0000-0000-0000-000000000004', 'Third Prize',  500000,10, 15, 4, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Special Event Template categories
INSERT INTO public.prize_categories (id, template_id, category_name, prize_amount, winners_count, runner_ups_count, display_order, created_at, updated_at)
VALUES
  ('c1000000-0000-0000-0000-000000000017', 'b1000000-0000-0000-0000-000000000005', 'Grand Prize',  20000000, 1,  5, 1, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000018', 'b1000000-0000-0000-0000-000000000005', 'First Prize',  10000000, 2,  8, 2, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000019', 'b1000000-0000-0000-0000-000000000005', 'Second Prize',  5000000, 5, 15, 3, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000020', 'b1000000-0000-0000-0000-000000000005', 'Third Prize',   2000000,10, 20, 4, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000021', 'b1000000-0000-0000-0000-000000000005', 'Fourth Prize',  1000000,20, 30, 5, NOW(), NOW()),
  ('c1000000-0000-0000-0000-000000000022', 'b1000000-0000-0000-0000-000000000005', 'Fifth Prize',    500000,50,100, 6, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SUBSCRIPTION TIERS (daily-draw auto-entry tiers)
-- ON CONFLICT DO NOTHING — existing BRONZE/SILVER/GOLD/PLATINUM rows in
-- 000_MASTER_PRODUCTION_SEED.sql are preserved; Basic/Silver/Gold/Platinum
-- names are added only if not already present.
-- ============================================================================

INSERT INTO public.subscription_tiers (name, description, entries, is_active, sort_order, created_at, updated_at)
VALUES
  ('Basic',
   'Perfect for casual players. Get automatic entry into daily draws. ₦50/day, ₦300/week, ₦1,000/month',
   1, true, 10, NOW(), NOW()),
  ('Silver',
   'Enhanced chances with multiple entries and bonus spins. ₦100/day, ₦600/week, ₦2,000/month',
   3, true, 11, NOW(), NOW()),
  ('Gold',
   'Premium tier with maximum entries, bonus spins, and priority support. ₦200/day, ₦1,200/week, ₦4,000/month',
   5, true, 12, NOW(), NOW()),
  ('Platinum',
   'VIP tier with unlimited entries, maximum bonus spins, priority support, and exclusive prizes. ₦500/day, ₦3,000/week, ₦10,000/month',
   10, true, 13, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- Summary
-- 3 draw types | 5 prize templates | 22 prize categories | 4 subscription tiers
-- ============================================================================
