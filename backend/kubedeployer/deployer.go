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
	tfplugin, err := deployer.NewTFPluginClient(mnemonic, deployer.WithNetwork(gridNet))
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

	return loadNewClusterState(ctx, tfplugin, cluster, networkName)
}

func AddNodesToCluster(ctx context.Context, gridNet, mnemonic string, cluster Cluster, sshKey string, leaderIP string, existingCluster *Cluster) (Cluster, error) {
	tfplugin, err := deployer.NewTFPluginClient(mnemonic, deployer.WithNetwork(gridNet))
	if err != nil {
		return Cluster{}, fmt.Errorf("failed to create TFPluginClient: %v", err)
	}
	defer tfplugin.Close()

	networkName := getNetworkName(cluster)

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

	net, err := workloadNetwork(networkName, cluster.Name, nodeIDs)
	if err != nil {
		return fmt.Errorf("failed to create network workload: %v", err)
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

	cluster.Network = networkName
	return cluster, nil
}

// mergeClusterStates merges the existing cluster state with new nodes
func mergeClusterStates(existingCluster, newNodesCluster Cluster) Cluster {
	return Cluster{
		Nodes:   append(existingCluster.Nodes, newNodesCluster.Nodes...),
		Network: existingCluster.Network,
		Token:   existingCluster.Token,
	}
}
