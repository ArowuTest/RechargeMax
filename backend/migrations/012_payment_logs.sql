-- ============================================================================
-- PAYMENT LOGS TABLE
-- Migration: 12
-- Date: 2026-01-30
-- Purpose: Comprehensive audit trail for all payment operations and API calls
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.payment_logs (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- References
    transaction_id UUID REFERENCES public.transactions(id) ON DELETE SET NULL,
    user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    
    -- Event details
    event_type TEXT NOT NULL CHECK (
        event_type IN (
            'INITIALIZE', 'VERIFY', 'CALLBACK', 'WEBHOOK',
            'REFUND', 'DISPUTE', 'CHARGEBACK', 'RETRY'
        )
    ),
    payment_provider TEXT DEFAULT 'PAYSTACK',
    payment_reference TEXT,
    
    -- Request/Response data
    request_payload JSONB,
    response_payload JSONB,
    status_code INTEGER,
    error_message TEXT,
    error_code TEXT,
    
    -- Request tracking
    ip_address INET,
    user_agent TEXT,
    request_id TEXT,
    
    -- Timing
    response_time_ms INTEGER,
    
    -- Payment details
    amount DECIMAL(12,2),
    currency TEXT DEFAULT 'NGN',
    payment_method TEXT,
    
    -- Status
    is_successful BOOLEAN,
    is_retry BOOLEAN DEFAULT false,
    retry_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX idx_payment_logs_transaction_id ON public.payment_logs(transaction_id);
CREATE INDEX idx_payment_logs_user_id ON public.payment_logs(user_id);
CREATE INDEX idx_payment_logs_event_type ON public.payment_logs(event_type);
CREATE INDEX idx_payment_logs_payment_reference ON public.payment_logs(payment_reference);
CREATE INDEX idx_payment_logs_created_at ON public.payment_logs(created_at);
CREATE INDEX idx_payment_logs_is_successful ON public.payment_logs(is_successful);
CREATE INDEX idx_payment_logs_status_code ON public.payment_logs(status_code);

-- Composite index for error tracking
CREATE INDEX idx_payment_logs_errors ON public.payment_logs(event_type, is_successful, created_at) 
    WHERE is_successful = false;

-- Index for slow requests
CREATE INDEX idx_payment_logs_slow_requests ON public.payment_logs(response_time_ms, created_at) 
    WHERE response_time_ms > 5000;

-- ============================================================================
-- FUNCTIONS
-- ============================================================================

-- Log payment event
CREATE OR REPLACE FUNCTION log_payment_event(
    p_transaction_id UUID,
    p_user_id UUID,
    p_event_type TEXT,
    p_payment_reference TEXT,
    p_request_payload JSONB DEFAULT NULL,
    p_response_payload JSONB DEFAULT NULL,
    p_status_code INTEGER DEFAULT NULL,
    p_error_message TEXT DEFAULT NULL,
    p_ip_address INET DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL,
    p_amount DECIMAL DEFAULT NULL,
    p_is_successful BOOLEAN DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
    v_log_id UUID;
BEGIN
    INSERT INTO public.payment_logs (
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
        is_successful
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
        p_is_successful
    ) RETURNING id INTO v_log_id;
    
    RETURN v_log_id;
END;
$$ LANGUAGE plpgsql;

-- Get payment event history
CREATE OR REPLACE FUNCTION get_payment_event_history(
    p_transaction_id UUID DEFAULT NULL,
    p_payment_reference TEXT DEFAULT NULL
)
RETURNS TABLE(
    id UUID,
    event_type TEXT,
    status_code INTEGER,
    is_successful BOOLEAN,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pl.id,
        pl.event_type,
        pl.status_code,
        pl.is_successful,
        pl.error_message,
        pl.created_at
    FROM public.payment_logs pl
    WHERE (p_transaction_id IS NULL OR pl.transaction_id = p_transaction_id)
    AND (p_payment_reference IS NULL OR pl.payment_reference = p_payment_reference)
    ORDER BY pl.created_at ASC;
END;
$$ LANGUAGE plpgsql;

-- Get payment error statistics
CREATE OR REPLACE FUNCTION get_payment_error_stats(
    p_hours INTEGER DEFAULT 24
)
RETURNS TABLE(
    total_requests BIGINT,
    failed_requests BIGINT,
    error_rate DECIMAL,
    most_common_error TEXT,
    error_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*)::BIGINT as total_requests,
        COUNT(*) FILTER (WHERE is_successful = false)::BIGINT as failed_requests,
        (COUNT(*) FILTER (WHERE is_successful = false)::DECIMAL / NULLIF(COUNT(*), 0) * 100) as error_rate,
        MODE() WITHIN GROUP (ORDER BY error_code) FILTER (WHERE is_successful = false) as most_common_error,
        COUNT(*) FILTER (WHERE error_code = MODE() WITHIN GROUP (ORDER BY error_code) FILTER (WHERE is_successful = false))::BIGINT as error_count
    FROM public.payment_logs
    WHERE created_at > NOW() - (p_hours || ' hours')::INTERVAL;
END;
$$ LANGUAGE plpgsql;

-- Get slow payment requests
CREATE OR REPLACE FUNCTION get_slow_payment_requests(
    p_threshold_ms INTEGER DEFAULT 5000,
    p_hours INTEGER DEFAULT 24
)
RETURNS TABLE(
    id UUID,
    event_type TEXT,
    payment_reference TEXT,
    response_time_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pl.id,
        pl.event_type,
        pl.payment_reference,
        pl.response_time_ms,
        pl.created_at
    FROM public.payment_logs pl
    WHERE pl.response_time_ms > p_threshold_ms
    AND pl.created_at > NOW() - (p_hours || ' hours')::INTERVAL
    ORDER BY pl.response_time_ms DESC;
END;
$$ LANGUAGE plpgsql;

-- Cleanup old payment logs
CREATE OR REPLACE FUNCTION cleanup_old_payment_logs(retention_days INTEGER DEFAULT 90)
RETURNS TABLE(deleted_count BIGINT) AS $$
DECLARE
    cutoff_date TIMESTAMP WITH TIME ZONE;
    rows_deleted BIGINT;
BEGIN
    cutoff_date := NOW() - (retention_days || ' days')::INTERVAL;
    
    -- Keep failed requests longer for debugging
    DELETE FROM public.payment_logs
    WHERE created_at < cutoff_date
    AND is_successful = true;
    
    GET DIAGNOSTICS rows_deleted = ROW_COUNT;
    
    RETURN QUERY SELECT rows_deleted;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE public.payment_logs IS 'Comprehensive audit trail for all payment operations and API calls';
COMMENT ON COLUMN public.payment_logs.event_type IS 'Type of payment event (INITIALIZE, VERIFY, CALLBACK, etc.)';
COMMENT ON COLUMN public.payment_logs.request_payload IS 'Full request data sent to payment provider';
COMMENT ON COLUMN public.payment_logs.response_payload IS 'Full response data received from payment provider';
COMMENT ON COLUMN public.payment_logs.response_time_ms IS 'Response time in milliseconds for performance monitoring';
COMMENT ON COLUMN public.payment_logs.is_retry IS 'Indicates if this is a retry attempt';
