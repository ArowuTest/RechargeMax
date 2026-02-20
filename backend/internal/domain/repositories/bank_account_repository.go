package repositories

import (
"context"
"github.com/google/uuid"
"rechargemax/internal/domain/entities"
)

type BankAccountRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.BankAccounts) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.BankAccounts, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.BankAccounts, error)
	Update(ctx context.Context, entity *entities.BankAccounts) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.BankAccounts, error)
	FindByAccountNumber(ctx context.Context, userID uuid.UUID, accountNumber string) (*entities.BankAccounts, error)
}
