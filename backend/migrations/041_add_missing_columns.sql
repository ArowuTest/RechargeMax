-- Migration 041: Add missing columns that exist in the application but were absent from earlier migrations.

-- 1. Add draw_code column to draws table (used to generate human-readable draw references)
ALTER TABLE draws
    ADD COLUMN IF NOT EXISTS draw_code VARCHAR(20) UNIQUE;

CREATE INDEX IF NOT EXISTS idx_draws_draw_code ON draws(draw_code);

-- 2. Add network_provider column to data_plans (used for network-specific plan filtering)
ALTER TABLE data_plans
    ADD COLUMN IF NOT EXISTS network_provider TEXT;

CREATE INDEX IF NOT EXISTS idx_data_plans_network_provider ON data_plans(network_provider);

-- 3. Widen subscription_code column from VARCHAR(20) to VARCHAR(50)
--    Paystack subscription codes (SUB_xxxxxxxxxxxxxxxx) can exceed 20 characters.
ALTER TABLE daily_subscriptions
    ALTER COLUMN subscription_code TYPE VARCHAR(50);
