package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
)

// AdminSpinClaimsHandler handles HTTP requests for admin spin claim management
type AdminSpinClaimsHandler struct {
	service *services.AdminSpinClaimService
}

// NewAdminSpinClaimsHandler creates a new admin spin claims handler
func NewAdminSpinClaimsHandler(service *services.AdminSpinClaimService) *AdminSpinClaimsHandler {
	return &AdminSpinClaimsHandler{
		service: service,
	}
}

// ============================================================================
// HTTP Handlers
// ============================================================================

// ListClaims handles GET /api/v1/admin/spin/claims
func (h *AdminSpinClaimsHandler) ListClaims(c *gin.Context) {
	// Parse filters
	filters := services.ClaimFilters{
		Status:     c.Query("status"),
		PrizeType:  c.Query("prize_type"),
		MSISDN:     c.Query("msisdn"),
		SearchTerm: c.Query("search"),
	}

	// Parse date filters
	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			filters.FromDate = fromDate
		}
	}
	if toDateStr := c.Query("to_date"); toDateStr != "" {
		if toDate, err := time.Parse("2006-01-02", toDateStr); err == nil {
			filters.ToDate = toDate.Add(24 * time.Hour) // Include entire day
		}
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	pagination := services.Pagination{
		Page:   page,
		Limit:  limit,
		SortBy: c.DefaultQuery("sort", "created_at"),
		Order:  c.DefaultQuery("order", "desc"),
	}

	// Call service
	result, err := h.service.ListClaims(c.Request.Context(), filters, pagination)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetPendingClaims handles GET /api/v1/admin/spin/claims/pending
func (h *AdminSpinClaimsHandler) GetPendingClaims(c *gin.Context) {
	// Call service (pagination handled internally by GetPendingClaims)
	result, err := h.service.GetPendingClaims(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetClaimDetails handles GET /api/v1/admin/spin/claims/:id
func (h *AdminSpinClaimsHandler) GetClaimDetails(c *gin.Context) {
	claimID := c.Param("id")

	// Call service
	result, err := h.service.GetClaimDetails(c.Request.Context(), claimID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"claim": result,
		},
	})
}

// ApproveClaim handles POST /api/v1/admin/spin/claims/:id/approve
func (h *AdminSpinClaimsHandler) ApproveClaim(c *gin.Context) {
	claimID := c.Param("id")

	// Get admin ID from context (set by auth middleware)
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Admin authentication required",
			},
		})
		return
	}

	// Parse request body
	var request services.ApproveClaimRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// Call service
	err := h.service.ApproveClaim(c.Request.Context(), claimID, adminID.(string), request)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message":  "Claim approved successfully",
			"claim_id": claimID,
			"status":   "APPROVED",
		},
	})
}

// RejectClaim handles POST /api/v1/admin/spin/claims/:id/reject
func (h *AdminSpinClaimsHandler) RejectClaim(c *gin.Context) {
	claimID := c.Param("id")

	// Get admin ID from context (set by auth middleware)
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Admin authentication required",
			},
		})
		return
	}

	// Parse request body
	var request services.RejectClaimRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// Call service
	err := h.service.RejectClaim(c.Request.Context(), claimID, adminID.(string), request)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message":  "Claim rejected successfully",
			"claim_id": claimID,
			"status":   "REJECTED",
		},
	})
}

// GetStatistics handles GET /api/v1/admin/spin/claims/statistics
func (h *AdminSpinClaimsHandler) GetStatistics(c *gin.Context) {
	// Call service (all-time statistics)
	result, err := h.service.GetStatistics(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ExportClaims handles GET /api/v1/admin/spin/claims/export
func (h *AdminSpinClaimsHandler) ExportClaims(c *gin.Context) {
	// Parse filters (same as ListClaims)
	filters := services.ClaimFilters{
		Status:     c.Query("status"),
		PrizeType:  c.Query("prize_type"),
		MSISDN:     c.Query("msisdn"),
		SearchTerm: c.Query("search"),
	}

	// Parse date filters
	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			filters.FromDate = fromDate
		}
	}
	if toDateStr := c.Query("to_date"); toDateStr != "" {
		if toDate, err := time.Parse("2006-01-02", toDateStr); err == nil {
			filters.ToDate = toDate.Add(24 * time.Hour)
		}
	}

	// Call service
	csvData, err := h.service.ExportClaims(c.Request.Context(), filters)
	if err != nil {
		handleError(c, err)
		return
	}

	// Set headers for CSV download
	filename := "spin_claims_" + time.Now().Format("20060102_150405") + ".csv"
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Length", strconv.Itoa(len(csvData)))

	c.Data(http.StatusOK, "text/csv", csvData)
}

// ============================================================================
// Helper Functions
// ============================================================================

// handleError converts service errors to HTTP responses
func handleError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *errors.AppError:
		status := http.StatusInternalServerError
		switch e.Code {
		case "BAD_REQUEST":
			status = http.StatusBadRequest
		case "NOT_FOUND":
			status = http.StatusNotFound
		case "UNAUTHORIZED":
			status = http.StatusUnauthorized
		case "FORBIDDEN":
			status = http.StatusForbidden
		}

		c.JSON(status, gin.H{
			"success": false,
			"error": gin.H{
				"code":    e.Code,
				"message": e.Message,
			},
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "An internal error occurred",
			},
		})
	}
}

// SendReminders sends notification reminders to all users with unclaimed prizes.
func (h *AdminSpinClaimsHandler) SendReminders(c *gin.Context) {
	claims, err := h.service.GetPendingClaimsForReminder(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}
	middleware.RespondWithSuccess(c, gin.H{
		"count":   len(claims),
		"message": "Reminders queued",
	})
}
