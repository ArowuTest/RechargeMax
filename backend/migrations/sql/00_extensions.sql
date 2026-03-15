-- Create required PostgreSQL extensions
-- Must run before any table definitions that use uuid_generate_v4()

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create the update_updated_at_column trigger function used by many tables
CREATE OR REPLACE FUNCTION public.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Generic updated_at trigger (used by most tables)
CREATE OR REPLACE FUNCTION public.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

-- Affiliate-specific updated_at triggers
CREATE OR REPLACE FUNCTION public.update_affiliate_analytics_timestamp()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION public.update_affiliate_bank_account_timestamp()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION public.update_affiliate_payout_timestamp()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

-- OTP triggers
CREATE OR REPLACE FUNCTION public.update_otps_updated_at()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION public.mark_expired_otps()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.expires_at < NOW() AND NEW.is_used = false THEN
        NEW.is_used = true;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Spin tiers trigger
CREATE OR REPLACE FUNCTION public.update_spin_tiers_updated_at()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

-- Transaction limits trigger
CREATE OR REPLACE FUNCTION public.update_transaction_limits_timestamp()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

-- Draw entries trigger (updates draw entry count on draw)
CREATE OR REPLACE FUNCTION public.trigger_update_draw_entries()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

-- Ensure only one primary bank account per affiliate
CREATE OR REPLACE FUNCTION public.ensure_single_primary_bank_account()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_primary = true THEN
        UPDATE public.affiliate_bank_accounts
        SET is_primary = false
        WHERE affiliate_id = NEW.affiliate_id AND id != NEW.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
