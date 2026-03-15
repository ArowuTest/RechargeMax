-- ============================================================
-- Table: notification_templates
-- ============================================================

CREATE TABLE public.notification_templates (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    template_key text NOT NULL,
    template_name text NOT NULL,
    description text,
    title_template text NOT NULL,
    body_template text NOT NULL,
    email_subject_template text,
    email_body_template text,
    sms_template text,
    variables jsonb DEFAULT '[]'::jsonb,
    supports_push boolean DEFAULT true,
    supports_email boolean DEFAULT true,
    supports_sms boolean DEFAULT false,
    supports_in_app boolean DEFAULT true,
    is_active boolean DEFAULT true,
    priority text DEFAULT 'NORMAL'::text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT notification_templates_priority_check CHECK ((priority = ANY (ARRAY['LOW'::text, 'NORMAL'::text, 'HIGH'::text, 'URGENT'::text]))),
    CONSTRAINT valid_template_key CHECK ((template_key ~ '^[a-z0-9_]+$'::text))
);

ALTER TABLE ONLY public.notification_templates
    ADD CONSTRAINT notification_templates_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.notification_templates
    ADD CONSTRAINT notification_templates_template_key_key UNIQUE (template_key);

CREATE INDEX idx_notification_templates_is_active ON public.notification_templates USING btree (is_active);

CREATE INDEX idx_notification_templates_template_key ON public.notification_templates USING btree (template_key);

CREATE TRIGGER update_notification_templates_updated_at BEFORE UPDATE ON public.notification_templates FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

ALTER TABLE public.notification_templates ENABLE ROW LEVEL SECURITY;
