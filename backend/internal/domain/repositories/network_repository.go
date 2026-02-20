package repositories

import (
	"context"
	
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// NetworkRepository defines the interface for network data operations
type NetworkRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.NetworkConfigs) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.NetworkConfigs, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.NetworkConfigs, error)
	Update(ctx context.Context, entity *entities.NetworkConfigs) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByCode(ctx context.Context, code string) (*entities.NetworkConfigs, error)
}
