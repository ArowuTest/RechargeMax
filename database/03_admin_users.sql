-- ============================================================
-- Table: admin_users
-- ============================================================

CREATE TABLE public.admin_users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    full_name text NOT NULL,
    role text DEFAULT 'ADMIN'::text,
    permissions jsonb DEFAULT '[]'::jsonb,
    is_active boolean DEFAULT true,
    last_login_at timestamp with time zone,
    login_attempts integer DEFAULT 0,
    locked_until timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT admin_users_role_check CHECK ((role = ANY (ARRAY['SUPER_ADMIN'::text, 'ADMIN'::text, 'MODERATOR'::text, 'SUPPORT'::text]))),
    CONSTRAINT valid_admin_email CHECK ((email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'::text))
);

ALTER TABLE ONLY public.admin_users
    ADD CONSTRAINT admin_users_email_key UNIQUE (email);

ALTER TABLE ONLY public.admin_users
    ADD CONSTRAINT admin_users_pkey PRIMARY KEY (id);

CREATE INDEX idx_admin_users_email ON public.admin_users USING btree (email);

CREATE INDEX idx_admin_users_is_active ON public.admin_users USING btree (is_active);

CREATE INDEX idx_admin_users_role ON public.admin_users USING btree (role);

CREATE TRIGGER update_admin_users_updated_at BEFORE UPDATE ON public.admin_users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

ALTER TABLE public.admin_users ENABLE ROW LEVEL SECURITY;
