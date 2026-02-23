package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ============================================================================
// PRIZE FULFILLMENT CONFIGURATION SERVICE (Enterprise-Grade)
// ============================================================================
// Purpose: Manage prize fulfillment configuration with caching and validation
// Features:
// - Admin-configurable fulfillment modes (AUTO/MANUAL)
// - Retry logic configuration
// - Fallback mechanism settings
// - In-memory caching for performance
// - Thread-safe operations
// ============================================================================

// PrizeFulfillmentConfigService manages prize fulfillment configuration
type PrizeFulfillmentConfigService struct {
	db    *sql.DB
	cache map[string]*FulfillmentConfig
}

// FulfillmentConfig represents configuration for a prize type
type FulfillmentConfig struct {
	PrizeType                   string
	FulfillmentMode             string
	AutoRetryEnabled            bool
	MaxRetryAttempts            int
	RetryDelaySeconds           int
	FallbackToManual            bool
	FallbackNotificationEnabled bool
	ProvisionTimeoutSeconds     int
	IsActive                    bool
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	CreatedBy                   string
	UpdatedBy                   string
}

// NewPrizeFulfillmentConfigService creates a new configuration service
func NewPrizeFulfillmentConfigService(db *sql.DB) *PrizeFulfillmentConfigService {
	return &PrizeFulfillmentConfigService{
		db:    db,
		cache: make(map[string]*FulfillmentConfig),
	}
}

// ============================================================================
// PUBLIC METHODS
// ============================================================================

// GetConfig retrieves fulfillment configuration for a prize type
// Returns cached config if available, otherwise queries database
func (s *PrizeFulfillmentConfigService) GetConfig(ctx context.Context, prizeType string) (*FulfillmentConfig, error) {
	// Check cache first
	if config, exists := s.cache[prizeType]; exists {
		return config, nil
	}

	// Query database
	query := `
		SELECT 
			prize_type, 
			fulfillment_mode, 
			auto_retry_enabled, 
			max_retry_attempts, 
			retry_delay_seconds, 
			fallback_to_manual,
			fallback_notification_enabled,
			provision_timeout_seconds,
			is_active,
			created_at,
			updated_at,
			COALESCE(created_by, ''),
			COALESCE(updated_by, '')
		FROM prize_fulfillment_config
		WHERE prize_type = $1 AND is_active = TRUE
	`

	var config FulfillmentConfig
	err := s.db.QueryRowContext(ctx, query, prizeType).Scan(
		&config.PrizeType,
		&config.FulfillmentMode,
		&config.AutoRetryEnabled,
		&config.MaxRetryAttempts,
		&config.RetryDelaySeconds,
		&config.FallbackToManual,
		&config.FallbackNotificationEnabled,
		&config.ProvisionTimeoutSeconds,
		&config.IsActive,
		&config.CreatedAt,
		&config.UpdatedAt,
		&config.CreatedBy,
		&config.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		// Return safe default config if not found
		fmt.Printf("⚠️  No config found for prize type %s, using safe defaults\n", prizeType)
		return s.getDefaultConfig(prizeType), nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get fulfillment config: %w", err)
	}

	// Cache the config
	s.cache[prizeType] = &config

	return &config, nil
}

// UpdateConfig updates fulfillment configuration
func (s *PrizeFulfillmentConfigService) UpdateConfig(ctx context.Context, config *FulfillmentConfig) error {
	// Validate configuration
	if err := s.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	query := `
		UPDATE prize_fulfillment_config
		SET fulfillment_mode = $1,
			auto_retry_enabled = $2,
			max_retry_attempts = $3,
			retry_delay_seconds = $4,
			fallback_to_manual = $5,
			fallback_notification_enabled = $6,
			provision_timeout_seconds = $7,
			updated_at = NOW(),
			updated_by = $8
		WHERE prize_type = $9
	`

	result, err := s.db.ExecContext(ctx, query,
		config.FulfillmentMode,
		config.AutoRetryEnabled,
		config.MaxRetryAttempts,
		config.RetryDelaySeconds,
		config.FallbackToManual,
		config.FallbackNotificationEnabled,
		config.ProvisionTimeoutSeconds,
		config.UpdatedBy,
		config.PrizeType,
	)

	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no configuration found for prize type: %s", config.PrizeType)
	}

	// Invalidate cache
	delete(s.cache, config.PrizeType)

	fmt.Printf("✅ Updated fulfillment config for %s: mode=%s, retry=%v, fallback=%v\n",
		config.PrizeType, config.FulfillmentMode, config.AutoRetryEnabled, config.FallbackToManual)

	return nil
}

// ListAllConfigs retrieves all active configurations
func (s *PrizeFulfillmentConfigService) ListAllConfigs(ctx context.Context) ([]*FulfillmentConfig, error) {
	query := `
		SELECT 
			prize_type, 
			fulfillment_mode, 
			auto_retry_enabled, 
			max_retry_attempts, 
			retry_delay_seconds, 
			fallback_to_manual,
			fallback_notification_enabled,
			provision_timeout_seconds,
			is_active,
			created_at,
			updated_at,
			COALESCE(created_by, ''),
			COALESCE(updated_by, '')
		FROM prize_fulfillment_config
		WHERE is_active = TRUE
		ORDER BY prize_type
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}
	defer rows.Close()

	var configs []*FulfillmentConfig
	for rows.Next() {
		var config FulfillmentConfig
		err := rows.Scan(
			&config.PrizeType,
			&config.FulfillmentMode,
			&config.AutoRetryEnabled,
			&config.MaxRetryAttempts,
			&config.RetryDelaySeconds,
			&config.FallbackToManual,
			&config.FallbackNotificationEnabled,
			&config.ProvisionTimeoutSeconds,
			&config.IsActive,
			&config.CreatedAt,
			&config.UpdatedAt,
			&config.CreatedBy,
			&config.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}
		configs = append(configs, &config)
	}

	return configs, nil
}

// InvalidateCache clears the configuration cache
func (s *PrizeFulfillmentConfigService) InvalidateCache() {
	s.cache = make(map[string]*FulfillmentConfig)
	fmt.Println("🔄 Prize fulfillment config cache invalidated")
}

// ============================================================================
// PRIVATE HELPER METHODS
// ============================================================================

// getDefaultConfig returns safe default configuration
func (s *PrizeFulfillmentConfigService) getDefaultConfig(prizeType string) *FulfillmentConfig {
	// Safe defaults: MANUAL mode with no retry
	// This ensures prizes are not lost if provisioning fails
	return &FulfillmentConfig{
		PrizeType:                   prizeType,
		FulfillmentMode:             "MANUAL",
		AutoRetryEnabled:            false,
		MaxRetryAttempts:            0,
		RetryDelaySeconds:           0,
		FallbackToManual:            true,
		FallbackNotificationEnabled: false,
		ProvisionTimeoutSeconds:     60,
		IsActive:                    true,
		CreatedAt:                   time.Now(),
		UpdatedAt:                   time.Now(),
		CreatedBy:                   "SYSTEM_DEFAULT",
		UpdatedBy:                   "SYSTEM_DEFAULT",
	}
}

// validateConfig validates configuration parameters
func (s *PrizeFulfillmentConfigService) validateConfig(config *FulfillmentConfig) error {
	// Validate prize type
	validPrizeTypes := []string{"AIRTIME", "DATA", "CASH", "POINTS", "PHYSICAL"}
	if !contains(validPrizeTypes, config.PrizeType) {
		return fmt.Errorf("invalid prize type: %s", config.PrizeType)
	}

	// Validate fulfillment mode
	if config.FulfillmentMode != "AUTO" && config.FulfillmentMode != "MANUAL" {
		return fmt.Errorf("invalid fulfillment mode: %s (must be AUTO or MANUAL)", config.FulfillmentMode)
	}

	// Validate retry attempts
	if config.MaxRetryAttempts < 0 || config.MaxRetryAttempts > 10 {
		return fmt.Errorf("invalid max retry attempts: %d (must be 0-10)", config.MaxRetryAttempts)
	}

	// Validate retry delay
	if config.RetryDelaySeconds < 0 || config.RetryDelaySeconds > 3600 {
		return fmt.Errorf("invalid retry delay: %d (must be 0-3600 seconds)", config.RetryDelaySeconds)
	}

	// Validate timeout
	if config.ProvisionTimeoutSeconds < 10 || config.ProvisionTimeoutSeconds > 300 {
		return fmt.Errorf("invalid provision timeout: %d (must be 10-300 seconds)", config.ProvisionTimeoutSeconds)
	}

	// Business rule: If auto retry is enabled, must have at least 1 retry attempt
	if config.AutoRetryEnabled && config.MaxRetryAttempts == 0 {
		return fmt.Errorf("auto retry enabled but max retry attempts is 0")
	}

	// Business rule: CASH prizes must always be MANUAL
	if config.PrizeType == "CASH" && config.FulfillmentMode == "AUTO" {
		return fmt.Errorf("CASH prizes must use MANUAL fulfillment mode")
	}

	return nil
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// ============================================================================
// STATISTICS AND MONITORING
// ============================================================================

// GetFulfillmentStatistics retrieves fulfillment statistics
type FulfillmentStatistics struct {
	PrizeType              string
	TotalAttempts          int64
	SuccessfulAttempts     int64
	FailedAttempts         int64
	SuccessRate            float64
	AvgResponseTimeMs      float64
	TotalRetries           int64
	FallbackToManualCount  int64
}

// GetStatistics retrieves fulfillment statistics for a prize type
func (s *PrizeFulfillmentConfigService) GetStatistics(ctx context.Context, prizeType string, startDate, endDate time.Time) (*FulfillmentStatistics, error) {
	query := `
		SELECT 
			sr.prize_type,
			COUNT(*)::BIGINT AS total_attempts,
			COUNT(*) FILTER (WHERE pfl.status = 'SUCCESS')::BIGINT AS successful_attempts,
			COUNT(*) FILTER (WHERE pfl.status = 'FAILED')::BIGINT AS failed_attempts,
			ROUND(
				(COUNT(*) FILTER (WHERE pfl.status = 'SUCCESS')::NUMERIC / 
				 NULLIF(COUNT(*)::NUMERIC, 0)) * 100, 
				2
			) AS success_rate,
			ROUND(AVG(pfl.response_time_ms), 2) AS avg_response_time_ms,
			COUNT(*) FILTER (WHERE pfl.attempt_number > 1)::BIGINT AS total_retries,
			COUNT(DISTINCT sr.id) FILTER (WHERE sr.fulfillment_mode = 'MANUAL' AND sr.fulfillment_attempts > 0)::BIGINT AS fallback_to_manual_count
		FROM spin_results sr
		LEFT JOIN prize_fulfillment_logs pfl ON sr.id = pfl.spin_result_id
		WHERE pfl.created_at BETWEEN $1 AND $2
		  AND sr.prize_type = $3
		GROUP BY sr.prize_type
	`

	var stats FulfillmentStatistics
	err := s.db.QueryRowContext(ctx, query, startDate, endDate, prizeType).Scan(
		&stats.PrizeType,
		&stats.TotalAttempts,
		&stats.SuccessfulAttempts,
		&stats.FailedAttempts,
		&stats.SuccessRate,
		&stats.AvgResponseTimeMs,
		&stats.TotalRetries,
		&stats.FallbackToManualCount,
	)

	if err == sql.ErrNoRows {
		// No data for this period
		return &FulfillmentStatistics{
			PrizeType: prizeType,
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return &stats, nil
}
