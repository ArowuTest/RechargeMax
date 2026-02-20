-- ============================================================================
-- FLEXIBLE PRIZE FULFILLMENT SYSTEM - DATABASE SCHEMA
-- ============================================================================
-- Supports both auto-provision and manual claim modes
-- Admin-configurable per prize type or globally
-- ============================================================================

-- 1. Prize Fulfillment Configuration Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS prize_fulfillment_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prize_type VARCHAR(50) NOT NULL, -- 'airtime', 'data', 'cash', 'goods', 'points'
    fulfillment_mode VARCHAR(20) NOT NULL DEFAULT 'manual_claim', -- 'auto_provision' or 'manual_claim'
    auto_provision_enabled BOOLEAN NOT NULL DEFAULT false,
    require_login_to_claim BOOLEAN NOT NULL DEFAULT true,
    claim_deadline_days INTEGER NOT NULL DEFAULT 30,
    allow_retry_on_failure BOOLEAN NOT NULL DEFAULT true,
    max_retry_attempts INTEGER NOT NULL DEFAULT 3,
    notification_template_id UUID,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID,
    
    UNIQUE(prize_type)
);

-- 2. Global Fulfillment Settings Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS global_fulfillment_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    setting_key VARCHAR(100) NOT NULL UNIQUE,
    setting_value TEXT NOT NULL,
    setting_type VARCHAR(20) NOT NULL, -- 'boolean', 'integer', 'string', 'json'
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 3. Update Winners Table to Support Both Modes
-- ============================================================================
ALTER TABLE winners ADD COLUMN IF NOT EXISTS fulfillment_mode VARCHAR(20) DEFAULT 'auto_provision';
ALTER TABLE winners ADD COLUMN IF NOT EXISTS claim_initiated_at TIMESTAMP;
ALTER TABLE winners ADD COLUMN IF NOT EXISTS claim_button_clicked BOOLEAN DEFAULT false;
ALTER TABLE winners ADD COLUMN IF NOT EXISTS provision_attempts INTEGER DEFAULT 0;
ALTER TABLE winners ADD COLUMN IF NOT EXISTS last_provision_attempt_at TIMESTAMP;
ALTER TABLE winners ADD COLUMN IF NOT EXISTS provider_transaction_id VARCHAR(255);
ALTER TABLE winners ADD COLUMN IF NOT EXISTS provider_response JSONB;

-- Add index for claim queries
CREATE INDEX IF NOT EXISTS idx_winners_claim_status ON winners(claim_status);
CREATE INDEX IF NOT EXISTS idx_winners_fulfillment_mode ON winners(fulfillment_mode);
CREATE INDEX IF NOT EXISTS idx_winners_provision_status ON winners(provision_status);

-- 4. Prize Claim Audit Log Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS prize_claim_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    winner_id UUID NOT NULL REFERENCES winners(id),
    action VARCHAR(50) NOT NULL, -- 'claim_initiated', 'provision_started', 'provision_success', 'provision_failed', 'retry_attempted'
    actor_type VARCHAR(20) NOT NULL, -- 'user', 'system', 'admin'
    actor_id UUID,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    FOREIGN KEY (winner_id) REFERENCES winners(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_claim_audit_winner_id ON prize_claim_audit_log(winner_id);
CREATE INDEX IF NOT EXISTS idx_claim_audit_created_at ON prize_claim_audit_log(created_at);

-- ============================================================================
-- SEED DATA: Default Prize Fulfillment Configurations
-- ============================================================================

INSERT INTO prize_fulfillment_config (prize_type, fulfillment_mode, auto_provision_enabled, require_login_to_claim, claim_deadline_days, allow_retry_on_failure, max_retry_attempts)
VALUES 
    ('airtime', 'manual_claim', true, true, 30, true, 3),
    ('data', 'manual_claim', true, true, 30, true, 3),
    ('cash', 'manual_claim', false, true, 30, false, 0),
    ('goods', 'manual_claim', false, true, 30, false, 0),
    ('points', 'auto_provision', true, false, 30, true, 3)
ON CONFLICT (prize_type) DO UPDATE SET
    updated_at = NOW();

-- ============================================================================
-- SEED DATA: Global Fulfillment Settings
-- ============================================================================

INSERT INTO global_fulfillment_settings (setting_key, setting_value, setting_type, description)
VALUES 
    ('default_fulfillment_mode', 'manual_claim', 'string', 'Default fulfillment mode for new prize types'),
    ('enable_claim_reminders', 'true', 'boolean', 'Send reminders to users with unclaimed prizes'),
    ('claim_reminder_days_before_deadline', '7,3,1', 'string', 'Days before deadline to send reminders (comma-separated)'),
    ('auto_expire_unclaimed_prizes', 'true', 'boolean', 'Automatically expire prizes after deadline'),
    ('allow_admin_override_mode', 'true', 'boolean', 'Allow admin to override fulfillment mode per winner'),
    ('vtpass_retry_delay_minutes', '5', 'integer', 'Delay between VTPass retry attempts'),
    ('max_concurrent_provisions', '10', 'integer', 'Maximum concurrent VTPass provision requests'),
    ('enable_provision_queue', 'true', 'boolean', 'Queue provisions instead of immediate processing')
ON CONFLICT (setting_key) DO UPDATE SET
    setting_value = EXCLUDED.setting_value,
    updated_at = NOW();

-- ============================================================================
-- FUNCTIONS: Helper Functions
-- ============================================================================

-- Function to get fulfillment config for a prize type
CREATE OR REPLACE FUNCTION get_prize_fulfillment_config(p_prize_type VARCHAR)
RETURNS TABLE (
    fulfillment_mode VARCHAR,
    auto_provision_enabled BOOLEAN,
    require_login_to_claim BOOLEAN,
    claim_deadline_days INTEGER,
    allow_retry_on_failure BOOLEAN,
    max_retry_attempts INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pfc.fulfillment_mode,
        pfc.auto_provision_enabled,
        pfc.require_login_to_claim,
        pfc.claim_deadline_days,
        pfc.allow_retry_on_failure,
        pfc.max_retry_attempts
    FROM prize_fulfillment_config pfc
    WHERE pfc.prize_type = p_prize_type
      AND pfc.is_active = true
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function to check if user can claim prize
CREATE OR REPLACE FUNCTION can_user_claim_prize(p_winner_id UUID)
RETURNS BOOLEAN AS $$
DECLARE
    v_claim_status VARCHAR;
    v_claim_deadline TIMESTAMP;
    v_fulfillment_mode VARCHAR;
BEGIN
    SELECT claim_status, claim_deadline, fulfillment_mode
    INTO v_claim_status, v_claim_deadline, v_fulfillment_mode
    FROM winners
    WHERE id = p_winner_id;
    
    -- Check if winner exists
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    
    -- Check if already claimed
    IF v_claim_status IN ('claimed', 'claim_submitted') THEN
        RETURN false;
    END IF;
    
    -- Check if deadline passed
    IF v_claim_deadline IS NOT NULL AND NOW() > v_claim_deadline THEN
        RETURN false;
    END IF;
    
    -- Check if manual claim mode
    IF v_fulfillment_mode != 'manual_claim' THEN
        RETURN false;
    END IF;
    
    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- VIEWS: Admin Dashboard Views
-- ============================================================================

-- View for unclaimed prizes requiring manual claim
CREATE OR REPLACE VIEW unclaimed_manual_prizes AS
SELECT 
    w.id,
    w.msisdn,
    w.prize_type,
    w.prize_description,
    w.prize_amount,
    w.airtime_amount,
    w.data_package,
    w.claim_status,
    w.claim_deadline,
    w.fulfillment_mode,
    w.provision_status,
    w.provision_attempts,
    w.created_at,
    EXTRACT(DAY FROM (w.claim_deadline - NOW())) AS days_until_deadline,
    d.name AS draw_name
FROM winners w
LEFT JOIN draws d ON w.draw_id = d.id
WHERE w.fulfillment_mode = 'manual_claim'
  AND w.claim_status IN ('pending', 'unclaimed')
  AND w.claim_deadline > NOW()
ORDER BY w.claim_deadline ASC;

-- View for failed provisions requiring retry
CREATE OR REPLACE VIEW failed_provisions AS
SELECT 
    w.id,
    w.msisdn,
    w.prize_type,
    w.prize_description,
    w.provision_status,
    w.provision_error,
    w.provision_attempts,
    w.last_provision_attempt_at,
    w.created_at,
    pfc.max_retry_attempts,
    pfc.allow_retry_on_failure
FROM winners w
JOIN prize_fulfillment_config pfc ON w.prize_type = pfc.prize_type
WHERE w.provision_status = 'failed'
  AND pfc.allow_retry_on_failure = true
  AND w.provision_attempts < pfc.max_retry_attempts
ORDER BY w.last_provision_attempt_at ASC;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE prize_fulfillment_config IS 'Configuration for prize fulfillment modes per prize type';
COMMENT ON TABLE global_fulfillment_settings IS 'Global settings for prize fulfillment system';
COMMENT ON TABLE prize_claim_audit_log IS 'Audit log for all prize claim and provision actions';
COMMENT ON COLUMN winners.fulfillment_mode IS 'Mode used for this prize: auto_provision or manual_claim';
COMMENT ON COLUMN winners.claim_button_clicked IS 'Whether user clicked claim button (for analytics)';
COMMENT ON COLUMN winners.provision_attempts IS 'Number of VTPass provision attempts made';
COMMENT ON FUNCTION get_prize_fulfillment_config IS 'Get fulfillment configuration for a specific prize type';
COMMENT ON FUNCTION can_user_claim_prize IS 'Check if user is allowed to claim a specific prize';
