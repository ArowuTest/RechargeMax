package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
	"rechargemax/internal/validation"
)

type WinnerHandler struct {
	winnerService *services.WinnerService
}

func NewWinnerHandler(winnerService *services.WinnerService) *WinnerHandler {
	return &WinnerHandler{winnerService: winnerService}
}

// GetMyWins godoc
// @Summary Get my wins
// @Description Get list of prizes won by the authenticated user
// @Tags winner
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /winner/my-wins [get]
func (h *WinnerHandler) GetMyWins(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	wins, err := h.winnerService.GetWinnersByMSISDN(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, wins)
}

// ClaimPrize godoc
// @Summary Claim a prize
// @Description Submit claim request for a won prize
// @Tags winner
// @Accept json
// @Produce json
// @Param id path string true "Winner ID"
// @Param request body validation.ClaimPrizeRequest true "Claim request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /winner/{id}/claim [post]
func (h *WinnerHandler) ClaimPrize(c *gin.Context) {
	winnerID := c.Param("id")
	msisdn := c.GetString("msisdn")
	
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	if winnerID == "" {
		middleware.RespondWithError(c, errors.BadRequest("Winner ID is required"))
		return
	}

	// Parse and validate UUID
	winnerIDUUID, err := uuid.Parse(winnerID)
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid winner ID"))
		return
	}
	
	var req struct {
		BankAccountID string `json:"bank_account_id,omitempty"`
		AccountNumber string `json:"account_number,omitempty"`
		BankCode      string `json:"bank_code,omitempty"`
		AccountName   string `json:"account_name,omitempty"`
		Address       string `json:"address,omitempty"`
		PhoneNumber   string `json:"phone_number,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Process claim with details
	claimDetails := map[string]interface{}{
		"bank_code":      req.BankCode,
		"account_number": req.AccountNumber,
		"account_name":   req.AccountName,
		"address":        req.Address,
		"phone_number":   req.PhoneNumber,
	}
	err = h.winnerService.ClaimPrize(c.Request.Context(), winnerIDUUID, msisdn, claimDetails)

	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log prize claim
	errors.Info("Prize claimed", map[string]interface{}{
		"msisdn":    msisdn,
		"winner_id": winnerID,
	})

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message": "Prize claim submitted successfully",
	})
}

// GetWinner godoc
// @Summary Get single winner details
// @Description Get details of a specific winner by ID
// @Tags winner
// @Produce json
// @Param id path string true "Winner ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /winner/{id} [get]
func (h *WinnerHandler) GetWinner(c *gin.Context) {
	winnerID := c.Param("id")
	msisdn := c.GetString("msisdn")
	
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	if winnerID == "" {
		middleware.RespondWithError(c, errors.BadRequest("Winner ID is required"))
		return
	}

	// Validate UUID
	if err := validation.ValidateUUID(winnerID); err != nil {
		middleware.RespondWithError(c, errors.BadRequest(err.Error()))
		return
	}

	// Get winner details
	winner, err := h.winnerService.GetWinnerByID(c.Request.Context(), winnerID, msisdn)
	if err != nil {
		middleware.RespondWithError(c, errors.NotFound("Winner not found"))
		return
	}

	middleware.RespondWithSuccess(c, winner)
}

// UpdateWinner godoc
// @Summary Update winner status (Admin only)
// @Description Update winner claim status, payment status, etc.
// @Tags admin, winner
// @Accept json
// @Produce json
// @Param id path string true "Winner ID"
// @Param request body map[string]interface{} true "Update request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /admin/winners/{id} [put]
func (h *WinnerHandler) UpdateWinner(c *gin.Context) {
	winnerID := c.Param("id")

	if winnerID == "" {
		middleware.RespondWithError(c, errors.BadRequest("Winner ID is required"))
		return
	}

	// Validate UUID
	if err := validation.ValidateUUID(winnerID); err != nil {
		middleware.RespondWithError(c, errors.BadRequest(err.Error()))
		return
	}

	var req struct {
		ClaimStatus   string `json:"claim_status"`
		PaymentStatus string `json:"payment_status"`
		Notes         string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Validate status values
	if req.ClaimStatus != "" {
		if err := validation.ValidateStatus(req.ClaimStatus); err != nil {
			middleware.RespondWithError(c, errors.BadRequest(err.Error()))
			return
		}
	}

	if req.PaymentStatus != "" {
		if err := validation.ValidateStatus(req.PaymentStatus); err != nil {
			middleware.RespondWithError(c, errors.BadRequest(err.Error()))
			return
		}
	}

	// Update winner
	winner, err := h.winnerService.UpdateWinnerStatus(c.Request.Context(), winnerID, services.UpdateWinnerRequest{
		ClaimStatus:   req.ClaimStatus,
		PaymentStatus: req.PaymentStatus,
		Notes:         req.Notes,
	})

	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log winner update
	errors.Info("Winner updated", map[string]interface{}{
		"winner_id":      winnerID,
		"claim_status":   req.ClaimStatus,
		"payment_status": req.PaymentStatus,
	})

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message": "Winner updated successfully",
		"winner":  winner,
	})
}
