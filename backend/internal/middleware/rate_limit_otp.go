package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// OTPRateLimiter implements a per-MSISDN sliding-window rate limit.
// Default: max 5 requests per 60-second window.
type OTPRateLimiter struct {
	mu      sync.Mutex
	windows map[string][]time.Time
	maxReqs int
	window  time.Duration
}

// NewOTPRateLimiter creates a new OTPRateLimiter with the given parameters.
func NewOTPRateLimiter(maxReqs int, window time.Duration) *OTPRateLimiter {
	rl := &OTPRateLimiter{
		windows: make(map[string][]time.Time),
		maxReqs: maxReqs,
		window:  window,
	}
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *OTPRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-rl.window)
	for key, ts := range rl.windows {
		valid := ts[:0]
		for _, t := range ts {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.windows, key)
		} else {
			rl.windows[key] = valid
		}
	}
}

// Allow returns true if the key is within the rate limit.
func (rl *OTPRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-rl.window)
	ts := rl.windows[key]
	valid := ts[:0]
	for _, t := range ts {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	if len(valid) >= rl.maxReqs {
		rl.windows[key] = valid
		return false
	}
	rl.windows[key] = append(valid, now)
	return true
}

// OTPRateLimitMiddleware reads the MSISDN from the request body and rate-limits.
// It falls back to IP address when MSISDN is not available.
func OTPRateLimitMiddleware(rl *OTPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		var body struct {
			MSISDN string `json:"msisdn"`
		}
		_ = json.Unmarshal(bodyBytes, &body)
		key := body.MSISDN
		if key == "" {
			key = c.ClientIP()
		}
		if !rl.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Too many OTP requests. Please wait before trying again.",
				"error":   "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
