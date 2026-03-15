-- Grant all permissions to the application database user
-- Needed for both Render's rechargemax_app and any future service roles

DO $$
DECLARE
    r RECORD;
    app_user TEXT := current_user;
BEGIN
    -- Grant to current user (the app user connecting to run migrations)
    FOR r IN SELECT tablename FROM pg_tables WHERE schemaname = 'public'
    LOOP
        EXECUTE 'GRANT ALL PRIVILEGES ON TABLE public.' || quote_ident(r.tablename) || ' TO ' || quote_ident(app_user);
    END LOOP;
    
    EXECUTE 'GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ' || quote_ident(app_user);
    EXECUTE 'GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO ' || quote_ident(app_user);
END $$;

-- Also grant to specific known usernames
DO $$
BEGIN
    -- Try granting to rechargemax_app (Render DB user)
    BEGIN
        GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO rechargemax_app;
        GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO rechargemax_app;
        GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO rechargemax_app;
    EXCEPTION WHEN undefined_object THEN
        -- User doesn't exist, skip
        NULL;
    END;
END $$;

-- Disable RLS for the app user (bypass RLS on all protected tables)
ALTER TABLE IF EXISTS public.admin_users FORCE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS public.users FORCE ROW LEVEL SECURITY;

-- Ensure permissive policies exist (idempotent)
DO $$
BEGIN
    BEGIN
        DROP POLICY IF EXISTS "service_full_access_admin_users" ON public.admin_users;
        CREATE POLICY "service_full_access_admin_users" ON public.admin_users FOR ALL USING (true) WITH CHECK (true);
    EXCEPTION WHEN OTHERS THEN NULL;
    END;
    
    BEGIN
        DROP POLICY IF EXISTS "service_full_access_users" ON public.users;
        CREATE POLICY "service_full_access_users" ON public.users FOR ALL USING (true) WITH CHECK (true);
    EXCEPTION WHEN OTHERS THEN NULL;
    END;
END $$;
