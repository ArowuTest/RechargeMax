package main

import (
	"context"

	"rechargemax/internal/pkg/safe"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"rechargemax/internal/application/jobs"
	"rechargemax/internal/application/services"
	"rechargemax/internal/domain/repositories"
	"rechargemax/internal/infrastructure/persistence"
	"rechargemax/internal/middleware"
	"rechargemax/internal/presentation/handlers"
)

func main() {
	log.Println("🚀 Starting RechargeMax Rewards Platform...")

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	}

	// Load configuration from environment
	config := loadConfig()
	
	// Initialize database
	db, err := initDatabase(config.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connected successfully")

	// Initialize repositories
	repos := initRepositories(db)
	log.Println("✅ Repositories initialized")

	// Initialize services
	svcs := initServices(repos, config, db)
	log.Println("✅ Services initialized")

	// Initialize handlers
	hdlrs := initHandlers(svcs, repos, config, db)
	log.Println("✅ Handlers initialized")

	// Setup router
	router := setupRouter(hdlrs, svcs, db)
	log.Println("✅ Router configured")

	// Start server
	srv := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── Background Jobs ──────────────────────────────────────────────────────
	// Commission release: auto-approve PENDING commissions past hold period,
	// credit affiliate wallets. Runs every 6 hours.
	serverCtx, serverCancel := context.WithCancel(context.Background())
	_ = serverCancel // cancelled on shutdown below
	commissionJob := jobs.NewCommissionReleaseJob(db)
	commissionJob.StartScheduled(serverCtx, 6*time.Hour)
	log.Println("✅ Commission release job started (interval: 6h)")

	// Start server in goroutine
	safe.Go(func() {
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Println("🎉 RechargeMax Rewards Platform - READY!")
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Printf("🌐 Server: http://localhost:%s", config.Port)
		log.Printf("📊 Health: http://localhost:%s/health", config.Port)
		log.Printf("🔌 API v1: http://localhost:%s/api/v1", config.Port)
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	})

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Shutting down server...")

	// Stop background jobs
	serverCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}

// Config holds application configuration
type Config struct {
	DatabaseURL       string
	Port              string
	PaystackKey       string
	FlutterwaveKey    string
	TermiiKey         string
	SendgridKey       string
	FCMKey            string
	JWTSecret         string
	AdminJWTSecret    string // separate secret for admin tokens
	Environment       string
	FrontendURL       string
	BackendURL        string
}

func loadConfig() *Config {
	config := &Config{
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		Port:           getEnv("PORT", "8080"),
		PaystackKey:    getEnv("PAYSTACK_SECRET_KEY", ""),
		FlutterwaveKey: getEnv("FLUTTERWAVE_SECRET_KEY", ""),
		TermiiKey:      getEnv("TERMII_API_KEY", ""),
		SendgridKey:    getEnv("SENDGRID_API_KEY", ""),
		FCMKey:         getEnv("FCM_SERVER_KEY", ""),
		JWTSecret:         getEnv("JWT_SECRET", ""),      // NO DEFAULT - MUST BE SET!
		AdminJWTSecret:    getEnv("ADMIN_JWT_SECRET", ""),  // Falls back to JWT_SECRET if not set
		Environment:    getEnv("ENVIRONMENT", "development"),
		FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:5173"),
		BackendURL:     getEnv("BACKEND_URL", "http://localhost:8080"),
	}
	
	// Validate required configuration
	if err := validateConfig(config); err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}
	
	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func validateConfig(config *Config) error {
	// Required fields
	if config.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	
	if config.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required - generate one with: openssl rand -hex 32")
	}
	
	// JWT secret strength validation
	if len(config.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long (current: %d chars) - use: openssl rand -hex 32", len(config.JWTSecret))
	}
	// Admin JWT secret — must be explicitly set and distinct from JWT_SECRET.
	// In production a shared secret is a hard failure: a compromised user token
	// would grant admin access. In development we allow a fallback with a warning.
	if config.AdminJWTSecret == "" {
		if config.Environment == "production" {
			return fmt.Errorf(
				"ADMIN_JWT_SECRET is required in production — generate one with: openssl rand -hex 32",
			)
		}
		log.Println("⚠️  ADMIN_JWT_SECRET not set — falling back to JWT_SECRET for development only")
		config.AdminJWTSecret = config.JWTSecret
	} else {
		if len(config.AdminJWTSecret) < 32 {
			return fmt.Errorf("ADMIN_JWT_SECRET must be at least 32 characters — use: openssl rand -hex 32")
		}
		if config.AdminJWTSecret == config.JWTSecret {
			if config.Environment == "production" {
				return fmt.Errorf(
					"ADMIN_JWT_SECRET must differ from JWT_SECRET in production — generate a separate secret",
				)
			}
			log.Println("⚠️  ADMIN_JWT_SECRET equals JWT_SECRET — use separate secrets in production")
		}
	}
	
	// In production, require real API keys
	if config.Environment == "production" {
		if config.PaystackKey == "" {
			return fmt.Errorf("production requires Paystack secret key")
		}
		if strings.HasPrefix(config.PaystackKey, "sk_test") {
			return fmt.Errorf("production requires live Paystack key (sk_live_...), not test key")
		}
		if config.TermiiKey == "" || config.TermiiKey == "test_termii_key" {
			return fmt.Errorf("production requires real Termii API key")
		}
	}
	
	log.Println("✅ Configuration validated successfully")
	return nil
}

func initDatabase(dbURL string) (*gorm.DB, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// STRATEGIC PRODUCTION APPROACH: Manual SQL Migrations
	// GORM AutoMigrate is NOT production-ready (no version control, no rollback, race conditions)
	// All schema changes MUST be done via versioned SQL files in /database/
	// Run migrations: ./scripts/run_migrations.sh  (reads from database/migrations/)
	// For production, use golang-migrate/migrate or pressly/goose for automated migration management
	log.Println("✅ Database connection established (using manual migrations)")

	// Verify database is accessible
	var result int
	if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("database verification failed: %w", err)
	}
	
	return db, nil
}

// Repositories holds all repository instances
type Repositories struct {
	User              repositories.UserRepository
	OTP               repositories.OTPRepository
	Transaction       repositories.TransactionRepository
	Subscription      repositories.SubscriptionRepository
	Spin              repositories.SpinRepository
	Wallet            repositories.WalletRepository
	WalletTransaction repositories.WalletTransactionRepository
	Affiliate         repositories.AffiliateRepository
	BankAccount       repositories.BankAccountRepository
	Withdrawal        repositories.WithdrawalRepository
	Draw              repositories.DrawRepository
	Winner            repositories.WinnerRepository
	NetworkCache      repositories.NetworkCacheRepository
	Network           repositories.NetworkRepository
	DataPlan          repositories.DataPlanRepository
	Device            repositories.DeviceRepository
	Notification      repositories.NotificationRepository
	AuditLog          repositories.AuditLogRepository
	Admin             repositories.AdminRepository
	PaymentLog        repositories.PaymentLogRepository
	SubscriptionTier  repositories.SubscriptionTierRepository
	USSDRecharge      repositories.USSDRechargeRepository
	PointsAdjustment  repositories.PointsAdjustmentRepository
	WheelPrize        repositories.WheelPrizeRepository
	// Prize Tier System Repositories
	DrawType          *persistence.DrawTypeRepositoryGORM
	PrizeTemplate     *persistence.PrizeTemplateRepositoryGORM
	PrizeCategory     *persistence.PrizeCategoryRepositoryGORM
	// Webhook Repository
	Webhook           *persistence.WebhookRepository
}

func initRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:              persistence.NewUserRepository(db),
		OTP:               persistence.NewOTPRepository(db),
		Transaction:       persistence.NewTransactionRepository(db),
		Subscription:      persistence.NewSubscriptionRepository(db),
		Spin:              persistence.NewSpinRepository(db),
		Wallet:            persistence.NewWalletRepository(db),
		WalletTransaction: persistence.NewWalletTransactionRepository(db),
		Affiliate:         persistence.NewAffiliateRepository(db),
		BankAccount:       persistence.NewBankAccountRepository(db),
		Withdrawal:        persistence.NewWithdrawalRepository(db),
		Draw:              persistence.NewDrawRepository(db),
		Winner:            persistence.NewWinnerRepository(db),
		NetworkCache:      persistence.NewNetworkCacheRepository(db),
		Network:           persistence.NewNetworkRepository(db),
		DataPlan:          persistence.NewDataPlanRepository(db),
		Device:            persistence.NewDeviceRepository(db),
		Notification:      persistence.NewNotificationRepository(db),
		AuditLog:          persistence.NewAuditLogRepository(db),
		Admin:             persistence.NewAdminRepository(db),
		PaymentLog:        persistence.NewPaymentLogRepository(db),
		SubscriptionTier:  persistence.NewSubscriptionTierRepository(db),
		USSDRecharge:      persistence.NewUSSDRechargeRepository(db),
		PointsAdjustment:  persistence.NewPointsAdjustmentRepository(db),
		WheelPrize:        persistence.NewWheelPrizeRepository(db),
		// Prize Tier System Repositories
		DrawType:          persistence.NewDrawTypeRepositoryGORM(db),
		PrizeTemplate:     persistence.NewPrizeTemplateRepositoryGORM(db),
		PrizeCategory:     persistence.NewPrizeCategoryRepositoryGORM(db),
		// Webhook Repository
		Webhook:           persistence.NewWebhookRepository(db),
	}
}

// Services holds all service instances
type Services struct {
	Auth         *services.AuthService
	Token        *services.TokenService
	User         *services.UserService
	Recharge     *services.RechargeService
	Payment      *services.PaymentService
	Wallet       *services.WalletService
	Affiliate    *services.AffiliateService
	Subscription *services.SubscriptionService
	Spin         *services.SpinService
	HLR          *services.HLRService
	Telecom      *services.TelecomService
	NetworkConfig *services.NetworkConfigService
	Draw         *services.DrawService
	Winner       *services.WinnerService
	Notification *services.NotificationService
	Device       *services.DeviceService
	PushNotification *services.PushNotificationService
	FraudDetection *services.FraudDetectionService
	SubscriptionTier *services.SubscriptionTierService
	USSDRecharge     *services.USSDRechargeService
	Points           *services.PointsService
	// Prize Tier System Services
	DrawType         *services.DrawTypeService
	PrizeTemplate    *services.PrizeTemplateService
	// Webhook Service
	Webhook          *services.WebhookService
	// Admin Spin Claims Service
	AdminSpinClaims  *services.AdminSpinClaimService
}

func initServices(repos *Repositories, config *Config, db *gorm.DB) *Services {
	// Initialize core services first
	paymentService := services.NewPaymentService(config.PaystackKey, config.FlutterwaveKey, repos.PaymentLog)
	hlrService := services.NewHLRService(repos.NetworkCache, config.TermiiKey)
	deviceService := services.NewDeviceService(repos.Device)
	pushNotificationService := services.NewPushNotificationService(config.FCMKey, deviceService)
	
	notificationService := services.NewNotificationService(
		repos.Notification,
		repos.Device,
		repos.User,
		config.TermiiKey,
		config.SendgridKey,
		config.FCMKey,
		db,
	)

	walletService := services.NewWalletService(
		repos.Wallet,
		repos.WalletTransaction,
		paymentService,
	)

	affiliateService := services.NewAffiliateService(
		db,
		repos.Affiliate,
		repos.User,
		persistence.NewAffiliateCommissionRepository(db), // Add commission repo
		repos.Transaction,
		walletService,
		notificationService,
	)

	telecomService := services.NewTelecomService(config.TermiiKey, "", "") // apiKey, apiSecret, baseURL

	// Initialize integrated telecom service with VTPass support
	sqlDB, _ := db.DB()
	telecomServiceIntegrated := services.NewTelecomServiceIntegrated(sqlDB)

	networkConfigService := services.NewNetworkConfigService(repos.Network, repos.DataPlan, hlrService)

	userService := services.NewUserService(
		repos.User,
		repos.Transaction,
		repos.BankAccount,
		repos.Withdrawal,
		repos.Transaction, // RechargeRepository (alias)
		repos.Spin,
		repos.Subscription,
		db,
	)

	subscriptionService := services.NewSubscriptionService(
		repos.Subscription,
		repos.User,
		paymentService,
		hlrService,
		db,
	)

	// Prize Fulfillment Config Service
	prizeFulfillmentConfigService := services.NewPrizeFulfillmentConfigService(sqlDB)

	spinService := services.NewSpinService(
		repos.Spin,
		repos.WheelPrize, // WheelPrizeRepository
		repos.User,
		repos.Transaction,
		hlrService,
		telecomServiceIntegrated,
		prizeFulfillmentConfigService,
		db, // Database connection for advisory locks
	)

	rechargeService := services.NewRechargeService(
		repos.Transaction, // RechargeRepository (alias)
		repos.User,
		repos.Transaction,
		repos.DataPlan,
		hlrService,
		telecomService,
		telecomServiceIntegrated, // New integrated service with VTPass
		paymentService,
		affiliateService,
		spinService,
			db,
			config.BackendURL,
			config.FrontendURL,
		)

	drawService := services.NewDrawService(
		db,
		repos.Draw,
		repos.Transaction, // RechargeRepository (alias)
		repos.Subscription,
		repos.Spin, // SpinResultRepository (alias)
	)

	winnerService := services.NewWinnerService(
		repos.Winner,
		repos.Draw,
		repos.User,
		repos.Spin,
		hlrService,
		telecomServiceIntegrated,
		notificationService,
		db,
	)

	authService := services.NewAuthService(
		repos.OTP,
		repos.User,
		config.JWTSecret,
		24 * time.Hour,
		config.TermiiKey,
		"production",
	)

	// TokenService requires TokenBlacklistRepository - using nil for now
	var tokenService *services.TokenService = nil

	fraudDetectionService := services.NewFraudDetectionService()

	// Initialize new services
	subscriptionTierService := services.NewSubscriptionTierService(
		repos.SubscriptionTier,
		repos.User,
		paymentService,
		notificationService,
	)

	ussdRechargeService := services.NewUSSDRechargeService(
		repos.USSDRecharge,
		repos.User,
		notificationService,
		repos.Draw,
	)

	pointsService := services.NewPointsService(
		repos.User,
		repos.Transaction, // RechargeRepository (alias)
		repos.USSDRecharge,
		repos.Subscription,
		repos.Spin,
		repos.PointsAdjustment,
		notificationService,
	)

	// Prize Tier System Services
	drawTypeService := services.NewDrawTypeService(repos.DrawType)
	prizeTemplateService := services.NewPrizeTemplateService(repos.PrizeTemplate, repos.PrizeCategory)

	// Webhook Service
	webhookService := services.NewWebhookService(
		repos.Webhook,
		rechargeService,
		subscriptionService,
		paymentService,
		config.PaystackKey,
	)

	// Admin Spin Claims Service
	adminSpinClaimsService := services.NewAdminSpinClaimService(
		repos.Spin,
		repos.User,
		db,
	)

	return &Services{
		Auth:         authService,
		Token:        tokenService,
		User:         userService,
		Recharge:     rechargeService,
		Payment:      paymentService,
		Wallet:       walletService,
		Affiliate:    affiliateService,
		Subscription: subscriptionService,
		Spin:         spinService,
		HLR:          hlrService,
		Telecom:      telecomService,
		NetworkConfig: networkConfigService,
		Draw:         drawService,
		Winner:       winnerService,
		Notification: notificationService,
		Device:       deviceService,
		PushNotification: pushNotificationService,
		FraudDetection: fraudDetectionService,
		SubscriptionTier: subscriptionTierService,
		USSDRecharge:     ussdRechargeService,
		Points:           pointsService,
		// Prize Tier System Services
		DrawType:         drawTypeService,
		PrizeTemplate:    prizeTemplateService,
		// Webhook Service
		Webhook:          webhookService,
		// Admin Spin Claims Service
		AdminSpinClaims:  adminSpinClaimsService,
	}
}

// Handlers holds all handler instances
type Handlers struct {
	Health       *handlers.HealthHandler
	Auth         *handlers.AuthHandler
	User         *handlers.UserHandler
	Recharge     *handlers.RechargeHandler
	Subscription *handlers.SubscriptionHandler
	Spin         *handlers.SpinHandler
	Affiliate    *handlers.AffiliateHandler
	Draw         *handlers.DrawHandler
	Winner       *handlers.WinnerHandler
	Notification *handlers.NotificationHandler
	Admin        *handlers.AdminHandler
	AdminAuth    *handlers.AdminAuthHandler
	AdminComprehensive *handlers.AdminComprehensiveHandler
	AdminSpinTiers *handlers.AdminSpinTiersHandler
	AdminUserManagement *handlers.AdminUserManagementHandler
	PlatformSettings       *handlers.PlatformSettingsHandler
	TransactionLimits      *handlers.TransactionLimitsHandler
	Network      *handlers.NetworkHandler
	Platform     *handlers.PlatformHandler
	Payment      *handlers.PaymentHandler
	Commission   *handlers.CommissionHandler
	ValidationStats *handlers.ValidationStatsHandler
	Webhook      *handlers.WebhookHandler
	AdminSpinClaims *handlers.AdminSpinClaimsHandler
}

func initHandlers(svcs *Services, repos *Repositories, appConfig *Config, db *gorm.DB) *Handlers {
	return &Handlers{
		Health:       handlers.NewHealthHandler(db),
		Auth:        handlers.NewAuthHandler(svcs.Auth, nil), // TokenService not implemented yet
		User:         handlers.NewUserHandler(svcs.User, svcs.Wallet),
		Recharge:     handlers.NewRechargeHandler(svcs.Recharge),
		Subscription: handlers.NewSubscriptionHandler(svcs.Subscription),
		Spin:         handlers.NewSpinHandler(svcs.Spin),
		Affiliate:    handlers.NewAffiliateHandler(svcs.Affiliate),
		Draw:         handlers.NewDrawHandler(svcs.Draw),
		Winner:       handlers.NewWinnerHandler(svcs.Winner),
		Notification: handlers.NewNotificationHandler(svcs.Notification),
		Admin:        handlers.NewAdminHandler(svcs.Draw, svcs.Winner, svcs.User),
		AdminAuth:    handlers.NewAdminAuthHandler(repos.Admin, appConfig.AdminJWTSecret),
			AdminComprehensive: handlers.NewAdminComprehensiveHandler(
				svcs.SubscriptionTier,
				svcs.USSDRecharge,
				svcs.Points,
				svcs.Draw,
				svcs.Winner,
				svcs.Spin,
				svcs.Recharge,
				svcs.User,
				svcs.Affiliate,
				svcs.Telecom,
				svcs.NetworkConfig,
				repos.Network,
				repos.DataPlan,
				svcs.Subscription,
				svcs.DrawType,
				svcs.PrizeTemplate,
				db,
			),
			AdminSpinTiers: handlers.NewAdminSpinTiersHandler(db),
		AdminUserManagement: handlers.NewAdminUserManagementHandler(repos.Admin),
		PlatformSettings:  handlers.NewPlatformSettingsHandler(db),
		TransactionLimits: handlers.NewTransactionLimitsHandler(db),
		Network: handlers.NewNetworkHandler(svcs.NetworkConfig, svcs.HLR),
		Platform: handlers.NewPlatformHandler(db),
		Payment: handlers.NewPaymentHandler(svcs.Payment, svcs.Recharge, svcs.Subscription, appConfig.FrontendURL),
		Commission: handlers.NewCommissionHandler(db),
		ValidationStats: handlers.NewValidationStatsHandler(db),
		Webhook: handlers.NewWebhookHandler(svcs.Webhook),
		AdminSpinClaims: handlers.NewAdminSpinClaimsHandler(svcs.AdminSpinClaims),
	}
}

func setupRouter(hdlrs *Handlers, svcs *Services, db *gorm.DB) *gin.Engine {
	// Set Gin mode based on environment
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	// Security middleware
	router.Use(middleware.RequestSizeLimitMiddleware(10 * 1024 * 1024)) // 10MB default limit
	router.Use(middleware.RateLimitMiddleware(100))      // 100 requests per minute

	// Health check (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "rechargemax-api",
			"version":   "1.0.0",
		})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		otpLimiter := middleware.NewOTPRateLimiter(db, 5, time.Minute)
		// OTP rate limiting is handled within the handler itself
		{
			auth.POST("/send-otp", middleware.OTPRateLimit(otpLimiter), hdlrs.Auth.SendOTP)
			auth.POST("/verify-otp", hdlrs.Auth.VerifyOTP)
		}

			// Network routes (public)
			networks := v1.Group("/networks")
			{
			networks.GET("", hdlrs.Network.GetNetworks)
			networks.GET("/:networkId/bundles", hdlrs.Network.GetDataBundles)
			networks.POST("/validate", hdlrs.Network.ValidatePhoneNetwork)
			networks.POST("/cached", hdlrs.Network.GetCachedNetwork)
			networks.POST("/validate-selection", hdlrs.Network.ValidateNetworkSelection)
			}

				// Public recharge routes (for guest users)
				recharge := v1.Group("/recharge")
				{
					recharge.POST("/airtime", hdlrs.Recharge.InitiateAirtimeRecharge)
					recharge.POST("/data", hdlrs.Recharge.InitiateDataRecharge)
					recharge.GET("/:id", hdlrs.Recharge.GetRecharge)
					recharge.GET("/reference/:reference", hdlrs.Recharge.GetRechargeByReference)
				}

			// Payment routes (public)
			payment := v1.Group("/payment")
			{
				payment.POST("/initialize", hdlrs.Payment.InitializePayment)
				payment.GET("/verify/:reference", hdlrs.Payment.VerifyPayment)
				payment.POST("/webhook", hdlrs.Payment.HandleWebhook)
				payment.GET("/callback", hdlrs.Payment.HandleCallback)
				payment.GET("/callback/success", hdlrs.Payment.HandleSuccess)
				payment.GET("/callback/cancel", hdlrs.Payment.HandleCancel)
			}

			// Webhook routes (public - for payment gateways)
			webhooks := v1.Group("/webhooks")
			{
				webhooks.POST("/paystack", hdlrs.Webhook.HandlePaystackWebhook)
			}

		// Platform routes (public)
		platform := v1.Group("/platform")
		{
			platform.GET("/statistics", hdlrs.Platform.GetStatistics)
		}

			// Winners routes (public)
			winners := v1.Group("/winners")
			{
				winners.GET("/recent", hdlrs.Platform.GetRecentWinners)
			}

			// Subscription config route (public - for displaying pricing)
			v1.GET("/subscription/config", hdlrs.Subscription.GetConfig)

			// Draws routes (public)
			draws := v1.Group("/draws")
			{
				draws.GET("", hdlrs.Draw.GetDraws)
				draws.GET("/active", hdlrs.Draw.GetActiveDraws)
				draws.GET("/:id", hdlrs.Draw.GetDrawByID)
				draws.GET("/:id/winners", hdlrs.Draw.GetDrawWinners)
				draws.GET("/my-entries", hdlrs.Draw.GetMyEntries)
			}

			// Spin routes (public - supports both guest and authenticated users)
			// Uses optional auth middleware to extract JWT if present
			spin := v1.Group("/spin", middleware.OptionalAuthMiddleware())
			{
				spin.POST("/play", hdlrs.Spin.PlaySpin) // Guest & auth spin support
				spin.GET("/eligibility", hdlrs.Spin.CheckEligibility) // Check spin eligibility
				spin.GET("/history", hdlrs.Spin.GetHistory) // Spin history
				spin.GET("/prizes", hdlrs.Spin.GetPrizes) // Public prizes list
			}

			// Spin tiers routes (public - for displaying prizes and progress)
			// Uses optional auth middleware to extract JWT if present
			spins := v1.Group("/spins", middleware.OptionalAuthMiddleware())
			{
				spins.GET("/tiers", hdlrs.Spin.GetTiers) // Get all spin tiers
				spins.GET("/tier-progress", hdlrs.Spin.GetTierProgress) // Get user's tier progress
			}


		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(svcs.Auth))
		{
				// User routes
				user := protected.Group("/user")
				{
					user.GET("/dashboard", hdlrs.User.GetDashboard)
					user.GET("/profile", hdlrs.User.GetProfile)
					// Note: UpdateProfile not implemented - users update via OTP re-verification
					user.GET("/wallet", hdlrs.User.GetWallet)
					user.GET("/transactions", hdlrs.User.GetTransactions)
					user.GET("/prizes", hdlrs.User.GetPrizes)
				}

			// Recharge routes
			recharge := protected.Group("/recharge")
			{
				recharge.POST("/initiate", hdlrs.Recharge.InitiateRecharge)
				recharge.GET("/history", hdlrs.Recharge.GetHistory)
				// Note: GetRecharge by ID and reference available via public routes
			}

			// Subscription routes
			subscription := protected.Group("/subscription")
			{
				subscription.POST("/create", hdlrs.Subscription.CreateSubscription)
				subscription.GET("/status", hdlrs.Subscription.GetSubscription)
				subscription.POST("/cancel", hdlrs.Subscription.CancelSubscription)
				subscription.GET("/history", hdlrs.Subscription.GetHistory)
				// Note: Subscription history available via user transactions endpoint
			}

			// Note: Spin routes moved to public section above to support guest spins
			// /spin/play, /spin/eligibility, /spin/history are all public

				// Affiliate routes
				affiliate := protected.Group("/affiliate")
				{
					affiliate.GET("/code", hdlrs.Affiliate.GetReferralCode)
					affiliate.GET("/stats", hdlrs.Affiliate.GetStats)
					affiliate.GET("/referrals", hdlrs.Affiliate.GetReferrals)
					affiliate.GET("/dashboard", hdlrs.Affiliate.GetDashboard)
					affiliate.POST("/register", hdlrs.Affiliate.Register)
					affiliate.GET("/referral-link", hdlrs.Affiliate.GetReferralCode) // Referral link alias
					affiliate.GET("/link", hdlrs.Affiliate.GetReferralLink)           // Referral link with full URL
					affiliate.GET("/commissions", hdlrs.Affiliate.GetCommissions)     // Commission history
					affiliate.GET("/earnings", hdlrs.Affiliate.GetEarnings)           // Earnings summary
					affiliate.POST("/payout", hdlrs.Affiliate.RequestPayout)          // Request payout
				}

			// Winner routes
			winner := protected.Group("/winner")
			{
				winner.GET("/my-wins", hdlrs.Winner.GetMyWins)
				winner.POST("/:id/claim", hdlrs.Winner.ClaimPrize)
			}

			// Notification routes
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", hdlrs.Notification.GetNotifications)
				notifications.GET("/unread-count", hdlrs.Notification.GetUnreadCount)
				notifications.POST("/:id/read", hdlrs.Notification.MarkAsRead)
			}
		}

			// Admin authentication routes (public)
			adminAuth := v1.Group("/admin/auth")
			{
				adminAuth.POST("/login", hdlrs.AdminAuth.Login)
				adminAuth.POST("/logout", hdlrs.AdminAuth.Logout)
			}
			// Legacy /admin/login alias for backward compat
			v1.POST("/admin/login", hdlrs.AdminAuth.Login)

			// Admin routes (require admin authentication)
			admin := v1.Group("/admin")
			admin.Use(middleware.AdminAuthMiddleware(svcs.Auth))
			admin.Use(middleware.AdminAuditMiddleware(db))
			{
			admin.GET("/dashboard", hdlrs.Admin.GetDashboardStats)
			admin.GET("/users", hdlrs.Admin.GetUsers)
			// Note: Individual user details available via GetUsers with filtering
			
				// Draw management
				admin.GET("/draws", hdlrs.Admin.GetDraws)
				admin.POST("/draws", hdlrs.Draw.CreateDraw)
				admin.PUT("/draws/:id", hdlrs.Draw.UpdateDraw)
				admin.POST("/draws/:id/execute", hdlrs.Draw.ExecuteDraw)
				admin.GET("/draws/:id/export", hdlrs.Draw.ExportEntries)
				admin.POST("/draws/:id/import-winners", hdlrs.Draw.ImportWinners)
			
			// Winner management
			// Note: Winners available via draw details and export endpoints

			// Subscription Tier Management
			admin.GET("/subscription-tiers", hdlrs.AdminComprehensive.GetSubscriptionTiers)
			admin.POST("/subscription-tiers", hdlrs.AdminComprehensive.CreateSubscriptionTier)
			admin.PUT("/subscription-tiers/:id", hdlrs.AdminComprehensive.UpdateSubscriptionTier)
			admin.DELETE("/subscription-tiers/:id", hdlrs.AdminComprehensive.DeleteSubscriptionTier)

			// Subscription Pricing
			admin.GET("/subscription-pricing/current", hdlrs.AdminComprehensive.GetCurrentPricing)
			admin.GET("/subscription-pricing/history", hdlrs.AdminComprehensive.GetPricingHistory)
			admin.POST("/subscription-pricing", hdlrs.AdminComprehensive.UpdatePricing)

			// Daily Subscription Monitoring
			admin.GET("/daily-subscriptions", hdlrs.AdminComprehensive.GetDailySubscriptions)
			admin.GET("/daily-subscriptions/:id", hdlrs.AdminComprehensive.GetDailySubscriptionDetails)
			admin.POST("/daily-subscriptions/:id/cancel", hdlrs.AdminComprehensive.CancelDailySubscription)
			admin.GET("/subscription-billings", hdlrs.AdminComprehensive.GetSubscriptionBillings)
			admin.GET("/daily-subscriptions/analytics", hdlrs.AdminComprehensive.GetSubscriptionAnalytics)
			admin.GET("/daily-subscriptions/config", hdlrs.AdminComprehensive.GetSubscriptionConfig)
			admin.PUT("/daily-subscriptions/config", hdlrs.AdminComprehensive.UpdateSubscriptionConfig)

			// USSD Recharge Monitoring
			admin.GET("/ussd/recharges", hdlrs.AdminComprehensive.GetUSSDRecharges)
			admin.GET("/ussd/statistics", hdlrs.AdminComprehensive.GetUSSDStatistics)
			admin.GET("/ussd/webhook-logs", hdlrs.AdminComprehensive.GetUSSDWebhookLogs)
			admin.POST("/ussd/retry-failed", hdlrs.AdminComprehensive.RetryFailedUSSDWebhooks)

			// User Points Management
			admin.GET("/points/users", hdlrs.AdminComprehensive.GetUsersWithPoints)
			admin.GET("/points/history", hdlrs.AdminComprehensive.GetPointsHistory)
			admin.POST("/points/adjust", hdlrs.AdminComprehensive.AdjustUserPoints)
			admin.GET("/points/statistics", hdlrs.AdminComprehensive.GetPointsStatistics)
			admin.GET("/points/export/users", hdlrs.AdminComprehensive.ExportUsersWithPoints)
			admin.GET("/points/export/history", hdlrs.AdminComprehensive.ExportPointsHistory)

			// Draw CSV Management
			admin.GET("/draws/:id/csv/export", hdlrs.AdminComprehensive.ExportDrawToCSV)
			admin.POST("/draws/:id/csv/import-winners", hdlrs.AdminComprehensive.ImportWinnersFromCSV)

			// Winner Claim Processing
			admin.GET("/winners/pending-claims", hdlrs.AdminComprehensive.GetPendingClaims)
			admin.POST("/winners/:id/approve-claim", hdlrs.AdminComprehensive.ApproveWinnerClaim)
			admin.POST("/winners/:id/reject-claim", hdlrs.AdminComprehensive.RejectWinnerClaim)
			admin.GET("/winners/claim-statistics", hdlrs.AdminComprehensive.GetClaimStatistics)
			// Prize fulfillment routes (alias of spin claims for backward compat)
			admin.GET("/prize-fulfillment/failed-provisions", hdlrs.AdminSpinClaims.GetPendingClaims)
			admin.POST("/prize-fulfillment/retry/:id", hdlrs.AdminSpinClaims.ApproveClaim)
			admin.POST("/prize-fulfillment/retry-all", hdlrs.AdminSpinClaims.GetPendingClaims)

			// Spin Wheel Prize Management
			admin.GET("/spin/config", hdlrs.AdminComprehensive.GetSpinConfig)
			admin.PUT("/spin/config", hdlrs.AdminComprehensive.UpdateSpinConfig)
			admin.GET("/spin/prizes", hdlrs.AdminComprehensive.GetAllPrizes)
			admin.POST("/spin/prizes", hdlrs.AdminComprehensive.CreatePrize)
			admin.PUT("/spin/prizes/:id", hdlrs.AdminComprehensive.UpdatePrize)
			admin.DELETE("/spin/prizes/:id", hdlrs.AdminComprehensive.DeletePrize)
			
			// Spin Tiers Management
			admin.GET("/spin-tiers", hdlrs.AdminSpinTiers.GetAllTiers)
			admin.GET("/spin-tiers/:id", hdlrs.AdminSpinTiers.GetTierByID)
			admin.POST("/spin-tiers", hdlrs.AdminSpinTiers.CreateTier)
			admin.PUT("/spin-tiers/:id", hdlrs.AdminSpinTiers.UpdateTier)
			admin.DELETE("/spin-tiers/:id", hdlrs.AdminSpinTiers.DeleteTier)
			
			// Spin Prize Claims Management
			admin.GET("/spin/claims", hdlrs.AdminSpinClaims.ListClaims)
			admin.GET("/spin/claims/pending", hdlrs.AdminSpinClaims.GetPendingClaims)
			admin.GET("/spin/claims/statistics", hdlrs.AdminSpinClaims.GetStatistics)
			admin.GET("/spin/claims/export", hdlrs.AdminSpinClaims.ExportClaims)
			admin.GET("/spin/claims/:id", hdlrs.AdminSpinClaims.GetClaimDetails)
			admin.POST("/spin/claims/:id/approve", hdlrs.AdminSpinClaims.ApproveClaim)
			admin.POST("/spin/claims/:id/reject", hdlrs.AdminSpinClaims.RejectClaim)
			
			// Commission Reconciliation
			admin.POST("/commissions/reconciliation", hdlrs.Commission.GetCommissionReconciliation)
			admin.POST("/commissions/export", hdlrs.Commission.ExportCommissionReport)
			
			// Validation Statistics
			admin.POST("/validation/stats", hdlrs.ValidationStats.GetValidationStats)
			admin.GET("/validation/stats", hdlrs.ValidationStats.GetValidationStats)
			
			// Recharge Monitoring APIs
			admin.GET("/recharge/transactions", hdlrs.AdminComprehensive.GetRechargeTransactions)
			admin.GET("/recharge/stats", hdlrs.AdminComprehensive.GetRechargeStats)
			admin.POST("/recharge/:id/retry", hdlrs.AdminComprehensive.RetryFailedRecharge)
			admin.GET("/recharge/vtpass/status", hdlrs.AdminComprehensive.GetVTPassStatus)
			admin.PUT("/recharge/provider-config", hdlrs.AdminComprehensive.UpdateProviderConfig)
			admin.GET("/recharge/network-configs", hdlrs.AdminComprehensive.GetNetworkConfigurations)
			admin.GET("/recharge/data-plans", hdlrs.AdminComprehensive.GetDataPlans)
			
			// Network Management (Full CRUD)
			admin.POST("/networks", hdlrs.AdminComprehensive.CreateNetwork)
			admin.PUT("/networks/:id", hdlrs.AdminComprehensive.UpdateNetwork)
			admin.DELETE("/networks/:id", hdlrs.AdminComprehensive.DeleteNetwork)
			
			// Data Plan Management (Full CRUD)
			admin.POST("/data-plans", hdlrs.AdminComprehensive.CreateDataPlan)
			admin.PUT("/data-plans/:id", hdlrs.AdminComprehensive.UpdateDataPlan)
			admin.DELETE("/data-plans/:id", hdlrs.AdminComprehensive.DeleteDataPlan)
			
			// User Management APIs
			admin.GET("/users/all", hdlrs.AdminComprehensive.GetAllUsers)
			admin.GET("/users/:id/details", hdlrs.AdminComprehensive.GetUserDetails)
			admin.PUT("/users/:id/status", hdlrs.AdminComprehensive.UpdateUserStatus)
			
			// Affiliate Management APIs
			admin.GET("/affiliates/all", hdlrs.AdminComprehensive.GetAllAffiliates)
			admin.GET("/affiliates/:id/stats", hdlrs.AdminComprehensive.GetAffiliateStats)
			admin.POST("/affiliates/:id/approve", hdlrs.AdminComprehensive.ApproveAffiliate)
			admin.POST("/affiliates/:id/reject", hdlrs.AdminComprehensive.RejectAffiliate)
			admin.POST("/affiliates/:id/suspend", hdlrs.AdminComprehensive.SuspendAffiliate)
			admin.GET("/affiliates/:id/commissions", hdlrs.AdminComprehensive.GetAffiliateCommissions)
			admin.PUT("/affiliates/:id/commission-rate", hdlrs.AdminComprehensive.UpdateAffiliateCommissionRate)
			admin.GET("/affiliates/:id/payouts", hdlrs.AdminComprehensive.GetAffiliatePayouts)
			admin.POST("/affiliates/:id/payout", hdlrs.AdminComprehensive.ProcessAffiliatePayout)
			admin.GET("/affiliates/analytics", hdlrs.AdminComprehensive.GetAffiliateAnalytics)
			
			// Prize Tier System - Draw Types, Templates & Categories
			admin.GET("/draw-types", hdlrs.AdminComprehensive.GetDrawTypes)
			admin.GET("/prize-templates", hdlrs.AdminComprehensive.GetPrizeTemplates)
			admin.GET("/prize-templates/:id", hdlrs.AdminComprehensive.GetPrizeTemplate)
			admin.POST("/prize-templates", hdlrs.AdminComprehensive.CreatePrizeTemplate)
			admin.PUT("/prize-templates/:id", hdlrs.AdminComprehensive.UpdatePrizeTemplate)
			admin.DELETE("/prize-templates/:id", hdlrs.AdminComprehensive.DeletePrizeTemplate)
			admin.POST("/prize-templates/:id/categories", hdlrs.AdminComprehensive.AddPrizeCategory)
			admin.PUT("/prize-categories/:id", hdlrs.AdminComprehensive.UpdatePrizeCategory)
			admin.DELETE("/prize-categories/:id", hdlrs.AdminComprehensive.DeletePrizeCategory)

			// Admin User Management
			admin.GET("/admins", hdlrs.AdminUserManagement.GetAllAdmins)
			admin.GET("/admins/:id", hdlrs.AdminUserManagement.GetAdminByID)
			admin.POST("/admins", hdlrs.AdminUserManagement.CreateAdmin)
			admin.PUT("/admins/:id", hdlrs.AdminUserManagement.UpdateAdmin)
			admin.DELETE("/admins/:id", hdlrs.AdminUserManagement.DeleteAdmin)
			admin.PUT("/admins/:id/status", hdlrs.AdminUserManagement.UpdateAdminStatus)

			// Platform Settings
			admin.GET("/settings", hdlrs.PlatformSettings.GetAllSettings)
			admin.PUT("/settings", hdlrs.PlatformSettings.UpdateSettings)
			admin.GET("/settings/category/:category", hdlrs.PlatformSettings.GetSettingsByCategory)
			admin.PUT("/settings/category/:category", hdlrs.PlatformSettings.UpdateCategorySettings)
				admin.GET("/settings/:key", hdlrs.PlatformSettings.GetSetting)
				admin.PUT("/settings/:key", hdlrs.PlatformSettings.UpdateSetting)

				// Audit Logs
				admin.GET("/audit-logs", middleware.AdminAuditLogsList(db))

			// Transaction Limits
			admin.GET("/transaction-limits", hdlrs.TransactionLimits.ListTransactionLimits)
			admin.GET("/transaction-limits/:id", hdlrs.TransactionLimits.GetTransactionLimit)
			admin.POST("/transaction-limits", hdlrs.TransactionLimits.CreateTransactionLimit)
			admin.PUT("/transaction-limits/:id", hdlrs.TransactionLimits.UpdateTransactionLimit)
			admin.DELETE("/transaction-limits/:id", hdlrs.TransactionLimits.DeleteTransactionLimit)
			}
	}

	return router
}
