-- Migration 041: Subscription recurring billing infrastructure
--
-- Changes to daily_subscriptions:
--   - Add paystack_authorization_code (stored after first successful payment — used for
--     all subsequent daily auto-charges without the user going through checkout again)
--   - Add consecutive_failures (auto-pause after 7 consecutive failures)
--   - Add paused_at
--   - Add billing_day_offset (for future per-subscription billing windows)
--   - Remove UNIQUE(user_id, subscription_date) — users can have MULTIPLE active
--     subscriptions simultaneously (e.g. 1-entry + 10-entry + 5-entry = 16/day)
--   - Remove the old status CHECK so we can use 'pending','active','cancelled','paused','expired'
--     without hitting the constraint mid-migration
--
-- New table subscription_billings:
--   - One row per daily billing attempt per subscription
--   - Tracks retry_count, next_retry_at, points_awarded (idempotency flag)
--   - Links back to daily_subscriptions via subscription_id FK

-- ── Step 1: remove the unique constraint that blocks multi-subscription ────────
ALTER TABLE public.daily_subscriptions
    DROP CONSTRAINT IF EXISTS daily_subscriptions_user_id_subscription_date_key;

-- ── Step 2: add missing columns (idempotent via IF NOT EXISTS in DO block) ──────
DO $$
BEGIN
    -- Paystack reusable authorization token for recurring charges
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'paystack_authorization_code'
    ) THEN
        ALTER TABLE public.daily_subscriptions
            ADD COLUMN paystack_authorization_code TEXT;
    END IF;

    -- Paystack customer code (needed alongside auth_code for charge API)
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'paystack_customer_code'
    ) THEN
        ALTER TABLE public.daily_subscriptions
            ADD COLUMN paystack_customer_code TEXT;
    END IF;

    -- Count of consecutive failed billing days (reset to 0 on any success)
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'consecutive_failures'
    ) THEN
        ALTER TABLE public.daily_subscriptions
            ADD COLUMN consecutive_failures INTEGER NOT NULL DEFAULT 0;
    END IF;

    -- When the subscription was auto-paused (after too many consecutive failures)
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'paused_at'
    ) THEN
        ALTER TABLE public.daily_subscriptions
            ADD COLUMN paused_at TIMESTAMP;
    END IF;

    -- Total lifetime amount successfully billed (kobo)
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'total_billed_amount'
    ) THEN
        ALTER TABLE public.daily_subscriptions
            ADD COLUMN total_billed_amount BIGINT NOT NULL DEFAULT 0;
    END IF;

    -- Total lifetime points awarded from this subscription
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'total_points_awarded'
    ) THEN
        ALTER TABLE public.daily_subscriptions
            ADD COLUMN total_points_awarded INTEGER NOT NULL DEFAULT 0;
    END IF;

    -- tier_id, bundle_quantity, daily_amount, next_billing_date, payment_method, auto_renew
    -- may already exist from the entity migration — add them only if missing
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'tier_id'
    ) THEN
        ALTER TABLE public.daily_subscriptions ADD COLUMN tier_id UUID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'bundle_quantity'
    ) THEN
        ALTER TABLE public.daily_subscriptions
            ADD COLUMN bundle_quantity INTEGER NOT NULL DEFAULT 1;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'daily_amount'
    ) THEN
        ALTER TABLE public.daily_subscriptions
            ADD COLUMN daily_amount BIGINT NOT NULL DEFAULT 2000;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'next_billing_date'
    ) THEN
        ALTER TABLE public.daily_subscriptions
            ADD COLUMN next_billing_date TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '1 day');
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'last_billing_date'
    ) THEN
        ALTER TABLE public.daily_subscriptions ADD COLUMN last_billing_date TIMESTAMP;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'payment_method'
    ) THEN
        ALTER TABLE public.daily_subscriptions
            ADD COLUMN payment_method VARCHAR(50) NOT NULL DEFAULT 'paystack';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'auto_renew'
    ) THEN
        ALTER TABLE public.daily_subscriptions
            ADD COLUMN auto_renew BOOLEAN NOT NULL DEFAULT TRUE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'cancelled_at'
    ) THEN
        ALTER TABLE public.daily_subscriptions ADD COLUMN cancelled_at TIMESTAMP;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'cancellation_reason'
    ) THEN
        ALTER TABLE public.daily_subscriptions ADD COLUMN cancellation_reason TEXT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'daily_subscriptions'
          AND column_name = 'total_entries'
    ) THEN
        ALTER TABLE public.daily_subscriptions ADD COLUMN total_entries INTEGER NOT NULL DEFAULT 0;
    END IF;
END $$;

-- ── Step 3: index for the billing job (needs to find subs due for billing fast) ─
CREATE INDEX IF NOT EXISTS idx_daily_subscriptions_billing
    ON public.daily_subscriptions (status, next_billing_date)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_daily_subscriptions_auth_code
    ON public.daily_subscriptions (paystack_authorization_code)
    WHERE paystack_authorization_code IS NOT NULL;

-- ── Step 4: subscription_billings table ────────────────────────────────────────
-- Authoritative record of every daily billing attempt.
-- Points/draw-entries are ONLY awarded when status = 'completed'.
-- Retry logic is driven by retry_count + next_retry_at.
CREATE TABLE IF NOT EXISTS public.subscription_billings (
    id                  UUID          NOT NULL DEFAULT uuid_generate_v4(),
    subscription_id     UUID          NOT NULL,      -- FK → daily_subscriptions.id
    msisdn              VARCHAR(25)   NOT NULL,
    billing_date        DATE          NOT NULL,       -- the calendar day this billing covers
    amount              BIGINT        NOT NULL,       -- kobo
    entries_to_award    INTEGER       NOT NULL DEFAULT 1,
    points_to_award     INTEGER       NOT NULL DEFAULT 1,
    status              VARCHAR(20)   NOT NULL DEFAULT 'pending',
        -- pending | attempted | completed | failed | skipped
    payment_reference   VARCHAR(255),                -- Paystack reference for this charge
    paystack_transaction_id BIGINT,                  -- Paystack transaction id
    gateway_response    TEXT,
    failure_reason      TEXT,
    retry_count         INTEGER       NOT NULL DEFAULT 0,
    max_retries         INTEGER       NOT NULL DEFAULT 3,
    next_retry_at       TIMESTAMP,                   -- NULL = no more retries
    points_awarded      BOOLEAN       NOT NULL DEFAULT FALSE,  -- idempotency flag
    processed_at        TIMESTAMP,
    created_at          TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP     NOT NULL DEFAULT NOW(),
    CONSTRAINT subscription_billings_pkey PRIMARY KEY (id),
    -- One billing record per subscription per calendar day (dedup guard)
    CONSTRAINT subscription_billings_sub_date_unique UNIQUE (subscription_id, billing_date),
    CONSTRAINT subscription_billings_status_check
        CHECK (status IN ('pending','attempted','completed','failed','skipped'))
);

CREATE INDEX IF NOT EXISTS idx_sub_billings_subscription
    ON public.subscription_billings (subscription_id);

CREATE INDEX IF NOT EXISTS idx_sub_billings_msisdn
    ON public.subscription_billings (msisdn);

CREATE INDEX IF NOT EXISTS idx_sub_billings_date
    ON public.subscription_billings (billing_date);

CREATE INDEX IF NOT EXISTS idx_sub_billings_status
    ON public.subscription_billings (status);

-- Billing job needs: find all pending/attempted billings that are due for retry
CREATE INDEX IF NOT EXISTS idx_sub_billings_retry
    ON public.subscription_billings (status, next_retry_at)
    WHERE status IN ('pending', 'attempted') AND next_retry_at IS NOT NULL;

COMMENT ON TABLE public.subscription_billings IS
    'One row per daily billing attempt per active subscription. '
    'Points and draw entries are awarded ONLY when status=completed and points_awarded=true. '
    'Retry schedule: attempt 1 at next_billing_date, '
    'retry 1 at +1h, retry 2 at +3h, retry 3 at +8h. '
    'After 3 retries mark status=failed and schedule next_billing_date=tomorrow.';
