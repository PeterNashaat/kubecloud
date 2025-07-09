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

	taskID, err := client.DeployCluster("test6")
	if err != nil {
		t.Fatalf("Deployment failed: %v", err)
	}
	t.Logf("Deployment started with task ID: %s", taskID)

	err = client.ListenToSSEWithLogger(taskID, t.Logf)
	if err != nil {
		t.Logf("SSE listening ended: %v", err)
	} else {
		t.Logf("SSE connection completed successfully")
	}
}

func TestAddNode(t *testing.T) {
	client := NewClient()

	err := client.Login("alaamahmoud.1223@gmail.com", "Password@22")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	t.Logf("Login successful")

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
		NodeID: 156,
	}

	taskID, err := client.AddNode("test6", newNode)
	if err != nil {
		t.Fatalf("Add node failed: %v", err)
	}
	t.Logf("Add node started with task ID: %s", taskID)

	err = client.ListenToSSEWithLogger(taskID, t.Logf)
	if err != nil {
		t.Logf("SSE listening ended: %v", err)
	} else {
		t.Logf("SSE connection completed successfully")
	}
}

func TestRemoveNode(t *testing.T) {
	client := NewClient()

	err := client.Login("alaamahmoud.1223@gmail.com", "Password@22")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	t.Logf("Login successful")

	err = client.RemoveNode("test6", "worker")
	if err != nil {
		t.Fatalf("Remove node failed: %v", err)
	}
	t.Logf("Node removed successfully")
}
