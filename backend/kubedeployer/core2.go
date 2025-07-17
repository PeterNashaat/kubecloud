package kubedeployer

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

func GetProjectName(userID, clusterName string) string {
	return "kc" + userID + clusterName
}

// PrepareCluster prepares the cluster by setting the names and ensuring a leader node exists
func (c *Cluster) PrepareCluster(userID string) error {
	projectName := GetProjectName(userID, c.Name)
	networkName := projectName + "net"

	c.ProjectName = projectName
	c.Network.Name = networkName

	hasLeader := false
	for idx, node := range c.Nodes {
		c.Nodes[idx].Name = projectName + node.Name
		if node.Type == NodeTypeLeader {
			hasLeader = true
		}
	}

	log.Info().Msgf("Prepared cluster: %s with network: %s", c.Name, c.Network.Name)

	if !hasLeader {
		for i, node := range c.Nodes {
			if node.Type == NodeTypeMaster {
				c.Nodes[i].Type = NodeTypeLeader
				break
			}
		}
	}

	return nil
}

// GetLeaderNode MUST return the leader node in the cluster
func (c *Cluster) GetLeaderNode() (Node, error) {
	for _, node := range c.Nodes {
		if node.Type == NodeTypeLeader {
			return node, nil
		}
	}
	return Node{}, fmt.Errorf("no leader node found in cluster %s", c.Name)
}

func (n *Node) AssignNodeIP(ctx context.Context, gridClient deployer.TFPluginClient, networkName string) error {
	ip, err := getIpForVm(ctx, gridClient, networkName, n.NodeID)
	if err != nil {
		return fmt.Errorf("failed to get IP for node %s: %v", n.Name, err)
	}
	n.IP = ip
	return nil
}

// AssignNodeIPs assigns IPs to each node in the cluster (after network is deployed)
func (c *Client) AssignNodeIPs(ctx context.Context, cluster *Cluster) error {
	for idx, node := range cluster.Nodes {
		ip, err := getIpForVm(ctx, c.GridClient, cluster.Network.Name, node.NodeID)
		if err != nil {
			return fmt.Errorf("failed to get IP for node %s: %v", node.Name, err)
		}
		cluster.Nodes[idx].IP = ip
	}

	return nil
}

// DeployNode deploys a node in the cluster and assigns the resulting node to the cluster
func (c *Client) DeployNode(ctx context.Context, cluster *Cluster, node Node) error {
	log.Info().Str("node_name", node.Name).Msg(">>>>>>>> Deploying node")

	var leaderIP string
	if node.Type == NodeTypeLeader {
		// For leader nodes, we don't need another leader's IP
		leaderIP = ""
	} else {
		// For non-leader nodes, we need the leader's IP to join the cluster
		leaderNode, err := cluster.GetLeaderNode()
		if err != nil {
			return fmt.Errorf("failed to get leader node IP: %v", err)
		}
		leaderIP = leaderNode.IP
	}
	log.Info().Str("node_name", node.Name).Str("type", string(node.Type)).Str("leader_ip", leaderIP).Msg("Checking node type")

	depl, err := deploymentFromNode(
		ctx,
		node,
		cluster.Name,
		cluster.Network.Name,
		leaderIP,
		cluster.Token,
		c.masterPubKey,
	)
	if err != nil {
		return fmt.Errorf("failed to create VM for node: %v", err)
	}

	log.Info().Str("ip", node.IP).Msg("Node IP assigned")

	log.Info().Str("node_name", node.Name).Msg("Starting deployment to grid")
	if err := c.GridClient.DeploymentDeployer.Deploy(ctx, &depl); err != nil {
		log.Error().Err(err).Str("node_name", node.Name).Msg("Failed to deploy node to grid")
		return fmt.Errorf("failed to deploy node %s: %v", node.Name, err)
	}
	log.Info().Str("node_name", node.Name).Msg("Grid deployment successful")

	log.Info().Str("node_name", node.Name).Msg("Extracting node info from deployment")
	res, err := nodeFromDeployment(ctx, depl)
	if err != nil {
		log.Error().Err(err).Str("node_name", node.Name).Msg("Failed to extract node from deployment")
		return fmt.Errorf("failed to get node from deployment: %v", err)
	}

	log.Info().
		Str("deployment_node_name", res.Name).
		Str("input_node_name", node.Name).
		Uint64("contract_id", res.ContractID).
		Str("ip", res.IP).
		Str("mycelium_ip", res.MyceliumIP).
		Msg("Node deployment completed")

	// used to handling adding new nodes or updating existing ones
	updated := false
	for i, n := range cluster.Nodes {
		log.Info().
			Str("cluster_node_name", n.Name).
			Str("deployment_node_name", res.Name).
			Bool("match", n.Name == res.Name).
			Msg("Comparing node names")
		if n.Name == res.Name {
			cluster.Nodes[i] = res
			updated = true
			log.Info().Str("node_name", res.Name).Msg("Updated existing node in cluster")
			break
		}
	}

	if !updated {
		cluster.Nodes = append(cluster.Nodes, res)
		log.Info().Str("node_name", res.Name).Msg("Added new node to cluster")
	}

	return nil
}

// DeployNetwork deploys the network in cluster and assign the resulting network
func (c *Client) DeployNetwork(ctx context.Context, cluster *Cluster) error {
	seen := make(map[uint32]bool)
	nodeIDs := make([]uint32, 0, len(cluster.Nodes))
	for _, node := range cluster.Nodes {
		if !seen[node.NodeID] {
			seen[node.NodeID] = true
			nodeIDs = append(nodeIDs, node.NodeID)
		}
	}

	var net workloads.ZNet
	var err error

	if len(cluster.Network.NodeDeploymentID) > 0 {
		log.Info().Msgf("updating network workload for network: %s", cluster.Network.Name)
		net = cluster.Network

		for _, nodeID := range nodeIDs {
			found := false
			for _, existingNodeID := range net.Nodes {
				if existingNodeID == nodeID {
					found = true
					break
				}
			}
			if !found {
				net.Nodes = append(net.Nodes, nodeID)
			}
		}

		if net.MyceliumKeys == nil {
			net.MyceliumKeys = make(map[uint32][]byte)
		}
		for _, nodeID := range nodeIDs {
			if _, exists := net.MyceliumKeys[nodeID]; !exists {
				key, err := workloads.RandomMyceliumKey()
				if err != nil {
					return fmt.Errorf("failed to generate mycelium key for node %d: %v", nodeID, err)
				}
				net.MyceliumKeys[nodeID] = key
			}
		}

		log.Info().Msgf("Appending nodes %v to existing network %s. Total nodes: %v", nodeIDs, cluster.Network.Name, net.Nodes)
	} else {
		log.Info().Msgf("Creating new network workload for network: %s", cluster.Network.Name)
		net, err = createNetworkWorkload(cluster.Network.Name, cluster.Name, nodeIDs)
		if err != nil {
			return fmt.Errorf("failed to create network workload: %v", err)
		}
	}

	log.Debug().Msgf("Deploying network %s with nodes %v", net.Name, net.Nodes)
	if err := c.GridClient.NetworkDeployer.Deploy(ctx, &net); err != nil {
		return fmt.Errorf("failed to deploy network: %v", err)
	}

	cluster.Network = net

	return nil
}

func (c *Client) CancelCluster(ctx context.Context, projectName string) error {
	if err := c.GridClient.CancelByProjectName(projectName); err != nil {
		return fmt.Errorf("failed to cancel deployment contracts by project name: %v", err)
	}

	return nil
}

func (c *Client) RemoveNode(ctx context.Context, cluster *Cluster, fullNodeName string) error {
	var nodeToRemove *Node
	var nodeIndex int
	for i, node := range cluster.Nodes {
		if node.Name == fullNodeName {
			nodeToRemove = &node
			nodeIndex = i
			break
		}
	}

	if nodeToRemove == nil {
		return fmt.Errorf("node %s not found in cluster", fullNodeName)
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
