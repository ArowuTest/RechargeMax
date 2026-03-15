-- Migration: Add transaction_code column to transactions table
-- Date: (see git history)
-- Description: Add unique transaction_code column for better transaction tracking

-- Add transaction_code column
ALTER TABLE transactions 
ADD COLUMN IF NOT EXISTS transaction_code VARCHAR(30);

-- Create unique index on transaction_code
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_transaction_code 
ON transactions(transaction_code);

-- Update existing transactions with generated transaction codes
UPDATE transactions 
SET transaction_code = 'TXN_' || SUBSTRING(msisdn FROM 8) || '_' || EXTRACT(EPOCH FROM created_at)::BIGINT
WHERE transaction_code IS NULL OR transaction_code = '';

-- Make transaction_code NOT NULL after populating existing records
ALTER TABLE transactions 
ALTER COLUMN transaction_code SET NOT NULL;
