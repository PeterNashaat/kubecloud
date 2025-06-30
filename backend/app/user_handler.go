package app

import (
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentmethod"
	"gorm.io/gorm"
)

// Handler struct holds configs for all handlers
type Handler struct {
	tokenManager    internal.TokenManager
	db              models.DB
	config          internal.Configuration
	mailService     internal.MailService
	proxyClient     proxy.Client
	substrateClient *substrate.Substrate
}

// NewHandler create new handler
func NewHandler(tokenManager internal.TokenManager, db models.DB, config internal.Configuration, mailService internal.MailService, gridproxy proxy.Client, substrateClient *substrate.Substrate) *Handler {
	return &Handler{
		tokenManager:    tokenManager,
		db:              db,
		config:          config,
		mailService:     mailService,
		proxyClient:     gridproxy,
		substrateClient: substrateClient,
	}
}

// RegisterInput struct for data needed when user register
type RegisterInput struct {
	Name            string `json:"name" binding:"required" validate:"min=3,max=64"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8,max=64"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

// LoginInput struct for login handler
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=3,max=64"`
}

// RefreshTokenInput struct when user refresh token
type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// EmailInput struct for user when forgetting password
type EmailInput struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyCodeInput struct takes verification code from user
type VerifyCodeInput struct {
	Email string `json:"email" binding:"required,email"`
	Code  int    `json:"code" binding:"required,numeric"`
}

// ChangePasswordInput struct for user to change password
type ChangePasswordInput struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8,max=64"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

// ChargeBalanceInput struct holds required data to charge users' balance
type ChargeBalanceInput struct {
	CardType     string `json:"card_type" binding:"required"`
	PaymentToken string `json:"payment_method_id" binding:"required"`
	Amount       uint64 `json:"amount" binding:"required"`
}

// RegisterHandler registers user to the system
func (h *Handler) RegisterHandler(c *gin.Context) {
	var request RegisterInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// password and confirm password should match
	if request.Password != request.ConfirmPassword {
		Error(c, http.StatusBadRequest, "Validation Error", "password and confirm password don't match")
		return
	}

	// check if user previously exists
	existingUser, getErr := h.db.GetUserByEmail(request.Email)
	if getErr != gorm.ErrRecordNotFound {
		if existingUser.Verified {
			Error(c, http.StatusConflict, "Conflict", "user already registered")
			return
		}

	}

	code := internal.GenerateRandomCode()
	log.Debug().Int("generated_code", code).Send()
	subject, body := h.mailService.SignUpMailContent(code, h.config.MailSender.Timeout, request.Name, h.config.Server.Host)

	err := h.mailService.SendMail(h.config.MailSender.Email, request.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Msg("failed to send verification code")
		InternalServerError(c)
		return
	}

	// hash password
	hashedPassword, err := internal.HashAndSaltPassword([]byte(request.Password))
	if err != nil {
		log.Error().Err(err).Msg("error hashing password")
		InternalServerError(c)
		return
	}

	isAdmin := internal.Contains(h.config.Admins, request.Email)

	mnemonic, _, err := internal.SetupUserOnTFChain(h.substrateClient, h.config)
	if err != nil {
		log.Error().Err(err).Msg("failed to setup user on TFChain")
		InternalServerError(c)
		return
	}
	customer, err := internal.CreateStripeCustomer(request.Name, request.Email)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	user := models.User{
		StripeCustomerID: customer.ID,
		Username:         request.Name,
		Email:            request.Email,
		Password:         hashedPassword,
		Admin:            isAdmin,
		Code:             code,
		Mnemonic:         mnemonic,
	}

	// If user exists but not verified
	if getErr != gorm.ErrRecordNotFound {
		if !existingUser.Verified {
			user.ID = existingUser.ID
			user.UpdatedAt = time.Now()
			err = h.db.UpdateUserByID(&user)
			if err != nil {
				log.Error().Err(err).Send()
				InternalServerError(c)
				return
			}
		}
	}

	// create user model in db
	if getErr != nil {
		err = h.db.RegisterUser(&user)
		if err != nil {
			log.Error().Err(err).Send()
			InternalServerError(c)
			return
		}
	}

	Success(c, http.StatusOK, "Verification code sent successfully", map[string]interface{}{
		"email":   request.Email,
		"timeout": fmt.Sprintf("%d seconds", h.config.MailSender.Timeout),
	})

}

// VerifyRegisterCode verifies email when signing uo
func (h *Handler) VerifyRegisterCode(c *gin.Context) {
	var request VerifyCodeInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user by email")
		Error(c, http.StatusBadRequest, "verification failed", "email or password is incorrect")
		return

	}

	if user.Verified {
		Error(c, http.StatusBadRequest, "verification failed", "user already registered")
		return
	}

	if user.Code != request.Code {
		Error(c, http.StatusBadRequest, "verification failed", "wrong code")
		return
	}

	if user.UpdatedAt.Add(time.Duration(h.config.MailSender.Timeout) * time.Second).Before(time.Now()) {
		Error(c, http.StatusBadRequest, "verification failed", "code has expired")
		return
	}

	err = h.db.UpdateUserVerification(user.ID, true)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return

	}

	subject, body := h.mailService.WelcomeMailContent(user.Username, h.config.Server.Host)
	err = h.mailService.SendMail(h.config.MailSender.Email, request.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		InternalServerError(c)
		return
	}
	Success(c, http.StatusCreated, "token pair generated", tokenPair)

}

// LoginUserHandler logs user into the system
func (h *Handler) LoginUserHandler(c *gin.Context) {
	var request LoginInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user by email")
		Error(c, http.StatusBadRequest, "verification failed", "email or password is incorrect")
		return

	}

	// verify password
	match := internal.VerifyPassword(user.Password, request.Password)
	if !match {
		Error(c, http.StatusUnauthorized, "login failed", "email or password is incorrect")
		return
	}

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		InternalServerError(c)
		return
	}
	Success(c, http.StatusCreated, "token pair generated", tokenPair)

}

// RefreshTokenHandler handles token refresh requests
func (h *Handler) RefreshTokenHandler(c *gin.Context) {
	var request RefreshTokenInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	accessToken, err := h.tokenManager.AccessTokenFromRefresh(request.RefreshToken)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusUnauthorized, "refresh token failed", "Invalid or expired refresh token")

		return
	}

	Success(c, http.StatusOK, "access token refreshed successfully", map[string]interface{}{
		"access_token": accessToken,
	})
}

// ForgotPasswordHandler sends user verification code
func (h *Handler) ForgotPasswordHandler(c *gin.Context) {
	var request EmailInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user ")
		Error(c, http.StatusNotFound, "user lookup failed", "failed to get user")
		return

	}

	code := internal.GenerateRandomCode()
	subject, body := h.mailService.ResetPasswordMailContent(code, h.config.MailSender.Timeout, user.Username, h.config.Server.Host)
	err = h.mailService.SendMail(h.config.MailSender.Email, request.Email, subject, body)

	if err != nil {
		log.Error().Err(err).Msg("failed to send verification code")
		InternalServerError(c)
		return
	}

	err = h.db.UpdateUserByID(
		&models.User{
			ID:        user.ID,
			UpdatedAt: time.Now(),
			Code:      code,
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("error updating user data")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Verification code sent", map[string]interface{}{
		"email":   request.Email,
		"timeout": fmt.Sprintf("%d seconds", h.config.MailSender.Timeout),
	})

}

// VerifyForgetPasswordCodeHandler verifies code sent to user when forgetting password
func (h *Handler) VerifyForgetPasswordCodeHandler(c *gin.Context) {
	var request VerifyCodeInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
			return

		}
		log.Error().Err(err).Msg("failed to get user by email")
		InternalServerError(c)
		return

	}

	if user.Code != request.Code {
		Error(c, http.StatusBadRequest, "Invalid code", "")
		return
	}

	if user.UpdatedAt.Add(time.Duration(h.config.MailSender.Timeout) * time.Second).Before(time.Now()) {
		Error(c, http.StatusBadRequest, "code expired", "verification code has expired")

		return
	}
	isAdmin := internal.Contains(h.config.Admins, request.Email)

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, isAdmin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		InternalServerError(c)
		return
	}
	Success(c, http.StatusCreated, "verification successful", tokenPair)

}

// ChangePasswordHandler changes password of user
func (h *Handler) ChangePasswordHandler(c *gin.Context) {
	var request ChangePasswordInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())

		return
	}

	if request.Password != request.ConfirmPassword {
		Error(c, http.StatusBadRequest, "password mismatch", "password and confirm password don't match")
		return
	}

	// hash password
	hashedPassword, err := internal.HashAndSaltPassword([]byte(request.Password))
	if err != nil {
		log.Error().Err(err).Msg("error hashing password")
		InternalServerError(c)
		return
	}

	err = h.db.UpdatePassword(request.Email, hashedPassword)
	if err == gorm.ErrRecordNotFound {
		log.Error().Err(err).Msg("user not found")
		Error(c, http.StatusNotFound, "user not found", err.Error())

		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return

	}

	Success(c, http.StatusOK, "password updated successfully", nil)

}

func (h *Handler) ChargeBalance(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		Error(c, http.StatusUnauthorized, "User ID not found in context", "")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		log.Error().Msg("Invalid user ID or type")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	var request ChargeBalanceInput
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if request.Amount <= 0 {
		Error(c, http.StatusBadRequest, "Amount must be greater than zero", "")
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User not found", "")
		return
	}

	paymentMethod, err := internal.CreatePaymentMethod(request.CardType, request.PaymentToken)
	if err != nil {
		log.Error().Err(err).Msg("error creating payment method")
		InternalServerError(c)
		return
	}

	_, err = paymentmethod.Attach(paymentMethod.ID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(user.StripeCustomerID),
	})
	if err != nil {
		log.Error().Err(err).Msg("error attaching payment method to customer")
		InternalServerError(c)
		return
	}

	intent, err := internal.CreatePaymentIntent(user.StripeCustomerID, paymentMethod.ID, h.config.Currency, request.Amount)
	if err != nil {
		log.Error().Err(err).Msg("error creating payment intent")
		InternalServerError(c)
		return
	}

	user.CreditCardBalance += float64(request.Amount)

	err = h.db.UpdateUserByID(&user)
	if err != nil {
		log.Error().Err(err).Msg("error updating user data")
		InternalServerError(c)
		return
	}

	err = internal.TransferTFTs(h.substrateClient, request.Amount, user.Mnemonic, h.config.SystemAccount.Mnemonics)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Balance is charged successfully", gin.H{
		"payment_intent_id": intent.ID,
		"new_balance":       user.CreditCardBalance,
	})

}

// GetUserBalance returns user's balance in usd
func (h *Handler) GetUserBalance(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		log.Error().Msg("user ID not found in context")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User not found", "")
		return
	}

	usdBalance, err := internal.GetUserBalanceUSD(h.substrateClient, user.Mnemonic)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
	}

	Success(c, http.StatusOK, "Balance fetched", gin.H{
		"balance_usd": usdBalance,
	})

}

func (h *Handler) RedeemVoucherHandler(c *gin.Context) {
	voucherCodeParam := c.Param("voucher_code")
	if voucherCodeParam == "" {
		Error(c, http.StatusBadRequest, "Voucher Code is required", "")
		return
	}
	userIDVal, exists := c.Get("user_id")
	if !exists {
		log.Error().Msg("user ID not found in context")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User not found", "")
		return
	}

	// check voucher exists
	voucher, err := h.db.GetVoucherByCode(voucherCodeParam)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "Voucher not found", "")
		return
	}

	// check voucher not redeemed
	if voucher.Redeemed {
		Error(c, http.StatusBadRequest, "voucher is already redeemed", "")
		return
	}

	// check on expiration time of voucher
	if voucher.ExpiresAt.Before(time.Now()) {
		Error(c, http.StatusBadRequest, "voucher is already expired", "")
		return

	}

	err = h.db.RedeemVoucher(voucher.Code)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	user.CreditedBalance += voucher.Value
	err = h.db.UpdateUserByID(&user)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	err = internal.TransferTFTs(h.substrateClient, uint64(voucher.Value), user.Mnemonic, h.config.SystemAccount.Mnemonics)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	Success(c, http.StatusOK, "Voucher is Redeemed Successfully", nil)

}
