-- ============================================================================
-- API LOGS AND WEBHOOK LOGS TABLES
-- Purpose: Monitor external API calls and incoming webhooks
-- ============================================================================

-- API Logs Table
CREATE TABLE IF NOT EXISTS public.api_logs_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Service identification
    service_name TEXT NOT NULL, -- 'PAYSTACK', 'MTN', 'AIRTEL', 'TERMII', etc.
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL CHECK (method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE')),
    
    -- Request details
    request_url TEXT,
    request_headers JSONB,
    request_payload JSONB,
    
    -- Response details
    response_status_code INTEGER,
    response_headers JSONB,
    response_payload JSONB,
    
    -- Performance
    response_time_ms INTEGER,
    
    -- Error tracking
    is_error BOOLEAN DEFAULT false,
    error_message TEXT,
    error_code TEXT,
    
    -- Context
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id),
    transaction_reference TEXT,
    ip_address TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Metadata
    metadata JSONB
);

-- Webhook Logs Table
CREATE TABLE IF NOT EXISTS public.webhook_logs_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Source identification
    source TEXT NOT NULL, -- 'PAYSTACK', 'MTN', 'AIRTEL', etc.
    event_type TEXT NOT NULL,
    
    -- Webhook data
    payload JSONB NOT NULL,
    headers JSONB,
    signature TEXT,
    
    -- Verification
    is_verified BOOLEAN DEFAULT false,
    verification_method TEXT,
    verification_error TEXT,
    
    -- Processing
    is_processed BOOLEAN DEFAULT false,
    processing_error TEXT,
    processing_attempts INTEGER DEFAULT 0,
    max_processing_attempts INTEGER DEFAULT 3,
    
    -- Related records
    transaction_reference TEXT,
    related_transaction_id UUID,
    
    -- Timestamps
    received_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE,
    processed_at TIMESTAMP WITH TIME ZONE,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    
    -- Request metadata
    ip_address TEXT,
    user_agent TEXT,
    
    -- Metadata
    metadata JSONB,
    
    CONSTRAINT valid_processing_attempts CHECK (processing_attempts <= max_processing_attempts)
);

-- Indexes for api_logs
CREATE INDEX idx_api_logs_service ON public.api_logs_2026_01_30_14_00(service_name);
CREATE INDEX idx_api_logs_created ON public.api_logs_2026_01_30_14_00(created_at DESC);
CREATE INDEX idx_api_logs_error ON public.api_logs_2026_01_30_14_00(is_error) WHERE is_error = true;
CREATE INDEX idx_api_logs_user ON public.api_logs_2026_01_30_14_00(user_id);
CREATE INDEX idx_api_logs_reference ON public.api_logs_2026_01_30_14_00(transaction_reference);
CREATE INDEX idx_api_logs_status ON public.api_logs_2026_01_30_14_00(response_status_code);

-- Indexes for webhook_logs
CREATE INDEX idx_webhook_source ON public.webhook_logs_2026_01_30_14_00(source);
CREATE INDEX idx_webhook_event ON public.webhook_logs_2026_01_30_14_00(event_type);
CREATE INDEX idx_webhook_received ON public.webhook_logs_2026_01_30_14_00(received_at DESC);
CREATE INDEX idx_webhook_processed ON public.webhook_logs_2026_01_30_14_00(is_processed);
CREATE INDEX idx_webhook_verified ON public.webhook_logs_2026_01_30_14_00(is_verified);
CREATE INDEX idx_webhook_reference ON public.webhook_logs_2026_01_30_14_00(transaction_reference);
CREATE INDEX idx_webhook_retry ON public.webhook_logs_2026_01_30_14_00(next_retry_at) WHERE is_processed = false AND next_retry_at IS NOT NULL;

-- Function to get API error rate by service
CREATE OR REPLACE FUNCTION get_api_error_rate(
    p_service_name TEXT,
    p_hours INTEGER DEFAULT 24
)
RETURNS DECIMAL(5,2) AS $$
DECLARE
    v_total INTEGER;
    v_errors INTEGER;
    v_error_rate DECIMAL(5,2);
BEGIN
    SELECT 
        COUNT(*),
        COUNT(*) FILTER (WHERE is_error = true)
    INTO v_total, v_errors
    FROM public.api_logs_2026_01_30_14_00
    WHERE service_name = p_service_name
    AND created_at >= NOW() - (p_hours || ' hours')::INTERVAL;
    
    IF v_total = 0 THEN
        RETURN 0;
    END IF;
    
    v_error_rate := (v_errors::DECIMAL / v_total::DECIMAL) * 100;
    RETURN ROUND(v_error_rate, 2);
END;
$$ LANGUAGE plpgsql;

-- Function to get average API response time
CREATE OR REPLACE FUNCTION get_api_avg_response_time(
    p_service_name TEXT,
    p_hours INTEGER DEFAULT 24
)
RETURNS INTEGER AS $$
DECLARE
    v_avg_time INTEGER;
BEGIN
    SELECT COALESCE(AVG(response_time_ms)::INTEGER, 0) INTO v_avg_time
    FROM public.api_logs_2026_01_30_14_00
    WHERE service_name = p_service_name
    AND created_at >= NOW() - (p_hours || ' hours')::INTERVAL
    AND is_error = false;
    
    RETURN v_avg_time;
END;
$$ LANGUAGE plpgsql;

-- Function to update webhook processing status
CREATE OR REPLACE FUNCTION update_webhook_processing()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_processed = true AND OLD.is_processed = false THEN
        NEW.processed_at = NOW();
    ELSIF NEW.is_verified = true AND OLD.is_verified = false THEN
        NEW.verified_at = NOW();
    END IF;
    
    -- Set next retry time if processing failed
    IF NEW.processing_error IS NOT NULL AND NEW.processing_attempts < NEW.max_processing_attempts THEN
        NEW.next_retry_at = NOW() + (POWER(2, NEW.processing_attempts) || ' minutes')::INTERVAL;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_webhook_processing
    BEFORE UPDATE ON public.webhook_logs_2026_01_30_14_00
    FOR EACH ROW
    EXECUTE FUNCTION update_webhook_processing();

COMMENT ON TABLE public.api_logs_2026_01_30_14_00 IS 'Comprehensive logging of all external API calls';
COMMENT ON TABLE public.webhook_logs_2026_01_30_14_00 IS 'Logging and processing of incoming webhooks';
