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
	engine.Register("send_verification_email", SendVerificationEmailStep(mail, config))
	engine.Register("setup_tfchain", SetupTFChainStep(substrate, config))
	engine.Register("create_stripe_customer", CreateStripeCustomerStep())
	engine.Register("save_user", SaveUserStep(db, config))
	engine.Register("update_user_verified", UpdateUserVerifiedStep(db))
	engine.Register("send_welcome_email", SendWelcomeEmailStep(mail, config))
	engine.Register("create_payment_intent", CreatePaymentIntentStep(config.Currency))
	engine.Register("update_user_balance", UpdateUserBalanceStep(db))
	engine.Register("transfer_tfts", TransferTFTsStep(substrate, config.SystemAccount.Mnemonic))
	engine.Register("cancel_payment_intent", CancelPaymentIntentStep())
	engine.Register("create_identity", CreateIdentityStep())
	engine.Register("reserve_node", ReserveNodeStep(db, substrate))
	engine.Register("unreserve_node", UnreserveNodeStep(db, substrate))

	engine.RegisterTemplate("user-registration", &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: "send_verification_email", RetryPolicy: &ewf.RetryPolicy{
				MaxAttempts: 3,
				Delay:       2,
			}},
			{Name: "setup_tfchain", RetryPolicy: &ewf.RetryPolicy{
				MaxAttempts: 5,
				Delay:       3,
			}},
			{Name: "create_stripe_customer", RetryPolicy: &ewf.RetryPolicy{
				MaxAttempts: 3,
				Delay:       2,
			}},
			{Name: "save_user", RetryPolicy: &ewf.RetryPolicy{
				MaxAttempts: 2,
				Delay:       2,
			}},
		},
	})

	engine.RegisterTemplate("user-verification", &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: "update_user_verified", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
			{Name: "send_welcome_email", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 3, Delay: 2}},
		},
	})

	engine.RegisterTemplate("charge-balance", &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: "create_payment_intent", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
			{Name: "transfer_tfts", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
			{Name: "cancel_payment_intent", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}}, // Compensation step
			{Name: "update_user_balance", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
		},
	})

	engine.RegisterTemplate("redeem-voucher", &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: "transfer_tfts", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
		},
	})

	engine.RegisterTemplate("reserve-node", &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: "create_identity", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
			{Name: "reserve_node", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
		},
	})

	engine.RegisterTemplate("unreserve-node", &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: "unreserve_node", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
		},
	})

}
