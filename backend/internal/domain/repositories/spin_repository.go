package repositories

import (
	"context"
	
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// SpinRepository defines the interface for spin data operations
type SpinRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.SpinResults) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.SpinResults, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.SpinResults, error)
	Update(ctx context.Context, entity *entities.SpinResults) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.SpinResults, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}
