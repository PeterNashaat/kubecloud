//go:build example

package tests

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestClient_ListDeployments(t *testing.T) {
	client := NewClient()

	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	deployments, err := client.ListDeployments()
	if err != nil {
		t.Errorf("Failed to list deployments: %v", err)
		return
	}
	t.Logf("Found %d deployments", len(deployments))

	for i, deployment := range deployments {
		deploymentJSON, _ := json.MarshalIndent(deployment, "", "  ")
		t.Logf("Deployment %d: %s", i+1, string(deploymentJSON))
	}
}

func TestClient_GetDeployment(t *testing.T) {
	client := NewClient()

	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	deployment, err := client.GetDeployment("mycluster")
	if err != nil {
		t.Errorf("Failed to get deployment 'my-k8s-cluster': %v", err)
		return
	}

	deploymentJSON, _ := json.MarshalIndent(deployment, "", "  ")
	t.Logf("Retrieved deployment 'my-k8s-cluster': %s", string(deploymentJSON))
}

func TestClient_GetKubeconfig(t *testing.T) {
	client := NewClient()

	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	kubeconfig, err := client.GetKubeconfig("mycluster")
	if err != nil {
		t.Errorf("Failed to get kubeconfig for 'mycluster': %v", err)
		return
	}

	if len(kubeconfig) == 0 {
		t.Log("Received empty kubeconfig")
		return
	}

	essentialKeys := []string{"apiVersion", "clusters", "contexts", "users"}
	for _, key := range essentialKeys {
		if !contains(kubeconfig, key) {
			t.Errorf("Kubeconfig missing essential key: %s", key)
		}
	}

	t.Logf("Successfully retrieved kubeconfig (size: %d bytes)", len(kubeconfig))
	t.Log(kubeconfig)
}

func contains(text, substr string) bool {
	return strings.Contains(text, substr)
}
