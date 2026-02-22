-- Migration: Create network_cache table
-- Date: 2026-02-22
-- Description: Create table to cache network lookups for phone numbers

CREATE TABLE IF NOT EXISTS network_cache (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    msisdn TEXT NOT NULL UNIQUE,
    network TEXT NOT NULL,
    last_verified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    cache_expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    lookup_source TEXT, -- 'hlr_api', 'prefix', 'user_selection', 'recharge_history'
    hlr_provider TEXT,
    hlr_response JSONB,
    is_valid BOOLEAN DEFAULT TRUE,
    invalidated_at TIMESTAMP WITH TIME ZONE,
    invalidation_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_network_cache_msisdn ON network_cache(msisdn);
CREATE INDEX IF NOT EXISTS idx_network_cache_expires ON network_cache(cache_expires_at);
CREATE INDEX IF NOT EXISTS idx_network_cache_valid ON network_cache(is_valid);

-- Add comment
COMMENT ON TABLE network_cache IS 'Caches network provider information for phone numbers to reduce HLR API calls';
