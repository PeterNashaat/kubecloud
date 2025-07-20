//go:build example

package tests

import (
	"kubecloud/kubedeployer"
	"testing"
)

const (
	testEmail    = "alaamahmoud.1223@gmail.com"
	testPassword = "Password@22"
)

func TestClient_DeployCluster(t *testing.T) {
	client := NewClient()

	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	cluster := kubedeployer.Cluster{
		Name:  "jrk8s02",
		Token: "test-token-123",
		Nodes: []kubedeployer.Node{
			{
				Name:     "leader",
				Type:     kubedeployer.NodeTypeLeader,
				CPU:      1,
				Memory:   2 * 1024,
				RootSize: 10240,
				DiskSize: 10240,
				EnvVars: map[string]string{
					"SSH_KEY": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDJ1t4Ug8EfykmJwAbYudyYYN/f7dZaVg3KGD2Pz0bd9pajAAASWYrss3h2ctCZWluM6KAt289RMNzxlNUkOMJ9WhCIxqDAwtg05h/J27qlaGCPP8BCEITwqNKsLwzmMZY1UFc+sSUyjd35d3kjtv+rzo2meaReZnUFNPisvxGoygftAE6unqNa7TKonVDS1YXzbpT8XdtCV1Y6ACx+3a82mFR07zgmY4BVOixNBy2Lzpq9KiZTz91Bmjg8dy4xUyWLiTmnye51hEBgUzPprjffZByYSb2Ag9hpNE1AdCGCli/0TbEwFn9iEroh/xmtvZRpux+L0OmO93z5Sz+RLiYXKiYVV5R5XYP8y5eYi48RY2qr82sUl5+WnKhI8nhzayO9yjPEp3aTvR1FdDDj5ocB7qKi47R8FXIuwzZf+kJ7ZYmMSG7N21zDIJrz6JGy9KMi7nX1sqy7NSqX3juAasIjx0IJsE8zv9qokZ83hgcDmTJjnI+YXimelhcHn4M52hU= omar@jarvis",
				},
				NodeID: 337,
			},
			{
				Name:     "master",
				Type:     kubedeployer.NodeTypeMaster,
				CPU:      1,
				Memory:   2 * 1024,
				RootSize: 10240,
				DiskSize: 10240,
				EnvVars: map[string]string{
					"SSH_KEY": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDJ1t4Ug8EfykmJwAbYudyYYN/f7dZaVg3KGD2Pz0bd9pajAAASWYrss3h2ctCZWluM6KAt289RMNzxlNUkOMJ9WhCIxqDAwtg05h/J27qlaGCPP8BCEITwqNKsLwzmMZY1UFc+sSUyjd35d3kjtv+rzo2meaReZnUFNPisvxGoygftAE6unqNa7TKonVDS1YXzbpT8XdtCV1Y6ACx+3a82mFR07zgmY4BVOixNBy2Lzpq9KiZTz91Bmjg8dy4xUyWLiTmnye51hEBgUzPprjffZByYSb2Ag9hpNE1AdCGCli/0TbEwFn9iEroh/xmtvZRpux+L0OmO93z5Sz+RLiYXKiYVV5R5XYP8y5eYi48RY2qr82sUl5+WnKhI8nhzayO9yjPEp3aTvR1FdDDj5ocB7qKi47R8FXIuwzZf+kJ7ZYmMSG7N21zDIJrz6JGy9KMi7nX1sqy7NSqX3juAasIjx0IJsE8zv9qokZ83hgcDmTJjnI+YXimelhcHn4M52hU= omar@jarvis",
				},
				NodeID: 179,
			},
			{
				Name:     "worker2",
				Type:     kubedeployer.NodeTypeWorker,
				CPU:      1,
				Memory:   2 * 1024,
				RootSize: 10240,
				DiskSize: 10240,
				EnvVars: map[string]string{
					"SSH_KEY": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDJ1t4Ug8EfykmJwAbYudyYYN/f7dZaVg3KGD2Pz0bd9pajAAASWYrss3h2ctCZWluM6KAt289RMNzxlNUkOMJ9WhCIxqDAwtg05h/J27qlaGCPP8BCEITwqNKsLwzmMZY1UFc+sSUyjd35d3kjtv+rzo2meaReZnUFNPisvxGoygftAE6unqNa7TKonVDS1YXzbpT8XdtCV1Y6ACx+3a82mFR07zgmY4BVOixNBy2Lzpq9KiZTz91Bmjg8dy4xUyWLiTmnye51hEBgUzPprjffZByYSb2Ag9hpNE1AdCGCli/0TbEwFn9iEroh/xmtvZRpux+L0OmO93z5Sz+RLiYXKiYVV5R5XYP8y5eYi48RY2qr82sUl5+WnKhI8nhzayO9yjPEp3aTvR1FdDDj5ocB7qKi47R8FXIuwzZf+kJ7ZYmMSG7N21zDIJrz6JGy9KMi7nX1sqy7NSqX3juAasIjx0IJsE8zv9qokZ83hgcDmTJjnI+YXimelhcHn4M52hU= omar@jarvis",
				},
				NodeID: 179,
			},
		},
	}

	taskID, err := client.DeployCluster(cluster)
	if err != nil {
		t.Errorf("Deployment failed: %v", err)
		return
	}
	t.Logf("Deployment started with task ID: %s", taskID)
}

func TestClient_AddNode(t *testing.T) {
	client := NewClient()

	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	newNode := kubedeployer.Node{
		Name:     "worker2",
		Type:     kubedeployer.NodeTypeWorker,
		CPU:      1,
		Memory:   2 * 1024,
		RootSize: 10240,
		DiskSize: 10240,
		EnvVars: map[string]string{
			"SSH_KEY": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDJ1t4Ug8EfykmJwAbYudyYYN/f7dZaVg3KGD2Pz0bd9pajAAASWYrss3h2ctCZWluM6KAt289RMNzxlNUkOMJ9WhCIxqDAwtg05h/J27qlaGCPP8BCEITwqNKsLwzmMZY1UFc+sSUyjd35d3kjtv+rzo2meaReZnUFNPisvxGoygftAE6unqNa7TKonVDS1YXzbpT8XdtCV1Y6ACx+3a82mFR07zgmY4BVOixNBy2Lzpq9KiZTz91Bmjg8dy4xUyWLiTmnye51hEBgUzPprjffZByYSb2Ag9hpNE1AdCGCli/0TbEwFn9iEroh/xmtvZRpux+L0OmO93z5Sz+RLiYXKiYVV5R5XYP8y5eYi48RY2qr82sUl5+WnKhI8nhzayO9yjPEp3aTvR1FdDDj5ocB7qKi47R8FXIuwzZf+kJ7ZYmMSG7N21zDIJrz6JGy9KMi7nX1sqy7NSqX3juAasIjx0IJsE8zv9qokZ83hgcDmTJjnI+YXimelhcHn4M52hU= omar@jarvis",
		},
		NodeID: 150,
	}

	taskID, err := client.AddNode("jrk8s02", newNode)
	if err != nil {
		t.Errorf("Add node failed: %v", err)
		return
	}
	t.Logf("Add node started with task ID: %s", taskID)
}

func TestClient_RemoveNode(t *testing.T) {
	client := NewClient()

	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	err = client.RemoveNode("jrk8s02", "worker2")
	if err != nil {
		t.Errorf("Remove node failed: %v", err)
		return
	}
	t.Log("Node removed successfully")
}

func TestClient_DeleteCluster(t *testing.T) {
	client := NewClient()

	if err := client.Login(testEmail, testPassword); err != nil {
		t.Errorf("Failed to login: %v", err)
		return
	}

	err := client.DeleteCluster("jrk8s02")
	if err != nil {
		t.Errorf("Failed to delete cluster: %v", err)
		return
	}
	t.Log("Cluster deleted successfully")
}

func TestClient_ListenToSSE(t *testing.T) {
	client := NewClient()
	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful. Listening to all SSE events...")

	if err := client.ListenToSSE(""); err != nil {
		t.Errorf("SSE listen failed: %v", err)
	}
}
