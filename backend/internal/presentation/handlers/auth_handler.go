package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
	"rechargemax/internal/validation"
)

type AuthHandler struct {
	authService  *services.AuthService
	tokenService *services.TokenService
}

func NewAuthHandler(authService *services.AuthService, tokenService *services.TokenService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		tokenService: tokenService,
	}
}

// SendOTP godoc
// @Summary Send OTP
// @Description Send OTP to user's phone number
// @Tags auth
// @Accept json
// @Produce json
// @Param request body validation.SendOTPRequest true "Send OTP Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /auth/send-otp [post]
func (h *AuthHandler) SendOTP(c *gin.Context) {
	var req validation.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Send OTP with purpose
	purpose := req.Purpose
	if purpose == "" {
		purpose = "login" // Default to login if not specified
	}
	if err := h.authService.SendOTP(c.Request.Context(), req.MSISDN, purpose); err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log successful OTP send
	errors.Info("OTP sent", map[string]interface{}{
		"msisdn": req.MSISDN,
	})

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message": "OTP sent successfully",
	})
}

// VerifyOTP godoc
// @Summary Verify OTP
// @Description Verify OTP and get authentication token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body validation.VerifyOTPRequest true "Verify OTP Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Router /auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req validation.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Verify OTP with purpose
	purpose := req.Purpose
	if purpose == "" {
		purpose = "login" // Default to login if not specified
	}
	token, user, isNew, err := h.authService.VerifyOTP(c.Request.Context(), req.MSISDN, req.OTP, purpose)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log successful authentication
	errors.Info("User authenticated", map[string]interface{}{
		"msisdn": req.MSISDN,
		"is_new": isNew,
	})

	// Log audit trail
	errors.LogAudit(user.ID.String(), "LOGIN", "auth", map[string]interface{}{
		"msisdn": req.MSISDN,
		"is_new": isNew,
	})

	// Set httpOnly secure cookie (primary auth storage — not accessible to JS)
	// Token also returned in body for mobile/API clients that can't use cookies
	middleware.SetAuthCookie(c, token, "user", 24*60*60) // 24 hours

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"token":  token,
		"user":   user,
		"is_new": isNew,
	})
}

// Logout godoc
// @Summary User logout
// @Description Logout user - clears auth cookie and token
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if exists {
		errors.LogAudit(userID.(string), "LOGOUT", "auth", nil)
	}

	// Clear the httpOnly auth cookie
	middleware.ClearAuthCookie(c, "user")

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message": "Logged out successfully",
	})
}

// AdminLogin godoc
// @Summary Admin login
// @Description Admin login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body validation.AdminLoginRequest true "Admin Login Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Router /auth/admin/login [post]
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req validation.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Admin login - authentication handled by AuthService
	// token, admin, err := h.authService.AdminLogin(c.Request.Context(), req.Email, req.Password)
	// For now, return not implemented
	middleware.RespondWithError(c, errors.ServiceUnavailable("Admin login not yet implemented"))
}

// AdminLogout godoc
// @Summary Admin logout
// @Description Admin logout - clears session/token
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/admin/logout [post]
func (h *AuthHandler) AdminLogout(c *gin.Context) {
	// Get admin ID from context (set by auth middleware)
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		middleware.RespondWithError(c, errors.Unauthorized("Not authenticated"))
		return
	}
	
	// Get token from context
	tokenStr, tokenExists := c.Get("token")
	if !tokenExists {
		middleware.RespondWithError(c, errors.BadRequest("Token not found in context"))
		return
	}
	
	// Parse admin ID to UUID
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid admin ID format"))
		return
	}
	
	// Blacklist the token
	err = h.tokenService.BlacklistToken(c.Request.Context(), tokenStr.(string), adminID, "logout")
	if err != nil {
		errors.Error("Failed to blacklist token", err, map[string]interface{}{
			"admin_id": adminID,
		})
		// Don't fail the logout even if blacklisting fails
		// The token will still expire naturally
	}
	
	// Log logout event
	errors.Info("Admin logout", map[string]interface{}{
		"admin_id": adminID,
	})
	
	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message": "Logged out successfully",
	})
}
