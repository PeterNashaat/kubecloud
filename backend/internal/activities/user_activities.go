package activities

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"

	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/xmonader/ewf"
	"gorm.io/gorm"
)

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

func SetupTFChainStep(client *substrate.Substrate, config internal.Configuration) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {

		mnemonic, _, err := internal.SetupUserOnTFChain(client, config)
		if err != nil {
			return err
		}

		state["mnemonic"] = mnemonic
		return nil
	}
}

func CreateStripeCustomerStep() ewf.StepFn {
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

		customer, err := internal.CreateStripeCustomer(name, email)
		if err != nil {
			return err
		}

		state["stripe_customer_id"] = customer.ID
		return nil
	}
}

func SaveUserStep(db models.DB, config internal.Configuration) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		nameVal, ok := state["name"]
		if !ok {
			return fmt.Errorf("missing 'name' in state")
		}
		name, ok := nameVal.(string)
		if !ok {
			return fmt.Errorf("'name' in state is not a string")
		}

		emailVal, ok := state["email"]
		if !ok {
			return fmt.Errorf("missing 'email' in state")
		}
		email, ok := emailVal.(string)
		if !ok {
			return fmt.Errorf("'email' in state is not a string")
		}

		passwordVal, ok := state["password"]
		if !ok {
			return fmt.Errorf("missing 'password' in state")
		}
		password, ok := passwordVal.(string)
		if !ok {
			return fmt.Errorf("'password' in state is not a string")
		}

		codeVal, ok := state["code"]
		if !ok {
			return fmt.Errorf("missing 'code' in state")
		}
		code, ok := codeVal.(int)
		if !ok {
			return fmt.Errorf("'code' in state is not a int")
		}

		mnemonicVal, ok := state["mnemonic"]
		if !ok {
			return fmt.Errorf("missing 'mnemonic' in state")
		}
		mnemonic, ok := mnemonicVal.(string)
		if !ok {
			return fmt.Errorf("'mnemonic' in state is not a string")
		}

		stripeIDVal, ok := state["stripe_customer_id"]
		if !ok {
			return fmt.Errorf("missing 'stripe_customer_id' in state")
		}
		stripeID, ok := stripeIDVal.(string)
		if !ok {
			return fmt.Errorf("'stripe_customer_id' in state is not a string")
		}

		// hash password
		hashedPassword, err := internal.HashAndSaltPassword([]byte(password))
		if err != nil {
			return err
		}

		isAdmin := internal.Contains(config.Admins, email)

		user := models.User{
			Username:         name,
			Email:            email,
			Password:         hashedPassword,
			Code:             code,
			Admin:            isAdmin,
			Mnemonic:         mnemonic,
			StripeCustomerID: stripeID,
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

func UpdateUserVerifiedStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		emailVal, ok := state["email"]
		if !ok {
			return fmt.Errorf("missing 'email' in state")
		}
		email, ok := emailVal.(string)
		if !ok {
			return fmt.Errorf("'email' in state is not a string")
		}

		user, err := db.GetUserByEmail(email)
		if err != nil {
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		err = db.UpdateUserVerification(user.ID, true)
		if err != nil {
			return fmt.Errorf("failed to update user verification: %w", err)
		}
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
		var amountInt int
		switch v := amountVal.(type) {
		case int:
			amountInt = v
		case float64:
			amountInt = int(v)
		default:
			return fmt.Errorf("'amount' in state is not a number")
		}

		intent, err := internal.CreatePaymentIntent(customerID, paymentMethodID, currency, uint64(amountInt))
		if err != nil {
			return fmt.Errorf("error creating payment intent: %w", err)
		}
		state["payment_intent_id"] = intent.ID
		return nil
	}
}

func CancelPaymentIntentStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		failed, ok := state["transfer_tfts_failed"].(bool)
		if !ok || !failed {
			return nil
		}
		paymentIntentID, _ := state["payment_intent_id"].(string)
		if paymentIntentID == "" {
			return nil
		}
		if err := internal.CancelPaymentIntent(paymentIntentID); err != nil {
			log.Error().Err(err).Msg("error canceling payment intent in compensation step")
			return err
		}
		return nil
	}
}

func TransferTFTsStep(substrate *substrate.Substrate, systemMnemonic string) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		amountVal, ok := state["amount"]
		if !ok {
			return fmt.Errorf("missing 'amount' in state")
		}
		var amount float64
		switch v := amountVal.(type) {
		case float64:
			amount = v
		case int:
			amount = float64(v)
		default:
			return fmt.Errorf("'amount' in state is not a float64 or int")
		}

		mnemonicVal, ok := state["mnemonic"]
		if !ok {
			return fmt.Errorf("missing 'mnemonic' in state")
		}
		mnemonic, ok := mnemonicVal.(string)
		if !ok {
			return fmt.Errorf("'mnemonic' in state is not a string")
		}

		err := internal.TransferTFTs(substrate, uint64(amount), mnemonic, systemMnemonic)
		if err != nil {
			log.Error().Err(err).Send()
			state["transfer_tfts_failed"] = true
			return err
		}
		state["transfer_tfts_failed"] = false
		return nil
	}
}

func UpdateUserBalanceStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'userID' in state is not a int")
		}

		amountVal, ok := state["amount"]
		if !ok {
			return fmt.Errorf("missing 'amount' in state")
		}
		var amount float64
		switch v := amountVal.(type) {
		case float64:
			amount = v
		case int:
			amount = float64(v)
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
