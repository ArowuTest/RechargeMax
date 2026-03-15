package services

import (
	"context"
	"math"
	"time"

	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// DTOs
// ─────────────────────────────────────────────────────────────────────────────

// ValidationStatsFilter constrains the report window.
type ValidationStatsFilter struct {
	StartDate string // YYYY-MM-DD
	EndDate   string // YYYY-MM-DD
}

// ValidationStats is the full network-validation statistics payload.
type ValidationStats struct {
	Summary            ValidationSummary            `json:"summary"`
	ValidationSources  ValidationSources             `json:"validation_sources"`
	ByNetwork          []ValidationByNetwork         `json:"by_network"`
	MismatchPatterns   []ValidationMismatch          `json:"mismatch_patterns"`
	ValidationTrend    []ValidationTrendPoint        `json:"validation_trend"`
}

// ValidationSummary is the top-level aggregate.
type ValidationSummary struct {
	TotalValidations      int64   `json:"total_validations"`
	SuccessfulValidations int64   `json:"successful_validations"`
	FailedValidations     int64   `json:"failed_validations"`
	SuccessRate           float64 `json:"success_rate"`
	MismatchRate          float64 `json:"mismatch_rate"`
}

// ValidationSources breaks down lookups by resolution method.
type ValidationSources struct {
	HLRAPICount      int64   `json:"hlr_api_count"`
	PrefixCount      int64   `json:"prefix_count"`
	CacheCount       int64   `json:"cache_count"`
	HLRAPIPercentage float64 `json:"hlr_api_percentage"`
	PrefixPercentage float64 `json:"prefix_percentage"`
	CachePercentage  float64 `json:"cache_percentage"`
}

// ValidationByNetwork is per-operator aggregation.
type ValidationByNetwork struct {
	Network               string   `json:"network"`
	TotalValidations      int64    `json:"total_validations"`
	SuccessfulValidations int64    `json:"successful_validations"`
	FailedValidations     int64    `json:"failed_validations"`
	SuccessRate           float64  `json:"success_rate"`
	CommonMismatches      []string `json:"common_mismatches"`
}

// ValidationMismatch is a selected-vs-actual network mismatch pattern.
type ValidationMismatch struct {
	SelectedNetwork string  `json:"selected_network"`
	ActualNetwork   string  `json:"actual_network"`
	Count           int64   `json:"count"`
	Percentage      float64 `json:"percentage"`
}

// ValidationTrendPoint is one day's bucket.
type ValidationTrendPoint struct {
	Date                  string  `json:"date"`
	TotalValidations      int64   `json:"total_validations"`
	SuccessfulValidations int64   `json:"successful_validations"`
	FailedValidations     int64   `json:"failed_validations"`
	SuccessRate           float64 `json:"success_rate"`
}

// ─────────────────────────────────────────────────────────────────────────────
// ValidationStatsService
// ─────────────────────────────────────────────────────────────────────────────

// ValidationStatsService runs network-validation analytics queries.
type ValidationStatsService struct {
	db *gorm.DB
}

// NewValidationStatsService constructs a ValidationStatsService.
func NewValidationStatsService(db *gorm.DB) *ValidationStatsService {
	return &ValidationStatsService{db: db}
}

// GetStats returns the full validation statistics report.
func (s *ValidationStatsService) GetStats(ctx context.Context, f ValidationStatsFilter) (*ValidationStats, error) {
	start, end, err := parseDateRange(f.StartDate, f.EndDate)
	if err != nil {
		return nil, err
	}
	db := s.db.WithContext(ctx)
	stats := &ValidationStats{}

	// ── Summary ───────────────────────────────────────────────────────────────
	type summaryRow struct {
		TotalValidations      int64
		SuccessfulValidations int64
		FailedValidations     int64
	}
	var sum summaryRow
	db.Table("transactions").
		Select(`COUNT(*) AS total_validations,
			SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) AS successful_validations,
			SUM(CASE WHEN status = 'FAILED'  THEN 1 ELSE 0 END) AS failed_validations`).
		Where("created_at BETWEEN ? AND ?", start, end).
		Scan(&sum)

	stats.Summary = ValidationSummary{
		TotalValidations:      sum.TotalValidations,
		SuccessfulValidations: sum.SuccessfulValidations,
		FailedValidations:     sum.FailedValidations,
	}
	if sum.TotalValidations > 0 {
		stats.Summary.SuccessRate = pct(sum.SuccessfulValidations, sum.TotalValidations)
		stats.Summary.MismatchRate = pct(sum.FailedValidations, sum.TotalValidations)
	}

	// ── Validation Sources ────────────────────────────────────────────────────
	type srcRow struct {
		LookupSource string
		Count        int64
	}
	var srcRows []srcRow
	db.Table("network_cache").
		Select("lookup_source, COUNT(*) AS count").
		Where("last_verified BETWEEN ? AND ?", start, end).
		Group("lookup_source").
		Scan(&srcRows)

	var totalSrc int64
	for _, r := range srcRows {
		totalSrc += r.Count
	}
	for _, r := range srcRows {
		p := pct(r.Count, totalSrc)
		switch r.LookupSource {
		case "hlr_api":
			stats.ValidationSources.HLRAPICount = r.Count
			stats.ValidationSources.HLRAPIPercentage = p
		case "prefix_fallback":
			stats.ValidationSources.PrefixCount = r.Count
			stats.ValidationSources.PrefixPercentage = p
		default:
			stats.ValidationSources.CacheCount += r.Count
			stats.ValidationSources.CachePercentage += p
		}
	}

	// ── By Network ────────────────────────────────────────────────────────────
	type netRow struct {
		Network               string
		TotalValidations      int64
		SuccessfulValidations int64
		FailedValidations     int64
	}
	var netRows []netRow
	db.Table("transactions").
		Select(`network,
			COUNT(*) AS total_validations,
			SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) AS successful_validations,
			SUM(CASE WHEN status = 'FAILED'  THEN 1 ELSE 0 END) AS failed_validations`).
		Where("created_at BETWEEN ? AND ?", start, end).
		Group("network").
		Scan(&netRows)
	for _, r := range netRows {
		stats.ByNetwork = append(stats.ByNetwork, ValidationByNetwork{
			Network:               r.Network,
			TotalValidations:      r.TotalValidations,
			SuccessfulValidations: r.SuccessfulValidations,
			FailedValidations:     r.FailedValidations,
			SuccessRate:           pct(r.SuccessfulValidations, r.TotalValidations),
			CommonMismatches:      []string{},
		})
	}

	// ── Mismatch Patterns (PostgreSQL: detected_network column if it exists) ──
	var colExists int64
	db.Raw(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name='transactions' AND column_name='detected_network'`).
		Scan(&colExists)
	if colExists > 0 {
		var totalMismatch int64
		db.Table("transactions").
			Where("created_at BETWEEN ? AND ? AND detected_network IS NOT NULL AND network != detected_network", start, end).
			Count(&totalMismatch)

		type mismatchRow struct {
			SelectedNetwork string
			ActualNetwork   string
			Count           int64
		}
		var mismatchRows []mismatchRow
		db.Table("transactions").
			Select("network AS selected_network, detected_network AS actual_network, COUNT(*) AS count").
			Where("created_at BETWEEN ? AND ? AND detected_network IS NOT NULL AND network != detected_network", start, end).
			Group("network, detected_network").
			Order("count DESC").
			Limit(10).
			Scan(&mismatchRows)
		for _, r := range mismatchRows {
			stats.MismatchPatterns = append(stats.MismatchPatterns, ValidationMismatch{
				SelectedNetwork: r.SelectedNetwork,
				ActualNetwork:   r.ActualNetwork,
				Count:           r.Count,
				Percentage:      pct(r.Count, totalMismatch),
			})
		}
	}
	if stats.MismatchPatterns == nil {
		stats.MismatchPatterns = []ValidationMismatch{}
	}

	// ── Validation Trend ──────────────────────────────────────────────────────
	type trendRow struct {
		Day                   string
		TotalValidations      int64
		SuccessfulValidations int64
		FailedValidations     int64
	}
	var trendRows []trendRow
	db.Table("transactions").
		Select(`DATE(created_at) AS day,
			COUNT(*) AS total_validations,
			SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) AS successful_validations,
			SUM(CASE WHEN status = 'FAILED'  THEN 1 ELSE 0 END) AS failed_validations`).
		Where("created_at BETWEEN ? AND ?", start, end).
		Group("DATE(created_at)").
		Order("day ASC").
		Scan(&trendRows)
	for _, r := range trendRows {
		stats.ValidationTrend = append(stats.ValidationTrend, ValidationTrendPoint{
			Date:                  r.Day,
			TotalValidations:      r.TotalValidations,
			SuccessfulValidations: r.SuccessfulValidations,
			FailedValidations:     r.FailedValidations,
			SuccessRate:           pct(r.SuccessfulValidations, r.TotalValidations),
		})
	}
	if stats.ValidationTrend == nil {
		stats.ValidationTrend = []ValidationTrendPoint{}
	}

	return stats, nil
}

// pct computes a rounded percentage: (num/den)*100, returns 0 when den=0.
func pct(num, den int64) float64 {
	if den == 0 {
		return 0
	}
	return math.Round(float64(num)/float64(den)*10000) / 100
}

// Ensure the time import is used.
var _ = time.Now
