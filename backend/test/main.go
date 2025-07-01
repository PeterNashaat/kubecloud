package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

type Client struct {
	httpClient  *http.Client
	accessToken string
	baseURL     string
}

// Request/Response structs
type RegisterRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

type DeploymentRequest struct {
	BlueprintID string                 `json:"blueprint_id"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type DeploymentResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type DeployResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskStatusResponse struct {
	TaskID      string                 `json:"task_id"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	BlueprintID string                 `json:"blueprint_id"`
	Parameters  map[string]interface{} `json:"parameters"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

type Blueprint struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "http://localhost:8080/api/v1",
	}
}

func (c *Client) makeRequest(method, endpoint string, body interface{}, needsAuth bool) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if needsAuth && c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	return c.httpClient.Do(req)
}

func (c *Client) Register(name, email, password, confirmPassword string) error {
	req := RegisterRequest{
		Name:            name,
		Email:           email,
		Password:        password,
		ConfirmPassword: confirmPassword,
	}

	resp, err := c.makeRequest("POST", "/user/register", req, false)
	if err != nil {
		return fmt.Errorf("register request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("✅ Registration successful for %s!\n", name)
	return nil
}

func (c *Client) Login(email, password string) error {
	req := LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := c.makeRequest("POST", "/user/login", req, false)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	c.accessToken = loginResp.AccessToken
	fmt.Printf("✅ Login successful. Welcome, %s!\n", loginResp.User.Name)
	return nil
}

func (c *Client) DeployClusterWithSSE(clusterName string) (string, error) {
	cluster := workloads.K8sCluster{
		Master: &workloads.K8sNode{
			VM: &workloads.VM{
				Name:        clusterName,
				NodeID:      155,
				CPU:         1,
				MemoryMB:    1024,
				NetworkName: fmt.Sprintf("%s_network", clusterName),
				Flist:       "https://hub.grid.tf/tf-official-apps/threefolddev-k3s-v1.31.0.flist",
				Entrypoint:  "/sbin/zinit init",
				EnvVars: map[string]string{
					"SSH_KEY":           "",
					"K3S_TOKEN":         "46YIi6OyKx",
					"K3S_DATA_DIR":      "/mnt/data",
					"K3S_FLANNEL_IFACE": "eth0",
					"K3S_NODE_NAME":     clusterName,
					"K3S_URL":           "",
				},
				RootfsSizeMB: 10 * 1024,
				Mounts:       []workloads.Mount{},
			},
			DiskSizeGB: 10,
		},
		Workers:     []workloads.K8sNode{},
		Token:       "46YIi6OyKx",
		NetworkName: fmt.Sprintf("%s_network", clusterName),
	}

	resp, err := c.makeRequest("POST", "/deploy", cluster, true)
	if err != nil {
		return "", fmt.Errorf("deploy request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("deploy request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deployResp DeployResponse
	if err := json.NewDecoder(resp.Body).Decode(&deployResp); err != nil {
		return "", fmt.Errorf("failed to decode deploy response: %w", err)
	}

	fmt.Printf("✅ Cluster deployment initiated successfully!\n")
	fmt.Printf("   Task ID: %s\n", deployResp.TaskID)
	return deployResp.TaskID, nil
}

func (c *Client) ListenToSSE(taskID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Received interrupt signal, closing SSE connection...")
		cancel()
	}()

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/events", nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	sseClient := &http.Client{
		Timeout: 0,
	}

	resp, err := sseClient.Do(req)
	if err != nil {
		return fmt.Errorf("SSE request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SSE connection failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("✅ SSE connection established (Status: %d)\n", resp.StatusCode)
	fmt.Printf("🔄 Listening for real-time updates for task %s...\n", taskID)
	fmt.Println("Press Ctrl+C to stop listening")

	scanner := bufio.NewScanner(resp.Body)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n✅ SSE connection closed gracefully")
			return nil
		default:
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					if ctx.Err() != nil {
						fmt.Println("\n✅ SSE connection closed gracefully")
						return nil
					}
					return fmt.Errorf("error reading SSE stream: %w", err)
				}
				fmt.Println("\n🔌 SSE connection closed by server")
				return nil
			}

			line := scanner.Text()
			if line != "" {
				fmt.Printf("🔍 Raw SSE line: %q\n", line)
			}

			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data != "" {
					fmt.Printf("📡 SSE Update: %s\n", data)
				}
			}
		}
	}
}

func main() {
	client := NewClient()

	// Test registration first
	// fmt.Println("\n1. Testing user registration...")
	// if err := client.Register("Test User", "testuser@example.com", "testpassword123", "testpassword123"); err != nil {
	// 	fmt.Printf("⚠️  Registration failed (might already exist): %v\n", err)
	// }

	fmt.Println("\n2. Logging in...")
	if err := client.Login("testuser@example.com", "testpassword123"); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	fmt.Println("\n3. Deploying k8s cluster...")
	taskID, err := client.DeployClusterWithSSE("my_k8s_cluster3")
	if err != nil {
		log.Fatalf("Deployment failed: %v", err)
	}

	fmt.Printf("\n4. Listening for deployment updates (Task ID: %s)...\n", taskID)
	if err := client.ListenToSSE(taskID); err != nil {
		log.Printf("SSE listening ended: %v", err)
	}

	fmt.Println("\n🏁 Test client finished!")
}
