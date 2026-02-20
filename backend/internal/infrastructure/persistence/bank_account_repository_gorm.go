package persistence

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// BankAccountRepositoryGORM implements the BankAccountRepository interface using GORM
type BankAccountRepositoryGORM struct {
	db *gorm.DB
}

// NewBankAccountRepository creates a new instance of BankAccountRepositoryGORM
func NewBankAccountRepository(db *gorm.DB) repositories.BankAccountRepository {
	return &BankAccountRepositoryGORM{db: db}
}

func (r *BankAccountRepositoryGORM) Create(ctx context.Context, entity *entities.BankAccounts) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *BankAccountRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.BankAccounts, error) {
	var entity entities.BankAccounts
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BankAccountRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.BankAccounts, error) {
	var entities []*entities.BankAccounts
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&entities).Error
	return entities, err
}

func (r *BankAccountRepositoryGORM) Update(ctx context.Context, entity *entities.BankAccounts) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *BankAccountRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.BankAccounts{}, "id = ?", id).Error
}

func (r *BankAccountRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.BankAccounts{}).Count(&count).Error
	return count, err
}

// FindByUserID retrieves all bank accounts for a specific user
func (r *BankAccountRepositoryGORM) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.BankAccounts, error) {
	var accounts []*entities.BankAccounts
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_primary DESC, created_at DESC").
		Find(&accounts).Error
	return accounts, err
}

// FindByAccountNumber finds a bank account by account number for a specific user
func (r *BankAccountRepositoryGORM) FindByAccountNumber(ctx context.Context, userID uuid.UUID, accountNumber string) (*entities.BankAccounts, error) {
	var account entities.BankAccounts
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND account_number = ?", userID, accountNumber).
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}
