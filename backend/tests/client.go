package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"kubecloud/app"
	"kubecloud/kubedeployer"
)

type Client struct {
	httpClient  *http.Client
	accessToken string
	baseURL     string
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
	req := app.RegisterInput{
		Name:            name,
		Email:           email,
		Password:        password,
		ConfirmPassword: confirmPassword,
	}

	resp, err := c.makeRequest("POST", "/user/register", req, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) Login(email, password string) error {
	req := app.LoginInput{
		Email:    email,
		Password: password,
	}

	resp, err := c.makeRequest("POST", "/user/login", req, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	if loginResp.Data.AccessToken == "" {
		return fmt.Errorf("no access token received in login response")
	}

	c.accessToken = loginResp.Data.AccessToken
	return nil
}

func (c *Client) DeployCluster(clusterName string) (string, error) {
	cluster := kubedeployer.Cluster{
		Name:  clusterName,
		Token: "test-token-123",
		Nodes: []kubedeployer.Node{
			{
				Name:     "leader",
				Type:     kubedeployer.NodeTypeLeader,
				CPU:      1,
				Memory:   2 * 1024, // 2 GB
				RootSize: 10240,    // 10 GB
				DiskSize: 10240,    // 10 GB
				EnvVars: map[string]string{
					"SSH_KEY": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDJ1t4Ug8EfykmJwAbYudyYYN/f7dZaVg3KGD2Pz0bd9pajAAASWYrss3h2ctCZWluM6KAt289RMNzxlNUkOMJ9WhCIxqDAwtg05h/J27qlaGCPP8BCEITwqNKsLwzmMZY1UFc+sSUyjd35d3kjtv+rzo2meaReZnUFNPisvxGoygftAE6unqNa7TKonVDS1YXzbpT8XdtCV1Y6ACx+3a82mFR07zgmY4BVOixNBy2Lzpq9KiZTz91Bmjg8dy4xUyWLiTmnye51hEBgUzPprjffZByYSb2Ag9hpNE1AdCGCli/0TbEwFn9iEroh/xmtvZRpux+L0OmO93z5Sz+RLiYXKiYVV5R5XYP8y5eYi48RY2qr82sUl5+WnKhI8nhzayO9yjPEp3aTvR1FdDDj5ocB7qKi47R8FXIuwzZf+kJ7ZYmMSG7N21zDIJrz6JGy9KMi7nX1sqy7NSqX3juAasIjx0IJsE8zv9qokZ83hgcDmTJjnI+YXimelhcHn4M52hU= omar@jarvis",
				},
				NodeID: 150,
			},
			{
				Name:     "master",
				Type:     kubedeployer.NodeTypeMaster,
				CPU:      1,
				Memory:   2 * 1024, // 2 GB
				RootSize: 10240,    // 10 GB
				DiskSize: 10240,    // 10 GB
				EnvVars: map[string]string{
					"SSH_KEY": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDJ1t4Ug8EfykmJwAbYudyYYN/f7dZaVg3KGD2Pz0bd9pajAAASWYrss3h2ctCZWluM6KAt289RMNzxlNUkOMJ9WhCIxqDAwtg05h/J27qlaGCPP8BCEITwqNKsLwzmMZY1UFc+sSUyjd35d3kjtv+rzo2meaReZnUFNPisvxGoygftAE6unqNa7TKonVDS1YXzbpT8XdtCV1Y6ACx+3a82mFR07zgmY4BVOixNBy2Lzpq9KiZTz91Bmjg8dy4xUyWLiTmnye51hEBgUzPprjffZByYSb2Ag9hpNE1AdCGCli/0TbEwFn9iEroh/xmtvZRpux+L0OmO93z5Sz+RLiYXKiYVV5R5XYP8y5eYi48RY2qr82sUl5+WnKhI8nhzayO9yjPEp3aTvR1FdDDj5ocB7qKi47R8FXIuwzZf+kJ7ZYmMSG7N21zDIJrz6JGy9KMi7nX1sqy7NSqX3juAasIjx0IJsE8zv9qokZ83hgcDmTJjnI+YXimelhcHn4M52hU= omar@jarvis",
				},
				NodeID: 152,
			},
			{
				Name:     "worker",
				Type:     kubedeployer.NodeTypeWorker,
				CPU:      1,
				Memory:   2 * 1024, // 2 GB
				RootSize: 10240,    // 10 GB
				DiskSize: 10240,    // 10 GB
				EnvVars: map[string]string{
					"SSH_KEY": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDJ1t4Ug8EfykmJwAbYudyYYN/f7dZaVg3KGD2Pz0bd9pajAAASWYrss3h2ctCZWluM6KAt289RMNzxlNUkOMJ9WhCIxqDAwtg05h/J27qlaGCPP8BCEITwqNKsLwzmMZY1UFc+sSUyjd35d3kjtv+rzo2meaReZnUFNPisvxGoygftAE6unqNa7TKonVDS1YXzbpT8XdtCV1Y6ACx+3a82mFR07zgmY4BVOixNBy2Lzpq9KiZTz91Bmjg8dy4xUyWLiTmnye51hEBgUzPprjffZByYSb2Ag9hpNE1AdCGCli/0TbEwFn9iEroh/xmtvZRpux+L0OmO93z5Sz+RLiYXKiYVV5R5XYP8y5eYi48RY2qr82sUl5+WnKhI8nhzayO9yjPEp3aTvR1FdDDj5ocB7qKi47R8FXIuwzZf+kJ7ZYmMSG7N21zDIJrz6JGy9KMi7nX1sqy7NSqX3juAasIjx0IJsE8zv9qokZ83hgcDmTJjnI+YXimelhcHn4M52hU= omar@jarvis",
				},
				NodeID: 155,
			},
		},
	}

	resp, err := c.makeRequest("POST", "/deployments", cluster, true)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("deploy request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deployResp app.DeployResponse
	if err := json.NewDecoder(resp.Body).Decode(&deployResp); err != nil {
		return "", err
	}

	return deployResp.TaskID, nil
}

func (c *Client) ListenToSSE(taskID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/events", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	sseClient := &http.Client{Timeout: 0}

	resp, err := sseClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SSE connection failed with status %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					if ctx.Err() != nil {
						return nil
					}
					return err
				}
				return nil
			}

			line := scanner.Text()
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimPrefix(line, "data:")
				if data != "" {
					fmt.Printf("SSE Update: %s\n", data)
				}
			}
		}
	}
}

func (c *Client) ListenToSSEWithLogger(taskID string, logFunc func(string, ...interface{})) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/events", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	sseClient := &http.Client{Timeout: 0}

	resp, err := sseClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SSE connection failed with status %d: %s", resp.StatusCode, string(body))
	}

	logFunc("SSE connection established, listening for deployment updates...")

	scanner := bufio.NewScanner(resp.Body)
	for {
		select {
		case <-ctx.Done():
			logFunc("SSE connection timeout reached")
			return nil
		default:
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					if ctx.Err() != nil {
						return nil
					}
					return err
				}
				logFunc("SSE connection ended normally")
				return nil
			}

			line := scanner.Text()
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimPrefix(line, "data:")
				if data != "" {
					logFunc("SSE Update: %s", data)
				}
			}
		}
	}
}

func (c *Client) ListDeployments() ([]interface{}, error) {
	resp, err := c.makeRequest("GET", "/deployments", nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list deployments failed with status %d: %s", resp.StatusCode, string(body))
	}

	var listResp struct {
		Deployments []interface{} `json:"deployments"`
		Count       int           `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode list response: %w", err)
	}

	return listResp.Deployments, nil
}

func (c *Client) GetDeployment(name string) (interface{}, error) {
	resp, err := c.makeRequest("GET", "/deployments/"+name, nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get deployment failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deployment interface{}
	if err := json.NewDecoder(resp.Body).Decode(&deployment); err != nil {
		return nil, fmt.Errorf("failed to decode deployment response: %w", err)
	}

	return deployment, nil
}

func (c *Client) GetKubeconfig(name string) (string, error) {
	resp, err := c.makeRequest("GET", "/deployments/"+name+"/kubeconfig", nil, true)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get kubeconfig failed with status %d: %s", resp.StatusCode, string(body))
	}

	kubeconfig, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read kubeconfig response: %w", err)
	}

	return string(kubeconfig), nil
}

func (c *Client) DeleteDeployment(name string) error {
	resp, err := c.makeRequest("DELETE", "/deployments/"+name, nil, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete deployment failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
