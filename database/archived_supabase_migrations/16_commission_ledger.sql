-- ============================================================================
-- COMMISSION LEDGER TABLE
-- Purpose: Track all commission calculations and payments
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.commission_ledger_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- User reference
    user_id UUID NOT NULL REFERENCES public.users_2026_01_30_14_00(id) ON DELETE CASCADE,
    
    -- Transaction reference
    transaction_id UUID REFERENCES public.transactions_2026_01_30_14_00(id),
    vtu_transaction_id UUID REFERENCES public.vtu_transactions_2026_01_30_14_00(id),
    
    -- Commission details
    commission_type TEXT NOT NULL CHECK (commission_type IN (
        'DIRECT_RECHARGE',      -- Commission from own recharge
        'REFERRAL_LEVEL_1',     -- Direct referral commission
        'REFERRAL_LEVEL_2',     -- 2nd level referral
        'REFERRAL_LEVEL_3',     -- 3rd level referral
        'AFFILIATE_COMMISSION', -- Affiliate program commission
        'BONUS',                -- Special bonus
        'CASHBACK',             -- Cashback rewards
        'LOYALTY_BONUS'         -- Loyalty tier bonus
    )),
    
    -- Calculation
    base_amount DECIMAL(12,2) NOT NULL, -- Transaction amount
    commission_rate DECIMAL(5,2) NOT NULL, -- Percentage
    commission_amount DECIMAL(12,2) NOT NULL,
    
    -- Referral chain (if applicable)
    referrer_id UUID REFERENCES public.users_2026_01_30_14_00(id),
    referral_level INTEGER,
    
    -- Status
    status TEXT DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'PAID', 'REVERSED', 'CANCELLED')),
    
    -- Payment tracking
    wallet_transaction_id UUID REFERENCES public.wallet_transactions_2026_01_30_14_00(id),
    paid_at TIMESTAMP WITH TIME ZONE,
    
    -- Reversal tracking
    reversed_at TIMESTAMP WITH TIME ZONE,
    reversal_reason TEXT,
    
    -- Metadata
    description TEXT,
    metadata JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    approved_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_commission_amount CHECK (commission_amount >= 0),
    CONSTRAINT valid_commission_rate CHECK (commission_rate >= 0 AND commission_rate <= 100),
    CONSTRAINT valid_base_amount CHECK (base_amount > 0)
);

-- Indexes
CREATE INDEX idx_commission_user ON public.commission_ledger_2026_01_30_14_00(user_id);
CREATE INDEX idx_commission_type ON public.commission_ledger_2026_01_30_14_00(commission_type);
CREATE INDEX idx_commission_status ON public.commission_ledger_2026_01_30_14_00(status);
CREATE INDEX idx_commission_referrer ON public.commission_ledger_2026_01_30_14_00(referrer_id);
CREATE INDEX idx_commission_transaction ON public.commission_ledger_2026_01_30_14_00(transaction_id);
CREATE INDEX idx_commission_created ON public.commission_ledger_2026_01_30_14_00(created_at DESC);
CREATE INDEX idx_commission_user_status ON public.commission_ledger_2026_01_30_14_00(user_id, status);

-- Function to calculate total pending commissions for a user
CREATE OR REPLACE FUNCTION get_user_pending_commissions(p_user_id UUID)
RETURNS DECIMAL(12,2) AS $$
DECLARE
    v_total DECIMAL(12,2);
BEGIN
    SELECT COALESCE(SUM(commission_amount), 0) INTO v_total
    FROM public.commission_ledger_2026_01_30_14_00
    WHERE user_id = p_user_id
    AND status IN ('PENDING', 'APPROVED');
    
    RETURN v_total;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate total paid commissions for a user
CREATE OR REPLACE FUNCTION get_user_paid_commissions(p_user_id UUID)
RETURNS DECIMAL(12,2) AS $$
DECLARE
    v_total DECIMAL(12,2);
BEGIN
    SELECT COALESCE(SUM(commission_amount), 0) INTO v_total
    FROM public.commission_ledger_2026_01_30_14_00
    WHERE user_id = p_user_id
    AND status = 'PAID';
    
    RETURN v_total;
END;
$$ LANGUAGE plpgsql;

-- Function to auto-update timestamps based on status changes
CREATE OR REPLACE FUNCTION update_commission_status_timestamps()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'APPROVED' AND OLD.status != 'APPROVED' THEN
        NEW.approved_at = NOW();
    ELSIF NEW.status = 'PAID' AND OLD.status != 'PAID' THEN
        NEW.paid_at = NOW();
    ELSIF NEW.status = 'REVERSED' AND OLD.status != 'REVERSED' THEN
        NEW.reversed_at = NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_commission_timestamps
    BEFORE UPDATE ON public.commission_ledger_2026_01_30_14_00
    FOR EACH ROW
    WHEN (NEW.status IS DISTINCT FROM OLD.status)
    EXECUTE FUNCTION update_commission_status_timestamps();

-- Function to create commission entry for a transaction
CREATE OR REPLACE FUNCTION create_transaction_commission(
    p_user_id UUID,
    p_transaction_id UUID,
    p_base_amount DECIMAL,
    p_commission_rate DECIMAL,
    p_commission_type TEXT DEFAULT 'DIRECT_RECHARGE'
)
RETURNS UUID AS $$
DECLARE
    v_commission_id UUID;
    v_commission_amount DECIMAL(12,2);
BEGIN
    v_commission_amount := p_base_amount * (p_commission_rate / 100);
    
    INSERT INTO public.commission_ledger_2026_01_30_14_00 (
        user_id,
        transaction_id,
        commission_type,
        base_amount,
        commission_rate,
        commission_amount,
        status
    ) VALUES (
        p_user_id,
        p_transaction_id,
        p_commission_type,
        p_base_amount,
        p_commission_rate,
        v_commission_amount,
        'APPROVED'
    ) RETURNING id INTO v_commission_id;
    
    RETURN v_commission_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON TABLE public.commission_ledger_2026_01_30_14_00 IS 'Complete ledger of all commission calculations and payments';
