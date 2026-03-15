-- ============================================================
-- Table: platform_settings
-- ============================================================

CREATE TABLE public.platform_settings (
    setting_key text NOT NULL,
    setting_value text NOT NULL,
    description text,
    is_public boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

ALTER TABLE ONLY public.platform_settings
    ADD CONSTRAINT platform_settings_pkey PRIMARY KEY (setting_key);

CREATE TRIGGER update_platform_settings_updated_at BEFORE UPDATE ON public.platform_settings FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
