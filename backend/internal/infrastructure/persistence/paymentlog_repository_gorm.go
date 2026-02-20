package persistence

import (
	"context"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type paymentlogRepositoryGORM struct {
	db *gorm.DB
}

// NewPaymentLogRepository creates a new GORM implementation
func NewPaymentLogRepository(db *gorm.DB) repositories.PaymentLogRepository {
	return &paymentlogRepositoryGORM{db: db}
}

// Create creates a new record
func (r *paymentlogRepositoryGORM) Create(ctx context.Context, entity *entities.PaymentLogs) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// FindByID finds a record by ID
func (r *paymentlogRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.PaymentLogs, error) {
	var entity entities.PaymentLogs
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// GetByReference finds a payment log by reference
func (r *paymentlogRepositoryGORM) GetByReference(ctx context.Context, reference string) (*entities.PaymentLogs, error) {
	var entity entities.PaymentLogs
	err := r.db.WithContext(ctx).Where("payment_reference = ?", reference).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindAll retrieves all records with pagination
func (r *paymentlogRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.PaymentLogs, error) {
	var entities []*entities.PaymentLogs
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&entities).Error
	return entities, err
}

// Update updates a record
func (r *paymentlogRepositoryGORM) Update(ctx context.Context, entity *entities.PaymentLogs) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete deletes a record
func (r *paymentlogRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.PaymentLogs{}, "id = ?", id).Error
}

// Count returns the total number of records
func (r *paymentlogRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.PaymentLogs{}).Count(&count).Error
	return count, err
}
