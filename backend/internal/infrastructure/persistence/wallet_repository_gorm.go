package persistence

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// WalletRepositoryGORM implements the WalletRepository interface using GORM
type WalletRepositoryGORM struct {
	db *gorm.DB
}

// NewWalletRepository creates a new instance of WalletRepositoryGORM
func NewWalletRepository(db *gorm.DB) repositories.WalletRepository {
	return &WalletRepositoryGORM{db: db}
}

func (r *WalletRepositoryGORM) Create(ctx context.Context, entity *entities.Wallet) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *WalletRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.Wallet, error) {
	var entity entities.Wallet
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *WalletRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.Wallet, error) {
	var wallets []*entities.Wallet
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&wallets).Error
	return wallets, err
}

func (r *WalletRepositoryGORM) Update(ctx context.Context, entity *entities.Wallet) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *WalletRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Wallet{}, "id = ?", id).Error
}

func (r *WalletRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Wallet{}).Count(&count).Error
	return count, err
}

// FindByMSISDN retrieves a wallet by MSISDN
func (r *WalletRepositoryGORM) FindByMSISDN(ctx context.Context, msisdn string) (*entities.Wallet, error) {
	var wallet entities.Wallet
	err := r.db.WithContext(ctx).
		Where("msisdn = ?", msisdn).
		First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// GetOrCreate retrieves an existing wallet or creates a new one
func (r *WalletRepositoryGORM) GetOrCreate(ctx context.Context, msisdn string) (*entities.Wallet, error) {
	wallet, err := r.FindByMSISDN(ctx, msisdn)
	if err == nil {
		return wallet, nil
	}
	
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	
	// Create new wallet
	wallet = &entities.Wallet{
		MSISDN:          msisdn,
		Balance:         0,
		PendingBalance:  0,
		TotalEarned:     0,
		TotalWithdrawn:  0,
		MinPayoutAmount: 100000, // ₦1000 default
		IsActive:        true,
		IsSuspended:     false,
	}
	
	err = r.Create(ctx, wallet)
	if err != nil {
		return nil, err
	}
	
	return wallet, nil
}

// UpdateBalance updates the available balance
func (r *WalletRepositoryGORM) UpdateBalance(ctx context.Context, walletID uuid.UUID, balanceDelta int64) error {
	return r.db.WithContext(ctx).
		Model(&entities.Wallet{}).
		Where("id = ?", walletID).
		UpdateColumn("balance", gorm.Expr("balance + ?", balanceDelta)).
		Error
}

// UpdatePendingBalance updates the pending balance
func (r *WalletRepositoryGORM) UpdatePendingBalance(ctx context.Context, walletID uuid.UUID, pendingDelta int64) error {
	return r.db.WithContext(ctx).
		Model(&entities.Wallet{}).
		Where("id = ?", walletID).
		UpdateColumn("pending_balance", gorm.Expr("pending_balance + ?", pendingDelta)).
		Error
}

// ReleasePendingToAvailable moves amount from pending to available balance
func (r *WalletRepositoryGORM) ReleasePendingToAvailable(ctx context.Context, walletID uuid.UUID, amount int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Decrease pending balance
		err := tx.Model(&entities.Wallet{}).
			Where("id = ? AND pending_balance >= ?", walletID, amount).
			UpdateColumn("pending_balance", gorm.Expr("pending_balance - ?", amount)).
			Error
		if err != nil {
			return err
		}
		
		// Increase available balance
		err = tx.Model(&entities.Wallet{}).
			Where("id = ?", walletID).
			UpdateColumn("balance", gorm.Expr("balance + ?", amount)).
			Error
		if err != nil {
			return err
		}
		
		return nil
	})
}

// GetTotalBalances returns the sum of all balances and pending balances
func (r *WalletRepositoryGORM) GetTotalBalances(ctx context.Context) (totalBalance int64, totalPending int64, err error) {
	type Result struct {
		TotalBalance int64
		TotalPending int64
	}
	
	var result Result
	err = r.db.WithContext(ctx).
		Model(&entities.Wallet{}).
		Select("SUM(balance) as total_balance, SUM(pending_balance) as total_pending").
		Where("is_active = ? AND is_suspended = ?", true, false).
		Scan(&result).Error
	
	return result.TotalBalance, result.TotalPending, err
}

// FindActiveWallets retrieves all active wallets
func (r *WalletRepositoryGORM) FindActiveWallets(ctx context.Context, limit, offset int) ([]*entities.Wallet, error) {
	var wallets []*entities.Wallet
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND is_suspended = ?", true, false).
		Order("balance DESC").
		Limit(limit).
		Offset(offset).
		Find(&wallets).Error
	return wallets, err
}

// FindSuspendedWallets retrieves all suspended wallets
func (r *WalletRepositoryGORM) FindSuspendedWallets(ctx context.Context, limit, offset int) ([]*entities.Wallet, error) {
	var wallets []*entities.Wallet
	err := r.db.WithContext(ctx).
		Where("is_suspended = ?", true).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&wallets).Error
	return wallets, err
}
