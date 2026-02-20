package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TransactionLimitsService handles transaction limit configuration and validation
type TransactionLimitsService struct {
	db *sql.DB
}

// TransactionLimit represents a transaction limit configuration
type TransactionLimit struct {
	ID                uuid.UUID  `json:"id"`
	LimitType         string     `json:"limit_type"`
	LimitScope        string     `json:"limit_scope"`
	MinAmount         int64      `json:"min_amount"`          // in kobo
	MaxAmount         int64      `json:"max_amount"`          // in kobo
	DailyLimit        *int64     `json:"daily_limit"`         // in kobo
	MonthlyLimit      *int64     `json:"monthly_limit"`       // in kobo
	IsActive          bool       `json:"is_active"`
	AppliesToUserTier *string    `json:"applies_to_user_tier"`
	Description       string     `json:"description"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	CreatedBy         *uuid.UUID `json:"created_by"`
	UpdatedBy         *uuid.UUID `json:"updated_by"`
}

// LimitCheckResult represents the result of a limit validation
type LimitCheckResult struct {
	IsValid       bool   `json:"is_valid"`
	MinAmount     int64  `json:"min_amount"`
	MaxAmount     int64  `json:"max_amount"`
	DailyLimit    *int64 `json:"daily_limit"`
	MonthlyLimit  *int64 `json:"monthly_limit"`
	CurrentDaily  int64  `json:"current_daily"`
	CurrentMonthly int64  `json:"current_monthly"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

// CreateLimitRequest represents a request to create a new limit
type CreateLimitRequest struct {
	LimitType         string  `json:"limit_type"`
	LimitScope        string  `json:"limit_scope"`
	MinAmount         int64   `json:"min_amount"`
	MaxAmount         int64   `json:"max_amount"`
	DailyLimit        *int64  `json:"daily_limit"`
	MonthlyLimit      *int64  `json:"monthly_limit"`
	AppliesToUserTier *string `json:"applies_to_user_tier"`
	Description       string  `json:"description"`
}

// UpdateLimitRequest represents a request to update an existing limit
type UpdateLimitRequest struct {
	MinAmount    *int64  `json:"min_amount"`
	MaxAmount    *int64  `json:"max_amount"`
	DailyLimit   *int64  `json:"daily_limit"`
	MonthlyLimit *int64  `json:"monthly_limit"`
	IsActive     *bool   `json:"is_active"`
	Description  *string `json:"description"`
}

// NewTransactionLimitsService creates a new transaction limits service
func NewTransactionLimitsService(db *sql.DB) *TransactionLimitsService {
	return &TransactionLimitsService{db: db}
}

// GetLimit retrieves a specific limit by ID
func (s *TransactionLimitsService) GetLimit(ctx context.Context, limitID uuid.UUID) (*TransactionLimit, error) {
	query := `
		SELECT id, limit_type, limit_scope, min_amount, max_amount, daily_limit, monthly_limit,
		       is_active, applies_to_user_tier, description, created_at, updated_at, created_by, updated_by
		FROM transaction_limits
		WHERE id = $1
	`

	limit := &TransactionLimit{}
	err := s.db.QueryRowContext(ctx, query, limitID).Scan(
		&limit.ID, &limit.LimitType, &limit.LimitScope, &limit.MinAmount, &limit.MaxAmount,
		&limit.DailyLimit, &limit.MonthlyLimit, &limit.IsActive, &limit.AppliesToUserTier,
		&limit.Description, &limit.CreatedAt, &limit.UpdatedAt, &limit.CreatedBy, &limit.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("limit not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get limit: %w", err)
	}

	return limit, nil
}

// ListLimits retrieves all transaction limits with optional filtering
func (s *TransactionLimitsService) ListLimits(ctx context.Context, limitType, limitScope string, activeOnly bool) ([]*TransactionLimit, error) {
	query := `
		SELECT id, limit_type, limit_scope, min_amount, max_amount, daily_limit, monthly_limit,
		       is_active, applies_to_user_tier, description, created_at, updated_at, created_by, updated_by
		FROM transaction_limits
		WHERE ($1 = '' OR limit_type = $1)
		  AND ($2 = '' OR limit_scope = $2)
		  AND ($3 = false OR is_active = true)
		ORDER BY limit_type, limit_scope, applies_to_user_tier NULLS LAST
	`

	rows, err := s.db.QueryContext(ctx, query, limitType, limitScope, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to list limits: %w", err)
	}
	defer rows.Close()

	var limits []*TransactionLimit
	for rows.Next() {
		limit := &TransactionLimit{}
		err := rows.Scan(
			&limit.ID, &limit.LimitType, &limit.LimitScope, &limit.MinAmount, &limit.MaxAmount,
			&limit.DailyLimit, &limit.MonthlyLimit, &limit.IsActive, &limit.AppliesToUserTier,
			&limit.Description, &limit.CreatedAt, &limit.UpdatedAt, &limit.CreatedBy, &limit.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan limit: %w", err)
		}
		limits = append(limits, limit)
	}

	return limits, nil
}

// CreateLimit creates a new transaction limit
func (s *TransactionLimitsService) CreateLimit(ctx context.Context, req *CreateLimitRequest, adminID uuid.UUID) (*TransactionLimit, error) {
	// Validate request
	if err := s.validateLimitRequest(req.LimitType, req.LimitScope, req.MinAmount, req.MaxAmount); err != nil {
		return nil, err
	}

	limitID := uuid.New()
	query := `
		INSERT INTO transaction_limits (
			id, limit_type, limit_scope, min_amount, max_amount, daily_limit, monthly_limit,
			applies_to_user_tier, description, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
		RETURNING id, created_at, updated_at
	`

	limit := &TransactionLimit{
		ID:                limitID,
		LimitType:         req.LimitType,
		LimitScope:        req.LimitScope,
		MinAmount:         req.MinAmount,
		MaxAmount:         req.MaxAmount,
		DailyLimit:        req.DailyLimit,
		MonthlyLimit:      req.MonthlyLimit,
		AppliesToUserTier: req.AppliesToUserTier,
		Description:       req.Description,
		IsActive:          true,
		CreatedBy:         &adminID,
		UpdatedBy:         &adminID,
	}

	err := s.db.QueryRowContext(ctx, query,
		limitID, req.LimitType, req.LimitScope, req.MinAmount, req.MaxAmount,
		req.DailyLimit, req.MonthlyLimit, req.AppliesToUserTier, req.Description, adminID,
	).Scan(&limit.ID, &limit.CreatedAt, &limit.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create limit: %w", err)
	}

	// Log audit trail
	if err := s.logAudit(ctx, limitID, "CREATE", nil, limit, adminID, ""); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to log audit: %v\n", err)
	}

	return limit, nil
}

// UpdateLimit updates an existing transaction limit
func (s *TransactionLimitsService) UpdateLimit(ctx context.Context, limitID uuid.UUID, req *UpdateLimitRequest, adminID uuid.UUID, reason string) (*TransactionLimit, error) {
	// Get current limit for audit
	oldLimit, err := s.GetLimit(ctx, limitID)
	if err != nil {
		return nil, err
	}

	// Build dynamic update query
	query := `UPDATE transaction_limits SET updated_by = $1, updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{adminID}
	argPos := 2

	if req.MinAmount != nil {
		query += fmt.Sprintf(", min_amount = $%d", argPos)
		args = append(args, *req.MinAmount)
		argPos++
	}
	if req.MaxAmount != nil {
		query += fmt.Sprintf(", max_amount = $%d", argPos)
		args = append(args, *req.MaxAmount)
		argPos++
	}
	if req.DailyLimit != nil {
		query += fmt.Sprintf(", daily_limit = $%d", argPos)
		args = append(args, *req.DailyLimit)
		argPos++
	}
	if req.MonthlyLimit != nil {
		query += fmt.Sprintf(", monthly_limit = $%d", argPos)
		args = append(args, *req.MonthlyLimit)
		argPos++
	}
	if req.IsActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argPos)
		args = append(args, *req.IsActive)
		argPos++
	}
	if req.Description != nil {
		query += fmt.Sprintf(", description = $%d", argPos)
		args = append(args, *req.Description)
		argPos++
	}

	query += fmt.Sprintf(" WHERE id = $%d RETURNING updated_at", argPos)
	args = append(args, limitID)

	var updatedAt time.Time
	err = s.db.QueryRowContext(ctx, query, args...).Scan(&updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update limit: %w", err)
	}

	// Get updated limit
	newLimit, err := s.GetLimit(ctx, limitID)
	if err != nil {
		return nil, err
	}

	// Log audit trail
	if err := s.logAudit(ctx, limitID, "UPDATE", oldLimit, newLimit, adminID, reason); err != nil {
		fmt.Printf("Failed to log audit: %v\n", err)
	}

	return newLimit, nil
}

// DeleteLimit soft-deletes a transaction limit by deactivating it
func (s *TransactionLimitsService) DeleteLimit(ctx context.Context, limitID uuid.UUID, adminID uuid.UUID, reason string) error {
	oldLimit, err := s.GetLimit(ctx, limitID)
	if err != nil {
		return err
	}

	query := `UPDATE transaction_limits SET is_active = false, updated_by = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = s.db.ExecContext(ctx, query, adminID, limitID)
	if err != nil {
		return fmt.Errorf("failed to delete limit: %w", err)
	}

	// Log audit trail
	if err := s.logAudit(ctx, limitID, "DEACTIVATE", oldLimit, nil, adminID, reason); err != nil {
		fmt.Printf("Failed to log audit: %v\n", err)
	}

	return nil
}

// CheckLimit validates if a transaction amount is within limits
func (s *TransactionLimitsService) CheckLimit(ctx context.Context, limitType string, amount int64, userID uuid.UUID, userTier string) (*LimitCheckResult, error) {
	// Get applicable limit
	query := `SELECT * FROM get_transaction_limit($1, 'PER_TRANSACTION', $2)`
	
	var minAmount, maxAmount int64
	var dailyLimit, monthlyLimit *int64
	
	err := s.db.QueryRowContext(ctx, query, limitType, userTier).Scan(&minAmount, &maxAmount, &dailyLimit, &monthlyLimit)
	if err == sql.ErrNoRows {
		return &LimitCheckResult{
			IsValid:      false,
			ErrorMessage: "No transaction limit configured for this type",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction limit: %w", err)
	}

	result := &LimitCheckResult{
		IsValid:      true,
		MinAmount:    minAmount,
		MaxAmount:    maxAmount,
		DailyLimit:   dailyLimit,
		MonthlyLimit: monthlyLimit,
	}

	// Check per-transaction limits
	if amount < minAmount {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("Amount below minimum limit of ₦%.2f", float64(minAmount)/100)
		return result, nil
	}

	if amount > maxAmount {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("Amount exceeds maximum limit of ₦%.2f", float64(maxAmount)/100)
		return result, nil
	}

	// Check daily cumulative limit if applicable
	if dailyLimit != nil {
		currentDaily, err := s.getDailySpend(ctx, userID, limitType)
		if err != nil {
			return nil, fmt.Errorf("failed to get daily spend: %w", err)
		}
		result.CurrentDaily = currentDaily

		if currentDaily+amount > *dailyLimit {
			result.IsValid = false
			result.ErrorMessage = fmt.Sprintf("Transaction would exceed daily limit of ₦%.2f", float64(*dailyLimit)/100)
			return result, nil
		}
	}

	// Check monthly cumulative limit if applicable
	if monthlyLimit != nil {
		currentMonthly, err := s.getMonthlySpend(ctx, userID, limitType)
		if err != nil {
			return nil, fmt.Errorf("failed to get monthly spend: %w", err)
		}
		result.CurrentMonthly = currentMonthly

		if currentMonthly+amount > *monthlyLimit {
			result.IsValid = false
			result.ErrorMessage = fmt.Sprintf("Transaction would exceed monthly limit of ₦%.2f", float64(*monthlyLimit)/100)
			return result, nil
		}
	}

	return result, nil
}

// Helper functions

func (s *TransactionLimitsService) validateLimitRequest(limitType, limitScope string, minAmount, maxAmount int64) error {
	validTypes := map[string]bool{"AIRTIME": true, "DATA": true, "SUBSCRIPTION": true, "WITHDRAWAL": true}
	validScopes := map[string]bool{"GLOBAL": true, "PER_USER": true, "PER_TRANSACTION": true, "DAILY_CUMULATIVE": true, "MONTHLY_CUMULATIVE": true}

	if !validTypes[limitType] {
		return fmt.Errorf("invalid limit type: %s", limitType)
	}
	if !validScopes[limitScope] {
		return fmt.Errorf("invalid limit scope: %s", limitScope)
	}
	if minAmount <= 0 || maxAmount <= 0 {
		return fmt.Errorf("amounts must be positive")
	}
	if minAmount > maxAmount {
		return fmt.Errorf("min_amount cannot be greater than max_amount")
	}

	return nil
}

func (s *TransactionLimitsService) getDailySpend(ctx context.Context, userID uuid.UUID, limitType string) (int64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1
		  AND transaction_type = $2
		  AND status = 'SUCCESS'
		  AND created_at >= CURRENT_DATE
	`

	var total int64
	err := s.db.QueryRowContext(ctx, query, userID, limitType).Scan(&total)
	return total, err
}

func (s *TransactionLimitsService) getMonthlySpend(ctx context.Context, userID uuid.UUID, limitType string) (int64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1
		  AND transaction_type = $2
		  AND status = 'SUCCESS'
		  AND created_at >= DATE_TRUNC('month', CURRENT_DATE)
	`

	var total int64
	err := s.db.QueryRowContext(ctx, query, userID, limitType).Scan(&total)
	return total, err
}

func (s *TransactionLimitsService) logAudit(ctx context.Context, limitID uuid.UUID, action string, oldLimit, newLimit *TransactionLimit, changedBy uuid.UUID, reason string) error {
	var oldValues, newValues []byte
	var err error

	if oldLimit != nil {
		oldValues, err = json.Marshal(oldLimit)
		if err != nil {
			return err
		}
	}

	if newLimit != nil {
		newValues, err = json.Marshal(newLimit)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO transaction_limits_audit (limit_id, action, old_values, new_values, changed_by, reason)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = s.db.ExecContext(ctx, query, limitID, action, oldValues, newValues, changedBy, reason)
	return err
}
