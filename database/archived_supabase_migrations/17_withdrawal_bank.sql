-- ============================================================================
-- WITHDRAWAL REQUESTS AND BANK ACCOUNTS TABLES
-- Purpose: Handle user withdrawal requests and bank account management
-- ============================================================================

-- Bank Accounts Table
CREATE TABLE IF NOT EXISTS public.bank_accounts_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- User reference
    user_id UUID NOT NULL REFERENCES public.users_2026_01_30_14_00(id) ON DELETE CASCADE,
    
    -- Bank details
    account_name TEXT NOT NULL,
    account_number TEXT NOT NULL,
    bank_name TEXT NOT NULL,
    bank_code TEXT NOT NULL,
    
    -- Verification
    is_verified BOOLEAN DEFAULT false,
    is_primary BOOLEAN DEFAULT false,
    verification_method TEXT, -- 'BVN', 'MANUAL', 'PAYSTACK'
    verification_data JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_account_number CHECK (LENGTH(account_number) = 10),
    UNIQUE(user_id, account_number, bank_code)
);

-- Withdrawal Requests Table
CREATE TABLE IF NOT EXISTS public.withdrawal_requests_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- User and bank reference
    user_id UUID NOT NULL REFERENCES public.users_2026_01_30_14_00(id) ON DELETE CASCADE,
    bank_account_id UUID NOT NULL REFERENCES public.bank_accounts_2026_01_30_14_00(id),
    
    -- Amount
    amount DECIMAL(12,2) NOT NULL,
    fee DECIMAL(12,2) DEFAULT 0,
    net_amount DECIMAL(12,2) NOT NULL, -- amount - fee
    
    -- Status workflow
    status TEXT DEFAULT 'PENDING' CHECK (status IN (
        'PENDING',      -- Awaiting approval
        'APPROVED',     -- Approved by admin
        'PROCESSING',   -- Being processed
        'COMPLETED',    -- Successfully completed
        'REJECTED',     -- Rejected by admin
        'FAILED',       -- Failed during processing
        'CANCELLED'     -- Cancelled by user
    )),
    
    -- Admin actions
    approved_by_admin_id UUID REFERENCES public.admin_users_2026_01_30_14_00(id),
    rejection_reason TEXT,
    admin_notes TEXT,
    
    -- Payment details
    transaction_reference TEXT UNIQUE,
    bank_reference TEXT,
    payment_provider TEXT, -- 'PAYSTACK', 'FLUTTERWAVE', 'MANUAL'
    provider_response JSONB,
    
    -- Wallet transaction
    wallet_transaction_id UUID REFERENCES public.wallet_transactions_2026_01_30_14_00(id),
    
    -- Timestamps
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    approved_at TIMESTAMP WITH TIME ZONE,
    processing_started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    rejected_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    request_ip TEXT,
    request_user_agent TEXT,
    
    -- Constraints
    CONSTRAINT valid_withdrawal_amount CHECK (amount > 0),
    CONSTRAINT valid_net_amount CHECK (net_amount = amount - fee)
);

-- Indexes for bank_accounts
CREATE INDEX idx_bank_accounts_user ON public.bank_accounts_2026_01_30_14_00(user_id);
CREATE INDEX idx_bank_accounts_verified ON public.bank_accounts_2026_01_30_14_00(is_verified);
CREATE INDEX idx_bank_accounts_primary ON public.bank_accounts_2026_01_30_14_00(user_id, is_primary) WHERE is_primary = true;

-- Indexes for withdrawal_requests
CREATE INDEX idx_withdrawal_user ON public.withdrawal_requests_2026_01_30_14_00(user_id);
CREATE INDEX idx_withdrawal_status ON public.withdrawal_requests_2026_01_30_14_00(status);
CREATE INDEX idx_withdrawal_requested ON public.withdrawal_requests_2026_01_30_14_00(requested_at DESC);
CREATE INDEX idx_withdrawal_pending ON public.withdrawal_requests_2026_01_30_14_00(status, requested_at) WHERE status = 'PENDING';

-- Ensure only one primary bank account per user
CREATE OR REPLACE FUNCTION ensure_single_primary_bank()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_primary = true THEN
        UPDATE public.bank_accounts_2026_01_30_14_00
        SET is_primary = false
        WHERE user_id = NEW.user_id
        AND id != NEW.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ensure_single_primary_bank
    AFTER INSERT OR UPDATE ON public.bank_accounts_2026_01_30_14_00
    FOR EACH ROW
    WHEN (NEW.is_primary = true)
    EXECUTE FUNCTION ensure_single_primary_bank();

-- Update timestamps on status change
CREATE OR REPLACE FUNCTION update_withdrawal_timestamps()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'APPROVED' AND OLD.status != 'APPROVED' THEN
        NEW.approved_at = NOW();
    ELSIF NEW.status = 'PROCESSING' AND OLD.status != 'PROCESSING' THEN
        NEW.processing_started_at = NOW();
    ELSIF NEW.status = 'COMPLETED' AND OLD.status != 'COMPLETED' THEN
        NEW.completed_at = NOW();
    ELSIF NEW.status = 'REJECTED' AND OLD.status != 'REJECTED' THEN
        NEW.rejected_at = NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_withdrawal_timestamps
    BEFORE UPDATE ON public.withdrawal_requests_2026_01_30_14_00
    FOR EACH ROW
    WHEN (NEW.status IS DISTINCT FROM OLD.status)
    EXECUTE FUNCTION update_withdrawal_timestamps();

COMMENT ON TABLE public.bank_accounts_2026_01_30_14_00 IS 'User bank account details for withdrawals';
COMMENT ON TABLE public.withdrawal_requests_2026_01_30_14_00 IS 'Withdrawal requests with approval workflow';
