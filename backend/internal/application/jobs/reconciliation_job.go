package jobs

import (
	"context"
	"fmt"
	"log"
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
				log.Println("[ReconciliationJob] Stopping")
				return
			case <-ticker.C:
				if err := j.Run(ctx); err != nil {
					log.Printf("[ReconciliationJob] Error: %v", err)
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

// Run executes one reconciliation pass.
func (j *ReconciliationJob) Run(ctx context.Context) error {
	log.Println("[ReconciliationJob] Starting pass…")

	cutoff := time.Now().Add(-1 * time.Hour)

	var rows []stuckTransaction
	err := j.db.WithContext(ctx).
		Raw(`
			SELECT t.id, t.payment_reference, t.amount, t.created_at, u.msisdn
			FROM transactions t
			INNER JOIN users u ON u.id = t.user_id
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
			log.Printf("[ReconciliationJob] verify %s: %v", row.PaymentReference, err)
			continue
		}

		processed++

		if verified && paidAmount == row.Amount {
			// ── Success path ──────────────────────────────────────────────
			if err := j.recharge.ProcessSuccessfulPayment(ctx, row.PaymentReference); err != nil {
				log.Printf("[ReconciliationJob] process %s: %v", row.PaymentReference, err)
				continue
			}
			succeeded++
			log.Printf("[ReconciliationJob] recovered %s (₦%.2f)", row.ID, float64(row.Amount)/100)

		} else {
			// ── Failure path ──────────────────────────────────────────────
			now := time.Now()
			if err := j.db.WithContext(ctx).Exec(
				`UPDATE transactions SET status='FAILED', updated_at=?, processed_at=? WHERE id=?`,
				now, now, row.ID,
			).Error; err != nil {
				log.Printf("[ReconciliationJob] mark failed %s: %v", row.ID, err)
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
					log.Printf("[ReconciliationJob] SMS to %s: %v", row.Msisdn, smsErr)
				}
			}

			log.Printf("[ReconciliationJob] marked failed %s", row.ID)
		}
	}

	log.Printf("[ReconciliationJob] done: total=%d succeeded=%d failed=%d", processed, succeeded, failed)
	return nil
}
