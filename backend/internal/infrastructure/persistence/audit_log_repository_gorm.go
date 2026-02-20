package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// AuditLogRepositoryGORM implements the AuditLogRepository interface using GORM
type AuditLogRepositoryGORM struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new instance of AuditLogRepositoryGORM
func NewAuditLogRepository(db *gorm.DB) repositories.AuditLogRepository {
	return &AuditLogRepositoryGORM{db: db}
}

func (r *AuditLogRepositoryGORM) Create(ctx context.Context, entity *entities.AuditLog) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *AuditLogRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.AuditLog, error) {
	var entity entities.AuditLog
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *AuditLogRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.AuditLog, error) {
	var logs []*entities.AuditLog
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

func (r *AuditLogRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.AuditLog{}).Count(&count).Error
	return count, err
}

// FindByAdminUserID retrieves audit logs for a specific admin user
func (r *AuditLogRepositoryGORM) FindByAdminUserID(ctx context.Context, adminUserID uuid.UUID, limit, offset int) ([]*entities.AuditLog, error) {
	var logs []*entities.AuditLog
	err := r.db.WithContext(ctx).
		Where("admin_user_id = ?", adminUserID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// FindByEntity retrieves audit logs for a specific entity
func (r *AuditLogRepositoryGORM) FindByEntity(ctx context.Context, entityType, entityID string, limit, offset int) ([]*entities.AuditLog, error) {
	var logs []*entities.AuditLog
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// FindByAction retrieves audit logs for a specific action
func (r *AuditLogRepositoryGORM) FindByAction(ctx context.Context, action string, limit, offset int) ([]*entities.AuditLog, error) {
	var logs []*entities.AuditLog
	err := r.db.WithContext(ctx).
		Where("action = ?", action).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// FindByDateRange retrieves audit logs within a date range
func (r *AuditLogRepositoryGORM) FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entities.AuditLog, error) {
	var logs []*entities.AuditLog
	err := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// FindRecentByAdmin retrieves recent audit logs for an admin
func (r *AuditLogRepositoryGORM) FindRecentByAdmin(ctx context.Context, adminUserID uuid.UUID, limit int) ([]*entities.AuditLog, error) {
	var logs []*entities.AuditLog
	err := r.db.WithContext(ctx).
		Where("admin_user_id = ?", adminUserID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// CountByAdminUserID counts audit logs for a specific admin
func (r *AuditLogRepositoryGORM) CountByAdminUserID(ctx context.Context, adminUserID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.AuditLog{}).
		Where("admin_user_id = ?", adminUserID).
		Count(&count).Error
	return count, err
}

// CountByAction counts audit logs for a specific action
func (r *AuditLogRepositoryGORM) CountByAction(ctx context.Context, action string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.AuditLog{}).
		Where("action = ?", action).
		Count(&count).Error
	return count, err
}

// DeleteOldLogs deletes audit logs older than specified days
func (r *AuditLogRepositoryGORM) DeleteOldLogs(ctx context.Context, daysOld int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -daysOld)
	result := r.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&entities.AuditLog{})
	return result.RowsAffected, result.Error
}
