package services

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// Entity
// ─────────────────────────────────────────────────────────────────────────────

// TransactionLimit is the DB entity for the transaction_limits table.
type TransactionLimit struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	LimitType  string    `json:"limit_type"`
	LimitScope string    `json:"limit_scope"`
	MinAmount  *float64  `json:"min_amount"`
	MaxAmount  *float64  `json:"max_amount"`
	DailyLimit *float64  `json:"daily_limit"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName tells GORM which table this entity maps to.
func (TransactionLimit) TableName() string { return "transaction_limits" }

// ─────────────────────────────────────────────────────────────────────────────
// TransactionLimitsService
// ─────────────────────────────────────────────────────────────────────────────

// TransactionLimitsService manages transaction-limit CRUD.
type TransactionLimitsService struct {
	db *gorm.DB
}

// NewTransactionLimitsService constructs a TransactionLimitsService.
func NewTransactionLimitsService(db *gorm.DB) *TransactionLimitsService {
	return &TransactionLimitsService{db: db}
}

// List returns all limits ordered by type and scope.
func (s *TransactionLimitsService) List(_ context.Context) ([]TransactionLimit, error) {
	var limits []TransactionLimit
	if err := s.db.Order("limit_type, limit_scope").Find(&limits).Error; err != nil {
		return nil, err
	}
	return limits, nil
}

// GetByID returns a single limit record.
func (s *TransactionLimitsService) GetByID(_ context.Context, id string) (*TransactionLimit, error) {
	var limit TransactionLimit
	if err := s.db.First(&limit, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &limit, nil
}

// Create inserts a new limit.
func (s *TransactionLimitsService) Create(_ context.Context, limit *TransactionLimit) error {
	return s.db.Create(limit).Error
}

// Update saves all fields of an existing limit.
func (s *TransactionLimitsService) Update(_ context.Context, limit *TransactionLimit) error {
	return s.db.Save(limit).Error
}

// Delete removes a limit by id.
func (s *TransactionLimitsService) Delete(_ context.Context, id string) error {
	return s.db.Delete(&TransactionLimit{}, "id = ?", id).Error
}
