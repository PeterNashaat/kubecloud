package kubedeployer

import (
	"context"
	"fmt"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

// Legacy wrapper functions for backward compatibility

// DeployCluster is a wrapper function for backward compatibility
// Deprecated: Use Client.CreateCluster instead
func DeployCluster(ctx context.Context, gridNet, mnemonic string, cluster Cluster, sshKey string, userID string) (Cluster, error) {
	tfplugin, err := deployer.NewTFPluginClient(mnemonic, deployer.WithNetwork(gridNet), deployer.WithLogs())
	if err != nil {
		return Cluster{}, fmt.Errorf("failed to create TFPluginClient: %v", err)
	}
	defer tfplugin.Close()

	deploymentNames := NewDeploymentNames(userID, cluster.Name)

	// Set the internal names for the cluster
	cluster.Name = deploymentNames.ProjectName

	if err := deployNetwork(ctx, tfplugin, cluster, deploymentNames); err != nil {
		return Cluster{}, err
	}

	ensureLeaderNode(&cluster)

	if err := deployNodes(ctx, tfplugin, cluster, deploymentNames, sshKey, ""); err != nil {
		return Cluster{}, err
	}

	// Load the complete cluster state including the full network workload
	cluster, err = loadNewClusterState(ctx, tfplugin, cluster, deploymentNames)
	if err != nil {
		return Cluster{}, fmt.Errorf("failed to load new cluster state: %v", err)
	}

	return cluster, nil
}

// AddNodesToCluster is a wrapper function for backward compatibility
// Deprecated: Use Client.AddClusterNode instead
func AddNodesToCluster(ctx context.Context, gridNet, mnemonic string, cluster Cluster, sshKey string, leaderIP string, existingCluster *Cluster, userID string) (Cluster, error) {
	tfplugin, err := deployer.NewTFPluginClient(mnemonic, deployer.WithNetwork(gridNet), deployer.WithLogs())
	if err != nil {
		return Cluster{}, fmt.Errorf("failed to create TFPluginClient: %v", err)
	}
	defer tfplugin.Close()

	deploymentNames := NewDeploymentNames(userID, cluster.Name)

	// Set the internal names for the cluster
	cluster.Name = deploymentNames.ProjectName

	cluster.Network = existingCluster.Network

	if err := deployNetwork(ctx, tfplugin, cluster, deploymentNames); err != nil {
		return Cluster{}, err
	}

	if err := deployNodes(ctx, tfplugin, cluster, deploymentNames, sshKey, leaderIP); err != nil {
		return Cluster{}, err
	}

	// Load state for the new nodes only
	newNodesCluster, err := loadNewClusterState(ctx, tfplugin, cluster, deploymentNames)
	if err != nil {
		return Cluster{}, err
	}

	return mergeClusterStates(*existingCluster, newNodesCluster), nil
}
