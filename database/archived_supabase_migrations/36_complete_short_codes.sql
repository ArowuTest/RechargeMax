-- Migration: Complete Short Code System Implementation
-- Purpose: Add remaining short codes for hybrid ID strategy
-- Author: Manus AI
-- Date: 2026-02-03
-- Note: users, draws, and affiliates already have short codes from previous migration

-- ============================================================================
-- 1. ADD SHORT CODE COLUMNS TO REMAINING TABLES
-- ============================================================================

-- Wheel Prizes: PRZ_CASH_001, PRZ_AIRTIME_042
ALTER TABLE wheel_prizes 
ADD COLUMN IF NOT EXISTS prize_code VARCHAR(20) UNIQUE;

-- Daily Subscriptions: SUB_1234_001
ALTER TABLE daily_subscriptions 
ADD COLUMN IF NOT EXISTS subscription_code VARCHAR(20) UNIQUE;

-- Spin Results: SPN_1234_20260203_01
ALTER TABLE spin_results 
ADD COLUMN IF NOT EXISTS spin_code VARCHAR(30) UNIQUE;

-- Draw Entries: ENT_1234_DRW001_01
ALTER TABLE draw_entries
ADD COLUMN IF NOT EXISTS entry_code VARCHAR(30) UNIQUE;

-- Transactions: Already has payment_reference (RCH_4567_1770118317)
-- No additional column needed

-- ============================================================================
-- 2. CREATE INDEXES FOR FAST LOOKUP
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_wheel_prizes_prize_code ON wheel_prizes(prize_code);
CREATE INDEX IF NOT EXISTS idx_daily_subscriptions_subscription_code ON daily_subscriptions(subscription_code);
CREATE INDEX IF NOT EXISTS idx_spin_results_spin_code ON spin_results(spin_code);
CREATE INDEX IF NOT EXISTS idx_draw_entries_entry_code ON draw_entries(entry_code);

-- ============================================================================
-- 3. CREATE SEQUENCE GENERATORS
-- ============================================================================

CREATE SEQUENCE IF NOT EXISTS prize_code_seq START 1 INCREMENT 1;
CREATE SEQUENCE IF NOT EXISTS subscription_code_seq START 1 INCREMENT 1;
CREATE SEQUENCE IF NOT EXISTS entry_code_seq START 1 INCREMENT 1;

-- ============================================================================
-- 4. CREATE HELPER FUNCTIONS FOR SHORT CODE GENERATION
-- ============================================================================

-- Function: Generate Prize Code (PRZ_CASH_001)
CREATE OR REPLACE FUNCTION generate_prize_code(prize_name TEXT)
RETURNS VARCHAR(20) AS $$
DECLARE
    category VARCHAR(10);
    next_val INTEGER;
    code VARCHAR(20);
BEGIN
    -- Extract category from prize name (first 4 chars, uppercase, letters only)
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
    FROM spin_results
    WHERE spin_code LIKE 'SPN_' || user_last4 || '_' || date_str || '_%';
    
    code := 'SPN_' || user_last4 || '_' || date_str || '_' || LPAD(sequence_num::TEXT, 2, '0');
    RETURN code;
END;
$$ LANGUAGE plpgsql;

-- Function: Generate Entry Code (ENT_1234_DRW001_01)
CREATE OR REPLACE FUNCTION generate_entry_code(user_id_param UUID, draw_id_param UUID)
RETURNS VARCHAR(30) AS $$
DECLARE
    user_last4 VARCHAR(4);
    draw_short VARCHAR(6);
    sequence_num INTEGER;
    code VARCHAR(30);
BEGIN
    -- Get last 4 digits of user's user_code
    SELECT RIGHT(user_code, 4) INTO user_last4
    FROM users WHERE id = user_id_param;
    
    -- Get draw code suffix (last 6 chars)
    SELECT RIGHT(draw_code, 6) INTO draw_short
    FROM draws WHERE id = draw_id_param;
    
    IF user_last4 IS NULL THEN
        user_last4 := '0000';
    END IF;
    
    IF draw_short IS NULL THEN
        draw_short := '000000';
    END IF;
    
    -- Get sequence for this user in this draw
    SELECT COALESCE(MAX(CAST(SUBSTRING(entry_code FROM '[0-9]+$') AS INTEGER)), 0) + 1
    INTO sequence_num
    FROM draw_entries
    WHERE entry_code LIKE 'ENT_' || user_last4 || '_' || draw_short || '_%';
    
    code := 'ENT_' || user_last4 || '_' || draw_short || '_' || LPAD(sequence_num::TEXT, 2, '0');
    RETURN code;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 5. CREATE TRIGGERS FOR AUTOMATIC SHORT CODE GENERATION
-- ============================================================================

-- Trigger: Auto-generate prize_code on wheel_prizes insert
CREATE OR REPLACE FUNCTION trigger_generate_prize_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.prize_code IS NULL THEN
        NEW.prize_code := generate_prize_code(NEW.name);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS auto_generate_prize_code ON wheel_prizes;
CREATE TRIGGER auto_generate_prize_code
    BEFORE INSERT ON wheel_prizes
    FOR EACH ROW
    EXECUTE FUNCTION trigger_generate_prize_code();

-- Trigger: Auto-generate subscription_code on daily_subscriptions insert
CREATE OR REPLACE FUNCTION trigger_generate_subscription_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.subscription_code IS NULL THEN
        NEW.subscription_code := generate_subscription_code(NEW.user_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS auto_generate_subscription_code ON daily_subscriptions;
CREATE TRIGGER auto_generate_subscription_code
    BEFORE INSERT ON daily_subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION trigger_generate_subscription_code();

-- Trigger: Auto-generate spin_code on spin_results insert
CREATE OR REPLACE FUNCTION trigger_generate_spin_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.spin_code IS NULL THEN
        NEW.spin_code := generate_spin_code(NEW.user_id, NEW.created_at);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS auto_generate_spin_code ON spin_results;
CREATE TRIGGER auto_generate_spin_code
    BEFORE INSERT ON spin_results
    FOR EACH ROW
    EXECUTE FUNCTION trigger_generate_spin_code();

-- Trigger: Auto-generate entry_code on draw_entries insert
CREATE OR REPLACE FUNCTION trigger_generate_entry_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.entry_code IS NULL THEN
        NEW.entry_code := generate_entry_code(NEW.user_id, NEW.draw_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS auto_generate_entry_code ON draw_entries;
CREATE TRIGGER auto_generate_entry_code
    BEFORE INSERT ON draw_entries
    FOR EACH ROW
    EXECUTE FUNCTION trigger_generate_entry_code();

-- ============================================================================
-- 6. BACKFILL EXISTING RECORDS (Using CTEs to avoid window function issues)
-- ============================================================================

-- Backfill users (if any are missing user_code)
WITH numbered_users AS (
    SELECT id, ROW_NUMBER() OVER (ORDER BY created_at) as rn
    FROM users
    WHERE user_code IS NULL
)
UPDATE users u
SET 
    user_code = 'USR_' || LPAD(nu.rn::TEXT, 4, '0'),
    display_name = '@recharge_' || LPAD(nu.rn::TEXT, 4, '0')
FROM numbered_users nu
WHERE u.id = nu.id;

-- Backfill draws (if any are missing draw_code)
WITH numbered_draws AS (
    SELECT 
        id,
        draw_time,
        ROW_NUMBER() OVER (PARTITION BY DATE_TRUNC('month', draw_time) ORDER BY created_at) as rn
    FROM draws
    WHERE draw_code IS NULL
)
UPDATE draws d
SET draw_code = 'DRW_' || TO_CHAR(nd.draw_time, 'YYYY_MM') || '_' || LPAD(nd.rn::TEXT, 3, '0')
FROM numbered_draws nd
WHERE d.id = nd.id;

-- Backfill wheel_prizes
WITH numbered_prizes AS (
    SELECT id, name, ROW_NUMBER() OVER (ORDER BY created_at) as rn
    FROM wheel_prizes
    WHERE prize_code IS NULL
)
UPDATE wheel_prizes wp
SET prize_code = 'PRZ_' || 
                 UPPER(SUBSTRING(REGEXP_REPLACE(np.name, '[^a-zA-Z]', '', 'g'), 1, 4)) || '_' || 
                 LPAD(np.rn::TEXT, 3, '0')
FROM numbered_prizes np
WHERE wp.id = np.id;

-- Backfill daily_subscriptions
WITH numbered_subs AS (
    SELECT 
        ds.id,
        ds.user_id,
        COALESCE(RIGHT(u.user_code, 4), '0000') as user_last4,
        ROW_NUMBER() OVER (PARTITION BY ds.user_id ORDER BY ds.created_at) as rn
    FROM daily_subscriptions ds
    LEFT JOIN users u ON ds.user_id = u.id
    WHERE ds.subscription_code IS NULL
)
UPDATE daily_subscriptions ds
SET subscription_code = 'SUB_' || ns.user_last4 || '_' || LPAD(ns.rn::TEXT, 3, '0')
FROM numbered_subs ns
WHERE ds.id = ns.id;

-- Backfill spin_results
WITH numbered_spins AS (
    SELECT 
        sr.id,
        sr.user_id,
        sr.created_at,
        COALESCE(RIGHT(u.user_code, 4), '0000') as user_last4,
        ROW_NUMBER() OVER (PARTITION BY sr.user_id, DATE(sr.created_at) ORDER BY sr.created_at) as rn
    FROM spin_results sr
    LEFT JOIN users u ON sr.user_id = u.id
    WHERE sr.spin_code IS NULL
)
UPDATE spin_results sr
SET spin_code = 'SPN_' || ns.user_last4 || '_' || 
                TO_CHAR(ns.created_at, 'YYYYMMDD') || '_' || 
                LPAD(ns.rn::TEXT, 2, '0')
FROM numbered_spins ns
WHERE sr.id = ns.id;

-- Backfill draw_entries
WITH numbered_entries AS (
    SELECT 
        de.id,
        de.user_id,
        de.draw_id,
        COALESCE(RIGHT(u.user_code, 4), '0000') as user_last4,
        COALESCE(RIGHT(d.draw_code, 6), '000000') as draw_short,
        ROW_NUMBER() OVER (PARTITION BY de.user_id, de.draw_id ORDER BY de.created_at) as rn
    FROM draw_entries de
    LEFT JOIN users u ON de.user_id = u.id
    LEFT JOIN draws d ON de.draw_id = d.id
    WHERE de.entry_code IS NULL
)
UPDATE draw_entries de
SET entry_code = 'ENT_' || ne.user_last4 || '_' || ne.draw_short || '_' || LPAD(ne.rn::TEXT, 2, '0')
FROM numbered_entries ne
WHERE de.id = ne.id;

-- ============================================================================
-- 7. ADD CONSTRAINTS
-- ============================================================================

-- Add check constraints for format validation
ALTER TABLE wheel_prizes ADD CONSTRAINT chk_prize_code_format 
    CHECK (prize_code ~ '^PRZ_[A-Z]{1,10}_[0-9]{3}$');

ALTER TABLE daily_subscriptions ADD CONSTRAINT chk_subscription_code_format 
    CHECK (subscription_code ~ '^SUB_[0-9]{4}_[0-9]{3}$');

ALTER TABLE spin_results ADD CONSTRAINT chk_spin_code_format 
    CHECK (spin_code ~ '^SPN_[0-9]{4}_[0-9]{8}_[0-9]{2}$');

ALTER TABLE draw_entries ADD CONSTRAINT chk_entry_code_format 
    CHECK (entry_code ~ '^ENT_[0-9]{4}_[0-9A-Z]{6}_[0-9]{2}$');

-- ============================================================================
-- 8. GRANT PERMISSIONS
-- ============================================================================

GRANT SELECT, UPDATE ON wheel_prizes TO rechargemax;
GRANT SELECT, UPDATE ON daily_subscriptions TO rechargemax;
GRANT SELECT, UPDATE ON spin_results TO rechargemax;
GRANT SELECT, UPDATE ON draw_entries TO rechargemax;
GRANT USAGE, SELECT ON prize_code_seq TO rechargemax;
GRANT USAGE, SELECT ON subscription_code_seq TO rechargemax;
GRANT USAGE, SELECT ON entry_code_seq TO rechargemax;

-- ============================================================================
-- 9. VERIFY MIGRATION
-- ============================================================================

DO $$
DECLARE
    user_count INTEGER;
    draw_count INTEGER;
    prize_count INTEGER;
    sub_count INTEGER;
    spin_count INTEGER;
    entry_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO user_count FROM users WHERE user_code IS NOT NULL;
    SELECT COUNT(*) INTO draw_count FROM draws WHERE draw_code IS NOT NULL;
    SELECT COUNT(*) INTO prize_count FROM wheel_prizes WHERE prize_code IS NOT NULL;
    SELECT COUNT(*) INTO sub_count FROM daily_subscriptions WHERE subscription_code IS NOT NULL;
    SELECT COUNT(*) INTO spin_count FROM spin_results WHERE spin_code IS NOT NULL;
    SELECT COUNT(*) INTO entry_count FROM draw_entries WHERE entry_code IS NOT NULL;
    
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Short Code Migration Complete!';
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Users with codes:         %', user_count;
    RAISE NOTICE 'Draws with codes:         %', draw_count;
    RAISE NOTICE 'Prizes with codes:        %', prize_count;
    RAISE NOTICE 'Subscriptions with codes: %', sub_count;
    RAISE NOTICE 'Spins with codes:         %', spin_count;
    RAISE NOTICE 'Entries with codes:       %', entry_count;
    RAISE NOTICE '============================================';
END $$;
