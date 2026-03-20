-- Migration 041: Fix spin_tier ranges to match intended design
--
-- Intended tier model (cumulative daily recharge → daily spin cap):
--   Bronze:   ₦1,000  – ₦2,499   → 1 spin/day
--   Silver:   ₦2,500  – ₦4,999   → 2 spins/day
--   Gold:     ₦5,000  – ₦9,999   → 3 spins/day
--   Platinum: ₦10,000+            → 5 spins/day  (no upper cap)
--
-- Problems fixed:
--   1. Platinum max_daily_amount was 9,999,999 kobo (₦99,999).
--      Anyone recharging ₦100,000+ fell off the tier table entirely and
--      received the conservative fallback of 1 spin instead of 5.
--      Fix: raise Platinum ceiling to 999,999,999,999 kobo (effectively unlimited).
--
--   2. The hardcoded SpinTiers slice in spin_tier_calculator.go had completely
--      different ranges (Bronze up to ₦4,999, Silver ₦5,000–₦9,999, etc.)
--      which was corrected in the same PR to match this DB definition.
--
-- All amounts are in KOBO (1 NGN = 100 kobo).

-- Fix Platinum upper bound to be effectively unlimited
UPDATE spin_tiers
SET    max_daily_amount = 999999999999,  -- ~₦10 billion, covers any realistic recharge
       description      = 'Recharge ₦10,000+ to earn 5 spins per day',
       updated_at       = NOW()
WHERE  tier_name = 'platinum'
  AND  max_daily_amount = 9999999;       -- only touch the old incorrect value

-- Ensure Bronze/Silver/Gold ranges are also correct (idempotent)
UPDATE spin_tiers SET min_daily_amount = 100000,  max_daily_amount = 249999,  updated_at = NOW() WHERE tier_name = 'bronze';
UPDATE spin_tiers SET min_daily_amount = 250000,  max_daily_amount = 499999,  updated_at = NOW() WHERE tier_name = 'silver';
UPDATE spin_tiers SET min_daily_amount = 500000,  max_daily_amount = 999999,  updated_at = NOW() WHERE tier_name = 'gold';
UPDATE spin_tiers SET min_daily_amount = 1000000, max_daily_amount = 999999999999, updated_at = NOW() WHERE tier_name = 'platinum';

-- Remove Diamond tier if it exists (not part of the intended design)
DELETE FROM spin_tiers WHERE tier_name = 'diamond';
