-- Migration 046: Fix valid_msisdn CHECK constraint — too restrictive regex
-- ─────────────────────────────────────────────────────────────────────────────
-- ROOT CAUSE:
--   The original CHECK uses: ^234[789][01][0-9]{8}$
--   This only allows 5th digit = 0 or 1.
--   Nigerian mobile numbers have 5th digits 0-9 (e.g. 0803x, 0813x, 0706x).
--   Any real user number like 08031234567 → 2348031234567 → 5th digit = 3
--   → CHECK VIOLATION → INSERT fails → 500 on every subscription/recharge.
--
-- CORRECT pattern: ^234[789][0-9]{9}$
--   234      = Nigeria country code
--   [789]    = valid mobile prefix (no landlines)
--   [0-9]{9} = any 9 remaining digits (positions 5-13)
--   Total: 13 characters (e.g. 2348031234567)
-- ─────────────────────────────────────────────────────────────────────────────

-- 1. daily_subscriptions
ALTER TABLE public.daily_subscriptions
    DROP CONSTRAINT IF EXISTS valid_msisdn;

ALTER TABLE public.daily_subscriptions
    ADD CONSTRAINT valid_msisdn
    CHECK (msisdn ~ '^234[789][0-9]{9}$');

-- 2. transactions
ALTER TABLE public.transactions
    DROP CONSTRAINT IF EXISTS valid_msisdn;

ALTER TABLE public.transactions
    ADD CONSTRAINT valid_msisdn
    CHECK (msisdn ~ '^234[789][0-9]{9}$');

-- 3. users
ALTER TABLE public.users
    DROP CONSTRAINT IF EXISTS valid_msisdn;

ALTER TABLE public.users
    ADD CONSTRAINT valid_msisdn
    CHECK (msisdn ~ '^234[789][0-9]{9}$');

-- 4. otp_verifications (slightly different pattern — same fix)
ALTER TABLE public.otp_verifications
    DROP CONSTRAINT IF EXISTS valid_msisdn_otp;

ALTER TABLE public.otp_verifications
    ADD CONSTRAINT valid_msisdn_otp
    CHECK (msisdn ~ '^(234|0)?[789][0-9]{9,10}$');
