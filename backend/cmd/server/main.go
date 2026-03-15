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

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"rechargemax/internal/application/jobs"
	"rechargemax/internal/application/services"
	"rechargemax/internal/domain/repositories"
	"rechargemax/internal/infrastructure/persistence"
	"rechargemax/internal/presentation/handlers"
	"rechargemax/internal/routes"
	"rechargemax/migrations"
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

	// Run embedded SQL migrations — creates all tables from scratch if they don't exist
	migrations.RunAll(db)

	// Seed essential data
	seedDatabase(db)

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
	router := routes.Register(hdlrs, svcs, db)
	log.Println("✅ Router configured")

	// Start server
	srv := &http.Server{
		Addr:              ":" + config.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,   // Slowloris mitigation
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,   // allow chunked/streamed responses
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,            // 1 MB header cap
	}

	// ── Background Jobs ──────────────────────────────────────────────────────
	// Commission release: auto-approve PENDING commissions past hold period,
	// credit affiliate wallets. Runs every 6 hours.
	serverCtx, serverCancel := context.WithCancel(context.Background())
	_ = serverCancel // cancelled on shutdown below
	commissionJob := jobs.NewCommissionReleaseJob(db)
	commissionJob.StartScheduled(serverCtx, 6*time.Hour)
	log.Println("✅ Commission release job started (interval: 6h)")

	reconciliationJob := jobs.NewReconciliationJob(db, svcs.Payment, svcs.Recharge, svcs.Notification)
	reconciliationJob.StartScheduled(serverCtx, 1*time.Hour)
	log.Println("✅ Reconciliation job started (interval: 1h)")

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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	// Retry loop: Render managed postgres may take a few seconds to accept connections
	// on first deploy. Retry up to 15 times with 2s backoff (30s total).
	var db *gorm.DB
	var err error
	maxRetries := 15
	for i := 1; i <= maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("⏳ DB connection attempt %d/%d failed: %v — retrying in 2s...", i, maxRetries, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Connection pool tuning — enough headroom for concurrent requests while
	// preventing resource exhaustion on a shared DB instance.
	sqlDB.SetMaxOpenConns(25)                 // reduced for basic_256mb plan
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	// Ping with retries
	for i := 1; i <= maxRetries; i++ {
		if err = sqlDB.Ping(); err == nil {
			break
		}
		log.Printf("⏳ DB ping attempt %d/%d failed — retrying in 2s...", i, maxRetries)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("database ping failed after %d attempts: %w", maxRetries, err)
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


func initServices(repos *Repositories, config *Config, db *gorm.DB) *services.Registry {
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
		db,
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

	fraudDetectionService := services.NewFraudDetectionService(db)

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
		fraudDetectionService,
		notificationService,
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
		paymentService,
		notificationService,
		db,
	)

	authService := services.NewAuthService(
		repos.OTP,
		repos.User,
		config.JWTSecret,
		config.AdminJWTSecret,
		24*time.Hour,
		config.TermiiKey,
		config.Environment,
		notificationService, // BUG-004: wire NotificationService for production SMS
	)

	// Wire TokenBlacklistRepository and TokenService (SEC-003)
	tokenBlacklistRepo := persistence.NewTokenBlacklistRepository(db)
	tokenService := services.NewTokenService(tokenBlacklistRepo)


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

	// Analytics & admin-only services
	platformService := services.NewPlatformService(db)
	commissionService := services.NewCommissionService(db)
	validationStatsService := services.NewValidationStatsService(db)
	spinTiersService := services.NewSpinTiersService(db)
	transactionLimitsService := services.NewTransactionLimitsService(db)
	platformSettingsService := services.NewPlatformSettingsService(db)

	return &services.Registry{
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
		Platform:            platformService,
		CommissionReport:    commissionService,
		ValidationStats:     validationStatsService,
		SpinTiers:           spinTiersService,
		TransactionLimits:   transactionLimitsService,
		PlatformSettings:    platformSettingsService,
	}
}


func initHandlers(svcs *services.Registry, repos *Repositories, appConfig *Config, db *gorm.DB) *handlers.Registry {
	return &handlers.Registry{
		Health:       handlers.NewHealthHandler(db),
		Auth:        handlers.NewAuthHandler(svcs.Auth, svcs.Token),
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
				svcs.PlatformSettings,
				db,
			),
			AdminSpinTiers: handlers.NewAdminSpinTiersHandler(svcs.SpinTiers),
		AdminUserManagement: handlers.NewAdminUserManagementHandler(repos.Admin),
		PlatformSettings:  handlers.NewPlatformSettingsHandler(svcs.PlatformSettings),
		TransactionLimits: handlers.NewTransactionLimitsHandler(svcs.TransactionLimits),
		Network: handlers.NewNetworkHandler(svcs.NetworkConfig, svcs.HLR),
		Platform: handlers.NewPlatformHandler(svcs.Platform),
		Payment: handlers.NewPaymentHandler(svcs.Payment, svcs.Recharge, svcs.Subscription, appConfig.FrontendURL),
		Commission: handlers.NewCommissionHandler(svcs.CommissionReport),
		ValidationStats: handlers.NewValidationStatsHandler(svcs.ValidationStats),
		Webhook: handlers.NewWebhookHandler(svcs.Webhook),
		AdminSpinClaims: handlers.NewAdminSpinClaimsHandler(svcs.AdminSpinClaims),
	}
}


// seedDatabase inserts essential seed data if tables are empty.
// This is idempotent - it only inserts when count = 0.
func seedDatabase(db *gorm.DB) {
	log.Println("🌱 Checking seed data...")

	// 1. Seed admin user — always UPSERT so password/role is always correct
	{
		var adminCount int64
		_ = db.Table("admin_users").Count(&adminCount)
		log.Printf("  ℹ️  admin_users count = %d (upserting default admin)", adminCount)
		adminSQL := `INSERT INTO admin_users (id, email, password_hash, full_name, role, permissions, is_active, created_at, updated_at)
VALUES ('950e8400-e29b-41d4-a716-446655440001',
        'admin@rechargemax.ng',
        '$2a$10$GSv3/EaeIzohXsGy6jIMfuoOCMkBLZJF/OiqtG7kVdVoD/dKXypoe',
        'Super Administrator',
        'SUPER_ADMIN',
        '["view_analytics","manage_users","manage_transactions","manage_networks","manage_prizes","manage_affiliates","manage_settings","manage_admins","view_monitoring","manage_draws"]',
        true,
        NOW(), NOW())
ON CONFLICT (email) DO UPDATE SET
        password_hash = EXCLUDED.password_hash,
        is_active = true,
        role = EXCLUDED.role`
		if err := db.Exec(adminSQL).Error; err != nil {
			log.Printf("❌ Admin seed FAILED: %v", err)
		} else {
			log.Println("  ✅ Admin user upserted (admin@rechargemax.ng / Admin@123456)")
		}
		// Verify the insert worked
		var verifyCount int64
		_ = db.Raw("SELECT COUNT(*) FROM admin_users WHERE email = ?", "admin@rechargemax.ng").Scan(&verifyCount)
		log.Printf("  ℹ️  Verify admin exists: count=%d", verifyCount)
	}

	// 2. Seed network configs
	var netCount int64
	if err := db.Table("network_configs").Count(&netCount).Error; err == nil && netCount == 0 {
		sql := `INSERT INTO network_configs (id, network_name, network_code, is_active, airtime_enabled, data_enabled, commission_rate, minimum_amount, maximum_amount, logo_url, brand_color, sort_order, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001','MTN Nigeria','MTN',true,true,true,2.50,5000,5000000,'','#FFCC00',1,NOW(),NOW()),
('550e8400-e29b-41d4-a716-446655440002','Airtel Nigeria','AIRTEL',true,true,true,2.50,5000,5000000,'','#FF0000',2,NOW(),NOW()),
('550e8400-e29b-41d4-a716-446655440003','Glo Mobile','GLO',true,true,true,3.00,5000,5000000,'','#00AA00',3,NOW(),NOW()),
('550e8400-e29b-41d4-a716-446655440004','9mobile','9MOBILE',true,true,true,3.50,5000,5000000,'','#006600',4,NOW(),NOW())
ON CONFLICT (id) DO NOTHING`
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("⚠️  Network configs seed warning: %v", err)
		} else {
			log.Println("  ✅ Network configs seeded (MTN, Airtel, Glo, 9mobile)")
		}
	}

	// 3. Seed subscription tiers
	var tierCount int64
	if err := db.Table("subscription_tiers").Count(&tierCount).Error; err == nil && tierCount == 0 {
		sql := `INSERT INTO subscription_tiers (id, name, description, entries, is_active, sort_order, created_at, updated_at) VALUES
('a50e8400-e29b-41d4-a716-446655440001','Bronze','Basic daily subscription',1,true,1,NOW(),NOW()),
('a50e8400-e29b-41d4-a716-446655440002','Silver','Enhanced daily subscription',2,true,2,NOW(),NOW()),
('a50e8400-e29b-41d4-a716-446655440003','Gold','Premium daily subscription',5,true,3,NOW(),NOW()),
('a50e8400-e29b-41d4-a716-446655440004','Platinum','Elite daily subscription',10,true,4,NOW(),NOW())
ON CONFLICT (id) DO NOTHING`
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("⚠️  Subscription tiers seed warning: %v", err)
		} else {
			log.Println("  ✅ Subscription tiers seeded (Bronze, Silver, Gold, Platinum)")
		}
	}

	// 4. Seed data plans
	var planCount int64
	if err := db.Table("data_plans").Count(&planCount).Error; err == nil && planCount == 0 {
		sql := `INSERT INTO data_plans (id, plan_code, plan_name, data_amount, price, network_provider, validity_days, is_active, sort_order, created_at, updated_at) VALUES
('d50e8400-0001-41d4-a716-446655440001','MTN_500MB','500MB Data','500MB',300,'MTN',30,true,1,NOW(),NOW()),
('d50e8400-0002-41d4-a716-446655440002','MTN_1GB','1GB Data','1GB',500,'MTN',30,true,2,NOW(),NOW()),
('d50e8400-0003-41d4-a716-446655440003','MTN_2GB','2GB Data','2GB',1000,'MTN',30,true,3,NOW(),NOW()),
('d50e8400-0004-41d4-a716-446655440004','MTN_5GB','5GB Data','5GB',2000,'MTN',30,true,4,NOW(),NOW()),
('d50e8400-0005-41d4-a716-446655440005','MTN_10GB','10GB Data','10GB',3500,'MTN',30,true,5,NOW(),NOW()),
('d50e8400-0011-41d4-a716-446655440011','GLO_500MB','500MB Data','500MB',250,'GLO',30,true,1,NOW(),NOW()),
('d50e8400-0012-41d4-a716-446655440012','GLO_1GB','1GB Data','1GB',450,'GLO',30,true,2,NOW(),NOW()),
('d50e8400-0013-41d4-a716-446655440013','GLO_2GB','2GB Data','2GB',900,'GLO',30,true,3,NOW(),NOW()),
('d50e8400-0014-41d4-a716-446655440014','GLO_5GB','5GB Data','5GB',1800,'GLO',30,true,4,NOW(),NOW()),
('d50e8400-0021-41d4-a716-446655440021','AIRTEL_500MB','500MB Data','500MB',300,'AIRTEL',30,true,1,NOW(),NOW()),
('d50e8400-0022-41d4-a716-446655440022','AIRTEL_1GB','1GB Data','1GB',500,'AIRTEL',30,true,2,NOW(),NOW()),
('d50e8400-0023-41d4-a716-446655440023','AIRTEL_2GB','2GB Data','2GB',1000,'AIRTEL',30,true,3,NOW(),NOW()),
('d50e8400-0024-41d4-a716-446655440024','AIRTEL_5GB','5GB Data','5GB',2000,'AIRTEL',30,true,4,NOW(),NOW()),
('d50e8400-0031-41d4-a716-446655440031','9MOBILE_500MB','500MB Data','500MB',250,'9MOBILE',30,true,1,NOW(),NOW()),
('d50e8400-0032-41d4-a716-446655440032','9MOBILE_1GB','1GB Data','1GB',400,'9MOBILE',30,true,2,NOW(),NOW()),
('d50e8400-0033-41d4-a716-446655440033','9MOBILE_2GB','2GB Data','2GB',800,'9MOBILE',30,true,3,NOW(),NOW()),
('d50e8400-0034-41d4-a716-446655440034','9MOBILE_5GB','5GB Data','5GB',1600,'9MOBILE',30,true,4,NOW(),NOW())
ON CONFLICT (id) DO NOTHING`
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("⚠️  Data plans seed warning: %v", err)
		} else {
			log.Println("  ✅ Data plans seeded (4 networks × 4 plans)")
		}
	}

	log.Println("🌱 Seed check complete")
}
