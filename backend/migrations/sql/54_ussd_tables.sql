-- ============================================================
-- Table: ussd_recharges
-- Recharges initiated directly through telecom USSD (not via app).
-- Points awarded when webhook received from telecom provider.
-- ============================================================
CREATE TABLE IF NOT EXISTS public.ussd_recharges (
    id              UUID          NOT NULL DEFAULT uuid_generate_v4(),
    user_id         UUID,
    msisdn          VARCHAR(20)   NOT NULL,
    network         VARCHAR(20)   NOT NULL,
    amount          BIGINT        NOT NULL,   -- kobo
    recharge_type   VARCHAR(20)   NOT NULL,   -- airtime, data
    product_code    VARCHAR(100),
    transaction_ref VARCHAR(255)  NOT NULL,
    provider_ref    VARCHAR(255),
    points_earned   INTEGER       NOT NULL DEFAULT 0,
    status          VARCHAR(20)   NOT NULL DEFAULT 'completed',
    recharge_date   TIMESTAMP     NOT NULL,
    received_at     TIMESTAMP     NOT NULL,
    processed_at    TIMESTAMP,
    webhook_payload TEXT,
    notes           TEXT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    CONSTRAINT ussd_recharges_pkey     PRIMARY KEY (id),
    CONSTRAINT ussd_recharges_ref_uniq UNIQUE (transaction_ref)
);

CREATE INDEX IF NOT EXISTS idx_ussd_recharges_msisdn   ON public.ussd_recharges (msisdn);
CREATE INDEX IF NOT EXISTS idx_ussd_recharges_user     ON public.ussd_recharges (user_id);
CREATE INDEX IF NOT EXISTS idx_ussd_recharges_date     ON public.ussd_recharges (recharge_date);
CREATE INDEX IF NOT EXISTS idx_ussd_recharges_status   ON public.ussd_recharges (status);

-- ============================================================
-- Table: ussd_webhook_logs
-- Raw log of every inbound USSD webhook from telecom providers.
-- ============================================================
CREATE TABLE IF NOT EXISTS public.ussd_webhook_logs (
    id                UUID          NOT NULL DEFAULT uuid_generate_v4(),
    provider          VARCHAR(20)   NOT NULL,
    endpoint          VARCHAR(255)  NOT NULL,
    method            VARCHAR(10)   NOT NULL,
    headers           TEXT,
    body              TEXT,
    ip_address        VARCHAR(50),
    status            VARCHAR(20)   NOT NULL DEFAULT 'received',
    processing_error  TEXT,
    ussd_recharge_id  UUID,
    received_at       TIMESTAMP     NOT NULL,
    processed_at      TIMESTAMP,
    created_at        TIMESTAMP     NOT NULL DEFAULT NOW(),
    CONSTRAINT ussd_webhook_logs_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_ussd_webhook_logs_provider   ON public.ussd_webhook_logs (provider);
CREATE INDEX IF NOT EXISTS idx_ussd_webhook_logs_received   ON public.ussd_webhook_logs (received_at);
CREATE INDEX IF NOT EXISTS idx_ussd_webhook_logs_recharge   ON public.ussd_webhook_logs (ussd_recharge_id);
