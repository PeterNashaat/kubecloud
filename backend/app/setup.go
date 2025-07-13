package app

import (
	"fmt"
	"kubecloud/internal"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func SetUp(t testing.TB) *App {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()

	configPath := filepath.Join(dir, "config.json")
	dbPath := filepath.Join(dir, "testing.db")
	workflowPath := filepath.Join(dir, "workflow_testing.db")

	config := fmt.Sprintf(`
{
  "server": {
    "host": "localhost",
    "port": "3000"
  },
  "database": {
    "file": "%s"
  },
  "jwt_token": {
    "secret": "secret",
    "access_expiry_minutes": 60,
    "refresh_expiry_hours": 24
  },
  "admins": [],
  "mailSender": {
    "email": "email@domain.com",
    "sendgrid_key": "sendgrid_key",
    "timeout": 60
  },
  "currency": "usd",
  "stripe_secret": "sk_test",
  "tfchain_url": "wss://tfchain.dev.grid.tf/wss",
  "gridproxy_url": "https://gridproxy.dev.grid.tf/",
  "voucher_name_length": 5,
  "terms_and_conditions": {
    "document_link": "https://manual.grid.tf/labs/knowledge_base/terms_conditions_all3",
    "document_hash": "6f2b4109704ba2883d978a7b94e5f295"
  },
  "activation_service_url": "https://activation.dev.grid.tf/activation/activate",
  "system_account": {
    "mnemonic": "winner giant reward damage expose pulse recipe manual brand volcano dry avoid",
    "network": "dev"
  },
  "graphql_url": "https://graphql.dev.grid.tf/graphql",
  "firesquid_url": "https://firesquid.dev.grid.tf/graphql",
  "redis": {
    "host": "localhost",
    "port": 6379,
    "password": "",
    "db": 0
  },
  "grid": {
    "mne": "winner giant reward damage expose pulse recipe manual brand volcano dry avoid",
    "net": "dev"
  },
  "deployer_workers_num": 3,
  "invoice": {
    "name": "Name",
    "address": "Address",
    "governorate": "Cairo Governorate"
  },
  "workflow_db_file": "%s",
  "ssh": {
    "private_key_path": "/tmp/test_id_rsa",
    "public_key_path": "/tmp/test_id_rsa.pub"
  }
}
`, dbPath, workflowPath)

	err := os.WriteFile(configPath, []byte(config), 0644)
	assert.NoError(t, err)

	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	configuration, err := internal.LoadConfig()
	assert.NoError(t, err)

	app, err := NewApp(configuration)
	assert.NoError(t, err)

	return app
}

func GetAuthToken(t *testing.T, app *App, id int, email, username string, isAdmin bool) string {
	tokenPair, err := app.handlers.tokenManager.CreateTokenPair(id, username, isAdmin)
	assert.NoError(t, err)
	return tokenPair.AccessToken
}
