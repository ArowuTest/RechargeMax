-- ============================================================
-- Table: vtu_transactions
-- VTPass/VTU transaction log — one row per external API call.
-- Referenced by reconciliation_job for FAILED status updates.
-- ============================================================
CREATE TABLE IF NOT EXISTS public.vtu_transactions (
    id                      UUID          NOT NULL DEFAULT uuid_generate_v4(),
    transaction_reference   VARCHAR(255)  NOT NULL,
    parent_transaction_id   UUID,
    user_id                 UUID,
    phone_number            VARCHAR(20)   NOT NULL,
    network_provider        VARCHAR(20)   NOT NULL,
    recharge_type           VARCHAR(20)   NOT NULL,
    amount                  BIGINT        NOT NULL,           -- kobo
    data_bundle             VARCHAR(100),
    data_bundle_code        VARCHAR(100),
    provider_used           VARCHAR(100),
    provider_transaction_id VARCHAR(255),
    provider_reference      VARCHAR(255),
    provider_response       JSONB,
    provider_status         VARCHAR(50),
    status                  VARCHAR(20)   NOT NULL DEFAULT 'PENDING',
    retry_count             INTEGER       DEFAULT 0,
    max_retries             INTEGER       DEFAULT 3,
    user_agent              TEXT,
    ip_address              VARCHAR(50),
    device_info             JSONB,
    created_at              TIMESTAMP     NOT NULL DEFAULT NOW(),
    processing_started_at   TIMESTAMP,
    completed_at            TIMESTAMP,
    failed_at               TIMESTAMP,
    error_message           TEXT,
    error_code              VARCHAR(50),
    last_error_at           TIMESTAMP,
    is_reconciled           BOOLEAN       DEFAULT FALSE,
    reconciled_at           TIMESTAMP,
    reconciliation_notes    TEXT,
    CONSTRAINT vtu_transactions_pkey         PRIMARY KEY (id),
    CONSTRAINT vtu_transactions_ref_unique   UNIQUE (transaction_reference)
);

CREATE INDEX IF NOT EXISTS idx_vtu_transactions_parent   ON public.vtu_transactions (parent_transaction_id);
CREATE INDEX IF NOT EXISTS idx_vtu_transactions_user     ON public.vtu_transactions (user_id);
CREATE INDEX IF NOT EXISTS idx_vtu_transactions_status   ON public.vtu_transactions (status);
CREATE INDEX IF NOT EXISTS idx_vtu_transactions_created  ON public.vtu_transactions (created_at);
