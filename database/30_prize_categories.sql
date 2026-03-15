-- ============================================================
-- Table: prize_categories
-- ============================================================

CREATE TABLE public.prize_categories (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    template_id uuid,
    draw_id uuid,
    category_name character varying(100) NOT NULL,
    prize_amount numeric(15,2) NOT NULL,
    winners_count integer DEFAULT 1 NOT NULL,
    runner_ups_count integer DEFAULT 1 NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_parent CHECK (((template_id IS NOT NULL) OR (draw_id IS NOT NULL)))
);

ALTER TABLE ONLY public.prize_categories
    ADD CONSTRAINT prize_categories_pkey PRIMARY KEY (id);

CREATE INDEX idx_prize_categories_draw ON public.prize_categories USING btree (draw_id);

CREATE INDEX idx_prize_categories_order ON public.prize_categories USING btree (display_order);

CREATE INDEX idx_prize_categories_template ON public.prize_categories USING btree (template_id);

ALTER TABLE ONLY public.prize_categories
    ADD CONSTRAINT prize_categories_draw_id_fkey FOREIGN KEY (draw_id) REFERENCES public.draws(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.prize_categories
    ADD CONSTRAINT prize_categories_template_id_fkey FOREIGN KEY (template_id) REFERENCES public.prize_templates(id) ON DELETE CASCADE;
