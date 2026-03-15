package middleware

import (
	"context"
	"net/http"

	"rechargemax/internal/pkg/safe"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware adds a timeout to each request context
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context with timeout context
		c.Request = c.Request.WithContext(ctx)

		// Channel to signal when request is done
		done := make(chan struct{})

		safe.Go(func() {
			c.Next()
			close(done)
		})

		select {
		case <-done:
			// Request completed successfully
			return
		case <-ctx.Done():
			// Timeout occurred
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"error": "Request timeout",
				"code":  "REQUEST_TIMEOUT",
			})
		}
	}
}

// DefaultTimeoutMiddleware creates a timeout middleware with 10 second timeout
func DefaultTimeoutMiddleware() gin.HandlerFunc {
	return TimeoutMiddleware(10 * time.Second)
}

// LongTimeoutMiddleware creates a timeout middleware with 30 second timeout
// Use for endpoints that may take longer (e.g., report generation, exports)
func LongTimeoutMiddleware() gin.HandlerFunc {
	return TimeoutMiddleware(30 * time.Second)
}
