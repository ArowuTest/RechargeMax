package handlers

import (
	"io"
	"net/http"

	"log"

	"rechargemax/internal/application/services"
	"rechargemax/internal/middleware"
	"rechargemax/internal/errors"

	"github.com/gin-gonic/gin"
)

// WebhookHandler handles webhook requests
type WebhookHandler struct {
	webhookService *services.WebhookService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(webhookService *services.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// HandlePaystackWebhook godoc
// @Summary Handle Paystack webhook
// @Description Receive and process webhook events from Paystack
// @Tags webhooks
// @Accept json
// @Produce json
// @Param x-paystack-signature header string true "Paystack signature"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /webhooks/paystack [post]
func (h *WebhookHandler) HandlePaystackWebhook(c *gin.Context) {
	// Get signature from header
	signature := c.GetHeader("x-paystack-signature")
	if signature == "" {
		log.Println("[Webhook] WARN: Webhook received without signature")
		middleware.RespondWithError(c, errors.Unauthorized("Missing webhook signature"))
		return
	}

	// Read raw body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[Webhook] ERROR: Failed to read webhook body: %v", err)
		middleware.RespondWithError(c, errors.BadRequest("Failed to read request body"))
		return
	}

	// Log webhook receipt
	log.Printf("[Webhook] Received from %s - Signature present: %v, Body size: %d",
		c.ClientIP(), signature != "", len(body))

	// Process webhook
	if err := h.webhookService.ProcessPaystackWebhook(c.Request.Context(), body, signature); err != nil {
	log.Printf("[Webhook] ERROR: Failed to process webhook: %v", err)
		middleware.RespondWithError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Webhook processed successfully",
	})
}
