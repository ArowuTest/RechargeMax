-- ============================================================
-- Table: otp_rate_limits
-- ============================================================

CREATE TABLE public.otp_rate_limits (
    id bigint NOT NULL,
    key character varying(64) NOT NULL,
    requested_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.otp_rate_limits ALTER COLUMN id SET DEFAULT nextval('public.otp_rate_limits_id_seq'::regclass);

ALTER TABLE ONLY public.otp_rate_limits
    ADD CONSTRAINT otp_rate_limits_pkey PRIMARY KEY (id);

CREATE INDEX idx_otp_rate_limits_key_time ON public.otp_rate_limits USING btree (key, requested_at);

CREATE INDEX idx_otp_rate_limits_time ON public.otp_rate_limits USING btree (requested_at);
