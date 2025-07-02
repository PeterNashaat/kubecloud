package app

import (
	"kubecloud/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// NotificationResponse represents a notification response
type NotificationResponse struct {
	ID        uint                      `json:"id"`
	Type      models.NotificationType   `json:"type"`
	Title     string                    `json:"title"`
	Message   string                    `json:"message"`
	Data      string                    `json:"data,omitempty"`
	TaskID    string                    `json:"task_id,omitempty"`
	Status    models.NotificationStatus `json:"status"`
	CreatedAt string                    `json:"created_at"`
	ReadAt    *string                   `json:"read_at,omitempty"`
}

// GetNotificationsHandler retrieves user notifications with pagination
func (h *Handler) GetNotificationsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	notifications, err := h.db.GetUserNotifications(userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notifications"})
		return
	}

	// Convert to response format
	var response []NotificationResponse
	for _, notification := range notifications {
		notificationResp := NotificationResponse{
			ID:        notification.ID,
			Type:      notification.Type,
			Title:     notification.Title,
			Message:   notification.Message,
			Data:      notification.Data,
			TaskID:    notification.TaskID,
			Status:    notification.Status,
			CreatedAt: notification.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if notification.ReadAt != nil {
			readAtStr := notification.ReadAt.Format("2006-01-02T15:04:05Z07:00")
			notificationResp.ReadAt = &readAtStr
		}

		response = append(response, notificationResp)
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": response,
		"limit":         limit,
		"offset":        offset,
		"count":         len(response),
	})
}

// MarkNotificationReadHandler marks a specific notification as read
func (h *Handler) MarkNotificationReadHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationIDStr := c.Param("notification_id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.db.MarkNotificationAsRead(uint(notificationID), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllNotificationsReadHandler marks all notifications as read for a user
func (h *Handler) MarkAllNotificationsReadHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := h.db.MarkAllNotificationsAsRead(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// GetUnreadNotificationCountHandler returns the count of unread notifications
func (h *Handler) GetUnreadNotificationCountHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := h.db.GetUnreadNotificationCount(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notification count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// DeleteNotificationHandler deletes a specific notification
func (h *Handler) DeleteNotificationHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationIDStr := c.Param("notification_id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.db.DeleteNotification(uint(notificationID), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}
