-- Migration: Create Spin Tiers Table
-- Purpose: Make spin tier system configurable by admin instead of hardcoded
-- Date: 2026-02-01
-- Author: Manus AI Agent

BEGIN;

-- ============================================
-- CREATE SPIN_TIERS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS spin_tiers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Tier details
    tier_name TEXT NOT NULL UNIQUE,
    tier_display_name TEXT NOT NULL,
    
    -- Amount thresholds (in kobo)
    min_daily_amount BIGINT NOT NULL,
    max_daily_amount BIGINT NOT NULL,
    
    -- Spins configuration
    spins_per_day INTEGER NOT NULL,
    
    -- Display configuration
    tier_color TEXT,
    tier_icon TEXT,
    tier_badge TEXT,
    description TEXT,
    
    -- Ordering and status
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID,
    updated_by UUID,
    
    -- Constraints
    CONSTRAINT positive_amounts CHECK (min_daily_amount >= 0 AND max_daily_amount > min_daily_amount),
    CONSTRAINT positive_spins CHECK (spins_per_day > 0),
    CONSTRAINT valid_sort_order CHECK (sort_order >= 0)
);

-- ============================================
-- CREATE INDEXES
-- ============================================

CREATE INDEX idx_spin_tiers_is_active ON spin_tiers(is_active);
CREATE INDEX idx_spin_tiers_sort_order ON spin_tiers(sort_order);
CREATE INDEX idx_spin_tiers_amount_range ON spin_tiers(min_daily_amount, max_daily_amount);

-- ============================================
-- SEED INITIAL TIER DATA
-- ============================================

-- Insert the current tier system (amounts in kobo)
INSERT INTO spin_tiers (
    tier_name,
    tier_display_name,
    min_daily_amount,
    max_daily_amount,
    spins_per_day,
    tier_color,
    tier_icon,
    tier_badge,
    description,
    sort_order,
    is_active
) VALUES
(
    'BRONZE',
    'Bronze',
    100000,     -- ₦1,000
    499999,     -- ₦4,999.99
    1,
    '#CD7F32',  -- Bronze color
    '🥉',
    'bronze-badge.svg',
    'Entry level tier for daily recharges between ₦1,000 and ₦4,999',
    1,
    true
),
(
    'SILVER',
    'Silver',
    500000,     -- ₦5,000
    999999,     -- ₦9,999.99
    2,
    '#C0C0C0',  -- Silver color
    '🥈',
    'silver-badge.svg',
    'Mid-level tier for daily recharges between ₦5,000 and ₦9,999',
    2,
    true
),
(
    'GOLD',
    'Gold',
    1000000,    -- ₦10,000
    1999999,    -- ₦19,999.99
    3,
    '#FFD700',  -- Gold color
    '🥇',
    'gold-badge.svg',
    'Premium tier for daily recharges between ₦10,000 and ₦19,999',
    3,
    true
),
(
    'PLATINUM',
    'Platinum',
    2000000,    -- ₦20,000
    4999999,    -- ₦49,999.99
    5,
    '#E5E4E2',  -- Platinum color
    '💎',
    'platinum-badge.svg',
    'Elite tier for daily recharges between ₦20,000 and ₦49,999',
    4,
    true
),
(
    'DIAMOND',
    'Diamond',
    5000000,    -- ₦50,000
    99999999999, -- Effectively unlimited
    10,
    '#B9F2FF',  -- Diamond color
    '💠',
    'diamond-badge.svg',
    'Ultimate tier for daily recharges of ₦50,000 and above',
    5,
    true
);

-- ============================================
-- CREATE TRIGGER FOR UPDATED_AT
-- ============================================

CREATE OR REPLACE FUNCTION update_spin_tiers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_spin_tiers_updated_at
    BEFORE UPDATE ON spin_tiers
    FOR EACH ROW
    EXECUTE FUNCTION update_spin_tiers_updated_at();

-- ============================================
-- VERIFICATION
-- ============================================

DO $$
DECLARE
    tier_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO tier_count FROM spin_tiers WHERE is_active = true;
    
    IF tier_count = 5 THEN
        RAISE NOTICE 'SUCCESS: Spin tiers table created and seeded with 5 tiers';
    ELSE
        RAISE EXCEPTION 'ERROR: Expected 5 active tiers, found %', tier_count;
    END IF;
END $$;

COMMIT;

-- ============================================
-- ROLLBACK SCRIPT
-- ============================================

-- To rollback this migration, run:
-- BEGIN;
-- DROP TRIGGER IF EXISTS trigger_spin_tiers_updated_at ON spin_tiers;
-- DROP FUNCTION IF EXISTS update_spin_tiers_updated_at();
-- DROP TABLE IF EXISTS spin_tiers CASCADE;
-- COMMIT;
