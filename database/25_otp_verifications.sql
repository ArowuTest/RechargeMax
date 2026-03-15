-- ============================================================
-- Table: otp_verifications
-- ============================================================

CREATE TABLE public.otp_verifications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    msisdn text NOT NULL,
    user_id uuid,
    otp_code_hash text NOT NULL,
    purpose text DEFAULT 'LOGIN'::text NOT NULL,
    is_verified boolean DEFAULT false,
    is_expired boolean DEFAULT false,
    is_revoked boolean DEFAULT false,
    attempts integer DEFAULT 0,
    max_attempts integer DEFAULT 5,
    last_attempt_at timestamp with time zone,
    request_ip inet,
    request_user_agent text,
    device_fingerprint text,
    verified_at timestamp with time zone,
    verified_ip inet,
    verified_user_agent text,
    created_at timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    metadata jsonb DEFAULT '{}'::jsonb,
    CONSTRAINT chk_otp_msisdn_format CHECK ((msisdn ~ '^234[7-9][0-1][0-9]{8}$'::text)),
    CONSTRAINT otp_verifications_purpose_check CHECK ((purpose = ANY (ARRAY['LOGIN'::text, 'REGISTRATION'::text, 'PASSWORD_RESET'::text, 'TRANSACTION_VERIFICATION'::text, 'PHONE_VERIFICATION'::text, 'WITHDRAWAL'::text, 'PROFILE_UPDATE'::text, 'TWO_FACTOR_AUTH'::text]))),
    CONSTRAINT valid_attempts CHECK (((attempts >= 0) AND (attempts <= max_attempts))),
    CONSTRAINT valid_expiry CHECK ((expires_at > created_at)),
    CONSTRAINT valid_msisdn_otp CHECK ((msisdn ~ '^(234|0)?[789][01][0-9]{8,9}$'::text))
);

ALTER TABLE ONLY public.otp_verifications
    ADD CONSTRAINT otp_verifications_pkey PRIMARY KEY (id);

CREATE INDEX idx_otp_verifications_active_lookup ON public.otp_verifications USING btree (msisdn, purpose, is_verified, expires_at) WHERE ((is_verified = false) AND (is_expired = false) AND (is_revoked = false));

CREATE INDEX idx_otp_verifications_created_at ON public.otp_verifications USING btree (created_at);

CREATE INDEX idx_otp_verifications_expires_at ON public.otp_verifications USING btree (expires_at) WHERE (is_verified = false);

CREATE INDEX idx_otp_verifications_is_verified ON public.otp_verifications USING btree (is_verified);

CREATE INDEX idx_otp_verifications_msisdn ON public.otp_verifications USING btree (msisdn);

CREATE INDEX idx_otp_verifications_purpose ON public.otp_verifications USING btree (purpose);

CREATE INDEX idx_otp_verifications_user_id ON public.otp_verifications USING btree (user_id);

CREATE TRIGGER trigger_mark_expired_otps BEFORE UPDATE ON public.otp_verifications FOR EACH ROW EXECUTE FUNCTION public.mark_expired_otps();

ALTER TABLE ONLY public.otp_verifications
    ADD CONSTRAINT otp_verifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
