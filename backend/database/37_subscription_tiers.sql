-- ============================================================
-- Table: subscription_tiers
-- ============================================================

CREATE TABLE public.subscription_tiers (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    description text,
    entries integer DEFAULT 1 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT subscription_tiers_entries_check CHECK ((entries > 0))
);

ALTER TABLE ONLY public.subscription_tiers
    ADD CONSTRAINT subscription_tiers_name_key UNIQUE (name);

ALTER TABLE ONLY public.subscription_tiers
    ADD CONSTRAINT subscription_tiers_pkey PRIMARY KEY (id);

CREATE INDEX idx_subscription_tiers_is_active ON public.subscription_tiers USING btree (is_active);
