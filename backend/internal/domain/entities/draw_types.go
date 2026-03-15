package entities

import (
	"time"

	"github.com/google/uuid"
)

// DrawType represents the draw_types table
type DrawType struct {
	ID          uuid.UUID  `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name        string     `json:"name" gorm:"column:name;not null;uniqueIndex" validate:"required"`
	Description *string    `json:"description" gorm:"column:description"`
	IsActive    bool       `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt   *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   *time.Time `json:"updated_at" gorm:"column:updated_at"`

	// Relationships
	PrizeTemplates []PrizeTemplate `json:"prize_templates,omitempty" gorm:"foreignKey:DrawTypeID"`
}

// TableName specifies the table name for DrawType
func (DrawType) TableName() string {
	return "draw_types"
}
