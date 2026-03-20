package handlers

import (
	"github.com/gin-gonic/gin"

	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
)

type SubscriptionHandler struct {
	subscriptionService *services.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService *services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionService: subscriptionService}
}

// createRequest is the unified body for subscription creation.
// All fields are optional at binding time — service validates what's needed.
type createSubscriptionBody struct {
	MSISDN        string `json:"msisdn"`
	PhoneNumber   string `json:"phone_number"`  // alternate field name frontend may send
	Network       string `json:"network"`
	PaymentMethod string `json:"payment_method"`
}

// resolveMSISDN returns the MSISDN to use for this request.
// Priority: JWT token (authenticated user) > body.msisdn > body.phone_number
func resolveMSISDN(c *gin.Context, body *createSubscriptionBody) string {
	if msisdn := c.GetString("msisdn"); msisdn != "" {
		return msisdn
	}
	if body.MSISDN != "" {
		return body.MSISDN
	}
	return body.PhoneNumber
}

// CreateSubscription godoc
// @Summary Subscribe to the daily ₦20 draw
// @Description Creates a daily subscription. Works for both authenticated users
//              (MSISDN taken from JWT) and guests (MSISDN provided in body).
// @Tags subscription
// @Accept  json
// @Produce json
// @Param   request body createSubscriptionBody false "Subscription request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 409 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /subscription/create [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var body createSubscriptionBody
	// ShouldBindJSON is best-effort; we don't fail on a missing body
	_ = c.ShouldBindJSON(&body)

	msisdn := resolveMSISDN(c, &body)
	if msisdn == "" {
		middleware.RespondWithError(c, errors.BadRequest("Phone number is required"))
		return
	}

	result, err := h.subscriptionService.CreateSubscription(c.Request.Context(), services.CreateSubscriptionRequest{
		MSISDN:        msisdn,
		Network:       body.Network,
		PaymentMethod: body.PaymentMethod,
	})
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}
	middleware.RespondWithSuccess(c, result)
}

// GetSubscription godoc
// @Summary Get active subscription status
// @Tags subscription
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} errors.ErrorResponse
// @Router /subscription/status [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		// Allow guests to query by MSISDN query param
		msisdn = c.Query("msisdn")
	}
	if msisdn == "" {
		middleware.RespondWithError(c, errors.BadRequest("Phone number is required"))
		return
	}

	subscription, err := h.subscriptionService.GetSubscription(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}
	middleware.RespondWithSuccess(c, subscription)
}

// CancelSubscription godoc
// @Summary Cancel active subscription
// @Tags subscription
// @Accept  json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} errors.ErrorResponse
// @Router /subscription/cancel [post]
func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		var body struct {
			MSISDN string `json:"msisdn"`
		}
		_ = c.ShouldBindJSON(&body)
		msisdn = body.MSISDN
	}
	if msisdn == "" {
		middleware.RespondWithError(c, errors.BadRequest("Phone number is required"))
		return
	}

	if err := h.subscriptionService.CancelSubscription(c.Request.Context(), msisdn); err != nil {
		middleware.RespondWithError(c, err)
		return
	}
	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message": "Subscription cancelled successfully",
	})
}

// GetConfig godoc
// @Summary Get subscription pricing configuration
// @Tags subscription
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /subscription/config [get]
func (h *SubscriptionHandler) GetConfig(c *gin.Context) {
	config, err := h.subscriptionService.GetConfig(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, errors.Internal("Failed to fetch subscription config"))
		return
	}
	middleware.RespondWithSuccess(c, config)
}

// Subscribe is an alias for CreateSubscription (matches frontend /subscribe endpoint)
func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	h.CreateSubscription(c)
}

// Unsubscribe is an alias for CancelSubscription
func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	h.CancelSubscription(c)
}

// GetHistory godoc
// @Summary Get subscription history
// @Tags subscription
// @Produce json
// @Param page  query int false "Page"  default(1)
// @Param limit query int false "Limit" default(50)
// @Success 200 {object} map[string]interface{}
// @Router /subscription/history [get]
func (h *SubscriptionHandler) GetHistory(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		msisdn = c.Query("msisdn")
	}
	if msisdn == "" {
		middleware.RespondWithError(c, errors.BadRequest("Phone number is required"))
		return
	}

	history, err := h.subscriptionService.GetSubscriptionHistory(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}
	middleware.RespondWithSuccess(c, map[string]interface{}{
		"subscriptions": history,
		"total":         len(history),
	})
}
