package jobs

import (
	"go.uber.org/zap"
	"rechargemax/internal/logger"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ────────────────────────────────────────────────────────────────────────────
// Interfaces
// ────────────────────────────────────────────────────────────────────────────

// ReconciliationPaymentVerifier can verify whether a payment reference was paid
// and return the amount in kobo (int64).
type ReconciliationPaymentVerifier interface {
	VerifyPaystackPayment(ctx context.Context, reference string) (paid bool, amount int64, err error)
}

// ReconciliationRechargeProcessor can process a completed payment reference.
type ReconciliationRechargeProcessor interface {
	ProcessSuccessfulPayment(ctx context.Context, paymentRef string) error
	RecoverProcessingTransactions(ctx context.Context) error
}

// ReconciliationNotifier can send a user-facing notification.
type ReconciliationNotifier interface {
	SendSMS(ctx context.Context, msisdn, message string) error
}

// ────────────────────────────────────────────────────────────────────────────
// Job
// ────────────────────────────────────────────────────────────────────────────

// ReconciliationJob repairs PENDING transactions that have been stuck for more
// than 1 hour by re-verifying payment and either fulfilling or failing them.
type ReconciliationJob struct {
	db          *gorm.DB
	payment     ReconciliationPaymentVerifier
	recharge    ReconciliationRechargeProcessor
	notifier    ReconciliationNotifier
}

// NewReconciliationJob constructs a ReconciliationJob.
// notifier is optional (may be nil).
func NewReconciliationJob(
	db *gorm.DB,
	payment ReconciliationPaymentVerifier,
	recharge ReconciliationRechargeProcessor,
	notifier ReconciliationNotifier,
) *ReconciliationJob {
	return &ReconciliationJob{db: db, payment: payment, recharge: recharge, notifier: notifier}
}

// StartScheduled runs the job on a background goroutine at the given interval.
func (j *ReconciliationJob) StartScheduled(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.Info("[ReconciliationJob] Stopping")
				return
			case <-ticker.C:
				if err := j.Run(ctx); err != nil {
					logger.Error("[ReconciliationJob] Error", zap.Error(err))
				}
			}
		}
	}()
}

// stuckTransaction holds the minimal fields needed per row.
type stuckTransaction struct {
	ID               string    `gorm:"column:id"`
	PaymentReference string    `gorm:"column:payment_reference"`
	Amount           int64     `gorm:"column:amount"`
	Msisdn           string    `gorm:"column:msisdn"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

// RunProcessingRecovery immediately requeues all PROCESSING transactions older than 5 minutes.
// Called once at server startup to recover any transactions stuck from a previous run.
func (j *ReconciliationJob) RunProcessingRecovery(ctx context.Context) {
	go func() {
		// Wait 30 seconds after startup so the server is fully initialised.
		time.Sleep(30 * time.Second)
		logger.Info("[ReconciliationJob] Running startup PROCESSING recovery")
		if err := j.recharge.RecoverProcessingTransactions(ctx); err != nil {
			logger.Error("[ReconciliationJob] PROCESSING recovery error", zap.Error(err))
		} else {
			logger.Info("[ReconciliationJob] PROCESSING recovery complete")
		}
	}()
}

// Run executes one reconciliation pass.
func (j *ReconciliationJob) Run(ctx context.Context) error {
	logger.Info("[ReconciliationJob] Starting pass…")

	cutoff := time.Now().Add(-1 * time.Hour)

	var rows []stuckTransaction
	err := j.db.WithContext(ctx).
		// NOTE: We select t.msisdn directly from the transactions table rather than
		// JOINing users. The previous INNER JOIN silently excluded guest transactions
		// (user_id = NULL) from reconciliation, leaving them stuck in PENDING forever.
		Raw(`
			SELECT t.id, t.payment_reference, t.amount, t.created_at, t.msisdn
			FROM transactions t
			WHERE t.status = 'PENDING'
			AND t.created_at < ?
			AND t.processed_at IS NULL
			ORDER BY t.created_at ASC
			LIMIT 100
		`, cutoff).
		Scan(&rows).Error
	if err != nil {
		return fmt.Errorf("query stuck transactions: %w", err)
	}

	var processed, succeeded, failed int

	for _, row := range rows {
		if ctx.Err() != nil {
			break
		}

		verified, paidAmount, err := j.payment.VerifyPaystackPayment(ctx, row.PaymentReference)
		if err != nil {
			logger.Info("[ReconciliationJob] verify", zap.Error(err), zap.Any("row.PaymentReference", row.PaymentReference))
			continue
		}

		processed++

		if verified && paidAmount == row.Amount {
			// ── Success path ──────────────────────────────────────────────
			if err := j.recharge.ProcessSuccessfulPayment(ctx, row.PaymentReference); err != nil {
				logger.Info("[ReconciliationJob] process", zap.Error(err), zap.Any("row.PaymentReference", row.PaymentReference))
				continue
			}
			succeeded++
			logger.Info("[ReconciliationJob] recovered", zap.String("id", row.ID), zap.Float64("amount_naira", float64(row.Amount)/100))

		} else {
			// ── Failure path ──────────────────────────────────────────────
			now := time.Now()
			if err := j.db.WithContext(ctx).Exec(
				`UPDATE transactions SET status='FAILED', updated_at=?, processed_at=? WHERE id=?`,
				now, now, row.ID,
			).Error; err != nil {
				logger.Error("[ReconciliationJob] mark failed", zap.Error(err), zap.String("id", row.ID))
				continue
			}

			// Also mark the linked VTU transaction if present
			j.db.WithContext(ctx).Exec(
				`UPDATE vtu_transactions
				 SET status='FAILED',
				     error_message='Payment verification failed during reconciliation',
				     failed_at=?
				 WHERE parent_transaction_id=?`,
				now, row.ID,
			)

			failed++

			if j.notifier != nil {
				msg := fmt.Sprintf(
					"Your recharge (Ref: %s) could not be completed. "+
						"If you were charged, a refund will be processed within 24 hours.",
					row.PaymentReference,
				)
				if smsErr := j.notifier.SendSMS(ctx, row.Msisdn, msg); smsErr != nil {
					logger.Info("[ReconciliationJob] SMS to", zap.Error(smsErr), zap.Any("row.Msisdn", row.Msisdn))
				}
			}

			logger.Error("[ReconciliationJob] marked failed", zap.String("id", row.ID))
		}
	}

	logger.Info("[ReconciliationJob] done", zap.Int("processed", processed), zap.Int("succeeded", succeeded), zap.Int("failed", failed))

	// Also recover any PROCESSING transactions that the background requery loop missed.
	if err := j.recharge.RecoverProcessingTransactions(ctx); err != nil {
		logger.Error("[ReconciliationJob] PROCESSING recovery error", zap.Error(err))
	}

	return nil
}
