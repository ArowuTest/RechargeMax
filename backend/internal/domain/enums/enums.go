// Package enums defines all typed string constants used across the domain.
// Using typed constants instead of bare strings prevents typos, enables
// IDE auto-complete, and makes enum values grep-able in one place.
package enums

// ---------------------------------------------------------------------------
// Transaction / Recharge
// ---------------------------------------------------------------------------

// TransactionStatus represents the lifecycle state of a transaction.
type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "PENDING"
	TransactionStatusProcessing TransactionStatus = "PROCESSING"
	TransactionStatusSuccess    TransactionStatus = "SUCCESS"
	TransactionStatusFailed     TransactionStatus = "FAILED"
	TransactionStatusCancelled  TransactionStatus = "CANCELLED"
	TransactionStatusReversed   TransactionStatus = "REVERSED"
)

// TransactionType distinguishes airtime from data recharges.
type TransactionType string

const (
	TransactionTypeAirtime TransactionType = "airtime"
	TransactionTypeData    TransactionType = "data"
)

// ---------------------------------------------------------------------------
// Draw
// ---------------------------------------------------------------------------

// DrawStatus represents a draw's lifecycle.
type DrawStatus string

const (
	DrawStatusUpcoming  DrawStatus = "UPCOMING"
	DrawStatusActive    DrawStatus = "ACTIVE"
	DrawStatusCompleted DrawStatus = "COMPLETED"
	DrawStatusCancelled DrawStatus = "CANCELLED"
)

// ---------------------------------------------------------------------------
// Winner / Prize Claim
// ---------------------------------------------------------------------------

// ClaimStatus is the state of a winner's prize claim.
type ClaimStatus string

const (
	ClaimStatusPending           ClaimStatus = "PENDING"
	ClaimStatusClaimed           ClaimStatus = "CLAIMED"
	ClaimStatusExpired           ClaimStatus = "EXPIRED"
	ClaimStatusPendingAdminReview ClaimStatus = "PENDING_ADMIN_REVIEW"
	ClaimStatusApproved          ClaimStatus = "APPROVED"
	ClaimStatusRejected          ClaimStatus = "REJECTED"
)

// ProvisionStatus tracks prize fulfilment state.
type ProvisionStatus string

const (
	ProvisionStatusPending    ProvisionStatus = "PENDING"
	ProvisionStatusProcessing ProvisionStatus = "PROCESSING"
	ProvisionStatusCompleted  ProvisionStatus = "COMPLETED"
	ProvisionStatusFailed     ProvisionStatus = "FAILED"
)

// ---------------------------------------------------------------------------
// Spin Wheel
// ---------------------------------------------------------------------------

// SpinClaimStatus mirrors winner claim but for spin-wheel prizes.
type SpinClaimStatus string

const (
	SpinClaimStatusPending  SpinClaimStatus = "PENDING"
	SpinClaimStatusClaimed  SpinClaimStatus = "CLAIMED"
	SpinClaimStatusExpired  SpinClaimStatus = "EXPIRED"
	SpinClaimStatusApproved SpinClaimStatus = "APPROVED"
	SpinClaimStatusRejected SpinClaimStatus = "REJECTED"
)

// ---------------------------------------------------------------------------
// Subscription
// ---------------------------------------------------------------------------

// SubscriptionStatus represents a daily subscription's state.
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusPaused    SubscriptionStatus = "paused"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
)

// ---------------------------------------------------------------------------
// Affiliate / Commission
// ---------------------------------------------------------------------------

// AffiliateStatus represents the approval state of an affiliate account.
type AffiliateStatus string

const (
	AffiliateStatusPending   AffiliateStatus = "PENDING"
	AffiliateStatusActive    AffiliateStatus = "ACTIVE"
	AffiliateStatusSuspended AffiliateStatus = "SUSPENDED"
	AffiliateStatusRejected  AffiliateStatus = "REJECTED"
)

// CommissionStatus represents the release state of a commission record.
type CommissionStatus string

const (
	CommissionStatusPending  CommissionStatus = "PENDING"
	CommissionStatusApproved CommissionStatus = "APPROVED"
	CommissionStatusPaid     CommissionStatus = "PAID"
	CommissionStatusRejected CommissionStatus = "REJECTED"
)

// PayoutStatus represents the state of an affiliate payout.
type PayoutStatus string

const (
	PayoutStatusPending    PayoutStatus = "PENDING"
	PayoutStatusProcessing PayoutStatus = "PROCESSING"
	PayoutStatusCompleted  PayoutStatus = "COMPLETED"
	PayoutStatusFailed     PayoutStatus = "FAILED"
	PayoutStatusCancelled  PayoutStatus = "CANCELLED"
)

// ---------------------------------------------------------------------------
// Wallet
// ---------------------------------------------------------------------------

// WalletTransactionType categorises a wallet ledger entry.
type WalletTransactionType string

const (
	WalletTxCredit          WalletTransactionType = "credit"
	WalletTxDebit           WalletTransactionType = "debit"
	WalletTxPendingCredit   WalletTransactionType = "pending_credit"
	WalletTxPendingRelease  WalletTransactionType = "pending_release"
)

// WalletTransactionStatus represents the settled state of a wallet entry.
type WalletTransactionStatus string

const (
	WalletTxStatusPending   WalletTransactionStatus = "pending"
	WalletTxStatusCompleted WalletTransactionStatus = "completed"
	WalletTxStatusFailed    WalletTransactionStatus = "failed"
	WalletTxStatusReversed  WalletTransactionStatus = "reversed"
)

// ---------------------------------------------------------------------------
// Notification
// ---------------------------------------------------------------------------

// NotificationPriority sets the urgency level of a notification.
type NotificationPriority string

const (
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityLow    NotificationPriority = "low"
)

// NotificationType classifies the event that triggered the notification.
type NotificationType string

const (
	NotificationTypeDrawWinner          NotificationType = "draw_winner"
	NotificationTypePrizeClaimed        NotificationType = "prize_claimed"
	NotificationTypePayoutCompleted     NotificationType = "payout_completed"
	NotificationTypeSpinWin             NotificationType = "spin_win"
	NotificationTypeCommissionEarned    NotificationType = "commission_earned"
	NotificationTypeWithdrawalProcessed NotificationType = "withdrawal_processed"
	NotificationTypeSystem              NotificationType = "system"
	NotificationTypeAnnouncement        NotificationType = "announcement"
)

// ---------------------------------------------------------------------------
// Loyalty Tier
// ---------------------------------------------------------------------------

// LoyaltyTier represents a user's reward tier.
type LoyaltyTier string

const (
	LoyaltyTierBronze   LoyaltyTier = "BRONZE"
	LoyaltyTierSilver   LoyaltyTier = "SILVER"
	LoyaltyTierGold     LoyaltyTier = "GOLD"
	LoyaltyTierPlatinum LoyaltyTier = "PLATINUM"
)

// ---------------------------------------------------------------------------
// Network
// ---------------------------------------------------------------------------

// NetworkCode identifies a Nigerian mobile network operator.
type NetworkCode string

const (
	NetworkMTN     NetworkCode = "MTN"
	NetworkAirtel  NetworkCode = "AIRTEL"
	NetworkGlo     NetworkCode = "GLO"
	Network9Mobile NetworkCode = "9MOBILE"
)

// ---------------------------------------------------------------------------
// OTP
// ---------------------------------------------------------------------------

// OTPPurpose classifies why an OTP was issued.
type OTPPurpose string

const (
	OTPPurposeLogin          OTPPurpose = "login"
	OTPPurposeRegistration   OTPPurpose = "registration"
	OTPPurposePasswordReset  OTPPurpose = "password_reset"
	OTPPurposeWithdrawal     OTPPurpose = "withdrawal"
	OTPPurposePrizeClaim     OTPPurpose = "prize_claim"
)
