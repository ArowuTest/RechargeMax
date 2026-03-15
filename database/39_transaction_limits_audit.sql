-- ============================================================
-- Table: transaction_limits_audit
-- ============================================================

CREATE TABLE public.transaction_limits_audit (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    limit_id uuid NOT NULL,
    action character varying(20) NOT NULL,
    old_values jsonb,
    new_values jsonb,
    changed_by uuid,
    changed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ip_address character varying(45),
    user_agent text,
    reason text
);

ALTER TABLE ONLY public.transaction_limits_audit
    ADD CONSTRAINT transaction_limits_audit_pkey PRIMARY KEY (id);

CREATE INDEX idx_limits_audit_changed_at ON public.transaction_limits_audit USING btree (changed_at DESC);

CREATE INDEX idx_limits_audit_changed_by ON public.transaction_limits_audit USING btree (changed_by);

CREATE INDEX idx_limits_audit_limit_id ON public.transaction_limits_audit USING btree (limit_id);

ALTER TABLE ONLY public.transaction_limits_audit
    ADD CONSTRAINT transaction_limits_audit_limit_id_fkey FOREIGN KEY (limit_id) REFERENCES public.transaction_limits(id) ON DELETE CASCADE;
