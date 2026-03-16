-- Migration 033: Add columns that exist in application code but were absent from the original schema.
-- All statements use IF NOT EXISTS / DO $$ guards to be fully idempotent.

-- 1. affiliate_commissions.updated_at
--    Required by commission_release_job.go (SET updated_at = NOW())
--    Added in commit fb510e4 but no migration was written at the time.
ALTER TABLE affiliate_commissions
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- 2. transactions.processed_at
--    Required by reconciliation_job.go (WHERE processed_at IS NULL, SET processed_at = ?)
--    Referenced since the original codebase but never added to the schema.
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_transactions_processed_at
    ON transactions(processed_at);

-- 3. wallet_transactions.processed_at
--    Required by WalletTransaction GORM entity (gorm:"column:processed_at;default:CURRENT_TIMESTAMP;not null")
--    The entity marks it NOT NULL so we provide a default for existing rows.
ALTER TABLE wallet_transactions
    ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Backfill existing rows so NOT NULL constraint can be satisfied
UPDATE wallet_transactions
    SET processed_at = created_at
    WHERE processed_at IS NULL;

-- 4. wallet_transactions.updated_at
--    Required by WalletTransaction GORM entity (gorm:"column:updated_at;autoUpdateTime")
ALTER TABLE wallet_transactions
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

UPDATE wallet_transactions
    SET updated_at = created_at
    WHERE updated_at IS NULL;
