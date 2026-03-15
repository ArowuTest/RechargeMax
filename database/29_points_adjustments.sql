-- ============================================================
-- Table: points_adjustments
-- ============================================================

CREATE TABLE public.points_adjustments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    points integer NOT NULL,
    reason character varying(255) NOT NULL,
    description text,
    created_by uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.points_adjustments
    ADD CONSTRAINT points_adjustments_pkey PRIMARY KEY (id);

CREATE INDEX idx_points_adjustments_created_at ON public.points_adjustments USING btree (created_at DESC);

CREATE INDEX idx_points_adjustments_created_by ON public.points_adjustments USING btree (created_by);

CREATE INDEX idx_points_adjustments_user_id ON public.points_adjustments USING btree (user_id);

ALTER TABLE ONLY public.points_adjustments
    ADD CONSTRAINT fk_points_adjustments_admin FOREIGN KEY (created_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.points_adjustments
    ADD CONSTRAINT fk_points_adjustments_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
