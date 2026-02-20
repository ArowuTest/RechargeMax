package handlers

import (
	"fmt"
	
	"github.com/gin-gonic/gin"
	
	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
)

// TestHandler handles test/debug endpoints
type TestHandler struct {
	rechargeService     *services.RechargeService
	subscriptionService *services.SubscriptionService
}

// NewTestHandler creates a new test handler
func NewTestHandler(rechargeService *services.RechargeService, subscriptionService *services.SubscriptionService) *TestHandler {
	return &TestHandler{
		rechargeService:     rechargeService,
		subscriptionService: subscriptionService,
	}
}

// ProcessPaymentManually godoc
// @Summary Manually process a payment (TEST ONLY)
// @Description Manually trigger payment processing for testing without webhook
// @Tags test
// @Accept json
// @Produce json
// @Param request body map[string]string true "Payment reference"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /test/process-payment [post]
func (h *TestHandler) ProcessPaymentManually(c *gin.Context) {
	var req struct {
		Reference string `json:"reference" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}
	
	fmt.Printf("[TEST] Manually processing payment for reference: %s\n", req.Reference)
	
	// Determine transaction type from reference prefix
	if len(req.Reference) >= 4 {
		prefix := req.Reference[:4]
		
		switch prefix {
		case "RCH_":
			// Process recharge
			fmt.Printf("[TEST] Processing recharge payment...\n")
			if err := h.rechargeService.ProcessSuccessfulPayment(c.Request.Context(), req.Reference); err != nil {
				fmt.Printf("[TEST ERROR] Failed to process recharge: %v\n", err)
				middleware.RespondWithError(c, errors.Internal("Failed to process payment: "+err.Error()))
				return
			}
			fmt.Printf("[TEST] ✅ Recharge payment processed successfully\n")
			
		case "SUB_":
			// Process subscription
			fmt.Printf("[TEST] Processing subscription payment...\n")
			if err := h.subscriptionService.ProcessSuccessfulPayment(c.Request.Context(), req.Reference); err != nil {
				fmt.Printf("[TEST ERROR] Failed to process subscription: %v\n", err)
				middleware.RespondWithError(c, errors.Internal("Failed to process payment: "+err.Error()))
				return
			}
			fmt.Printf("[TEST] ✅ Subscription payment processed successfully\n")
			
		default:
			middleware.RespondWithError(c, errors.BadRequest("Unknown transaction type"))
			return
		}
	} else {
		middleware.RespondWithError(c, errors.BadRequest("Invalid reference format"))
		return
	}
	
	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message":   "Payment processed successfully",
		"reference": req.Reference,
	})
}
