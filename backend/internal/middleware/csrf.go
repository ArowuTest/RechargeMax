package middleware

// csrf.go — PostgreSQL-backed CSRF token store (INFRA-001)
//
// Architecture:
//  * CSRFStore is an interface so the in-memory store can be swapped for
//    the DB-backed store at startup without changing middleware call sites.
//  * InitCSRF(db) must be called from main.go before any request arrives.
//  * GET, HEAD, OPTIONS, and the payment webhook bypass CSRF validation.
//  * Tokens are single-use: they are deleted from the store after one successful
//    validation, preventing replay across multiple requests.

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

// CSRFStorer is the interface both the in-memory and DB-backed stores satisfy.
type CSRFStorer interface {
	// Store persists a token with the given TTL.
	Store(token string, ttl time.Duration) error
	// Validate checks whether the token exists and has not expired.
	// It deletes the token on success (single-use).
	Validate(token string) error
}

// ---------------------------------------------------------------------------
// Global store (set by InitCSRF; defaults to in-memory for safety)
// ---------------------------------------------------------------------------

var activeCSRFStore CSRFStorer = newInMemoryCSRFStore()

// InitCSRF swaps the global store for a PostgreSQL-backed one.
// Call this once from main.go after the DB connection is established.
func InitCSRF(db *gorm.DB) {
	if db == nil {
		log.Println("[csrf] WARNING: nil db passed to InitCSRF — falling back to in-memory store")
		return
	}
	activeCSRFStore = &postgresCSRFStore{db: db}
	log.Println("[csrf] PostgreSQL-backed CSRF store activated")
}

// ---------------------------------------------------------------------------
// PostgreSQL-backed store
// ---------------------------------------------------------------------------

type postgresCSRFStore struct {
	db *gorm.DB
}

func (s *postgresCSRFStore) Store(token string, ttl time.Duration) error {
	return s.db.Exec(
		`INSERT INTO csrf_tokens (token, expires_at) VALUES ($1, $2)
		 ON CONFLICT (token) DO UPDATE SET expires_at = EXCLUDED.expires_at`,
		token, time.Now().Add(ttl),
	).Error
}

func (s *postgresCSRFStore) Validate(token string) error {
	res := s.db.Exec(
		`DELETE FROM csrf_tokens WHERE token = $1 AND expires_at > NOW()`,
		token,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("token not found or expired")
	}
	return nil
}

// ---------------------------------------------------------------------------
// In-memory fallback store (used before InitCSRF is called)
// ---------------------------------------------------------------------------

type inMemoryCSRFStore struct {
	tokens map[string]time.Time
	ch     chan func()
}

func newInMemoryCSRFStore() *inMemoryCSRFStore {
	s := &inMemoryCSRFStore{
		tokens: make(map[string]time.Time),
		ch:     make(chan func(), 256),
	}
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		for {
			select {
			case fn := <-s.ch:
				fn()
			case <-ticker.C:
				now := time.Now()
				for tok, exp := range s.tokens {
					if now.After(exp) {
						delete(s.tokens, tok)
					}
				}
			}
		}
	}()
	return s
}

func (s *inMemoryCSRFStore) Store(token string, ttl time.Duration) error {
	done := make(chan struct{})
	s.ch <- func() {
		s.tokens[token] = time.Now().Add(ttl)
		close(done)
	}
	<-done
	return nil
}

func (s *inMemoryCSRFStore) Validate(token string) error {
	result := make(chan error, 1)
	s.ch <- func() {
		exp, ok := s.tokens[token]
		if !ok || time.Now().After(exp) {
			delete(s.tokens, token)
			result <- fmt.Errorf("token not found or expired")
			return
		}
		delete(s.tokens, token) // single-use
		result <- nil
	}
	return <-result
}

// ---------------------------------------------------------------------------
// Public helpers
// ---------------------------------------------------------------------------

// GenerateCSRFToken generates a 32-byte URL-safe random token.
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ---------------------------------------------------------------------------
// Gin middleware
// ---------------------------------------------------------------------------

// CSRFMiddleware validates the X-CSRF-Token header for state-changing methods.
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Safe methods + webhook bypass
		switch c.Request.Method {
		case "GET", "HEAD", "OPTIONS":
			c.Next()
			return
		}
		if c.Request.URL.Path == "/api/v1/payment/webhook" {
			c.Next()
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token missing"})
			c.Abort()
			return
		}

		if err := activeCSRFStore.Validate(token); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or expired CSRF token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCSRFTokenHandler issues a new CSRF token to the caller.
func GetCSRFTokenHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := GenerateCSRFToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSRF token"})
			return
		}
		const ttl = 24 * time.Hour // extend to 24h so tokens survive normal browser sessions
		if err := activeCSRFStore.Store(token, ttl); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store CSRF token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"csrf_token": token,
			"expires_in": int(ttl.Seconds()),
		})
	}
}

// CleanupExpiredTokens is a no-op for the DB store (handled by DB TTL query).
// Kept for backward compat with code that calls it from main.go.
func CleanupExpiredTokens() {
	// DB store: the DELETE WHERE expires_at > NOW() in Validate handles cleanup.
	// In-memory store: handled by the background goroutine in newInMemoryCSRFStore.
}
