-- ============================================================
-- Table: data_plans
-- ============================================================

CREATE TABLE public.data_plans (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    network_id uuid,
    plan_name text NOT NULL,
    data_amount text NOT NULL,
    price numeric(10,2) NOT NULL,
    validity_days integer NOT NULL,
    plan_code text NOT NULL,
    is_active boolean DEFAULT true,
    sort_order integer DEFAULT 0,
    description text,
    terms_and_conditions text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    network_provider text,
    CONSTRAINT positive_price CHECK ((price > (0)::numeric)),
    CONSTRAINT positive_validity CHECK ((validity_days > 0))
);

ALTER TABLE ONLY public.data_plans
    ADD CONSTRAINT data_plans_network_id_plan_code_key UNIQUE (network_id, plan_code);

ALTER TABLE ONLY public.data_plans
    ADD CONSTRAINT data_plans_pkey PRIMARY KEY (id);

CREATE INDEX idx_data_plans_is_active ON public.data_plans USING btree (is_active);

CREATE INDEX idx_data_plans_network_id ON public.data_plans USING btree (network_id);

CREATE INDEX idx_data_plans_network_provider ON public.data_plans USING btree (network_provider);

CREATE INDEX idx_data_plans_price ON public.data_plans USING btree (price);

CREATE TRIGGER update_data_plans_updated_at BEFORE UPDATE ON public.data_plans FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

ALTER TABLE ONLY public.data_plans
    ADD CONSTRAINT data_plans_network_id_fkey FOREIGN KEY (network_id) REFERENCES public.network_configs(id) ON DELETE CASCADE;
