package app

import (
	"fmt"
	"kubecloud/internal"
	"kubecloud/kubedeployer"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"golang.org/x/crypto/ssh"
)

const (
	KUBECLOUD_KEY = "kubecloud/"
)

// DeployResponse represents the response structure for deployment requests
type DeployResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// HandleAsyncDeploy handles asynchronous Kubernetes cluster deployment requests.
//
// This endpoint accepts a JSON payload containing cluster configuration and queues
// a deployment task for processing. The deployment is handled asynchronously via
// Redis task queue and workers.
//
// Request: POST /deployments
// Content-Type: application/json
// Body: kubedeployer.Cluster JSON structure containing:
//   - name: cluster name (becomes project name)
//   - network: optional network configuration
//   - token: optional k3s token
//   - nodes: array of node configurations with CPU, memory, storage specs
//
// Response: 202 Accepted with deployment task information
//
//	{
//	  "task_id": "uuid-string",
//	  "status": "pending",
//	  "message": "Deployment task queued successfully",
//	  "created_at": "2025-01-01T12:00:00Z"
//	}
//
// Authentication: Requires valid user JWT token
// Authorization: User can only deploy to their own account
func (h *Handler) HandleAsyncDeploy(c *gin.Context) {
	var cluster kubedeployer.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request json format"})
		return
	}
	// TODO: add an early validation

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

// HandleListDeployments retrieves all Kubernetes cluster deployments for the authenticated user.
//
// This endpoint returns a paginated list of all deployments associated with the current user,
// including deployment metadata and cluster configuration details.
//
// Request: GET /deployments
// Authentication: Requires valid user JWT token
//
// Response: 200 OK with deployment list
//
//	{
//	  "deployments": [
//	    {
//	      "id": 123,
//	      "project_name": "my-cluster",
//	      "cluster": { /* kubedeployer.Cluster object */ },
//	      "created_at": "2025-01-01T12:00:00Z",
//	      "updated_at": "2025-01-01T12:00:00Z"
//	    }
//	  ],
//	  "count": 1
//	}
//
// Authorization: Returns only deployments owned by the authenticated user
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

// HandleGetDeployment retrieves detailed information for a specific deployment by name.
//
// This endpoint returns comprehensive details about a single Kubernetes cluster deployment,
// including the full cluster configuration and node specifications.
//
// Request: GET /deployments/{name}
// Path Parameters:
//   - name: The project name of the deployment to retrieve
//
// Response: 200 OK with deployment details
//
//	{
//	  "id": 123,
//	  "project_name": "my-cluster",
//	  "cluster": {
//	    "name": "my-cluster",
//	    "nodes": [
//	      {
//	        "name": "leader",
//	        "type": "leader",
//	        "node_id": 1,
//	        "cpu": 2,
//	        "memory": 4096,
//	        "ip": "10.20.0.1",
//	        "mycelium_ip": "400:1234::1",
//	        "planetary_ip": "302:9e63::1"
//	      }
//	    ]
//	  },
//	  "created_at": "2025-01-01T12:00:00Z",
//	  "updated_at": "2025-01-01T12:00:00Z"
//	}
//
// Authentication: Requires valid user JWT token
// Authorization: User can only access their own deployments
// Errors:
//   - 404 Not Found: Deployment doesn't exist or doesn't belong to user
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

// HandleGetKubeconfig retrieves the kubectl configuration file for a deployed cluster.
//
// This endpoint connects via SSH to the leader/master node of the specified cluster
// and downloads the kubeconfig file. The configuration is post-processed to use
// external IP addresses for connectivity from outside the cluster network.
//
// Request: GET /deployments/{name}/kubeconfig
// Path Parameters:
//   - name: The project name of the deployment
//
// Response: 200 OK with kubeconfig YAML file
// Content-Type: application/x-yaml
// Content-Disposition: attachment; filename="{name}-kubeconfig.yaml"
//
// The response body contains a standard kubectl configuration file that can be
// used to connect to the cluster from external clients.
//
// Authentication: Requires valid user JWT token
// Authorization: User can only access kubeconfig for their own deployments
//
// Technical Details:
// - Attempts SSH connection to leader node first, falls back to master nodes
// - Tries multiple kubeconfig retrieval commands for compatibility
// - Uses mycelium IP first, falls back to planetary IP
// - Includes retry logic for transient network issues
//
// Errors:
//   - 404 Not Found: Deployment doesn't exist or doesn't belong to user
//   - 500 Internal Server Error: SSH connection failed or kubeconfig not found
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

	privateKeyBytes, err := os.ReadFile(h.config.SSH.PrivateKeyPath)
	if err != nil {
		log.Error().Err(err).Str("key_path", h.config.SSH.PrivateKeyPath).Msg("Failed to read SSH private key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read SSH configuration"})
		return
	}

	kubeconfig, err := h.getKubeconfigViaSSH(string(privateKeyBytes), targetNode)
	if err != nil {
		log.Error().Err(err).Str("node_name", targetNode.Name).Msg("Failed to retrieve kubeconfig via SSH")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve kubeconfig: " + err.Error()})
		return
	}

	// c.Header("Content-Type", "application/x-yaml")
	// c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-kubeconfig.yaml\"", projectName))
	// c.String(http.StatusOK, kubeconfig)
	c.JSON(http.StatusOK, gin.H{"kubeconfig": kubeconfig})
}

// getKubeconfigViaSSH connects to a cluster node via SSH and retrieves the kubeconfig file.
//
// This method attempts to connect to the specified node using both mycelium and planetary
// IP addresses, trying multiple commands to locate and retrieve the kubeconfig file.
// It includes retry logic and fallback mechanisms for robustness.
//
// Parameters:
//   - privateKey: SSH private key in PEM format for authentication
//   - node: Target node containing IP addresses and metadata
//
// Returns:
//   - string: Post-processed kubeconfig YAML content
//   - error: Connection or retrieval error
//
// The method tries the following kubeconfig locations in order:
// 1. kubectl config view --minify --raw (if kubectl is available)
// 2. /etc/rancher/k3s/k3s.yaml (standard k3s location)
// 3. ~/.kube/config (standard kubectl location)
func (h *Handler) getKubeconfigViaSSH(privateKey string, node *kubedeployer.Node) (string, error) {
	ips := []string{}
	if node.MyceliumIP != "" {
		ips = append(ips, node.MyceliumIP)
	}
	if node.PlanetaryIP != "" {
		ips = append(ips, node.PlanetaryIP)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no valid IP addresses found for node %s", node.Name)
	}

	var lastErr error
	for _, ip := range ips {
		log.Debug().Str("ip", ip).Str("node", node.Name).Msg("Attempting SSH connection")

		commands := []string{
			"kubectl config view --minify --raw",
			"cat /etc/rancher/k3s/k3s.yaml",
			"cat ~/.kube/config",
		}

		var kubeconfig string
		var cmdErr error

		for _, cmd := range commands {
			kubeconfig, cmdErr = h.executeSSHCommand(privateKey, ip, cmd)
			if cmdErr == nil && strings.Contains(kubeconfig, "apiVersion") && strings.Contains(kubeconfig, "clusters") {
				return kubeconfig, nil
			}
			if cmdErr != nil {
				log.Debug().Err(cmdErr).Str("ip", ip).Str("command", cmd).Msg("Command failed, trying next")
			}
		}

		if cmdErr != nil {
			log.Warn().Err(cmdErr).Str("ip", ip).Str("node", node.Name).Msg("All commands failed on this IP, trying next IP")
			lastErr = cmdErr
		} else {
			lastErr = fmt.Errorf("no valid kubeconfig found on node %s at IP %s", node.Name, ip)
		}
	}

	return "", fmt.Errorf("failed to retrieve kubeconfig from any IP address: %v", lastErr)
}

// executeSSHCommand establishes an SSH connection and executes a single command.
//
// This method handles SSH connection establishment with retry logic for transient
// network issues. It uses the provided private key for authentication and connects
// as the root user with a reasonable timeout.
//
// Parameters:
//   - privateKey: SSH private key in PEM format
//   - address: Target IP address (IPv4 or IPv6)
//   - command: Shell command to execute on the remote host
//
// Returns:
//   - string: Combined stdout and stderr output from the command
//   - error: Connection or execution error
//
// Connection details:
// - Uses root user for authentication
// - 15-second connection timeout
// - Up to 3 connection retry attempts
// - Insecure host key verification (for lab environments)
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

	port := "22"
	var client *ssh.Client
	for attempt := 1; attempt <= 3; attempt++ {
		client, err = ssh.Dial("tcp", net.JoinHostPort(address, port), config)
		if err == nil {
			break
		}
		if attempt < 3 {
			log.Debug().Err(err).Str("address", address).Int("attempt", attempt).Msg("SSH connection attempt failed, retrying")
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	if err != nil {
		return "", fmt.Errorf("could not establish SSH connection to %s after 3 attempts: %w", address, err)
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

// HandleDeleteDeployment cancels all ThreeFold Grid contracts and removes a deployment.
//
// This endpoint performs a complete cleanup of a Kubernetes cluster deployment by:
// 1. Canceling all associated ThreeFold Grid contracts (compute, storage, network)
// 2. Removing the deployment record from the database
//
// The operation is atomic - if contract cancellation fails, the database record
// is preserved to allow retry attempts.
//
// Request: DELETE /deployments/{name}
// Path Parameters:
//   - name: The project name of the deployment to delete
//
// Response: 200 OK on successful deletion
//
//	{
//	  "message": "deployment deleted successfully",
//	  "name": "cluster-name"
//	}
//
// Authentication: Requires valid user JWT token
// Authorization: User can only delete their own deployments
//
// Technical Details:
// - Uses grid client's CancelByProjectName to cancel all related contracts
// - Project names are prefixed with "kubecloud/" for contract identification
// - Cancellation includes all node contracts, network contracts, and storage
// - Database cleanup only occurs after successful contract cancellation
//
// Errors:
//   - 404 Not Found: Deployment doesn't exist or doesn't belong to user
//   - 500 Internal Server Error: Contract cancellation or database operation failed
//
// Note: This operation cannot be undone. All cluster data and configurations
// will be permanently destroyed.
func (h *Handler) HandleDeleteDeployment(c *gin.Context) {
	deploymentName := c.Param("name")
	if deploymentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deployment name is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	log.Debug().Str("user_id", userIDStr).Str("deployment_name", deploymentName).Msg("Starting deployment deletion")

	cluster, err := h.db.GetClusterByName(userIDStr, deploymentName)
	if err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Str("deployment_name", deploymentName).Msg("Failed to find deployment")
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	cl, err := cluster.GetClusterResult()
	if err != nil {
		log.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployment details"})
		return
	}

	var contracts []uint64
	for _, node := range cl.Nodes {
		if node.ContractID != 0 {
			contracts = append(contracts, node.ContractID)
		}
	}

	// get user client
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Msg("Failed to parse user ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user ID"})
		return
	}
	user, err := h.db.GetUserByID(userIDInt)
	if err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Msg("Failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	gridClient, err := deployer.NewTFPluginClient(user.Mnemonic, deployer.WithNetwork(h.config.SystemAccount.Network))
	if err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Msg("Failed to create grid client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create grid client"})
		return
	}

	if err := gridClient.CancelByProjectName(cl.Name); err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Str("deployment_name", deploymentName).Msg("Failed to cancel deployment contracts by project name")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel deployment contracts"})
		return
	}

	if err := h.db.DeleteCluster(userIDStr, deploymentName); err != nil {
		log.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to delete deployment from database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove deployment from database"})
		return
	}

	log.Info().Str("user_id", userIDStr).Str("deployment_name", deploymentName).Str("project_name", deploymentName).Msg("Successfully deleted deployment")

	c.JSON(http.StatusOK, gin.H{
		"message": "deployment deleted successfully",
		"name":    deploymentName,
	})
}

// HandleAddNodeToDeployment adds a new node to an existing Kubernetes cluster deployment.
//
// This endpoint accepts a JSON payload containing node configuration and queues
// a deployment task for processing. The new node deployment is handled asynchronously via
// Redis task queue and workers, similar to the main deployment process.
//
// Request: POST /deployments/{name}/nodes
// Content-Type: application/json
// Body: kubedeployer.Cluster JSON structure containing:
//   - name: cluster name (should match the deployment name)
//   - nodes: array with the new node configuration (CPU, memory, storage specs)
//
// Response: 202 Accepted with deployment task information
//
//	{
//	  "task_id": "uuid-string",
//	  "status": "pending",
//	  "message": "Node addition task queued successfully",
//	  "created_at": "2025-01-01T12:00:00Z"
//	}
//
// Authentication: Requires valid user JWT token
// Authorization: User can only modify their own deployments
func (h *Handler) HandleAddNodeToDeployment(c *gin.Context) {
	deploymentName := c.Param("name")
	if deploymentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deployment name is required"})
		return
	}

	var cluster kubedeployer.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request json format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)

	// Verify the deployment exists and belongs to the user
	existingCluster, err := h.db.GetClusterByName(userIDStr, deploymentName)
	if err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Str("deployment_name", deploymentName).Msg("Failed to find deployment")
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	// Ensure cluster name matches deployment name
	cluster.Name = deploymentName

	taskID := uuid.New().String()
	task := &internal.DeploymentTask{
		TaskID:    taskID,
		UserID:    userIDStr,
		Status:    internal.TaskStatusPending,
		CreatedAt: time.Now(),
		Payload:   cluster,
	}

	if err := h.redis.AddTask(c.Request.Context(), task); err != nil {
		log.Error().Err(err).Str("task_id", taskID).Msg("Failed to add node addition task to Task queue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue node addition task"})
		return
	}

	log.Info().Str("user_id", userIDStr).Str("deployment_name", deploymentName).Str("task_id", taskID).Int("cluster_id", existingCluster.ID).Msg("Queued node addition task")

	response := DeployResponse{
		TaskID:    taskID,
		Status:    string(internal.TaskStatusPending),
		Message:   "Node addition task queued successfully",
		CreatedAt: task.CreatedAt,
	}

	c.JSON(http.StatusAccepted, response)
}

// HandleRemoveNodeFromDeployment removes a node from an existing Kubernetes cluster deployment.
//
// This endpoint removes a specific node from the cluster by canceling its contract
// and updating the deployment configuration. The operation is performed synchronously
// since it primarily involves contract cancellation.
//
// Request: DELETE /deployments/{name}/nodes/{node_name}
// Path Parameters:
//   - name: The project name of the deployment
//   - node_name: The name of the node to remove from the cluster
//
// Response: 200 OK with removal confirmation
//
//	{
//	  "message": "Node removed successfully",
//	  "deployment_name": "my-cluster",
//	  "node_name": "worker-1"
//	}
//
// Authentication: Requires valid user JWT token
// Authorization: User can only modify their own deployments
// Errors:
//   - 404 Not Found: Deployment or node doesn't exist or doesn't belong to user
//   - 400 Bad Request: Cannot remove the last remaining node or master/leader node
func (h *Handler) HandleRemoveNodeFromDeployment(c *gin.Context) {
	deploymentName := c.Param("name")
	nodeName := c.Param("node_name")

	if deploymentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deployment name is required"})
		return
	}

	if nodeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node name is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	log.Debug().Str("user_id", userIDStr).Str("deployment_name", deploymentName).Str("node_name", nodeName).Msg("Starting node removal")

	cluster, err := h.db.GetClusterByName(userIDStr, deploymentName)
	if err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Str("deployment_name", deploymentName).Msg("Failed to find deployment")
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	cl, err := cluster.GetClusterResult()
	if err != nil {
		log.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployment details"})
		return
	}

	// Find the node to remove
	var nodeToRemove *kubedeployer.Node
	var nodeIndex int
	for i, node := range cl.Nodes {
		if node.Name == nodeName {
			nodeToRemove = &node
			nodeIndex = i
			break
		}
	}

	if nodeToRemove == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found in deployment"})
		return
	}

	if nodeToRemove.Type == kubedeployer.NodeTypeLeader {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove leader nodes"})
		return
	}

	// TODO: latest master nodes should not be removed
	if len(cl.Nodes) <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove the last remaining node"})
		return
	}

	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Msg("Failed to parse user ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user ID"})
		return
	}

	user, err := h.db.GetUserByID(userIDInt)
	if err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Msg("Failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	gridClient, err := deployer.NewTFPluginClient(user.Mnemonic, deployer.WithNetwork(h.config.SystemAccount.Network))
	if err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Msg("Failed to create grid client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create grid client"})
		return
	}

	var contractsToCancel []uint64
	if nodeToRemove.ContractID != 0 {
		contractsToCancel = append(contractsToCancel, nodeToRemove.ContractID)
	}

	networkWorkload := cl.NetworkWorkload
	if networkContractID, exists := networkWorkload.NodeDeploymentID[nodeToRemove.NodeID]; exists && networkContractID != 0 {
		networkStillInUse := false
		for _, otherNode := range cl.Nodes {
			if otherNode.NodeID == nodeToRemove.NodeID {
				continue
			}
			if otherNetworkContractID, otherExists := networkWorkload.NodeDeploymentID[otherNode.NodeID]; otherExists && otherNetworkContractID != 0 {
				networkStillInUse = true
				break
			}
		}

		if !networkStillInUse {
			contractsToCancel = append(contractsToCancel, networkContractID)
		}
	}

	if len(contractsToCancel) > 0 {
		if err := gridClient.BatchCancelContract(contractsToCancel); err != nil {
			log.Error().Err(err).Uints64("contract_ids", contractsToCancel).Msg("Failed to cancel contracts")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel node and/or network contracts"})
			return
		}
	}

	// Update cluster state - remove the node from the cluster
	updatedNodes := make([]kubedeployer.Node, 0, len(cl.Nodes)-1)
	updatedNodes = append(updatedNodes, cl.Nodes[:nodeIndex]...)
	updatedNodes = append(updatedNodes, cl.Nodes[nodeIndex+1:]...)
	cl.Nodes = updatedNodes

	// Update network workload - remove the node from network if it was canceled
	if networkContractID, exists := cl.NetworkWorkload.NodeDeploymentID[nodeToRemove.NodeID]; exists {
		networkWasCanceled := false
		for _, contractID := range contractsToCancel {
			if contractID == networkContractID {
				networkWasCanceled = true
				break
			}
		}

		delete(cl.NetworkWorkload.NodeDeploymentID, nodeToRemove.NodeID)

		var updatedNetworkNodes []uint32
		for _, nodeID := range cl.NetworkWorkload.Nodes {
			if nodeID != nodeToRemove.NodeID {
				updatedNetworkNodes = append(updatedNetworkNodes, nodeID)
			}
		}
		cl.NetworkWorkload.Nodes = updatedNetworkNodes

		if cl.NetworkWorkload.NodesIPRange != nil {
			delete(cl.NetworkWorkload.NodesIPRange, nodeToRemove.NodeID)
		}
		if cl.NetworkWorkload.MyceliumKeys != nil {
			delete(cl.NetworkWorkload.MyceliumKeys, nodeToRemove.NodeID)
		}
		if cl.NetworkWorkload.Keys != nil {
			delete(cl.NetworkWorkload.Keys, nodeToRemove.NodeID)
		}
		if cl.NetworkWorkload.WGPort != nil {
			delete(cl.NetworkWorkload.WGPort, nodeToRemove.NodeID)
		}

		if networkWasCanceled {
			log.Info().Uint32("node_id", nodeToRemove.NodeID).Msg("Cleaned up network workload data for canceled network contract")
		}
	}

	if err := cluster.SetClusterResult(cl); err != nil {
		log.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to serialize updated cluster result")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update deployment configuration"})
		return
	}

	if err := h.db.UpdateCluster(&cluster); err != nil {
		log.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to update cluster in database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update deployment in database"})
		return
	}

	log.Info().Str("user_id", userIDStr).Str("deployment_name", deploymentName).Str("node_name", nodeName).Uint64("contract_id", nodeToRemove.ContractID).Msg("Successfully removed node from deployment")

	c.JSON(http.StatusOK, gin.H{
		"message":         "Node removed successfully",
		"deployment_name": deploymentName,
		"node_name":       nodeName,
	})
}
