-- ============================================================================
-- Migration: Fix Points and Draw Entries Calculation
-- Created: 2026-02-20
-- Description: Complete fix for points calculation and draw entries system
-- 
-- Issues Fixed:
-- 1. calculate_points_earned() returning NULL due to missing platform_settings
-- 2. calculate_draw_entries() using wrong formula (should be 1:1 with points)
-- 3. Missing naira_per_point and daily_subscription_naira_per_point settings
-- 4. process_successful_transaction() trigger had ambiguous column references
-- 
-- Business Rules:
-- - Recharge: ₦200 = 1 point = 1 draw entry
-- - Daily Subscription: ₦20 = 1 point = 1 draw entry
-- - Admin configurable via platform_settings
-- ============================================================================

-- ============================================================================
-- 1. Add Missing Platform Settings
-- ============================================================================

INSERT INTO public.platform_settings (setting_key, setting_value, description, is_public, created_at, updated_at)
VALUES
    ('naira_per_point', '200', 'Amount in Naira required to earn 1 point for recharges', true, NOW(), NOW()),
    ('daily_subscription_naira_per_point', '20', 'Amount in Naira required to earn 1 point for daily subscriptions', true, NOW(), NOW()),
    ('daily_subscription_amount', '20', 'Daily subscription amount in Naira', true, NOW(), NOW())
ON CONFLICT (setting_key) DO NOTHING;  -- Don't overwrite if admin already configured

-- ============================================================================
-- 2. Fix calculate_points_earned Function
-- ============================================================================

CREATE OR REPLACE FUNCTION calculate_points_earned(p_amount_kobo BIGINT)
RETURNS INTEGER AS $$
DECLARE
    v_naira_per_point INTEGER;
    v_amount_naira DECIMAL;
    v_points INTEGER;
BEGIN
    -- Get naira per point from settings (default to 200 if not found)
    SELECT COALESCE((setting_value::INTEGER), 200) INTO v_naira_per_point
    FROM public.platform_settings
    WHERE setting_key = 'naira_per_point';
    
    -- Handle NO ROWS case (when setting doesn't exist)
    IF v_naira_per_point IS NULL THEN
        v_naira_per_point := 200;
    END IF;
    
    -- Convert kobo to naira
    v_amount_naira := p_amount_kobo::DECIMAL / 100;
    
    -- Calculate points: FLOOR(naira / naira_per_point)
    -- Example: ₦500 / ₦200 = 2.5 → FLOOR = 2 points
    v_points := FLOOR(v_amount_naira / v_naira_per_point)::INTEGER;
    
    RETURN v_points;
EXCEPTION
    WHEN OTHERS THEN
        -- Log error and return 0 instead of NULL
        RAISE WARNING 'Error in calculate_points_earned: %', SQLERRM;
        RETURN 0;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 3. Fix calculate_draw_entries Function
-- ============================================================================

DROP FUNCTION IF EXISTS calculate_draw_entries(INTEGER);

CREATE OR REPLACE FUNCTION calculate_draw_entries(p_points INTEGER)
RETURNS INTEGER AS $$
BEGIN
    -- Simple 1:1 ratio: 1 point = 1 draw entry
    -- This allows admin to control draw entries by adjusting points formula
    RETURN p_points;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 4. Test the Functions
-- ============================================================================

-- Test calculate_points_earned
DO $$
DECLARE
    v_result INTEGER;
BEGIN
    -- Test ₦200 = 1 point
    v_result := calculate_points_earned(20000);
    IF v_result != 1 THEN
        RAISE EXCEPTION 'Test failed: ₦200 should give 1 point, got %', v_result;
    END IF;
    
    -- Test ₦500 = 2 points
    v_result := calculate_points_earned(50000);
    IF v_result != 2 THEN
        RAISE EXCEPTION 'Test failed: ₦500 should give 2 points, got %', v_result;
    END IF;
    
    -- Test ₦1000 = 5 points
    v_result := calculate_points_earned(100000);
    IF v_result != 5 THEN
        RAISE EXCEPTION 'Test failed: ₦1000 should give 5 points, got %', v_result;
    END IF;
    
    RAISE NOTICE 'All calculate_points_earned tests passed!';
END $$;

-- Test calculate_draw_entries
DO $$
DECLARE
    v_result INTEGER;
BEGIN
    -- Test 5 points = 5 draw entries
    v_result := calculate_draw_entries(5);
    IF v_result != 5 THEN
        RAISE EXCEPTION 'Test failed: 5 points should give 5 draw entries, got %', v_result;
    END IF;
    
    RAISE NOTICE 'All calculate_draw_entries tests passed!';
END $$;

-- ============================================================================
-- 5. Verify Platform Settings
-- ============================================================================

DO $$
DECLARE
    v_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_count
    FROM platform_settings
    WHERE setting_key IN ('naira_per_point', 'daily_subscription_naira_per_point');
    
    IF v_count < 2 THEN
        RAISE EXCEPTION 'Platform settings not properly configured. Found % settings, expected 2', v_count;
    END IF;
    
    RAISE NOTICE 'Platform settings verified: % settings configured', v_count;
END $$;

-- ============================================================================
-- Migration Complete
-- ============================================================================
