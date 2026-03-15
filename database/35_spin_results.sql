-- ============================================================
-- Table: spin_results
-- ============================================================

CREATE TABLE public.spin_results (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    transaction_id uuid,
    msisdn text NOT NULL,
    prize_id uuid,
    prize_name text NOT NULL,
    prize_type text NOT NULL,
    prize_value bigint NOT NULL,
    claim_status text DEFAULT 'PENDING'::text,
    claimed_at timestamp with time zone,
    claim_reference text,
    created_at timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone DEFAULT (now() + '30 days'::interval),
    reviewed_by uuid,
    reviewed_at timestamp without time zone,
    rejection_reason text,
    admin_notes text,
    payment_reference character varying(100),
    bank_account_number text,
    bank_account_name text,
    bank_name text,
    spin_code character varying(30),
    fulfillment_mode character varying(20) DEFAULT 'AUTO'::character varying,
    fulfillment_attempts integer DEFAULT 0,
    last_fulfillment_attempt timestamp without time zone,
    fulfillment_error text,
    can_retry boolean DEFAULT true,
    provision_started_at timestamp without time zone,
    provision_completed_at timestamp without time zone,
    CONSTRAINT check_fulfillment_mode CHECK (((fulfillment_mode)::text = ANY ((ARRAY['AUTO'::character varying, 'MANUAL'::character varying])::text[]))),
    CONSTRAINT chk_spin_results_claim_status CHECK ((claim_status = ANY (ARRAY['PENDING'::text, 'CLAIMED'::text, 'EXPIRED'::text, 'PENDING_ADMIN_REVIEW'::text, 'APPROVED'::text, 'REJECTED'::text]))),
    CONSTRAINT positive_prize_value CHECK (((prize_value)::numeric >= (0)::numeric))
);

ALTER TABLE ONLY public.spin_results
    ADD CONSTRAINT spin_results_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.spin_results
    ADD CONSTRAINT spin_results_spin_code_key UNIQUE (spin_code);

CREATE INDEX idx_spin_results_can_retry ON public.spin_results USING btree (can_retry) WHERE (can_retry = true);

CREATE INDEX idx_spin_results_claim_status ON public.spin_results USING btree (claim_status);

CREATE INDEX idx_spin_results_created_at ON public.spin_results USING btree (created_at DESC);

CREATE INDEX idx_spin_results_created_at2 ON public.spin_results USING btree (created_at DESC);

CREATE INDEX idx_spin_results_fulfillment_mode ON public.spin_results USING btree (fulfillment_mode);

CREATE INDEX idx_spin_results_msisdn ON public.spin_results USING btree (msisdn);

CREATE INDEX idx_spin_results_msisdn2 ON public.spin_results USING btree (msisdn);

CREATE INDEX idx_spin_results_msisdn_claim_status ON public.spin_results USING btree (msisdn, claim_status);

CREATE INDEX idx_spin_results_prize_type ON public.spin_results USING btree (prize_type);

CREATE INDEX idx_spin_results_reviewed_at ON public.spin_results USING btree (reviewed_at DESC);

CREATE INDEX idx_spin_results_reviewed_by ON public.spin_results USING btree (reviewed_by);

CREATE UNIQUE INDEX idx_spin_results_spin_code ON public.spin_results USING btree (spin_code);

CREATE INDEX idx_spin_results_transaction_id ON public.spin_results USING btree (transaction_id);

CREATE INDEX idx_spin_results_user_id ON public.spin_results USING btree (user_id);

ALTER TABLE ONLY public.spin_results
    ADD CONSTRAINT spin_results_prize_id_fkey FOREIGN KEY (prize_id) REFERENCES public.wheel_prizes(id);

ALTER TABLE ONLY public.spin_results
    ADD CONSTRAINT spin_results_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.admin_users(id);

ALTER TABLE ONLY public.spin_results
    ADD CONSTRAINT spin_results_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id);

ALTER TABLE ONLY public.spin_results
    ADD CONSTRAINT spin_results_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE public.spin_results ENABLE ROW LEVEL SECURITY;
