package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type pointsAdjustmentRepositoryGorm struct {
	db *gorm.DB
}

// NewPointsAdjustmentRepository creates a new points adjustment repository
func NewPointsAdjustmentRepository(db *gorm.DB) repositories.PointsAdjustmentRepository {
	return &pointsAdjustmentRepositoryGorm{db: db}
}

// Create creates a new points adjustment record
func (r *pointsAdjustmentRepositoryGorm) Create(ctx context.Context, adjustment *entities.PointsAdjustment) error {
	return r.db.WithContext(ctx).Create(adjustment).Error
}

// FindByID finds a points adjustment by ID
func (r *pointsAdjustmentRepositoryGorm) FindByID(ctx context.Context, id uuid.UUID) (*entities.PointsAdjustment, error) {
	var adjustment entities.PointsAdjustment
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("AdminUser").
		Where("id = ?", id).
		First(&adjustment).Error
	if err != nil {
		return nil, err
	}
	return &adjustment, nil
}

// FindByUserID finds all points adjustments for a user
func (r *pointsAdjustmentRepositoryGorm) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.PointsAdjustment, error) {
	var adjustments []*entities.PointsAdjustment
	err := r.db.WithContext(ctx).
		Preload("AdminUser").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&adjustments).Error
	return adjustments, err
}

// FindByAdminID finds all points adjustments made by an admin
func (r *pointsAdjustmentRepositoryGorm) FindByAdminID(ctx context.Context, adminID uuid.UUID) ([]*entities.PointsAdjustment, error) {
	var adjustments []*entities.PointsAdjustment
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("admin_id = ?", adminID).
		Order("created_at DESC").
		Find(&adjustments).Error
	return adjustments, err
}

// FindAll finds all points adjustments within a date range
func (r *pointsAdjustmentRepositoryGorm) FindAll(ctx context.Context, startDate, endDate time.Time) ([]*entities.PointsAdjustment, error) {
	var adjustments []*entities.PointsAdjustment
	query := r.db.WithContext(ctx).
		Preload("User").
		Preload("AdminUser")

	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	err := query.Order("created_at DESC").Find(&adjustments).Error
	return adjustments, err
}
