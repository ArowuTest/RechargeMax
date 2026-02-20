-- Migration: Add Short Code System for Hybrid ID Strategy
-- Purpose: Add user-facing short codes while keeping UUID for internal use
-- Author: Manus AI
-- Date: 2026-02-03

-- ============================================================================
-- 1. ADD SHORT CODE COLUMNS
-- ============================================================================

-- Users: USR_0001, USR_0002, etc.
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS user_code VARCHAR(12) UNIQUE,
ADD COLUMN IF NOT EXISTS display_name VARCHAR(50); -- For @recharge_xxxx

-- Draws: DRW_2026_02_001
ALTER TABLE draws 
ADD COLUMN IF NOT EXISTS draw_code VARCHAR(20) UNIQUE;

-- Prizes: PRZ_CASH_001, PRZ_AIRTIME_042
ALTER TABLE prizes 
ADD COLUMN IF NOT EXISTS prize_code VARCHAR(20) UNIQUE;

-- Subscriptions: SUB_1234_001
ALTER TABLE subscriptions 
ADD COLUMN IF NOT EXISTS subscription_code VARCHAR(20) UNIQUE;

-- Spin Opportunities: SPN_1234_20260203_01
ALTER TABLE spin_opportunities 
ADD COLUMN IF NOT EXISTS spin_code VARCHAR(30) UNIQUE;

-- Affiliate: REF_JOHN (referral_code might already exist, check first)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'affiliates' AND column_name = 'referral_code'
    ) THEN
        ALTER TABLE affiliates ADD COLUMN referral_code VARCHAR(20) UNIQUE;
    END IF;
END $$;

-- ============================================================================
-- 2. CREATE INDEXES FOR FAST LOOKUP
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_users_user_code ON users(user_code);
CREATE INDEX IF NOT EXISTS idx_users_display_name ON users(display_name);
CREATE INDEX IF NOT EXISTS idx_draws_draw_code ON draws(draw_code);
CREATE INDEX IF NOT EXISTS idx_prizes_prize_code ON prizes(prize_code);
CREATE INDEX IF NOT EXISTS idx_subscriptions_subscription_code ON subscriptions(subscription_code);
CREATE INDEX IF NOT EXISTS idx_spin_opportunities_spin_code ON spin_opportunities(spin_code);
CREATE INDEX IF NOT EXISTS idx_affiliates_referral_code ON affiliates(referral_code);

-- ============================================================================
-- 3. CREATE SEQUENCE GENERATORS
-- ============================================================================

CREATE SEQUENCE IF NOT EXISTS user_code_seq START 1 INCREMENT 1;
CREATE SEQUENCE IF NOT EXISTS draw_code_seq START 1 INCREMENT 1;
CREATE SEQUENCE IF NOT EXISTS prize_code_seq START 1 INCREMENT 1;
CREATE SEQUENCE IF NOT EXISTS subscription_code_seq START 1 INCREMENT 1;

-- ============================================================================
-- 4. CREATE HELPER FUNCTIONS FOR SHORT CODE GENERATION
-- ============================================================================

-- Function: Generate User Code (USR_0001)
CREATE OR REPLACE FUNCTION generate_user_code()
RETURNS VARCHAR(12) AS $$
DECLARE
    next_val INTEGER;
    code VARCHAR(12);
BEGIN
    next_val := nextval('user_code_seq');
    code := 'USR_' || LPAD(next_val::TEXT, 4, '0');
    RETURN code;
END;
$$ LANGUAGE plpgsql;

-- Function: Generate Draw Code (DRW_2026_02_001)
CREATE OR REPLACE FUNCTION generate_draw_code(draw_date_param TIMESTAMP)
RETURNS VARCHAR(20) AS $$
DECLARE
    year_month VARCHAR(7);
    sequence_num INTEGER;
    code VARCHAR(20);
BEGIN
    year_month := TO_CHAR(draw_date_param, 'YYYY_MM');
    
    -- Get sequence number for this month
    SELECT COALESCE(MAX(CAST(SUBSTRING(draw_code FROM '[0-9]+$') AS INTEGER)), 0) + 1
    INTO sequence_num
    FROM draws
    WHERE draw_code LIKE 'DRW_' || year_month || '_%';
    
    code := 'DRW_' || year_month || '_' || LPAD(sequence_num::TEXT, 3, '0');
    RETURN code;
END;
$$ LANGUAGE plpgsql;

-- Function: Generate Prize Code (PRZ_CASH_001)
CREATE OR REPLACE FUNCTION generate_prize_code(prize_name TEXT)
RETURNS VARCHAR(20) AS $$
DECLARE
    category VARCHAR(10);
    next_val INTEGER;
    code VARCHAR(20);
BEGIN
    -- Extract category from prize name (first 4 chars, uppercase)
    category := UPPER(SUBSTRING(REGEXP_REPLACE(prize_name, '[^a-zA-Z]', '', 'g'), 1, 4));
    IF LENGTH(category) = 0 THEN
        category := 'ITEM';
    END IF;
    
    next_val := nextval('prize_code_seq');
    code := 'PRZ_' || category || '_' || LPAD(next_val::TEXT, 3, '0');
    RETURN code;
END;
$$ LANGUAGE plpgsql;

-- Function: Generate Subscription Code (SUB_1234_001)
CREATE OR REPLACE FUNCTION generate_subscription_code(user_id_param UUID)
RETURNS VARCHAR(20) AS $$
DECLARE
    user_last4 VARCHAR(4);
    next_val INTEGER;
    code VARCHAR(20);
BEGIN
    -- Get last 4 digits of user's user_code
    SELECT RIGHT(user_code, 4) INTO user_last4
    FROM users WHERE id = user_id_param;
    
    IF user_last4 IS NULL THEN
        user_last4 := '0000';
    END IF;
    
    next_val := nextval('subscription_code_seq');
    code := 'SUB_' || user_last4 || '_' || LPAD(next_val::TEXT, 3, '0');
    RETURN code;
END;
$$ LANGUAGE plpgsql;

-- Function: Generate Spin Code (SPN_1234_20260203_01)
CREATE OR REPLACE FUNCTION generate_spin_code(user_id_param UUID, spin_date TIMESTAMP)
RETURNS VARCHAR(30) AS $$
DECLARE
    user_last4 VARCHAR(4);
    date_str VARCHAR(8);
    sequence_num INTEGER;
    code VARCHAR(30);
BEGIN
    -- Get last 4 digits of user's user_code
    SELECT RIGHT(user_code, 4) INTO user_last4
    FROM users WHERE id = user_id_param;
    
    IF user_last4 IS NULL THEN
        user_last4 := '0000';
    END IF;
    
    date_str := TO_CHAR(spin_date, 'YYYYMMDD');
    
    -- Get sequence for this user on this date
    SELECT COALESCE(MAX(CAST(SUBSTRING(spin_code FROM '[0-9]+$') AS INTEGER)), 0) + 1
    INTO sequence_num
    FROM spin_opportunities
    WHERE spin_code LIKE 'SPN_' || user_last4 || '_' || date_str || '_%';
    
    code := 'SPN_' || user_last4 || '_' || date_str || '_' || LPAD(sequence_num::TEXT, 2, '0');
    RETURN code;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 5. CREATE TRIGGERS FOR AUTOMATIC SHORT CODE GENERATION
-- ============================================================================

-- Trigger: Auto-generate user_code on user insert
CREATE OR REPLACE FUNCTION trigger_generate_user_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.user_code IS NULL THEN
        NEW.user_code := generate_user_code();
        NEW.display_name := '@recharge_' || RIGHT(NEW.user_code, 4);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS auto_generate_user_code ON users;
CREATE TRIGGER auto_generate_user_code
    BEFORE INSERT ON users
    FOR EACH ROW
    EXECUTE FUNCTION trigger_generate_user_code();

-- Trigger: Auto-generate draw_code on draw insert
CREATE OR REPLACE FUNCTION trigger_generate_draw_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.draw_code IS NULL THEN
        NEW.draw_code := generate_draw_code(NEW.draw_date);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS auto_generate_draw_code ON draws;
CREATE TRIGGER auto_generate_draw_code
    BEFORE INSERT ON draws
    FOR EACH ROW
    EXECUTE FUNCTION trigger_generate_draw_code();

-- Trigger: Auto-generate prize_code on prize insert
CREATE OR REPLACE FUNCTION trigger_generate_prize_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.prize_code IS NULL THEN
        NEW.prize_code := generate_prize_code(NEW.name);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS auto_generate_prize_code ON prizes;
CREATE TRIGGER auto_generate_prize_code
    BEFORE INSERT ON prizes
    FOR EACH ROW
    EXECUTE FUNCTION trigger_generate_prize_code();

-- Trigger: Auto-generate subscription_code on subscription insert
CREATE OR REPLACE FUNCTION trigger_generate_subscription_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.subscription_code IS NULL THEN
        NEW.subscription_code := generate_subscription_code(NEW.user_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS auto_generate_subscription_code ON subscriptions;
CREATE TRIGGER auto_generate_subscription_code
    BEFORE INSERT ON subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION trigger_generate_subscription_code();

-- Trigger: Auto-generate spin_code on spin_opportunity insert
CREATE OR REPLACE FUNCTION trigger_generate_spin_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.spin_code IS NULL THEN
        NEW.spin_code := generate_spin_code(NEW.user_id, NEW.created_at);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS auto_generate_spin_code ON spin_opportunities;
CREATE TRIGGER auto_generate_spin_code
    BEFORE INSERT ON spin_opportunities
    FOR EACH ROW
    EXECUTE FUNCTION trigger_generate_spin_code();

-- ============================================================================
-- 6. BACKFILL EXISTING RECORDS
-- ============================================================================

-- Backfill users
UPDATE users 
SET 
    user_code = 'USR_' || LPAD(ROW_NUMBER() OVER (ORDER BY created_at)::TEXT, 4, '0'),
    display_name = '@recharge_' || LPAD(ROW_NUMBER() OVER (ORDER BY created_at)::TEXT, 4, '0')
WHERE user_code IS NULL;

-- Backfill draws
UPDATE draws 
SET draw_code = 'DRW_' || TO_CHAR(draw_date, 'YYYY_MM') || '_' || 
                LPAD(ROW_NUMBER() OVER (PARTITION BY DATE_TRUNC('month', draw_date) ORDER BY created_at)::TEXT, 3, '0')
WHERE draw_code IS NULL;

-- Backfill prizes
UPDATE prizes 
SET prize_code = 'PRZ_' || UPPER(SUBSTRING(REGEXP_REPLACE(name, '[^a-zA-Z]', '', 'g'), 1, 4)) || '_' || 
                 LPAD(ROW_NUMBER() OVER (ORDER BY created_at)::TEXT, 3, '0')
WHERE prize_code IS NULL;

-- Backfill subscriptions
UPDATE subscriptions s
SET subscription_code = 'SUB_' || COALESCE(RIGHT(u.user_code, 4), '0000') || '_' || 
                        LPAD(ROW_NUMBER() OVER (PARTITION BY s.user_id ORDER BY s.created_at)::TEXT, 3, '0')
FROM users u
WHERE s.user_id = u.id AND s.subscription_code IS NULL;

-- Backfill spin_opportunities
UPDATE spin_opportunities so
SET spin_code = 'SPN_' || COALESCE(RIGHT(u.user_code, 4), '0000') || '_' || 
                TO_CHAR(so.created_at, 'YYYYMMDD') || '_' || 
                LPAD(ROW_NUMBER() OVER (PARTITION BY so.user_id, DATE(so.created_at) ORDER BY so.created_at)::TEXT, 2, '0')
FROM users u
WHERE so.user_id = u.id AND so.spin_code IS NULL;

-- ============================================================================
-- 7. ADD CONSTRAINTS
-- ============================================================================

-- Ensure user_code is not null for new records (allow NULL for migration)
-- ALTER TABLE users ALTER COLUMN user_code SET NOT NULL; -- Enable after backfill

-- Add check constraints for format validation
ALTER TABLE users ADD CONSTRAINT chk_user_code_format 
    CHECK (user_code ~ '^USR_[0-9]{4}$');

ALTER TABLE draws ADD CONSTRAINT chk_draw_code_format 
    CHECK (draw_code ~ '^DRW_[0-9]{4}_[0-9]{2}_[0-9]{3}$');

ALTER TABLE prizes ADD CONSTRAINT chk_prize_code_format 
    CHECK (prize_code ~ '^PRZ_[A-Z]{1,10}_[0-9]{3}$');

ALTER TABLE subscriptions ADD CONSTRAINT chk_subscription_code_format 
    CHECK (subscription_code ~ '^SUB_[0-9]{4}_[0-9]{3}$');

ALTER TABLE spin_opportunities ADD CONSTRAINT chk_spin_code_format 
    CHECK (spin_code ~ '^SPN_[0-9]{4}_[0-9]{8}_[0-9]{2}$');

-- ============================================================================
-- 8. GRANT PERMISSIONS
-- ============================================================================

GRANT SELECT, UPDATE ON users TO rechargemax;
GRANT SELECT, UPDATE ON draws TO rechargemax;
GRANT SELECT, UPDATE ON prizes TO rechargemax;
GRANT SELECT, UPDATE ON subscriptions TO rechargemax;
GRANT SELECT, UPDATE ON spin_opportunities TO rechargemax;
GRANT USAGE, SELECT ON user_code_seq TO rechargemax;
GRANT USAGE, SELECT ON draw_code_seq TO rechargemax;
GRANT USAGE, SELECT ON prize_code_seq TO rechargemax;
GRANT USAGE, SELECT ON subscription_code_seq TO rechargemax;

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================

-- Verify migration
DO $$
DECLARE
    user_count INTEGER;
    draw_count INTEGER;
    prize_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO user_count FROM users WHERE user_code IS NOT NULL;
    SELECT COUNT(*) INTO draw_count FROM draws WHERE draw_code IS NOT NULL;
    SELECT COUNT(*) INTO prize_count FROM prizes WHERE prize_code IS NOT NULL;
    
    RAISE NOTICE 'Short Code Migration Complete:';
    RAISE NOTICE '  - Users with codes: %', user_count;
    RAISE NOTICE '  - Draws with codes: %', draw_count;
    RAISE NOTICE '  - Prizes with codes: %', prize_count;
END $$;
