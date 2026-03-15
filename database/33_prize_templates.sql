-- ============================================================
-- Table: prize_templates
-- ============================================================

CREATE TABLE public.prize_templates (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(100) NOT NULL,
    draw_type_id uuid,
    description text,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.prize_templates
    ADD CONSTRAINT prize_templates_pkey PRIMARY KEY (id);

CREATE INDEX idx_prize_templates_active ON public.prize_templates USING btree (is_active);

CREATE INDEX idx_prize_templates_draw_type ON public.prize_templates USING btree (draw_type_id);

ALTER TABLE ONLY public.prize_templates
    ADD CONSTRAINT prize_templates_draw_type_id_fkey FOREIGN KEY (draw_type_id) REFERENCES public.draw_types(id) ON DELETE CASCADE;
