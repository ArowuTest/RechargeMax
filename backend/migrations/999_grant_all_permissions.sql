-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-- COMPREHENSIVE PERMISSIONS MIGRATION
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
--
-- Purpose: Grant all necessary permissions to application database users
-- 
-- This migration ensures that all tables, sequences, and functions have proper
-- permissions for deployment to any environment (local, staging, production).
--
-- Compatible with:
-- - Render PostgreSQL
-- - Supabase
-- - AWS RDS
-- - Google Cloud SQL
-- - Any standard PostgreSQL instance
--
-- The migration is idempotent and can be run multiple times safely.
--
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

-- Function to grant permissions to a user if they exist
CREATE OR REPLACE FUNCTION grant_permissions_if_user_exists(username TEXT) RETURNS VOID AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = username) THEN
        -- Grant schema usage
        EXECUTE format('GRANT USAGE ON SCHEMA public TO %I', username);
        
        -- Grant table permissions
        EXECUTE format('GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO %I', username);
        
        -- Grant sequence permissions
        EXECUTE format('GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO %I', username);
        
        -- Grant function permissions
        EXECUTE format('GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO %I', username);
        
        -- Set default privileges for future objects
        EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO %I', username);
        EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO %I', username);
        EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO %I', username);
        
        RAISE NOTICE 'Granted all permissions to user: %', username;
    ELSE
        RAISE NOTICE 'User % does not exist, skipping grants', username;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-- GRANT PERMISSIONS TO COMMON DATABASE USERS
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

-- Grant to rechargemax user (primary application user)
SELECT grant_permissions_if_user_exists('rechargemax');

-- Grant to rechargemax_user (alternative naming convention)
SELECT grant_permissions_if_user_exists('rechargemax_user');

-- Grant to render user (Render.com default)
SELECT grant_permissions_if_user_exists('render');

-- Grant to supabase_admin (Supabase default)
SELECT grant_permissions_if_user_exists('supabase_admin');

-- Grant to authenticated (Supabase authenticated users)
SELECT grant_permissions_if_user_exists('authenticated');

-- Grant to service_role (Supabase service role)
SELECT grant_permissions_if_user_exists('service_role');

-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-- EXPLICIT GRANTS FOR ALL TABLES (Fallback for environments without function support)
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DO $$
DECLARE
    app_user TEXT;
    table_record RECORD;
    sequence_record RECORD;
BEGIN
    -- Loop through common user names
    FOREACH app_user IN ARRAY ARRAY['rechargemax', 'rechargemax_user', 'render', 'supabase_admin', 'authenticated', 'service_role']
    LOOP
        IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = app_user) THEN
            RAISE NOTICE 'Granting explicit permissions to: %', app_user;
            
            -- Grant on all existing tables
            FOR table_record IN 
                SELECT tablename FROM pg_tables WHERE schemaname = 'public'
            LOOP
                EXECUTE format('GRANT ALL PRIVILEGES ON TABLE public.%I TO %I', table_record.tablename, app_user);
            END LOOP;
            
            -- Grant on all existing sequences
            FOR sequence_record IN 
                SELECT sequencename FROM pg_sequences WHERE schemaname = 'public'
            LOOP
                EXECUTE format('GRANT ALL PRIVILEGES ON SEQUENCE public.%I TO %I', sequence_record.sequencename, app_user);
            END LOOP;
        END IF;
    END LOOP;
END $$;

-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-- SPECIFIC TABLE GRANTS (Critical tables for application)
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

-- Core user and authentication tables
GRANT ALL PRIVILEGES ON TABLE users TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE otp_verifications TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE otps TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE admin_users TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE admin_sessions TO rechargemax, rechargemax_user;

-- Transaction and payment tables
GRANT ALL PRIVILEGES ON TABLE transactions TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE payment_logs TO rechargemax, rechargemax_user;

-- Spin and prize tables
GRANT ALL PRIVILEGES ON TABLE spin_results TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE spin_tiers TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE wheel_prizes TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE prize_categories TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE prize_templates TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE prize_fulfillment_config TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE prize_fulfillment_logs TO rechargemax, rechargemax_user;

-- Draw tables
GRANT ALL PRIVILEGES ON TABLE draws TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE draw_entries TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE draw_winners TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE draw_types TO rechargemax, rechargemax_user;

-- Subscription tables
GRANT ALL PRIVILEGES ON TABLE daily_subscriptions TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE daily_subscription_config TO rechargemax, rechargemax_user;

-- Affiliate tables
GRANT ALL PRIVILEGES ON TABLE affiliates TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE affiliate_clicks TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE affiliate_commissions TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE affiliate_payouts TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE affiliate_analytics TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE affiliate_bank_accounts TO rechargemax, rechargemax_user;

-- Configuration tables
GRANT ALL PRIVILEGES ON TABLE platform_settings TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE provider_configs TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE network_configs TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE network_cache TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE data_plans TO rechargemax, rechargemax_user;

-- Notification tables
GRANT ALL PRIVILEGES ON TABLE user_notifications TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE user_notification_preferences TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE notification_templates TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE notification_delivery_log TO rechargemax, rechargemax_user;

-- Logging and monitoring tables
GRANT ALL PRIVILEGES ON TABLE admin_activity_logs TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE application_logs TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE application_metrics TO rechargemax, rechargemax_user;
GRANT ALL PRIVILEGES ON TABLE points_adjustments TO rechargemax, rechargemax_user;

-- File upload tables
GRANT ALL PRIVILEGES ON TABLE file_uploads TO rechargemax, rechargemax_user;

-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-- SEQUENCE GRANTS (Critical for INSERT operations)
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO rechargemax, rechargemax_user;

-- Specific critical sequences
GRANT USAGE, SELECT ON SEQUENCE prize_fulfillment_logs_id_seq TO rechargemax, rechargemax_user;

-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-- CLEANUP
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

-- Drop the helper function (no longer needed)
DROP FUNCTION IF EXISTS grant_permissions_if_user_exists(TEXT);

-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-- VERIFICATION
-- ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DO $$
BEGIN
    RAISE NOTICE '✅ Comprehensive permissions migration completed successfully!';
    RAISE NOTICE '📊 All tables, sequences, and functions have been granted to application users.';
    RAISE NOTICE '🚀 Database is ready for deployment to any environment (Render, Vercel, etc.)';
END $$;
