package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeDeploymentUpdate NotificationType = "deployment_update"
	NotificationTypeTaskUpdate       NotificationType = "task_update"
	NotificationTypeConnected        NotificationType = "connected"
	NotificationTypeError            NotificationType = "error"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusUnread NotificationStatus = "unread"
	NotificationStatusRead   NotificationStatus = "read"
)

type NotificationSeverity string

const (
	NotificationSeverityInfo    NotificationSeverity = "info"
	NotificationSeverityError   NotificationSeverity = "error"
	NotificationSeverityWarning NotificationSeverity = "warning"
	NotificationSeveritySuccess NotificationSeverity = "success"
)

// Notification represents a persistent notification
type Notification struct {
	ID        uuid.UUID            `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    int                  `json:"user_id" gorm:"not null;index"`
	TaskID    string               `json:"task_id,omitempty" gorm:"index"`
	Type      NotificationType     `json:"type" gorm:"not null"`
	Severity  NotificationSeverity `json:"severity" gorm:"not null"`
	Channels  []string             `json:"channels" gorm:"not null"`
	Payload   map[string]string    `json:"payload" gorm:"not null"`
	Status    NotificationStatus   `json:"status" gorm:"default:'unread'"`
	CreatedAt time.Time            `json:"created_at" gorm:"autoCreateTime"`
	ReadAt    *time.Time           `json:"read_at,omitempty"`
}

// CreateNotification creates a new notification
func (s *GormDB) CreateNotification(notification *Notification) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.db.Create(notification).Error
}

// GetUserNotifications retrieves notifications for a user with pagination
func (s *GormDB) GetUserNotifications(userID int, limit, offset int) ([]Notification, error) {
	var notifications []Notification
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// MarkNotificationAsRead marks a specific notification as read
func (s *GormDB) MarkNotificationAsRead(notificationID uint, userID int) error {
	now := time.Now()
	result := s.db.Model(&Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":  NotificationStatusRead,
			"read_at": &now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// MarkAllNotificationsAsRead marks all notifications as read for a user
func (s *GormDB) MarkAllNotificationsAsRead(userID int) error {
	now := time.Now()
	return s.db.Model(&Notification{}).
		Where("user_id = ? AND status = ?", userID, NotificationStatusUnread).
		Updates(map[string]interface{}{
			"status":  NotificationStatusRead,
			"read_at": &now,
		}).Error
}

// DeleteNotification deletes a notification for a user
func (s *GormDB) DeleteNotification(notificationID uint, userID int) error {
	result := s.db.Where("id = ? AND user_id = ?", notificationID, userID).Delete(&Notification{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// GetUnreadNotifications retrieves only unread notifications for a user with pagination
func (s *GormDB) GetUnreadNotifications(userID int, limit, offset int) ([]Notification, error) {
	var notifications []Notification
	err := s.db.Where("user_id = ? AND status = ?", userID, NotificationStatusUnread).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// DeleteAllNotifications deletes all notifications for a user
func (s *GormDB) DeleteAllNotifications(userID int) error {
	return s.db.Where("user_id = ?", userID).Delete(&Notification{}).Error
}

// MarkNotificationAsUnread marks a specific notification as unread
func (s *GormDB) MarkNotificationAsUnread(notificationID uint, userID int) error {
	result := s.db.Model(&Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":  NotificationStatusUnread,
			"read_at": nil,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
