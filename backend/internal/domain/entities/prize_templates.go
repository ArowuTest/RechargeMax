package entities

import (
	"time"
)

// PrizeTemplate represents the prize_templates table
type PrizeTemplate struct {
	ID          uint       `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name        string          `json:"name" gorm:"column:name;not null" validate:"required"`
	DrawTypeID  uint      `json:"draw_type_id" gorm:"column:draw_type_id;not null"`
	Description *string         `json:"description" gorm:"column:description"`
	IsDefault   bool            `json:"is_default" gorm:"column:is_default;default:false"`
	IsActive    bool            `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt   *time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   *time.Time      `json:"updated_at" gorm:"column:updated_at"`
	
	// Relationships
	DrawType    *DrawType        `json:"draw_type,omitempty" gorm:"foreignKey:DrawTypeID"`
	PrizeCategories  []PrizeCategory `json:"prize_categories,omitempty" gorm:"foreignKey:PrizeTemplateID"`
}

// TableName specifies the table name for PrizeTemplate
func (PrizeTemplate) TableName() string {
	return "prize_templates"
}
