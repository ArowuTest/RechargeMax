package repositories

import (
"context"
"github.com/google/uuid"
"rechargemax/internal/domain/entities"
)

type WithdrawalRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.WithdrawalRequests) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.WithdrawalRequests, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.WithdrawalRequests, error)
	Update(ctx context.Context, entity *entities.WithdrawalRequests) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.WithdrawalRequests, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}
