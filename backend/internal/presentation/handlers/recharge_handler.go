package handlers

import (
	"encoding/json"
	"fmt"

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
	fmt.Printf("[DEBUG] InitiateAirtimeRecharge called, msisdn from context: %s\n", msisdn)

	var req validation.AirtimeRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[ERROR] Failed to bind JSON: %v\n", err)
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}
	fmt.Printf("[DEBUG] Request parsed: phone=%s, network=%s, amount=%.2f\n", req.PhoneNumber, req.Network, req.Amount)

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
	}

	fmt.Printf("[DEBUG] Calling CreateRecharge service...\n")
	result, err := h.rechargeService.CreateRecharge(c.Request.Context(), rechargeReq)
	if err != nil {
		fmt.Printf("[ERROR] CreateRecharge failed: %v\n", err)
		middleware.RespondWithError(c, err)
		return
	}
	fmt.Printf("[DEBUG] CreateRecharge succeeded, result: %+v\n", result)

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
	fmt.Printf("[DEBUG] GetRechargeByReference called with reference: %s\n", reference)

	if reference == "" {
		fmt.Printf("[ERROR] Reference is empty\n")
		middleware.RespondWithError(c, errors.BadRequest("Payment reference is required"))
		return
	}

	// Get recharge details by reference
	fmt.Printf("[DEBUG] Calling rechargeService.GetRechargeByReference...\n")
	recharge, err := h.rechargeService.GetRechargeByReference(c.Request.Context(), reference)
	if err != nil {
		fmt.Printf("[ERROR] GetRechargeByReference failed: %v\n", err)
		middleware.RespondWithError(c, errors.NotFound("Recharge not found"))
		return
	}
	fmt.Printf("[DEBUG] Recharge found: %+v\n", recharge)

	middleware.RespondWithSuccess(c, recharge)
}
