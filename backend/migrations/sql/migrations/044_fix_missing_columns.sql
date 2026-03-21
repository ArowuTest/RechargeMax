-- Migration 044: Fix missing columns and sequences from incomplete earlier migrations
-- ─────────────────────────────────────────────────────────────────────────────
-- Fixes two runtime errors from the deployed logs:
--
-- 1. ERROR: column "next_retry_at" does not exist (SQLSTATE 42703)
--    subscription_billings was created by an older version of migration 041 that
--    didn't include all columns. CREATE TABLE IF NOT EXISTS never re-runs for an
--    existing table, so missing columns must be added via ADD COLUMN IF NOT EXISTS.
--
-- 2. ERROR: null value in column "id" of relation "provider_configs" (SQLSTATE 23502)
--    provider_configs.id is BIGINT referencing provider_configs_id_seq, but that
--    sequence was never created. The ALTER COLUMN SET DEFAULT silently failed.
-- ─────────────────────────────────────────────────────────────────────────────

-- ── 1. provider_configs: create the missing sequence ──────────────────────────
CREATE SEQUENCE IF NOT EXISTS public.provider_configs_id_seq
    AS bigint
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

-- Wire the sequence as the column default (idempotent ALTER)
ALTER TABLE public.provider_configs
    ALTER COLUMN id SET DEFAULT nextval('public.provider_configs_id_seq'::regclass);

-- Advance the sequence past any existing rows so the next INSERT doesn't collide
SELECT setval(
    'public.provider_configs_id_seq',
    GREATEST(COALESCE((SELECT MAX(id) FROM public.provider_configs), 0), 1)
);

-- ── 2. subscription_billings: add all columns that may be missing ─────────────
ALTER TABLE public.subscription_billings
    ADD COLUMN IF NOT EXISTS next_retry_at           TIMESTAMP,
    ADD COLUMN IF NOT EXISTS points_awarded          BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS processed_at            TIMESTAMP,
    ADD COLUMN IF NOT EXISTS paystack_transaction_id BIGINT,
    ADD COLUMN IF NOT EXISTS gateway_response        TEXT,
    ADD COLUMN IF NOT EXISTS failure_reason          TEXT,
    ADD COLUMN IF NOT EXISTS max_retries             INTEGER NOT NULL DEFAULT 3,
    ADD COLUMN IF NOT EXISTS entries_to_award        INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS points_to_award         INTEGER NOT NULL DEFAULT 1;

-- Ensure the retry index exists (depends on next_retry_at being present)
CREATE INDEX IF NOT EXISTS idx_sub_billings_retry
    ON public.subscription_billings (status, next_retry_at)
    WHERE status IN ('pending', 'attempted') AND next_retry_at IS NOT NULL;

-- ── 3. Re-seed provider_configs now that the sequence is fixed ────────────────
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
