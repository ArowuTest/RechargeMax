package persistence

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type affiliateCommissionRepositoryGORM struct {
	db *gorm.DB
}

// NewAffiliateCommissionRepository creates a new GORM implementation
func NewAffiliateCommissionRepository(db *gorm.DB) repositories.AffiliateCommissionRepository {
	return &affiliateCommissionRepositoryGORM{db: db}
}

// Create creates a new record
func (r *affiliateCommissionRepositoryGORM) Create(ctx context.Context, entity *entities.AffiliateCommissions) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// FindByID finds a record by ID
func (r *affiliateCommissionRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.AffiliateCommissions, error) {
	var entity entities.AffiliateCommissions
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindAll retrieves all records with pagination
func (r *affiliateCommissionRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.AffiliateCommissions, error) {
	var entities []*entities.AffiliateCommissions
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&entities).Error
	return entities, err
}

// Update updates a record
func (r *affiliateCommissionRepositoryGORM) Update(ctx context.Context, entity *entities.AffiliateCommissions) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete deletes a record
func (r *affiliateCommissionRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.AffiliateCommissions{}, "id = ?", id).Error
}

// Count returns the total number of records
func (r *affiliateCommissionRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.AffiliateCommissions{}).Count(&count).Error
	return count, err
}

// FindByAffiliateID finds commissions by affiliate ID
func (r *affiliateCommissionRepositoryGORM) FindByAffiliateID(ctx context.Context, affiliateID uuid.UUID, limit, offset int) ([]*entities.AffiliateCommissions, error) {
	var commissions []*entities.AffiliateCommissions
	err := r.db.WithContext(ctx).
		Where("affiliate_id = ?", affiliateID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&commissions).Error
	return commissions, err
}

// FindByUserID finds commissions by user ID
func (r *affiliateCommissionRepositoryGORM) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.AffiliateCommissions, error) {
	var commissions []*entities.AffiliateCommissions
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&commissions).Error
	return commissions, err
}

// FindByStatus finds commissions by status
func (r *affiliateCommissionRepositoryGORM) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.AffiliateCommissions, error) {
	var commissions []*entities.AffiliateCommissions
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&commissions).Error
	return commissions, err
}

// GetTotalByAffiliateID gets total commission amount for an affiliate
func (r *affiliateCommissionRepositoryGORM) GetTotalByAffiliateID(ctx context.Context, affiliateID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&entities.AffiliateCommissions{}).
		Where("affiliate_id = ?", affiliateID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// GetPendingByAffiliateID gets pending commission amount for an affiliate
func (r *affiliateCommissionRepositoryGORM) GetPendingByAffiliateID(ctx context.Context, affiliateID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&entities.AffiliateCommissions{}).
		Where("affiliate_id = ? AND status = ?", affiliateID, "PENDING").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// SumByAffiliateIDAndStatus sums commission amounts for an affiliate by status
func (r *affiliateCommissionRepositoryGORM) SumByAffiliateIDAndStatus(ctx context.Context, affiliateID uuid.UUID, status string) (float64, error) {
	var result struct {
		Total float64
	}
	err := r.db.WithContext(ctx).
		Model(&entities.AffiliateCommissions{}).
		Select("COALESCE(SUM(commission_amount), 0) as total").
		Where("affiliate_id = ? AND status = ?", affiliateID, status).
		Scan(&result).Error
	return result.Total, err
}

// CountByAffiliateID counts total commissions for an affiliate
func (r *affiliateCommissionRepositoryGORM) CountByAffiliateID(ctx context.Context, affiliateID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.AffiliateCommissions{}).
		Where("affiliate_id = ?", affiliateID).
		Count(&count).Error
	return count, err
}
