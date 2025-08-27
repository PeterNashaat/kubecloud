package notification

import (
	"fmt"
	"strings"

	"kubecloud/internal"
	"kubecloud/models"
)

type SSENotifier struct {
	sse *internal.SSEManager
}

func NewSSENotifier(sse *internal.SSEManager) *SSENotifier {
	return &SSENotifier{sse: sse}
}

func (n *SSENotifier) GetType() string {
	return ChannelUI
}

func (n *SSENotifier) Notify(notification models.Notification, receiver ...string) error {
	if n.sse == nil {
		return fmt.Errorf("sse manager is nil")
	}

	msgType := string(notification.Type)
	data := notification.Payload
	if data == nil {
		data = map[string]string{}
	}
	if notification.Severity != "" {
		data["severity"] = string(notification.Severity)
	}
	if len(notification.Channels) > 0 {
		data["channels"] = strings.Join(notification.Channels, ",")
	}

	n.sse.Notify(notification.UserID, msgType, data, notification.TaskID)
	return nil
}
