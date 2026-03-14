-- Migration 039: Create wallets and wallet_transactions tables
-- These tables support the affiliate payout wallet system.

-- Wallets table: one per affiliate MSISDN
CREATE TABLE IF NOT EXISTS wallets (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    msisdn              VARCHAR(20)  NOT NULL UNIQUE,
    balance             BIGINT       NOT NULL DEFAULT 0 CHECK (balance >= 0),
    pending_balance     BIGINT       NOT NULL DEFAULT 0 CHECK (pending_balance >= 0),
    total_earned        BIGINT       NOT NULL DEFAULT 0,
    total_withdrawn     BIGINT       NOT NULL DEFAULT 0,
    min_payout_amount   BIGINT       NOT NULL DEFAULT 100000,  -- ₦1000 in kobo
    is_active           BOOLEAN      NOT NULL DEFAULT TRUE,
    is_suspended        BOOLEAN      NOT NULL DEFAULT FALSE,
    suspension_reason   TEXT,
    last_transaction_at TIMESTAMPTZ,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallets_msisdn ON wallets(msisdn);

-- Wallet transactions ledger
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id       UUID         NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    msisdn          VARCHAR(20)  NOT NULL,
    type            VARCHAR(30)  NOT NULL CHECK (type IN ('credit','debit','hold','release','adjustment')),
    amount          BIGINT       NOT NULL CHECK (amount > 0),
    balance_before  BIGINT       NOT NULL,
    balance_after   BIGINT       NOT NULL,
    reference       VARCHAR(100) NOT NULL UNIQUE,
    description     TEXT,
    status          VARCHAR(20)  NOT NULL DEFAULT 'completed' CHECK (status IN ('pending','completed','failed','reversed')),
    metadata        JSONB,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_msisdn    ON wallet_transactions(msisdn);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_reference ON wallet_transactions(reference);
