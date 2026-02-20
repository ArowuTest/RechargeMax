package repositories

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// SubscriptionRepository defines the interface for subscription data operations
type SubscriptionRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.DailySubscriptions) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.DailySubscriptions, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.DailySubscriptions, error)
	Update(ctx context.Context, entity *entities.DailySubscriptions) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.DailySubscriptions, error)
	FindByUserIDAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*entities.DailySubscriptions, error)
	FindByDate(ctx context.Context, date time.Time) ([]*entities.DailySubscriptions, error)
	FindByPaymentRef(ctx context.Context, paymentRef string) (*entities.DailySubscriptions, error)
	
	// Analytics and filtering methods
	CountByStatus(ctx context.Context, status string) (int64, error)
	FindActiveByMSISDN(ctx context.Context, msisdn string) (*entities.DailySubscriptions, error)
}
