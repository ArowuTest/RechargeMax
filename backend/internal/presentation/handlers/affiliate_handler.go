package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"rechargemax/internal/application/services"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
	"rechargemax/internal/validation"
)

type AffiliateHandler struct {
	affiliateService *services.AffiliateService
}

func NewAffiliateHandler(affiliateService *services.AffiliateService) *AffiliateHandler {
	return &AffiliateHandler{affiliateService: affiliateService}
}

// GetReferralCode godoc
// @Summary Get referral code
// @Description Get user's affiliate referral code
// @Tags affiliate
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /affiliate/referral-code [get]
func (h *AffiliateHandler) GetReferralCode(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	affiliate, err := h.affiliateService.GetAffiliateByMSISDN(c.Request.Context(), msisdn)
	if err != nil {
		// User is not registered as an affiliate - return empty response
		middleware.RespondWithSuccess(c, map[string]interface{}{
			"referral_code":   nil,
			"is_affiliate":    false,
			"message":         "Not registered as an affiliate",
		})
		return
	}

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"referral_code": affiliate.AffiliateCode,
		"is_affiliate":  true,
	})
}

// GetStats godoc
// @Summary Get affiliate stats
// @Description Get affiliate statistics and performance metrics
// @Tags affiliate
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /affiliate/stats [get]
func (h *AffiliateHandler) GetStats(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	stats, err := h.affiliateService.GetAffiliateStats(c.Request.Context(), msisdn)
	if err != nil {
		// User is not registered as an affiliate - return empty stats
		middleware.RespondWithSuccess(c, map[string]interface{}{
			"is_affiliate":       false,
			"total_referrals":    0,
			"total_earnings":     0,
			"pending_earnings":   0,
			"message":            "Not registered as an affiliate",
		})
		return
	}

	middleware.RespondWithSuccess(c, stats)
}

// GetReferrals godoc
// @Summary Get referrals
// @Description Get list of users referred by affiliate
// @Tags affiliate
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /affiliate/referrals [get]
func (h *AffiliateHandler) GetReferrals(c *gin.Context) {
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

	// Get affiliate stats which includes referral count
	stats, err := h.affiliateService.GetAffiliateStats(c.Request.Context(), msisdn)
	if err != nil {
		// User is not an affiliate - return empty referrals list
		middleware.RespondWithSuccess(c, map[string]interface{}{
			"referrals":    []interface{}{},
			"total":        0,
			"page":         pagination.Page,
			"limit":        pagination.Limit,
			"is_affiliate": false,
		})
		return
	}
	
	middleware.RespondWithSuccess(c, map[string]interface{}{
		"referrals": stats,
		"page":      pagination.Page,
		"limit":     pagination.Limit,
	})
}

// Register godoc
// @Summary Register as affiliate
// @Description Register user as an affiliate
// @Tags affiliate
// @Accept json
// @Produce json
// @Param request body validation.RegisterAffiliateRequest true "Affiliate registration request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /affiliate/register [post]
func (h *AffiliateHandler) Register(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	var req validation.RegisterAffiliateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Set MSISDN from JWT context if not provided in body
	if req.MSISDN == "" {
		req.MSISDN = msisdn
	}

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Service will validate affiliate eligibility
	affiliateReq := services.RegisterAffiliateRequest{
		MSISDN:        msisdn,
		BankName:      req.BankName,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
	}

	affiliate, err := h.affiliateService.RegisterAffiliate(c.Request.Context(), affiliateReq)
	if err != nil {
		fmt.Printf("[AFFILIATE REGISTER ERROR] %v\n", err)
		middleware.RespondWithError(c, err)
		return
	}

	// Log affiliate registration
	errors.Info("Affiliate registered", map[string]interface{}{
		"msisdn":         msisdn,
		"affiliate_code": affiliate.AffiliateCode,
	})

	middleware.RespondWithSuccess(c, affiliate)
}

// GetDashboard godoc
// @Summary Get affiliate dashboard
// @Description Get affiliate dashboard with stats and earnings
// @Tags affiliate
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /affiliate/dashboard [get]
func (h *AffiliateHandler) GetDashboard(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	// Check if user is an affiliate
	affiliate, err := h.affiliateService.GetAffiliateByMSISDN(c.Request.Context(), msisdn)
	if err != nil {
		// User is not an affiliate
		middleware.RespondWithSuccess(c, map[string]interface{}{
			"is_affiliate": false,
		})
		return
	}

	// Get stats
	stats, err := h.affiliateService.GetAffiliateStats(c.Request.Context(), msisdn)
	if err != nil {
		stats = nil
	}

	// Build response in format the frontend expects
	response := map[string]interface{}{
		"is_affiliate": true,
		"affiliate": affiliate,
		"statistics":  stats,
		"bank_accounts": []interface{}{},
		"referral_link": affiliate.ReferralLink,
	}
	middleware.RespondWithSuccess(c, response)
}

// GetReferralLink godoc
// @Summary Get referral link
// @Description Get affiliate referral link
// @Tags affiliate
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /affiliate/referral-link [get]
func (h *AffiliateHandler) GetReferralLink(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	affiliate, err := h.affiliateService.GetAffiliateByMSISDN(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Generate referral link (base URL should come from config)
	baseURL := "https://rechargemax.com" // NOTE: Should be loaded from environment config
	referralLink := baseURL + "/register?ref=" + affiliate.AffiliateCode

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"referral_link": referralLink,
		"referral_code": affiliate.AffiliateCode,
	})
}

// GetCommissions godoc
// @Summary Get affiliate commissions
// @Description Get list of affiliate commissions with pagination
// @Tags affiliate
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /affiliate/commissions [get]
func (h *AffiliateHandler) GetCommissions(c *gin.Context) {
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

	commissions, err := h.affiliateService.GetCommissions(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"commissions": commissions,
		"page":        pagination.Page,
		"limit":       pagination.Limit,
	})
}

// RequestPayout godoc
// @Summary Request affiliate payout
// @Description Request payout of affiliate earnings
// @Tags affiliate
// @Accept json
// @Produce json
// @Param request body validation.PayoutRequest true "Payout request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /affiliate/payout [post]
func (h *AffiliateHandler) RequestPayout(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	var req validation.PayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Service will validate withdrawal eligibility
	// Convert Naira to kobo (multiply by 100)
	amountKobo := int64(req.Amount * 100)
	payout, err := h.affiliateService.RequestPayout(c.Request.Context(), msisdn, amountKobo)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log payout request
	errors.Info("Payout requested", map[string]interface{}{
		"msisdn": msisdn,
		"amount": req.Amount,
	})

	middleware.RespondWithSuccess(c, payout)
}

// GetEarnings godoc
// @Summary Get affiliate earnings summary
// @Description Get total earnings, pending, and paid amounts for affiliate
// @Tags affiliate
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /affiliate/earnings [get]
func (h *AffiliateHandler) GetEarnings(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	// Get earnings summary
	earnings, err := h.affiliateService.GetEarningsSummary(c.Request.Context(), msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, earnings)
}

// TrackClick records an affiliate link click for attribution.
// The click source and referral code are optional — pulled from query/body.
// Authenticated users get their MSISDN recorded; unauthenticated calls are allowed.
func (h *AffiliateHandler) TrackClick(c *gin.Context) {
	var req struct {
		ReferralCode string `json:"referral_code"`
		Source       string `json:"source"`
	}
	_ = c.ShouldBindJSON(&req) // body is optional

	msisdn := c.GetString("msisdn")

	if err := h.affiliateService.RecordClick(c.Request.Context(), req.ReferralCode, msisdn, req.Source); err != nil {
		// Non-critical: log but don't return an error to the caller
		errors.Info("affiliate click track failed", map[string]interface{}{
			"error":         err.Error(),
			"referral_code": req.ReferralCode,
		})
	}

	middleware.RespondWithSuccess(c, map[string]interface{}{"tracked": true})
}
