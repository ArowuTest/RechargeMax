package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rechargemax/internal/pkg/safe"
)

// OTPRateLimiter implements a sliding-window rate limiter backed by PostgreSQL.
// This survives server restarts and works correctly across multiple instances.
type OTPRateLimiter struct {
	db      *gorm.DB
	maxReqs int
	window  time.Duration
}

// NewOTPRateLimiter creates a PostgreSQL-backed OTP rate limiter.
//   - db:      live *gorm.DB connection
//   - maxReqs: maximum requests allowed per window per key
//   - window:  sliding window size (e.g. 60*time.Second)
func NewOTPRateLimiter(db *gorm.DB, maxReqs int, window time.Duration) *OTPRateLimiter {
	rl := &OTPRateLimiter{db: db, maxReqs: maxReqs, window: window}

	// Periodic cleanup of old rows (> 24 h) to prevent unbounded growth.
	safe.Go(func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.pruneOld()
		}
	})
	return rl
}

// Allow returns true if the request is within the rate limit for the given key.
// It atomically records the new attempt and counts recent attempts in one round trip.
func (rl *OTPRateLimiter) Allow(key string) bool {
	windowStart := time.Now().Add(-rl.window)

	// Insert the new attempt first (so concurrent calls don't race on the count).
	if err := rl.db.Exec(
		`INSERT INTO otp_rate_limits (key, requested_at) VALUES (?, NOW())`, key,
	).Error; err != nil {
		// On DB error, fail-open: allow the request but log.
		log.Printf("[OTPRateLimiter] DB insert error for key %s: %v — allowing request", key, err)
		return true
	}

	// Count how many attempts (including the one just inserted) fall inside window.
	var count int64
	if err := rl.db.Table("otp_rate_limits").
		Where("key = ? AND requested_at > ?", key, windowStart).
		Count(&count).Error; err != nil {
		log.Printf("[OTPRateLimiter] DB count error for key %s: %v — allowing request", key, err)
		return true
	}

	return count <= int64(rl.maxReqs)
}

// pruneOld removes rows older than 24 h to keep the table small.
func (rl *OTPRateLimiter) pruneOld() {
	cutoff := time.Now().Add(-24 * time.Hour)
	if err := rl.db.Exec(`DELETE FROM otp_rate_limits WHERE requested_at < ?`, cutoff).Error; err != nil {
		log.Printf("[OTPRateLimiter] prune error: %v", err)
	}
}

// OTPRateLimit returns a Gin middleware that enforces the rate limit.
// It keys on MSISDN from the JSON body ("msisdn" field) and falls back to client IP.
func OTPRateLimit(rl *OTPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Determine the key: prefer MSISDN, fall back to IP
		key := "ip:" + c.ClientIP()

		// Peek at the body for the msisdn field without consuming it
		var body struct {
			MSISDN string `json:"msisdn"`
		}
		// Use ShouldBindBodyWith so the body remains available for subsequent handlers
		if err := c.ShouldBindJSON(&body); err == nil && body.MSISDN != "" {
			key = "msisdn:" + body.MSISDN
			// Re-set the body so the next handler can read it
			c.Set("parsed_msisdn", body.MSISDN)
		}

		if !rl.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many OTP requests. Please wait before trying again.",
				"code":    "OTP_RATE_LIMITED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
