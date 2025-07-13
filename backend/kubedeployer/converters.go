package kubedeployer

import (
	"fmt"
	"net"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/zos"
)

const (
	MYC_NET_SEED_LEN = 32
	MYC_IP_SEED_LEN  = 6
	K3S_FLIST        = "https://hub.threefold.me/hanafy.3bot/ahmedhanafy725-k3s-full.flist"
	K3S_ENTRYPOINT   = "/sbin/zinit init"
	K3S_DATA_DIR     = "/mydisk"
	K3S_IFACE        = "mycelium-br"
	K3S_TOKEN        = "randomely_generated_token"
)

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

	networkWorkload := newNodesCluster.NetworkWorkload
	if len(networkWorkload.NodeDeploymentID) == 0 {
		networkWorkload = existingCluster.NetworkWorkload
	}

	return Cluster{
		Name:            existingCluster.Name,
		Token:           existingCluster.Token,
		Nodes:           mergedNodes,
		NetworkWorkload: networkWorkload,
	}
}

func createWorkloadsFromNode(node Node, deploymentNames DeploymentNames, networkName string, token string, vmIP, leaderIP, sshKey string) (workloads.VM, workloads.Disk, error) {
	netSeed, err := getRandomMyceliumNetSeed()
	if err != nil {
		return workloads.VM{}, workloads.Disk{}, err
	}

	// Use prefixed node name for internal workload name
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
