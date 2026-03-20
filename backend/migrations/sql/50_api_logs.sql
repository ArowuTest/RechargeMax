-- ============================================================
-- Table: api_logs
-- Outbound API call log (VTPass, Paystack, HLR, etc.)
-- ============================================================
CREATE TABLE IF NOT EXISTS public.api_logs (
    id                      UUID          NOT NULL DEFAULT uuid_generate_v4(),
    service_name            VARCHAR(100)  NOT NULL,
    endpoint                VARCHAR(500)  NOT NULL,
    method                  VARCHAR(10)   NOT NULL,
    request_url             TEXT,
    request_headers         JSONB,
    request_payload         JSONB,
    response_status_code    INTEGER,
    response_headers        JSONB,
    response_payload        JSONB,
    response_time_ms        INTEGER,
    is_error                BOOLEAN       DEFAULT FALSE,
    error_message           TEXT,
    error_code              VARCHAR(50),
    user_id                 UUID,
    transaction_reference   VARCHAR(255),
    ip_address              VARCHAR(50),
    created_at              TIMESTAMP     NOT NULL DEFAULT NOW(),
    metadata                JSONB,
    CONSTRAINT api_logs_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_api_logs_service    ON public.api_logs (service_name);
CREATE INDEX IF NOT EXISTS idx_api_logs_created    ON public.api_logs (created_at);
CREATE INDEX IF NOT EXISTS idx_api_logs_is_error   ON public.api_logs (is_error);
CREATE INDEX IF NOT EXISTS idx_api_logs_user       ON public.api_logs (user_id);
