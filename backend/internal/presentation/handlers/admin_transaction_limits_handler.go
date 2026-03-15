package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
)

// TransactionLimit mirrors the transaction_limits table.
type TransactionLimit struct {
	ID                uuid.UUID  `gorm:"primaryKey;type:uuid" json:"id"`
	LimitType         string     `json:"limit_type"`
	LimitScope        string     `json:"limit_scope"`
	MinAmount         int64      `json:"min_amount"`
	MaxAmount         int64      `json:"max_amount"`
	DailyLimit        *int64     `json:"daily_limit"`
	MonthlyLimit      *int64     `json:"monthly_limit"`
	IsActive          bool       `json:"is_active"`
	AppliesToUserTier *string    `json:"applies_to_user_tier"`
	Description       string     `json:"description"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (TransactionLimit) TableName() string { return "transaction_limits" }

// TransactionLimitsHandler handles admin CRUD for transaction_limits.
type TransactionLimitsHandler struct {
	db *gorm.DB
}

func NewTransactionLimitsHandler(db *gorm.DB) *TransactionLimitsHandler {
	return &TransactionLimitsHandler{db: db}
}

// ListTransactionLimits GET /admin/transaction-limits
func (h *TransactionLimitsHandler) ListTransactionLimits(c *gin.Context) {
	var limits []TransactionLimit
	if err := h.db.Order("limit_type, limit_scope").Find(&limits).Error; err != nil {
		middleware.RespondWithError(c, errors.Internal("Failed to fetch transaction limits"))
		return
	}
	middleware.RespondWithSuccess(c, limits)
}

// GetTransactionLimit GET /admin/transaction-limits/:id
func (h *TransactionLimitsHandler) GetTransactionLimit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid ID"))
		return
	}
	var limit TransactionLimit
	if err := h.db.First(&limit, "id = ?", id).Error; err != nil {
		middleware.RespondWithError(c, errors.NotFound("Transaction limit not found"))
		return
	}
	middleware.RespondWithSuccess(c, limit)
}

// CreateTransactionLimit POST /admin/transaction-limits
func (h *TransactionLimitsHandler) CreateTransactionLimit(c *gin.Context) {
	var req TransactionLimit
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request: "+err.Error()))
		return
	}
	req.ID = uuid.New()
	if err := h.db.Create(&req).Error; err != nil {
		middleware.RespondWithError(c, errors.Internal("Failed to create limit: "+err.Error()))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": req})
}

// UpdateTransactionLimit PUT /admin/transaction-limits/:id
func (h *TransactionLimitsHandler) UpdateTransactionLimit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid ID"))
		return
	}
	var req TransactionLimit
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request: "+err.Error()))
		return
	}
	req.ID = id
	if err := h.db.Save(&req).Error; err != nil {
		middleware.RespondWithError(c, errors.Internal("Failed to update limit: "+err.Error()))
		return
	}
	middleware.RespondWithSuccess(c, req)
}

// DeleteTransactionLimit DELETE /admin/transaction-limits/:id
func (h *TransactionLimitsHandler) DeleteTransactionLimit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid ID"))
		return
	}
	if err := h.db.Delete(&TransactionLimit{}, "id = ?", id).Error; err != nil {
		middleware.RespondWithError(c, errors.Internal("Failed to delete limit"))
		return
	}
	middleware.RespondWithSuccess(c, gin.H{"deleted": true})
}
