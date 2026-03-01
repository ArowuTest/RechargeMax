package repositories

import (
	"context"
	
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create operations
	Create(ctx context.Context, user *entities.Users) error
	CreateBatch(ctx context.Context, users []*entities.Users) error
	CreateUserWithDefaults(ctx context.Context, msisdn string, referredBy *uuid.UUID) (*entities.Users, error)
	
	// Read operations
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Users, error)
	FindByMSISDN(ctx context.Context, msisdn string) (*entities.Users, error)
	FindByReferralCode(ctx context.Context, code string) (*entities.Users, error)
	FindByEmail(ctx context.Context, email string) (*entities.Users, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.Users, error)
	FindByLoyaltyTier(ctx context.Context, tier string) ([]*entities.Users, error)
	FindActiveUsers(ctx context.Context, limit, offset int) ([]*entities.Users, error)
	Count(ctx context.Context) (int64, error)
	
	// Update operations
	Update(ctx context.Context, user *entities.Users) error
	UpdateStatus(ctx context.Context, userID uuid.UUID, isActive bool) error
	UpdatePoints(ctx context.Context, userID uuid.UUID, points int) error
	UpdateLoyaltyTier(ctx context.Context, userID uuid.UUID, tier string) error
	UpdateRechargeStats(ctx context.Context, userID uuid.UUID, amount float64) error
	
	// Delete operations
	Delete(ctx context.Context, userID uuid.UUID) error
}
