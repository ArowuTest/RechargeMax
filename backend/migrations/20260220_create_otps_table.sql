-- Migration: Create OTPs table for phone number verification
-- Date: 2026-02-20
-- Description: Creates the otps table for storing one-time passwords used in authentication flow

-- Create OTPs table
CREATE TABLE IF NOT EXISTS otps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    msisdn VARCHAR(20) NOT NULL,
    code VARCHAR(6) NOT NULL,
    purpose VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_otps_msisdn ON otps(msisdn);
CREATE INDEX IF NOT EXISTS idx_otps_expires_at ON otps(expires_at);
CREATE INDEX IF NOT EXISTS idx_otps_code ON otps(code);
CREATE INDEX IF NOT EXISTS idx_otps_is_used ON otps(is_used);

-- Add comment to table
COMMENT ON TABLE otps IS 'Stores one-time passwords for phone number verification and authentication';

-- Add comments to columns
COMMENT ON COLUMN otps.id IS 'Unique identifier for the OTP record';
COMMENT ON COLUMN otps.msisdn IS 'Phone number in international format (e.g., 2348012345678)';
COMMENT ON COLUMN otps.code IS 'Six-digit OTP code sent to user';
COMMENT ON COLUMN otps.purpose IS 'Purpose of OTP: REGISTRATION, LOGIN, PASSWORD_RESET, etc.';
COMMENT ON COLUMN otps.expires_at IS 'Timestamp when OTP expires (typically 10 minutes from creation)';
COMMENT ON COLUMN otps.is_used IS 'Flag indicating if OTP has been used';
COMMENT ON COLUMN otps.used_at IS 'Timestamp when OTP was used';
COMMENT ON COLUMN otps.created_at IS 'Timestamp when OTP was created';
COMMENT ON COLUMN otps.updated_at IS 'Timestamp when OTP record was last updated';

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_otps_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_otps_updated_at
    BEFORE UPDATE ON otps
    FOR EACH ROW
    EXECUTE FUNCTION update_otps_updated_at();

-- Add cleanup function for expired OTPs (optional, for maintenance)
CREATE OR REPLACE FUNCTION cleanup_expired_otps()
RETURNS void AS $$
BEGIN
    DELETE FROM otps WHERE expires_at < NOW() - INTERVAL '24 hours';
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_otps() IS 'Removes OTP records that expired more than 24 hours ago';
