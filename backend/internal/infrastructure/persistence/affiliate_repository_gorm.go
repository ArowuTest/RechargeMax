package persistence

import (
	"context"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type affiliateRepositoryGORM struct {
	db *gorm.DB
}

// NewAffiliateRepository creates a new GORM implementation
func NewAffiliateRepository(db *gorm.DB) repositories.AffiliateRepository {
	return &affiliateRepositoryGORM{db: db}
}

// Create creates a new record
func (r *affiliateRepositoryGORM) Create(ctx context.Context, entity *entities.Affiliates) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// FindByID finds a record by ID
func (r *affiliateRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.Affiliates, error) {
	var entity entities.Affiliates
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindAll retrieves all records with pagination
func (r *affiliateRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.Affiliates, error) {
	var entities []*entities.Affiliates
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&entities).Error
	return entities, err
}

// Update updates a record
func (r *affiliateRepositoryGORM) Update(ctx context.Context, entity *entities.Affiliates) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete deletes a record
func (r *affiliateRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Affiliates{}, "id = ?", id).Error
}

// Count returns the total number of records
func (r *affiliateRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Affiliates{}).Count(&count).Error
	return count, err
}

// FindByUserID finds an affiliate by user ID
func (r *affiliateRepositoryGORM) FindByUserID(ctx context.Context, userID uuid.UUID) (*entities.Affiliates, error) {
	var affiliate entities.Affiliates
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&affiliate).Error
	if err != nil {
		return nil, err
	}
	return &affiliate, nil
}

// FindByAffiliateCode finds an affiliate by affiliate code
func (r *affiliateRepositoryGORM) FindByAffiliateCode(ctx context.Context, code string) (*entities.Affiliates, error) {
	var affiliate entities.Affiliates
	err := r.db.WithContext(ctx).
		Where("affiliate_code = ?", code).
		First(&affiliate).Error
	if err != nil {
		return nil, err
	}
	return &affiliate, nil
}

// FindByMSISDN finds an affiliate by MSISDN (via user relationship)
func (r *affiliateRepositoryGORM) FindByMSISDN(ctx context.Context, msisdn string) (*entities.Affiliates, error) {
	var affiliate entities.Affiliates
	err := r.db.WithContext(ctx).
		Joins("JOIN users_2026_01_30_14_00 ON users_2026_01_30_14_00.id = affiliates_2026_01_30_14_00.user_id").
		Where("users_2026_01_30_14_00.msisdn = ?", msisdn).
		First(&affiliate).Error
	if err != nil {
		return nil, err
	}
	return &affiliate, nil
}

// FindByEmail finds an affiliate by email (via user relationship)
func (r *affiliateRepositoryGORM) FindByEmail(ctx context.Context, email string) (*entities.Affiliates, error) {
	var affiliate entities.Affiliates
	err := r.db.WithContext(ctx).
		Joins("JOIN users_2026_01_30_14_00 ON users_2026_01_30_14_00.id = affiliates_2026_01_30_14_00.user_id").
		Where("users_2026_01_30_14_00.email = ?", email).
		First(&affiliate).Error
	if err != nil {
		return nil, err
	}
	return &affiliate, nil
}

// FindByStatus finds affiliates by status
func (r *affiliateRepositoryGORM) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Affiliates, error) {
	var affiliates []*entities.Affiliates
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&affiliates).Error
	return affiliates, err
}
