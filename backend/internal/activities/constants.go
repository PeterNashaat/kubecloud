package activities

const (
	// Workflow names
	WorkflowChargeBalance    = "charge-balance"
	WorkflowUserRegistration = "user-registration"
	WorkflowUserVerification = "user-verification"
	WorkflowRedeemVoucher    = "redeem-voucher"
	WorkflowReserveNode      = "reserve-node"
	WorkflowUnreserveNode    = "unreserve-node"

	// Step names
	StepCreatePaymentIntent   = "create_payment_intent"
	StepTransferTFTs          = "transfer_tfts"
	StepCancelPaymentIntent   = "cancel_payment_intent"
	StepUpdateUserBalance     = "update_user_balance"
	StepSendVerificationEmail = "send_verification_email"
	StepSetupTFChain          = "setup_tfchain"
	StepCreateStripeCustomer  = "create_stripe_customer"
	StepSaveUser              = "save_user"
	StepUpdateUserVerified    = "update_user_verified"
	StepSendWelcomeEmail      = "send_welcome_email"
	StepCreateIdentity        = "create_identity"
	StepReserveNode           = "reserve_node"
	StepUnreserveNode         = "unreserve-node"
)
