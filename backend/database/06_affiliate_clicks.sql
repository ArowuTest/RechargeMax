-- ============================================================
-- Table: affiliate_clicks
-- ============================================================

CREATE TABLE public.affiliate_clicks (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    affiliate_id uuid,
    ip_address inet,
    user_agent text,
    referrer_url text,
    landing_page text,
    converted boolean DEFAULT false,
    conversion_transaction_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    converted_at timestamp with time zone
);

ALTER TABLE ONLY public.affiliate_clicks
    ADD CONSTRAINT affiliate_clicks_pkey PRIMARY KEY (id);

CREATE INDEX idx_affiliate_clicks_affiliate_id ON public.affiliate_clicks USING btree (affiliate_id);

CREATE INDEX idx_affiliate_clicks_converted ON public.affiliate_clicks USING btree (converted);

CREATE INDEX idx_affiliate_clicks_created_at ON public.affiliate_clicks USING btree (created_at DESC);

ALTER TABLE ONLY public.affiliate_clicks
    ADD CONSTRAINT affiliate_clicks_affiliate_id_fkey FOREIGN KEY (affiliate_id) REFERENCES public.affiliates(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.affiliate_clicks
    ADD CONSTRAINT affiliate_clicks_conversion_transaction_id_fkey FOREIGN KEY (conversion_transaction_id) REFERENCES public.transactions(id);
