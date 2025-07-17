package activities

import (
	"kubecloud/internal"
	"kubecloud/models"
	"time"

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
	engine.Register(StepUpdateCreditCardBalance, UpdateCreditCardBalanceStep(db))
	engine.Register(StepTransferTFTs, TransferTFTsStep(substrate, config.SystemAccount.Mnemonic))
	engine.Register(StepCancelPaymentIntent, CancelPaymentIntentStep())
	engine.Register(StepCreateIdentity, CreateIdentityStep())
	engine.Register(StepReserveNode, ReserveNodeStep(db, substrate))
	engine.Register(StepUnreserveNode, UnreserveNodeStep(db, substrate))
	engine.Register(StepUpdateCreditedBalance, UpdateCreditedBalanceStep(db))

	engine.RegisterTemplate(WorkflowUserRegistration, &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
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
		},
	})

	engine.RegisterTemplate(WorkflowUserVerification, &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepUpdateUserVerified, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
			{Name: StepSendWelcomeEmail, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 3, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		},
	})

	engine.RegisterTemplate(WorkflowChargeBalance, &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepCreatePaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
			{Name: StepTransferTFTs, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
			{Name: StepCancelPaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}}, // Compensation step
			{Name: StepUpdateCreditCardBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		},
	})

	engine.RegisterTemplate(WorkflowRedeemVoucher, &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepTransferTFTs, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
			{Name: StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		},
	})

	engine.RegisterTemplate(WorkflowReserveNode, &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepCreateIdentity, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
			{Name: StepReserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		},
	})

	engine.RegisterTemplate(WorkflowUnreserveNode, &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepUnreserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		},
	})

}
