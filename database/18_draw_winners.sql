-- ============================================================
-- Table: draw_winners
-- ============================================================

CREATE TABLE public.draw_winners (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    draw_id uuid,
    user_id uuid,
    msisdn text NOT NULL,
    "position" integer NOT NULL,
    prize_amount numeric(10,2) NOT NULL,
    claim_status text DEFAULT 'PENDING'::text,
    claimed_at timestamp with time zone,
    claim_reference text,
    created_at timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone DEFAULT (now() + '30 days'::interval),
    prize_category_id uuid,
    category_name character varying(100),
    is_runner_up boolean DEFAULT false,
    is_forfeited boolean DEFAULT false,
    promoted_from uuid,
    first_name text,
    last_name text,
    prize_type text DEFAULT 'cash'::text NOT NULL,
    prize_description text DEFAULT ''::text NOT NULL,
    data_package text,
    airtime_amount bigint,
    network text,
    auto_provision boolean DEFAULT false NOT NULL,
    provision_status text,
    provision_reference text,
    provisioned_at timestamp with time zone,
    provision_error text,
    claim_deadline timestamp with time zone,
    payout_status text DEFAULT 'pending'::text NOT NULL,
    payout_method text,
    bank_code text,
    bank_name text,
    account_number text,
    account_name text,
    payout_reference text,
    payout_error text,
    shipping_address text,
    shipping_phone text,
    shipping_status text,
    tracking_number text,
    shipped_at timestamp with time zone,
    delivered_at timestamp with time zone,
    notification_sent boolean DEFAULT false NOT NULL,
    notification_sent_at timestamp with time zone,
    notification_channels text,
    notes text,
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT draw_winners_claim_status_check CHECK ((claim_status = ANY (ARRAY['PENDING'::text, 'CLAIMED'::text, 'EXPIRED'::text]))),
    CONSTRAINT positive_position CHECK (("position" > 0)),
    CONSTRAINT positive_prize_amount CHECK ((prize_amount > (0)::numeric))
);

ALTER TABLE ONLY public.draw_winners
    ADD CONSTRAINT draw_winners_draw_id_position_key UNIQUE (draw_id, "position");

ALTER TABLE ONLY public.draw_winners
    ADD CONSTRAINT draw_winners_pkey PRIMARY KEY (id);

CREATE INDEX idx_draw_winners_category ON public.draw_winners USING btree (prize_category_id);

CREATE INDEX idx_draw_winners_claim_status ON public.draw_winners USING btree (claim_status);

CREATE INDEX idx_draw_winners_draw_id ON public.draw_winners USING btree (draw_id);

CREATE INDEX idx_draw_winners_forfeited ON public.draw_winners USING btree (draw_id, is_forfeited);

CREATE INDEX idx_draw_winners_payout ON public.draw_winners USING btree (payout_status);

CREATE INDEX idx_draw_winners_provision ON public.draw_winners USING btree (provision_status) WHERE (auto_provision = true);

CREATE INDEX idx_draw_winners_runner_up ON public.draw_winners USING btree (draw_id, is_runner_up);

CREATE INDEX idx_draw_winners_user_id ON public.draw_winners USING btree (user_id);

ALTER TABLE ONLY public.draw_winners
    ADD CONSTRAINT draw_winners_draw_id_fkey FOREIGN KEY (draw_id) REFERENCES public.draws(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.draw_winners
    ADD CONSTRAINT draw_winners_prize_category_id_fkey FOREIGN KEY (prize_category_id) REFERENCES public.prize_categories(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.draw_winners
    ADD CONSTRAINT draw_winners_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE public.draw_winners ENABLE ROW LEVEL SECURITY;
