package entities

import (
	"time"

	"github.com/google/uuid"
)

// DrawWinners represents the draw_winners table
type DrawWinners struct {
	ID             uuid.UUID  `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	DrawID         uuid.UUID  `json:"draw_id" gorm:"column:draw_id;not null;index"`
	UserID         *uuid.UUID `json:"user_id" gorm:"column:user_id;index"`
	Msisdn         string     `json:"msisdn" gorm:"column:msisdn;not null" validate:"required"`
	Position       int        `json:"position" gorm:"column:position;not null" validate:"required"`
	PrizeAmount    int64      `json:"prize_amount" gorm:"column:prize_amount;type:bigint;not null" validate:"required"` // Prize amount in kobo
	ClaimedAt      *time.Time `json:"claimed_at" gorm:"column:claimed_at"`
	ClaimReference string     `json:"claim_reference" gorm:"column:claim_reference"`
	IsRunnerUp     bool       `json:"is_runner_up" gorm:"column:is_runner_up;default:false"`
	IsForfeited    bool       `json:"is_forfeited" gorm:"column:is_forfeited;default:false"`
	PromotedFrom     *uuid.UUID `json:"promoted_from" gorm:"column:promoted_from;type:uuid"`
	PrizeCategoryID  *uint      `json:"prize_category_id" gorm:"column:prize_category_id;index"`
	CategoryName     *string    `json:"category_name" gorm:"column:category_name"`
	CreatedAt      *time.Time `json:"created_at" gorm:"column:created_at"`
	ExpiresAt      *time.Time `json:"expires_at" gorm:"column:expires_at"`
}

// TableName specifies the table name for DrawWinners
func (DrawWinners) TableName() string {
	return "draw_winners"
}
