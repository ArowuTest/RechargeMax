package repositories

import (
	"context"
	
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// AdminRepository defines the interface for admin data operations
type AdminRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.AdminUsers) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.AdminUsers, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.AdminUsers, error)
	Update(ctx context.Context, entity *entities.AdminUsers) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Authentication methods
	GetByEmail(ctx context.Context, email string) (*entities.AdminUsers, error)
	GetByID(ctx context.Context, id string) (*entities.AdminUsers, error)
	UpdateLastLogin(ctx context.Context, id string) error
}
