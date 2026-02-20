package entities

import (
	"time"

	"github.com/google/uuid"
)

// DataPlans represents the data_plans table
type DataPlans struct {
	ID                 uuid.UUID  `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	NetworkProvider    string     `json:"network_provider" gorm:"column:network_provider;not null"`
	PlanName           string     `json:"plan_name" gorm:"column:plan_name;not null" validate:"required"`
	DataAmount         string     `json:"data_amount" gorm:"column:data_amount;not null" validate:"required"`
	Price              float64    `json:"price" gorm:"column:price;type:numeric(10,2);not null" validate:"required"`
	ValidityDays       int        `json:"validity_days" gorm:"column:validity_days;not null" validate:"required"`
	PlanCode           string     `json:"plan_code" gorm:"column:plan_code;not null" validate:"required"`
	IsActive           *bool      `json:"is_active" gorm:"column:is_active"`
	SortOrder          *int       `json:"sort_order" gorm:"column:sort_order"`
	Description        string     `json:"description" gorm:"column:description"`
	TermsAndConditions string     `json:"terms_and_conditions" gorm:"column:terms_and_conditions"`
	CreatedAt          *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt          *time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for DataPlans
func (DataPlans) TableName() string {
	return "data_plans"
}
