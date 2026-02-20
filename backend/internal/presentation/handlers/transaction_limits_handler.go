package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
)

// TransactionLimitsHandler handles transaction limits management endpoints
type TransactionLimitsHandler struct {
	limitsService *services.TransactionLimitsService
}

// NewTransactionLimitsHandler creates a new transaction limits handler
func NewTransactionLimitsHandler(limitsService *services.TransactionLimitsService) *TransactionLimitsHandler {
	return &TransactionLimitsHandler{
		limitsService: limitsService,
	}
}

// GetLimit retrieves a specific limit by ID
func (h *TransactionLimitsHandler) GetLimit(c *gin.Context) {
	limitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid limit ID format"))
		return
	}

	limit, err := h.limitsService.GetLimit(c.Request.Context(), limitID)
	if err != nil {
		middleware.RespondWithError(c, errors.NotFound("Transaction limit"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   limit,
	})
}

// ListLimits retrieves all transaction limits with optional filtering
func (h *TransactionLimitsHandler) ListLimits(c *gin.Context) {
	limitType := c.Query("limit_type")
	limitScope := c.Query("limit_scope")
	activeOnly := c.Query("active_only") == "true"

	limits, err := h.limitsService.ListLimits(c.Request.Context(), limitType, limitScope, activeOnly)
	if err != nil {
		middleware.RespondWithError(c, errors.Internal("Failed to retrieve limits"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   limits,
		"count":  len(limits),
	})
}

// CreateLimit creates a new transaction limit
func (h *TransactionLimitsHandler) CreateLimit(c *gin.Context) {
	var req services.CreateLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request body"))
		return
	}

	// Get admin user ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, errors.Unauthorized("Admin authentication required"))
		return
	}

	limit, err := h.limitsService.CreateLimit(c.Request.Context(), &req, adminID.(uuid.UUID))
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Transaction limit created successfully",
		"data":    limit,
	})
}

// UpdateLimit updates an existing transaction limit
func (h *TransactionLimitsHandler) UpdateLimit(c *gin.Context) {
	limitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid limit ID format"))
		return
	}

	var req services.UpdateLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request body"))
		return
	}

	// Get admin user ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, errors.Unauthorized("Admin authentication required"))
		return
	}

	reason := c.Query("reason")
	limit, err := h.limitsService.UpdateLimit(c.Request.Context(), limitID, &req, adminID.(uuid.UUID), reason)
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Transaction limit updated successfully",
		"data":    limit,
	})
}

// DeleteLimit deactivates a transaction limit
func (h *TransactionLimitsHandler) DeleteLimit(c *gin.Context) {
	limitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid limit ID format"))
		return
	}

	// Get admin user ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, errors.Unauthorized("Admin authentication required"))
		return
	}

	reason := c.Query("reason")
	err = h.limitsService.DeleteLimit(c.Request.Context(), limitID, adminID.(uuid.UUID), reason)
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Transaction limit deactivated successfully",
	})
}

// CheckLimit validates if a transaction amount is within limits
func (h *TransactionLimitsHandler) CheckLimit(c *gin.Context) {
	limitType := c.Query("limit_type")
	if limitType == "" {
		middleware.RespondWithError(c, errors.BadRequest("limit_type is required"))
		return
	}

	var amount int64
	if err := c.ShouldBindQuery(&struct {
		Amount int64 `form:"amount" binding:"required"`
	}{Amount: amount}); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid amount parameter"))
		return
	}

	// Get user ID and tier from context (if authenticated)
	var userID uuid.UUID
	var userTier string = "bronze" // Default tier
	
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uuid.UUID)
		// TODO: Get user tier from database or context
	}

	result, err := h.limitsService.CheckLimit(c.Request.Context(), limitType, amount, userID, userTier)
	if err != nil {
		middleware.RespondWithError(c, errors.Internal("Failed to check limit"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

// GetAuditTrail retrieves audit trail for limit changes
func (h *TransactionLimitsHandler) GetAuditTrail(c *gin.Context) {
	_, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid limit ID format"))
		return
	}

	// TODO: Implement audit trail retrieval from database
	// For now, return empty array
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   []interface{}{},
		"count":  0,
	})
}
