package activities

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/xmonader/ewf"
)

func hookWorkflowStarted(ctx context.Context, w *ewf.Workflow) {
	log.Info().Str("workflow_name", w.Name).Msg("Starting workflow")
}

func hookStepStarted(ctx context.Context, w *ewf.Workflow, step *ewf.Step) {
	log.Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Starting step")
}

func hookWorkflowDone(_ context.Context, wf *ewf.Workflow, err error) {
	if err != nil {
		log.Error().Err(err).Str("workflow_name", wf.Name).Msg("Workflow failed")
	} else {
		log.Info().Str("workflow_name", wf.Name).Msg("Workflow completed successfully")
	}
}

func hookStepDone(_ context.Context, w *ewf.Workflow, step *ewf.Step, err error) {
	if err != nil {
		log.Error().Err(err).Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step failed")
	} else {
		log.Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step completed successfully")
	}
}

var baseWorkflowTemplate = ewf.WorkflowTemplate{
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
