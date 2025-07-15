package kubedeployer

import (
	"context"
	"fmt"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/zosbase/pkg/netlight/resource"
)

func deploymentFromNode(
	ctx context.Context,
	node Node,
	projectName string,
	networkName string,
	leaderIP string,
	token string,
	masterSSH string,
) (workloads.Deployment, error) {
	netSeed, err := getRandomMyceliumNetSeed()
	if err != nil {
		return workloads.Deployment{}, err
	}

	disk := workloads.Disk{
		Name:   fmt.Sprintf("%s_data", node.Name),
		SizeGB: node.DiskSize / 1024,
	}

	vm := workloads.VM{
		Name:         node.Name,
		NodeID:       node.NodeID,
		CPU:          node.CPU,
		MemoryMB:     node.Memory,
		RootfsSizeMB: node.RootSize,
		EnvVars:      node.EnvVars,
		Flist:        node.Flist,
		Entrypoint:   node.Entrypoint,
		NetworkName:  networkName,
		IP:           node.IP,
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

	if token == "" {
		vm.EnvVars["K3S_TOKEN"] = K3S_TOKEN
	} else {
		vm.EnvVars["K3S_TOKEN"] = token
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

	vm.EnvVars["SSH_KEY"] = node.EnvVars["SSH_KEY"] + "\n" + masterSSH

	depl := workloads.NewDeployment(
		node.Name,
		node.NodeID,
		projectName, nil,
		networkName,
		[]workloads.Disk{disk}, nil,
		[]workloads.VM{vm}, nil, nil, nil,
	)

	return depl, nil
}

func nodeFromDeployment(
	ctx context.Context,
	depl workloads.Deployment,
) (Node, error) {
	vm := depl.Vms[0]
	var node Node

	node.Name = vm.Name
	node.NodeID = vm.NodeID
	node.CPU = vm.CPU
	node.Memory = vm.MemoryMB
	node.RootSize = vm.RootfsSizeMB
	node.EnvVars = vm.EnvVars
	node.Flist = vm.Flist
	node.Entrypoint = vm.Entrypoint

	seed := node.EnvVars["NET_SEED"]
	inspections, err := resource.InspectMycelium([]byte(seed))
	if err != nil {
		return Node{}, fmt.Errorf("failed to inspect mycelium for node %s: %v", node.Name, err)
	}

	node.MyceliumIP = inspections.IP().String()
	node.IP = vm.IP
	node.PlanetaryIP = vm.PlanetaryIP
	node.ContractID = depl.ContractID

	return node, nil
}
