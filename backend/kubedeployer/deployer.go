package kubedeployer

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

func DeployCluster(ctx context.Context, tfplugin deployer.TFPluginClient, cluster Cluster) (Cluster, error) {
	// 1. Deploy network on all related nodes
	gridNodes := []uint32{}
	for _, node := range cluster.Nodes {
		gridNodes = append(gridNodes, node.NodeID)
	}

	networkName := cluster.Network
	if networkName == "" {
		networkName = cluster.Name + "_network"
	}
	net, err := workloadNetwork(networkName, cluster.Name, gridNodes)
	if err != nil {
		return Cluster{}, err
	}

	log.Debug().Msgf("Deploying network %s with nodes %v", net.Name, net.Nodes)
	if err := tfplugin.NetworkDeployer.Deploy(context.Background(), &net); err != nil {
		return Cluster{}, fmt.Errorf("failed to deploy network: %v", err)
	}

	// 2. make a deployment for each node in the cluster
	// TODO: if no leader node is defined, pick the first master node as the leader
	leaderIP := ""
	for _, node := range cluster.Nodes {
		// TODO: if multiple vms on same node, a single deployment should be created
		ip, err := getIpForVm(ctx, tfplugin, networkName, node.NodeID)
		if err != nil {
			return Cluster{}, fmt.Errorf("failed to get IP for node %d: %v", node.NodeID, err)
		}

		if node.Type == NodeTypeLeader {
			leaderIP = ip
		}

		vm, disk, err := workloadsFromNode(node, networkName, cluster.Token, ip, leaderIP)
		if err != nil {
			return Cluster{}, fmt.Errorf("failed to create workloads for node %s: %v", node.Name, err)
		}

		depl := workloads.NewDeployment(
			cluster.Name+node.Name,
			node.NodeID, "", nil,
			net.Name,
			[]workloads.Disk{disk}, nil,
			[]workloads.VM{vm}, nil, nil, nil,
		)

		log.Debug().Msgf("Deploying node %s in cluster %s", node.Name, cluster.Name)
		if err := tfplugin.DeploymentDeployer.Deploy(context.Background(), &depl); err != nil {
			// TODO: if err we should rollback the network deployment and previous VMs
			return Cluster{}, fmt.Errorf("failed to deploy VMs: %v", err)
		}
	}

	// 3. Load all deployments to get the IPs
	for idx, node := range cluster.Nodes {
		result, err := tfplugin.State.LoadDeploymentFromGrid(context.Background(), node.NodeID, cluster.Name+node.Name)
		if err != nil {
			return Cluster{}, fmt.Errorf("failed to load deployment: %v", err)
		}

		cluster.Nodes[idx].IP = result.Vms[0].IP
		cluster.Nodes[idx].MyceliumIP = result.Vms[0].MyceliumIP
		cluster.Nodes[idx].PlanetaryIP = result.Vms[0].PlanetaryIP
	}

	return cluster, nil
}
