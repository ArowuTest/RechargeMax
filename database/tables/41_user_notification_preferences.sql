-- ============================================================
-- Table: user_notification_preferences
-- ============================================================

CREATE TABLE public.user_notification_preferences (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    transaction_notifications jsonb DEFAULT '{"sms": false, "push": true, "email": true}'::jsonb,
    prize_notifications jsonb DEFAULT '{"sms": true, "push": true, "email": true}'::jsonb,
    draw_notifications jsonb DEFAULT '{"sms": false, "push": true, "email": true}'::jsonb,
    affiliate_notifications jsonb DEFAULT '{"sms": false, "push": true, "email": true}'::jsonb,
    promotional_notifications jsonb DEFAULT '{"sms": false, "push": true, "email": false}'::jsonb,
    security_notifications jsonb DEFAULT '{"sms": true, "push": true, "email": true}'::jsonb,
    do_not_disturb_start time without time zone,
    do_not_disturb_end time without time zone,
    timezone text DEFAULT 'Africa/Lagos'::text,
    preferred_language text DEFAULT 'en'::text,
    email_frequency text DEFAULT 'immediate'::text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT user_notification_preferences_email_frequency_check CHECK ((email_frequency = ANY (ARRAY['immediate'::text, 'daily'::text, 'weekly'::text, 'never'::text])))
);

ALTER TABLE ONLY public.user_notification_preferences
    ADD CONSTRAINT user_notification_preferences_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_notification_preferences
    ADD CONSTRAINT user_notification_preferences_user_id_key UNIQUE (user_id);

CREATE INDEX idx_user_preferences_user_id ON public.user_notification_preferences USING btree (user_id);

CREATE TRIGGER update_user_preferences_updated_at BEFORE UPDATE ON public.user_notification_preferences FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

ALTER TABLE ONLY public.user_notification_preferences
    ADD CONSTRAINT user_notification_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE public.user_notification_preferences ENABLE ROW LEVEL SECURITY;
