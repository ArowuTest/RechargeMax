package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"rechargemax/internal/application/services"
)

// CommissionHandler handles admin commission-reconciliation endpoints.
type CommissionHandler struct {
	commissionSvc *services.CommissionService
}

// NewCommissionHandler creates a new CommissionHandler.
func NewCommissionHandler(commissionSvc *services.CommissionService) *CommissionHandler {
	return &CommissionHandler{commissionSvc: commissionSvc}
}

// commissionRequest is the shared body for both reconciliation and CSV export.
type commissionRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	Network   string `json:"network"`
	Provider  string `json:"provider"`
}

// GetCommissionReconciliation returns the full reconciliation report.
func (h *CommissionHandler) GetCommissionReconciliation(c *gin.Context) {
	var req commissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	report, err := h.commissionSvc.GetReconciliation(c.Request.Context(), services.CommissionFilter{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Network:   req.Network,
		Provider:  req.Provider,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": report})
}

// ExportCommissionReport streams the report as a CSV download.
func (h *CommissionHandler) ExportCommissionReport(c *gin.Context) {
	var req commissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	csvBytes, err := h.commissionSvc.ExportCSV(c.Request.Context(), services.CommissionFilter{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Network:   req.Network,
		Provider:  req.Provider,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	filename := fmt.Sprintf("commission_report_%s_to_%s.csv", req.StartDate, req.EndDate)
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "text/csv", csvBytes)
}
