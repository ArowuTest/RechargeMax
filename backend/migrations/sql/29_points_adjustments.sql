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

-- NOTE: FK fk_points_adjustments_admin intentionally removed here.
-- created_by is populated from admin_users.id (a separate table from users).
-- Adding a FK to users.id causes a 23503 violation because admin UUIDs
-- don't exist in the users table. Migration 050 drops this wrong FK and
-- makes created_by nullable. See migration 050 for the correct constraint state.

ALTER TABLE ONLY public.points_adjustments
    ADD CONSTRAINT fk_points_adjustments_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
