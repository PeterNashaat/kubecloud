package activities

import (
	"context"
	"fmt"
	"kubecloud/internal/notification"
	"kubecloud/models"

	"github.com/xmonader/ewf"
)

func SendNotification(db models.DB, notifiers map[string]notification.Notifier) ewf.StepFn {
	return func(ctx context.Context, wf ewf.State) error {
		raw, ok := wf["notification"]
		if !ok {
			return fmt.Errorf("missing notification in workflow state")
		}
		notif, ok := raw.(*models.Notification)
		if !ok || notif == nil {
			return fmt.Errorf("invalid notification in workflow state")
		}
		for _, notifChan := range notif.Channels {
			err := notifiers[notifChan].Notify(*notif)
			if err != nil {
				return fmt.Errorf("failed to send notification (id: %v) to %s: %w", notif.ID, notifChan, err)
			}
		}
		return nil
	}
}
