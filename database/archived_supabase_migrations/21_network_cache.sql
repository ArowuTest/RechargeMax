-- ============================================================================
-- NETWORK CACHE TABLE MIGRATION
-- Created: 2026-01-31
-- Description: HLR lookup result caching for network detection
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.network_cache (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- MSISDN (unique identifier)
    msisdn TEXT UNIQUE NOT NULL,
    
    -- Network information
    network TEXT NOT NULL CHECK (network IN ('MTN', 'AIRTEL', 'GLO', '9MOBILE')),
    
    -- Lookup metadata
    lookup_source TEXT NOT NULL CHECK (lookup_source IN ('hlr_api', 'user_selection', 'prefix_fallback')),
    confidence_level TEXT DEFAULT 'high' CHECK (confidence_level IN ('high', 'medium', 'low')),
    
    -- Provider information (optional)
    hlr_provider TEXT, -- 'termii', 'africas_talking', 'infobip'
    hlr_response JSONB, -- Full HLR API response for debugging
    
    -- Cache management
    last_verified_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    cache_expires_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() + INTERVAL '60 days',
    verification_count INTEGER DEFAULT 1,
    
    -- Invalidation tracking
    is_valid BOOLEAN DEFAULT true,
    invalidated_at TIMESTAMP WITH TIME ZONE,
    invalidation_reason TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_msisdn CHECK (msisdn ~ '^234[789][01][0-9]{8}$')
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_network_cache_msisdn ON public.network_cache(msisdn);
CREATE INDEX IF NOT EXISTS idx_network_cache_network ON public.network_cache(network);
CREATE INDEX IF NOT EXISTS idx_network_cache_expires ON public.network_cache(cache_expires_at);
CREATE INDEX IF NOT EXISTS idx_network_cache_valid ON public.network_cache(is_valid) WHERE is_valid = true;
CREATE INDEX IF NOT EXISTS idx_network_cache_source ON public.network_cache(lookup_source);
CREATE INDEX IF NOT EXISTS idx_network_cache_confidence ON public.network_cache(confidence_level);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_network_cache_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_network_cache_updated_at
    BEFORE UPDATE ON public.network_cache
    FOR EACH ROW
    EXECUTE FUNCTION update_network_cache_updated_at();

-- Function to clean up expired cache entries
CREATE OR REPLACE FUNCTION cleanup_expired_network_cache()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM public.network_cache
    WHERE cache_expires_at < NOW()
    OR (is_valid = false AND invalidated_at < NOW() - INTERVAL '30 days');
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to invalidate cache for a specific MSISDN
CREATE OR REPLACE FUNCTION invalidate_network_cache(
    p_msisdn TEXT,
    p_reason TEXT DEFAULT 'manual_invalidation'
)
RETURNS BOOLEAN AS $$
BEGIN
    UPDATE public.network_cache
    SET 
        is_valid = false,
        invalidated_at = NOW(),
        invalidation_reason = p_reason,
        updated_at = NOW()
    WHERE msisdn = p_msisdn;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Function to get cached network with expiry check
CREATE OR REPLACE FUNCTION get_cached_network(p_msisdn TEXT)
RETURNS TABLE (
    network TEXT,
    confidence_level TEXT,
    is_expired BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        nc.network,
        nc.confidence_level,
        (nc.cache_expires_at < NOW() OR nc.is_valid = false) AS is_expired
    FROM public.network_cache nc
    WHERE nc.msisdn = p_msisdn;
END;
$$ LANGUAGE plpgsql;

-- Comments
COMMENT ON TABLE public.network_cache IS 'HLR lookup result caching for network detection (60-day TTL)';
COMMENT ON COLUMN public.network_cache.msisdn IS 'Nigerian MSISDN in format 234XXXXXXXXXX';
COMMENT ON COLUMN public.network_cache.network IS 'Detected mobile network operator';
COMMENT ON COLUMN public.network_cache.lookup_source IS 'Source of network detection: hlr_api (highest confidence), user_selection (medium), prefix_fallback (lowest)';
COMMENT ON COLUMN public.network_cache.confidence_level IS 'Confidence in the cached result: high (HLR), medium (user), low (prefix)';
COMMENT ON COLUMN public.network_cache.cache_expires_at IS 'Cache expiry timestamp (default 60 days from creation)';
COMMENT ON FUNCTION cleanup_expired_network_cache() IS 'Removes expired cache entries and old invalidated records';
COMMENT ON FUNCTION invalidate_network_cache(TEXT, TEXT) IS 'Marks a cache entry as invalid (e.g., after failed recharge)';
COMMENT ON FUNCTION get_cached_network(TEXT) IS 'Retrieves cached network with expiry status';
