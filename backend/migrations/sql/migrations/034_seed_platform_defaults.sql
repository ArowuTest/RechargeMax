-- Migration 034: Seed default platform_settings rows that are expected by background jobs
-- and application services. All inserts use ON CONFLICT DO NOTHING so existing admin
-- configuration is never overwritten.

INSERT INTO platform_settings (setting_key, setting_value, description, is_public, created_at, updated_at)
VALUES
    -- Commission release job: hold period before auto-approving affiliate commissions
    ('affiliate.commission_hold_days', '7',
     'Number of days before PENDING affiliate commissions are auto-approved', false, NOW(), NOW()),

    -- Reconciliation job: how many minutes before a PENDING transaction is considered stuck
    ('reconciliation.stuck_threshold_minutes', '30',
     'Minutes after which a PENDING transaction is reconciled against Paystack', false, NOW(), NOW()),

    -- Spin/draw defaults used by platform services
    ('spin.enabled', 'true',
     'Whether the spin-to-win feature is globally enabled', true, NOW(), NOW()),

    ('draw.entries_per_naira', '1',
     'Draw entries awarded per 100 naira recharged', true, NOW(), NOW()),

    -- Wallet / affiliate payout settings
    ('affiliate.minimum_payout_amount', '500',
     'Minimum balance (in kobo) before affiliate payout is processed', false, NOW(), NOW())

ON CONFLICT (setting_key) DO NOTHING;
