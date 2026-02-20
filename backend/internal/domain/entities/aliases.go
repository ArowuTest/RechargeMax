package entities

// Type aliases for backward compatibility
// Services written with singular entity names can work with plural entity definitions

// Singular aliases for plural entity types
type Draw = Draws
type User = Users
type Affiliate = Affiliates
type Transaction = Transactions
type Recharge = Transactions // Recharges are transactions
type WheelPrize = WheelPrizes
type WheelSpin = SpinResults // Wheel spins are spin results
type Subscription = DailySubscriptions
type NetworkConfig = NetworkConfigs
type DataPlan = DataPlans
type AdminUser = AdminUsers
type AdminSession = AdminSessions
type AdminActivityLog = AdminActivityLogs
type BankAccount = BankAccounts
type WithdrawalRequest = WithdrawalRequests
type Withdrawal = WithdrawalRequests
type AffiliateCommission = AffiliateCommissions
type AffiliatePayout = AffiliatePayouts
type AffiliateClick = AffiliateClicks
type AffiliateBankAccount = AffiliateBankAccounts
type AffiliateAnalytic = AffiliateAnalytics
type DrawEntry = DrawEntries
type DrawWinner = DrawWinners
type PaymentLog = PaymentLogs
type VtuTransaction = VtuTransactions
type ApiLog = ApiLogs
type WebhookLog = WebhookLogs
type ApplicationLog = ApplicationLogs
type ApplicationMetric = ApplicationMetrics
type PlatformSetting = PlatformSettings
type NotificationTemplate = NotificationTemplates
type NotificationDeliveryLogEntry = NotificationDeliveryLog
type UserNotificationPreference = UserNotificationPreferences
type UserNotification = UserNotifications
type OtpVerification = OtpVerifications
type FileUpload = FileUploads
type ServicePrice = ServicePricing
