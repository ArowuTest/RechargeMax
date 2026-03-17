-- Migration 035: Fix OTP code column to support bcrypt hashes (SEC-009)
-- The code column was varchar(6) but bcrypt hashes are 60 characters.
-- This caused OTP creation to fail with a string-too-long error.
ALTER TABLE otps ALTER COLUMN code TYPE varchar(100);
