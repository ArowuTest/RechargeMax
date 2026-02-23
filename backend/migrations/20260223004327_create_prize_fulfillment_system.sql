-- ============================================================================
-- MIGRATION: Prize Fulfillment System (Enterprise-Grade)
-- Description: Implements flexible auto/manual fulfillment with retry logic,
--              fallback mechanisms, and comprehensive audit trails
-- Author: RechargeMax Team
-- Date: 2026-02-23
-- Version: 1.0.0
-- ============================================================================

-- TABLE 1: Prize Fulfillment Configuration
CREATE TABLE IF NOT EXISTS prize_fulfillment_config (
    id SERIAL PRIMARY KEY,
    prize_type VARCHAR(20) NOT NULL,
    fulfillment_mode VARCHAR(20) NOT NULL DEFAULT 'AUTO',
    auto_retry_enabled BOOLEAN DEFAULT TRUE,
    max_retry_attempts INTEGER DEFAULT 3,
    retry_delay_seconds INTEGER DEFAULT 300,
    fallback_to_manual BOOLEAN DEFAULT TRUE,
    fallback_notification_enabled BOOLEAN DEFAULT TRUE,
    provision_timeout_seconds INTEGER DEFAULT 60,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    CONSTRAINT unique_prize_type UNIQUE(prize_type),
    CONSTRAINT check_fulfillment_mode CHECK (fulfillment_mode IN ('AUTO', 'MANUAL')),
    CONSTRAINT check_prize_type CHECK (prize_type IN ('AIRTIME', 'DATA', 'CASH', 'POINTS', 'PHYSICAL')),
    CONSTRAINT check_max_retry_attempts CHECK (max_retry_attempts >= 0 AND max_retry_attempts <= 10),
    CONSTRAINT check_retry_delay CHECK (retry_delay_seconds >= 0 AND retry_delay_seconds <= 3600),
    CONSTRAINT check_timeout CHECK (provision_timeout_seconds >= 10 AND provision_timeout_seconds <= 300)
);

CREATE INDEX idx_fulfillment_config_prize_type ON prize_fulfillment_config(prize_type);
CREATE INDEX idx_fulfillment_config_active ON prize_fulfillment_config(is_active);

-- TABLE 2: Prize Fulfillment Logs
CREATE TABLE IF NOT EXISTS prize_fulfillment_logs (
    id BIGSERIAL PRIMARY KEY,
    spin_result_id UUID NOT NULL,
    attempt_number INTEGER NOT NULL,
    fulfillment_mode VARCHAR(20) NOT NULL,
    provider_name VARCHAR(50),
    provider_reference VARCHAR(100),
    provider_transaction_id BIGINT,
    request_payload JSONB,
    response_payload JSONB,
    status VARCHAR(20) NOT NULL,
    error_code VARCHAR(50),
    error_message TEXT,
    response_time_ms INTEGER,
    detected_network VARCHAR(20),
    msisdn VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT check_status CHECK (status IN ('SUCCESS', 'FAILED', 'PENDING', 'TIMEOUT', 'CANCELLED')),
    CONSTRAINT check_attempt_number CHECK (attempt_number > 0)
);

CREATE INDEX idx_fulfillment_logs_spin_result ON prize_fulfillment_logs(spin_result_id);
CREATE INDEX idx_fulfillment_logs_status ON prize_fulfillment_logs(status);
CREATE INDEX idx_fulfillment_logs_created_at ON prize_fulfillment_logs(created_at DESC);
CREATE INDEX idx_fulfillment_logs_provider_ref ON prize_fulfillment_logs(provider_reference);
CREATE INDEX idx_fulfillment_logs_msisdn ON prize_fulfillment_logs(msisdn);

-- TABLE 3: Update spin_results
ALTER TABLE spin_results
ADD COLUMN IF NOT EXISTS fulfillment_mode VARCHAR(20) DEFAULT 'AUTO',
ADD COLUMN IF NOT EXISTS fulfillment_attempts INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS last_fulfillment_attempt TIMESTAMP,
ADD COLUMN IF NOT EXISTS fulfillment_error TEXT,
ADD COLUMN IF NOT EXISTS can_retry BOOLEAN DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS provision_started_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS provision_completed_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS claim_reference VARCHAR(100);

ALTER TABLE spin_results
ADD CONSTRAINT IF NOT EXISTS check_fulfillment_mode 
CHECK (fulfillment_mode IN ('AUTO', 'MANUAL'));

CREATE INDEX IF NOT EXISTS idx_spin_results_fulfillment_mode ON spin_results(fulfillment_mode);
CREATE INDEX IF NOT EXISTS idx_spin_results_can_retry ON spin_results(can_retry) WHERE can_retry = TRUE;

-- SEED DATA
INSERT INTO prize_fulfillment_config (prize_type, fulfillment_mode, auto_retry_enabled, max_retry_attempts, retry_delay_seconds, fallback_to_manual, fallback_notification_enabled, provision_timeout_seconds, is_active, created_by) VALUES
    ('AIRTIME', 'AUTO', TRUE, 3, 300, TRUE, TRUE, 60, TRUE, 'SYSTEM'),
    ('DATA', 'AUTO', TRUE, 3, 300, TRUE, TRUE, 60, TRUE, 'SYSTEM'),
    ('CASH', 'MANUAL', FALSE, 0, 0, FALSE, FALSE, 60, TRUE, 'SYSTEM'),
    ('POINTS', 'AUTO', FALSE, 0, 0, FALSE, FALSE, 10, TRUE, 'SYSTEM'),
    ('PHYSICAL', 'MANUAL', FALSE, 0, 0, FALSE, FALSE, 60, TRUE, 'SYSTEM')
ON CONFLICT (prize_type) DO NOTHING;
