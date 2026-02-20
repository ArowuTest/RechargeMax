package persistence

import (
	"context"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type spinRepositoryGORM struct {
	db *gorm.DB
}

// NewSpinRepository creates a new GORM implementation
func NewSpinRepository(db *gorm.DB) repositories.SpinRepository {
	return &spinRepositoryGORM{db: db}
}

// Create creates a new record
func (r *spinRepositoryGORM) Create(ctx context.Context, entity *entities.SpinResults) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// FindByID finds a record by ID
func (r *spinRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.SpinResults, error) {
	var entity entities.SpinResults
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindAll retrieves all records with pagination
func (r *spinRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.SpinResults, error) {
	var entities []*entities.SpinResults
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&entities).Error
	return entities, err
}

// Update updates a record
func (r *spinRepositoryGORM) Update(ctx context.Context, entity *entities.SpinResults) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete deletes a record
func (r *spinRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.SpinResults{}, "id = ?", id).Error
}

// Count returns the total number of records
func (r *spinRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.SpinResults{}).Count(&count).Error
	return count, err
}

// FindByUserID retrieves spin results for a specific user with pagination
func (r *spinRepositoryGORM) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.SpinResults, error) {
	var results []*entities.SpinResults
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&results).Error
	return results, err
}

// CountByUserID counts spin results for a specific user
func (r *spinRepositoryGORM) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.SpinResults{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
