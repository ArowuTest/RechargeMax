-- ============================================================
-- Table: spin_tiers
-- ============================================================

CREATE TABLE public.spin_tiers (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tier_name text NOT NULL,
    tier_display_name text NOT NULL,
    min_daily_amount bigint NOT NULL,
    max_daily_amount bigint NOT NULL,
    spins_per_day integer NOT NULL,
    tier_color text,
    tier_icon text,
    tier_badge text,
    description text,
    sort_order integer DEFAULT 0,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    CONSTRAINT positive_amounts CHECK (((min_daily_amount >= 0) AND (max_daily_amount > min_daily_amount))),
    CONSTRAINT positive_spins CHECK ((spins_per_day > 0)),
    CONSTRAINT valid_sort_order CHECK ((sort_order >= 0))
);

ALTER TABLE ONLY public.spin_tiers
    ADD CONSTRAINT spin_tiers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.spin_tiers
    ADD CONSTRAINT spin_tiers_tier_name_key UNIQUE (tier_name);

CREATE INDEX idx_spin_tiers_amount_range ON public.spin_tiers USING btree (min_daily_amount, max_daily_amount);

CREATE INDEX idx_spin_tiers_is_active ON public.spin_tiers USING btree (is_active);

CREATE INDEX idx_spin_tiers_sort_order ON public.spin_tiers USING btree (sort_order);

CREATE TRIGGER trigger_spin_tiers_updated_at BEFORE UPDATE ON public.spin_tiers FOR EACH ROW EXECUTE FUNCTION public.update_spin_tiers_updated_at();
