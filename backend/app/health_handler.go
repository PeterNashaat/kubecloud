package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"kubecloud/models/sqlite"

	"github.com/gin-gonic/gin"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
)

type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type HealthChecker func(ctx context.Context) HealthStatus

func (h *Handler) checkDatabase(ctx context.Context) HealthStatus {
	sqliteDB, ok := h.db.(*sqlite.Sqlite)
	if !ok {
		return HealthStatus{Status: "unhealthy", Message: "not a sqlite DB"}
	}
	db, err := sqliteDB.SQLDB()
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	return HealthStatus{Status: "healthy"}
}

func (h *Handler) checkRedis(ctx context.Context) HealthStatus {
	if h.redis == nil || h.redis.Client() == nil {
		return HealthStatus{Status: "unhealthy", Message: "redis client not initialized"}
	}
	err := h.redis.Client().Ping(ctx).Err()
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	return HealthStatus{Status: "healthy"}
}

func (h *Handler) checkGridProxy(ctx context.Context) HealthStatus {
	if strings.TrimSpace(h.config.GridProxyURL) == "" {
		return HealthStatus{Status: "unhealthy", Message: "gridproxy URL not set"}
	}
	client := &http.Client{Timeout: 2 * time.Second}
	healthURL, err := url.JoinPath(h.config.GridProxyURL, "health")
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	resp, err := client.Get(healthURL)
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return HealthStatus{Status: "unhealthy", Message: fmt.Sprintf("unexpected status: %s", resp.Status)}
	}
	return HealthStatus{Status: "healthy"}
}

// tfchainHealthURL converts wss://.../ws to https://.../health
func tfchainHealthURL(tfchainURL string) (string, error) {
	if tfchainURL == "" {
		return "", fmt.Errorf("tfchain_url is empty")
	}
	url := strings.Replace(tfchainURL, "wss://", "https://", 1)
	url = strings.Replace(url, "/ws", "/health", 1)
	return url, nil
}

func (h *Handler) checkTFChainHealth(ctx context.Context) HealthStatus {
	healthURL, err := tfchainHealthURL(h.config.TFChainURL)
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return HealthStatus{Status: "unhealthy", Message: fmt.Sprintf("unexpected status: %s", resp.Status)}
	}
	var health struct {
		Peers           int  `json:"peers"`
		IsSyncing       bool `json:"isSyncing"`
		ShouldHavePeers bool `json:"shouldHavePeers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	if health.IsSyncing || !health.ShouldHavePeers || health.Peers == 0 {
		return HealthStatus{Status: "unhealthy", Message: fmt.Sprintf("syncing: %v, shouldHavePeers: %v, peers: %d", health.IsSyncing, health.ShouldHavePeers, health.Peers)}
	}
	return HealthStatus{Status: "healthy"}
}

// TODO: Cache identity/account to avoid deriving on every request for performance
func (h *Handler) checkSystemAccountBalance(ctx context.Context) HealthStatus {
	if h.substrateClient == nil {
		return HealthStatus{Status: "unhealthy", Message: "client not initialized"}
	}
	identity, idErr := substrate.NewIdentityFromSr25519Phrase(h.config.SystemAccount.Mnemonic)
	if idErr != nil {
		return HealthStatus{Status: "unhealthy", Message: idErr.Error()}
	}
	address := identity.Address()
	account, accErr := substrate.FromAddress(address)
	if accErr != nil {
		return HealthStatus{Status: "unhealthy", Message: accErr.Error()}
	}
	_, err := h.substrateClient.GetBalance(account)
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	return HealthStatus{Status: "healthy"}
}

func (h *Handler) checkActivationService(ctx context.Context) HealthStatus {
	if strings.TrimSpace(h.config.ActivationServiceURL) == "" {
		return HealthStatus{Status: "unhealthy", Message: "activation service URL not set"}
	}
	client := &http.Client{Timeout: 2 * time.Second}
	healthURL, err := url.JoinPath(h.config.ActivationServiceURL, "health")
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	resp, err := client.Get(healthURL)
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return HealthStatus{Status: "unhealthy", Message: fmt.Sprintf("unexpected status: %s", resp.Status)}
	}
	return HealthStatus{Status: "healthy"}
}

func (h *Handler) checkGraphQL(ctx context.Context) HealthStatus {
	_, err := h.graphqlClient.Query("{ __typename }", map[string]interface{}{})
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	return HealthStatus{Status: "healthy"}
}

func (h *Handler) checkFiresquid(ctx context.Context) HealthStatus {
	_, err := h.firesquidClient.Query("{ __typename }", map[string]interface{}{})
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	return HealthStatus{Status: "healthy"}
}

func (h *Handler) HealthHandler(c *gin.Context) {
	ctx := c.Request.Context()
	checks := map[string]HealthChecker{
		"database":               h.checkDatabase,
		"redis":                  h.checkRedis,
		"gridproxy":              h.checkGridProxy,
		"tfchain_health":         h.checkTFChainHealth,
		"system_account_balance": h.checkSystemAccountBalance,
		"activation_service":     h.checkActivationService,
		"graphql":                h.checkGraphQL,
		"firesquid":              h.checkFiresquid,
	}
	resp := make(map[string]HealthStatus)
	for name, check := range checks {
		resp[name] = check(ctx)
	}
	c.JSON(http.StatusOK, resp)
}
