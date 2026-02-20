package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"rechargemax/internal/errors"
	"rechargemax/internal/validation"
)

// ErrorHandler is a middleware that handles errors and panics
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		startTime := time.Now()
		
		// Recover from panics
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				errors.Error("Panic recovered", nil, map[string]interface{}{
					"error": err,
					"path":  c.Request.URL.Path,
				})
				
				// Return internal server error
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error": map[string]interface{}{
						"code":    errors.ErrCodeInternal,
						"message": "Internal server error",
					},
				})
				
				c.Abort()
			}
		}()
		
		// Process request
		c.Next()
		
		// Log request
		duration := time.Since(startTime)
		errors.LogRequest(c.Request.Context(), c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
		
		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last().Err
			
			// Handle different error types
			if appErr, ok := errors.IsAppError(err); ok {
				// AppError - use structured response
				c.JSON(appErr.HTTPStatus, appErr.ToResponse())
				return
			}
			
			// Validation errors
			if valErrs, ok := err.(validation.ValidationErrors); ok {
				errors.LogValidationError(c.Request.URL.Path, valErrs)
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error": map[string]interface{}{
						"code":    errors.ErrCodeValidation,
						"message": "Validation failed",
					},
					"details": map[string]interface{}{
						"errors": valErrs,
					},
				})
				return
			}
			
			// Generic error
			errors.Error("Unhandled error", err, map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": map[string]interface{}{
					"code":    errors.ErrCodeInternal,
					"message": "Internal server error",
				},
			})
		}
	}
}

// RespondWithError sends an error response
func RespondWithError(c *gin.Context, err error) {
	if appErr, ok := errors.IsAppError(err); ok {
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}
	
	// Generic error
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": map[string]interface{}{
			"code":    errors.ErrCodeInternal,
			"message": "Internal server error",
		},
	})
}

// RespondWithSuccess sends a success response
func RespondWithSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// RespondWithValidationError sends a validation error response
func RespondWithValidationError(c *gin.Context, validationErrors interface{}) {
	errors.LogValidationError(c.Request.URL.Path, validationErrors)
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error": map[string]interface{}{
			"code":    errors.ErrCodeValidation,
			"message": "Validation failed",
		},
		"details": map[string]interface{}{
			"errors": validationErrors,
		},
	})
}
