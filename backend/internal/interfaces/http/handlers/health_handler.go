package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Checks    map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents a single health check
type HealthCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// BasicHealthCheck returns a simple health check (for load balancers)
func (h *HealthHandler) BasicHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// DetailedHealthCheck returns detailed health information
func (h *HealthHandler) DetailedHealthCheck(c *gin.Context) {
	checks := make(map[string]HealthCheck)
	overallStatus := "healthy"
	
	// Check database connection
	dbCheck := h.checkDatabase()
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		overallStatus = "unhealthy"
	}
	
	// Check database write capability
	dbWriteCheck := h.checkDatabaseWrite()
	checks["database_write"] = dbWriteCheck
	if dbWriteCheck.Status != "healthy" {
		overallStatus = "degraded"
	}
	
	response := HealthCheckResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0", // TODO: Get from config or build info
		Checks:    checks,
	}
	
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if overallStatus == "degraded" {
		statusCode = http.StatusOK // Still return 200 for degraded
	}
	
	c.JSON(statusCode, response)
}

// checkDatabase checks database connectivity
func (h *HealthHandler) checkDatabase() HealthCheck {
	start := time.Now()
	
	sqlDB, err := h.db.DB()
	if err != nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: "Failed to get database connection: " + err.Error(),
		}
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: "Database ping failed: " + err.Error(),
		}
	}
	
	latency := time.Since(start)
	
	return HealthCheck{
		Status:  "healthy",
		Message: "Database connection is healthy",
		Latency: latency.String(),
	}
}

// checkDatabaseWrite checks database write capability
func (h *HealthHandler) checkDatabaseWrite() HealthCheck {
	start := time.Now()
	
	// Try to execute a simple write query (to a health check table if it exists)
	// For now, just check if we can start a transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: "Cannot start database transaction: " + tx.Error.Error(),
		}
	}
	
	// Rollback immediately (we don't actually want to write anything)
	tx.Rollback()
	
	latency := time.Since(start)
	
	return HealthCheck{
		Status:  "healthy",
		Message: "Database write capability is healthy",
		Latency: latency.String(),
	}
}

// ReadinessCheck checks if the service is ready to accept traffic
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Check if database is accessible
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"ready": false,
			"reason": "database connection unavailable",
		})
		return
	}
	
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"ready": false,
			"reason": "database not responding",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"ready": true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// LivenessCheck checks if the service is alive (for Kubernetes)
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"alive": true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
