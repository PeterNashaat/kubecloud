package notification

import (
	"context"
	"fmt"
	"kubecloud/models"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xmonader/ewf"
)

const (
	ChannelUI    = "ui"
	ChannelEmail = "email"
)

type Notifier interface {
	Notify(notification models.Notification, receiver ...string) error
	GetType() string
}

type ChannelRule struct {
	Channels []string
	Severity models.NotificationSeverity
}

type NotificationTemplate struct {
	Default  ChannelRule
	ByStatus map[string]ChannelRule
}

type NotificationServiceInterface interface {
	Send(ctx context.Context, notificationType string, payload map[string]string, userID string, taskID ...string) error
	GetNotifiers() map[string]Notifier
	RegisterTemplate(notificationType models.NotificationType, template NotificationTemplate)
}

type NotificationService struct {
	db        models.DB
	notifiers map[string]Notifier
	engine    *ewf.Engine
	templates map[models.NotificationType]NotificationTemplate
}

func NewNotificationService(db models.DB, engine *ewf.Engine, notifiers ...Notifier) *NotificationService {
	notifiersMap := make(map[string]Notifier)
	for _, notifier := range notifiers {
		notifiersMap[notifier.GetType()] = notifier
	}

	s := &NotificationService{
		db:        db,
		notifiers: notifiersMap,
		engine:    engine,
		templates: make(map[models.NotificationType]NotificationTemplate),
	}

	s.RegisterTemplate(models.NotificationTypeDeployment, NotificationTemplate{
		Default: ChannelRule{Channels: []string{ChannelUI}, Severity: models.NotificationSeverityInfo},
		ByStatus: map[string]ChannelRule{
			"started":   {Channels: []string{ChannelUI}, Severity: models.NotificationSeverityInfo},
			"succeeded": {Channels: []string{ChannelUI, ChannelEmail}, Severity: models.NotificationSeveritySuccess},
			"failed":    {Channels: []string{ChannelUI, ChannelEmail}, Severity: models.NotificationSeverityError},
			"deleted":   {Channels: []string{ChannelUI}, Severity: models.NotificationSeverityWarning},
		},
	})

	s.RegisterTemplate(models.NotificationTypeBilling, NotificationTemplate{
		Default: ChannelRule{Channels: []string{ChannelUI}, Severity: models.NotificationSeverityInfo},
		ByStatus: map[string]ChannelRule{
			// Adding funds events
			"funds_succeeded": {Channels: []string{ChannelUI, ChannelEmail}, Severity: models.NotificationSeveritySuccess},
			"funds_failed":    {Channels: []string{ChannelUI, ChannelEmail}, Severity: models.NotificationSeverityError},
			// Voucher
			"voucher_redeemed": {Channels: []string{ChannelUI}, Severity: models.NotificationSeveritySuccess},
		},
	})

	s.RegisterTemplate(models.NotificationTypeUser, NotificationTemplate{
		Default: ChannelRule{Channels: []string{ChannelUI}, Severity: models.NotificationSeverityInfo},
		ByStatus: map[string]ChannelRule{
			"password_changed": {Channels: []string{ChannelUI, ChannelEmail}, Severity: models.NotificationSeveritySuccess},
			"ssh_key_added":    {Channels: []string{ChannelUI, ChannelEmail}, Severity: models.NotificationSeveritySuccess},
		},
	})
	return s
}

func (s *NotificationService) RegisterTemplate(notificationType models.NotificationType, template NotificationTemplate) {
	s.templates[notificationType] = template
}

func (s *NotificationService) GetNotifiers() map[string]Notifier {
	return s.notifiers
}

func (s *NotificationService) HandleNotificationSSE(c *gin.Context) {
	uiNotifier, ok := s.notifiers[ChannelUI]
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSE notifier not available"})
		return
	}

	sseNotifier, ok := uiNotifier.(*SSENotifier)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSE notifier not available"})
		return
	}
	sseNotifier.GetSSEManager().HandleSSE(c)

}

func (s *NotificationService) Send(ctx context.Context, notificationType models.NotificationType, payload map[string]string, userID string, taskID ...string) error {
	tpl, ok := s.templates[notificationType]
	if !ok {
		return fmt.Errorf("notification template not found for type: %s", notificationType)
	}

	rule := tpl.Default
	if status, ok := payload["status"]; ok && tpl.ByStatus != nil {
		if r, exists := tpl.ByStatus[status]; exists {
			rule = r
		}
	}

	notification := &models.Notification{
		ID:       uuid.NewString(),
		UserID:   userID,
		Type:     notificationType,
		Channels: append([]string{}, rule.Channels...),
		Severity: rule.Severity,
		Payload:  payload,
	}

	if len(taskID) > 0 {
		notification.TaskID = taskID[0]
	}

	if err := s.db.CreateNotification(notification); err != nil {
		return err
	}

	workflow, err := s.engine.NewWorkflow("send-notification")
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}
	workflow.State["notification"] = notification
	s.engine.RunAsync(ctx, workflow)

	return nil
}
