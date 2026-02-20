package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// OTPAttemptLimiter tracks OTP verification attempts to prevent brute force
type OTPAttemptLimiter struct {
	attempts map[string]*OTPAttempts
	mu       sync.RWMutex
}

// OTPAttempts tracks attempts for a single phone number
type OTPAttempts struct {
	count      int
	firstAttempt time.Time
	lockedUntil  *time.Time
}

var (
	otpLimiter     *OTPAttemptLimiter
	otpLimiterOnce sync.Once
)

// GetOTPAttemptLimiter returns the singleton OTP attempt limiter
func GetOTPAttemptLimiter() *OTPAttemptLimiter {
	otpLimiterOnce.Do(func() {
		otpLimiter = &OTPAttemptLimiter{
			attempts: make(map[string]*OTPAttempts),
		}
		go otpLimiter.cleanup()
	})
	return otpLimiter
}

// CheckAndRecordAttempt checks if an OTP verification attempt is allowed
// Returns error if attempt should be blocked
func (l *OTPAttemptLimiter) CheckAndRecordAttempt(msisdn string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	now := time.Now()
	
	attempts, exists := l.attempts[msisdn]
	if !exists {
		// First attempt
		l.attempts[msisdn] = &OTPAttempts{
			count:        1,
			firstAttempt: now,
			lockedUntil:  nil,
		}
		return nil
	}
	
	// Check if locked
	if attempts.lockedUntil != nil && now.Before(*attempts.lockedUntil) {
		remaining := attempts.lockedUntil.Sub(now)
		return fmt.Errorf("too many failed attempts. Please try again in %d minutes", int(remaining.Minutes())+1)
	}
	
	// Reset if window expired (5 minutes)
	if now.Sub(attempts.firstAttempt) > 5*time.Minute {
		attempts.count = 1
		attempts.firstAttempt = now
		attempts.lockedUntil = nil
		return nil
	}
	
	// Increment attempt count
	attempts.count++
	
	// Lock after 5 failed attempts
	if attempts.count >= 5 {
		lockUntil := now.Add(15 * time.Minute) // Lock for 15 minutes
		attempts.lockedUntil = &lockUntil
		return fmt.Errorf("too many failed OTP attempts. Account locked for 15 minutes")
	}
	
	// Warn after 3 attempts
	if attempts.count >= 3 {
		remaining := 5 - attempts.count
		return fmt.Errorf("invalid OTP. %d attempts remaining before account lock", remaining)
	}
	
	return nil
}

// RecordSuccess resets the attempt counter on successful verification
func (l *OTPAttemptLimiter) RecordSuccess(msisdn string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	delete(l.attempts, msisdn)
}

// cleanup removes stale entries
func (l *OTPAttemptLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for msisdn, attempts := range l.attempts {
			// Remove if lock expired and no recent attempts
			if attempts.lockedUntil != nil && now.After(*attempts.lockedUntil) {
				if now.Sub(attempts.firstAttempt) > 30*time.Minute {
					delete(l.attempts, msisdn)
				}
			} else if now.Sub(attempts.firstAttempt) > 30*time.Minute {
				delete(l.attempts, msisdn)
			}
		}
		l.mu.Unlock()
	}
}

// OTPVerificationMiddleware enforces OTP attempt limits
func OTPVerificationMiddleware() gin.HandlerFunc {
	limiter := GetOTPAttemptLimiter()
	
	return func(c *gin.Context) {
		var requestBody struct {
			MSISDN string `json:"msisdn"`
			Code   string `json:"code"`
		}
		
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request",
			})
			c.Abort()
			return
		}
		
		// Check attempt limit
		if err := limiter.CheckAndRecordAttempt(requestBody.MSISDN); err != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": err.Error(),
				"error":   "otp_attempts_exceeded",
			})
			c.Abort()
			return
		}
		
		// Store MSISDN in context for success handler
		c.Set("otp_msisdn", requestBody.MSISDN)
		
		c.Next()
		
		// If verification was successful (status 200), reset attempts
		if c.Writer.Status() == http.StatusOK {
			limiter.RecordSuccess(requestBody.MSISDN)
		}
	}
}
