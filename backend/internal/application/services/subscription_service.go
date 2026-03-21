package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
	"rechargemax/internal/errors"
	"rechargemax/internal/logger"
	"rechargemax/internal/utils"
)

// SubscriptionService handles subscription operations.
type SubscriptionService struct {
	subscriptionRepo repositories.SubscriptionRepository
	userRepo         repositories.UserRepository
	paymentService   *PaymentService
	hlrService       *HLRService
	db               *gorm.DB
}

// CreateSubscriptionRequest is the input to CreateSubscription.
// All fields are optional at parse time — the handler resolves MSISDN from JWT
// or body before calling, and PaymentMethod defaults to "paystack".
type CreateSubscriptionRequest struct {
	MSISDN        string `json:"msisdn"`
	Network       string `json:"network"`
	PaymentMethod string `json:"payment_method"`
	// Entries is the number of daily draw entries (and points) the user wants.
	// Each entry costs ₦20 (PricePerEntry kobo).
	// Min 1, max 100.  Defaults to 1 if not provided.
	Entries int `json:"entries"`
}

// SubscriptionResponse is the API-facing subscription DTO.
type SubscriptionResponse struct {
	ID              uuid.UUID  `json:"id"`
	SubscriptionCode string    `json:"subscription_code"`
	MSISDN          string     `json:"msisdn"`
	Network         string     `json:"network"`
	Status          string     `json:"status"`
	PaymentMethod   string     `json:"payment_method"`
	Entries         int        `json:"entries"`           // entries awarded per day
	DailyAmount     int64      `json:"daily_amount"`      // kobo
	DailyAmountNGN  float64    `json:"daily_amount_ngn"`  // naira (display)
	NextBilling     time.Time  `json:"next_billing"`
	CreatedAt       time.Time  `json:"created_at"`
	CancelledAt     *time.Time `json:"cancelled_at,omitempty"`
	PaymentURL      string     `json:"payment_url,omitempty"`
	// Summary across all active subscriptions for this user
	TotalDailyEntries int     `json:"total_daily_entries,omitempty"`
	TotalDailyPoints  int     `json:"total_daily_points,omitempty"`
}

// NewSubscriptionService creates a new subscription service.
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

// CreateSubscription creates a new recurring daily subscription line.
//
// Key design decisions:
//  - Multiple active lines per user are allowed (see entities.DailySubscription).
//  - Pricing: ₦20 per entry → entries=5 costs ₦100/day (10,000 kobo).
//  - Status starts as "pending" until Paystack confirms the first payment.
//  - On confirmation (ProcessFirstPayment), auth_code is stored and status → active.
//  - From that point the SubscriptionBillingJob handles daily auto-renewal.
func (s *SubscriptionService) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*SubscriptionResponse, error) {
	// ── 1. Resolve & normalise MSISDN ─────────────────────────────────────────
	if req.MSISDN == "" {
		return nil, errors.BadRequest("Phone number is required")
	}
	msisdn, err := utils.NormalizeMSISDN(req.MSISDN)
	if err != nil {
		return nil, errors.BadRequest(fmt.Sprintf("Invalid phone number: %s", req.MSISDN))
	}

	// ── 2. Entries & amount ────────────────────────────────────────────────────
	if req.Entries <= 0 {
		req.Entries = 1
	}
	if req.Entries > 100 {
		return nil, errors.BadRequest("Maximum 100 entries per subscription")
	}
	dailyAmountKobo := entities.PricePerEntry * int64(req.Entries) // ₦20 × entries

	// ── 3. Payment method default ──────────────────────────────────────────────
	if req.PaymentMethod == "" {
		req.PaymentMethod = "paystack"
	}

	// ── 4. Network detection (non-fatal) ──────────────────────────────────────
	networkHint := req.Network
	detectedNetworkResult, _ := s.hlrService.DetectNetwork(ctx, msisdn, &networkHint)
	detectedNetwork := "MTN" // safe default
	if detectedNetworkResult != nil && detectedNetworkResult.Network != "" {
		detectedNetwork = detectedNetworkResult.Network
	}

	// ── 5. Resolve user (optional — guest subscriptions are allowed) ──────────
	var userID *uuid.UUID
	var userEmail string
	if user, err := s.userRepo.FindByMSISDN(ctx, msisdn); err == nil && user != nil {
		userID = &user.ID
		userEmail = fmt.Sprintf("%s@rechargemax.ng", msisdn)
	}

	// ── 6. Resolve default tier ────────────────────────────────────────────────
	tierID := s.resolveDefaultTierID(ctx)

	// ── 7. Build the subscription row ─────────────────────────────────────────
	// Purge any stale 'pending' rows for this MSISDN+entries combination that were
	// never activated (payment abandoned). This prevents the uniqueIndex on
	// subscription_code from blocking a legitimate retry.
	s.db.WithContext(ctx).
		Where("msisdn = ? AND status = 'pending' AND bundle_quantity = ?", msisdn, req.Entries).
		Delete(&entities.DailySubscription{})

	now := time.Now()
	// Use UUID fragment for guaranteed uniqueness — unix timestamp alone collides
	// on rapid retries or multiple subscriptions created in the same second.
	newID := uuid.New()
	code := fmt.Sprintf("SUB%s%s", msisdn[len(msisdn)-4:], strings.ToUpper(newID.String()[:8]))
	entries := req.Entries
	nextBilling := tomorrow(now)

	sub := &entities.DailySubscription{
		ID:               newID,
		SubscriptionCode: code,
		UserID:           userID,
		MSISDN:           msisdn,
		TierID:           &tierID,
		BundleQuantity:   entries,
		TotalEntries:     0, // incremented per successful billing day
		DailyAmount:      dailyAmountKobo,
		Amount:           dailyAmountKobo, // legacy column
		DrawEntriesEarned: &entries,
		PointsEarned:     &entries,
		Status:           "pending",
		AutoRenew:        true,
		NextBillingDate:  nextBilling,
		PaymentMethod:    req.PaymentMethod,
		SubscriptionDate: now,
		CustomerEmail:    ptrString(userEmail),
	}

	if err := s.subscriptionRepo.Create(ctx, sub); err != nil {
		logger.Error("[SubscriptionService] Create failed",
			zap.String("msisdn", msisdn), zap.Error(err))
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "23505") {
			// Last-resort: if UUID-based code still collides (extremely unlikely),
			// generate a fresh one and retry once.
			retryID := uuid.New()
			sub.ID = retryID
			sub.SubscriptionCode = fmt.Sprintf("SUB%s%s", msisdn[len(msisdn)-4:], strings.ToUpper(retryID.String()[:8]))
			if retryErr := s.subscriptionRepo.Create(ctx, sub); retryErr != nil {
				return nil, errors.Conflict("Could not create subscription — please try again")
			}
		} else {
			return nil, fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	// ── 8. Build response ──────────────────────────────────────────────────────
	resp := &SubscriptionResponse{
		ID:             sub.ID,
		SubscriptionCode: sub.SubscriptionCode,
		MSISDN:         msisdn,
		Network:        detectedNetwork,
		Status:         sub.Status,
		PaymentMethod:  req.PaymentMethod,
		Entries:        entries,
		DailyAmount:    dailyAmountKobo,
		DailyAmountNGN: float64(dailyAmountKobo) / 100,
		NextBilling:    nextBilling,
		CreatedAt:      sub.CreatedAt,
	}

	// ── 9. Initialise first payment ────────────────────────────────────────────
	// The first payment MUST go through the Paystack checkout UI so we can
	// capture a reusable authorization_code for subsequent auto-charges.
	// We ALWAYS use the Paystack checkout for the first payment (even if the
	// user later chooses DCB for renewals).
	ref := fmt.Sprintf("SUB_%s_%d", sub.ID.String()[:8], now.Unix())
	payReq := PaymentRequest{
		Amount:    dailyAmountKobo,
		Email:     userEmail,
		Reference: ref,
		Metadata: map[string]interface{}{
			"msisdn":           msisdn,
			"type":             "subscription_first_payment",
			"subscription_id":  sub.ID.String(),
			"entries":          entries,
		},
	}
	payURL, payErr := s.paymentService.InitializePayment(ctx, payReq)
	if payErr != nil {
		// Non-fatal: row created, user can retry payment later
		logger.Error("[SubscriptionService] Payment init failed", zap.Error(payErr))
		resp.PaymentURL = ""
	} else {
		resp.PaymentURL = payURL
		sub.PaymentReference = &ref
		_ = s.subscriptionRepo.Update(ctx, sub)
	}

	// Compute user's total daily commitment across all active subscriptions
	resp.TotalDailyEntries, resp.TotalDailyPoints = s.totalDailyCommitment(ctx, msisdn)

	logger.Info("[SubscriptionService] Created subscription",
		zap.String("id", sub.ID.String()),
		zap.String("msisdn", msisdn),
		zap.Int("entries", entries),
		zap.Float64("daily_ngn", resp.DailyAmountNGN),
	)
	return resp, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ProcessFirstPayment  — called by webhook on charge.success for SUB_ prefix
// ─────────────────────────────────────────────────────────────────────────────

// ProcessFirstPayment activates a subscription after the first Paystack
// checkout payment succeeds and stores the reusable authorization_code.
//
// authCode:     the Paystack authorization.authorization_code from the webhook
// customerCode: the Paystack customer.customer_code from the webhook
func (s *SubscriptionService) ProcessFirstPayment(ctx context.Context, paymentRef, authCode, customerCode string) error {
	sub, err := s.subscriptionRepo.FindByPaymentRef(ctx, paymentRef)
	if err != nil {
		return fmt.Errorf("subscription not found for ref %s: %w", paymentRef, err)
	}
	if sub.Status == "active" {
		return nil // idempotent
	}

	now := time.Now()
	sub.Status = "active"
	sub.SubscriptionDate = now
	sub.LastBillingDate = &now
	sub.NextBillingDate = tomorrow(now)
	sub.PaystackAuthorizationCode = ptrString(authCode)
	sub.PaystackCustomerCode = ptrString(customerCode)
	isPaid := true
	sub.IsPaid = &isPaid

	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to activate subscription: %w", err)
	}

	// Create today's billing record as completed (first payment already confirmed)
	today := truncateToDay(now)
	billing := &entities.SubscriptionBilling{
		ID:             uuid.New(),
		SubscriptionID: sub.ID,
		MSISDN:         sub.MSISDN,
		BillingDate:    today,
		Amount:         sub.DailyAmount,
		EntriesToAward: sub.BundleQuantity,
		PointsToAward:  sub.BundleQuantity,
		Status:         "completed",
		PaymentReference: sub.PaymentReference,
		ProcessedAt:    &now,
	}
	// Upsert — ignore conflict if somehow already exists
	if dbErr := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(billing).Error; dbErr != nil {
		logger.Error("[SubscriptionService] Failed to create first billing record", zap.Error(dbErr))
	}

	// Award points + entries for today
	return s.awardSubscriptionPoints(ctx, sub, billing)
}

// ProcessSuccessfulPayment is the legacy entry point kept for backward-compat.
// New callers should use ProcessFirstPayment or ProcessRecurringPayment.
func (s *SubscriptionService) ProcessSuccessfulPayment(ctx context.Context, paymentRef string) error {
	return s.ProcessFirstPayment(ctx, paymentRef, "", "")
}

// ─────────────────────────────────────────────────────────────────────────────
// ProcessRecurringPayment — called by webhook on charge.success for RCR_ prefix
// ─────────────────────────────────────────────────────────────────────────────

// ProcessRecurringPayment handles the webhook confirmation of a daily auto-charge.
// It marks the corresponding subscription_billing row as completed and awards points.
func (s *SubscriptionService) ProcessRecurringPayment(ctx context.Context, paymentRef string, txID int64) error {
	// Find the billing record waiting for this reference
	var billing entities.SubscriptionBilling
	err := s.db.WithContext(ctx).
		Where("payment_reference = ?", paymentRef).
		First(&billing).Error
	if err != nil {
		// Could be a race — log and ignore (job will retry)
		logger.Error("[SubscriptionService] Billing record not found for recurring ref",
			zap.String("ref", paymentRef), zap.Error(err))
		return nil
	}

	if billing.Status == "completed" {
		return nil // idempotent
	}

	now := time.Now()
	billing.Status = "completed"
	billing.PaystackTransactionID = &txID
	billing.ProcessedAt = &now
	if err := s.db.WithContext(ctx).Save(&billing).Error; err != nil {
		return fmt.Errorf("failed to update billing record: %w", err)
	}

	// Load the parent subscription to award points
	sub, err := s.subscriptionRepo.FindByID(ctx, billing.SubscriptionID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Update subscription billing metadata
	sub.LastBillingDate = &now
	sub.NextBillingDate = tomorrow(now)
	sub.ConsecutiveFailures = 0 // reset on success
	sub.TotalBilledAmount += billing.Amount
	_ = s.subscriptionRepo.Update(ctx, sub)

	return s.awardSubscriptionPoints(ctx, sub, &billing)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetSubscription / GetSubscriptions
// ─────────────────────────────────────────────────────────────────────────────

func (s *SubscriptionService) GetSubscription(ctx context.Context, msisdn string) (*SubscriptionResponse, error) {
	msisdn = s.normMSISDN(msisdn)
	active, err := s.subscriptionRepo.FindActiveByMSISDN(ctx, msisdn)
	if err == nil && active != nil {
		return s.toResponse(active), nil
	}
	// Fall back to user lookup
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

// GetAllActiveSubscriptions returns all active subscriptions for a MSISDN
// (multiple lines possible).
func (s *SubscriptionService) GetAllActiveSubscriptions(ctx context.Context, msisdn string) ([]*entities.DailySubscription, error) {
	msisdn = s.normMSISDN(msisdn)
	var subs []*entities.DailySubscription
	err := s.db.WithContext(ctx).
		Where("msisdn = ? AND LOWER(status) = 'active'", msisdn).
		Order("created_at ASC").
		Find(&subs).Error
	return subs, err
}

// ─────────────────────────────────────────────────────────────────────────────
// CancelSubscription
// ─────────────────────────────────────────────────────────────────────────────

func (s *SubscriptionService) CancelSubscription(ctx context.Context, msisdn string) error {
	return s.CancelSubscriptionByID(ctx, msisdn, uuid.Nil)
}

// CancelSubscriptionByID cancels a specific subscription line by its ID.
// If id is uuid.Nil, the most recently created active subscription is cancelled.
func (s *SubscriptionService) CancelSubscriptionByID(ctx context.Context, msisdn string, id uuid.UUID) error {
	msisdn = s.normMSISDN(msisdn)

	var sub *entities.DailySubscription

	if id != uuid.Nil {
		var found entities.DailySubscription
		if err := s.db.WithContext(ctx).
			Where("id = ? AND msisdn = ?", id, msisdn).
			First(&found).Error; err != nil {
			return errors.NotFound("Subscription not found")
		}
		sub = &found
	} else {
		// Cancel the most recent active line
		active, err := s.subscriptionRepo.FindActiveByMSISDN(ctx, msisdn)
		if err != nil || active == nil {
			return errors.NotFound("No active subscription found to cancel")
		}
		sub = active
	}

	now := time.Now()
	sub.Status = "cancelled"
	sub.CancelledAt = &now
	sub.CancellationReason = "user_requested"
	sub.AutoRenew = false

	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	// Mark any pending billing records for future dates as skipped
	s.db.WithContext(ctx).
		Model(&entities.SubscriptionBilling{}).
		Where("subscription_id = ? AND billing_date > ? AND status = 'pending'",
			sub.ID, truncateToDay(now)).
		Updates(map[string]interface{}{
			"status":     "skipped",
			"updated_at": now,
		})

	logger.Info("[SubscriptionService] Subscription cancelled",
		zap.String("id", sub.ID.String()),
		zap.String("msisdn", msisdn),
	)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetSubscriptionHistory
// ─────────────────────────────────────────────────────────────────────────────

func (s *SubscriptionService) GetSubscriptionHistory(ctx context.Context, msisdn string) ([]SubscriptionResponse, error) {
	msisdn = s.normMSISDN(msisdn)
	var subs []*entities.DailySubscription
	if err := s.db.WithContext(ctx).
		Where("msisdn = ?", msisdn).
		Order("subscription_date DESC").
		Find(&subs).Error; err != nil {
		return nil, err
	}
	result := make([]SubscriptionResponse, 0, len(subs))
	for _, sub := range subs {
		r := s.toResponse(sub)
		result = append(result, *r)
	}
	return result, nil
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
			"price_per_entry":      int64(2000),
			"draw_entries_earned":  &entries,
			"is_paid":              &isPaid,
			"description":          "Daily ₦20 subscription — 1 point + 1 draw entry per day",
			"terms_and_conditions": "Renews automatically every 24 hours. Cancel anytime.",
		}, nil
	}
	return map[string]interface{}{
		"id":                   cfg.ID,
		"amount":               cfg.Amount,
		"price_per_entry":      entities.PricePerEntry,
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

// awardSubscriptionPoints credits points and draw entries to the user.
// It checks billing.PointsAwarded first — safe to call multiple times.
func (s *SubscriptionService) awardSubscriptionPoints(ctx context.Context, sub *entities.DailySubscription, billing *entities.SubscriptionBilling) error {
	if billing.PointsAwarded {
		return nil // already awarded — idempotent
	}
	if billing.Status != "completed" {
		return nil // don't award for non-completed billings
	}

	points := billing.PointsToAward
	entries := billing.EntriesToAward

	// Resolve user_id
	resolvedUserID := sub.UserID
	if resolvedUserID == nil {
		user, err := s.userRepo.FindByMSISDN(ctx, sub.MSISDN)
		if err == nil && user != nil {
			resolvedUserID = &user.ID
		}
	}

	if resolvedUserID != nil {
		// Award points
		if err := s.db.WithContext(ctx).
			Model(&entities.User{}).
			Where("id = ?", *resolvedUserID).
			UpdateColumn("total_points", gorm.Expr("total_points + ?", points)).
			Error; err != nil {
			return fmt.Errorf("failed to award points: %w", err)
		}
	}

	// Mark billing record as awarded (atomic update — prevents double-award)
	now := time.Now()
	if err := s.db.WithContext(ctx).
		Model(&entities.SubscriptionBilling{}).
		Where("id = ? AND points_awarded = false", billing.ID).
		Updates(map[string]interface{}{
			"points_awarded": true,
			"updated_at":     now,
		}).Error; err != nil {
		return fmt.Errorf("failed to mark points awarded: %w", err)
	}

	// Update subscription lifetime totals
	s.db.WithContext(ctx).
		Model(&entities.DailySubscription{}).
		Where("id = ?", sub.ID).
		Updates(map[string]interface{}{
			"total_entries":        gorm.Expr("total_entries + ?", entries),
			"total_points_awarded": gorm.Expr("total_points_awarded + ?", points),
			"updated_at":           now,
		})

	logger.Info("[SubscriptionService] Points awarded",
		zap.String("subscription_id", sub.ID.String()),
		zap.String("msisdn", sub.MSISDN),
		zap.Int("points", points),
		zap.Int("entries", entries),
	)
	return nil
}

// resolveDefaultTierID returns the UUID of the Bronze spin tier.
func (s *SubscriptionService) resolveDefaultTierID(ctx context.Context) uuid.UUID {
	var tier struct{ ID uuid.UUID `gorm:"column:id"` }
	s.db.WithContext(ctx).Table("spin_tiers").
		Where("LOWER(tier_name) = 'bronze' AND is_active = true").
		Order("sort_order ASC").Limit(1).Scan(&tier)
	if tier.ID == uuid.Nil {
		s.db.WithContext(ctx).Table("spin_tiers").
			Where("is_active = true").Order("sort_order ASC").Limit(1).Scan(&tier)
	}
	if tier.ID == uuid.Nil {
		return uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}
	return tier.ID
}

// totalDailyCommitment sums entries + points across all ACTIVE subscription lines.
func (s *SubscriptionService) totalDailyCommitment(ctx context.Context, msisdn string) (totalEntries int, totalPoints int) {
	var subs []*entities.DailySubscription
	s.db.WithContext(ctx).
		Where("msisdn = ? AND LOWER(status) = 'active'", msisdn).
		Find(&subs)
	for _, sub := range subs {
		totalEntries += sub.BundleQuantity
		totalPoints += sub.BundleQuantity
	}
	return
}

func (s *SubscriptionService) normMSISDN(msisdn string) string {
	if n, err := utils.NormalizeMSISDN(msisdn); err == nil {
		return n
	}
	return msisdn
}

func (s *SubscriptionService) toResponse(sub *entities.DailySubscription) *SubscriptionResponse {
	nb := sub.NextBillingDate
	if nb.IsZero() {
		nb = sub.SubscriptionDate.Add(24 * time.Hour)
	}
	amt := sub.DailyAmount
	if amt == 0 {
		amt = sub.Amount
	}
	if amt == 0 {
		amt = entities.PricePerEntry * int64(sub.BundleQuantity)
	}
	pm := sub.PaymentMethod
	if pm == "" {
		pm = "paystack"
	}
	entries := sub.BundleQuantity
	if entries == 0 {
		entries = 1
	}
	return &SubscriptionResponse{
		ID:               sub.ID,
		SubscriptionCode: sub.SubscriptionCode,
		MSISDN:           sub.MSISDN,
		Network:          "auto",
		Status:           sub.Status,
		PaymentMethod:    pm,
		Entries:          entries,
		DailyAmount:      amt,
		DailyAmountNGN:   float64(amt) / 100,
		NextBilling:      nb,
		CreatedAt:        sub.CreatedAt,
		CancelledAt:      sub.CancelledAt,
	}
}

func (s *SubscriptionService) getUserEmail(ctx context.Context, msisdn string) string {
	return fmt.Sprintf("%s@rechargemax.ng", msisdn)
}

// ─────────────────────────────────────────────────────────────────────────────
// Date helpers
// ─────────────────────────────────────────────────────────────────────────────

// tomorrow returns 08:00 Africa/Lagos (UTC+1) the next calendar day.
// Using 08:00 WAT ensures billing attempts happen during business hours.
func tomorrow(from time.Time) time.Time {
	y, m, d := from.Date()
	// 07:00 UTC = 08:00 WAT
	return time.Date(y, m, d+1, 7, 0, 0, 0, time.UTC)
}

// truncateToDay returns the UTC date at 00:00 (used as billing_date key).
func truncateToDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}



// ptrString returns a pointer to a string, or nil for empty strings.
func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// AwardPointsForBilling is the exported wrapper used by SubscriptionBillingJob.
func (s *SubscriptionService) AwardPointsForBilling(ctx context.Context, sub *entities.DailySubscription, billing *entities.SubscriptionBilling) error {
	return s.awardSubscriptionPoints(ctx, sub, billing)
}
