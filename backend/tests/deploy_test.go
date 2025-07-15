package main

import (
	"testing"

	"kubecloud/kubedeployer"
)

func TestDeployment(t *testing.T) {
	client := NewClient()

	err := client.Login("alaamahmoud.1223@gmail.com", "Password@22")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	t.Logf("Login successful")

	taskID, err := client.DeployCluster(clusterName)
	if err != nil {
		t.Fatalf("Deployment failed: %v", err)
	}
	t.Logf("Deployment started with task ID: %s", taskID)
}

func TestAddNode(t *testing.T) {
	client := NewClient()

	err := client.Login("alaamahmoud.1223@gmail.com", "Password@22")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	t.Logf("Login successful")

	newNode := kubedeployer.Node{
		Name:     workerNodeName,
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

	taskID, err := client.AddNode(clusterName, newNode)
	if err != nil {
		t.Fatalf("Add node failed: %v", err)
	}
	t.Logf("Add node started with task ID: %s", taskID)
}

func TestRemoveNode(t *testing.T) {
	client := NewClient()

	err := client.Login("alaamahmoud.1223@gmail.com", "Password@22")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	t.Logf("Login successful")

	err = client.RemoveNode(clusterName, workerNodeName)
	if err != nil {
		t.Fatalf("Remove node failed: %v", err)
	}
	t.Logf("Node removed successfully")
}

func TestDeleteDeployment(t *testing.T) {
	client := NewClient()

	if err := client.Login("alaamahmoud.1223@gmail.com", "Password@22"); err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	err := client.DeleteDeployment(clusterName)
	if err != nil {
		t.Fatalf("Failed to delete deployment: %v", err)
	}

	t.Log("Deployment deleted successfully")
}
