-- ============================================================
-- Table: wallet_transactions
-- ============================================================

CREATE TABLE public.wallet_transactions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    wallet_id uuid NOT NULL,
    msisdn character varying(20) NOT NULL,
    type character varying(30) NOT NULL,
    amount bigint NOT NULL,
    balance_before bigint NOT NULL,
    balance_after bigint NOT NULL,
    reference character varying(100) NOT NULL,
    description text,
    status character varying(20) DEFAULT 'completed'::character varying NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT wallet_transactions_amount_check CHECK ((amount > 0)),
    CONSTRAINT wallet_transactions_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'completed'::character varying, 'failed'::character varying, 'reversed'::character varying])::text[]))),
    CONSTRAINT wallet_transactions_type_check CHECK (((type)::text = ANY ((ARRAY['credit'::character varying, 'debit'::character varying, 'hold'::character varying, 'release'::character varying, 'adjustment'::character varying])::text[])))
);

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_reference_key UNIQUE (reference);

CREATE INDEX idx_wallet_transactions_msisdn ON public.wallet_transactions USING btree (msisdn);

CREATE INDEX idx_wallet_transactions_reference ON public.wallet_transactions USING btree (reference);

CREATE INDEX idx_wallet_transactions_wallet_id ON public.wallet_transactions USING btree (wallet_id);

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_wallet_id_fkey FOREIGN KEY (wallet_id) REFERENCES public.wallets(id) ON DELETE CASCADE;
