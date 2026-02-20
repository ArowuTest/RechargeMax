-- ============================================================================
-- WALLET TRANSACTIONS TABLE
-- Purpose: Track all wallet debits and credits for complete financial audit
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.wallet_transactions_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- User reference
    user_id UUID NOT NULL REFERENCES public.users_2026_01_30_14_00(id) ON DELETE CASCADE,
    
    -- Transaction details
    transaction_type TEXT NOT NULL CHECK (transaction_type IN ('CREDIT', 'DEBIT')),
    amount DECIMAL(12,2) NOT NULL,
    
    -- Balance tracking
    balance_before DECIMAL(12,2) NOT NULL,
    balance_after DECIMAL(12,2) NOT NULL,
    
    -- Transaction source/reason
    source TEXT NOT NULL CHECK (source IN (
        'RECHARGE', 'COMMISSION', 'REFERRAL_BONUS', 'SPIN_WIN', 'DRAW_WIN',
        'WITHDRAWAL', 'REFUND', 'REVERSAL', 'ADMIN_CREDIT', 'ADMIN_DEBIT',
        'SUBSCRIPTION_FEE', 'PLATFORM_FEE', 'CASHBACK'
    )),
    
    -- References
    reference TEXT UNIQUE NOT NULL,
    related_transaction_id UUID, -- Reference to transactions, withdrawals, etc.
    description TEXT,
    
    -- Status
    status TEXT DEFAULT 'COMPLETED' CHECK (status IN ('PENDING', 'COMPLETED', 'REVERSED', 'FAILED')),
    
    -- Metadata
    metadata JSONB,
    admin_id UUID, -- If admin initiated
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    reversed_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_amount CHECK (amount > 0),
    CONSTRAINT valid_balance_calculation CHECK (
        (transaction_type = 'CREDIT' AND balance_after = balance_before + amount) OR
        (transaction_type = 'DEBIT' AND balance_after = balance_before - amount)
    )
);

-- Indexes
CREATE INDEX idx_wallet_trans_user ON public.wallet_transactions_2026_01_30_14_00(user_id);
CREATE INDEX idx_wallet_trans_type ON public.wallet_transactions_2026_01_30_14_00(transaction_type);
CREATE INDEX idx_wallet_trans_source ON public.wallet_transactions_2026_01_30_14_00(source);
CREATE INDEX idx_wallet_trans_status ON public.wallet_transactions_2026_01_30_14_00(status);
CREATE INDEX idx_wallet_trans_reference ON public.wallet_transactions_2026_01_30_14_00(reference);
CREATE INDEX idx_wallet_trans_created ON public.wallet_transactions_2026_01_30_14_00(created_at DESC);
CREATE INDEX idx_wallet_trans_user_created ON public.wallet_transactions_2026_01_30_14_00(user_id, created_at DESC);

-- Function to get user wallet balance
CREATE OR REPLACE FUNCTION get_user_wallet_balance(p_user_id UUID)
RETURNS DECIMAL(12,2) AS $$
DECLARE
    v_balance DECIMAL(12,2);
BEGIN
    SELECT COALESCE(balance_after, 0) INTO v_balance
    FROM public.wallet_transactions_2026_01_30_14_00
    WHERE user_id = p_user_id
    AND status = 'COMPLETED'
    ORDER BY created_at DESC
    LIMIT 1;
    
    RETURN COALESCE(v_balance, 0);
END;
$$ LANGUAGE plpgsql;

-- Function to validate sufficient balance before debit
CREATE OR REPLACE FUNCTION validate_wallet_balance()
RETURNS TRIGGER AS $$
DECLARE
    v_current_balance DECIMAL(12,2);
BEGIN
    IF NEW.transaction_type = 'DEBIT' THEN
        v_current_balance := get_user_wallet_balance(NEW.user_id);
        
        IF v_current_balance < NEW.amount THEN
            RAISE EXCEPTION 'Insufficient wallet balance. Current: %, Required: %', v_current_balance, NEW.amount;
        END IF;
        
        -- Verify balance_before matches current balance
        IF NEW.balance_before != v_current_balance THEN
            RAISE EXCEPTION 'Balance mismatch. Expected: %, Provided: %', v_current_balance, NEW.balance_before;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_validate_wallet_balance
    BEFORE INSERT ON public.wallet_transactions_2026_01_30_14_00
    FOR EACH ROW
    EXECUTE FUNCTION validate_wallet_balance();

-- Function to auto-update completed_at
CREATE OR REPLACE FUNCTION update_wallet_transaction_completed()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'COMPLETED' AND OLD.status != 'COMPLETED' THEN
        NEW.completed_at = NOW();
    ELSIF NEW.status = 'REVERSED' AND OLD.status != 'REVERSED' THEN
        NEW.reversed_at = NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_wallet_completed
    BEFORE UPDATE ON public.wallet_transactions_2026_01_30_14_00
    FOR EACH ROW
    WHEN (NEW.status IS DISTINCT FROM OLD.status)
    EXECUTE FUNCTION update_wallet_transaction_completed();

COMMENT ON TABLE public.wallet_transactions_2026_01_30_14_00 IS 'Complete audit trail of all wallet transactions with balance tracking';
