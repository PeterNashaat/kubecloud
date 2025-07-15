//go:build example

package kubedeployer

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/zos"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func ExampleClient_CreateCluster() {
	mnemonic := os.Getenv("MNE")
	gridNet := os.Getenv("NETWORK")
	sshKey := os.Getenv("SSH_KEY")

	client, err := NewClient(context.Background(), mnemonic, gridNet, sshKey, "1")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	cluster := Cluster{
		Name:  "jrk8s12",
		Token: "secure-cluster-token",
		Nodes: []Node{
			{
				Name:     "leader",
				Type:     NodeTypeLeader,
				NodeID:   150,
				CPU:      1,
				Memory:   2 * 1024,
				RootSize: 10 * 1024,
				DiskSize: 10 * 1024,
				EnvVars: map[string]string{
					"SSH_KEY": client.masterPubKey,
				},
			},
			{
				Name:     "worker",
				Type:     NodeTypeWorker,
				NodeID:   156,
				CPU:      1,
				Memory:   2 * 1024,
				RootSize: 10 * 1024,
				DiskSize: 10 * 1024,
				EnvVars: map[string]string{
					"SSH_KEY": client.masterPubKey,
				},
			},
		},
	}

	deployedCluster, err := client.CreateCluster(ctx, cluster)
	if err != nil {
		log.Fatalf("Failed to create cluster: %v", err)
	}

	// Output: cluster created
	for _, node := range deployedCluster.Nodes {
		fmt.Printf("Node %s: IP=%s, MyceliumIP=%s, PlanetaryIP=%s\n",
			node.Name, node.IP, node.MyceliumIP, node.PlanetaryIP)
	}
	// fmt.Printf("Network: %+v\n", deployedCluster.Network)
}

func ExampleClient_AddClusterNode() {
	mnemonic := os.Getenv("MNE")
	gridNet := os.Getenv("NETWORK")
	sshKey := os.Getenv("SSH_KEY")

	client, err := NewClient(context.Background(), mnemonic, gridNet, sshKey, "1")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Loaded from db
	existingCluster := Cluster{
		Name:  "jrk8s12",
		Token: "secure-cluster-token",
		Nodes: []Node{
			{
				Name:     "leader",
				Type:     NodeTypeLeader,
				NodeID:   150,
				CPU:      1,
				Memory:   2 * 1024,
				RootSize: 10 * 1024,
				DiskSize: 10 * 1024,
				EnvVars: map[string]string{
					"SSH_KEY": client.masterPubKey,
				},
				IP: "10.20.2.2", // mimic IP for the leader node
			},
			{
				Name:     "worker",
				Type:     NodeTypeWorker,
				NodeID:   156,
				CPU:      1,
				Memory:   2 * 1024,
				RootSize: 10 * 1024,
				DiskSize: 10 * 1024,
				EnvVars: map[string]string{
					"SSH_KEY": client.masterPubKey,
				},
				IP: "10.20.3.2", // mimic IP for the worker node
			},
		},
		Network: network(),
	}

	newNodeCluster := Cluster{
		Name: "jrk8s12",
		Nodes: []Node{
			{
				Name:     "worker2",
				Type:     NodeTypeWorker,
				NodeID:   156,
				CPU:      1,
				Memory:   2 * 1024,
				RootSize: 10 * 1024,
				DiskSize: 10 * 1024,
				EnvVars: map[string]string{
					"SSH_KEY": sshKey,
				},
			},
		},
	}

	updatedCluster, err := client.AddClusterNode(ctx, newNodeCluster, &existingCluster)
	if err != nil {
		log.Fatalf("Failed to add node: %v", err)
	}

	// Output: Node added successfully!
	fmt.Printf("Node added successfully! Total nodes: %d\n", len(updatedCluster.Nodes))
}

func ExampleClient_RemoveClusterNode() {
	mnemonic := os.Getenv("MNE")
	gridNet := os.Getenv("NETWORK")
	sshKey := os.Getenv("SSH_KEY")

	client, err := NewClient(context.Background(), mnemonic, gridNet, sshKey, "1")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Loaded from db
	existingCluster := Cluster{
		Name:  "jrk8s12",
		Token: "secure-cluster-token",
		Nodes: []Node{
			{
				Name:     "leader",
				Type:     NodeTypeLeader,
				NodeID:   150,
				CPU:      1,
				Memory:   2 * 1024,
				RootSize: 10 * 1024,
				DiskSize: 10 * 1024,
				EnvVars: map[string]string{
					"SSH_KEY": client.masterPubKey,
				},
				IP: "10.20.2.2", // mimic IP for the leader node
			},
			{
				Name:     "kc1jrk8s12worker", // name after deployment
				Type:     NodeTypeWorker,
				NodeID:   156,
				CPU:      1,
				Memory:   2 * 1024,
				RootSize: 10 * 1024,
				DiskSize: 10 * 1024,
				EnvVars: map[string]string{
					"SSH_KEY": client.masterPubKey,
				},
				IP:         "10.20.3.2", // mimic IP for the worker node
				ContractID: 219957,      // mimic contract ID for the worker node (got from dashboard)
			},
		},
		Network: network(),
	}

	nodeNameToRemove := "worker"
	err = client.RemoveClusterNode(ctx, &existingCluster, nodeNameToRemove)
	if err != nil {
		log.Fatalf("Failed to remove node: %v", err)
	}

	// Output: Node 'worker' removed successfully! Remaining nodes: 2
	fmt.Printf("Node '%s' removed successfully! Remaining nodes: %d\n",
		nodeNameToRemove, len(existingCluster.Nodes))
}

func ExampleClient_DeleteCluster() {
	mnemonic := os.Getenv("MNE")
	gridNet := os.Getenv("NETWORK")
	sshKey := os.Getenv("SSH_KEY")

	client, err := NewClient(context.Background(), mnemonic, gridNet, sshKey, "1")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	projectName := "jrk8s12"
	err = client.DeleteCluster(ctx, projectName)
	if err != nil {
		log.Fatalf("Failed to delete cluster: %v", err)
	}

	// Output: Cluster deleted successfully!
	fmt.Printf("Cluster '%s' deleted successfully!\n", projectName)
}

// helper function to create a network workload, generated from the output of deploying a cluster
func network() workloads.ZNet {
	myceliumKeys := map[uint32][]byte{
		150: {77, 24, 114, 231, 39, 71, 101, 224, 108, 40, 14, 30, 24, 15, 253, 125, 177, 99, 239, 99, 92, 70, 186, 168, 58, 184, 183, 71, 199, 156, 73, 112},
		156: {214, 43, 198, 98, 103, 234, 179, 45, 192, 13, 224, 57, 40, 242, 108, 126, 174, 190, 89, 25, 75, 207, 56, 73, 122, 152, 38, 123, 215, 65, 127, 254},
	}
	nodesIPRange := map[uint32]zos.IPNet{
		150: {IPNet: net.IPNet{
			IP:   net.IPv4(10, 20, 2, 0),
			Mask: net.CIDRMask(24, 32),
		}},
		156: {IPNet: net.IPNet{
			IP:   net.IPv4(10, 20, 3, 0),
			Mask: net.CIDRMask(24, 32),
		}},
	}
	nodeDeploymentID := map[uint32]uint64{
		150: 219955,
		156: 219954,
	}
	wgPort := map[uint32]int{
		150: 11826,
		156: 31666,
	}
	key150, err := wgtypes.ParseKey("nPQvgcoLDrlylNpw9unGWuiJPJCyKCAc5nZXGlJXfjM=")
	if err != nil {
		log.Fatalf("Failed to parse key for node 150: %v", err)
	}
	key156, err := wgtypes.ParseKey("XAE/1h2Ds0MbnBb6ifwg0jFixDqKFXZ40h5BdvaRtsM=")
	if err != nil {
		log.Fatalf("Failed to parse key for node 156: %v", err)
	}
	keys := map[uint32]wgtypes.Key{
		150: key150,
		156: key156,
	}
	externalSK, err := wgtypes.ParseKey("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	if err != nil {
		log.Fatalf("Failed to parse external secret key: %v", err)
	}
	network := workloads.ZNet{
		Name:        "kc1jrk8s12net",
		Description: "",
		Nodes:       []uint32{150, 156},
		IPRange: zos.IPNet{IPNet: net.IPNet{
			IP:   net.IPv4(10, 20, 0, 0),
			Mask: net.CIDRMask(16, 32),
		}},
		AddWGAccess:      false,
		MyceliumKeys:     myceliumKeys,
		SolutionType:     "kc1jrk8s12",
		AccessWGConfig:   "",
		ExternalIP:       nil,
		ExternalSK:       externalSK,
		PublicNodeID:     0,
		NodesIPRange:     nodesIPRange,
		NodeDeploymentID: nodeDeploymentID,
		WGPort:           wgPort,
		Keys:             keys,
	}

	return network
}
