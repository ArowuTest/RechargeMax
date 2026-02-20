package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// WalletTransactionRepositoryGORM implements the WalletTransactionRepository interface using GORM
type WalletTransactionRepositoryGORM struct {
	db *gorm.DB
}

// NewWalletTransactionRepository creates a new instance of WalletTransactionRepositoryGORM
func NewWalletTransactionRepository(db *gorm.DB) repositories.WalletTransactionRepository {
	return &WalletTransactionRepositoryGORM{db: db}
}

func (r *WalletTransactionRepositoryGORM) Create(ctx context.Context, entity *entities.WalletTransaction) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *WalletTransactionRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.WalletTransaction, error) {
	var entity entities.WalletTransaction
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *WalletTransactionRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.WalletTransaction, error) {
	var transactions []*entities.WalletTransaction
	err := r.db.WithContext(ctx).
		Order("processed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

func (r *WalletTransactionRepositoryGORM) Update(ctx context.Context, entity *entities.WalletTransaction) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *WalletTransactionRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.WalletTransaction{}, "id = ?", id).Error
}

func (r *WalletTransactionRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.WalletTransaction{}).Count(&count).Error
	return count, err
}

// FindByWalletID retrieves transactions for a specific wallet
func (r *WalletTransactionRepositoryGORM) FindByWalletID(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*entities.WalletTransaction, error) {
	var transactions []*entities.WalletTransaction
	err := r.db.WithContext(ctx).
		Where("wallet_id = ?", walletID).
		Order("processed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// FindByTransactionID retrieves a transaction by its transaction ID
func (r *WalletTransactionRepositoryGORM) FindByTransactionID(ctx context.Context, transactionID string) (*entities.WalletTransaction, error) {
	var transaction entities.WalletTransaction
	err := r.db.WithContext(ctx).
		Where("transaction_id = ?", transactionID).
		First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// FindByReference retrieves transactions by reference type and ID
func (r *WalletTransactionRepositoryGORM) FindByReference(ctx context.Context, referenceType, referenceID string) ([]*entities.WalletTransaction, error) {
	var transactions []*entities.WalletTransaction
	err := r.db.WithContext(ctx).
		Where("reference_type = ? AND reference_id = ?", referenceType, referenceID).
		Order("processed_at DESC").
		Find(&transactions).Error
	return transactions, err
}

// FindByDateRange retrieves transactions within a date range for a wallet
func (r *WalletTransactionRepositoryGORM) FindByDateRange(ctx context.Context, walletID uuid.UUID, startDate, endDate time.Time) ([]*entities.WalletTransaction, error) {
	var transactions []*entities.WalletTransaction
	err := r.db.WithContext(ctx).
		Where("wallet_id = ? AND processed_at BETWEEN ? AND ?", walletID, startDate, endDate).
		Order("processed_at DESC").
		Find(&transactions).Error
	return transactions, err
}

// CountByWalletID counts transactions for a specific wallet
func (r *WalletTransactionRepositoryGORM) CountByWalletID(ctx context.Context, walletID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.WalletTransaction{}).
		Where("wallet_id = ?", walletID).
		Count(&count).Error
	return count, err
}

// GetTotalByType calculates the total amount for a specific transaction type
func (r *WalletTransactionRepositoryGORM) GetTotalByType(ctx context.Context, walletID uuid.UUID, txType string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&entities.WalletTransaction{}).
		Where("wallet_id = ? AND type = ? AND status = ?", walletID, txType, "completed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}
