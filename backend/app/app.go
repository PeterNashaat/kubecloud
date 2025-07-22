package app

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/middlewares"
	"kubecloud/models/sqlite"
	"net/http"
	"os"
	"strings"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v82"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Import the generated docs package
	_ "kubecloud/docs"
)

// App holds all configurations for the app
type App struct {
	router        *gin.Engine
	httpServer    *http.Server
	config        internal.Configuration
	handlers      Handler
	db            *sqlite.Sqlite
	redis         *internal.RedisClient
	sseManager    *internal.SSEManager
	workerManager *internal.WorkerManager
	gridClient    deployer.TFPluginClient
	appCancel     context.CancelFunc
}

// NewApp create new instance of the app with all configs
func NewApp(config internal.Configuration) (*App, error) {
	router := gin.Default()

	stripe.Key = config.StripeSecret

	tokenHandler := internal.NewTokenHandler(
		config.JwtToken.Secret,
		time.Duration(config.JwtToken.AccessExpiryMinutes)*time.Minute,
		time.Duration(config.JwtToken.RefreshExpiryHours)*time.Hour,
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

	redisClient, err := internal.NewRedisClient(config.Redis)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Redis client")
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	sseManager := internal.NewSSEManager(redisClient, db)

	// start gridclient
	gridClient, err := deployer.NewTFPluginClient(
		config.SystemAccount.Mnemonic,
		deployer.WithNetwork(config.SystemAccount.Network),
		// TODO: remove this after testing
		// deployer.WithSubstrateURL("wss://tfchain.dev.grid.tf/ws"),
		// deployer.WithProxyURL("https://gridproxy.dev.grid.tf"),
		// deployer.WithRelayURL("wss://relay.dev.grid.tf"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TF grid client: %w", err)
	}

	sshPublicKeyBytes, err := os.ReadFile(config.SSH.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSH public key from %s: %w", config.SSH.PublicKeyPath, err)
	}
	sshPublicKey := strings.TrimSpace(string(sshPublicKeyBytes))

	_, appCancel := context.WithCancel(context.Background())

	// Derive sponsor (system) account SS58 address once
	sponsorKeyPair, err := internal.KeyPairFromMnemonic(config.SystemAccount.Mnemonic)
	if err != nil {
		appCancel()
		return nil, fmt.Errorf("failed to create sponsor keypair from system account: %w", err)
	}
	sponsorAddress, err := internal.AccountAddressFromKeypair(sponsorKeyPair)
	if err != nil {
		appCancel()
		return nil, fmt.Errorf("failed to create sponsor address from keypair: %w", err)
	}

	workerManager := internal.NewWorkerManager(redisClient, sseManager, config.DeployerWorkersNum, sshPublicKey, db, config.SystemAccount.Network)

	// Validate KYC configuration
	if strings.TrimSpace(config.KYCVerifierAPIURL) == "" {
		appCancel()
		return nil, fmt.Errorf("KYC verifier API URL is required")
	}
	if strings.TrimSpace(config.KYCChallengeDomain) == "" {
		appCancel()
		return nil, fmt.Errorf("KYC challenge domain is required")
	}

	// Initialize KYC client
	kycClient := internal.NewKYCClient(
		config.KYCVerifierAPIURL,
		config.KYCChallengeDomain,
		nil, // Use default http.Client
	)

	handler := NewHandler(tokenHandler, db, config, mailService, gridProxy,
		substrateClient, graphqlClient, firesquidClient, redisClient,
		sseManager, config.SystemAccount.Network, kycClient, sponsorKeyPair, sponsorAddress)

	app := &App{
		router:        router,
		config:        config,
		handlers:      *handler,
		redis:         redisClient,
		db:            db,
		sseManager:    sseManager,
		workerManager: workerManager,
		appCancel:     appCancel,
		gridClient:    gridClient,
	}

	app.registerHandlers()

	app.workerManager.Start()

	return app, nil

}

// registerHandlers registers all routes
func (app *App) registerHandlers() {
	app.router.Use(middlewares.CorsMiddleware())
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
				authGroup.PUT("/change_password", app.handlers.ChangePasswordHandler)
				authGroup.GET("/nodes", app.handlers.ListReservedNodeHandler)
				authGroup.POST("/nodes/:node_id", app.handlers.ReserveNodeHandler)
				authGroup.DELETE("/nodes/unreserve/:contract_id", app.handlers.UnreserveNodeHandler)
				authGroup.POST("/balance/charge", app.handlers.ChargeBalance)
				authGroup.GET("/balance", app.handlers.GetUserBalance)
				authGroup.PUT("/redeem/:voucher_code", app.handlers.RedeemVoucherHandler)
				authGroup.GET("/invoice/:invoice_id", app.handlers.DownloadInvoiceHandler)
				authGroup.GET("/invoice/", app.handlers.ListUserInvoicesHandler)
				// SSH Key management
				authGroup.GET("/ssh-keys", app.handlers.ListSSHKeysHandler)
				authGroup.POST("/ssh-keys", app.handlers.AddSSHKeyHandler)
				authGroup.DELETE("/ssh-keys/:ssh_key_id", app.handlers.DeleteSSHKeyHandler)
			}
		}

		deployerGroup := v1.Group("")
		deployerGroup.Use(middlewares.UserMiddleware(app.handlers.tokenManager))
		{
			deployerGroup.GET("/events", app.sseManager.HandleSSE)

			deploymentGroup := deployerGroup.Group("/deployments")
			{
				deploymentGroup.POST("", app.handlers.HandleAsyncDeploy)
				deploymentGroup.GET("", app.handlers.HandleListDeployments)
				deploymentGroup.GET("/:name", app.handlers.HandleGetDeployment)
				deploymentGroup.GET("/:name/kubeconfig", app.handlers.HandleGetKubeconfig)
				deploymentGroup.DELETE("/:name", app.handlers.HandleDeleteDeployment)

				// Node management routes
				deploymentGroup.POST("/:name/nodes", app.handlers.HandleAddNodeToDeployment)
				deploymentGroup.DELETE("/:name/nodes/:node_name", app.handlers.HandleRemoveNodeFromDeployment)
			}

			// TODO: Task routes
			// deployerGroup.GET("/tasks", app.handlers.ListUserTasksHandler)
			// deployerGroup.GET("/tasks/:task_id", app.handlers.GetTaskStatusHandler)

			// Notification routes
			deployerGroup.GET("/notifications", app.handlers.GetNotificationsHandler)
			deployerGroup.PUT("/notifications/:notification_id/read", app.handlers.MarkNotificationReadHandler)
			deployerGroup.PUT("/notifications/read-all", app.handlers.MarkAllNotificationsReadHandler)
			deployerGroup.GET("/notifications/unread-count", app.handlers.GetUnreadNotificationCountHandler)
			deployerGroup.DELETE("/notifications/:notification_id", app.handlers.DeleteNotificationHandler)
		}

	}
	app.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (app *App) StartBackgroundWorkers() {
	go app.handlers.MonthlyInvoicesHandler()
	go app.handlers.TrackUserDebt(app.gridClient)
}

// Run starts the server
func (app *App) Run() error {
	addr := fmt.Sprintf("%s:%s", app.config.Server.Host, app.config.Server.Port)

	app.StartBackgroundWorkers()
	app.httpServer = &http.Server{
		Addr:    addr,
		Handler: app.router,
	}

	log.Info().Msgf("Starting server at http://%s", addr)

	if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Failed to start server")
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the server and worker manager
func (app *App) Shutdown(ctx context.Context) error {
	// First, cancel the app context to signal all components to stop
	if app.appCancel != nil {
		app.appCancel()
	}

	if app.httpServer != nil {
		if err := app.httpServer.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown HTTP server")
		}
	}

	if app.workerManager != nil {
		app.workerManager.Stop()
	}

	if app.sseManager != nil {
		app.sseManager.Stop()
	}

	if app.redis != nil {
		if err := app.redis.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Redis connection")
		}
	}

	if app.db != nil {
		if err := app.db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}

	// app.gridClient.Close()

	return nil
}
