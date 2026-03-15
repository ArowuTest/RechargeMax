package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rechargemax/internal/application/services"
)

// ValidationStatsHandler handles admin network-validation analytics.
type ValidationStatsHandler struct {
	validationSvc *services.ValidationStatsService
}

// NewValidationStatsHandler creates a new ValidationStatsHandler.
func NewValidationStatsHandler(validationSvc *services.ValidationStatsService) *ValidationStatsHandler {
	return &ValidationStatsHandler{validationSvc: validationSvc}
}

// validationStatsRequest is the request body.
type validationStatsRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

// GetValidationStats returns the network-validation statistics report.
func (h *ValidationStatsHandler) GetValidationStats(c *gin.Context) {
	var req validationStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	stats, err := h.validationSvc.GetStats(c.Request.Context(), services.ValidationStatsFilter{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}
