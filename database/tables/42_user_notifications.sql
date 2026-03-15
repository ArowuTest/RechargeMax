-- ============================================================
-- Table: user_notifications
-- ============================================================

CREATE TABLE public.user_notifications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    template_id uuid,
    title text NOT NULL,
    body text NOT NULL,
    notification_type text NOT NULL,
    reference_id uuid,
    reference_type text,
    channels jsonb DEFAULT '["in_app"]'::jsonb,
    is_read boolean DEFAULT false,
    read_at timestamp with time zone,
    delivery_status jsonb DEFAULT '{}'::jsonb,
    delivery_attempts integer DEFAULT 0,
    last_delivery_attempt timestamp with time zone,
    priority text DEFAULT 'NORMAL'::text,
    scheduled_for timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT user_notifications_priority_check CHECK ((priority = ANY (ARRAY['LOW'::text, 'NORMAL'::text, 'HIGH'::text, 'URGENT'::text]))),
    CONSTRAINT valid_notification_type CHECK ((notification_type = ANY (ARRAY['transaction'::text, 'prize'::text, 'draw'::text, 'affiliate'::text, 'system'::text, 'promotional'::text, 'security'::text])))
);

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications_pkey PRIMARY KEY (id);

CREATE INDEX idx_user_notifications_created_at ON public.user_notifications USING btree (created_at DESC);

CREATE INDEX idx_user_notifications_is_read ON public.user_notifications USING btree (is_read);

CREATE INDEX idx_user_notifications_reference ON public.user_notifications USING btree (reference_type, reference_id);

CREATE INDEX idx_user_notifications_scheduled_for ON public.user_notifications USING btree (scheduled_for);

CREATE INDEX idx_user_notifications_type ON public.user_notifications USING btree (notification_type);

CREATE INDEX idx_user_notifications_user_id ON public.user_notifications USING btree (user_id);

CREATE TRIGGER update_user_notifications_updated_at BEFORE UPDATE ON public.user_notifications FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications_template_id_fkey FOREIGN KEY (template_id) REFERENCES public.notification_templates(id);

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE public.user_notifications ENABLE ROW LEVEL SECURITY;
