-- ============================================================
-- Table: affiliate_bank_accounts
-- ============================================================

CREATE TABLE public.affiliate_bank_accounts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    affiliate_id uuid,
    bank_name text NOT NULL,
    account_number text NOT NULL,
    account_name text NOT NULL,
    is_verified boolean DEFAULT false,
    is_primary boolean DEFAULT false,
    verified_at timestamp with time zone,
    verified_by uuid,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT valid_account_number CHECK ((length(account_number) >= 10))
);

ALTER TABLE ONLY public.affiliate_bank_accounts
    ADD CONSTRAINT affiliate_bank_accounts_pkey PRIMARY KEY (id);

CREATE INDEX idx_affiliate_bank_accounts_affiliate_id ON public.affiliate_bank_accounts USING btree (affiliate_id);

CREATE INDEX idx_affiliate_bank_accounts_is_primary ON public.affiliate_bank_accounts USING btree (is_primary) WHERE (is_primary = true);

CREATE INDEX idx_affiliate_bank_accounts_is_verified ON public.affiliate_bank_accounts USING btree (is_verified);

CREATE TRIGGER trigger_ensure_single_primary_bank_account BEFORE INSERT OR UPDATE ON public.affiliate_bank_accounts FOR EACH ROW WHEN ((new.is_primary = true)) EXECUTE FUNCTION public.ensure_single_primary_bank_account();

CREATE TRIGGER trigger_update_affiliate_bank_account_timestamp BEFORE UPDATE ON public.affiliate_bank_accounts FOR EACH ROW EXECUTE FUNCTION public.update_affiliate_bank_account_timestamp();

ALTER TABLE ONLY public.affiliate_bank_accounts
    ADD CONSTRAINT affiliate_bank_accounts_affiliate_id_fkey FOREIGN KEY (affiliate_id) REFERENCES public.affiliates(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.affiliate_bank_accounts
    ADD CONSTRAINT affiliate_bank_accounts_verified_by_fkey FOREIGN KEY (verified_by) REFERENCES public.admin_users(id);
