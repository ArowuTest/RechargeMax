package persistence

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type wheelPrizeRepositoryGORM struct {
	db *gorm.DB
}

// NewWheelPrizeRepository creates a new GORM implementation
func NewWheelPrizeRepository(db *gorm.DB) repositories.WheelPrizeRepository {
	return &wheelPrizeRepositoryGORM{db: db}
}

// Create creates a new record
func (r *wheelPrizeRepositoryGORM) Create(ctx context.Context, entity *entities.WheelPrizes) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// FindByID finds a record by ID
func (r *wheelPrizeRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.WheelPrizes, error) {
	var entity entities.WheelPrizes
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindAll retrieves all records with pagination
func (r *wheelPrizeRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.WheelPrizes, error) {
	var entities []*entities.WheelPrizes
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("sort_order ASC").
		Find(&entities).Error
	return entities, err
}

// Update updates a record
func (r *wheelPrizeRepositoryGORM) Update(ctx context.Context, entity *entities.WheelPrizes) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete deletes a record
func (r *wheelPrizeRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.WheelPrizes{}, "id = ?", id).Error
}

// Count returns the total number of records
func (r *wheelPrizeRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.WheelPrizes{}).Count(&count).Error
	return count, err
}

// FindActive finds all active wheel prizes
func (r *wheelPrizeRepositoryGORM) FindActive(ctx context.Context) ([]*entities.WheelPrizes, error) {
	var prizes []*entities.WheelPrizes
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC").
		Find(&prizes).Error
	return prizes, err
}

// FindByProbabilityRange finds prizes within a probability range
func (r *wheelPrizeRepositoryGORM) FindByProbabilityRange(ctx context.Context, minProb, maxProb float64) ([]*entities.WheelPrizes, error) {
	var prizes []*entities.WheelPrizes
	err := r.db.WithContext(ctx).
		Where("probability >= ? AND probability <= ?", minProb, maxProb).
		Order("probability DESC").
		Find(&prizes).Error
	return prizes, err
}


// FindByStatus finds prizes by status (active, inactive, etc.)
func (r *wheelPrizeRepositoryGORM) FindByStatus(ctx context.Context, status string) ([]*entities.WheelPrizes, error) {
	var prizes []*entities.WheelPrizes
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&prizes).Error
	return prizes, err
}
