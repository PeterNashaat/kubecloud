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

type Cluster struct {
	Name  string `json:"name"` // the projectname in metadata, to get/list all related deployments
	Token string `json:"token"`
	Nodes []Node `json:"nodes"`

	// Computed
	Network         string         `json:"network"`
	NetworkWorkload workloads.ZNet `json:"network_workload"` // the network workload created for this cluster
}

type Node struct {
	Name   string   `json:"name"` // name of the deployment
	Type   NodeType `json:"type"`
	NodeID uint32   `json:"node_id"`

	CPU      uint8             `json:"cpu"`
	Memory   uint64            `json:"memory"`    // Memory in MB
	RootSize uint64            `json:"root_size"` // Storage in MB
	DiskSize uint64            `json:"disk_size"` // Storage in MB
	EnvVars  map[string]string `json:"env_vars"`  // SSH_KEY, etc.

	// Optional fields
	Flist      string `json:"flist,omitempty"`
	Entrypoint string `json:"entrypoint,omitempty"`

	// Computed
	IP          string `json:"ip,omitempty"`
	MyceliumIP  string `json:"mycelium_ip,omitempty"`
	PlanetaryIP string `json:"planetary_ip,omitempty"`
	ContractID  uint64 `json:"contract_id,omitempty"`
}

func workloadsFromNode(node Node, networkName string, token string, vmIP, leaderIP, sshKey string) (workloads.VM, workloads.Disk, error) {
	netSeed, err := getRandomMyceliumNetSeed()
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

		NetworkName: networkName,
		IP:          vmIP,
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
