package repositories

import (
	"context"
	"time"
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

type DrawRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.Draw) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Draw, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.Draw, error)
	Update(ctx context.Context, entity *entities.Draw) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Draw, error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entities.Draw, error)
	FindUpcoming(ctx context.Context, limit int) ([]*entities.Draw, error)
	FindCompleted(ctx context.Context, limit, offset int) ([]*entities.Draw, error)
	GetActiveDraw(ctx context.Context) (*entities.Draw, error)
	UpdateStatus(ctx context.Context, drawID uuid.UUID, status string) error
	UpdateStats(ctx context.Context, drawID uuid.UUID, totalEntries, totalParticipants int) error
	
	// Entry management methods
	CreateEntry(ctx context.Context, entry *entities.DrawEntries) error
	GetDrawEntries(ctx context.Context, drawID uuid.UUID) ([]map[string]interface{}, error)
	CountEntriesByUser(ctx context.Context, drawID uuid.UUID, userID uuid.UUID) (int64, error)
}
