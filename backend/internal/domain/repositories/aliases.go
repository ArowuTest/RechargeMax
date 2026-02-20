package repositories

// Type aliases for backward compatibility
// These allow services written for more granular repositories
// to work with the consolidated repository implementations

// ReferralRepository - Referrals are tracked in Users table via referred_by field
// Use UserRepository for referral operations
type ReferralRepository = UserRepository

// RechargeRepository is an alias for TransactionRepository
// Recharges are a type of transaction in the system
type RechargeRepository = TransactionRepository

// SpinResultRepository is an alias for SpinRepository
// Spin results are managed by the spin repository
type SpinResultRepository = SpinRepository
