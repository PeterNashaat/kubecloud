package internal

import (
	"context"
	"fmt"
	"net"
	"slices"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/zos"
)

func buildNetwork(name, projectName string, nodes []uint32, addMycelium bool) (workloads.ZNet, error) {
	keys := make(map[uint32][]byte)
	if addMycelium {
		for _, node := range nodes {
			key, err := workloads.RandomMyceliumKey()
			if err != nil {
				return workloads.ZNet{}, err
			}
			keys[node] = key
		}
	}
	return workloads.ZNet{
		Name:  name,
		Nodes: nodes,
		IPRange: zos.IPNet{IPNet: net.IPNet{
			IP:   net.IPv4(10, 20, 0, 0),
			Mask: net.CIDRMask(16, 32),
		}},
		MyceliumKeys: keys,
		SolutionType: projectName,
	}, nil
}

// DeployKubernetesCluster deploys a kubernetes cluster
func DeployKubernetesCluster(ctx context.Context, t deployer.TFPluginClient, master workloads.K8sNode, workers []workloads.K8sNode, sshKey, k8sFlist string) (workloads.K8sCluster, error) {
	networkName := fmt.Sprintf("%snetwork", master.Name)
	projectName := fmt.Sprintf("kubernetes/%s", master.Name)
	networkNodes := []uint32{master.NodeID}
	for _, worker := range workers {
		if !slices.Contains(networkNodes, worker.NodeID) {
			networkNodes = append(networkNodes, worker.NodeID)
		}
	}

	network, err := buildNetwork(networkName, projectName, networkNodes, len(master.MyceliumIPSeed) != 0)
	if err != nil {
		return workloads.K8sCluster{}, err
	}

	master.NetworkName = networkName
	for i := range workers {
		workers[i].NetworkName = networkName
	}

	cluster := workloads.K8sCluster{
		Master:       &master,
		Workers:      workers,
		Token:        "securetoken",
		SolutionType: projectName,
		SSHKey:       sshKey,
		Flist:        k8sFlist,
		NetworkName:  networkName,
	}
	log.Debug().Msg("deploying network")
	err = t.NetworkDeployer.Deploy(ctx, &network)
	if err != nil {
		return workloads.K8sCluster{}, errors.Wrapf(err, "failed to deploy network on nodes %v", network.Nodes)
	}

	log.Debug().Msg("deploying cluster")
	err = t.K8sDeployer.Deploy(ctx, &cluster)
	if err != nil {
		log.Warn().Msg("error happened while deploying. removing network")
		revertErr := t.NetworkDeployer.Cancel(ctx, &network)
		if revertErr != nil {
			log.Error().Err(revertErr).Msg("failed to remove network")
		}
		return workloads.K8sCluster{}, errors.Wrap(err, "failed to deploy kubernetes cluster")
	}
	nodeIDs := []uint32{master.NodeID}
	for _, worker := range workers {
		nodeIDs = append(nodeIDs, worker.NodeID)
	}
	return t.State.LoadK8sFromGrid(
		ctx,
		nodeIDs,
		master.Name,
	)
}
