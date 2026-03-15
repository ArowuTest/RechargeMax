-- ============================================================
-- Table: admin_sessions
-- ============================================================

CREATE TABLE public.admin_sessions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    admin_user_id uuid,
    session_token text NOT NULL,
    ip_address inet,
    user_agent text,
    is_active boolean DEFAULT true,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    last_accessed_at timestamp with time zone DEFAULT now()
);

ALTER TABLE ONLY public.admin_sessions
    ADD CONSTRAINT admin_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.admin_sessions
    ADD CONSTRAINT admin_sessions_session_token_key UNIQUE (session_token);

CREATE INDEX idx_admin_sessions_admin_user_id ON public.admin_sessions USING btree (admin_user_id);

CREATE INDEX idx_admin_sessions_expires_at ON public.admin_sessions USING btree (expires_at);

CREATE INDEX idx_admin_sessions_session_token ON public.admin_sessions USING btree (session_token);

ALTER TABLE ONLY public.admin_sessions
    ADD CONSTRAINT admin_sessions_admin_user_id_fkey FOREIGN KEY (admin_user_id) REFERENCES public.admin_users(id) ON DELETE CASCADE;
