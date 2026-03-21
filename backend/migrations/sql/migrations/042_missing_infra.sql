-- =============================================================
-- Migration 042: missing infrastructure
--
-- 1. msisdn_blacklist table (fraud_detection_service queries it)
-- 2. get_active_provider() function (telecom_service_integrated calls it)
-- 3. wheel_prizes: add variation_code + network_provider columns
--    and seed the correct VTPass variation codes
-- 4. Seed a default ACTIVE draw so platform stats / draw-entry
--    code paths don't generate noisy "record not found" logs
-- =============================================================

-- Ensure provider_configs sequence exists BEFORE the INSERT below.
-- Migration 044 also creates it, but migrations run alphabetically so 042
-- executes first. Creating it here (idempotent IF NOT EXISTS) guarantees
-- the INSERT has a working DEFAULT for the id column.
CREATE SEQUENCE IF NOT EXISTS public.provider_configs_id_seq
    AS bigint START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;

ALTER TABLE public.provider_configs
    ALTER COLUMN id SET DEFAULT nextval('public.provider_configs_id_seq'::regclass);

-- ─────────────────────────────────────────────────────────────
-- 1. msisdn_blacklist
-- ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS public.msisdn_blacklist (
    id          bigserial PRIMARY KEY,
    msisdn      varchar(20) NOT NULL,
    reason      text,
    blacklisted_by  varchar(100) DEFAULT 'system',
    is_active   boolean NOT NULL DEFAULT true,
    blacklisted_at  timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz,          -- NULL = permanent
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_msisdn_blacklist_msisdn
    ON public.msisdn_blacklist (msisdn)
    WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_msisdn_blacklist_active
    ON public.msisdn_blacklist (is_active);

-- ─────────────────────────────────────────────────────────────
-- 2. get_active_provider() PostgreSQL function
--    Returns the highest-priority active row from provider_configs
--    for a given (network, service_type) pair.
--    Falls back gracefully when provider_configs is empty.
-- ─────────────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION public.get_active_provider(
    p_network      text,
    p_service_type text
)
RETURNS TABLE (
    id            bigint,
    network       varchar(50),
    service_type  varchar(50),
    provider_mode varchar(50),
    provider_name varchar(100),
    priority      integer,
    config        jsonb
)
LANGUAGE sql
STABLE
AS $$
    SELECT
        pc.id,
        pc.network,
        pc.service_type,
        pc.provider_mode,
        pc.provider_name,
        pc.priority,
        pc.config
    FROM public.provider_configs pc
    WHERE pc.network       = UPPER(p_network)
      AND pc.service_type  = UPPER(p_service_type)
      AND pc.is_active     = true
    ORDER BY pc.priority ASC
    LIMIT 1;
$$;

-- Seed a VTPass VTU row for every network × service_type combination
-- so the function always returns a row in production.
-- ON CONFLICT DO NOTHING so existing admin-managed rows are preserved.
INSERT INTO public.provider_configs
    (network, service_type, provider_mode, provider_name, priority, config, is_active)
VALUES
    ('MTN',     'AIRTIME', 'VTU', 'VTPass', 1, '{"mode":"production"}'::jsonb, true),
    ('GLO',     'AIRTIME', 'VTU', 'VTPass', 1, '{"mode":"production"}'::jsonb, true),
    ('AIRTEL',  'AIRTIME', 'VTU', 'VTPass', 1, '{"mode":"production"}'::jsonb, true),
    ('9MOBILE', 'AIRTIME', 'VTU', 'VTPass', 1, '{"mode":"production"}'::jsonb, true),
    ('MTN',     'DATA',    'VTU', 'VTPass', 1, '{"mode":"production"}'::jsonb, true),
    ('GLO',     'DATA',    'VTU', 'VTPass', 1, '{"mode":"production"}'::jsonb, true),
    ('AIRTEL',  'DATA',    'VTU', 'VTPass', 1, '{"mode":"production"}'::jsonb, true),
    ('9MOBILE', 'DATA',    'VTU', 'VTPass', 1, '{"mode":"production"}'::jsonb, true)
ON CONFLICT ON CONSTRAINT unique_active_provider DO NOTHING;

-- ─────────────────────────────────────────────────────────────
-- 3. wheel_prizes: add variation_code + network_provider columns
--    so admins can override the hardcoded Go fallback map
-- ─────────────────────────────────────────────────────────────
ALTER TABLE public.wheel_prizes
    ADD COLUMN IF NOT EXISTS variation_code    text,
    ADD COLUMN IF NOT EXISTS network_provider  text;  -- NULL = applies to all networks

CREATE INDEX IF NOT EXISTS idx_wheel_prizes_variation
    ON public.wheel_prizes (network_provider, prize_value)
    WHERE variation_code IS NOT NULL;

-- Seed correct VTPass variation codes for existing DATA prizes.
-- These update only the rows that have no variation_code yet.
-- prize_value is stored in KOBO.  Codes match VTPass plan catalogue.
UPDATE public.wheel_prizes
SET
    variation_code   = CASE
        -- MTN
        WHEN network_provider = 'MTN'     AND prize_value = 10000  THEN 'mtn-100mb-100'
        WHEN network_provider = 'MTN'     AND prize_value = 20000  THEN 'mtn-200mb-200'
        WHEN network_provider = 'MTN'     AND prize_value = 50000  THEN 'mtn-500mb-500'
        WHEN network_provider = 'MTN'     AND prize_value = 100000 THEN 'mtn-1gb-1000'
        WHEN network_provider = 'MTN'     AND prize_value = 200000 THEN 'mtn-2gb-1200'
        -- GLO
        WHEN network_provider = 'GLO'     AND prize_value = 10000  THEN 'glo-100mb-100'
        WHEN network_provider = 'GLO'     AND prize_value = 50000  THEN 'glo-500mb-500'
        WHEN network_provider = 'GLO'     AND prize_value = 100000 THEN 'glo-1gb-1000'
        WHEN network_provider = 'GLO'     AND prize_value = 200000 THEN 'glo-2gb-2000'
        -- AIRTEL
        WHEN network_provider = 'AIRTEL'  AND prize_value = 10000  THEN 'airtel-100mb-100'
        WHEN network_provider = 'AIRTEL'  AND prize_value = 50000  THEN 'airtel-500mb-500'
        WHEN network_provider = 'AIRTEL'  AND prize_value = 100000 THEN 'airtel-1gb-1000'
        WHEN network_provider = 'AIRTEL'  AND prize_value = 200000 THEN 'airtel-2gb-2000'
        -- 9MOBILE
        WHEN network_provider = '9MOBILE' AND prize_value = 50000  THEN 'etisalat-500mb-500'
        WHEN network_provider = '9MOBILE' AND prize_value = 100000 THEN 'etisalat-1gb-1000'
        WHEN network_provider = '9MOBILE' AND prize_value = 200000 THEN 'etisalat-2gb-2000'
        ELSE variation_code   -- leave unchanged if no match
    END
WHERE prize_type = 'DATA'
  AND variation_code IS NULL;

-- ─────────────────────────────────────────────────────────────
-- 4. Seed a default ACTIVE draw
--    If no draw is running, platform stats silently shows null
--    and recharge draw-entry code logs noisy "record not found".
--    Insert a rolling monthly draw that stays active until the
--    admin creates a real one.
-- ─────────────────────────────────────────────────────────────
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM public.draws
        WHERE status = 'ACTIVE' AND end_time > now()
    ) THEN
        INSERT INTO public.draws (
            id,
            name,
            type,
            description,
            status,
            start_time,
            end_time,
            draw_time,
            prize_pool,
            winners_count,
            total_entries,
            draw_code,
            created_at,
            updated_at
        ) VALUES (
            gen_random_uuid(),
            'Monthly Grand Draw',
            'MONTHLY',
            'Recharge and win amazing prizes every month!',
            'ACTIVE',
            date_trunc('month', now()),                    -- 1st of current month
            date_trunc('month', now()) + INTERVAL '1 month' - INTERVAL '1 second', -- last second of month
            date_trunc('month', now()) + INTERVAL '1 month' - INTERVAL '1 second',
            500000.00,   -- ₦500,000 prize pool (admin can update)
            5,           -- 5 winners
            0,
            'DRAW-' || TO_CHAR(now(), 'YYYYMM'),
            now(),
            now()
        );
    END IF;
END;
$$;
