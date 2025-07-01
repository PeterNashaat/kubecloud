package models

import (
	"time"
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

// Notification represents a persistent notification
type Notification struct {
	ID        uint               `json:"id" gorm:"primaryKey"`
	UserID    string             `json:"user_id" gorm:"not null;index"`
	Type      NotificationType   `json:"type" gorm:"not null"`
	Title     string             `json:"title" gorm:"not null"`
	Message   string             `json:"message" gorm:"not null"`
	Data      string             `json:"data,omitempty" gorm:"type:text"` // JSON string for additional data
	TaskID    string             `json:"task_id,omitempty" gorm:"index"`
	Status    NotificationStatus `json:"status" gorm:"default:'unread'"`
	CreatedAt time.Time          `json:"created_at" gorm:"autoCreateTime"`
	ReadAt    *time.Time         `json:"read_at,omitempty"`
}
