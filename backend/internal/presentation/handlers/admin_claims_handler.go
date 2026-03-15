package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WINNER CLAIM PROCESSING
// ============================================================================

// GetPendingClaims returns pending winner claims
func (h *AdminComprehensiveHandler) GetPendingClaims(c *gin.Context) {
	ctx := c.Request.Context()

	// Get all winners (page 1, 100 per page, no draw filter)
	winners, _, err := h.winnerService.GetAllWinners(ctx, 1, 100, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve pending claims",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    winners,
	})
}

// ApproveWinnerClaim approves a winner claim
func (h *AdminComprehensiveHandler) ApproveWinnerClaim(c *gin.Context) {
	winnerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid winner ID",
		})
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	err = h.winnerService.ApproveClaim(ctx, winnerID.String(), req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to approve claim",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Winner claim approved successfully",
	})
}

// RejectWinnerClaim rejects a winner claim
func (h *AdminComprehensiveHandler) RejectWinnerClaim(c *gin.Context) {
	winnerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid winner ID",
		})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	err = h.winnerService.RejectClaim(ctx, winnerID.String(), req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to reject claim",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Winner claim rejected successfully",
	})
}

// GetClaimStatistics returns claim statistics
func (h *AdminComprehensiveHandler) GetClaimStatistics(c *gin.Context) {
	// TODO: Implement claim statistics aggregation
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_claims":    0,
			"pending_claims":  0,
			"approved_claims": 0,
			"rejected_claims": 0,
		},
	})
}
