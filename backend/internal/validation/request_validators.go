package validation

import (
	"fmt"
	"time"

	"rechargemax/internal/utils"
)

// RechargeRequest validation
type RechargeRequest struct {
	MSISDN  string  `json:"msisdn"`
	Amount  float64 `json:"amount"`
	Network string  `json:"network"`
	Type    string  `json:"type"` // AIRTIME or DATA
}

func (r *RechargeRequest) Validate() error {
	var errs ValidationErrors
	
	if err := ValidateMSISDN(r.MSISDN); err != nil {
		errs.Add("msisdn", err.Error())
	}
	
	if err := ValidateRechargeAmount(r.Amount); err != nil {
		errs.Add("amount", err.Error())
	}
	
	if err := ValidateNetwork(r.Network); err != nil {
		errs.Add("network", err.Error())
	}
	
	if err := ValidateTransactionType(r.Type); err != nil {
		errs.Add("type", err.Error())
	}
	
	// Validate phone number matches network
	if err := ValidatePhoneNetworkMatch(r.MSISDN, r.Network); err != nil {
		errs.Add("network", err.Error())
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// SubscriptionRequest validation
type SubscriptionRequest struct {
	MSISDN string `json:"msisdn"`
}

func (r *SubscriptionRequest) Validate() error {
	var errs ValidationErrors
	
	if err := ValidateMSISDN(r.MSISDN); err != nil {
		errs.Add("msisdn", err.Error())
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// SpinRequest validation
type SpinRequest struct {
	TransactionID string `json:"transaction_id"`
}

func (r *SpinRequest) Validate() error {
	var errs ValidationErrors
	
	if err := ValidateUUID(r.TransactionID); err != nil {
		errs.Add("transaction_id", err.Error())
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// ClaimPrizeRequest validation
type ClaimPrizeRequest struct {
	WinnerID string `json:"winner_id"`
}

func (r *ClaimPrizeRequest) Validate() error {
	var errs ValidationErrors
	
	if err := ValidateUUID(r.WinnerID); err != nil {
		errs.Add("winner_id", err.Error())
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// UpdateProfileRequest validation
type UpdateProfileRequest struct {
	Email     string `json:"email,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

func (r *UpdateProfileRequest) Validate() error {
	var errs ValidationErrors
	
	if r.Email != "" {
		if err := ValidateEmail(r.Email); err != nil {
			errs.Add("email", err.Error())
		}
	}
	
	if r.FirstName != "" && len(r.FirstName) < 2 {
		errs.Add("first_name", "first name must be at least 2 characters")
	}
	
	if r.LastName != "" && len(r.LastName) < 2 {
		errs.Add("last_name", "last name must be at least 2 characters")
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// RegisterRequest validation
type RegisterRequest struct {
	MSISDN       string `json:"msisdn"`
	Password     string `json:"password"`
	Email        string `json:"email,omitempty"`
	ReferralCode string `json:"referral_code,omitempty"`
}

func (r *RegisterRequest) Validate() error {
	var errs ValidationErrors
	
	if err := ValidateMSISDN(r.MSISDN); err != nil {
		errs.Add("msisdn", err.Error())
	}
	
	if err := ValidatePassword(r.Password); err != nil {
		errs.Add("password", err.Error())
	}
	
	if r.Email != "" {
		if err := ValidateEmail(r.Email); err != nil {
			errs.Add("email", err.Error())
		}
	}
	
	if r.ReferralCode != "" {
		if err := ValidateReferralCode(r.ReferralCode); err != nil {
			errs.Add("referral_code", err.Error())
		}
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// LoginRequest validation
type LoginRequest struct {
	MSISDN   string `json:"msisdn"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	var errs ValidationErrors
	
	if err := ValidateMSISDN(r.MSISDN); err != nil {
		errs.Add("msisdn", err.Error())
	}
	
	if r.Password == "" {
		errs.Add("password", "password is required")
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// CreatePrizeRequest validation (Admin)
type CreatePrizeRequest struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Value       float64 `json:"value"`
	Probability float64 `json:"probability"`
	Description string  `json:"description,omitempty"`
}

func (r *CreatePrizeRequest) Validate() error {
	var errs ValidationErrors
	
	if r.Name == "" {
		errs.Add("name", "prize name is required")
	}
	
	if err := ValidatePrizeType(r.Type); err != nil {
		errs.Add("type", err.Error())
	}
	
	if err := ValidatePrizeValue(r.Type, r.Value); err != nil {
		errs.Add("value", err.Error())
	}
	
	if err := ValidateProbability(r.Probability); err != nil {
		errs.Add("probability", err.Error())
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// UpdateCommissionRateRequest validation (Admin)
type UpdateCommissionRateRequest struct {
	Rate float64 `json:"rate"`
}

func (r *UpdateCommissionRateRequest) Validate() error {
	var errs ValidationErrors
	
	if err := ValidateCommissionRate(r.Rate); err != nil {
		errs.Add("rate", err.Error())
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// CreateDrawRequest validation (Admin)
type CreateDrawRequest struct {
	Name        string `json:"name"`
	DrawDate    string `json:"draw_date"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Description string `json:"description,omitempty"`
}

func (r *CreateDrawRequest) Validate() error {
	var errs ValidationErrors
	
	if r.Name == "" {
		errs.Add("name", "draw name is required")
	}
	
	if r.DrawDate != "" {
		drawDate, err := time.Parse("2006-01-02", r.DrawDate)
		if err != nil {
			errs.Add("draw_date", "invalid date format, use YYYY-MM-DD")
		} else if err := ValidateDate(drawDate, false); err != nil {
			errs.Add("draw_date", err.Error())
		}
	}
	
	if err := ValidateDateRange(r.StartDate, r.EndDate); err != nil {
		errs.Add("date_range", err.Error())
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// UpdateNetworkRequest validation (Admin)
type UpdateNetworkRequest struct {
	NetworkCode string  `json:"network_code"`
	IsActive    *bool   `json:"is_active,omitempty"`
	MinAmount   float64 `json:"min_amount,omitempty"`
	MaxAmount   float64 `json:"max_amount,omitempty"`
}

func (r *UpdateNetworkRequest) Validate() error {
	var errs ValidationErrors
	
	if err := ValidateNetwork(r.NetworkCode); err != nil {
		errs.Add("network_code", err.Error())
	}
	
	if r.MinAmount > 0 && r.MaxAmount > 0 {
		if r.MinAmount > r.MaxAmount {
			errs.Add("amount_range", "min_amount cannot be greater than max_amount")
		}
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// PaginationRequest validation
type PaginationRequest struct {
	Page  int `json:"page" form:"page"`
	Limit int `json:"limit" form:"limit"`
}

func (r *PaginationRequest) Validate() error {
	var errs ValidationErrors
	
	// Set defaults
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Limit == 0 {
		r.Limit = 20
	}
	
	if err := ValidatePagination(r.Page, r.Limit); err != nil {
		errs.Add("pagination", err.Error())
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// DateRangeRequest validation
type DateRangeRequest struct {
	StartDate string `json:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" form:"end_date"`
}

func (r *DateRangeRequest) Validate() error {
	var errs ValidationErrors
	
	if r.StartDate == "" {
		errs.Add("start_date", "start date is required")
	}
	
	if r.EndDate == "" {
		errs.Add("end_date", "end date is required")
	}
	
	if r.StartDate != "" && r.EndDate != "" {
		if err := ValidateDateRange(r.StartDate, r.EndDate); err != nil {
			errs.Add("date_range", err.Error())
		}
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// WithdrawalRequest validation
type WithdrawalRequest struct {
	Amount        float64 `json:"amount"`
	BankCode      string  `json:"bank_code"`
	AccountNumber string  `json:"account_number"`
	AccountName   string  `json:"account_name"`
}

func (r *WithdrawalRequest) Validate() error {
	var errs ValidationErrors
	
	if r.Amount < 1000 {
		errs.Add("amount", "minimum withdrawal amount is ₦1,000")
	}
	
	if r.BankCode == "" {
		errs.Add("bank_code", "bank code is required")
	}
	
	if r.AccountNumber == "" {
		errs.Add("account_number", "account number is required")
	} else if len(r.AccountNumber) != 10 {
		errs.Add("account_number", "account number must be 10 digits")
	}
	
	if r.AccountName == "" {
		errs.Add("account_name", "account name is required")
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// Helper function to validate request body
func ValidateRequest(req interface{}) error {
	type validator interface {
		Validate() error
	}
	
	if v, ok := req.(validator); ok {
		return v.Validate()
	}
	
	return fmt.Errorf("request type does not implement Validate() method")
}

// ============================================================================
// ADDITIONAL REQUEST VALIDATORS FOR HANDLER INTEGRATION
// ============================================================================

// Auth Requests
type SendOTPRequest struct {
	MSISDN  string `json:"msisdn"`
	Purpose string `json:"purpose"` // REGISTRATION, LOGIN, PASSWORD_RESET
}

func (r *SendOTPRequest) Validate() error {
	var errs ValidationErrors
	if err := ValidateMSISDN(r.MSISDN); err != nil {
		errs.Add("msisdn", err.Error())
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

type VerifyOTPRequest struct {
	MSISDN  string `json:"msisdn"`
	OTP     string `json:"otp"`
	Purpose string `json:"purpose"` // REGISTRATION, LOGIN, PASSWORD_RESET
}

func (r *VerifyOTPRequest) Validate() error {
	var errs ValidationErrors
	if err := ValidateMSISDN(r.MSISDN); err != nil {
		errs.Add("msisdn", err.Error())
	}
	if r.OTP == "" || len(r.OTP) != 6 {
		errs.Add("otp", "OTP must be exactly 6 digits")
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *AdminLoginRequest) Validate() error {
	var errs ValidationErrors
	if err := ValidateEmail(r.Email); err != nil {
		errs.Add("email", err.Error())
	}
	if r.Password == "" {
		errs.Add("password", "Password is required")
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// Recharge Requests
type AirtimeRechargeRequest struct {
	PhoneNumber   string  `json:"phone_number"`
	Network       string  `json:"network"`
	Amount        float64 `json:"amount"`
	// AffiliateCode is the ?ref=AFFxxxx value captured by the frontend tracking hook.
	// Optional — silently ignored if empty or invalid.
	AffiliateCode string  `json:"affiliate_code"`
}

func (r *AirtimeRechargeRequest) Validate() error {
	var errs ValidationErrors
	// Normalise MSISDN to canonical international format (234...) in-place
	if normalized, err := utils.NormalizeMSISDN(r.PhoneNumber); err == nil {
		r.PhoneNumber = normalized
	}
	if err := ValidateMSISDN(r.PhoneNumber); err != nil {
		errs.Add("phone_number", err.Error())
	}
	if err := ValidateNetwork(r.Network); err != nil {
		errs.Add("network", err.Error())
	}
	if err := ValidateRechargeAmount(r.Amount); err != nil {
		errs.Add("amount", err.Error())
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

type DataRechargeRequest struct {
	PhoneNumber   string  `json:"phone_number"`
	Network       string  `json:"network"`
	BundleID      string  `json:"bundle_id"`
	Amount        float64 `json:"amount"`
	AffiliateCode string  `json:"affiliate_code"`
}

func (r *DataRechargeRequest) Validate() error {
	var errs ValidationErrors
	// Normalise MSISDN to canonical international format (234...) in-place
	if normalized, err := utils.NormalizeMSISDN(r.PhoneNumber); err == nil {
		r.PhoneNumber = normalized
	}
	if err := ValidateMSISDN(r.PhoneNumber); err != nil {
		errs.Add("phone_number", err.Error())
	}
	if err := ValidateNetwork(r.Network); err != nil {
		errs.Add("network", err.Error())
	}
	if r.BundleID == "" {
		errs.Add("bundle_id", "Bundle ID is required")
	}
	if err := ValidateRechargeAmount(r.Amount); err != nil {
		errs.Add("amount", err.Error())
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// Payment Requests
type InitiatePaymentRequest struct {
	Amount      float64                `json:"amount"`
	Email       string                 `json:"email"`
	CallbackURL string                 `json:"callback_url,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (r *InitiatePaymentRequest) Validate() error {
	var errs ValidationErrors
	if r.Amount < 100 {
		errs.Add("amount", "amount must be at least ₦100")
	} else if r.Amount > 1000000 {
		errs.Add("amount", "amount must not exceed ₦1,000,000")
	}
	if err := ValidateEmail(r.Email); err != nil {
		errs.Add("email", err.Error())
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// Subscription Requests
type SubscribeRequest struct {
	MSISDN string `json:"msisdn"`
}

func (r *SubscribeRequest) Validate() error {
	var errs ValidationErrors
	if err := ValidateMSISDN(r.MSISDN); err != nil {
		errs.Add("msisdn", err.Error())
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

type CancelSubscriptionRequest struct {
	MSISDN string `json:"msisdn"`
}

func (r *CancelSubscriptionRequest) Validate() error {
	var errs ValidationErrors
	if err := ValidateMSISDN(r.MSISDN); err != nil {
		errs.Add("msisdn", err.Error())
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// Affiliate Requests
type RegisterAffiliateRequest struct {
	MSISDN        string `json:"msisdn"`
	Email         string `json:"email,omitempty"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
}

func (r *RegisterAffiliateRequest) Validate() error {
	var errs ValidationErrors
	if err := ValidateMSISDN(r.MSISDN); err != nil {
		errs.Add("msisdn", err.Error())
	}
	if r.Email != "" {
		if err := ValidateEmail(r.Email); err != nil {
			errs.Add("email", err.Error())
		}
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

type RequestPayoutRequest struct {
	Amount        float64 `json:"amount"`
	AccountNumber string  `json:"account_number"`
	BankCode      string  `json:"bank_code"`
	AccountName   string  `json:"account_name"`
}

func (r *RequestPayoutRequest) Validate() error {
	var errs ValidationErrors
	if r.Amount < 100 {
		errs.Add("amount", "amount must be at least ₦100")
	} else if r.Amount > 1000000 {
		errs.Add("amount", "amount must not exceed ₦1,000,000")
	}
	if r.AccountNumber == "" {
		errs.Add("account_number", "Account number is required")
	}
	if r.BankCode == "" {
		errs.Add("bank_code", "Bank code is required")
	}
	if r.AccountName == "" {
		errs.Add("account_name", "Account name is required")
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// UpdatePrizeRequest validation
type UpdatePrizeRequest struct {
	Name        string  `json:"name,omitempty"`
	Type        string  `json:"type,omitempty"`
	Value       float64 `json:"value,omitempty"`
	Probability float64 `json:"probability,omitempty"`
	Description string  `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

func (r *UpdatePrizeRequest) Validate() error {
	var errs ValidationErrors
	
	// Only validate fields that are being updated
	if r.Type != "" {
		if err := ValidatePrizeType(r.Type); err != nil {
			errs.Add("type", err.Error())
		}
	}
	
	if r.Value != 0 {
		if r.Value < 0 {
			errs.Add("value", "prize value must be positive")
		}
	}
	
	if r.Probability != 0 {
		if err := ValidateProbability(r.Probability); err != nil {
			errs.Add("probability", err.Error())
		}
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// PayoutRequest validation
type PayoutRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

func (r *PayoutRequest) Validate() error {
	var errs ValidationErrors
	
	if r.Amount <= 0 {
		errs.Add("amount", "amount must be greater than 0")
	}
	
	// Minimum payout amount
	if r.Amount < 1000 {
		errs.Add("amount", "minimum payout amount is ₦1,000")
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}


// ValidatePhoneNetworkRequest validation
type ValidatePhoneNetworkRequest struct {
	PhoneNumber     string `json:"phone_number"`
	ExpectedNetwork string `json:"expected_network"`
}

func (r *ValidatePhoneNetworkRequest) Validate() error {
	var errs ValidationErrors
	// Normalise MSISDN to canonical international format (234...) in-place
	if normalized, err := utils.NormalizeMSISDN(r.PhoneNumber); err == nil {
		r.PhoneNumber = normalized
	}
	if err := ValidateMSISDN(r.PhoneNumber); err != nil {
		errs.Add("phone_number", err.Error())
	}
	
	if err := ValidateNetwork(r.ExpectedNetwork); err != nil {
		errs.Add("expected_network", err.Error())
	}
	
	if errs.HasErrors() {
		return errs
	}
	return nil
}
