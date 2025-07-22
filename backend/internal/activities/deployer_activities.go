package activities

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/kubedeployer"
	"kubecloud/models"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xmonader/ewf"
)

type ClientConfig struct {
	SSHPublicKey string               `json:"ssh_public_key"`
	Mnemonic     string               `json:"mnemonic"`
	UserID       string               `json:"user_id"`
	Network      string               `json:"network"`
	SSE          *internal.SSEManager `json:"sse"`
	DB           models.DB            `json:"db"`
}

var (
	criticalRetryPolicy = &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ConstantBackoff(5 * time.Second)}
	standardRetryPolicy = &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}
)

func isWorkloadAlreadyDeployedError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "exists: conflict")
}

func DeployNetworkStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := getKubeClient(state)
		if err != nil {
			return err
		}

		cluster, err := getCluster(state)
		if err != nil {
			return err
		}

		if cluster.ProjectName == "" {
			// this is a first not a retry
			if err := cluster.PrepareCluster(config.UserID); err != nil {
				return fmt.Errorf("failed to prepare cluster: %w", err)
			}
		}

		if err := kubeClient.DeployNetwork(ctx, &cluster); err != nil {
			if isWorkloadAlreadyDeployedError(err) {
				return fmt.Errorf("network already deployed for cluster %s: %w", cluster.Name, ewf.ErrFailWorkflowNow)
			}
			return fmt.Errorf("failed to deploy network: %w", err)
		}

		state["cluster"] = cluster
		return nil
	}
}

func UpdateNetworkStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := getKubeClient(state)
		if err != nil {
			return err
		}

		cluster, err := getCluster(state)
		if err != nil {
			return fmt.Errorf("failed to get cluster from state while updating network: %w", err)
		}

		node, err := getFromState[kubedeployer.Node](state, "node")
		if err != nil {
			return err
		}

		node.Name = kubedeployer.GetNodeName(config.UserID, cluster.Name, node.Name)
		cluster.Nodes = append(cluster.Nodes, node)

		if err := kubeClient.DeployNetwork(ctx, &cluster); err != nil {
			return fmt.Errorf("failed to update network: %w", err)
		}

		state["cluster"] = cluster
		state["node"] = node
		return nil
	}
}

func AddNodeStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := getKubeClient(state)
		if err != nil {
			return err
		}

		cluster, err := getCluster(state)
		if err != nil {
			return err
		}

		node, err := getFromState[kubedeployer.Node](state, "node")
		if err != nil {
			return err
		}
		node.OriginalName = node.Name

		if err := node.AssignNodeIP(ctx, kubeClient.GridClient, cluster.Network.Name); err != nil {
			return fmt.Errorf("failed to assign IP for node %s: %w", node.Name, err)
		}

		if err := kubeClient.DeployNode(ctx, &cluster, node, config.SSHPublicKey); err != nil {
			return fmt.Errorf("failed to deploy node %s to existing cluster: %w", node.Name, err)
		}

		state["cluster"] = cluster
		return nil
	}
}

func DeployNodeStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := getKubeClient(state)
		if err != nil {
			return err
		}

		cluster, err := getCluster(state)
		if err != nil {
			return err
		}

		nodeIdx, ok := state["node_index"].(int)
		if !ok {
			nodeIdx = 0
		}
		node := cluster.Nodes[nodeIdx]

		if err := node.AssignNodeIP(ctx, kubeClient.GridClient, cluster.Network.Name); err != nil {
			return fmt.Errorf("failed to assign node IPs: %w", err)
		}
		cluster.Nodes[nodeIdx].IP = node.IP

		if err := kubeClient.DeployNode(ctx, &cluster, node, config.SSHPublicKey); err != nil {
			if isWorkloadAlreadyDeployedError(err) {
				return fmt.Errorf("node already deployed for cluster %s: %w", cluster.Name, ewf.ErrFailWorkflowNow)
			}
		}

		state["cluster"] = cluster
		state["node_index"] = nodeIdx + 1
		return nil
	}
}

func StoreDeploymentStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		cluster, err := getCluster(state)
		if err != nil {
			return err
		}

		config, err := getConfig(state)
		if err != nil {
			return err
		}

		dbCluster := &models.Cluster{
			ProjectName: cluster.ProjectName,
		}

		if err := dbCluster.SetClusterResult(cluster); err != nil {
			return fmt.Errorf("failed to set cluster result: %w", err)
		}

		existingCluster, err := config.DB.GetClusterByName(config.UserID, cluster.ProjectName)
		if err != nil { // cluster not found, create a new one
			if err := config.DB.CreateCluster(config.UserID, dbCluster); err != nil {
				return fmt.Errorf("failed to create cluster in database: %w", err)
			}
		} else { // cluster exists, update it
			existingCluster.Result = dbCluster.Result
			if err := config.DB.UpdateCluster(&existingCluster); err != nil {
				return fmt.Errorf("failed to update cluster in database: %w", err)
			}
		}

		return nil
	}
}

func CancelDeploymentStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		kubeClient, err := getKubeClient(state)
		if err != nil {
			return err
		}

		projectName, ok := state["project_name"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'project_name' in state")
		}

		if err := kubeClient.CancelCluster(ctx, projectName); err != nil {
			return fmt.Errorf("failed to cancel deployment: %w", err)
		}

		return nil
	}
}

func RemoveClusterFromDBStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return err
		}

		projectName, ok := state["project_name"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'project_name' in state")
		}

		if err := config.DB.DeleteCluster(config.UserID, projectName); err != nil {
			return fmt.Errorf("failed to delete cluster from database: %w", err)
		}

		return nil
	}
}

func RemoveDeploymentNodeStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := getKubeClient(state)
		if err != nil {
			return err
		}

		existingCluster, err := getCluster(state)
		if err != nil {
			return err
		}

		nodeName, ok := state["node_name"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'node_name' in state")
		}

		nodeName = kubedeployer.GetNodeName(config.UserID, existingCluster.Name, nodeName)

		if err := kubeClient.RemoveNode(ctx, &existingCluster, nodeName); err != nil {
			return fmt.Errorf("failed to remove node %s from existing cluster: %w", nodeName, err)
		}

		state["cluster"] = existingCluster
		return nil
	}
}

func NewDynamicDeployWorkflowTemplate(engine *ewf.Engine, wfName string, nodesNum int) {
	steps := []ewf.Step{
		{Name: StepDeployNetwork, RetryPolicy: criticalRetryPolicy},
	}

	for i := 0; i < nodesNum; i++ {
		stepName := fmt.Sprintf("deploy_node_%d", i) // TODO: should be cleaned
		engine.Register(stepName, DeployNodeStep())
		steps = append(steps, ewf.Step{Name: stepName, RetryPolicy: criticalRetryPolicy})
	}

	steps = append(steps, ewf.Step{Name: StepStoreDeployment, RetryPolicy: standardRetryPolicy})

	workflow := BaseWFTemplate
	workflow.Steps = steps

	engine.RegisterTemplate(wfName, &workflow)
}

func validateConfig(config ClientConfig) error {
	if config.SSHPublicKey == "" {
		return fmt.Errorf("missing SSH public key in config")
	}
	if config.Mnemonic == "" {
		return fmt.Errorf("missing mnemonic in config")
	}
	if config.UserID == "" {
		return fmt.Errorf("missing user ID in config")
	}
	if config.Network == "" {
		return fmt.Errorf("missing network in config")
	}
	return nil
}

func SetupClient(ctx context.Context, wf *ewf.Workflow) {
	config, ok := wf.State["config"].(ClientConfig)
	if !ok {
		log.Error().Msg("Missing or invalid 'config' in workflow state")
		return
	}

	if err := validateConfig(config); err != nil {
		log.Error().Err(err).Msg("Invalid workflow configuration")
		return
	}

	kubeClient, err := kubedeployer.NewClient(config.Mnemonic, config.Network)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create kubeclient")
		return
	}

	wf.State["kubeclient"] = kubeClient
}

func CloseClient(ctx context.Context, wf *ewf.Workflow, err error) {
	if kubeClient, ok := wf.State["kubeclient"].(*kubedeployer.Client); ok {
		kubeClient.Close()
		delete(wf.State, "kubeclient")
	} else {
		log.Warn().Msg("No kubeclient found in workflow state to close")
	}

	if err != nil {
		log.Error().Err(err).Str("workflow_name", wf.Name).Msg("Workflow completed with error")
	} else {
		log.Info().Str("workflow_name", wf.Name).Msg("Workflow completed successfully")
	}
}

func NotifyUser(ctx context.Context, wf *ewf.Workflow, err error) {
	config, ok := wf.State["config"].(ClientConfig)
	if !ok {
		log.Error().Msg("Missing or invalid 'config' in workflow state")
		return
	}

	notificationData := map[string]interface{}{
		"type":    "workflow_update",
		"message": "Workflow failed",
	}

	if err != nil {
		notificationData["data"] = map[string]interface{}{"name": wf.Name, "error": err.Error()}
	} else {
		cluster, clusterErr := getCluster(wf.State)
		if clusterErr != nil {
			notificationData = map[string]interface{}{
				"type":    "workflow_update",
				"message": "Workflow completed",
				"data":    map[string]interface{}{"name": wf.Name, "error": false},
			}
		} else {
			notificationData = map[string]interface{}{
				"type":    "workflow_update",
				"message": "Workflow completed successfully",
				"data":    map[string]interface{}{"name": wf.Name, "cluster": cluster, "error": false},
			}
		}
	}

	config.SSE.Notify(config.UserID, "workflow_update", notificationData)
}

var BaseWFTemplate = ewf.WorkflowTemplate{
	BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
		func(ctx context.Context, w *ewf.Workflow) {
			log.Info().Str("workflow_name", w.Name).Msg("Starting workflow")
		},
		SetupClient,
	},
	AfterWorkflowHooks: []ewf.AfterWorkflowHook{
		NotifyUser,
		CloseClient,
	},
	BeforeStepHooks: []ewf.BeforeStepHook{
		func(ctx context.Context, w *ewf.Workflow, step *ewf.Step) {
			log.Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Starting step")
		},
	},
	AfterStepHooks: []ewf.AfterStepHook{
		func(ctx context.Context, w *ewf.Workflow, step *ewf.Step, err error) {
			if err != nil {
				log.Error().Err(err).Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step failed")
			} else {
				log.Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step completed successfully")
			}
		},
	},
}

func registerDeploymentActivities(engine *ewf.Engine) {
	engine.Register(StepDeployNetwork, DeployNetworkStep())
	engine.Register(StepDeployNode, DeployNodeStep())
	engine.Register(StepRemoveCluster, CancelDeploymentStep())
	engine.Register(StepAddNode, AddNodeStep())
	engine.Register(StepUpdateNetwork, UpdateNetworkStep())
	engine.Register(StepRemoveNode, RemoveDeploymentNodeStep())
	engine.Register(StepStoreDeployment, StoreDeploymentStep())
	engine.Register(StepRemoveClusterFromDB, RemoveClusterFromDBStep())

	deleteWFTemplate := BaseWFTemplate
	deleteWFTemplate.Steps = []ewf.Step{
		{Name: StepRemoveCluster, RetryPolicy: standardRetryPolicy},
		{Name: StepRemoveClusterFromDB, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(WorkflowDeleteCluster, &deleteWFTemplate)

	addNodeWFTemplate := BaseWFTemplate
	addNodeWFTemplate.Steps = []ewf.Step{
		{Name: StepUpdateNetwork, RetryPolicy: criticalRetryPolicy},
		{Name: StepAddNode, RetryPolicy: standardRetryPolicy},
		{Name: StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(WorkflowAddNode, &addNodeWFTemplate)

	removeNodeWFTemplate := BaseWFTemplate
	removeNodeWFTemplate.Steps = []ewf.Step{
		{Name: StepRemoveNode, RetryPolicy: standardRetryPolicy},
		{Name: StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(WorkflowRemoveNode, &removeNodeWFTemplate)
}

func getFromState[T any](state ewf.State, key string) (T, error) {
	value, ok := state[key].(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("missing or invalid '%s' in state", key)
	}
	return value, nil
}

func getKubeClient(state ewf.State) (*kubedeployer.Client, error) {
	return getFromState[*kubedeployer.Client](state, "kubeclient")
}

func getCluster(state ewf.State) (kubedeployer.Cluster, error) {
	return getFromState[kubedeployer.Cluster](state, "cluster")
}

func getConfig(state ewf.State) (ClientConfig, error) {
	return getFromState[ClientConfig](state, "config")
}
