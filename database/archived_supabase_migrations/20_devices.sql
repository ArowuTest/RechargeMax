-- ============================================================================
-- DEVICES TABLE MIGRATION
-- Created: 2026-01-31
-- Description: Mobile device registration for push notifications
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- User reference
    msisdn TEXT NOT NULL,
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id) ON DELETE CASCADE,
    
    -- Device identification
    device_id TEXT UNIQUE NOT NULL,
    fcm_token TEXT, -- Firebase Cloud Messaging token
    
    -- Device information
    platform TEXT NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    app_version TEXT,
    os_version TEXT,
    device_model TEXT,
    device_name TEXT,
    
    -- Location (optional)
    country_code TEXT DEFAULT 'NG',
    timezone TEXT DEFAULT 'Africa/Lagos',
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_notifications_enabled BOOLEAN DEFAULT true,
    
    -- Tracking
    last_active_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_notification_at TIMESTAMP WITH TIME ZONE,
    notification_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_msisdn CHECK (msisdn ~ '^234[789][01][0-9]{8}$')
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_devices_msisdn ON public.devices(msisdn);
CREATE INDEX IF NOT EXISTS idx_devices_user_id ON public.devices(user_id);
CREATE INDEX IF NOT EXISTS idx_devices_device_id ON public.devices(device_id);
CREATE INDEX IF NOT EXISTS idx_devices_fcm_token ON public.devices(fcm_token) WHERE fcm_token IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_devices_platform ON public.devices(platform);
CREATE INDEX IF NOT EXISTS idx_devices_active ON public.devices(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_devices_last_active ON public.devices(last_active_at DESC);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_devices_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_devices_updated_at
    BEFORE UPDATE ON public.devices
    FOR EACH ROW
    EXECUTE FUNCTION update_devices_updated_at();

-- Function to clean up old inactive devices (optional maintenance)
CREATE OR REPLACE FUNCTION cleanup_inactive_devices()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM public.devices
    WHERE is_active = false
    AND last_active_at < NOW() - INTERVAL '180 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Comments
COMMENT ON TABLE public.devices IS 'Mobile device registration for push notifications';
COMMENT ON COLUMN public.devices.device_id IS 'Unique device identifier (UUID or device-specific ID)';
COMMENT ON COLUMN public.devices.fcm_token IS 'Firebase Cloud Messaging token for push notifications';
COMMENT ON COLUMN public.devices.platform IS 'Device platform: ios, android, or web';
COMMENT ON FUNCTION cleanup_inactive_devices() IS 'Removes devices inactive for more than 180 days';
