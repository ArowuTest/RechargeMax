-- Migration: Add admin review columns to spin_results
-- Date: 2026-02-19
-- Purpose: Enable admin review and approval workflow for prize claims

-- Add admin review columns to spin_results table
ALTER TABLE spin_results 
ADD COLUMN IF NOT EXISTS reviewed_by TEXT,
ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS rejection_reason TEXT,
ADD COLUMN IF NOT EXISTS admin_notes TEXT,
ADD COLUMN IF NOT EXISTS payment_reference TEXT,
ADD COLUMN IF NOT EXISTS bank_account_number TEXT,
ADD COLUMN IF NOT EXISTS bank_account_name TEXT,
ADD COLUMN IF NOT EXISTS bank_name TEXT;

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_spin_results_reviewed_by ON spin_results(reviewed_by);
CREATE INDEX IF NOT EXISTS idx_spin_results_reviewed_at ON spin_results(reviewed_at);

-- Add comments
COMMENT ON COLUMN spin_results.reviewed_by IS 'Admin user ID who reviewed the claim';
COMMENT ON COLUMN spin_results.reviewed_at IS 'Timestamp when claim was reviewed';
COMMENT ON COLUMN spin_results.rejection_reason IS 'Reason for claim rejection';
COMMENT ON COLUMN spin_results.admin_notes IS 'Admin notes about the claim';
COMMENT ON COLUMN spin_results.payment_reference IS 'Payment reference for approved cash prizes';
COMMENT ON COLUMN spin_results.bank_account_number IS 'User bank account number for cash prizes';
COMMENT ON COLUMN spin_results.bank_account_name IS 'User bank account name for cash prizes';
COMMENT ON COLUMN spin_results.bank_name IS 'User bank name for cash prizes';
