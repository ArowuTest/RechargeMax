-- ============================================================
-- Table: affiliate_payouts
-- ============================================================

CREATE TABLE public.affiliate_payouts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    affiliate_id uuid,
    payout_batch_id uuid DEFAULT public.uuid_generate_v4(),
    total_amount numeric(12,2) NOT NULL,
    commission_count integer DEFAULT 0 NOT NULL,
    commission_ids jsonb DEFAULT '[]'::jsonb,
    payout_method text DEFAULT 'BANK_TRANSFER'::text,
    bank_name text,
    account_number text,
    account_name text,
    payout_status text DEFAULT 'PENDING'::text,
    payout_reference text,
    payout_fee numeric(12,2) DEFAULT 0.00,
    net_amount numeric(12,2) NOT NULL,
    processed_at timestamp with time zone,
    processed_by uuid,
    failure_reason text,
    notes text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT affiliate_payouts_payout_method_check CHECK ((payout_method = ANY (ARRAY['BANK_TRANSFER'::text, 'MOBILE_MONEY'::text, 'WALLET'::text]))),
    CONSTRAINT affiliate_payouts_payout_status_check CHECK ((payout_status = ANY (ARRAY['PENDING'::text, 'PROCESSING'::text, 'COMPLETED'::text, 'FAILED'::text, 'CANCELLED'::text]))),
    CONSTRAINT affiliate_payouts_total_amount_check CHECK ((total_amount > (0)::numeric))
);

ALTER TABLE ONLY public.affiliate_payouts
    ADD CONSTRAINT affiliate_payouts_pkey PRIMARY KEY (id);

CREATE INDEX idx_affiliate_payouts_affiliate_id ON public.affiliate_payouts USING btree (affiliate_id);

CREATE INDEX idx_affiliate_payouts_batch_id ON public.affiliate_payouts USING btree (payout_batch_id);

CREATE INDEX idx_affiliate_payouts_created_at ON public.affiliate_payouts USING btree (created_at);

CREATE INDEX idx_affiliate_payouts_pending ON public.affiliate_payouts USING btree (affiliate_id, payout_status, created_at) WHERE (payout_status = 'PENDING'::text);

CREATE INDEX idx_affiliate_payouts_processed_at ON public.affiliate_payouts USING btree (processed_at);

CREATE INDEX idx_affiliate_payouts_reference ON public.affiliate_payouts USING btree (payout_reference);

CREATE INDEX idx_affiliate_payouts_status ON public.affiliate_payouts USING btree (payout_status);

CREATE TRIGGER trigger_update_affiliate_payout_timestamp BEFORE UPDATE ON public.affiliate_payouts FOR EACH ROW EXECUTE FUNCTION public.update_affiliate_payout_timestamp();

ALTER TABLE ONLY public.affiliate_payouts
    ADD CONSTRAINT affiliate_payouts_affiliate_id_fkey FOREIGN KEY (affiliate_id) REFERENCES public.affiliates(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.affiliate_payouts
    ADD CONSTRAINT affiliate_payouts_processed_by_fkey FOREIGN KEY (processed_by) REFERENCES public.admin_users(id);
