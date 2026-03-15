package entities

// ---------------------------------------------------------------------------
// Backward-compatibility aliases
//
// All entity structs are now SINGULAR (idiomatic Go).
// Services/repos written before the rename still compile via these aliases.
// Remove an alias once all callers have been updated to use the singular name.
// ---------------------------------------------------------------------------

// Core domain — plural aliases
type Draws = Draw
type Users = User
type Affiliates = Affiliate
type Transactions = Transaction
type NetworkConfigs = NetworkConfig
type DataPlans = DataPlan
type DailySubscriptions = DailySubscription

// Admin — plural aliases
type AdminUsers = AdminUser
type AdminSessions = AdminSession
type AdminActivityLogs = AdminActivityLog

// Financial
type BankAccounts = BankAccount
type WithdrawalRequests = WithdrawalRequest
type AffiliateCommissions = AffiliateCommission
type AffiliatePayouts = AffiliatePayout
type AffiliateClicks = AffiliateClick
type AffiliateBankAccounts = AffiliateBankAccount
type AffiliateAnalytics = AffiliateAnalytic

// Draw / prize
type DrawEntries = DrawEntry
type DrawWinners = DrawWinner
type PaymentLogs = PaymentLog
type VtuTransactions = VtuTransaction

// Logging / observability
type ApiLogs = APILog
type WebhookLogs = WebhookLog
type ApplicationLogs = ApplicationLog
type ApplicationMetrics = ApplicationMetric

// Settings / config / notifications
type PlatformSettings = PlatformSetting
type NotificationTemplates = NotificationTemplate
type UserNotificationPreferences = UserNotificationPreference
type FileUploads = FileUpload
type ServicePricing = ServicePrice
type SpinResults = SpinResult
type WheelPrizes = WheelPrize

// Semantic aliases (different concept name → same underlying type)
type Recharge = Transaction          // recharges are transactions
type WheelSpin = SpinResult          // wheel spins are spin results
type Subscription = DailySubscription
type Withdrawal = WithdrawalRequest

// Repository aliases (kept in domain/repositories/aliases.go; mirrored here for clarity)
// ReferralRepository → UserRepository  (referrals tracked via users.referred_by)
// RechargeRepository → TransactionRepository
// SpinResultRepository → SpinRepository
