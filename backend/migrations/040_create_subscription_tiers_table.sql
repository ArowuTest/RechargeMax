-- Migration 040: Create subscription_tiers table
-- Defines the available daily subscription tiers (e.g. Basic ₦20/day = 1 draw entry).

CREATE TABLE IF NOT EXISTS subscription_tiers (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT        NOT NULL UNIQUE,
    description TEXT,
    entries     INTEGER     NOT NULL DEFAULT 1 CHECK (entries > 0),
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    sort_order  INTEGER     NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscription_tiers_is_active ON subscription_tiers(is_active);

-- Seed the default Basic tier (₦20/day = 1 draw entry)
INSERT INTO subscription_tiers (name, description, entries, is_active, sort_order)
VALUES ('Basic', '₦20 per day — 1 guaranteed daily draw entry', 1, TRUE, 1)
ON CONFLICT (name) DO NOTHING;
