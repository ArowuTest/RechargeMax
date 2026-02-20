-- Prize Tier System Seed Data
-- This file contains default draw types and prize templates for testing

-- ============================================================================
-- DRAW TYPES
-- ============================================================================

INSERT INTO draw_types (id, name, description, is_active, created_at, updated_at)
VALUES
  (1, 'Daily', 'Daily draw with smaller prize pool', true, NOW(), NOW()),
  (2, 'Weekly', 'Weekly draw with mega prize pool', true, NOW(), NOW()),
  (3, 'Special', 'Special event draw with custom prizes', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- PRIZE TEMPLATES - DAILY
-- ============================================================================

-- Daily Standard Template
INSERT INTO prize_templates (id, name, draw_type_id, description, is_default, is_active, created_at, updated_at)
VALUES
  (1, 'Daily Standard Template', 1, 'Standard daily draw with 3 prize tiers', true, true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Daily Prize Categories
INSERT INTO prize_categories (id, prize_template_id, category_name, prize_amount, winner_count, runner_up_count, display_order, created_at, updated_at)
VALUES
  (1, 1, 'Jackpot', 500000, 1, 2, 1, NOW(), NOW()),
  (2, 1, 'First Prize', 200000, 3, 5, 2, NOW(), NOW()),
  (3, 1, 'Second Prize', 100000, 5, 10, 3, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Daily Premium Template
INSERT INTO prize_templates (id, name, draw_type_id, description, is_default, is_active, created_at, updated_at)
VALUES
  (2, 'Daily Premium Template', 1, 'Premium daily draw with 4 prize tiers', false, true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Daily Premium Prize Categories
INSERT INTO prize_categories (id, prize_template_id, category_name, prize_amount, winner_count, runner_up_count, display_order, created_at, updated_at)
VALUES
  (4, 2, 'Grand Prize', 1000000, 1, 3, 1, NOW(), NOW()),
  (5, 2, 'First Prize', 300000, 2, 4, 2, NOW(), NOW()),
  (6, 2, 'Second Prize', 150000, 5, 8, 3, NOW(), NOW()),
  (7, 2, 'Third Prize', 50000, 10, 15, 4, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- PRIZE TEMPLATES - WEEKLY
-- ============================================================================

-- Weekly Mega Template
INSERT INTO prize_templates (id, name, draw_type_id, description, is_default, is_active, created_at, updated_at)
VALUES
  (3, 'Weekly Mega Template', 2, 'Mega weekly draw with 5 prize tiers', true, true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Weekly Mega Prize Categories
INSERT INTO prize_categories (id, prize_template_id, category_name, prize_amount, winner_count, runner_up_count, display_order, created_at, updated_at)
VALUES
  (8, 3, 'Mega Jackpot', 10000000, 1, 3, 1, NOW(), NOW()),
  (9, 3, 'First Prize', 5000000, 2, 5, 2, NOW(), NOW()),
  (10, 3, 'Second Prize', 2000000, 5, 10, 3, NOW(), NOW()),
  (11, 3, 'Third Prize', 1000000, 10, 20, 4, NOW(), NOW()),
  (12, 3, 'Fourth Prize', 500000, 20, 30, 5, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Weekly Standard Template
INSERT INTO prize_templates (id, name, draw_type_id, description, is_default, is_active, created_at, updated_at)
VALUES
  (4, 'Weekly Standard Template', 2, 'Standard weekly draw with 4 prize tiers', false, true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Weekly Standard Prize Categories
INSERT INTO prize_categories (id, prize_template_id, category_name, prize_amount, winner_count, runner_up_count, display_order, created_at, updated_at)
VALUES
  (13, 4, 'Jackpot', 5000000, 1, 2, 1, NOW(), NOW()),
  (14, 4, 'First Prize', 2000000, 3, 5, 2, NOW(), NOW()),
  (15, 4, 'Second Prize', 1000000, 5, 10, 3, NOW(), NOW()),
  (16, 4, 'Third Prize', 500000, 10, 15, 4, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- PRIZE TEMPLATES - SPECIAL
-- ============================================================================

-- Special Event Template
INSERT INTO prize_templates (id, name, draw_type_id, description, is_default, is_active, created_at, updated_at)
VALUES
  (5, 'Special Event Template', 3, 'Special event draw with custom prize structure', true, true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Special Event Prize Categories
INSERT INTO prize_categories (id, prize_template_id, category_name, prize_amount, winner_count, runner_up_count, display_order, created_at, updated_at)
VALUES
  (17, 5, 'Grand Prize', 20000000, 1, 5, 1, NOW(), NOW()),
  (18, 5, 'First Prize', 10000000, 2, 8, 2, NOW(), NOW()),
  (19, 5, 'Second Prize', 5000000, 5, 15, 3, NOW(), NOW()),
  (20, 5, 'Third Prize', 2000000, 10, 20, 4, NOW(), NOW()),
  (21, 5, 'Fourth Prize', 1000000, 20, 30, 5, NOW(), NOW()),
  (22, 5, 'Fifth Prize', 500000, 50, 100, 6, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SUMMARY
-- ============================================================================

-- Total Prize Templates: 5
-- Total Prize Categories: 22
-- 
-- Daily Templates:
--   1. Daily Standard Template (3 categories) - Total Pool: ₦1,300,000
--   2. Daily Premium Template (4 categories) - Total Pool: ₦2,250,000
-- 
-- Weekly Templates:
--   3. Weekly Mega Template (5 categories) - Total Pool: ₦40,000,000
--   4. Weekly Standard Template (4 categories) - Total Pool: ₦14,500,000
-- 
-- Special Templates:
--   5. Special Event Template (6 categories) - Total Pool: ₦82,000,000

-- Reset sequences to continue from next available ID
SELECT setval('draw_types_id_seq', (SELECT MAX(id) FROM draw_types), true);
SELECT setval('prize_templates_id_seq', (SELECT MAX(id) FROM prize_templates), true);
SELECT setval('prize_categories_id_seq', (SELECT MAX(id) FROM prize_categories), true);
