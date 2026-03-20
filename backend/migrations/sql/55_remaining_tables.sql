-- ============================================================
-- Table: devices
-- Push notification device registrations (FCM tokens).
-- ============================================================
CREATE TABLE IF NOT EXISTS public.devices (
    id                          UUID          NOT NULL DEFAULT uuid_generate_v4(),
    msisdn                      VARCHAR(20)   NOT NULL,
    device_id                   VARCHAR(255)  NOT NULL,
    fcm_token                   VARCHAR(500),
    platform                    VARCHAR(10)   NOT NULL,  -- ios, android, web
    app_version                 VARCHAR(20),
    device_model                VARCHAR(100),
    os_version                  VARCHAR(50),
    is_active                   BOOLEAN       NOT NULL DEFAULT TRUE,
    last_active                 TIMESTAMP     NOT NULL DEFAULT NOW(),
    last_notification_sent_at   TIMESTAMP,
    notification_count          INTEGER       NOT NULL DEFAULT 0,
    created_at                  TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMP     NOT NULL DEFAULT NOW(),
    CONSTRAINT devices_pkey           PRIMARY KEY (id),
    CONSTRAINT devices_device_id_uniq UNIQUE (device_id)
);

CREATE INDEX IF NOT EXISTS idx_devices_msisdn     ON public.devices (msisdn);
CREATE INDEX IF NOT EXISTS idx_devices_fcm_token  ON public.devices (fcm_token);
CREATE INDEX IF NOT EXISTS idx_devices_active     ON public.devices (is_active);

-- ============================================================
-- Table: file_uploads
-- Uploaded files (KYC docs, profile photos, etc.)
-- ============================================================
CREATE TABLE IF NOT EXISTS public.file_uploads (
    id          UUID          NOT NULL DEFAULT uuid_generate_v4(),
    file_name   VARCHAR(255)  NOT NULL,
    file_path   TEXT          NOT NULL,
    file_size   INTEGER       NOT NULL,
    mime_type   VARCHAR(100)  NOT NULL,
    bucket_name VARCHAR(100)  NOT NULL,
    file_type   VARCHAR(50)   NOT NULL,
    description TEXT,
    is_active   BOOLEAN       DEFAULT TRUE,
    is_verified BOOLEAN       DEFAULT FALSE,
    upload_ip   VARCHAR(50),
    user_agent  TEXT,
    created_at  TIMESTAMP     DEFAULT NOW(),
    updated_at  TIMESTAMP     DEFAULT NOW(),
    CONSTRAINT file_uploads_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_file_uploads_file_type ON public.file_uploads (file_type);

-- ============================================================
-- Table: webhook_events
-- Deduplicated webhook event registry (idempotency keys).
-- ============================================================
CREATE TABLE IF NOT EXISTS public.webhook_events (
    id              UUID          NOT NULL DEFAULT uuid_generate_v4(),
    event_id        VARCHAR(255)  NOT NULL,
    source          VARCHAR(100)  NOT NULL,
    event_type      VARCHAR(100)  NOT NULL,
    processed       BOOLEAN       NOT NULL DEFAULT FALSE,
    payload         JSONB,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMP,
    CONSTRAINT webhook_events_pkey          PRIMARY KEY (id),
    CONSTRAINT webhook_events_event_id_uniq UNIQUE (event_id)
);

CREATE INDEX IF NOT EXISTS idx_webhook_events_source     ON public.webhook_events (source);
CREATE INDEX IF NOT EXISTS idx_webhook_events_processed  ON public.webhook_events (processed);

-- ============================================================
-- Table: service_pricing
-- Dynamic per-service pricing configuration (e.g. ₦20/day sub).
-- ============================================================
CREATE TABLE IF NOT EXISTS public.service_pricing (
    id              UUID          NOT NULL DEFAULT uuid_generate_v4(),
    service_name    VARCHAR(100)  NOT NULL,
    price           BIGINT        NOT NULL,  -- kobo
    currency        VARCHAR(10)   NOT NULL DEFAULT 'NGN',
    is_active       BOOLEAN       NOT NULL DEFAULT TRUE,
    effective_from  TIMESTAMP     NOT NULL DEFAULT NOW(),
    effective_to    TIMESTAMP,
    description     TEXT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    CONSTRAINT service_pricing_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_service_pricing_name    ON public.service_pricing (service_name);
CREATE INDEX IF NOT EXISTS idx_service_pricing_active  ON public.service_pricing (is_active);

-- ============================================================
-- Table: subscription_pricing
-- Global entry-price configuration for the daily draw.
-- ============================================================
CREATE TABLE IF NOT EXISTS public.subscription_pricing (
    id              UUID          NOT NULL DEFAULT uuid_generate_v4(),
    price_per_entry BIGINT        NOT NULL,  -- kobo (₦20 = 2000 kobo)
    currency        VARCHAR(10)   DEFAULT 'NGN',
    is_active       BOOLEAN       DEFAULT TRUE,
    effective_from  TIMESTAMP     NOT NULL,
    effective_to    TIMESTAMP,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    CONSTRAINT subscription_pricing_pkey PRIMARY KEY (id)
);

-- ============================================================
-- Table: subscription_billings
-- Per-day billing record for the ₦20/day subscription.
-- ============================================================
CREATE TABLE IF NOT EXISTS public.subscription_billings (
    id                  UUID          NOT NULL DEFAULT uuid_generate_v4(),
    subscription_id     UUID          NOT NULL,
    msisdn              VARCHAR(20)   NOT NULL,
    billing_date        TIMESTAMP     NOT NULL,
    amount              BIGINT        NOT NULL,  -- kobo
    entries_awarded     INTEGER       NOT NULL,
    points_earned       INTEGER       NOT NULL DEFAULT 0,
    status              VARCHAR(20)   NOT NULL DEFAULT 'pending',
    payment_reference   VARCHAR(255),
    payment_method      VARCHAR(100),
    failure_reason      TEXT,
    processed_at        TIMESTAMP,
    created_at          TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP     NOT NULL DEFAULT NOW(),
    CONSTRAINT subscription_billings_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_subscription_billings_sub     ON public.subscription_billings (subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscription_billings_msisdn  ON public.subscription_billings (msisdn);
CREATE INDEX IF NOT EXISTS idx_subscription_billings_date    ON public.subscription_billings (billing_date);
CREATE INDEX IF NOT EXISTS idx_subscription_billings_status  ON public.subscription_billings (status);
