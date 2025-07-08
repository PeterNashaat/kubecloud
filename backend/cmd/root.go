package cmd

import (
	"context"
	"fmt"
	"kubecloud/app"
	"kubecloud/internal"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func addFlags() {
	rootCmd.PersistentFlags().StringP("config", "c", "./config.json", "Path to the configuration file (default: ./config.json)")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	// === Server ===
	bindStringFlag(rootCmd, "server.host", "", "Server host")
	bindStringFlag(rootCmd, "server.port", "", "Server port")

	// === Database ===
	bindStringFlag(rootCmd, "database.file", "", "Database file path")

	// === JWT Token ===
	bindStringFlag(rootCmd, "token.secret", "", "JWT secret")
	bindIntFlag(rootCmd, "token.access_expiry_minutes", 60, "Access token expiry (minutes)")
	bindIntFlag(rootCmd, "token.refresh_expiry_hours", 24, "Refresh token expiry (hours)")

	// === Admins ===
	bindStringFlag(rootCmd, "admins", "", "Comma-separated list of admin emails")

	// === Mail Sender ===
	bindStringFlag(rootCmd, "mailSender.email", "", "Sender email")
	bindStringFlag(rootCmd, "mailSender.sendgrid_key", "", "SendGrid API key")
	bindIntFlag(rootCmd, "mailSender.timeout", 60, "Send timeout (seconds)")

	// === Stripe ===
	bindStringFlag(rootCmd, "currency", "", "Currency (e.g., USD)")
	bindStringFlag(rootCmd, "stripe_secret", "", "Stripe secret")

	// === Voucher ===
	bindIntFlag(rootCmd, "voucher_name_length", 6, "Voucher name length")

	// === URLs ===
	bindStringFlag(rootCmd, "gridproxy_url", "", "GridProxy URL")
	bindStringFlag(rootCmd, "tfchain_url", "", "TFChain URL")
	bindStringFlag(rootCmd, "activation_service_url", "", "Activation Service URL")
	bindStringFlag(rootCmd, "graphql_url", "", "GraphQL URL")
	bindStringFlag(rootCmd, "firesquid_url", "", "Firesquid URL")

	// === Terms and Conditions ===
	bindStringFlag(rootCmd, "terms_and_conditions.document_link", "", "Terms document link")
	bindStringFlag(rootCmd, "terms_and_conditions.document_hash", "", "Terms document hash")

	// === System Account ===
	bindStringFlag(rootCmd, "system_account.mnemonic", "", "System account mnemonic")
	bindStringFlag(rootCmd, "system_account.network", "", "System account network")

	// === Redis ===
	bindStringFlag(rootCmd, "redis.host", "", "Redis host")
	bindIntFlag(rootCmd, "redis.port", 6379, "Redis port")
	bindStringFlag(rootCmd, "redis.password", "", "Redis password")
	bindIntFlag(rootCmd, "redis.db", 0, "Redis DB number")

	// === Grid ===
	bindStringFlag(rootCmd, "grid.mnemonic", "", "Grid mnemonic")
	bindStringFlag(rootCmd, "grid.net", "", "Grid network")

	// === Deployer Workers ===
	bindIntFlag(rootCmd, "deployer_workers_num", 1, "Number of deployer workers")

	// === Invoice ===
	bindStringFlag(rootCmd, "invoice.name", "", "Invoice company name")
	bindStringFlag(rootCmd, "invoice.address", "", "Invoice address")
	bindStringFlag(rootCmd, "invoice.governorate", "", "Invoice governorate")
}

func init() {
	cobra.OnInitialize(initConfig)

	viper.SetEnvPrefix("kubecloud") // Prefix for environment variables
	viper.AutomaticEnv()            // Automatically bind environment variables

	// Map environment variables to their corresponding keys
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	addFlags()
}

func initConfig() {
	configFile := viper.GetString("config")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("json")
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Msg("No configuration file found, using defaults")
	}
}

var rootCmd = &cobra.Command{
	Use:   "KubeCloud",
	Short: "Deploy secure, decentralized Kubernetes clusters on TFGrid with Mycelium networking and QSFS storage.",
	Long: `KubeCloud is a CLI tool that helps you deploy and manage Kubernetes clusters on the decentralized TFGrid.

It supports:
- GPU and dedicated nodes for high-performance workloads
- Built-in storage using QSFS with backup and restore
- Private networking with Mycelium (no public IPs needed)
- Web gateway (WebGW) access to expose services
- Usage-based billing with USD pricing set by farmers
- Secure access control through Mycelium whitelisting
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := internal.LoadConfig()
		if err != nil {
			log.Error().Err(err).Msg("Failed to read configurations")
			return fmt.Errorf("failed to read configuration: %w", err)
		}

		app, err := app.NewApp(config)
		if err != nil {
			return fmt.Errorf("failed to create new app: %w", err)
		}

		return gracefulShutdown(app)
	},
}

func gracefulShutdown(app *app.App) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info().Msg("Starting KubeCloud server")

		if err := app.Run(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Failed to start server")
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info().Msg("Shutting down...")
	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
		return err
	}

	log.Info().Msg("Server gracefully stopped.")
	return nil
}

func Execute() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	err := rootCmd.Execute()
	if err != nil {
		log.Error().Err(err).Msg("Command execution failed")
		os.Exit(1)
	}
}

func bindStringFlag(cmd *cobra.Command, key, defaultVal, usage string) {
	cmd.PersistentFlags().String(key, defaultVal, usage)
	viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key))
}

func bindIntFlag(cmd *cobra.Command, key string, defaultVal int, usage string) {
	cmd.PersistentFlags().Int(key, defaultVal, usage)
	viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key))
	viper.BindEnv(key)
	viper.Set(key, viper.GetInt(key))
}
