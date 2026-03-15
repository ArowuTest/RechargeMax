package services

// Registry holds every application service instance.
// It is constructed in main.go and passed to routes.Register.
// Routes depend on services only through this registry — keeping the
// dependency arrow pointing inward (routes → services, never services → routes).
type Registry struct {
	Auth             *AuthService
	Token            *TokenService
	User             *UserService
	Recharge         *RechargeService
	Payment          *PaymentService
	Wallet           *WalletService
	Affiliate        *AffiliateService
	Subscription     *SubscriptionService
	Spin             *SpinService
	HLR              *HLRService
	Telecom          *TelecomService
	NetworkConfig    *NetworkConfigService
	Draw             *DrawService
	Winner           *WinnerService
	Notification     *NotificationService
	Device           *DeviceService
	PushNotification *PushNotificationService
	FraudDetection   *FraudDetectionService
	SubscriptionTier *SubscriptionTierService
	USSDRecharge     *USSDRechargeService
	Points           *PointsService
	DrawType         *DrawTypeService
	PrizeTemplate    *PrizeTemplateService
	Webhook          *WebhookService
	AdminSpinClaims     *AdminSpinClaimService
	Platform            *PlatformService
	CommissionReport    *CommissionService
	ValidationStats     *ValidationStatsService
	SpinTiers           *SpinTiersService
	TransactionLimits   *TransactionLimitsService
	PlatformSettings    *PlatformSettingsService
}
