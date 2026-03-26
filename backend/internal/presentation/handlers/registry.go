package handlers

// Registry holds every handler instance.
// It is constructed in main.go via NewRegistry and passed to routes.Register.
// Keeping the Registry in the handlers package means routes/ only imports
// handlers/ — the dependency arrow stays clean.
type Registry struct {
	Health              *HealthHandler
	Auth                *AuthHandler
	User                *UserHandler
	Recharge            *RechargeHandler
	Subscription        *SubscriptionHandler
	Spin                *SpinHandler
	Affiliate           *AffiliateHandler
	Draw                *DrawHandler
	Winner              *WinnerHandler
	Notification        *NotificationHandler
	Admin               *AdminHandler
	AdminAuth           *AdminAuthHandler
	AdminComprehensive  *AdminComprehensiveHandler
	AdminSpinTiers      *AdminSpinTiersHandler
	AdminUserManagement *AdminUserManagementHandler
	PlatformSettings    *PlatformSettingsHandler
	TransactionLimits   *TransactionLimitsHandler
	Network             *NetworkHandler
	Platform            *PlatformHandler
	Payment             *PaymentHandler
	Commission          *CommissionHandler
	ValidationStats     *ValidationStatsHandler
	Webhook             *WebhookHandler
	AdminSpinClaims     *AdminSpinClaimsHandler
	Monitoring          *MonitoringHandler
}
