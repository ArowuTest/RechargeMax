package repositories

import (
	"context"
	"time"
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

type WalletTransactionRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.WalletTransaction) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.WalletTransaction, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.WalletTransaction, error)
	Update(ctx context.Context, entity *entities.WalletTransaction) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByWalletID(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*entities.WalletTransaction, error)
	FindByTransactionID(ctx context.Context, transactionID string) (*entities.WalletTransaction, error)
	FindByReference(ctx context.Context, referenceType, referenceID string) ([]*entities.WalletTransaction, error)
	FindByDateRange(ctx context.Context, walletID uuid.UUID, startDate, endDate time.Time) ([]*entities.WalletTransaction, error)
	CountByWalletID(ctx context.Context, walletID uuid.UUID) (int64, error)
	GetTotalByType(ctx context.Context, walletID uuid.UUID, txType string) (int64, error)
}
