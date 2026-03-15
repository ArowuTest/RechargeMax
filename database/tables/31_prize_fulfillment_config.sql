-- ============================================================
-- Table: prize_fulfillment_config
-- ============================================================

CREATE TABLE public.prize_fulfillment_config (
    id integer NOT NULL,
    prize_type character varying(20) NOT NULL,
    fulfillment_mode character varying(20) DEFAULT 'AUTO'::character varying NOT NULL,
    auto_retry_enabled boolean DEFAULT true,
    max_retry_attempts integer DEFAULT 3,
    retry_delay_seconds integer DEFAULT 300,
    fallback_to_manual boolean DEFAULT true,
    fallback_notification_enabled boolean DEFAULT true,
    provision_timeout_seconds integer DEFAULT 60,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    created_by character varying(100),
    updated_by character varying(100),
    CONSTRAINT check_fulfillment_mode CHECK (((fulfillment_mode)::text = ANY ((ARRAY['AUTO'::character varying, 'MANUAL'::character varying])::text[]))),
    CONSTRAINT check_max_retry_attempts CHECK (((max_retry_attempts >= 0) AND (max_retry_attempts <= 10))),
    CONSTRAINT check_prize_type CHECK (((prize_type)::text = ANY ((ARRAY['AIRTIME'::character varying, 'DATA'::character varying, 'CASH'::character varying, 'POINTS'::character varying, 'PHYSICAL'::character varying])::text[]))),
    CONSTRAINT check_retry_delay CHECK (((retry_delay_seconds >= 0) AND (retry_delay_seconds <= 3600))),
    CONSTRAINT check_timeout CHECK (((provision_timeout_seconds >= 10) AND (provision_timeout_seconds <= 300)))
);

ALTER TABLE ONLY public.prize_fulfillment_config ALTER COLUMN id SET DEFAULT nextval('public.prize_fulfillment_config_id_seq'::regclass);

ALTER TABLE ONLY public.prize_fulfillment_config
    ADD CONSTRAINT prize_fulfillment_config_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.prize_fulfillment_config
    ADD CONSTRAINT unique_prize_type UNIQUE (prize_type);

CREATE INDEX idx_fulfillment_config_active ON public.prize_fulfillment_config USING btree (is_active);

CREATE INDEX idx_fulfillment_config_prize_type ON public.prize_fulfillment_config USING btree (prize_type);
