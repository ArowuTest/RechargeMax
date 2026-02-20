package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"rechargemax/internal/utils"
)

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var messages []string
	for _, err := range v {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// Add adds a validation error
func (v *ValidationErrors) Add(field, message string) {
	*v = append(*v, ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are validation errors
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

// Nigerian phone number regex (supports various formats)
var nigerianPhoneRegex = regexp.MustCompile(`^(\+?234|0)[7-9][0-1]\d{8}$`)

// Email regex (basic validation)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateMSISDN validates Nigerian phone numbers
func ValidateMSISDN(msisdn string) error {
	if msisdn == "" {
		return fmt.Errorf("phone number is required")
	}
	
	// Remove spaces and dashes
	cleaned := strings.ReplaceAll(strings.ReplaceAll(msisdn, " ", ""), "-", "")
	
	if !nigerianPhoneRegex.MatchString(cleaned) {
		return fmt.Errorf("invalid Nigerian phone number format")
	}
	
	return nil
}

// NormalizeMSISDN normalizes Nigerian phone numbers to international format
// Uses the centralized utils.NormalizeMSISDN for consistency
func NormalizeMSISDN(msisdn string) (string, error) {
	return utils.NormalizeMSISDN(msisdn)
}

// ValidateEmail validates email addresses
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	
	return nil
}

// ValidateUUID validates UUID strings
func ValidateUUID(id string) error {
	if id == "" {
		return fmt.Errorf("ID is required")
	}
	
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format")
	}
	
	return nil
}

// ValidateAmount validates monetary amounts (in kobo)
func ValidateAmount(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	
	return nil
}

// ValidateDate validates date strings and ensures they're not in the past
func ValidateDate(date time.Time, allowPast bool) error {
	if date.IsZero() {
		return fmt.Errorf("date is required")
	}
	
	if !allowPast && date.Before(time.Now()) {
		return fmt.Errorf("date cannot be in the past")
	}
	
	return nil
}

// ValidateString validates string fields
func ValidateString(value, fieldName string, minLength, maxLength int) error {
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	
	if len(value) < minLength {
		return fmt.Errorf("%s must be at least %d characters", fieldName, minLength)
	}
	
	if maxLength > 0 && len(value) > maxLength {
		return fmt.Errorf("%s must not exceed %d characters", fieldName, maxLength)
	}
	
	return nil
}

// ValidateInt validates integer fields
func ValidateInt(value int, fieldName string, min, max int) error {
	if value < min {
		return fmt.Errorf("%s must be at least %d", fieldName, min)
	}
	
	if max > 0 && value > max {
		return fmt.Errorf("%s must not exceed %d", fieldName, max)
	}
	
	return nil
}


// ValidateRechargeAmount validates recharge amounts
func ValidateRechargeAmount(amount float64) error {
	if amount < 100 { // Minimum ₦100
		return fmt.Errorf("recharge amount must be at least ₦100")
	}
	
	if amount > 100000 { // Maximum ₦100,000
		return fmt.Errorf("recharge amount must not exceed ₦100,000")
	}
	
	return nil
}

// ValidateNetwork validates network names
func ValidateNetwork(network string) error {
	validNetworks := map[string]bool{
		"MTN":     true,
		"GLO":     true,
		"AIRTEL":  true,
		"9MOBILE": true,
	}
	
	if !validNetworks[strings.ToUpper(network)] {
		return fmt.Errorf("invalid network: must be MTN, GLO, AIRTEL, or 9MOBILE")
	}
	
	return nil
}

// ValidateTransactionType validates transaction types
func ValidateTransactionType(txType string) error {
	validTypes := map[string]bool{
		"airtime": true,
		"data":    true,
	}
	
	if !validTypes[strings.ToLower(txType)] {
		return fmt.Errorf("invalid transaction type: must be airtime or data")
	}
	
	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	
	if len(password) > 100 {
		return fmt.Errorf("password must not exceed 100 characters")
	}
	
	// Check for at least one number
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	
	// Check for at least one letter
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}
	
	return nil
}

// ValidateReferralCode validates referral codes
func ValidateReferralCode(code string) error {
	if code == "" {
		return nil // Referral code is optional
	}
	
	if len(code) < 6 || len(code) > 20 {
		return fmt.Errorf("referral code must be between 6 and 20 characters")
	}
	
	// Only alphanumeric characters allowed
	validCode := regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(code)
	if !validCode {
		return fmt.Errorf("referral code must contain only letters and numbers")
	}
	
	return nil
}

// ValidatePrizeType validates prize types
func ValidatePrizeType(prizeType string) error {
	validTypes := map[string]bool{
		"airtime":  true,
		"data":     true,
		"cash":     true,
		"points":   true,
		"physical": true,
	}
	
	if !validTypes[strings.ToLower(prizeType)] {
		return fmt.Errorf("invalid prize type")
	}
	
	return nil
}

// ValidateProbability validates probability values (0-100)
func ValidateProbability(probability float64) error {
	if probability < 0 || probability > 100 {
		return fmt.Errorf("probability must be between 0 and 100")
	}
	
	return nil
}

// ValidateCommissionRate validates commission rates (0-100)
func ValidateCommissionRate(rate float64) error {
	if rate < 0 || rate > 100 {
		return fmt.Errorf("commission rate must be between 0 and 100")
	}
	
	return nil
}


// ValidateDateRange validates date range inputs
func ValidateDateRange(startDate, endDate string) error {
	if startDate == "" || endDate == "" {
		return nil // Optional date range
	}
	
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("invalid start date format, use YYYY-MM-DD")
	}
	
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("invalid end date format, use YYYY-MM-DD")
	}
	
	if end.Before(start) {
		return fmt.Errorf("end date must be after start date")
	}
	
	return nil
}

// ValidatePagination validates pagination parameters
func ValidatePagination(page, limit int) error {
	if page < 1 {
		return fmt.Errorf("page must be at least 1")
	}
	
	if limit < 1 {
		return fmt.Errorf("limit must be at least 1")
	}
	
	if limit > 100 {
		return fmt.Errorf("limit must not exceed 100")
	}
	
	return nil
}


// ValidateStatus validates status values
func ValidateStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":   true,
		"active":    true,
		"completed": true,
		"cancelled": true,
		"failed":    true,
		"claimed":   true,
		"unclaimed": true,
	}
	
	if !validStatuses[strings.ToLower(status)] {
		return fmt.Errorf("invalid status")
	}
	
	return nil
}


// ValidatePhoneNetworkMatch validates that a phone number belongs to the specified network
func ValidatePhoneNetworkMatch(msisdn string, network string) error {
	if msisdn == "" || network == "" {
		return nil // Skip validation if either is empty (will be caught by other validators)
	}
	
	isValid, err := utils.ValidatePhoneNetwork(msisdn, network)
	if err != nil {
		return fmt.Errorf("unable to validate network: %v", err)
	}
	
	if !isValid {
		detectedNetwork, _ := utils.DetectNetwork(msisdn)
		return fmt.Errorf("phone number does not belong to %s network (detected: %s)", network, detectedNetwork)
	}
	
	return nil
}
