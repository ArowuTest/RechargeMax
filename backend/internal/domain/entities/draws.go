package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Draws represents the draws table
type Draws struct {
	ID       uuid.UUID `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	DrawCode string    `json:"draw_code" gorm:"column:draw_code;uniqueIndex;size:20"`

	// Draw details
	Name        string  `json:"name" gorm:"column:name;not null" validate:"required"`
	Type        string  `json:"type" gorm:"column:type;not null;check:type IN ('DAILY','WEEKLY','MONTHLY','SPECIAL')" validate:"required,oneof=DAILY WEEKLY MONTHLY SPECIAL"`
	Description *string `json:"description" gorm:"column:description"`

	// Draw configuration
	Status   string     `json:"status" gorm:"column:status;default:UPCOMING;check:status IN ('UPCOMING','ACTIVE','COMPLETED','CANCELLED')" validate:"oneof=UPCOMING ACTIVE COMPLETED CANCELLED"`
	StartTime time.Time `json:"start_time" gorm:"column:start_time;not null" validate:"required"`
	EndTime   time.Time `json:"end_time" gorm:"column:end_time;not null" validate:"required"`
	DrawTime  *time.Time `json:"draw_time" gorm:"column:draw_time"`

	// Prize configuration
	PrizePool        float64 `json:"prize_pool" gorm:"column:prize_pool;type:decimal(12,2);not null" validate:"required,min=0"`
	WinnersCount     int     `json:"winners_count" gorm:"column:winners_count;default:1"`
	RunnerUpsCount   int     `json:"runner_ups_count" gorm:"column:runner_ups_count;default:1"`
	DrawTypeID       *uuid.UUID `json:"draw_type_id" gorm:"column:draw_type_id;type:uuid;index"`       // Links to draw_types table
	PrizeTemplateID  *uuid.UUID `json:"prize_template_id" gorm:"column:prize_template_id;type:uuid;index"` // Links to prize_templates table

	// Draw statistics
	TotalEntries  int `json:"total_entries" gorm:"column:total_entries;default:0"`
	TotalWinners  int `json:"total_winners" gorm:"column:total_winners;default:0"`

	// Draw results (JSONB)
	Results datatypes.JSON `json:"results" gorm:"column:results;type:jsonb"`

	// Timestamps
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	CompletedAt *time.Time     `json:"completed_at" gorm:"column:completed_at"`

	// Associations
	Winners []*DrawWinners `json:"winners,omitempty" gorm:"foreignKey:DrawID"`
	Entries []*DrawEntries `json:"entries,omitempty" gorm:"foreignKey:DrawID"`
}

// TableName specifies the table name for Draws
func (Draws) TableName() string {
	return "draws"
}

// BeforeCreate hook
func (d *Draws) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// IsActive checks if the draw is currently active
func (d *Draws) IsActive() bool {
	now := time.Now()
	return d.Status == "ACTIVE" && now.After(d.StartTime) && now.Before(d.EndTime)
}

// IsUpcoming checks if the draw is upcoming
func (d *Draws) IsUpcoming() bool {
	return d.Status == "UPCOMING" && time.Now().Before(d.StartTime)
}

// IsCompleted checks if the draw is completed
func (d *Draws) IsCompleted() bool {
	return d.Status == "COMPLETED"
}
