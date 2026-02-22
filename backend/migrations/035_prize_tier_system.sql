-- Prize Tier System Migration
-- Implements draw types, prize templates, and prize categories

-- 1. Create draw_types table
CREATE TABLE IF NOT EXISTS draw_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Create prize_templates table
CREATE TABLE IF NOT EXISTS prize_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    draw_type_id UUID REFERENCES draw_types(id) ON DELETE CASCADE,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create prize_categories table
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

-- 4. Update draw_winners table to link to prize categories
ALTER TABLE draw_winners ADD COLUMN IF NOT EXISTS prize_category_id UUID REFERENCES prize_categories(id) ON DELETE SET NULL;
ALTER TABLE draw_winners ADD COLUMN IF NOT EXISTS category_name VARCHAR(100);
ALTER TABLE draw_winners ADD COLUMN IF NOT EXISTS prize_amount DECIMAL(15,2);

-- 5. Add draw_type_id to draws table
ALTER TABLE draws ADD COLUMN IF NOT EXISTS draw_type_id UUID REFERENCES draw_types(id) ON DELETE SET NULL;

-- 6. Create indexes
CREATE INDEX IF NOT EXISTS idx_prize_categories_template ON prize_categories(template_id);
CREATE INDEX IF NOT EXISTS idx_prize_categories_draw ON prize_categories(draw_id);
CREATE INDEX IF NOT EXISTS idx_prize_categories_order ON prize_categories(display_order);
CREATE INDEX IF NOT EXISTS idx_draw_winners_category ON draw_winners(prize_category_id);
CREATE INDEX IF NOT EXISTS idx_draws_type ON draws(draw_type_id);

-- 7. Insert default draw types
INSERT INTO draw_types (id, name, description) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Daily', 'Daily draws with smaller prizes'),
    ('22222222-2222-2222-2222-222222222222', 'Weekly', 'Weekly draws with larger prizes'),
    ('33333333-3333-3333-3333-333333333333', 'Special', 'Special event draws with custom prizes')
ON CONFLICT (name) DO NOTHING;

-- 8. Insert default prize templates
INSERT INTO prize_templates (id, name, draw_type_id, description) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Standard Daily Draw', '11111111-1111-1111-1111-111111111111', 'Standard prize structure for daily draws'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Standard Weekly Draw', '22222222-2222-2222-2222-222222222222', 'Standard prize structure for weekly draws')
ON CONFLICT DO NOTHING;

-- 9. Insert default prize categories for Daily template
INSERT INTO prize_categories (template_id, category_name, prize_amount, winners_count, runner_ups_count, display_order) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Winner', 10000.00, 1, 2, 1),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Consolation', 2000.00, 3, 2, 2)
ON CONFLICT DO NOTHING;

-- 10. Insert default prize categories for Weekly template
INSERT INTO prize_categories (template_id, category_name, prize_amount, winners_count, runner_ups_count, display_order) VALUES
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Jackpot', 500000.00, 1, 3, 1),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'First Prize', 200000.00, 1, 2, 2),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Second Prize', 100000.00, 1, 2, 3),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Third Prize', 50000.00, 1, 2, 4),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Consolation', 10000.00, 5, 3, 5)
ON CONFLICT DO NOTHING;
