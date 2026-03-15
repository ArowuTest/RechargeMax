package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// UserService handles user-related operations
type UserService struct {
	userRepo         repositories.UserRepository
	transactionRepo  repositories.TransactionRepository
	bankAccountRepo  repositories.BankAccountRepository
	withdrawalRepo   repositories.WithdrawalRepository
	rechargeRepo     repositories.RechargeRepository
	spinRepo         repositories.SpinRepository
	subscriptionRepo repositories.SubscriptionRepository
	db               *gorm.DB
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// UserProfile represents user profile data
type UserProfile struct {
	ID           uuid.UUID  `json:"id"`
	MSISDN       string     `json:"msisdn"`
	FirstName    string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Email        string     `json:"email"`
	LoyaltyTier  string     `json:"loyalty_tier"`
	TotalPoints  int64      `json:"total_points"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	ReferralCode string     `json:"referral_code"`
	UserCode     string     `json:"user_code"`
	FullName     string     `json:"full_name"`
}

// UserSummaryResponse represents user summary data
type UserSummaryResponse struct {
	User                *UserProfile           `json:"user"`
	Stats               map[string]interface{} `json:"stats"`
	RecentTransactions  []ActivityItem         `json:"recent_transactions"`
}

// ActivityItem represents a user activity item (DEPRECATED - use TransactionItem)
type ActivityItem struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Amount      int64     `json:"amount,omitempty"`
	Points      int64     `json:"points,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// TransactionItem represents a transaction for dashboard display
type TransactionItem struct {
	ID              uuid.UUID `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	NetworkProvider string    `json:"network_provider"`
	RechargeType    string    `json:"recharge_type"`
	Amount          int64     `json:"amount"`
	PointsEarned    int64     `json:"points_earned"`
	Status          string    `json:"status"`
}

// DashboardSummary represents dashboard summary statistics
type DashboardSummary struct {
	TotalTransactions         int64 `json:"total_transactions"`
	TotalPrizes              int64 `json:"total_prizes"`
	PendingPrizes            int64 `json:"pending_prizes"`
	ClaimedPrizes            int64 `json:"claimed_prizes"`
	TotalAmountRecharged     int64 `json:"total_amount_recharged"`
	TotalSubscriptions       int64 `json:"total_subscriptions"`
	TotalSubscriptionAmount  int64 `json:"total_subscription_amount"`
	TotalSubscriptionEntries int64 `json:"total_subscription_entries"`
	TotalSubscriptionPoints  int64 `json:"total_subscription_points"`
}

// DashboardResponse represents dashboard data
type DashboardResponse struct {
	User                *UserProfile           `json:"user"`
	Stats               map[string]interface{} `json:"stats"`
	Summary             *DashboardSummary      `json:"summary"`
	RecentTransactions  []TransactionItem      `json:"recent_transactions"`
	Subscriptions       []SubscriptionItem     `json:"subscriptions"`
	Prizes              []PrizeItem            `json:"prizes"`
	PendingSpins        int64                  `json:"pending_spins"`
	UnclaimedPrizes     int64                  `json:"unclaimed_prizes"`
	NextDrawDate        string                 `json:"next_draw_date"`
}

// SubscriptionItem represents a subscription entry
type SubscriptionItem struct {
	ID              uuid.UUID `json:"id"`
	TransactionDate time.Time `json:"transaction_date"`
	Reference       string    `json:"reference"`
	Amount          int64     `json:"amount"`
	Entries         int       `json:"entries"`
	PointsEarned    int       `json:"points_earned"`
	Status          string    `json:"status"`
}

// PrizeItem represents a prize won by user
type PrizeItem struct {
	ID          uuid.UUID `json:"id"`
	PrizeName   string    `json:"prize_name"`
	PrizeType   string    `json:"prize_type"`
	PrizeValue  int64     `json:"prize_value"`
	Status      string     `json:"status"`
	WonAt       time.Time  `json:"won_at"`
	WonDate     string     `json:"won_date"`
	ClaimedAt   *time.Time `json:"claimed_at,omitempty"`
	ClaimDate   *string    `json:"claim_date,omitempty"`
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repositories.UserRepository,
	transactionRepo repositories.TransactionRepository,
	bankAccountRepo repositories.BankAccountRepository,
	withdrawalRepo repositories.WithdrawalRepository,
	rechargeRepo repositories.RechargeRepository,
	spinRepo repositories.SpinRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	db *gorm.DB,
) *UserService {
	return &UserService{
		userRepo:         userRepo,
		transactionRepo:  transactionRepo,
		bankAccountRepo:  bankAccountRepo,
		withdrawalRepo:   withdrawalRepo,
		rechargeRepo:     rechargeRepo,
		spinRepo:         spinRepo,
		subscriptionRepo: subscriptionRepo,
		db:               db,
	}
}

// GetUserProfile gets user profile by MSISDN
func (s *UserService) GetUserProfile(ctx context.Context, msisdn string) (*UserProfile, error) {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &UserProfile{
		ID:           user.ID,
		MSISDN:       user.MSISDN,
		FirstName:    user.FullName,
		LastName:     "",
		FullName:     user.FullName,
		Email:        user.Email,
		LoyaltyTier:  user.LoyaltyTier,
		TotalPoints:  int64(user.TotalPoints),
		IsActive:     user.IsActive,
		LastLoginAt:  user.LastLoginAt,
		CreatedAt:    user.CreatedAt,
		ReferralCode: user.ReferralCode,
		UserCode:     user.UserCode,
	}, nil
}

// UpdateUserProfile updates user profile information
func (s *UserService) UpdateUserProfile(ctx context.Context, msisdn string, req UpdateProfileRequest) (*UserProfile, error) {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields
	if req.FirstName != "" {
		user.FullName = req.FirstName
	}
	if req.LastName != "" {
		// LastName not supported in Users entity
	}
	if req.Email != "" {
		// Check if email is already taken by another user
		existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
		if err == nil && existingUser.MSISDN != msisdn {
			return nil, fmt.Errorf("email already in use by another account")
		}
		user.Email = req.Email
	}

	// Save updates
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &UserProfile{
		ID:           user.ID,
		MSISDN:       user.MSISDN,
		FirstName:    user.FullName,
		LastName:     "",
		FullName:     user.FullName,
		Email:        user.Email,
		LoyaltyTier:  user.LoyaltyTier,
		TotalPoints:  int64(user.TotalPoints),
		IsActive:     user.IsActive,
		LastLoginAt:  user.LastLoginAt,
		CreatedAt:    user.CreatedAt,
		ReferralCode: user.ReferralCode,
		UserCode:     user.UserCode,
	}, nil
}

// GetDashboard gets user dashboard data
func (s *UserService) GetDashboard(ctx context.Context, msisdn string) (*DashboardResponse, error) {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get stats
	stats := make(map[string]interface{})

	// Total recharges
	totalRecharges, err := s.rechargeRepo.CountByUserID(ctx, user.ID)
	if err == nil {
		stats["total_recharges"] = totalRecharges
	}

	// Total spins
	totalSpins, err := s.spinRepo.CountByUserID(ctx, user.ID)
	if err == nil {
		stats["total_spins"] = totalSpins
	}

	// Pending spins
	// Get all spins for this user and count pending ones
	allSpins, err := s.spinRepo.FindByUserID(ctx, user.ID, 100, 0)
	var pendingSpins int64 = 0
	if err == nil {
		for _, spin := range allSpins {
			if spin.ClaimStatus == "PENDING" {
				pendingSpins++
			}
		}
		stats["pending_spins"] = pendingSpins
	} else {
		pendingSpins = 0
	}

	// Total points
	stats["total_points"] = user.TotalPoints
	stats["loyalty_tier"] = user.LoyaltyTier

		// Get recent transactions (last 10)
		recentTransactions := s.getRecentTransactions(ctx, user.ID)

	// Calculate next draw date (assuming daily draws at 8 PM)
	nextDraw := time.Now().AddDate(0, 0, 1)
	nextDraw = time.Date(nextDraw.Year(), nextDraw.Month(), nextDraw.Day(), 20, 0, 0, 0, nextDraw.Location())

	// Build summary statistics
	claimedPrizes := totalSpins - pendingSpins
	if claimedPrizes < 0 {
		claimedPrizes = 0
	}

		// Get total amount recharged
		totalAmountRecharged := s.getTotalAmountRecharged(ctx, user.ID)

		// Get subscription stats
		totalSubscriptions, totalSubAmount, totalSubEntries, totalSubPoints := s.getSubscriptionStats(ctx, user.MSISDN)

		summary := &DashboardSummary{
			TotalTransactions:         totalRecharges,
			TotalPrizes:              totalSpins,
			PendingPrizes:            pendingSpins,
			ClaimedPrizes:            claimedPrizes,
			TotalAmountRecharged:     totalAmountRecharged,
			TotalSubscriptions:       totalSubscriptions,
			TotalSubscriptionAmount:  totalSubAmount,
			TotalSubscriptionEntries: totalSubEntries,
			TotalSubscriptionPoints:  totalSubPoints,
		}

		// Get subscriptions
		subscriptions := s.getUserSubscriptions(ctx, user.MSISDN)

		// Get prizes
		prizes := s.getUserPrizes(ctx, user.ID)

		return &DashboardResponse{
			User: &UserProfile{
				ID:           user.ID,
				MSISDN:       user.MSISDN,
				FirstName:    user.FullName,
				LastName:     "",
				FullName:     user.FullName,
				Email:        user.Email,
				LoyaltyTier:  user.LoyaltyTier,
				TotalPoints:  int64(user.TotalPoints),
				IsActive:     user.IsActive,
				LastLoginAt:  user.LastLoginAt,
				CreatedAt:    user.CreatedAt,
				ReferralCode: user.ReferralCode,
				UserCode:     user.UserCode,
			},
			Stats:              stats,
			Summary:            summary,
			RecentTransactions: recentTransactions,
			Subscriptions:      subscriptions,
			Prizes:             prizes,
			PendingSpins:       pendingSpins,
			UnclaimedPrizes:    s.getUnclaimedPrizesCount(ctx, user.MSISDN),
			NextDrawDate:       nextDraw.Format("2006-01-02 15:04:05"),
		}, nil
}

// getRecentActivity gets recent user activity
func (s *UserService) getRecentActivity(ctx context.Context, userID uuid.UUID) ([]ActivityItem, error) {
	var activities []ActivityItem

	// Get recent recharges (last 5)
	recharges, err := s.rechargeRepo.FindByUserID(ctx, userID, 5, 0)
	if err == nil {
		for _, r := range recharges {
			activities = append(activities, ActivityItem{
				Type:        "recharge",
				Description: fmt.Sprintf("Recharged %s %s", r.NetworkProvider, r.RechargeType),
				Amount:      int64(r.Amount),
				Points:      int64(r.PointsEarned),
				CreatedAt:   r.CreatedAt,
			})
		}
	}

	// Get recent spins (last 5)
	spins, err := s.spinRepo.FindByUserID(ctx, userID, 5, 0)
	if err == nil {
		for _, sp := range spins {
			activities = append(activities, ActivityItem{
				Type:        "spin",
				Description: fmt.Sprintf("Won: %s", sp.PrizeName),
				Points:      0, // Points not stored in SpinResults
				CreatedAt:   sp.CreatedAt,
			})
		}
	}

	// Sort by created_at descending (most recent first)
	for i := 0; i < len(activities)-1; i++ {
		for j := i + 1; j < len(activities); j++ {
			if activities[j].CreatedAt.After(activities[i].CreatedAt) {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}

	// Return top 10
	if len(activities) > 10 {
		activities = activities[:10]
	}

	return activities, nil
}

// GetTransactions gets user transaction history
func (s *UserService) GetTransactions(ctx context.Context, msisdn string, limit, offset int) ([]*entities.Transactions, error) {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	transactions, err := s.transactionRepo.FindByUserID(ctx, user.ID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	return transactions, nil
}

// GetBankAccounts gets user bank accounts
func (s *UserService) GetBankAccounts(ctx context.Context, msisdn string) ([]*entities.BankAccounts, error) {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	accounts, err := s.bankAccountRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank accounts: %w", err)
	}

	return accounts, nil
}

// AddBankAccount adds a new bank account for user
func (s *UserService) AddBankAccount(ctx context.Context, msisdn string, account *entities.BankAccount) error {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	account.UserId = user.ID
	account.Id = uuid.New()
	account.CreatedAt = time.Now()

	if err := s.bankAccountRepo.Create(ctx, account); err != nil {
		return fmt.Errorf("failed to add bank account: %w", err)
	}

	return nil
}

// RequestWithdrawal creates a withdrawal request
func (s *UserService) RequestWithdrawal(ctx context.Context, msisdn string, amount int64, bankAccountID uuid.UUID) (*entities.Withdrawal, error) {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify bank account belongs to user
	account, err := s.bankAccountRepo.FindByID(ctx, bankAccountID)
	if err != nil {
		return nil, fmt.Errorf("bank account not found: %w", err)
	}
	if account.UserId != user.ID {
		return nil, fmt.Errorf("bank account does not belong to user")
	}

	// Create withdrawal
	withdrawal := &entities.Withdrawal{
		Id:            uuid.New(),
		UserId:        user.ID,
		BankAccountId: bankAccountID,
		Amount:        amount,
		Status:        "pending",
	}

	if err := s.withdrawalRepo.Create(ctx, withdrawal); err != nil {
		return nil, fmt.Errorf("failed to create withdrawal: %w", err)
	}

	return withdrawal, nil
}

// GetWithdrawals gets user withdrawal history
func (s *UserService) GetWithdrawals(ctx context.Context, msisdn string, limit, offset int) ([]*entities.WithdrawalRequests, error) {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	withdrawals, err := s.withdrawalRepo.FindByUserID(ctx, user.ID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}

	return withdrawals, nil
}

// GetUserPoints gets current points for a user
func (s *UserService) GetUserPoints(ctx context.Context, msisdn string) (map[string]interface{}, error) {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		// Return zero points for non-existent users
		return map[string]interface{}{
			"total_points":    int64(0),
			"loyalty_tier":    "bronze",
			"next_tier":       "silver",
			"points_to_next":  int64(1000),
		}, nil
	}

	// Calculate points to next tier
	var nextTier string
	var pointsToNext int64

	switch user.LoyaltyTier {
	case "bronze":
		nextTier = "silver"
		pointsToNext = int64(1000 - user.TotalPoints)
		if pointsToNext < 0 {
			pointsToNext = 0
		}
	case "silver":
		nextTier = "gold"
		pointsToNext = int64(5000 - user.TotalPoints)
		if pointsToNext < 0 {
			pointsToNext = 0
		}
	case "gold":
		nextTier = "gold"
		pointsToNext = 0
	}

	return map[string]interface{}{
		"total_points":   user.TotalPoints,
		"loyalty_tier":   user.LoyaltyTier,
		"next_tier":      nextTier,
		"points_to_next": pointsToNext,
	}, nil
}

// DeactivateUser deactivates a user account
func (s *UserService) DeactivateUser(ctx context.Context, msisdn string) error {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	// Use targeted update to avoid full-row save and unique constraint violations
	return s.userRepo.UpdateStatus(ctx, user.ID, false)
}

// ReactivateUser reactivates a user account
func (s *UserService) ReactivateUser(ctx context.Context, msisdn string) error {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	// Use targeted update to avoid full-row save and unique constraint violations
	return s.userRepo.UpdateStatus(ctx, user.ID, true)
}

// GetUserByID gets a user by ID
func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// PrizeResponse represents a user prize
type PrizeResponse struct {
	ID          uuid.UUID  `json:"id"`
	PrizeName   string     `json:"prize_name"`
	PrizeType   string     `json:"prize_type"`
	PrizeValue  string     `json:"prize_value"`
	WonAt       time.Time  `json:"won_at"`
	Status      string     `json:"status"`
	ClaimedAt   *time.Time `json:"claimed_at,omitempty"`
}

// GetUserPrizes retrieves all prizes won by a user
func (s *UserService) GetUserPrizes(ctx context.Context, msisdn string) ([]PrizeResponse, error) {
	// Get user
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get all spin results for the user (limit 100, offset 0)
	spins, err := s.spinRepo.FindByUserID(ctx, user.ID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get user prizes: %w", err)
	}

	// Convert to response format
	var result []PrizeResponse
	for _, spin := range spins {
		result = append(result, PrizeResponse{
			ID:          spin.ID,
			PrizeName:   spin.PrizeName,
			PrizeType:   spin.PrizeType,
			PrizeValue:  fmt.Sprintf("%.2f", float64(spin.PrizeValue)/100.0), // Convert kobo to naira
			WonAt:       spin.CreatedAt,
			Status:      spin.ClaimStatus,
			ClaimedAt:   spin.ClaimedAt,
		})
	}

	return result, nil
}


// GetUserCount returns the total number of users
func (s *UserService) GetUserCount(ctx context.Context) (int64, error) {
	count, err := s.userRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}
	return count, nil
}

// GetAllUsers returns paginated list of all users
func (s *UserService) GetAllUsers(ctx context.Context, page, perPage int) ([]*UserProfile, int64, error) {
	// Calculate offset
	offset := (page - 1) * perPage
	
	// Get users from repository
	users, err := s.userRepo.FindAll(ctx, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}
	
	// Get total count
	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}
	
	// Convert to profile format
	var profiles []*UserProfile
	for _, user := range users {
		profiles = append(profiles, &UserProfile{
			ID:           user.ID,
			MSISDN:       user.MSISDN,
			FirstName:    user.FullName,
			LastName:     "",
			FullName:     user.FullName,
			Email:        user.Email,
			LoyaltyTier:  user.LoyaltyTier,
			TotalPoints:  int64(user.TotalPoints),
			IsActive:     user.IsActive,
			LastLoginAt:  user.LastLoginAt,
			CreatedAt:    user.CreatedAt,
			ReferralCode: user.ReferralCode,
			UserCode:     user.UserCode,
		})
	}
	
	return profiles, total, nil
}

// GetUserByID returns a single user by ID (accepts string ID)

// getUnclaimedPrizesCount gets count of unclaimed prizes for a user
func (s *UserService) getUnclaimedPrizesCount(ctx context.Context, msisdn string) int64 {
	// In production, this would query the winners table:
	// winners, err := s.winnerRepo.FindByMSISDN(ctx, msisdn)
	// if err != nil {
	//     return 0
	// }
	// 
	// var unclaimedCount int64 = 0
	// for _, winner := range winners {
	//     if winner.ClaimStatus == "pending" || winner.ClaimStatus == "unclaimed" {
	//         unclaimedCount++
	//     }
	// }
	// return unclaimedCount
	
	// For now, return 0 (no unclaimed prizes)
	// When WinnerRepository is properly integrated, uncomment above code
	return 0
}

// GetUserAnalytics retrieves user analytics data
func (s *UserService) GetUserAnalytics(ctx context.Context) (map[string]interface{}, error) {
	// Get all users (using large limit to get all)
	users, err := s.userRepo.FindAll(ctx, 100000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	
	totalUsers := int64(len(users))
	var activeUsers int64
	var suspendedUsers int64
	var bannedUsers int64
	
	// Count users by status
	for _, user := range users {
		if user.IsActive {
			activeUsers++
		}
		// Note: User entity doesn't have Status field, so we use IsActive
		// In production, you might want to add a Status field to User entity
	}
	
	// Calculate today's new users
	todayStart := time.Now().Truncate(24 * time.Hour)
	var todayNewUsers int64
	
	for _, user := range users {
		if user.CreatedAt.After(todayStart) {
			todayNewUsers++
		}
	}
	
	// Calculate this month's new users
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
	var monthNewUsers int64
	
	for _, user := range users {
		if user.CreatedAt.After(monthStart) {
			monthNewUsers++
		}
	}
	
	analytics := map[string]interface{}{
		"total_users":      totalUsers,
		"active_users":     activeUsers,
		"suspended_users":  suspendedUsers,
		"banned_users":     bannedUsers,
		"today_new_users":  todayNewUsers,
		"month_new_users":  monthNewUsers,
	}
	
	return analytics, nil
}

// UpdateUser updates user status and points (admin function)
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, status string, points int64) (*entities.User, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	
	// Update status if provided
	if status != "" {
		switch status {
		case "ACTIVE":
			user.IsActive = true
		case "SUSPENDED", "BANNED":
			user.IsActive = false
		default:
			return nil, fmt.Errorf("invalid status: %s", status)
		}
	}
	
	// Update points if provided
	if points != 0 {
		user.TotalPoints += int(points)
		if user.TotalPoints < 0 {
			user.TotalPoints = 0
		}
	}
	
	// Save user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	return user, nil
}


// getTotalAmountRecharged calculates total amount recharged by user
func (s *UserService) getTotalAmountRecharged(ctx context.Context, userID uuid.UUID) int64 {
	recharges, err := s.rechargeRepo.FindByUserID(ctx, userID, 10000, 0) // Get all recharges
	if err != nil {
		return 0
	}

	var total int64
	for _, r := range recharges {
		total += int64(r.Amount)
	}
	return total / 100 // Convert kobo to naira
}

// getSubscriptionStats calculates subscription statistics
func (s *UserService) getSubscriptionStats(ctx context.Context, msisdn string) (count int64, amount int64, entries int64, points int64) {
	// Try to find active subscription for this MSISDN
	activeSub, err := s.subscriptionRepo.FindActiveByMSISDN(ctx, msisdn)
	if err != nil {
		// No active subscription found
		return 0, 0, 0, 0
	}

	// If active subscription exists, count it
	if activeSub != nil {
		count = 1
		amount = int64(activeSub.Amount)
		entries = 1
		// Calculate points (assuming same logic as recharge: 1 point per N100)
		points = int64(activeSub.Amount) / 100
	}
	return
}

// getUserSubscriptions gets user's active subscriptions
func (s *UserService) getUserSubscriptions(ctx context.Context, msisdn string) []SubscriptionItem {
	// Try to find active subscription for this MSISDN
	activeSub, err := s.subscriptionRepo.FindActiveByMSISDN(ctx, msisdn)
	if err != nil || activeSub == nil {
		return []SubscriptionItem{} // Return empty array instead of nil
	}

	// Map entity fields to response fields
	var entries int
	if activeSub.DrawEntriesEarned != nil {
		entries = *activeSub.DrawEntriesEarned
	} else {
		entries = 1 // Default
	}

	var pointsEarned int
	if activeSub.PointsEarned != nil {
		pointsEarned = *activeSub.PointsEarned
	} else {
		pointsEarned = 0
	}

	var result []SubscriptionItem
	result = append(result, SubscriptionItem{
		ID:              activeSub.ID,
		TransactionDate: activeSub.SubscriptionDate,
		Reference:       activeSub.SubscriptionCode,
		Amount:          activeSub.Amount,
		Entries:         entries,
		PointsEarned:    pointsEarned,
		Status:          activeSub.Status,
	})
	return result
}

// getRecentTransactions gets recent transactions from transactions table
func (s *UserService) getRecentTransactions(ctx context.Context, userID uuid.UUID) []TransactionItem {
	// Query transactions table for recent transactions
	transactions, err := s.transactionRepo.FindByUserID(ctx, userID, 10, 0)
	if err != nil {
		return []TransactionItem{} // Return empty array instead of nil
	}

	var result []TransactionItem
	for _, t := range transactions {
		result = append(result, TransactionItem{
			ID:              t.ID,
			CreatedAt:       t.CreatedAt,
			NetworkProvider: t.NetworkProvider,
			RechargeType:    t.RechargeType,
			Amount:          t.Amount / 100, // Convert kobo to naira
			PointsEarned:    int64(t.PointsEarned),
			Status:          t.Status,
		})
	}

	return result
}

// getUserPrizes gets user's prizes from spin results
func (s *UserService) getUserPrizes(ctx context.Context, userID uuid.UUID) []PrizeItem {
	spins, err := s.spinRepo.FindByUserID(ctx, userID, 100, 0)
	if err != nil {
		return []PrizeItem{} // Return empty array instead of nil
	}

	var result []PrizeItem
	for _, spin := range spins {
		result = append(result, PrizeItem{
			ID:          spin.ID,
			PrizeName:   spin.PrizeName,
			PrizeType:   spin.PrizeType,
			PrizeValue:  int64(spin.PrizeValue) / 100, // Convert kobo to naira
			Status:      spin.ClaimStatus,
			WonAt:       spin.CreatedAt,
			WonDate:     spin.CreatedAt.Format("2006-01-02 15:04:05"),
			ClaimedAt:   spin.ClaimedAt,
			ClaimDate:   func() *string { if spin.ClaimedAt != nil { s := spin.ClaimedAt.Format("2006-01-02 15:04:05"); return &s }; return nil }(),
		})
	}
	return result
}
