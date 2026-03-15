package middleware

import (
	"bytes"
	"encoding/json"

	"rechargemax/internal/pkg/safe"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type auditLogEntry struct {
	ID          uuid.UUID  `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	AdminUserID *uuid.UUID `gorm:"column:admin_user_id"`
	UserID      *uuid.UUID `gorm:"column:user_id"`
	Action      string     `gorm:"column:action"`
	EntityType  string     `gorm:"column:entity_type"`
	EntityID    string     `gorm:"column:entity_id"`
	OldValue    []byte     `gorm:"column:old_value;type:jsonb"`
	NewValue    []byte     `gorm:"column:new_value;type:jsonb"`
	IPAddress   string     `gorm:"column:ip_address"`
	UserAgent   string     `gorm:"column:user_agent"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (auditLogEntry) TableName() string { return "audit_logs" }

func AdminAuditMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		c.Next()

		action := deriveAction(c.Request.Method, c.FullPath())
		var adminUID *uuid.UUID
		if raw, ok := c.Get("admin_id"); ok {
			if s, ok2 := raw.(string); ok2 {
				if id, err := uuid.Parse(s); err == nil {
					adminUID = &id
				}
			}
		}
		entry := auditLogEntry{
			Action:      action,
			EntityType:  entityTypeFromPath(c.FullPath()),
			EntityID:    c.Param("id"),
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			AdminUserID: adminUID,
		}
		if len(bodyBytes) > 0 {
			if len(bodyBytes) > 32768 {
				bodyBytes = bodyBytes[:32768]
			}
			if json.Valid(bodyBytes) {
				entry.NewValue = bodyBytes
			}
		}
		safe.Go(func() { db.Create(&entry) })
	}
}

func AdminAuditLogsList(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var logs []auditLogEntry
		limit, offset := 50, 0
		if v := c.Query("limit"); v != "" {
			fmt.Sscanf(v, "%d", &limit)
		}
		if v := c.Query("offset"); v != "" {
			fmt.Sscanf(v, "%d", &offset)
		}
		if limit > 200 {
			limit = 200
		}
		query := db.Order("created_at DESC").Limit(limit).Offset(offset)
		if action := c.Query("action"); action != "" {
			query = query.Where("action = ?", action)
		}
		if adminID := c.Query("admin_id"); adminID != "" {
			query = query.Where("admin_user_id::text = ?", adminID)
		}
		var total int64
		db.Model(&auditLogEntry{}).Count(&total)
		query.Find(&logs)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"logs": logs, "total": total},
		})
	}
}

func deriveAction(method, path string) string {
	switch method {
	case http.MethodPost:
		return "CREATE:" + entityTypeFromPath(path)
	case http.MethodPut:
		return "UPDATE:" + entityTypeFromPath(path)
	case http.MethodPatch:
		return "PATCH:" + entityTypeFromPath(path)
	case http.MethodDelete:
		return "DELETE:" + entityTypeFromPath(path)
	}
	return method + ":" + path
}

func entityTypeFromPath(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for i, s := range segments {
		if s == "admin" && i+1 < len(segments) {
			if next := segments[i+1]; next != "" && !strings.HasPrefix(next, ":") {
				return next
			}
		}
	}
	if len(segments) > 0 {
		return segments[len(segments)-1]
	}
	return "unknown"
}
