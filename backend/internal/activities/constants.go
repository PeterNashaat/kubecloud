package activities

const (
	// Workflow names
	WorkflowChargeBalance      = "charge-balance"
	WorkflowAdminCreditBalance = "admin-credit-balance"
	WorkflowUserRegistration   = "user-registration"
	WorkflowUserVerification   = "user-verification"
	WorkflowRedeemVoucher      = "redeem-voucher"
	WorkflowReserveNode        = "reserve-node"
	WorkflowUnreserveNode      = "unreserve-node"
	WorkflowDeleteCluster      = "delete-cluster"
	WorkflowAddNode            = "add-node"
	WorkflowRemoveNode         = "remove-node"

	// Step names
	StepCreatePaymentIntent     = "create_payment_intent"
	StepCreatePendingRecord     = "create_pending_record"
	StepUpdateCreditCardBalance = "update_user_balance"
	StepSendVerificationEmail   = "send_verification_email"
	StepSetupTFChain            = "setup_tfchain"
	StepCreateStripeCustomer    = "create_stripe_customer"
	StepCreateKYCSponsorship    = "create_kyc_sponsorship"
	StepSaveUser                = "save_user"
	StepUpdateUserVerified      = "update_user_verified"
	StepSendWelcomeEmail        = "send_welcome_email"
	StepCreateIdentity          = "create_identity"
	StepReserveNode             = "reserve_node"
	StepUnreserveNode           = "unreserve-node"
	StepUpdateCreditedBalance   = "update-credited-balance"
	StepRemoveNode              = "remove-node"
	StepStoreDeployment         = "store-deployment"
	StepAddNode                 = "add-node"
	StepUpdateNetwork           = "update-network"
	StepRemoveCluster           = "remove-cluster"
	StepRemoveClusterFromDB     = "remove-cluster-from-db"
	StepDeployNode              = "deploy-node"
	StepDeployNetwork           = "deploy-network"
)
