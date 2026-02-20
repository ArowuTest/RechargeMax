package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// PointsAdjustmentRepository defines operations for points adjustments
type PointsAdjustmentRepository interface {
	Create(ctx context.Context, adjustment *entities.PointsAdjustment) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.PointsAdjustment, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.PointsAdjustment, error)
	FindByAdminID(ctx context.Context, adminID uuid.UUID) ([]*entities.PointsAdjustment, error)
	FindAll(ctx context.Context, startDate, endDate time.Time) ([]*entities.PointsAdjustment, error)
}
