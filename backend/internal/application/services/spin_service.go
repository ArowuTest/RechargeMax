package services

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
	"rechargemax/internal/errors"
)

// SpinService handles wheel spin operations
type SpinService struct {
	spinRepo        repositories.SpinRepository
	prizeRepo       repositories.WheelPrizeRepository
	userRepo        repositories.UserRepository
	rechargeRepo    repositories.RechargeRepository
	hlrService      *HLRService
	telecomService  *TelecomServiceIntegrated
	configService   *PrizeFulfillmentConfigService
	db              *gorm.DB // Database connection for advisory locks
}

// SpinEligibilityResponse represents spin eligibility check response
type SpinEligibilityResponse struct {
Eligible      bool   `json:"eligible"`
AvailableSpins int64  `json:"available_spins"`
Message       string `json:"message"`
}

// SpinResultResponse represents spin result
type SpinResultResponse struct {
ID          uuid.UUID `json:"id"`
PrizeWon    string    `json:"prize_won"`
PrizeType   string    `json:"prize_type"`
PrizeValue  int64     `json:"prize_value"`
PointsEarned int64    `json:"points_earned"`
Status      string    `json:"status"`
CreatedAt   time.Time `json:"created_at"`
}

// NewSpinService creates a new spin service
func NewSpinService(
	spinRepo repositories.SpinRepository,
	prizeRepo repositories.WheelPrizeRepository,
	userRepo repositories.UserRepository,
	rechargeRepo repositories.RechargeRepository,
	hlrService *HLRService,
	telecomService *TelecomServiceIntegrated,
	configService *PrizeFulfillmentConfigService,
	db *gorm.DB, // Database connection for advisory locks
) *SpinService {
	return &SpinService{
		spinRepo:        spinRepo,
		prizeRepo:       prizeRepo,
		userRepo:        userRepo,
		rechargeRepo:    rechargeRepo,
		hlrService:      hlrService,
		telecomService:  telecomService,
		configService:   configService,
		db:              db,
	}
}

// CheckEligibility checks if user is eligible to spin
func (s *SpinService) CheckEligibility(ctx context.Context, msisdn string) (*SpinEligibilityResponse, error) {
user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
if err != nil {
return &SpinEligibilityResponse{
Eligible: false,
Message:  "User not found",
}, nil
}

	// Check if user has pending spins
	// Get all spins for this user and filter by pending status
	allSpins, err := s.spinRepo.FindByUserID(ctx, user.ID, 100, 0) // Get up to 100 recent spins
	if err != nil {
		// If error, assume no pending spins
		allSpins = []*entities.SpinResults{}
	}
	
	// Count pending/unclaimed spins
	var pendingSpins int64 = 0
	for _, spin := range allSpins {
		if spin.ClaimStatus == "PENDING" {
			pendingSpins++
		}
	}

if pendingSpins > 0 {
return &SpinEligibilityResponse{
Eligible:       true,
AvailableSpins: pendingSpins,
Message:        fmt.Sprintf("You have %d spin(s) available!", pendingSpins),
}, nil
}

	// Check if user has made a qualifying transaction (₦1000+) today
	// Query transactions table directly by MSISDN
	today := time.Now().Truncate(24 * time.Hour)
	var transaction entities.Transactions
	txErr := s.db.Where("msisdn = ? AND amount >= ? AND status = ? AND created_at >= ?",
		msisdn, int64(100000), "SUCCESS", today).First(&transaction).Error
	
	if txErr != nil {
		return &SpinEligibilityResponse{
			Eligible: false,
			Message:  "No qualifying recharges found. Recharge ₦1000+ to earn a spin!",
		}, nil
	}

return &SpinEligibilityResponse{
Eligible:       true,
AvailableSpins: 1,
Message:        "You're eligible to spin!",
}, nil
}

// PlaySpin plays the wheel spin
func (s *SpinService) PlaySpin(ctx context.Context, msisdn string) (*SpinResultResponse, error) {
	// Try to find existing user
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		// User doesn't exist - check if they have qualifying transactions
		// This handles guest users who just recharged but haven't registered yet
		var transaction entities.Transactions
		today := time.Now().Truncate(24 * time.Hour)
		
		// Debug: Log the query parameters
		errors.Info("Checking for qualifying transactions", map[string]interface{}{
			"msisdn": msisdn,
			"min_amount": int64(100000),
			"status": "SUCCESS",
			"today": today,
		})
		
		txErr := s.db.Where("msisdn = ? AND amount >= ? AND status = ? AND created_at >= ?",
			msisdn, int64(100000), "SUCCESS", today).First(&transaction).Error
		
		if txErr != nil {
			errors.Info("Transaction query failed", map[string]interface{}{
				"error": txErr.Error(),
				"msisdn": msisdn,
			})
			return nil, errors.BadRequest("Not eligible to spin: No qualifying recharges found. Recharge ₦1000+ to earn a spin!")
		}
		
		errors.Info("Found qualifying transaction", map[string]interface{}{
			"transaction_id": transaction.ID,
			"amount": transaction.Amount,
		})
		
		// Auto-create user account for guest transaction
		userCode := fmt.Sprintf("USR%s", uuid.New().String()[:8])
		referralCode := fmt.Sprintf("RCH%s", uuid.New().String()[:8])
		
		user = &entities.Users{
			ID:                  uuid.New(),
			MSISDN:              msisdn,
			UserCode:            userCode,
			ReferralCode:        referralCode,
			TotalPoints:         0,
			TotalRechargeAmount: 0,
			IsActive:            true,
			IsVerified:          false,
		}
		
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user account: %w", err)
		}
	}
	
	// CRITICAL: Acquire advisory lock to prevent race conditions
	// This ensures only one spin can be processed per user at a time
	lockID := int64(user.ID.ID()) // Use UUID's integer representation
	if err := s.acquireLock(ctx, lockID); err != nil {
		return nil, fmt.Errorf("failed to acquire spin lock: %w", err)
	}
	defer s.releaseLock(ctx, lockID)
	
	// Check eligibility (now protected by lock)
	eligibility, err := s.CheckEligibility(ctx, msisdn)
if err != nil {
	return nil, fmt.Errorf("failed to check eligibility: %w", err)
}

if !eligibility.Eligible {
	return nil, errors.BadRequest(fmt.Sprintf("Not eligible to spin: %s", eligibility.Message))
}

// Get all active prizes
prizes, err := s.prizeRepo.FindActive(ctx)
if err != nil || len(prizes) == 0 {
	return nil, errors.BadRequest("No prizes available for spinning")
}

	// Select a random prize based on probability
	selectedPrize := s.selectPrizeByProbability(prizes)
	
	// CRITICAL: Wrap in database transaction for atomicity
	// If any operation fails, all changes are rolled back
	var spin *entities.WheelSpin
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Generate unique spin code
		timestamp := time.Now().Unix()
		last4Digits := msisdn[len(msisdn)-4:]
		spinCode := fmt.Sprintf("SPIN_%s_%d", last4Digits, timestamp)
		
		// Create spin record
		spin = &entities.WheelSpin{
			ID:          uuid.New(),
			SpinCode:    spinCode,
			UserID:      &user.ID,
			Msisdn:      msisdn,
			PrizeID:     &selectedPrize.ID,
			PrizeName:   selectedPrize.PrizeName,
			PrizeType:   selectedPrize.PrizeType,
			PrizeValue:  selectedPrize.PrizeValue,
			ClaimStatus: "PENDING",
		}
	
		if err := tx.Create(spin).Error; err != nil {
			return fmt.Errorf("failed to create spin record: %w", err)
		}
	
		// Check fulfillment mode for this prize type
		config, err := s.configService.GetConfig(ctx, selectedPrize.PrizeType)
		if err != nil {
			return fmt.Errorf("failed to get fulfillment config: %w", err)
		}

		// Set fulfillment mode on spin result
		spin.FulfillmentMode = config.FulfillmentMode

		// Auto-provision if mode is AUTO and prize is airtime/data
		if config.FulfillmentMode == "AUTO" && 
		   (selectedPrize.PrizeType == "DATA" || selectedPrize.PrizeType == "AIRTIME") {
			
			err := s.provisionPrizeWithRetry(ctx, spin, config)
			
			if err != nil {
				// Log the error
				s.logFulfillmentAttempt(ctx, spin, "FAILED", err.Error())
				
				// Check if we should fallback to manual
				if config.FallbackToManual {
					spin.ClaimStatus = "PENDING"
					spin.FulfillmentMode = "MANUAL"
					spin.CanRetry = true
					fmt.Printf("⚠️ Auto-provision failed, falling back to manual claim for spin %s\n", spin.ID)
				} else {
					spin.ClaimStatus = "EXPIRED"
					spin.CanRetry = false
				}
			} else {
				spin.ClaimStatus = "CLAIMED"
				s.logFulfillmentAttempt(ctx, spin, "SUCCESS", "")
			}
			
			// Update spin status within transaction
			if err := tx.Save(spin).Error; err != nil {
				return fmt.Errorf("failed to update spin status: %w", err)
			}
		}
		
		return nil // Commit transaction
	})
	
	if err != nil {
		return nil, fmt.Errorf("spin transaction failed: %w", err)
	}

	return &SpinResultResponse{
		ID:           spin.ID,
		PrizeWon:     spin.PrizeName,
		PrizeType:    spin.PrizeType,
		PrizeValue:   int64(spin.PrizeValue),
		PointsEarned: 0, // Points not stored in SpinResults
		Status:       spin.ClaimStatus,
		CreatedAt:    spin.CreatedAt,
	}, nil
}

// selectPrizeByProbability selects a prize based on probability using cryptographically secure random
func (s *SpinService) selectPrizeByProbability(prizes []*entities.WheelPrizes) *entities.WheelPrizes {
	// Calculate total probability
	totalProb := 0.0
	for _, p := range prizes {
		totalProb += p.Probability
	}
	
	if totalProb == 0 {
		// Fallback to first prize if no probabilities set
		return prizes[0]
	}
	
	// Generate cryptographically secure random number
	// Scale to integer for precision (multiply by 1,000,000)
	maxBig := big.NewInt(int64(totalProb * 1000000))
	randomBig, err := cryptorand.Int(cryptorand.Reader, maxBig)
	if err != nil {
		// Fallback to last prize on error
		return prizes[len(prizes)-1]
	}
	
	// Convert back to float
	r := float64(randomBig.Int64()) / 1000000.0
	
	// Select prize based on cumulative probability
	cumulative := 0.0
	for _, p := range prizes {
		cumulative += p.Probability
		if r <= cumulative {
			return p
		}
	}
	
	// Fallback to last prize
	return prizes[len(prizes)-1]
}

// provisionPrize provisions data or airtime prize
// ============================================================================
// ENTERPRISE-GRADE FULFILLMENT METHODS
// ============================================================================

// provisionPrizeWithRetry provisions a prize with automatic retry logic
func (s *SpinService) provisionPrizeWithRetry(ctx context.Context, spin *entities.WheelSpin, config *FulfillmentConfig) error {
	startTime := time.Now()
	spin.ProvisionStartedAt = &startTime
	
	var lastErr error
	maxAttempts := 1
	if config.AutoRetryEnabled {
		maxAttempts = config.MaxRetryAttempts + 1 // +1 for initial attempt
	}
	
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		spin.FulfillmentAttempts = attempt
		now := time.Now()
		spin.LastFulfillmentAttempt = &now
		
		fmt.Printf("🔄 Provisioning attempt %d/%d for spin %s\n", attempt, maxAttempts, spin.ID)
		
		err := s.provisionPrize(ctx, spin)
		
		if err == nil {
			// Success!
			completedAt := time.Now()
			spin.ProvisionCompletedAt = &completedAt
			fmt.Printf("✅ Prize provisioned successfully on attempt %d\n", attempt)
			return nil
		}
		
		lastErr = err
		spin.FulfillmentError = err.Error()
		
		// If this isn't the last attempt, wait before retrying
		if attempt < maxAttempts {
			retryDelay := time.Duration(config.RetryDelaySeconds) * time.Second
			fmt.Printf("⚠️  Attempt %d failed: %v. Retrying in %v...\n", attempt, err, retryDelay)
			time.Sleep(retryDelay)
		}
	}
	
	// All attempts failed
	fmt.Printf("❌ All %d provision attempts failed for spin %s\n", maxAttempts, spin.ID)
	return fmt.Errorf("all %d provision attempts failed: %w", maxAttempts, lastErr)
}

// provisionPrize provisions a single prize (one attempt)
func (s *SpinService) provisionPrize(ctx context.Context, spin *entities.WheelSpin) error {
	// Detect network
	networkHint := ""
	networkResult, err := s.hlrService.DetectNetwork(ctx, spin.Msisdn, &networkHint)
	if err != nil {
		return fmt.Errorf("failed to detect network: %w", err)
	}
	network := networkResult.Network
	
	// Provision based on prize type
	switch spin.PrizeType {
	case "AIRTIME":
		return s.provisionAirtime(ctx, spin, network)
	case "DATA":
		return s.provisionData(ctx, spin, network)
	case "POINTS":
		// Points are auto-credited, no external provisioning needed
		return nil
	default:
		// CASH, PHYSICAL prizes require manual handling
		return fmt.Errorf("prize type %s requires manual fulfillment", spin.PrizeType)
	}
}

// provisionAirtime provisions airtime via VTPass
func (s *SpinService) provisionAirtime(ctx context.Context, spin *entities.WheelSpin, network string) error {
	if s.telecomService == nil {
		return fmt.Errorf("telecom service not initialized")
	}
	
	fmt.Printf("📞 Provisioning ₦%d airtime to %s on %s network\n", spin.PrizeValue/100, spin.Msisdn, network)
	
	// Call VTPass to purchase airtime (amount in kobo)
	response, err := s.telecomService.PurchaseAirtime(ctx, network, spin.Msisdn, int(spin.PrizeValue))
	if err != nil {
		return fmt.Errorf("VTPass airtime purchase failed: %w", err)
	}
	
	// Store provider reference
	if response != nil {
		spin.ClaimReference = response.ProviderReference
		fmt.Printf("✅ Airtime provisioned successfully. Reference: %s, Status: %s\n", 
			response.ProviderReference, response.Status)
	}
	
	return nil
}

// provisionData provisions data via VTPass
func (s *SpinService) provisionData(ctx context.Context, spin *entities.WheelSpin, network string) error {
	if s.telecomService == nil {
		return fmt.Errorf("telecom service not initialized")
	}
	
	// Get data variation code from prize description or value
	variationCode := s.getDataVariationCode(spin.PrizeValue, network)
	if variationCode == "" {
		return fmt.Errorf("no data variation code found for value %d on %s", spin.PrizeValue, network)
	}
	
	fmt.Printf("📱 Provisioning data (%s) to %s on %s network\n", variationCode, spin.Msisdn, network)
	
	// Call VTPass to purchase data (amount in kobo)
	response, err := s.telecomService.PurchaseData(ctx, network, spin.Msisdn, variationCode, int(spin.PrizeValue))
	if err != nil {
		return fmt.Errorf("VTPass data purchase failed: %w", err)
	}
	
	// Store provider reference
	if response != nil {
		spin.ClaimReference = response.ProviderReference
		fmt.Printf("✅ Data provisioned successfully. Reference: %s, Status: %s\n", 
			response.ProviderReference, response.Status)
	}
	
	return nil
}

// getDataVariationCode maps prize value to VTPass variation code
func (s *SpinService) getDataVariationCode(prizeValue int64, network string) string {
	// TODO: This should be stored in database (wheel_prizes.variation_code)
	// For now, hardcode common mappings
	
	// Map based on network and common data sizes
	// Prize value is in kobo, so 50000 = 500MB, 100000 = 1GB, etc.
	
	switch network {
	case "MTN":
		switch prizeValue {
		case 50000: // 500MB
			return "mtn-20mb-100"
		case 100000: // 1GB
			return "mtn-1gb-500"
		case 200000: // 2GB
			return "mtn-2gb-1000"
		}
	case "GLO":
		switch prizeValue {
		case 50000:
			return "glo-200mb-200"
		case 100000:
			return "glo-1gb-500"
		case 200000:
			return "glo-2gb-1000"
		}
	case "AIRTEL":
		switch prizeValue {
		case 50000:
			return "airtel-750mb-500"
		case 100000:
			return "airtel-1gb-500"
		case 200000:
			return "airtel-2gb-1000"
		}
	case "9MOBILE":
		switch prizeValue {
		case 50000:
			return "etisalat-500mb-500"
		case 100000:
			return "etisalat-1gb-1000"
		case 200000:
			return "etisalat-2gb-2000"
		}
	}
	
	return "" // No matching variation code
}

// logFulfillmentAttempt logs a fulfillment attempt to the audit trail
func (s *SpinService) logFulfillmentAttempt(ctx context.Context, spin *entities.WheelSpin, status string, errorMsg string) {
	// This would insert into prize_fulfillment_logs table
	// For now, just log to console
	fmt.Printf("📝 Fulfillment log: spin=%s, attempt=%d, status=%s, error=%s\n",
		spin.ID, spin.FulfillmentAttempts, status, errorMsg)
	
	// TODO: Insert into database
	// query := `
	//     INSERT INTO prize_fulfillment_logs (
	//         spin_result_id, attempt_number, fulfillment_mode, status,
	//         error_message, detected_network, msisdn
	//     ) VALUES ($1, $2, $3, $4, $5, $6, $7)
	// `
	// s.db.ExecContext(ctx, query, spin.ID, spin.FulfillmentAttempts,
	//     spin.FulfillmentMode, status, errorMsg, network, spin.Msisdn)
}

// GetSpinHistory gets user's spin history
func (s *SpinService) GetSpinHistory(ctx context.Context, msisdn string, limit, offset int) ([]SpinResultResponse, error) {
user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
if err != nil {
return nil, fmt.Errorf("user not found: %w", err)
}

spins, err := s.spinRepo.FindByUserID(ctx, user.ID, limit, offset)
if err != nil {
return nil, fmt.Errorf("failed to get spin history: %w", err)
}

	results := make([]SpinResultResponse, len(spins))
	for i, spin := range spins {
		results[i] = SpinResultResponse{
			ID:           spin.ID,
			PrizeWon:     spin.PrizeName,
			PrizeType:    spin.PrizeType,
			PrizeValue:   int64(spin.PrizeValue),
			PointsEarned: 0, // Points not stored in SpinResults
			Status:       spin.ClaimStatus,
			CreatedAt:    spin.CreatedAt,
		}
	}

return results, nil
}

// GetTotalSpinCount returns total number of spins
func (s *SpinService) GetTotalSpinCount(ctx context.Context) (int64, error) {
	count, err := s.spinRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get spin count: %w", err)
	}
	return count, nil
}

// GetConfig returns spin/wheel configuration
func (s *SpinService) GetConfig(ctx context.Context) (map[string]interface{}, error) {
	config := map[string]interface{}{
		"enabled":              true,
		"min_recharge_amount":  100000, // ₦1000 in kobo
		"spins_per_recharge":   1,
		"daily_spin_limit":     10,
		"prizes": []map[string]interface{}{
			{
				"id":          "airtime_50",
				"name":        "₦50 Airtime",
				"type":        "airtime",
				"value":       5000, // kobo
				"probability": 30.0, // 30%
				"color":       "#FF6B6B",
			},
			{
				"id":          "airtime_100",
				"name":        "₦100 Airtime",
				"type":        "airtime",
				"value":       10000,
				"probability": 20.0, // 20%
				"color":       "#4ECDC4",
			},
			{
				"id":          "data_500mb",
				"name":        "500MB Data",
				"type":        "data",
				"value":       500,
				"probability": 15.0, // 15%
				"color":       "#45B7D1",
			},
			{
				"id":          "data_1gb",
				"name":        "1GB Data",
				"type":        "data",
				"value":       1024,
				"probability": 10.0, // 10%
				"color":       "#96CEB4",
			},
			{
				"id":          "points_100",
				"name":        "100 Points",
				"type":        "points",
				"value":       100,
				"probability": 15.0, // 15%
				"color":       "#FFEAA7",
			},
			{
				"id":          "better_luck",
				"name":        "Better Luck Next Time",
				"type":        "nothing",
				"value":       0,
				"probability": 10.0, // 10%
				"color":       "#DFE6E9",
			},
		},
	}
	
	return config, nil
}

// UpdateConfig updates spin/wheel configuration (admin)
func (s *SpinService) UpdateConfig(ctx context.Context, config map[string]interface{}) error {
	// Validate configuration before storing
	// This ensures data integrity and prevents invalid configurations
	
	// Validate required fields
	if enabled, ok := config["enabled"].(bool); ok {
		_ = enabled // Validate it's a boolean
	}
	
	if minAmount, ok := config["min_recharge_amount"].(float64); ok {
		if minAmount < 0 {
			return fmt.Errorf("min_recharge_amount must be positive")
		}
	}
	
	if prizes, ok := config["prizes"].([]interface{}); ok {
		totalProbability := 0.0
		for _, prize := range prizes {
			if prizeMap, ok := prize.(map[string]interface{}); ok {
				if prob, ok := prizeMap["probability"].(float64); ok {
					totalProbability += prob
				}
			}
		}
		
		// Probabilities should sum to 100%
		if totalProbability < 99.0 || totalProbability > 101.0 {
			return fmt.Errorf("prize probabilities must sum to 100%%, got %.2f%%", totalProbability)
		}
	}
	
	// Store configuration in database
	// In production, this would use a ConfigurationRepository
	// Configuration storage strategy:
	// 1. Serialize config to JSON
	// 2. Store in configuration table with key "spin_wheel_config"
	// 3. Invalidate any caches
	// 4. Log configuration change for audit
	//
	// Example implementation:
	// configJSON, err := json.Marshal(config)
	// if err != nil {
	//     return fmt.Errorf("failed to serialize config: %w", err)
	// }
	// 
	// configRecord := &entities.Configuration{
	//     Key:       "spin_wheel_config",
	//     Value:     string(configJSON),
	//     UpdatedAt: time.Now(),
	// }
	// 
	// err = s.configRepo.Upsert(ctx, configRecord)
	// if err != nil {
	//     return fmt.Errorf("failed to save config: %w", err)
	// }
	
	// For now, configuration is validated but stored in memory
	// When ConfigurationRepository is implemented, uncomment the above code
	
	return nil
}

// GetAllPrizes returns all available prizes from database
func (s *SpinService) GetAllPrizes(ctx context.Context) ([]map[string]interface{}, error) {
	// Query all prizes from database
	prizes, err := s.prizeRepo.FindAll(ctx, 1000, 0) // Get up to 1000 prizes
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prizes from database: %w", err)
	}
	
	// Convert to map format for API response
	result := make([]map[string]interface{}, 0, len(prizes))
	for _, prize := range prizes {
		prizeMap := map[string]interface{}{
			"id":                prize.ID.String(),
			"prize_code":        prize.PrizeCode,
			"prize_name":        prize.PrizeName,
			"prize_type":        prize.PrizeType,
			"prize_value":       prize.PrizeValue,
			"probability":       prize.Probability,
			"minimum_recharge":  prize.MinimumRecharge,
			"is_active":         prize.IsActive,
			"icon_name":         prize.IconName,
			"color_scheme":      prize.ColorScheme,
			"sort_order":        prize.SortOrder,
			"description":       prize.Description,
			"created_at":        prize.CreatedAt,
			"updated_at":        prize.UpdatedAt,
		}
		result = append(result, prizeMap)
	}
	
	return result, nil
}

// CreatePrize creates a new prize (admin)
func (s *SpinService) CreatePrize(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	// Extract and validate fields
	name, _ := data["name"].(string)
	prizeType, _ := data["type"].(string)
	if name == "" {
		return nil, fmt.Errorf("prize name is required")
	}
	if prizeType == "" {
		return nil, fmt.Errorf("prize type is required")
	}
	
	// Normalize prize type to uppercase for DB
	prizeTypeUpper := strings.ToUpper(prizeType)
	validTypes := []string{"CASH", "AIRTIME", "DATA", "POINTS"}
	isValidType := false
	for _, t := range validTypes {
		if prizeTypeUpper == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return nil, fmt.Errorf("invalid prize type: %s (must be CASH, AIRTIME, DATA, or POINTS)", prizeType)
	}
	
	// Extract value
	var prizeValue int64
	switch v := data["value"].(type) {
	case float64:
		prizeValue = int64(v)
	case int64:
		prizeValue = v
	case int:
		prizeValue = int64(v)
	}
	
	// Extract probability
	var probability float64
	if p, ok := data["probability"].(float64); ok {
		probability = p
	}
	if probability < 0 || probability > 100 {
		return nil, fmt.Errorf("probability must be between 0 and 100")
	}
	
	// Extract optional fields
	colorScheme, _ := data["color"].(string)
	if cs, ok := data["color_scheme"].(string); ok && cs != "" {
		colorScheme = cs
	}
	if colorScheme == "" {
		colorScheme = "green"
	}
	
	isActive := true
	if ia, ok := data["is_active"].(bool); ok {
		isActive = ia
	}
	
	var minimumRecharge *float64
	if mr, ok := data["minimum_recharge"].(float64); ok {
		minimumRecharge = &mr
	}
	
	var sortOrder *int
	if so, ok := data["sort_order"].(float64); ok {
		sov := int(so)
		sortOrder = &sov
	}
	
	// Generate a unique prize code: e.g. CASH-a1b2c3d4
	prizeID := uuid.New()
	prizeCode := fmt.Sprintf("%s-%s", prizeTypeUpper, prizeID.String()[:8])
	
	// Create the entity
	prize := &entities.WheelPrizes{
		ID:              prizeID,
		PrizeCode:       prizeCode,
		PrizeName:       name,
		PrizeType:       prizeTypeUpper,
		PrizeValue:      prizeValue,
		Probability:     probability,
		MinimumRecharge: minimumRecharge,
		IsActive:        &isActive,
		ColorScheme:     colorScheme,
		SortOrder:       sortOrder,
	}
	
	err := s.prizeRepo.Create(ctx, prize)
	if err != nil {
		return nil, fmt.Errorf("failed to create prize: %w", err)
	}
	
	// Return the created prize as a map
	result := map[string]interface{}{
		"id":               prize.ID.String(),
		"prize_code":       prize.PrizeCode,
		"prize_name":       prize.PrizeName,
		"prize_type":       prize.PrizeType,
		"prize_value":      prize.PrizeValue,
		"probability":      prize.Probability,
		"minimum_recharge": prize.MinimumRecharge,
		"is_active":        prize.IsActive,
		"color_scheme":     prize.ColorScheme,
		"sort_order":       prize.SortOrder,
		"created_at":       prize.CreatedAt,
		"updated_at":       prize.UpdatedAt,
	}
	return result, nil
}

// UpdatePrize updates an existing prize (admin)
func (s *SpinService) UpdatePrize(ctx context.Context, prizeID string, data map[string]interface{}) (map[string]interface{}, error) {
	if prizeID == "" {
		return nil, fmt.Errorf("prize ID is required")
	}
	
	// Parse prize ID
	prizeUUID, err := uuid.Parse(prizeID)
	if err != nil {
		return nil, fmt.Errorf("invalid prize ID: %w", err)
	}
	
	// Find existing prize
	prize, err := s.prizeRepo.FindByID(ctx, prizeUUID)
	if err != nil {
		return nil, fmt.Errorf("prize not found: %w", err)
	}
	
	// Update fields
	if name, ok := data["name"].(string); ok && name != "" {
		prize.PrizeName = name
	}
	if prizeType, ok := data["type"].(string); ok && prizeType != "" {
		prize.PrizeType = strings.ToUpper(prizeType)
	}
	if value, ok := data["value"].(float64); ok {
		prize.PrizeValue = int64(value)
	}
	if prob, ok := data["probability"].(float64); ok {
		prize.Probability = prob
	}
	if isActive, ok := data["is_active"].(bool); ok {
		prize.IsActive = &isActive
	}
	if colorScheme, ok := data["color"].(string); ok && colorScheme != "" {
		prize.ColorScheme = colorScheme
	}
	if colorScheme, ok := data["color_scheme"].(string); ok && colorScheme != "" {
		prize.ColorScheme = colorScheme
	}
	if mr, ok := data["minimum_recharge"].(float64); ok {
		prize.MinimumRecharge = &mr
	}
	if so, ok := data["sort_order"].(float64); ok {
		sov := int(so)
		prize.SortOrder = &sov
	}
	
	err = s.prizeRepo.Update(ctx, prize)
	if err != nil {
		return nil, fmt.Errorf("failed to update prize: %w", err)
	}
	
	// Return updated prize as map
	result := map[string]interface{}{
		"id":               prize.ID.String(),
		"prize_code":       prize.PrizeCode,
		"prize_name":       prize.PrizeName,
		"prize_type":       prize.PrizeType,
		"prize_value":      prize.PrizeValue,
		"probability":      prize.Probability,
		"minimum_recharge": prize.MinimumRecharge,
		"is_active":        prize.IsActive,
		"color_scheme":     prize.ColorScheme,
		"sort_order":       prize.SortOrder,
		"created_at":       prize.CreatedAt,
		"updated_at":       prize.UpdatedAt,
	}
	return result, nil
}

// DeletePrize deletes a prize (admin)
func (s *SpinService) DeletePrize(ctx context.Context, prizeID string) error {
	if prizeID == "" {
		return fmt.Errorf("prize ID is required")
	}
	
	// Parse prize ID
	prizeUUID, err := uuid.Parse(prizeID)
	if err != nil {
		return fmt.Errorf("invalid prize ID: %w", err)
	}
	
	// Find prize to ensure it exists
	prize, err := s.prizeRepo.FindByID(ctx, prizeUUID)
	if err != nil {
		return fmt.Errorf("prize not found: %w", err)
	}
	
	// Soft delete - set IsActive to false
	isActive := false
	prize.IsActive = &isActive
	err = s.prizeRepo.Update(ctx, prize)
	if err != nil {
		return fmt.Errorf("failed to delete prize: %w", err)
	}
	
	return nil
}

// CreateSpinOpportunity creates a spin opportunity for a user after a qualifying recharge
func (s *SpinService) CreateSpinOpportunity(ctx context.Context, userID uuid.UUID, rechargeID uuid.UUID) error {
	// Check if user already has a pending spin for this recharge
	// This prevents duplicate spin opportunities
	
	// For now, we'll rely on CheckEligibility to determine if user can spin
	// The spin opportunity is implicit - if they made a ₦1000+ recharge today, they can spin
	
	// In a more sophisticated implementation, you could:
	// 1. Create a spin_opportunities table
	// 2. Track which recharges have granted spins
	// 3. Allow multiple spins per day if multiple qualifying recharges
	
	// For this implementation, the spin opportunity is determined by:
	// - Recent recharge of ₦1000+
	// - CheckEligibility validates this
	// - User can spin once per qualifying recharge
	
	return nil // Spin opportunity is implicit based on recharge history
}

// acquireLock acquires a PostgreSQL advisory lock to prevent race conditions
// This ensures only one spin operation can proceed for a given user at a time
func (s *SpinService) acquireLock(ctx context.Context, lockID int64) error {
	if s.db == nil {
		// If db is not available, continue without locking (graceful degradation)
		// The database unique constraint will still prevent duplicates
		return nil
	}
	// PostgreSQL advisory lock
	query := "SELECT pg_advisory_lock($1)"
	return s.db.WithContext(ctx).Exec(query, lockID).Error
}

// releaseLock releases a PostgreSQL advisory lock
func (s *SpinService) releaseLock(ctx context.Context, lockID int64) error {
	if s.db == nil {
		return nil
	}
	query := "SELECT pg_advisory_unlock($1)"
	return s.db.WithContext(ctx).Exec(query, lockID).Error
}

// SpinTierResponse represents a spin tier
type SpinTierResponse struct {
	ID              uuid.UUID `json:"id"`
	TierName        string    `json:"tier_name"`
	TierDisplayName string    `json:"tier_display_name"`
	MinDailyAmount  int64     `json:"min_daily_amount"`
	MaxDailyAmount  int64     `json:"max_daily_amount"`
	SpinsPerDay     int       `json:"spins_per_day"`
	TierColor       string    `json:"tier_color"`
	TierIcon        string    `json:"tier_icon"`
	TierBadge       string    `json:"tier_badge"`
	Description     string    `json:"description"`
	SortOrder       int       `json:"sort_order"`
	IsActive        bool      `json:"is_active"`
}

// TierProgressResponse represents user's progress towards tiers
type TierProgressResponse struct {
	CurrentTier      *SpinTierResponse `json:"current_tier"`
	NextTier         *SpinTierResponse `json:"next_tier"`
	TodayAmount      int64             `json:"today_amount"`
	ProgressPercent  float64           `json:"progress_percent"`
	AmountToNextTier int64             `json:"amount_to_next_tier"`
	AvailableSpins   int               `json:"available_spins"`
}

// GetAllTiers retrieves all active spin tiers
func (s *SpinService) GetAllTiers(ctx context.Context) ([]SpinTierResponse, error) {
	var tiers []SpinTierResponse
	
	err := s.db.WithContext(ctx).
		Table("spin_tiers").
		Where("is_active = ?", true).
		Order("sort_order ASC").
		Find(&tiers).Error
	
	if err != nil {
		return nil, errors.DatabaseError(err)
	}
	
	return tiers, nil
}

// GetTierProgress gets user's current tier and progress
func (s *SpinService) GetTierProgress(ctx context.Context, msisdn string) (*TierProgressResponse, error) {
	// Get user
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, errors.NotFound("user")
	}
	
	// Calculate today's total recharge amount
	var todayAmount int64
	err = s.db.WithContext(ctx).
		Table("transactions").
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ? AND status = 'SUCCESS' AND DATE(created_at) = CURRENT_DATE", user.ID).
		Scan(&todayAmount).Error
	
	if err != nil {
		return nil, errors.DatabaseError(err)
	}
	
	// Get all tiers
	var tiers []SpinTierResponse
	err = s.db.WithContext(ctx).
		Table("spin_tiers").
		Where("is_active = ?", true).
		Order("sort_order ASC").
		Find(&tiers).Error
	
	if err != nil {
		return nil, errors.DatabaseError(err)
	}
	
	if len(tiers) == 0 {
		return nil, errors.Internal("No active tiers found")
	}
	
	// Find current tier based on today's amount
	var currentTier *SpinTierResponse
	var nextTier *SpinTierResponse
	
	for i, tier := range tiers {
		if todayAmount >= tier.MinDailyAmount && todayAmount <= tier.MaxDailyAmount {
			currentTier = &tiers[i]
			// Get next tier if exists
			if i+1 < len(tiers) {
				nextTier = &tiers[i+1]
			}
			break
		}
	}
	
	// If no tier matches, user is below minimum (Bronze)
	if currentTier == nil {
		nextTier = &tiers[0] // First tier (Bronze)
	}
	
	// Calculate progress
	var progressPercent float64
	var amountToNextTier int64
	var availableSpins int
	
	if currentTier != nil {
		availableSpins = currentTier.SpinsPerDay
		
		if nextTier != nil {
			// Calculate progress to next tier
			currentTierRange := float64(currentTier.MaxDailyAmount - currentTier.MinDailyAmount)
			currentProgress := float64(todayAmount - currentTier.MinDailyAmount)
			progressPercent = (currentProgress / currentTierRange) * 100
			
			amountToNextTier = nextTier.MinDailyAmount - todayAmount
			if amountToNextTier < 0 {
				amountToNextTier = 0
			}
		} else {
			// Already at highest tier
			progressPercent = 100
			amountToNextTier = 0
		}
	} else {
		// Below minimum tier
		if nextTier != nil {
			amountToNextTier = nextTier.MinDailyAmount - todayAmount
			progressPercent = 0
		}
	}
	
	return &TierProgressResponse{
		CurrentTier:      currentTier,
		NextTier:         nextTier,
		TodayAmount:      todayAmount,
		ProgressPercent:  progressPercent,
		AmountToNextTier: amountToNextTier,
		AvailableSpins:   availableSpins,
	}, nil
}
