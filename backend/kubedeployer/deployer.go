// DEPRECATED: used with the old Workers
package kubedeployer

import (
	"context"
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/zos"
	"github.com/threefoldtech/zosbase/pkg/netlight/resource"
)

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

// DeploymentNames holds all the name conversions for a deployment
type DeploymentNames struct {
	UserID              string
	OriginalClusterName string
	ProjectName         string
	NetworkName         string
}

// NewDeploymentNames creates a new naming context for a deployment
func NewDeploymentNames(userID, originalClusterName string) DeploymentNames {
	// Kubernetes only allow alphanumeric characters with - and .
	// Grid only allow alphanumeric characters with _
	// So we should allow only alphanumeric characters with
	projectName := "kc" + userID + originalClusterName
	return DeploymentNames{
		UserID:              userID,
		OriginalClusterName: originalClusterName, // OriginalClusterName is used for logging and debugging
		ProjectName:         projectName,         // used as a clusterName and as a projectName in the contracts metadata
		NetworkName:         projectName + "net", // used as a networkName
	}
}

// GetNodeName returns the prefixed node name
func (dn DeploymentNames) GetNodeName(originalNodeName string) string {
	return dn.ProjectName + originalNodeName
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

func mergeClusterStates(existingCluster, newNodesCluster Cluster) Cluster {
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

	networkWorkload := newNodesCluster.Network
	if len(networkWorkload.NodeDeploymentID) == 0 {
		networkWorkload = existingCluster.Network
	}

	return Cluster{
		Name:    existingCluster.Name,
		Token:   existingCluster.Token,
		Nodes:   mergedNodes,
		Network: networkWorkload,
	}
}

func createWorkloadsFromNode(node Node, deploymentNames DeploymentNames, networkName string, token string, vmIP, leaderIP, sshKey string) (workloads.VM, workloads.Disk, error) {
	netSeed, err := getRandomMyceliumNetSeed()
	if err != nil {
		return workloads.VM{}, workloads.Disk{}, err
	}
	workloadName := deploymentNames.GetNodeName(node.Name)

	disk := workloads.Disk{
		Name:   fmt.Sprintf("%s_data", workloadName),
		SizeGB: node.DiskSize / 1024,
	}

	vm := workloads.VM{
		Name:         workloadName,
		NodeID:       node.NodeID,
		CPU:          node.CPU,
		MemoryMB:     node.Memory,
		RootfsSizeMB: node.RootSize,
		Planetary:    true,
		EnvVars:      node.EnvVars,

		Flist:      node.Flist,
		Entrypoint: node.Entrypoint,

		NetworkName: networkName,
		IP:          vmIP,
		Mounts: []workloads.Mount{
			{
				Name:       disk.Name,
				MountPoint: K3S_DATA_DIR,
			},
		},
	}

	vm.EnvVars["K3S_NODE_NAME"] = workloadName
	vm.EnvVars["NET_SEED"] = netSeed
	vm.EnvVars["DUAL_STACK"] = "true"
	vm.EnvVars["MASTER"] = "false"
	vm.EnvVars["HA"] = "false"
	vm.EnvVars["K3S_URL"] = ""

	if node.Type == NodeTypeMaster || node.Type == NodeTypeLeader {
		vm.EnvVars["MASTER"] = "true"
		vm.EnvVars["HA"] = "true"
	}
	if node.Type != NodeTypeLeader {
		vm.EnvVars["K3S_URL"] = fmt.Sprintf("https://%s:6443", leaderIP)
	}

	if vm.EnvVars["K3S_TOKEN"] == "" {
		vm.EnvVars["K3S_TOKEN"] = K3S_TOKEN
	}
	if vm.EnvVars["K3S_FLANNEL_IFACE"] == "" {
		vm.EnvVars["K3S_FLANNEL_IFACE"] = K3S_IFACE
	}
	if vm.EnvVars["K3S_DATA_DIR"] == "" {
		vm.EnvVars["K3S_DATA_DIR"] = K3S_DATA_DIR
	}
	if vm.Flist == "" {
		vm.Flist = K3S_FLIST
	}
	if vm.Entrypoint == "" {
		vm.Entrypoint = K3S_ENTRYPOINT
	}

	vm.EnvVars["SSH_KEY"] = node.EnvVars["SSH_KEY"] + "\n" + sshKey

	return vm, disk, nil
}

func createNetworkWorkload(networkName, projectName string, nodes []uint32) (workloads.ZNet, error) {
	keys := make(map[uint32][]byte)
	for _, node := range nodes {
		key, err := workloads.RandomMyceliumKey()
		if err != nil {
			return workloads.ZNet{}, err
		}
		keys[node] = key
	}

	return workloads.ZNet{
		Name:  networkName,
		Nodes: nodes,
		IPRange: zos.IPNet{IPNet: net.IPNet{
			IP:   net.IPv4(10, 20, 0, 0),
			Mask: net.CIDRMask(16, 32),
		}},
		MyceliumKeys: keys,
		SolutionType: projectName,
	}, nil
}
