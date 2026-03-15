-- ============================================================
-- Table: notification_delivery_log
-- ============================================================

CREATE TABLE public.notification_delivery_log (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    notification_id uuid,
    channel text NOT NULL,
    delivery_status text NOT NULL,
    provider_name text,
    provider_message_id text,
    provider_response jsonb,
    error_code text,
    error_message text,
    retry_count integer DEFAULT 0,
    attempted_at timestamp with time zone DEFAULT now(),
    delivered_at timestamp with time zone,
    CONSTRAINT valid_channel CHECK ((channel = ANY (ARRAY['push'::text, 'email'::text, 'sms'::text, 'in_app'::text]))),
    CONSTRAINT valid_delivery_status CHECK ((delivery_status = ANY (ARRAY['pending'::text, 'sent'::text, 'delivered'::text, 'failed'::text, 'bounced'::text, 'opened'::text, 'clicked'::text])))
);

ALTER TABLE ONLY public.notification_delivery_log
    ADD CONSTRAINT notification_delivery_log_pkey PRIMARY KEY (id);

CREATE INDEX idx_delivery_log_attempted_at ON public.notification_delivery_log USING btree (attempted_at DESC);

CREATE INDEX idx_delivery_log_channel ON public.notification_delivery_log USING btree (channel);

CREATE INDEX idx_delivery_log_notification_id ON public.notification_delivery_log USING btree (notification_id);

CREATE INDEX idx_delivery_log_status ON public.notification_delivery_log USING btree (delivery_status);

ALTER TABLE ONLY public.notification_delivery_log
    ADD CONSTRAINT notification_delivery_log_notification_id_fkey FOREIGN KEY (notification_id) REFERENCES public.user_notifications(id) ON DELETE CASCADE;

ALTER TABLE public.notification_delivery_log ENABLE ROW LEVEL SECURITY;
