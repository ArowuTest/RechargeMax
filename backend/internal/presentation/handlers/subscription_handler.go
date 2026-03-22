package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

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
	Entries       int    `json:"entries"`        // number of daily draw entries (1–100)
	Amount        int64  `json:"amount"`         // total amount in kobo (informational — backend recalculates)
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
		Entries:       body.Entries, // pass through — service defaults to 1 if 0
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

// GetActiveLines returns ALL active subscription lines for a MSISDN,
// plus aggregate totals (total_daily_entries, total_daily_cost).
// This supports the multi-line subscription UI.
func (h *SubscriptionHandler) GetActiveLines(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		msisdn = c.Query("msisdn")
	}
	if msisdn == "" {
		middleware.RespondWithError(c, errors.BadRequest("Phone number is required"))
		return
	}

	subs, err := h.subscriptionService.GetAllActiveSubscriptions(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Build response list and compute aggregates
	type lineItem struct {
		ID            string  `json:"id"`
		Code          string  `json:"code"`
		Entries       int     `json:"entries"`
		DailyAmountNGN float64 `json:"daily_amount_ngn"`
		Status        string  `json:"status"`
		NextBilling   string  `json:"next_billing"`
		CreatedAt     string  `json:"created_at"`
	}
	lines := make([]lineItem, 0, len(subs))
	totalEntries := 0
	totalDailyCostNGN := 0.0
	for _, s := range subs {
		entries := s.BundleQuantity
		if entries == 0 {
			entries = 1
		}
		amtNGN := float64(s.DailyAmount) / 100
		if amtNGN == 0 {
			amtNGN = s.Amount // legacy naira field
		}
		totalEntries += entries
		totalDailyCostNGN += amtNGN
		lines = append(lines, lineItem{
			ID:            s.ID.String(),
			Code:          s.SubscriptionCode,
			Entries:       entries,
			DailyAmountNGN: amtNGN,
			Status:        s.Status,
			NextBilling:   s.NextBillingDate.Format("2006-01-02T15:04:05Z"),
			CreatedAt:     s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"lines":                lines,
		"total_active_lines":   len(lines),
		"total_daily_entries":  totalEntries,
		"total_daily_cost_ngn": totalDailyCostNGN,
	})
}

// CancelLine cancels a specific subscription line by its ID.
func (h *SubscriptionHandler) CancelLine(c *gin.Context) {
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

	lineID := c.Param("id")
	parsedID, err := uuid.Parse(lineID)
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid subscription line ID"))
		return
	}

	if err := h.subscriptionService.CancelSubscriptionByID(c.Request.Context(), msisdn, parsedID); err != nil {
		middleware.RespondWithError(c, err)
		return
	}
	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message": "Subscription line cancelled successfully",
	})
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
