-- ============================================================================
-- ADMIN ACTIVITY LOGS TABLE
-- Migration: 09
-- Date: 2026-01-30
-- Purpose: Audit trail for all admin actions and API calls
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.admin_activity_logs_2026_01_30_14_00 (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Admin identification
    admin_user_id UUID REFERENCES public.admin_users_2026_01_30_14_00(id) ON DELETE SET NULL,
    admin_session_id UUID REFERENCES public.admin_sessions_2026_01_30_14_00(id) ON DELETE SET NULL,
    
    -- Action details
    action TEXT NOT NULL,
    resource TEXT,
    resource_id TEXT,
    
    -- HTTP details
    method TEXT,
    endpoint TEXT,
    request_data JSONB,
    response_status INTEGER,
    response_data JSONB,
    
    -- Request tracking
    ip_address INET,
    user_agent TEXT,
    duration_ms INTEGER,
    
    -- Security monitoring
    is_suspicious BOOLEAN DEFAULT false,
    risk_score INTEGER DEFAULT 0 CHECK (risk_score >= 0 AND risk_score <= 100),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX idx_admin_activity_logs_admin_user_id ON public.admin_activity_logs_2026_01_30_14_00(admin_user_id);
CREATE INDEX idx_admin_activity_logs_session_id ON public.admin_activity_logs_2026_01_30_14_00(admin_session_id);
CREATE INDEX idx_admin_activity_logs_action ON public.admin_activity_logs_2026_01_30_14_00(action);
CREATE INDEX idx_admin_activity_logs_resource ON public.admin_activity_logs_2026_01_30_14_00(resource);
CREATE INDEX idx_admin_activity_logs_resource_id ON public.admin_activity_logs_2026_01_30_14_00(resource_id);
CREATE INDEX idx_admin_activity_logs_created_at ON public.admin_activity_logs_2026_01_30_14_00(created_at);
CREATE INDEX idx_admin_activity_logs_is_suspicious ON public.admin_activity_logs_2026_01_30_14_00(is_suspicious) WHERE is_suspicious = true;
CREATE INDEX idx_admin_activity_logs_risk_score ON public.admin_activity_logs_2026_01_30_14_00(risk_score) WHERE risk_score > 50;

-- Composite index for security monitoring
CREATE INDEX idx_admin_activity_logs_security ON public.admin_activity_logs_2026_01_30_14_00(admin_user_id, created_at, is_suspicious);

-- ============================================================================
-- FUNCTIONS
-- ============================================================================

-- Log admin action
CREATE OR REPLACE FUNCTION log_admin_action(
    p_admin_user_id UUID,
    p_session_id UUID,
    p_action TEXT,
    p_resource TEXT,
    p_resource_id TEXT,
    p_method TEXT DEFAULT NULL,
    p_endpoint TEXT DEFAULT NULL,
    p_request_data JSONB DEFAULT NULL,
    p_response_status INTEGER DEFAULT NULL,
    p_ip_address INET DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
    v_log_id UUID;
BEGIN
    INSERT INTO public.admin_activity_logs_2026_01_30_14_00 (
        admin_user_id,
        admin_session_id,
        action,
        resource,
        resource_id,
        method,
        endpoint,
        request_data,
        response_status,
        ip_address,
        user_agent
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
        p_user_agent
    ) RETURNING id INTO v_log_id;
    
    RETURN v_log_id;
END;
$$ LANGUAGE plpgsql;

-- Get admin activity summary
CREATE OR REPLACE FUNCTION get_admin_activity_summary(
    p_admin_user_id UUID,
    p_days INTEGER DEFAULT 30
)
RETURNS TABLE(
    total_actions BIGINT,
    suspicious_actions BIGINT,
    avg_risk_score NUMERIC,
    most_common_action TEXT,
    action_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*)::BIGINT as total_actions,
        COUNT(*) FILTER (WHERE is_suspicious = true)::BIGINT as suspicious_actions,
        AVG(risk_score)::NUMERIC as avg_risk_score,
        MODE() WITHIN GROUP (ORDER BY action) as most_common_action,
        COUNT(*) FILTER (WHERE action = MODE() WITHIN GROUP (ORDER BY action))::BIGINT as action_count
    FROM public.admin_activity_logs_2026_01_30_14_00
    WHERE admin_user_id = p_admin_user_id
    AND created_at > NOW() - (p_days || ' days')::INTERVAL;
END;
$$ LANGUAGE plpgsql;

-- Cleanup old logs
CREATE OR REPLACE FUNCTION cleanup_old_admin_logs(retention_days INTEGER DEFAULT 90)
RETURNS TABLE(deleted_count BIGINT) AS $$
DECLARE
    cutoff_date TIMESTAMP WITH TIME ZONE;
    rows_deleted BIGINT;
BEGIN
    cutoff_date := NOW() - (retention_days || ' days')::INTERVAL;
    
    DELETE FROM public.admin_activity_logs_2026_01_30_14_00
    WHERE created_at < cutoff_date
    AND is_suspicious = false
    AND risk_score < 50;
    
    GET DIAGNOSTICS rows_deleted = ROW_COUNT;
    
    RETURN QUERY SELECT rows_deleted;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE public.admin_activity_logs_2026_01_30_14_00 IS 'Audit trail for all admin actions and API calls';
COMMENT ON COLUMN public.admin_activity_logs_2026_01_30_14_00.action IS 'Action performed (e.g., CREATE, UPDATE, DELETE, APPROVE)';
COMMENT ON COLUMN public.admin_activity_logs_2026_01_30_14_00.resource IS 'Resource type (e.g., USER, AFFILIATE, DRAW, PRIZE)';
COMMENT ON COLUMN public.admin_activity_logs_2026_01_30_14_00.resource_id IS 'ID of the affected resource';
COMMENT ON COLUMN public.admin_activity_logs_2026_01_30_14_00.is_suspicious IS 'Flag for suspicious activity detection';
COMMENT ON COLUMN public.admin_activity_logs_2026_01_30_14_00.risk_score IS 'Risk score (0-100) for security monitoring';
