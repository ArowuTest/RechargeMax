-- ============================================================================
-- SERVICE PRICING TABLE
-- Purpose: Manage dynamic pricing and service availability for network services
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.service_pricing_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Service identification
    network_provider TEXT NOT NULL CHECK (network_provider IN ('MTN', 'AIRTEL', 'GLO', 'NINE_MOBILE')),
    service_type TEXT NOT NULL CHECK (service_type IN ('AIRTIME', 'DATA')),
    data_bundle_code TEXT, -- For DATA type, reference to specific bundle
    
    -- Pricing information
    base_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    selling_price DECIMAL(12,2) NOT NULL,
    commission_rate DECIMAL(5,2) NOT NULL DEFAULT 0, -- Percentage
    platform_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    
    -- Limits
    min_amount DECIMAL(12,2) NOT NULL DEFAULT 50,
    max_amount DECIMAL(12,2) NOT NULL DEFAULT 50000,
    
    -- Availability
    is_active BOOLEAN DEFAULT true,
    is_featured BOOLEAN DEFAULT false,
    sort_order INTEGER DEFAULT 0,
    
    -- Metadata
    description TEXT,
    validity_period TEXT, -- e.g., "30 days", "1 month"
    data_volume TEXT, -- e.g., "1GB", "5GB"
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT unique_service_pricing UNIQUE(network_provider, service_type, data_bundle_code),
    CONSTRAINT valid_pricing CHECK (selling_price >= base_price),
    CONSTRAINT valid_commission CHECK (commission_rate >= 0 AND commission_rate <= 100),
    CONSTRAINT valid_limits CHECK (max_amount >= min_amount)
);

-- Indexes
CREATE INDEX idx_service_pricing_network ON public.service_pricing_2026_01_30_14_00(network_provider);
CREATE INDEX idx_service_pricing_type ON public.service_pricing_2026_01_30_14_00(service_type);
CREATE INDEX idx_service_pricing_active ON public.service_pricing_2026_01_30_14_00(is_active);
CREATE INDEX idx_service_pricing_featured ON public.service_pricing_2026_01_30_14_00(is_featured, sort_order);

-- Updated timestamp trigger
CREATE OR REPLACE FUNCTION update_service_pricing_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_service_pricing_updated_at
    BEFORE UPDATE ON public.service_pricing_2026_01_30_14_00
    FOR EACH ROW
    EXECUTE FUNCTION update_service_pricing_updated_at();

-- Seed data for common services
INSERT INTO public.service_pricing_2026_01_30_14_00 
    (network_provider, service_type, selling_price, base_price, commission_rate, min_amount, max_amount, is_active)
VALUES
    ('MTN', 'AIRTIME', 100, 98, 2.0, 50, 50000, true),
    ('AIRTEL', 'AIRTIME', 100, 98, 2.0, 50, 50000, true),
    ('GLO', 'AIRTIME', 100, 98, 2.0, 50, 50000, true),
    ('NINE_MOBILE', 'AIRTIME', 100, 98, 2.0, 50, 50000, true)
ON CONFLICT (network_provider, service_type, data_bundle_code) DO NOTHING;

COMMENT ON TABLE public.service_pricing_2026_01_30_14_00 IS 'Manages pricing and availability for network recharge services';
