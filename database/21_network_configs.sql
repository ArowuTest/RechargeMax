-- ============================================================
-- Table: network_configs
-- ============================================================

CREATE TABLE public.network_configs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    network_name text NOT NULL,
    network_code text NOT NULL,
    is_active boolean DEFAULT true,
    airtime_enabled boolean DEFAULT true,
    data_enabled boolean DEFAULT true,
    commission_rate numeric(5,2) DEFAULT 2.50,
    minimum_amount numeric(10,2) DEFAULT 50.00,
    maximum_amount numeric(10,2) DEFAULT 50000.00,
    logo_url text,
    brand_color text,
    sort_order integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT positive_commission_rate CHECK ((commission_rate >= (0)::numeric)),
    CONSTRAINT valid_amount_range CHECK ((maximum_amount > minimum_amount))
);

ALTER TABLE ONLY public.network_configs
    ADD CONSTRAINT network_configs_network_code_key UNIQUE (network_code);

ALTER TABLE ONLY public.network_configs
    ADD CONSTRAINT network_configs_pkey PRIMARY KEY (id);

CREATE INDEX idx_network_configs_is_active ON public.network_configs USING btree (is_active);

CREATE INDEX idx_network_configs_network_code ON public.network_configs USING btree (network_code);

CREATE INDEX idx_network_configs_sort_order ON public.network_configs USING btree (sort_order);

CREATE TRIGGER update_network_configs_updated_at BEFORE UPDATE ON public.network_configs FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
