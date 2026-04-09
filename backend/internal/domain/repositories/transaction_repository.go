package repositories

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// TransactionRepository defines the interface for transaction data operations
type TransactionRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.Transactions) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Transactions, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.Transactions, error)
	Update(ctx context.Context, entity *entities.Transactions) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business queries
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Transactions, error)
	FindByReference(ctx context.Context, reference string) (*entities.Transactions, error)
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Transactions, error)
	GetTotalRevenue(ctx context.Context) (float64, error)
	GetRevenueByDate(ctx context.Context, date time.Time) (float64, error)
	CountPendingWithdrawals(ctx context.Context) (int, error)
	CountActiveSubscriptions(ctx context.Context) (int, error)
	CountSpinsByDate(ctx context.Context, date time.Time) (int, error)
	// CountEligibleForSpin counts transactions eligible for spin (amount >= 1000, no spin yet)
	CountEligibleForSpin(ctx context.Context, userID uuid.UUID) (int64, error)
	// FindEligibleForSpin finds a transaction eligible for spin
	FindEligibleForSpin(ctx context.Context, userID uuid.UUID) (*entities.Transactions, error)
	// CountByUserID counts transactions for a specific user
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	// CountByMSISDN counts successful transactions for a given MSISDN (used as
	// a user_id-independent fallback when user_id is NULL on legacy rows)
	CountByMSISDN(ctx context.Context, msisdn string) (int64, error)
	// SumSuccessfulAmountByMSISDNSince returns the total kobo amount of all
	// SUCCESS transactions for a given MSISDN on or after the given timestamp.
	// Used by CheckEligibility to determine the user's cumulative daily recharge
	// total without bypassing the repository layer.
	SumSuccessfulAmountByMSISDNSince(ctx context.Context, msisdn string, since time.Time) (int64, error)
	
	// Recharge-specific methods (Transactions are recharges)
	FindByPaymentRef(ctx context.Context, paymentRef string) (*entities.Transactions, error)
	FindByMSISDN(ctx context.Context, msisdn string, limit, offset int) ([]*entities.Transactions, error)
	
	// Fraud detection and analytics methods
	FindByPaymentReference(ctx context.Context, reference string) (*entities.Transactions, error)
	CountByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error)
	SumAmountByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (float64, error)
}
