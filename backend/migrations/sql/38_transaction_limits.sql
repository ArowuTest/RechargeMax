-- ============================================================
-- Table: transaction_limits
-- ============================================================

CREATE TABLE public.transaction_limits (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    limit_type character varying(50) NOT NULL,
    limit_scope character varying(50) NOT NULL,
    min_amount bigint DEFAULT 10000 NOT NULL,
    max_amount bigint DEFAULT 10000000 NOT NULL,
    daily_limit bigint,
    monthly_limit bigint,
    is_active boolean DEFAULT true NOT NULL,
    applies_to_user_tier character varying(50),
    description text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by uuid,
    updated_by uuid,
    CONSTRAINT positive_amounts CHECK (((min_amount > 0) AND (max_amount > 0))),
    CONSTRAINT valid_amount_range CHECK ((min_amount <= max_amount))
);

ALTER TABLE ONLY public.transaction_limits
    ADD CONSTRAINT transaction_limits_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.transaction_limits
    ADD CONSTRAINT unique_limit_config UNIQUE (limit_type, limit_scope, applies_to_user_tier);

CREATE INDEX idx_transaction_limits_active ON public.transaction_limits USING btree (is_active) WHERE (is_active = true);

CREATE INDEX idx_transaction_limits_tier ON public.transaction_limits USING btree (applies_to_user_tier);

CREATE INDEX idx_transaction_limits_type_scope ON public.transaction_limits USING btree (limit_type, limit_scope);

CREATE TRIGGER transaction_limits_updated_at BEFORE UPDATE ON public.transaction_limits FOR EACH ROW EXECUTE FUNCTION public.update_transaction_limits_timestamp();
