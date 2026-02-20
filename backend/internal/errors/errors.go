package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeBadRequest          ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden           ErrorCode = "FORBIDDEN"
	ErrCodeNotFound            ErrorCode = "NOT_FOUND"
	ErrCodeConflict            ErrorCode = "CONFLICT"
	ErrCodeValidation          ErrorCode = "VALIDATION_ERROR"
	ErrCodeRateLimitExceeded   ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeInsufficientBalance ErrorCode = "INSUFFICIENT_BALANCE"
	
	// Server errors (5xx)
	ErrCodeInternal           ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalAPIError   ErrorCode = "EXTERNAL_API_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	
	// Business logic errors
	ErrCodeIneligible         ErrorCode = "INELIGIBLE"
	ErrCodeAlreadyExists      ErrorCode = "ALREADY_EXISTS"
	ErrCodeExpired            ErrorCode = "EXPIRED"
	ErrCodeSuspended          ErrorCode = "SUSPENDED"
	ErrCodeFraudDetected      ErrorCode = "FRAUD_DETECTED"
	ErrCodePaymentFailed      ErrorCode = "PAYMENT_FAILED"
	ErrCodeProvisioningFailed ErrorCode = "PROVISIONING_FAILED"
)

// AppError represents a standardized application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Err        error                  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// New creates a new AppError
func New(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Predefined errors

// Client errors
func BadRequest(message string) *AppError {
	return New(ErrCodeBadRequest, message, http.StatusBadRequest)
}

func Unauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	return New(ErrCodeForbidden, message, http.StatusForbidden)
}

func NotFound(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

func Conflict(message string) *AppError {
	return New(ErrCodeConflict, message, http.StatusConflict)
}

func ValidationError(message string) *AppError {
	return New(ErrCodeValidation, message, http.StatusBadRequest)
}

func RateLimitExceeded() *AppError {
	return New(ErrCodeRateLimitExceeded, "Rate limit exceeded, please try again later", http.StatusTooManyRequests)
}

func InsufficientBalance(available, required float64) *AppError {
	return New(
		ErrCodeInsufficientBalance,
		fmt.Sprintf("Insufficient balance: available ₦%.2f, required ₦%.2f", available, required),
		http.StatusBadRequest,
	)
}

// Server errors
func Internal(message string) *AppError {
	return New(ErrCodeInternal, message, http.StatusInternalServerError)
}

func DatabaseError(err error) *AppError {
	return New(ErrCodeDatabaseError, "Database operation failed", http.StatusInternalServerError).WithError(err)
}

func ExternalAPIError(service string, err error) *AppError {
	return New(
		ErrCodeExternalAPIError,
		fmt.Sprintf("%s service unavailable", service),
		http.StatusServiceUnavailable,
	).WithError(err)
}

func ServiceUnavailable(message string) *AppError {
	return New(ErrCodeServiceUnavailable, message, http.StatusServiceUnavailable)
}

// Business logic errors
func Ineligible(reason string) *AppError {
	return New(ErrCodeIneligible, reason, http.StatusBadRequest)
}

func AlreadyExists(resource string) *AppError {
	return New(ErrCodeAlreadyExists, fmt.Sprintf("%s already exists", resource), http.StatusConflict)
}

func Expired(resource string) *AppError {
	return New(ErrCodeExpired, fmt.Sprintf("%s has expired", resource), http.StatusBadRequest)
}

func Suspended(reason string) *AppError {
	return New(ErrCodeSuspended, fmt.Sprintf("Account suspended: %s", reason), http.StatusForbidden)
}

func FraudDetected(reason string) *AppError {
	return New(ErrCodeFraudDetected, fmt.Sprintf("Fraud detected: %s", reason), http.StatusForbidden)
}

func PaymentFailed(reason string) *AppError {
	return New(ErrCodePaymentFailed, fmt.Sprintf("Payment failed: %s", reason), http.StatusBadRequest)
}

func ProvisioningFailed(reason string) *AppError {
	return New(ErrCodeProvisioningFailed, fmt.Sprintf("Provisioning failed: %s", reason), http.StatusInternalServerError)
}

// ErrorResponse is the JSON response structure for errors
type ErrorResponse struct {
	Success bool                   `json:"success"`
	Error   *ErrorDetail           `json:"error"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// ToResponse converts AppError to ErrorResponse
func (e *AppError) ToResponse() *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    e.Code,
			Message: e.Message,
		},
		Details: e.Details,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	
	appErr, ok := err.(*AppError)
	return appErr, ok
}

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	if appErr, ok := IsAppError(err); ok {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// Wrap wraps a standard error into an AppError
func Wrap(err error, code ErrorCode, message string, httpStatus int) *AppError {
	return New(code, message, httpStatus).WithError(err)
}
