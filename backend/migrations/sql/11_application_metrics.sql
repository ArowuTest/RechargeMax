-- ============================================================
-- Table: application_metrics
-- ============================================================

CREATE TABLE public.application_metrics (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    metric_name text NOT NULL,
    metric_value numeric(15,4) NOT NULL,
    metric_unit text,
    tags jsonb,
    dimensions jsonb,
    recorded_at timestamp with time zone DEFAULT now()
);

ALTER TABLE ONLY public.application_metrics
    ADD CONSTRAINT application_metrics_pkey PRIMARY KEY (id);

CREATE INDEX idx_application_metrics_metric_name ON public.application_metrics USING btree (metric_name);

CREATE INDEX idx_application_metrics_recorded_at ON public.application_metrics USING btree (recorded_at DESC);
