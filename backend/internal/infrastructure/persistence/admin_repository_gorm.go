package persistence

import (
	"context"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type adminRepositoryGORM struct {
	db *gorm.DB
}

// NewAdminRepository creates a new GORM implementation
func NewAdminRepository(db *gorm.DB) repositories.AdminRepository {
	return &adminRepositoryGORM{db: db}
}

// Create creates a new record
func (r *adminRepositoryGORM) Create(ctx context.Context, entity *entities.AdminUsers) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// FindByID finds a record by ID
func (r *adminRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.AdminUsers, error) {
	var entity entities.AdminUsers
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindAll retrieves all records with pagination
func (r *adminRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.AdminUsers, error) {
	var entities []*entities.AdminUsers
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&entities).Error
	return entities, err
}

// Update updates a record
func (r *adminRepositoryGORM) Update(ctx context.Context, entity *entities.AdminUsers) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete deletes a record
func (r *adminRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.AdminUsers{}, "id = ?", id).Error
}

// Count returns the total number of records
func (r *adminRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.AdminUsers{}).Count(&count).Error
	return count, err
}

// GetByEmail finds an admin by email
func (r *adminRepositoryGORM) GetByEmail(ctx context.Context, email string) (*entities.AdminUsers, error) {
	var entity entities.AdminUsers
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// GetByID finds an admin by ID (string version for JWT claims)
func (r *adminRepositoryGORM) GetByID(ctx context.Context, id string) (*entities.AdminUsers, error) {
	var entity entities.AdminUsers
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// UpdateLastLogin updates the last login timestamp
func (r *adminRepositoryGORM) UpdateLastLogin(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&entities.AdminUsers{}).
		Where("id = ?", id).
		Update("last_login_at", gorm.Expr("NOW()")).
		Error
}
