-- ============================================================
-- Table: wallets
-- ============================================================

CREATE TABLE public.wallets (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    msisdn character varying(20) NOT NULL,
    balance bigint DEFAULT 0 NOT NULL,
    pending_balance bigint DEFAULT 0 NOT NULL,
    total_earned bigint DEFAULT 0 NOT NULL,
    total_withdrawn bigint DEFAULT 0 NOT NULL,
    min_payout_amount bigint DEFAULT 100000 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    is_suspended boolean DEFAULT false NOT NULL,
    suspension_reason text,
    last_transaction_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT wallets_balance_check CHECK ((balance >= 0)),
    CONSTRAINT wallets_pending_balance_check CHECK ((pending_balance >= 0))
);

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_msisdn_key UNIQUE (msisdn);

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_pkey PRIMARY KEY (id);

CREATE INDEX idx_wallets_msisdn ON public.wallets USING btree (msisdn);
