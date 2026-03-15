-- ============================================================
-- Table: affiliate_analytics
-- ============================================================

CREATE TABLE public.affiliate_analytics (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    affiliate_id uuid,
    analytics_date date NOT NULL,
    total_clicks integer DEFAULT 0,
    unique_clicks integer DEFAULT 0,
    conversions integer DEFAULT 0,
    conversion_rate numeric(5,2) DEFAULT 0.00,
    total_commission numeric(12,2) DEFAULT 0.00,
    recharge_commissions numeric(12,2) DEFAULT 0.00,
    subscription_commissions numeric(12,2) DEFAULT 0.00,
    top_referrer_country text,
    top_device_type text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

ALTER TABLE ONLY public.affiliate_analytics
    ADD CONSTRAINT affiliate_analytics_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.affiliate_analytics
    ADD CONSTRAINT unique_affiliate_date UNIQUE (affiliate_id, analytics_date);

CREATE INDEX idx_affiliate_analytics_affiliate_id ON public.affiliate_analytics USING btree (affiliate_id);

CREATE INDEX idx_affiliate_analytics_conversions ON public.affiliate_analytics USING btree (conversions);

CREATE INDEX idx_affiliate_analytics_date ON public.affiliate_analytics USING btree (analytics_date);

CREATE TRIGGER trigger_update_affiliate_analytics_timestamp BEFORE UPDATE ON public.affiliate_analytics FOR EACH ROW EXECUTE FUNCTION public.update_affiliate_analytics_timestamp();

ALTER TABLE ONLY public.affiliate_analytics
    ADD CONSTRAINT affiliate_analytics_affiliate_id_fkey FOREIGN KEY (affiliate_id) REFERENCES public.affiliates(id) ON DELETE CASCADE;
