package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-playground/validator"
)

// Configuration struct holds all configs for the app
type Configuration struct {
	Server               Server             `json:"server" validate:"required,dive"`
	Database             DB                 `json:"database" validate:"required"`
	JWT                  JwtToken           `json:"token" validate:"required"`
	Admins               []string           `json:"admins"`
	MailSender           MailSender         `json:"mailSender"`
	Currency             string             `json:"currency" validate:"required"`
	StripeSecret         string             `json:"stripe_secret" validate:"required"`
	VoucherNameLength    int                `json:"voucher_name_length"  validate:"required,gt=0"`
	GridProxyURL         string             `json:"gridproxy_url" validate:"required"`
	TFChainURL           string             `json:"tfchain_url" validate:"required"`
	TermsANDConditions   TermsANDConditions `json:"terms_and_conditions"`
	ActivationServiceURL string             `json:"activation_service_url" validate:"required"`
	GraphqlURL           string             `json:"graphql_url" validate:"required"`
	SystemAccount        GridAccount        `json:"system_account"`
}

// Server struct holds server's information
type Server struct {
	Host string `json:"host" validate:"required,hostname|ip"`
	Port string `json:"port" validate:"required,numeric"`
}

// DB struct holds database file
type DB struct {
	File string `json:"file" validate:"required"`
}

// JWT Token struct holds info required for JWT Tokens
type JwtToken struct {
	Secret                   string `json:"secret" validate:"required"`
	AccessTokenExpiryMinutes int    `json:"access_token_expiry_minutes" validate:"required,gt=0"` // in minutes
	RefreshTokenExpiryHours  int    `json:"refresh_token_expiry_hours" validate:"required,gt=0"`  // in hours
}

// MailSender struct to hold sender's email, password
type MailSender struct {
	Email       string `json:"email" validate:"required,email"`
	SendGridKey string `json:"sendgrid_key" validate:"required"`
	Timeout     int    `json:"timeout" validate:"min=30"`
}

// TermsANDConditions holds required data for accepting terms and conditions
type TermsANDConditions struct {
	DocumentLink string `json:"document_link" validate:"required"`
	DocumentHash string `json:"document_hash" validate:"required"`
}

// GridAccount holds data for system's account
type GridAccount struct {
	Mnemonics string `json:"mnemonics" validate:"required"`
	Network   string `json:"network" validate:"required"`
}

// ReadConfFile read configurations of json file
func ReadConfFile(path string) (Configuration, error) {
	config := Configuration{}
	file, err := os.Open(path)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	if err := dec.Decode(&config); err != nil {
		return Configuration{}, fmt.Errorf("failed to load config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, ve := range validationErrors {
				fmt.Printf("Validation error on field '%s': %s\n", ve.Namespace(), ve.Tag())
			}
		}
		return Configuration{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}
