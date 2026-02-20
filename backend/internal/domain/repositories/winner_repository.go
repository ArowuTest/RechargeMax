package repositories

import (
	"context"
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

type WinnerRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.Winner) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Winner, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.Winner, error)
	Update(ctx context.Context, entity *entities.Winner) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByDrawID(ctx context.Context, drawID uuid.UUID) ([]*entities.Winner, error)
	FindByMSISDN(ctx context.Context, msisdn string, limit, offset int) ([]*entities.Winner, error)
	FindByClaimStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Winner, error)
	FindByPayoutStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Winner, error)
	FindByProvisionStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Winner, error)
	FindPendingProvisioning(ctx context.Context) ([]*entities.Winner, error)
	FindExpiredClaims(ctx context.Context) ([]*entities.Winner, error)
	CountByMSISDN(ctx context.Context, msisdn string) (int64, error)
	UpdateClaimStatus(ctx context.Context, winnerID uuid.UUID, status string) error
	UpdatePayoutStatus(ctx context.Context, winnerID uuid.UUID, status string) error
	UpdateProvisionStatus(ctx context.Context, winnerID uuid.UUID, status string) error
	
	// User-specific queries
	FindUnclaimedByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Winner, error)
	CountUnclaimedByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	FindUnclaimedBeforeDeadline(ctx context.Context, deadline string) ([]*entities.Winner, error)
}
