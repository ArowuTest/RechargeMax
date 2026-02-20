package repositories

import (
	"context"
	"rechargemax/internal/domain/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DataPlanRepository handles data plan database operations
type DataPlanRepository interface {
	FindByNetwork(ctx context.Context, networkID uuid.UUID) ([]entities.DataPlans, error)
	FindByNetworkCode(ctx context.Context, networkCode string) ([]entities.DataPlans, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entities.DataPlans, error)
	FindAll(ctx context.Context) ([]entities.DataPlans, error)
	Create(ctx context.Context, plan *entities.DataPlans) error
	Update(ctx context.Context, plan *entities.DataPlans) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// DataPlanRepositoryImpl implements DataPlanRepository
type DataPlanRepositoryImpl struct {
	db *gorm.DB
}

// NewDataPlanRepository creates a new data plan repository
func NewDataPlanRepository(db *gorm.DB) DataPlanRepository {
	return &DataPlanRepositoryImpl{db: db}
}

// FindByNetwork finds all active data plans for a specific network
func (r *DataPlanRepositoryImpl) FindByNetwork(ctx context.Context, networkID uuid.UUID) ([]entities.DataPlans, error) {
	var plans []entities.DataPlans
	err := r.db.WithContext(ctx).
		Where("network_id = ? AND is_active = ?", networkID, true).
		Order("sort_order ASC, price ASC").
		Find(&plans).Error
	return plans, err
}

// FindByNetworkCode finds all active data plans for a network by its code
func (r *DataPlanRepositoryImpl) FindByNetworkCode(ctx context.Context, networkCode string) ([]entities.DataPlans, error) {
	var plans []entities.DataPlans
	
		// Query data plans directly using network_provider field
	err := r.db.WithContext(ctx).
		Where("network_provider = ? AND is_active = ?", networkCode, true).
		Order("sort_order ASC, price ASC").
		Find(&plans).Error
	
	return plans, err
}

// FindByID finds a data plan by ID
func (r *DataPlanRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entities.DataPlans, error) {
	var plan entities.DataPlans
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// FindAll finds all data plans
func (r *DataPlanRepositoryImpl) FindAll(ctx context.Context) ([]entities.DataPlans, error) {
	var plans []entities.DataPlans
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, price ASC").
		Find(&plans).Error
	return plans, err
}

// Create creates a new data plan
func (r *DataPlanRepositoryImpl) Create(ctx context.Context, plan *entities.DataPlans) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

// Update updates a data plan
func (r *DataPlanRepositoryImpl) Update(ctx context.Context, plan *entities.DataPlans) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

// Delete deletes a data plan (soft delete by setting is_active to false)
func (r *DataPlanRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entities.DataPlans{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}
