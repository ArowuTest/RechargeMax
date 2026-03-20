package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"rechargemax/internal/application/services"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/logger"
)

// SubscriptionBillingJob processes daily recurring subscription charges.
//
// It runs every BillingJobInterval (default 15 min) and:
//
//  1. Finds every active subscription whose next_billing_date has passed and
//     for which today's subscription_billings row doesn't yet exist (or
//     exists in pending/attempted state with a next_retry_at that has passed).
//
//  2. Charges the stored Paystack authorization_code via
//     PaymentService.ChargeAuthorization().
//
//  3a. On success → marks billing "attempted" (webhook will set "completed").
//      The webhook's ProcessRecurringPayment() awards points.
//
//  3b. On charge API error (not a Paystack webhook event, just the HTTP call) →
//      schedules a retry according to RetryDelays:
//        retry 0 → +1h, retry 1 → +3h, retry 2 → +8h
//      After max_retries exhausted → marks billing "failed",
//      increments subscription.consecutive_failures,
//      auto-pauses subscription if consecutive_failures >= ConsecutiveFailureLimit.
//
//  4. Sets next_billing_date = tomorrow(now) on the subscription so the next
//     run of the job picks it up the following day.
//
// Edge cases handled:
//   - No auth code: subscription never had a successful first payment;
//     skip and leave next_billing_date unchanged (user must complete checkout).
//   - Double-billing guard: UNIQUE(subscription_id, billing_date) in DB.
//   - Idempotent points award: SubscriptionBilling.PointsAwarded flag.
//   - Cancellation mid-day: cancelled subscriptions are skipped.
//   - Multiple active lines per user: each line is billed independently.
//   - Paused subscriptions: skipped until admin/user resumes.

const BillingJobInterval = 15 * time.Minute

// SubscriptionBillingJob drives daily auto-renewal.
type SubscriptionBillingJob struct {
	db                  *gorm.DB
	subscriptionService *services.SubscriptionService
	paymentService      *services.PaymentService
	stopCh              chan struct{}
}

// NewSubscriptionBillingJob creates the job.
func NewSubscriptionBillingJob(
	db *gorm.DB,
	subscriptionService *services.SubscriptionService,
	paymentService *services.PaymentService,
) *SubscriptionBillingJob {
	return &SubscriptionBillingJob{
		db:                  db,
		subscriptionService: subscriptionService,
		paymentService:      paymentService,
		stopCh:              make(chan struct{}),
	}
}

// Start launches the job ticker in a background goroutine.
func (j *SubscriptionBillingJob) Start() {
	logger.Info("[SubscriptionBillingJob] Started", zap.Duration("interval", BillingJobInterval))
	go func() {
		// Run immediately on startup (catches any missed billings after a deploy)
		j.runOnce(context.Background())

		ticker := time.NewTicker(BillingJobInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				j.runOnce(context.Background())
			case <-j.stopCh:
				logger.Info("[SubscriptionBillingJob] Stopped")
				return
			}
		}
	}()
}

// Stop signals the job goroutine to exit.
func (j *SubscriptionBillingJob) Stop() {
	close(j.stopCh)
}

// ─────────────────────────────────────────────────────────────────────────────
// Core run loop
// ─────────────────────────────────────────────────────────────────────────────

func (j *SubscriptionBillingJob) runOnce(ctx context.Context) {
	now := time.Now()
	logger.Info("[SubscriptionBillingJob] Running billing cycle", zap.Time("now", now))

	// ── Phase 1: process new billings (subscription.next_billing_date has passed) ──
	newCount, newErr := j.processNewBillings(ctx, now)
	if newErr != nil {
		logger.Error("[SubscriptionBillingJob] processNewBillings error", zap.Error(newErr))
	}

	// ── Phase 2: retry pending/attempted billings whose next_retry_at has passed ──
	retryCount, retryErr := j.processRetries(ctx, now)
	if retryErr != nil {
		logger.Error("[SubscriptionBillingJob] processRetries error", zap.Error(retryErr))
	}

	if newCount+retryCount > 0 {
		logger.Info("[SubscriptionBillingJob] Cycle complete",
			zap.Int("new_billings", newCount),
			zap.Int("retries", retryCount),
		)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 1 — new billings
// ─────────────────────────────────────────────────────────────────────────────

func (j *SubscriptionBillingJob) processNewBillings(ctx context.Context, now time.Time) (int, error) {
	// Find active subscriptions that are due for billing and have a stored auth code
	var subs []entities.DailySubscription
	err := j.db.WithContext(ctx).
		Where(`
			LOWER(status) = 'active'
			AND auto_renew = true
			AND paystack_authorization_code IS NOT NULL
			AND next_billing_date <= ?`, now).
		Find(&subs).Error
	if err != nil {
		return 0, fmt.Errorf("query active subscriptions: %w", err)
	}

	processed := 0
	for i := range subs {
		sub := &subs[i]
		if err := j.billSubscription(ctx, sub, now); err != nil {
			logger.Error("[SubscriptionBillingJob] Failed to bill subscription",
				zap.String("id", sub.ID.String()),
				zap.String("msisdn", sub.MSISDN),
				zap.Error(err),
			)
			continue
		}
		processed++
	}
	return processed, nil
}

// billSubscription creates a billing record and attempts the charge for one subscription.
func (j *SubscriptionBillingJob) billSubscription(ctx context.Context, sub *entities.DailySubscription, now time.Time) error {
	today := truncateDay(now)
	ref := fmt.Sprintf("RCR_%s_%s", sub.ID.String()[:8], today.Format("20060102"))

	// ── Double-billing guard: create billing record with UNIQUE constraint ─────
	billing := &entities.SubscriptionBilling{
		ID:             uuid.New(),
		SubscriptionID: sub.ID,
		MSISDN:         sub.MSISDN,
		BillingDate:    today,
		Amount:         sub.DailyAmount,
		EntriesToAward: sub.BundleQuantity,
		PointsToAward:  sub.BundleQuantity,
		Status:         "pending",
		PaymentReference: ptrStr(ref),
		MaxRetries:     len(entities.RetryDelays),
	}

	result := j.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(billing)
	if result.Error != nil {
		return fmt.Errorf("create billing record: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		// Row already existed (double-run guard hit) — check its status
		var existing entities.SubscriptionBilling
		j.db.WithContext(ctx).
			Where("subscription_id = ? AND billing_date = ?", sub.ID, today).
			First(&existing)
		if existing.Status == "completed" || existing.Status == "skipped" {
			// Already done — advance next_billing_date and return
			return j.advanceNextBilling(ctx, sub, now)
		}
		// Use the existing record for retry logic
		billing = &existing
	}

	// ── Attempt the charge ─────────────────────────────────────────────────────
	return j.attemptCharge(ctx, sub, billing, now)
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 2 — retries
// ─────────────────────────────────────────────────────────────────────────────

func (j *SubscriptionBillingJob) processRetries(ctx context.Context, now time.Time) (int, error) {
	var billings []entities.SubscriptionBilling
	err := j.db.WithContext(ctx).
		Where(`status IN ('pending','attempted') AND next_retry_at <= ?`, now).
		Find(&billings).Error
	if err != nil {
		return 0, fmt.Errorf("query retries: %w", err)
	}

	processed := 0
	for i := range billings {
		b := &billings[i]
		// Load the parent subscription
		sub, err := j.loadSub(ctx, b.SubscriptionID)
		if err != nil || sub == nil {
			continue
		}
		if sub.Status != "active" {
			// Subscription was cancelled/paused after billing record was created
			j.markBillingSkipped(ctx, b)
			continue
		}
		if err := j.attemptCharge(ctx, sub, b, now); err != nil {
			logger.Error("[SubscriptionBillingJob] Retry charge failed",
				zap.String("billing_id", b.ID.String()),
				zap.Error(err),
			)
		}
		processed++
	}
	return processed, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Charge attempt (shared by Phase 1 and Phase 2)
// ─────────────────────────────────────────────────────────────────────────────

func (j *SubscriptionBillingJob) attemptCharge(
	ctx context.Context,
	sub *entities.DailySubscription,
	billing *entities.SubscriptionBilling,
	now time.Time,
) error {
	authCode := ""
	if sub.PaystackAuthorizationCode != nil {
		authCode = *sub.PaystackAuthorizationCode
	}
	email := fmt.Sprintf("%s@rechargemax.ng", sub.MSISDN)

	ref := ""
	if billing.PaymentReference != nil {
		ref = *billing.PaymentReference
	}

	// Mark as attempted immediately so a crash mid-flight doesn't leave it in pending
	billing.Status = "attempted"
	billing.RetryCount++ // reflects this attempt
	j.db.WithContext(ctx).Save(billing)

	// Call Paystack charge authorization API
	chargeResult, err := j.paymentService.ChargeAuthorization(ctx, services.ChargeAuthRequest{
		AuthorizationCode: authCode,
		Email:             email,
		Amount:            billing.Amount,
		Reference:         ref,
		Metadata: map[string]interface{}{
			"subscription_id": sub.ID.String(),
			"msisdn":          sub.MSISDN,
			"billing_date":    billing.BillingDate.Format("2006-01-02"),
			"type":            "subscription_renewal",
		},
	})

	if err == nil && chargeResult.Status == "success" {
		// Immediate success (some channels return success inline, not via webhook)
		now2 := time.Now()
		billing.Status = "completed"
		billing.ProcessedAt = &now2
		billing.NextRetryAt = nil
		if chargeResult.TransactionID != 0 {
			billing.PaystackTransactionID = &chargeResult.TransactionID
		}
		j.db.WithContext(ctx).Save(billing)

		// Award points immediately
		_ = j.subscriptionService.AwardPointsForBilling(ctx, sub, billing)
		return j.advanceNextBilling(ctx, sub, now)
	}

	if err == nil && chargeResult.Status == "pending" {
		// Webhook will fire later with confirmation — mark as attempted, wait for webhook
		billing.NextRetryAt = nil // don't auto-retry; webhook drives this
		j.db.WithContext(ctx).Save(billing)
		return nil
	}

	// ── Charge failed (err != nil or status == "failed") ─────────────────────
	failureReason := "unknown"
	if err != nil {
		failureReason = err.Error()
	} else if chargeResult != nil {
		failureReason = chargeResult.GatewayResponse
	}
	billing.GatewayResponse = failureReason

	if billing.RetryCount <= billing.MaxRetries {
		// Schedule next retry
		delayIdx := billing.RetryCount - 1
		if delayIdx >= len(entities.RetryDelays) {
			delayIdx = len(entities.RetryDelays) - 1
		}
		nextRetry := now.Add(entities.RetryDelays[delayIdx])
		billing.Status = "pending" // back to pending for retry
		billing.NextRetryAt = &nextRetry
		j.db.WithContext(ctx).Save(billing)

		logger.Info("[SubscriptionBillingJob] Charge failed — retry scheduled",
			zap.String("subscription_id", sub.ID.String()),
			zap.String("msisdn", sub.MSISDN),
			zap.Int("retry_count", billing.RetryCount),
			zap.Time("next_retry_at", nextRetry),
			zap.String("reason", failureReason),
		)
		return nil
	}

	// ── All retries exhausted ─────────────────────────────────────────────────
	billing.Status = "failed"
	billing.FailureReason = failureReason
	billing.NextRetryAt = nil
	j.db.WithContext(ctx).Save(billing)

	sub.ConsecutiveFailures++
	if sub.ConsecutiveFailures >= entities.ConsecutiveFailureLimit {
		pauseTime := time.Now()
		sub.Status = "paused"
		sub.PausedAt = &pauseTime
		logger.Info("[SubscriptionBillingJob] Subscription auto-paused after consecutive failures",
			zap.String("id", sub.ID.String()),
			zap.String("msisdn", sub.MSISDN),
			zap.Int("consecutive_failures", sub.ConsecutiveFailures),
		)
	}
	// Always advance next_billing_date to tomorrow even on failure
	// so the job tries again the next day (after the pause threshold resets)
	_ = j.advanceNextBilling(ctx, sub, now)

	logger.Info("[SubscriptionBillingJob] Day billing failed — no points awarded",
		zap.String("subscription_id", sub.ID.String()),
		zap.String("msisdn", sub.MSISDN),
		zap.String("billing_date", billing.BillingDate.Format("2006-01-02")),
		zap.String("reason", failureReason),
	)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func (j *SubscriptionBillingJob) advanceNextBilling(ctx context.Context, sub *entities.DailySubscription, now time.Time) error {
	nextBilling := tomorrowAt8(now)
	sub.NextBillingDate = nextBilling
	return j.db.WithContext(ctx).
		Model(sub).
		Updates(map[string]interface{}{
			"next_billing_date":    nextBilling,
			"consecutive_failures": sub.ConsecutiveFailures,
			"status":               sub.Status,
			"paused_at":            sub.PausedAt,
			"updated_at":           time.Now(),
		}).Error
}

func (j *SubscriptionBillingJob) markBillingSkipped(ctx context.Context, b *entities.SubscriptionBilling) {
	now := time.Now()
	j.db.WithContext(ctx).
		Model(b).
		Updates(map[string]interface{}{
			"status":     "skipped",
			"updated_at": now,
		})
}

func (j *SubscriptionBillingJob) loadSub(ctx context.Context, id uuid.UUID) (*entities.DailySubscription, error) {
	var sub entities.DailySubscription
	err := j.db.WithContext(ctx).Where("id = ?", id).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Date utilities
// ─────────────────────────────────────────────────────────────────────────────

// truncateDay returns the UTC date at 00:00 for billing_date key.
func truncateDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// tomorrowAt8 returns 07:00 UTC (= 08:00 WAT) the next calendar day.
func tomorrowAt8(from time.Time) time.Time {
	y, m, d := from.Date()
	return time.Date(y, m, d+1, 7, 0, 0, 0, time.UTC)
}

func ptrStr(s string) *string { return &s }
