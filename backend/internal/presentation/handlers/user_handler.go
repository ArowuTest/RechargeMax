package handlers

import (
	"github.com/gin-gonic/gin"
	
	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
	"rechargemax/internal/validation"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userService   *services.UserService
	walletService *services.WalletService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService, walletService *services.WalletService) *UserHandler {
	return &UserHandler{
		userService:   userService,
		walletService: walletService,
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get authenticated user's profile information
// @Tags user
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	profile, err := h.userService.GetUserProfile(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, profile)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user's profile information
// @Tags user
// @Accept json
// @Produce json
// @Param request body validation.UpdateProfileRequest true "Profile update request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	var req validation.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Update profile
	profile, err := h.userService.UpdateUserProfile(c.Request.Context(), msisdn, services.UpdateProfileRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})

	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log profile update
	errors.Info("Profile updated", map[string]interface{}{
		"msisdn": msisdn,
		"email":  req.Email,
	})

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message": "Profile updated successfully",
		"profile": profile,
	})
}

// GetWallet godoc
// @Summary Get user wallet
// @Description Get authenticated user's wallet balance and details
// @Tags user
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /user/wallet [get]
func (h *UserHandler) GetWallet(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	wallet, err := h.walletService.GetWallet(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, wallet)
}

// GetTransactions godoc
// @Summary Get user transactions
// @Description Get authenticated user's transaction history with pagination
// @Tags user
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /user/transactions [get]
func (h *UserHandler) GetTransactions(c *gin.Context) {
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

	// Get transactions and total count in parallel
	transactions, err := h.userService.GetTransactions(c.Request.Context(), msisdn, pagination.Limit, offset)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	total, err := h.userService.CountTransactions(c.Request.Context(), msisdn)
	if err != nil {
		total = 0 // non-fatal: return data without total
	}

	totalPages := int64(0)
	if pagination.Limit > 0 {
		totalPages = (total + int64(pagination.Limit) - 1) / int64(pagination.Limit)
	}

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"transactions": transactions,
		"page":         pagination.Page,
		"limit":        pagination.Limit,
		"total":        total,
		"total_pages":  totalPages,
	})
}

// GetPrizes godoc
// @Summary Get user prizes
// @Description Get list of prizes won by the authenticated user
// @Tags user
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /user/prizes [get]
func (h *UserHandler) GetPrizes(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	prizes, err := h.userService.GetUserPrizes(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, prizes)
}


// GetDashboard godoc
// @Summary Get user dashboard
// @Description Get comprehensive dashboard data for authenticated user
// @Tags user
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /user/dashboard [get]
func (h *UserHandler) GetDashboard(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	dashboard, err := h.userService.GetDashboard(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, dashboard)
}
