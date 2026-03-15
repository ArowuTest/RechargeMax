-- ============================================================
-- Table: payment_logs
-- ============================================================

CREATE TABLE public.payment_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    transaction_id uuid,
    user_id uuid,
    event_type text NOT NULL,
    payment_provider text DEFAULT 'PAYSTACK'::text,
    payment_reference text,
    request_payload jsonb,
    response_payload jsonb,
    status_code integer,
    error_message text,
    error_code text,
    ip_address inet,
    user_agent text,
    request_id text,
    response_time_ms integer,
    amount numeric(12,2),
    currency text DEFAULT 'NGN'::text,
    payment_method text,
    is_successful boolean,
    is_retry boolean DEFAULT false,
    retry_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT payment_logs_event_type_check CHECK ((event_type = ANY (ARRAY['INITIALIZE'::text, 'VERIFY'::text, 'CALLBACK'::text, 'WEBHOOK'::text, 'REFUND'::text, 'DISPUTE'::text, 'CHARGEBACK'::text, 'RETRY'::text])))
);

ALTER TABLE ONLY public.payment_logs
    ADD CONSTRAINT payment_logs_pkey PRIMARY KEY (id);

CREATE INDEX idx_payment_logs_created_at ON public.payment_logs USING btree (created_at);

CREATE INDEX idx_payment_logs_errors ON public.payment_logs USING btree (event_type, is_successful, created_at) WHERE (is_successful = false);

CREATE INDEX idx_payment_logs_event_type ON public.payment_logs USING btree (event_type);

CREATE INDEX idx_payment_logs_is_successful ON public.payment_logs USING btree (is_successful);

CREATE INDEX idx_payment_logs_payment_reference ON public.payment_logs USING btree (payment_reference);

CREATE INDEX idx_payment_logs_slow_requests ON public.payment_logs USING btree (response_time_ms, created_at) WHERE (response_time_ms > 5000);

CREATE INDEX idx_payment_logs_status_code ON public.payment_logs USING btree (status_code);

CREATE INDEX idx_payment_logs_transaction_id ON public.payment_logs USING btree (transaction_id);

CREATE INDEX idx_payment_logs_user_id ON public.payment_logs USING btree (user_id);

ALTER TABLE ONLY public.payment_logs
    ADD CONSTRAINT payment_logs_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.payment_logs
    ADD CONSTRAINT payment_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;
