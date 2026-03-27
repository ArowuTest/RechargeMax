package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type AdminUserManagementHandler struct {
	adminRepo repositories.AdminRepository
}

func NewAdminUserManagementHandler(adminRepo repositories.AdminRepository) *AdminUserManagementHandler {
	return &AdminUserManagementHandler{
		adminRepo: adminRepo,
	}
}

// GetAllAdmins returns list of all admin users
func (h *AdminUserManagementHandler) GetAllAdmins(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	// Get admins
	admins, err := h.adminRepo.FindAll(ctx, perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve admin users",
		})
		return
	}

	// Get total count
	total, err := h.adminRepo.Count(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to count admin users",
		})
		return
	}

	// Remove password hashes from response
	for _, admin := range admins {
		admin.PasswordHash = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    admins,
		"pagination": gin.H{
			"page":     page,
			"per_page": perPage,
			"total":    total,
		},
	})
}

// GetAdminByID returns a specific admin user
func (h *AdminUserManagementHandler) GetAdminByID(c *gin.Context) {
	ctx := c.Request.Context()
	adminID := c.Param("id")

	admin, err := h.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Admin user not found",
		})
		return
	}

	// Remove password hash
	admin.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    admin,
	})
}

// CreateAdmin creates a new admin user
func (h *AdminUserManagementHandler) CreateAdmin(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Email       string   `json:"email" binding:"required,email"`
		Password    string   `json:"password"`
		FullName    string   `json:"full_name" binding:"required"`
		Role        string   `json:"role" binding:"required"`
		IsActive    *bool    `json:"is_active"`
		Permissions []string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Validate role
	validRoles := map[string]bool{
		"SUPER_ADMIN": true,
		"ADMIN":       true,
		"MODERATOR":   true,
		"VIEWER":      true,
	}
	if !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid role. Must be one of: SUPER_ADMIN, ADMIN, MODERATOR, VIEWER",
		})
		return
	}

	// Check if email already exists
	existing, _ := h.adminRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Admin user with this email already exists",
		})
		return
	}

	// Auto-generate a temporary password if none provided
	tempPassword := ""
	if req.Password == "" {
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#$"
		rand.Seed(time.Now().UnixNano())
		b := make([]byte, 12)
		for i := range b {
			b[i] = charset[rand.Intn(len(charset))]
		}
		req.Password = string(b)
		tempPassword = req.Password
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to hash password",
		})
		return
	}

	// Set default active status
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Build permissions — use provided list or fall back to role-based defaults
	permsByRole := map[string][]string{
		"SUPER_ADMIN": {"view_analytics", "manage_users", "manage_transactions", "manage_networks", "manage_prizes", "manage_affiliates", "manage_settings", "manage_admins", "view_monitoring", "manage_draws"},
		"ADMIN":       {"view_analytics", "manage_users", "manage_transactions", "manage_networks", "manage_prizes", "manage_affiliates", "view_monitoring", "manage_draws"},
		"MODERATOR":   {"view_analytics", "manage_users", "manage_transactions", "view_monitoring"},
		"VIEWER":      {"view_analytics", "view_monitoring"},
	}
	permsToUse := req.Permissions
	if len(permsToUse) == 0 {
		permsToUse = permsByRole[req.Role]
	}
	permsJSON, _ := json.Marshal(permsToUse)

	// Create admin user
	admin := &entities.AdminUsers{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		Role:         req.Role,
		IsActive:     &isActive,
		Permissions:  datatypes.JSON(permsJSON),
	}

	if err := h.adminRepo.Create(ctx, admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create admin user: " + err.Error(),
		})
		return
	}

	// Remove password hash from response
	admin.PasswordHash = ""

	response := gin.H{
		"success": true,
		"message": "Admin user created successfully",
		"data":    admin,
	}
	if tempPassword != "" {
		response["temporary_password"] = tempPassword
		response["message"] = "Admin user created successfully. Please share the temporary password with the new admin."
	}
	c.JSON(http.StatusCreated, response)
}

// UpdateAdmin updates an existing admin user
func (h *AdminUserManagementHandler) UpdateAdmin(c *gin.Context) {
	ctx := c.Request.Context()
	adminID := c.Param("id")

	var req struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
		FullName *string `json:"full_name"`
		Role     *string `json:"role"`
		IsActive *bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get existing admin
	admin, err := h.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Admin user not found",
		})
		return
	}

	// Update fields if provided
	if req.Email != nil {
		// Check if new email already exists
		existing, _ := h.adminRepo.GetByEmail(ctx, *req.Email)
		if existing != nil && existing.ID != adminID {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Admin user with this email already exists",
			})
			return
		}
		admin.Email = *req.Email
	}

	if req.Password != nil {
		if len(*req.Password) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Password must be at least 8 characters",
			})
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to hash password",
			})
			return
		}
		admin.PasswordHash = string(hashedPassword)
	}

	if req.FullName != nil {
		admin.FullName = *req.FullName
	}

	if req.Role != nil {
		validRoles := map[string]bool{
			"SUPER_ADMIN": true,
			"ADMIN":       true,
			"MODERATOR":   true,
			"VIEWER":      true,
		}
		if !validRoles[*req.Role] {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid role. Must be one of: SUPER_ADMIN, ADMIN, MODERATOR, VIEWER",
			})
			return
		}
		admin.Role = *req.Role
	}

	if req.IsActive != nil {
		admin.IsActive = req.IsActive
	}

	// Update admin
	if err := h.adminRepo.Update(ctx, admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update admin user",
		})
		return
	}

	// Remove password hash from response
	admin.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Admin user updated successfully",
		"data":    admin,
	})
}

// DeleteAdmin deletes an admin user
func (h *AdminUserManagementHandler) DeleteAdmin(c *gin.Context) {
	ctx := c.Request.Context()
	adminID := c.Param("id")

	// Parse UUID
	id, err := uuid.Parse(adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid admin ID format",
		})
		return
	}

	// Check if admin exists
	_, err = h.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Admin user not found",
		})
		return
	}

	// Prevent deleting the last super admin
	allAdmins, _ := h.adminRepo.FindAll(ctx, 1000, 0)
	superAdminCount := 0
	for _, a := range allAdmins {
		if a.Role == "super_admin" {
			superAdminCount++
		}
	}
	// Check the target admin's role
	targetAdmin, _ := h.adminRepo.GetByID(ctx, adminID)
	if targetAdmin != nil && targetAdmin.Role == "super_admin" && superAdminCount <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Cannot delete the last super admin account",
		})
		return
	}

	// Delete admin
	if err := h.adminRepo.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete admin user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Admin user deleted successfully",
	})
}

// UpdateAdminStatus updates only the active status of an admin
func (h *AdminUserManagementHandler) UpdateAdminStatus(c *gin.Context) {
	ctx := c.Request.Context()
	adminID := c.Param("id")

	var req struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Get existing admin
	admin, err := h.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Admin user not found",
		})
		return
	}

	// Update status
	admin.IsActive = &req.IsActive

	if err := h.adminRepo.Update(ctx, admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update admin status",
		})
		return
	}

	// Remove password hash from response
	admin.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Admin status updated successfully",
		"data":    admin,
	})
}
