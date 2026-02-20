-- ============================================================================
-- RECHARGEMAX CORE DATABASE SCHEMA
-- Comprehensive database schema for the RechargeMax platform
-- Created: 2026-01-30 14:00 UTC
-- ============================================================================

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- USERS TABLE
-- ============================================================================

CREATE TABLE public.users_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auth_user_id UUID UNIQUE, -- References auth.users(id)
    
    -- Basic user information
    msisdn TEXT UNIQUE NOT NULL,
    full_name TEXT,
    email TEXT,
    phone_verified BOOLEAN DEFAULT false,
    email_verified BOOLEAN DEFAULT false,
    
    -- Profile information
    date_of_birth DATE,
    gender TEXT CHECK (gender IN ('MALE', 'FEMALE', 'OTHER')),
    state TEXT,
    city TEXT,
    address TEXT,
    
    -- Gamification and loyalty
    total_points INTEGER DEFAULT 0,
    loyalty_tier TEXT DEFAULT 'BRONZE' CHECK (loyalty_tier IN ('BRONZE', 'SILVER', 'GOLD', 'PLATINUM')),
    total_recharge_amount DECIMAL(12,2) DEFAULT 0,
    total_transactions INTEGER DEFAULT 0,
    last_recharge_date TIMESTAMP WITH TIME ZONE,
    
    -- Referral system
    referral_code TEXT UNIQUE,
    referred_by UUID REFERENCES public.users_2026_01_30_14_00(id),
    total_referrals INTEGER DEFAULT 0,
    
    -- Account status
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    kyc_status TEXT DEFAULT 'PENDING' CHECK (kyc_status IN ('PENDING', 'VERIFIED', 'REJECTED')),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_msisdn CHECK (msisdn ~ '^234[789][01][0-9]{8}$'),
    CONSTRAINT valid_email CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$' OR email IS NULL)
);

-- ============================================================================
-- ADMIN USERS TABLE
-- ============================================================================

CREATE TABLE public.admin_users_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Admin credentials
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    
    -- Admin profile
    full_name TEXT NOT NULL,
    role TEXT DEFAULT 'ADMIN' CHECK (role IN ('SUPER_ADMIN', 'ADMIN', 'MODERATOR', 'SUPPORT')),
    
    -- Permissions (JSON array of permission strings)
    permissions JSONB DEFAULT '[]'::jsonb,
    
    -- Admin status
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE,
    login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_admin_email CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- ============================================================================
-- ADMIN SESSIONS TABLE
-- ============================================================================

CREATE TABLE public.admin_sessions_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_user_id UUID REFERENCES public.admin_users_2026_01_30_14_00(id) ON DELETE CASCADE,
    
    -- Session details
    session_token TEXT UNIQUE NOT NULL,
    ip_address INET,
    user_agent TEXT,
    
    -- Session status
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- NETWORK CONFIGURATIONS TABLE
-- ============================================================================

CREATE TABLE public.network_configs_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Network details
    network_name TEXT NOT NULL,
    network_code TEXT UNIQUE NOT NULL,
    
    -- Network capabilities
    is_active BOOLEAN DEFAULT true,
    airtime_enabled BOOLEAN DEFAULT true,
    data_enabled BOOLEAN DEFAULT true,
    
    -- Pricing and limits
    commission_rate DECIMAL(5,2) DEFAULT 2.50, -- Percentage
    minimum_amount DECIMAL(10,2) DEFAULT 50.00,
    maximum_amount DECIMAL(10,2) DEFAULT 50000.00,
    
    -- Display settings
    logo_url TEXT,
    brand_color TEXT,
    sort_order INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT positive_commission_rate CHECK (commission_rate >= 0),
    CONSTRAINT valid_amount_range CHECK (maximum_amount > minimum_amount)
);

-- ============================================================================
-- DATA PLANS TABLE
-- ============================================================================

CREATE TABLE public.data_plans_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    network_id UUID REFERENCES public.network_configs_2026_01_30_14_00(id) ON DELETE CASCADE,
    
    -- Plan details
    plan_name TEXT NOT NULL,
    data_amount TEXT NOT NULL, -- e.g., "1GB", "500MB"
    price DECIMAL(10,2) NOT NULL,
    validity_days INTEGER NOT NULL,
    plan_code TEXT NOT NULL,
    
    -- Plan status
    is_active BOOLEAN DEFAULT true,
    sort_order INTEGER DEFAULT 0,
    
    -- Plan metadata
    description TEXT,
    terms_and_conditions TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT positive_price CHECK (price > 0),
    CONSTRAINT positive_validity CHECK (validity_days > 0),
    UNIQUE(network_id, plan_code)
);

-- ============================================================================
-- TRANSACTIONS TABLE
-- ============================================================================

CREATE TABLE public.transactions_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id),
    
    -- Transaction details
    msisdn TEXT NOT NULL,
    network_provider TEXT NOT NULL,
    recharge_type TEXT NOT NULL CHECK (recharge_type IN ('AIRTIME', 'DATA')),
    amount DECIMAL(10,2) NOT NULL,
    data_plan_id UUID REFERENCES public.data_plans_2026_01_30_14_00(id),
    
    -- Payment details
    payment_method TEXT NOT NULL CHECK (payment_method IN ('CARD', 'BANK_TRANSFER', 'USSD', 'WALLET')),
    payment_reference TEXT UNIQUE,
    payment_gateway TEXT,
    
    -- Transaction status
    status TEXT DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSING', 'SUCCESS', 'FAILED', 'CANCELLED')),
    provider_reference TEXT,
    provider_response JSONB,
    failure_reason TEXT,
    
    -- Rewards and gamification
    points_earned INTEGER DEFAULT 0,
    draw_entries INTEGER DEFAULT 0,
    spin_eligible BOOLEAN DEFAULT false,
    
    -- Customer information (for guest transactions)
    customer_email TEXT,
    customer_name TEXT,
    
    -- Metadata
    ip_address INET,
    user_agent TEXT,
    affiliate_code TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT positive_amount CHECK (amount > 0),
    CONSTRAINT valid_msisdn CHECK (msisdn ~ '^234[789][01][0-9]{8}$')
);

-- ============================================================================
-- WHEEL PRIZES TABLE
-- ============================================================================

CREATE TABLE public.wheel_prizes_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Prize details
    prize_name TEXT NOT NULL,
    prize_type TEXT NOT NULL CHECK (prize_type IN ('CASH', 'AIRTIME', 'DATA', 'POINTS')),
    prize_value DECIMAL(10,2) NOT NULL,
    
    -- Prize probability and conditions
    probability DECIMAL(5,2) NOT NULL, -- Percentage (0-100)
    minimum_recharge DECIMAL(10,2) DEFAULT 0,
    
    -- Prize status and display
    is_active BOOLEAN DEFAULT true,
    icon_name TEXT,
    color_scheme TEXT,
    sort_order INTEGER DEFAULT 0,
    
    -- Prize metadata
    description TEXT,
    terms_and_conditions TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT positive_prize_value CHECK (prize_value > 0),
    CONSTRAINT valid_probability CHECK (probability >= 0 AND probability <= 100)
);

-- ============================================================================
-- SPIN RESULTS TABLE
-- ============================================================================

CREATE TABLE public.spin_results_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id),
    transaction_id UUID REFERENCES public.transactions_2026_01_30_14_00(id),
    
    -- Spin details
    msisdn TEXT NOT NULL,
    prize_id UUID REFERENCES public.wheel_prizes_2026_01_30_14_00(id),
    prize_name TEXT NOT NULL,
    prize_type TEXT NOT NULL,
    prize_value DECIMAL(10,2) NOT NULL,
    
    -- Claim status
    claim_status TEXT DEFAULT 'PENDING' CHECK (claim_status IN ('PENDING', 'CLAIMED', 'EXPIRED')),
    claimed_at TIMESTAMP WITH TIME ZONE,
    claim_reference TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '30 days'),
    
    -- Constraints
    CONSTRAINT positive_prize_value CHECK (prize_value >= 0)
);

-- ============================================================================
-- DAILY SUBSCRIPTION CONFIG TABLE
-- ============================================================================

CREATE TABLE public.daily_subscription_config_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Subscription configuration
    amount DECIMAL(5,2) NOT NULL,
    draw_entries_earned INTEGER DEFAULT 1,
    is_paid BOOLEAN DEFAULT true,
    
    -- Configuration details
    description TEXT,
    terms_and_conditions TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT positive_amount CHECK (amount > 0),
    CONSTRAINT positive_entries CHECK (draw_entries_earned > 0)
);

-- ============================================================================
-- DAILY SUBSCRIPTIONS TABLE
-- ============================================================================

CREATE TABLE public.daily_subscriptions_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id),
    
    -- Subscription details
    msisdn TEXT NOT NULL,
    subscription_date DATE NOT NULL,
    amount DECIMAL(5,2) NOT NULL,
    draw_entries_earned INTEGER DEFAULT 1,
    points_earned INTEGER DEFAULT 0,
    
    -- Payment details
    payment_reference TEXT,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'cancelled', 'expired')),
    is_paid BOOLEAN DEFAULT false,
    
    -- Customer information (for guest subscriptions)
    customer_email TEXT,
    customer_name TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT positive_amount CHECK (amount > 0),
    CONSTRAINT valid_msisdn CHECK (msisdn ~ '^234[789][01][0-9]{8}$'),
    UNIQUE(user_id, subscription_date)
);

-- ============================================================================
-- DRAWS TABLE
-- ============================================================================

CREATE TABLE public.draws_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Draw details
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('DAILY', 'WEEKLY', 'MONTHLY', 'SPECIAL')),
    description TEXT,
    
    -- Draw configuration
    status TEXT DEFAULT 'UPCOMING' CHECK (status IN ('UPCOMING', 'ACTIVE', 'COMPLETED', 'CANCELLED')),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    draw_time TIMESTAMP WITH TIME ZONE,
    
    -- Prize configuration
    prize_pool DECIMAL(12,2) NOT NULL,
    winners_count INTEGER DEFAULT 1,
    entry_cost DECIMAL(5,2) DEFAULT 0,
    
    -- Draw statistics
    total_entries INTEGER DEFAULT 0,
    
    -- Draw results
    results JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT positive_prize_pool CHECK (prize_pool > 0),
    CONSTRAINT positive_winners_count CHECK (winners_count > 0),
    CONSTRAINT valid_timing CHECK (end_time > start_time)
);

-- ============================================================================
-- DRAW ENTRIES TABLE
-- ============================================================================

CREATE TABLE public.draw_entries_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    draw_id UUID REFERENCES public.draws_2026_01_30_14_00(id) ON DELETE CASCADE,
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id),
    
    -- Entry details
    msisdn TEXT NOT NULL,
    entries_count INTEGER DEFAULT 1,
    
    -- Entry source
    source_type TEXT NOT NULL CHECK (source_type IN ('TRANSACTION', 'SUBSCRIPTION', 'BONUS', 'MANUAL')),
    source_transaction_id UUID REFERENCES public.transactions_2026_01_30_14_00(id),
    source_subscription_id UUID REFERENCES public.daily_subscriptions_2026_01_30_14_00(id),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT positive_entries CHECK (entries_count > 0)
);

-- ============================================================================
-- DRAW WINNERS TABLE
-- ============================================================================

CREATE TABLE public.draw_winners_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    draw_id UUID REFERENCES public.draws_2026_01_30_14_00(id) ON DELETE CASCADE,
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id),
    
    -- Winner details
    msisdn TEXT NOT NULL,
    position INTEGER NOT NULL, -- 1st, 2nd, 3rd place
    prize_amount DECIMAL(10,2) NOT NULL,
    
    -- Claim status
    claim_status TEXT DEFAULT 'PENDING' CHECK (claim_status IN ('PENDING', 'CLAIMED', 'EXPIRED')),
    claimed_at TIMESTAMP WITH TIME ZONE,
    claim_reference TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '30 days'),
    
    -- Constraints
    CONSTRAINT positive_prize_amount CHECK (prize_amount > 0),
    CONSTRAINT positive_position CHECK (position > 0),
    UNIQUE(draw_id, position)
);

-- ============================================================================
-- AFFILIATES TABLE
-- ============================================================================

CREATE TABLE public.affiliates_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id) ON DELETE CASCADE,
    
    -- Affiliate details
    affiliate_code TEXT UNIQUE NOT NULL,
    status TEXT DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'SUSPENDED', 'REJECTED')),
    
    -- Affiliate tier and commission
    tier TEXT DEFAULT 'BRONZE' CHECK (tier IN ('BRONZE', 'SILVER', 'GOLD', 'PLATINUM')),
    commission_rate DECIMAL(5,2) DEFAULT 5.00, -- Percentage
    
    -- Affiliate statistics
    total_referrals INTEGER DEFAULT 0,
    active_referrals INTEGER DEFAULT 0,
    total_commission DECIMAL(10,2) DEFAULT 0,
    
    -- Affiliate profile
    business_name TEXT,
    website_url TEXT,
    social_media_handles JSONB,
    
    -- Bank details for payouts
    bank_name TEXT,
    account_number TEXT,
    account_name TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    approved_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT positive_commission_rate CHECK (commission_rate >= 0)
);

-- ============================================================================
-- AFFILIATE CLICKS TABLE
-- ============================================================================

CREATE TABLE public.affiliate_clicks_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    affiliate_id UUID REFERENCES public.affiliates_2026_01_30_14_00(id) ON DELETE CASCADE,
    
    -- Click details
    ip_address INET,
    user_agent TEXT,
    referrer_url TEXT,
    landing_page TEXT,
    
    -- Conversion tracking
    converted BOOLEAN DEFAULT false,
    conversion_transaction_id UUID REFERENCES public.transactions_2026_01_30_14_00(id),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    converted_at TIMESTAMP WITH TIME ZONE
);

-- ============================================================================
-- AFFILIATE COMMISSIONS TABLE
-- ============================================================================

CREATE TABLE public.affiliate_commissions_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    affiliate_id UUID REFERENCES public.affiliates_2026_01_30_14_00(id) ON DELETE CASCADE,
    transaction_id UUID REFERENCES public.transactions_2026_01_30_14_00(id),
    
    -- Commission details
    commission_amount DECIMAL(10,2) NOT NULL,
    commission_rate DECIMAL(5,2) NOT NULL,
    transaction_amount DECIMAL(10,2) NOT NULL,
    
    -- Commission status
    status TEXT DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'PAID', 'CANCELLED')),
    
    -- Payout details
    payout_reference TEXT,
    payout_method TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    earned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    paid_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT positive_commission_amount CHECK (commission_amount > 0),
    CONSTRAINT positive_transaction_amount CHECK (transaction_amount > 0)
);

-- ============================================================================
-- PLATFORM SETTINGS TABLE
-- ============================================================================

CREATE TABLE public.platform_settings_2026_01_30_14_00 (
    setting_key TEXT PRIMARY KEY,
    setting_value TEXT NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- APPLICATION LOGS TABLE
-- ============================================================================

CREATE TABLE public.application_logs_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Log details
    level TEXT NOT NULL CHECK (level IN ('DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL')),
    message TEXT NOT NULL,
    context JSONB,
    
    -- Request details
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id),
    ip_address INET,
    user_agent TEXT,
    request_id TEXT,
    
    -- Error details
    error_code TEXT,
    stack_trace TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- APPLICATION METRICS TABLE
-- ============================================================================

CREATE TABLE public.application_metrics_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Metric details
    metric_name TEXT NOT NULL,
    metric_value DECIMAL(15,4) NOT NULL,
    metric_unit TEXT,
    
    -- Metric metadata
    tags JSONB,
    dimensions JSONB,
    
    -- Timestamps
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================

-- Users indexes
CREATE INDEX idx_users_auth_user_id ON public.users_2026_01_30_14_00(auth_user_id);
CREATE INDEX idx_users_msisdn ON public.users_2026_01_30_14_00(msisdn);
CREATE INDEX idx_users_referral_code ON public.users_2026_01_30_14_00(referral_code);
CREATE INDEX idx_users_referred_by ON public.users_2026_01_30_14_00(referred_by);
CREATE INDEX idx_users_loyalty_tier ON public.users_2026_01_30_14_00(loyalty_tier);

-- Admin users indexes
CREATE INDEX idx_admin_users_email ON public.admin_users_2026_01_30_14_00(email);
CREATE INDEX idx_admin_users_role ON public.admin_users_2026_01_30_14_00(role);
CREATE INDEX idx_admin_users_is_active ON public.admin_users_2026_01_30_14_00(is_active);

-- Admin sessions indexes
CREATE INDEX idx_admin_sessions_admin_user_id ON public.admin_sessions_2026_01_30_14_00(admin_user_id);
CREATE INDEX idx_admin_sessions_session_token ON public.admin_sessions_2026_01_30_14_00(session_token);
CREATE INDEX idx_admin_sessions_expires_at ON public.admin_sessions_2026_01_30_14_00(expires_at);

-- Network configs indexes
CREATE INDEX idx_network_configs_network_code ON public.network_configs_2026_01_30_14_00(network_code);
CREATE INDEX idx_network_configs_is_active ON public.network_configs_2026_01_30_14_00(is_active);
CREATE INDEX idx_network_configs_sort_order ON public.network_configs_2026_01_30_14_00(sort_order);

-- Data plans indexes
CREATE INDEX idx_data_plans_network_id ON public.data_plans_2026_01_30_14_00(network_id);
CREATE INDEX idx_data_plans_is_active ON public.data_plans_2026_01_30_14_00(is_active);
CREATE INDEX idx_data_plans_price ON public.data_plans_2026_01_30_14_00(price);

-- Transactions indexes
CREATE INDEX idx_transactions_user_id ON public.transactions_2026_01_30_14_00(user_id);
CREATE INDEX idx_transactions_msisdn ON public.transactions_2026_01_30_14_00(msisdn);
CREATE INDEX idx_transactions_status ON public.transactions_2026_01_30_14_00(status);
CREATE INDEX idx_transactions_created_at ON public.transactions_2026_01_30_14_00(created_at DESC);
CREATE INDEX idx_transactions_network_provider ON public.transactions_2026_01_30_14_00(network_provider);
CREATE INDEX idx_transactions_payment_reference ON public.transactions_2026_01_30_14_00(payment_reference);

-- Wheel prizes indexes
CREATE INDEX idx_wheel_prizes_is_active ON public.wheel_prizes_2026_01_30_14_00(is_active);
CREATE INDEX idx_wheel_prizes_prize_type ON public.wheel_prizes_2026_01_30_14_00(prize_type);
CREATE INDEX idx_wheel_prizes_sort_order ON public.wheel_prizes_2026_01_30_14_00(sort_order);

-- Spin results indexes
CREATE INDEX idx_spin_results_user_id ON public.spin_results_2026_01_30_14_00(user_id);
CREATE INDEX idx_spin_results_transaction_id ON public.spin_results_2026_01_30_14_00(transaction_id);
CREATE INDEX idx_spin_results_claim_status ON public.spin_results_2026_01_30_14_00(claim_status);
CREATE INDEX idx_spin_results_created_at ON public.spin_results_2026_01_30_14_00(created_at DESC);

-- Daily subscriptions indexes
CREATE INDEX idx_daily_subscriptions_user_id ON public.daily_subscriptions_2026_01_30_14_00(user_id);
CREATE INDEX idx_daily_subscriptions_msisdn ON public.daily_subscriptions_2026_01_30_14_00(msisdn);
CREATE INDEX idx_daily_subscriptions_subscription_date ON public.daily_subscriptions_2026_01_30_14_00(subscription_date);
CREATE INDEX idx_daily_subscriptions_status ON public.daily_subscriptions_2026_01_30_14_00(status);

-- Draws indexes
CREATE INDEX idx_draws_status ON public.draws_2026_01_30_14_00(status);
CREATE INDEX idx_draws_type ON public.draws_2026_01_30_14_00(type);
CREATE INDEX idx_draws_start_time ON public.draws_2026_01_30_14_00(start_time);
CREATE INDEX idx_draws_end_time ON public.draws_2026_01_30_14_00(end_time);

-- Draw entries indexes
CREATE INDEX idx_draw_entries_draw_id ON public.draw_entries_2026_01_30_14_00(draw_id);
CREATE INDEX idx_draw_entries_user_id ON public.draw_entries_2026_01_30_14_00(user_id);
CREATE INDEX idx_draw_entries_source_type ON public.draw_entries_2026_01_30_14_00(source_type);

-- Draw winners indexes
CREATE INDEX idx_draw_winners_draw_id ON public.draw_winners_2026_01_30_14_00(draw_id);
CREATE INDEX idx_draw_winners_user_id ON public.draw_winners_2026_01_30_14_00(user_id);
CREATE INDEX idx_draw_winners_claim_status ON public.draw_winners_2026_01_30_14_00(claim_status);

-- Affiliates indexes
CREATE INDEX idx_affiliates_user_id ON public.affiliates_2026_01_30_14_00(user_id);
CREATE INDEX idx_affiliates_affiliate_code ON public.affiliates_2026_01_30_14_00(affiliate_code);
CREATE INDEX idx_affiliates_status ON public.affiliates_2026_01_30_14_00(status);
CREATE INDEX idx_affiliates_tier ON public.affiliates_2026_01_30_14_00(tier);

-- Affiliate clicks indexes
CREATE INDEX idx_affiliate_clicks_affiliate_id ON public.affiliate_clicks_2026_01_30_14_00(affiliate_id);
CREATE INDEX idx_affiliate_clicks_created_at ON public.affiliate_clicks_2026_01_30_14_00(created_at DESC);
CREATE INDEX idx_affiliate_clicks_converted ON public.affiliate_clicks_2026_01_30_14_00(converted);

-- Affiliate commissions indexes
CREATE INDEX idx_affiliate_commissions_affiliate_id ON public.affiliate_commissions_2026_01_30_14_00(affiliate_id);
CREATE INDEX idx_affiliate_commissions_transaction_id ON public.affiliate_commissions_2026_01_30_14_00(transaction_id);
CREATE INDEX idx_affiliate_commissions_status ON public.affiliate_commissions_2026_01_30_14_00(status);
CREATE INDEX idx_affiliate_commissions_earned_at ON public.affiliate_commissions_2026_01_30_14_00(earned_at DESC);

-- Application logs indexes
CREATE INDEX idx_application_logs_level ON public.application_logs_2026_01_30_14_00(level);
CREATE INDEX idx_application_logs_user_id ON public.application_logs_2026_01_30_14_00(user_id);
CREATE INDEX idx_application_logs_created_at ON public.application_logs_2026_01_30_14_00(created_at DESC);

-- Application metrics indexes
CREATE INDEX idx_application_metrics_metric_name ON public.application_metrics_2026_01_30_14_00(metric_name);
CREATE INDEX idx_application_metrics_recorded_at ON public.application_metrics_2026_01_30_14_00(recorded_at DESC);

-- ============================================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply triggers to tables with updated_at columns
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON public.users_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_admin_users_updated_at 
    BEFORE UPDATE ON public.admin_users_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_network_configs_updated_at 
    BEFORE UPDATE ON public.network_configs_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_data_plans_updated_at 
    BEFORE UPDATE ON public.data_plans_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at 
    BEFORE UPDATE ON public.transactions_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_wheel_prizes_updated_at 
    BEFORE UPDATE ON public.wheel_prizes_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_daily_subscription_config_updated_at 
    BEFORE UPDATE ON public.daily_subscription_config_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_draws_updated_at 
    BEFORE UPDATE ON public.draws_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_affiliates_updated_at 
    BEFORE UPDATE ON public.affiliates_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_platform_settings_updated_at 
    BEFORE UPDATE ON public.platform_settings_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

