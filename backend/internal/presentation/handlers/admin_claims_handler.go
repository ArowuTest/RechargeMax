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

// GetClaimStatistics returns claim statistics aggregated from the winners table
func (h *AdminComprehensiveHandler) GetClaimStatistics(c *gin.Context) {
	ctx := c.Request.Context()

	type claimStat struct {
		ClaimStatus string `gorm:"column:claim_status"`
		Count       int64  `gorm:"column:count"`
	}
	var stats []claimStat
	// Table is draw_winners (not 'winners')
	if err := h.db.WithContext(ctx).
		Table("draw_winners").
		Select("claim_status, COUNT(*) as count").
		Group("claim_status").
		Scan(&stats).Error; err != nil {
		// Return zeroed stats rather than 500 — empty draw history is valid
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"total_claims":    int64(0),
				"pending_claims":  int64(0),
				"approved_claims": int64(0),
				"rejected_claims": int64(0),
			},
		})
		return
	}

	result := gin.H{
		"total_claims":    int64(0),
		"pending_claims":  int64(0),
		"approved_claims": int64(0),
		"rejected_claims": int64(0),
	}
	var total int64
	for _, s := range stats {
		total += s.Count
		switch s.ClaimStatus {
		case "PENDING":
			result["pending_claims"] = s.Count
		case "CLAIMED":
			result["approved_claims"] = s.Count
		case "REJECTED":
			result["rejected_claims"] = s.Count
		case "EXPIRED":
			result["expired_claims"] = s.Count
		}
	}
	result["total_claims"] = total

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
