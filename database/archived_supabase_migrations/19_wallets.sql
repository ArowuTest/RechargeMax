-- ============================================================================
-- WALLETS TABLE MIGRATION
-- Created: 2026-01-31
-- Description: Main wallet balance table for affiliate earnings
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- User reference
    msisdn TEXT UNIQUE NOT NULL,
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id) ON DELETE CASCADE,
    
    -- Balance fields (stored in kobo - smallest currency unit)
    balance BIGINT DEFAULT 0 CHECK (balance >= 0),
    pending_balance BIGINT DEFAULT 0 CHECK (pending_balance >= 0),
    total_earned BIGINT DEFAULT 0,
    total_withdrawn BIGINT DEFAULT 0,
    
    -- Withdrawal settings
    min_payout_amount BIGINT DEFAULT 100000, -- ₦1000 minimum (100000 kobo)
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_suspended BOOLEAN DEFAULT false,
    suspension_reason TEXT,
    
    -- Tracking
    last_transaction_at TIMESTAMP WITH TIME ZONE,
    last_withdrawal_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_msisdn CHECK (msisdn ~ '^234[789][01][0-9]{8}$')
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_wallets_msisdn ON public.wallets(msisdn);
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON public.wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_wallets_balance ON public.wallets(balance) WHERE balance > 0;
CREATE INDEX IF NOT EXISTS idx_wallets_pending ON public.wallets(pending_balance) WHERE pending_balance > 0;
CREATE INDEX IF NOT EXISTS idx_wallets_active ON public.wallets(is_active) WHERE is_active = true;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_wallets_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_wallets_updated_at
    BEFORE UPDATE ON public.wallets
    FOR EACH ROW
    EXECUTE FUNCTION update_wallets_updated_at();

-- Comments
COMMENT ON TABLE public.wallets IS 'Main wallet balance table for affiliate earnings and withdrawals';
COMMENT ON COLUMN public.wallets.balance IS 'Available balance in kobo (₦1 = 100 kobo)';
COMMENT ON COLUMN public.wallets.pending_balance IS 'Pending balance during holding period in kobo';
COMMENT ON COLUMN public.wallets.min_payout_amount IS 'Minimum withdrawal amount in kobo (default ₦1000)';
