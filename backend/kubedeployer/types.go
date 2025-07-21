package kubedeployer

import (
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

type NodeType string

const (
	NodeTypeWorker NodeType = "worker"
	NodeTypeMaster NodeType = "master"
	NodeTypeLeader NodeType = "leader"
)

type Cluster struct {
	Name  string `json:"name"`
	Token string `json:"token"`
	Nodes []Node `json:"nodes"`

	// Computed
	Network     workloads.ZNet `json:"network,omitempty"`
	ProjectName string         `json:"project_name,omitempty"`
}

type Node struct {
	Name   string   `json:"name"`
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
