-- Migration 048: Add failed_attempts column to otps table (SEC-008)
-- Tracks brute-force attempts on OTP verification.
ALTER TABLE otps ADD COLUMN IF NOT EXISTS failed_attempts INTEGER NOT NULL DEFAULT 0;
