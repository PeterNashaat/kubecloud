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

func addFlags() error {
	rootCmd.PersistentFlags().StringP("config", "c", "./config.json", "Path to the configuration file (default: ./config.json)")
	if err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")); err != nil {
		return fmt.Errorf("failed to bind config flag: %w", err)
	}

	// === Server ===
	if err := bindStringFlag(rootCmd, "server.host", "", "Server host"); err != nil {
		return fmt.Errorf("failed to bind server.host flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "server.port", "", "Server port"); err != nil {
		return fmt.Errorf("failed to bind server.port flag: %w", err)
	}

	// === Database ===
	if err := bindStringFlag(rootCmd, "database.file", "", "Database file path"); err != nil {
		return fmt.Errorf("failed to bind database.file flag: %w", err)
	}

	// === JWT Token ===
	if err := bindStringFlag(rootCmd, "jwt_token.secret", "", "JWT secret"); err != nil {
		return fmt.Errorf("failed to bind jwt_token.secret flag: %w", err)
	}
	if err := bindIntFlag(rootCmd, "jwt_token.access_expiry_minutes", 60, "Access token expiry (minutes)"); err != nil {
		return fmt.Errorf("failed to bind jwt_token.access_expiry_minutes flag: %w", err)
	}
	if err := bindIntFlag(rootCmd, "jwt_token.refresh_expiry_hours", 24, "Refresh token expiry (hours)"); err != nil {
		return fmt.Errorf("failed to bind jwt_token.refresh_expiry_hours flag: %w", err)
	}

	// === Admins ===
	if err := bindStringFlag(rootCmd, "admins", "", "Comma-separated list of admin emails"); err != nil {
		return fmt.Errorf("failed to bind admins flag: %w", err)
	}

	// === Mail Sender ===
	if err := bindStringFlag(rootCmd, "mailSender.email", "", "Sender email"); err != nil {
		return fmt.Errorf("failed to bind mailSender.email flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "mailSender.sendgrid_key", "", "SendGrid API key"); err != nil {
		return fmt.Errorf("failed to bind mailSender.sendgrid_key flag: %w", err)
	}
	if err := bindIntFlag(rootCmd, "mailSender.timeout", 60, "Send timeout (seconds)"); err != nil {
		return fmt.Errorf("failed to bind mailSender.timeout flag: %w", err)
	}

	// === Stripe ===
	if err := bindStringFlag(rootCmd, "currency", "", "Currency (e.g., USD)"); err != nil {
		return fmt.Errorf("failed to bind currency flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "stripe_secret", "", "Stripe secret"); err != nil {
		return fmt.Errorf("failed to bind stripe_secret flag: %w", err)
	}

	// === Voucher ===
	if err := bindIntFlag(rootCmd, "voucher_name_length", 6, "Voucher name length"); err != nil {
		return fmt.Errorf("failed to bind voucher_name_length flag: %w", err)
	}

	// === URLs ===
	if err := bindStringFlag(rootCmd, "gridproxy_url", "", "GridProxy URL"); err != nil {
		return fmt.Errorf("failed to bind gridproxy_url flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "tfchain_url", "", "TFChain URL"); err != nil {
		return fmt.Errorf("failed to bind tfchain_url flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "activation_service_url", "", "Activation Service URL"); err != nil {
		return fmt.Errorf("failed to bind activation_service_url flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "graphql_url", "", "GraphQL URL"); err != nil {
		return fmt.Errorf("failed to bind graphql_url flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "firesquid_url", "", "Firesquid URL"); err != nil {
		return fmt.Errorf("failed to bind firesquid_url flag: %w", err)
	}

	// === Terms and Conditions ===
	if err := bindStringFlag(rootCmd, "terms_and_conditions.document_link", "", "Terms document link"); err != nil {
		return fmt.Errorf("failed to bind terms_and_conditions.document_link flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "terms_and_conditions.document_hash", "", "Terms document hash"); err != nil {
		return fmt.Errorf("failed to bind terms_and_conditions.document_hash flag: %w", err)
	}

	// === System Account ===
	if err := bindStringFlag(rootCmd, "system_account.mnemonic", "", "System account mnemonic"); err != nil {
		return fmt.Errorf("failed to bind system_account.mnemonic flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "system_account.network", "", "System account network"); err != nil {
		return fmt.Errorf("failed to bind system_account.network flag: %w", err)
	}

	// === Redis ===
	if err := bindStringFlag(rootCmd, "redis.host", "", "Redis host"); err != nil {
		return fmt.Errorf("failed to bind redis.host flag: %w", err)
	}
	if err := bindIntFlag(rootCmd, "redis.port", 6379, "Redis port"); err != nil {
		return fmt.Errorf("failed to bind redis.port flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "redis.password", "", "Redis password"); err != nil {
		return fmt.Errorf("failed to bind redis.password flag: %w", err)
	}
	if err := bindIntFlag(rootCmd, "redis.db", 0, "Redis DB number"); err != nil {
		return fmt.Errorf("failed to bind redis.db flag: %w", err)
	}

	// === Grid ===
	if err := bindStringFlag(rootCmd, "grid.mnemonic", "", "Grid mnemonic"); err != nil {
		return fmt.Errorf("failed to bind grid.mnemonic flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "grid.net", "", "Grid network"); err != nil {
		return fmt.Errorf("failed to bind grid.net flag: %w", err)
	}

	// === Deployer Workers ===
	if err := bindIntFlag(rootCmd, "deployer_workers_num", 1, "Number of deployer workers"); err != nil {
		return fmt.Errorf("failed to bind deployer_workers_num flag: %w", err)
	}

	// === Invoice ===
	if err := bindStringFlag(rootCmd, "invoice.name", "", "Invoice company name"); err != nil {
		return fmt.Errorf("failed to bind invoice.name flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "invoice.address", "", "Invoice address"); err != nil {
		return fmt.Errorf("failed to bind invoice.address flag: %w", err)
	}
	if err := bindStringFlag(rootCmd, "invoice.governorate", "", "Invoice governorate"); err != nil {
		return fmt.Errorf("failed to bind invoice.governorate flag: %w", err)
	}

	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	viper.SetEnvPrefix("kubecloud") // Prefix for environment variables
	viper.AutomaticEnv()            // Automatically bind environment variables

	// Map environment variables to their corresponding keys
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := addFlags(); err != nil {
		log.Fatal().Err(err).Msg("Failed to add flags")
	}
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

func bindStringFlag(cmd *cobra.Command, key, defaultVal, usage string) error {
	cmd.PersistentFlags().String(key, defaultVal, usage)
	return viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key))
}

func bindIntFlag(cmd *cobra.Command, key string, defaultVal int, usage string) error {
	cmd.PersistentFlags().Int(key, defaultVal, usage)
	if err := viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		return fmt.Errorf("failed to bind flag %s: %w", key, err)
	}

	if err := viper.BindEnv(key); err != nil {
		return fmt.Errorf("failed to bind environment variable for flag %s: %w", key, err)
	}
	// Ensure the value is set as an integer in viper
	viper.Set(key, viper.GetInt(key))

	return nil
}
