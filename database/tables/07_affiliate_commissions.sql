-- ============================================================
-- Table: affiliate_commissions
-- ============================================================

CREATE TABLE public.affiliate_commissions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    affiliate_id uuid,
    transaction_id uuid,
    commission_amount numeric(10,2) NOT NULL,
    commission_rate numeric(5,2) NOT NULL,
    transaction_amount numeric(10,2) NOT NULL,
    status text DEFAULT 'PENDING'::text,
    payout_reference text,
    payout_method text,
    created_at timestamp with time zone DEFAULT now(),
    earned_at timestamp with time zone DEFAULT now(),
    paid_at timestamp with time zone,
    CONSTRAINT affiliate_commissions_status_check CHECK ((status = ANY (ARRAY['PENDING'::text, 'APPROVED'::text, 'PAID'::text, 'CANCELLED'::text]))),
    CONSTRAINT positive_commission_amount CHECK ((commission_amount > (0)::numeric)),
    CONSTRAINT positive_transaction_amount CHECK ((transaction_amount > (0)::numeric))
);

ALTER TABLE ONLY public.affiliate_commissions
    ADD CONSTRAINT affiliate_commissions_pkey PRIMARY KEY (id);

CREATE INDEX idx_affiliate_commissions_affiliate_id ON public.affiliate_commissions USING btree (affiliate_id);

CREATE INDEX idx_affiliate_commissions_earned_at ON public.affiliate_commissions USING btree (earned_at DESC);

CREATE INDEX idx_affiliate_commissions_status ON public.affiliate_commissions USING btree (status);

CREATE INDEX idx_affiliate_commissions_transaction_id ON public.affiliate_commissions USING btree (transaction_id);

ALTER TABLE ONLY public.affiliate_commissions
    ADD CONSTRAINT affiliate_commissions_affiliate_id_fkey FOREIGN KEY (affiliate_id) REFERENCES public.affiliates(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.affiliate_commissions
    ADD CONSTRAINT affiliate_commissions_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id);

ALTER TABLE public.affiliate_commissions ENABLE ROW LEVEL SECURITY;
