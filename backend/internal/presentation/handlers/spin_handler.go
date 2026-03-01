package handlers

import (
	"github.com/gin-gonic/gin"
	
	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
	"rechargemax/internal/validation"
)

type SpinHandler struct {
	spinService *services.SpinService
}

func NewSpinHandler(spinService *services.SpinService) *SpinHandler {
	return &SpinHandler{spinService: spinService}
}

// CheckEligibility godoc
// @Summary Check spin eligibility
// @Description Check if user is eligible to spin the wheel
// @Tags spin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /spin/eligibility [get]
func (h *SpinHandler) CheckEligibility(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	// Service will validate spin eligibility
	eligibility, err := h.spinService.CheckEligibility(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, eligibility)
}

// PlaySpin godoc
// @Summary Play spin
// @Description Spin the wheel and get a prize
// @Tags spin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /spin/play [post]
func (h *SpinHandler) PlaySpin(c *gin.Context) {
	// DEBUG: Log incoming request
	errors.Info("PlaySpin called", map[string]interface{}{
		"method": c.Request.Method,
		"path": c.Request.URL.Path,
	})
	
	// Support both authenticated users (JWT) and guest users (request body)
	// OptionalAuthMiddleware sets msisdn in context if JWT is present
	msisdn := c.GetString("msisdn")
	errors.Info("MSISDN from context", map[string]interface{}{
		"msisdn": msisdn,
		"has_auth_header": c.GetHeader("Authorization") != "",
	})
	
	// If no MSISDN from JWT context, try to get from request body (guest spin)
	if msisdn == "" {
		var req struct {
			MSISDN string `json:"msisdn" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errors.Info("Failed to bind JSON", map[string]interface{}{
				"error": err.Error(),
			})
			middleware.RespondWithError(c, errors.BadRequest("MSISDN required for spin"))
			return
		}
		// Normalise guest MSISDN to canonical international format (234...)
		if normalized, err := validation.NormalizeMSISDN(req.MSISDN); err == nil {
			msisdn = normalized
		} else {
			msisdn = req.MSISDN
		}
		errors.Info("Guest spin request (normalised)", map[string]interface{}{
			"msisdn": msisdn,
		})
	} else {
		errors.Info("Authenticated spin request", map[string]interface{}{
			"msisdn": msisdn,
		})
	}

	// Service will validate spin eligibility
	errors.Info("Calling PlaySpin service", map[string]interface{}{
		"msisdn": msisdn,
	})
	result, err := h.spinService.PlaySpin(c.Request.Context(), msisdn)
	if err != nil {
		errors.Info("PlaySpin service error", map[string]interface{}{
			"error": err.Error(),
			"msisdn": msisdn,
		})
		middleware.RespondWithError(c, err)
		return
	}

	// Log spin result
	errors.Info("Spin played", map[string]interface{}{
		"msisdn":     msisdn,
		"prize_type": result.PrizeType,
		"prize_won":  result.PrizeWon,
	})

	middleware.RespondWithSuccess(c, result)
}

// GetHistory godoc
// @Summary Get spin history
// @Description Get user's spin history with pagination
// @Tags spin
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /spin/history [get]
func (h *SpinHandler) GetHistory(c *gin.Context) {
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

	// Calculate offset from page
	offset := (pagination.Page - 1) * pagination.Limit

	history, err := h.spinService.GetSpinHistory(c.Request.Context(), msisdn, pagination.Limit, offset)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"spins": history,
		"page":  pagination.Page,
		"limit": pagination.Limit,
	})
}

// Spin godoc
// @Summary Spin the wheel
// @Description Spin the wheel and get a prize (alias for PlaySpin)
// @Tags spin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /spin [post]
func (h *SpinHandler) Spin(c *gin.Context) {
	// Alias for PlaySpin to match frontend expectations
	h.PlaySpin(c)
}

// GetTiers godoc
// @Summary Get all spin tiers
// @Description Get all active spin tiers with prize information
// @Tags spin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} errors.ErrorResponse
// @Router /spins/tiers [get]
func (h *SpinHandler) GetTiers(c *gin.Context) {
	tiers, err := h.spinService.GetAllTiers(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"tiers": tiers,
	})
}

// GetTierProgress godoc
// @Summary Get user's tier progress
// @Description Get user's progress towards different spin tiers
// @Tags spin
// @Produce json
// @Param msisdn query string true "User's MSISDN"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /spins/tier-progress [get]
func (h *SpinHandler) GetTierProgress(c *gin.Context) {
	msisdn := c.Query("msisdn")
	if msisdn == "" {
		// Try to get from JWT context
		msisdn = c.GetString("msisdn")
	}
	// Normalise MSISDN if provided as query param (may be in local format)
	if msisdn != "" {
		if normalized, err := validation.NormalizeMSISDN(msisdn); err == nil {
			msisdn = normalized
		}
	}
	if msisdn == "" {
		middleware.RespondWithError(c, errors.BadRequest("MSISDN required"))
		return
	}

	progress, err := h.spinService.GetTierProgress(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, progress)
}
