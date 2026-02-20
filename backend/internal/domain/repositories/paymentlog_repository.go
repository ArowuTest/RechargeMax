package repositories

import (
	"context"
	
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// PaymentLogRepository defines the interface for paymentlog data operations
type PaymentLogRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.PaymentLogs) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.PaymentLogs, error)
	GetByReference(ctx context.Context, reference string) (*entities.PaymentLogs, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.PaymentLogs, error)
	Update(ctx context.Context, entity *entities.PaymentLogs) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}
