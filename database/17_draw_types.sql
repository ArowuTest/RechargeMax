-- ============================================================
-- Table: draw_types
-- ============================================================

CREATE TABLE public.draw_types (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(50) NOT NULL,
    description text,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.draw_types
    ADD CONSTRAINT draw_types_name_key UNIQUE (name);

ALTER TABLE ONLY public.draw_types
    ADD CONSTRAINT draw_types_pkey PRIMARY KEY (id);

CREATE INDEX idx_draw_types_active ON public.draw_types USING btree (is_active);
