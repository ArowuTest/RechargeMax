package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rechargemax/internal/application/services"
)

// AdminAuthMiddlewareWithBlacklist creates authentication middleware for admin users with token blacklist support
func AdminAuthMiddlewareWithBlacklist(authService *services.AuthService, tokenService *services.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Admin authorization required",
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization header format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		ctx := context.Background()

		// Check if token is blacklisted
		isBlacklisted, err := tokenService.IsTokenBlacklisted(ctx, token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to verify token status",
			})
			c.Abort()
			return
		}

		if isBlacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token has been revoked. Please login again",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := authService.ValidateToken(ctx, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired admin token",
			})
			c.Abort()
			return
		}

		// Admin tokens must have type "admin"
		if claims.Type != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Admin access required. Regular user tokens cannot access admin endpoints",
			})
			c.Abort()
			return
		}

		// Verify admin exists and is active
		if claims.UserID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid admin token: missing admin ID",
			})
			c.Abort()
			return
		}

		// Set admin context for handlers
		c.Set("admin_id", claims.UserID)
		c.Set("msisdn", claims.MSISDN)
		c.Set("authenticated", true)
		c.Set("is_admin", true)
		c.Set("token", token) // Store token for potential blacklisting

		c.Next()
	}
}
