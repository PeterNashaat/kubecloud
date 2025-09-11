package activities

import (
	"context"
	"fmt"

	"kubecloud/internal/logger"
	"kubecloud/internal/notification"
	"kubecloud/models"

	"github.com/xmonader/ewf"
)

func hookWorkflowStarted(ctx context.Context, w *ewf.Workflow) {
	logger.GetLogger().Info().Str("workflow_name", w.Name).Msg("Starting workflow")
}

func hookStepStarted(ctx context.Context, w *ewf.Workflow, step *ewf.Step) {
	logger.GetLogger().Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Starting step")
}

func hookWorkflowDone(_ context.Context, wf *ewf.Workflow, err error) {
	if err != nil {
		logger.GetLogger().Error().Err(err).Str("workflow_name", wf.Name).Msg("Workflow failed")
	} else {
		logger.GetLogger().Info().Str("workflow_name", wf.Name).Msg("Workflow completed successfully")
	}
}

func hookStepDone(_ context.Context, w *ewf.Workflow, step *ewf.Step, err error) {
	if err != nil {
		logger.GetLogger().Error().Err(err).Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step failed")
	} else {
		logger.GetLogger().Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step completed successfully")
	}
}

func newKubecloudWorkflowTemplate() ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
			hookWorkflowStarted,
		},
		AfterWorkflowHooks: []ewf.AfterWorkflowHook{
			hookWorkflowDone,
		},
		BeforeStepHooks: []ewf.BeforeStepHook{
			hookStepStarted,
		},
		AfterStepHooks: []ewf.AfterStepHook{
			hookStepDone,
		},
	}
}

func getOrdinalSuffix(n int) string {
	switch n % 100 {
	case 11, 12, 13:
		return "th"
	default:
		switch n % 10 {
		case 1:
			return "st"
		case 2:
			return "nd"
		case 3:
			return "rd"
		default:
			return "th"
		}
	}
}

func getDeployNodeStepName(index int) string {
	return fmt.Sprintf("deploy-%d%s-node", index, getOrdinalSuffix(index))
}

// hookVerificationWorkflowCompleted is called after user verification workflow completes
func hookVerificationWorkflowCompleted(notificationService *notification.NotificationService) ewf.AfterWorkflowHook {
	return func(ctx context.Context, w *ewf.Workflow, err error) {
		userID := w.State["user_id"].(int)
		notifcation := models.NewNotification(userID, "user_registration", map[string]string{}, models.WithNoPersist())
		if err != nil {
			notifcation.Severity = models.NotificationSeverityError
			notifcation.Payload = notification.MergePayload(notification.CommonPayload{
				Message: "User verification failed. Please try again later.",
				Subject: "User verification",
			}, map[string]string{})
			err = notificationService.Send(ctx, notifcation)
			if err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to send user verification notification")
			}
			return
		}

		notifcation.Severity = models.NotificationSeveritySuccess
		notifcation.Payload = notification.MergePayload(notification.CommonPayload{
			Message: "User verification completed successfully",
			Subject: "User verification",
		}, map[string]string{})
		err = notificationService.Send(ctx, notifcation)
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to send user verification notification")
		}
	}
}
