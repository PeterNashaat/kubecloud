package activities

import (
	"context"
	"fmt"
	"kubecloud/kubedeployer"

	"github.com/rs/zerolog/log"
	"github.com/xmonader/ewf"
)

func ensureClient(ctx context.Context, state ewf.State) (*kubedeployer.Client, error) {
	client, ok := state["kubeclient"].(*kubedeployer.Client)
	if ok {
		log.Info().Msg("Found existing kubedeployer client in state")
		return client, nil
	}

	// create client again if not found in state
	mnemonic, ok := state["mnemonic"].(string)
	if !ok {
		log.Error().Msg("Missing 'mnemonic' in state")
		return nil, fmt.Errorf("missing 'mnemonic' in state")
	}

	gridNet, ok := state["grid_net"].(string)
	if !ok {
		log.Error().Msg("Missing 'grid_net' in state")
		return nil, fmt.Errorf("missing 'grid_net' in state")
	}

	masterPubKey, ok := state["master_pub_key"].(string)
	if !ok {
		log.Error().Msg("Missing 'master_pub_key' in state")
		return nil, fmt.Errorf("missing 'master_pub_key' in state")
	}

	userID, ok := state["user_id"].(string)
	if !ok {
		log.Error().Msg("Missing 'user_id' in state")
		return nil, fmt.Errorf("missing 'user_id' in state")
	}

	kubeClient, err := kubedeployer.NewClient(mnemonic, gridNet, masterPubKey, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create kubedeployer client")
		return nil, fmt.Errorf("failed to create kubedeployer client: %w", err)
	}

	log.Info().Msg("Successfully created kubedeployer client")
	return kubeClient, nil
}

func SetClientStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		kubeClient, err := ensureClient(ctx, state)
		if err != nil {
			log.Error().Err(err).Msg("Failed to ensure kubedeployer client")
			return fmt.Errorf("failed to ensure kubedeployer client: %w", err)
		}

		state["kubeclient"] = kubeClient
		return nil
	}
}

func DeployNetworkStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		log.Info().Msg("Deploying network for cluster")
		kubeClient, err := ensureClient(ctx, state)
		if err != nil {
			log.Error().Err(err).Msg("Failed to ensure kubedeployer client in deploy network step")
			return fmt.Errorf("failed to ensure kubedeployer client: %w", err)
		}

		cluster, ok := state["cluster"].(kubedeployer.Cluster)
		if !ok {
			log.Error().Msg("Missing or invalid 'cluster' in state")
			return fmt.Errorf("missing or invalid 'cluster' in state")
		}

		log.Info().Str("cluster_name", cluster.Name).Msg("Preparing cluster")
		if err := cluster.PrepareCluster(kubeClient.UserID); err != nil {
			log.Error().Err(err).Str("cluster_name", cluster.Name).Msg("Failed to prepare cluster")
			return fmt.Errorf("failed to prepare cluster: %w", err)
		}

		log.Info().Str("cluster_name", cluster.Name).Msg("Starting network deployment")
		if err := kubeClient.DeployNetwork(ctx, &cluster); err != nil {
			log.Error().Err(err).Str("cluster_name", cluster.Name).Msg("Failed to deploy network")
			return fmt.Errorf("failed to deploy network: %w", err)
		}

		state["cluster"] = cluster
		log.Info().Str("cluster_name", cluster.Name).Msg("Network deployed successfully for cluster")
		return nil
	}
}

func DeployNodesStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		log.Info().Msg("Starting node deployment for cluster")
		kubeClient, err := ensureClient(ctx, state)
		if err != nil {
			log.Error().Err(err).Msg("Failed to ensure kubedeployer client in deploy nodes step")
			return fmt.Errorf("failed to ensure kubedeployer client: %w", err)
		}

		cluster, ok := state["cluster"].(kubedeployer.Cluster)
		if !ok {
			log.Error().Msg("Missing or invalid 'cluster' in state")
			return fmt.Errorf("missing or invalid 'cluster' in state")
		}

		log.Info().Str("cluster_name", cluster.Name).Int("node_count", len(cluster.Nodes)).Msg("Assigning node IPs")
		if err := kubeClient.AssignNodeIPs(ctx, &cluster); err != nil {
			log.Error().Err(err).Str("cluster_name", cluster.Name).Msg("Failed to assign node IPs")
			return fmt.Errorf("failed to assign node IPs: %w", err)
		}

		for idx, node := range cluster.Nodes {
			if node.ContractID != 0 {
				log.Info().Str("node_name", node.Name).Uint64("contract_id", node.ContractID).Msg("Node deployment already exists, skipping")
				continue
			}

			// Deploy and update the node on cluster.Nodes
			if err := kubeClient.DeployNode(ctx, &cluster, node); err != nil {
				log.Error().Err(err).Str("node_name", node.Name).Int("node_index", idx).Msg("Failed to deploy node")
				return fmt.Errorf("failed to deploy node %s (index %d): %w", node.Name, idx, err)
			}
			log.Info().Str("node_name", node.Name).Msg("Node deployed successfully")
		}

		log.Info().Str("cluster_name", cluster.Name).Int("node_count", len(cluster.Nodes)).Msg("All nodes deployed successfully for cluster")
		return nil
	}
}

func StoreDeploymentStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		return nil
	}
}

func NotifyUserStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		return nil
	}
}

func CancelDeploymentStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		kubeClient, err := ensureClient(ctx, state)
		if err != nil {
			log.Error().Err(err).Msg("Failed to ensure kubedeployer client in cancel deployment step")
			return fmt.Errorf("failed to ensure kubedeployer client: %w", err)
		}

		projectName, ok := state["project_name"].(string)
		if !ok {
			log.Error().Msg("Missing or invalid 'project_name' in state")
			return fmt.Errorf("missing or invalid 'project_name' in state")
		}

		kubeClient.CancelCluster(ctx, projectName)
		log.Info().Str("project_name", projectName).Msg("Deployment canceled successfully")
		return nil
	}
}

func AddDeploymentNodeStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		kubeClient, err := ensureClient(ctx, state)
		if err != nil {
			log.Error().Err(err).Msg("Failed to ensure kubedeployer client in add deployment node step")
			return fmt.Errorf("failed to ensure kubedeployer client: %w", err)
		}

		addedCluster, ok := state["added_cluster"].(kubedeployer.Cluster)
		if !ok {
			log.Error().Msg("Missing or invalid 'cluster' in state")
			return fmt.Errorf("missing or invalid 'cluster' in state")
		}

		existingCluster, ok := state["existing_cluster"].(kubedeployer.Cluster)
		if !ok {
			log.Error().Msg("Missing or invalid 'existing_cluster' in state")
			return fmt.Errorf("missing or invalid 'existing_cluster' in state")
		}

		if err := addedCluster.PrepareCluster(kubeClient.UserID); err != nil {
			log.Error().Err(err).Str("cluster_name", addedCluster.Name).Msg("Failed to prepare added cluster")
			return fmt.Errorf("failed to prepare added cluster: %w", err)
		}

		if err := kubeClient.AssignNodeIPs(ctx, &addedCluster); err != nil {
			log.Error().Err(err).Msg("Failed to assign node IPs for added cluster")
			return fmt.Errorf("failed to assign node IPs for added cluster: %w", err)
		}

		for _, node := range addedCluster.Nodes {
			if node.ContractID != 0 {
				log.Info().Str("node_name", node.Name).Uint64("contract_id", node.ContractID).Msg("Node already deployed, skipping")
				continue
			}

			// assign nodes to the existing cluster (merge)
			if err := kubeClient.DeployNode(ctx, &existingCluster, node); err != nil {
				log.Error().Err(err).Str("node_name", node.Name).Msg("Failed to deploy node to existing cluster")
				return fmt.Errorf("failed to deploy node %s to existing cluster: %w", node.Name, err)
			}
		}

		log.Info().Str("existing_cluster_name", existingCluster.Name).Msg("All nodes deployed successfully for existing cluster")
		return nil
	}
}

func RemoveDeploymentNodeStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		kubeClient, err := ensureClient(ctx, state)
		if err != nil {
			log.Error().Err(err).Msg("Failed to ensure kubedeployer client in remove deployment node step")
			return fmt.Errorf("failed to ensure kubedeployer client: %w", err)
		}

		existingCluster, ok := state["existing_cluster"].(kubedeployer.Cluster)
		if !ok {
			log.Error().Msg("Missing or invalid 'existing_cluster' in state")
			return fmt.Errorf("missing or invalid 'existing_cluster' in state")
		}

		nodeName, ok := state["node_name"].(string)
		if !ok {
			log.Error().Msg("Missing or invalid 'node_name' in state")
			return fmt.Errorf("missing or invalid 'node_name' in state")
		}

		if err := kubeClient.RemoveClusterNode(ctx, &existingCluster, nodeName); err != nil {
			log.Error().Err(err).Str("node_name", nodeName).Msg("Failed to remove node from existing cluster")
			return fmt.Errorf("failed to remove node %s from existing cluster: %w", nodeName, err)
		}
		return nil
	}
}

func registerDeploymentActivities(engine *ewf.Engine) {
	engine.Register("setup_client", SetClientStep())
	engine.Register("deploy_network", DeployNetworkStep())
	engine.Register("deploy_nodes", DeployNodesStep())

	engine.Register("store_deployment", StoreDeploymentStep())
	engine.Register("notify_user", NotifyUserStep())
	engine.Register("cancel_deployment", CancelDeploymentStep())
	engine.Register("add_deployment_node", AddDeploymentNodeStep())
	engine.Register("remove_deployment_node", RemoveDeploymentNodeStep())

	engine.RegisterTemplate("deploy_cluster", &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: "setup_client", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
			{Name: "deploy_network", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 3, Delay: 5}},
			{Name: "deploy_nodes", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 3, Delay: 5}},
		},
	})

	engine.RegisterTemplate("cancel_deployment", &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: "cancel_deployment", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
			{Name: "notify_user", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
		},
	})
	engine.RegisterTemplate("add_deployment_node", &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: "add_deployment_node", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
			{Name: "notify_user", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
		},
	})
	engine.RegisterTemplate("remove_deployment_node", &ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: "remove_deployment_node", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
			{Name: "notify_user", RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, Delay: 2}},
		},
	})
}
