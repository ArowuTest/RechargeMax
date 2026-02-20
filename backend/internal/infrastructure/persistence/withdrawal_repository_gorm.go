package persistence

import (
"context"

"github.com/google/uuid"
"gorm.io/gorm"
"rechargemax/internal/domain/entities"
"rechargemax/internal/domain/repositories"
)

// WithdrawalRepositoryGORM implements the WithdrawalRepository interface using GORM
type WithdrawalRepositoryGORM struct {
db *gorm.DB
}

// NewWithdrawalRepository creates a new instance of WithdrawalRepositoryGORM
func NewWithdrawalRepository(db *gorm.DB) repositories.WithdrawalRepository {
return &WithdrawalRepositoryGORM{db: db}
}

func (r *WithdrawalRepositoryGORM) Create(ctx context.Context, entity *entities.WithdrawalRequests) error {
return r.db.WithContext(ctx).Create(entity).Error
}

func (r *WithdrawalRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.WithdrawalRequests, error) {
var entity entities.WithdrawalRequests
err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
if err != nil {
return nil, err
}
return &entity, nil
}

func (r *WithdrawalRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.WithdrawalRequests, error) {
var entities []*entities.WithdrawalRequests
err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&entities).Error
return entities, err
}

func (r *WithdrawalRepositoryGORM) Update(ctx context.Context, entity *entities.WithdrawalRequests) error {
return r.db.WithContext(ctx).Save(entity).Error
}

func (r *WithdrawalRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
return r.db.WithContext(ctx).Delete(&entities.WithdrawalRequests{}, "id = ?", id).Error
}

func (r *WithdrawalRepositoryGORM) Count(ctx context.Context) (int64, error) {
var count int64
err := r.db.WithContext(ctx).Model(&entities.WithdrawalRequests{}).Count(&count).Error
return count, err
}

// FindByUserID retrieves all withdrawal requests for a specific user with pagination
func (r *WithdrawalRepositoryGORM) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.WithdrawalRequests, error) {
	var withdrawals []*entities.WithdrawalRequests
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("requested_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&withdrawals).Error
	return withdrawals, err
}

// CountByUserID counts withdrawal requests for a specific user
func (r *WithdrawalRepositoryGORM) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.WithdrawalRequests{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
