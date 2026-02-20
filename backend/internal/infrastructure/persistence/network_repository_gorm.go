package persistence

import (
	"context"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type networkRepositoryGORM struct {
	db *gorm.DB
}

// NewNetworkRepository creates a new GORM implementation
func NewNetworkRepository(db *gorm.DB) repositories.NetworkRepository {
	return &networkRepositoryGORM{db: db}
}

// Create creates a new record
func (r *networkRepositoryGORM) Create(ctx context.Context, entity *entities.NetworkConfigs) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// FindByID finds a record by ID
func (r *networkRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.NetworkConfigs, error) {
	var entity entities.NetworkConfigs
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindAll retrieves all records with pagination
func (r *networkRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.NetworkConfigs, error) {
	var entities []*entities.NetworkConfigs
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&entities).Error
	return entities, err
}

// Update updates a record
func (r *networkRepositoryGORM) Update(ctx context.Context, entity *entities.NetworkConfigs) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete deletes a record
func (r *networkRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.NetworkConfigs{}, "id = ?", id).Error
}

// Count returns the total number of records
func (r *networkRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.NetworkConfigs{}).Count(&count).Error
	return count, err
}

// FindByCode finds a network configuration by network code (e.g., "MTN", "GLO")
func (r *networkRepositoryGORM) FindByCode(ctx context.Context, code string) (*entities.NetworkConfigs, error) {
	var network entities.NetworkConfigs
	err := r.db.WithContext(ctx).
		Where("network_code = ? OR network_name = ?", code, code).
		First(&network).Error
	if err != nil {
		return nil, err
	}
	return &network, nil
}
