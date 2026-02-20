package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// RECONCILIATION JOB - Fix Issue #17
// ============================================================================
// This job reconciles stuck transactions by:
// - Finding PENDING transactions older than 1 hour
// - Verifying payment status with Paystack
// - Processing if payment succeeded
// - Marking as failed if payment failed
// - Sending notifications to users
// ============================================================================

type ReconciliationJob struct {
	db                  *sql.DB
	paymentService      PaymentServiceInterface
	rechargeService     RechargeServiceInterface
	notificationService NotificationServiceInterface
}

func NewReconciliationJob(
	db *sql.DB,
	paymentService PaymentServiceInterface,
	rechargeService RechargeServiceInterface,
	notificationService NotificationServiceInterface,
) *ReconciliationJob {
	return &ReconciliationJob{
		db:                  db,
		paymentService:      paymentService,
		rechargeService:     rechargeService,
		notificationService: notificationService,
	}
}

// Run executes the reconciliation job
func (j *ReconciliationJob) Run(ctx context.Context) error {
	fmt.Println("[ReconciliationJob] Starting reconciliation...")

	// Find stuck transactions
	rows, err := j.db.QueryContext(ctx, `
		SELECT t.id, t.payment_reference, t.amount, t.created_at, u.msisdn, u.email
		FROM transactions t
		INNER JOIN users u ON u.id = t.user_id
		WHERE t.type = 'RECHARGE'
		AND t.status = 'PENDING'
		AND t.created_at < NOW() - INTERVAL '1 hour'
		AND t.processed_at IS NULL
		ORDER BY t.created_at ASC
		LIMIT 100
	`)

	if err != nil {
		return fmt.Errorf("failed to query stuck transactions: %w", err)
	}
	defer rows.Close()

	var processed int
	var succeeded int
	var failed int

	for rows.Next() {
		var transactionID uuid.UUID
		var paymentRef string
		var amount int64
		var createdAt time.Time
		var msisdn string
		var email sql.NullString

		err := rows.Scan(&transactionID, &paymentRef, &amount, &createdAt, &msisdn, &email)
		if err != nil {
			fmt.Printf("[ReconciliationJob] Error scanning row: %v\n", err)
			continue
		}

		// Verify payment with Paystack
		verified, paidAmount, err := j.paymentService.VerifyPayment(ctx, paymentRef)
		if err != nil {
			fmt.Printf("[ReconciliationJob] Error verifying payment %s: %v\n", paymentRef, err)
			continue
		}

		processed++

		if verified && paidAmount == amount {
			// Payment succeeded - process recharge
			err = j.rechargeService.ProcessSuccessfulPayment(ctx, paymentRef)
			if err != nil {
				fmt.Printf("[ReconciliationJob] Error processing payment %s: %v\n", paymentRef, err)
				continue
			}

			succeeded++
			fmt.Printf("[ReconciliationJob] Processed stuck transaction %s (₦%.2f)\n",
				transactionID, float64(amount)/100)

		} else {
			// Payment failed or not found - mark as failed
			_, err = j.db.ExecContext(ctx, `
				UPDATE transactions
				SET status = 'FAILED',
				    updated_at = NOW(),
				    processed_at = NOW()
				WHERE id = $1
			`, transactionID)

			if err != nil {
				fmt.Printf("[ReconciliationJob] Error marking transaction as failed %s: %v\n", transactionID, err)
				continue
			}

			// Update VTU transaction
			_, err = j.db.ExecContext(ctx, `
				UPDATE vtu_transactions
				SET status = 'FAILED',
				    error_message = 'Payment verification failed during reconciliation',
				    failed_at = NOW()
				WHERE parent_transaction_id = $1
			`, transactionID)

			if err != nil {
				fmt.Printf("[ReconciliationJob] Error updating VTU transaction %s: %v\n", transactionID, err)
			}

			failed++

			// Send notification
			if j.notificationService != nil {
				j.notificationService.SendMultiChannel(ctx, msisdn,
					"Transaction Failed",
					fmt.Sprintf("Your recharge transaction (Ref: %s) could not be completed. If you were charged, you will be refunded within 24 hours.",
						paymentRef),
					"recharge",
					map[string]interface{}{
						"transaction_id": transactionID.String(),
						"status":         "FAILED",
						"reference":      paymentRef,
					},
				)
			}

			fmt.Printf("[ReconciliationJob] Marked stuck transaction as failed %s\n", transactionID)
		}
	}

	fmt.Printf("[ReconciliationJob] Completed: processed=%d, succeeded=%d, failed=%d\n",
		processed, succeeded, failed)

	return nil
}

// Interfaces for dependency injection

type PaymentServiceInterface interface {
	VerifyPayment(ctx context.Context, reference string) (bool, int64, error)
}

type RechargeServiceInterface interface {
	ProcessSuccessfulPayment(ctx context.Context, paymentRef string) error
}

type NotificationServiceInterface interface {
	SendMultiChannel(ctx context.Context, msisdn, title, message, category string, data map[string]interface{}) error
}
