package kubedeployer

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/zosbase/pkg/netlight/resource"
)

func deployNetwork(ctx context.Context, tfplugin deployer.TFPluginClient, cluster Cluster, deploymentNames DeploymentNames) error {
	// one network for deployments on the same node
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
		log.Info().Msgf("updating network workload for network: %s", deploymentNames.NetworkName)
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

		log.Info().Msgf("Appending nodes %v to existing network %s. Total nodes: %v", nodeIDs, deploymentNames.NetworkName, net.Nodes)
	} else {
		log.Info().Msgf("Creating new network workload for network: %s", deploymentNames.NetworkName)
		net, err = createNetworkWorkload(deploymentNames.NetworkName, deploymentNames.ProjectName, nodeIDs)
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

func deployNodes(ctx context.Context, tfplugin deployer.TFPluginClient, cluster Cluster, deploymentNames DeploymentNames, sshKey, leaderIP string) error {
	// assign IPs to nodes early to avoid conflicts later
	nodeIPs := make(map[string]string)
	for _, node := range cluster.Nodes {
		ip, err := getIpForVm(ctx, tfplugin, deploymentNames.NetworkName, node.NodeID)
		if err != nil {
			return fmt.Errorf("failed to get IP for node %s (NodeID: %d): %v", node.Name, node.NodeID, err)
		}
		nodeIPs[node.Name] = ip
	}

	if leaderIP == "" {
		for _, node := range cluster.Nodes {
			if node.Type == NodeTypeLeader {
				leaderIP = nodeIPs[node.Name]
				break
			}
		}
	}

	for _, node := range cluster.Nodes {
		if err := deployNode(ctx, tfplugin, node, cluster, deploymentNames, sshKey, leaderIP, nodeIPs[node.Name]); err != nil {
			return err
		}
	}

	return nil
}

func deployNode(ctx context.Context, tfplugin deployer.TFPluginClient, node Node, cluster Cluster, deploymentNames DeploymentNames, sshKey, leaderIP, nodeIP string) error {
	vm, disk, err := createWorkloadsFromNode(node, deploymentNames, deploymentNames.NetworkName, cluster.Token, nodeIP, leaderIP, sshKey)
	if err != nil {
		return fmt.Errorf("failed to create workloads for node %s: %v", node.Name, err)
	}

	deploymentName := deploymentNames.GetNodeName(node.Name)
	depl := workloads.NewDeployment(
		deploymentName,
		node.NodeID, deploymentNames.ProjectName, nil,
		deploymentNames.NetworkName,
		[]workloads.Disk{disk}, nil,
		[]workloads.VM{vm}, nil, nil, nil,
	)

	log.Debug().Msgf("Deploying node %s in cluster %s", node.Name, cluster.Name)
	if err := tfplugin.DeploymentDeployer.Deploy(ctx, &depl); err != nil {
		return fmt.Errorf("failed to deploy node %s: %v", node.Name, err)
	}

	return nil
}

func loadNewClusterState(ctx context.Context, tfplugin deployer.TFPluginClient, cluster Cluster, deploymentNames DeploymentNames) (Cluster, error) {
	for idx, node := range cluster.Nodes {
		nodeName := deploymentNames.GetNodeName(node.Name)
		result, err := tfplugin.State.LoadDeploymentFromGrid(ctx, node.NodeID, nodeName)
		if err != nil {
			return Cluster{}, fmt.Errorf("failed to load deployment for node %s: %v", node.Name, err)
		}

		seed := cluster.Nodes[idx].EnvVars["NET_SEED"]
		inspections, err := resource.InspectMycelium([]byte(seed))
		if err != nil {
			return Cluster{}, fmt.Errorf("failed to inspect mycelium for node %s: %v", node.Name, err)
		}

		cluster.Nodes[idx].MyceliumIP = inspections.IP().String()
		cluster.Nodes[idx].IP = result.Vms[0].IP
		cluster.Nodes[idx].PlanetaryIP = result.Vms[0].PlanetaryIP
		cluster.Nodes[idx].ContractID = result.ContractID
	}

	netWorkload, err := tfplugin.State.LoadNetworkFromGrid(ctx, deploymentNames.NetworkName)
	if err != nil {
		return Cluster{}, fmt.Errorf("failed to load complete network workload from grid: %v", err)
	}

	cluster.Network = netWorkload
	return cluster, nil
}
