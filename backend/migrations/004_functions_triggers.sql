-- ============================================================================
-- DATABASE FUNCTIONS AND TRIGGERS
-- Automated business logic and system operations
-- Created: 2026-01-30 14:00 UTC
-- ============================================================================

-- ============================================================================
-- UTILITY FUNCTIONS
-- ============================================================================

-- Function to generate unique referral codes
CREATE OR REPLACE FUNCTION generate_referral_code(user_name TEXT)
RETURNS TEXT AS $$
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
$$ LANGUAGE plpgsql;

-- Function to calculate loyalty tier based on total recharge amount
CREATE OR REPLACE FUNCTION calculate_loyalty_tier(total_amount DECIMAL)
RETURNS TEXT AS $$
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
$$ LANGUAGE plpgsql;

-- Function to calculate points earned from amount
CREATE OR REPLACE FUNCTION calculate_points_earned(amount DECIMAL)
RETURNS INTEGER AS $$
DECLARE
    points_per_naira INTEGER;
BEGIN
    -- Get points per naira from settings (default to 1)
    SELECT COALESCE((setting_value::INTEGER), 1) INTO points_per_naira
    FROM public.platform_settings 
    WHERE setting_key = 'points_per_naira';
    
    RETURN (amount * points_per_naira)::INTEGER;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate draw entries from points
CREATE OR REPLACE FUNCTION calculate_draw_entries(points INTEGER)
RETURNS INTEGER AS $$
DECLARE
    entries_per_points INTEGER;
BEGIN
    -- Get draw entries per 200 points from settings (default to 1)
    SELECT COALESCE((setting_value::INTEGER), 1) INTO entries_per_points
    FROM public.platform_settings 
    WHERE setting_key = 'draw_entries_per_200_points';
    
    RETURN (points / 200 * entries_per_points)::INTEGER;
END;
$$ LANGUAGE plpgsql;

-- Function to get active wheel prizes for a recharge amount
CREATE OR REPLACE FUNCTION get_eligible_wheel_prizes(recharge_amount DECIMAL)
RETURNS TABLE(
    prize_id UUID,
    prize_name TEXT,
    prize_type TEXT,
    prize_value DECIMAL,
    probability DECIMAL,
    icon_name TEXT,
    color_scheme TEXT
) AS $$
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
$$ LANGUAGE plpgsql;

-- Function to select random wheel prize based on probabilities
CREATE OR REPLACE FUNCTION select_wheel_prize(recharge_amount DECIMAL)
RETURNS TABLE(
    prize_id UUID,
    prize_name TEXT,
    prize_type TEXT,
    prize_value DECIMAL
) AS $$
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
$$ LANGUAGE plpgsql;

-- ============================================================================
-- USER MANAGEMENT FUNCTIONS
-- ============================================================================

-- Function to create or update user profile
CREATE OR REPLACE FUNCTION upsert_user_profile(
    p_auth_user_id UUID,
    p_msisdn TEXT,
    p_full_name TEXT DEFAULT NULL,
    p_email TEXT DEFAULT NULL
)
RETURNS UUID AS $$
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
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to update user statistics after transaction
CREATE OR REPLACE FUNCTION update_user_statistics(p_user_id UUID)
RETURNS VOID AS $$
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
$$ LANGUAGE plpgsql;

-- ============================================================================
-- TRANSACTION PROCESSING FUNCTIONS
-- ============================================================================

-- Function to process successful transaction
CREATE OR REPLACE FUNCTION process_successful_transaction(p_transaction_id UUID)
RETURNS VOID AS $$
DECLARE
    transaction_record RECORD;
    points_earned INTEGER;
    draw_entries INTEGER;
    spin_wheel_minimum DECIMAL;
    prize_record RECORD;
BEGIN
    -- Get transaction details
    SELECT * INTO transaction_record
    FROM public.transactions
    WHERE id = p_transaction_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Transaction not found: %', p_transaction_id;
    END IF;
    
    -- Calculate points and draw entries
    points_earned := calculate_points_earned(transaction_record.amount);
    draw_entries := calculate_draw_entries(points_earned);
    
    -- Update transaction with calculated values
    UPDATE public.transactions
    SET 
        points_earned = points_earned,
        draw_entries = draw_entries,
        completed_at = NOW()
    WHERE id = p_transaction_id;
    
    -- Update user points if user exists
    IF transaction_record.user_id IS NOT NULL THEN
        UPDATE public.users
        SET 
            total_points = total_points + points_earned,
            updated_at = NOW()
        WHERE id = transaction_record.user_id;
        
        -- Update user statistics
        PERFORM update_user_statistics(transaction_record.user_id);
    END IF;
    
    -- Check if eligible for spin wheel
    SELECT COALESCE((setting_value::DECIMAL), 1000) INTO spin_wheel_minimum
    FROM public.platform_settings 
    WHERE setting_key = 'spin_wheel_minimum';
    
    IF transaction_record.amount >= spin_wheel_minimum THEN
        -- Select and create spin result
        SELECT * INTO prize_record
        FROM select_wheel_prize(transaction_record.amount)
        LIMIT 1;
        
        IF FOUND THEN
            INSERT INTO public.spin_results (
                id, user_id, transaction_id, msisdn, prize_id, 
                prize_name, prize_type, prize_value, claim_status
            ) VALUES (
                uuid_generate_v4(),
                transaction_record.user_id,
                p_transaction_id,
                transaction_record.msisdn,
                prize_record.prize_id,
                prize_record.prize_name,
                prize_record.prize_type,
                prize_record.prize_value,
                'PENDING'
            );
        END IF;
    END IF;
    
    -- Add draw entries to active draws
    IF draw_entries > 0 THEN
        INSERT INTO public.draw_entries (
            id, draw_id, user_id, msisdn, entries_count, source_type, source_transaction_id
        )
        SELECT 
            uuid_generate_v4(),
            d.id,
            transaction_record.user_id,
            transaction_record.msisdn,
            draw_entries,
            'TRANSACTION',
            p_transaction_id
        FROM public.draws d
        WHERE d.status = 'ACTIVE' AND d.type = 'DAILY';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- AFFILIATE SYSTEM FUNCTIONS
-- ============================================================================

-- Function to process affiliate commission
CREATE OR REPLACE FUNCTION process_affiliate_commission(
    p_transaction_id UUID,
    p_affiliate_code TEXT DEFAULT NULL
)
RETURNS VOID AS $$
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
$$ LANGUAGE plpgsql;

-- ============================================================================
-- DRAW SYSTEM FUNCTIONS
-- ============================================================================

-- Function to conduct draw and select winners
CREATE OR REPLACE FUNCTION conduct_draw(p_draw_id UUID)
RETURNS TABLE(winner_user_id UUID, winner_msisdn TEXT, prize_amount DECIMAL) AS $$
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
$$ LANGUAGE plpgsql;

-- ============================================================================
-- TRIGGER FUNCTION TO UPDATE DRAW ENTRIES
-- ============================================================================

-- Trigger function to update draw total entries
CREATE OR REPLACE FUNCTION trigger_update_draw_entries()
RETURNS TRIGGER AS $$
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
$$ LANGUAGE plpgsql;

-- ============================================================================
-- TRIGGER FUNCTION TO PROCESS TRANSACTIONS
-- ============================================================================

-- Trigger function to process transaction completion
CREATE OR REPLACE FUNCTION trigger_process_transaction()
RETURNS TRIGGER AS $$
BEGIN
    -- Only process when status changes to SUCCESS
    IF NEW.status = 'SUCCESS' AND (OLD.status IS NULL OR OLD.status != 'SUCCESS') THEN
        PERFORM process_successful_transaction(NEW.id);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- CREATE TRIGGERS
-- ============================================================================

-- Draw entries update trigger
CREATE TRIGGER update_draw_entries_trigger
    AFTER INSERT OR UPDATE OR DELETE ON public.draw_entries
    FOR EACH ROW EXECUTE FUNCTION trigger_update_draw_entries();

-- Transaction processing trigger
CREATE TRIGGER process_transaction_trigger
    AFTER UPDATE ON public.transactions
    FOR EACH ROW EXECUTE FUNCTION trigger_process_transaction();

-- ============================================================================
-- ADMIN FUNCTIONS
-- ============================================================================

-- Function to get platform statistics
CREATE OR REPLACE FUNCTION get_platform_statistics()
RETURNS TABLE(
    total_users INTEGER,
    total_transactions INTEGER,
    total_revenue DECIMAL,
    active_affiliates INTEGER,
    pending_prizes INTEGER,
    active_draws INTEGER
) AS $$
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
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get user activity summary
CREATE OR REPLACE FUNCTION get_user_activity_summary(p_user_id UUID)
RETURNS TABLE(
    total_recharge DECIMAL,
    total_points INTEGER,
    loyalty_tier TEXT,
    pending_prizes INTEGER,
    draw_entries INTEGER,
    affiliate_earnings DECIMAL
) AS $$
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
$$ LANGUAGE plpgsql SECURITY DEFINER;

