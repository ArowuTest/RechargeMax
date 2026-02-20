package repositories

import (
	"context"
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

type NetworkCacheRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.NetworkCache) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.NetworkCache, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.NetworkCache, error)
	Update(ctx context.Context, entity *entities.NetworkCache) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByMSISDN(ctx context.Context, msisdn string) (*entities.NetworkCache, error)
	FindValidCache(ctx context.Context, msisdn string) (*entities.NetworkCache, error)
	Invalidate(ctx context.Context, msisdn string, reason string) error
	DeleteExpired(ctx context.Context) (int64, error)
	CountByNetwork(ctx context.Context, network string) (int64, error)
	CountByLookupSource(ctx context.Context, source string) (int64, error)
}
