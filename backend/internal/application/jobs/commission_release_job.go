package jobs

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CommissionReleaseJob auto-approves affiliate commissions after the hold period expires.
// Business rules:
//   - PENDING commissions older than `holdDays` are moved to APPROVED status
//   - APPROVED commissions are written to the affiliate wallet balance
//   - Runs on a schedule (default: every 6 hours)
type CommissionReleaseJob struct {
	db       *gorm.DB
	holdDays int // configurable hold period in days
}

func NewCommissionReleaseJob(db *gorm.DB) *CommissionReleaseJob {
	return &CommissionReleaseJob{
		db:       db,
		holdDays: 7, // 7-day hold by default; overridden by platform_settings
	}
}

// getHoldDays reads the commission hold period from platform_settings.
func (j *CommissionReleaseJob) getHoldDays(ctx context.Context) int {
	var setting struct {
		Value string `gorm:"column:setting_value"`
	}
	if err := j.db.WithContext(ctx).
		Table("platform_settings").
		Where("setting_key = ?", "affiliate.commission_hold_days").
		Select("setting_value").
		First(&setting).Error; err == nil {
		var days int
		if n, err2 := fmt.Sscanf(setting.Value, "%d", &days); n == 1 && err2 == nil && days > 0 {
			return days
		}
	}
	return j.holdDays
}

// Run executes the commission release job.
func (j *CommissionReleaseJob) Run(ctx context.Context) error {
	fmt.Println("[CommissionReleaseJob] Starting...")

	holdDays := j.getHoldDays(ctx)
	holdCutoff := time.Now().AddDate(0, 0, -holdDays)

	// Step 1: Auto-approve PENDING commissions past the hold period
	result := j.db.WithContext(ctx).Exec(`
		UPDATE affiliate_commissions
		SET    status     = 'APPROVED',
		       updated_at = NOW()
		WHERE  status    = 'PENDING'
		AND    earned_at < $1
	`, holdCutoff)
	if result.Error != nil {
		return fmt.Errorf("failed to auto-approve commissions: %w", result.Error)
	}
	fmt.Printf("[CommissionReleaseJob] Auto-approved %d commissions (hold > %d days)\n",
		result.RowsAffected, holdDays)

	// Step 2: Credit APPROVED commissions to affiliate wallet balance
	// We do this in a transaction to ensure atomicity per affiliate
	type affSummary struct {
		AffiliateID string
		TotalAmount int64
	}
	var summaries []affSummary
	if err := j.db.WithContext(ctx).Raw(`
		SELECT affiliate_id::text, SUM(commission_amount) AS total_amount
		FROM   affiliate_commissions
		WHERE  status = 'APPROVED'
		GROUP  BY affiliate_id
	`).Scan(&summaries).Error; err != nil {
		return fmt.Errorf("failed to query approved commissions: %w", err)
	}

	credited := 0
	for _, s := range summaries {
		err := j.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// Credit wallet
			if err := tx.Exec(`
				INSERT INTO wallets (id, user_id, balance, currency, status, created_at, updated_at)
				SELECT uuid_generate_v4(), u.id, $1, 'NGN', 'active', NOW(), NOW()
				FROM   users u
				INNER  JOIN affiliates a ON a.msisdn = u.msisdn
				WHERE  a.id::text = $2
				ON CONFLICT (user_id) DO UPDATE
				  SET balance    = wallets.balance + EXCLUDED.balance,
				      updated_at = NOW()
			`, s.TotalAmount, s.AffiliateID).Error; err != nil {
				return fmt.Errorf("wallet credit failed for affiliate %s: %w", s.AffiliateID, err)
			}

			// Mark commissions as PAID
			return tx.Exec(`
				UPDATE affiliate_commissions
				SET    status  = 'PAID',
				       paid_at = NOW()
				WHERE  affiliate_id::text = $1
				AND    status = 'APPROVED'
			`, s.AffiliateID).Error
		})
		if err != nil {
			fmt.Printf("[CommissionReleaseJob] ERROR for affiliate %s: %v\n", s.AffiliateID, err)
			continue
		}
		credited++
	}

	fmt.Printf("[CommissionReleaseJob] Credited commissions for %d affiliates\n", credited)
	return nil
}

// StartScheduled launches the job on a ticker and runs it until ctx is cancelled.
func (j *CommissionReleaseJob) StartScheduled(ctx context.Context, interval time.Duration) {
	go func() {
		// Run immediately on start
		if err := j.Run(ctx); err != nil {
			fmt.Printf("[CommissionReleaseJob] Initial run error: %v\n", err)
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("[CommissionReleaseJob] Stopping.")
				return
			case <-ticker.C:
				if err := j.Run(ctx); err != nil {
					fmt.Printf("[CommissionReleaseJob] Error: %v\n", err)
				}
			}
		}
	}()
}
