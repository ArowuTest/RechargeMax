package handlers

import (
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ValidationStatsHandler struct {
	db *gorm.DB
}

func NewValidationStatsHandler(db *gorm.DB) *ValidationStatsHandler {
	return &ValidationStatsHandler{db: db}
}

// ValidationStatsRequest represents the request for validation statistics
type ValidationStatsRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

// ValidationStatsResponse represents the validation statistics data
type ValidationStatsResponse struct {
	Summary struct {
		TotalValidations      int64   `json:"total_validations"`
		SuccessfulValidations int64   `json:"successful_validations"`
		FailedValidations     int64   `json:"failed_validations"`
		SuccessRate           float64 `json:"success_rate"`
		MismatchRate          float64 `json:"mismatch_rate"`
	} `json:"summary"`

	ValidationSources struct {
		HLRAPICount      int64   `json:"hlr_api_count"`
		PrefixCount      int64   `json:"prefix_count"`
		CacheCount       int64   `json:"cache_count"`
		HLRAPIPercentage float64 `json:"hlr_api_percentage"`
		PrefixPercentage float64 `json:"prefix_percentage"`
		CachePercentage  float64 `json:"cache_percentage"`
	} `json:"validation_sources"`

	ByNetwork []struct {
		Network               string   `json:"network"`
		TotalValidations      int64    `json:"total_validations"`
		SuccessfulValidations int64    `json:"successful_validations"`
		FailedValidations     int64    `json:"failed_validations"`
		SuccessRate           float64  `json:"success_rate"`
		CommonMismatches      []string `json:"common_mismatches"`
	} `json:"by_network"`

	MismatchPatterns []struct {
		SelectedNetwork string  `json:"selected_network"`
		ActualNetwork   string  `json:"actual_network"`
		Count           int64   `json:"count"`
		Percentage      float64 `json:"percentage"`
	} `json:"mismatch_patterns"`

	ValidationTrend []struct {
		Date                  string  `json:"date"`
		TotalValidations      int64   `json:"total_validations"`
		SuccessfulValidations int64   `json:"successful_validations"`
		FailedValidations     int64   `json:"failed_validations"`
		SuccessRate           float64 `json:"success_rate"`
	} `json:"validation_trend"`
}

// GetValidationStats returns real network validation statistics from DB
func (h *ValidationStatsHandler) GetValidationStats(c *gin.Context) {
	var req ValidationStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}
	endDate = endDate.Add(24*time.Hour - time.Second)

	response := ValidationStatsResponse{}

	// ── Summary: derive from transactions table ────────────────────────────────
	// "validation" = any transaction that ran through network detection.
	// Success = status SUCCESS; Failed = status FAILED.
	type summaryRow struct {
		TotalValidations      int64
		SuccessfulValidations int64
		FailedValidations     int64
	}
	var sumRow summaryRow
	h.db.Table("transactions").
		Select(`
			COUNT(*) AS total_validations,
			SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) AS successful_validations,
			SUM(CASE WHEN status = 'FAILED'  THEN 1 ELSE 0 END) AS failed_validations
		`).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&sumRow)

	response.Summary.TotalValidations = sumRow.TotalValidations
	response.Summary.SuccessfulValidations = sumRow.SuccessfulValidations
	response.Summary.FailedValidations = sumRow.FailedValidations
	if sumRow.TotalValidations > 0 {
		response.Summary.SuccessRate = math.Round(float64(sumRow.SuccessfulValidations)/float64(sumRow.TotalValidations)*10000) / 100
		response.Summary.MismatchRate = math.Round(float64(sumRow.FailedValidations)/float64(sumRow.TotalValidations)*10000) / 100
	}

	// ── Validation Sources: from network_cache table ───────────────────────────
	type srcRow struct {
		LookupSource string
		Count        int64
	}
	var srcRows []srcRow
	h.db.Table("network_cache").
		Select("lookup_source, COUNT(*) AS count").
		Where("last_verified BETWEEN ? AND ?", startDate, endDate).
		Group("lookup_source").
		Scan(&srcRows)

	var totalSrc int64
	for _, r := range srcRows {
		totalSrc += r.Count
	}
	for _, r := range srcRows {
		pct := float64(0)
		if totalSrc > 0 {
			pct = math.Round(float64(r.Count)/float64(totalSrc)*10000) / 100
		}
		switch r.LookupSource {
		case "hlr_api":
			response.ValidationSources.HLRAPICount = r.Count
			response.ValidationSources.HLRAPIPercentage = pct
		case "prefix_fallback":
			response.ValidationSources.PrefixCount = r.Count
			response.ValidationSources.PrefixPercentage = pct
		case "cache", "user_selection":
			response.ValidationSources.CacheCount += r.Count
			response.ValidationSources.CachePercentage += pct
		}
	}

	// ── By Network ─────────────────────────────────────────────────────────────
	type netRow struct {
		Network               string
		TotalValidations      int64
		SuccessfulValidations int64
		FailedValidations     int64
	}
	var netRows []netRow
	h.db.Table("transactions").
		Select(`
			network,
			COUNT(*) AS total_validations,
			SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) AS successful_validations,
			SUM(CASE WHEN status = 'FAILED'  THEN 1 ELSE 0 END) AS failed_validations
		`).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("network").
		Scan(&netRows)

	for _, r := range netRows {
		rate := float64(0)
		if r.TotalValidations > 0 {
			rate = math.Round(float64(r.SuccessfulValidations)/float64(r.TotalValidations)*10000) / 100
		}
		response.ByNetwork = append(response.ByNetwork, struct {
			Network               string   `json:"network"`
			TotalValidations      int64    `json:"total_validations"`
			SuccessfulValidations int64    `json:"successful_validations"`
			FailedValidations     int64    `json:"failed_validations"`
			SuccessRate           float64  `json:"success_rate"`
			CommonMismatches      []string `json:"common_mismatches"`
		}{
			Network: r.Network, TotalValidations: r.TotalValidations,
			SuccessfulValidations: r.SuccessfulValidations, FailedValidations: r.FailedValidations,
			SuccessRate: rate, CommonMismatches: []string{},
		})
	}

	// ── Mismatch Patterns ─────────────────────────────────────────────────────
	// Mismatch = transaction where user-selected network != HLR-detected network.
	// Stored in transactions as network (selected) vs detected_network columns (if present).
	type mismatchRow struct {
		SelectedNetwork string
		ActualNetwork   string
		Count           int64
	}
	var mismatchRows []mismatchRow
	// Only query if detected_network column exists to avoid runtime errors
	var colExists int64
	h.db.Raw(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name='transactions' AND column_name='detected_network'`).Scan(&colExists)
	if colExists > 0 {
		var totalMismatch int64
		h.db.Table("transactions").
			Where("created_at BETWEEN ? AND ? AND detected_network IS NOT NULL AND network != detected_network", startDate, endDate).
			Count(&totalMismatch)

		h.db.Table("transactions").
			Select("network AS selected_network, detected_network AS actual_network, COUNT(*) AS count").
			Where("created_at BETWEEN ? AND ? AND detected_network IS NOT NULL AND network != detected_network", startDate, endDate).
			Group("network, detected_network").
			Order("count DESC").
			Limit(10).
			Scan(&mismatchRows)

		for _, r := range mismatchRows {
			pct := float64(0)
			if totalMismatch > 0 {
				pct = math.Round(float64(r.Count)/float64(totalMismatch)*10000) / 100
			}
			response.MismatchPatterns = append(response.MismatchPatterns, struct {
				SelectedNetwork string  `json:"selected_network"`
				ActualNetwork   string  `json:"actual_network"`
				Count           int64   `json:"count"`
				Percentage      float64 `json:"percentage"`
			}{SelectedNetwork: r.SelectedNetwork, ActualNetwork: r.ActualNetwork, Count: r.Count, Percentage: pct})
		}
	}
	if response.MismatchPatterns == nil {
		response.MismatchPatterns = []struct {
			SelectedNetwork string  `json:"selected_network"`
			ActualNetwork   string  `json:"actual_network"`
			Count           int64   `json:"count"`
			Percentage      float64 `json:"percentage"`
		}{}
	}

	// ── Validation Trend (by day) ──────────────────────────────────────────────
	type trendRow struct {
		Day                   string
		TotalValidations      int64
		SuccessfulValidations int64
		FailedValidations     int64
	}
	var trendRows []trendRow
	h.db.Table("transactions").
		Select(`
			DATE(created_at) AS day,
			COUNT(*) AS total_validations,
			SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) AS successful_validations,
			SUM(CASE WHEN status = 'FAILED'  THEN 1 ELSE 0 END) AS failed_validations
		`).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("day ASC").
		Scan(&trendRows)

	for _, r := range trendRows {
		rate := float64(0)
		if r.TotalValidations > 0 {
			rate = math.Round(float64(r.SuccessfulValidations)/float64(r.TotalValidations)*10000) / 100
		}
		response.ValidationTrend = append(response.ValidationTrend, struct {
			Date                  string  `json:"date"`
			TotalValidations      int64   `json:"total_validations"`
			SuccessfulValidations int64   `json:"successful_validations"`
			FailedValidations     int64   `json:"failed_validations"`
			SuccessRate           float64 `json:"success_rate"`
		}{
			Date: r.Day, TotalValidations: r.TotalValidations,
			SuccessfulValidations: r.SuccessfulValidations, FailedValidations: r.FailedValidations,
			SuccessRate: rate,
		})
	}
	if response.ValidationTrend == nil {
		response.ValidationTrend = []struct {
			Date                  string  `json:"date"`
			TotalValidations      int64   `json:"total_validations"`
			SuccessfulValidations int64   `json:"successful_validations"`
			FailedValidations     int64   `json:"failed_validations"`
			SuccessRate           float64 `json:"success_rate"`
		}{}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}
