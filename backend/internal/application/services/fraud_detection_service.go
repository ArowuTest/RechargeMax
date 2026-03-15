package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// fraudConfig holds tunable thresholds.
// These can be made configurable via platform_settings in a future iteration.
const (
	maxAmountKobo         int64 = 50_000_000 // ₦500,000 hard ceiling per transaction
	maxTxPerHour                = 15          // velocity: max transactions in 1 hour
	maxFailedRecharge1h         = 5           // too many failures in 1 hour → suspicious
	maxDailyRechargeKobo  int64 = 200_000_000 // ₦2,000,000 daily cap per MSISDN
)

// FraudDetectionService performs lightweight, database-backed fraud checks.
// It is intentionally non-blocking: a DB error is logged but does not block
// the transaction (fail-open strategy to preserve uptime).
type FraudDetectionService struct {
	db *gorm.DB
}

// NewFraudDetectionService creates a FraudDetectionService.
// Pass a *gorm.DB to enable velocity/blacklist checks; nil disables DB checks.
func NewFraudDetectionService(db ...*gorm.DB) *FraudDetectionService {
	svc := &FraudDetectionService{}
	if len(db) > 0 {
		svc.db = db[0]
	}
	return svc
}

// CheckTransaction checks if a single transaction is potentially fraudulent.
// Returns (isFraud bool, reason string, error).
func (s *FraudDetectionService) CheckTransaction(ctx context.Context, msisdn string, amount int64) (bool, string, error) {
	// 1. Hard amount ceiling
	if amount > maxAmountKobo {
		return true, fmt.Sprintf("amount ₦%.2f exceeds maximum allowed ₦%.2f",
			float64(amount)/100, float64(maxAmountKobo)/100), nil
	}

	if s.db == nil {
		return false, "", nil
	}

	// 2. MSISDN blacklist check
	var blacklistCount int64
	if err := s.db.WithContext(ctx).
		Table("msisdn_blacklist").
		Where("msisdn = ? AND is_active = true", msisdn).
		Count(&blacklistCount).Error; err != nil {
		log.Printf("[fraud] blacklist check error: %v", err)
	} else if blacklistCount > 0 {
		return true, "MSISDN is blacklisted", nil
	}

	// 3. Transaction velocity (hourly)
	windowStart := time.Now().Add(-1 * time.Hour)
	var txCount int64
	if err := s.db.WithContext(ctx).
		Table("transactions").
		Where("msisdn = ? AND created_at >= ?", msisdn, windowStart).
		Count(&txCount).Error; err != nil {
		log.Printf("[fraud] velocity check error: %v", err)
	} else if txCount >= maxTxPerHour {
		return true, fmt.Sprintf("transaction velocity exceeded: %d transactions in 1 hour", txCount), nil
	}

	// 4. Daily cumulative amount cap
	dayStart := time.Now().Truncate(24 * time.Hour)
	var dailyTotal struct{ Total int64 }
	if err := s.db.WithContext(ctx).
		Table("transactions").
		Select("COALESCE(SUM(amount), 0) AS total").
		Where("msisdn = ? AND status = 'SUCCESS' AND created_at >= ?", msisdn, dayStart).
		Scan(&dailyTotal).Error; err != nil {
		log.Printf("[fraud] daily cap check error: %v", err)
	} else if dailyTotal.Total+amount > maxDailyRechargeKobo {
		return true, fmt.Sprintf("daily limit exceeded: ₦%.2f cumulative", float64(dailyTotal.Total)/100), nil
	}

	return false, "", nil
}

// CheckRecharge checks if a recharge is potentially fraudulent.
// Returns (isFraud bool, reason string, error).
func (s *FraudDetectionService) CheckRecharge(ctx context.Context, msisdn string, amount int64) (bool, string, error) {
	// Delegate to CheckTransaction — recharge is a subtype of transaction
	isFraud, reason, err := s.CheckTransaction(ctx, msisdn, amount)
	if err != nil || isFraud {
		return isFraud, reason, err
	}

	if s.db == nil {
		return false, "", nil
	}

	// Extra check: too many failed recharges in the last hour (credential stuffing / card testing)
	windowStart := time.Now().Add(-1 * time.Hour)
	var failedCount int64
	if err := s.db.WithContext(ctx).
		Table("transactions").
		Where("msisdn = ? AND status = 'FAILED' AND created_at >= ?", msisdn, windowStart).
		Count(&failedCount).Error; err != nil {
		log.Printf("[fraud] failed-recharge check error: %v", err)
	} else if failedCount >= maxFailedRecharge1h {
		return true, fmt.Sprintf("too many failed recharges: %d in the last hour", failedCount), nil
	}

	return false, "", nil
}
