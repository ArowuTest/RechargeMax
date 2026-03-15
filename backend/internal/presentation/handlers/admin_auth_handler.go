package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"rechargemax/internal/domain/repositories"
	"rechargemax/internal/middleware"
)

type AdminAuthHandler struct {
	adminRepo repositories.AdminRepository
	jwtSecret string
}

func NewAdminAuthHandler(adminRepo repositories.AdminRepository, jwtSecret string) *AdminAuthHandler {
	return &AdminAuthHandler{
		adminRepo: adminRepo,
		jwtSecret: jwtSecret,
	}
}

// LoginRequest represents admin login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents admin login response
type LoginResponse struct {
	Success bool        `json:"success"`
	Token   string      `json:"token,omitempty"`
	Admin   interface{} `json:"admin,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Login handles admin login
func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	ctx := c.Request.Context()

	// Get admin by email
	admin, err := h.adminRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Error:   "Invalid credentials",
		})
		return
	}

	// Check if admin is active
	if admin.IsActive == nil || !*admin.IsActive {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Error:   "Account is inactive",
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Error:   "Invalid credentials",
		})
		return
	}

	// Generate JWT token
	token, err := h.generateToken(admin.ID, admin.Email, admin.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, LoginResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	// Update last login
	if err := h.adminRepo.UpdateLastLogin(ctx, admin.ID); err != nil {
		// Log error but don't fail the login
		// TODO: Add proper logging
	}

	// Set httpOnly cookie for browser-based admin panel
	middleware.SetAuthCookie(c, token, "admin", 24*60*60) // 24 hours

	// Return success response
	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Token:   token,
		Admin: gin.H{
			"id":         admin.ID,
			"email":      admin.Email,
			"full_name":  admin.FullName,
			"role":       admin.Role,
			"permissions": admin.Permissions,
		},
	})
}

// generateToken generates a JWT token for admin
func (h *AdminAuthHandler) generateToken(adminID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"admin_id": adminID,
		"email":    email,
		"role":     role,
		"type":     "admin",
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

// RefreshToken handles admin token refresh
func (h *AdminAuthHandler) RefreshToken(c *gin.Context) {
	// Get admin ID from context (set by auth middleware)
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	ctx := c.Request.Context()

	// Get admin details
	admin, err := h.adminRepo.GetByID(ctx, adminID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Error:   "Invalid admin",
		})
		return
	}

	// Check if admin is still active
	if admin.IsActive == nil || !*admin.IsActive {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Error:   "Account is inactive",
		})
		return
	}

	// Generate new token
	token, err := h.generateToken(admin.ID, admin.Email, admin.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, LoginResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Token:   token,
	})
}

// Logout handles admin logout
func (h *AdminAuthHandler) Logout(c *gin.Context) {
	// Clear the httpOnly admin auth cookie
	middleware.ClearAuthCookie(c, "admin")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// GetProfile returns the current admin's profile
func (h *AdminAuthHandler) GetProfile(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	ctx := c.Request.Context()

	admin, err := h.adminRepo.GetByID(ctx, adminID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":          admin.ID,
			"email":       admin.Email,
			"full_name":   admin.FullName,
			"role":        admin.Role,
			"permissions": admin.Permissions,
			"is_active":   admin.IsActive,
			"created_at":  admin.CreatedAt,
			"last_login":  admin.LastLogin,
		},
	})
}
