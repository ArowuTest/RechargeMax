package repositories

import (
	"context"

	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// AffiliateCommissionRepository defines the interface for affiliate commission data operations
type AffiliateCommissionRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.AffiliateCommissions) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.AffiliateCommissions, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.AffiliateCommissions, error)
	Update(ctx context.Context, entity *entities.AffiliateCommissions) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)

	// Business operations
	FindByAffiliateID(ctx context.Context, affiliateID uuid.UUID, limit, offset int) ([]*entities.AffiliateCommissions, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.AffiliateCommissions, error)
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.AffiliateCommissions, error)
	GetTotalByAffiliateID(ctx context.Context, affiliateID uuid.UUID) (float64, error)
	GetPendingByAffiliateID(ctx context.Context, affiliateID uuid.UUID) (float64, error)
	
	// Analytics and aggregation methods
	SumByAffiliateIDAndStatus(ctx context.Context, affiliateID uuid.UUID, status string) (float64, error)
	CountByAffiliateID(ctx context.Context, affiliateID uuid.UUID) (int64, error)
}
