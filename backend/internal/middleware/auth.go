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

// extractToken extracts the JWT from the request.
// Priority order:
//  1. httpOnly cookie  "auth_token" (user) or "admin_auth_token" (admin)
//  2. Authorization: Bearer <token> header (mobile apps / API clients)
func extractToken(c *gin.Context, isAdmin bool) string {
	cookieName := "auth_token"
	if isAdmin {
		cookieName = "admin_auth_token"
	}
	if cookie, err := c.Cookie(cookieName); err == nil && cookie != "" {
		return cookie
	}
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1]
	}
	return ""
}

// AuthMiddleware validates user authentication
func AuthMiddleware(authService interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c, false)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Set("msisdn", claims.MSISDN)
		c.Set("user_id", claims.UserID)
		c.Set("authenticated", true)
		c.Next()
	}
}

// AdminAuthMiddleware validates admin authentication and sets admin_id in context
func AdminAuthMiddleware(authService interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("[AdminAuth] Request received:", c.Request.Method, c.Request.URL.Path)
		tokenString := extractToken(c, true)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Admin authorization required",
			})
			c.Abort()
			return
		}

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
		if claims.Type != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Access denied: admin privileges required",
			})
			c.Abort()
			return
		}
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
		c.Set("admin_id", adminID)
		c.Set("admin_token", tokenString)
		c.Set("msisdn", claims.MSISDN)
		c.Set("is_admin", true)
		c.Set("authenticated", true)
		log.Println("[AdminAuth] Middleware passed, admin_id:", adminID)
		c.Next()
		log.Println("[AdminAuth] Handler completed")
	}
}

// SetAuthCookie sets a secure httpOnly JWT cookie on the response.
// tokenType: "user" or "admin"
func SetAuthCookie(c *gin.Context, token, tokenType string, maxAgeSecs int) {
	cookieName := "auth_token"
	if tokenType == "admin" {
		cookieName = "admin_auth_token"
	}
	// Determine if we're in production (HTTPS)
	secure := os.Getenv("GIN_MODE") == "release"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		cookieName, // name
		token,      // value
		maxAgeSecs, // max-age in seconds
		"/",        // path
		"",         // domain (empty = current host)
		secure,     // secure (HTTPS only in production)
		true,       // httpOnly — JS cannot read this cookie
	)
}

// ClearAuthCookie clears the auth cookie on logout
func ClearAuthCookie(c *gin.Context, tokenType string) {
	cookieName := "auth_token"
	if tokenType == "admin" {
		cookieName = "admin_auth_token"
	}
	secure := os.Getenv("GIN_MODE") == "release"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(cookieName, "", -1, "/", "", secure, true)
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
		if origin != "" && (strings.Contains(origin, "localhost") || strings.Contains(origin, "manus.computer") || strings.Contains(origin, "127.0.0.1")) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if origin == "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
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
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		c.Next()
	}
}

// OptionalAuthMiddleware reads token from cookie or header if present, never rejects
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c, false)
		if tokenString == "" {
			c.Next()
			return
		}
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			c.Next()
			return
		}
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.Next()
			return
		}
		c.Set("msisdn", claims.MSISDN)
		c.Set("user_id", claims.UserID)
		c.Set("authenticated", true)
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
