-- ═══════════════════════════════════════════════════════════════════════════
-- Migration 054: Create commission_payouts table
-- ═══════════════════════════════════════════════════════════════════════════
-- Previously RequestPayout() was a commented-out stub.  This table makes
-- payouts a first-class, auditable entity.
--
-- Flow:
--   Admin approves commissions  →  status = APPROVED
--   Admin clicks "Initiate Payout"  →  commission_payouts row created (PENDING)
--   Paystack transfer initiated     →  transfer_reference populated
--   Paystack webhook confirms        →  status = PAID, paid_at set
--   Affiliate balance decremented   →  affiliate.total_commission reduced
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS public.commission_payouts (
    id                  UUID        DEFAULT public.uuid_generate_v4() NOT NULL,
    affiliate_id        UUID        NOT NULL,

    -- Amount in kobo (integer, consistent with commission_amount)
    amount_kobo         BIGINT      NOT NULL,

    -- Naira display amount (amount_kobo / 100)
    amount_ngn          NUMERIC(12,2) GENERATED ALWAYS AS (amount_kobo / 100.0) STORED,

    status              TEXT        NOT NULL DEFAULT 'PENDING'
                            CONSTRAINT commission_payouts_status_check
                            CHECK (status IN ('PENDING','IN_TRANSIT','PAID','FAILED','CANCELLED')),

    -- Bank details snapshot at time of payout (denormalised for audit)
    bank_name           TEXT        NOT NULL DEFAULT '',
    account_number      TEXT        NOT NULL DEFAULT '',
    account_name        TEXT        NOT NULL DEFAULT '',

    -- Paystack transfer details
    transfer_reference  TEXT,
    transfer_code       TEXT,          -- Paystack transfer_code for status polling
    paystack_receipt    JSONB,         -- full Paystack webhook payload stored for audit

    -- Who initiated and when
    initiated_by        UUID,          -- admin user id
    initiated_at        TIMESTAMPTZ,
    paid_at             TIMESTAMPTZ,
    failed_reason       TEXT,

    -- Weekly payout batch identifier (YYYY-WNN e.g. 2026-W13)
    payout_week         TEXT,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT commission_payouts_pkey PRIMARY KEY (id),
    CONSTRAINT commission_payouts_amount_positive CHECK (amount_kobo > 0)
);

-- FK to affiliates
ALTER TABLE public.commission_payouts
    ADD CONSTRAINT commission_payouts_affiliate_id_fkey
        FOREIGN KEY (affiliate_id) REFERENCES public.affiliates(id) ON DELETE RESTRICT;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_commission_payouts_affiliate_id
    ON public.commission_payouts (affiliate_id);

CREATE INDEX IF NOT EXISTS idx_commission_payouts_status
    ON public.commission_payouts (status);

CREATE INDEX IF NOT EXISTS idx_commission_payouts_payout_week
    ON public.commission_payouts (payout_week);

CREATE INDEX IF NOT EXISTS idx_commission_payouts_initiated_at
    ON public.commission_payouts (initiated_at DESC);

-- Auto-update updated_at
CREATE TRIGGER update_commission_payouts_updated_at
    BEFORE UPDATE ON public.commission_payouts
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

-- ── Link commissions to payouts ─────────────────────────────────────────────
-- When commissions are batched into a payout we record which payout they
-- belong to.  This allows partial payout tracking.
ALTER TABLE public.affiliate_commissions
    ADD COLUMN IF NOT EXISTS payout_id UUID REFERENCES public.commission_payouts(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_affiliate_commissions_payout_id
    ON public.affiliate_commissions (payout_id)
    WHERE payout_id IS NOT NULL;

-- ── Add status APPROVED to affiliate_commissions ────────────────────────────
-- Original status check only had PENDING | APPROVED | PAID | CANCELLED.
-- Keeping this as-is — APPROVED = ready for payout, PAID = payout completed.
-- No schema change needed here; constraint already covers it.
