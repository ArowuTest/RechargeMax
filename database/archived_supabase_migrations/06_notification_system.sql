-- Comprehensive notification management for RechargeMax platform
-- Created: 2026-01-30 14:00 UTC
-- ============================================================================

-- ============================================================================
-- NOTIFICATION TEMPLATES
-- ============================================================================

CREATE TABLE public.notification_templates_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Template details
    template_key TEXT UNIQUE NOT NULL,
    template_name TEXT NOT NULL,
    description TEXT,
    
    -- Template content
    title_template TEXT NOT NULL,
    body_template TEXT NOT NULL,
    email_subject_template TEXT,
    email_body_template TEXT,
    sms_template TEXT,
    
    -- Template variables (JSON array of variable names)
    variables JSONB DEFAULT '[]'::jsonb,
    
    -- Delivery channels
    supports_push BOOLEAN DEFAULT true,
    supports_email BOOLEAN DEFAULT true,
    supports_sms BOOLEAN DEFAULT false,
    supports_in_app BOOLEAN DEFAULT true,
    
    -- Template settings
    is_active BOOLEAN DEFAULT true,
    priority TEXT DEFAULT 'NORMAL' CHECK (priority IN ('LOW', 'NORMAL', 'HIGH', 'URGENT')),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_template_key CHECK (template_key ~ '^[a-z0-9_]+$')
);

-- ============================================================================
-- USER NOTIFICATIONS
-- ============================================================================

CREATE TABLE public.user_notifications_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id) ON DELETE CASCADE,
    template_id UUID REFERENCES public.notification_templates_2026_01_30_14_00(id),
    
    -- Notification content
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    
    -- Notification metadata
    notification_type TEXT NOT NULL, -- 'transaction', 'prize', 'draw', 'affiliate', 'system'
    reference_id UUID, -- ID of related entity (transaction, prize, etc.)
    reference_type TEXT, -- Type of related entity
    
    -- Delivery channels
    channels JSONB DEFAULT '["in_app"]'::jsonb, -- Array of delivery channels
    
    -- Status tracking
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP WITH TIME ZONE,
    
    -- Delivery status
    delivery_status JSONB DEFAULT '{}'::jsonb, -- Status per channel
    delivery_attempts INTEGER DEFAULT 0,
    last_delivery_attempt TIMESTAMP WITH TIME ZONE,
    
    -- Priority and scheduling
    priority TEXT DEFAULT 'NORMAL' CHECK (priority IN ('LOW', 'NORMAL', 'HIGH', 'URGENT')),
    scheduled_for TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_notification_type CHECK (notification_type IN ('transaction', 'prize', 'draw', 'affiliate', 'system', 'promotional', 'security'))
);

-- ============================================================================
-- NOTIFICATION DELIVERY LOG
-- ============================================================================

CREATE TABLE public.notification_delivery_log_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    notification_id UUID REFERENCES public.user_notifications_2026_01_30_14_00(id) ON DELETE CASCADE,
    
    -- Delivery details
    channel TEXT NOT NULL, -- 'push', 'email', 'sms', 'in_app'
    delivery_status TEXT NOT NULL, -- 'pending', 'sent', 'delivered', 'failed', 'bounced'
    
    -- Provider details
    provider_name TEXT, -- 'firebase', 'resend', 'twilio', etc.
    provider_message_id TEXT,
    provider_response JSONB,
    
    -- Error handling
    error_code TEXT,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    
    -- Timestamps
    attempted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    delivered_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_channel CHECK (channel IN ('push', 'email', 'sms', 'in_app')),
    CONSTRAINT valid_delivery_status CHECK (delivery_status IN ('pending', 'sent', 'delivered', 'failed', 'bounced', 'opened', 'clicked'))
);

-- ============================================================================
-- USER NOTIFICATION PREFERENCES
-- ============================================================================

CREATE TABLE public.user_notification_preferences_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id) ON DELETE CASCADE,
    
    -- Notification type preferences
    transaction_notifications JSONB DEFAULT '{"push": true, "email": true, "sms": false}'::jsonb,
    prize_notifications JSONB DEFAULT '{"push": true, "email": true, "sms": true}'::jsonb,
    draw_notifications JSONB DEFAULT '{"push": true, "email": true, "sms": false}'::jsonb,
    affiliate_notifications JSONB DEFAULT '{"push": true, "email": true, "sms": false}'::jsonb,
    promotional_notifications JSONB DEFAULT '{"push": true, "email": false, "sms": false}'::jsonb,
    security_notifications JSONB DEFAULT '{"push": true, "email": true, "sms": true}'::jsonb,
    
    -- Global preferences
    do_not_disturb_start TIME,
    do_not_disturb_end TIME,
    timezone TEXT DEFAULT 'Africa/Lagos',
    
    -- Contact preferences
    preferred_language TEXT DEFAULT 'en',
    email_frequency TEXT DEFAULT 'immediate' CHECK (email_frequency IN ('immediate', 'daily', 'weekly', 'never')),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    UNIQUE(user_id)
);

-- ============================================================================
-- ENABLE RLS ON ALL NOTIFICATION TABLES
-- ============================================================================

ALTER TABLE public.notification_templates_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.user_notifications_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.notification_delivery_log_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.user_notification_preferences_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- RLS POLICIES
-- ============================================================================

-- Notification Templates (public read, service role manage)
CREATE POLICY "templates_select_public" ON public.notification_templates_2026_01_30_14_00
    FOR SELECT USING (is_active = true);

CREATE POLICY "templates_service_manage" ON public.notification_templates_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- User Notifications (users see their own)
CREATE POLICY "notifications_select_own" ON public.user_notifications_2026_01_30_14_00
    FOR SELECT USING (auth.uid() = (SELECT auth_user_id FROM public.users_2026_01_30_14_00 WHERE id = user_id));

CREATE POLICY "notifications_update_own" ON public.user_notifications_2026_01_30_14_00
    FOR UPDATE USING (auth.uid() = (SELECT auth_user_id FROM public.users_2026_01_30_14_00 WHERE id = user_id));

CREATE POLICY "notifications_service_manage" ON public.user_notifications_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Delivery Log (service role only)
CREATE POLICY "delivery_log_service_only" ON public.notification_delivery_log_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- User Preferences (users manage their own)
CREATE POLICY "preferences_select_own" ON public.user_notification_preferences_2026_01_30_14_00
    FOR SELECT USING (auth.uid() = (SELECT auth_user_id FROM public.users_2026_01_30_14_00 WHERE id = user_id));

CREATE POLICY "preferences_insert_own" ON public.user_notification_preferences_2026_01_30_14_00
    FOR INSERT WITH CHECK (auth.uid() = (SELECT auth_user_id FROM public.users_2026_01_30_14_00 WHERE id = user_id));

CREATE POLICY "preferences_update_own" ON public.user_notification_preferences_2026_01_30_14_00
    FOR UPDATE USING (auth.uid() = (SELECT auth_user_id FROM public.users_2026_01_30_14_00 WHERE id = user_id));

CREATE POLICY "preferences_service_manage" ON public.user_notification_preferences_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Notification Templates
CREATE INDEX idx_notification_templates_template_key ON public.notification_templates_2026_01_30_14_00(template_key);
CREATE INDEX idx_notification_templates_is_active ON public.notification_templates_2026_01_30_14_00(is_active);

-- User Notifications
CREATE INDEX idx_user_notifications_user_id ON public.user_notifications_2026_01_30_14_00(user_id);
CREATE INDEX idx_user_notifications_type ON public.user_notifications_2026_01_30_14_00(notification_type);
CREATE INDEX idx_user_notifications_is_read ON public.user_notifications_2026_01_30_14_00(is_read);
CREATE INDEX idx_user_notifications_created_at ON public.user_notifications_2026_01_30_14_00(created_at DESC);
CREATE INDEX idx_user_notifications_scheduled_for ON public.user_notifications_2026_01_30_14_00(scheduled_for);
CREATE INDEX idx_user_notifications_reference ON public.user_notifications_2026_01_30_14_00(reference_type, reference_id);

-- Delivery Log
CREATE INDEX idx_delivery_log_notification_id ON public.notification_delivery_log_2026_01_30_14_00(notification_id);
CREATE INDEX idx_delivery_log_channel ON public.notification_delivery_log_2026_01_30_14_00(channel);
CREATE INDEX idx_delivery_log_status ON public.notification_delivery_log_2026_01_30_14_00(delivery_status);
CREATE INDEX idx_delivery_log_attempted_at ON public.notification_delivery_log_2026_01_30_14_00(attempted_at DESC);

-- User Preferences
CREATE INDEX idx_user_preferences_user_id ON public.user_notification_preferences_2026_01_30_14_00(user_id);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Updated_at triggers
CREATE TRIGGER update_notification_templates_updated_at 
    BEFORE UPDATE ON public.notification_templates_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_notifications_updated_at 
    BEFORE UPDATE ON public.user_notifications_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_preferences_updated_at 
    BEFORE UPDATE ON public.user_notification_preferences_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- NOTIFICATION FUNCTIONS
-- ============================================================================

-- Function to create notification from template
CREATE OR REPLACE FUNCTION create_notification_from_template(
    p_user_id UUID,
    p_template_key TEXT,
    p_variables JSONB DEFAULT '{}'::jsonb,
    p_reference_id UUID DEFAULT NULL,
    p_reference_type TEXT DEFAULT NULL,
    p_channels JSONB DEFAULT '["in_app"]'::jsonb
)
RETURNS UUID AS $$
DECLARE
    template_record RECORD;
    notification_id UUID;
    processed_title TEXT;
    processed_body TEXT;
    var_key TEXT;
    var_value TEXT;
BEGIN
    -- Get template
    SELECT * INTO template_record
    FROM public.notification_templates_2026_01_30_14_00
    WHERE template_key = p_template_key AND is_active = true;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Template not found: %', p_template_key;
    END IF;
    
    -- Process template variables
    processed_title := template_record.title_template;
    processed_body := template_record.body_template;
    
    -- Replace variables in title and body
    FOR var_key IN SELECT jsonb_object_keys(p_variables) LOOP
        var_value := p_variables ->> var_key;
        processed_title := REPLACE(processed_title, '{{' || var_key || '}}', var_value);
        processed_body := REPLACE(processed_body, '{{' || var_key || '}}', var_value);
    END LOOP;
    
    -- Create notification
    notification_id := uuid_generate_v4();
    
    INSERT INTO public.user_notifications_2026_01_30_14_00 (
        id, user_id, template_id, title, body, notification_type,
        reference_id, reference_type, channels, priority
    ) VALUES (
        notification_id,
        p_user_id,
        template_record.id,
        processed_title,
        processed_body,
        COALESCE(p_reference_type, 'system'),
        p_reference_id,
        p_reference_type,
        p_channels,
        template_record.priority
    );
    
    RETURN notification_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to mark notification as read
CREATE OR REPLACE FUNCTION mark_notification_read(p_notification_id UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE public.user_notifications_2026_01_30_14_00
    SET 
        is_read = true,
        read_at = NOW(),
        updated_at = NOW()
    WHERE id = p_notification_id
    AND auth.uid() = (SELECT auth_user_id FROM public.users_2026_01_30_14_00 WHERE id = user_id);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get user's unread notification count
CREATE OR REPLACE FUNCTION get_unread_notification_count(p_user_id UUID DEFAULT NULL)
RETURNS INTEGER AS $$
DECLARE
    target_user_id UUID;
BEGIN
    -- Use provided user_id or get from auth
    target_user_id := COALESCE(
        p_user_id,
        (SELECT id FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
    );
    
    RETURN (
        SELECT COUNT(*)::INTEGER
        FROM public.user_notifications_2026_01_30_14_00
        WHERE user_id = target_user_id
        AND is_read = false
        AND (expires_at IS NULL OR expires_at > NOW())
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

