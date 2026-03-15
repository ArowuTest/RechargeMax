-- ============================================================
-- Table: daily_subscription_config
-- ============================================================

CREATE TABLE public.daily_subscription_config (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    amount numeric(5,2) NOT NULL,
    draw_entries_earned integer DEFAULT 1,
    is_paid boolean DEFAULT true,
    description text,
    terms_and_conditions text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT positive_amount CHECK ((amount > (0)::numeric)),
    CONSTRAINT positive_entries CHECK ((draw_entries_earned > 0))
);

ALTER TABLE ONLY public.daily_subscription_config
    ADD CONSTRAINT daily_subscription_config_pkey PRIMARY KEY (id);

CREATE TRIGGER update_daily_subscription_config_updated_at BEFORE UPDATE ON public.daily_subscription_config FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
