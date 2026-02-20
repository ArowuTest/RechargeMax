-- Migration: Provider Configuration System
-- Description: Add provider configuration for admin-controlled switching between VTU, Direct, and Simulation modes
-- Date: 2026-02-02

-- =====================================================
-- 1. CREATE PROVIDER CONFIGURATION TABLE
-- =====================================================

CREATE TABLE IF NOT EXISTS provider_configurations (
    id BIGSERIAL PRIMARY KEY,
    network VARCHAR(20) NOT NULL, -- MTN, GLO, AIRTEL, 9MOBILE, or 'GLOBAL' for all networks
    service_type VARCHAR(20) NOT NULL, -- AIRTIME, DATA, or 'ALL' for both
    provider_mode VARCHAR(20) NOT NULL DEFAULT 'SIMULATION', -- VTU, DIRECT, SIMULATION
    provider_name VARCHAR(50), -- vtpass, direct_mtn, direct_glo, etc.
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 1, -- For fallback ordering
    config JSONB, -- Provider-specific configuration (API keys, endpoints, etc.)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    
    CONSTRAINT provider_configurations_network_check CHECK (network IN ('MTN', 'GLO', 'AIRTEL', '9MOBILE', 'GLOBAL')),
    CONSTRAINT provider_configurations_service_type_check CHECK (service_type IN ('AIRTIME', 'DATA', 'ALL')),
    CONSTRAINT provider_configurations_provider_mode_check CHECK (provider_mode IN ('VTU', 'DIRECT', 'SIMULATION')),
    CONSTRAINT provider_configurations_priority_check CHECK (priority > 0),
    
    -- Unique constraint: one config per network+service_type+provider_mode combination
    UNIQUE (network, service_type, provider_mode)
);

CREATE INDEX idx_provider_configurations_network ON provider_configurations(network);
CREATE INDEX idx_provider_configurations_enabled ON provider_configurations(is_enabled);
CREATE INDEX idx_provider_configurations_mode ON provider_configurations(provider_mode);
CREATE INDEX idx_provider_configurations_priority ON provider_configurations(priority);

COMMENT ON TABLE provider_configurations IS 'Configuration for recharge providers with admin-controlled mode switching';
COMMENT ON COLUMN provider_configurations.network IS 'Network provider (MTN, GLO, AIRTEL, 9MOBILE) or GLOBAL for all';
COMMENT ON COLUMN provider_configurations.service_type IS 'Service type (AIRTIME, DATA) or ALL for both';
COMMENT ON COLUMN provider_configurations.provider_mode IS 'Provider mode: VTU (aggregator), DIRECT (network), SIMULATION (testing)';
COMMENT ON COLUMN provider_configurations.provider_name IS 'Specific provider name (vtpass, direct_mtn, etc.)';
COMMENT ON COLUMN provider_configurations.priority IS 'Priority for fallback (1 = highest)';
COMMENT ON COLUMN provider_configurations.config IS 'Provider-specific configuration (API keys, endpoints, etc.)';

-- =====================================================
-- 2. CREATE PROVIDER TRANSACTION LOG
-- =====================================================

CREATE TABLE IF NOT EXISTS provider_transaction_logs (
    id BIGSERIAL PRIMARY KEY,
    transaction_id BIGINT NOT NULL,
    provider_config_id BIGINT NOT NULL REFERENCES provider_configurations(id),
    provider_mode VARCHAR(20) NOT NULL,
    provider_name VARCHAR(50),
    request_payload JSONB,
    response_payload JSONB,
    status VARCHAR(20) NOT NULL, -- SUCCESS, FAILED, PENDING, TIMEOUT
    error_message TEXT,
    response_time_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT provider_transaction_logs_status_check CHECK (status IN ('SUCCESS', 'FAILED', 'PENDING', 'TIMEOUT', 'REVERSED'))
);

CREATE INDEX idx_provider_transaction_logs_transaction_id ON provider_transaction_logs(transaction_id);
CREATE INDEX idx_provider_transaction_logs_provider_config_id ON provider_transaction_logs(provider_config_id);
CREATE INDEX idx_provider_transaction_logs_status ON provider_transaction_logs(status);
CREATE INDEX idx_provider_transaction_logs_created_at ON provider_transaction_logs(created_at DESC);

COMMENT ON TABLE provider_transaction_logs IS 'Audit log for all provider transactions';
COMMENT ON COLUMN provider_transaction_logs.response_time_ms IS 'Provider API response time in milliseconds';

-- =====================================================
-- 3. INSERT DEFAULT CONFIGURATIONS
-- =====================================================

-- Global default: Simulation mode for all networks and services
INSERT INTO provider_configurations (network, service_type, provider_mode, provider_name, is_enabled, priority, config)
VALUES 
    ('GLOBAL', 'ALL', 'SIMULATION', 'simulation', true, 3, '{"success_rate": 0.95}'::jsonb)
ON CONFLICT (network, service_type, provider_mode) DO NOTHING;

-- VTPass configuration for all networks (disabled by default until API keys are configured)
INSERT INTO provider_configurations (network, service_type, provider_mode, provider_name, is_enabled, priority, config)
VALUES 
    ('GLOBAL', 'ALL', 'VTU', 'vtpass', false, 2, '{
        "api_key": "",
        "public_key": "",
        "secret_key": "",
        "is_sandbox": true,
        "base_url": "https://sandbox.vtpass.com/api"
    }'::jsonb)
ON CONFLICT (network, service_type, provider_mode) DO NOTHING;

-- Direct network integration placeholders (disabled by default until partnerships are signed)
INSERT INTO provider_configurations (network, service_type, provider_mode, provider_name, is_enabled, priority, config)
VALUES 
    ('MTN', 'ALL', 'DIRECT', 'direct_mtn', false, 1, '{
        "api_key": "",
        "api_secret": "",
        "base_url": "",
        "commission_rate": 0.05
    }'::jsonb),
    ('GLO', 'ALL', 'DIRECT', 'direct_glo', false, 1, '{
        "api_key": "",
        "api_secret": "",
        "base_url": "",
        "commission_rate": 0.05
    }'::jsonb),
    ('AIRTEL', 'ALL', 'DIRECT', 'direct_airtel', false, 1, '{
        "api_key": "",
        "api_secret": "",
        "base_url": "",
        "commission_rate": 0.05
    }'::jsonb),
    ('9MOBILE', 'ALL', 'DIRECT', 'direct_9mobile', false, 1, '{
        "api_key": "",
        "api_secret": "",
        "base_url": "",
        "commission_rate": 0.05
    }'::jsonb)
ON CONFLICT (network, service_type, provider_mode) DO NOTHING;

-- =====================================================
-- 4. CREATE FUNCTION TO GET ACTIVE PROVIDER
-- =====================================================

CREATE OR REPLACE FUNCTION get_active_provider(
    p_network VARCHAR(20),
    p_service_type VARCHAR(20)
)
RETURNS TABLE (
    id BIGINT,
    network VARCHAR(20),
    service_type VARCHAR(20),
    provider_mode VARCHAR(20),
    provider_name VARCHAR(50),
    priority INTEGER,
    config JSONB
) AS $$
BEGIN
    -- Try to find network-specific configuration first
    RETURN QUERY
    SELECT 
        pc.id,
        pc.network,
        pc.service_type,
        pc.provider_mode,
        pc.provider_name,
        pc.priority,
        pc.config
    FROM provider_configurations pc
    WHERE pc.is_enabled = true
      AND (pc.network = p_network OR pc.network = 'GLOBAL')
      AND (pc.service_type = p_service_type OR pc.service_type = 'ALL')
    ORDER BY 
        CASE WHEN pc.network = p_network THEN 0 ELSE 1 END, -- Network-specific first
        pc.priority ASC, -- Then by priority
        pc.id ASC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_active_provider IS 'Get the active provider configuration for a network and service type';

-- =====================================================
-- 5. CREATE FUNCTION TO LOG PROVIDER TRANSACTION
-- =====================================================

CREATE OR REPLACE FUNCTION log_provider_transaction(
    p_transaction_id BIGINT,
    p_provider_config_id BIGINT,
    p_provider_mode VARCHAR(20),
    p_provider_name VARCHAR(50),
    p_request_payload JSONB,
    p_response_payload JSONB,
    p_status VARCHAR(20),
    p_error_message TEXT,
    p_response_time_ms INTEGER
)
RETURNS BIGINT AS $$
DECLARE
    v_log_id BIGINT;
BEGIN
    INSERT INTO provider_transaction_logs (
        transaction_id,
        provider_config_id,
        provider_mode,
        provider_name,
        request_payload,
        response_payload,
        status,
        error_message,
        response_time_ms
    ) VALUES (
        p_transaction_id,
        p_provider_config_id,
        p_provider_mode,
        p_provider_name,
        p_request_payload,
        p_response_payload,
        p_status,
        p_error_message,
        p_response_time_ms
    )
    RETURNING id INTO v_log_id;
    
    RETURN v_log_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION log_provider_transaction IS 'Log a provider transaction for audit and monitoring';

-- =====================================================
-- 6. CREATE VIEW FOR PROVIDER ANALYTICS
-- =====================================================

CREATE OR REPLACE VIEW provider_analytics AS
SELECT 
    pc.network,
    pc.service_type,
    pc.provider_mode,
    pc.provider_name,
    COUNT(ptl.id) as total_transactions,
    COUNT(CASE WHEN ptl.status = 'SUCCESS' THEN 1 END) as successful_transactions,
    COUNT(CASE WHEN ptl.status = 'FAILED' THEN 1 END) as failed_transactions,
    COUNT(CASE WHEN ptl.status = 'PENDING' THEN 1 END) as pending_transactions,
    ROUND(
        COUNT(CASE WHEN ptl.status = 'SUCCESS' THEN 1 END)::NUMERIC / 
        NULLIF(COUNT(ptl.id), 0) * 100, 
        2
    ) as success_rate_percent,
    AVG(ptl.response_time_ms) as avg_response_time_ms,
    MAX(ptl.response_time_ms) as max_response_time_ms,
    MIN(ptl.response_time_ms) as min_response_time_ms
FROM provider_configurations pc
LEFT JOIN provider_transaction_logs ptl ON pc.id = ptl.provider_config_id
WHERE ptl.created_at >= CURRENT_DATE - INTERVAL '30 days' OR ptl.id IS NULL
GROUP BY pc.id, pc.network, pc.service_type, pc.provider_mode, pc.provider_name;

COMMENT ON VIEW provider_analytics IS 'Provider performance analytics for the last 30 days';

-- =====================================================
-- 7. CREATE AUDIT LOG TRIGGER
-- =====================================================

CREATE OR REPLACE FUNCTION audit_provider_configuration_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_logs (
            table_name,
            record_id,
            action,
            old_data,
            new_data,
            changed_by,
            changed_at
        ) VALUES (
            'provider_configurations',
            NEW.id,
            'UPDATE',
            row_to_json(OLD),
            row_to_json(NEW),
            NEW.updated_by,
            CURRENT_TIMESTAMP
        );
    ELSIF TG_OP = 'INSERT' THEN
        INSERT INTO audit_logs (
            table_name,
            record_id,
            action,
            new_data,
            changed_by,
            changed_at
        ) VALUES (
            'provider_configurations',
            NEW.id,
            'INSERT',
            row_to_json(NEW),
            NEW.created_by,
            CURRENT_TIMESTAMP
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_audit_provider_configuration_changes
AFTER INSERT OR UPDATE ON provider_configurations
FOR EACH ROW
EXECUTE FUNCTION audit_provider_configuration_changes();

-- =====================================================
-- 8. GRANT PERMISSIONS
-- =====================================================

-- Grant permissions to application user (adjust username as needed)
-- GRANT SELECT, INSERT, UPDATE ON provider_configurations TO rechargemax_app;
-- GRANT SELECT, INSERT ON provider_transaction_logs TO rechargemax_app;
-- GRANT SELECT ON provider_analytics TO rechargemax_app;
-- GRANT EXECUTE ON FUNCTION get_active_provider TO rechargemax_app;
-- GRANT EXECUTE ON FUNCTION log_provider_transaction TO rechargemax_app;

-- =====================================================
-- MIGRATION COMPLETE
-- =====================================================

-- Verify installation
SELECT 'Provider Configuration System installed successfully' as status;
SELECT COUNT(*) as default_configs_count FROM provider_configurations;
