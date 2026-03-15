-- ============================================================
-- Table: transactions
-- ============================================================

CREATE TABLE public.transactions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    msisdn text NOT NULL,
    network_provider text NOT NULL,
    recharge_type text NOT NULL,
    amount bigint NOT NULL,
    data_plan_id uuid,
    payment_method text NOT NULL,
    payment_reference text,
    payment_gateway text,
    status text DEFAULT 'PENDING'::text,
    provider_reference text,
    provider_response jsonb,
    failure_reason text,
    points_earned integer DEFAULT 0,
    draw_entries integer DEFAULT 0,
    spin_eligible boolean DEFAULT false,
    customer_email text,
    customer_name text,
    ip_address inet,
    user_agent text,
    affiliate_code text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    completed_at timestamp with time zone,
    transaction_code character varying(30) NOT NULL,
    CONSTRAINT positive_amount CHECK (((amount)::numeric > (0)::numeric)),
    CONSTRAINT transactions_payment_method_check CHECK ((payment_method = ANY (ARRAY['CARD'::text, 'BANK_TRANSFER'::text, 'USSD'::text, 'WALLET'::text]))),
    CONSTRAINT transactions_recharge_type_check CHECK ((recharge_type = ANY (ARRAY['AIRTIME'::text, 'DATA'::text]))),
    CONSTRAINT transactions_status_check CHECK ((status = ANY (ARRAY['PENDING'::text, 'PROCESSING'::text, 'SUCCESS'::text, 'FAILED'::text, 'CANCELLED'::text]))),
    CONSTRAINT valid_msisdn CHECK ((msisdn ~ '^234[789][01][0-9]{8}$'::text))
);

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_payment_reference_key UNIQUE (payment_reference);

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);

CREATE INDEX idx_transactions_created_at ON public.transactions USING btree (created_at DESC);

CREATE INDEX idx_transactions_msisdn ON public.transactions USING btree (msisdn);

CREATE INDEX idx_transactions_msisdn2 ON public.transactions USING btree (msisdn);

CREATE INDEX idx_transactions_network_provider ON public.transactions USING btree (network_provider);

CREATE INDEX idx_transactions_payment_reference ON public.transactions USING btree (payment_reference);

CREATE INDEX idx_transactions_status ON public.transactions USING btree (status);

CREATE UNIQUE INDEX idx_transactions_transaction_code ON public.transactions USING btree (transaction_code);

CREATE INDEX idx_transactions_user_id ON public.transactions USING btree (user_id);

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON public.transactions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_data_plan_id_fkey FOREIGN KEY (data_plan_id) REFERENCES public.data_plans(id);

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE public.transactions ENABLE ROW LEVEL SECURITY;
