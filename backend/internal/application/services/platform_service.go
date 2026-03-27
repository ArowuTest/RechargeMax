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
	if err := db.Model(&entities.Transaction{}).Where("status = ?", "SUCCESS").Count(&totalTransactions).Error; err != nil {
		return nil, err
	}

	var totalPrizes int64
	if err := db.Model(&entities.DrawWinners{}).Count(&totalPrizes).Error; err != nil {
		return nil, err
	}

	// Use Find (not First) so GORM does not log a noisy "record not found"
	// error when no draw is currently active — that is expected behaviour.
	var activeDraws []entities.Draw
	db.Where("status = ? AND end_time > ?", "ACTIVE", time.Now()).
		Order("created_at DESC").
		Limit(1).
		Find(&activeDraws)
	var activeDrawData map[string]interface{}
	if len(activeDraws) > 0 {
		activeDraw := activeDraws[0]
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

// PublicWinner is the public-facing winner record (no sensitive MSISDN).
type PublicWinner struct {
	ID               string    `json:"id"`
	DrawID           string    `json:"draw_id"`
	DrawName         string    `json:"draw_name"`
	DrawType         string    `json:"draw_type"`
	MaskedMSISDN     string    `json:"masked_msisdn"`
	Position         int       `json:"position"`
	PrizeType        string    `json:"prize_type"`
	PrizeDescription string    `json:"prize_description"`
	PrizeAmount      int64     `json:"prize_amount"` // kobo
	ClaimStatus      string    `json:"claim_status"`
	WonAt            time.Time `json:"won_at"`
}

// GetPublicWinners returns a paginated list of winners for the public wall.
func (s *PlatformService) GetPublicWinners(ctx context.Context, page, limit int) ([]PublicWinner, int64, error) {
	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit

	var total int64
	s.db.WithContext(ctx).Table("draw_winners").Count(&total)

	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT
			dw.id,
			dw.draw_id,
			COALESCE(d.name, 'Prize Draw')   AS draw_name,
			COALESCE(d.type, 'DAILY')         AS draw_type,
			dw.msisdn,
			dw.position,
			dw.prize_type,
			dw.prize_description,
			COALESCE(dw.prize_amount, 0)      AS prize_amount,
			dw.claim_status,
			dw.created_at                     AS won_at
		FROM draw_winners dw
		LEFT JOIN draws d ON d.id = dw.draw_id
		ORDER BY dw.created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var winners []PublicWinner
	for rows.Next() {
		var w PublicWinner
		var msisdn string
		if err := rows.Scan(&w.ID, &w.DrawID, &w.DrawName, &w.DrawType,
			&msisdn, &w.Position, &w.PrizeType, &w.PrizeDescription,
			&w.PrizeAmount, &w.ClaimStatus, &w.WonAt); err != nil {
			continue
		}
		// mask: show first 4 + last 2 digits
		if len(msisdn) >= 6 {
			w.MaskedMSISDN = msisdn[:4] + "****" + msisdn[len(msisdn)-2:]
		} else {
			w.MaskedMSISDN = "****"
		}
		winners = append(winners, w)
	}
	if winners == nil {
		winners = []PublicWinner{}
	}
	return winners, total, nil
}

// GetPublicWinnerByID returns a single winner record (public, masked).
func (s *PlatformService) GetPublicWinnerByID(ctx context.Context, id string) (*PublicWinner, error) {
	row := s.db.WithContext(ctx).Raw(`
		SELECT
			dw.id,
			dw.draw_id,
			COALESCE(d.name, 'Prize Draw')   AS draw_name,
			COALESCE(d.type, 'DAILY')         AS draw_type,
			dw.msisdn,
			dw.position,
			dw.prize_type,
			dw.prize_description,
			COALESCE(dw.prize_amount, 0)      AS prize_amount,
			dw.claim_status,
			dw.created_at                     AS won_at
		FROM draw_winners dw
		LEFT JOIN draws d ON d.id = dw.draw_id
		WHERE dw.id = ?
	`, id).Row()

	var w PublicWinner
	var msisdn string
	if err := row.Scan(&w.ID, &w.DrawID, &w.DrawName, &w.DrawType,
		&msisdn, &w.Position, &w.PrizeType, &w.PrizeDescription,
		&w.PrizeAmount, &w.ClaimStatus, &w.WonAt); err != nil {
		return nil, err
	}
	if len(msisdn) >= 6 {
		w.MaskedMSISDN = msisdn[:4] + "****" + msisdn[len(msisdn)-2:]
	} else {
		w.MaskedMSISDN = "****"
	}
	return &w, nil
}
