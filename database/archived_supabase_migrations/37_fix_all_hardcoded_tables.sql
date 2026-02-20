-- Comprehensive fix for all hardcoded table names in database functions
-- This script replaces all instances of partitioned table names with base table names

-- Fix check_otp_rate_limit
CREATE OR REPLACE FUNCTION check_otp_rate_limit(p_msisdn TEXT, p_purpose TEXT)
RETURNS BOOLEAN AS $$
DECLARE
    recent_count INTEGER;
BEGIN
    SELECT COUNT(*)
    INTO recent_count
    FROM otps
    WHERE msisdn = p_msisdn
    AND purpose = p_purpose
    AND created_at > NOW() - INTERVAL '1 hour';
    
    RETURN recent_count < 5; -- Max 5 OTPs per hour
END;
$$ LANGUAGE plpgsql;

-- Fix verify_otp
CREATE OR REPLACE FUNCTION verify_otp(p_msisdn TEXT, p_otp_code TEXT, p_purpose TEXT)
RETURNS UUID AS $$
DECLARE
    v_user_id UUID;
    v_otp RECORD;
BEGIN
    -- Get the OTP record
    SELECT * INTO v_otp
    FROM otps
    WHERE msisdn = p_msisdn
    AND otp_code = p_otp_code
    AND purpose = p_purpose
    AND is_used = false
    AND expires_at > NOW()
    ORDER BY created_at DESC
    LIMIT 1;
    
    IF NOT FOUND THEN
        RETURN NULL;
    END IF;
    
    -- Mark OTP as used
    UPDATE otps
    SET is_used = true, used_at = NOW()
    WHERE id = v_otp.id;
    
    -- Get or create user
    SELECT id INTO v_user_id
    FROM users
    WHERE msisdn = p_msisdn;
    
    IF NOT FOUND THEN
        INSERT INTO users (msisdn, is_verified, created_at, updated_at)
        VALUES (p_msisdn, true, NOW(), NOW())
        RETURNING id INTO v_user_id;
    END IF;
    
    RETURN v_user_id;
END;
$$ LANGUAGE plpgsql;

-- Fix get_user_id
CREATE OR REPLACE FUNCTION get_user_id(p_msisdn TEXT)
RETURNS UUID AS $$
DECLARE
    v_user_id UUID;
BEGIN
    SELECT id INTO v_user_id
    FROM users
    WHERE msisdn = p_msisdn;
    
    RETURN v_user_id;
END;
$$ LANGUAGE plpgsql;

-- Fix upsert_user_profile
CREATE OR REPLACE FUNCTION upsert_user_profile(
    p_auth_user_id UUID,
    p_msisdn TEXT,
    p_full_name TEXT,
    p_email TEXT
)
RETURNS UUID AS $$
DECLARE
    v_user_id UUID;
BEGIN
    -- Check if user exists by auth_user_id
    SELECT id INTO v_user_id
    FROM users
    WHERE auth_user_id = p_auth_user_id;
    
    IF FOUND THEN
        -- Update existing user
        UPDATE users
        SET 
            msisdn = COALESCE(p_msisdn, msisdn),
            full_name = COALESCE(p_full_name, full_name),
            email = COALESCE(p_email, email),
            updated_at = NOW()
        WHERE id = v_user_id;
    ELSE
        -- Check if user exists by MSISDN
        SELECT id INTO v_user_id
        FROM users
        WHERE msisdn = p_msisdn;
        
        IF FOUND THEN
            -- Update existing user with auth_user_id
            UPDATE users
            SET 
                auth_user_id = p_auth_user_id,
                full_name = COALESCE(p_full_name, full_name),
                email = COALESCE(p_email, email),
                updated_at = NOW()
            WHERE id = v_user_id;
        ELSE
            -- Create new user
            INSERT INTO users (auth_user_id, msisdn, full_name, email, is_verified, created_at, updated_at)
            VALUES (p_auth_user_id, p_msisdn, p_full_name, p_email, true, NOW(), NOW())
            RETURNING id INTO v_user_id;
        END IF;
    END IF;
    
    RETURN v_user_id;
END;
$$ LANGUAGE plpgsql;

-- Fix process_affiliate_commission
CREATE OR REPLACE FUNCTION process_affiliate_commission(p_transaction_id UUID, p_affiliate_code TEXT)
RETURNS VOID AS $$
DECLARE
    v_transaction RECORD;
    v_affiliate RECORD;
    v_commission_amount INTEGER;
BEGIN
    -- Get transaction
    SELECT * INTO v_transaction
    FROM transactions
    WHERE id = p_transaction_id;
    
    IF NOT FOUND THEN
        RETURN;
    END IF;
    
    -- Get affiliate
    SELECT * INTO v_affiliate
    FROM affiliates
    WHERE affiliate_code = p_affiliate_code
    AND status = 'APPROVED';
    
    IF NOT FOUND THEN
        RETURN;
    END IF;
    
    -- Calculate commission
    v_commission_amount := FLOOR((v_transaction.amount * v_affiliate.commission_rate / 100));
    
    -- Insert commission record
    INSERT INTO affiliate_commissions (
        affiliate_id,
        transaction_id,
        commission_amount,
        commission_rate,
        transaction_amount,
        status,
        created_at
    ) VALUES (
        v_affiliate.id,
        p_transaction_id,
        v_commission_amount,
        v_affiliate.commission_rate,
        v_transaction.amount,
        'PENDING',
        NOW()
    );
    
    -- Update affiliate totals
    UPDATE affiliates
    SET 
        total_commission = total_commission + v_commission_amount,
        updated_at = NOW()
    WHERE id = v_affiliate.id;
END;
$$ LANGUAGE plpgsql;

-- Fix log_payment_event
CREATE OR REPLACE FUNCTION log_payment_event(
    p_transaction_id UUID,
    p_user_id UUID,
    p_event_type TEXT,
    p_payment_reference TEXT,
    p_request_payload JSONB,
    p_response_payload JSONB,
    p_status_code INTEGER,
    p_error_message TEXT,
    p_ip_address INET,
    p_user_agent TEXT,
    p_amount NUMERIC,
    p_is_successful BOOLEAN
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO payment_logs (
        transaction_id,
        user_id,
        event_type,
        payment_reference,
        request_payload,
        response_payload,
        status_code,
        error_message,
        ip_address,
        user_agent,
        amount,
        is_successful,
        created_at
    ) VALUES (
        p_transaction_id,
        p_user_id,
        p_event_type,
        p_payment_reference,
        p_request_payload,
        p_response_payload,
        p_status_code,
        p_error_message,
        p_ip_address,
        p_user_agent,
        p_amount,
        p_is_successful,
        NOW()
    );
END;
$$ LANGUAGE plpgsql;

-- Fix log_admin_action
CREATE OR REPLACE FUNCTION log_admin_action(
    p_admin_user_id UUID,
    p_session_id UUID,
    p_action TEXT,
    p_resource TEXT,
    p_resource_id TEXT,
    p_method TEXT,
    p_endpoint TEXT,
    p_request_data JSONB,
    p_response_status INTEGER,
    p_ip_address INET,
    p_user_agent TEXT
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO admin_activity_logs (
        admin_user_id,
        session_id,
        action,
        resource,
        resource_id,
        method,
        endpoint,
        request_data,
        response_status,
        ip_address,
        user_agent,
        created_at
    ) VALUES (
        p_admin_user_id,
        p_session_id,
        p_action,
        p_resource,
        p_resource_id,
        p_method,
        p_endpoint,
        p_request_data,
        p_response_status,
        p_ip_address,
        p_user_agent,
        NOW()
    );
END;
$$ LANGUAGE plpgsql;

-- Fix get_user_wallet_balance
CREATE OR REPLACE FUNCTION get_user_wallet_balance(p_user_id UUID)
RETURNS INTEGER AS $$
DECLARE
    v_balance INTEGER;
BEGIN
    SELECT balance INTO v_balance
    FROM wallets
    WHERE user_id = p_user_id;
    
    RETURN COALESCE(v_balance, 0);
END;
$$ LANGUAGE plpgsql;

-- Fix mark_notification_read
CREATE OR REPLACE FUNCTION mark_notification_read(p_notification_id UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE notifications
    SET 
        is_read = true,
        read_at = NOW(),
        updated_at = NOW()
    WHERE id = p_notification_id;
END;
$$ LANGUAGE plpgsql;

-- Fix get_unread_notification_count
CREATE OR REPLACE FUNCTION get_unread_notification_count(p_user_id UUID)
RETURNS INTEGER AS $$
DECLARE
    v_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_count
    FROM notifications
    WHERE user_id = p_user_id
    AND is_read = false;
    
    RETURN v_count;
END;
$$ LANGUAGE plpgsql;

-- Fix cleanup_old_otps
CREATE OR REPLACE FUNCTION cleanup_old_otps()
RETURNS INTEGER AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    DELETE FROM otps
    WHERE created_at < NOW() - INTERVAL '24 hours';
    
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Fix cleanup_old_payment_logs
CREATE OR REPLACE FUNCTION cleanup_old_payment_logs()
RETURNS INTEGER AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    DELETE FROM payment_logs
    WHERE created_at < NOW() - INTERVAL '90 days';
    
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Fix cleanup_old_admin_logs
CREATE OR REPLACE FUNCTION cleanup_old_admin_logs()
RETURNS INTEGER AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    DELETE FROM admin_activity_logs
    WHERE created_at < NOW() - INTERVAL '180 days';
    
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

SELECT 'All critical hardcoded table names fixed' AS status;
