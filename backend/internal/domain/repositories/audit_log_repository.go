package repositories

import (
	"context"
	"time"
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

type AuditLogRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.AuditLog) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.AuditLog, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.AuditLog, error)
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByAdminUserID(ctx context.Context, adminUserID uuid.UUID, limit, offset int) ([]*entities.AuditLog, error)
	FindByEntity(ctx context.Context, entityType, entityID string, limit, offset int) ([]*entities.AuditLog, error)
	FindByAction(ctx context.Context, action string, limit, offset int) ([]*entities.AuditLog, error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entities.AuditLog, error)
	FindRecentByAdmin(ctx context.Context, adminUserID uuid.UUID, limit int) ([]*entities.AuditLog, error)
	CountByAdminUserID(ctx context.Context, adminUserID uuid.UUID) (int64, error)
	CountByAction(ctx context.Context, action string) (int64, error)
	DeleteOldLogs(ctx context.Context, daysOld int) (int64, error)
}
