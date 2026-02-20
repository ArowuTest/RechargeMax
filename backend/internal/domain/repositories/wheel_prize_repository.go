package repositories

import (
	"context"

	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// WheelPrizeRepository defines the interface for wheel prize data operations
type WheelPrizeRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.WheelPrizes) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.WheelPrizes, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.WheelPrizes, error)
	Update(ctx context.Context, entity *entities.WheelPrizes) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)

	// Business operations
	FindActive(ctx context.Context) ([]*entities.WheelPrizes, error)
	FindByProbabilityRange(ctx context.Context, minProb, maxProb float64) ([]*entities.WheelPrizes, error)
	FindByStatus(ctx context.Context, status string) ([]*entities.WheelPrizes, error)
}
