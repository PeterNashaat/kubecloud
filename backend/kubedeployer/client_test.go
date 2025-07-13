//go:build example

package kubedeployer

import (
	"context"
	"fmt"
	"log"
	"os"
)

func ExampleClient_CreateCluster() {
	mnemonic := os.Getenv("MNE")
	gridNet := os.Getenv("NETWORK")
	sshKey := os.Getenv("SSH_KEY")

	client, err := NewClient(mnemonic, gridNet, sshKey, "1")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	cluster := Cluster{
		Name:  "jrk8s08",
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
				Name:     "workerx1",
				Type:     NodeTypeWorker,
				NodeID:   150,
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
}

func ExampleClient_AddClusterNode() {
	mnemonic := os.Getenv("MNE")
	gridNet := os.Getenv("NETWORK")
	sshKey := os.Getenv("SSH_KEY")

	client, err := NewClient(mnemonic, gridNet, sshKey, "1")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	existingCluster := Cluster{
		Name: "jrk8s04",
	}

	newNodeCluster := Cluster{
		Name: "jrk8s04",
		Nodes: []Node{
			{
				Name:     "worker2",
				Type:     NodeTypeWorker,
				NodeID:   789,
				CPU:      1,
				Memory:   2048,
				RootSize: 10240,
				DiskSize: 25000,
				EnvVars: map[string]string{
					"SSH_KEY": sshKey,
				},
			},
		},
	}

	var leaderIP string
	for _, node := range existingCluster.Nodes {
		if node.Type == NodeTypeLeader {
			leaderIP = node.IP
			break
		}
	}

	updatedCluster, err := client.AddClusterNode(ctx, newNodeCluster, leaderIP, &existingCluster)
	if err != nil {
		log.Fatalf("Failed to add node: %v", err)
	}

	// Output: Node added successfully! Total nodes: 3
	fmt.Printf("Node added successfully! Total nodes: %d\n", len(updatedCluster.Nodes))
}

func ExampleClient_RemoveClusterNode() {
	mnemonic := os.Getenv("MNE")
	gridNet := os.Getenv("NETWORK")
	sshKey := os.Getenv("SSH_KEY")

	client, err := NewClient(mnemonic, gridNet, sshKey, "1")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	cluster := Cluster{
		Name: "jrk8s04",
		Nodes: []Node{
			{Name: "leader", Type: NodeTypeLeader},
			{Name: "workerx1", Type: NodeTypeWorker},
			{Name: "worker2", Type: NodeTypeWorker},
		},
	}

	nodeNameToRemove := "worker2"
	err = client.RemoveClusterNode(ctx, &cluster, nodeNameToRemove)
	if err != nil {
		log.Fatalf("Failed to remove node: %v", err)
	}

	// Output: Node 'worker2' removed successfully! Remaining nodes: 2
	fmt.Printf("Node '%s' removed successfully! Remaining nodes: %d\n",
		nodeNameToRemove, len(cluster.Nodes))
}

func ExampleClient_DeleteCluster() {
	mnemonic := os.Getenv("MNE")
	gridNet := os.Getenv("NETWORK")
	sshKey := os.Getenv("SSH_KEY")

	client, err := NewClient(mnemonic, gridNet, sshKey, "1")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	projectName := "jrk8s04"
	err = client.DeleteCluster(ctx, projectName)
	if err != nil {
		log.Fatalf("Failed to delete cluster: %v", err)
	}

	// Output: Cluster deleted successfully!
	fmt.Printf("Cluster '%s' deleted successfully!\n", projectName)
}
