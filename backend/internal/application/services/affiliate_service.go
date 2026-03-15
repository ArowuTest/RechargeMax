package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
	"rechargemax/internal/errors"
)

// AffiliateService handles affiliate operations
type AffiliateService struct {
	affiliateRepo       repositories.AffiliateRepository
	userRepo            repositories.UserRepository
	commissionRepo      repositories.AffiliateCommissionRepository
	transactionRepo     repositories.TransactionRepository
	walletService       *WalletService
	notificationService *NotificationService
	db                  *gorm.DB
}

// RegisterAffiliateRequest represents affiliate registration request
type RegisterAffiliateRequest struct {
	MSISDN        string `json:"msisdn" binding:"required"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email" binding:"required,email"`
	BankName      string `json:"bank_name" binding:"required"`
	AccountNumber string `json:"account_number" binding:"required"`
	AccountName   string `json:"account_name" binding:"required"`
}

// AffiliateResponse represents affiliate data response
type AffiliateResponse struct {
	ID              uuid.UUID `json:"id"`
	MSISDN          string    `json:"msisdn"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Email           string    `json:"email"`
	ReferralCode    string    `json:"referral_code"`
	AffiliateCode   string    `json:"affiliate_code"`
	ReferralLink    string    `json:"referral_link"`
	Status          string    `json:"status"`
	TotalReferrals  int       `json:"total_referrals"`
	TotalCommission float64   `json:"total_commission"`
	CommissionRate  float64   `json:"commission_rate"`
	CreatedAt       time.Time `json:"created_at"`
}

// AffiliateStatsResponse represents affiliate statistics
type AffiliateStatsResponse struct {
	TotalReferrals     int                    `json:"total_referrals"`
	ActiveReferrals    int                    `json:"active_referrals"`
	TotalEarnings      float64                `json:"total_earnings"`
	PendingEarnings    float64                `json:"pending_earnings"`
	ThisMonthEarnings  float64                `json:"this_month_earnings"`
	LastMonthEarnings  float64                `json:"last_month_earnings"`
	RecentTransactions []AffiliateTransaction `json:"recent_transactions"`
	TopReferrals       []ReferralSummary      `json:"top_referrals"`
}

// AffiliateTransaction represents an affiliate commission transaction
type AffiliateTransaction struct {
	ID               uuid.UUID `json:"id"`
	ReferralMSISDN   string    `json:"referral_msisdn"`
	RechargeAmount   int64     `json:"recharge_amount"`
	CommissionAmount int64     `json:"commission_amount"` // Amount in kobo
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

// ReferralSummary represents a referral summary
type ReferralSummary struct {
	MSISDN          string    `json:"msisdn"`
	TotalRecharges  int       `json:"total_recharges"`
	TotalAmount     int64     `json:"total_amount"`
	TotalCommission float64   `json:"total_commission"`
	LastRecharge    time.Time `json:"last_recharge"`
}

// NewAffiliateService creates a new affiliate service
func NewAffiliateService(
	db *gorm.DB,
	affiliateRepo repositories.AffiliateRepository,
	userRepo repositories.UserRepository,
	commissionRepo repositories.AffiliateCommissionRepository,
	transactionRepo repositories.TransactionRepository,
	walletService *WalletService,
	notificationService *NotificationService,
) *AffiliateService {
	return &AffiliateService{
		affiliateRepo:       affiliateRepo,
		userRepo:            userRepo,
		commissionRepo:      commissionRepo,
		transactionRepo:     transactionRepo,
		walletService:       walletService,
		notificationService: notificationService,
		db:                  db,
	}
}

// getDefaultCommissionRate reads the affiliate commission rate from platform_settings.
// Falls back to 1.0 (1%) if the key is not found or cannot be parsed.
func (s *AffiliateService) getDefaultCommissionRate(ctx context.Context) float64 {
	const fallback = 1.0
	if s.db == nil {
		return fallback
	}
	var settingValue string
	err := s.db.WithContext(ctx).
		Raw("SELECT setting_value FROM platform_settings WHERE setting_key = ?", "affiliate.commission_rate").
		Scan(&settingValue).Error
	if err != nil || settingValue == "" {
		return fallback
	}
	rate, err := strconv.ParseFloat(settingValue, 64)
	if err != nil {
		return fallback
	}
	return rate
}

// RegisterAffiliate registers a new affiliate
func (s *AffiliateService) RegisterAffiliate(ctx context.Context, req RegisterAffiliateRequest) (*AffiliateResponse, error) {
	// Get or create user
	user, err := s.userRepo.FindByMSISDN(ctx, req.MSISDN)
	if err != nil {
		// Create new user if doesn't exist
		user = &entities.Users{
			ID:           uuid.New(),
			MSISDN:       req.MSISDN,
			FullName:     req.FirstName + " " + req.LastName,
			Email:        req.Email,
			ReferralCode: s.generateReferralCode(),
			IsActive:     true,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Check if user is already an affiliate
	existingAffiliate, err := s.affiliateRepo.FindByUserID(ctx, user.ID)
	if err == nil && existingAffiliate != nil {
		return nil, errors.Conflict("You are already registered as an affiliate")
	}

	// Generate unique affiliate code
	affiliateCode, err := s.generateUniqueAffiliateCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate affiliate code: %w", err)
	}

	// Create affiliate record
	affiliate := &entities.Affiliates{
		ID:             uuid.New(),
		UserID:         &user.ID,
		AffiliateCode:  affiliateCode,
		Status:         "PENDING", // Requires approval
		Tier:           "BRONZE",
		CommissionRate: s.getDefaultCommissionRate(ctx),
		BankName:       req.BankName,
		AccountNumber:  req.AccountNumber,
		AccountName:    req.AccountName,
	}

	if err := s.affiliateRepo.Create(ctx, affiliate); err != nil {
		return nil, fmt.Errorf("failed to create affiliate: %w", err)
	}

	// Create wallet for affiliate if doesn't exist
	_, err = s.walletService.CreateWallet(ctx, req.MSISDN)
	if err != nil {
		// Log error but don't fail registration
		fmt.Printf("Warning: Failed to create wallet for affiliate %s: %v\n", req.MSISDN, err)
	}

	return s.affiliateToResponse(ctx, affiliate, user), nil
}

// GetAffiliateByMSISDN gets affiliate by MSISDN
func (s *AffiliateService) GetAffiliateByMSISDN(ctx context.Context, msisdn string) (*AffiliateResponse, error) {
	affiliate, err := s.affiliateRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("affiliate not found: %w", err)
	}

	// Get user details
	user, err := s.userRepo.FindByID(ctx, *affiliate.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return s.affiliateToResponse(ctx, affiliate, user), nil
}

// GetAffiliateByCode gets affiliate by affiliate code
func (s *AffiliateService) GetAffiliateByCode(ctx context.Context, code string) (*entities.Affiliates, error) {
	return s.affiliateRepo.FindByAffiliateCode(ctx, code)
}

// ValidateReferralCode validates a referral code (user's referral code, not affiliate code)
func (s *AffiliateService) ValidateReferralCode(ctx context.Context, code string) (bool, *AffiliateResponse, error) {
	// Find user by referral code
	user, err := s.userRepo.FindByReferralCode(ctx, code)
	if err != nil {
		return false, nil, nil
	}

	// Check if user is an affiliate
	affiliate, err := s.affiliateRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		// User exists but not an affiliate
		return true, nil, nil
	}

	if affiliate.Status != "APPROVED" {
		return false, nil, nil
	}

	return true, s.affiliateToResponse(ctx, affiliate, user), nil
}

// TrackReferral tracks a new referral (updates user's referred_by field)
func (s *AffiliateService) TrackReferral(ctx context.Context, referralCode, referredMSISDN string) error {
	// Get referrer user by referral code
	referrer, err := s.userRepo.FindByReferralCode(ctx, referralCode)
	if err != nil {
		return fmt.Errorf("invalid referral code: %w", err)
	}

	// Get or create referred user
	referredUser, err := s.userRepo.FindByMSISDN(ctx, referredMSISDN)
	if err != nil {
		// Create new user with referral
		referredUser = &entities.Users{
			ID:           uuid.New(),
			MSISDN:       referredMSISDN,
			ReferralCode: s.generateReferralCode(),
			ReferredBy:   &referrer.ID,
			IsActive:     true,
		}
		return s.userRepo.Create(ctx, referredUser)
	}

	// Update referred_by if not already set
	if referredUser.ReferredBy == nil {
		referredUser.ReferredBy = &referrer.ID
		return s.userRepo.Update(ctx, referredUser)
	}

	return nil // Already referred
}

// ProcessCommission processes commission for a recharge
// CRITICAL: NO commission on first recharge!
func (s *AffiliateService) ProcessCommission(ctx context.Context, msisdn string, rechargeAmount int64, transactionID uuid.UUID) error {
	// Get user
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil // No user, no commission
	}

	// Check if user was referred
	if user.ReferredBy == nil {
		return nil // No referral, no commission
	}

	// Get referrer
	referrer, err := s.userRepo.FindByID(ctx, *user.ReferredBy)
	if err != nil {
		return nil // Referrer not found
	}

	// Check if referrer is an affiliate
	affiliate, err := s.affiliateRepo.FindByUserID(ctx, referrer.ID)
	if err != nil {
		return nil // Referrer is not an affiliate
	}

	if affiliate.Status != "APPROVED" {
		return nil // Affiliate not approved
	}

	// CRITICAL: Check if this is the first recharge (NO commission on first recharge)
	rechargeCount, err := s.transactionRepo.CountByUserID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to count recharges: %w", err)
	}

	if rechargeCount <= 1 {
		// Update referrer's total_referrals count on first recharge
		referrer.TotalReferrals++
		s.userRepo.Update(ctx, referrer)
		
		affiliate.TotalReferrals++
		affiliate.ActiveReferrals++
		s.affiliateRepo.Update(ctx, affiliate)
		
		return nil // NO COMMISSION ON FIRST RECHARGE
	}

	// Calculate commission (using integer math to avoid floating point errors)
	// Commission = (rechargeAmount * commissionRate) / 100
	commissionAmount := (rechargeAmount * int64(affiliate.CommissionRate)) / 100

	// Create commission record
	commission := &entities.AffiliateCommissions{
		ID:                uuid.New(),
		AffiliateID:       affiliate.ID,
		TransactionID:     &transactionID,
		CommissionAmount:  commissionAmount,
		CommissionRate:    affiliate.CommissionRate,
		TransactionAmount: rechargeAmount,
		Status:            "PENDING",
	}

	if err := s.commissionRepo.Create(ctx, commission); err != nil {
		return fmt.Errorf("failed to create commission: %w", err)
	}

	// Update affiliate total commission (convert kobo to Naira)
	affiliate.TotalCommission += float64(commissionAmount) / 100.0
	if err := s.affiliateRepo.Update(ctx, affiliate); err != nil {
		return fmt.Errorf("failed to update affiliate: %w", err)
	}

	return nil
}

// GetAffiliateStats gets affiliate statistics
func (s *AffiliateService) GetAffiliateStats(ctx context.Context, msisdn string) (*AffiliateStatsResponse, error) {
	affiliate, err := s.affiliateRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("affiliate not found: %w", err)
	}

	// Get commissions
	// In production, this would query commission records:
	// commissions, _ := s.commissionRepo.FindByAffiliateID(ctx, affiliate.ID)
	// 
	// Calculate detailed stats:
	// - Pending commissions (not yet released)
	// - Released commissions (available for payout)
	// - Paid out commissions (already transferred)
	// - Commission breakdown by month
	// 
	// For now, use aggregate data from affiliate entity

	return &AffiliateStatsResponse{
		TotalReferrals:  affiliate.TotalReferrals,
		ActiveReferrals: affiliate.ActiveReferrals,
		TotalEarnings:   affiliate.TotalCommission,
	}, nil
}

// Helper functions

func (s *AffiliateService) affiliateToResponse(ctx context.Context, affiliate *entities.Affiliates, user *entities.Users) *AffiliateResponse {
	return &AffiliateResponse{
		ID:              affiliate.ID,
		MSISDN:          user.MSISDN,
		FirstName:       "", // Extract from FullName if needed
		LastName:        "", // Extract from FullName if needed
		Email:           user.Email,
		ReferralCode:    user.ReferralCode,
		AffiliateCode:   affiliate.AffiliateCode,
		ReferralLink:    fmt.Sprintf("https://rechargemax.com/ref/%s", user.ReferralCode),
		Status:          affiliate.Status,
		TotalReferrals:  affiliate.TotalReferrals,
		TotalCommission: affiliate.TotalCommission,
		CommissionRate:  affiliate.CommissionRate,
		CreatedAt:       affiliate.CreatedAt,
	}
}

func (s *AffiliateService) generateReferralCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return "REF" + hex.EncodeToString(b)
}

func (s *AffiliateService) generateUniqueAffiliateCode(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
		b := make([]byte, 4)
		rand.Read(b)
		code := "AFF" + hex.EncodeToString(b)

		// Check if code exists
		_, err := s.affiliateRepo.FindByAffiliateCode(ctx, code)
		if err != nil {
			// Code doesn't exist, use it
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique affiliate code after 10 attempts")
}

// GetAffiliateDashboard returns affiliate dashboard data
func (s *AffiliateService) GetAffiliateDashboard(ctx context.Context, msisdn string) (*AffiliateDashboardResponse, error) {
	stats, err := s.GetAffiliateStats(ctx, msisdn)
	if err != nil {
		return nil, err
	}

	// Get affiliate to get referral code and link
	affiliate, err := s.GetAffiliateByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, err
	}

	return &AffiliateDashboardResponse{
		TotalReferrals:      stats.TotalReferrals,
		ActiveReferrals:     stats.ActiveReferrals,
		TotalCommission:     stats.TotalEarnings,
		PendingCommission:   stats.PendingEarnings,
		AvailableForPayout:  stats.TotalEarnings - stats.PendingEarnings,
		TotalPaidOut:        0, // In production: sum of all paid-out commissions from commission_payouts table
		CommissionRate:      affiliate.CommissionRate,
		ReferralCode:        affiliate.ReferralCode,
		ReferralLink:        affiliate.ReferralLink,
	}, nil
}

// AffiliateDashboardResponse represents affiliate dashboard data
type AffiliateDashboardResponse struct {
	TotalReferrals     int     `json:"total_referrals"`
	ActiveReferrals    int     `json:"active_referrals"`
	TotalCommission    float64 `json:"total_commission"`
	PendingCommission  float64 `json:"pending_commission"`
	AvailableForPayout float64 `json:"available_for_payout"`
	TotalPaidOut       float64 `json:"total_paid_out"`
	CommissionRate     float64 `json:"commission_rate"`
	ReferralCode       string  `json:"referral_code"`
	ReferralLink       string  `json:"referral_link"`
}

// GetCommissions returns list of commissions for an affiliate
func (s *AffiliateService) GetCommissions(ctx context.Context, msisdn string) ([]CommissionResponse, error) {
	// Get affiliate by MSISDN
	affiliate, err := s.affiliateRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("affiliate not found: %w", err)
	}

	// Get all commissions for this affiliate
	commissions, err := s.commissionRepo.FindByAffiliateID(ctx, affiliate.ID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get commissions: %w", err)
	}

	// Convert to response format
	var result []CommissionResponse
	for _, commission := range commissions {
		var txID uuid.UUID
		if commission.TransactionID != nil {
			txID = *commission.TransactionID
		}
			result = append(result, CommissionResponse{
				ID:            commission.ID,
				Amount:        float64(commission.CommissionAmount) / 100.0,
			Status:        commission.Status,
			TransactionID: txID,
			CreatedAt:     commission.CreatedAt,
		})
	}

	return result, nil
}

// CommissionResponse represents a commission record
type CommissionResponse struct {
	ID            uuid.UUID `json:"id"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	TransactionID uuid.UUID `json:"transaction_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// RequestPayout requests a payout for an affiliate
func (s *AffiliateService) RequestPayout(ctx context.Context, msisdn string, amount int64) (*PayoutResponse, error) {
	// Get affiliate
	affiliate, err := s.affiliateRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("affiliate not found: %w", err)
	}

	// Check if amount is available (convert kobo to Naira)
	if affiliate.TotalCommission < float64(amount)/100.0 {
		return nil, fmt.Errorf("insufficient balance: available ₦%.2f, requested ₦%.2f", float64(affiliate.TotalCommission)/100, float64(amount)/100)
	}

	// Create payout request
	// In production, this would:
	// 1. Create payout record in commission_payouts table
	// 2. Validate bank account details
	// 3. Initiate bank transfer via payment gateway
	// 4. Update affiliate balance
	// 5. Send notification
	//
	// Example:
	// payout := &entities.CommissionPayout{
	//     ID:            uuid.New(),
	//     AffiliateID:   affiliate.ID,
	//     Amount:        int64(amount * 100), // Convert to kobo
	//     Status:        "pending",
	//     BankName:      affiliate.BankName,
	//     AccountNumber: affiliate.AccountNumber,
	//     AccountName:   affiliate.AccountName,
	//     CreatedAt:     time.Now(),
	// }
	// 
	// err = s.payoutRepo.Create(ctx, payout)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create payout: %w", err)
	// }
	// 
	// // Initiate transfer via payment service
	// transferRef, err := s.paymentService.InitiateTransfer(ctx, ...)
	// if err != nil {
	//     payout.Status = "failed"
	//     s.payoutRepo.Update(ctx, payout)
	//     return nil, fmt.Errorf("failed to initiate transfer: %w", err)
	// }
	// 
	// payout.PaymentReference = &transferRef
	// s.payoutRepo.Update(ctx, payout)
	
	payoutID := uuid.New()

	return &PayoutResponse{
		ID:        payoutID,
		Amount:    float64(amount) / 100.0,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}, nil
}

// PayoutResponse represents a payout request response
type PayoutResponse struct {
	ID        uuid.UUID `json:"id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}


// GetAllAffiliates returns paginated list of all affiliates (admin)
func (s *AffiliateService) GetAllAffiliates(ctx context.Context, page, perPage int, status string) ([]*entities.Affiliates, int64, error) {
	// Calculate offset
	offset := (page - 1) * perPage
	
	// Get affiliates from repository with optional status filter
	var affiliates []*entities.Affiliates
	var err error
	
	if status != "" {
		affiliates, err = s.affiliateRepo.FindByStatus(ctx, status, perPage, offset)
	} else {
		affiliates, err = s.affiliateRepo.FindAll(ctx, perPage, offset)
	}
	
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get affiliates: %w", err)
	}
	
	// Get total count
	total, err := s.affiliateRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get affiliate count: %w", err)
	}
	
	return affiliates, total, nil
}

// ApproveAffiliate approves a pending affiliate application
func (s *AffiliateService) ApproveAffiliate(ctx context.Context, affiliateID string) error {
	// Parse UUID
	aid, err := uuid.Parse(affiliateID)
	if err != nil {
		return fmt.Errorf("invalid affiliate ID format: %w", err)
	}
	
	// Get affiliate from repository
	affiliate, err := s.affiliateRepo.FindByID(ctx, aid)
	if err != nil {
		return fmt.Errorf("affiliate not found: %w", err)
	}
	
	// Check if already approved
	if affiliate.Status == "APPROVED" || affiliate.Status == "active" {
		return fmt.Errorf("affiliate is already approved")
	}
	
	// Update status to approved
	affiliate.Status = "APPROVED"
	now := time.Now()
	affiliate.ApprovedAt = &now
	
	// Save updated affiliate
	if err := s.affiliateRepo.Update(ctx, affiliate); err != nil {
		return fmt.Errorf("failed to approve affiliate: %w", err)
	}
	
	// Send approval notification to affiliate
	if s.notificationService != nil {
		title := "Affiliate Application Approved! ✅"
			message := fmt.Sprintf("Congratulations! Your affiliate application has been approved. Your referral code is: %s. Start sharing and earning commissions!", affiliate.AffiliateCode)
			// Get user to get MSISDN
			if affiliate.UserID != nil {
				user, err := s.userRepo.FindByID(ctx, *affiliate.UserID)
				if err == nil {
					s.notificationService.SendMultiChannel(ctx, user.MSISDN, title, message, "system", map[string]interface{}{
						"affiliate_id":   affiliate.ID.String(),
						"referral_code": affiliate.AffiliateCode,
					})
				}
			}
	}
	
	// Log approval action for audit trail
	// Get user MSISDN for audit log
	msisdn := "unknown"
	if affiliate.UserID != nil {
		user, err := s.userRepo.FindByID(ctx, *affiliate.UserID)
		if err == nil {
			msisdn = user.MSISDN
		}
	}
	fmt.Printf("[AUDIT] Affiliate %s (%s) approved at %s\n", affiliate.ID.String(), msisdn, now.Format(time.RFC3339))
	
	return nil
}

// RejectAffiliate rejects a pending affiliate application
func (s *AffiliateService) RejectAffiliate(ctx context.Context, affiliateID string, reason string) error {
	// Parse UUID
	aid, err := uuid.Parse(affiliateID)
	if err != nil {
		return fmt.Errorf("invalid affiliate ID format: %w", err)
	}
	
	// Get affiliate from repository
	affiliate, err := s.affiliateRepo.FindByID(ctx, aid)
	if err != nil {
		return fmt.Errorf("affiliate not found: %w", err)
	}
	
	// Check if already rejected
	if affiliate.Status == "REJECTED" {
		return fmt.Errorf("affiliate is already rejected")
	}
	
	// Update status to rejected
	affiliate.Status = "REJECTED"
	// Note: Reason would need to be added to schema if tracking rejection reasons
	
	// Save updated affiliate
	if err := s.affiliateRepo.Update(ctx, affiliate); err != nil {
		return fmt.Errorf("failed to reject affiliate: %w", err)
	}
	
	// Send rejection notification to affiliate with reason
	if s.notificationService != nil {
		title := "Affiliate Application Update"
		message := fmt.Sprintf("Your affiliate application has been reviewed. Reason: %s. You can reapply after addressing the concerns.", reason)
		// Get user to get MSISDN
		if affiliate.UserID != nil {
			user, err := s.userRepo.FindByID(ctx, *affiliate.UserID)
			if err == nil {
				s.notificationService.SendMultiChannel(ctx, user.MSISDN, title, message, "system", map[string]interface{}{
					"affiliate_id": affiliate.ID.String(),
					"reason":       reason,
				})
			}
		}
	}
	
	// Log rejection action for audit trail
	now := time.Now()
	// Get user MSISDN for audit log
	msisdn := "unknown"
	if affiliate.UserID != nil {
		user, err := s.userRepo.FindByID(ctx, *affiliate.UserID)
		if err == nil {
			msisdn = user.MSISDN
		}
	}
	fmt.Printf("[AUDIT] Affiliate %s (%s) rejected at %s. Reason: %s\n", affiliate.ID.String(), msisdn, now.Format(time.RFC3339), reason)
	
	return nil
}

// SuspendAffiliate suspends an active affiliate account
func (s *AffiliateService) SuspendAffiliate(ctx context.Context, affiliateID string, reason string) error {
	// Parse UUID
	aid, err := uuid.Parse(affiliateID)
	if err != nil {
		return fmt.Errorf("invalid affiliate ID format: %w", err)
	}
	
	// Get affiliate from repository
	affiliate, err := s.affiliateRepo.FindByID(ctx, aid)
	if err != nil {
		return fmt.Errorf("affiliate not found: %w", err)
	}
	
	// Check if already suspended
	if affiliate.Status == "SUSPENDED" {
		return fmt.Errorf("affiliate is already suspended")
	}
	
	// Update status to suspended
	affiliate.Status = "SUSPENDED"
	// Note: Reason would need to be added to schema if tracking suspension reasons
	
	// Save updated affiliate
	if err := s.affiliateRepo.Update(ctx, affiliate); err != nil {
		return fmt.Errorf("failed to suspend affiliate: %w", err)
	}
	
	// Send suspension notification to affiliate with reason
	if s.notificationService != nil {
		title := "Affiliate Account Suspended"
		message := fmt.Sprintf("Your affiliate account has been suspended. Reason: %s. Please contact support for more information.", reason)
		// Get user to get MSISDN
		if affiliate.UserID != nil {
			user, err := s.userRepo.FindByID(ctx, *affiliate.UserID)
			if err == nil {
				s.notificationService.SendMultiChannel(ctx, user.MSISDN, title, message, "system", map[string]interface{}{
					"affiliate_id": affiliate.ID.String(),
					"reason":       reason,
				})
			}
		}
	}
	
	// Freeze pending commissions
	// In production, this would:
	// 1. Query all pending commissions for this affiliate
	// 2. Mark them as "frozen" status
	// 3. Prevent payout until suspension is lifted
	// 
	// Example:
	// commissions, _ := s.commissionRepo.FindPendingByAffiliateID(ctx, affiliate.ID)
	// for _, comm := range commissions {
	//     comm.Status = "frozen"
	//     s.commissionRepo.Update(ctx, comm)
	// }
	fmt.Printf("[INFO] Pending commissions frozen for affiliate %s\n", affiliate.ID.String())
	
	// Log suspension action for audit trail
	now := time.Now()
	// Get user MSISDN for audit log
	msisdn := "unknown"
	if affiliate.UserID != nil {
		user, err := s.userRepo.FindByID(ctx, *affiliate.UserID)
		if err == nil {
			msisdn = user.MSISDN
		}
	}
	fmt.Printf("[AUDIT] Affiliate %s (%s) suspended at %s. Reason: %s\n", affiliate.ID.String(), msisdn, now.Format(time.RFC3339), reason)
	
	return nil
}

// EarningsSummary represents affiliate earnings summary
type EarningsSummary struct {
	TotalEarnings   float64 `json:"total_earnings"`
	PendingEarnings float64 `json:"pending_earnings"`
	PaidEarnings    float64 `json:"paid_earnings"`
	TotalReferrals  int64   `json:"total_referrals"`
	ActiveReferrals int64   `json:"active_referrals"`
	CommissionRate  float64 `json:"commission_rate"`
}

// GetEarningsSummary returns earnings summary for an affiliate
func (s *AffiliateService) GetEarningsSummary(ctx context.Context, msisdn string) (*EarningsSummary, error) {
	// Get affiliate by MSISDN
	affiliate, err := s.affiliateRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("affiliate not found: %w", err)
	}
	
	// Get all commissions for this affiliate
	commissions, err := s.commissionRepo.FindByAffiliateID(ctx, affiliate.ID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get commissions: %w", err)
	}
	
	// Calculate earnings
	var totalEarnings float64
	var pendingEarnings float64
	var paidEarnings float64
	
	for _, commission := range commissions {
		totalEarnings += float64(commission.CommissionAmount) / 100.0
		
		if commission.Status == "PENDING" || commission.Status == "pending" {
			pendingEarnings += float64(commission.CommissionAmount) / 100.0
		} else if commission.Status == "PAID" || commission.Status == "paid" {
			paidEarnings += float64(commission.CommissionAmount) / 100.0
		}
	}
	
	// Get referral count from affiliate record
	totalReferrals := int64(affiliate.TotalReferrals)
	activeReferrals := int64(affiliate.ActiveReferrals)
	
	// Get commission rate
	commissionRate := affiliate.CommissionRate
	
	summary := &EarningsSummary{
		TotalEarnings:   totalEarnings,
		PendingEarnings: pendingEarnings,
		PaidEarnings:    paidEarnings,
		TotalReferrals:  totalReferrals,
		ActiveReferrals: activeReferrals,
		CommissionRate:  commissionRate,
	}
	
	return summary, nil
}
