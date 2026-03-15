-- Migration: Fix points calculation formula
-- Purpose: Correct points calculation to ₦200 = 1 point (rounded down)
-- Date: (see git history)

-- Drop old function with incorrect calculation
DROP FUNCTION IF EXISTS public.calculate_points_earned(numeric);

-- Create corrected function: ₦200 = 1 point (rounded down)
CREATE OR REPLACE FUNCTION public.calculate_points_earned(p_amount_kobo BIGINT)
RETURNS INTEGER AS $$
DECLARE
    v_naira_per_point INTEGER;
    v_amount_naira DECIMAL;
BEGIN
    -- Get naira per point from settings (default to 200)
    -- Setting: naira_per_point = 200 means ₦200 = 1 point
    SELECT COALESCE((setting_value::INTEGER), 200) INTO v_naira_per_point
    FROM public.platform_settings
    WHERE setting_key = 'naira_per_point';

    -- Convert kobo to naira (divide by 100)
    v_amount_naira := p_amount_kobo::DECIMAL / 100;

    -- Calculate points: FLOOR(amount in naira / naira per point)
    -- Examples:
    --   ₦199 (19,900 kobo) = FLOOR(199 / 200) = 0 points
    --   ₦200 (20,000 kobo) = FLOOR(200 / 200) = 1 point
    --   ₦399 (39,900 kobo) = FLOOR(399 / 200) = 1 point
    --   ₦400 (40,000 kobo) = FLOOR(400 / 200) = 2 points
    --   ₦2,000 (200,000 kobo) = FLOOR(2000 / 200) = 10 points
    RETURN FLOOR(v_amount_naira / v_naira_per_point)::INTEGER;
END;
$$ LANGUAGE plpgsql;

-- Add or update the naira_per_point setting
INSERT INTO platform_settings (setting_key, setting_value, description, created_at, updated_at)
VALUES ('naira_per_point', '200', 'Amount in Naira required to earn 1 point (₦200 = 1 point)', NOW(), NOW())
ON CONFLICT (setting_key) DO UPDATE 
SET setting_value = '200', 
    description = 'Amount in Naira required to earn 1 point (₦200 = 1 point)',
    updated_at = NOW();

-- Remove deprecated points_per_naira setting (inverse logic)
DELETE FROM platform_settings WHERE setting_key = 'points_per_naira';

-- Add comment for documentation
COMMENT ON FUNCTION calculate_points_earned IS 'Calculates loyalty points earned based on recharge amount. Formula: FLOOR(amount_naira / naira_per_point). Default: ₦200 = 1 point, rounded down.';
