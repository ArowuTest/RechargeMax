-- ============================================================
-- Table: wheel_prizes
-- ============================================================

CREATE TABLE public.wheel_prizes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    prize_name text NOT NULL,
    prize_type text NOT NULL,
    prize_value bigint NOT NULL,
    probability numeric(5,2) NOT NULL,
    minimum_recharge numeric(10,2) DEFAULT 0,
    is_active boolean DEFAULT true,
    icon_name text,
    color_scheme text,
    sort_order integer DEFAULT 0,
    description text,
    terms_and_conditions text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT positive_prize_value CHECK (((prize_value)::numeric > (0)::numeric)),
    CONSTRAINT valid_probability CHECK (((probability >= (0)::numeric) AND (probability <= (100)::numeric))),
    CONSTRAINT wheel_prizes_prize_type_check CHECK ((prize_type = ANY (ARRAY['CASH'::text, 'AIRTIME'::text, 'DATA'::text, 'POINTS'::text])))
);

ALTER TABLE ONLY public.wheel_prizes
    ADD CONSTRAINT wheel_prizes_pkey PRIMARY KEY (id);

CREATE INDEX idx_wheel_prizes_is_active ON public.wheel_prizes USING btree (is_active);

CREATE INDEX idx_wheel_prizes_prize_type ON public.wheel_prizes USING btree (prize_type);

CREATE INDEX idx_wheel_prizes_sort_order ON public.wheel_prizes USING btree (sort_order);

CREATE TRIGGER update_wheel_prizes_updated_at BEFORE UPDATE ON public.wheel_prizes FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();
