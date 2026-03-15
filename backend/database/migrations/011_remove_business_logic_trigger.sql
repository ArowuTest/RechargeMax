-- Migration: Remove business logic trigger from transactions table
-- Date: (see git history)
-- Description: Removes the process_transaction_trigger and trigger_process_transaction function
--              as part of the strategic move to application-layer business logic

-- This trigger was calling a non-existent function process_successful_transaction()
-- which caused webhook processing failures after payment success.

-- Drop the trigger from transactions table
DROP TRIGGER IF EXISTS process_transaction_trigger ON transactions;

-- Drop the trigger function
DROP FUNCTION IF EXISTS trigger_process_transaction() CASCADE;

-- Verify only timestamp triggers remain
COMMENT ON TABLE transactions IS 'Recharge transactions table - business logic handled in application layer (RechargeService.ProcessSuccessfulPayment)';

-- Add comment to clarify trigger removal
SELECT 'Business logic trigger removed - all processing now in application layer' AS migration_status;
