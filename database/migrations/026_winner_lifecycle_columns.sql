-- Migration: Add winner lifecycle columns to draw_winners
-- These columns are used by winner_service for prize claim/provision/notification lifecycle

ALTER TABLE draw_winners 
  ADD COLUMN IF NOT EXISTS first_name         TEXT,
  ADD COLUMN IF NOT EXISTS last_name          TEXT,
  ADD COLUMN IF NOT EXISTS prize_type         TEXT NOT NULL DEFAULT 'cash',
  ADD COLUMN IF NOT EXISTS prize_description  TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS data_package       TEXT,
  ADD COLUMN IF NOT EXISTS airtime_amount     BIGINT,
  ADD COLUMN IF NOT EXISTS network            TEXT,
  ADD COLUMN IF NOT EXISTS auto_provision     BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS provision_status   TEXT,
  ADD COLUMN IF NOT EXISTS provision_reference TEXT,
  ADD COLUMN IF NOT EXISTS provisioned_at     TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS provision_error    TEXT,
  ADD COLUMN IF NOT EXISTS claim_deadline     TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS payout_status      TEXT NOT NULL DEFAULT 'pending',
  ADD COLUMN IF NOT EXISTS payout_method      TEXT,
  ADD COLUMN IF NOT EXISTS bank_code          TEXT,
  ADD COLUMN IF NOT EXISTS bank_name          TEXT,
  ADD COLUMN IF NOT EXISTS account_number     TEXT,
  ADD COLUMN IF NOT EXISTS account_name       TEXT,
  ADD COLUMN IF NOT EXISTS payout_reference   TEXT,
  ADD COLUMN IF NOT EXISTS payout_error       TEXT,
  ADD COLUMN IF NOT EXISTS shipping_address   TEXT,
  ADD COLUMN IF NOT EXISTS shipping_phone     TEXT,
  ADD COLUMN IF NOT EXISTS shipping_status    TEXT,
  ADD COLUMN IF NOT EXISTS tracking_number    TEXT,
  ADD COLUMN IF NOT EXISTS shipped_at         TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS delivered_at       TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS notification_sent  BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS notification_sent_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS notification_channels TEXT,
  ADD COLUMN IF NOT EXISTS notes              TEXT,
  ADD COLUMN IF NOT EXISTS updated_at         TIMESTAMPTZ DEFAULT NOW();

-- prize_amount already exists as NUMERIC(10,2) but Winner entity uses BIGINT (kobo)
-- Keep numeric - just ensure it's compatible

-- Backfill prize_type for existing rows
UPDATE draw_winners SET prize_type = 'cash' WHERE prize_type IS NULL OR prize_type = '';

-- Backfill claim_deadline from expires_at where available
UPDATE draw_winners SET claim_deadline = expires_at WHERE claim_deadline IS NULL AND expires_at IS NOT NULL;

-- Add useful index for provision_status  
CREATE INDEX IF NOT EXISTS idx_draw_winners_provision ON draw_winners(provision_status) WHERE auto_provision = TRUE;
CREATE INDEX IF NOT EXISTS idx_draw_winners_payout ON draw_winners(payout_status);

COMMENT ON TABLE draw_winners IS 'Winner records with full prize claim/provision/notification lifecycle';
