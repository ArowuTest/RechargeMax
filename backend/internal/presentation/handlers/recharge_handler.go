package handlers

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
	"rechargemax/internal/validation"
)

type RechargeHandler struct {
	rechargeService *services.RechargeService
}

func NewRechargeHandler(rechargeService *services.RechargeService) *RechargeHandler {
	return &RechargeHandler{
		rechargeService: rechargeService,
	}
}

// InitiateRecharge godoc
// @Summary Initiate recharge
// @Description Initiate a recharge (airtime, data, or subscription)
// @Tags recharge
// @Accept json
// @Produce json
// @Param request body validation.RechargeRequest true "Recharge Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /recharge [post]
func (h *RechargeHandler) InitiateRecharge(c *gin.Context) {
	var req validation.RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Convert amount to kobo
	amountKobo := int64(req.Amount * 100)
	
	// Create recharge
	var networkPtr *string
	if req.Network != "" {
		networkPtr = &req.Network
	}
	recharge, err := h.rechargeService.CreateRecharge(c.Request.Context(), services.CreateRechargeRequest{
		MSISDN:        req.MSISDN,
		Amount:        amountKobo,
		Network:       networkPtr,
		RechargeType:  req.Type,
		DataPackage:   "", // DataPackage not in validation request
		PaymentMethod: "CARD", // Paystack gateway, CARD payment method
	})

	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log transaction
	errors.LogTransaction("RECHARGE", recharge.ID.String(), req.MSISDN, float64(amountKobo)/100, "INITIATED")

	middleware.RespondWithSuccess(c, recharge)
}

// InitiateAirtimeRecharge godoc
// @Summary Initiate airtime recharge
// @Description Initiate an airtime recharge for a phone number
// @Tags recharge
// @Accept json
// @Produce json
// @Param request body validation.AirtimeRechargeRequest true "Airtime recharge request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /recharge/airtime [post]
func (h *RechargeHandler) InitiateAirtimeRecharge(c *gin.Context) {
	// Get msisdn from auth context (if authenticated) or use phone number from request
	msisdn := c.GetString("msisdn")

	var req validation.AirtimeRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Convert amount to kobo
	amountKobo := int64(req.Amount * 100)

	// Convert network to pointer
	networkPtr := &req.Network

	// Create recharge request
	rechargeReq := services.CreateRechargeRequest{
		MSISDN:        req.PhoneNumber,
		Network:       networkPtr,
		Amount:        amountKobo,
		RechargeType:  "AIRTIME",
		PaymentMethod: "CARD", // Paystack gateway, CARD payment method
		AffiliateCode: req.AffiliateCode,
	}

	result, err := h.rechargeService.CreateRecharge(c.Request.Context(), rechargeReq)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log transaction
	errors.LogTransaction("AIRTIME_RECHARGE", result.ID.String(), msisdn, float64(amountKobo)/100, "INITIATED")

	middleware.RespondWithSuccess(c, result)
}

// InitiateDataRecharge godoc
// @Summary Initiate data recharge
// @Description Initiate a data bundle recharge for a phone number
// @Tags recharge
// @Accept json
// @Produce json
// @Param request body validation.DataRechargeRequest true "Data recharge request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /recharge/data [post]
func (h *RechargeHandler) InitiateDataRecharge(c *gin.Context) {
	// Get msisdn from auth context (if authenticated) or use phone number from request
	msisdn := c.GetString("msisdn")

	var req validation.DataRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Convert amount to kobo
	amountKobo := int64(req.Amount * 100)

	// Convert network to pointer
	networkPtr := &req.Network

	// Create recharge request
	rechargeReq := services.CreateRechargeRequest{
		MSISDN:        req.PhoneNumber,
		Network:       networkPtr,
		Amount:        amountKobo,
		RechargeType:  "DATA",
		DataPackage:   req.BundleID,
		PaymentMethod: "CARD", // Paystack gateway, CARD payment method
		AffiliateCode: req.AffiliateCode,
	}

	result, err := h.rechargeService.CreateRecharge(c.Request.Context(), rechargeReq)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log transaction
	errors.LogTransaction("DATA_RECHARGE", result.ID.String(), msisdn, float64(amountKobo)/100, "INITIATED")

	middleware.RespondWithSuccess(c, result)
}

// GetRechargeHistory godoc
// @Summary Get recharge history
// @Description Get user's recharge history with pagination
// @Tags recharge
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /recharge/history [get]
func (h *RechargeHandler) GetRechargeHistory(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	// Get pagination parameters
	var req validation.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = validation.PaginationRequest{Page: 1, Limit: 50}
	}

	// Validate pagination
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Calculate offset
	offset := (req.Page - 1) * req.Limit

	history, err := h.rechargeService.GetRechargeHistory(c.Request.Context(), msisdn, req.Limit, offset)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"data":  history,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// GetHistory is an alias for GetRechargeHistory for backward compatibility
func (h *RechargeHandler) GetHistory(c *gin.Context) {
	h.GetRechargeHistory(c)
}

// GetRecharge godoc
// @Summary Get single recharge
// @Description Get details of a specific recharge by ID
// @Tags recharge
// @Produce json
// @Param id path string true "Recharge ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /recharge/{id} [get]
func (h *RechargeHandler) GetRecharge(c *gin.Context) {
	rechargeID := c.Param("id")

	if rechargeID == "" {
		middleware.RespondWithError(c, errors.BadRequest("Recharge ID is required"))
		return
	}

	// Parse UUID
	rid, err := uuid.Parse(rechargeID)
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid recharge ID format"))
		return
	}

	// Get recharge details
	recharge, err := h.rechargeService.GetRechargeByID(c.Request.Context(), rid)
	if err != nil {
		middleware.RespondWithError(c, errors.NotFound("Recharge not found"))
		return
	}

	// SECURITY: If the caller is authenticated, ensure they only see their own transaction.
	// (Route uses OptionalAuth so unauthenticated callers — e.g. the Paystack callback flow — are allowed.)
	if callerMsisdn := c.GetString("msisdn"); callerMsisdn != "" {
		if recharge.MSISDN != "" && recharge.MSISDN != callerMsisdn {
			middleware.RespondWithError(c, errors.NotFound("Recharge not found"))
			return
		}
	}

	middleware.RespondWithSuccess(c, recharge)
}

// HandleTelecomWebhook godoc
// @Summary Handle telecom provider webhook
// @Description Receive async recharge confirmation from telecom provider
// @Tags webhook, recharge
// @Accept json
// @Produce json
// @Param provider query string false "Provider name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Router /webhooks/telecom [post]
func (h *RechargeHandler) HandleTelecomWebhook(c *gin.Context) {
	// Get provider from query parameter
	provider := c.DefaultQuery("provider", "unknown")

	// Read raw body
	rawBody, err := c.GetRawData()
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Failed to read request body"))
		return
	}

	// Parse webhook payload
	var payload map[string]interface{}
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid JSON payload"))
		return
	}

	// Extract reference and status
	reference, _ := payload["reference"].(string)
	status, _ := payload["status"].(string)

	if reference == "" {
		middleware.RespondWithError(c, errors.BadRequest("Missing reference in webhook payload"))
		return
	}

	// Process telecom confirmation
	err = h.rechargeService.ProcessTelecomConfirmation(c.Request.Context(), reference, status, provider, payload)
	if err != nil {
		// Log error but return 200 to prevent retries
		errors.Error("Telecom webhook processing failed", err, map[string]interface{}{
			"provider":  provider,
			"reference": reference,
			"status":    status,
		})
		
		middleware.RespondWithSuccess(c, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Log successful webhook processing
	errors.Info("Telecom webhook processed", map[string]interface{}{
		"provider":  provider,
		"reference": reference,
		"status":    status,
	})

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message":   "Telecom webhook processed successfully",
		"reference": reference,
		"status":    status,
		"provider":  provider,
	})
}

// ProcessStuckRecharge resets a stuck PROCESSING transaction to PENDING and retries it synchronously.
// This is a debug/recovery endpoint — returns the exact error if processing fails.
func (h *RechargeHandler) ProcessStuckRecharge(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		c.JSON(400, gin.H{"success": false, "error": "reference required"})
		return
	}

	ctx := c.Request.Context()

	// Get the transaction
	recharge, err := h.rechargeService.GetRechargeByReference(ctx, reference)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "transaction not found: " + err.Error()})
		return
	}

	// Only process PROCESSING or PENDING transactions
	if recharge.Status != "PROCESSING" && recharge.Status != "PENDING" {
		c.JSON(200, gin.H{"success": true, "message": "transaction already in status: " + recharge.Status, "status": recharge.Status})
		return
	}

	// Reset to PENDING if stuck in PROCESSING
	if recharge.Status == "PROCESSING" {
		if err := h.rechargeService.ResetToPending(ctx, recharge.ID); err != nil {
			c.JSON(500, gin.H{"success": false, "error": "failed to reset to pending: " + err.Error()})
			return
		}
	}

	// Process synchronously (not in goroutine) so we get the exact error
	if err := h.rechargeService.ProcessSuccessfulPayment(ctx, reference); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "processing failed: " + err.Error(), "reference": reference})
		return
	}

	// Check final status
	final, _ := h.rechargeService.GetRechargeByReference(ctx, reference)
	finalStatus := "UNKNOWN"
	if final != nil {
		finalStatus = final.Status
	}

	c.JSON(200, gin.H{"success": true, "message": "processed successfully", "reference": reference, "status": finalStatus})
}

// GetRechargeByReference godoc
// @Summary Get recharge by payment reference
// @Description Get recharge transaction details by payment reference
// @Tags recharge
// @Accept json
// @Produce json
// @Param reference path string true "Payment Reference"
// @Success 200 {object} entities.Transactions
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /api/v1/recharge/reference/{reference} [get]
func (h *RechargeHandler) GetRechargeByReference(c *gin.Context) {
	reference := c.Param("reference")

	if reference == "" {
		middleware.RespondWithError(c, errors.BadRequest("Payment reference is required"))
		return
	}

	recharge, err := h.rechargeService.GetRechargeByReference(c.Request.Context(), reference)
	if err != nil {
		middleware.RespondWithError(c, errors.NotFound("Recharge not found"))
		return
	}

	// SECURITY: If authenticated, verify the caller owns this transaction.
	// Unauthenticated callers (Paystack redirect callback) are still allowed through.
	if callerMsisdn := c.GetString("msisdn"); callerMsisdn != "" {
		if recharge.MSISDN != "" && recharge.MSISDN != callerMsisdn {
			middleware.RespondWithError(c, errors.NotFound("Recharge not found"))
			return
		}
	}

	middleware.RespondWithSuccess(c, recharge)
}
