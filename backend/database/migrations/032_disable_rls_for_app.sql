-- Disable RLS restrictions for the app database user
-- The app user connects as rechargemax_app and needs full access to all tables.
-- We use FORCE ROW LEVEL SECURITY + USING(true) policy approach.

-- First ensure we have permissive policies on all tables with RLS enabled
DO $$
DECLARE
    t RECORD;
BEGIN
    FOR t IN 
        SELECT tablename, schemaname 
        FROM pg_tables 
        WHERE schemaname = 'public'
    LOOP
        -- Enable RLS (idempotent)
        BEGIN
            EXECUTE format('ALTER TABLE public.%I ENABLE ROW LEVEL SECURITY', t.tablename);
        EXCEPTION WHEN OTHERS THEN NULL;
        END;
        
        -- Drop any existing catch-all policy and recreate it
        BEGIN
            EXECUTE format('DROP POLICY IF EXISTS "allow_all_service_access" ON public.%I', t.tablename);
            EXECUTE format(
                'CREATE POLICY "allow_all_service_access" ON public.%I FOR ALL TO PUBLIC USING (true) WITH CHECK (true)',
                t.tablename
            );
        EXCEPTION WHEN OTHERS THEN NULL;
        END;
    END LOOP;
END $$;

-- Also ensure admin_users table specifically has the policy
DROP POLICY IF EXISTS "service_full_access_admin_users" ON public.admin_users;
CREATE POLICY "service_full_access_admin_users" 
    ON public.admin_users 
    FOR ALL 
    TO PUBLIC 
    USING (true) 
    WITH CHECK (true);

-- Grant explicit permissions
GRANT ALL ON ALL TABLES IN SCHEMA public TO rechargemax_app;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO rechargemax_app;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO rechargemax_app;

-- Completely disable RLS on critical tables (simplest approach)
ALTER TABLE public.admin_users DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.users DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.transactions DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.spin_results DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.daily_subscriptions DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.draws DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.draw_entries DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.draw_winners DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.affiliates DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.affiliate_commissions DISABLE ROW LEVEL SECURITY;
