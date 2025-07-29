package activities

import (
	"context"
	"kubecloud/internal"
	"kubecloud/models"
	"time"

	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/xmonader/ewf"
)

var userWorkfowTemplate = ewf.WorkflowTemplate{
	BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
		func(ctx context.Context, w *ewf.Workflow) {
			log.Info().Str("workflow_name", w.Name).Msg("Starting workflow")
		},
	},
	BeforeStepHooks: []ewf.BeforeStepHook{
		func(ctx context.Context, w *ewf.Workflow, step *ewf.Step) {
			log.Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Starting step")
		},
	},
	AfterStepHooks: []ewf.AfterStepHook{
		func(ctx context.Context, w *ewf.Workflow, step *ewf.Step, err error) {
			if err != nil {
				log.Error().Err(err).Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step failed")
			} else {
				log.Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step completed successfully")
			}
		},
	},
	AfterWorkflowHooks: []ewf.AfterWorkflowHook{
		func(ctx context.Context, w *ewf.Workflow, err error) {
			if err != nil {
				log.Error().Err(err).Str("workflow_name", w.Name).Msg("Workflow completed with error")
			} else {
				log.Info().Str("workflow_name", w.Name).Msg("Workflow completed successfully")
			}
		},
	},
}

func RegisterEWFWorkflows(
	engine *ewf.Engine,
	config internal.Configuration,
	db models.DB,
	mail internal.MailService,
	substrate *substrate.Substrate,
	sse *internal.SSEManager,
) {
	engine.Register(StepSendVerificationEmail, SendVerificationEmailStep(mail, config))
	engine.Register(StepSetupTFChain, SetupTFChainStep(substrate, config))
	engine.Register(StepCreateStripeCustomer, CreateStripeCustomerStep())
	engine.Register(StepSaveUser, SaveUserStep(db, config))
	engine.Register(StepUpdateUserVerified, UpdateUserVerifiedStep(db))
	engine.Register(StepSendWelcomeEmail, SendWelcomeEmailStep(mail, config))
	engine.Register(StepCreatePaymentIntent, CreatePaymentIntentStep(config.Currency))
	engine.Register(StepUpdateCreditCardBalance, UpdateCreditCardBalanceStep(db))
	engine.Register(StepTransferTFTs, TransferTFTsStep(substrate, config.SystemAccount.Mnemonic))
	engine.Register(StepCancelPaymentIntent, CancelPaymentIntentStep())
	engine.Register(StepCreateIdentity, CreateIdentityStep())
	engine.Register(StepReserveNode, ReserveNodeStep(db, substrate))
	engine.Register(StepUnreserveNode, UnreserveNodeStep(db, substrate))
	engine.Register(StepUpdateCreditedBalance, UpdateCreditedBalanceStep(db))
	engine.Register(StepRedeemVoucher, RedeemVoucherStep(db))

	registerWorkflowTemplate := userWorkfowTemplate
	registerWorkflowTemplate.Steps = []ewf.Step{
		{Name: StepSendVerificationEmail, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: StepSetupTFChain, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 5,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: StepCreateStripeCustomer, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: StepSaveUser, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 2,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
	}
	engine.RegisterTemplate(WorkflowUserRegistration, &registerWorkflowTemplate)

	userVerificationTemplate := userWorkfowTemplate
	userVerificationTemplate.Steps = []ewf.Step{
		{Name: StepUpdateUserVerified, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepSendWelcomeEmail, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 3, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowUserVerification, &userVerificationTemplate)

	chargeBalanceTemplate := userWorkfowTemplate
	chargeBalanceTemplate.Steps = []ewf.Step{
		{Name: StepCreatePaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepTransferTFTs, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepCancelPaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}}, // Compensation step
		{Name: StepUpdateCreditCardBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowChargeBalance, &chargeBalanceTemplate)

	redeemVoucherTemplate := userWorkfowTemplate
	redeemVoucherTemplate.Steps = []ewf.Step{
		{Name: StepTransferTFTs, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepRedeemVoucher, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowRedeemVoucher, &redeemVoucherTemplate)

	reserveNodeTemplate := userWorkfowTemplate
	reserveNodeTemplate.Steps = []ewf.Step{
		{Name: StepCreateIdentity, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepReserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowReserveNode, &reserveNodeTemplate)

	unreserveNodeTemplate := userWorkfowTemplate
	unreserveNodeTemplate.Steps = []ewf.Step{
		{Name: StepUnreserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowUnreserveNode, &unreserveNodeTemplate)

	registerDeploymentActivities(engine, db, sse)
}
