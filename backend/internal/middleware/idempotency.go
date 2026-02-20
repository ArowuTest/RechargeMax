package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// IdempotencyChecker provides idempotency checking for webhooks
type IdempotencyChecker struct {
	processed map[string]time.Time
	mu        sync.RWMutex
	ttl       time.Duration
}

// NewIdempotencyChecker creates a new idempotency checker
func NewIdempotencyChecker(ttl time.Duration) *IdempotencyChecker {
	checker := &IdempotencyChecker{
		processed: make(map[string]time.Time),
		ttl:       ttl,
	}
	
	// Start cleanup goroutine
	go checker.cleanup()
	
	return checker
}

// IsProcessed checks if a reference has already been processed
func (ic *IdempotencyChecker) IsProcessed(reference string) bool {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	
	processedAt, exists := ic.processed[reference]
	if !exists {
		return false
	}
	
	// Check if entry has expired
	if time.Since(processedAt) > ic.ttl {
		return false
	}
	
	return true
}

// MarkAsProcessed marks a reference as processed
func (ic *IdempotencyChecker) MarkAsProcessed(reference string) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	
	ic.processed[reference] = time.Now()
}

// cleanup removes expired entries periodically
func (ic *IdempotencyChecker) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		ic.mu.Lock()
		now := time.Now()
		for ref, processedAt := range ic.processed {
			if now.Sub(processedAt) > ic.ttl {
				delete(ic.processed, ref)
			}
		}
		ic.mu.Unlock()
	}
}

// WebhookIdempotencyMiddleware creates middleware for webhook idempotency
func WebhookIdempotencyMiddleware(checker *IdempotencyChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get reference from request (will be set by handler after parsing)
		// This middleware runs after the handler extracts the reference
		c.Next()
	}
}
