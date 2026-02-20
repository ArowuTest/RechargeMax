package repositories

import (
	"context"
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

type WalletRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.Wallet) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Wallet, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.Wallet, error)
	Update(ctx context.Context, entity *entities.Wallet) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByMSISDN(ctx context.Context, msisdn string) (*entities.Wallet, error)
	GetOrCreate(ctx context.Context, msisdn string) (*entities.Wallet, error)
	UpdateBalance(ctx context.Context, walletID uuid.UUID, balanceDelta int64) error
	UpdatePendingBalance(ctx context.Context, walletID uuid.UUID, pendingDelta int64) error
	ReleasePendingToAvailable(ctx context.Context, walletID uuid.UUID, amount int64) error
	GetTotalBalances(ctx context.Context) (totalBalance int64, totalPending int64, err error)
	FindActiveWallets(ctx context.Context, limit, offset int) ([]*entities.Wallet, error)
	FindSuspendedWallets(ctx context.Context, limit, offset int) ([]*entities.Wallet, error)
}
