-- Migration: Admin Spin Prize Claims Management
-- Date: 2026-02-18
-- Description: Add columns and constraints for admin management of spin prize claims

-- ============================================================================
-- 1. Add new columns to spin_results table for admin review
-- ============================================================================

ALTER TABLE spin_results
ADD COLUMN IF NOT EXISTS reviewed_by UUID REFERENCES admin_users(id),
ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS rejection_reason TEXT,
ADD COLUMN IF NOT EXISTS admin_notes TEXT,
ADD COLUMN IF NOT EXISTS payment_reference VARCHAR(100);

-- ============================================================================
-- 2. Update claim_status constraint to include APPROVED and REJECTED
-- ============================================================================

ALTER TABLE spin_results
DROP CONSTRAINT IF EXISTS chk_spin_results_claim_status;

ALTER TABLE spin_results
ADD CONSTRAINT chk_spin_results_claim_status 
CHECK (claim_status IN ('PENDING', 'CLAIMED', 'EXPIRED', 'PENDING_ADMIN_REVIEW', 'APPROVED', 'REJECTED'));

-- ============================================================================
-- 3. Create indexes for admin query performance
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_spin_results_claim_status ON spin_results(claim_status);
CREATE INDEX IF NOT EXISTS idx_spin_results_created_at2 ON spin_results(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_spin_results_prize_type ON spin_results(prize_type);
CREATE INDEX IF NOT EXISTS idx_spin_results_reviewed_at ON spin_results(reviewed_at DESC);
CREATE INDEX IF NOT EXISTS idx_spin_results_msisdn_claim_status ON spin_results(msisdn, claim_status);
CREATE INDEX IF NOT EXISTS idx_spin_results_reviewed_by ON spin_results(reviewed_by);

-- ============================================================================
-- 4. Create admin_activity_logs table if it doesn't exist
-- ============================================================================

-- admin_activity_logs already created in 009; add missing columns if needed
ALTER TABLE admin_activity_logs
    ADD COLUMN IF NOT EXISTS admin_email VARCHAR(255),
    ADD COLUMN IF NOT EXISTS action_type VARCHAR(50),
    ADD COLUMN IF NOT EXISTS resource_type VARCHAR(50),
    ADD COLUMN IF NOT EXISTS details JSONB,
    ADD COLUMN IF NOT EXISTS ip_address INET,
    ADD COLUMN IF NOT EXISTS user_agent TEXT;

-- ============================================================================
-- 5. Create indexes for admin_activity_logs
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_admin_activity_logs_admin_user_id ON admin_activity_logs(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_activity_logs_resource ON admin_activity_logs(resource, resource_id);
CREATE INDEX IF NOT EXISTS idx_admin_activity_logs_created_at ON admin_activity_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_activity_logs_action_type ON admin_activity_logs(action);

-- ============================================================================
-- 6. Add comments for documentation
-- ============================================================================

COMMENT ON COLUMN spin_results.reviewed_by IS 'Admin user ID who reviewed the claim';
COMMENT ON COLUMN spin_results.reviewed_at IS 'Timestamp when claim was reviewed by admin';
COMMENT ON COLUMN spin_results.rejection_reason IS 'Reason provided by admin for rejecting the claim';
COMMENT ON COLUMN spin_results.admin_notes IS 'Internal notes added by admin during review';
COMMENT ON COLUMN spin_results.payment_reference IS 'Payment reference number for approved cash prizes';

COMMENT ON TABLE admin_activity_logs IS 'Audit trail for all admin actions on spin prize claims';
COMMENT ON COLUMN admin_activity_logs.action_type IS 'Type of admin action performed';
COMMENT ON COLUMN admin_activity_logs.resource_type IS 'Type of resource being acted upon';
COMMENT ON COLUMN admin_activity_logs.resource_id IS 'ID of the resource being acted upon';
COMMENT ON COLUMN admin_activity_logs.details IS 'JSON object containing additional action details';

-- ============================================================================
-- Migration complete
-- ============================================================================
