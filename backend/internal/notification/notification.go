package notification

import (
	"kubecloud/models"
)

const (
	ChannelUI   = "ui"
	ChannelEmail = "email"
)

type Notifier interface {
	Notify(notification *models.Notification) error
	GetType()string
}

type NotificationServiceInterface interface {
	Send(notificationType string, payload any, userID string) error
	GetUserNotifications(userID string, limit, offset int) ([]models.Notification, error)
	MarkAsRead(notificationID string) error
}

type NotificationService struct {
	db       models.DB
	channels map[string][]Notifier
}

func NewNotificationService(db models.DB, notifiers ...Notifier) *NotificationService {
	channels := make(map[string][]Notifier)
	for _, notifier := range notifiers {
		channels[notifier.GetType()] = append(channels[notifier.GetType()], notifier)
	}
	
	return &NotificationService{
		db:       db,
		channels: channels,
	}
}

func (s *NotificationService) Send(notificationType string, payload any, userID string) error {
	return nil
}
