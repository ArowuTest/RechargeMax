-- Migration: Create provider_configs table and related functions
-- Purpose: Enable dynamic provider configuration for VTU services
-- Date: 2026-02-20

-- Create provider_configs table
CREATE TABLE IF NOT EXISTS provider_configs (
    id BIGSERIAL PRIMARY KEY,
    network VARCHAR(50) NOT NULL,
    service_type VARCHAR(50) NOT NULL,
    provider_mode VARCHAR(50) NOT NULL,  -- VTU, DIRECT, SIMULATION
    provider_name VARCHAR(100) NOT NULL,
    priority INTEGER DEFAULT 1,
    config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Ensure unique active provider per network/service combination at each priority level
    CONSTRAINT unique_active_provider UNIQUE (network, service_type, priority, is_active)
);

-- Create index for fast lookups
CREATE INDEX IF NOT EXISTS idx_provider_configs_lookup 
ON provider_configs(network, service_type, is_active, priority);

-- Create updated_at trigger
CREATE TRIGGER update_provider_configs_updated_at
    BEFORE UPDATE ON provider_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create get_active_provider function
CREATE OR REPLACE FUNCTION get_active_provider(p_network VARCHAR, p_service_type VARCHAR)
RETURNS TABLE (
    id BIGINT,
    network VARCHAR,
    service_type VARCHAR,
    provider_mode VARCHAR,
    provider_name VARCHAR,
    priority INTEGER,
    config JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pc.id,
        pc.network,
        pc.service_type,
        pc.provider_mode,
        pc.provider_name,
        pc.priority,
        pc.config
    FROM provider_configs pc
    WHERE pc.network = p_network 
      AND pc.service_type = p_service_type
      AND pc.is_active = true
    ORDER BY pc.priority ASC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Seed initial provider configurations for all networks
-- MTN configurations
INSERT INTO provider_configs (network, service_type, provider_mode, provider_name, priority, config, is_active) VALUES
('MTN', 'AIRTIME', 'VTU', 'VTPass', 1, '{"mode": "sandbox", "api_key_env": "VTPASS_API_KEY", "public_key_env": "VTPASS_PUBLIC_KEY", "secret_key_env": "VTPASS_SECRET_KEY"}', true),
('MTN', 'DATA', 'VTU', 'VTPass', 1, '{"mode": "sandbox", "api_key_env": "VTPASS_API_KEY", "public_key_env": "VTPASS_PUBLIC_KEY", "secret_key_env": "VTPASS_SECRET_KEY"}', true);

-- GLO configurations
INSERT INTO provider_configs (network, service_type, provider_mode, provider_name, priority, config, is_active) VALUES
('GLO', 'AIRTIME', 'VTU', 'VTPass', 1, '{"mode": "sandbox", "api_key_env": "VTPASS_API_KEY", "public_key_env": "VTPASS_PUBLIC_KEY", "secret_key_env": "VTPASS_SECRET_KEY"}', true),
('GLO', 'DATA', 'VTU', 'VTPass', 1, '{"mode": "sandbox", "api_key_env": "VTPASS_API_KEY", "public_key_env": "VTPASS_PUBLIC_KEY", "secret_key_env": "VTPASS_SECRET_KEY"}', true);

-- AIRTEL configurations
INSERT INTO provider_configs (network, service_type, provider_mode, provider_name, priority, config, is_active) VALUES
('AIRTEL', 'AIRTIME', 'VTU', 'VTPass', 1, '{"mode": "sandbox", "api_key_env": "VTPASS_API_KEY", "public_key_env": "VTPASS_PUBLIC_KEY", "secret_key_env": "VTPASS_SECRET_KEY"}', true),
('AIRTEL', 'DATA', 'VTU', 'VTPass', 1, '{"mode": "sandbox", "api_key_env": "VTPASS_API_KEY", "public_key_env": "VTPASS_PUBLIC_KEY", "secret_key_env": "VTPASS_SECRET_KEY"}', true);

-- 9MOBILE configurations
INSERT INTO provider_configs (network, service_type, provider_mode, provider_name, priority, config, is_active) VALUES
('9MOBILE', 'AIRTIME', 'VTU', 'VTPass', 1, '{"mode": "sandbox", "api_key_env": "VTPASS_API_KEY", "public_key_env": "VTPASS_PUBLIC_KEY", "secret_key_env": "VTPASS_SECRET_KEY"}', true),
('9MOBILE', 'DATA', 'VTU', 'VTPass', 1, '{"mode": "sandbox", "api_key_env": "VTPASS_API_KEY", "public_key_env": "VTPASS_PUBLIC_KEY", "secret_key_env": "VTPASS_SECRET_KEY"}', true);

-- Add simulation providers as fallback (priority 2)
INSERT INTO provider_configs (network, service_type, provider_mode, provider_name, priority, config, is_active) VALUES
('MTN', 'AIRTIME', 'SIMULATION', 'SimulationProvider', 2, '{"success_rate": 0.9}', false),
('MTN', 'DATA', 'SIMULATION', 'SimulationProvider', 2, '{"success_rate": 0.9}', false),
('GLO', 'AIRTIME', 'SIMULATION', 'SimulationProvider', 2, '{"success_rate": 0.9}', false),
('GLO', 'DATA', 'SIMULATION', 'SimulationProvider', 2, '{"success_rate": 0.9}', false),
('AIRTEL', 'AIRTIME', 'SIMULATION', 'SimulationProvider', 2, '{"success_rate": 0.9}', false),
('AIRTEL', 'DATA', 'SIMULATION', 'SimulationProvider', 2, '{"success_rate": 0.9}', false),
('9MOBILE', 'AIRTIME', 'SIMULATION', 'SimulationProvider', 2, '{"success_rate": 0.9}', false),
('9MOBILE', 'DATA', 'SIMULATION', 'SimulationProvider', 2, '{"success_rate": 0.9}', false);

-- Create function to log provider transactions
CREATE OR REPLACE FUNCTION log_provider_transaction(
    p_transaction_id BIGINT,
    p_provider_config_id BIGINT,
    p_provider_mode VARCHAR,
    p_provider_name VARCHAR,
    p_request_payload JSONB,
    p_response_payload JSONB,
    p_status VARCHAR,
    p_message TEXT,
    p_response_time_ms BIGINT
) RETURNS BIGINT AS $$
DECLARE
    v_log_id BIGINT;
BEGIN
    -- Note: This function is a placeholder for future provider_transaction_logs table
    -- For now, return a dummy ID
    -- TODO: Create provider_transaction_logs table in future migration
    RETURN 1;
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON TABLE provider_configs IS 'Stores configuration for VTU service providers (VTPass, etc.)';
COMMENT ON COLUMN provider_configs.provider_mode IS 'VTU = VTU aggregator, DIRECT = Direct network API, SIMULATION = Test mode';
COMMENT ON COLUMN provider_configs.priority IS 'Lower number = higher priority. Used for failover.';
COMMENT ON COLUMN provider_configs.config IS 'JSON configuration specific to provider (API keys, endpoints, etc.)';
COMMENT ON FUNCTION get_active_provider IS 'Returns the highest priority active provider for a network/service combination';
