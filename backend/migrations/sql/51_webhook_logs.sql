-- ============================================================
-- Table: webhook_logs
-- Inbound webhook log (Paystack, VTPass callbacks, etc.)
-- ============================================================
CREATE TABLE IF NOT EXISTS public.webhook_logs (
    id                      UUID          NOT NULL DEFAULT uuid_generate_v4(),
    source                  VARCHAR(100)  NOT NULL,
    event_type              VARCHAR(100)  NOT NULL,
    payload                 JSONB         NOT NULL,
    headers                 JSONB,
    signature               TEXT,
    is_verified             BOOLEAN       DEFAULT FALSE,
    verification_method     VARCHAR(100),
    verification_error      TEXT,
    is_processed            BOOLEAN       DEFAULT FALSE,
    processing_error        TEXT,
    processing_attempts     INTEGER       DEFAULT 0,
    max_processing_attempts INTEGER       DEFAULT 3,
    transaction_reference   VARCHAR(255),
    related_transaction_id  UUID,
    received_at             TIMESTAMP     NOT NULL DEFAULT NOW(),
    verified_at             TIMESTAMP,
    processed_at            TIMESTAMP,
    next_retry_at           TIMESTAMP,
    ip_address              VARCHAR(50),
    user_agent              TEXT,
    metadata                JSONB,
    CONSTRAINT webhook_logs_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_webhook_logs_source      ON public.webhook_logs (source);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_event       ON public.webhook_logs (event_type);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_processed   ON public.webhook_logs (is_processed);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_received    ON public.webhook_logs (received_at);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_tx_ref      ON public.webhook_logs (transaction_reference);
