package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DBLoggingMiddleware writes a row to application_logs for every request.
// Inserts are fire-and-forget in a goroutine — zero impact on request latency.
// Only INFO-level rows are written for 2xx/3xx responses; 4xx → WARN; 5xx → ERROR.
// Health-check noise (/health, /api/v1/health) is suppressed to keep the table clean.
func DBLoggingMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path  := c.Request.URL.Path

		// Skip noisy health-check polling
		if path == "/health" || path == "/api/v1/health" {
			c.Next()
			return
		}

		c.Next()

		status  := c.Writer.Status()
		latency := time.Since(start)
		ip      := c.ClientIP()
		reqID   := c.GetString("request_id")
		method  := c.Request.Method
		ua      := c.Request.UserAgent()

		level := "INFO"
		switch {
		case status >= http.StatusInternalServerError:
			level = "ERROR"
		case status >= http.StatusBadRequest:
			level = "WARN"
		}

		ctx := map[string]interface{}{
			"method":     method,
			"path":       path,
			"status":     status,
			"latency_ms": latency.Milliseconds(),
		}
		ctxJSON, _ := json.Marshal(ctx)

		msg := method + " " + path

		// Async insert — never block the HTTP response
		go func() {
			db.Exec(
				`INSERT INTO application_logs
				 (level, message, context, ip_address, user_agent, request_id, created_at)
				 VALUES (?, ?, ?, ?, ?, ?, NOW())`,
				level,
				msg,
				string(ctxJSON),
				ip,
				ua,
				reqID,
			)
		}()
	}
}
