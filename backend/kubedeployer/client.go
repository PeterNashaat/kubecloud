package kubedeployer

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

type Client struct {
	ctx          context.Context
	GridClient   deployer.TFPluginClient
	gridNet      string
	mnemonic     string
	masterPubKey string
	UserID       string
}

func NewClient(ctx context.Context, mnemonic, gridNet, masterPubKey, userID string) (*Client, error) {
	tfplugin, err := deployer.NewTFPluginClient(
		mnemonic,
		deployer.WithNetwork(gridNet),
		// TODO: for testing
		// deployer.WithLogs(),
		deployer.WithSubstrateURL("wss://tfchain.dev.grid.tf/ws"),
		deployer.WithRelayURL("wss://relay.dev.grid.tf"),
		deployer.WithProxyURL("https://gridproxy.dev.grid.tf"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TFPluginClient: %v", err)
	}

	return &Client{
		ctx:          ctx,
		GridClient:   tfplugin,
		gridNet:      gridNet,
		mnemonic:     mnemonic,
		masterPubKey: masterPubKey,
		UserID:       userID,
	}, nil
}

func (c *Client) Close() {
	c.GridClient.Close()
}

func (c *Client) CreateCluster(ctx context.Context, cluster Cluster) (Cluster, error) {
	deploymentNames := NewDeploymentNames(c.UserID, cluster.Name)
	cluster.Name = deploymentNames.ProjectName

	if err := deployNetwork(ctx, c.GridClient, cluster, deploymentNames); err != nil {
		return Cluster{}, err
	}

	ensureLeaderNode(&cluster)

	if err := deployNodes(ctx, c.GridClient, cluster, deploymentNames, c.masterPubKey, ""); err != nil {
		c.rollbackCreateCluster(ctx, deploymentNames)
		return Cluster{}, err
	}

	cluster, err := loadNewClusterState(ctx, c.GridClient, cluster, deploymentNames)
	if err != nil {
		return Cluster{}, fmt.Errorf("failed to load new cluster state: %v", err)
	}

	return cluster, nil
}

func (c *Client) AddClusterNode(ctx context.Context, newCluster Cluster, existingCluster *Cluster) (Cluster, error) {
	// get leader ip from existing cluster
	ensureLeaderNode(existingCluster)
	leaderIP := ""
	for _, node := range existingCluster.Nodes {
		if node.Type == NodeTypeLeader {
			leaderIP = node.IP
			break
		}
	}

	deploymentNames := NewDeploymentNames(c.UserID, newCluster.Name)
	newCluster.Name = deploymentNames.ProjectName
	newCluster.Network = existingCluster.Network

	if err := deployNetwork(ctx, c.GridClient, newCluster, deploymentNames); err != nil {
		return Cluster{}, err
	}

	if err := deployNodes(ctx, c.GridClient, newCluster, deploymentNames, c.masterPubKey, leaderIP); err != nil {
		c.rollbackAddClusterNode(ctx, newCluster, deploymentNames)
		return Cluster{}, err
	}

	newNodesCluster, err := loadNewClusterState(ctx, c.GridClient, newCluster, deploymentNames)
	if err != nil {
		return Cluster{}, err
	}

	return mergeClusterStates(*existingCluster, newNodesCluster), nil
}

func (c *Client) DeleteCluster(ctx context.Context, clusterName string) error {
	deploymentNames := NewDeploymentNames(c.UserID, clusterName)

	if err := c.GridClient.CancelByProjectName(deploymentNames.ProjectName); err != nil {
		return fmt.Errorf("failed to cancel deployment contracts by project name: %v", err)
	}

	return nil
}

func (c *Client) RemoveClusterNode(ctx context.Context, cluster *Cluster, nodeName string) error {
	deploymentNames := NewDeploymentNames(c.UserID, cluster.Name)
	fullNodeName := deploymentNames.GetNodeName(nodeName)

	var nodeToRemove *Node
	var nodeIndex int
	for i, node := range cluster.Nodes {
		fmt.Println("full node name:", fullNodeName)
		fmt.Println("Found node to remove:", node.Name)
		if node.Name == fullNodeName {
			nodeToRemove = &node
			nodeIndex = i
			break
		}
	}

	if nodeToRemove == nil {
		return fmt.Errorf("node %s not found in cluster", nodeName)
	}

	if nodeToRemove.Type == NodeTypeLeader {
		return fmt.Errorf("cannot remove leader nodes")
	}

	var contractsToCancel []uint64
	if nodeToRemove.ContractID != 0 {
		contractsToCancel = append(contractsToCancel, nodeToRemove.ContractID)
	}

	networkWorkload := cluster.Network
	// is the network still used by other nodes?
	if networkContractID, exists := networkWorkload.NodeDeploymentID[nodeToRemove.NodeID]; exists && networkContractID != 0 {
		networkStillInUse := false
		for _, otherNode := range cluster.Nodes {
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
		log.Info().Msgf("Removing node %s with contracts: %v", nodeToRemove.Name, contractsToCancel)
		if err := c.GridClient.BatchCancelContract(contractsToCancel); err != nil {
			return fmt.Errorf("failed to cancel node and/or network contracts: %v", err)
		}
	}

	// Update cluster state
	updatedNodes := make([]Node, 0, len(cluster.Nodes)-1)
	updatedNodes = append(updatedNodes, cluster.Nodes[:nodeIndex]...)
	updatedNodes = append(updatedNodes, cluster.Nodes[nodeIndex+1:]...)
	cluster.Nodes = updatedNodes

	if networkContractID, exists := cluster.Network.NodeDeploymentID[nodeToRemove.NodeID]; exists {
		networkWasCanceled := false
		for _, contractID := range contractsToCancel {
			if contractID == networkContractID {
				networkWasCanceled = true
				break
			}
		}

		delete(cluster.Network.NodeDeploymentID, nodeToRemove.NodeID)

		var updatedNetworkNodes []uint32
		for _, nodeID := range cluster.Network.Nodes {
			if nodeID != nodeToRemove.NodeID {
				updatedNetworkNodes = append(updatedNetworkNodes, nodeID)
			}
		}
		cluster.Network.Nodes = updatedNetworkNodes

		if cluster.Network.NodesIPRange != nil {
			delete(cluster.Network.NodesIPRange, nodeToRemove.NodeID)
		}
		if cluster.Network.MyceliumKeys != nil {
			delete(cluster.Network.MyceliumKeys, nodeToRemove.NodeID)
		}
		if cluster.Network.Keys != nil {
			delete(cluster.Network.Keys, nodeToRemove.NodeID)
		}
		if cluster.Network.WGPort != nil {
			delete(cluster.Network.WGPort, nodeToRemove.NodeID)
		}

		if networkWasCanceled {
			log.Info().Uint32("node_id", nodeToRemove.NodeID).Msg("Cleaned up network workload data for canceled network contract")
		}
	}

	return nil
}

func (c *Client) rollbackCreateCluster(ctx context.Context, deploymentNames DeploymentNames) {
	log.Warn().Str("project_name", deploymentNames.ProjectName).Msg("Rolling back cluster creation")

	if err := c.GridClient.CancelByProjectName(deploymentNames.ProjectName); err != nil {
		log.Error().Err(err).Str("project_name", deploymentNames.ProjectName).Msg("Failed to rollback cluster creation")
	}
}

func (c *Client) rollbackAddClusterNode(ctx context.Context, cluster Cluster, deploymentNames DeploymentNames) {
	log.Warn().Str("project_name", deploymentNames.ProjectName).Msg("Rolling back node addition")

	var contractsToCancel []uint64

	for _, node := range cluster.Nodes {
		nodeName := deploymentNames.GetNodeName(node.Name)
		result, err := c.GridClient.State.LoadDeploymentFromGrid(ctx, node.NodeID, nodeName)
		if err == nil && result.ContractID != 0 {
			contractsToCancel = append(contractsToCancel, result.ContractID)
		}
	}

	if err := c.GridClient.BatchCancelContract(contractsToCancel); err != nil {
		log.Error().Err(err).Uints64("contract_ids", contractsToCancel).Msg("Failed to rollback node contracts")
	}
}
