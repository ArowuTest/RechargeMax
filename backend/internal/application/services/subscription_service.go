package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
	"rechargemax/internal/errors"
	"rechargemax/internal/logger"
	"rechargemax/internal/utils"
)

// SubscriptionService handles subscription operations
type SubscriptionService struct {
	subscriptionRepo repositories.SubscriptionRepository
	userRepo         repositories.UserRepository
	paymentService   *PaymentService
	hlrService       *HLRService
	db               *gorm.DB
}

// CreateSubscriptionRequest represents subscription creation request.
// PaymentMethod is optional — defaults to "paystack" if not provided.
type CreateSubscriptionRequest struct {
	MSISDN        string `json:"msisdn"`
	Network       string `json:"network"`
	PaymentMethod string `json:"payment_method"`
}

// SubscriptionResponse represents subscription response
type SubscriptionResponse struct {
	ID            uuid.UUID  `json:"id"`
	MSISDN        string     `json:"msisdn"`
	Network       string     `json:"network"`
	Status        string     `json:"status"`
	PaymentMethod string     `json:"payment_method"`
	DailyAmount   int64      `json:"daily_amount"`
	NextBilling   time.Time  `json:"next_billing"`
	CreatedAt     time.Time  `json:"created_at"`
	CancelledAt   *time.Time `json:"cancelled_at,omitempty"`
	PaymentURL    string     `json:"payment_url,omitempty"`
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(
	subscriptionRepo repositories.SubscriptionRepository,
	userRepo repositories.UserRepository,
	paymentService *PaymentService,
	hlrService *HLRService,
	db *gorm.DB,
) *SubscriptionService {
	return &SubscriptionService{
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		paymentService:   paymentService,
		hlrService:       hlrService,
		db:               db,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CreateSubscription
// ─────────────────────────────────────────────────────────────────────────────

// CreateSubscription creates a ₦20/day subscription.
//
// MSISDN resolution order:
//  1. req.MSISDN set by handler from JWT token (authenticated user)
//  2. req.MSISDN supplied in request body (guest / unauthenticated user)
//
// The handler already copies the JWT msisdn into req.MSISDN before calling
// this function.  If it is still empty we return a clear validation error.
func (s *SubscriptionService) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*SubscriptionResponse, error) {
	// ── 1. Resolve & normalise MSISDN ────────────────────────────────────────
	if req.MSISDN == "" {
		return nil, errors.BadRequest("Phone number is required")
	}
	normalised, err := utils.NormalizeMSISDN(req.MSISDN)
	if err != nil {
		return nil, errors.BadRequest(fmt.Sprintf("Invalid phone number format: %s", req.MSISDN))
	}
	req.MSISDN = normalised

	// ── 2. Default payment method ─────────────────────────────────────────────
	if req.PaymentMethod == "" {
		req.PaymentMethod = "paystack"
	}

	// ── 3. Optional network detection (non-fatal) ─────────────────────────────
	networkHint := req.Network
	detectedNetwork := ""
	if n, e := s.hlrService.DetectNetwork(ctx, req.MSISDN, &networkHint); e == nil {
		detectedNetwork = n
	}
	if detectedNetwork == "" {
		detectedNetwork = "MTN" // safe default so DCB path works for MTN subscribers
	}

	// ── 4. Resolve user (optional — guest subscriptions are allowed) ──────────
	var userID *uuid.UUID
	var userEmail string
	user, err := s.userRepo.FindByMSISDN(ctx, req.MSISDN)
	if err == nil && user != nil {
		userID = &user.ID
		userEmail = fmt.Sprintf("%s@rechargemax.ng", req.MSISDN)
	}

	// ── 5. Check for existing active subscription (idempotency) ──────────────
	existing, lookupErr := s.subscriptionRepo.FindActiveByMSISDN(ctx, req.MSISDN)
	if lookupErr == nil && existing != nil {
		return nil, errors.Conflict("You already have an active subscription. Cancel the current one to re-subscribe.")
	}

	// ── 6. Resolve tier_id — required NOT NULL in daily_subscriptions entity ──
	// Use the first active spin tier as the FK value (Bronze by default).
	// This FK is informational only for subscriptions; it has no NOT NULL
	// constraint in the actual SQL schema (see 14_daily_subscriptions.sql).
	// We provide it anyway so the entity validates correctly.
	tierID := s.resolveDefaultTierID(ctx)

	// ── 7. Build subscription record ──────────────────────────────────────────
	now := time.Now()
	subscriptionCode := fmt.Sprintf("SUB_%s_%d", req.MSISDN[len(req.MSISDN)-4:], now.Unix())
	paymentMethod := req.PaymentMethod
	nextBilling := now.Add(24 * time.Hour)
	dailyAmount := int64(2000) // ₦20 in kobo
	drawEntries := 1
	pointsEarned := 1

	subscription := &entities.DailySubscription{
		ID:               uuid.New(),
		SubscriptionCode: subscriptionCode,
		UserID:           userID,
		MSISDN:           req.MSISDN,
		TierID:           tierID,
		BundleQuantity:   1,
		TotalEntries:     drawEntries,
		DailyAmount:      dailyAmount,
		Amount:           dailyAmount, // legacy compat column
		DrawEntriesEarned: &drawEntries,
		PointsEarned:     &pointsEarned,
		Status:           "pending",
		AutoRenew:        true,
		NextBillingDate:  nextBilling,
		PaymentMethod:    paymentMethod,
		SubscriptionDate: now,
		CustomerEmail:    ptrString(userEmail),
	}

	if err := s.subscriptionRepo.Create(ctx, subscription); err != nil {
		logger.Error("[SubscriptionService] Create failed", zap.String("msisdn", req.MSISDN), zap.Error(err))
		if strings.Contains(err.Error(), "23505") ||
			strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.Conflict("You already have a subscription for today")
		}
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// ── 8. Build response ─────────────────────────────────────────────────────
	response := &SubscriptionResponse{
		ID:            subscription.ID,
		MSISDN:        subscription.MSISDN,
		Network:       detectedNetwork,
		Status:        subscription.Status,
		PaymentMethod: paymentMethod,
		DailyAmount:   dailyAmount,
		NextBilling:   nextBilling,
		CreatedAt:     subscription.CreatedAt,
	}

	// ── 9. Payment initialisation ─────────────────────────────────────────────
	if paymentMethod == "dcb" && detectedNetwork == "MTN" {
		// Direct carrier billing — activate immediately
		subscription.Status = "active"
		_ = s.subscriptionRepo.Update(ctx, subscription)
		response.Status = "active"
		// Award points immediately on DCB
		_ = s.awardSubscriptionPoints(ctx, subscription)
	} else {
		// Paystack / card — generate payment link
		reference := fmt.Sprintf("SUB_%s_%d", subscription.ID.String()[:8], now.Unix())
		paymentReq := PaymentRequest{
			Amount:    2000,
			Email:     userEmail,
			Reference: reference,
			Metadata: map[string]interface{}{
				"msisdn":          req.MSISDN,
				"type":            "subscription",
				"subscription_id": subscription.ID.String(),
			},
		}
		payURL, err := s.paymentService.InitializePayment(ctx, paymentReq)
		if err != nil {
			// Non-fatal: subscription row is created, payment can be retried
			logger.Error("[SubscriptionService] Payment init failed", zap.Error(err))
		} else {
			response.PaymentURL = payURL
			// Store payment reference on the subscription row
			subscription.PaymentReference = &reference
			_ = s.subscriptionRepo.Update(ctx, subscription)
		}
	}

	logger.Info("[SubscriptionService] Subscription created",
		zap.String("msisdn", req.MSISDN),
		zap.String("id", subscription.ID.String()),
		zap.String("status", subscription.Status),
	)
	return response, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetSubscription
// ─────────────────────────────────────────────────────────────────────────────

func (s *SubscriptionService) GetSubscription(ctx context.Context, msisdn string) (*SubscriptionResponse, error) {
	normalised, err := utils.NormalizeMSISDN(msisdn)
	if err == nil {
		msisdn = normalised
	}

	active, err := s.subscriptionRepo.FindActiveByMSISDN(ctx, msisdn)
	if err == nil && active != nil {
		return s.toResponse(active), nil
	}

	// Fall back to user-based lookup
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, errors.NotFound("No subscription found for this number")
	}
	subs, err := s.subscriptionRepo.FindByUserID(ctx, user.ID)
	if err != nil || len(subs) == 0 {
		return nil, errors.NotFound("No subscription found")
	}
	latest := subs[0]
	for _, sub := range subs {
		if sub.Status == "active" {
			latest = sub
			break
		}
	}
	return s.toResponse(latest), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// CancelSubscription
// ─────────────────────────────────────────────────────────────────────────────

func (s *SubscriptionService) CancelSubscription(ctx context.Context, msisdn string) error {
	normalised, err := utils.NormalizeMSISDN(msisdn)
	if err == nil {
		msisdn = normalised
	}

	active, err := s.subscriptionRepo.FindActiveByMSISDN(ctx, msisdn)
	if err != nil || active == nil {
		// Fall back to user-based lookup
		user, userErr := s.userRepo.FindByMSISDN(ctx, msisdn)
		if userErr != nil {
			return errors.NotFound("No active subscription found")
		}
		subs, subsErr := s.subscriptionRepo.FindByUserID(ctx, user.ID)
		if subsErr != nil {
			return errors.NotFound("No active subscription found")
		}
		for _, sub := range subs {
			if sub.Status == "active" {
				active = sub
				break
			}
		}
	}

	if active == nil {
		return errors.NotFound("No active subscription found to cancel")
	}

	now := time.Now()
	active.Status = "cancelled"
	active.CancelledAt = &now
	active.CancellationReason = "user_requested"

	if err := s.subscriptionRepo.Update(ctx, active); err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	logger.Info("[SubscriptionService] Subscription cancelled",
		zap.String("msisdn", msisdn),
		zap.String("id", active.ID.String()),
	)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetSubscriptionHistory
// ─────────────────────────────────────────────────────────────────────────────

func (s *SubscriptionService) GetSubscriptionHistory(ctx context.Context, msisdn string) ([]SubscriptionResponse, error) {
	normalised, err := utils.NormalizeMSISDN(msisdn)
	if err == nil {
		msisdn = normalised
	}

	// Try MSISDN-direct lookup first (works for both guests and registered users)
	var allSubs []*entities.DailySubscription
	if dbErr := s.db.WithContext(ctx).
		Where("msisdn = ?", msisdn).
		Order("subscription_date DESC").
		Find(&allSubs).Error; dbErr != nil || len(allSubs) == 0 {
		// Fall back to user_id lookup
		user, userErr := s.userRepo.FindByMSISDN(ctx, msisdn)
		if userErr != nil {
			return []SubscriptionResponse{}, nil // empty, not an error
		}
		allSubs, _ = s.subscriptionRepo.FindByUserID(ctx, user.ID)
	}

	result := make([]SubscriptionResponse, 0, len(allSubs))
	for _, sub := range allSubs {
		result = append(result, *s.toResponse(sub))
	}
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ProcessSuccessfulPayment  (Paystack webhook callback)
// ─────────────────────────────────────────────────────────────────────────────

func (s *SubscriptionService) ProcessSuccessfulPayment(ctx context.Context, paymentRef string) error {
	subscription, err := s.subscriptionRepo.FindByPaymentRef(ctx, paymentRef)
	if err != nil {
		return fmt.Errorf("subscription not found for ref %s: %w", paymentRef, err)
	}
	if subscription.Status == "active" {
		return nil // already processed — idempotent
	}

	subscription.Status = "active"
	subscription.SubscriptionDate = time.Now()
	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to activate subscription: %w", err)
	}

	// Award points + draw entry
	return s.awardSubscriptionPoints(ctx, subscription)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetConfig / UpdateConfig
// ─────────────────────────────────────────────────────────────────────────────

func (s *SubscriptionService) GetConfig(ctx context.Context) (map[string]interface{}, error) {
	var cfg entities.DailySubscriptionConfig
	if err := s.db.WithContext(ctx).First(&cfg).Error; err != nil {
		entries := 1
		isPaid := true
		return map[string]interface{}{
			"amount":               int64(2000),
			"draw_entries_earned":  &entries,
			"is_paid":              &isPaid,
			"description":          "Daily ₦20 subscription — 1 point + 1 draw entry per day",
			"terms_and_conditions": "",
		}, nil
	}
	return map[string]interface{}{
		"id":                   cfg.ID,
		"amount":               cfg.Amount,
		"draw_entries_earned":  cfg.DrawEntriesEarned,
		"is_paid":              cfg.IsPaid,
		"description":          cfg.Description,
		"terms_and_conditions": cfg.TermsAndConditions,
		"updated_at":           cfg.UpdatedAt,
	}, nil
}

func (s *SubscriptionService) UpdateConfig(ctx context.Context, config map[string]interface{}) error {
	var cfg entities.DailySubscriptionConfig
	if err := s.db.WithContext(ctx).First(&cfg).Error; err != nil {
		cfg = entities.DailySubscriptionConfig{}
	}
	if v, ok := config["amount"]; ok {
		switch val := v.(type) {
		case float64:
			cfg.Amount = int64(val)
		case int64:
			cfg.Amount = val
		case int:
			cfg.Amount = int64(val)
		}
	}
	if v, ok := config["draw_entries_earned"]; ok {
		if val, ok := v.(float64); ok {
			n := int(val)
			cfg.DrawEntriesEarned = &n
		}
	}
	if v, ok := config["is_paid"]; ok {
		if val, ok := v.(bool); ok {
			cfg.IsPaid = &val
		}
	}
	if v, ok := config["description"]; ok {
		cfg.Description = fmt.Sprintf("%v", v)
	}
	if v, ok := config["terms_and_conditions"]; ok {
		cfg.TermsAndConditions = fmt.Sprintf("%v", v)
	}
	return s.db.WithContext(ctx).Save(&cfg).Error
}

// ─────────────────────────────────────────────────────────────────────────────
// Admin helpers
// ─────────────────────────────────────────────────────────────────────────────

func (s *SubscriptionService) GetActiveSubscriptionCount(ctx context.Context) (int64, error) {
	return s.subscriptionRepo.CountByStatus(ctx, "active")
}

func (s *SubscriptionService) GetAllSubscriptions(ctx context.Context, page, perPage int) ([]*entities.DailySubscriptions, int64, error) {
	offset := (page - 1) * perPage
	subs, err := s.subscriptionRepo.FindAll(ctx, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get subscriptions: %w", err)
	}
	total, err := s.subscriptionRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get subscription count: %w", err)
	}
	return subs, total, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal helpers
// ─────────────────────────────────────────────────────────────────────────────

// resolveDefaultTierID returns the UUID of the Bronze spin_tier (lowest tier).
// The DailySubscription.TierID field is a UUID NOT NULL in the entity struct
// even though the actual SQL column allows NULL.  We always set it to avoid
// any ORM-level validation failure.
func (s *SubscriptionService) resolveDefaultTierID(ctx context.Context) uuid.UUID {
	var tier struct {
		ID uuid.UUID `gorm:"column:id"`
	}
	err := s.db.WithContext(ctx).
		Table("spin_tiers").
		Where("LOWER(tier_name) = ? AND is_active = true", "bronze").
		Order("sort_order ASC").
		Limit(1).
		Scan(&tier).Error
	if err != nil || tier.ID == uuid.Nil {
		// Fallback: pick any active tier
		s.db.WithContext(ctx).
			Table("spin_tiers").
			Where("is_active = true").
			Order("sort_order ASC").
			Limit(1).
			Scan(&tier)
	}
	if tier.ID == uuid.Nil {
		// Absolute fallback: generate a deterministic nil-safe UUID
		return uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}
	return tier.ID
}

// awardSubscriptionPoints credits 1 point and 1 draw entry to the user
// associated with the subscription.
func (s *SubscriptionService) awardSubscriptionPoints(ctx context.Context, sub *entities.DailySubscription) error {
	if sub.UserID == nil {
		// Guest subscription — find user by MSISDN
		user, err := s.userRepo.FindByMSISDN(ctx, sub.MSISDN)
		if err != nil || user == nil {
			logger.Info("[SubscriptionService] No user found for points award — guest subscription",
				zap.String("msisdn", sub.MSISDN))
			return nil
		}
		sub.UserID = &user.ID
	}

	// Update user points
	if err := s.db.WithContext(ctx).
		Model(&entities.User{}).
		Where("id = ?", sub.UserID).
		UpdateColumn("total_points", gorm.Expr("total_points + 1")).
		Error; err != nil {
		logger.Error("[SubscriptionService] Failed to award points", zap.Error(err))
		return err
	}

	logger.Info("[SubscriptionService] Awarded 1 point + 1 draw entry",
		zap.String("user_id", sub.UserID.String()),
		zap.String("msisdn", sub.MSISDN),
	)
	return nil
}

// toResponse converts a DailySubscription entity to SubscriptionResponse
func (s *SubscriptionService) toResponse(sub *entities.DailySubscription) *SubscriptionResponse {
	nextBilling := sub.NextBillingDate
	if nextBilling.IsZero() {
		nextBilling = sub.SubscriptionDate.Add(24 * time.Hour)
	}
	amount := sub.DailyAmount
	if amount == 0 {
		amount = sub.Amount
	}
	if amount == 0 {
		amount = 2000 // ₦20 in kobo
	}
	pm := sub.PaymentMethod
	if pm == "" {
		pm = "paystack"
	}
	return &SubscriptionResponse{
		ID:            sub.ID,
		MSISDN:        sub.MSISDN,
		Network:       "auto",
		Status:        sub.Status,
		PaymentMethod: pm,
		DailyAmount:   amount,
		NextBilling:   nextBilling,
		CreatedAt:     sub.CreatedAt,
		CancelledAt:   sub.CancelledAt,
	}
}

// getUserEmail is kept for backward compatibility
func (s *SubscriptionService) getUserEmail(ctx context.Context, msisdn string) string {
	return fmt.Sprintf("%s@rechargemax.ng", msisdn)
}

func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
