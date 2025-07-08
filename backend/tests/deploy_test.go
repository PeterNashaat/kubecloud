package main

import (
	"testing"
)

func TestDeployment(t *testing.T) {
	client := NewClient()

	err := client.Login("testuser@example.com", "testpassword123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	t.Logf("Login successful")

	taskID, err := client.DeployCluster("test2")
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
