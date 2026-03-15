-- ============================================================
-- Table: users
-- ============================================================

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    auth_user_id uuid,
    msisdn text NOT NULL,
    full_name text,
    email text,
    phone_verified boolean DEFAULT false,
    email_verified boolean DEFAULT false,
    date_of_birth date,
    gender text,
    state text,
    city text,
    address text,
    total_points integer DEFAULT 0,
    loyalty_tier text DEFAULT 'BRONZE'::text,
    total_recharge_amount bigint DEFAULT 0,
    total_transactions integer DEFAULT 0,
    last_recharge_date timestamp with time zone,
    referral_code text,
    referred_by uuid,
    total_referrals integer DEFAULT 0,
    is_active boolean DEFAULT true,
    is_verified boolean DEFAULT false,
    kyc_status text DEFAULT 'PENDING'::text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    last_login_at timestamp with time zone,
    user_code character varying(20),
    CONSTRAINT chk_users_msisdn_format CHECK ((msisdn ~ '^234[7-9][0-1][0-9]{8}$'::text)),
    CONSTRAINT users_gender_check CHECK (((gender = ANY (ARRAY['MALE'::text, 'FEMALE'::text, 'OTHER'::text, ''::text])) OR (gender IS NULL))),
    CONSTRAINT users_kyc_status_check CHECK ((kyc_status = ANY (ARRAY['PENDING'::text, 'VERIFIED'::text, 'REJECTED'::text]))),
    CONSTRAINT users_loyalty_tier_check CHECK ((loyalty_tier = ANY (ARRAY['BRONZE'::text, 'SILVER'::text, 'GOLD'::text, 'PLATINUM'::text]))),
    CONSTRAINT valid_email CHECK (((email = ''::text) OR (email IS NULL) OR (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'::text))),
    CONSTRAINT valid_msisdn CHECK ((msisdn ~ '^234[789][01][0-9]{8}$'::text))
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_auth_user_id_key UNIQUE (auth_user_id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_msisdn_key UNIQUE (msisdn);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_referral_code_key UNIQUE (referral_code);

CREATE INDEX idx_users_auth_user_id ON public.users USING btree (auth_user_id);

CREATE INDEX idx_users_loyalty_tier ON public.users USING btree (loyalty_tier);

CREATE INDEX idx_users_msisdn ON public.users USING btree (msisdn);

CREATE UNIQUE INDEX idx_users_referral_code ON public.users USING btree (referral_code) WHERE (referral_code IS NOT NULL);

CREATE INDEX idx_users_referred_by ON public.users USING btree (referred_by);

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_referred_by_fkey FOREIGN KEY (referred_by) REFERENCES public.users(id);

ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;
