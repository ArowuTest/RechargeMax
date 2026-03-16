package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	
	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
	"rechargemax/internal/validation"
)

type PaymentHandler struct {
	paymentService      *services.PaymentService
	rechargeService     *services.RechargeService
	subscriptionService *services.SubscriptionService
	frontendURL         string
}

func NewPaymentHandler(paymentService *services.PaymentService, rechargeService *services.RechargeService, subscriptionService *services.SubscriptionService, frontendURL string) *PaymentHandler {
	return &PaymentHandler{
		paymentService:      paymentService,
		rechargeService:     rechargeService,
		subscriptionService: subscriptionService,
		frontendURL:         frontendURL,
	}
}

// InitializePayment godoc
// @Summary Initialize a payment
// @Description Initialize payment with Paystack or Flutterwave
// @Tags payment
// @Accept json
// @Produce json
// @Param request body validation.InitiatePaymentRequest true "Payment request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /payment/initialize [post]
func (h *PaymentHandler) InitializePayment(c *gin.Context) {
	var req validation.InitiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Initialize payment
	authURL, err := h.paymentService.InitializePayment(c.Request.Context(), services.PaymentRequest{
		Amount:      int64(req.Amount * 100), // Convert naira to kobo
		Email:       req.Email,
		CallbackURL: req.CallbackURL,
		Metadata:    req.Metadata,
	})

	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log payment initialization
	errors.Info("Payment initialized", map[string]interface{}{
		"email":             req.Email,
		"amount":            req.Amount,
		"authorization_url": authURL,
		"metadata":          req.Metadata,
	})

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"authorization_url": authURL,
	})
}

// VerifyPayment godoc
// @Summary Verify a payment
// @Description Verify payment status using reference
// @Tags payment
// @Accept json
// @Produce json
// @Param reference path string true "Payment reference"
// @Param gateway query string false "Payment gateway" default(paystack)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /payment/verify/{reference} [get]
func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
	reference := c.Param("reference")

	if reference == "" {
		middleware.RespondWithError(c, errors.BadRequest("Payment reference is required"))
		return
	}

	// Get gateway from query parameter (default to paystack)
	gateway := c.DefaultQuery("gateway", "paystack")

	// Verify payment
	success, amount, err := h.paymentService.VerifyPayment(c.Request.Context(), reference, gateway)

	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log payment verification
	errors.Info("Payment verified", map[string]interface{}{
		"reference": reference,
		"success":   success,
		"gateway":   gateway,
	})
	
	middleware.RespondWithSuccess(c, map[string]interface{}{
		"success":   success,
		"reference": reference,
		"amount":    amount,
	})
}

// HandleWebhook godoc
// @Summary Handle payment webhook
// @Description Receive and process payment webhooks from Paystack/Flutterwave
// @Tags payment
// @Accept json
// @Produce json
// @Param gateway query string false "Payment gateway" default(paystack)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Router /payment/webhook [post]
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
	// Get gateway from query parameter (default to paystack)
	gateway := c.DefaultQuery("gateway", "paystack")

	// Get webhook signature from header
	var signature string
	switch gateway {
	case "paystack":
		signature = c.GetHeader("x-paystack-signature")
	case "flutterwave":
		signature = c.GetHeader("verif-hash")
	default:
		middleware.RespondWithError(c, errors.BadRequest("Invalid payment gateway"))
		return
	}

	if signature == "" {
		middleware.RespondWithError(c, errors.Unauthorized("Missing webhook signature"))
		return
	}

	// Read raw body for signature verification
	rawBody, err := c.GetRawData()
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Failed to read request body"))
		return
	}

	// Process webhook
	reference, status, err := h.paymentService.ProcessWebhook(rawBody, signature, gateway)
	if err != nil {
		// Log error but return 200 to prevent webhook retries for invalid signatures
		errors.Error("Payment webhook processing failed", err, map[string]interface{}{
			"gateway": gateway,
		})
		
		middleware.RespondWithSuccess(c, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// If webhook doesn't contain relevant event, return success
	if reference == "" || status == "" {
		middleware.RespondWithSuccess(c, map[string]interface{}{
			"message": "Webhook received but no action required",
		})
		return
	}

	// Check idempotency - prevent duplicate processing
	// This is critical for webhooks which may be retried by payment providers
	if h.paymentService.IsPaymentProcessed(c.Request.Context(), reference) {
		errors.Info("Payment already processed (idempotency check)", map[string]interface{}{
			"reference": reference,
			"gateway":   gateway,
		})
		middleware.RespondWithSuccess(c, map[string]interface{}{
			"message": "Payment already processed",
		})
		return
	}

	// Process payment based on status
	if status == "completed" {
		// Determine transaction type from reference prefix
		// RCH_ = Recharge, SUB_ = Subscription
		if len(reference) >= 4 {
			prefix := reference[:4]

			switch prefix {
			case "RCH_":
				// Process recharge
				if err := h.rechargeService.ProcessSuccessfulPayment(c.Request.Context(), reference); err != nil {
					// Log error but return 200 to acknowledge webhook
					errors.Error("Failed to process recharge payment", err, map[string]interface{}{
						"reference": reference,
					})
					
					middleware.RespondWithSuccess(c, map[string]interface{}{
						"success": false,
						"error":   err.Error(),
						"message": "Failed to process recharge",
					})
					return
				}

				// Log successful recharge processing
				errors.Info("Recharge payment processed", map[string]interface{}{
					"reference": reference,
					"status":    status,
				})

			case "SUB_":
				// Process subscription
				if err := h.subscriptionService.ProcessSuccessfulPayment(c.Request.Context(), reference); err != nil {
					// Log error but return 200 to acknowledge webhook
					errors.Error("Failed to process subscription payment", err, map[string]interface{}{
						"reference": reference,
					})
					
					middleware.RespondWithSuccess(c, map[string]interface{}{
						"success": false,
						"error":   err.Error(),
						"message": "Failed to process subscription",
					})
					return
				}

				// Log successful subscription processing
				errors.Info("Subscription payment processed", map[string]interface{}{
					"reference": reference,
					"status":    status,
				})

			default:
				// Unknown transaction type
				errors.Warning("Unknown payment reference prefix", map[string]interface{}{
					"reference": reference,
					"prefix":    prefix,
				})
				
				middleware.RespondWithSuccess(c, map[string]interface{}{
					"success": false,
					"error":   "Unknown transaction type",
					"message": "Payment reference prefix not recognized",
				})
				return
			}
		}

		middleware.RespondWithSuccess(c, map[string]interface{}{
			"message":   "Payment webhook processed successfully",
			"reference": reference,
			"status":    status,
		})
		return
	}

	// For failed payments, just acknowledge
	errors.Info("Payment webhook received", map[string]interface{}{
		"reference": reference,
		"status":    status,
	})

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message":   "Payment webhook processed",
		"reference": reference,
		"status":    status,
	})
}

// HandleCallback godoc
// @Summary Handle payment callback
// @Description Handle user redirect after payment completion
// @Tags payment
// @Produce json,html
// @Param reference query string true "Payment reference"
// @Param gateway query string false "Payment gateway" default(paystack)
// @Success 200 {string} string "Redirect or JSON"
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /payment/callback [get]
func (h *PaymentHandler) HandleCallback(c *gin.Context) {
	reference := c.Query("reference")

	if reference == "" {
		middleware.RespondWithError(c, errors.BadRequest("Payment reference is required"))
		return
	}

	// Get gateway from query parameter (default to paystack)
	gateway := c.DefaultQuery("gateway", "paystack")

	// Verify payment
	success, _, err := h.paymentService.VerifyPayment(c.Request.Context(), reference, gateway)

	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	if success {
		if len(reference) >= 4 {
			prefix := reference[:4]
			switch prefix {
			case "RCH_":
				// Fire-and-forget: VTPass can take 5-30s (especially in sandbox).
				// Redirect the browser immediately so the user is not left on a blank page.
				// The frontend polls /recharge/reference/:ref every 3s until status=SUCCESS.
				// Idempotency is enforced inside ProcessSuccessfulPayment (no-op if not PENDING).
				go func() {
					if err := h.rechargeService.ProcessSuccessfulPayment(context.Background(), reference); err != nil {
						errors.Error("Async VTPass processing failed in callback", err, map[string]interface{}{
							"reference": reference,
						})
					}
				}()
			case "SUB_":
				go func() {
					if err := h.subscriptionService.ProcessSuccessfulPayment(context.Background(), reference); err != nil {
						errors.Error("Async subscription processing failed in callback", err, map[string]interface{}{
							"reference": reference,
						})
					}
				}()
			}
		}

		// Check if this is an AJAX call or a browser redirect from Paystack
		isAPICall := strings.Contains(c.GetHeader("Accept"), "application/json")
		if isAPICall {
			c.JSON(http.StatusOK, gin.H{
				"success":   true,
				"message":   "Payment verified, recharge processing",
				"reference": reference,
			})
		} else {
			// Browser redirect: go to frontend immediately, let it poll for VTPass result
			c.Redirect(http.StatusFound, h.frontendURL+"/?payment=success&reference="+reference)
		}
	} else {
		isAPICall := strings.Contains(c.GetHeader("Accept"), "application/json")
		if isAPICall {
			c.JSON(http.StatusOK, gin.H{
				"success":   false,
				"message":   "Payment verification failed",
				"reference": reference,
			})
		} else {
			c.Redirect(http.StatusFound, h.frontendURL+"/?payment=failed&reference="+reference)
		}
	}
}

// HandleSuccess godoc
// @Summary Payment success page
// @Description Show payment success page with transaction details
// @Tags payment
// @Produce json
// @Param reference query string true "Payment reference"
// @Param redirect_url query string false "Frontend redirect URL"
// @Success 200 {object} map[string]interface{}
// @Router /payment/callback/success [get]
func (h *PaymentHandler) HandleSuccess(c *gin.Context) {
	reference := c.Query("reference")

	// Log successful payment callback
	errors.Info("Payment success callback", map[string]interface{}{
		"reference": reference,
	})

	// For web clients with redirect URL, redirect to frontend
	frontendURL := c.Query("redirect_url")
	if frontendURL != "" {
		c.Redirect(http.StatusFound, frontendURL+"?status=success&reference="+reference)
		return
	}

	// Return JSON
	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message":   "Payment completed successfully",
		"reference": reference,
		"status":    "completed",
	})
}

// HandleCancel godoc
// @Summary Payment cancellation page
// @Description Show payment cancellation page
// @Tags payment
// @Produce json
// @Param reference query string false "Payment reference"
// @Param redirect_url query string false "Frontend redirect URL"
// @Success 200 {object} map[string]interface{}
// @Router /payment/callback/cancel [get]
func (h *PaymentHandler) HandleCancel(c *gin.Context) {
	reference := c.Query("reference")

	// Log cancelled payment callback
	errors.Info("Payment cancel callback", map[string]interface{}{
		"reference": reference,
	})

	// For web clients with redirect URL, redirect to frontend
	frontendURL := c.Query("redirect_url")
	if frontendURL != "" {
		c.Redirect(http.StatusFound, frontendURL+"?status=cancelled&reference="+reference)
		return
	}

	// Return JSON
	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message":   "Payment was cancelled",
		"reference": reference,
		"status":    "cancelled",
	})
}
