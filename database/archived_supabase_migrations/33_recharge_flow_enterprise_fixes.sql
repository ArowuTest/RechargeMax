-- ============================================================================
-- RECHARGE FLOW ENTERPRISE FIXES - Migration 031
-- ============================================================================
-- This migration implements all database-level fixes for the 18 critical issues
-- identified in the recharge flow analysis
-- ============================================================================

-- ISSUE #3 FIX: Add idempotency_key to transactions
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS idempotency_key TEXT;
CREATE INDEX IF NOT EXISTS idx_transactions_idempotency ON transactions(idempotency_key) WHERE idempotency_key IS NOT NULL;

-- ISSUE #4 FIX: Add processed_at timestamp for idempotency
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_transactions_processed ON transactions(processed_at) WHERE processed_at IS NOT NULL;

-- ISSUE #11 FIX: Add unique constraint to prevent concurrent recharges
-- (Partial - enforced in application logic with rate limiting)

-- ISSUE #13 FIX: Add daily and monthly recharge tracking
CREATE TABLE IF NOT EXISTS recharge_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    period_type TEXT NOT NULL CHECK (period_type IN ('DAILY', 'MONTHLY')),
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    total_amount BIGINT NOT NULL DEFAULT 0,
    transaction_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, period_type, period_start)
);

CREATE INDEX IF NOT EXISTS idx_recharge_limits_user ON recharge_limits(user_id);
CREATE INDEX IF NOT EXISTS idx_recharge_limits_period ON recharge_limits(period_start, period_end);

-- ISSUE #17 FIX: Add reconciliation tracking
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS reconciled_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS reconciliation_attempts INTEGER DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_transactions_reconciliation ON transactions(status, created_at) WHERE status = 'PENDING' AND processed_at IS NULL;

-- ISSUE #8 FIX: Ensure retry fields exist in vtu_transactions
-- (Already exist in schema, just verify)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vtu_transactions' AND column_name = 'retry_count') THEN
        ALTER TABLE vtu_transactions ADD COLUMN retry_count INTEGER DEFAULT 0;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vtu_transactions' AND column_name = 'max_retries') THEN
        ALTER TABLE vtu_transactions ADD COLUMN max_retries INTEGER DEFAULT 3;
    END IF;
END $$;

-- Add constraint to ensure retry_count doesn't exceed max_retries
ALTER TABLE vtu_transactions DROP CONSTRAINT IF EXISTS valid_retry_count;
ALTER TABLE vtu_transactions ADD CONSTRAINT valid_retry_count CHECK (retry_count <= max_retries);

-- ISSUE #5 FIX: Add provider response tracking for reconciliation
ALTER TABLE vtu_transactions ADD COLUMN IF NOT EXISTS provider_request JSONB;
ALTER TABLE vtu_transactions ADD COLUMN IF NOT EXISTS provider_response_raw JSONB;

-- ISSUE #16 FIX: Add payment verification tracking
CREATE TABLE IF NOT EXISTS payment_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    payment_reference TEXT NOT NULL,
    verification_source TEXT NOT NULL CHECK (verification_source IN ('WEBHOOK', 'RECONCILIATION', 'MANUAL')),
    verified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    paystack_response JSONB,
    amount_verified BIGINT,
    status_verified TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_verifications_transaction ON payment_verifications(transaction_id);
CREATE INDEX IF NOT EXISTS idx_payment_verifications_reference ON payment_verifications(payment_reference);

-- ISSUE #18 FIX: Add notification tracking
CREATE TABLE IF NOT EXISTS recharge_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    notification_type TEXT NOT NULL CHECK (notification_type IN ('SUCCESS', 'FAILURE', 'PROCESSING', 'REFUND')),
    channel TEXT NOT NULL CHECK (channel IN ('SMS', 'EMAIL', 'PUSH', 'IN_APP')),
    recipient TEXT NOT NULL,
    message TEXT NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    delivery_status TEXT DEFAULT 'PENDING',
    delivery_response JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recharge_notifications_transaction ON recharge_notifications(transaction_id);
CREATE INDEX IF NOT EXISTS idx_recharge_notifications_status ON recharge_notifications(delivery_status) WHERE delivery_status = 'PENDING';

-- Add function to update recharge limits
CREATE OR REPLACE FUNCTION update_recharge_limits()
RETURNS TRIGGER AS $$
BEGIN
    -- Update daily limit
    INSERT INTO recharge_limits (user_id, period_type, period_start, period_end, total_amount, transaction_count)
    VALUES (
        NEW.user_id,
        'DAILY',
        DATE_TRUNC('day', NEW.created_at),
        DATE_TRUNC('day', NEW.created_at) + INTERVAL '1 day',
        NEW.amount,
        1
    )
    ON CONFLICT (user_id, period_type, period_start)
    DO UPDATE SET
        total_amount = recharge_limits.total_amount + NEW.amount,
        transaction_count = recharge_limits.transaction_count + 1,
        updated_at = NOW();

    -- Update monthly limit
    INSERT INTO recharge_limits (user_id, period_type, period_start, period_end, total_amount, transaction_count)
    VALUES (
        NEW.user_id,
        'MONTHLY',
        DATE_TRUNC('month', NEW.created_at),
        DATE_TRUNC('month', NEW.created_at) + INTERVAL '1 month',
        NEW.amount,
        1
    )
    ON CONFLICT (user_id, period_type, period_start)
    DO UPDATE SET
        total_amount = recharge_limits.total_amount + NEW.amount,
        transaction_count = recharge_limits.transaction_count + 1,
        updated_at = NOW();

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to update recharge limits
DROP TRIGGER IF EXISTS trigger_update_recharge_limits ON transactions;
CREATE TRIGGER trigger_update_recharge_limits
    AFTER INSERT ON transactions
    FOR EACH ROW
    WHEN (NEW.type = 'RECHARGE' AND NEW.status IN ('SUCCESS', 'COMPLETED'))
    EXECUTE FUNCTION update_recharge_limits();

-- Add function to track payment verifications
CREATE OR REPLACE FUNCTION log_payment_verification()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.processed_at IS NOT NULL AND OLD.processed_at IS NULL THEN
        INSERT INTO payment_verifications (
            transaction_id,
            payment_reference,
            verification_source,
            amount_verified,
            status_verified
        ) VALUES (
            NEW.id,
            NEW.payment_reference,
            'WEBHOOK',
            NEW.amount,
            NEW.status
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to log payment verifications
DROP TRIGGER IF EXISTS trigger_log_payment_verification ON transactions;
CREATE TRIGGER trigger_log_payment_verification
    AFTER UPDATE ON transactions
    FOR EACH ROW
    WHEN (NEW.type = 'RECHARGE')
    EXECUTE FUNCTION log_payment_verification();

-- Add system configuration for recharge limits
INSERT INTO system_config (config_key, config_value, description, category, created_at, updated_at)
VALUES
    ('recharge_min_amount', '5000', 'Minimum recharge amount in kobo (₦50)', 'recharge', NOW(), NOW()),
    ('recharge_max_amount', '5000000', 'Maximum recharge amount per transaction in kobo (₦50,000)', 'recharge', NOW(), NOW()),
    ('recharge_daily_limit', '20000000', 'Maximum daily recharge amount per user in kobo (₦200,000)', 'recharge', NOW(), NOW()),
    ('recharge_monthly_limit', '100000000', 'Maximum monthly recharge amount per user in kobo (₦1,000,000)', 'recharge', NOW(), NOW()),
    ('recharge_concurrent_limit', '1', 'Maximum concurrent pending recharges per user', 'recharge', NOW(), NOW()),
    ('recharge_idempotency_window', '300', 'Idempotency window in seconds (5 minutes)', 'recharge', NOW(), NOW()),
    ('recharge_vtu_timeout', '30', 'VTU provider timeout in seconds', 'recharge', NOW(), NOW()),
    ('recharge_max_retries', '3', 'Maximum retry attempts for failed VTU recharges', 'recharge', NOW(), NOW()),
    ('recharge_reconciliation_interval', '3600', 'Reconciliation job interval in seconds (1 hour)', 'recharge', NOW(), NOW()),
    ('recharge_mode', 'simulation', 'Recharge mode: simulation, vtu, or direct', 'recharge', NOW(), NOW()),
    ('recharge_vtu_provider', 'vtpass', 'VTU provider: vtpass or shago', 'recharge', NOW(), NOW())
ON CONFLICT (config_key) DO UPDATE SET
    config_value = EXCLUDED.config_value,
    description = EXCLUDED.description,
    updated_at = NOW();

-- Add audit log for recharge operations
CREATE TABLE IF NOT EXISTS recharge_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID REFERENCES transactions(id),
    event_type TEXT NOT NULL,
    event_data JSONB,
    user_id UUID REFERENCES users(id),
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recharge_audit_transaction ON recharge_audit_log(transaction_id);
CREATE INDEX IF NOT EXISTS idx_recharge_audit_user ON recharge_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_recharge_audit_event ON recharge_audit_log(event_type, created_at DESC);

-- Add function to log recharge events
CREATE OR REPLACE FUNCTION log_recharge_event()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO recharge_audit_log (transaction_id, event_type, event_data, user_id)
        VALUES (NEW.id, 'RECHARGE_CREATED', row_to_json(NEW), NEW.user_id);
    ELSIF TG_OP = 'UPDATE' AND NEW.status != OLD.status THEN
        INSERT INTO recharge_audit_log (transaction_id, event_type, event_data, user_id)
        VALUES (NEW.id, 'STATUS_CHANGED', jsonb_build_object(
            'old_status', OLD.status,
            'new_status', NEW.status,
            'transaction', row_to_json(NEW)
        ), NEW.user_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to log recharge events
DROP TRIGGER IF EXISTS trigger_log_recharge_event ON transactions;
CREATE TRIGGER trigger_log_recharge_event
    AFTER INSERT OR UPDATE ON transactions
    FOR EACH ROW
    WHEN (NEW.type = 'RECHARGE')
    EXECUTE FUNCTION log_recharge_event();

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_transactions_user_type_status ON transactions(user_id, type, status);
CREATE INDEX IF NOT EXISTS idx_transactions_created_status ON transactions(created_at DESC, status) WHERE type = 'RECHARGE';
CREATE INDEX IF NOT EXISTS idx_vtu_transactions_status_retry ON vtu_transactions(status, retry_count) WHERE status IN ('PENDING', 'PROCESSING');

-- Add comments for documentation
COMMENT ON TABLE recharge_limits IS 'Tracks daily and monthly recharge limits per user';
COMMENT ON TABLE payment_verifications IS 'Logs all payment verification attempts for audit trail';
COMMENT ON TABLE recharge_notifications IS 'Tracks all notifications sent for recharge transactions';
COMMENT ON TABLE recharge_audit_log IS 'Complete audit trail of all recharge operations';

COMMENT ON COLUMN transactions.idempotency_key IS 'Client-provided key to prevent duplicate recharges';
COMMENT ON COLUMN transactions.processed_at IS 'Timestamp when payment was processed (for idempotency)';
COMMENT ON COLUMN transactions.reconciled_at IS 'Timestamp when transaction was reconciled';
COMMENT ON COLUMN transactions.reconciliation_attempts IS 'Number of reconciliation attempts';

-- Grant permissions (adjust as needed for your setup)
-- GRANT SELECT, INSERT, UPDATE ON recharge_limits TO rechargemax_app;
-- GRANT SELECT, INSERT ON payment_verifications TO rechargemax_app;
-- GRANT SELECT, INSERT ON recharge_notifications TO rechargemax_app;
-- GRANT SELECT, INSERT ON recharge_audit_log TO rechargemax_app;

-- Migration complete
SELECT 'Recharge flow enterprise fixes migration completed successfully' AS status;
