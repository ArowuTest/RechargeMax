-- ============================================================================
-- OTP VERIFICATIONS TABLE
-- Migration: 08
-- Date: 2026-01-30
-- Purpose: Secure OTP verification for authentication and sensitive operations
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.otp_verifications (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- User identification
    msisdn TEXT NOT NULL,
    user_id UUID REFERENCES public.users(id) ON DELETE CASCADE,
    
    -- OTP details
    otp_code_hash TEXT NOT NULL,
    purpose TEXT NOT NULL DEFAULT 'LOGIN' CHECK (
        purpose IN (
            'LOGIN', 'REGISTRATION', 'PASSWORD_RESET',
            'TRANSACTION_VERIFICATION', 'PHONE_VERIFICATION',
            'WITHDRAWAL', 'PROFILE_UPDATE', 'TWO_FACTOR_AUTH'
        )
    ),
    
    -- Verification status
    is_verified BOOLEAN DEFAULT false,
    is_expired BOOLEAN DEFAULT false,
    is_revoked BOOLEAN DEFAULT false,
    
    -- Security tracking
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 5,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    
    -- IP and device tracking
    request_ip INET,
    request_user_agent TEXT,
    device_fingerprint TEXT,
    
    -- Verification details
    verified_at TIMESTAMP WITH TIME ZONE,
    verified_ip INET,
    verified_user_agent TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Constraints
    CONSTRAINT valid_msisdn_otp CHECK (msisdn ~ '^(234|0)?[789][01][0-9]{8,9}$'),
    CONSTRAINT valid_attempts CHECK (attempts >= 0 AND attempts <= max_attempts),
    CONSTRAINT valid_expiry CHECK (expires_at > created_at)
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX idx_otp_verifications_msisdn ON public.otp_verifications(msisdn);
CREATE INDEX idx_otp_verifications_user_id ON public.otp_verifications(user_id);
CREATE INDEX idx_otp_verifications_purpose ON public.otp_verifications(purpose);
CREATE INDEX idx_otp_verifications_is_verified ON public.otp_verifications(is_verified);
CREATE INDEX idx_otp_verifications_expires_at ON public.otp_verifications(expires_at) WHERE is_verified = false;
CREATE INDEX idx_otp_verifications_created_at ON public.otp_verifications(created_at);

-- Composite index for active OTP lookup
CREATE INDEX idx_otp_verifications_active_lookup ON public.otp_verifications(msisdn, purpose, is_verified, expires_at) 
    WHERE is_verified = false AND is_expired = false AND is_revoked = false;

-- ============================================================================
-- FUNCTIONS
-- ============================================================================

-- Generate OTP code
CREATE OR REPLACE FUNCTION generate_otp_code(length INTEGER DEFAULT 6)
RETURNS TEXT AS $$
DECLARE
    otp_code TEXT;
    i INTEGER;
BEGIN
    otp_code := '';
    FOR i IN 1..length LOOP
        otp_code := otp_code || floor(random() * 10)::TEXT;
    END LOOP;
    RETURN otp_code;
END;
$$ LANGUAGE plpgsql;

-- Check rate limit
CREATE OR REPLACE FUNCTION check_otp_rate_limit(
    p_msisdn TEXT,
    p_purpose TEXT,
    p_time_window_minutes INTEGER DEFAULT 5,
    p_max_requests INTEGER DEFAULT 3
)
RETURNS BOOLEAN AS $$
DECLARE
    request_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO request_count
    FROM public.otp_verifications
    WHERE msisdn = p_msisdn
    AND purpose = p_purpose
    AND created_at > NOW() - (p_time_window_minutes || ' minutes')::INTERVAL;
    
    RETURN request_count < p_max_requests;
END;
$$ LANGUAGE plpgsql;

-- Verify OTP
CREATE OR REPLACE FUNCTION verify_otp(
    p_msisdn TEXT,
    p_otp_code TEXT,
    p_purpose TEXT
)
RETURNS TABLE(
    success BOOLEAN,
    message TEXT,
    otp_id UUID,
    user_id UUID
) AS $$
DECLARE
    v_otp RECORD;
    v_otp_hash TEXT;
BEGIN
    -- Hash the provided OTP
    v_otp_hash := encode(digest(p_otp_code, 'sha256'), 'hex');
    
    -- Find matching OTP
    SELECT * INTO v_otp
    FROM public.otp_verifications
    WHERE msisdn = p_msisdn
    AND purpose = p_purpose
    AND otp_code_hash = v_otp_hash
    AND is_verified = false
    AND is_expired = false
    AND is_revoked = false
    AND expires_at > NOW()
    ORDER BY created_at DESC
    LIMIT 1;
    
    IF NOT FOUND THEN
        RETURN QUERY SELECT false, 'Invalid or expired OTP'::TEXT, NULL::UUID, NULL::UUID;
        RETURN;
    END IF;
    
    IF v_otp.attempts >= v_otp.max_attempts THEN
        RETURN QUERY SELECT false, 'Maximum attempts exceeded'::TEXT, v_otp.id, v_otp.user_id;
        RETURN;
    END IF;
    
    -- Mark as verified
    UPDATE public.otp_verifications
    SET is_verified = true,
        verified_at = NOW(),
        attempts = attempts + 1
    WHERE id = v_otp.id;
    
    RETURN QUERY SELECT true, 'OTP verified successfully'::TEXT, v_otp.id, v_otp.user_id;
END;
$$ LANGUAGE plpgsql;

-- Cleanup old OTPs
CREATE OR REPLACE FUNCTION cleanup_old_otps(retention_days INTEGER DEFAULT 30)
RETURNS TABLE(deleted_count BIGINT) AS $$
DECLARE
    cutoff_date TIMESTAMP WITH TIME ZONE;
    rows_deleted BIGINT;
BEGIN
    cutoff_date := NOW() - (retention_days || ' days')::INTERVAL;
    
    DELETE FROM public.otp_verifications
    WHERE created_at < cutoff_date
    AND (is_verified = true OR is_expired = true OR is_revoked = true);
    
    GET DIAGNOSTICS rows_deleted = ROW_COUNT;
    
    RETURN QUERY SELECT rows_deleted;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Automatic expiry trigger
CREATE OR REPLACE FUNCTION mark_expired_otps()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.expires_at < NOW() AND NEW.is_expired = false THEN
        NEW.is_expired := true;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_mark_expired_otps
    BEFORE UPDATE ON public.otp_verifications
    FOR EACH ROW
    EXECUTE FUNCTION mark_expired_otps();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE public.otp_verifications IS 'Stores OTP verification codes for authentication and sensitive operations';
COMMENT ON COLUMN public.otp_verifications.otp_code_hash IS 'SHA-256 hash of OTP code for security';
COMMENT ON COLUMN public.otp_verifications.purpose IS 'Purpose of OTP: LOGIN, REGISTRATION, TRANSACTION_VERIFICATION, etc.';
COMMENT ON COLUMN public.otp_verifications.attempts IS 'Number of verification attempts';
COMMENT ON COLUMN public.otp_verifications.max_attempts IS 'Maximum allowed attempts before blocking';
