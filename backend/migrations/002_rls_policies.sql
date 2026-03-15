-- RLS Policies: Disabled for Go-native JWT auth (not Supabase auth)
-- All access control is enforced at the application layer via JWT middleware.
-- RLS remains enabled on tables but policies use permissive TRUE to allow
-- the Go backend (connecting as superuser/service role) full access.

ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.admin_users ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.spin_results ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.daily_subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.draws ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.draw_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.draw_winners ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.affiliates ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.affiliate_commissions ENABLE ROW LEVEL SECURITY;

-- Grant full access to the rechargemax service role (Go backend)
CREATE POLICY "service_full_access_users"             ON public.users             FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "service_full_access_admin_users"       ON public.admin_users       FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "service_full_access_transactions"      ON public.transactions      FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "service_full_access_spin_results"      ON public.spin_results      FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "service_full_access_daily_subs"        ON public.daily_subscriptions FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "service_full_access_draws"             ON public.draws             FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "service_full_access_draw_entries"      ON public.draw_entries      FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "service_full_access_draw_winners"      ON public.draw_winners      FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "service_full_access_affiliates"        ON public.affiliates        FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "service_full_access_affiliate_comms"   ON public.affiliate_commissions FOR ALL USING (true) WITH CHECK (true);
