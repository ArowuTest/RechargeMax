package entities

import (
	"time"

	"github.com/google/uuid"
)

// WheelPrizes represents the wheel_prizes table
type WheelPrizes struct {
	ID                 uuid.UUID  `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	PrizeCode          string     `json:"prize_code" gorm:"column:prize_code;uniqueIndex;size:20"`
	PrizeName          string     `json:"prize_name" gorm:"column:prize_name;not null" validate:"required"`
	PrizeType          string     `json:"prize_type" gorm:"column:prize_type;not null" validate:"required,oneof=CASH AIRTIME DATA POINTS"`
	PrizeValue         int64      `json:"prize_value" gorm:"column:prize_value;type:bigint;not null" validate:"required"`
	Probability        float64    `json:"probability" gorm:"column:probability;not null" validate:"required"`
	MinimumRecharge    *float64   `json:"minimum_recharge" gorm:"column:minimum_recharge"`
	IsActive           *bool      `json:"is_active" gorm:"column:is_active"`
	IconName           string     `json:"icon_name" gorm:"column:icon_name"`
	ColorScheme        string     `json:"color_scheme" gorm:"column:color_scheme"`
	SortOrder          *int       `json:"sort_order" gorm:"column:sort_order"`
	Description        string     `json:"description" gorm:"column:description"`
	TermsAndConditions string     `json:"terms_and_conditions" gorm:"column:terms_and_conditions"`
	CreatedAt          time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName specifies the table name for WheelPrizes
func (WheelPrizes) TableName() string {
	return "wheel_prizes"
}
