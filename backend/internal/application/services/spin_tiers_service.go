package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rechargemax/internal/utils"
)

// ─────────────────────────────────────────────────────────────────────────────
// DTOs
// ─────────────────────────────────────────────────────────────────────────────

// UpdateSpinTierRequest carries the mutable fields for a tier update.
// All fields are optional (pointer = nil means "no change").
type UpdateSpinTierRequest struct {
	TierDisplayName *string
	MinDailyAmount  *int64
	MaxDailyAmount  *int64
	SpinsPerDay     *int
	TierColor       *string
	TierIcon        *string
	TierBadge       *string
	Description     *string
	SortOrder       *int
	IsActive        *bool
}

// CreateSpinTierRequest carries required fields for a new tier.
type CreateSpinTierRequest struct {
	TierName        string
	TierDisplayName string
	MinDailyAmount  int64
	MaxDailyAmount  int64
	SpinsPerDay     int
	TierColor       string
	TierIcon        string
	TierBadge       string
	Description     string
	SortOrder       int
}

// ─────────────────────────────────────────────────────────────────────────────
// SpinTiersService
// ─────────────────────────────────────────────────────────────────────────────

// SpinTiersService manages spin-tier configuration.
type SpinTiersService struct {
	db         *gorm.DB
	calculator *utils.SpinTierCalculatorDB
}

// NewSpinTiersService constructs a SpinTiersService.
func NewSpinTiersService(db *gorm.DB) *SpinTiersService {
	return &SpinTiersService{db: db, calculator: utils.NewSpinTierCalculatorDB(db)}
}

// ListAll returns all tiers (active + inactive) ordered by sort_order.
func (s *SpinTiersService) ListAll(_ context.Context) ([]utils.SpinTierDB, error) {
	var tiers []utils.SpinTierDB
	if err := s.db.Order("sort_order ASC").Find(&tiers).Error; err != nil {
		return nil, err
	}
	return tiers, nil
}

// GetByID returns a single tier or an error.
func (s *SpinTiersService) GetByID(_ context.Context, id string) (*utils.SpinTierDB, error) {
	return s.calculator.GetTierByIDFromDB(id)
}

// Update applies the non-nil fields of req to the tier identified by id.
func (s *SpinTiersService) Update(_ context.Context, id string, req UpdateSpinTierRequest, adminID string) (*utils.SpinTierDB, error) {
	if _, err := s.calculator.GetTierByIDFromDB(id); err != nil {
		return nil, fmt.Errorf("tier not found: %w", err)
	}

	updates := map[string]interface{}{"updated_by": adminID}
	if req.TierDisplayName != nil {
		updates["tier_display_name"] = *req.TierDisplayName
	}
	if req.MinDailyAmount != nil {
		if *req.MinDailyAmount < 0 {
			return nil, fmt.Errorf("min_daily_amount cannot be negative")
		}
		updates["min_daily_amount"] = *req.MinDailyAmount
	}
	if req.MaxDailyAmount != nil {
		if *req.MaxDailyAmount <= 0 {
			return nil, fmt.Errorf("max_daily_amount must be positive")
		}
		updates["max_daily_amount"] = *req.MaxDailyAmount
	}
	if req.SpinsPerDay != nil {
		if *req.SpinsPerDay <= 0 {
			return nil, fmt.Errorf("spins_per_day must be positive")
		}
		updates["spins_per_day"] = *req.SpinsPerDay
	}
	if req.TierColor != nil {
		updates["tier_color"] = *req.TierColor
	}
	if req.TierIcon != nil {
		updates["tier_icon"] = *req.TierIcon
	}
	if req.TierBadge != nil {
		updates["tier_badge"] = *req.TierBadge
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.SortOrder != nil {
		if *req.SortOrder < 0 {
			return nil, fmt.Errorf("sort_order cannot be negative")
		}
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := s.db.Model(&utils.SpinTierDB{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.calculator.ValidateTierConfiguration(); err != nil {
		return nil, fmt.Errorf("tier configuration conflict after update: %w", err)
	}
	return s.calculator.GetTierByIDFromDB(id)
}

// Create inserts a new tier.  Rolls back if the config becomes invalid.
func (s *SpinTiersService) Create(_ context.Context, req CreateSpinTierRequest, adminID string) (*utils.SpinTierDB, error) {
	if req.MinDailyAmount < 0 {
		return nil, fmt.Errorf("min_daily_amount cannot be negative")
	}
	if req.MaxDailyAmount <= req.MinDailyAmount {
		return nil, fmt.Errorf("max_daily_amount must be greater than min_daily_amount")
	}
	if req.SpinsPerDay <= 0 {
		return nil, fmt.Errorf("spins_per_day must be positive")
	}

	tierID := uuid.New().String()
	tier := utils.SpinTierDB{
		ID:              tierID,
		TierName:        req.TierName,
		TierDisplayName: req.TierDisplayName,
		MinDailyAmount:  req.MinDailyAmount,
		MaxDailyAmount:  req.MaxDailyAmount,
		SpinsPerDay:     req.SpinsPerDay,
		TierColor:       req.TierColor,
		TierIcon:        req.TierIcon,
		TierBadge:       req.TierBadge,
		Description:     req.Description,
		SortOrder:       req.SortOrder,
		IsActive:        true,
	}
	if adminID != "" {
		tier.CreatedBy = &adminID
		tier.UpdatedBy = &adminID
	}

	if err := s.db.Create(&tier).Error; err != nil {
		return nil, err
	}
	if err := s.calculator.ValidateTierConfiguration(); err != nil {
		s.db.Delete(&tier) // rollback
		return nil, fmt.Errorf("tier configuration conflict after create: %w", err)
	}
	return &tier, nil
}

// Delete soft-deletes a tier by setting is_active = false.
func (s *SpinTiersService) Delete(_ context.Context, id string) (*utils.SpinTierDB, error) {
	tier, err := s.calculator.GetTierByIDFromDB(id)
	if err != nil {
		return nil, fmt.Errorf("tier not found: %w", err)
	}
	if err := s.db.Model(&utils.SpinTierDB{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return nil, err
	}
	return tier, nil
}

// ValidateConfiguration runs the tier range validation and returns a slice of human-readable error strings.
// An empty slice means the configuration is valid.
func (s *SpinTiersService) ValidateConfiguration(ctx context.Context) ([]string, error) {
	err := s.calculator.ValidateTierConfiguration()
	if err != nil {
		return []string{err.Error()}, nil
	}
	return []string{}, nil
}
