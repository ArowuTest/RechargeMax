-- Migration: Add Performance Indexes
-- Description: Adds indexes to improve query performance across all tables
-- Created: 2026-02-01
-- Champion Developer Review: Database Optimization

-- ============================================================================
-- TRANSACTIONS TABLE INDEXES
-- ============================================================================

-- Index for user transaction history queries
CREATE INDEX IF NOT EXISTS idx_transactions_user_created 
ON transactions(user_id, created_at DESC) 
WHERE deleted_at IS NULL;

-- Index for status-based queries (pending, processing)
CREATE INDEX IF NOT EXISTS idx_transactions_status_created 
ON transactions(status, created_at DESC) 
WHERE deleted_at IS NULL;

-- Index for payment reference lookups (webhook processing)
CREATE INDEX IF NOT EXISTS idx_transactions_payment_ref 
ON transactions(payment_reference) 
WHERE payment_reference IS NOT NULL;

-- Index for network provider analytics
CREATE INDEX IF NOT EXISTS idx_transactions_network_created 
ON transactions(network_provider, created_at DESC);

-- Index for recharge type filtering
CREATE INDEX IF NOT EXISTS idx_transactions_type_created 
ON transactions(recharge_type, created_at DESC);

-- Composite index for user + status queries
CREATE INDEX IF NOT EXISTS idx_transactions_user_status 
ON transactions(user_id, status) 
WHERE deleted_at IS NULL;

-- ============================================================================
-- VTU_TRANSACTIONS TABLE INDEXES
-- ============================================================================

-- Index for user VTU transaction history
CREATE INDEX IF NOT EXISTS idx_vtu_user_created 
ON vtu_transactions(user_id, created_at DESC) 
WHERE user_id IS NOT NULL;

-- Index for phone number lookups
CREATE INDEX IF NOT EXISTS idx_vtu_phone_created 
ON vtu_transactions(phone_number, created_at DESC);

-- Index for status-based queries
CREATE INDEX IF NOT EXISTS idx_vtu_status_created 
ON vtu_transactions(status, created_at DESC);

-- Index for provider reference lookups
CREATE INDEX IF NOT EXISTS idx_vtu_provider_ref 
ON vtu_transactions(provider_reference) 
WHERE provider_reference IS NOT NULL;

-- Index for retry processing
CREATE INDEX IF NOT EXISTS idx_vtu_retry 
ON vtu_transactions(status, retry_count) 
WHERE status IN ('PENDING', 'FAILED') AND retry_count < max_retries;

-- ============================================================================
-- USERS TABLE INDEXES
-- ============================================================================

-- Index for email lookups (login)
CREATE INDEX IF NOT EXISTS idx_users_email 
ON users(email) 
WHERE email IS NOT NULL AND deleted_at IS NULL;

-- Index for referral code lookups
CREATE INDEX IF NOT EXISTS idx_users_referral_code 
ON users(referral_code) 
WHERE referral_code IS NOT NULL;

-- Index for loyalty tier analytics
CREATE INDEX IF NOT EXISTS idx_users_loyalty_tier 
ON users(loyalty_tier, total_recharge_amount DESC);

-- Index for active users
CREATE INDEX IF NOT EXISTS idx_users_active 
ON users(is_active, last_login_at DESC) 
WHERE deleted_at IS NULL;

-- ============================================================================
-- WHEEL_SPINS TABLE INDEXES
-- ============================================================================

-- Index for user spin history
CREATE INDEX IF NOT EXISTS idx_spins_user_created 
ON wheel_spins(user_id, created_at DESC);

-- Index for daily spin eligibility checks
CREATE INDEX IF NOT EXISTS idx_spins_user_date 
ON wheel_spins(user_id, DATE(created_at));

-- Index for prize tracking
CREATE INDEX IF NOT EXISTS idx_spins_prize_created 
ON wheel_spins(prize_id, created_at DESC) 
WHERE prize_id IS NOT NULL;

-- ============================================================================
-- SPIN_RESULTS TABLE INDEXES
-- ============================================================================

-- Index for user spin results
CREATE INDEX IF NOT EXISTS idx_spin_results_user 
ON spin_results(user_id, created_at DESC);

-- Index for prize value analytics
CREATE INDEX IF NOT EXISTS idx_spin_results_prize_value 
ON spin_results(prize_value DESC, created_at DESC);

-- ============================================================================
-- WALLET_TRANSACTIONS TABLE INDEXES
-- ============================================================================

-- Index for user wallet history
CREATE INDEX IF NOT EXISTS idx_wallet_user_created 
ON wallet_transactions(user_id, created_at DESC);

-- Index for transaction type filtering
CREATE INDEX IF NOT EXISTS idx_wallet_type_created 
ON wallet_transactions(transaction_type, created_at DESC);

-- Index for status-based queries
CREATE INDEX IF NOT EXISTS idx_wallet_status_created 
ON wallet_transactions(status, created_at DESC);

-- Index for reference lookups
CREATE INDEX IF NOT EXISTS idx_wallet_reference 
ON wallet_transactions(reference);

-- ============================================================================
-- WITHDRAWAL_REQUESTS TABLE INDEXES
-- ============================================================================

-- Index for user withdrawal history
CREATE INDEX IF NOT EXISTS idx_withdrawal_user_created 
ON withdrawal_requests(user_id, requested_at DESC);

-- Index for status-based admin queries
CREATE INDEX IF NOT EXISTS idx_withdrawal_status_requested 
ON withdrawal_requests(status, requested_at DESC);

-- Index for pending approvals
CREATE INDEX IF NOT EXISTS idx_withdrawal_pending 
ON withdrawal_requests(status, requested_at ASC) 
WHERE status = 'PENDING';

-- ============================================================================
-- OTP_CODES TABLE INDEXES
-- ============================================================================

-- Index for OTP verification (most common query)
CREATE INDEX IF NOT EXISTS idx_otp_msisdn_code 
ON otp_codes(msisdn, code, expires_at) 
WHERE is_used = false;

-- Index for cleanup of expired OTPs
CREATE INDEX IF NOT EXISTS idx_otp_expires 
ON otp_codes(expires_at ASC) 
WHERE is_used = false;

-- ============================================================================
-- PAYMENT_LOGS TABLE INDEXES
-- ============================================================================

-- Index for transaction reference lookups
CREATE INDEX IF NOT EXISTS idx_payment_logs_txn_ref 
ON payment_logs(transaction_reference);

-- Index for user payment history
CREATE INDEX IF NOT EXISTS idx_payment_logs_user_created 
ON payment_logs(user_id, created_at DESC) 
WHERE user_id IS NOT NULL;

-- Index for gateway + status queries
CREATE INDEX IF NOT EXISTS idx_payment_logs_gateway_status 
ON payment_logs(gateway, status, created_at DESC);

-- ============================================================================
-- DRAWS TABLE INDEXES
-- ============================================================================

-- Index for active draws
CREATE INDEX IF NOT EXISTS idx_draws_status_start 
ON draws(status, start_date DESC);

-- Index for upcoming draws
CREATE INDEX IF NOT EXISTS idx_draws_upcoming 
ON draws(start_date ASC) 
WHERE status = 'UPCOMING';

-- ============================================================================
-- DRAW_ENTRIES TABLE INDEXES
-- ============================================================================

-- Index for user entries in a draw
CREATE INDEX IF NOT EXISTS idx_draw_entries_user_draw 
ON draw_entries(user_id, draw_id);

-- Index for draw entries count
CREATE INDEX IF NOT EXISTS idx_draw_entries_draw_created 
ON draw_entries(draw_id, created_at DESC);

-- ============================================================================
-- DRAW_WINNERS TABLE INDEXES
-- ============================================================================

-- Index for user wins
CREATE INDEX IF NOT EXISTS idx_draw_winners_user 
ON draw_winners(user_id, created_at DESC) 
WHERE user_id IS NOT NULL;

-- Index for draw winners
CREATE INDEX IF NOT EXISTS idx_draw_winners_draw 
ON draw_winners(draw_id, position ASC);

-- Index for unclaimed prizes
CREATE INDEX IF NOT EXISTS idx_draw_winners_unclaimed 
ON draw_winners(claimed_at, expires_at) 
WHERE claimed_at IS NULL AND expires_at > NOW();

-- ============================================================================
-- AFFILIATE_COMMISSIONS TABLE INDEXES
-- ============================================================================

-- Index for affiliate earnings
CREATE INDEX IF NOT EXISTS idx_affiliate_comm_affiliate 
ON affiliate_commissions(affiliate_user_id, created_at DESC);

-- Index for referral tracking
CREATE INDEX IF NOT EXISTS idx_affiliate_comm_referred 
ON affiliate_commissions(referred_user_id);

-- Index for payout processing
CREATE INDEX IF NOT EXISTS idx_affiliate_comm_payout_status 
ON affiliate_commissions(payout_status, created_at ASC);

-- ============================================================================
-- DATA_PLANS TABLE INDEXES
-- ============================================================================

-- Index for active plans by network
CREATE INDEX IF NOT EXISTS idx_data_plans_network_active 
ON data_plans(network_provider, is_active) 
WHERE is_active = true;

-- Index for plan code lookups
CREATE INDEX IF NOT EXISTS idx_data_plans_code 
ON data_plans(plan_code) 
WHERE is_active = true;

-- ============================================================================
-- NOTIFICATIONS TABLE INDEXES
-- ============================================================================

-- Index for user notifications
CREATE INDEX IF NOT EXISTS idx_notifications_user_created 
ON notifications(user_id, created_at DESC);

-- Index for unread notifications
CREATE INDEX IF NOT EXISTS idx_notifications_unread 
ON notifications(user_id, is_read, created_at DESC) 
WHERE is_read = false;

-- ============================================================================
-- VERIFICATION
-- ============================================================================

-- Query to verify all indexes were created
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
    AND indexname LIKE 'idx_%'
ORDER BY tablename, indexname;

-- ============================================================================
-- ROLLBACK
-- ============================================================================

/*
-- To rollback, drop all indexes:

DROP INDEX IF EXISTS idx_transactions_user_created;
DROP INDEX IF EXISTS idx_transactions_status_created;
DROP INDEX IF EXISTS idx_transactions_payment_ref;
DROP INDEX IF EXISTS idx_transactions_network_created;
DROP INDEX IF EXISTS idx_transactions_type_created;
DROP INDEX IF EXISTS idx_transactions_user_status;

DROP INDEX IF EXISTS idx_vtu_user_created;
DROP INDEX IF EXISTS idx_vtu_phone_created;
DROP INDEX IF EXISTS idx_vtu_status_created;
DROP INDEX IF EXISTS idx_vtu_provider_ref;
DROP INDEX IF EXISTS idx_vtu_retry;

DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_referral_code;
DROP INDEX IF EXISTS idx_users_loyalty_tier;
DROP INDEX IF EXISTS idx_users_active;

DROP INDEX IF EXISTS idx_spins_user_created;
DROP INDEX IF EXISTS idx_spins_user_date;
DROP INDEX IF EXISTS idx_spins_prize_created;

DROP INDEX IF EXISTS idx_spin_results_user;
DROP INDEX IF EXISTS idx_spin_results_prize_value;

DROP INDEX IF EXISTS idx_wallet_user_created;
DROP INDEX IF EXISTS idx_wallet_type_created;
DROP INDEX IF EXISTS idx_wallet_status_created;
DROP INDEX IF EXISTS idx_wallet_reference;

DROP INDEX IF EXISTS idx_withdrawal_user_created;
DROP INDEX IF EXISTS idx_withdrawal_status_requested;
DROP INDEX IF EXISTS idx_withdrawal_pending;

DROP INDEX IF EXISTS idx_otp_msisdn_code;
DROP INDEX IF EXISTS idx_otp_expires;

DROP INDEX IF EXISTS idx_payment_logs_txn_ref;
DROP INDEX IF EXISTS idx_payment_logs_user_created;
DROP INDEX IF EXISTS idx_payment_logs_gateway_status;

DROP INDEX IF EXISTS idx_draws_status_start;
DROP INDEX IF EXISTS idx_draws_upcoming;

DROP INDEX IF EXISTS idx_draw_entries_user_draw;
DROP INDEX IF EXISTS idx_draw_entries_draw_created;

DROP INDEX IF EXISTS idx_draw_winners_user;
DROP INDEX IF EXISTS idx_draw_winners_draw;
DROP INDEX IF EXISTS idx_draw_winners_unclaimed;

DROP INDEX IF EXISTS idx_affiliate_comm_affiliate;
DROP INDEX IF EXISTS idx_affiliate_comm_referred;
DROP INDEX IF EXISTS idx_affiliate_comm_payout_status;

DROP INDEX IF EXISTS idx_data_plans_network_active;
DROP INDEX IF EXISTS idx_data_plans_code;

DROP INDEX IF EXISTS idx_notifications_user_created;
DROP INDEX IF EXISTS idx_notifications_unread;
*/
