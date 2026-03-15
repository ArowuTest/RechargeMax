-- ============================================================
-- Table: otps
-- ============================================================

CREATE TABLE public.otps (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    msisdn character varying(20) NOT NULL,
    code character varying(6) NOT NULL,
    purpose character varying(50) NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    is_used boolean DEFAULT false,
    used_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

ALTER TABLE ONLY public.otps
    ADD CONSTRAINT otps_pkey PRIMARY KEY (id);

CREATE INDEX idx_otps_code ON public.otps USING btree (code);

CREATE INDEX idx_otps_expires_at ON public.otps USING btree (expires_at);

CREATE INDEX idx_otps_is_used ON public.otps USING btree (is_used);

CREATE INDEX idx_otps_msisdn ON public.otps USING btree (msisdn);

CREATE TRIGGER trigger_otps_updated_at BEFORE UPDATE ON public.otps FOR EACH ROW EXECUTE FUNCTION public.update_otps_updated_at();
