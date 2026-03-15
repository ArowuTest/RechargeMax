package services
import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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
		db:                      db,
		backendURL:              backendURL,
		frontendURL:             frontendURL,
	}
}

// CreateRecharge creates a new recharge transaction
func (s *RechargeService) CreateRecharge(ctx context.Context, req CreateRechargeRequest) (*RechargeResponse, error) {
	// Normalize phone number to international format for database storage
	normalizedMSISDN := normalizePhoneToInternational(req.MSISDN)
	
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
		Msisdn:          normalizedMSISDN, // Use normalized format for database
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
		MSISDN:          recharge.Msisdn,
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

	if recharge.Status != "PENDING" {
		return fmt.Errorf("recharge is not in pending status, current status: %s", recharge.Status)
	}

	// ✅ NEW: Attempt VTU with automatic retry (up to 2 immediate retries)
	vtuResponse, vtuErr := s.attemptVTUWithRetry(ctx, recharge, 2)

	if vtuErr != nil {
		// All retries exhausted - initiate refund
		return s.handleFailedRechargeWithRefund(ctx, recharge, vtuErr.Error())
	}

	// Check VTU response status
	if !vtuResponse.Success {
		// Check if it's a PENDING status (needs requery, not refund)
		if vtuResponse.Status == "PROCESSING" || vtuResponse.Status == "PENDING" {
			return s.handlePendingRecharge(ctx, recharge, vtuResponse)
		}
		
		// All retries exhausted - initiate refund
		return s.handleFailedRechargeWithRefund(ctx, recharge, vtuResponse.Message)
	}

	// Calculate points earned (₦200 = 1 point)
	// ₦200 = 20000 kobo, so points = amount / 20000
	pointsEarned := recharge.Amount / 20000
	
	// Calculate draw entries with loyalty tier multiplier
	// Get user's current total points to determine their tier
	var userForTier entities.Users
	_ = s.db.Where("msisdn = ?", recharge.Msisdn).First(&userForTier).Error
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
			return fmt.Errorf("failed to update recharge status: %w", err)
		}

		// Find or create user account (auto-create for guest transactions)
		// Query within transaction to see committed data
		var user entities.Users
		err := tx.Where("msisdn = ?", recharge.Msisdn).First(&user).Error
		if err != nil {
			// User doesn't exist - auto-create for guest transaction
			fmt.Printf("Auto-creating user account for guest transaction: %s\n", recharge.Msisdn)
			
			// Generate unique codes
			userCode := fmt.Sprintf("USR%s", uuid.New().String()[:8])
			referralCode := fmt.Sprintf("RCH%s", uuid.New().String()[:8])
			
			newUser := &entities.Users{
				MSISDN:              recharge.Msisdn,
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
				return fmt.Errorf("failed to create user account: %w", err)
			}
				user = *newUser
			
			// Link transaction to newly created user
			if err := tx.Model(&entities.Transactions{}).Where("id = ?", recharge.ID).Update("user_id", user.ID).Error; err != nil {
				return fmt.Errorf("failed to link transaction to user: %w", err)
			}
		} else {
			// User exists - update points and stats
			user.TotalPoints += int(pointsEarned)
			user.TotalRechargeAmount += recharge.Amount
			
			if err := tx.Save(&user).Error; err != nil {
				return fmt.Errorf("failed to update user points: %w", err)
			}
		}

		// Process affiliate commission if applicable
		if s.affiliateService != nil {
			if err := s.affiliateService.ProcessCommission(ctx, recharge.Msisdn, recharge.Amount, recharge.ID); err != nil {
				// Log error but don't fail the recharge
				fmt.Printf("Failed to process affiliate commission: %v\n", err)
				// Don't return error - commission failure shouldn't rollback recharge
			}
		}

		// Create wheel spin opportunity if eligible
		if isWheelEligible && s.spinService != nil {
			if err := s.spinService.CreateSpinOpportunity(ctx, user.ID, recharge.ID); err != nil {
				// Log error but don't fail the recharge
				fmt.Printf("Failed to create spin opportunity: %v\n", err)
				// Don't return error - spin opportunity failure shouldn't rollback recharge
			}
		}

		// Create draw_entries rows for the active draw
		s.createRechargeDrawEntries(ctx, tx, &user, int(drawEntries), recharge.Msisdn)

		return nil // Commit transaction
	})
	
	if err != nil {
		return fmt.Errorf("payment processing transaction failed: %w", err)
	}

	// Update the user's cached loyalty_tier field based on their new total points.
	// This is done asynchronously outside the transaction so a failure here never
	// rolls back the recharge.
	go func() {
		var updatedUser entities.Users
		if dbErr := s.db.Where("msisdn = ?", recharge.Msisdn).First(&updatedUser).Error; dbErr == nil {
			newTier := computeLoyaltyTier(s.db, context.Background(), int64(updatedUser.TotalPoints))
			if newTier != updatedUser.LoyaltyTier {
				if dbErr2 := s.db.Model(&entities.Users{}).
					Where("id = ?", updatedUser.ID).
					Update("loyalty_tier", newTier).Error; dbErr2 != nil {
					fmt.Printf("[Recharge] Loyalty tier update failed for %s: %v\n", recharge.Msisdn, dbErr2)
				} else {
					fmt.Printf("[Recharge] Loyalty tier updated %s: %s -> %s\n",
						recharge.Msisdn, updatedUser.LoyaltyTier, newTier)
				}
			}
		}
	}()

	return nil
}

// computeLoyaltyTier returns the tier name for the given points balance.
func computeLoyaltyTier(db *gorm.DB, ctx context.Context, points int64) string {
	type setting struct {
		Key   string  `gorm:"column:setting_key"`
		Value float64 `gorm:"column:setting_value"`
	}
	defaults := map[string]float64{
		"loyalty.silver_min_points":   500,
		"loyalty.gold_min_points":     2000,
		"loyalty.platinum_min_points": 5000,
	}
	var rows []struct {
		SettingKey   string `gorm:"column:setting_key"`
		SettingValue string `gorm:"column:setting_value"`
	}
	if err := db.WithContext(ctx).
		Table("platform_settings").
		Where("setting_key LIKE 'loyalty.%_min_points'").
		Find(&rows).Error; err == nil {
		for _, r := range rows {
			var v float64
			if n, _ := fmt.Sscanf(r.SettingValue, "%f", &v); n == 1 {
				defaults[r.SettingKey] = v
			}
		}
	}
	switch {
	case float64(points) >= defaults["loyalty.platinum_min_points"]:
		return "PLATINUM"
	case float64(points) >= defaults["loyalty.gold_min_points"]:
		return "GOLD"
	case float64(points) >= defaults["loyalty.silver_min_points"]:
		return "SILVER"
	default:
		return "BRONZE"
	}
}

// GetRechargeHistory retrieves recharge history for a user
func (s *RechargeService) GetRechargeHistory(ctx context.Context, msisdn string, limit, offset int) ([]*entities.Recharge, error) {
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
		if recharge.Status == "completed" || recharge.Status == "success" {
			totalRevenue += float64(recharge.Amount) / 100.0
			successfulCount++
		} else if recharge.Status == "failed" {
			failedCount++
		} else if recharge.Status == "pending" || recharge.Status == "PENDING" {
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
	if recharge.Status == "completed" || recharge.Status == "success" {
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
		s.hlrService.InvalidateCache(ctx, recharge.Msisdn, "telecom_confirmation_failed")
		
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
	if recharge.Status == "completed" {
		// Get user details for notification
		user, err := s.userRepo.FindByMSISDN(ctx, recharge.Msisdn)
		if err == nil && user != nil {
			// Send SMS notification
			notificationMsg := fmt.Sprintf("Your ₦%.2f %s recharge was successful! You earned %d points. Ref: %s",
				float64(recharge.Amount)/100,
				recharge.RechargeType,
				recharge.Amount/20000, // ₦200 = 1 point
				recharge.PaymentReference,
			)
			// Note: Actual SMS sending would be handled by NotificationService
			// For now, we log it (in production, call notificationService.SendSMS)
			_ = notificationMsg
		}
	}
	
	// Process affiliate commission if applicable
	// Get user to check if they were referred
	user, err := s.userRepo.FindByMSISDN(ctx, recharge.Msisdn)
	if err == nil && user.ReferredBy != nil {
		// Calculate commission (1% default, configurable by admin)
		commissionRate := 0.01 // 1% - should be fetched from config
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
	
	// Log telecom confirmation for audit trail
	// Create audit log entry
	auditLog := map[string]interface{}{
		"event":              "telecom_confirmation",
		"recharge_id":        recharge.ID.String(),
		"msisdn":             recharge.Msisdn,
		"reference":          reference,
		"status":             status,
		"provider":           provider,
		"payload":            payload,
		"timestamp":          time.Now(),
		"previous_status":    recharge.Status,
		"new_status":         status,
	}
	
	// In production, this would be logged to a proper audit system
	// For now, we just acknowledge it
	_ = auditLog
	
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
			Where("status = ? AND type = ?", "COMPLETED", "RECHARGE").
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
			Where("status = ? AND type = ? AND created_at >= ?", "COMPLETED", "RECHARGE", todayStart).
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
			Where("status = ? AND type = ? AND created_at >= ?", "COMPLETED", "RECHARGE", monthStart).
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
			
			fmt.Printf("🔄 Retry attempt %d/%d for transaction %s (waiting %v)\\n", 
				attempt, maxRetries, recharge.ID, waitDuration)
			time.Sleep(waitDuration)
		}
		
		// Attempt VTU based on recharge type
		var err error
		if recharge.RechargeType == "AIRTIME" {
			vtuResponse, err = s.telecomServiceIntegrated.PurchaseAirtime(
				ctx,
				recharge.NetworkProvider,
				recharge.Msisdn,
				int(recharge.Amount),
			)
		} else if recharge.RechargeType == "DATA" {
			// For data, we need the variation code (data bundle ID)
			// TODO: Store variation_code in transactions table
			variationCode := "" // Placeholder
			vtuResponse, err = s.telecomServiceIntegrated.PurchaseData(
				ctx,
				recharge.NetworkProvider,
				recharge.Msisdn,
				variationCode,
				int(recharge.Amount),
			)
		} else {
			return nil, fmt.Errorf("invalid recharge type: %s", recharge.RechargeType)
		}
		
		// Check if successful
		if err == nil && vtuResponse != nil && vtuResponse.Success {
			if attempt > 0 {
				fmt.Printf("✅ Retry successful on attempt %d for transaction %s\\n", attempt, recharge.ID)
			}
			return vtuResponse, nil
		}
		
		// Store last error
		if err != nil {
			lastErr = err
		} else if vtuResponse != nil {
			lastErr = fmt.Errorf("VTU failed: %s", vtuResponse.Message)
		}
		
		fmt.Printf("❌ Attempt %d failed for transaction %s: %v\\n", attempt, recharge.ID, lastErr)
	}
	
	// All retries exhausted
	fmt.Printf("⚠️  All %d retry attempts exhausted for transaction %s\\n", maxRetries+1, recharge.ID)
	return vtuResponse, lastErr
}

// handleFailedRechargeWithRefund handles a failed recharge after all retries exhausted
func (s *RechargeService) handleFailedRechargeWithRefund(ctx context.Context, recharge *entities.Recharge, failureReason string) error {
	// Update recharge status to FAILED
	recharge.Status = "FAILED"
	recharge.FailureReason = failureReason
	s.rechargeRepo.Update(ctx, recharge)
	
	// Invalidate network cache if recharge failed due to wrong network
	s.hlrService.InvalidateCache(ctx, recharge.Msisdn, "recharge_failed")
	
	// ✅ Initiate automatic refund after retries exhausted
	if s.paymentService != nil && recharge.PaymentReference != "" {
		fmt.Printf("💰 Initiating refund for transaction %s after failed retries\\n", recharge.ID)
		
		refundErr := s.paymentService.RefundPayment(
			ctx,
			recharge.PaymentReference,
			recharge.PaymentGateway,
			recharge.Amount,
			fmt.Sprintf("VTU recharge failed after retries: %s", failureReason),
		)
		
		if refundErr != nil {
			// Log refund failure but don't fail the transaction update
			fmt.Printf("❌ CRITICAL: Failed to refund transaction %s (₦%d): %v\\n", 
				recharge.ID, recharge.Amount/100, refundErr)
			fmt.Printf("⚠️  MANUAL ACTION REQUIRED: Admin must manually refund payment reference: %s\\n", 
				recharge.PaymentReference)
			// TODO: Queue for manual review/alert admin
		} else {
			// Refund successful - update transaction
			fmt.Printf("✅ Refund initiated successfully for transaction %s\\n", recharge.ID)
			recharge.Status = "REFUNDED"
			s.rechargeRepo.Update(ctx, recharge)
			
			// TODO: Notify customer via SMS when notification service is available
			// amountNaira := recharge.Amount / 100
			// message := fmt.Sprintf("Your ₦%d recharge could not be completed after multiple attempts. A refund of ₦%d has been initiated and will be processed within 5-7 business days. We apologize for the inconvenience.", amountNaira, amountNaira)
			// s.notificationService.SendSMS(ctx, recharge.Msisdn, message)
		}
	}
	
	return fmt.Errorf("recharge failed after retries: %s", failureReason)
}

// handlePendingRecharge handles a PENDING VTU response (requires requery, not refund)
func (s *RechargeService) handlePendingRecharge(ctx context.Context, recharge *entities.Recharge, vtuResponse *VTUResponse) error {
	// Update status to PROCESSING (will be requeried by background job)
	recharge.Status = "PROCESSING"
	recharge.ProviderReference = vtuResponse.ProviderReference
	s.rechargeRepo.Update(ctx, recharge)
	
	fmt.Printf("⏳ Transaction %s is PENDING with VTPass - will requery for status\\n", recharge.ID)
	
	// TODO: Notify customer that recharge is being processed when notification service is available
	// amountNaira := recharge.Amount / 100
	// message := fmt.Sprintf("Your ₦%d recharge is being processed. You'll be notified once it's complete.", amountNaira)
	// s.notificationService.SendSMS(ctx, recharge.Msisdn, message)
	
	// TODO: Schedule background job to requery VTPass after 2 minutes
	// For now, return success - background job will handle requery
	return nil
}

// getTierMultiplier returns the draw entry multiplier for a user's point balance.
// Multipliers are loaded from platform_settings with fallback to hardcoded defaults.
func getTierMultiplier(db *gorm.DB, ctx context.Context, totalPoints int) float64 {
	settings := map[string]float64{
		"loyalty.silver_min_points":   500,
		"loyalty.gold_min_points":     2000,
		"loyalty.platinum_min_points": 5000,
		"loyalty.silver_multiplier":   1.25,
		"loyalty.gold_multiplier":     1.5,
		"loyalty.platinum_multiplier": 2.0,
	}

	// Try to load overrides from DB
	var rows []struct {
		Key   string `gorm:"column:setting_key"`
		Value string `gorm:"column:setting_value"`
	}
	db.WithContext(ctx).Table("platform_settings").
		Where("setting_key LIKE 'loyalty.%'").
		Select("setting_key, setting_value").
		Find(&rows)

	for _, row := range rows {
		if v, err := strconv.ParseFloat(row.Value, 64); err == nil {
			settings[row.Key] = v
		}
	}

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

	// Find the most recent active draw
	var activeDraw entities.Draw
	if err := tx.WithContext(ctx).Where("status = 'ACTIVE'").Order("start_time DESC").First(&activeDraw).Error; err != nil {
		// No active draw - entries will be created via CSV import at draw time
		return
	}

	now := time.Now()
	count := entryCount
	entry := entities.DrawEntries{
		ID:           uuid.New(),
		DrawID:       activeDraw.ID,
		UserID:       &user.ID,
		Msisdn:       msisdn,
		EntriesCount: &count,
		CreatedAt:    &now,
	}
	if err := tx.WithContext(ctx).Create(&entry).Error; err != nil {
		// Log but don't fail the transaction - draw entries are non-critical
		fmt.Printf("Failed to create draw entries for %s: %v\n", msisdn, err)
		return
	}

	// Increment total_entries counter on the draw
	tx.WithContext(ctx).Model(&activeDraw).
		Update("total_entries", gorm.Expr("total_entries + ?", entryCount))
}
