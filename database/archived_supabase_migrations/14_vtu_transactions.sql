-- ============================================================================
-- VTU TRANSACTIONS TABLE
-- Purpose: Detailed logging of all Virtual Top-Up transactions
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.vtu_transactions_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Transaction identification
    transaction_reference TEXT UNIQUE NOT NULL,
    parent_transaction_id UUID REFERENCES public.transactions_2026_01_30_14_00(id),
    
    -- User and recipient information
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id),
    phone_number TEXT NOT NULL,
    
    -- Service details
    network_provider TEXT NOT NULL CHECK (network_provider IN ('MTN', 'AIRTEL', 'GLO', 'NINE_MOBILE')),
    recharge_type TEXT NOT NULL CHECK (recharge_type IN ('AIRTIME', 'DATA')),
    amount DECIMAL(12,2) NOT NULL,
    data_bundle TEXT, -- For DATA type
    data_bundle_code TEXT,
    
    -- Provider details
    provider_used TEXT, -- Which VTU provider was used
    provider_transaction_id TEXT,
    provider_reference TEXT,
    provider_response JSONB,
    provider_status TEXT,
    
    -- Transaction status
    status TEXT DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'REVERSED')),
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    
    -- Request metadata
    user_agent TEXT,
    ip_address TEXT,
    device_info JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processing_started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    
    -- Error tracking
    error_message TEXT,
    error_code TEXT,
    last_error_at TIMESTAMP WITH TIME ZONE,
    
    -- Reconciliation
    is_reconciled BOOLEAN DEFAULT false,
    reconciled_at TIMESTAMP WITH TIME ZONE,
    reconciliation_notes TEXT,
    
    CONSTRAINT valid_amount CHECK (amount > 0),
    CONSTRAINT valid_retry_count CHECK (retry_count <= max_retries)
);

-- Indexes for performance
CREATE INDEX idx_vtu_trans_reference ON public.vtu_transactions_2026_01_30_14_00(transaction_reference);
CREATE INDEX idx_vtu_trans_user ON public.vtu_transactions_2026_01_30_14_00(user_id);
CREATE INDEX idx_vtu_trans_phone ON public.vtu_transactions_2026_01_30_14_00(phone_number);
CREATE INDEX idx_vtu_trans_status ON public.vtu_transactions_2026_01_30_14_00(status);
CREATE INDEX idx_vtu_trans_network ON public.vtu_transactions_2026_01_30_14_00(network_provider);
CREATE INDEX idx_vtu_trans_provider ON public.vtu_transactions_2026_01_30_14_00(provider_used);
CREATE INDEX idx_vtu_trans_created ON public.vtu_transactions_2026_01_30_14_00(created_at DESC);
CREATE INDEX idx_vtu_trans_reconciled ON public.vtu_transactions_2026_01_30_14_00(is_reconciled) WHERE is_reconciled = false;

-- Function to check for duplicate transactions (idempotency)
CREATE OR REPLACE FUNCTION check_vtu_transaction_duplicate()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if transaction with same reference already exists
    IF EXISTS (
        SELECT 1 FROM public.vtu_transactions_2026_01_30_14_00 
        WHERE transaction_reference = NEW.transaction_reference 
        AND id != NEW.id
    ) THEN
        RAISE EXCEPTION 'Duplicate transaction reference: %', NEW.transaction_reference;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_check_vtu_duplicate
    BEFORE INSERT OR UPDATE ON public.vtu_transactions_2026_01_30_14_00
    FOR EACH ROW
    EXECUTE FUNCTION check_vtu_transaction_duplicate();

-- Function to auto-update completed_at timestamp
CREATE OR REPLACE FUNCTION update_vtu_transaction_completed()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'COMPLETED' AND OLD.status != 'COMPLETED' THEN
        NEW.completed_at = NOW();
    ELSIF NEW.status = 'FAILED' AND OLD.status != 'FAILED' THEN
        NEW.failed_at = NOW();
        NEW.last_error_at = NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_vtu_completed
    BEFORE UPDATE ON public.vtu_transactions_2026_01_30_14_00
    FOR EACH ROW
    WHEN (NEW.status IS DISTINCT FROM OLD.status)
    EXECUTE FUNCTION update_vtu_transaction_completed();

COMMENT ON TABLE public.vtu_transactions_2026_01_30_14_00 IS 'Detailed logging of all VTU transactions for idempotency and reconciliation';
