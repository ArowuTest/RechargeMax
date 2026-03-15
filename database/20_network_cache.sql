-- ============================================================
-- Table: network_cache
-- ============================================================

CREATE TABLE public.network_cache (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    msisdn text NOT NULL,
    network text NOT NULL,
    last_verified_at timestamp with time zone DEFAULT now() NOT NULL,
    cache_expires_at timestamp with time zone NOT NULL,
    lookup_source text,
    hlr_provider text,
    hlr_response jsonb,
    is_valid boolean DEFAULT true,
    invalidated_at timestamp with time zone,
    invalidation_reason text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

ALTER TABLE ONLY public.network_cache
    ADD CONSTRAINT network_cache_msisdn_key UNIQUE (msisdn);

ALTER TABLE ONLY public.network_cache
    ADD CONSTRAINT network_cache_pkey PRIMARY KEY (id);

CREATE INDEX idx_network_cache_expires ON public.network_cache USING btree (cache_expires_at);

CREATE INDEX idx_network_cache_msisdn ON public.network_cache USING btree (msisdn);

CREATE INDEX idx_network_cache_valid ON public.network_cache USING btree (is_valid);
