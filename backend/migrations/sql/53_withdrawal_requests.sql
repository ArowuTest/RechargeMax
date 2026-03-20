-- ============================================================
-- Table: withdrawal_requests
-- User requests to withdraw cash prizes to their bank account.
-- ============================================================
CREATE TABLE IF NOT EXISTS public.withdrawal_requests (
    id                      UUID          NOT NULL DEFAULT uuid_generate_v4(),
    user_id                 UUID          NOT NULL,
    bank_account_id         UUID          NOT NULL,
    amount                  BIGINT        NOT NULL,   -- kobo
    fee                     BIGINT        NOT NULL DEFAULT 0,
    net_amount              BIGINT        NOT NULL,   -- kobo
    status                  VARCHAR(30)   NOT NULL DEFAULT 'PENDING',
    approved_by_admin_id    UUID,
    rejection_reason        TEXT,
    admin_notes             TEXT,
    transaction_reference   VARCHAR(255),
    bank_reference          VARCHAR(255),
    payment_provider        VARCHAR(100),
    provider_response       JSONB,
    wallet_transaction_id   UUID,
    requested_at            TIMESTAMP     NOT NULL DEFAULT NOW(),
    approved_at             TIMESTAMP,
    processing_started_at   TIMESTAMP,
    completed_at            TIMESTAMP,
    rejected_at             TIMESTAMP,
    request_ip              VARCHAR(50),
    request_user_agent      TEXT,
    CONSTRAINT withdrawal_requests_pkey     PRIMARY KEY (id),
    CONSTRAINT withdrawal_requests_ref_uniq UNIQUE (transaction_reference)
);

CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_user    ON public.withdrawal_requests (user_id);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_status  ON public.withdrawal_requests (status);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_created ON public.withdrawal_requests (requested_at);
