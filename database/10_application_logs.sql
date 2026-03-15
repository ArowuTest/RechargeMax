-- ============================================================
-- Table: application_logs
-- ============================================================

CREATE TABLE public.application_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    level text NOT NULL,
    message text NOT NULL,
    context jsonb,
    user_id uuid,
    ip_address inet,
    user_agent text,
    request_id text,
    error_code text,
    stack_trace text,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT application_logs_level_check CHECK ((level = ANY (ARRAY['DEBUG'::text, 'INFO'::text, 'WARN'::text, 'ERROR'::text, 'FATAL'::text])))
);

ALTER TABLE ONLY public.application_logs
    ADD CONSTRAINT application_logs_pkey PRIMARY KEY (id);

CREATE INDEX idx_application_logs_created_at ON public.application_logs USING btree (created_at DESC);

CREATE INDEX idx_application_logs_level ON public.application_logs USING btree (level);

CREATE INDEX idx_application_logs_user_id ON public.application_logs USING btree (user_id);

ALTER TABLE ONLY public.application_logs
    ADD CONSTRAINT application_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);
