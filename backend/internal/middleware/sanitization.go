package middleware

import (
	"html"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// SanitizationMiddleware sanitizes user input to prevent XSS and injection attacks
func SanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			sanitized := make([]string, len(values))
			for i, value := range values {
				sanitized[i] = SanitizeString(value)
			}
			c.Request.URL.Query()[key] = sanitized
		}

		// Note: Request body sanitization is handled by GORM's parameterized queries
		// which prevent SQL injection. Additional JSON validation happens at the
		// handler level with struct tags.

		c.Next()
	}
}

// SanitizeString removes potentially dangerous characters and HTML
func SanitizeString(input string) string {
	// HTML escape
	sanitized := html.EscapeString(input)

	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// Remove control characters except newline and tab
	sanitized = removeControlCharacters(sanitized)

	return sanitized
}

// SanitizePhoneNumber validates and sanitizes phone numbers
func SanitizePhoneNumber(phone string) string {
	// Remove all non-digit characters except +
	re := regexp.MustCompile(`[^0-9+]`)
	sanitized := re.ReplaceAllString(phone, "")

	// Remove multiple + signs, keep only first
	if strings.Count(sanitized, "+") > 1 {
		parts := strings.Split(sanitized, "+")
		sanitized = "+" + strings.Join(parts[1:], "")
	}

	return sanitized
}

// SanitizeEmail validates and sanitizes email addresses
func SanitizeEmail(email string) string {
	// Convert to lowercase
	sanitized := strings.ToLower(strings.TrimSpace(email))

	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(sanitized) {
		return "" // Invalid email
	}

	return sanitized
}

// SanitizeAlphanumeric allows only alphanumeric characters and spaces
func SanitizeAlphanumeric(input string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	return re.ReplaceAllString(input, "")
}

// removeControlCharacters removes control characters except newline and tab
func removeControlCharacters(input string) string {
	var result strings.Builder
	for _, r := range input {
		// Allow newline (10), tab (9), and printable characters (32-126)
		if r == 9 || r == 10 || (r >= 32 && r <= 126) || r > 126 {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ValidateAmount ensures amount is positive and within reasonable limits
func ValidateAmount(amount int64) bool {
	// Amount in kobo, so 50 Naira = 5000 kobo minimum
	// Maximum 10 million Naira = 1,000,000,000 kobo
	return amount >= 5000 && amount <= 1000000000
}

// ValidateTransactionReference ensures transaction reference is alphanumeric
func ValidateTransactionReference(ref string) bool {
	// Transaction references should be alphanumeric with hyphens/underscores
	re := regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	return re.MatchString(ref) && len(ref) >= 10 && len(ref) <= 100
}
