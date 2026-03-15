package entities

import (
	"time"

	"github.com/google/uuid"
)

// PrizeCategory represents the prize_categories table
// DB column: template_id (FK to prize_templates.id)
type PrizeCategory struct {
	ID              uuid.UUID  `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	PrizeTemplateID uuid.UUID  `json:"prize_template_id" gorm:"column:template_id;type:uuid;index;not null"`
	DrawID          *uuid.UUID `json:"draw_id" gorm:"column:draw_id;type:uuid;index"`
	CategoryName    string     `json:"category_name" gorm:"column:category_name;not null" validate:"required"`
	PrizeAmount     float64    `json:"prize_amount" gorm:"column:prize_amount;not null" validate:"required,gt=0"`
	WinnerCount     int        `json:"winner_count" gorm:"column:winners_count;not null;default:1" validate:"required,gt=0"`
	RunnerUpCount   int        `json:"runner_up_count" gorm:"column:runner_ups_count;not null;default:1" validate:"required,gte=0"`
	DisplayOrder    int        `json:"display_order" gorm:"column:display_order;not null;default:0"`
	CreatedAt       *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       *time.Time `json:"updated_at" gorm:"column:updated_at"`

	// Relationships
	PrizeTemplate *PrizeTemplate `json:"prize_template,omitempty" gorm:"foreignKey:PrizeTemplateID"`
	Draw          *Draw         `json:"draw,omitempty" gorm:"foreignKey:DrawID"`
}

// TableName specifies the table name for PrizeCategory
func (PrizeCategory) TableName() string {
	return "prize_categories"
}
