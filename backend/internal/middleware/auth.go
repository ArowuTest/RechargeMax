package middleware

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TokenBlacklister is the minimal interface the auth middlewares need.
// Implemented by *services.TokenService; use it to avoid circular imports.
type TokenBlacklister interface {
	IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error)
}

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
func AuthMiddleware(authService interface{}, blacklister ...TokenBlacklister) gin.HandlerFunc {
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

		// SEC-003: Check token blacklist (handles logout/revocation)
		if len(blacklister) > 0 && blacklister[0] != nil {
			if blacklisted, _ := blacklister[0].IsTokenBlacklisted(c.Request.Context(), tokenString); blacklisted {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
				c.Abort()
				return
			}
		}

		c.Set("msisdn", claims.MSISDN)
		c.Set("user_id", claims.UserID)
		c.Set("authenticated", true)
		c.Next()
	}
}

// AdminAuthMiddleware validates admin authentication and sets admin_id in context
func AdminAuthMiddleware(authService interface{}, blacklister ...TokenBlacklister) gin.HandlerFunc {
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
			secret := os.Getenv("ADMIN_JWT_SECRET")
			if secret == "" {
				secret = os.Getenv("JWT_SECRET") // backward-compat fallback
			}
			return []byte(secret), nil
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
		// SEC-003: Check token blacklist — allows instant admin token revocation on logout
		if len(blacklister) > 0 && blacklister[0] != nil {
			if blacklisted, _ := blacklister[0].IsTokenBlacklisted(c.Request.Context(), tokenString); blacklisted {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Token has been revoked — please log in again",
				})
				c.Abort()
				return
			}
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
// buildAllowedOrigins returns a list of trusted origins from env + hardcoded dev origins.
func buildAllowedOrigins() []string {
	origins := []string{
		"http://localhost:3000", "http://localhost:5173",
		"http://127.0.0.1:3000", "http://127.0.0.1:5173",
	}
	if extra := os.Getenv("ALLOWED_ORIGINS"); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			if o = strings.TrimSpace(o); o != "" {
				origins = append(origins, o)
			}
		}
	}
	return origins
}

// isAllowedOrigin performs exact origin matching only.
// Suffix matching (e.g. HasSuffix) is intentionally absent: it would allow
// "https://evil-rechargemax.com" to pass when "https://rechargemax.com" is
// in the allow-list (SEC-006).
func isAllowedOrigin(origin string, allowed []string) bool {
	for _, a := range allowed {
		if origin == a {
			return true
		}
	}
	return false
}

func CORSMiddleware() gin.HandlerFunc {
	allowedOrigins := buildAllowedOrigins()
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" && isAllowedOrigin(origin, allowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
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

// SecurityHeadersMiddleware adds defensive HTTP security headers to every response.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	isProd := os.Getenv("ENVIRONMENT") == "production"
	return func(c *gin.Context) {
		// Clickjacking prevention
		c.Header("X-Frame-Options", "DENY")
		// MIME-type sniffing prevention
		c.Header("X-Content-Type-Options", "nosniff")
		// Legacy XSS filter
		c.Header("X-XSS-Protection", "1; mode=block")
		// Referrer leakage control
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// Disable unneeded browser APIs
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		// CSP: API is JSON-only — deny all framing and embedding
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		// HSTS: production only — enforce HTTPS for 1 year
		if isProd {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
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

// RequestIDMiddleware generates a unique request ID per request and attaches
// it to both the Gin context and the X-Request-ID response header.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prefer forwarded header (from load balancer) if present
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = generateRequestID()
		}
		c.Set("request_id", reqID)
		c.Header("X-Request-ID", reqID)
		c.Next()
	}
}

// generateRequestID returns a 16-byte hex request ID.
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := crand.Read(b); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

// RequestSizeLimitMiddleware rejects requests whose body exceeds maxSize bytes.
// Uses http.MaxBytesReader so the body is consumed lazily and the error is
// surfaced to the handler immediately.
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("request body too large (max %d MB)", maxSize/(1024*1024)),
			})
			c.Abort()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// RateLimitMiddleware is a per-IP sliding-window rate limiter backed by
// an in-memory store (sync.Map). For multi-instance deployments move the
// store to Redis; for now this protects a single instance.
//
// requestsPerMinute: maximum requests allowed per IP per 60-second window.
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	type entry struct {
		mu      sync.Mutex
		count   int
		resetAt time.Time
	}
	var store sync.Map

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		v, _ := store.LoadOrStore(ip, &entry{resetAt: now.Add(time.Minute)})
		e := v.(*entry)

		e.mu.Lock()
		if now.After(e.resetAt) {
			e.count = 0
			e.resetAt = now.Add(time.Minute)
		}
		e.count++
		count := e.count
		e.mu.Unlock()

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(requestsPerMinute-count, 0)))

		if count > requestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests — please slow down",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
