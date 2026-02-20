package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type subscriptionTierRepositoryGorm struct {
	db *gorm.DB
}

// NewSubscriptionTierRepository creates a new subscription tier repository
func NewSubscriptionTierRepository(db *gorm.DB) repositories.SubscriptionTierRepository {
	return &subscriptionTierRepositoryGorm{db: db}
}

// Create creates a new subscription tier
func (r *subscriptionTierRepositoryGorm) Create(ctx context.Context, tier *entities.SubscriptionTier) error {
	return r.db.WithContext(ctx).Create(tier).Error
}

// Update updates an existing subscription tier
func (r *subscriptionTierRepositoryGorm) Update(ctx context.Context, tier *entities.SubscriptionTier) error {
	return r.db.WithContext(ctx).Save(tier).Error
}

// Delete soft deletes a subscription tier
func (r *subscriptionTierRepositoryGorm) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entities.SubscriptionTier{}).Error
}

// FindByID finds a subscription tier by ID
func (r *subscriptionTierRepositoryGorm) FindByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionTier, error) {
	var tier entities.SubscriptionTier
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tier).Error
	if err != nil {
		return nil, err
	}
	return &tier, nil
}

// FindAll finds all subscription tiers
func (r *subscriptionTierRepositoryGorm) FindAll(ctx context.Context) ([]*entities.SubscriptionTier, error) {
	var tiers []*entities.SubscriptionTier
	err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&tiers).Error
	return tiers, err
}

// FindActive finds all active subscription tiers
func (r *subscriptionTierRepositoryGorm) FindActive(ctx context.Context) ([]*entities.SubscriptionTier, error) {
	var tiers []*entities.SubscriptionTier
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, created_at ASC").
		Find(&tiers).Error
	return tiers, err
}

// CreatePricing creates a new pricing record
func (r *subscriptionTierRepositoryGorm) CreatePricing(ctx context.Context, pricing *entities.SubscriptionPricing) error {
	return r.db.WithContext(ctx).Create(pricing).Error
}

// UpdatePricing updates an existing pricing record
func (r *subscriptionTierRepositoryGorm) UpdatePricing(ctx context.Context, pricing *entities.SubscriptionPricing) error {
	return r.db.WithContext(ctx).Save(pricing).Error
}

// GetCurrentPricing gets the current active pricing
func (r *subscriptionTierRepositoryGorm) GetCurrentPricing(ctx context.Context) (*entities.SubscriptionPricing, error) {
	var pricing entities.SubscriptionPricing
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("effective_from DESC").
		First(&pricing).Error
	if err != nil {
		return nil, err
	}
	return &pricing, nil
}

// GetPricingHistory gets all pricing history
func (r *subscriptionTierRepositoryGorm) GetPricingHistory(ctx context.Context) ([]*entities.SubscriptionPricing, error) {
	var history []*entities.SubscriptionPricing
	err := r.db.WithContext(ctx).Order("effective_from DESC").Find(&history).Error
	return history, err
}

// CreateDailySubscription creates a new daily subscription
func (r *subscriptionTierRepositoryGorm) CreateDailySubscription(ctx context.Context, sub *entities.DailySubscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

// UpdateDailySubscription updates an existing daily subscription
func (r *subscriptionTierRepositoryGorm) UpdateDailySubscription(ctx context.Context, sub *entities.DailySubscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

// FindDailySubscriptionByID finds a daily subscription by ID
func (r *subscriptionTierRepositoryGorm) FindDailySubscriptionByID(ctx context.Context, id uuid.UUID) (*entities.DailySubscription, error) {
	var sub entities.DailySubscription
	err := r.db.WithContext(ctx).
		Preload("Tier").
		Where("id = ?", id).
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// FindDailySubscriptionsByMSISDN finds all daily subscriptions for an MSISDN
func (r *subscriptionTierRepositoryGorm) FindDailySubscriptionsByMSISDN(ctx context.Context, msisdn string) ([]*entities.DailySubscription, error) {
	var subs []*entities.DailySubscription
	err := r.db.WithContext(ctx).
		Preload("Tier").
		Where("msisdn = ?", msisdn).
		Order("created_at DESC").
		Find(&subs).Error
	return subs, err
}

// FindActiveDailySubscriptions finds all active daily subscriptions
func (r *subscriptionTierRepositoryGorm) FindActiveDailySubscriptions(ctx context.Context) ([]*entities.DailySubscription, error) {
	var subs []*entities.DailySubscription
	err := r.db.WithContext(ctx).
		Preload("Tier").
		Where("status = ?", "active").
		Where("auto_renew = ?", true).
		Find(&subs).Error
	return subs, err
}

// FindDailySubscriptionsDueForBilling finds subscriptions due for billing on a specific date
func (r *subscriptionTierRepositoryGorm) FindDailySubscriptionsDueForBilling(ctx context.Context, date time.Time) ([]*entities.DailySubscription, error) {
	var subs []*entities.DailySubscription
	err := r.db.WithContext(ctx).
		Preload("Tier").
		Where("status = ?", "active").
		Where("auto_renew = ?", true).
		Where("next_billing_date <= ?", date).
		Find(&subs).Error
	return subs, err
}

// CreateBilling creates a new billing record
func (r *subscriptionTierRepositoryGorm) CreateBilling(ctx context.Context, billing *entities.SubscriptionBilling) error {
	return r.db.WithContext(ctx).Create(billing).Error
}

// UpdateBilling updates an existing billing record
func (r *subscriptionTierRepositoryGorm) UpdateBilling(ctx context.Context, billing *entities.SubscriptionBilling) error {
	return r.db.WithContext(ctx).Save(billing).Error
}

// FindBillingByID finds a billing record by ID
func (r *subscriptionTierRepositoryGorm) FindBillingByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionBilling, error) {
	var billing entities.SubscriptionBilling
	err := r.db.WithContext(ctx).
		Preload("Subscription").
		Preload("Subscription.Tier").
		Where("id = ?", id).
		First(&billing).Error
	if err != nil {
		return nil, err
	}
	return &billing, nil
}

// FindBillingsBySubscriptionID finds all billing records for a subscription
func (r *subscriptionTierRepositoryGorm) FindBillingsBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entities.SubscriptionBilling, error) {
	var billings []*entities.SubscriptionBilling
	err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("billing_date DESC").
		Find(&billings).Error
	return billings, err
}

// FindBillingsByMSISDN finds billing records for an MSISDN within a date range
func (r *subscriptionTierRepositoryGorm) FindBillingsByMSISDN(ctx context.Context, msisdn string, startDate, endDate time.Time) ([]*entities.SubscriptionBilling, error) {
	var billings []*entities.SubscriptionBilling
	query := r.db.WithContext(ctx).
		Preload("Subscription").
		Preload("Subscription.Tier").
		Where("msisdn = ?", msisdn)

	if !startDate.IsZero() {
		query = query.Where("billing_date >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("billing_date <= ?", endDate)
	}

	err := query.Order("billing_date DESC").Find(&billings).Error
	return billings, err
}

// FindPendingBillings finds all pending billing records
func (r *subscriptionTierRepositoryGorm) FindPendingBillings(ctx context.Context) ([]*entities.SubscriptionBilling, error) {
	var billings []*entities.SubscriptionBilling
	err := r.db.WithContext(ctx).
		Preload("Subscription").
		Preload("Subscription.Tier").
		Where("status = ?", "pending").
		Order("billing_date ASC").
		Find(&billings).Error
	return billings, err
}
