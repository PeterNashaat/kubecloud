package kubedeployer

import (
	"context"
	"os"
	"testing"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

func TestDeployCluster(t *testing.T) {
	mnemonic := os.Getenv("MNEMONIC")
	network := os.Getenv("NETWORK")

	var cluster = Cluster{
		Name:  "clusterdeployer3",
		Token: "randomtoken",
		Nodes: []Node{
			{
				Name:     "leader",
				Type:     NodeTypeLeader,
				CPU:      1,
				Memory:   2 * 1024, // 1 GB
				RootSize: 10240,    // 10 GB
				DiskSize: 10240,    // 10 GB
				EnvVars: map[string]string{
					"SSH_KEY": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDJ1t4Ug8EfykmJwAbYudyYYN/f7dZaVg3KGD2Pz0bd9pajAAASWYrss3h2ctCZWluM6KAt289RMNzxlNUkOMJ9WhCIxqDAwtg05h/J27qlaGCPP8BCEITwqNKsLwzmMZY1UFc+sSUyjd35d3kjtv+rzo2meaReZnUFNPisvxGoygftAE6unqNa7TKonVDS1YXzbpT8XdtCV1Y6ACx+3a82mFR07zgmY4BVOixNBy2Lzpq9KiZTz91Bmjg8dy4xUyWLiTmnye51hEBgUzPprjffZByYSb2Ag9hpNE1AdCGCli/0TbEwFn9iEroh/xmtvZRpux+L0OmO93z5Sz+RLiYXKiYVV5R5XYP8y5eYi48RY2qr82sUl5+WnKhI8nhzayO9yjPEp3aTvR1FdDDj5ocB7qKi47R8FXIuwzZf+kJ7ZYmMSG7N21zDIJrz6JGy9KMi7nX1sqy7NSqX3juAasIjx0IJsE8zv9qokZ83hgcDmTJjnI+YXimelhcHn4M52hU= omar@jarvis",
				},
				NodeID: 150,
			},
			{
				Name:     "master",
				Type:     NodeTypeMaster,
				CPU:      1,
				Memory:   2 * 1024, // 1 GB
				RootSize: 10240,    // 10 GB
				DiskSize: 10240,    // 10 GB
				EnvVars: map[string]string{
					"SSH_KEY": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDJ1t4Ug8EfykmJwAbYudyYYN/f7dZaVg3KGD2Pz0bd9pajAAASWYrss3h2ctCZWluM6KAt289RMNzxlNUkOMJ9WhCIxqDAwtg05h/J27qlaGCPP8BCEITwqNKsLwzmMZY1UFc+sSUyjd35d3kjtv+rzo2meaReZnUFNPisvxGoygftAE6unqNa7TKonVDS1YXzbpT8XdtCV1Y6ACx+3a82mFR07zgmY4BVOixNBy2Lzpq9KiZTz91Bmjg8dy4xUyWLiTmnye51hEBgUzPprjffZByYSb2Ag9hpNE1AdCGCli/0TbEwFn9iEroh/xmtvZRpux+L0OmO93z5Sz+RLiYXKiYVV5R5XYP8y5eYi48RY2qr82sUl5+WnKhI8nhzayO9yjPEp3aTvR1FdDDj5ocB7qKi47R8FXIuwzZf+kJ7ZYmMSG7N21zDIJrz6JGy9KMi7nX1sqy7NSqX3juAasIjx0IJsE8zv9qokZ83hgcDmTJjnI+YXimelhcHn4M52hU= omar@jarvis",
				},
				NodeID: 152,
			},
			{
				Name:     "worker",
				Type:     NodeTypeWorker,
				CPU:      1,
				Memory:   2 * 1024, // 1 GB
				RootSize: 10240,    // 10 GB
				DiskSize: 10240,    // 10 GB
				EnvVars: map[string]string{
					"SSH_KEY": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDJ1t4Ug8EfykmJwAbYudyYYN/f7dZaVg3KGD2Pz0bd9pajAAASWYrss3h2ctCZWluM6KAt289RMNzxlNUkOMJ9WhCIxqDAwtg05h/J27qlaGCPP8BCEITwqNKsLwzmMZY1UFc+sSUyjd35d3kjtv+rzo2meaReZnUFNPisvxGoygftAE6unqNa7TKonVDS1YXzbpT8XdtCV1Y6ACx+3a82mFR07zgmY4BVOixNBy2Lzpq9KiZTz91Bmjg8dy4xUyWLiTmnye51hEBgUzPprjffZByYSb2Ag9hpNE1AdCGCli/0TbEwFn9iEroh/xmtvZRpux+L0OmO93z5Sz+RLiYXKiYVV5R5XYP8y5eYi48RY2qr82sUl5+WnKhI8nhzayO9yjPEp3aTvR1FdDDj5ocB7qKi47R8FXIuwzZf+kJ7ZYmMSG7N21zDIJrz6JGy9KMi7nX1sqy7NSqX3juAasIjx0IJsE8zv9qokZ83hgcDmTJjnI+YXimelhcHn4M52hU= omar@jarvis",
				},
				NodeID: 155,
			},
		},
	}

	tfplugin, err := deployer.NewTFPluginClient(mnemonic,
		deployer.WithNetwork(network),
		deployer.WithSubstrateURL("wss://tfchain.dev.grid.tf/ws"),
		deployer.WithRelayURL("wss://relay.dev.grid.tf"),
		deployer.WithLogs(),
	)
	if err != nil {
		t.Fatalf("failed to create TF plugin client: %v", err)
	}

	cls, err := DeployCluster(context.Background(), tfplugin, cluster, "")
	if err != nil {
		t.Fatalf("failed to deploy cluster: %v", err)
	}

	// TODO: cleanup
	t.Log("Cluster deployed successfully", cls)
}
