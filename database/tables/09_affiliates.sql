-- ============================================================
-- Table: affiliates
-- ============================================================

CREATE TABLE public.affiliates (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    affiliate_code text NOT NULL,
    status text DEFAULT 'PENDING'::text,
    tier text DEFAULT 'BRONZE'::text,
    commission_rate numeric(5,2) DEFAULT 5.00,
    total_referrals integer DEFAULT 0,
    active_referrals integer DEFAULT 0,
    total_commission numeric(10,2) DEFAULT 0,
    business_name text,
    website_url text,
    social_media_handles jsonb,
    bank_name text,
    account_number text,
    account_name text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    approved_at timestamp with time zone,
    CONSTRAINT affiliates_status_check CHECK ((status = ANY (ARRAY['PENDING'::text, 'APPROVED'::text, 'SUSPENDED'::text, 'REJECTED'::text]))),
    CONSTRAINT affiliates_tier_check CHECK ((tier = ANY (ARRAY['BRONZE'::text, 'SILVER'::text, 'GOLD'::text, 'PLATINUM'::text]))),
    CONSTRAINT positive_commission_rate CHECK ((commission_rate >= (0)::numeric))
);

ALTER TABLE ONLY public.affiliates
    ADD CONSTRAINT affiliates_affiliate_code_key UNIQUE (affiliate_code);

ALTER TABLE ONLY public.affiliates
    ADD CONSTRAINT affiliates_pkey PRIMARY KEY (id);

CREATE INDEX idx_affiliates_affiliate_code ON public.affiliates USING btree (affiliate_code);

CREATE INDEX idx_affiliates_status ON public.affiliates USING btree (status);

CREATE INDEX idx_affiliates_tier ON public.affiliates USING btree (tier);

CREATE INDEX idx_affiliates_user_id ON public.affiliates USING btree (user_id);

CREATE TRIGGER update_affiliates_updated_at BEFORE UPDATE ON public.affiliates FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

ALTER TABLE ONLY public.affiliates
    ADD CONSTRAINT affiliates_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE public.affiliates ENABLE ROW LEVEL SECURITY;
