package entities

import (
	"time"

	"github.com/google/uuid"
)

// WheelPrize represents the wheel_prizes table
type WheelPrize struct {
	ID                 uuid.UUID `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	PrizeCode          string    `json:"prize_code" gorm:"column:prize_code;uniqueIndex;size:50"`
	PrizeName          string    `json:"prize_name" gorm:"column:prize_name;not null" validate:"required"`
	PrizeType          string    `json:"prize_type" gorm:"column:prize_type;not null" validate:"required,oneof=CASH AIRTIME DATA POINTS TICKETS NO_WIN"`
	PrizeValue         int64     `json:"prize_value" gorm:"column:prize_value;type:bigint;not null" validate:"required"` // Value in kobo
	Probability        float64   `json:"probability" gorm:"column:probability;not null" validate:"required"`
	MinimumRecharge    *float64  `json:"minimum_recharge" gorm:"column:minimum_recharge"`
	IsActive           *bool     `json:"is_active" gorm:"column:is_active"`
	IconName           string    `json:"icon_name" gorm:"column:icon_name"`
	ColorScheme        string    `json:"color_scheme" gorm:"column:color_scheme"`
	SortOrder          *int      `json:"sort_order" gorm:"column:sort_order"`
	Description        string    `json:"description" gorm:"column:description"`
	TermsAndConditions string    `json:"terms_and_conditions" gorm:"column:terms_and_conditions"`
	// IsNoWin marks this slot as a "no prize" outcome (e.g. "Try Again", "Better Luck Next Time").
	// When the wheel lands on an IsNoWin slot the backend does NOT create a spin_result record
	// and returns no_win=true so the frontend shows a "try again" message instead of a win toast.
	// NoWinMessage is the text displayed to the user (overrides the default if set).
	IsNoWin     *bool  `json:"is_no_win"     gorm:"column:is_no_win;default:false"`
	NoWinMessage string `json:"no_win_message" gorm:"column:no_win_message;size:200"`

	CreatedAt          time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName specifies the table name for WheelPrize
func (WheelPrize) TableName() string {
	return "wheel_prizes"
}
