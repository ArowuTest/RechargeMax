package handlers

import (
	"github.com/google/uuid"
"net/http"
"strconv"

"github.com/gin-gonic/gin"

"rechargemax/internal/application/services"
)

type NotificationHandler struct {
notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
return &NotificationHandler{notificationService: notificationService}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
msisdn := c.GetString("msisdn")
limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

notifications, total, err := h.notificationService.GetNotifications(c.Request.Context(), msisdn, limit, offset)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, gin.H{"data": notifications, "total": total})
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
msisdn := c.GetString("msisdn")

count, err := h.notificationService.GetUnreadCount(c.Request.Context(), msisdn)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, gin.H{"data": gin.H{"unread_count": count}})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	err = h.notificationService.MarkAsRead(c.Request.Context(), notificationID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}
