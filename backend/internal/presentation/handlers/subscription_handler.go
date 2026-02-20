package handlers

import (
	"github.com/gin-gonic/gin"
	
	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
	"rechargemax/internal/validation"
)

type SubscriptionHandler struct {
	subscriptionService *services.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService *services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionService: subscriptionService}
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Create a new daily subscription for the user
// @Tags subscription
// @Accept json
// @Produce json
// @Param request body validation.SubscribeRequest true "Subscription request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /subscription [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	var req validation.SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Use authenticated user's MSISDN
	req.MSISDN = msisdn

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Service will validate subscription eligibility
	result, err := h.subscriptionService.CreateSubscription(c.Request.Context(), services.CreateSubscriptionRequest{
		MSISDN: msisdn,
	})

	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log subscription creation
	errors.Info("Subscription created", map[string]interface{}{
		"msisdn":          msisdn,
		"subscription_id": result.ID,
	})

	middleware.RespondWithSuccess(c, result)
}

// GetSubscription godoc
// @Summary Get active subscription
// @Description Get user's active subscription details
// @Tags subscription
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /subscription [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	subscription, err := h.subscriptionService.GetSubscription(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, errors.NotFound("No active subscription found"))
		return
	}

	middleware.RespondWithSuccess(c, subscription)
}

// CancelSubscription godoc
// @Summary Cancel subscription
// @Description Cancel user's active subscription
// @Tags subscription
// @Accept json
// @Produce json
// @Param request body validation.CancelSubscriptionRequest true "Cancel request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /subscription/cancel [post]
func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	var req validation.CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Use authenticated user's MSISDN
	req.MSISDN = msisdn

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Cancel subscription
	err := h.subscriptionService.CancelSubscription(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log subscription cancellation
	errors.Info("Subscription cancelled", map[string]interface{}{
		"msisdn": msisdn,
	})

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message": "Subscription cancelled successfully",
	})
}

// Subscribe godoc
// @Summary Subscribe to daily service
// @Description Subscribe user to daily recharge service (alias for CreateSubscription)
// @Tags subscription
// @Accept json
// @Produce json
// @Param request body validation.SubscribeRequest true "Subscription request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /subscription/subscribe [post]
func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	// Alias for CreateSubscription to match frontend expectations
	h.CreateSubscription(c)
}

// Unsubscribe godoc
// @Summary Unsubscribe from daily service
// @Description Unsubscribe user from daily recharge service (alias for CancelSubscription)
// @Tags subscription
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /subscription/unsubscribe [post]
func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	// Alias for CancelSubscription to match frontend expectations
	h.CancelSubscription(c)
}

// GetHistory godoc
// @Summary Get subscription history
// @Description Get user's subscription history with pagination
// @Tags subscription
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /subscription/history [get]
func (h *SubscriptionHandler) GetHistory(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	// Parse pagination parameters
	var pagination validation.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid pagination parameters"))
		return
	}

	// Validate pagination
	if err := pagination.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Get subscription history
	history, err := h.subscriptionService.GetSubscriptionHistory(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"subscriptions": history,
		"page":          pagination.Page,
		"limit":         pagination.Limit,
	})
}
