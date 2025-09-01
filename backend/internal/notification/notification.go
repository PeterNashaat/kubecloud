package notification

import (
	"context"
	"fmt"
	"kubecloud/internal"
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

func NewNotificationService(db models.DB, engine *ewf.Engine, notificationConfig internal.NotificationConfig, notifiers ...Notifier) (*NotificationService, error) {
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

	// Load notification templates from configuration file
	if err := s.loadTemplatesFromConfigFile(notificationConfig); err != nil {
		return nil, fmt.Errorf("failed to load notification templates: %w", err)
	}

	return s, nil
}

func (s *NotificationService) RegisterTemplate(notificationType models.NotificationType, template NotificationTemplate) {
	s.templates[notificationType] = template
}

func (s *NotificationService) loadTemplatesFromConfigFile(notificationConfig internal.NotificationConfig) error {
	for templateName, templateConfig := range notificationConfig.TemplateTypes {
		notificationType := models.NotificationType(templateName)

		defaultRule := ChannelRule{
			Channels: templateConfig.Default.Channels,
			Severity: models.NotificationSeverity(templateConfig.Default.Severity),
		}

		byStatusRules := make(map[string]ChannelRule)
		for status, ruleConfig := range templateConfig.ByStatus {
			byStatusRules[status] = ChannelRule{
				Channels: ruleConfig.Channels,
				Severity: models.NotificationSeverity(ruleConfig.Severity),
			}
		}

		s.RegisterTemplate(notificationType, NotificationTemplate{
			Default:  defaultRule,
			ByStatus: byStatusRules,
		})
	}

	return nil
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
