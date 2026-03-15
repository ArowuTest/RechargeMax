-- ============================================================
-- Table: prize_fulfillment_logs
-- ============================================================

CREATE TABLE public.prize_fulfillment_logs (
    id bigint NOT NULL,
    spin_result_id uuid NOT NULL,
    attempt_number integer NOT NULL,
    fulfillment_mode character varying(20) NOT NULL,
    provider_name character varying(50),
    provider_reference character varying(100),
    provider_transaction_id bigint,
    request_payload jsonb,
    response_payload jsonb,
    status character varying(20) NOT NULL,
    error_code character varying(50),
    error_message text,
    response_time_ms integer,
    detected_network character varying(20),
    msisdn character varying(20),
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT check_attempt_number CHECK ((attempt_number > 0)),
    CONSTRAINT check_status CHECK (((status)::text = ANY ((ARRAY['SUCCESS'::character varying, 'FAILED'::character varying, 'PENDING'::character varying, 'TIMEOUT'::character varying, 'CANCELLED'::character varying])::text[])))
);

ALTER TABLE ONLY public.prize_fulfillment_logs ALTER COLUMN id SET DEFAULT nextval('public.prize_fulfillment_logs_id_seq'::regclass);

ALTER TABLE ONLY public.prize_fulfillment_logs
    ADD CONSTRAINT prize_fulfillment_logs_pkey PRIMARY KEY (id);

CREATE INDEX idx_fulfillment_logs_created_at ON public.prize_fulfillment_logs USING btree (created_at DESC);

CREATE INDEX idx_fulfillment_logs_msisdn ON public.prize_fulfillment_logs USING btree (msisdn);

CREATE INDEX idx_fulfillment_logs_provider_ref ON public.prize_fulfillment_logs USING btree (provider_reference);

CREATE INDEX idx_fulfillment_logs_spin_result ON public.prize_fulfillment_logs USING btree (spin_result_id);

CREATE INDEX idx_fulfillment_logs_status ON public.prize_fulfillment_logs USING btree (status);
