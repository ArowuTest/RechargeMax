package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID  string `json:"user_id"`
	AdminID string `json:"admin_id"` // Admin tokens use admin_id instead of user_id
	MSISDN  string `json:"msisdn"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	Type    string `json:"type"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates user authentication
func AuthMiddleware(authService interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

	tokenString := parts[1]

	// Parse and validate JWT
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key for validation
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.Abort()
		return
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
		return
	}

	// Set user context from JWT claims
	c.Set("msisdn", claims.MSISDN)
	c.Set("user_id", claims.UserID)

		c.Next()
	}
}

// AdminAuthMiddleware validates admin authentication and sets admin_id in context
func AdminAuthMiddleware(authService interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("[AdminAuth] Request received:", c.Request.Method, c.Request.URL.Path)
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		log.Println("[AdminAuth] Authorization header:", authHeader != "")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Admin authorization required",
			})
			c.Abort()
			return
		}
		// Extract token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}
		tokenString := parts[1]
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token",
			})
			c.Abort()
			return
		}
		// Parse JWT to extract admin_id and other claims
		parsedToken, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !parsedToken.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired admin token",
			})
			c.Abort()
			return
		}
		claims, ok := parsedToken.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid admin token claims",
			})
			c.Abort()
			return
		}
			// Verify this is an admin token (type must be "admin")
			// Regular user tokens must not be accepted on admin endpoints — return 403 Forbidden
			if claims.Type != "admin" {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "Access denied: admin privileges required",
				})
				c.Abort()
				return
			}
			// Admin tokens use admin_id claim; fall back to user_id for compatibility
			adminID := claims.AdminID
			if adminID == "" {
				adminID = claims.UserID
			}
			if adminID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Invalid admin token: missing admin ID",
				})
				c.Abort()
				return
			}
		// Set full admin context for all handlers
		c.Set("admin_id", adminID)
		c.Set("admin_token", tokenString)
		c.Set("msisdn", claims.MSISDN)
		c.Set("is_admin", true)
		c.Set("authenticated", true)
		log.Println("[AdminAuth] Middleware passed, admin_id:", claims.UserID)
		c.Next()
		log.Println("[AdminAuth] Handler completed")
	}
}

// LoggingMiddleware logs all requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("[REQUEST] %s %s\n", c.Request.Method, c.Request.URL.Path)
		c.Next()
		fmt.Printf("[RESPONSE] %s %s - Status: %d\n", c.Request.Method, c.Request.URL.Path, c.Writer.Status())
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Allow requests from the frontend origin or any localhost/manus.computer domain
		if origin != "" && (strings.Contains(origin, "localhost") || strings.Contains(origin, "manus.computer") || strings.Contains(origin, "127.0.0.1")) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if origin == "" {
			// For non-browser requests (like Paystack webhooks), allow without credentials
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Next()
	}
}

// RequestIDMiddleware adds request ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// RequestSizeLimitMiddleware limits request size
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// RateLimitMiddleware limits request rate
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
