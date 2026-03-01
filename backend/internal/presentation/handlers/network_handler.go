package handlers

import (
	"github.com/gin-gonic/gin"
	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
	"rechargemax/internal/validation"
)

type NetworkHandler struct {
	networkService *services.NetworkConfigService
	hlrService     *services.HLRService
}

func NewNetworkHandler(networkService *services.NetworkConfigService, hlrService *services.HLRService) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
		hlrService:     hlrService,
	}
}

// GetNetworks godoc
// @Summary Get all available networks
// @Description Get list of all telecom networks (MTN, Glo, Airtel, 9mobile)
// @Tags networks
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} errors.ErrorResponse
// @Router /networks [get]
func (h *NetworkHandler) GetNetworks(c *gin.Context) {
	networks, err := h.networkService.GetAllNetworks(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, networks)
}

// GetDataBundles godoc
// @Summary Get data bundles for a network
// @Description Get available data bundles for a specific network
// @Tags networks
// @Produce json
// @Param networkId path string true "Network ID or Code (e.g., MTN, GLO, AIRTEL, 9MOBILE)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /networks/{networkId}/bundles [get]
func (h *NetworkHandler) GetDataBundles(c *gin.Context) {
	networkID := c.Param("networkId")

	if networkID == "" {
		middleware.RespondWithError(c, errors.BadRequest("Network ID is required"))
		return
	}

	// Validate network code
	if err := validation.ValidateNetwork(networkID); err != nil {
		middleware.RespondWithError(c, errors.BadRequest(err.Error()))
		return
	}

	bundles, err := h.networkService.GetDataBundles(c.Request.Context(), networkID)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, bundles)
}

// ValidatePhoneNetwork godoc
// @Summary Validate phone number network
// @Description Validate that a phone number belongs to the specified network using HLR lookup
// @Tags networks
// @Accept json
// @Produce json
// @Param request body validation.ValidatePhoneNetworkRequest true "Validate Phone Network Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /networks/validate [post]
func (h *NetworkHandler) ValidatePhoneNetwork(c *gin.Context) {
	var req validation.ValidatePhoneNetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Validate phone number network
	result, err := h.networkService.ValidatePhoneNetwork(c.Request.Context(), req.PhoneNumber, req.ExpectedNetwork)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, result)
}

// GetCachedNetworkRequest represents request to get cached network
type GetCachedNetworkRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

// ValidateNetworkSelectionRequest represents request to validate network selection
type ValidateNetworkSelectionRequest struct {
	PhoneNumber     string `json:"phone_number" binding:"required"`
	SelectedNetwork string `json:"selected_network" binding:"required"`
}

// GetCachedNetwork godoc
// @Summary Get cached network for phone number
// @Description Get network from recent successful recharges (last 30 days) to auto-suggest network
// @Tags networks
// @Accept json
// @Produce json
// @Param request body GetCachedNetworkRequest true "Get Cached Network Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /networks/cached [post]
func (h *NetworkHandler) GetCachedNetwork(c *gin.Context) {
	var req GetCachedNetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}
	// Normalise MSISDN to canonical international format (234...) in-place
	if normalized, err := validation.NormalizeMSISDN(req.PhoneNumber); err == nil {
		req.PhoneNumber = normalized
	}

	// Get cached network from recent recharges
	result, err := h.hlrService.GetCachedNetworkForUser(c.Request.Context(), req.PhoneNumber)
	if err != nil {
		// No cache found - this is not an error, just means first time user
		middleware.RespondWithSuccess(c, gin.H{
			"cached": false,
			"message": "No recent recharge history found for this number",
		})
		return
	}

	middleware.RespondWithSuccess(c, gin.H{
		"cached":         true,
		"network":        result.ActualNetwork,
		"last_recharged": result.LastRecharged,
		"message":        result.Message,
		"confidence":     result.Confidence,
	})
}

// ValidateNetworkSelection godoc
// @Summary Validate network selection before payment
// @Description Validate that selected network matches the actual network via HLR API or prefix
// @Tags networks
// @Accept json
// @Produce json
// @Param request body ValidateNetworkSelectionRequest true "Validate Network Selection Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 422 {object} errors.ErrorResponse
// @Router /networks/validate-selection [post]
func (h *NetworkHandler) ValidateNetworkSelection(c *gin.Context) {
	var req ValidateNetworkSelectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}
	// Normalise MSISDN to canonical international format (234...) in-place
	if normalized, err := validation.NormalizeMSISDN(req.PhoneNumber); err == nil {
		req.PhoneNumber = normalized
	}

	// Validate network selection
	result, err := h.hlrService.ValidateNetworkSelection(c.Request.Context(), req.PhoneNumber, req.SelectedNetwork)
	if err != nil {
		middleware.RespondWithError(c, errors.Internal("Network validation failed: "+err.Error()))
		return
	}

	// If validation failed (network mismatch), return 422 with details
	if !result.IsValid {
		c.JSON(422, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NETWORK_MISMATCH",
				"message": result.Message,
			},
			"data": gin.H{
				"selected_network":   result.SelectedNetwork,
				"actual_network":     result.ActualNetwork,
				"validation_source":  result.ValidationSource,
				"confidence":         result.Confidence,
			},
		})
		return
	}

	// Validation passed
	middleware.RespondWithSuccess(c, gin.H{
		"valid":              true,
		"network":            result.ActualNetwork,
		"validation_source":  result.ValidationSource,
		"confidence":         result.Confidence,
		"message":            result.Message,
	})
}
