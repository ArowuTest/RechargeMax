-- ============================================================================
-- AFFILIATE PAYOUTS TABLE
-- Migration: 10
-- Date: 2026-01-30
-- Purpose: Track affiliate commission payouts and payment processing
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.affiliate_payouts_2026_01_30_14_00 (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- References
    affiliate_id UUID REFERENCES public.affiliates_2026_01_30_14_00(id) ON DELETE SET NULL,
    payout_batch_id UUID DEFAULT uuid_generate_v4(),
    
    -- Payout details
    total_amount DECIMAL(12,2) NOT NULL CHECK (total_amount > 0),
    commission_count INTEGER NOT NULL DEFAULT 0,
    commission_ids JSONB DEFAULT '[]'::jsonb,
    
    -- Payment method
    payout_method TEXT DEFAULT 'BANK_TRANSFER' CHECK (
        payout_method IN ('BANK_TRANSFER', 'MOBILE_MONEY', 'WALLET')
    ),
    
    -- Bank details
    bank_name TEXT,
    account_number TEXT,
    account_name TEXT,
    
    -- Status tracking
    payout_status TEXT DEFAULT 'PENDING' CHECK (
        payout_status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'CANCELLED')
    ),
    payout_reference TEXT,
    
    -- Financial details
    payout_fee DECIMAL(12,2) DEFAULT 0.00,
    net_amount DECIMAL(12,2) NOT NULL,
    
    -- Processing details
    processed_at TIMESTAMP WITH TIME ZONE,
    processed_by UUID REFERENCES public.admin_users_2026_01_30_14_00(id),
    failure_reason TEXT,
    notes TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX idx_affiliate_payouts_affiliate_id ON public.affiliate_payouts_2026_01_30_14_00(affiliate_id);
CREATE INDEX idx_affiliate_payouts_batch_id ON public.affiliate_payouts_2026_01_30_14_00(payout_batch_id);
CREATE INDEX idx_affiliate_payouts_status ON public.affiliate_payouts_2026_01_30_14_00(payout_status);
CREATE INDEX idx_affiliate_payouts_reference ON public.affiliate_payouts_2026_01_30_14_00(payout_reference);
CREATE INDEX idx_affiliate_payouts_created_at ON public.affiliate_payouts_2026_01_30_14_00(created_at);
CREATE INDEX idx_affiliate_payouts_processed_at ON public.affiliate_payouts_2026_01_30_14_00(processed_at);

-- Composite index for pending payouts
CREATE INDEX idx_affiliate_payouts_pending ON public.affiliate_payouts_2026_01_30_14_00(affiliate_id, payout_status, created_at) 
    WHERE payout_status = 'PENDING';

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Update timestamp trigger
CREATE OR REPLACE FUNCTION update_affiliate_payout_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_affiliate_payout_timestamp
    BEFORE UPDATE ON public.affiliate_payouts_2026_01_30_14_00
    FOR EACH ROW
    EXECUTE FUNCTION update_affiliate_payout_timestamp();

-- ============================================================================
-- FUNCTIONS
-- ============================================================================

-- Create payout for affiliate
CREATE OR REPLACE FUNCTION create_affiliate_payout(
    p_affiliate_id UUID,
    p_commission_ids UUID[],
    p_bank_name TEXT,
    p_account_number TEXT,
    p_account_name TEXT
)
RETURNS UUID AS $$
DECLARE
    v_payout_id UUID;
    v_total_amount DECIMAL(12,2);
    v_commission_count INTEGER;
    v_payout_fee DECIMAL(12,2);
    v_net_amount DECIMAL(12,2);
BEGIN
    -- Calculate total amount and count
    SELECT 
        SUM(commission_amount),
        COUNT(*)
    INTO v_total_amount, v_commission_count
    FROM public.affiliate_commissions_2026_01_30_14_00
    WHERE id = ANY(p_commission_ids)
    AND affiliate_id = p_affiliate_id
    AND status = 'APPROVED';
    
    IF v_total_amount IS NULL OR v_total_amount <= 0 THEN
        RAISE EXCEPTION 'No approved commissions found';
    END IF;
    
    -- Calculate fee (e.g., 1% or ₦100, whichever is higher)
    v_payout_fee := GREATEST(v_total_amount * 0.01, 100);
    v_net_amount := v_total_amount - v_payout_fee;
    
    -- Create payout record
    INSERT INTO public.affiliate_payouts_2026_01_30_14_00 (
        affiliate_id,
        total_amount,
        commission_count,
        commission_ids,
        bank_name,
        account_number,
        account_name,
        payout_fee,
        net_amount
    ) VALUES (
        p_affiliate_id,
        v_total_amount,
        v_commission_count,
        to_jsonb(p_commission_ids),
        p_bank_name,
        p_account_number,
        p_account_name,
        v_payout_fee,
        v_net_amount
    ) RETURNING id INTO v_payout_id;
    
    -- Update commission status
    UPDATE public.affiliate_commissions_2026_01_30_14_00
    SET status = 'PAID',
        paid_at = NOW()
    WHERE id = ANY(p_commission_ids);
    
    RETURN v_payout_id;
END;
$$ LANGUAGE plpgsql;

-- Get payout statistics
CREATE OR REPLACE FUNCTION get_affiliate_payout_stats(p_affiliate_id UUID)
RETURNS TABLE(
    total_payouts BIGINT,
    total_paid DECIMAL,
    pending_amount DECIMAL,
    last_payout_date TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*)::BIGINT as total_payouts,
        COALESCE(SUM(net_amount) FILTER (WHERE payout_status = 'COMPLETED'), 0) as total_paid,
        COALESCE(SUM(net_amount) FILTER (WHERE payout_status = 'PENDING'), 0) as pending_amount,
        MAX(processed_at) FILTER (WHERE payout_status = 'COMPLETED') as last_payout_date
    FROM public.affiliate_payouts_2026_01_30_14_00
    WHERE affiliate_id = p_affiliate_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE public.affiliate_payouts_2026_01_30_14_00 IS 'Tracks affiliate commission payouts and payment processing';
COMMENT ON COLUMN public.affiliate_payouts_2026_01_30_14_00.payout_batch_id IS 'Groups multiple payouts processed together';
COMMENT ON COLUMN public.affiliate_payouts_2026_01_30_14_00.commission_ids IS 'JSON array of commission IDs included in this payout';
COMMENT ON COLUMN public.affiliate_payouts_2026_01_30_14_00.payout_fee IS 'Processing fee deducted from payout';
COMMENT ON COLUMN public.affiliate_payouts_2026_01_30_14_00.net_amount IS 'Amount after deducting fees';
