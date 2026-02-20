-- Comprehensive security policies for the RechargeMax platform
-- Created: 2026-01-30 14:00 UTC
-- ============================================================================

-- Enable RLS on all tables
ALTER TABLE public.users_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.admin_users_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.admin_sessions_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.network_configs_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.data_plans_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.transactions_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.wheel_prizes_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.spin_results_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.daily_subscription_config_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.daily_subscriptions_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.draws_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.draw_entries_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.draw_winners_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.affiliates_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.affiliate_clicks_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.affiliate_commissions_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.platform_settings_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.application_logs_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.application_metrics_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- USERS TABLE POLICIES
-- ============================================================================

-- Users can view their own profile
CREATE POLICY "users_select_own" ON public.users_2026_01_30_14_00
    FOR SELECT USING (auth.uid() = auth_user_id);

-- Users can update their own profile
CREATE POLICY "users_update_own" ON public.users_2026_01_30_14_00
    FOR UPDATE USING (auth.uid() = auth_user_id);

-- Users can insert their own profile (registration)
CREATE POLICY "users_insert_own" ON public.users_2026_01_30_14_00
    FOR INSERT WITH CHECK (auth.uid() = auth_user_id);

-- Service role can manage all users
CREATE POLICY "users_service_role_all" ON public.users_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- ============================================================================
-- ADMIN SYSTEM POLICIES
-- ============================================================================

-- Only service role can access admin users
CREATE POLICY "admin_users_service_only" ON public.admin_users_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Only service role can access admin sessions
CREATE POLICY "admin_sessions_service_only" ON public.admin_sessions_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- ============================================================================
-- NETWORK AND DATA PLANS POLICIES
-- ============================================================================

-- Everyone can view active network configs (for frontend display)
CREATE POLICY "network_configs_select_active" ON public.network_configs_2026_01_30_14_00
    FOR SELECT USING (is_active = true);

-- Service role can manage network configs
CREATE POLICY "network_configs_service_manage" ON public.network_configs_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Everyone can view active data plans
CREATE POLICY "data_plans_select_active" ON public.data_plans_2026_01_30_14_00
    FOR SELECT USING (is_active = true);

-- Service role can manage data plans
CREATE POLICY "data_plans_service_manage" ON public.data_plans_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- ============================================================================
-- TRANSACTIONS POLICIES
-- ============================================================================

-- Users can view their own transactions
CREATE POLICY "transactions_select_own" ON public.transactions_2026_01_30_14_00
    FOR SELECT USING (
        auth.uid() IS NOT NULL AND (
            user_id = (SELECT id FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
            OR msisdn = (SELECT msisdn FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
        )
    );

-- Service role can manage all transactions
CREATE POLICY "transactions_service_manage" ON public.transactions_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Anonymous users can insert transactions (guest recharges)
CREATE POLICY "transactions_anonymous_insert" ON public.transactions_2026_01_30_14_00
    FOR INSERT WITH CHECK (true);

-- ============================================================================
-- GAMIFICATION POLICIES
-- ============================================================================

-- Everyone can view active wheel prizes
CREATE POLICY "wheel_prizes_select_active" ON public.wheel_prizes_2026_01_30_14_00
    FOR SELECT USING (is_active = true);

-- Service role can manage wheel prizes
CREATE POLICY "wheel_prizes_service_manage" ON public.wheel_prizes_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Users can view their own spin results
CREATE POLICY "spin_results_select_own" ON public.spin_results_2026_01_30_14_00
    FOR SELECT USING (
        auth.uid() IS NOT NULL AND (
            user_id = (SELECT id FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
            OR msisdn = (SELECT msisdn FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
        )
    );

-- Users can update their own spin results (for claiming)
CREATE POLICY "spin_results_update_own" ON public.spin_results_2026_01_30_14_00
    FOR UPDATE USING (
        auth.uid() IS NOT NULL AND (
            user_id = (SELECT id FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
            OR msisdn = (SELECT msisdn FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
        )
    );

-- Service role can manage spin results
CREATE POLICY "spin_results_service_manage" ON public.spin_results_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- ============================================================================
-- DAILY SUBSCRIPTION POLICIES
-- ============================================================================

-- Everyone can view daily subscription config
CREATE POLICY "daily_sub_config_select_all" ON public.daily_subscription_config_2026_01_30_14_00
    FOR SELECT USING (true);

-- Service role can manage daily subscription config
CREATE POLICY "daily_sub_config_service_manage" ON public.daily_subscription_config_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Users can view their own subscriptions
CREATE POLICY "daily_subs_select_own" ON public.daily_subscriptions_2026_01_30_14_00
    FOR SELECT USING (
        auth.uid() IS NOT NULL AND (
            user_id = (SELECT id FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
            OR msisdn = (SELECT msisdn FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
        )
    );

-- Service role can manage all subscriptions
CREATE POLICY "daily_subs_service_manage" ON public.daily_subscriptions_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Anonymous users can insert subscriptions (guest subscriptions)
CREATE POLICY "daily_subs_anonymous_insert" ON public.daily_subscriptions_2026_01_30_14_00
    FOR INSERT WITH CHECK (true);

-- ============================================================================
-- DRAW SYSTEM POLICIES
-- ============================================================================

-- Everyone can view active draws
CREATE POLICY "draws_select_active" ON public.draws_2026_01_30_14_00
    FOR SELECT USING (status IN ('ACTIVE', 'UPCOMING'));

-- Service role can manage draws
CREATE POLICY "draws_service_manage" ON public.draws_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Users can view their own draw entries
CREATE POLICY "draw_entries_select_own" ON public.draw_entries_2026_01_30_14_00
    FOR SELECT USING (
        auth.uid() IS NOT NULL AND (
            user_id = (SELECT id FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
            OR msisdn = (SELECT msisdn FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
        )
    );

-- Service role can manage draw entries
CREATE POLICY "draw_entries_service_manage" ON public.draw_entries_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Everyone can view draw winners (public information)
CREATE POLICY "draw_winners_select_all" ON public.draw_winners_2026_01_30_14_00
    FOR SELECT USING (true);

-- Users can update their own winner records (for claiming)
CREATE POLICY "draw_winners_update_own" ON public.draw_winners_2026_01_30_14_00
    FOR UPDATE USING (
        auth.uid() IS NOT NULL AND (
            user_id = (SELECT id FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
            OR msisdn = (SELECT msisdn FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
        )
    );

-- Service role can manage draw winners
CREATE POLICY "draw_winners_service_manage" ON public.draw_winners_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- ============================================================================
-- AFFILIATE SYSTEM POLICIES
-- ============================================================================

-- Users can view their own affiliate record
CREATE POLICY "affiliates_select_own" ON public.affiliates_2026_01_30_14_00
    FOR SELECT USING (
        auth.uid() IS NOT NULL AND 
        user_id = (SELECT id FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
    );

-- Users can update their own affiliate record
CREATE POLICY "affiliates_update_own" ON public.affiliates_2026_01_30_14_00
    FOR UPDATE USING (
        auth.uid() IS NOT NULL AND 
        user_id = (SELECT id FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
    );

-- Users can insert their own affiliate application
CREATE POLICY "affiliates_insert_own" ON public.affiliates_2026_01_30_14_00
    FOR INSERT WITH CHECK (
        auth.uid() IS NOT NULL AND 
        user_id = (SELECT id FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
    );

-- Service role can manage affiliates
CREATE POLICY "affiliates_service_manage" ON public.affiliates_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Service role can manage affiliate clicks
CREATE POLICY "affiliate_clicks_service_manage" ON public.affiliate_clicks_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Anonymous users can insert clicks (tracking)
CREATE POLICY "affiliate_clicks_anonymous_insert" ON public.affiliate_clicks_2026_01_30_14_00
    FOR INSERT WITH CHECK (true);

-- Users can view their own commissions
CREATE POLICY "affiliate_commissions_select_own" ON public.affiliate_commissions_2026_01_30_14_00
    FOR SELECT USING (
        auth.uid() IS NOT NULL AND 
        affiliate_id IN (
            SELECT id FROM public.affiliates_2026_01_30_14_00 
            WHERE user_id = (SELECT id FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid())
        )
    );

-- Service role can manage commissions
CREATE POLICY "affiliate_commissions_service_manage" ON public.affiliate_commissions_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- ============================================================================
-- SYSTEM TABLES POLICIES
-- ============================================================================

-- Everyone can view public platform settings
CREATE POLICY "platform_settings_select_public" ON public.platform_settings_2026_01_30_14_00
    FOR SELECT USING (is_public = true);

-- Service role can manage all platform settings
CREATE POLICY "platform_settings_service_manage" ON public.platform_settings_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Only service role can access application logs
CREATE POLICY "application_logs_service_only" ON public.application_logs_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Only service role can access application metrics
CREATE POLICY "application_metrics_service_only" ON public.application_metrics_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- ============================================================================
-- HELPER FUNCTIONS FOR POLICIES
-- ============================================================================

-- Function to check if user is admin (for future use)
CREATE OR REPLACE FUNCTION is_admin_user()
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM public.admin_users_2026_01_30_14_00 
        WHERE email = auth.email() AND is_active = true
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get user ID from auth
CREATE OR REPLACE FUNCTION get_user_id()
RETURNS UUID AS $$
BEGIN
    RETURN (SELECT id FROM public.users_2026_01_30_14_00 WHERE auth_user_id = auth.uid());
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

