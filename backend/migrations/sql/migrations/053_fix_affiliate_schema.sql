-- ═══════════════════════════════════════════════════════════════════════════
-- Migration 053: Fix affiliate schema gaps
-- ═══════════════════════════════════════════════════════════════════════════
-- Problems addressed:
--  1. affiliates.click_count   — column missing from 09_affiliates.sql
--  2. affiliates.commission_rate constraint — no floor/ceiling enforced
--  3. affiliate_commissions.commission_amount/transaction_amount — DB uses
--     numeric(10,2) but Go entity declares bigint (stores kobo integers).
--     Migrate to bigint so arithmetic stays integer throughout.
--  4. affiliates.referral_code — entity declares it, migration never added it.
--     We standardise on affiliate_code for sharing; referral_code kept as an
--     internal alias populated from user.referral_code at approval time.
--  5. users.referred_by — index missing (used in every commission lookup).
--  6. platform_settings seed — ensure affiliate commission defaults exist.
-- ═══════════════════════════════════════════════════════════════════════════

-- ── 1. Add click_count to affiliates ───────────────────────────────────────
ALTER TABLE public.affiliates
    ADD COLUMN IF NOT EXISTS click_count INTEGER NOT NULL DEFAULT 0;

-- ── 2. Add referral_code to affiliates ─────────────────────────────────────
ALTER TABLE public.affiliates
    ADD COLUMN IF NOT EXISTS referral_code TEXT;

-- back-fill from users where the join is possible
UPDATE public.affiliates a
SET    referral_code = u.referral_code
FROM   public.users u
WHERE  a.user_id = u.id
  AND  a.referral_code IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_affiliates_referral_code
    ON public.affiliates (referral_code)
    WHERE referral_code IS NOT NULL;

-- ── 3. Enforce commission_rate floor 0.50 % and ceiling 1.50 % ─────────────
-- Drop old permissive constraint first (may not exist — ignore error)
ALTER TABLE public.affiliates
    DROP CONSTRAINT IF EXISTS positive_commission_rate;

ALTER TABLE public.affiliates
    ADD CONSTRAINT affiliate_commission_rate_range
        CHECK (commission_rate >= 0.50 AND commission_rate <= 1.50);

-- Reset existing out-of-range values to the new default
UPDATE public.affiliates
SET    commission_rate = 1.00
WHERE  commission_rate > 1.50 OR commission_rate < 0.50;

-- ── 4. Fix affiliate_commissions column types kobo-safe ────────────────────
-- commission_amount and transaction_amount store kobo (integer cents).
-- numeric(10,2) works but causes implicit casting; bigint is the correct type.
ALTER TABLE public.affiliate_commissions
    ALTER COLUMN commission_amount   TYPE BIGINT USING commission_amount::BIGINT,
    ALTER COLUMN transaction_amount  TYPE BIGINT USING transaction_amount::BIGINT;

-- Remove the numeric > 0 constraints and re-add them for bigint
ALTER TABLE public.affiliate_commissions
    DROP CONSTRAINT IF EXISTS positive_commission_amount,
    DROP CONSTRAINT IF EXISTS positive_transaction_amount;

ALTER TABLE public.affiliate_commissions
    ADD CONSTRAINT positive_commission_amount  CHECK (commission_amount  > 0),
    ADD CONSTRAINT positive_transaction_amount CHECK (transaction_amount > 0);

-- ── 5. Index users.referred_by (used on every commission lookup) ────────────
CREATE INDEX IF NOT EXISTS idx_users_referred_by
    ON public.users (referred_by)
    WHERE referred_by IS NOT NULL;

-- ── 6. Platform settings — affiliate commission defaults ────────────────────
INSERT INTO public.platform_settings (setting_key, setting_value, description, created_at, updated_at)
VALUES
    ('affiliate.commission_rate_percent',  '1.00',   'Default affiliate commission rate (%). Range 0.50–1.50', NOW(), NOW()),
    ('affiliate.commission_rate_floor',    '0.50',   'Minimum affiliate commission rate (%) — hard floor',     NOW(), NOW()),
    ('affiliate.commission_rate_ceiling',  '1.50',   'Maximum affiliate commission rate (%) — hard ceiling',   NOW(), NOW()),
    ('affiliate.min_payout_ngn',           '1000',   'Minimum affiliate payout threshold in Naira',            NOW(), NOW()),
    ('affiliate.payout_day',               'MONDAY', 'Day of week admin is notified to run weekly payouts',    NOW(), NOW())
ON CONFLICT (setting_key) DO UPDATE
    SET setting_value = EXCLUDED.setting_value,
        updated_at    = NOW();
