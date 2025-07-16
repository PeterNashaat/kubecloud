package activities

import (
	"kubecloud/internal"
	"kubecloud/models"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/xmonader/ewf"
)

func RegisterEWFWorkflows(
	engine *ewf.Engine,
	config internal.Configuration,
	db models.DB,
	mail internal.MailService,
	substrate *substrate.Substrate,
) {
	engine.Register(StepSendVerificationEmail, SendVerificationEmailStep(mail, config))
	engine.Register(StepSetupTFChain, SetupTFChainStep(substrate, config))
	engine.Register(StepCreateStripeCustomer, CreateStripeCustomerStep())
	engine.Register(StepSaveUser, SaveUserStep(db, config))
	engine.Register(StepUpdateUserVerified, UpdateUserVerifiedStep(db))
	engine.Register(StepSendWelcomeEmail, SendWelcomeEmailStep(mail, config))
	engine.Register(StepCreatePaymentIntent, CreatePaymentIntentStep(config.Currency))
	engine.Register(StepUpdateUserBalance, UpdateUserBalanceStep(db))
	engine.Register(StepTransferTFTs, TransferTFTsStep(substrate, config.SystemAccount.Mnemonic))
	engine.Register(StepCancelPaymentIntent, CancelPaymentIntentStep())
	engine.Register(StepCreateIdentity, CreateIdentityStep())
	engine.Register(StepReserveNode, ReserveNodeStep(db, substrate))
	engine.Register(StepUnreserveNode, UnreserveNodeStep(db, substrate))

	engine.RegisterTemplate(WorkflowUserRegistration, &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepSendVerificationEmail, RetryPolicy: &ewf.RetryPolicy{
				MaxAttempts: 3,
				Delay:       2,
			}},
			{Name: StepSetupTFChain, RetryPolicy: &ewf.RetryPolicy{
				MaxAttempts: 5,
				Delay:       3,
			}},
			{Name: StepCreateStripeCustomer, RetryPolicy: &ewf.RetryPolicy{
				MaxAttempts: 3,
				Delay:       2,
			}},
			{Name: StepSaveUser, RetryPolicy: &ewf.RetryPolicy{
				MaxAttempts: 2,
				Delay:       2,
			}},
		},
	})

	engine.RegisterTemplate(WorkflowUserVerification, &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepUpdateUserVerified, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
			{Name: StepSendWelcomeEmail, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 3, Delay: 2}},
		},
	})

	engine.RegisterTemplate(WorkflowChargeBalance, &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepCreatePaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
			{Name: StepTransferTFTs, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
			{Name: StepCancelPaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}}, // Compensation step
			{Name: StepUpdateUserBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
		},
	})

	engine.RegisterTemplate(WorkflowRedeemVoucher, &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepTransferTFTs, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
		},
	})

	engine.RegisterTemplate(WorkflowReserveNode, &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepCreateIdentity, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
			{Name: StepReserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
		},
	})

	engine.RegisterTemplate(WorkflowUnreserveNode, &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepUnreserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
		},
	})

}
