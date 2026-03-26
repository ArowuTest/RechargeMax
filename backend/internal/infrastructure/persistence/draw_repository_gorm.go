package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// DrawRepositoryGORM implements the DrawRepository interface using GORM
type DrawRepositoryGORM struct {
	db *gorm.DB
}

// NewDrawRepository creates a new instance of DrawRepositoryGORM
func NewDrawRepository(db *gorm.DB) repositories.DrawRepository {
	return &DrawRepositoryGORM{db: db}
}

func (r *DrawRepositoryGORM) Create(ctx context.Context, entity *entities.Draw) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *DrawRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.Draw, error) {
	var entity entities.Draw
	err := r.db.WithContext(ctx).
		Preload("Winners").
		First(&entity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *DrawRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.Draw, error) {
	var draws []*entities.Draw
	err := r.db.WithContext(ctx).
		Order("draw_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&draws).Error
	return draws, err
}

func (r *DrawRepositoryGORM) Update(ctx context.Context, entity *entities.Draw) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *DrawRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Draw{}, "id = ?", id).Error
}

func (r *DrawRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Draw{}).Count(&count).Error
	return count, err
}

// FindByStatus retrieves draws by status
func (r *DrawRepositoryGORM) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Draw, error) {
	var draws []*entities.Draw
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("draw_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&draws).Error
	return draws, err
}

// FindByDateRange retrieves draws within a date range
func (r *DrawRepositoryGORM) FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entities.Draw, error) {
	var draws []*entities.Draw
	err := r.db.WithContext(ctx).
		Where("draw_time BETWEEN ? AND ?", startDate, endDate).
		Order("draw_time DESC").
		Find(&draws).Error
	return draws, err
}

// FindUpcoming retrieves upcoming draws
func (r *DrawRepositoryGORM) FindUpcoming(ctx context.Context, limit int) ([]*entities.Draw, error) {
	var draws []*entities.Draw
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("draw_date >= ? AND status IN (?)", now, []string{"scheduled", "entries_exported"}).
		Order("draw_date ASC").
		Limit(limit).
		Find(&draws).Error
	return draws, err
}

// FindCompleted retrieves completed draws
func (r *DrawRepositoryGORM) FindCompleted(ctx context.Context, limit, offset int) ([]*entities.Draw, error) {
	var draws []*entities.Draw
	err := r.db.WithContext(ctx).
		Where("status = ?", "completed").
		Order("draw_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&draws).Error
	return draws, err
}

// GetActiveDraw retrieves the currently active draw
func (r *DrawRepositoryGORM) GetActiveDraw(ctx context.Context) (*entities.Draw, error) {
	var draw entities.Draw
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("start_date <= ? AND end_date >= ? AND status = ?", now, now, "scheduled").
		Order("draw_date ASC").
		First(&draw).Error
	if err != nil {
		return nil, err
	}
	return &draw, nil
}

// UpdateStatus updates the status of a draw
func (r *DrawRepositoryGORM) UpdateStatus(ctx context.Context, drawID uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&entities.Draw{}).
		Where("id = ?", drawID).
		Update("status", status).
		Error
}

// UpdateStats updates the statistics of a draw
func (r *DrawRepositoryGORM) UpdateStats(ctx context.Context, drawID uuid.UUID, totalEntries, totalParticipants int) error {
	return r.db.WithContext(ctx).
		Model(&entities.Draw{}).
		Where("id = ?", drawID).
		Updates(map[string]interface{}{
			"total_entries":      totalEntries,
			"total_participants": totalParticipants,
		}).
		Error
}

// GetDrawEntries retrieves all entries for a draw (aggregated from user points)
func (r *DrawRepositoryGORM) GetDrawEntries(ctx context.Context, drawID uuid.UUID) ([]map[string]interface{}, error) {
	// This aggregates entries from users' points
	// Each point = 1 entry
	var results []map[string]interface{}
	
	// Query to get users with their points as entries
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.id as user_id, users.msisdn, users.total_points as entries").
		Where("users.total_points > 0").
		Scan(&results).Error
	
	return results, err
}

// CountEntriesByUser counts total entries for a specific user in a draw
func (r *DrawRepositoryGORM) CountEntriesByUser(ctx context.Context, drawID uuid.UUID, userID uuid.UUID) (int64, error) {
	// Each point = 1 entry
	var user struct {
		TotalPoints int64
	}
	
	err := r.db.WithContext(ctx).
		Table("users").
		Select("total_points").
		Where("id = ?", userID).
		Scan(&user).Error
	
	if err != nil {
		return 0, err
	}
	
	return user.TotalPoints, nil
}

// CreateEntry creates a new draw entry
func (r *DrawRepositoryGORM) CreateEntry(ctx context.Context, entry *entities.DrawEntries) error {
	return r.db.WithContext(ctx).Omit("updated_at").Create(entry).Error
}
