package services

import (
	"context"
	"time"

	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
)

// PlatformStats is the platform-wide statistics payload.
type PlatformStats struct {
	TotalUsers        int64                  `json:"totalUsers"`
	TotalTransactions int64                  `json:"totalTransactions"`
	TotalPrizes       int64                  `json:"totalPrizes"`
	ActiveDraw        map[string]interface{} `json:"activeDraw"`
}

// RecentWinner is the display record returned by GetRecentWinners.
type RecentWinner struct {
	FullName         string    `json:"full_name"`
	PrizeDescription string    `json:"prize_description"`
	PrizeValue       float64   `json:"prize_value"`
	CreatedAt        time.Time `json:"created_at"`
	NetworkProvider  string    `json:"network_provider"`
	Position         int       `json:"position"`
}

// PlatformService encapsulates all platform-wide read queries.
type PlatformService struct {
	db *gorm.DB
}

// NewPlatformService constructs a PlatformService.
func NewPlatformService(db *gorm.DB) *PlatformService {
	return &PlatformService{db: db}
}

// GetStatistics returns aggregate platform statistics.
func (s *PlatformService) GetStatistics(ctx context.Context) (*PlatformStats, error) {
	db := s.db.WithContext(ctx)

	var totalUsers int64
	if err := db.Model(&entities.User{}).Where("is_active = ?", true).Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	var totalTransactions int64
	if err := db.Model(&entities.Transaction{}).Where("status = ?", "completed").Count(&totalTransactions).Error; err != nil {
		return nil, err
	}

	var totalPrizes int64
	if err := db.Model(&entities.DrawWinners{}).Count(&totalPrizes).Error; err != nil {
		return nil, err
	}

	var activeDraw entities.Draw
	var activeDrawData map[string]interface{}
	err := db.Where("status = ? AND end_time > ?", "ACTIVE", time.Now()).
		Order("created_at DESC").
		First(&activeDraw).Error
	if err == nil {
		var entries int64
		db.Model(&entities.DrawEntry{}).Where("draw_id = ?", activeDraw.ID).Count(&entries)
		activeDrawData = map[string]interface{}{
			"name":      activeDraw.Name,
			"prizePool": activeDraw.PrizePool,
			"endTime":   activeDraw.EndTime,
			"entries":   entries,
		}
	}

	return &PlatformStats{
		TotalUsers:        totalUsers,
		TotalTransactions: totalTransactions,
		TotalPrizes:       totalPrizes,
		ActiveDraw:        activeDrawData,
	}, nil
}

// GetRecentWinners returns the most recent n winners.
func (s *PlatformService) GetRecentWinners(ctx context.Context, limit int) ([]RecentWinner, error) {
	if limit <= 0 {
		limit = 4
	}
	var winners []RecentWinner
	err := s.db.WithContext(ctx).Table("draw_winners").
		Select("users.full_name, CONCAT('Position ', draw_winners.position, ' Prize') as prize_description, draw_winners.prize_amount as prize_value, draw_winners.created_at, users.msisdn as network_provider, draw_winners.position").
		Joins("LEFT JOIN users ON draw_winners.user_id = users.id").
		Where("draw_winners.claim_status IN (?)", []string{"CLAIMED", "PROCESSING"}).
		Order("draw_winners.created_at DESC").
		Limit(limit).
		Scan(&winners).Error
	return winners, err
}
