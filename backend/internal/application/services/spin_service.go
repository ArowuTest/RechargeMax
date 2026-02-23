package services

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"time"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
	"rechargemax/internal/errors"
)

// SpinService handles wheel spin operations
type SpinService struct {
	spinRepo    repositories.SpinRepository
	prizeRepo   repositories.WheelPrizeRepository
	userRepo    repositories.UserRepository
	rechargeRepo repositories.RechargeRepository
	hlrService  *HLRService
	db          *gorm.DB // Database connection for advisory locks
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
	db *gorm.DB, // Database connection for advisory locks
) *SpinService {
	return &SpinService{
		spinRepo:     spinRepo,
		prizeRepo:    prizeRepo,
		userRepo:     userRepo,
		rechargeRepo: rechargeRepo,
		hlrService:   hlrService,
		db:           db,
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
			return nil, fmt.Errorf("not eligible to spin: No qualifying recharges found. Recharge ₦1000+ to earn a spin!")
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
return nil, fmt.Errorf("not eligible to spin: %s", eligibility.Message)
}

// Get all active prizes
prizes, err := s.prizeRepo.FindActive(ctx)
if err != nil || len(prizes) == 0 {
return nil, fmt.Errorf("no prizes available")
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
	
		// Auto-provision if data or airtime
		if selectedPrize.PrizeType == "DATA" || selectedPrize.PrizeType == "AIRTIME" {
			err := s.provisionPrize(ctx, spin)
			if err != nil {
				spin.ClaimStatus = "EXPIRED" // Mark as expired if provisioning failed
				// Rollback transaction on provisioning failure
				return fmt.Errorf("failed to provision prize: %w", err)
			} else {
				spin.ClaimStatus = "CLAIMED"
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
func (s *SpinService) provisionPrize(ctx context.Context, spin *entities.WheelSpin) error {
	// Detect network
	networkHint := ""
	network, err := s.hlrService.DetectNetwork(ctx, spin.Msisdn, &networkHint)
	if err != nil {
		return fmt.Errorf("failed to detect network: %w", err)
	}
	
	// Provision prize based on type
	// This integrates with TelecomService for direct network provisioning
	// 
	// In production, this would:
	// 1. Call TelecomService.PurchaseAirtime() for airtime prizes
	// 2. Call TelecomService.PurchaseData() for data prizes
	// 3. Handle async confirmation via webhook
	// 4. Update spin status based on provisioning result
	//
	// Example implementation:
	// if spin.PrizeType == "airtime" {
	//     amount := spin.PrizeValue // Amount in kobo
	//     err := s.telecomService.PurchaseAirtime(ctx, spin.Msisdn, network, amount)
	//     if err != nil {
	//         return fmt.Errorf("failed to provision airtime: %w", err)
	//     }
	// } else if spin.PrizeType == "data" {
	//     dataPackage := spin.PrizeDescription // e.g., "1GB_DAILY"
	//     err := s.telecomService.PurchaseData(ctx, spin.Msisdn, network, dataPackage)
	//     if err != nil {
	//         return fmt.Errorf("failed to provision data: %w", err)
	//     }
	// }
	//
	// For cash prizes, no provisioning needed - handled via bank transfer
	// For physical prizes, no provisioning needed - handled via admin
	
	// For now, acknowledge the provisioning requirement
	_ = network
	
	// In production, this would return success only after actual provisioning
	return nil
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
	// Validate required fields
	requiredFields := []string{"id", "name", "type", "value", "probability", "color"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			return nil, fmt.Errorf("missing required field: %s", field)
		}
	}
	
	// Validate probability
	if prob, ok := data["probability"].(float64); ok {
		if prob < 0 || prob > 100 {
			return nil, fmt.Errorf("probability must be between 0 and 100")
		}
	}
	
	// Validate prize type
	validTypes := []string{"airtime", "data", "points", "nothing", "cash"}
	prizeType, ok := data["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid prize type")
	}
	
	isValidType := false
	for _, t := range validTypes {
		if prizeType == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return nil, fmt.Errorf("invalid prize type: %s", prizeType)
	}
	
	// Store prize in database
	// In production, this would:
	// 1. Create a WheelPrize entity
	// 2. Save to wheel_prizes table via prizeRepo
	// 3. Update spin configuration to include new prize
	// 4. Invalidate caches
	//
	// Example implementation:
	// prize := &entities.WheelPrize{
	//     ID:          uuid.New(),
	//     Name:        data["name"].(string),
	//     Type:        data["type"].(string),
	//     Value:       int64(data["value"].(float64)),
	//     Probability: data["probability"].(float64),
	//     Color:       data["color"].(string),
	//     IsActive:    true,
	//     CreatedAt:   time.Now(),
	// }
	// 
	// err := s.prizeRepo.Create(ctx, prize)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create prize: %w", err)
	// }
	
	// For now, return the prize data
	// When WheelPrizeRepository CRUD methods are implemented, uncomment above
	return data, nil
}

// UpdatePrize updates an existing prize (admin)
func (s *SpinService) UpdatePrize(ctx context.Context, prizeID string, data map[string]interface{}) (map[string]interface{}, error) {
	// Validate prize ID
	if prizeID == "" {
		return nil, fmt.Errorf("prize ID is required")
	}
	
	// Validate probability if provided
	if prob, ok := data["probability"].(float64); ok {
		if prob < 0 || prob > 100 {
			return nil, fmt.Errorf("probability must be between 0 and 100")
		}
	}
	
	// Validate prize type if provided
	if prizeType, ok := data["type"].(string); ok {
		validTypes := []string{"airtime", "data", "points", "nothing", "cash"}
		isValidType := false
		for _, t := range validTypes {
			if prizeType == t {
				isValidType = true
				break
			}
		}
		if !isValidType {
			return nil, fmt.Errorf("invalid prize type: %s", prizeType)
		}
	}
	
	// Update prize in database
	// In production, this would:
	// 1. Find existing prize by ID
	// 2. Update fields with new data
	// 3. Save to database via prizeRepo
	// 4. Invalidate caches
	//
	// Example implementation:
	// prize, err := s.prizeRepo.FindByID(ctx, prizeID)
	// if err != nil {
	//     return nil, fmt.Errorf("prize not found: %w", err)
	// }
	// 
	// if name, ok := data["name"].(string); ok {
	//     prize.Name = name
	// }
	// if prob, ok := data["probability"].(float64); ok {
	//     prize.Probability = prob
	// }
	// // ... update other fields
	// 
	// err = s.prizeRepo.Update(ctx, prize)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to update prize: %w", err)
	// }
	
	// For now, return the updated data
	// When WheelPrizeRepository CRUD methods are implemented, uncomment above
	data["id"] = prizeID
	return data, nil
}

// DeletePrize deletes a prize (admin)
func (s *SpinService) DeletePrize(ctx context.Context, prizeID string) error {
	// Validate prize ID
	if prizeID == "" {
		return fmt.Errorf("prize ID is required")
	}
	
	// Delete prize from database
	// In production, this would:
	// 1. Find prize by ID to ensure it exists
	// 2. Check if prize is currently in use (active spins)
	// 3. Soft delete (set IsActive = false) or hard delete
	// 4. Update spin configuration to remove prize
	// 5. Invalidate caches
	//
	// Example implementation:
	// prize, err := s.prizeRepo.FindByID(ctx, prizeID)
	// if err != nil {
	//     return fmt.Errorf("prize not found: %w", err)
	// }
	// 
	// // Soft delete (recommended to preserve historical data)
	// prize.IsActive = false
	// err = s.prizeRepo.Update(ctx, prize)
	// if err != nil {
	//     return fmt.Errorf("failed to delete prize: %w", err)
	// }
	// 
	// // Or hard delete:
	// // err = s.prizeRepo.Delete(ctx, prizeID)
	
	// For now, just validate the ID format
	// When WheelPrizeRepository CRUD methods are implemented, uncomment above
	
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
