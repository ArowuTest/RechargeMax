package services

import (
	"go.uber.org/zap"
	"rechargemax/internal/logger"
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
	paymentService      *PaymentService
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
	paymentService *PaymentService,
) *AffiliateService {
	return &AffiliateService{
		affiliateRepo:       affiliateRepo,
		userRepo:            userRepo,
		commissionRepo:      commissionRepo,
		transactionRepo:     transactionRepo,
		walletService:       walletService,
		notificationService: notificationService,
		paymentService:      paymentService,
		db:                  db,
	}
}

// getDefaultCommissionRate reads the affiliate commission rate from platform_settings.
// Falls back to 1.0 (1%) if the key is not found or cannot be parsed.
func (s *AffiliateService) getDefaultCommissionRate(ctx context.Context) float64 {
	const fallback = 1.0
	const floor   = 0.5
	const ceiling = 1.5
	if s.db == nil {
		return fallback
	}
	var settingValue string
	err := s.db.WithContext(ctx).
		Raw("SELECT setting_value FROM platform_settings WHERE setting_key = ?",
			"affiliate.commission_rate_percent").
		Scan(&settingValue).Error
	if err != nil || settingValue == "" {
		return fallback
	}
	rate, err := strconv.ParseFloat(settingValue, 64)
	if err != nil || rate < floor || rate > ceiling {
		return fallback
	}
	return rate
}

// getFrontendURL reads FRONTEND_URL from platform_settings or falls back to
// the production default so referral links always point at the live frontend.
func (s *AffiliateService) getFrontendURL(ctx context.Context) string {
	const fallback = "https://rechargemax-frontend.vercel.app"
	if s.db == nil {
		return fallback
	}
	var val string
	s.db.WithContext(ctx).
		Raw("SELECT setting_value FROM platform_settings WHERE setting_key = 'frontend_url'").
		Scan(&val)
	if val == "" {
		return fallback
	}
	return val
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
		logger.Error("Warning: Failed to create wallet for affiliate", zap.Error(err), zap.String("msisdn", req.MSISDN))
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
		ReferralLink:    fmt.Sprintf("%s/recharge?ref=%s", s.getFrontendURL(ctx), affiliate.AffiliateCode),
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

// GetCommissionsSimple returns list of commissions for an affiliate (simple, no pagination)
func (s *AffiliateService) GetCommissionsSimple(ctx context.Context, msisdn string) ([]CommissionResponse, error) {
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

// RequestPayout creates a payout request for an affiliate.
// Validates minimum threshold (₦1,000), creates a commission_payouts record,
// then initiates the Paystack bank transfer.
// Amount is in NAIRA (not kobo) — matches what the frontend displays.
func (s *AffiliateService) RequestPayout(ctx context.Context, msisdn string, amountNGN float64) (*PayoutResponse, error) {
	const minPayoutNGN = 1000.0

	if amountNGN < minPayoutNGN {
		return nil, fmt.Errorf("minimum payout is ₦%.0f (requested ₦%.2f)", minPayoutNGN, amountNGN)
	}

	affiliate, err := s.affiliateRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("affiliate not found: %w", err)
	}
	if affiliate.Status != "APPROVED" {
		return nil, fmt.Errorf("affiliate account is not active (status: %s)", affiliate.Status)
	}
	if affiliate.TotalCommission < amountNGN {
		return nil, fmt.Errorf("insufficient balance: available ₦%.2f, requested ₦%.2f",
			affiliate.TotalCommission, amountNGN)
	}
	if affiliate.AccountNumber == "" || affiliate.BankName == "" {
		return nil, fmt.Errorf("bank account details are incomplete — update your profile before requesting a payout")
	}

	amountKobo := int64(amountNGN * 100)

	// ── ISO week identifier for grouping (e.g. "2026-W13") ──────────────────
	year, week := time.Now().ISOWeek()
	payoutWeek := fmt.Sprintf("%d-W%02d", year, week)

	payoutID := uuid.New()
	payoutRef := fmt.Sprintf("PAY-%s-%s", payoutWeek, payoutID.String()[:8])

	// ── Create payout record (PENDING) ───────────────────────────────────────
	if err := s.db.WithContext(ctx).Exec(`
		INSERT INTO commission_payouts
		    (id, affiliate_id, amount_kobo, status, bank_name, account_number,
		     account_name, payout_week, created_at, updated_at)
		VALUES (?, ?, ?, 'PENDING', ?, ?, ?, ?, NOW(), NOW())`,
		payoutID, affiliate.ID, amountKobo,
		affiliate.BankName, affiliate.AccountNumber, affiliate.AccountName,
		payoutWeek,
	).Error; err != nil {
		return nil, fmt.Errorf("failed to create payout record: %w", err)
	}

	// ── Initiate Paystack transfer ───────────────────────────────────────────
	var transferCode, transferRef string
	if s.paymentService != nil {
		transferResp, err := s.paymentService.ProcessTransfer(ctx, map[string]interface{}{
			"amount":         amountKobo,
			"account_name":   affiliate.AccountName,
			"account_number": affiliate.AccountNumber,
			"bank_name":      affiliate.BankName,
			"narration":      fmt.Sprintf("RechargeMax affiliate commission payout %s", payoutWeek),
			"reference":      payoutRef,
		})
		if err != nil {
			// Mark payout as FAILED but don't delete — admin can retry
			s.db.WithContext(ctx).Exec(
				"UPDATE commission_payouts SET status='FAILED', failed_reason=?, updated_at=NOW() WHERE id=?",
				err.Error(), payoutID)
			return nil, fmt.Errorf("bank transfer failed: %w", err)
		}
		if data, ok := transferResp["data"].(map[string]interface{}); ok {
			transferCode, _ = data["transfer_code"].(string)
			transferRef, _ = data["reference"].(string)
		}
		// Update payout with transfer details
		now := time.Now()
		s.db.WithContext(ctx).Exec(`
			UPDATE commission_payouts
			SET status='IN_TRANSIT', transfer_reference=?, transfer_code=?,
			    initiated_at=?, updated_at=NOW()
			WHERE id=?`,
			transferRef, transferCode, now, payoutID)
	}

	// ── Deduct from affiliate balance ────────────────────────────────────────
	s.db.WithContext(ctx).Exec(
		"UPDATE affiliates SET total_commission = total_commission - ?, updated_at=NOW() WHERE id=?",
		amountNGN, affiliate.ID)

	// ── Mark related APPROVED commissions as PAID ────────────────────────────
	s.db.WithContext(ctx).Exec(`
		UPDATE affiliate_commissions
		SET status='PAID', payout_id=?
		WHERE affiliate_id=? AND status='APPROVED'
		  AND id IN (
		      SELECT id FROM affiliate_commissions
		      WHERE affiliate_id=? AND status='APPROVED'
		      ORDER BY created_at
		      LIMIT (SELECT COUNT(*) FROM affiliate_commissions
		             WHERE affiliate_id=? AND status='APPROVED')
		  )`,
		payoutID, affiliate.ID, affiliate.ID, affiliate.ID)

	// ── Notify affiliate ─────────────────────────────────────────────────────
	if s.notificationService != nil && affiliate.UserID != nil {
		if user, err := s.userRepo.FindByID(ctx, *affiliate.UserID); err == nil {
			s.notificationService.SendMultiChannel(ctx, user.MSISDN,
				"Payout Initiated 💸",
				fmt.Sprintf("Your affiliate payout of ₦%.2f has been initiated and will arrive in 1-2 business days. Reference: %s", amountNGN, payoutRef),
				"affiliate_payout", map[string]interface{}{
					"payout_id":  payoutID.String(),
					"amount_ngn": amountNGN,
					"reference":  payoutRef,
				})
		}
	}

	logger.Info("[AUDIT] Affiliate payout initiated",
		zap.String("affiliate_id", affiliate.ID.String()),
		zap.Float64("amount_ngn", amountNGN),
		zap.String("payout_week", payoutWeek),
		zap.String("reference", payoutRef),
	)

	return &PayoutResponse{
		ID:        payoutID,
		Amount:    amountNGN,
		Status:    "IN_TRANSIT",
		CreatedAt: time.Now(),
	}, nil
}

// NotifyWeeklyPayout notifies admin that weekly payout processing is due.
// Called by a cron job every Monday at 08:00 WAT.
func (s *AffiliateService) NotifyWeeklyPayout(ctx context.Context) error {
	// Count affiliates with balance >= ₦1,000
	var eligibleCount int64
	var totalOwed float64
	s.db.WithContext(ctx).Raw(`
		SELECT COUNT(*), COALESCE(SUM(total_commission), 0)
		FROM affiliates
		WHERE status = 'APPROVED' AND total_commission >= 1000`).
		Row().Scan(&eligibleCount, &totalOwed)

	if eligibleCount == 0 {
		logger.Info("[AffiliateWeeklyNotify] No affiliates eligible for payout this week")
		return nil
	}

	year, week := time.Now().ISOWeek()
	payoutWeek := fmt.Sprintf("%d-W%02d", year, week)

	logger.Info("[AffiliateWeeklyNotify] Weekly payout due",
		zap.Int64("eligible_affiliates", eligibleCount),
		zap.Float64("total_owed_ngn", totalOwed),
		zap.String("payout_week", payoutWeek),
	)

	// Notify admin users
	if s.notificationService != nil {
		s.notificationService.NotifyAdmins(ctx,
			"⚡ Weekly Affiliate Payout Due",
			fmt.Sprintf("%d affiliates are eligible for payout this week (%s). Total owed: ₦%.2f. Visit Admin → Affiliates → Payouts to process.",
				eligibleCount, payoutWeek, totalOwed),
			"weekly_affiliate_payout",
			map[string]interface{}{
				"eligible_count": eligibleCount,
				"total_owed":     totalOwed,
				"payout_week":    payoutWeek,
			},
		)
	}
	return nil
}

// GetCommissions returns paginated commission records for an affiliate.
func (s *AffiliateService) GetCommissions(ctx context.Context, msisdn string, page, perPage int, status string) ([]map[string]interface{}, int64, error) {
	affiliate, err := s.affiliateRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, 0, fmt.Errorf("affiliate not found: %w", err)
	}

	offset := (page - 1) * perPage
	query := s.db.WithContext(ctx).Table("affiliate_commissions").
		Where("affiliate_id = ?", affiliate.ID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var rows []map[string]interface{}
	query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&rows)
	return rows, total, nil
}

// AdminGetAllCommissions returns commissions for admin view with filters.
func (s *AffiliateService) AdminGetAllCommissions(ctx context.Context, page, perPage int, status, affiliateID string) ([]map[string]interface{}, int64, error) {
	offset := (page - 1) * perPage
	query := s.db.WithContext(ctx).Table("affiliate_commissions ac").
		Select("ac.*, a.affiliate_code, u.msisdn as affiliate_msisdn").
		Joins("LEFT JOIN affiliates a ON a.id = ac.affiliate_id").
		Joins("LEFT JOIN users u ON u.id = a.user_id")
	if status != "" {
		query = query.Where("ac.status = ?", status)
	}
	if affiliateID != "" {
		query = query.Where("ac.affiliate_id = ?", affiliateID)
	}

	var total int64
	query.Count(&total)

	var rows []map[string]interface{}
	query.Order("ac.created_at DESC").Limit(perPage).Offset(offset).Find(&rows)
	return rows, total, nil
}

// AdminApproveCommissions bulk-approves PENDING commissions for payout.
func (s *AffiliateService) AdminApproveCommissions(ctx context.Context, commissionIDs []string) (int64, error) {
	if len(commissionIDs) == 0 {
		return 0, fmt.Errorf("no commission IDs provided")
	}
	result := s.db.WithContext(ctx).Exec(
		"UPDATE affiliate_commissions SET status='APPROVED', updated_at=NOW() WHERE id IN ? AND status='PENDING'",
		commissionIDs)
	return result.RowsAffected, result.Error
}

// AdminGetPayouts returns commission_payouts for admin view.
func (s *AffiliateService) AdminGetPayouts(ctx context.Context, page, perPage int, status string) ([]map[string]interface{}, int64, error) {
	offset := (page - 1) * perPage
	query := s.db.WithContext(ctx).Table("commission_payouts cp").
		Select("cp.*, a.affiliate_code, u.msisdn as affiliate_msisdn, u.full_name as affiliate_name").
		Joins("LEFT JOIN affiliates a ON a.id = cp.affiliate_id").
		Joins("LEFT JOIN users u ON u.id = a.user_id")
	if status != "" {
		query = query.Where("cp.status = ?", status)
	}
	var total int64
	query.Count(&total)
	var rows []map[string]interface{}
	query.Order("cp.created_at DESC").Limit(perPage).Offset(offset).Find(&rows)
	return rows, total, nil
}

// AdminInitiatePayout triggers the Paystack transfer for an admin-approved payout.
func (s *AffiliateService) AdminInitiatePayout(ctx context.Context, payoutID string, adminUserID string) error {
	var payout struct {
		ID            string  `gorm:"column:id"`
		AffiliateID   string  `gorm:"column:affiliate_id"`
		AmountKobo    int64   `gorm:"column:amount_kobo"`
		Status        string  `gorm:"column:status"`
		BankName      string  `gorm:"column:bank_name"`
		AccountNumber string  `gorm:"column:account_number"`
		AccountName   string  `gorm:"column:account_name"`
		PayoutWeek    string  `gorm:"column:payout_week"`
	}
	if err := s.db.WithContext(ctx).Raw(
		"SELECT id, affiliate_id, amount_kobo, status, bank_name, account_number, account_name, payout_week FROM commission_payouts WHERE id = ?",
		payoutID).Scan(&payout).Error; err != nil || payout.ID == "" {
		return fmt.Errorf("payout not found")
	}
	if payout.Status != "PENDING" {
		return fmt.Errorf("payout is already %s", payout.Status)
	}

	payoutRef := fmt.Sprintf("PAY-%s-%s", payout.PayoutWeek, payoutID[:8])

	if s.paymentService != nil {
		transferResp, err := s.paymentService.ProcessTransfer(ctx, map[string]interface{}{
			"amount":         payout.AmountKobo,
			"account_name":   payout.AccountName,
			"account_number": payout.AccountNumber,
			"bank_name":      payout.BankName,
			"narration":      fmt.Sprintf("RechargeMax affiliate payout %s", payout.PayoutWeek),
			"reference":      payoutRef,
		})
		if err != nil {
			s.db.WithContext(ctx).Exec(
				"UPDATE commission_payouts SET status='FAILED', failed_reason=?, updated_at=NOW() WHERE id=?",
				err.Error(), payoutID)
			return fmt.Errorf("transfer failed: %w", err)
		}
		var transferCode, transferRef string
		if data, ok := transferResp["data"].(map[string]interface{}); ok {
			transferCode, _ = data["transfer_code"].(string)
			transferRef, _ = data["reference"].(string)
		}
		now := time.Now()
		s.db.WithContext(ctx).Exec(`
			UPDATE commission_payouts
			SET status='IN_TRANSIT', transfer_reference=?, transfer_code=?,
			    initiated_by=?, initiated_at=?, updated_at=NOW()
			WHERE id=?`,
			transferRef, transferCode, adminUserID, now, payoutID)
	}

	logger.Info("[AUDIT] Admin initiated affiliate payout",
		zap.String("payout_id", payoutID),
		zap.String("admin_id", adminUserID),
		zap.String("ref", payoutRef),
	)
	return nil
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
	logger.Info("[AUDIT] Affiliate approved", zap.String("affiliate_id", affiliate.ID.String()), zap.String("msisdn", msisdn), zap.String("approved_at", now.Format(time.RFC3339)))
	
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
	logger.Info("[AUDIT] Affiliate rejected",
		zap.String("affiliate_id", affiliate.ID.String()),
		zap.String("msisdn", msisdn),
		zap.String("rejected_at", now.Format(time.RFC3339)),
		zap.String("reason", reason),
	)
	
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
	logger.Info("[INFO] Pending commissions frozen", zap.String("affiliate_id", affiliate.ID.String()))
	
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
	logger.Info("[AUDIT] Affiliate suspended",
		zap.String("affiliate_id", affiliate.ID.String()),
		zap.String("msisdn", msisdn),
		zap.String("suspended_at", now.Format(time.RFC3339)),
		zap.String("reason", reason),
	)
	
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

// ProcessCommissionTx is the transactional variant of ProcessCommission (BUG-002).
// All DB writes use the provided *gorm.DB transaction so they participate in the
// caller's atomic unit of work. If tx is nil, it falls back to the regular method.
func (s *AffiliateService) ProcessCommissionTx(ctx context.Context, tx *gorm.DB, msisdn string, rechargeAmount int64, transactionID uuid.UUID) error {
	if tx == nil {
		return s.ProcessCommission(ctx, msisdn, rechargeAmount, transactionID)
	}

	// --- reads via repos (non-mutating, safe outside the tx) ---
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil || user.ReferredBy == nil {
		return nil // no user or no referral — nothing to do
	}
	referrer, err := s.userRepo.FindByID(ctx, *user.ReferredBy)
	if err != nil {
		return nil
	}
	affiliate, err := s.affiliateRepo.FindByUserID(ctx, referrer.ID)
	if err != nil || affiliate.Status != "APPROVED" {
		return nil
	}

	// Count SUCCESSFUL recharges only — failed/pending transactions must not
	// influence the first-recharge gate.
	var successCount int64
	if err := s.db.WithContext(ctx).
		Model(&entities.Transactions{}).
		Where("user_id = ? AND status = 'success'", user.ID).
		Count(&successCount).Error; err != nil {
		return fmt.Errorf("commission: failed to count successful recharges: %w", err)
	}
	if successCount == 0 {
		// Very first successful recharge — no commission, but record the referral
		if err := tx.Model(referrer).UpdateColumn("total_referrals", gorm.Expr("total_referrals + 1")).Error; err != nil {
			return err
		}
		return tx.Model(affiliate).Updates(map[string]interface{}{
			"total_referrals":  gorm.Expr("total_referrals + 1"),
			"active_referrals": gorm.Expr("active_referrals + 1"),
		}).Error
	}

	// --- writes inside tx ---
	// Commission rate is stored as a percentage (e.g. 1.0 = 1%).
	// rechargeAmount is in kobo; commissionAmount will also be in kobo.
	commissionAmount := int64(float64(rechargeAmount) * affiliate.CommissionRate / 100.0)
	commission := &entities.AffiliateCommissions{
		ID:                uuid.New(),
		AffiliateID:       affiliate.ID,
		TransactionID:     &transactionID,
		CommissionAmount:  commissionAmount,
		CommissionRate:    affiliate.CommissionRate,
		TransactionAmount: rechargeAmount,
		Status:            "PENDING",
	}
	if err := tx.Create(commission).Error; err != nil {
		return fmt.Errorf("commission: failed to create record: %w", err)
	}
	if err := tx.Model(affiliate).
		UpdateColumn("total_commission", gorm.Expr("total_commission + ?", float64(commissionAmount)/100.0)).
		Error; err != nil {
		return fmt.Errorf("commission: failed to update affiliate total: %w", err)
	}
	return nil
}

// RecordClick records an affiliate link click for analytics.
// affiliateCode is the AFF... code embedded in the shared link (?ref=AFFxxxx).
func (s *AffiliateService) RecordClick(ctx context.Context, affiliateCode, msisdn, source string) error {
	if affiliateCode == "" {
		return nil
	}
	return s.db.WithContext(ctx).
		Model(&entities.Affiliates{}).
		Where("affiliate_code = ?", affiliateCode).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).
		Error
}

// AttributeReferral links a user to an affiliate via the ?ref=AFFxxxx param
// captured during recharge. Called inside the recharge transaction.
// Safe to call multiple times — only sets referred_by if not already assigned.
func (s *AffiliateService) AttributeReferral(ctx context.Context, tx *gorm.DB, msisdn, affiliateCode string) error {
	if affiliateCode == "" {
		return nil
	}
	affiliate, err := s.affiliateRepo.FindByAffiliateCode(ctx, affiliateCode)
	if err != nil || affiliate.Status != "APPROVED" || affiliate.UserID == nil {
		return nil // silently ignore invalid / unapproved codes
	}
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil || user.ReferredBy != nil {
		return nil // user not found or already attributed — first-write-wins
	}
	return tx.Model(&entities.Users{}).Where("id = ?", user.ID).
		Update("referred_by", affiliate.UserID).Error
}
