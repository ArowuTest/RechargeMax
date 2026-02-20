package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRFToken represents a CSRF token with expiration
type CSRFToken struct {
	Token     string
	ExpiresAt time.Time
}

// CSRFStore stores CSRF tokens in memory
type CSRFStore struct {
	tokens map[string]*CSRFToken
	mu     sync.RWMutex
}

var csrfStore = &CSRFStore{
	tokens: make(map[string]*CSRFToken),
}

// GenerateCSRFToken generates a new CSRF token
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CSRFMiddleware provides CSRF protection for state-changing operations
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for GET, HEAD, OPTIONS (safe methods)
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Skip CSRF for webhook endpoints (verified by signature)
		if c.Request.URL.Path == "/api/v1/payment/webhook" {
			c.Next()
			return
		}

		// Get token from header
		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token missing",
			})
			c.Abort()
			return
		}

		// Validate token
		csrfStore.mu.RLock()
		storedToken, exists := csrfStore.tokens[token]
		csrfStore.mu.RUnlock()

		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid CSRF token",
			})
			c.Abort()
			return
		}

		// Check expiration
		if time.Now().After(storedToken.ExpiresAt) {
			// Remove expired token
			csrfStore.mu.Lock()
			delete(csrfStore.tokens, token)
			csrfStore.mu.Unlock()

			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token expired",
			})
			c.Abort()
			return
		}

		// Token is valid, proceed
		c.Next()
	}
}

// GetCSRFTokenHandler returns a new CSRF token
func GetCSRFTokenHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := GenerateCSRFToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate CSRF token",
			})
			return
		}

		// Store token with 1 hour expiration
		csrfStore.mu.Lock()
		csrfStore.tokens[token] = &CSRFToken{
			Token:     token,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		csrfStore.mu.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"csrf_token": token,
			"expires_in": 3600, // 1 hour in seconds
		})
	}
}

// CleanupExpiredTokens periodically removes expired CSRF tokens
func CleanupExpiredTokens() {
	ticker := time.NewTicker(15 * time.Minute)
	go func() {
		for range ticker.C {
			csrfStore.mu.Lock()
			now := time.Now()
			for token, storedToken := range csrfStore.tokens {
				if now.After(storedToken.ExpiresAt) {
					delete(csrfStore.tokens, token)
				}
			}
			csrfStore.mu.Unlock()
		}
	}()
}
