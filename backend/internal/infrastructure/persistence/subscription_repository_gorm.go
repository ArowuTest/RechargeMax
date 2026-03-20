package persistence

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type subscriptionRepositoryGORM struct {
	db *gorm.DB
}

// NewSubscriptionRepository creates a new GORM implementation
func NewSubscriptionRepository(db *gorm.DB) repositories.SubscriptionRepository {
	return &subscriptionRepositoryGORM{db: db}
}

// Create creates a new record
func (r *subscriptionRepositoryGORM) Create(ctx context.Context, entity *entities.DailySubscriptions) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// FindByID finds a record by ID
func (r *subscriptionRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.DailySubscriptions, error) {
	var entity entities.DailySubscriptions
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindAll retrieves all records with pagination
func (r *subscriptionRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.DailySubscriptions, error) {
	var entities []*entities.DailySubscriptions
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&entities).Error
	return entities, err
}

// Update updates a record
func (r *subscriptionRepositoryGORM) Update(ctx context.Context, entity *entities.DailySubscriptions) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete deletes a record
func (r *subscriptionRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.DailySubscriptions{}, "id = ?", id).Error
}

// Count returns the total number of records
func (r *subscriptionRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.DailySubscriptions{}).Count(&count).Error
	return count, err
}

// FindByUserID retrieves all subscriptions for a specific user
func (r *subscriptionRepositoryGORM) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.DailySubscriptions, error) {
	var results []*entities.DailySubscriptions
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("subscription_date DESC").
		Find(&results).Error
	return results, err
}

// FindByUserIDAndDate retrieves subscription for a specific user and date
func (r *subscriptionRepositoryGORM) FindByUserIDAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*entities.DailySubscriptions, error) {
	var result entities.DailySubscriptions
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND DATE(subscription_date) = DATE(?)", userID, date).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// FindByDate retrieves all subscriptions for a specific date
func (r *subscriptionRepositoryGORM) FindByDate(ctx context.Context, date time.Time) ([]*entities.DailySubscriptions, error) {
	var results []*entities.DailySubscriptions
	err := r.db.WithContext(ctx).
		Where("DATE(subscription_date) = DATE(?)", date).
		Find(&results).Error
	return results, err
}


// FindByPaymentRef finds a subscription by payment reference
func (r *subscriptionRepositoryGORM) FindByPaymentRef(ctx context.Context, paymentRef string) (*entities.DailySubscriptions, error) {
	var result entities.DailySubscriptions
	err := r.db.WithContext(ctx).
		Where("payment_reference = ? OR payment_reference = ?", paymentRef, &paymentRef).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CountByStatus counts subscriptions by status
func (r *subscriptionRepositoryGORM) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.DailySubscriptions{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}

// FindActiveByMSISDN finds an active subscription by MSISDN
func (r *subscriptionRepositoryGORM) FindActiveByMSISDN(ctx context.Context, msisdn string) (*entities.DailySubscriptions, error) {
	var subscription entities.DailySubscriptions
	err := r.db.WithContext(ctx).
		Where("msisdn = ? AND LOWER(status) = 'active'", msisdn).
		Order("created_at DESC").
		First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}
