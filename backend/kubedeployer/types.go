package kubedeployer

import (
	"fmt"
	"net"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/zos"
)

type NodeType string

const (
	NodeTypeWorker NodeType = "worker"
	NodeTypeMaster NodeType = "master"
	NodeTypeLeader NodeType = "leader"
)

// final project name is
// kubecloud/cluster.name and kubecloud/cluster.name_network for net TODO: add username prefix as a namespace
// cluster.name+cluster.node.name for each deployment
// list -> get all contract for twin ide then filter where projecName start with kubecloud/
// get -> get all contracts with project name kubecloud/cluster.name

type Cluster struct {
	Name    string // the projectname in metadata, to get/list all related deployments
	Network string
	Token   string
	Nodes   []Node
}

type Node struct {
	Name   string // name of the deployment
	Type   NodeType
	NodeID uint32

	CPU      uint8
	Memory   uint64            // Memory in MB
	RootSize uint64            // Storage in MB
	DiskSize uint64            // Storage in MB
	EnvVars  map[string]string // SSH_KEY, etc.

	// Optional fields
	Flist      string
	Entrypoint string

	// Computed
	IP          string
	MyceliumIP  string
	PlanetaryIP string
}

func workloadsFromNode(node Node, networkName string, token string, vmIP, leaderIP, sshKey string) (workloads.VM, workloads.Disk, error) {
	netSeed, err := getRandomMyceliumNetSeed()
	if err != nil {
		return workloads.VM{}, workloads.Disk{}, err
	}
	ipSeed, err := getRandomMyceliumIPSeed()
	if err != nil {
		return workloads.VM{}, workloads.Disk{}, err
	}

	disk := workloads.Disk{
		Name:   fmt.Sprintf("%s_data", node.Name),
		SizeGB: node.DiskSize / 1024, // Convert MB to GB
	}

	vm := workloads.VM{
		Name:         node.Name,
		NodeID:       node.NodeID,
		CPU:          node.CPU,
		MemoryMB:     node.Memory,
		RootfsSizeMB: node.RootSize,
		Planetary:    true,
		EnvVars:      node.EnvVars,

		Flist:      node.Flist,
		Entrypoint: node.Entrypoint,

		MyceliumIPSeed: ipSeed,
		NetworkName:    networkName,
		IP:             vmIP,
		Mounts: []workloads.Mount{
			{
				Name:       disk.Name,
				MountPoint: K3S_DATA_DIR,
			},
		},
	}

	vm.EnvVars["K3S_NODE_NAME"] = node.Name
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

	// append master ssh key to the VM env vars
	vm.EnvVars["SSH_KEY"] = node.EnvVars["SSH_KEY"] + "\n" + sshKey

	return vm, disk, nil
}

func workloadNetwork(networkName, projectName string, nodes []uint32) (workloads.ZNet, error) {
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
