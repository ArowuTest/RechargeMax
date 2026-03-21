package services
import (
	"context"
	"fmt"
	"sync"

	"rechargemax/internal/logger"
	"rechargemax/internal/pkg/safe"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
	"rechargemax/internal/errors"
)

// RechargeService handles all recharge-related operations
type RechargeService struct {
	rechargeRepo            repositories.RechargeRepository
	userRepo                repositories.UserRepository
	transactionRepo         repositories.TransactionRepository
	dataPlanRepo            repositories.DataPlanRepository
	hlrService              *HLRService
	telecomService          *TelecomService // Legacy - kept for backward compatibility
	telecomServiceIntegrated *TelecomServiceIntegrated // New integrated service with VTPass
	paymentService          *PaymentService
	affiliateService        *AffiliateService
	spinService             *SpinService
	fraudService            *FraudDetectionService
	notificationService     *NotificationService
	db                      *gorm.DB
	backendURL              string
	frontendURL             string
}

// CreateRechargeRequest represents a recharge creation request
type CreateRechargeRequest struct {
	MSISDN           string  `json:"msisdn" binding:"required"`
	Amount           int64   `json:"amount" binding:"required,min=10000"` // Minimum ₦100 (10000 kobo)
	Network          *string `json:"network"`                             // Optional - will use HLR if not provided
	RechargeType     string  `json:"recharge_type" binding:"required,oneof=airtime data"`
	DataPackage      string  `json:"data_package"`
	PaymentMethod    string  `json:"payment_method" binding:"required,oneof=paystack flutterwave"`
}

// RechargeResponse represents the response after creating a recharge
type RechargeResponse struct {
	ID              uuid.UUID `json:"id"`
	MSISDN          string    `json:"msisdn"`
	Amount          int64     `json:"amount"`
	Network         string    `json:"network"`
	RechargeType    string    `json:"recharge_type"`
	DataPackage     string    `json:"data_package,omitempty"`
	Status          string    `json:"status"`
	PaymentRef      string    `json:"payment_ref"`
	PaymentURL      string    `json:"payment_url,omitempty"`
	PointsEarned    int64     `json:"points_earned"`
	IsWheelEligible bool      `json:"is_wheel_eligible"`
	CreatedAt       time.Time `json:"created_at"`
}

// NewRechargeService creates a new recharge service
func NewRechargeService(
	rechargeRepo repositories.RechargeRepository,
	userRepo repositories.UserRepository,
	transactionRepo repositories.TransactionRepository,
	dataPlanRepo repositories.DataPlanRepository,
	hlrService *HLRService,
	telecomService *TelecomService,
	telecomServiceIntegrated *TelecomServiceIntegrated,
	paymentService *PaymentService,
	affiliateService *AffiliateService,
	spinService *SpinService,
	fraudService *FraudDetectionService,
	notificationService *NotificationService,
	db *gorm.DB,
	backendURL string,
	frontendURL string,
) *RechargeService {
	return &RechargeService{
		rechargeRepo:            rechargeRepo,
		userRepo:                userRepo,
		transactionRepo:         transactionRepo,
		dataPlanRepo:            dataPlanRepo,
		hlrService:              hlrService,
		telecomService:          telecomService,
		telecomServiceIntegrated: telecomServiceIntegrated,
		paymentService:          paymentService,
		affiliateService:        affiliateService,
		spinService:             spinService,
		fraudService:            fraudService,
		notificationService:     notificationService,
		db:                      db,
		backendURL:              backendURL,
		frontendURL:             frontendURL,
	}
}

// CreateRecharge creates a new recharge transaction
func (s *RechargeService) CreateRecharge(ctx context.Context, req CreateRechargeRequest) (*RechargeResponse, error) {
	// Normalize phone number to international format for database storage
	normalizedMSISDN := normalizePhoneToInternational(req.MSISDN)

	// BUG-003: fraud detection gate (velocity + amount ceiling + blacklist)
	if s.fraudService != nil {
		if isFraud, reason, err := s.fraudService.CheckRecharge(ctx, normalizedMSISDN, req.Amount); err != nil {
			logger.Warn("fraud check error (non-blocking)", zap.Error(err))
		} else if isFraud {
			return nil, fmt.Errorf("transaction declined: %s", reason)
		}
	}
	
	// NEW VALIDATION LOGIC:
	// 1. If user selected network, validate it BEFORE proceeding
	// 2. If no selection, try to get from cache (recent recharge)
	// 3. If validation fails, return error (don't proceed to payment)
	validationResult, err := s.hlrService.ValidateAndDetectNetwork(ctx, req.MSISDN, req.Network)
	if err != nil {
		return nil, errors.BadRequest("Network validation failed: " + err.Error())
	}
	
	// Check if validation passed
	if !validationResult.IsValid {
		return nil, errors.BadRequest("Network mismatch: " + validationResult.Message)
	}
	
	// Use validated network
	network := validationResult.ActualNetwork

	// Validate data package for data recharges
	if req.RechargeType == "data" && req.DataPackage == "" {
		return nil, fmt.Errorf("data package is required for data recharges")
	}

	// Get or create user
	user, err := s.getOrCreateUser(ctx, normalizedMSISDN)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create user: %w", err)
	}

	// Generate payment reference and transaction code
	paymentRef := fmt.Sprintf("RCH_%s_%s", req.MSISDN[len(req.MSISDN)-4:], uuid.New().String()[:8]) // Use UUID for guaranteed uniqueness
	transactionCode := fmt.Sprintf("TXN_%s_%d", req.MSISDN[len(req.MSISDN)-4:], time.Now().UnixNano()/1000000) // Use milliseconds for uniqueness

	// Create recharge record
	recharge := &entities.Recharge{
		ID:              uuid.New(),
		UserID:          &user.ID, // Link transaction to user
		TransactionCode: transactionCode,
		MSISDN:          normalizedMSISDN, // Use normalized format for database
		Amount:          req.Amount,
		NetworkProvider: network,
		RechargeType:    strings.ToUpper(req.RechargeType), // DB constraint requires uppercase (AIRTIME/DATA)
		Status:          "PENDING",
		PaymentMethod:   req.PaymentMethod,
		PaymentReference: paymentRef,
	}

	if err := s.rechargeRepo.Create(ctx, recharge); err != nil {
		return nil, fmt.Errorf("failed to create recharge: %w", err)
	}

	// Initialize payment
	email := "user@rechargemax.com" // Default email
	if user.Email != "" {
		email = user.Email
	}

	// Callback goes to BACKEND endpoint which verifies payment and redirects to frontend
	paymentURL, err := s.paymentService.InitializePayment(ctx, PaymentRequest{
		Amount:      req.Amount,
		Email:       email,
		Reference:   paymentRef,
		CallbackURL: fmt.Sprintf("%s/api/v1/payment/callback?reference=%s&gateway=paystack", s.backendURL, paymentRef),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize payment: %w", err)
	}

	return &RechargeResponse{
		ID:              recharge.ID,
		MSISDN:          recharge.MSISDN,
		Amount:          recharge.Amount,
		Network:         recharge.NetworkProvider,
		RechargeType:    recharge.RechargeType,
		DataPackage:     "", // DataPackage not stored in Transactions
		Status:          recharge.Status,
		PaymentRef:      recharge.PaymentReference,
		PaymentURL:      paymentURL,
		PointsEarned:    0, // Will be calculated after successful payment
		IsWheelEligible: false,
		CreatedAt:       recharge.CreatedAt,
	}, nil
}

// ProcessSuccessfulPayment processes a successful payment and fulfills the recharge
func (s *RechargeService) ProcessSuccessfulPayment(ctx context.Context, paymentRef string) error {
	// Get recharge by payment reference
	recharge, err := s.rechargeRepo.FindByPaymentRef(ctx, paymentRef)
	if err != nil {
		return fmt.Errorf("recharge not found: %w", err)
	}

	// Atomically claim the transaction by setting PENDING → PROCESSING.
	// Uses WHERE status='PENDING' so only ONE concurrent goroutine wins;
	// the others see RowsAffected=0 and return immediately (idempotent).
	claim := s.db.WithContext(ctx).
		Model(&entities.Transactions{}).
		Where("id = ? AND status = 'PENDING'", recharge.ID).
		Update("status", "PROCESSING")
	if claim.Error != nil {
		return fmt.Errorf("failed to claim transaction: %w", claim.Error)
	}
	if claim.RowsAffected == 0 {
		// Another goroutine is already processing (or it's already done).
		return nil
	}
	recharge.Status = "PROCESSING"

	// Attempt VTU — no immediate retries for PENDING responses.
	// VTPass sandbox (and sometimes production) returns code=000 with status=initiated
	// on the first call, meaning the network has accepted the request but not yet confirmed
	// delivery. Retrying immediately would just get the same response.
	// Instead, we detect PENDING immediately and hand off to handlePendingRecharge,
	// which starts a background requery loop (polls every 30s for up to 15 minutes).
	vtuResponse, vtuErr := s.attemptVTUWithRetry(ctx, recharge, 0) // 0 retries — detect PENDING fast

	if vtuErr != nil {
		// Hard error (network failure, invalid credentials, etc.) — initiate refund
		return s.handleFailedRechargeWithRefund(ctx, recharge, vtuErr.Error())
	}

	// Check VTU response status
	if !vtuResponse.Success {
		// Check if it's a PENDING/PROCESSING status (needs requery, not refund).
		// VTPass returns code=000 with status=initiated or code=011 for in-flight transactions.
		if vtuResponse.Status == "PROCESSING" || vtuResponse.Status == "PENDING" {
			return s.handlePendingRecharge(ctx, recharge, vtuResponse)
		}
		
		// Definitive failure — initiate refund
		return s.handleFailedRechargeWithRefund(ctx, recharge, vtuResponse.Message)
	}

	// Calculate points earned (₦200 = 1 point)
	// ₦200 = 20000 kobo, so points = amount / 20000
	pointsEarned := recharge.Amount / 20000
	
	// Calculate draw entries with loyalty tier multiplier
	// Get user's current total points to determine their tier
	var userForTier entities.Users
	_ = s.db.Where("msisdn = ?", recharge.MSISDN).First(&userForTier).Error
	multiplier := getTierMultiplier(s.db, ctx, userForTier.TotalPoints)
	drawEntries := int64(float64(pointsEarned) * multiplier)
	if drawEntries < pointsEarned {
		drawEntries = pointsEarned // Minimum 1:1 ratio
	}

	// Check if eligible for wheel spin (₦1000 minimum)
	// ₦1000 = 100000 kobo
	isWheelEligible := recharge.Amount >= 100000

	// CRITICAL: Wrap in database transaction for atomicity
	// Update recharge + award points + process commission atomically
	logger.Info("ProcessSuccessfulPayment: starting DB transaction",
		zap.String("payment_ref", paymentRef),
		zap.String("msisdn", recharge.MSISDN),
		zap.Int64("amount", recharge.Amount),
		zap.Int64("points_earned", pointsEarned),
		zap.Bool("is_wheel_eligible", isWheelEligible),
	)

	var txUser entities.Users // hoisted so post-commit code can access the user ID
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Update recharge status to SUCCESS with points and spin eligibility
		recharge.Status = "SUCCESS"
		if err := tx.Model(&entities.Transactions{}).Where("id = ?", recharge.ID).Updates(map[string]interface{}{
			"status":        "SUCCESS",
			"points_earned": pointsEarned,
			"draw_entries":  drawEntries,
			"spin_eligible": isWheelEligible,
			"completed_at":  time.Now(),
		}).Error; err != nil {
			logger.Error("ProcessSuccessfulPayment: failed to update recharge status", zap.String("payment_ref", paymentRef), zap.Error(err))
			return fmt.Errorf("failed to update recharge status: %w", err)
		}
		logger.Info("ProcessSuccessfulPayment: recharge status updated to SUCCESS", zap.String("payment_ref", paymentRef))

		// Find or create user account (auto-create for guest transactions)
		// Query within transaction to see committed data
		var user entities.Users
		err := tx.Where("msisdn = ?", recharge.MSISDN).First(&user).Error
		if err != nil {
			// User doesn't exist - auto-create for guest transaction
			logger.Info("auto-creating user account for guest transaction", zap.String("msisdn", recharge.MSISDN))
			
			// Generate unique codes
			userCode := fmt.Sprintf("USR%s", uuid.New().String()[:8])
			referralCode := fmt.Sprintf("RCH%s", uuid.New().String()[:8])
			
			newUser := &entities.Users{
				MSISDN:              recharge.MSISDN,
				Email:               recharge.CustomerEmail,
				FullName:            recharge.CustomerName,
				UserCode:            userCode,
				ReferralCode:        referralCode,
				TotalPoints:         int(pointsEarned),
				TotalRechargeAmount: recharge.Amount,
				IsActive:            true,
				IsVerified:          false, // Not verified until they complete registration
			}
			
			if err := tx.Create(newUser).Error; err != nil {
				logger.Error("ProcessSuccessfulPayment: failed to create user account", zap.String("msisdn", recharge.MSISDN), zap.Error(err))
				return fmt.Errorf("failed to create user account: %w", err)
			}
			user = *newUser
			txUser = user
			logger.Info("ProcessSuccessfulPayment: user auto-created", zap.String("msisdn", recharge.MSISDN), zap.String("user_id", user.ID.String()))
		
			// Link transaction to newly created user
			if err := tx.Model(&entities.Transactions{}).Where("id = ?", recharge.ID).Update("user_id", user.ID).Error; err != nil {
				logger.Error("ProcessSuccessfulPayment: failed to link transaction to user", zap.Error(err))
				return fmt.Errorf("failed to link transaction to user: %w", err)
			}
		} else {
			// User exists - update points and stats
			logger.Info("ProcessSuccessfulPayment: updating existing user", zap.String("msisdn", recharge.MSISDN), zap.String("user_id", user.ID.String()))
			user.TotalPoints += int(pointsEarned)
			user.TotalRechargeAmount += recharge.Amount
			
			if err := tx.Save(&user).Error; err != nil {
				logger.Error("ProcessSuccessfulPayment: failed to update user points", zap.String("msisdn", recharge.MSISDN), zap.Error(err))
				return fmt.Errorf("failed to update user points: %w", err)
			}
			txUser = user
		}

		// Process affiliate commission atomically inside this transaction (BUG-002).
		// Using ProcessCommissionTx ensures the commission record and the recharge
		// update are committed together or rolled back together.
		if s.affiliateService != nil {
			if err := s.affiliateService.ProcessCommissionTx(ctx, tx, recharge.MSISDN, recharge.Amount, recharge.ID); err != nil {
				// Log but don't fail — a commission calculation error should not
				// block the user's recharge from completing.
				logger.Warn("affiliate commission error (non-fatal)", zap.Error(err))
			}
		}

		// NOTE: CreateSpinOpportunity is called AFTER the transaction commits (below).
		// Calling it inside the transaction causes a deadlock: the outer tx holds a
		// row-lock on the transactions row, and CreateSpinOpportunity's s.db UPDATE
		// on the same row blocks, preventing the transaction from ever committing.
		// spin_eligible is already set to true in the Updates() call above, so no
		// data is lost by moving this call post-commit.

		// Create draw_entries rows for the active draw
		s.createRechargeDrawEntries(ctx, tx, &user, int(drawEntries), recharge.MSISDN)

		return nil // Commit transaction
	})
	
	if err != nil {
		logger.Error("ProcessSuccessfulPayment: DB transaction FAILED",
			zap.String("payment_ref", paymentRef),
			zap.String("msisdn", recharge.MSISDN),
			zap.Error(err),
		)
		return fmt.Errorf("payment processing transaction failed: %w", err)
	}
	logger.Info("ProcessSuccessfulPayment: DB transaction committed successfully",
		zap.String("payment_ref", paymentRef),
		zap.String("msisdn", recharge.MSISDN),
	)

	// Create wheel spin opportunity POST-COMMIT to avoid deadlock.
	// spin_eligible is already persisted by the transaction above; this call
	// records the grant in the WheelSpin/spin_results table so CheckEligibility
	// can enforce the one-spin-per-qualifying-recharge rule.
	if isWheelEligible && s.spinService != nil && txUser.ID != uuid.Nil {
		if spinErr := s.spinService.CreateSpinOpportunity(ctx, txUser.ID, recharge.ID); spinErr != nil {
			// Non-fatal: spin_eligible is already true in the DB; the user can still spin.
			logger.Error("post-commit CreateSpinOpportunity failed (non-fatal)", zap.Error(spinErr))
		}
	}

	// Update the user's cached loyalty_tier field based on their new total points.
	// This is done asynchronously outside the transaction so a failure here never
	// rolls back the recharge.
	safe.Go(func() {
		var updatedUser entities.Users
		if dbErr := s.db.Where("msisdn = ?", recharge.MSISDN).First(&updatedUser).Error; dbErr == nil {
			newTier := computeLoyaltyTier(s.db, context.Background(), int64(updatedUser.TotalPoints))
			if newTier != updatedUser.LoyaltyTier {
				if dbErr2 := s.db.Model(&entities.Users{}).
					Where("id = ?", updatedUser.ID).
					Update("loyalty_tier", newTier).Error; dbErr2 != nil {
					logger.Error("loyalty tier update failed", zap.String("msisdn", recharge.MSISDN), zap.Error(dbErr2))
				} else {
					logger.Info("loyalty tier updated", zap.String("msisdn", recharge.MSISDN), zap.String("old_tier", updatedUser.LoyaltyTier), zap.String("new_tier", newTier))
				}
			}
		}
	})

	return nil
}

// ---------------------------------------------------------------------------
// PERF-001: platform_settings in-memory cache (5-minute TTL)
// Avoids a DB round-trip on every recharge for read-heavy loyalty settings.
// ---------------------------------------------------------------------------

var platformSettingsCache struct {
	sync.RWMutex
	data      map[string]float64
	expiresAt time.Time
}

const platformSettingsCacheTTL = 5 * time.Minute

// loadPlatformSettingsLocked refreshes the cache from the DB.
// Must be called with the write-lock held.
func loadPlatformSettingsLocked(db *gorm.DB, ctx context.Context) map[string]float64 {
	defaults := map[string]float64{
		"loyalty.silver_min_points":   500,
		"loyalty.gold_min_points":     2000,
		"loyalty.platinum_min_points": 5000,
		"loyalty.silver_multiplier":   1.25,
		"loyalty.gold_multiplier":     1.5,
		"loyalty.platinum_multiplier": 2.0,
	}
	var rows []struct {
		Key   string  `gorm:"column:setting_key"`
		Value float64 `gorm:"column:setting_value"`
	}
	if err := db.WithContext(ctx).
		Table("platform_settings").
		Where("setting_key LIKE 'loyalty.%'").
		Select("setting_key, setting_value").
		Scan(&rows).Error; err == nil {
		for _, r := range rows {
			defaults[r.Key] = r.Value
		}
	}
	return defaults
}

// getPlatformSettings returns loyalty settings, using a 5-min in-memory cache.
func getPlatformSettings(db *gorm.DB, ctx context.Context) map[string]float64 {
	platformSettingsCache.RLock()
	if time.Now().Before(platformSettingsCache.expiresAt) && platformSettingsCache.data != nil {
		data := platformSettingsCache.data
		platformSettingsCache.RUnlock()
		return data
	}
	platformSettingsCache.RUnlock()

	platformSettingsCache.Lock()
	defer platformSettingsCache.Unlock()
	// Double-checked locking
	if time.Now().Before(platformSettingsCache.expiresAt) && platformSettingsCache.data != nil {
		return platformSettingsCache.data
	}
	platformSettingsCache.data = loadPlatformSettingsLocked(db, ctx)
	platformSettingsCache.expiresAt = time.Now().Add(platformSettingsCacheTTL)
	return platformSettingsCache.data
}

// computeLoyaltyTier returns the tier name for the given points balance.
func computeLoyaltyTier(db *gorm.DB, ctx context.Context, points int64) string {
	settings := getPlatformSettings(db, ctx)
	switch {
	case float64(points) >= settings["loyalty.platinum_min_points"]:
		return "PLATINUM"
	case float64(points) >= settings["loyalty.gold_min_points"]:
		return "GOLD"
	case float64(points) >= settings["loyalty.silver_min_points"]:
		return "SILVER"
	default:
		return "BRONZE"
	}
}

// GetRechargeHistory retrieves recharge history for a user
func (s *RechargeService) GetRechargeHistory(ctx context.Context, msisdn string, limit, offset int) ([]*entities.Recharge, error) {
	// If no MSISDN filter, return all transactions (admin use case)
	if msisdn == "" {
		return s.rechargeRepo.FindAll(ctx, limit, offset)
	}
	return s.rechargeRepo.FindByMSISDN(ctx, msisdn, limit, offset)
}

// GetRechargeByID retrieves a specific recharge by ID
func (s *RechargeService) GetRechargeByID(ctx context.Context, id uuid.UUID) (*entities.Recharge, error) {
	return s.rechargeRepo.FindByID(ctx, id)
}

// GetRechargeByPaymentRef retrieves a recharge by payment reference
func (s *RechargeService) GetRechargeByPaymentRef(ctx context.Context, paymentRef string) (*entities.Recharge, error) {
	return s.rechargeRepo.FindByPaymentRef(ctx, paymentRef)
}

// GetRechargeByReference retrieves a recharge by payment reference (alias for callback)
func (s *RechargeService) GetRechargeByReference(ctx context.Context, reference string) (*entities.Recharge, error) {
	return s.rechargeRepo.FindByPaymentReference(ctx, reference)
}

// GetNetworks returns available networks
func (s *RechargeService) GetNetworks(ctx context.Context) ([]NetworkInfo, error) {
	if s.telecomService == nil {
		return nil, fmt.Errorf("telecom service not configured")
	}

	providers, err := s.telecomService.GetNetworkProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network providers: %w", err)
	}

	networks := make([]NetworkInfo, 0, len(providers))
	for _, provider := range providers {
		network := NetworkInfo{
			Code:     provider.Code,
			Name:     provider.Name,
			Logo:     provider.Logo,
			IsActive: provider.IsActive,
		}
		networks = append(networks, network)
	}

	return networks, nil
}

// GetDataPackages returns available data packages for a network
func (s *RechargeService) GetDataPackages(ctx context.Context, network string) ([]DataPackageInfo, error) {
	if s.telecomService == nil {
		return nil, fmt.Errorf("telecom service not configured")
	}

	packages, err := s.telecomService.GetDataPackages(ctx, network)
	if err != nil {
		return nil, err
	}

	var result []DataPackageInfo
	for _, pkg := range packages {
		result = append(result, DataPackageInfo{
			Code:     pkg.ID,
			Name:     pkg.Name,
			Size:     pkg.DataSize,
			Validity: "30 days", // Default validity
			Price:    pkg.Amount,
			Network:  pkg.Network,
			IsActive: true, // Default active
		})
	}

	return result, nil
}

// NetworkInfo represents network information
type NetworkInfo struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Logo     string `json:"logo"`
	IsActive bool   `json:"is_active"`
}

// DataPackageInfo represents data package information
type DataPackageInfo struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Size     string `json:"size"`
	Validity string `json:"validity"`
	Price    int64  `json:"price"`
	Network  string `json:"network"`
	IsActive bool   `json:"is_active"`
}

// getOrCreateUser gets existing user or creates new one
// Uses centralized CreateUserWithDefaults to ensure consistency across the platform
func (s *RechargeService) getOrCreateUser(ctx context.Context, msisdn string) (*entities.User, error) {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err == nil && user != nil {
		return user, nil
	}

	// User doesn't exist - create using centralized function
	// This ensures all users get proper defaults (referral code, loyalty tier, etc.)
	user, err = s.userRepo.CreateUserWithDefaults(ctx, msisdn, nil)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// TelecomRechargeRequest represents a request to telecom provider
type TelecomRechargeRequest struct {
	MSISDN      string
	Amount      int64
	Network     string
	Type        string
	DataPackage string
	Reference   string
	CallbackURL string
}

// TelecomRechargeResponse represents response from telecom provider
type TelecomRechargeResponse struct {
	Success    bool
	Message    string
	NetworkRef string
}


// GetStats returns recharge statistics for admin dashboard
func (s *RechargeService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// Get total recharge count
	totalCount, err := s.rechargeRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get recharge count: %w", err)
	}
	
	// Get all recharges to calculate revenue
	recharges, err := s.rechargeRepo.FindAll(ctx, 10000, 0) // Get all (up to 10k)
	if err != nil {
		return nil, fmt.Errorf("failed to get recharges: %w", err)
	}
	
	// Calculate total revenue and successful recharges
	var totalRevenue float64
	var successfulCount int64
	var failedCount int64
	var pendingCount int64
	
	for _, recharge := range recharges {
		if recharge.Status == "SUCCESS" {
			totalRevenue += float64(recharge.Amount) / 100.0
			successfulCount++
		} else if recharge.Status == "FAILED" {
			failedCount++
		} else if recharge.Status == "PENDING" {
			pendingCount++
		}
	}
	
	stats := map[string]interface{}{
		"total":            totalCount,
		"successful":       successfulCount,
		"failed":           failedCount,
		"pending":          pendingCount,
		"revenue":          totalRevenue,
		"revenue_kobo":     int64(totalRevenue),
		"revenue_naira":    totalRevenue / 100,
		"success_rate":     float64(successfulCount) / float64(totalCount) * 100,
	}
	
	return stats, nil
}

// GetRechargeByID returns a single recharge by ID

// ProcessTelecomConfirmation processes async confirmation from telecom provider
func (s *RechargeService) ProcessTelecomConfirmation(ctx context.Context, reference string, status string, provider string, payload map[string]interface{}) error {
	// Find recharge by payment reference
	recharge, err := s.rechargeRepo.FindByPaymentRef(ctx, reference)
	if err != nil {
		return fmt.Errorf("recharge not found for reference %s: %w", reference, err)
	}
	
	// Check if already processed
	if recharge.Status == "SUCCESS" {
		return nil // Already processed, idempotent
	}
	
		// Update status based on telecom response
	switch status {
	case "success", "completed", "successful":
		recharge.Status = "SUCCESS"
		// Note: Points are already awarded in ProcessSuccessfulPayment, no need to duplicate here
		
	case "failed", "error":
		recharge.Status = "FAILED"
		
		// Invalidate network cache if recharge failed
		s.hlrService.InvalidateCache(ctx, recharge.MSISDN, "telecom_confirmation_failed")
		
	case "pending", "processing":
		recharge.Status = "PROCESSING"
		
	default:
		return fmt.Errorf("unknown telecom status: %s", status)
	}
	
	// Update recharge in database
	if err := s.rechargeRepo.Update(ctx, recharge); err != nil {
		return fmt.Errorf("failed to update recharge status: %w", err)
	}
	
	// Send notification to user about recharge status
	if recharge.Status == "SUCCESS" {
			// Get user details for notification
		user, err := s.userRepo.FindByMSISDN(ctx, recharge.MSISDN)
		if err == nil && user != nil {
			// Send SMS notification
			notificationMsg := fmt.Sprintf("Your ₦%.2f %s recharge was successful! You earned %d points. Ref: %s",
				float64(recharge.Amount)/100,
				recharge.RechargeType,
				recharge.Amount/20000, // ₦200 = 1 point
				recharge.PaymentReference,
			)
			// Send real SMS via NotificationService
			if s.notificationService != nil {
				go s.notificationService.SendSMS(ctx, recharge.MSISDN, notificationMsg)
			}
		}
	}
	
	// Process affiliate commission if applicable
	// Get user to check if they were referred
	user, err := s.userRepo.FindByMSISDN(ctx, recharge.MSISDN)
	if err == nil && user.ReferredBy != nil {
		// Calculate commission (1% default, configurable by admin)
		commissionRate := 0.01 // default 1%
		var rateSetting struct{ SettingValue string }
		if s.db.WithContext(ctx).
			Table("platform_settings").
			Where("setting_key = ?", "affiliate.commission_rate_percent").
			First(&rateSetting).Error == nil {
			var parsed float64
			if _, scanErr := fmt.Sscanf(rateSetting.SettingValue, "%f", &parsed); scanErr == nil && parsed > 0 {
				commissionRate = parsed / 100.0
			}
		}
		commissionAmount := int64(float64(recharge.Amount) * commissionRate)
		
		// Process commission via AffiliateService
		if s.affiliateService != nil {
			err := s.affiliateService.ProcessCommission(ctx, user.MSISDN, commissionAmount, recharge.ID)
			if err != nil {
				// Log error but don't fail the recharge
				// Commission can be processed manually later
				_ = err
			}
		}
	}
	
	// Write to audit_logs table for compliance trail
	s.db.WithContext(ctx).Exec(`
		INSERT INTO audit_logs (id, entity_type, entity_id, action, description, created_at)
		VALUES (gen_random_uuid(), 'recharge', ?, ?, ?, NOW())`,
		recharge.ID,
		"telecom_confirmation",
		fmt.Sprintf("status=%s provider=%s ref=%s", status, provider, reference),
	)

	return nil
}

	// GetRevenueAnalytics retrieves revenue analytics data using optimized aggregate queries
	func (s *RechargeService) GetRevenueAnalytics(ctx context.Context) (map[string]interface{}, error) {
		// Calculate time boundaries
		now := time.Now()
		todayStart := now.Truncate(24 * time.Hour)
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		
		// Total revenue and count (all time) - using aggregate query
		var totalResult struct {
			TotalRevenue float64
			TotalCount   int64
		}
		err := s.db.WithContext(ctx).
			Table("transactions").
			Select("COALESCE(SUM(amount), 0) as total_revenue, COUNT(*) as total_count").
			Where("status = ? AND type = ?", "SUCCESS", "RECHARGE").
			Scan(&totalResult).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get total revenue: %w", err)
		}
		
		// Today's revenue and count - using aggregate query
		var todayResult struct {
			TotalRevenue float64
			TotalCount   int64
		}
		err = s.db.WithContext(ctx).
			Table("transactions").
			Select("COALESCE(SUM(amount), 0) as total_revenue, COUNT(*) as total_count").
			Where("status = ? AND type = ? AND created_at >= ?", "SUCCESS", "RECHARGE", todayStart).
			Scan(&todayResult).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get today's revenue: %w", err)
		}
		
		// This month's revenue and count - using aggregate query
		var monthResult struct {
			TotalRevenue float64
			TotalCount   int64
		}
		err = s.db.WithContext(ctx).
			Table("transactions").
			Select("COALESCE(SUM(amount), 0) as total_revenue, COUNT(*) as total_count").
			Where("status = ? AND type = ? AND created_at >= ?", "SUCCESS", "RECHARGE", monthStart).
			Scan(&monthResult).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get month's revenue: %w", err)
		}
		
		// Calculate average
		var averageRecharge float64
		if totalResult.TotalCount > 0 {
			averageRecharge = totalResult.TotalRevenue / float64(totalResult.TotalCount)
		}
		
		analytics := map[string]interface{}{
			"total_revenue":    totalResult.TotalRevenue,
			"total_recharges":  totalResult.TotalCount,
			"today_revenue":    todayResult.TotalRevenue,
			"today_recharges":  todayResult.TotalCount,
			"month_revenue":    monthResult.TotalRevenue,
			"month_recharges":  monthResult.TotalCount,
			"average_recharge": averageRecharge,
		}
	
	return analytics, nil
}

// normalizePhoneToInternational converts phone number to international format (234...)
// Accepts: 08031234567 or 2348031234567
// Returns: 2348031234567
func normalizePhoneToInternational(phone string) string {
	// Remove all non-digit characters
	digitsOnly := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			digitsOnly += string(char)
		}
	}
	
	// If starts with 0 (local format), replace with 234
	if len(digitsOnly) == 11 && digitsOnly[0] == '0' {
		return "234" + digitsOnly[1:]
	}
	
	// If already in international format, return as-is
	if len(digitsOnly) == 13 && digitsOnly[:3] == "234" {
		return digitsOnly
	}
	
	// Fallback: return as-is (will fail validation)
	return digitsOnly
}

// GetAllDataPlans returns all active data plans for admin management
func (s *RechargeService) GetAllDataPlans() ([]*entities.DataPlans, error) {
	ctx := context.Background()
	plans, err := s.dataPlanRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data plans: %w", err)
	}
	
	// Convert []entities.DataPlans to []*entities.DataPlans
	result := make([]*entities.DataPlans, len(plans))
	for i := range plans {
		result[i] = &plans[i]
	}
	
	return result, nil
}

// attemptVTUWithRetry attempts VTU recharge with automatic retry
// Retries up to maxRetries times with exponential backoff (30s, 2min)
func (s *RechargeService) attemptVTUWithRetry(ctx context.Context, recharge *entities.Recharge, maxRetries int) (*VTUResponse, error) {
	var lastErr error
	var vtuResponse *VTUResponse
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 30s, 2min
			waitDuration := time.Duration(30 * (1 << (attempt - 1))) * time.Second
			if waitDuration > 2*time.Minute {
				waitDuration = 2 * time.Minute
			}
			
			logger.Info("retry attempt", zap.Int("attempt", attempt), zap.Int("max_retries", maxRetries), zap.String("transaction_id", recharge.ID.String()), zap.Duration("wait", waitDuration))
			time.Sleep(waitDuration)
		}
		
		// Attempt VTU based on recharge type
		var err error
		if recharge.RechargeType == "AIRTIME" {
			vtuResponse, err = s.telecomServiceIntegrated.PurchaseAirtime(
				ctx,
				recharge.NetworkProvider,
				recharge.MSISDN,
				int(recharge.Amount),
			)
		} else if recharge.RechargeType == "DATA" {
			// For data, use the DataPlanID as the variation code.
			// DataPlanID is the UUID of the selected data plan which maps to a VTPass variation code.
			variationCode := ""
			if recharge.DataPlanID != nil {
				variationCode = recharge.DataPlanID.String()
			}
			vtuResponse, err = s.telecomServiceIntegrated.PurchaseData(
				ctx,
				recharge.NetworkProvider,
				recharge.MSISDN,
				variationCode,
				int(recharge.Amount),
			)
		} else {
			return nil, fmt.Errorf("invalid recharge type: %s", recharge.RechargeType)
		}
		
		// Check if successful
		if err == nil && vtuResponse != nil && vtuResponse.Success {
			if attempt > 0 {
				logger.Info("retry successful", zap.Int("attempt", attempt), zap.String("transaction_id", recharge.ID.String()))
			}
			return vtuResponse, nil
		}
		
		// Store last error
		if err != nil {
			lastErr = err
		} else if vtuResponse != nil {
			lastErr = fmt.Errorf("VTU failed: %s", vtuResponse.Message)
		}
		
		logger.Warn("attempt failed", zap.Int("attempt", attempt), zap.String("transaction_id", recharge.ID.String()), zap.Error(lastErr))
	}
	
	// All retries exhausted
	logger.Warn("all retry attempts exhausted", zap.Int("total_attempts", maxRetries+1), zap.String("transaction_id", recharge.ID.String()))
	return vtuResponse, lastErr
}

// handleFailedRechargeWithRefund handles a failed recharge after all retries exhausted
func (s *RechargeService) handleFailedRechargeWithRefund(ctx context.Context, recharge *entities.Recharge, failureReason string) error {
	// Update recharge status to FAILED
	recharge.Status = "FAILED"
	recharge.FailureReason = failureReason
	s.rechargeRepo.Update(ctx, recharge)
	
	// Invalidate network cache if recharge failed due to wrong network
	s.hlrService.InvalidateCache(ctx, recharge.MSISDN, "recharge_failed")
	
	// ✅ Initiate automatic refund after retries exhausted
	if s.paymentService != nil && recharge.PaymentReference != "" {
		logger.Info("initiating refund after failed retries", zap.String("transaction_id", recharge.ID.String()))
		
		refundErr := s.paymentService.RefundPayment(
			ctx,
			recharge.PaymentReference,
			recharge.PaymentGateway,
			recharge.Amount,
			fmt.Sprintf("VTU recharge failed after retries: %s", failureReason),
		)
		
		if refundErr != nil {
			// Log refund failure but don't fail the transaction update
			logger.Error("CRITICAL: failed to refund transaction", zap.String("transaction_id", recharge.ID.String()), zap.Int64("amount_naira", recharge.Amount/100), zap.Error(refundErr))
			logger.Error("MANUAL ACTION REQUIRED: admin must manually refund payment", zap.String("payment_reference", recharge.PaymentReference))
			// Write audit log for admin review queue
			s.db.WithContext(ctx).Exec(`
				INSERT INTO audit_logs (id, entity_type, entity_id, action, description, created_at)
				VALUES (gen_random_uuid(), 'recharge', ?, 'REFUND_FAILED_MANUAL_REQUIRED', ?, NOW())`,
				recharge.ID,
				"CRITICAL: Auto-refund failed. Manual refund required for payment ref: "+recharge.PaymentReference,
			)
		} else {
			// Refund successful - update transaction
			// Use CANCELLED status (matches DB CHECK constraint: PENDING/PROCESSING/SUCCESS/FAILED/CANCELLED)
			logger.Info("refund initiated successfully", zap.String("transaction_id", recharge.ID.String()))
			recharge.Status = "CANCELLED"
			s.rechargeRepo.Update(ctx, recharge)
			
			// Notify customer of refund
			if s.notificationService != nil {
				amountNaira := recharge.Amount / 100
				msg := fmt.Sprintf("Your ₦%d recharge could not be completed. A refund has been initiated and will be processed within 5-7 business days. Sorry for the inconvenience.", amountNaira)
				go s.notificationService.SendSMS(ctx, recharge.MSISDN, msg)
			}
		}
	}
	
	return fmt.Errorf("recharge failed after retries: %s", failureReason)
}

// handlePendingRecharge handles a PENDING VTU response (requires requery, not refund)
func (s *RechargeService) handlePendingRecharge(ctx context.Context, recharge *entities.Recharge, vtuResponse *VTUResponse) error {
	// Use the VTPass request_id as the provider reference for requery.
	// When VTPass returns code=011 (PROCESSING), the TransactionID field is empty,
	// but the RequestID we sent is echoed back and is the correct requery key.
	providerRef := vtuResponse.ProviderReference
	if providerRef == "" {
		providerRef = vtuResponse.VTPassRequestID
	}

	// Update status to PROCESSING and persist the requery reference.
	recharge.Status = "PROCESSING"
	recharge.ProviderReference = providerRef
	s.rechargeRepo.Update(ctx, recharge)

	logger.Info("transaction is PENDING with VTPass, starting requery loop",
		zap.String("transaction_id", recharge.ID.String()),
		zap.String("provider_ref", providerRef),
	)

	// Notify customer that recharge is being processed
	if s.notificationService != nil {
		amountNaira := recharge.Amount / 100
		msg := fmt.Sprintf("Your ₦%d recharge is being processed. You will be notified once it is complete.", amountNaira)
		go s.notificationService.SendSMS(ctx, recharge.MSISDN, msg)
	}

	// Background requery loop: poll VTPass every 30 seconds for up to 15 minutes (30 attempts).
	// This ensures the transaction resolves quickly once VTPass delivers the airtime,
	// rather than waiting for the hourly reconciliation job.
	go func() {
		bgCtx := context.Background()
		const (
			maxAttempts  = 30
			pollInterval = 30 * time.Second
		)
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			time.Sleep(pollInterval)

			// Re-fetch the latest recharge state to avoid acting on stale data.
			latest, fetchErr := s.rechargeRepo.FindByID(bgCtx, recharge.ID)
			if fetchErr != nil {
				logger.Warn("requery: could not fetch latest recharge",
					zap.String("transaction_id", recharge.ID.String()), zap.Error(fetchErr))
				continue
			}
			// Stop if another path already resolved this transaction.
			if latest.Status != "PROCESSING" {
				logger.Info("requery: transaction already resolved, stopping loop",
					zap.String("transaction_id", recharge.ID.String()),
					zap.String("status", latest.Status),
				)
				return
			}

			logger.Info("requery attempt",
				zap.Int("attempt", attempt),
				zap.Int("max", maxAttempts),
				zap.String("transaction_id", recharge.ID.String()),
			)
			if reqErr := s.requeryVTPassTransaction(bgCtx, latest); reqErr != nil {
				logger.Warn("requery attempt failed",
					zap.Int("attempt", attempt),
					zap.String("transaction_id", recharge.ID.String()),
					zap.Error(reqErr),
				)
				continue
			}
			// Re-check status after successful requery call.
			resolved, _ := s.rechargeRepo.FindByID(bgCtx, recharge.ID)
			if resolved != nil && resolved.Status != "PROCESSING" {
				logger.Info("requery: transaction resolved",
					zap.String("transaction_id", recharge.ID.String()),
					zap.String("status", resolved.Status),
				)
				return
			}
		}
		logger.Warn("requery loop exhausted without resolution",
			zap.String("transaction_id", recharge.ID.String()),
		)
	}()
	return nil
}

// getTierMultiplier returns the draw entry multiplier for a user's point balance.
// Multipliers are loaded from platform_settings with fallback to hardcoded defaults.
func getTierMultiplier(db *gorm.DB, ctx context.Context, totalPoints int) float64 {
	settings := getPlatformSettings(db, ctx)
	points := float64(totalPoints)
	switch {
	case points >= settings["loyalty.platinum_min_points"]:
		return settings["loyalty.platinum_multiplier"]
	case points >= settings["loyalty.gold_min_points"]:
		return settings["loyalty.gold_multiplier"]
	case points >= settings["loyalty.silver_min_points"]:
		return settings["loyalty.silver_multiplier"]
	default:
		return 1.0
	}
}

// createRechargeDrawEntries creates individual draw_entries rows for an active draw.
// Called inside the payment-processing DB transaction so entries are created atomically.
func (s *RechargeService) createRechargeDrawEntries(ctx context.Context, tx *gorm.DB, user *entities.Users, entryCount int, msisdn string) {
	if entryCount <= 0 {
		return
	}

	// Find the most recent active draw.
	// Use Find+Limit (not First) to avoid GORM logging "record not found"
	// as an error every time no draw is running — that is expected behaviour.
	var activeDraws []entities.Draw
	tx.WithContext(ctx).Where("status = 'ACTIVE'").Order("start_time DESC").Limit(1).Find(&activeDraws)
	if len(activeDraws) == 0 {
		// No active draw - entries will be created via CSV import at draw time
		return
	}
	activeDraw := activeDraws[0]

	now := time.Now()
	count := entryCount
	entry := entities.DrawEntries{
		ID:           uuid.New(),
		DrawID:       activeDraw.ID,
		UserID:       &user.ID,
		MSISDN:       msisdn,
		EntriesCount: &count,
		CreatedAt:    &now,
	}
	if err := tx.WithContext(ctx).Create(&entry).Error; err != nil {
		// Log but don't fail the transaction - draw entries are non-critical
		logger.Error("failed to create draw entries", zap.String("msisdn", msisdn), zap.Error(err))
		return
	}

	// Increment total_entries counter on the draw
	tx.WithContext(ctx).Model(&activeDraw).
		Update("total_entries", gorm.Expr("total_entries + ?", entryCount))
}

// requeryVTPassTransaction re-checks a PROCESSING transaction with VTPass
// and finalises its status. Called from a background goroutine.
func (s *RechargeService) requeryVTPassTransaction(ctx context.Context, recharge *entities.Recharge) error {
	if recharge.ProviderReference == "" {
		// Defensive guard — callers should handle the no-ref case before reaching here.
		// RecoverProcessingTransactions handles this via handleFailedRechargeWithRefund.
		return fmt.Errorf("requeryVTPass: no provider_reference for transaction %s", recharge.ID)
	}
	status, err := s.telecomServiceIntegrated.QueryTransactionStatus(ctx, recharge.ProviderReference)
	if err != nil {
		return fmt.Errorf("QueryTransactionStatus: %w", err)
	}
	switch status {
	case "SUCCESS", "DELIVERED":
		// Calculate points, draw entries, and spin eligibility — same logic as ProcessSuccessfulPayment.
		// Previously this block only set status="SUCCESS" and missed awarding rewards,
		// which caused the frontend poll to show 0 points / no spin wheel after a delayed VTPass response.
		pointsEarned := recharge.Amount / 20000 // ₦200 (20000 kobo) = 1 point

		var userForTier entities.Users
		_ = s.db.WithContext(ctx).Where("msisdn = ?", recharge.MSISDN).First(&userForTier).Error
		multiplier  := getTierMultiplier(s.db, ctx, userForTier.TotalPoints)
		drawEntries := int64(float64(pointsEarned) * multiplier)
		if drawEntries < pointsEarned {
			drawEntries = pointsEarned
		}
		isWheelEligible := recharge.Amount >= 100000 // ₦1,000 = 100,000 kobo

		now := time.Now()
		if err := s.db.WithContext(ctx).
			Model(&entities.Transactions{}).
			Where("id = ?", recharge.ID).
			Updates(map[string]interface{}{
				"status":        "SUCCESS",
				"points_earned": pointsEarned,
				"draw_entries":  drawEntries,
				"spin_eligible": isWheelEligible,
				"completed_at":  now,
			}).Error; err != nil {
			return fmt.Errorf("requeryVTPass: failed to update transaction: %w", err)
		}

		// Award points to the user account
		if pointsEarned > 0 {
			s.db.WithContext(ctx).
				Model(&entities.Users{}).
				Where("msisdn = ?", recharge.MSISDN).
				Updates(map[string]interface{}{
					"total_points": gorm.Expr("total_points + ?", pointsEarned),
				})
		}

		if s.notificationService != nil {
			msg := fmt.Sprintf("Your ₦%d recharge has been processed successfully! You earned %d points. Thank you!", recharge.Amount/100, pointsEarned)
			go s.notificationService.SendSMS(ctx, recharge.MSISDN, msg)
		}
		logger.Info("requeryVTPass: transaction completed",
			zap.String("ref", recharge.PaymentReference),
			zap.Int64("points", pointsEarned),
			zap.Int64("draw_entries", drawEntries),
			zap.Bool("spin_eligible", isWheelEligible),
		)

	case "FAILED":
		return s.handleFailedRechargeWithRefund(ctx, recharge, "VTPass requery returned FAILED")
	// PENDING / PROCESSING → do nothing; reconciliation job will retry later
	}
	return nil
}

// UpdateRecharge persists changes to a recharge/transaction record.
func (s *RechargeService) UpdateRecharge(ctx context.Context, recharge *entities.Recharge) error {
	return s.rechargeRepo.Update(ctx, recharge)
}

// RecoverProcessingTransactions finds all PROCESSING transactions older than 5 minutes
// and attempts to resolve them.
//
// Concurrency safety:
//   - Uses SELECT … FOR UPDATE SKIP LOCKED so multiple server instances or
//     overlapping job runs each claim a disjoint set of rows — no double-processing.
//   - Work is dispatched through a bounded worker pool (max 5 concurrent VTPass
//     calls) so we never hammer the external API with 100 simultaneous requests.
//
// Two resolution paths:
//   a) provider_reference present  → requery VTPass to get final status.
//   b) provider_reference missing  → transaction was stuck before VTPass was ever
//      called (crash, timeout before request was sent). Mark FAILED and refund;
//      the payment was taken but the VTU call never happened.
func (s *RechargeService) RecoverProcessingTransactions(ctx context.Context) error {
	cutoff := time.Now().Add(-5 * time.Minute)

	// FOR UPDATE SKIP LOCKED: each instance claims its own rows; overlapping runs
	// are safe without any external distributed lock.
	var rows []entities.Recharge
	if err := s.db.WithContext(ctx).
		Raw(`SELECT * FROM transactions
		     WHERE status = 'PROCESSING' AND created_at < ?
		     ORDER BY created_at ASC
		     LIMIT 50
		     FOR UPDATE SKIP LOCKED`, cutoff).
		Scan(&rows).Error; err != nil {
		return fmt.Errorf("RecoverProcessingTransactions: query: %w", err)
	}

	if len(rows) == 0 {
		logger.Info("RecoverProcessingTransactions: no stuck transactions found")
		return nil
	}

	withRef  := 0
	withoutRef := 0
	for _, r := range rows {
		if r.ProviderReference == "" { withoutRef++ } else { withRef++ }
	}
	logger.Info("RecoverProcessingTransactions: found stuck transactions",
		zap.Int("total",       len(rows)),
		zap.Int("with_ref",    withRef),
		zap.Int("without_ref", withoutRef),
	)

	// Bounded worker pool — at most 5 concurrent external API calls.
	const maxWorkers = 5
	sem := make(chan struct{}, maxWorkers)

	for i := range rows {
		recharge := &rows[i]
		sem <- struct{}{} // acquire slot
		go func(r *entities.Recharge) {
			defer func() { <-sem }() // release slot
			bgCtx := context.Background()

			if r.ProviderReference == "" {
				// ── Path B: VTPass was never called ──────────────────────
				// The transaction was claimed (PROCESSING) but the process
				// died before the VTPass request was sent.  The customer's
				// payment was taken but no airtime was delivered.
				// Mark FAILED and trigger a refund.
				logger.Warn("RecoverProcessingTransactions: no provider_reference — marking FAILED and refunding",
					zap.String("transaction_id", r.ID.String()),
					zap.String("payment_ref",    r.PaymentReference),
				)
				if err := s.handleFailedRechargeWithRefund(bgCtx, r,
					"Transaction stuck in PROCESSING with no VTPass reference — auto-recovered"); err != nil {
					logger.Error("RecoverProcessingTransactions: refund failed",
						zap.String("transaction_id", r.ID.String()),
						zap.Error(err),
					)
				}
				return
			}

			// ── Path A: requery VTPass ────────────────────────────────────
			if reqErr := s.requeryVTPassTransaction(bgCtx, r); reqErr != nil {
				logger.Error("RecoverProcessingTransactions: requery failed",
					zap.String("transaction_id", r.ID.String()),
					zap.Error(reqErr),
				)
			}
		}(recharge)
	}

	// Wait for all workers to finish before returning.
	for i := 0; i < maxWorkers; i++ {
		sem <- struct{}{}
	}

	return nil
}

// ResetToPending resets FAILED or stuck PROCESSING transactions back to PENDING so they can be retried.
// Used by the admin retry endpoint.
func (s *RechargeService) ResetToPending(ctx context.Context, id uuid.UUID) error {
	return s.db.Model(&entities.Transactions{}).
		Where("id = ? AND status IN ('FAILED', 'PROCESSING')", id).
		Updates(map[string]interface{}{
			"status":         "PENDING",
			"failure_reason": "",
		}).Error
}
