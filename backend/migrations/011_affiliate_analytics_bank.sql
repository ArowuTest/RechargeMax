-- ============================================================================
-- AFFILIATE ANALYTICS AND BANK ACCOUNTS TABLES
-- Migration: 11
-- Date: 2026-01-30
-- Purpose: Track affiliate performance analytics and bank account details
-- ============================================================================

-- ============================================================================
-- AFFILIATE ANALYTICS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.affiliate_analytics (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- References
    affiliate_id UUID REFERENCES public.affiliates(id) ON DELETE CASCADE,
    
    -- Analytics period
    analytics_date DATE NOT NULL,
    
    -- Click metrics
    total_clicks INTEGER DEFAULT 0,
    unique_clicks INTEGER DEFAULT 0,
    
    -- Conversion metrics
    conversions INTEGER DEFAULT 0,
    conversion_rate DECIMAL(5,2) DEFAULT 0.00,
    
    -- Commission metrics
    total_commission DECIMAL(12,2) DEFAULT 0.00,
    recharge_commissions DECIMAL(12,2) DEFAULT 0.00,
    subscription_commissions DECIMAL(12,2) DEFAULT 0.00,
    
    -- Demographic insights
    top_referrer_country TEXT,
    top_device_type TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT unique_affiliate_date UNIQUE (affiliate_id, analytics_date)
);

-- Analytics Indexes
CREATE INDEX idx_affiliate_analytics_affiliate_id ON public.affiliate_analytics(affiliate_id);
CREATE INDEX idx_affiliate_analytics_date ON public.affiliate_analytics(analytics_date);
CREATE INDEX idx_affiliate_analytics_conversions ON public.affiliate_analytics(conversions);

-- ============================================================================
-- AFFILIATE BANK ACCOUNTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.affiliate_bank_accounts (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- References
    affiliate_id UUID REFERENCES public.affiliates(id) ON DELETE CASCADE,
    
    -- Bank details
    bank_name TEXT NOT NULL,
    account_number TEXT NOT NULL,
    account_name TEXT NOT NULL,
    
    -- Verification
    is_verified BOOLEAN DEFAULT false,
    is_primary BOOLEAN DEFAULT false,
    verified_at TIMESTAMP WITH TIME ZONE,
    verified_by UUID REFERENCES public.admin_users(id),
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_account_number CHECK (length(account_number) >= 10)
);

-- Bank Accounts Indexes
CREATE INDEX idx_affiliate_bank_accounts_affiliate_id ON public.affiliate_bank_accounts(affiliate_id);
CREATE INDEX idx_affiliate_bank_accounts_is_verified ON public.affiliate_bank_accounts(is_verified);
CREATE INDEX idx_affiliate_bank_accounts_is_primary ON public.affiliate_bank_accounts(is_primary) WHERE is_primary = true;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Update timestamp for analytics
CREATE OR REPLACE FUNCTION update_affiliate_analytics_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_affiliate_analytics_timestamp
    BEFORE UPDATE ON public.affiliate_analytics
    FOR EACH ROW
    EXECUTE FUNCTION update_affiliate_analytics_timestamp();

-- Update timestamp for bank accounts
CREATE OR REPLACE FUNCTION update_affiliate_bank_account_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_affiliate_bank_account_timestamp
    BEFORE UPDATE ON public.affiliate_bank_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_affiliate_bank_account_timestamp();

-- Ensure only one primary bank account per affiliate
CREATE OR REPLACE FUNCTION ensure_single_primary_bank_account()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_primary = true THEN
        UPDATE public.affiliate_bank_accounts
        SET is_primary = false
        WHERE affiliate_id = NEW.affiliate_id
        AND id != NEW.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ensure_single_primary_bank_account
    BEFORE INSERT OR UPDATE ON public.affiliate_bank_accounts
    FOR EACH ROW
    WHEN (NEW.is_primary = true)
    EXECUTE FUNCTION ensure_single_primary_bank_account();

-- ============================================================================
-- FUNCTIONS
-- ============================================================================

-- Update daily analytics
CREATE OR REPLACE FUNCTION update_affiliate_daily_analytics(
    p_affiliate_id UUID,
    p_date DATE DEFAULT CURRENT_DATE
)
RETURNS VOID AS $$
DECLARE
    v_total_clicks INTEGER;
    v_unique_clicks INTEGER;
    v_conversions INTEGER;
    v_conversion_rate DECIMAL(5,2);
    v_total_commission DECIMAL(12,2);
    v_recharge_commissions DECIMAL(12,2);
    v_subscription_commissions DECIMAL(12,2);
BEGIN
    -- Calculate metrics
    SELECT 
        COUNT(*),
        COUNT(DISTINCT ip_address),
        COUNT(*) FILTER (WHERE converted = true)
    INTO v_total_clicks, v_unique_clicks, v_conversions
    FROM public.affiliate_clicks
    WHERE affiliate_id = p_affiliate_id
    AND DATE(created_at) = p_date;
    
    -- Calculate conversion rate
    v_conversion_rate := CASE 
        WHEN v_total_clicks > 0 THEN (v_conversions::DECIMAL / v_total_clicks) * 100
        ELSE 0
    END;
    
    -- Calculate commissions
    SELECT 
        COALESCE(SUM(commission_amount), 0),
        COALESCE(SUM(commission_amount) FILTER (WHERE transaction_type = 'RECHARGE'), 0),
        COALESCE(SUM(commission_amount) FILTER (WHERE transaction_type = 'SUBSCRIPTION'), 0)
    INTO v_total_commission, v_recharge_commissions, v_subscription_commissions
    FROM public.affiliate_commissions
    WHERE affiliate_id = p_affiliate_id
    AND DATE(created_at) = p_date;
    
    -- Insert or update analytics
    INSERT INTO public.affiliate_analytics (
        affiliate_id,
        analytics_date,
        total_clicks,
        unique_clicks,
        conversions,
        conversion_rate,
        total_commission,
        recharge_commissions,
        subscription_commissions
    ) VALUES (
        p_affiliate_id,
        p_date,
        v_total_clicks,
        v_unique_clicks,
        v_conversions,
        v_conversion_rate,
        v_total_commission,
        v_recharge_commissions,
        v_subscription_commissions
    )
    ON CONFLICT (affiliate_id, analytics_date)
    DO UPDATE SET
        total_clicks = EXCLUDED.total_clicks,
        unique_clicks = EXCLUDED.unique_clicks,
        conversions = EXCLUDED.conversions,
        conversion_rate = EXCLUDED.conversion_rate,
        total_commission = EXCLUDED.total_commission,
        recharge_commissions = EXCLUDED.recharge_commissions,
        subscription_commissions = EXCLUDED.subscription_commissions,
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Get affiliate performance summary
CREATE OR REPLACE FUNCTION get_affiliate_performance(
    p_affiliate_id UUID,
    p_days INTEGER DEFAULT 30
)
RETURNS TABLE(
    total_clicks BIGINT,
    total_conversions BIGINT,
    avg_conversion_rate DECIMAL,
    total_earned DECIMAL,
    best_day DATE,
    best_day_commission DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        SUM(a.total_clicks)::BIGINT,
        SUM(a.conversions)::BIGINT,
        AVG(a.conversion_rate),
        SUM(a.total_commission),
        (SELECT analytics_date FROM public.affiliate_analytics 
         WHERE affiliate_id = p_affiliate_id 
         ORDER BY total_commission DESC LIMIT 1),
        (SELECT total_commission FROM public.affiliate_analytics 
         WHERE affiliate_id = p_affiliate_id 
         ORDER BY total_commission DESC LIMIT 1)
    FROM public.affiliate_analytics a
    WHERE a.affiliate_id = p_affiliate_id
    AND a.analytics_date > CURRENT_DATE - p_days;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE public.affiliate_analytics IS 'Daily analytics for affiliate performance tracking';
COMMENT ON TABLE public.affiliate_bank_accounts IS 'Bank account details for affiliate payouts';
COMMENT ON COLUMN public.affiliate_bank_accounts.is_primary IS 'Primary bank account for payouts (only one per affiliate)';
