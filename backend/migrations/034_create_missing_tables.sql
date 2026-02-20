-- ============================================================================
-- Migration: Create Missing Tables for Production Readiness
-- Created: 2026-02-20
-- Description: Create 4 missing tables needed for complete system functionality
-- ============================================================================

-- ============================================================================
-- 1. Create draw_types table
-- ============================================================================

CREATE TABLE IF NOT EXISTS draw_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE draw_types IS 'Types of draws (e.g., daily, weekly, monthly, special)';

-- ============================================================================
-- 2. Create prize_templates table
-- ============================================================================

CREATE TABLE IF NOT EXISTS prize_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    draw_type_id UUID REFERENCES draw_types(id) ON DELETE CASCADE,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE prize_templates IS 'Reusable prize templates for different draw types';

-- ============================================================================
-- 3. Create prize_categories table
-- ============================================================================

CREATE TABLE IF NOT EXISTS prize_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id UUID REFERENCES prize_templates(id) ON DELETE CASCADE,
    draw_id UUID REFERENCES draws(id) ON DELETE CASCADE,
    category_name VARCHAR(100) NOT NULL,
    prize_amount DECIMAL(15,2) NOT NULL,
    winners_count INT NOT NULL DEFAULT 1,
    runner_ups_count INT NOT NULL DEFAULT 1,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_parent CHECK (template_id IS NOT NULL OR draw_id IS NOT NULL)
);

COMMENT ON TABLE prize_categories IS 'Prize categories for draws (e.g., 1st place, 2nd place, consolation)';

-- ============================================================================
-- 4. Create points_adjustments table
-- ============================================================================

CREATE TABLE IF NOT EXISTS points_adjustments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    points INT NOT NULL,
    reason VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_points_adjustments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_points_adjustments_admin FOREIGN KEY (created_by) REFERENCES admin_users(id)
);

COMMENT ON TABLE points_adjustments IS 'Tracks manual points adjustments made by administrators';
COMMENT ON COLUMN points_adjustments.points IS 'Points amount (positive for add, negative for deduct)';
COMMENT ON COLUMN points_adjustments.reason IS 'Reason for adjustment (e.g., manual_adjustment, compensation, correction)';
COMMENT ON COLUMN points_adjustments.created_by IS 'Admin user ID who made the adjustment';

-- ============================================================================
-- 5. Create indexes for performance
-- ============================================================================

-- draw_types indexes
CREATE INDEX IF NOT EXISTS idx_draw_types_active ON draw_types(is_active);

-- prize_templates indexes
CREATE INDEX IF NOT EXISTS idx_prize_templates_draw_type ON prize_templates(draw_type_id);
CREATE INDEX IF NOT EXISTS idx_prize_templates_active ON prize_templates(is_active);

-- prize_categories indexes
CREATE INDEX IF NOT EXISTS idx_prize_categories_template ON prize_categories(template_id);
CREATE INDEX IF NOT EXISTS idx_prize_categories_draw ON prize_categories(draw_id);
CREATE INDEX IF NOT EXISTS idx_prize_categories_order ON prize_categories(display_order);

-- points_adjustments indexes
CREATE INDEX IF NOT EXISTS idx_points_adjustments_user_id ON points_adjustments(user_id);
CREATE INDEX IF NOT EXISTS idx_points_adjustments_created_by ON points_adjustments(created_by);
CREATE INDEX IF NOT EXISTS idx_points_adjustments_created_at ON points_adjustments(created_at DESC);

-- ============================================================================
-- 6. Add columns to existing tables if they don't exist
-- ============================================================================

-- Add draw_type_id to draws table
ALTER TABLE draws ADD COLUMN IF NOT EXISTS draw_type_id UUID REFERENCES draw_types(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_draws_type ON draws(draw_type_id);

-- Add prize_category_id to draw_winners table
ALTER TABLE draw_winners ADD COLUMN IF NOT EXISTS prize_category_id UUID REFERENCES prize_categories(id) ON DELETE SET NULL;
ALTER TABLE draw_winners ADD COLUMN IF NOT EXISTS category_name VARCHAR(100);
ALTER TABLE draw_winners ADD COLUMN IF NOT EXISTS prize_amount DECIMAL(15,2);
CREATE INDEX IF NOT EXISTS idx_draw_winners_category ON draw_winners(prize_category_id);

-- ============================================================================
-- 7. Seed default draw types
-- ============================================================================

INSERT INTO draw_types (id, name, description, is_active) VALUES
    ('d1111111-1111-1111-1111-111111111111', 'daily', 'Daily draw for all participants', true),
    ('d2222222-2222-2222-2222-222222222222', 'weekly', 'Weekly draw with bigger prizes', true),
    ('d3333333-3333-3333-3333-333333333333', 'monthly', 'Monthly grand draw', true),
    ('d4444444-4444-4444-4444-444444444444', 'special', 'Special promotional draws', true)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- Migration Complete
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Migration 034 completed successfully!';
    RAISE NOTICE 'Created tables: draw_types, prize_templates, prize_categories, points_adjustments';
    RAISE NOTICE 'Added columns: draws.draw_type_id, draw_winners.prize_category_id';
END $$;
