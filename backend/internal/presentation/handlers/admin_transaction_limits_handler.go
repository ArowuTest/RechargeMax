package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rechargemax/internal/application/services"
)

// TransactionLimitsHandler handles CRUD for transaction limits.
type TransactionLimitsHandler struct {
	limitsSvc *services.TransactionLimitsService
}

// NewTransactionLimitsHandler creates a new TransactionLimitsHandler.
func NewTransactionLimitsHandler(limitsSvc *services.TransactionLimitsService) *TransactionLimitsHandler {
	return &TransactionLimitsHandler{limitsSvc: limitsSvc}
}

// ListTransactionLimits returns all limits.
func (h *TransactionLimitsHandler) ListTransactionLimits(c *gin.Context) {
	limits, err := h.limitsSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to retrieve transaction limits"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": limits})
}

// GetTransactionLimit returns a single limit by id.
func (h *TransactionLimitsHandler) GetTransactionLimit(c *gin.Context) {
	limit, err := h.limitsSvc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Transaction limit not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": limit})
}

// CreateTransactionLimit inserts a new limit.
func (h *TransactionLimitsHandler) CreateTransactionLimit(c *gin.Context) {
	var req services.TransactionLimit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err := h.limitsSvc.Create(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create transaction limit"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": req})
}

// UpdateTransactionLimit replaces an existing limit.
func (h *TransactionLimitsHandler) UpdateTransactionLimit(c *gin.Context) {
	var req services.TransactionLimit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	req.ID = c.Param("id")
	if err := h.limitsSvc.Update(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update transaction limit"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": req})
}

// DeleteTransactionLimit removes a limit.
func (h *TransactionLimitsHandler) DeleteTransactionLimit(c *gin.Context) {
	if err := h.limitsSvc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete transaction limit"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Transaction limit deleted"})
}
