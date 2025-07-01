package app

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/middlewares"
	"kubecloud/models/sqlite"
	"net/http"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v82"
)

// App holds all configurations for the app
type App struct {
	router     *gin.Engine
	httpServer *http.Server
	config     internal.Configuration
	handlers   Handler
}

// NewApp create new instance of the app with all configs
func NewApp(config internal.Configuration) (*App, error) {
	router := gin.Default()

	stripe.Key = config.StripeSecret

	tokenHandler := internal.NewTokenHandler(
		config.JWT.Secret,
		time.Duration(config.JWT.AccessTokenExpiryMinutes)*time.Minute,
		time.Duration(config.JWT.RefreshTokenExpiryHours)*time.Hour,
	)

	db, err := sqlite.NewSqliteStorage(config.Database.File)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user storage")
		return nil, fmt.Errorf("failed to create user storage: %w", err)
	}

	mailService := internal.NewMailService(config.MailSender.SendGridKey)

	gridProxy := proxy.NewRetryingClient(proxy.NewClient(config.GridProxyURL))

	manager := substrate.NewManager(config.TFChainURL)
	substrateClient, err := manager.Substrate()

	if err != nil {
		log.Error().Err(err).Msg("failed to connect to substrate client")
		return nil, fmt.Errorf("failed to connect to substrate client: %w", err)
	}

	graphqlURL := []string{config.GraphqlURL}
	graphqlClient, err := graphql.NewGraphQl(graphqlURL...)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to graphql client")
		return nil, fmt.Errorf("failed to connect to graphql client: %w", err)
	}

	firesquidURL := []string{config.FiresquidURL}
	firesquidClient, err := graphql.NewGraphQl(firesquidURL...)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to firesquid client")
		return nil, fmt.Errorf("failed to connect to firesquid client: %w", err)
	}

	handler := NewHandler(tokenHandler, db, config, mailService, gridProxy, substrateClient, graphqlClient, firesquidClient)

	app := &App{
		router:   router,
		config:   config,
		handlers: *handler,
	}

	app.registerHandlers()

	return app, nil

}

// registerHandlers registers all routes
func (app *App) registerHandlers() {
	v1 := app.router.Group("/api/v1")
	{
		v1.GET("/nodes", app.handlers.ListNodesHandler)

		adminGroup := v1.Group("")
		adminGroup.Use(middlewares.AdminMiddleware(app.handlers.tokenManager))
		{
			usersGroup := adminGroup.Group("/users")
			{
				usersGroup.GET("", app.handlers.ListUsersHandler)
				usersGroup.DELETE("/:user_id", app.handlers.DeleteUsersHandler)
				usersGroup.POST("/:user_id/credit", app.handlers.CreditUserHandler)
			}

			adminGroup.GET("/invoices", app.handlers.ListAllInvoicesHandler)

			vouchersGroup := adminGroup.Group("/vouchers")
			{
				vouchersGroup.POST("/generate", app.handlers.GenerateVouchersHandler)
				vouchersGroup.GET("", app.handlers.ListVouchersHandler)

			}

		}

		userGroup := v1.Group("/user")
		{
			userGroup.POST("/register", app.handlers.RegisterHandler)
			userGroup.POST("/register/verify", app.handlers.VerifyRegisterCode)
			userGroup.POST("/login", app.handlers.LoginUserHandler)
			userGroup.POST("/refresh", app.handlers.RefreshTokenHandler)
			userGroup.POST("/forgot_password", app.handlers.ForgotPasswordHandler)
			userGroup.POST("/forgot_password/verify", app.handlers.VerifyForgetPasswordCodeHandler)

			authGroup := userGroup.Group("")
			authGroup.Use(middlewares.UserMiddleware(app.handlers.tokenManager))
			{
				authGroup.GET("/", app.handlers.GetUserHandler)
				authGroup.POST("/change_password", app.handlers.ChangePasswordHandler)
				authGroup.POST("/nodes/:node_id", app.handlers.ReserveNodeHandler)
				authGroup.GET("/nodes", app.handlers.ListReservedNodeHandler)
				authGroup.POST("/nodes/unreserve/:contract_id", app.handlers.UnreserveNodeHandler)
				authGroup.POST("/charge_balance", app.handlers.ChargeBalance)
				authGroup.GET("/balance", app.handlers.GetUserBalance)
				authGroup.PUT("/redeem/:voucher_code", app.handlers.RedeemVoucherHandler)
				authGroup.GET("/invoice/:invoice_id", app.handlers.DownloadInvoiceHandler)
				authGroup.GET("/invoice/", app.handlers.ListUserInvoicesHandler)
			}

		}

	}

}

func (app *App) StartBackgroundWorkers() {
	go app.handlers.MonthlyInvoicesHandler()
}

// Run starts the server
func (app *App) Run() error {
	addr := fmt.Sprintf("%s:%s", app.config.Server.Host, app.config.Server.Port)
	app.httpServer = &http.Server{
		Addr:    addr,
		Handler: app.router,
	}

	log.Info().Msgf("Starting server at http://%s", addr)

	if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Failed to start server")
	}

	return app.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (app *App) Shutdown(ctx context.Context) error {
	if app.httpServer != nil {
		return app.httpServer.Shutdown(ctx)
	}
	return nil
}
