package repositories

import (
	"context"
	"time"

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
	CountPendingByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	// CountTodayByMSISDN counts all spin_results rows for a given MSISDN on or
	// after the given timestamp (typically today's UTC midnight). Used by
	// CheckEligibility to determine how many spins the user has already played
	// today, regardless of claim_status (PENDING means spin was played, prize
	// is awaiting collection — it must NOT be treated as a fresh spin slot).
	CountTodayByMSISDN(ctx context.Context, msisdn string, since time.Time) (int64, error)
}
