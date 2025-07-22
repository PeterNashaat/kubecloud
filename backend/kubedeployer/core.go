package kubedeployer

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

func (c *Cluster) GetLeaderNode() (Node, error) {
	return c.Nodes[0], nil
}

func (n *Node) AssignNodeIP(ctx context.Context, gridClient deployer.TFPluginClient, networkName string) error {
	log.Debug().Msgf("Assigning IP for node %s in network %s", n.Name, networkName)
	ip, err := getIpForVm(ctx, gridClient, networkName, n.NodeID)
	if err != nil {
		return fmt.Errorf("failed to get IP for node %s: %v", n.Name, err)
	}
	n.IP = ip
	return nil
}

func (c *Client) DeployNode(ctx context.Context, cluster *Cluster, node Node, masterPubKey string) error {
	log.Debug().Msgf("Deploying node %s in cluster %s", node.Name, cluster.Name)
	var leaderIP string
	if node.Type == NodeTypeLeader {
		leaderIP = ""
	} else {
		leaderNode, err := cluster.GetLeaderNode()
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get leader node for cluster %s", cluster.Name)
			return fmt.Errorf("failed to get leader node IP: %v", err)
		}

		leaderIP = leaderNode.IP
	}

	depl, err := deploymentFromNode(
		node,
		cluster.ProjectName,
		cluster.Network.Name,
		leaderIP,
		cluster.Token,
		masterPubKey,
	)
	if err != nil {
		return fmt.Errorf("failed to create VM for node: %v", err)
	}

	log.Debug().Str("node_name", node.Name).Msg("Starting deployment to grid")
	if err := c.GridClient.DeploymentDeployer.Deploy(ctx, &depl); err != nil {
		log.Error().Err(err).Str("node_name", node.Name).Msg("Failed to deploy node to grid")
		return fmt.Errorf("failed to deploy node %s: %v", node.Name, err)
	}
	log.Debug().Str("node_name", node.Name).Msg("Grid deployment successful")

	res, err := nodeFromDeployment(depl)
	if err != nil {
		log.Error().Err(err).Str("node_name", node.Name).Msg("Failed to extract node from deployment")
		return fmt.Errorf("failed to get node from deployment: %v", err)
	}
	res.OriginalName = node.OriginalName

	// used to handling adding new nodes or updating existing ones
	updated := false
	for i, n := range cluster.Nodes {
		if n.Name == res.Name {
			cluster.Nodes[i] = res
			updated = true
			log.Debug().Str("node_name", res.Name).Msg("Updated existing node in cluster")
			break
		}
	}

	if !updated {
		cluster.Nodes = append(cluster.Nodes, res)
		log.Debug().Str("node_name", res.Name).Msg("Added new node to cluster")
	}

	return nil
}

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
		log.Debug().Msgf("updating network workload for network: %s", cluster.Network.Name)
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

		log.Debug().Msgf("Appending nodes %v to existing network %s. Total nodes: %v", nodeIDs, cluster.Network.Name, net.Nodes)
	} else {
		log.Debug().Msgf("Creating new network workload for network: %s", cluster.Network.Name)
		net, err = createNetworkWorkload(cluster.Network.Name, cluster.ProjectName, nodeIDs)
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

func (c *Client) RemoveNode(ctx context.Context, cluster *Cluster, nodeName string) error {
	var nodeToRemove *Node
	var nodeIndex int
	for i, node := range cluster.Nodes {
		if node.Name == nodeName {
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
		log.Debug().Msgf("Removing node %s with contracts: %v", nodeToRemove.Name, contractsToCancel)
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
			log.Debug().Uint32("node_id", nodeToRemove.NodeID).Msg("Cleaned up network workload data for canceled network contract")
		}
	}

	return nil
}
