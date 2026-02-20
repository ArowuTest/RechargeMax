package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type ussdRechargeRepositoryGorm struct {
	db *gorm.DB
}

// NewUSSDRechargeRepository creates a new USSD recharge repository
func NewUSSDRechargeRepository(db *gorm.DB) repositories.USSDRechargeRepository {
	return &ussdRechargeRepositoryGorm{db: db}
}

// Create creates a new USSD recharge record
func (r *ussdRechargeRepositoryGorm) Create(ctx context.Context, recharge *entities.USSDRecharge) error {
	return r.db.WithContext(ctx).Create(recharge).Error
}

// Update updates an existing USSD recharge record
func (r *ussdRechargeRepositoryGorm) Update(ctx context.Context, recharge *entities.USSDRecharge) error {
	return r.db.WithContext(ctx).Save(recharge).Error
}

// FindByID finds a USSD recharge by ID
func (r *ussdRechargeRepositoryGorm) FindByID(ctx context.Context, id uuid.UUID) (*entities.USSDRecharge, error) {
	var recharge entities.USSDRecharge
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("id = ?", id).
		First(&recharge).Error
	if err != nil {
		return nil, err
	}
	return &recharge, nil
}

// FindByTransactionRef finds a USSD recharge by transaction reference
func (r *ussdRechargeRepositoryGorm) FindByTransactionRef(ctx context.Context, transactionRef string) (*entities.USSDRecharge, error) {
	var recharge entities.USSDRecharge
	err := r.db.WithContext(ctx).
		Where("transaction_ref = ?", transactionRef).
		First(&recharge).Error
	if err != nil {
		return nil, err
	}
	return &recharge, nil
}

// FindByMSISDN finds USSD recharges by MSISDN within a date range
func (r *ussdRechargeRepositoryGorm) FindByMSISDN(ctx context.Context, msisdn string, startDate, endDate time.Time) ([]*entities.USSDRecharge, error) {
	var recharges []*entities.USSDRecharge
	query := r.db.WithContext(ctx).Where("msisdn = ?", msisdn)

	if !startDate.IsZero() {
		query = query.Where("received_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("received_at <= ?", endDate)
	}

	err := query.Order("received_at DESC").Find(&recharges).Error
	return recharges, err
}

// FindByUserID finds USSD recharges by user ID
func (r *ussdRechargeRepositoryGorm) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.USSDRecharge, error) {
	var recharges []*entities.USSDRecharge
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("received_at DESC").
		Find(&recharges).Error
	return recharges, err
}

// FindUnprocessed finds all unprocessed USSD recharges
func (r *ussdRechargeRepositoryGorm) FindUnprocessed(ctx context.Context) ([]*entities.USSDRecharge, error) {
	var recharges []*entities.USSDRecharge
	err := r.db.WithContext(ctx).
		Where("processed_at IS NULL").
		Where("status = ?", "success").
		Order("received_at ASC").
		Find(&recharges).Error
	return recharges, err
}

// CreateWebhookLog creates a new webhook log
func (r *ussdRechargeRepositoryGorm) CreateWebhookLog(ctx context.Context, log *entities.USSDWebhookLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// UpdateWebhookLog updates an existing webhook log
func (r *ussdRechargeRepositoryGorm) UpdateWebhookLog(ctx context.Context, log *entities.USSDWebhookLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// FindWebhookLogByID finds a webhook log by ID
func (r *ussdRechargeRepositoryGorm) FindWebhookLogByID(ctx context.Context, id uuid.UUID) (*entities.USSDWebhookLog, error) {
	var log entities.USSDWebhookLog
	err := r.db.WithContext(ctx).
		Preload("USSDRecharge").
		Where("id = ?", id).
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// FindWebhookLogs finds webhook logs by provider within a date range
func (r *ussdRechargeRepositoryGorm) FindWebhookLogs(ctx context.Context, provider string, startDate, endDate time.Time) ([]*entities.USSDWebhookLog, error) {
	var logs []*entities.USSDWebhookLog
	query := r.db.WithContext(ctx)

	if provider != "" {
		query = query.Where("provider = ?", provider)
	}
	if !startDate.IsZero() {
		query = query.Where("received_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("received_at <= ?", endDate)
	}

	err := query.Order("received_at DESC").Find(&logs).Error
	return logs, err
}

// FindFailedWebhookLogs finds all failed webhook logs
func (r *ussdRechargeRepositoryGorm) FindFailedWebhookLogs(ctx context.Context) ([]*entities.USSDWebhookLog, error) {
	var logs []*entities.USSDWebhookLog
	err := r.db.WithContext(ctx).
		Where("status = ?", "failed").
		Where("processed_at IS NULL").
		Order("received_at ASC").
		Find(&logs).Error
	return logs, err
}
