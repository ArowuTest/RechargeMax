package repositories

import (
	"context"
	
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// AffiliateRepository defines the interface for affiliate data operations
type AffiliateRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.Affiliates) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Affiliates, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.Affiliates, error)
	Update(ctx context.Context, entity *entities.Affiliates) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)

	// Business operations
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entities.Affiliates, error)
	FindByAffiliateCode(ctx context.Context, code string) (*entities.Affiliates, error)
	FindByMSISDN(ctx context.Context, msisdn string) (*entities.Affiliates, error)
	FindByEmail(ctx context.Context, email string) (*entities.Affiliates, error)
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Affiliates, error)
}
