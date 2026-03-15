-- ============================================================
-- Table: draws
-- ============================================================

CREATE TABLE public.draws (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    description text,
    status text DEFAULT 'UPCOMING'::text,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone NOT NULL,
    draw_time timestamp with time zone,
    prize_pool numeric(12,2) NOT NULL,
    winners_count integer DEFAULT 1,
    total_entries integer DEFAULT 0,
    results jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    completed_at timestamp with time zone,
    draw_type_id uuid,
    runner_ups_count integer DEFAULT 1,
    draw_code character varying(20),
    prize_template_id uuid,
    CONSTRAINT draws_status_check CHECK ((status = ANY (ARRAY['UPCOMING'::text, 'ACTIVE'::text, 'COMPLETED'::text, 'CANCELLED'::text]))),
    CONSTRAINT draws_type_check CHECK ((type = ANY (ARRAY['DAILY'::text, 'WEEKLY'::text, 'MONTHLY'::text, 'SPECIAL'::text]))),
    CONSTRAINT positive_prize_pool CHECK ((prize_pool > (0)::numeric)),
    CONSTRAINT positive_winners_count CHECK ((winners_count > 0)),
    CONSTRAINT valid_timing CHECK ((end_time > start_time))
);

ALTER TABLE ONLY public.draws
    ADD CONSTRAINT draws_draw_code_key UNIQUE (draw_code);

ALTER TABLE ONLY public.draws
    ADD CONSTRAINT draws_pkey PRIMARY KEY (id);

CREATE INDEX idx_draws_draw_code ON public.draws USING btree (draw_code);

CREATE INDEX idx_draws_end_time ON public.draws USING btree (end_time);

CREATE INDEX idx_draws_start_time ON public.draws USING btree (start_time);

CREATE INDEX idx_draws_status ON public.draws USING btree (status);

CREATE INDEX idx_draws_type ON public.draws USING btree (type);

CREATE TRIGGER update_draws_updated_at BEFORE UPDATE ON public.draws FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

ALTER TABLE ONLY public.draws
    ADD CONSTRAINT draws_draw_type_id_fkey FOREIGN KEY (draw_type_id) REFERENCES public.draw_types(id) ON DELETE SET NULL NOT VALID;

ALTER TABLE ONLY public.draws
    ADD CONSTRAINT draws_prize_template_id_fkey FOREIGN KEY (prize_template_id) REFERENCES public.prize_templates(id) ON DELETE SET NULL NOT VALID;

ALTER TABLE public.draws ENABLE ROW LEVEL SECURITY;
