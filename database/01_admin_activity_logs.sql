-- ============================================================
-- Table: admin_activity_logs
-- ============================================================

CREATE TABLE public.admin_activity_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    admin_user_id uuid,
    admin_session_id uuid,
    action text NOT NULL,
    resource text,
    resource_id text,
    method text,
    endpoint text,
    request_data jsonb,
    response_status integer,
    response_data jsonb,
    ip_address inet,
    user_agent text,
    duration_ms integer,
    is_suspicious boolean DEFAULT false,
    risk_score integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    admin_email character varying(255),
    action_type character varying(50),
    resource_type character varying(50),
    details jsonb,
    CONSTRAINT admin_activity_logs_risk_score_check CHECK (((risk_score >= 0) AND (risk_score <= 100)))
);

ALTER TABLE ONLY public.admin_activity_logs
    ADD CONSTRAINT admin_activity_logs_pkey PRIMARY KEY (id);

CREATE INDEX idx_admin_activity_logs_action ON public.admin_activity_logs USING btree (action);

CREATE INDEX idx_admin_activity_logs_action_type ON public.admin_activity_logs USING btree (action);

CREATE INDEX idx_admin_activity_logs_admin_user_id ON public.admin_activity_logs USING btree (admin_user_id);

CREATE INDEX idx_admin_activity_logs_created_at ON public.admin_activity_logs USING btree (created_at);

CREATE INDEX idx_admin_activity_logs_is_suspicious ON public.admin_activity_logs USING btree (is_suspicious) WHERE (is_suspicious = true);

CREATE INDEX idx_admin_activity_logs_resource ON public.admin_activity_logs USING btree (resource);

CREATE INDEX idx_admin_activity_logs_resource_id ON public.admin_activity_logs USING btree (resource_id);

CREATE INDEX idx_admin_activity_logs_risk_score ON public.admin_activity_logs USING btree (risk_score) WHERE (risk_score > 50);

CREATE INDEX idx_admin_activity_logs_security ON public.admin_activity_logs USING btree (admin_user_id, created_at, is_suspicious);

CREATE INDEX idx_admin_activity_logs_session_id ON public.admin_activity_logs USING btree (admin_session_id);

ALTER TABLE ONLY public.admin_activity_logs
    ADD CONSTRAINT admin_activity_logs_admin_session_id_fkey FOREIGN KEY (admin_session_id) REFERENCES public.admin_sessions(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.admin_activity_logs
    ADD CONSTRAINT admin_activity_logs_admin_user_id_fkey FOREIGN KEY (admin_user_id) REFERENCES public.admin_users(id) ON DELETE SET NULL;
