-- ============================================================
-- Table: bank_accounts
-- User bank accounts for cash prize withdrawals.
-- ============================================================
CREATE TABLE IF NOT EXISTS public.bank_accounts (
    id                  UUID          NOT NULL DEFAULT uuid_generate_v4(),
    user_id             UUID          NOT NULL,
    account_name        VARCHAR(255)  NOT NULL,
    account_number      VARCHAR(20)   NOT NULL,
    bank_name           VARCHAR(255)  NOT NULL,
    bank_code           VARCHAR(20)   NOT NULL,
    is_verified         BOOLEAN       DEFAULT FALSE,
    is_primary          BOOLEAN       DEFAULT FALSE,
    verification_method VARCHAR(100),
    verification_data   JSONB,
    created_at          TIMESTAMP     NOT NULL DEFAULT NOW(),
    verified_at         TIMESTAMP,
    last_used_at        TIMESTAMP,
    CONSTRAINT bank_accounts_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_bank_accounts_user      ON public.bank_accounts (user_id);
CREATE INDEX IF NOT EXISTS idx_bank_accounts_primary   ON public.bank_accounts (user_id, is_primary);
