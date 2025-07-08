package app

import (
	"fmt"
	"kubecloud/internal"
	"kubecloud/kubedeployer"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

const (
	KUBECLOUD_KEY = "kubecloud/"
)

type DeployResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) HandleAsyncDeploy(c *gin.Context) {
	var cluster kubedeployer.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request json format"})
		return
	}
	// TODO: add validation

	// create task and add to queue
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := fmt.Sprintf("%v", userID)
	taskID := uuid.New().String()
	task := &internal.DeploymentTask{
		TaskID:    taskID,
		UserID:    id,
		Status:    internal.TaskStatusPending,
		CreatedAt: time.Now(),
		Payload:   cluster,
	}

	if err := h.redis.AddTask(c.Request.Context(), task); err != nil {
		log.Error().Err(err).Str("task_id", taskID).Msg("Failed to add task to Task queue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue deployment task"})
		return
	}

	response := DeployResponse{
		TaskID:    taskID,
		Status:    string(internal.TaskStatusPending),
		Message:   "Deployment task queued successfully",
		CreatedAt: task.CreatedAt,
	}

	c.JSON(http.StatusAccepted, response)
}

// HandleListDeployments lists all deployments for the authenticated user
func (h *Handler) HandleListDeployments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := fmt.Sprintf("%v", userID)
	clusters, err := h.db.ListUserClusters(id)
	if err != nil {
		log.Error().Err(err).Str("user_id", id).Msg("Failed to list user clusters")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployments"})
		return
	}

	// Convert clusters to response format
	deployments := make([]gin.H, 0, len(clusters))
	for _, cluster := range clusters {
		clusterResult, err := cluster.GetClusterResult()
		if err != nil {
			log.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
			continue
		}

		deployments = append(deployments, gin.H{
			"id":           cluster.ID,
			"project_name": cluster.ProjectName,
			"cluster":      clusterResult,
			"created_at":   cluster.CreatedAt,
			"updated_at":   cluster.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"deployments": deployments,
		"count":       len(deployments),
	})
}

// HandleGetDeployment gets a specific deployment by name for the authenticated user
func (h *Handler) HandleGetDeployment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectName := c.Param("name")
	if projectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project name is required"})
		return
	}

	id := fmt.Sprintf("%v", userID)
	cluster, err := h.db.GetClusterByName(id, projectName)
	if err != nil {
		log.Error().Err(err).Str("user_id", id).Str("project_name", projectName).Msg("Failed to get cluster")
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	clusterResult, err := cluster.GetClusterResult()
	if err != nil {
		log.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployment details"})
		return
	}

	response := gin.H{
		"id":           cluster.ID,
		"project_name": cluster.ProjectName,
		"cluster":      clusterResult,
		"created_at":   cluster.CreatedAt,
		"updated_at":   cluster.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) HandleGetKubeconfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectName := c.Param("name")
	if projectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project name is required"})
		return
	}

	id := fmt.Sprintf("%v", userID)
	cluster, err := h.db.GetClusterByName(id, projectName)
	if err != nil {
		log.Error().Err(err).Str("user_id", id).Str("project_name", projectName).Msg("Failed to get cluster")
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	clusterResult, err := cluster.GetClusterResult()
	if err != nil {
		log.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployment details"})
		return
	}

	// Find the leader or master node
	var targetNode *kubedeployer.Node
	for _, node := range clusterResult.Nodes {
		if node.Type == kubedeployer.NodeTypeLeader {
			targetNode = &node
			break
		}
	}

	// Fallback to master if no leader found
	if targetNode == nil {
		for _, node := range clusterResult.Nodes {
			if node.Type == kubedeployer.NodeTypeMaster {
				targetNode = &node
				break
			}
		}
	}

	if targetNode == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No leader or master node found in deployment"})
		return
	}

	// Read SSH private key
	privateKeyBytes, err := os.ReadFile(h.config.SSH.PrivateKeyPath)
	if err != nil {
		log.Error().Err(err).Str("key_path", h.config.SSH.PrivateKeyPath).Msg("Failed to read SSH private key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read SSH configuration"})
		return
	}

	// Try to get kubeconfig via SSH
	kubeconfig, err := h.getKubeconfigViaSSH(string(privateKeyBytes), targetNode)
	if err != nil {
		log.Error().Err(err).Str("node_name", targetNode.Name).Msg("Failed to retrieve kubeconfig via SSH")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve kubeconfig: " + err.Error()})
		return
	}

	c.Header("Content-Type", "application/x-yaml")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-kubeconfig.yaml\"", projectName))
	c.String(http.StatusOK, kubeconfig)
}

func (h *Handler) getKubeconfigViaSSH(privateKey string, node *kubedeployer.Node) (string, error) {
	// Try planetary IP first, then mycelium IP as fallback
	ips := []string{}
	if node.PlanetaryIP != "" {
		ips = append(ips, node.PlanetaryIP)
	}
	if node.MyceliumIP != "" {
		ips = append(ips, node.MyceliumIP)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no valid IP addresses found for node %s", node.Name)
	}

	var lastErr error
	for _, ip := range ips {
		log.Info().Str("ip", ip).Str("node", node.Name).Msg("Attempting SSH connection")

		kubeconfig, err := h.executeSSHCommand(privateKey, ip, "kubectl config view --minify --raw")
		if err != nil {
			log.Warn().Err(err).Str("ip", ip).Str("node", node.Name).Msg("SSH connection failed, trying next IP")
			lastErr = err
			continue
		}

		// Validate that we got a valid kubeconfig
		if strings.Contains(kubeconfig, "apiVersion") && strings.Contains(kubeconfig, "clusters") {
			log.Info().Str("ip", ip).Str("node", node.Name).Msg("Successfully retrieved kubeconfig")
			return kubeconfig, nil
		} else {
			lastErr = fmt.Errorf("invalid kubeconfig received from node %s", node.Name)
		}
	}

	return "", fmt.Errorf("failed to retrieve kubeconfig from any IP address: %v", lastErr)
}

func (h *Handler) executeSSHCommand(privateKey, address, command string) (string, error) {
	key, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("could not parse SSH private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		Timeout: 30 * time.Second,
	}

	// Connect to SSH
	port := "22"
	client, err := ssh.Dial("tcp", net.JoinHostPort(address, port), config)
	if err != nil {
		return "", fmt.Errorf("could not establish SSH connection to %s: %w", address, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("could not create SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("could not execute command '%s': %w, output: %s", command, err, string(output))
	}

	return string(output), nil
}
