-- ============================================================
-- Table: provider_configs
-- ============================================================

-- Create sequence first so the ALTER COLUMN SET DEFAULT succeeds
CREATE SEQUENCE IF NOT EXISTS public.provider_configs_id_seq
    AS bigint START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;

CREATE TABLE IF NOT EXISTS public.provider_configs (
    id bigint NOT NULL DEFAULT nextval('public.provider_configs_id_seq'::regclass),
    network character varying(50) NOT NULL,
    service_type character varying(50) NOT NULL,
    provider_mode character varying(50) NOT NULL,
    provider_name character varying(100) NOT NULL,
    priority integer DEFAULT 1,
    config jsonb DEFAULT '{}'::jsonb NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

ALTER TABLE ONLY public.provider_configs
    ADD CONSTRAINT provider_configs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.provider_configs
    ADD CONSTRAINT unique_active_provider UNIQUE (network, service_type, priority, is_active);

CREATE INDEX idx_provider_configs_lookup ON public.provider_configs USING btree (network, service_type, is_active, priority);

CREATE TRIGGER update_provider_configs_updated_at BEFORE UPDATE ON public.provider_configs FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
