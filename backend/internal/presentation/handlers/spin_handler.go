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
	// MSISDN resolution — two paths:
	//
	// Path A (authenticated user — JWT present):
	//   The MSISDN is extracted from the verified JWT by OptionalAuthMiddleware.
	//   The request body is ignored for MSISDN purposes.
	//
	// Path B (guest user — no JWT):
	//   The MSISDN is taken from the request body.
	//   The spin service will validate that this MSISDN has a qualifying
	//   recharge within the last 4 hours before allowing the spin.
	//   This preserves the "recharge → spin → log in to claim" UX while
	//   making it impractical to spin on behalf of an arbitrary number
	//   (the attacker would also need to have recharged for that number
	//   within the last 4 hours).
	msisdn := c.GetString("msisdn") // set by OptionalAuthMiddleware when JWT is valid

	if msisdn == "" {
		// Path B-prime: JWT was present and valid but its msisdn claim is empty
		// (legacy tokens issued before MSISDN normalisation).  If user_id is in
		// context we can still identify the user — look up their MSISDN by ID.
		if userID := c.GetString("user_id"); userID != "" {
			resolved, err := h.spinService.ResolveMSISDNFromUserID(c.Request.Context(), userID)
			if err == nil && resolved != "" {
				msisdn = resolved
				errors.Info("Resolved MSISDN from user_id (stale JWT)", map[string]interface{}{
					"user_id": userID,
					"msisdn":  msisdn,
				})
			}
		}
	}

	if msisdn == "" {
		// Guest path: require MSISDN in request body (or optional msisdn param)
		var req struct {
			MSISDN string `json:"msisdn"`
		}
		_ = c.ShouldBindJSON(&req) // non-binding — missing field is handled below
		if req.MSISDN != "" {
			// Normalise to international format (234XXXXXXXXXX)
			if normalized, err := validation.NormalizeMSISDN(req.MSISDN); err == nil {
				msisdn = normalized
			} else {
				msisdn = req.MSISDN
			}
		}
		if msisdn == "" {
			middleware.RespondWithError(c, errors.BadRequest("Phone number required. Please provide your MSISDN."))
			return
		}
		errors.Info("Guest spin request", map[string]interface{}{"msisdn": msisdn})
	} else {
		errors.Info("Authenticated spin request", map[string]interface{}{"msisdn": msisdn})
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
	// Support both authenticated users (JWT) and guest users (query param)
	// OptionalAuthMiddleware sets msisdn in context if JWT is present
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		// Fall back to msisdn query parameter for guest access
		msisdn = c.Query("msisdn")
		if msisdn != "" {
			// Normalise MSISDN from query param
			if normalized, err := validation.NormalizeMSISDN(msisdn); err == nil {
				msisdn = normalized
			}
		}
	}
	if msisdn == "" {
		middleware.RespondWithError(c, errors.BadRequest("MSISDN required: provide via Authorization header or msisdn query parameter"))
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

// GetPrizes godoc
// @Summary Get all spin prizes
// @Description Get all available spin prizes (public endpoint)
// @Tags spin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} errors.ErrorResponse
// @Router /spin/prizes [get]
func (h *SpinHandler) GetPrizes(c *gin.Context) {
	prizes, err := h.spinService.GetAllPrizes(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"data": prizes,
	})
}

// DebugSpins is a TEMPORARY admin diagnostic endpoint — remove after investigation.
// Use msisdn=__ALL__ to dump all rows in the table.
func (h *SpinHandler) DebugSpins(c *gin.Context) {
	msisdn := c.Query("msisdn")
	if msisdn == "" {
		msisdn = "__ALL__"
	}
	rows, total, err := h.spinService.DebugSpinResults(c.Request.Context(), msisdn)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"msisdn": msisdn, "total_in_table": total, "returned": len(rows), "rows": rows})
}
