package kubedeployer

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/zosbase/pkg/netlight/resource"
)

func DeployCluster(ctx context.Context, gridNet, mnemonic string, cluster Cluster, sshKey string) (Cluster, error) {
	tfplugin, err := deployer.NewTFPluginClient(mnemonic, deployer.WithNetwork(gridNet), deployer.WithLogs())
	if err != nil {
		return Cluster{}, fmt.Errorf("failed to create TFPluginClient: %v", err)
	}
	defer tfplugin.Close()

	networkName := getNetworkName(cluster)

	if err := deployNetwork(ctx, tfplugin, cluster, networkName); err != nil {
		return Cluster{}, err
	}

	ensureLeaderNode(&cluster)

	if err := deployNodes(ctx, tfplugin, cluster, networkName, sshKey, ""); err != nil {
		return Cluster{}, err
	}

	// Load the complete cluster state including the full network workload
	cluster, err = loadNewClusterState(ctx, tfplugin, cluster, networkName)
	if err != nil {
		return Cluster{}, fmt.Errorf("failed to load new cluster state: %v", err)
	}

	return cluster, nil
}

func AddNodesToCluster(ctx context.Context, gridNet, mnemonic string, cluster Cluster, sshKey string, leaderIP string, existingCluster *Cluster) (Cluster, error) {
	tfplugin, err := deployer.NewTFPluginClient(mnemonic, deployer.WithNetwork(gridNet), deployer.WithLogs())
	if err != nil {
		return Cluster{}, fmt.Errorf("failed to create TFPluginClient: %v", err)
	}
	defer tfplugin.Close()

	networkName := getNetworkName(cluster)

	cluster.NetworkWorkload = existingCluster.NetworkWorkload
	cluster.Network = existingCluster.Network

	if err := deployNetwork(ctx, tfplugin, cluster, networkName); err != nil {
		return Cluster{}, err
	}

	if err := deployNodes(ctx, tfplugin, cluster, networkName, sshKey, leaderIP); err != nil {
		return Cluster{}, err
	}

	// Load state for the new nodes only
	newNodesCluster, err := loadNewClusterState(ctx, tfplugin, cluster, networkName)
	if err != nil {
		return Cluster{}, err
	}

	return mergeClusterStates(*existingCluster, newNodesCluster), nil
}

func getNetworkName(cluster Cluster) string {
	if cluster.Network != "" {
		return cluster.Network
	}
	return cluster.Name + "_network"
}

func deployNetwork(ctx context.Context, tfplugin deployer.TFPluginClient, cluster Cluster, networkName string) error {
	nodeIDs := make([]uint32, len(cluster.Nodes))
	for i, node := range cluster.Nodes {
		nodeIDs[i] = node.NodeID
	}

	var net workloads.ZNet
	var err error

	if len(cluster.NetworkWorkload.NodeDeploymentID) > 0 {
		log.Info().Msgf("Using existing network workload for network: %s", networkName)
		net = cluster.NetworkWorkload

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

		log.Info().Msgf("Appending nodes %v to existing network %s. Total nodes: %v", nodeIDs, networkName, net.Nodes)
	} else {
		// Create new network workload
		log.Info().Msgf("Creating new network workload for network: %s", networkName)
		net, err = workloadNetwork(networkName, cluster.Name, nodeIDs)
		if err != nil {
			return fmt.Errorf("failed to create network workload: %v", err)
		}
	}

	log.Debug().Msgf("Deploying network %s with nodes %v", net.Name, net.Nodes)
	if err := tfplugin.NetworkDeployer.Deploy(ctx, &net); err != nil {
		return fmt.Errorf("failed to deploy network: %v", err)
	}

	return nil
}

func ensureLeaderNode(cluster *Cluster) {
	hasLeader := false
	for _, node := range cluster.Nodes {
		if node.Type == NodeTypeLeader {
			hasLeader = true
			break
		}
	}

	if !hasLeader {
		for i, node := range cluster.Nodes {
			if node.Type == NodeTypeMaster {
				cluster.Nodes[i].Type = NodeTypeLeader
				break
			}
		}
	}
}

func deployNodes(ctx context.Context, tfplugin deployer.TFPluginClient, cluster Cluster, networkName, sshKey, leaderIP string) error {
	// TODO: verify the leaderIP is consistent
	if leaderIP == "" {
		for _, node := range cluster.Nodes {
			if node.Type == NodeTypeLeader {
				ip, err := getIpForVm(ctx, tfplugin, networkName, node.NodeID)
				if err != nil {
					return fmt.Errorf("failed to get IP for leader node %d: %v", node.NodeID, err)
				}
				leaderIP = ip
				break
			}
		}
	}

	for _, node := range cluster.Nodes {
		if err := deployNode(ctx, tfplugin, node, cluster, networkName, sshKey, leaderIP); err != nil {
			return err
		}
	}

	return nil
}

func deployNode(ctx context.Context, tfplugin deployer.TFPluginClient, node Node, cluster Cluster, networkName, sshKey, leaderIP string) error {
	ip, err := getIpForVm(ctx, tfplugin, networkName, node.NodeID)
	if err != nil {
		return fmt.Errorf("failed to get IP for node %d: %v", node.NodeID, err)
	}

	vm, disk, err := workloadsFromNode(node, networkName, cluster.Token, ip, leaderIP, sshKey)
	if err != nil {
		return fmt.Errorf("failed to create workloads for node %s: %v", node.Name, err)
	}

	depl := workloads.NewDeployment(
		cluster.Name+node.Name,
		node.NodeID, cluster.Name, nil,
		networkName,
		[]workloads.Disk{disk}, nil,
		[]workloads.VM{vm}, nil, nil, nil,
	)

	log.Debug().Msgf("Deploying node %s in cluster %s", node.Name, cluster.Name)
	if err := tfplugin.DeploymentDeployer.Deploy(ctx, &depl); err != nil {
		return fmt.Errorf("failed to deploy node %s: %v", node.Name, err)
	}

	return nil
}

func loadNewClusterState(ctx context.Context, tfplugin deployer.TFPluginClient, cluster Cluster, networkName string) (Cluster, error) {
	for idx, node := range cluster.Nodes {
		result, err := tfplugin.State.LoadDeploymentFromGrid(ctx, node.NodeID, cluster.Name+node.Name)
		if err != nil {
			return Cluster{}, fmt.Errorf("failed to load deployment for node %s: %v", node.Name, err)
		}

		seed := cluster.Nodes[idx].EnvVars["NET_SEED"]
		if seed == "" {
			seed = result.Vms[0].EnvVars["NET_SEED"]
		}
		if seed == "" {
			return Cluster{}, fmt.Errorf("NET_SEED env var is missing for node %s", node.Name)
		}

		inspections, err := resource.InspectMycelium([]byte(seed))
		if err != nil {
			return Cluster{}, fmt.Errorf("failed to inspect mycelium for node %s: %v", node.Name, err)
		}

		cluster.Nodes[idx].MyceliumIP = inspections.IP().String()
		cluster.Nodes[idx].IP = result.Vms[0].IP
		cluster.Nodes[idx].PlanetaryIP = result.Vms[0].PlanetaryIP
		cluster.Nodes[idx].ContractID = result.ContractID
	}

	netWorkload, err := tfplugin.State.LoadNetworkFromGrid(ctx, networkName)
	if err != nil {
		return Cluster{}, fmt.Errorf("failed to load complete network workload from grid: %v", err)
	}

	cluster.Network = networkName
	cluster.NetworkWorkload = netWorkload
	return cluster, nil
}

// mergeClusterStates merges the existing cluster state with new nodes
func mergeClusterStates(existingCluster, newNodesCluster Cluster) Cluster {
	// Merge nodes
	mergedNodes := make([]Node, len(existingCluster.Nodes))
	copy(mergedNodes, existingCluster.Nodes)

	for _, newNode := range newNodesCluster.Nodes {
		found := false
		for _, existingNode := range existingCluster.Nodes {
			if existingNode.NodeID == newNode.NodeID {
				found = true
				break
			}
		}
		if !found {
			mergedNodes = append(mergedNodes, newNode)
		}
	}

	networkWorkload := newNodesCluster.NetworkWorkload
	if len(networkWorkload.NodeDeploymentID) == 0 {
		networkWorkload = existingCluster.NetworkWorkload
	}

	return Cluster{
		Name:            existingCluster.Name,
		Token:           existingCluster.Token,
		Nodes:           mergedNodes,
		Network:         existingCluster.Network,
		NetworkWorkload: networkWorkload,
	}
}
