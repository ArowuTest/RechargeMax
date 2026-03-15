-- ============================================================
-- Table: draw_entries
-- ============================================================

CREATE TABLE public.draw_entries (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    draw_id uuid,
    user_id uuid,
    msisdn text NOT NULL,
    entries_count integer DEFAULT 1,
    source_type text NOT NULL,
    source_transaction_id uuid,
    source_subscription_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT draw_entries_source_type_check CHECK ((source_type = ANY (ARRAY['TRANSACTION'::text, 'SUBSCRIPTION'::text, 'BONUS'::text, 'MANUAL'::text]))),
    CONSTRAINT positive_entries CHECK ((entries_count > 0))
);

ALTER TABLE ONLY public.draw_entries
    ADD CONSTRAINT draw_entries_pkey PRIMARY KEY (id);

CREATE INDEX idx_draw_entries_draw_id ON public.draw_entries USING btree (draw_id);

CREATE INDEX idx_draw_entries_source_type ON public.draw_entries USING btree (source_type);

CREATE INDEX idx_draw_entries_user_id ON public.draw_entries USING btree (user_id);

CREATE TRIGGER update_draw_entries_trigger AFTER INSERT OR DELETE OR UPDATE ON public.draw_entries FOR EACH ROW EXECUTE FUNCTION public.trigger_update_draw_entries();

ALTER TABLE ONLY public.draw_entries
    ADD CONSTRAINT draw_entries_draw_id_fkey FOREIGN KEY (draw_id) REFERENCES public.draws(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.draw_entries
    ADD CONSTRAINT draw_entries_source_subscription_id_fkey FOREIGN KEY (source_subscription_id) REFERENCES public.daily_subscriptions(id);

ALTER TABLE ONLY public.draw_entries
    ADD CONSTRAINT draw_entries_source_transaction_id_fkey FOREIGN KEY (source_transaction_id) REFERENCES public.transactions(id);

ALTER TABLE ONLY public.draw_entries
    ADD CONSTRAINT draw_entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE public.draw_entries ENABLE ROW LEVEL SECURITY;
