package activities

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"

	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
	"gorm.io/gorm"
)

func CreateUserStep(config internal.Configuration, db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		emailVal, ok := state["email"]
		if !ok {
			return fmt.Errorf("missing 'email' in state")
		}
		email, ok := emailVal.(string)
		if !ok {
			return fmt.Errorf("'email' in state is not a string")
		}

		nameVal, ok := state["name"]
		if !ok {
			return fmt.Errorf("missing 'name' in state")
		}
		name, ok := nameVal.(string)
		if !ok {
			return fmt.Errorf("'name' in state is not a string")
		}

		passwordVal, ok := state["password"]
		if !ok {
			return fmt.Errorf("missing 'password' in state")
		}
		password, ok := passwordVal.(string)
		if !ok {
			return fmt.Errorf("'password' in state is not a string")
		}

		hashedPassword, err := internal.HashAndSaltPassword([]byte(password))
		if err != nil {
			return fmt.Errorf("hashing password failed: %w", err)
		}

		user := models.User{
			Username: name,
			Email:    email,
			Password: hashedPassword,
			Admin:    internal.Contains(config.Admins, email),
		}

		existingUser, err := db.GetUserByEmail(email)
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		if err == nil && !existingUser.Verified {
			user.ID = existingUser.ID
			if updateErr := db.UpdateUserByID(&user); updateErr != nil {
				return fmt.Errorf("failed to update user: %w", updateErr)
			}
			return nil
		}

		err = db.RegisterUser(&user)
		if err != nil {
			return fmt.Errorf("user registration failed: %w", err)
		}

		return nil
	}
}

func SendVerificationEmailStep(mailService internal.MailService, config internal.Configuration) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		emailVal, ok := state["email"]
		if !ok {
			return fmt.Errorf("missing 'email' in state")
		}
		email, ok := emailVal.(string)
		if !ok {
			return fmt.Errorf("'email' in state is not a string")
		}

		nameVal, ok := state["name"]
		if !ok {
			return fmt.Errorf("missing 'name' in state")
		}
		name, ok := nameVal.(string)
		if !ok {
			return fmt.Errorf("'name' in state is not a string")
		}

		code := internal.GenerateRandomCode()
		subject, body := mailService.SignUpMailContent(code, config.MailSender.Timeout, name, config.Server.Host)

		if err := mailService.SendMail(config.MailSender.Email, email, subject, body); err != nil {
			return fmt.Errorf("send mail failed: %w", err)
		}

		state["code"] = code
		return nil
	}
}

func UpdateCodeStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		emailVal, ok := state["email"]
		if !ok {
			return fmt.Errorf("missing 'email' in state")
		}
		email, ok := emailVal.(string)
		if !ok {
			return fmt.Errorf("'email' in state is not a string")
		}

		codeVal, ok := state["code"]
		if !ok {
			return fmt.Errorf("missing 'code' in state")
		}
		code, ok := codeVal.(int)
		if !ok {
			return fmt.Errorf("'code' in state is not a int")
		}

		existingUser, err := db.GetUserByEmail(email)
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		existingUser.Code = code
		return db.UpdateUserByID(&existingUser)
	}
}

func SetupTFChainStep(client *substrate.Substrate, config internal.Configuration, sse *internal.SSEManager, db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		sse.Notify(fmt.Sprintf("%d", userID), "user_registration", "Registering user is in progress")

		mnemonic, _, err := internal.SetupUserOnTFChain(client, config)
		if err != nil {
			return err
		}

		if err := db.UpdateUserByID(&models.User{
			ID:       userID,
			Mnemonic: mnemonic,
		}); err != nil {
			return fmt.Errorf("failed to update user mnemonic: %w", err)
		}

		state["mnemonic"] = mnemonic
		return nil
	}
}

func CreateStripeCustomerStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		emailVal, ok := state["email"]
		if !ok {
			return fmt.Errorf("missing 'email' in state")
		}
		email, ok := emailVal.(string)
		if !ok {
			return fmt.Errorf("'email' in state is not a string")
		}

		nameVal, ok := state["name"]
		if !ok {
			return fmt.Errorf("missing 'name' in state")
		}
		name, ok := nameVal.(string)
		if !ok {
			return fmt.Errorf("'name' in state is not a string")
		}

		customer, err := internal.CreateStripeCustomer(name, email)
		if err != nil {
			return err
		}

		if err := db.UpdateUserByID(&models.User{
			ID:               userID,
			StripeCustomerID: customer.ID,
		}); err != nil {
			return fmt.Errorf("failed to update user stripe customer: %w", err)
		}

		state["stripe_customer_id"] = customer.ID
		return nil
	}
}

func CreateKYCSponsorship(kycClient *internal.KYCClient, sse *internal.SSEManager, sponsorAddress string, sponsorKeyPair subkey.KeyPair, db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		mnemonicVal, ok := state["mnemonic"]
		if !ok {
			return fmt.Errorf("missing 'mnemonic' in state")
		}
		mnemonic, ok := mnemonicVal.(string)
		if !ok {
			return fmt.Errorf("'mnemonic' in state is not a string")
		}

		state["sponsored"] = false
		state["sponsee_address"] = ""

		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		sse.Notify(fmt.Sprintf("%d", userID), "user_registration", "Account verification is in progress")

		// Set user.AccountAddress from mnemonic
		sponseeKeyPair, err := internal.KeyPairFromMnemonic(mnemonic)
		if err != nil {
			log.Error().Err(err).Msg("failed to create keypair for SS58 address")
			return err
		}

		sponseeAddress, err := internal.AccountAddressFromKeypair(sponseeKeyPair)
		if err != nil {
			log.Error().Err(err).Msg("failed to get SS58 address")
			return err
		}

		if err := kycClient.CreateSponsorship(ctx, sponsorAddress, sponsorKeyPair, sponseeAddress, sponseeKeyPair); err != nil {
			return fmt.Errorf("failed to create KYC sponsorship: %w", err)
		}

		if err := db.UpdateUserByID(&models.User{
			ID:             userID,
			Sponsored:      true,
			AccountAddress: sponseeAddress,
			Verified:       true,
		}); err != nil {
			return fmt.Errorf("failed to update user data: %w", err)
		}

		state["sponsee_address"] = sponseeAddress
		state["sponsored"] = true
		return nil
	}
}

func SendWelcomeEmailStep(mailService internal.MailService, config internal.Configuration) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		emailVal, ok := state["email"]
		if !ok {
			return fmt.Errorf("missing 'email' in state")
		}
		email, ok := emailVal.(string)
		if !ok {
			return fmt.Errorf("'email' in state is not a string")
		}

		nameVal, ok := state["name"]
		if !ok {
			return fmt.Errorf("missing 'name' in state")
		}
		name, ok := nameVal.(string)
		if !ok {
			return fmt.Errorf("'name' in state is not a string")
		}

		subject, body := mailService.WelcomeMailContent(name, config.Server.Host)
		if err := mailService.SendMail(config.MailSender.Email, email, subject, body); err != nil {
			return fmt.Errorf("send mail failed: %w", err)
		}
		return nil
	}
}

func CreatePaymentIntentStep(currency string) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		customerIDVal, ok := state["stripe_customer_id"]
		if !ok {
			return fmt.Errorf("missing 'stripe_customer_id' in state")
		}
		customerID, ok := customerIDVal.(string)
		if !ok {
			return fmt.Errorf("'stripe_customer_id' in state is not a string")
		}
		paymentMethodIDVal, ok := state["payment_method_id"]
		if !ok {
			return fmt.Errorf("missing 'payment_method_id' in state")
		}
		paymentMethodID, ok := paymentMethodIDVal.(string)
		if !ok {
			return fmt.Errorf("'payment_method_id' in state is not a string")
		}
		amountVal, ok := state["amount"]
		if !ok {
			return fmt.Errorf("missing 'amount' in state")
		}
		var amount uint64
		switch v := amountVal.(type) {
		case int:
			amount = uint64(v)
		case uint64:
			amount = v
		case float64:
			amount = uint64(v)
		default:
			return fmt.Errorf("'amount' in state is not a number")
		}

		intent, err := internal.CreatePaymentIntent(customerID, paymentMethodID, currency, amount)
		if err != nil {
			return fmt.Errorf("error creating payment intent: %w", err)
		}
		state["payment_intent_id"] = intent.ID
		return nil
	}
}

func CreatePendingRecord(substrateClient *substrate.Substrate, db models.DB, systemMnemonic string) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		amountVal, ok := state["amount"]
		if !ok {
			return fmt.Errorf("missing 'amount' in state")
		}
		var amount uint64
		switch v := amountVal.(type) {
		case int:
			amount = uint64(v)
		case uint64:
			amount = v
		case float64:
			amount = uint64(v)
		default:
			return fmt.Errorf("'amount' in state is not a number")
		}

		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		requestedTFTs, err := internal.FromUSDMillicentToTFT(substrateClient, amount)
		if err != nil {
			log.Error().Err(err).Msg("error converting usd")
			return err
		}

		if err = db.CreatePendingRecord(&models.PendingRecord{
			UserID:    userID,
			TFTAmount: requestedTFTs,
		}); err != nil {
			log.Error().Err(err).Send()
			return err
		}

		return nil
	}
}

func UpdateCreditCardBalanceStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		amountVal, ok := state["amount"]
		if !ok {
			return fmt.Errorf("missing 'amount' in state")
		}
		var amount uint64
		switch v := amountVal.(type) {
		case float64:
			amount = uint64(v)
		case uint64:
			amount = v
		case int:
			amount = uint64(v)
		default:
			return fmt.Errorf("'amount' in state is not a float64 or int")
		}

		user, err := db.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		user.CreditCardBalance += amount
		if err := db.UpdateUserByID(&user); err != nil {
			return fmt.Errorf("error updating user: %w", err)
		}

		state["new_balance"] = user.CreditCardBalance
		state["mnemonic"] = user.Mnemonic
		return nil
	}
}

func UpdateCreditedBalanceStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		amountVal, ok := state["amount"]
		if !ok {
			return fmt.Errorf("missing 'amount' in state")
		}
		var amount uint64
		switch v := amountVal.(type) {
		case float64:
			amount = uint64(v)
		case uint64:
			amount = v
		case int:
			amount = uint64(v)
		default:
			return fmt.Errorf("'amount' in state is not a float64 or int")
		}

		user, err := db.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("user is not found: %w", err)
		}

		user.CreditedBalance += amount
		if err := db.UpdateUserByID(&user); err != nil {
			return fmt.Errorf("error updating user: %w", err)
		}
		state["new_balance"] = user.CreditedBalance
		return nil
	}
}
