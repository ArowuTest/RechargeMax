-- Migration: Transaction Limits Configuration System
-- Description: Implements configurable transaction limits for fraud prevention and risk management
-- Author: Manus AI
-- Date: (see git history)

-- Create transaction_limits table
CREATE TABLE IF NOT EXISTS transaction_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    limit_type VARCHAR(50) NOT NULL, -- 'AIRTIME', 'DATA', 'SUBSCRIPTION', 'WITHDRAWAL'
    limit_scope VARCHAR(50) NOT NULL, -- 'GLOBAL', 'PER_USER', 'PER_TRANSACTION'
    min_amount BIGINT NOT NULL DEFAULT 10000, -- Minimum amount in kobo (₦100)
    max_amount BIGINT NOT NULL DEFAULT 10000000, -- Maximum amount in kobo (₦100,000)
    daily_limit BIGINT, -- Daily cumulative limit in kobo (optional)
    monthly_limit BIGINT, -- Monthly cumulative limit in kobo (optional)
    is_active BOOLEAN NOT NULL DEFAULT true,
    applies_to_user_tier VARCHAR(50), -- 'bronze', 'silver', 'gold', 'platinum', NULL for all
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID, -- Admin user who created this limit
    updated_by UUID, -- Admin user who last updated this limit
    
    CONSTRAINT unique_limit_config UNIQUE (limit_type, limit_scope, applies_to_user_tier),
    CONSTRAINT valid_amount_range CHECK (min_amount <= max_amount),
    CONSTRAINT positive_amounts CHECK (min_amount > 0 AND max_amount > 0)
);

-- Create index for faster lookups
CREATE INDEX idx_transaction_limits_active ON transaction_limits(is_active) WHERE is_active = true;
CREATE INDEX idx_transaction_limits_type_scope ON transaction_limits(limit_type, limit_scope);
CREATE INDEX idx_transaction_limits_tier ON transaction_limits(applies_to_user_tier);

-- Create audit log for limit changes
CREATE TABLE IF NOT EXISTS transaction_limits_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    limit_id UUID NOT NULL REFERENCES transaction_limits(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL, -- 'CREATE', 'UPDATE', 'DELETE', 'ACTIVATE', 'DEACTIVATE'
    old_values JSONB,
    new_values JSONB,
    changed_by UUID, -- Admin user who made the change
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45),
    user_agent TEXT,
    reason TEXT -- Optional reason for the change
);

-- Create index for audit trail queries
CREATE INDEX idx_limits_audit_limit_id ON transaction_limits_audit(limit_id);
CREATE INDEX idx_limits_audit_changed_at ON transaction_limits_audit(changed_at DESC);
CREATE INDEX idx_limits_audit_changed_by ON transaction_limits_audit(changed_by);

-- Create trigger function for automatic updated_at timestamp
CREATE OR REPLACE FUNCTION update_transaction_limits_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
CREATE TRIGGER transaction_limits_updated_at
    BEFORE UPDATE ON transaction_limits
    FOR EACH ROW
    EXECUTE FUNCTION update_transaction_limits_timestamp();

-- Insert default transaction limits (production-ready values)
INSERT INTO transaction_limits (limit_type, limit_scope, min_amount, max_amount, daily_limit, monthly_limit, applies_to_user_tier, description) VALUES
-- Global limits for all users
('AIRTIME', 'PER_TRANSACTION', 10000, 10000000, NULL, NULL, NULL, 'Global per-transaction limit for airtime recharge (₦100 - ₦100,000)'),
('DATA', 'PER_TRANSACTION', 10000, 10000000, NULL, NULL, NULL, 'Global per-transaction limit for data recharge (₦100 - ₦100,000)'),
('SUBSCRIPTION', 'PER_TRANSACTION', 2000, 2000, NULL, NULL, NULL, 'Daily subscription limit (₦20)'),

-- Daily cumulative limits by user tier
('AIRTIME', 'DAILY_CUMULATIVE', 10000, 50000000, 50000000, NULL, 'bronze', 'Bronze tier daily airtime limit (₦500,000)'),
('AIRTIME', 'DAILY_CUMULATIVE', 10000, 100000000, 100000000, NULL, 'silver', 'Silver tier daily airtime limit (₦1,000,000)'),
('AIRTIME', 'DAILY_CUMULATIVE', 10000, 200000000, 200000000, NULL, 'gold', 'Gold tier daily airtime limit (₦2,000,000)'),
('AIRTIME', 'DAILY_CUMULATIVE', 10000, 500000000, 500000000, NULL, 'platinum', 'Platinum tier daily airtime limit (₦5,000,000)'),

('DATA', 'DAILY_CUMULATIVE', 10000, 50000000, 50000000, NULL, 'bronze', 'Bronze tier daily data limit (₦500,000)'),
('DATA', 'DAILY_CUMULATIVE', 10000, 100000000, 100000000, NULL, 'silver', 'Silver tier daily data limit (₦1,000,000)'),
('DATA', 'DAILY_CUMULATIVE', 10000, 200000000, 200000000, NULL, 'gold', 'Gold tier daily data limit (₦2,000,000)'),
('DATA', 'DAILY_CUMULATIVE', 10000, 500000000, 500000000, NULL, 'platinum', 'Platinum tier daily data limit (₦5,000,000)'),

-- Monthly cumulative limits
('AIRTIME', 'MONTHLY_CUMULATIVE', 10000, 1000000000, NULL, 1000000000, NULL, 'Global monthly airtime limit (₦10,000,000)'),
('DATA', 'MONTHLY_CUMULATIVE', 10000, 1000000000, NULL, 1000000000, NULL, 'Global monthly data limit (₦10,000,000)')
ON CONFLICT DO NOTHING;

-- Create helper function to get active limit for a specific context
CREATE OR REPLACE FUNCTION get_transaction_limit(
    p_limit_type VARCHAR(50),
    p_limit_scope VARCHAR(50),
    p_user_tier VARCHAR(50) DEFAULT NULL
)
RETURNS TABLE (
    min_amount BIGINT,
    max_amount BIGINT,
    daily_limit BIGINT,
    monthly_limit BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        tl.min_amount,
        tl.max_amount,
        tl.daily_limit,
        tl.monthly_limit
    FROM transaction_limits tl
    WHERE tl.limit_type = p_limit_type
      AND tl.limit_scope = p_limit_scope
      AND tl.is_active = true
      AND (tl.applies_to_user_tier = p_user_tier OR tl.applies_to_user_tier IS NULL)
    ORDER BY 
        CASE WHEN tl.applies_to_user_tier IS NOT NULL THEN 1 ELSE 2 END, -- Prioritize tier-specific limits
        tl.created_at DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON TABLE transaction_limits IS 'Configurable transaction limits for fraud prevention and risk management';
COMMENT ON COLUMN transaction_limits.limit_type IS 'Type of transaction: AIRTIME, DATA, SUBSCRIPTION, WITHDRAWAL';
COMMENT ON COLUMN transaction_limits.limit_scope IS 'Scope of limit: GLOBAL, PER_USER, PER_TRANSACTION, DAILY_CUMULATIVE, MONTHLY_CUMULATIVE';
COMMENT ON COLUMN transaction_limits.min_amount IS 'Minimum transaction amount in kobo';
COMMENT ON COLUMN transaction_limits.max_amount IS 'Maximum transaction amount in kobo';
COMMENT ON COLUMN transaction_limits.daily_limit IS 'Daily cumulative limit in kobo (optional)';
COMMENT ON COLUMN transaction_limits.monthly_limit IS 'Monthly cumulative limit in kobo (optional)';
COMMENT ON COLUMN transaction_limits.applies_to_user_tier IS 'User loyalty tier this limit applies to (NULL for all tiers)';

COMMENT ON TABLE transaction_limits_audit IS 'Audit trail for all changes to transaction limits';
COMMENT ON FUNCTION get_transaction_limit IS 'Helper function to retrieve active transaction limit for a specific context';
