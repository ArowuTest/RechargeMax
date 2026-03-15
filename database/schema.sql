-- ============================================================================
-- RechargeMax Database Schema
-- ============================================================================
-- This is the CANONICAL schema - the ground truth of what the database
-- looks like after all migrations have been applied.
--
-- For a FRESH database: run this file first, then run seeds.
-- For an EXISTING database: run only the files in migrations/ folder.
--
-- Regenerate with:
--   pg_dump -U rechargemax -d rechargemax --schema-only --no-owner --no-acl -f schema.sql
-- ============================================================================

--
-- PostgreSQL database dump
--



SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: calculate_draw_entries(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_draw_entries(p_points integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Simple 1:1 ratio: 1 point = 1 draw entry
    -- This allows admin to control draw entries by adjusting points formula
    RETURN p_points;
END;
$$;


--
-- Name: calculate_loyalty_tier(numeric); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_loyalty_tier(total_amount numeric) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF total_amount >= 100000 THEN
        RETURN 'PLATINUM';
    ELSIF total_amount >= 50000 THEN
        RETURN 'GOLD';
    ELSIF total_amount >= 20000 THEN
        RETURN 'SILVER';
    ELSE
        RETURN 'BRONZE';
    END IF;
END;
$$;


--
-- Name: calculate_points_earned(bigint); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_points_earned(p_amount_kobo bigint) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_naira_per_point INTEGER;
    v_amount_naira DECIMAL;
    v_points INTEGER;
BEGIN
    -- Get naira per point from settings (default to 200 if not found)
    SELECT COALESCE((setting_value::INTEGER), 200) INTO v_naira_per_point
    FROM public.platform_settings
    WHERE setting_key = 'naira_per_point';
    
    -- Handle NO ROWS case (when setting doesn't exist)
    IF v_naira_per_point IS NULL THEN
        v_naira_per_point := 200;
    END IF;
    
    -- Convert kobo to naira
    v_amount_naira := p_amount_kobo::DECIMAL / 100;
    
    -- Calculate points: FLOOR(naira / naira_per_point)
    -- Example: ₦500 / ₦200 = 2.5 → FLOOR = 2 points
    v_points := FLOOR(v_amount_naira / v_naira_per_point)::INTEGER;
    
    RETURN v_points;
EXCEPTION
    WHEN OTHERS THEN
        -- Log error and return 0 instead of NULL
        RAISE WARNING 'Error in calculate_points_earned: %', SQLERRM;
        RETURN 0;
END;
$$;


--
-- Name: check_otp_rate_limit(text, text, integer, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.check_otp_rate_limit(p_msisdn text, p_purpose text, p_time_window_minutes integer DEFAULT 5, p_max_requests integer DEFAULT 3) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    request_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO request_count
    FROM public.otp_verifications
    WHERE msisdn = p_msisdn
    AND purpose = p_purpose
    AND created_at > NOW() - (p_time_window_minutes || ' minutes')::INTERVAL;
    
    RETURN request_count < p_max_requests;
END;
$$;


--
-- Name: cleanup_expired_otps(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.cleanup_expired_otps() RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    DELETE FROM otps WHERE expires_at < NOW() - INTERVAL '24 hours';
END;
$$;


--
-- Name: cleanup_old_admin_logs(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.cleanup_old_admin_logs(retention_days integer DEFAULT 90) RETURNS TABLE(deleted_count bigint)
    LANGUAGE plpgsql
    AS $$
DECLARE
    cutoff_date TIMESTAMP WITH TIME ZONE;
    rows_deleted BIGINT;
BEGIN
    cutoff_date := NOW() - (retention_days || ' days')::INTERVAL;
    
    DELETE FROM public.admin_activity_logs
    WHERE created_at < cutoff_date
    AND is_suspicious = false
    AND risk_score < 50;
    
    GET DIAGNOSTICS rows_deleted = ROW_COUNT;
    
    RETURN QUERY SELECT rows_deleted;
END;
$$;


--
-- Name: cleanup_old_otps(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.cleanup_old_otps(retention_days integer DEFAULT 30) RETURNS TABLE(deleted_count bigint)
    LANGUAGE plpgsql
    AS $$
DECLARE
    cutoff_date TIMESTAMP WITH TIME ZONE;
    rows_deleted BIGINT;
BEGIN
    cutoff_date := NOW() - (retention_days || ' days')::INTERVAL;
    
    DELETE FROM public.otp_verifications
    WHERE created_at < cutoff_date
    AND (is_verified = true OR is_expired = true OR is_revoked = true);
    
    GET DIAGNOSTICS rows_deleted = ROW_COUNT;
    
    RETURN QUERY SELECT rows_deleted;
END;
$$;


--
-- Name: cleanup_old_payment_logs(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.cleanup_old_payment_logs(retention_days integer DEFAULT 90) RETURNS TABLE(deleted_count bigint)
    LANGUAGE plpgsql
    AS $$
DECLARE
    cutoff_date TIMESTAMP WITH TIME ZONE;
    rows_deleted BIGINT;
BEGIN
    cutoff_date := NOW() - (retention_days || ' days')::INTERVAL;
    
    -- Keep failed requests longer for debugging
    DELETE FROM public.payment_logs
    WHERE created_at < cutoff_date
    AND is_successful = true;
    
    GET DIAGNOSTICS rows_deleted = ROW_COUNT;
    
    RETURN QUERY SELECT rows_deleted;
END;
$$;


--
-- Name: conduct_draw(uuid); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.conduct_draw(p_draw_id uuid) RETURNS TABLE(winner_user_id uuid, winner_msisdn text, prize_amount numeric)
    LANGUAGE plpgsql
    AS $$
DECLARE
    draw_record RECORD;
    total_entries INTEGER;
    winners_needed INTEGER;
    prize_per_winner DECIMAL;
    entry_record RECORD;
    selected_entries UUID[];
    random_position INTEGER;
    current_position INTEGER := 0;
BEGIN
    -- Get draw details
    SELECT * INTO draw_record
    FROM public.draws
    WHERE id = p_draw_id AND status = 'ACTIVE';
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Draw not found or not active: %', p_draw_id;
    END IF;
    
    -- Get total entries
    SELECT COALESCE(SUM(entries_count), 0) INTO total_entries
    FROM public.draw_entries
    WHERE draw_id = p_draw_id;
    
    IF total_entries = 0 THEN
        RAISE EXCEPTION 'No entries found for draw: %', p_draw_id;
    END IF;
    
    winners_needed := draw_record.winners_count;
    prize_per_winner := draw_record.prize_pool / winners_needed;
    
    -- Select random winners
    FOR i IN 1..winners_needed LOOP
        random_position := (RANDOM() * total_entries)::INTEGER + 1;
        current_position := 0;
        
        -- Find the entry at the random position
        FOR entry_record IN 
            SELECT de.user_id, de.msisdn, de.entries_count
            FROM public.draw_entries de
            WHERE de.draw_id = p_draw_id
            ORDER BY de.created_at
        LOOP
            current_position := current_position + entry_record.entries_count;
            
            IF current_position >= random_position THEN
                -- Check if this user hasn't already won
                IF NOT (entry_record.user_id = ANY(selected_entries)) THEN
                    selected_entries := array_append(selected_entries, entry_record.user_id);
                    
                    -- Create winner record
                    INSERT INTO public.draw_winners (
                        id, draw_id, user_id, msisdn, prize_amount, 
                        claim_status, created_at
                    ) VALUES (
                        uuid_generate_v4(),
                        p_draw_id,
                        entry_record.user_id,
                        entry_record.msisdn,
                        prize_per_winner,
                        'PENDING',
                        NOW()
                    );
                    
                    -- Return winner info
                    winner_user_id := entry_record.user_id;
                    winner_msisdn := entry_record.msisdn;
                    prize_amount := prize_per_winner;
                    RETURN NEXT;
                    
                    EXIT; -- Move to next winner
                END IF;
            END IF;
        END LOOP;
    END LOOP;
    
    -- Update draw status
    UPDATE public.draws
    SET 
        status = 'COMPLETED',
        completed_at = NOW(),
        updated_at = NOW()
    WHERE id = p_draw_id;
END;
$$;


--
-- Name: create_affiliate_payout(uuid, uuid[], text, text, text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.create_affiliate_payout(p_affiliate_id uuid, p_commission_ids uuid[], p_bank_name text, p_account_number text, p_account_name text) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_payout_id UUID;
    v_total_amount DECIMAL(12,2);
    v_commission_count INTEGER;
    v_payout_fee DECIMAL(12,2);
    v_net_amount DECIMAL(12,2);
BEGIN
    -- Calculate total amount and count
    SELECT 
        SUM(commission_amount),
        COUNT(*)
    INTO v_total_amount, v_commission_count
    FROM public.affiliate_commissions
    WHERE id = ANY(p_commission_ids)
    AND affiliate_id = p_affiliate_id
    AND status = 'APPROVED';
    
    IF v_total_amount IS NULL OR v_total_amount <= 0 THEN
        RAISE EXCEPTION 'No approved commissions found';
    END IF;
    
    -- Calculate fee (e.g., 1% or ₦100, whichever is higher)
    v_payout_fee := GREATEST(v_total_amount * 0.01, 100);
    v_net_amount := v_total_amount - v_payout_fee;
    
    -- Create payout record
    INSERT INTO public.affiliate_payouts (
        affiliate_id,
        total_amount,
        commission_count,
        commission_ids,
        bank_name,
        account_number,
        account_name,
        payout_fee,
        net_amount
    ) VALUES (
        p_affiliate_id,
        v_total_amount,
        v_commission_count,
        to_jsonb(p_commission_ids),
        p_bank_name,
        p_account_number,
        p_account_name,
        v_payout_fee,
        v_net_amount
    ) RETURNING id INTO v_payout_id;
    
    -- Update commission status
    UPDATE public.affiliate_commissions
    SET status = 'PAID',
        paid_at = NOW()
    WHERE id = ANY(p_commission_ids);
    
    RETURN v_payout_id;
END;
$$;


--
-- Name: create_notification_from_template(uuid, text, jsonb, uuid, text, jsonb); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.create_notification_from_template(p_user_id uuid, p_template_key text, p_variables jsonb DEFAULT '{}'::jsonb, p_reference_id uuid DEFAULT NULL::uuid, p_reference_type text DEFAULT NULL::text, p_channels jsonb DEFAULT '["in_app"]'::jsonb) RETURNS uuid
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
DECLARE
    template_record RECORD;
    notification_id UUID;
    processed_title TEXT;
    processed_body TEXT;
    var_key TEXT;
    var_value TEXT;
BEGIN
    -- Get template
    SELECT * INTO template_record
    FROM public.notification_templates
    WHERE template_key = p_template_key AND is_active = true;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Template not found: %', p_template_key;
    END IF;
    
    -- Process template variables
    processed_title := template_record.title_template;
    processed_body := template_record.body_template;
    
    -- Replace variables in title and body
    FOR var_key IN SELECT jsonb_object_keys(p_variables) LOOP
        var_value := p_variables ->> var_key;
        processed_title := REPLACE(processed_title, '{{' || var_key || '}}', var_value);
        processed_body := REPLACE(processed_body, '{{' || var_key || '}}', var_value);
    END LOOP;
    
    -- Create notification
    notification_id := uuid_generate_v4();
    
    INSERT INTO public.user_notifications (
        id, user_id, template_id, title, body, notification_type,
        reference_id, reference_type, channels, priority
    ) VALUES (
        notification_id,
        p_user_id,
        template_record.id,
        processed_title,
        processed_body,
        COALESCE(p_reference_type, 'system'),
        p_reference_id,
        p_reference_type,
        p_channels,
        template_record.priority
    );
    
    RETURN notification_id;
END;
$$;


--
-- Name: ensure_single_primary_bank_account(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.ensure_single_primary_bank_account() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.is_primary = true THEN
        UPDATE public.affiliate_bank_accounts
        SET is_primary = false
        WHERE affiliate_id = NEW.affiliate_id
        AND id != NEW.id;
    END IF;
    RETURN NEW;
END;
$$;


--
-- Name: generate_otp_code(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.generate_otp_code(length integer DEFAULT 6) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
    otp_code TEXT;
    i INTEGER;
BEGIN
    otp_code := '';
    FOR i IN 1..length LOOP
        otp_code := otp_code || floor(random() * 10)::TEXT;
    END LOOP;
    RETURN otp_code;
END;
$$;


--
-- Name: generate_referral_code(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.generate_referral_code(user_name text) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
    base_code TEXT;
    final_code TEXT;
    counter INTEGER := 0;
BEGIN
    -- Create base code from user name (first 4 chars + year)
    base_code := UPPER(LEFT(REGEXP_REPLACE(user_name, '[^a-zA-Z]', '', 'g'), 4)) || '2026';
    final_code := base_code;
    
    -- Ensure uniqueness
    WHILE EXISTS (SELECT 1 FROM public.users WHERE referral_code = final_code) LOOP
        counter := counter + 1;
        final_code := base_code || counter::TEXT;
    END LOOP;
    
    RETURN final_code;
END;
$$;


--
-- Name: get_active_provider(character varying, character varying); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_active_provider(p_network character varying, p_service_type character varying) RETURNS TABLE(id bigint, network character varying, service_type character varying, provider_mode character varying, provider_name character varying, priority integer, config jsonb)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pc.id,
        pc.network,
        pc.service_type,
        pc.provider_mode,
        pc.provider_name,
        pc.priority,
        pc.config
    FROM provider_configs pc
    WHERE pc.network = p_network 
      AND pc.service_type = p_service_type
      AND pc.is_active = true
    ORDER BY pc.priority ASC
    LIMIT 1;
END;
$$;


--
-- Name: get_admin_activity_summary(uuid, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_admin_activity_summary(p_admin_user_id uuid, p_days integer DEFAULT 30) RETURNS TABLE(total_actions bigint, suspicious_actions bigint, avg_risk_score numeric, most_common_action text, action_count bigint)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*)::BIGINT as total_actions,
        COUNT(*) FILTER (WHERE is_suspicious = true)::BIGINT as suspicious_actions,
        AVG(risk_score)::NUMERIC as avg_risk_score,
        MODE() WITHIN GROUP (ORDER BY action) as most_common_action,
        COUNT(*) FILTER (WHERE action = MODE() WITHIN GROUP (ORDER BY action))::BIGINT as action_count
    FROM public.admin_activity_logs
    WHERE admin_user_id = p_admin_user_id
    AND created_at > NOW() - (p_days || ' days')::INTERVAL;
END;
$$;


--
-- Name: get_affiliate_payout_stats(uuid); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_affiliate_payout_stats(p_affiliate_id uuid) RETURNS TABLE(total_payouts bigint, total_paid numeric, pending_amount numeric, last_payout_date timestamp with time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*)::BIGINT as total_payouts,
        COALESCE(SUM(net_amount) FILTER (WHERE payout_status = 'COMPLETED'), 0) as total_paid,
        COALESCE(SUM(net_amount) FILTER (WHERE payout_status = 'PENDING'), 0) as pending_amount,
        MAX(processed_at) FILTER (WHERE payout_status = 'COMPLETED') as last_payout_date
    FROM public.affiliate_payouts
    WHERE affiliate_id = p_affiliate_id;
END;
$$;


--
-- Name: get_affiliate_performance(uuid, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_affiliate_performance(p_affiliate_id uuid, p_days integer DEFAULT 30) RETURNS TABLE(total_clicks bigint, total_conversions bigint, avg_conversion_rate numeric, total_earned numeric, best_day date, best_day_commission numeric)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        SUM(a.total_clicks)::BIGINT,
        SUM(a.conversions)::BIGINT,
        AVG(a.conversion_rate),
        SUM(a.total_commission),
        (SELECT analytics_date FROM public.affiliate_analytics 
         WHERE affiliate_id = p_affiliate_id 
         ORDER BY total_commission DESC LIMIT 1),
        (SELECT total_commission FROM public.affiliate_analytics 
         WHERE affiliate_id = p_affiliate_id 
         ORDER BY total_commission DESC LIMIT 1)
    FROM public.affiliate_analytics a
    WHERE a.affiliate_id = p_affiliate_id
    AND a.analytics_date > CURRENT_DATE - p_days;
END;
$$;


--
-- Name: get_eligible_wheel_prizes(numeric); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_eligible_wheel_prizes(recharge_amount numeric) RETURNS TABLE(prize_id uuid, prize_name text, prize_type text, prize_value numeric, probability numeric, icon_name text, color_scheme text)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        wp.id,
        wp.prize_name,
        wp.prize_type,
        wp.prize_value,
        wp.probability,
        wp.icon_name,
        wp.color_scheme
    FROM public.wheel_prizes wp
    WHERE wp.is_active = true 
    AND wp.minimum_recharge <= recharge_amount
    ORDER BY wp.sort_order;
END;
$$;


--
-- Name: get_payment_error_stats(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_payment_error_stats(p_hours integer DEFAULT 24) RETURNS TABLE(total_requests bigint, failed_requests bigint, error_rate numeric, most_common_error text, error_count bigint)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*)::BIGINT as total_requests,
        COUNT(*) FILTER (WHERE is_successful = false)::BIGINT as failed_requests,
        (COUNT(*) FILTER (WHERE is_successful = false)::DECIMAL / NULLIF(COUNT(*), 0) * 100) as error_rate,
        MODE() WITHIN GROUP (ORDER BY error_code) FILTER (WHERE is_successful = false) as most_common_error,
        COUNT(*) FILTER (WHERE error_code = MODE() WITHIN GROUP (ORDER BY error_code) FILTER (WHERE is_successful = false))::BIGINT as error_count
    FROM public.payment_logs
    WHERE created_at > NOW() - (p_hours || ' hours')::INTERVAL;
END;
$$;


--
-- Name: get_payment_event_history(uuid, text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_payment_event_history(p_transaction_id uuid DEFAULT NULL::uuid, p_payment_reference text DEFAULT NULL::text) RETURNS TABLE(id uuid, event_type text, status_code integer, is_successful boolean, error_message text, created_at timestamp with time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pl.id,
        pl.event_type,
        pl.status_code,
        pl.is_successful,
        pl.error_message,
        pl.created_at
    FROM public.payment_logs pl
    WHERE (p_transaction_id IS NULL OR pl.transaction_id = p_transaction_id)
    AND (p_payment_reference IS NULL OR pl.payment_reference = p_payment_reference)
    ORDER BY pl.created_at ASC;
END;
$$;


--
-- Name: get_platform_statistics(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_platform_statistics() RETURNS TABLE(total_users integer, total_transactions integer, total_revenue numeric, active_affiliates integer, pending_prizes integer, active_draws integer)
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        (SELECT COUNT(*)::INTEGER FROM public.users WHERE is_active = true),
        (SELECT COUNT(*)::INTEGER FROM public.transactions WHERE status = 'SUCCESS'),
        (SELECT COALESCE(SUM(amount), 0) FROM public.transactions WHERE status = 'SUCCESS'),
        (SELECT COUNT(*)::INTEGER FROM public.affiliates WHERE status = 'APPROVED'),
        (SELECT COUNT(*)::INTEGER FROM public.spin_results WHERE claim_status = 'PENDING'),
        (SELECT COUNT(*)::INTEGER FROM public.draws WHERE status = 'ACTIVE');
END;
$$;


--
-- Name: get_slow_payment_requests(integer, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_slow_payment_requests(p_threshold_ms integer DEFAULT 5000, p_hours integer DEFAULT 24) RETURNS TABLE(id uuid, event_type text, payment_reference text, response_time_ms integer, created_at timestamp with time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pl.id,
        pl.event_type,
        pl.payment_reference,
        pl.response_time_ms,
        pl.created_at
    FROM public.payment_logs pl
    WHERE pl.response_time_ms > p_threshold_ms
    AND pl.created_at > NOW() - (p_hours || ' hours')::INTERVAL
    ORDER BY pl.response_time_ms DESC;
END;
$$;


--
-- Name: get_transaction_limit(character varying, character varying, character varying); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_transaction_limit(p_limit_type character varying, p_limit_scope character varying, p_user_tier character varying DEFAULT NULL::character varying) RETURNS TABLE(min_amount bigint, max_amount bigint, daily_limit bigint, monthly_limit bigint)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        tl.min_amount,
        tl.max_amount,
        tl.daily_limit,
        tl.monthly_limit
    FROM transaction_limits tl
    WHERE tl.limit_type = p_limit_type
      AND tl.limit_scope = p_limit_scope
      AND tl.is_active = true
      AND (tl.applies_to_user_tier = p_user_tier OR tl.applies_to_user_tier IS NULL)
    ORDER BY 
        CASE WHEN tl.applies_to_user_tier IS NOT NULL THEN 1 ELSE 2 END, -- Prioritize tier-specific limits
        tl.created_at DESC
    LIMIT 1;
END;
$$;


--
-- Name: get_unread_notification_count(uuid); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_unread_notification_count(p_user_id uuid DEFAULT NULL::uuid) RETURNS integer
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
DECLARE
    target_user_id UUID;
BEGIN
    -- Use provided user_id or get from auth
    target_user_id := COALESCE(
        p_user_id,
        NULL
    );
    
    RETURN (
        SELECT COUNT(*)::INTEGER
        FROM public.user_notifications
        WHERE user_id = target_user_id
        AND is_read = false
        AND (expires_at IS NULL OR expires_at > NOW())
    );
END;
$$;


--
-- Name: get_user_activity_summary(uuid); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_user_activity_summary(p_user_id uuid) RETURNS TABLE(total_recharge numeric, total_points integer, loyalty_tier text, pending_prizes integer, draw_entries integer, affiliate_earnings numeric)
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        u.total_recharge_amount,
        u.total_points,
        u.loyalty_tier,
        (SELECT COUNT(*)::INTEGER FROM public.spin_results WHERE user_id = p_user_id AND claim_status = 'PENDING'),
        (SELECT COALESCE(SUM(entries_count), 0)::INTEGER FROM public.draw_entries WHERE user_id = p_user_id),
        (SELECT COALESCE(SUM(commission_amount), 0) FROM public.affiliate_commissions ac 
         JOIN public.affiliates a ON a.id = ac.affiliate_id 
         WHERE a.user_id = p_user_id AND ac.status = 'APPROVED')
    FROM public.users u
    WHERE u.id = p_user_id;
END;
$$;


--
-- Name: log_admin_action(uuid, uuid, text, text, text, text, text, jsonb, integer, inet, text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.log_admin_action(p_admin_user_id uuid, p_session_id uuid, p_action text, p_resource text, p_resource_id text, p_method text DEFAULT NULL::text, p_endpoint text DEFAULT NULL::text, p_request_data jsonb DEFAULT NULL::jsonb, p_response_status integer DEFAULT NULL::integer, p_ip_address inet DEFAULT NULL::inet, p_user_agent text DEFAULT NULL::text) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_log_id UUID;
BEGIN
    INSERT INTO public.admin_activity_logs (
        admin_user_id,
        admin_session_id,
        action,
        resource,
        resource_id,
        method,
        endpoint,
        request_data,
        response_status,
        ip_address,
        user_agent
    ) VALUES (
        p_admin_user_id,
        p_session_id,
        p_action,
        p_resource,
        p_resource_id,
        p_method,
        p_endpoint,
        p_request_data,
        p_response_status,
        p_ip_address,
        p_user_agent
    ) RETURNING id INTO v_log_id;
    
    RETURN v_log_id;
END;
$$;


--
-- Name: log_payment_event(uuid, uuid, text, text, jsonb, jsonb, integer, text, inet, text, numeric, boolean); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.log_payment_event(p_transaction_id uuid, p_user_id uuid, p_event_type text, p_payment_reference text, p_request_payload jsonb DEFAULT NULL::jsonb, p_response_payload jsonb DEFAULT NULL::jsonb, p_status_code integer DEFAULT NULL::integer, p_error_message text DEFAULT NULL::text, p_ip_address inet DEFAULT NULL::inet, p_user_agent text DEFAULT NULL::text, p_amount numeric DEFAULT NULL::numeric, p_is_successful boolean DEFAULT NULL::boolean) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_log_id UUID;
BEGIN
    INSERT INTO public.payment_logs (
        transaction_id,
        user_id,
        event_type,
        payment_reference,
        request_payload,
        response_payload,
        status_code,
        error_message,
        ip_address,
        user_agent,
        amount,
        is_successful
    ) VALUES (
        p_transaction_id,
        p_user_id,
        p_event_type,
        p_payment_reference,
        p_request_payload,
        p_response_payload,
        p_status_code,
        p_error_message,
        p_ip_address,
        p_user_agent,
        p_amount,
        p_is_successful
    ) RETURNING id INTO v_log_id;
    
    RETURN v_log_id;
END;
$$;


--
-- Name: log_provider_transaction(bigint, bigint, character varying, character varying, jsonb, jsonb, character varying, text, bigint); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.log_provider_transaction(p_transaction_id bigint, p_provider_config_id bigint, p_provider_mode character varying, p_provider_name character varying, p_request_payload jsonb, p_response_payload jsonb, p_status character varying, p_message text, p_response_time_ms bigint) RETURNS bigint
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_log_id BIGINT;
BEGIN
    -- Note: This function is a placeholder for future provider_transaction_logs table
    -- For now, return a dummy ID
    -- TODO: Create provider_transaction_logs table in future migration
    RETURN 1;
END;
$$;


--
-- Name: mark_expired_otps(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.mark_expired_otps() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.expires_at < NOW() AND NEW.is_expired = false THEN
        NEW.is_expired := true;
    END IF;
    RETURN NEW;
END;
$$;


--
-- Name: mark_notification_read(uuid); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.mark_notification_read(p_notification_id uuid) RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
BEGIN
    UPDATE public.user_notifications
    SET 
        is_read = true,
        read_at = NOW(),
        updated_at = NOW()
    WHERE id = p_notification_id
    AND true;
END;
$$;


--
-- Name: process_affiliate_commission(uuid, text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.process_affiliate_commission(p_transaction_id uuid, p_affiliate_code text DEFAULT NULL::text) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    transaction_record RECORD;
    affiliate_record RECORD;
    commission_amount DECIMAL;
    commission_rate DECIMAL;
BEGIN
    -- Get transaction details
    SELECT * INTO transaction_record
    FROM public.transactions
    WHERE id = p_transaction_id AND status = 'SUCCESS';
    
    IF NOT FOUND THEN
        RETURN; -- Transaction not found or not successful
    END IF;
    
    -- Find affiliate (from parameter or user's referrer)
    IF p_affiliate_code IS NOT NULL THEN
        SELECT * INTO affiliate_record
        FROM public.affiliates
        WHERE affiliate_code = p_affiliate_code AND status = 'APPROVED';
    ELSIF transaction_record.user_id IS NOT NULL THEN
        SELECT a.* INTO affiliate_record
        FROM public.affiliates a
        JOIN public.users u ON u.referred_by = a.user_id
        WHERE u.id = transaction_record.user_id AND a.status = 'APPROVED';
    END IF;
    
    IF NOT FOUND THEN
        RETURN; -- No eligible affiliate found
    END IF;
    
    -- Calculate commission
    commission_rate := affiliate_record.commission_rate / 100.0;
    commission_amount := transaction_record.amount * commission_rate;
    
    -- Create commission record
    INSERT INTO public.affiliate_commissions (
        id, affiliate_id, transaction_id, commission_amount, 
        commission_rate, status, earned_at
    ) VALUES (
        uuid_generate_v4(),
        affiliate_record.id,
        p_transaction_id,
        commission_amount,
        affiliate_record.commission_rate,
        'PENDING',
        NOW()
    );
    
    -- Update affiliate statistics
    UPDATE public.affiliates
    SET 
        total_commission = total_commission + commission_amount,
        updated_at = NOW()
    WHERE id = affiliate_record.id;
END;
$$;


--
-- Name: select_wheel_prize(numeric); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.select_wheel_prize(recharge_amount numeric) RETURNS TABLE(prize_id uuid, prize_name text, prize_type text, prize_value numeric)
    LANGUAGE plpgsql
    AS $$
DECLARE
    total_probability DECIMAL := 0;
    random_value DECIMAL;
    cumulative_probability DECIMAL := 0;
    prize_record RECORD;
BEGIN
    -- Calculate total probability for eligible prizes
    SELECT SUM(wp.probability) INTO total_probability
    FROM public.wheel_prizes wp
    WHERE wp.is_active = true AND wp.minimum_recharge <= recharge_amount;
    
    -- Generate random value
    random_value := RANDOM() * total_probability;
    
    -- Select prize based on probability
    FOR prize_record IN 
        SELECT wp.id, wp.prize_name, wp.prize_type, wp.prize_value, wp.probability
        FROM public.wheel_prizes wp
        WHERE wp.is_active = true AND wp.minimum_recharge <= recharge_amount
        ORDER BY wp.sort_order
    LOOP
        cumulative_probability := cumulative_probability + prize_record.probability;
        IF random_value <= cumulative_probability THEN
            prize_id := prize_record.id;
            prize_name := prize_record.prize_name;
            prize_type := prize_record.prize_type;
            prize_value := prize_record.prize_value;
            RETURN NEXT;
            RETURN;
        END IF;
    END LOOP;
    
    -- Fallback to first prize if no selection made
    SELECT wp.id, wp.prize_name, wp.prize_type, wp.prize_value
    INTO prize_id, prize_name, prize_type, prize_value
    FROM public.wheel_prizes wp
    WHERE wp.is_active = true AND wp.minimum_recharge <= recharge_amount
    ORDER BY wp.sort_order
    LIMIT 1;
    
    RETURN NEXT;
END;
$$;


--
-- Name: trigger_update_draw_entries(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.trigger_update_draw_entries() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Update total entries for the draw
    UPDATE public.draws
    SET 
        total_entries = (
            SELECT COALESCE(SUM(entries_count), 0)
            FROM public.draw_entries
            WHERE draw_id = COALESCE(NEW.draw_id, OLD.draw_id)
        ),
        updated_at = NOW()
    WHERE id = COALESCE(NEW.draw_id, OLD.draw_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$;


--
-- Name: update_affiliate_analytics_timestamp(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_affiliate_analytics_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


--
-- Name: update_affiliate_bank_account_timestamp(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_affiliate_bank_account_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


--
-- Name: update_affiliate_daily_analytics(uuid, date); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_affiliate_daily_analytics(p_affiliate_id uuid, p_date date DEFAULT CURRENT_DATE) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_total_clicks INTEGER;
    v_unique_clicks INTEGER;
    v_conversions INTEGER;
    v_conversion_rate DECIMAL(5,2);
    v_total_commission DECIMAL(12,2);
    v_recharge_commissions DECIMAL(12,2);
    v_subscription_commissions DECIMAL(12,2);
BEGIN
    -- Calculate metrics
    SELECT 
        COUNT(*),
        COUNT(DISTINCT ip_address),
        COUNT(*) FILTER (WHERE converted = true)
    INTO v_total_clicks, v_unique_clicks, v_conversions
    FROM public.affiliate_clicks
    WHERE affiliate_id = p_affiliate_id
    AND DATE(created_at) = p_date;
    
    -- Calculate conversion rate
    v_conversion_rate := CASE 
        WHEN v_total_clicks > 0 THEN (v_conversions::DECIMAL / v_total_clicks) * 100
        ELSE 0
    END;
    
    -- Calculate commissions
    SELECT 
        COALESCE(SUM(commission_amount), 0),
        COALESCE(SUM(commission_amount) FILTER (WHERE transaction_type = 'RECHARGE'), 0),
        COALESCE(SUM(commission_amount) FILTER (WHERE transaction_type = 'SUBSCRIPTION'), 0)
    INTO v_total_commission, v_recharge_commissions, v_subscription_commissions
    FROM public.affiliate_commissions
    WHERE affiliate_id = p_affiliate_id
    AND DATE(created_at) = p_date;
    
    -- Insert or update analytics
    INSERT INTO public.affiliate_analytics (
        affiliate_id,
        analytics_date,
        total_clicks,
        unique_clicks,
        conversions,
        conversion_rate,
        total_commission,
        recharge_commissions,
        subscription_commissions
    ) VALUES (
        p_affiliate_id,
        p_date,
        v_total_clicks,
        v_unique_clicks,
        v_conversions,
        v_conversion_rate,
        v_total_commission,
        v_recharge_commissions,
        v_subscription_commissions
    )
    ON CONFLICT (affiliate_id, analytics_date)
    DO UPDATE SET
        total_clicks = EXCLUDED.total_clicks,
        unique_clicks = EXCLUDED.unique_clicks,
        conversions = EXCLUDED.conversions,
        conversion_rate = EXCLUDED.conversion_rate,
        total_commission = EXCLUDED.total_commission,
        recharge_commissions = EXCLUDED.recharge_commissions,
        subscription_commissions = EXCLUDED.subscription_commissions,
        updated_at = NOW();
END;
$$;


--
-- Name: update_affiliate_payout_timestamp(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_affiliate_payout_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


--
-- Name: update_otps_updated_at(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_otps_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


--
-- Name: update_spin_tiers_updated_at(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_spin_tiers_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


--
-- Name: update_transaction_limits_timestamp(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_transaction_limits_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


--
-- Name: update_user_statistics(uuid); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_user_statistics(p_user_id uuid) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    total_amount DECIMAL;
    total_count INTEGER;
    last_date TIMESTAMP WITH TIME ZONE;
    new_tier TEXT;
BEGIN
    -- Calculate statistics from successful transactions
    SELECT 
        COALESCE(SUM(amount), 0),
        COUNT(*),
        MAX(completed_at)
    INTO total_amount, total_count, last_date
    FROM public.transactions
    WHERE user_id = p_user_id AND status = 'SUCCESS';
    
    -- Calculate new loyalty tier
    new_tier := calculate_loyalty_tier(total_amount);
    
    -- Update user record
    UPDATE public.users
    SET 
        total_recharge_amount = total_amount,
        total_transactions = total_count,
        last_recharge_date = last_date,
        loyalty_tier = new_tier,
        updated_at = NOW()
    WHERE id = p_user_id;
END;
$$;


--
-- Name: upsert_user_profile(uuid, text, text, text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.upsert_user_profile(p_auth_user_id uuid, p_msisdn text, p_full_name text DEFAULT NULL::text, p_email text DEFAULT NULL::text) RETURNS uuid
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
DECLARE
    user_id UUID;
    referral_code TEXT;
BEGIN
    -- Check if user already exists
    SELECT id INTO user_id 
    FROM public.users 
    WHERE auth_user_id = p_auth_user_id OR msisdn = p_msisdn;
    
    IF user_id IS NOT NULL THEN
        -- Update existing user
        UPDATE public.users 
        SET 
            full_name = COALESCE(p_full_name, full_name),
            email = COALESCE(p_email, email),
            auth_user_id = COALESCE(p_auth_user_id, auth_user_id),
            updated_at = NOW()
        WHERE id = user_id;
    ELSE
        -- Create new user
        user_id := uuid_generate_v4();
        referral_code := generate_referral_code(COALESCE(p_full_name, p_msisdn));
        
        INSERT INTO public.users (
            id, auth_user_id, msisdn, full_name, email, referral_code
        ) VALUES (
            user_id, p_auth_user_id, p_msisdn, p_full_name, p_email, referral_code
        );
    END IF;
    
    RETURN user_id;
END;
$$;


--
-- Name: verify_otp(text, text, text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.verify_otp(p_msisdn text, p_otp_code text, p_purpose text) RETURNS TABLE(success boolean, message text, otp_id uuid, user_id uuid)
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_otp RECORD;
    v_otp_hash TEXT;
BEGIN
    -- Hash the provided OTP
    v_otp_hash := encode(digest(p_otp_code, 'sha256'), 'hex');
    
    -- Find matching OTP
    SELECT * INTO v_otp
    FROM public.otp_verifications
    WHERE msisdn = p_msisdn
    AND purpose = p_purpose
    AND otp_code_hash = v_otp_hash
    AND is_verified = false
    AND is_expired = false
    AND is_revoked = false
    AND expires_at > NOW()
    ORDER BY created_at DESC
    LIMIT 1;
    
    IF NOT FOUND THEN
        RETURN QUERY SELECT false, 'Invalid or expired OTP'::TEXT, NULL::UUID, NULL::UUID;
        RETURN;
    END IF;
    
    IF v_otp.attempts >= v_otp.max_attempts THEN
        RETURN QUERY SELECT false, 'Maximum attempts exceeded'::TEXT, v_otp.id, v_otp.user_id;
        RETURN;
    END IF;
    
    -- Mark as verified
    UPDATE public.otp_verifications
    SET is_verified = true,
        verified_at = NOW(),
        attempts = attempts + 1
    WHERE id = v_otp.id;
    
    RETURN QUERY SELECT true, 'OTP verified successfully'::TEXT, v_otp.id, v_otp.user_id;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admin_activity_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.admin_activity_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    admin_user_id uuid,
    admin_session_id uuid,
    action text NOT NULL,
    resource text,
    resource_id text,
    method text,
    endpoint text,
    request_data jsonb,
    response_status integer,
    response_data jsonb,
    ip_address inet,
    user_agent text,
    duration_ms integer,
    is_suspicious boolean DEFAULT false,
    risk_score integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    admin_email character varying(255),
    action_type character varying(50),
    resource_type character varying(50),
    details jsonb,
    CONSTRAINT admin_activity_logs_risk_score_check CHECK (((risk_score >= 0) AND (risk_score <= 100)))
);


--
-- Name: admin_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.admin_sessions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    admin_user_id uuid,
    session_token text NOT NULL,
    ip_address inet,
    user_agent text,
    is_active boolean DEFAULT true,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    last_accessed_at timestamp with time zone DEFAULT now()
);


--
-- Name: admin_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.admin_users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    full_name text NOT NULL,
    role text DEFAULT 'ADMIN'::text,
    permissions jsonb DEFAULT '[]'::jsonb,
    is_active boolean DEFAULT true,
    last_login_at timestamp with time zone,
    login_attempts integer DEFAULT 0,
    locked_until timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT admin_users_role_check CHECK ((role = ANY (ARRAY['SUPER_ADMIN'::text, 'ADMIN'::text, 'MODERATOR'::text, 'SUPPORT'::text]))),
    CONSTRAINT valid_admin_email CHECK ((email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'::text))
);


--
-- Name: affiliate_analytics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.affiliate_analytics (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    affiliate_id uuid,
    analytics_date date NOT NULL,
    total_clicks integer DEFAULT 0,
    unique_clicks integer DEFAULT 0,
    conversions integer DEFAULT 0,
    conversion_rate numeric(5,2) DEFAULT 0.00,
    total_commission numeric(12,2) DEFAULT 0.00,
    recharge_commissions numeric(12,2) DEFAULT 0.00,
    subscription_commissions numeric(12,2) DEFAULT 0.00,
    top_referrer_country text,
    top_device_type text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: affiliate_bank_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.affiliate_bank_accounts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    affiliate_id uuid,
    bank_name text NOT NULL,
    account_number text NOT NULL,
    account_name text NOT NULL,
    is_verified boolean DEFAULT false,
    is_primary boolean DEFAULT false,
    verified_at timestamp with time zone,
    verified_by uuid,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT valid_account_number CHECK ((length(account_number) >= 10))
);


--
-- Name: affiliate_clicks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.affiliate_clicks (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    affiliate_id uuid,
    ip_address inet,
    user_agent text,
    referrer_url text,
    landing_page text,
    converted boolean DEFAULT false,
    conversion_transaction_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    converted_at timestamp with time zone
);


--
-- Name: affiliate_commissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.affiliate_commissions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    affiliate_id uuid,
    transaction_id uuid,
    commission_amount numeric(10,2) NOT NULL,
    commission_rate numeric(5,2) NOT NULL,
    transaction_amount numeric(10,2) NOT NULL,
    status text DEFAULT 'PENDING'::text,
    payout_reference text,
    payout_method text,
    created_at timestamp with time zone DEFAULT now(),
    earned_at timestamp with time zone DEFAULT now(),
    paid_at timestamp with time zone,
    CONSTRAINT affiliate_commissions_status_check CHECK ((status = ANY (ARRAY['PENDING'::text, 'APPROVED'::text, 'PAID'::text, 'CANCELLED'::text]))),
    CONSTRAINT positive_commission_amount CHECK ((commission_amount > (0)::numeric)),
    CONSTRAINT positive_transaction_amount CHECK ((transaction_amount > (0)::numeric))
);


--
-- Name: affiliate_payouts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.affiliate_payouts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    affiliate_id uuid,
    payout_batch_id uuid DEFAULT public.uuid_generate_v4(),
    total_amount numeric(12,2) NOT NULL,
    commission_count integer DEFAULT 0 NOT NULL,
    commission_ids jsonb DEFAULT '[]'::jsonb,
    payout_method text DEFAULT 'BANK_TRANSFER'::text,
    bank_name text,
    account_number text,
    account_name text,
    payout_status text DEFAULT 'PENDING'::text,
    payout_reference text,
    payout_fee numeric(12,2) DEFAULT 0.00,
    net_amount numeric(12,2) NOT NULL,
    processed_at timestamp with time zone,
    processed_by uuid,
    failure_reason text,
    notes text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT affiliate_payouts_payout_method_check CHECK ((payout_method = ANY (ARRAY['BANK_TRANSFER'::text, 'MOBILE_MONEY'::text, 'WALLET'::text]))),
    CONSTRAINT affiliate_payouts_payout_status_check CHECK ((payout_status = ANY (ARRAY['PENDING'::text, 'PROCESSING'::text, 'COMPLETED'::text, 'FAILED'::text, 'CANCELLED'::text]))),
    CONSTRAINT affiliate_payouts_total_amount_check CHECK ((total_amount > (0)::numeric))
);


--
-- Name: affiliates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.affiliates (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    affiliate_code text NOT NULL,
    status text DEFAULT 'PENDING'::text,
    tier text DEFAULT 'BRONZE'::text,
    commission_rate numeric(5,2) DEFAULT 5.00,
    total_referrals integer DEFAULT 0,
    active_referrals integer DEFAULT 0,
    total_commission numeric(10,2) DEFAULT 0,
    business_name text,
    website_url text,
    social_media_handles jsonb,
    bank_name text,
    account_number text,
    account_name text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    approved_at timestamp with time zone,
    CONSTRAINT affiliates_status_check CHECK ((status = ANY (ARRAY['PENDING'::text, 'APPROVED'::text, 'SUSPENDED'::text, 'REJECTED'::text]))),
    CONSTRAINT affiliates_tier_check CHECK ((tier = ANY (ARRAY['BRONZE'::text, 'SILVER'::text, 'GOLD'::text, 'PLATINUM'::text]))),
    CONSTRAINT positive_commission_rate CHECK ((commission_rate >= (0)::numeric))
);


--
-- Name: application_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.application_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    level text NOT NULL,
    message text NOT NULL,
    context jsonb,
    user_id uuid,
    ip_address inet,
    user_agent text,
    request_id text,
    error_code text,
    stack_trace text,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT application_logs_level_check CHECK ((level = ANY (ARRAY['DEBUG'::text, 'INFO'::text, 'WARN'::text, 'ERROR'::text, 'FATAL'::text])))
);


--
-- Name: application_metrics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.application_metrics (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    metric_name text NOT NULL,
    metric_value numeric(15,4) NOT NULL,
    metric_unit text,
    tags jsonb,
    dimensions jsonb,
    recorded_at timestamp with time zone DEFAULT now()
);


--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    admin_user_id uuid,
    user_id uuid,
    action character varying(100) NOT NULL,
    entity_type character varying(100),
    entity_id text,
    old_value jsonb,
    new_value jsonb,
    ip_address inet,
    user_agent text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: daily_subscription_config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.daily_subscription_config (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    amount numeric(5,2) NOT NULL,
    draw_entries_earned integer DEFAULT 1,
    is_paid boolean DEFAULT true,
    description text,
    terms_and_conditions text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT positive_amount CHECK ((amount > (0)::numeric)),
    CONSTRAINT positive_entries CHECK ((draw_entries_earned > 0))
);


--
-- Name: daily_subscriptions; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: data_plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.data_plans (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    network_id uuid,
    plan_name text NOT NULL,
    data_amount text NOT NULL,
    price numeric(10,2) NOT NULL,
    validity_days integer NOT NULL,
    plan_code text NOT NULL,
    is_active boolean DEFAULT true,
    sort_order integer DEFAULT 0,
    description text,
    terms_and_conditions text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    network_provider text,
    CONSTRAINT positive_price CHECK ((price > (0)::numeric)),
    CONSTRAINT positive_validity CHECK ((validity_days > 0))
);


--
-- Name: draw_entries; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: draw_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.draw_types (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(50) NOT NULL,
    description text,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: draw_winners; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: draws; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.draws (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    description text,
    status text DEFAULT 'UPCOMING'::text,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone NOT NULL,
    draw_time timestamp with time zone,
    prize_pool numeric(12,2) NOT NULL,
    winners_count integer DEFAULT 1,
    total_entries integer DEFAULT 0,
    results jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    completed_at timestamp with time zone,
    draw_type_id uuid,
    runner_ups_count integer DEFAULT 1,
    draw_code character varying(20),
    prize_template_id uuid,
    CONSTRAINT draws_status_check CHECK ((status = ANY (ARRAY['UPCOMING'::text, 'ACTIVE'::text, 'COMPLETED'::text, 'CANCELLED'::text]))),
    CONSTRAINT draws_type_check CHECK ((type = ANY (ARRAY['DAILY'::text, 'WEEKLY'::text, 'MONTHLY'::text, 'SPECIAL'::text]))),
    CONSTRAINT positive_prize_pool CHECK ((prize_pool > (0)::numeric)),
    CONSTRAINT positive_winners_count CHECK ((winners_count > 0)),
    CONSTRAINT valid_timing CHECK ((end_time > start_time))
);


--
-- Name: network_cache; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.network_cache (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    msisdn text NOT NULL,
    network text NOT NULL,
    last_verified_at timestamp with time zone DEFAULT now() NOT NULL,
    cache_expires_at timestamp with time zone NOT NULL,
    lookup_source text,
    hlr_provider text,
    hlr_response jsonb,
    is_valid boolean DEFAULT true,
    invalidated_at timestamp with time zone,
    invalidation_reason text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: network_configs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.network_configs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    network_name text NOT NULL,
    network_code text NOT NULL,
    is_active boolean DEFAULT true,
    airtime_enabled boolean DEFAULT true,
    data_enabled boolean DEFAULT true,
    commission_rate numeric(5,2) DEFAULT 2.50,
    minimum_amount numeric(10,2) DEFAULT 50.00,
    maximum_amount numeric(10,2) DEFAULT 50000.00,
    logo_url text,
    brand_color text,
    sort_order integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT positive_commission_rate CHECK ((commission_rate >= (0)::numeric)),
    CONSTRAINT valid_amount_range CHECK ((maximum_amount > minimum_amount))
);


--
-- Name: notification_delivery_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification_delivery_log (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    notification_id uuid,
    channel text NOT NULL,
    delivery_status text NOT NULL,
    provider_name text,
    provider_message_id text,
    provider_response jsonb,
    error_code text,
    error_message text,
    retry_count integer DEFAULT 0,
    attempted_at timestamp with time zone DEFAULT now(),
    delivered_at timestamp with time zone,
    CONSTRAINT valid_channel CHECK ((channel = ANY (ARRAY['push'::text, 'email'::text, 'sms'::text, 'in_app'::text]))),
    CONSTRAINT valid_delivery_status CHECK ((delivery_status = ANY (ARRAY['pending'::text, 'sent'::text, 'delivered'::text, 'failed'::text, 'bounced'::text, 'opened'::text, 'clicked'::text])))
);


--
-- Name: notification_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification_templates (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    template_key text NOT NULL,
    template_name text NOT NULL,
    description text,
    title_template text NOT NULL,
    body_template text NOT NULL,
    email_subject_template text,
    email_body_template text,
    sms_template text,
    variables jsonb DEFAULT '[]'::jsonb,
    supports_push boolean DEFAULT true,
    supports_email boolean DEFAULT true,
    supports_sms boolean DEFAULT false,
    supports_in_app boolean DEFAULT true,
    is_active boolean DEFAULT true,
    priority text DEFAULT 'NORMAL'::text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT notification_templates_priority_check CHECK ((priority = ANY (ARRAY['LOW'::text, 'NORMAL'::text, 'HIGH'::text, 'URGENT'::text]))),
    CONSTRAINT valid_template_key CHECK ((template_key ~ '^[a-z0-9_]+$'::text))
);


--
-- Name: otp_rate_limits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.otp_rate_limits (
    id bigint NOT NULL,
    key character varying(64) NOT NULL,
    requested_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: otp_rate_limits_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.otp_rate_limits_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: otp_rate_limits_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.otp_rate_limits_id_seq OWNED BY public.otp_rate_limits.id;


--
-- Name: otp_verifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.otp_verifications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    msisdn text NOT NULL,
    user_id uuid,
    otp_code_hash text NOT NULL,
    purpose text DEFAULT 'LOGIN'::text NOT NULL,
    is_verified boolean DEFAULT false,
    is_expired boolean DEFAULT false,
    is_revoked boolean DEFAULT false,
    attempts integer DEFAULT 0,
    max_attempts integer DEFAULT 5,
    last_attempt_at timestamp with time zone,
    request_ip inet,
    request_user_agent text,
    device_fingerprint text,
    verified_at timestamp with time zone,
    verified_ip inet,
    verified_user_agent text,
    created_at timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    metadata jsonb DEFAULT '{}'::jsonb,
    CONSTRAINT chk_otp_msisdn_format CHECK ((msisdn ~ '^234[7-9][0-1][0-9]{8}$'::text)),
    CONSTRAINT otp_verifications_purpose_check CHECK ((purpose = ANY (ARRAY['LOGIN'::text, 'REGISTRATION'::text, 'PASSWORD_RESET'::text, 'TRANSACTION_VERIFICATION'::text, 'PHONE_VERIFICATION'::text, 'WITHDRAWAL'::text, 'PROFILE_UPDATE'::text, 'TWO_FACTOR_AUTH'::text]))),
    CONSTRAINT valid_attempts CHECK (((attempts >= 0) AND (attempts <= max_attempts))),
    CONSTRAINT valid_expiry CHECK ((expires_at > created_at)),
    CONSTRAINT valid_msisdn_otp CHECK ((msisdn ~ '^(234|0)?[789][01][0-9]{8,9}$'::text))
);


--
-- Name: otps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.otps (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    msisdn character varying(20) NOT NULL,
    code character varying(6) NOT NULL,
    purpose character varying(50) NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    is_used boolean DEFAULT false,
    used_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: payment_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payment_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    transaction_id uuid,
    user_id uuid,
    event_type text NOT NULL,
    payment_provider text DEFAULT 'PAYSTACK'::text,
    payment_reference text,
    request_payload jsonb,
    response_payload jsonb,
    status_code integer,
    error_message text,
    error_code text,
    ip_address inet,
    user_agent text,
    request_id text,
    response_time_ms integer,
    amount numeric(12,2),
    currency text DEFAULT 'NGN'::text,
    payment_method text,
    is_successful boolean,
    is_retry boolean DEFAULT false,
    retry_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT payment_logs_event_type_check CHECK ((event_type = ANY (ARRAY['INITIALIZE'::text, 'VERIFY'::text, 'CALLBACK'::text, 'WEBHOOK'::text, 'REFUND'::text, 'DISPUTE'::text, 'CHARGEBACK'::text, 'RETRY'::text])))
);


--
-- Name: platform_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.platform_settings (
    setting_key text NOT NULL,
    setting_value text NOT NULL,
    description text,
    is_public boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: points_adjustments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.points_adjustments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    points integer NOT NULL,
    reason character varying(255) NOT NULL,
    description text,
    created_by uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: prize_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.prize_categories (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    template_id uuid,
    draw_id uuid,
    category_name character varying(100) NOT NULL,
    prize_amount numeric(15,2) NOT NULL,
    winners_count integer DEFAULT 1 NOT NULL,
    runner_ups_count integer DEFAULT 1 NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_parent CHECK (((template_id IS NOT NULL) OR (draw_id IS NOT NULL)))
);


--
-- Name: prize_fulfillment_config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.prize_fulfillment_config (
    id integer NOT NULL,
    prize_type character varying(20) NOT NULL,
    fulfillment_mode character varying(20) DEFAULT 'AUTO'::character varying NOT NULL,
    auto_retry_enabled boolean DEFAULT true,
    max_retry_attempts integer DEFAULT 3,
    retry_delay_seconds integer DEFAULT 300,
    fallback_to_manual boolean DEFAULT true,
    fallback_notification_enabled boolean DEFAULT true,
    provision_timeout_seconds integer DEFAULT 60,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    created_by character varying(100),
    updated_by character varying(100),
    CONSTRAINT check_fulfillment_mode CHECK (((fulfillment_mode)::text = ANY ((ARRAY['AUTO'::character varying, 'MANUAL'::character varying])::text[]))),
    CONSTRAINT check_max_retry_attempts CHECK (((max_retry_attempts >= 0) AND (max_retry_attempts <= 10))),
    CONSTRAINT check_prize_type CHECK (((prize_type)::text = ANY ((ARRAY['AIRTIME'::character varying, 'DATA'::character varying, 'CASH'::character varying, 'POINTS'::character varying, 'PHYSICAL'::character varying])::text[]))),
    CONSTRAINT check_retry_delay CHECK (((retry_delay_seconds >= 0) AND (retry_delay_seconds <= 3600))),
    CONSTRAINT check_timeout CHECK (((provision_timeout_seconds >= 10) AND (provision_timeout_seconds <= 300)))
);


--
-- Name: prize_fulfillment_config_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.prize_fulfillment_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: prize_fulfillment_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.prize_fulfillment_config_id_seq OWNED BY public.prize_fulfillment_config.id;


--
-- Name: prize_fulfillment_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.prize_fulfillment_logs (
    id bigint NOT NULL,
    spin_result_id uuid NOT NULL,
    attempt_number integer NOT NULL,
    fulfillment_mode character varying(20) NOT NULL,
    provider_name character varying(50),
    provider_reference character varying(100),
    provider_transaction_id bigint,
    request_payload jsonb,
    response_payload jsonb,
    status character varying(20) NOT NULL,
    error_code character varying(50),
    error_message text,
    response_time_ms integer,
    detected_network character varying(20),
    msisdn character varying(20),
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT check_attempt_number CHECK ((attempt_number > 0)),
    CONSTRAINT check_status CHECK (((status)::text = ANY ((ARRAY['SUCCESS'::character varying, 'FAILED'::character varying, 'PENDING'::character varying, 'TIMEOUT'::character varying, 'CANCELLED'::character varying])::text[])))
);


--
-- Name: prize_fulfillment_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.prize_fulfillment_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: prize_fulfillment_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.prize_fulfillment_logs_id_seq OWNED BY public.prize_fulfillment_logs.id;


--
-- Name: prize_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.prize_templates (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(100) NOT NULL,
    draw_type_id uuid,
    description text,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: provider_configs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.provider_configs (
    id bigint NOT NULL,
    network character varying(50) NOT NULL,
    service_type character varying(50) NOT NULL,
    provider_mode character varying(50) NOT NULL,
    provider_name character varying(100) NOT NULL,
    priority integer DEFAULT 1,
    config jsonb DEFAULT '{}'::jsonb NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: provider_configs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.provider_configs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: provider_configs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.provider_configs_id_seq OWNED BY public.provider_configs.id;


--
-- Name: spin_results; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.spin_results (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    transaction_id uuid,
    msisdn text NOT NULL,
    prize_id uuid,
    prize_name text NOT NULL,
    prize_type text NOT NULL,
    prize_value bigint NOT NULL,
    claim_status text DEFAULT 'PENDING'::text,
    claimed_at timestamp with time zone,
    claim_reference text,
    created_at timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone DEFAULT (now() + '30 days'::interval),
    reviewed_by uuid,
    reviewed_at timestamp without time zone,
    rejection_reason text,
    admin_notes text,
    payment_reference character varying(100),
    bank_account_number text,
    bank_account_name text,
    bank_name text,
    spin_code character varying(30),
    fulfillment_mode character varying(20) DEFAULT 'AUTO'::character varying,
    fulfillment_attempts integer DEFAULT 0,
    last_fulfillment_attempt timestamp without time zone,
    fulfillment_error text,
    can_retry boolean DEFAULT true,
    provision_started_at timestamp without time zone,
    provision_completed_at timestamp without time zone,
    CONSTRAINT check_fulfillment_mode CHECK (((fulfillment_mode)::text = ANY ((ARRAY['AUTO'::character varying, 'MANUAL'::character varying])::text[]))),
    CONSTRAINT chk_spin_results_claim_status CHECK ((claim_status = ANY (ARRAY['PENDING'::text, 'CLAIMED'::text, 'EXPIRED'::text, 'PENDING_ADMIN_REVIEW'::text, 'APPROVED'::text, 'REJECTED'::text]))),
    CONSTRAINT positive_prize_value CHECK (((prize_value)::numeric >= (0)::numeric))
);


--
-- Name: spin_tiers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.spin_tiers (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tier_name text NOT NULL,
    tier_display_name text NOT NULL,
    min_daily_amount bigint NOT NULL,
    max_daily_amount bigint NOT NULL,
    spins_per_day integer NOT NULL,
    tier_color text,
    tier_icon text,
    tier_badge text,
    description text,
    sort_order integer DEFAULT 0,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    CONSTRAINT positive_amounts CHECK (((min_daily_amount >= 0) AND (max_daily_amount > min_daily_amount))),
    CONSTRAINT positive_spins CHECK ((spins_per_day > 0)),
    CONSTRAINT valid_sort_order CHECK ((sort_order >= 0))
);


--
-- Name: subscription_tiers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subscription_tiers (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    description text,
    entries integer DEFAULT 1 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT subscription_tiers_entries_check CHECK ((entries > 0))
);


--
-- Name: transaction_limits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transaction_limits (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    limit_type character varying(50) NOT NULL,
    limit_scope character varying(50) NOT NULL,
    min_amount bigint DEFAULT 10000 NOT NULL,
    max_amount bigint DEFAULT 10000000 NOT NULL,
    daily_limit bigint,
    monthly_limit bigint,
    is_active boolean DEFAULT true NOT NULL,
    applies_to_user_tier character varying(50),
    description text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by uuid,
    updated_by uuid,
    CONSTRAINT positive_amounts CHECK (((min_amount > 0) AND (max_amount > 0))),
    CONSTRAINT valid_amount_range CHECK ((min_amount <= max_amount))
);


--
-- Name: transaction_limits_audit; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transaction_limits_audit (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    limit_id uuid NOT NULL,
    action character varying(20) NOT NULL,
    old_values jsonb,
    new_values jsonb,
    changed_by uuid,
    changed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ip_address character varying(45),
    user_agent text,
    reason text
);


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    msisdn text NOT NULL,
    network_provider text NOT NULL,
    recharge_type text NOT NULL,
    amount bigint NOT NULL,
    data_plan_id uuid,
    payment_method text NOT NULL,
    payment_reference text,
    payment_gateway text,
    status text DEFAULT 'PENDING'::text,
    provider_reference text,
    provider_response jsonb,
    failure_reason text,
    points_earned integer DEFAULT 0,
    draw_entries integer DEFAULT 0,
    spin_eligible boolean DEFAULT false,
    customer_email text,
    customer_name text,
    ip_address inet,
    user_agent text,
    affiliate_code text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    completed_at timestamp with time zone,
    transaction_code character varying(30) NOT NULL,
    CONSTRAINT positive_amount CHECK (((amount)::numeric > (0)::numeric)),
    CONSTRAINT transactions_payment_method_check CHECK ((payment_method = ANY (ARRAY['CARD'::text, 'BANK_TRANSFER'::text, 'USSD'::text, 'WALLET'::text]))),
    CONSTRAINT transactions_recharge_type_check CHECK ((recharge_type = ANY (ARRAY['AIRTIME'::text, 'DATA'::text]))),
    CONSTRAINT transactions_status_check CHECK ((status = ANY (ARRAY['PENDING'::text, 'PROCESSING'::text, 'SUCCESS'::text, 'FAILED'::text, 'CANCELLED'::text]))),
    CONSTRAINT valid_msisdn CHECK ((msisdn ~ '^234[789][01][0-9]{8}$'::text))
);


--
-- Name: user_notification_preferences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_notification_preferences (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    transaction_notifications jsonb DEFAULT '{"sms": false, "push": true, "email": true}'::jsonb,
    prize_notifications jsonb DEFAULT '{"sms": true, "push": true, "email": true}'::jsonb,
    draw_notifications jsonb DEFAULT '{"sms": false, "push": true, "email": true}'::jsonb,
    affiliate_notifications jsonb DEFAULT '{"sms": false, "push": true, "email": true}'::jsonb,
    promotional_notifications jsonb DEFAULT '{"sms": false, "push": true, "email": false}'::jsonb,
    security_notifications jsonb DEFAULT '{"sms": true, "push": true, "email": true}'::jsonb,
    do_not_disturb_start time without time zone,
    do_not_disturb_end time without time zone,
    timezone text DEFAULT 'Africa/Lagos'::text,
    preferred_language text DEFAULT 'en'::text,
    email_frequency text DEFAULT 'immediate'::text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT user_notification_preferences_email_frequency_check CHECK ((email_frequency = ANY (ARRAY['immediate'::text, 'daily'::text, 'weekly'::text, 'never'::text])))
);


--
-- Name: user_notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_notifications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    template_id uuid,
    title text NOT NULL,
    body text NOT NULL,
    notification_type text NOT NULL,
    reference_id uuid,
    reference_type text,
    channels jsonb DEFAULT '["in_app"]'::jsonb,
    is_read boolean DEFAULT false,
    read_at timestamp with time zone,
    delivery_status jsonb DEFAULT '{}'::jsonb,
    delivery_attempts integer DEFAULT 0,
    last_delivery_attempt timestamp with time zone,
    priority text DEFAULT 'NORMAL'::text,
    scheduled_for timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT user_notifications_priority_check CHECK ((priority = ANY (ARRAY['LOW'::text, 'NORMAL'::text, 'HIGH'::text, 'URGENT'::text]))),
    CONSTRAINT valid_notification_type CHECK ((notification_type = ANY (ARRAY['transaction'::text, 'prize'::text, 'draw'::text, 'affiliate'::text, 'system'::text, 'promotional'::text, 'security'::text])))
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    auth_user_id uuid,
    msisdn text NOT NULL,
    full_name text,
    email text,
    phone_verified boolean DEFAULT false,
    email_verified boolean DEFAULT false,
    date_of_birth date,
    gender text,
    state text,
    city text,
    address text,
    total_points integer DEFAULT 0,
    loyalty_tier text DEFAULT 'BRONZE'::text,
    total_recharge_amount bigint DEFAULT 0,
    total_transactions integer DEFAULT 0,
    last_recharge_date timestamp with time zone,
    referral_code text,
    referred_by uuid,
    total_referrals integer DEFAULT 0,
    is_active boolean DEFAULT true,
    is_verified boolean DEFAULT false,
    kyc_status text DEFAULT 'PENDING'::text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    last_login_at timestamp with time zone,
    user_code character varying(20),
    CONSTRAINT chk_users_msisdn_format CHECK ((msisdn ~ '^234[7-9][0-1][0-9]{8}$'::text)),
    CONSTRAINT users_gender_check CHECK (((gender = ANY (ARRAY['MALE'::text, 'FEMALE'::text, 'OTHER'::text, ''::text])) OR (gender IS NULL))),
    CONSTRAINT users_kyc_status_check CHECK ((kyc_status = ANY (ARRAY['PENDING'::text, 'VERIFIED'::text, 'REJECTED'::text]))),
    CONSTRAINT users_loyalty_tier_check CHECK ((loyalty_tier = ANY (ARRAY['BRONZE'::text, 'SILVER'::text, 'GOLD'::text, 'PLATINUM'::text]))),
    CONSTRAINT valid_email CHECK (((email = ''::text) OR (email IS NULL) OR (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'::text))),
    CONSTRAINT valid_msisdn CHECK ((msisdn ~ '^234[789][01][0-9]{8}$'::text))
);


--
-- Name: wallet_transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.wallet_transactions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    wallet_id uuid NOT NULL,
    msisdn character varying(20) NOT NULL,
    type character varying(30) NOT NULL,
    amount bigint NOT NULL,
    balance_before bigint NOT NULL,
    balance_after bigint NOT NULL,
    reference character varying(100) NOT NULL,
    description text,
    status character varying(20) DEFAULT 'completed'::character varying NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT wallet_transactions_amount_check CHECK ((amount > 0)),
    CONSTRAINT wallet_transactions_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'completed'::character varying, 'failed'::character varying, 'reversed'::character varying])::text[]))),
    CONSTRAINT wallet_transactions_type_check CHECK (((type)::text = ANY ((ARRAY['credit'::character varying, 'debit'::character varying, 'hold'::character varying, 'release'::character varying, 'adjustment'::character varying])::text[])))
);


--
-- Name: wallets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.wallets (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    msisdn character varying(20) NOT NULL,
    balance bigint DEFAULT 0 NOT NULL,
    pending_balance bigint DEFAULT 0 NOT NULL,
    total_earned bigint DEFAULT 0 NOT NULL,
    total_withdrawn bigint DEFAULT 0 NOT NULL,
    min_payout_amount bigint DEFAULT 100000 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    is_suspended boolean DEFAULT false NOT NULL,
    suspension_reason text,
    last_transaction_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT wallets_balance_check CHECK ((balance >= 0)),
    CONSTRAINT wallets_pending_balance_check CHECK ((pending_balance >= 0))
);


--
-- Name: wheel_prizes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.wheel_prizes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    prize_name text NOT NULL,
    prize_type text NOT NULL,
    prize_value bigint NOT NULL,
    probability numeric(5,2) NOT NULL,
    minimum_recharge numeric(10,2) DEFAULT 0,
    is_active boolean DEFAULT true,
    icon_name text,
    color_scheme text,
    sort_order integer DEFAULT 0,
    description text,
    terms_and_conditions text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT positive_prize_value CHECK (((prize_value)::numeric > (0)::numeric)),
    CONSTRAINT valid_probability CHECK (((probability >= (0)::numeric) AND (probability <= (100)::numeric))),
    CONSTRAINT wheel_prizes_prize_type_check CHECK ((prize_type = ANY (ARRAY['CASH'::text, 'AIRTIME'::text, 'DATA'::text, 'POINTS'::text])))
);


--
-- Name: otp_rate_limits id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.otp_rate_limits ALTER COLUMN id SET DEFAULT nextval('public.otp_rate_limits_id_seq'::regclass);


--
-- Name: prize_fulfillment_config id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prize_fulfillment_config ALTER COLUMN id SET DEFAULT nextval('public.prize_fulfillment_config_id_seq'::regclass);


--
-- Name: prize_fulfillment_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prize_fulfillment_logs ALTER COLUMN id SET DEFAULT nextval('public.prize_fulfillment_logs_id_seq'::regclass);


--
-- Name: provider_configs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.provider_configs ALTER COLUMN id SET DEFAULT nextval('public.provider_configs_id_seq'::regclass);


--
-- Name: admin_activity_logs admin_activity_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_activity_logs
    ADD CONSTRAINT admin_activity_logs_pkey PRIMARY KEY (id);


--
-- Name: admin_sessions admin_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_sessions
    ADD CONSTRAINT admin_sessions_pkey PRIMARY KEY (id);


--
-- Name: admin_sessions admin_sessions_session_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_sessions
    ADD CONSTRAINT admin_sessions_session_token_key UNIQUE (session_token);


--
-- Name: admin_users admin_users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_users
    ADD CONSTRAINT admin_users_email_key UNIQUE (email);


--
-- Name: admin_users admin_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_users
    ADD CONSTRAINT admin_users_pkey PRIMARY KEY (id);


--
-- Name: affiliate_analytics affiliate_analytics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_analytics
    ADD CONSTRAINT affiliate_analytics_pkey PRIMARY KEY (id);


--
-- Name: affiliate_bank_accounts affiliate_bank_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_bank_accounts
    ADD CONSTRAINT affiliate_bank_accounts_pkey PRIMARY KEY (id);


--
-- Name: affiliate_clicks affiliate_clicks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_clicks
    ADD CONSTRAINT affiliate_clicks_pkey PRIMARY KEY (id);


--
-- Name: affiliate_commissions affiliate_commissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_commissions
    ADD CONSTRAINT affiliate_commissions_pkey PRIMARY KEY (id);


--
-- Name: affiliate_payouts affiliate_payouts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_payouts
    ADD CONSTRAINT affiliate_payouts_pkey PRIMARY KEY (id);


--
-- Name: affiliates affiliates_affiliate_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliates
    ADD CONSTRAINT affiliates_affiliate_code_key UNIQUE (affiliate_code);


--
-- Name: affiliates affiliates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliates
    ADD CONSTRAINT affiliates_pkey PRIMARY KEY (id);


--
-- Name: application_logs application_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.application_logs
    ADD CONSTRAINT application_logs_pkey PRIMARY KEY (id);


--
-- Name: application_metrics application_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.application_metrics
    ADD CONSTRAINT application_metrics_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: daily_subscription_config daily_subscription_config_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.daily_subscription_config
    ADD CONSTRAINT daily_subscription_config_pkey PRIMARY KEY (id);


--
-- Name: daily_subscriptions daily_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.daily_subscriptions
    ADD CONSTRAINT daily_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: daily_subscriptions daily_subscriptions_user_id_subscription_date_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.daily_subscriptions
    ADD CONSTRAINT daily_subscriptions_user_id_subscription_date_key UNIQUE (user_id, subscription_date);


--
-- Name: data_plans data_plans_network_id_plan_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.data_plans
    ADD CONSTRAINT data_plans_network_id_plan_code_key UNIQUE (network_id, plan_code);


--
-- Name: data_plans data_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.data_plans
    ADD CONSTRAINT data_plans_pkey PRIMARY KEY (id);


--
-- Name: draw_entries draw_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draw_entries
    ADD CONSTRAINT draw_entries_pkey PRIMARY KEY (id);


--
-- Name: draw_types draw_types_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draw_types
    ADD CONSTRAINT draw_types_name_key UNIQUE (name);


--
-- Name: draw_types draw_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draw_types
    ADD CONSTRAINT draw_types_pkey PRIMARY KEY (id);


--
-- Name: draw_winners draw_winners_draw_id_position_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draw_winners
    ADD CONSTRAINT draw_winners_draw_id_position_key UNIQUE (draw_id, "position");


--
-- Name: draw_winners draw_winners_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draw_winners
    ADD CONSTRAINT draw_winners_pkey PRIMARY KEY (id);


--
-- Name: draws draws_draw_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draws
    ADD CONSTRAINT draws_draw_code_key UNIQUE (draw_code);


--
-- Name: draws draws_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draws
    ADD CONSTRAINT draws_pkey PRIMARY KEY (id);


--
-- Name: network_cache network_cache_msisdn_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.network_cache
    ADD CONSTRAINT network_cache_msisdn_key UNIQUE (msisdn);


--
-- Name: network_cache network_cache_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.network_cache
    ADD CONSTRAINT network_cache_pkey PRIMARY KEY (id);


--
-- Name: network_configs network_configs_network_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.network_configs
    ADD CONSTRAINT network_configs_network_code_key UNIQUE (network_code);


--
-- Name: network_configs network_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.network_configs
    ADD CONSTRAINT network_configs_pkey PRIMARY KEY (id);


--
-- Name: notification_delivery_log notification_delivery_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_delivery_log
    ADD CONSTRAINT notification_delivery_log_pkey PRIMARY KEY (id);


--
-- Name: notification_templates notification_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_templates
    ADD CONSTRAINT notification_templates_pkey PRIMARY KEY (id);


--
-- Name: notification_templates notification_templates_template_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_templates
    ADD CONSTRAINT notification_templates_template_key_key UNIQUE (template_key);


--
-- Name: otp_rate_limits otp_rate_limits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.otp_rate_limits
    ADD CONSTRAINT otp_rate_limits_pkey PRIMARY KEY (id);


--
-- Name: otp_verifications otp_verifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.otp_verifications
    ADD CONSTRAINT otp_verifications_pkey PRIMARY KEY (id);


--
-- Name: otps otps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.otps
    ADD CONSTRAINT otps_pkey PRIMARY KEY (id);


--
-- Name: payment_logs payment_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_logs
    ADD CONSTRAINT payment_logs_pkey PRIMARY KEY (id);


--
-- Name: platform_settings platform_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_settings
    ADD CONSTRAINT platform_settings_pkey PRIMARY KEY (setting_key);


--
-- Name: points_adjustments points_adjustments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.points_adjustments
    ADD CONSTRAINT points_adjustments_pkey PRIMARY KEY (id);


--
-- Name: prize_categories prize_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prize_categories
    ADD CONSTRAINT prize_categories_pkey PRIMARY KEY (id);


--
-- Name: prize_fulfillment_config prize_fulfillment_config_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prize_fulfillment_config
    ADD CONSTRAINT prize_fulfillment_config_pkey PRIMARY KEY (id);


--
-- Name: prize_fulfillment_logs prize_fulfillment_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prize_fulfillment_logs
    ADD CONSTRAINT prize_fulfillment_logs_pkey PRIMARY KEY (id);


--
-- Name: prize_templates prize_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prize_templates
    ADD CONSTRAINT prize_templates_pkey PRIMARY KEY (id);


--
-- Name: provider_configs provider_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.provider_configs
    ADD CONSTRAINT provider_configs_pkey PRIMARY KEY (id);


--
-- Name: spin_results spin_results_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spin_results
    ADD CONSTRAINT spin_results_pkey PRIMARY KEY (id);


--
-- Name: spin_results spin_results_spin_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spin_results
    ADD CONSTRAINT spin_results_spin_code_key UNIQUE (spin_code);


--
-- Name: spin_tiers spin_tiers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spin_tiers
    ADD CONSTRAINT spin_tiers_pkey PRIMARY KEY (id);


--
-- Name: spin_tiers spin_tiers_tier_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spin_tiers
    ADD CONSTRAINT spin_tiers_tier_name_key UNIQUE (tier_name);


--
-- Name: subscription_tiers subscription_tiers_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscription_tiers
    ADD CONSTRAINT subscription_tiers_name_key UNIQUE (name);


--
-- Name: subscription_tiers subscription_tiers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscription_tiers
    ADD CONSTRAINT subscription_tiers_pkey PRIMARY KEY (id);


--
-- Name: transaction_limits_audit transaction_limits_audit_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transaction_limits_audit
    ADD CONSTRAINT transaction_limits_audit_pkey PRIMARY KEY (id);


--
-- Name: transaction_limits transaction_limits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transaction_limits
    ADD CONSTRAINT transaction_limits_pkey PRIMARY KEY (id);


--
-- Name: transactions transactions_payment_reference_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_payment_reference_key UNIQUE (payment_reference);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: provider_configs unique_active_provider; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.provider_configs
    ADD CONSTRAINT unique_active_provider UNIQUE (network, service_type, priority, is_active);


--
-- Name: affiliate_analytics unique_affiliate_date; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_analytics
    ADD CONSTRAINT unique_affiliate_date UNIQUE (affiliate_id, analytics_date);


--
-- Name: transaction_limits unique_limit_config; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transaction_limits
    ADD CONSTRAINT unique_limit_config UNIQUE (limit_type, limit_scope, applies_to_user_tier);


--
-- Name: prize_fulfillment_config unique_prize_type; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prize_fulfillment_config
    ADD CONSTRAINT unique_prize_type UNIQUE (prize_type);


--
-- Name: user_notification_preferences user_notification_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notification_preferences
    ADD CONSTRAINT user_notification_preferences_pkey PRIMARY KEY (id);


--
-- Name: user_notification_preferences user_notification_preferences_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notification_preferences
    ADD CONSTRAINT user_notification_preferences_user_id_key UNIQUE (user_id);


--
-- Name: user_notifications user_notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications_pkey PRIMARY KEY (id);


--
-- Name: users users_auth_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_auth_user_id_key UNIQUE (auth_user_id);


--
-- Name: users users_msisdn_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_msisdn_key UNIQUE (msisdn);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_referral_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_referral_code_key UNIQUE (referral_code);


--
-- Name: wallet_transactions wallet_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_pkey PRIMARY KEY (id);


--
-- Name: wallet_transactions wallet_transactions_reference_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_reference_key UNIQUE (reference);


--
-- Name: wallets wallets_msisdn_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_msisdn_key UNIQUE (msisdn);


--
-- Name: wallets wallets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_pkey PRIMARY KEY (id);


--
-- Name: wheel_prizes wheel_prizes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wheel_prizes
    ADD CONSTRAINT wheel_prizes_pkey PRIMARY KEY (id);


--
-- Name: idx_admin_activity_logs_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_activity_logs_action ON public.admin_activity_logs USING btree (action);


--
-- Name: idx_admin_activity_logs_action_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_activity_logs_action_type ON public.admin_activity_logs USING btree (action);


--
-- Name: idx_admin_activity_logs_admin_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_activity_logs_admin_user_id ON public.admin_activity_logs USING btree (admin_user_id);


--
-- Name: idx_admin_activity_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_activity_logs_created_at ON public.admin_activity_logs USING btree (created_at);


--
-- Name: idx_admin_activity_logs_is_suspicious; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_activity_logs_is_suspicious ON public.admin_activity_logs USING btree (is_suspicious) WHERE (is_suspicious = true);


--
-- Name: idx_admin_activity_logs_resource; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_activity_logs_resource ON public.admin_activity_logs USING btree (resource);


--
-- Name: idx_admin_activity_logs_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_activity_logs_resource_id ON public.admin_activity_logs USING btree (resource_id);


--
-- Name: idx_admin_activity_logs_risk_score; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_activity_logs_risk_score ON public.admin_activity_logs USING btree (risk_score) WHERE (risk_score > 50);


--
-- Name: idx_admin_activity_logs_security; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_activity_logs_security ON public.admin_activity_logs USING btree (admin_user_id, created_at, is_suspicious);


--
-- Name: idx_admin_activity_logs_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_activity_logs_session_id ON public.admin_activity_logs USING btree (admin_session_id);


--
-- Name: idx_admin_sessions_admin_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_sessions_admin_user_id ON public.admin_sessions USING btree (admin_user_id);


--
-- Name: idx_admin_sessions_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_sessions_expires_at ON public.admin_sessions USING btree (expires_at);


--
-- Name: idx_admin_sessions_session_token; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_sessions_session_token ON public.admin_sessions USING btree (session_token);


--
-- Name: idx_admin_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_users_email ON public.admin_users USING btree (email);


--
-- Name: idx_admin_users_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_users_is_active ON public.admin_users USING btree (is_active);


--
-- Name: idx_admin_users_role; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_users_role ON public.admin_users USING btree (role);


--
-- Name: idx_affiliate_analytics_affiliate_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_analytics_affiliate_id ON public.affiliate_analytics USING btree (affiliate_id);


--
-- Name: idx_affiliate_analytics_conversions; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_analytics_conversions ON public.affiliate_analytics USING btree (conversions);


--
-- Name: idx_affiliate_analytics_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_analytics_date ON public.affiliate_analytics USING btree (analytics_date);


--
-- Name: idx_affiliate_bank_accounts_affiliate_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_bank_accounts_affiliate_id ON public.affiliate_bank_accounts USING btree (affiliate_id);


--
-- Name: idx_affiliate_bank_accounts_is_primary; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_bank_accounts_is_primary ON public.affiliate_bank_accounts USING btree (is_primary) WHERE (is_primary = true);


--
-- Name: idx_affiliate_bank_accounts_is_verified; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_bank_accounts_is_verified ON public.affiliate_bank_accounts USING btree (is_verified);


--
-- Name: idx_affiliate_clicks_affiliate_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_clicks_affiliate_id ON public.affiliate_clicks USING btree (affiliate_id);


--
-- Name: idx_affiliate_clicks_converted; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_clicks_converted ON public.affiliate_clicks USING btree (converted);


--
-- Name: idx_affiliate_clicks_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_clicks_created_at ON public.affiliate_clicks USING btree (created_at DESC);


--
-- Name: idx_affiliate_commissions_affiliate_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_commissions_affiliate_id ON public.affiliate_commissions USING btree (affiliate_id);


--
-- Name: idx_affiliate_commissions_earned_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_commissions_earned_at ON public.affiliate_commissions USING btree (earned_at DESC);


--
-- Name: idx_affiliate_commissions_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_commissions_status ON public.affiliate_commissions USING btree (status);


--
-- Name: idx_affiliate_commissions_transaction_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_commissions_transaction_id ON public.affiliate_commissions USING btree (transaction_id);


--
-- Name: idx_affiliate_payouts_affiliate_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_payouts_affiliate_id ON public.affiliate_payouts USING btree (affiliate_id);


--
-- Name: idx_affiliate_payouts_batch_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_payouts_batch_id ON public.affiliate_payouts USING btree (payout_batch_id);


--
-- Name: idx_affiliate_payouts_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_payouts_created_at ON public.affiliate_payouts USING btree (created_at);


--
-- Name: idx_affiliate_payouts_pending; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_payouts_pending ON public.affiliate_payouts USING btree (affiliate_id, payout_status, created_at) WHERE (payout_status = 'PENDING'::text);


--
-- Name: idx_affiliate_payouts_processed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_payouts_processed_at ON public.affiliate_payouts USING btree (processed_at);


--
-- Name: idx_affiliate_payouts_reference; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_payouts_reference ON public.affiliate_payouts USING btree (payout_reference);


--
-- Name: idx_affiliate_payouts_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliate_payouts_status ON public.affiliate_payouts USING btree (payout_status);


--
-- Name: idx_affiliates_affiliate_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliates_affiliate_code ON public.affiliates USING btree (affiliate_code);


--
-- Name: idx_affiliates_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliates_status ON public.affiliates USING btree (status);


--
-- Name: idx_affiliates_tier; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliates_tier ON public.affiliates USING btree (tier);


--
-- Name: idx_affiliates_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_affiliates_user_id ON public.affiliates USING btree (user_id);


--
-- Name: idx_application_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_application_logs_created_at ON public.application_logs USING btree (created_at DESC);


--
-- Name: idx_application_logs_level; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_application_logs_level ON public.application_logs USING btree (level);


--
-- Name: idx_application_logs_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_application_logs_user_id ON public.application_logs USING btree (user_id);


--
-- Name: idx_application_metrics_metric_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_application_metrics_metric_name ON public.application_metrics USING btree (metric_name);


--
-- Name: idx_application_metrics_recorded_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_application_metrics_recorded_at ON public.application_metrics USING btree (recorded_at DESC);


--
-- Name: idx_audit_logs_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_action ON public.audit_logs USING btree (action);


--
-- Name: idx_audit_logs_admin_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_admin_user_id ON public.audit_logs USING btree (admin_user_id);


--
-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_created_at ON public.audit_logs USING btree (created_at DESC);


--
-- Name: idx_audit_logs_entity; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_entity ON public.audit_logs USING btree (entity_type, entity_id);


--
-- Name: idx_audit_logs_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_user_id ON public.audit_logs USING btree (user_id);


--
-- Name: idx_daily_subscriptions_msisdn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_daily_subscriptions_msisdn ON public.daily_subscriptions USING btree (msisdn);


--
-- Name: idx_daily_subscriptions_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_daily_subscriptions_status ON public.daily_subscriptions USING btree (status);


--
-- Name: idx_daily_subscriptions_subscription_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_daily_subscriptions_subscription_date ON public.daily_subscriptions USING btree (subscription_date);


--
-- Name: idx_daily_subscriptions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_daily_subscriptions_user_id ON public.daily_subscriptions USING btree (user_id);


--
-- Name: idx_data_plans_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_data_plans_is_active ON public.data_plans USING btree (is_active);


--
-- Name: idx_data_plans_network_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_data_plans_network_id ON public.data_plans USING btree (network_id);


--
-- Name: idx_data_plans_network_provider; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_data_plans_network_provider ON public.data_plans USING btree (network_provider);


--
-- Name: idx_data_plans_price; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_data_plans_price ON public.data_plans USING btree (price);


--
-- Name: idx_delivery_log_attempted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_delivery_log_attempted_at ON public.notification_delivery_log USING btree (attempted_at DESC);


--
-- Name: idx_delivery_log_channel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_delivery_log_channel ON public.notification_delivery_log USING btree (channel);


--
-- Name: idx_delivery_log_notification_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_delivery_log_notification_id ON public.notification_delivery_log USING btree (notification_id);


--
-- Name: idx_delivery_log_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_delivery_log_status ON public.notification_delivery_log USING btree (delivery_status);


--
-- Name: idx_draw_entries_draw_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draw_entries_draw_id ON public.draw_entries USING btree (draw_id);


--
-- Name: idx_draw_entries_source_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draw_entries_source_type ON public.draw_entries USING btree (source_type);


--
-- Name: idx_draw_entries_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draw_entries_user_id ON public.draw_entries USING btree (user_id);


--
-- Name: idx_draw_types_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draw_types_active ON public.draw_types USING btree (is_active);


--
-- Name: idx_draw_winners_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draw_winners_category ON public.draw_winners USING btree (prize_category_id);


--
-- Name: idx_draw_winners_claim_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draw_winners_claim_status ON public.draw_winners USING btree (claim_status);


--
-- Name: idx_draw_winners_draw_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draw_winners_draw_id ON public.draw_winners USING btree (draw_id);


--
-- Name: idx_draw_winners_forfeited; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draw_winners_forfeited ON public.draw_winners USING btree (draw_id, is_forfeited);


--
-- Name: idx_draw_winners_payout; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draw_winners_payout ON public.draw_winners USING btree (payout_status);


--
-- Name: idx_draw_winners_provision; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draw_winners_provision ON public.draw_winners USING btree (provision_status) WHERE (auto_provision = true);


--
-- Name: idx_draw_winners_runner_up; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draw_winners_runner_up ON public.draw_winners USING btree (draw_id, is_runner_up);


--
-- Name: idx_draw_winners_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draw_winners_user_id ON public.draw_winners USING btree (user_id);


--
-- Name: idx_draws_draw_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draws_draw_code ON public.draws USING btree (draw_code);


--
-- Name: idx_draws_end_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draws_end_time ON public.draws USING btree (end_time);


--
-- Name: idx_draws_start_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draws_start_time ON public.draws USING btree (start_time);


--
-- Name: idx_draws_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draws_status ON public.draws USING btree (status);


--
-- Name: idx_draws_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_draws_type ON public.draws USING btree (type);


--
-- Name: idx_fulfillment_config_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fulfillment_config_active ON public.prize_fulfillment_config USING btree (is_active);


--
-- Name: idx_fulfillment_config_prize_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fulfillment_config_prize_type ON public.prize_fulfillment_config USING btree (prize_type);


--
-- Name: idx_fulfillment_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fulfillment_logs_created_at ON public.prize_fulfillment_logs USING btree (created_at DESC);


--
-- Name: idx_fulfillment_logs_msisdn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fulfillment_logs_msisdn ON public.prize_fulfillment_logs USING btree (msisdn);


--
-- Name: idx_fulfillment_logs_provider_ref; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fulfillment_logs_provider_ref ON public.prize_fulfillment_logs USING btree (provider_reference);


--
-- Name: idx_fulfillment_logs_spin_result; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fulfillment_logs_spin_result ON public.prize_fulfillment_logs USING btree (spin_result_id);


--
-- Name: idx_fulfillment_logs_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fulfillment_logs_status ON public.prize_fulfillment_logs USING btree (status);


--
-- Name: idx_limits_audit_changed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_limits_audit_changed_at ON public.transaction_limits_audit USING btree (changed_at DESC);


--
-- Name: idx_limits_audit_changed_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_limits_audit_changed_by ON public.transaction_limits_audit USING btree (changed_by);


--
-- Name: idx_limits_audit_limit_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_limits_audit_limit_id ON public.transaction_limits_audit USING btree (limit_id);


--
-- Name: idx_network_cache_expires; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_network_cache_expires ON public.network_cache USING btree (cache_expires_at);


--
-- Name: idx_network_cache_msisdn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_network_cache_msisdn ON public.network_cache USING btree (msisdn);


--
-- Name: idx_network_cache_valid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_network_cache_valid ON public.network_cache USING btree (is_valid);


--
-- Name: idx_network_configs_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_network_configs_is_active ON public.network_configs USING btree (is_active);


--
-- Name: idx_network_configs_network_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_network_configs_network_code ON public.network_configs USING btree (network_code);


--
-- Name: idx_network_configs_sort_order; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_network_configs_sort_order ON public.network_configs USING btree (sort_order);


--
-- Name: idx_notification_templates_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notification_templates_is_active ON public.notification_templates USING btree (is_active);


--
-- Name: idx_notification_templates_template_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notification_templates_template_key ON public.notification_templates USING btree (template_key);


--
-- Name: idx_otp_rate_limits_key_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otp_rate_limits_key_time ON public.otp_rate_limits USING btree (key, requested_at);


--
-- Name: idx_otp_rate_limits_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otp_rate_limits_time ON public.otp_rate_limits USING btree (requested_at);


--
-- Name: idx_otp_verifications_active_lookup; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otp_verifications_active_lookup ON public.otp_verifications USING btree (msisdn, purpose, is_verified, expires_at) WHERE ((is_verified = false) AND (is_expired = false) AND (is_revoked = false));


--
-- Name: idx_otp_verifications_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otp_verifications_created_at ON public.otp_verifications USING btree (created_at);


--
-- Name: idx_otp_verifications_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otp_verifications_expires_at ON public.otp_verifications USING btree (expires_at) WHERE (is_verified = false);


--
-- Name: idx_otp_verifications_is_verified; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otp_verifications_is_verified ON public.otp_verifications USING btree (is_verified);


--
-- Name: idx_otp_verifications_msisdn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otp_verifications_msisdn ON public.otp_verifications USING btree (msisdn);


--
-- Name: idx_otp_verifications_purpose; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otp_verifications_purpose ON public.otp_verifications USING btree (purpose);


--
-- Name: idx_otp_verifications_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otp_verifications_user_id ON public.otp_verifications USING btree (user_id);


--
-- Name: idx_otps_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otps_code ON public.otps USING btree (code);


--
-- Name: idx_otps_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otps_expires_at ON public.otps USING btree (expires_at);


--
-- Name: idx_otps_is_used; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otps_is_used ON public.otps USING btree (is_used);


--
-- Name: idx_otps_msisdn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_otps_msisdn ON public.otps USING btree (msisdn);


--
-- Name: idx_payment_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_logs_created_at ON public.payment_logs USING btree (created_at);


--
-- Name: idx_payment_logs_errors; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_logs_errors ON public.payment_logs USING btree (event_type, is_successful, created_at) WHERE (is_successful = false);


--
-- Name: idx_payment_logs_event_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_logs_event_type ON public.payment_logs USING btree (event_type);


--
-- Name: idx_payment_logs_is_successful; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_logs_is_successful ON public.payment_logs USING btree (is_successful);


--
-- Name: idx_payment_logs_payment_reference; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_logs_payment_reference ON public.payment_logs USING btree (payment_reference);


--
-- Name: idx_payment_logs_slow_requests; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_logs_slow_requests ON public.payment_logs USING btree (response_time_ms, created_at) WHERE (response_time_ms > 5000);


--
-- Name: idx_payment_logs_status_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_logs_status_code ON public.payment_logs USING btree (status_code);


--
-- Name: idx_payment_logs_transaction_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_logs_transaction_id ON public.payment_logs USING btree (transaction_id);


--
-- Name: idx_payment_logs_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_logs_user_id ON public.payment_logs USING btree (user_id);


--
-- Name: idx_points_adjustments_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_points_adjustments_created_at ON public.points_adjustments USING btree (created_at DESC);


--
-- Name: idx_points_adjustments_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_points_adjustments_created_by ON public.points_adjustments USING btree (created_by);


--
-- Name: idx_points_adjustments_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_points_adjustments_user_id ON public.points_adjustments USING btree (user_id);


--
-- Name: idx_prize_categories_draw; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prize_categories_draw ON public.prize_categories USING btree (draw_id);


--
-- Name: idx_prize_categories_order; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prize_categories_order ON public.prize_categories USING btree (display_order);


--
-- Name: idx_prize_categories_template; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prize_categories_template ON public.prize_categories USING btree (template_id);


--
-- Name: idx_prize_templates_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prize_templates_active ON public.prize_templates USING btree (is_active);


--
-- Name: idx_prize_templates_draw_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prize_templates_draw_type ON public.prize_templates USING btree (draw_type_id);


--
-- Name: idx_provider_configs_lookup; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_provider_configs_lookup ON public.provider_configs USING btree (network, service_type, is_active, priority);


--
-- Name: idx_spin_results_can_retry; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_can_retry ON public.spin_results USING btree (can_retry) WHERE (can_retry = true);


--
-- Name: idx_spin_results_claim_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_claim_status ON public.spin_results USING btree (claim_status);


--
-- Name: idx_spin_results_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_created_at ON public.spin_results USING btree (created_at DESC);


--
-- Name: idx_spin_results_created_at2; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_created_at2 ON public.spin_results USING btree (created_at DESC);


--
-- Name: idx_spin_results_fulfillment_mode; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_fulfillment_mode ON public.spin_results USING btree (fulfillment_mode);


--
-- Name: idx_spin_results_msisdn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_msisdn ON public.spin_results USING btree (msisdn);


--
-- Name: idx_spin_results_msisdn2; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_msisdn2 ON public.spin_results USING btree (msisdn);


--
-- Name: idx_spin_results_msisdn_claim_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_msisdn_claim_status ON public.spin_results USING btree (msisdn, claim_status);


--
-- Name: idx_spin_results_prize_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_prize_type ON public.spin_results USING btree (prize_type);


--
-- Name: idx_spin_results_reviewed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_reviewed_at ON public.spin_results USING btree (reviewed_at DESC);


--
-- Name: idx_spin_results_reviewed_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_reviewed_by ON public.spin_results USING btree (reviewed_by);


--
-- Name: idx_spin_results_spin_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_spin_results_spin_code ON public.spin_results USING btree (spin_code);


--
-- Name: idx_spin_results_transaction_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_transaction_id ON public.spin_results USING btree (transaction_id);


--
-- Name: idx_spin_results_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_results_user_id ON public.spin_results USING btree (user_id);


--
-- Name: idx_spin_tiers_amount_range; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_tiers_amount_range ON public.spin_tiers USING btree (min_daily_amount, max_daily_amount);


--
-- Name: idx_spin_tiers_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_tiers_is_active ON public.spin_tiers USING btree (is_active);


--
-- Name: idx_spin_tiers_sort_order; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spin_tiers_sort_order ON public.spin_tiers USING btree (sort_order);


--
-- Name: idx_subscription_tiers_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_subscription_tiers_is_active ON public.subscription_tiers USING btree (is_active);


--
-- Name: idx_transaction_limits_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transaction_limits_active ON public.transaction_limits USING btree (is_active) WHERE (is_active = true);


--
-- Name: idx_transaction_limits_tier; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transaction_limits_tier ON public.transaction_limits USING btree (applies_to_user_tier);


--
-- Name: idx_transaction_limits_type_scope; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transaction_limits_type_scope ON public.transaction_limits USING btree (limit_type, limit_scope);


--
-- Name: idx_transactions_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_created_at ON public.transactions USING btree (created_at DESC);


--
-- Name: idx_transactions_msisdn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_msisdn ON public.transactions USING btree (msisdn);


--
-- Name: idx_transactions_msisdn2; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_msisdn2 ON public.transactions USING btree (msisdn);


--
-- Name: idx_transactions_network_provider; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_network_provider ON public.transactions USING btree (network_provider);


--
-- Name: idx_transactions_payment_reference; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_payment_reference ON public.transactions USING btree (payment_reference);


--
-- Name: idx_transactions_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_status ON public.transactions USING btree (status);


--
-- Name: idx_transactions_transaction_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_transactions_transaction_code ON public.transactions USING btree (transaction_code);


--
-- Name: idx_transactions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_user_id ON public.transactions USING btree (user_id);


--
-- Name: idx_user_notifications_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_notifications_created_at ON public.user_notifications USING btree (created_at DESC);


--
-- Name: idx_user_notifications_is_read; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_notifications_is_read ON public.user_notifications USING btree (is_read);


--
-- Name: idx_user_notifications_reference; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_notifications_reference ON public.user_notifications USING btree (reference_type, reference_id);


--
-- Name: idx_user_notifications_scheduled_for; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_notifications_scheduled_for ON public.user_notifications USING btree (scheduled_for);


--
-- Name: idx_user_notifications_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_notifications_type ON public.user_notifications USING btree (notification_type);


--
-- Name: idx_user_notifications_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_notifications_user_id ON public.user_notifications USING btree (user_id);


--
-- Name: idx_user_preferences_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_preferences_user_id ON public.user_notification_preferences USING btree (user_id);


--
-- Name: idx_users_auth_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_auth_user_id ON public.users USING btree (auth_user_id);


--
-- Name: idx_users_loyalty_tier; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_loyalty_tier ON public.users USING btree (loyalty_tier);


--
-- Name: idx_users_msisdn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_msisdn ON public.users USING btree (msisdn);


--
-- Name: idx_users_referral_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_referral_code ON public.users USING btree (referral_code) WHERE (referral_code IS NOT NULL);


--
-- Name: idx_users_referred_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_referred_by ON public.users USING btree (referred_by);


--
-- Name: idx_wallet_transactions_msisdn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wallet_transactions_msisdn ON public.wallet_transactions USING btree (msisdn);


--
-- Name: idx_wallet_transactions_reference; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wallet_transactions_reference ON public.wallet_transactions USING btree (reference);


--
-- Name: idx_wallet_transactions_wallet_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wallet_transactions_wallet_id ON public.wallet_transactions USING btree (wallet_id);


--
-- Name: idx_wallets_msisdn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wallets_msisdn ON public.wallets USING btree (msisdn);


--
-- Name: idx_wheel_prizes_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wheel_prizes_is_active ON public.wheel_prizes USING btree (is_active);


--
-- Name: idx_wheel_prizes_prize_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wheel_prizes_prize_type ON public.wheel_prizes USING btree (prize_type);


--
-- Name: idx_wheel_prizes_sort_order; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wheel_prizes_sort_order ON public.wheel_prizes USING btree (sort_order);


--
-- Name: transaction_limits transaction_limits_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER transaction_limits_updated_at BEFORE UPDATE ON public.transaction_limits FOR EACH ROW EXECUTE FUNCTION public.update_transaction_limits_timestamp();


--
-- Name: affiliate_bank_accounts trigger_ensure_single_primary_bank_account; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_ensure_single_primary_bank_account BEFORE INSERT OR UPDATE ON public.affiliate_bank_accounts FOR EACH ROW WHEN ((new.is_primary = true)) EXECUTE FUNCTION public.ensure_single_primary_bank_account();


--
-- Name: otp_verifications trigger_mark_expired_otps; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_mark_expired_otps BEFORE UPDATE ON public.otp_verifications FOR EACH ROW EXECUTE FUNCTION public.mark_expired_otps();


--
-- Name: otps trigger_otps_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_otps_updated_at BEFORE UPDATE ON public.otps FOR EACH ROW EXECUTE FUNCTION public.update_otps_updated_at();


--
-- Name: spin_tiers trigger_spin_tiers_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_spin_tiers_updated_at BEFORE UPDATE ON public.spin_tiers FOR EACH ROW EXECUTE FUNCTION public.update_spin_tiers_updated_at();


--
-- Name: affiliate_analytics trigger_update_affiliate_analytics_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_update_affiliate_analytics_timestamp BEFORE UPDATE ON public.affiliate_analytics FOR EACH ROW EXECUTE FUNCTION public.update_affiliate_analytics_timestamp();


--
-- Name: affiliate_bank_accounts trigger_update_affiliate_bank_account_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_update_affiliate_bank_account_timestamp BEFORE UPDATE ON public.affiliate_bank_accounts FOR EACH ROW EXECUTE FUNCTION public.update_affiliate_bank_account_timestamp();


--
-- Name: affiliate_payouts trigger_update_affiliate_payout_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_update_affiliate_payout_timestamp BEFORE UPDATE ON public.affiliate_payouts FOR EACH ROW EXECUTE FUNCTION public.update_affiliate_payout_timestamp();


--
-- Name: admin_users update_admin_users_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_admin_users_updated_at BEFORE UPDATE ON public.admin_users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: affiliates update_affiliates_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_affiliates_updated_at BEFORE UPDATE ON public.affiliates FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: daily_subscription_config update_daily_subscription_config_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_daily_subscription_config_updated_at BEFORE UPDATE ON public.daily_subscription_config FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: data_plans update_data_plans_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_data_plans_updated_at BEFORE UPDATE ON public.data_plans FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: draw_entries update_draw_entries_trigger; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_draw_entries_trigger AFTER INSERT OR DELETE OR UPDATE ON public.draw_entries FOR EACH ROW EXECUTE FUNCTION public.trigger_update_draw_entries();


--
-- Name: draws update_draws_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_draws_updated_at BEFORE UPDATE ON public.draws FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: network_configs update_network_configs_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_network_configs_updated_at BEFORE UPDATE ON public.network_configs FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: notification_templates update_notification_templates_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_notification_templates_updated_at BEFORE UPDATE ON public.notification_templates FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: platform_settings update_platform_settings_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_platform_settings_updated_at BEFORE UPDATE ON public.platform_settings FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: provider_configs update_provider_configs_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_provider_configs_updated_at BEFORE UPDATE ON public.provider_configs FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: transactions update_transactions_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON public.transactions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: user_notifications update_user_notifications_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_user_notifications_updated_at BEFORE UPDATE ON public.user_notifications FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: user_notification_preferences update_user_preferences_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_user_preferences_updated_at BEFORE UPDATE ON public.user_notification_preferences FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: wheel_prizes update_wheel_prizes_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_wheel_prizes_updated_at BEFORE UPDATE ON public.wheel_prizes FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: admin_activity_logs admin_activity_logs_admin_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_activity_logs
    ADD CONSTRAINT admin_activity_logs_admin_session_id_fkey FOREIGN KEY (admin_session_id) REFERENCES public.admin_sessions(id) ON DELETE SET NULL;


--
-- Name: admin_activity_logs admin_activity_logs_admin_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_activity_logs
    ADD CONSTRAINT admin_activity_logs_admin_user_id_fkey FOREIGN KEY (admin_user_id) REFERENCES public.admin_users(id) ON DELETE SET NULL;


--
-- Name: admin_sessions admin_sessions_admin_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_sessions
    ADD CONSTRAINT admin_sessions_admin_user_id_fkey FOREIGN KEY (admin_user_id) REFERENCES public.admin_users(id) ON DELETE CASCADE;


--
-- Name: affiliate_analytics affiliate_analytics_affiliate_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_analytics
    ADD CONSTRAINT affiliate_analytics_affiliate_id_fkey FOREIGN KEY (affiliate_id) REFERENCES public.affiliates(id) ON DELETE CASCADE;


--
-- Name: affiliate_bank_accounts affiliate_bank_accounts_affiliate_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_bank_accounts
    ADD CONSTRAINT affiliate_bank_accounts_affiliate_id_fkey FOREIGN KEY (affiliate_id) REFERENCES public.affiliates(id) ON DELETE CASCADE;


--
-- Name: affiliate_bank_accounts affiliate_bank_accounts_verified_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_bank_accounts
    ADD CONSTRAINT affiliate_bank_accounts_verified_by_fkey FOREIGN KEY (verified_by) REFERENCES public.admin_users(id);


--
-- Name: affiliate_clicks affiliate_clicks_affiliate_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_clicks
    ADD CONSTRAINT affiliate_clicks_affiliate_id_fkey FOREIGN KEY (affiliate_id) REFERENCES public.affiliates(id) ON DELETE CASCADE;


--
-- Name: affiliate_clicks affiliate_clicks_conversion_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_clicks
    ADD CONSTRAINT affiliate_clicks_conversion_transaction_id_fkey FOREIGN KEY (conversion_transaction_id) REFERENCES public.transactions(id);


--
-- Name: affiliate_commissions affiliate_commissions_affiliate_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_commissions
    ADD CONSTRAINT affiliate_commissions_affiliate_id_fkey FOREIGN KEY (affiliate_id) REFERENCES public.affiliates(id) ON DELETE CASCADE;


--
-- Name: affiliate_commissions affiliate_commissions_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_commissions
    ADD CONSTRAINT affiliate_commissions_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id);


--
-- Name: affiliate_payouts affiliate_payouts_affiliate_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_payouts
    ADD CONSTRAINT affiliate_payouts_affiliate_id_fkey FOREIGN KEY (affiliate_id) REFERENCES public.affiliates(id) ON DELETE SET NULL;


--
-- Name: affiliate_payouts affiliate_payouts_processed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliate_payouts
    ADD CONSTRAINT affiliate_payouts_processed_by_fkey FOREIGN KEY (processed_by) REFERENCES public.admin_users(id);


--
-- Name: affiliates affiliates_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliates
    ADD CONSTRAINT affiliates_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: application_logs application_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.application_logs
    ADD CONSTRAINT application_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: audit_logs audit_logs_admin_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_admin_user_id_fkey FOREIGN KEY (admin_user_id) REFERENCES public.admin_users(id) ON DELETE SET NULL;


--
-- Name: audit_logs audit_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: daily_subscriptions daily_subscriptions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.daily_subscriptions
    ADD CONSTRAINT daily_subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: data_plans data_plans_network_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.data_plans
    ADD CONSTRAINT data_plans_network_id_fkey FOREIGN KEY (network_id) REFERENCES public.network_configs(id) ON DELETE CASCADE;


--
-- Name: draw_entries draw_entries_draw_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draw_entries
    ADD CONSTRAINT draw_entries_draw_id_fkey FOREIGN KEY (draw_id) REFERENCES public.draws(id) ON DELETE CASCADE;


--
-- Name: draw_entries draw_entries_source_subscription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draw_entries
    ADD CONSTRAINT draw_entries_source_subscription_id_fkey FOREIGN KEY (source_subscription_id) REFERENCES public.daily_subscriptions(id);


--
-- Name: draw_entries draw_entries_source_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draw_entries
    ADD CONSTRAINT draw_entries_source_transaction_id_fkey FOREIGN KEY (source_transaction_id) REFERENCES public.transactions(id);


--
-- Name: draw_entries draw_entries_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draw_entries
    ADD CONSTRAINT draw_entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: draw_winners draw_winners_draw_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draw_winners
    ADD CONSTRAINT draw_winners_draw_id_fkey FOREIGN KEY (draw_id) REFERENCES public.draws(id) ON DELETE CASCADE;


--
-- Name: draw_winners draw_winners_prize_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draw_winners
    ADD CONSTRAINT draw_winners_prize_category_id_fkey FOREIGN KEY (prize_category_id) REFERENCES public.prize_categories(id) ON DELETE SET NULL;


--
-- Name: draw_winners draw_winners_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draw_winners
    ADD CONSTRAINT draw_winners_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: draws draws_draw_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draws
    ADD CONSTRAINT draws_draw_type_id_fkey FOREIGN KEY (draw_type_id) REFERENCES public.draw_types(id) ON DELETE SET NULL NOT VALID;


--
-- Name: draws draws_prize_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.draws
    ADD CONSTRAINT draws_prize_template_id_fkey FOREIGN KEY (prize_template_id) REFERENCES public.prize_templates(id) ON DELETE SET NULL NOT VALID;


--
-- Name: points_adjustments fk_points_adjustments_admin; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.points_adjustments
    ADD CONSTRAINT fk_points_adjustments_admin FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: points_adjustments fk_points_adjustments_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.points_adjustments
    ADD CONSTRAINT fk_points_adjustments_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: notification_delivery_log notification_delivery_log_notification_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_delivery_log
    ADD CONSTRAINT notification_delivery_log_notification_id_fkey FOREIGN KEY (notification_id) REFERENCES public.user_notifications(id) ON DELETE CASCADE;


--
-- Name: otp_verifications otp_verifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.otp_verifications
    ADD CONSTRAINT otp_verifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: payment_logs payment_logs_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_logs
    ADD CONSTRAINT payment_logs_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id) ON DELETE SET NULL;


--
-- Name: payment_logs payment_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_logs
    ADD CONSTRAINT payment_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: prize_categories prize_categories_draw_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prize_categories
    ADD CONSTRAINT prize_categories_draw_id_fkey FOREIGN KEY (draw_id) REFERENCES public.draws(id) ON DELETE CASCADE;


--
-- Name: prize_categories prize_categories_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prize_categories
    ADD CONSTRAINT prize_categories_template_id_fkey FOREIGN KEY (template_id) REFERENCES public.prize_templates(id) ON DELETE CASCADE;


--
-- Name: prize_templates prize_templates_draw_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prize_templates
    ADD CONSTRAINT prize_templates_draw_type_id_fkey FOREIGN KEY (draw_type_id) REFERENCES public.draw_types(id) ON DELETE CASCADE;


--
-- Name: spin_results spin_results_prize_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spin_results
    ADD CONSTRAINT spin_results_prize_id_fkey FOREIGN KEY (prize_id) REFERENCES public.wheel_prizes(id);


--
-- Name: spin_results spin_results_reviewed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spin_results
    ADD CONSTRAINT spin_results_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.admin_users(id);


--
-- Name: spin_results spin_results_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spin_results
    ADD CONSTRAINT spin_results_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id);


--
-- Name: spin_results spin_results_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spin_results
    ADD CONSTRAINT spin_results_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: transaction_limits_audit transaction_limits_audit_limit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transaction_limits_audit
    ADD CONSTRAINT transaction_limits_audit_limit_id_fkey FOREIGN KEY (limit_id) REFERENCES public.transaction_limits(id) ON DELETE CASCADE;


--
-- Name: transactions transactions_data_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_data_plan_id_fkey FOREIGN KEY (data_plan_id) REFERENCES public.data_plans(id);


--
-- Name: transactions transactions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: user_notification_preferences user_notification_preferences_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notification_preferences
    ADD CONSTRAINT user_notification_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_notifications user_notifications_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications_template_id_fkey FOREIGN KEY (template_id) REFERENCES public.notification_templates(id);


--
-- Name: user_notifications user_notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notifications
    ADD CONSTRAINT user_notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: users users_referred_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_referred_by_fkey FOREIGN KEY (referred_by) REFERENCES public.users(id);


--
-- Name: wallet_transactions wallet_transactions_wallet_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_wallet_id_fkey FOREIGN KEY (wallet_id) REFERENCES public.wallets(id) ON DELETE CASCADE;


--
-- Name: admin_users; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.admin_users ENABLE ROW LEVEL SECURITY;

--
-- Name: affiliate_commissions; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.affiliate_commissions ENABLE ROW LEVEL SECURITY;

--
-- Name: affiliates; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.affiliates ENABLE ROW LEVEL SECURITY;

--
-- Name: daily_subscriptions; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.daily_subscriptions ENABLE ROW LEVEL SECURITY;

--
-- Name: notification_delivery_log delivery_log_service_only; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY delivery_log_service_only ON public.notification_delivery_log USING (true);


--
-- Name: draw_entries; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.draw_entries ENABLE ROW LEVEL SECURITY;

--
-- Name: draw_winners; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.draw_winners ENABLE ROW LEVEL SECURITY;

--
-- Name: draws; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.draws ENABLE ROW LEVEL SECURITY;

--
-- Name: notification_delivery_log; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.notification_delivery_log ENABLE ROW LEVEL SECURITY;

--
-- Name: notification_templates; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.notification_templates ENABLE ROW LEVEL SECURITY;

--
-- Name: user_notifications notifications_select_own; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY notifications_select_own ON public.user_notifications FOR SELECT USING (true);


--
-- Name: user_notifications notifications_service_manage; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY notifications_service_manage ON public.user_notifications USING (true);


--
-- Name: user_notifications notifications_update_own; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY notifications_update_own ON public.user_notifications FOR UPDATE USING (true);


--
-- Name: user_notification_preferences preferences_insert_own; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY preferences_insert_own ON public.user_notification_preferences FOR INSERT WITH CHECK (true);


--
-- Name: user_notification_preferences preferences_select_own; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY preferences_select_own ON public.user_notification_preferences FOR SELECT USING (true);


--
-- Name: user_notification_preferences preferences_service_manage; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY preferences_service_manage ON public.user_notification_preferences USING (true);


--
-- Name: user_notification_preferences preferences_update_own; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY preferences_update_own ON public.user_notification_preferences FOR UPDATE USING (true);


--
-- Name: admin_users service_full_access_admin_users; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY service_full_access_admin_users ON public.admin_users USING (true) WITH CHECK (true);


--
-- Name: affiliate_commissions service_full_access_affiliate_comms; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY service_full_access_affiliate_comms ON public.affiliate_commissions USING (true) WITH CHECK (true);


--
-- Name: affiliates service_full_access_affiliates; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY service_full_access_affiliates ON public.affiliates USING (true) WITH CHECK (true);


--
-- Name: daily_subscriptions service_full_access_daily_subs; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY service_full_access_daily_subs ON public.daily_subscriptions USING (true) WITH CHECK (true);


--
-- Name: draw_entries service_full_access_draw_entries; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY service_full_access_draw_entries ON public.draw_entries USING (true) WITH CHECK (true);


--
-- Name: draw_winners service_full_access_draw_winners; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY service_full_access_draw_winners ON public.draw_winners USING (true) WITH CHECK (true);


--
-- Name: draws service_full_access_draws; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY service_full_access_draws ON public.draws USING (true) WITH CHECK (true);


--
-- Name: spin_results service_full_access_spin_results; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY service_full_access_spin_results ON public.spin_results USING (true) WITH CHECK (true);


--
-- Name: transactions service_full_access_transactions; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY service_full_access_transactions ON public.transactions USING (true) WITH CHECK (true);


--
-- Name: users service_full_access_users; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY service_full_access_users ON public.users USING (true) WITH CHECK (true);


--
-- Name: spin_results; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.spin_results ENABLE ROW LEVEL SECURITY;

--
-- Name: notification_templates templates_select_public; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY templates_select_public ON public.notification_templates FOR SELECT USING ((is_active = true));


--
-- Name: notification_templates templates_service_manage; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY templates_service_manage ON public.notification_templates USING (true);


--
-- Name: transactions; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.transactions ENABLE ROW LEVEL SECURITY;

--
-- Name: user_notification_preferences; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.user_notification_preferences ENABLE ROW LEVEL SECURITY;

--
-- Name: user_notifications; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.user_notifications ENABLE ROW LEVEL SECURITY;

--
-- Name: users; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

--
-- PostgreSQL database dump complete
--

\unrestrict 8fgD89hm54Q1WpmRaOWe2tmp7vM3sxUhpLy4gQ3SmeMsddYG6TOnF17xJYHR4MW

