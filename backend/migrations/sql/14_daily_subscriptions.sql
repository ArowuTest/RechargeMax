-- ============================================================
-- Table: daily_subscriptions
-- ============================================================

CREATE TABLE public.daily_subscriptions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    msisdn text NOT NULL,
    subscription_date date NOT NULL,
    amount numeric(5,2) NOT NULL,
    draw_entries_earned integer DEFAULT 1,
    points_earned integer DEFAULT 0,
    payment_reference text,
    status text DEFAULT 'active'::text,
    is_paid boolean DEFAULT false,
    customer_email text,
    customer_name text,
    created_at timestamp with time zone DEFAULT now(),
    subscription_code character varying(50),
    CONSTRAINT daily_subscriptions_status_check CHECK ((status = ANY (ARRAY['active'::text, 'pending'::text, 'cancelled'::text, 'expired'::text, 'paused'::text]))),
    CONSTRAINT positive_amount CHECK ((amount > (0)::numeric)),
    CONSTRAINT valid_msisdn CHECK ((msisdn ~ '^234[789][01][0-9]{8}$'::text))
);

ALTER TABLE ONLY public.daily_subscriptions
    ADD CONSTRAINT daily_subscriptions_pkey PRIMARY KEY (id);

-- NOTE: UNIQUE(user_id, subscription_date) intentionally removed here.
-- Migration 043 drops this constraint to support multi-line subscriptions
-- (multiple lines on the same day for the same user). Adding it here causes
-- a 23505 error on every restart because duplicate rows already exist.
-- The non-unique index idx_daily_subscriptions_user_date (added by 043) covers
-- query performance without enforcing the unwanted uniqueness.

CREATE INDEX idx_daily_subscriptions_msisdn ON public.daily_subscriptions USING btree (msisdn);

CREATE INDEX idx_daily_subscriptions_status ON public.daily_subscriptions USING btree (status);

CREATE INDEX idx_daily_subscriptions_subscription_date ON public.daily_subscriptions USING btree (subscription_date);

CREATE INDEX idx_daily_subscriptions_user_id ON public.daily_subscriptions USING btree (user_id);

ALTER TABLE ONLY public.daily_subscriptions
    ADD CONSTRAINT daily_subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE public.daily_subscriptions ENABLE ROW LEVEL SECURITY;
