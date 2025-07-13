package main

import (
	"encoding/json"
	"testing"
)

func TestGetters(t *testing.T) {
	client := NewClient()

	err := client.Login("testuser@example.com", "testpassword123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	t.Logf("Login successful")

	// Test listing deployments
	deployments, err := client.ListDeployments()
	if err != nil {
		t.Fatalf("Failed to list deployments: %v", err)
	}
	t.Logf("Found %d deployments", len(deployments))

	// Print deployments info
	for i, deployment := range deployments {
		deploymentJSON, _ := json.MarshalIndent(deployment, "", "  ")
		t.Logf("Deployment %d: %s", i+1, string(deploymentJSON))
	}

	// Test getting a specific deployment if any exist
	if len(deployments) > 0 {
		// Extract project name from first deployment
		if deploymentMap, ok := deployments[0].(map[string]interface{}); ok {
			if projectName, exists := deploymentMap["project_name"]; exists {
				if name, ok := projectName.(string); ok {
					deployment, err := client.GetDeployment(name)
					if err != nil {
						t.Logf("Failed to get deployment '%s': %v", name, err)
					} else {
						deploymentJSON, _ := json.MarshalIndent(deployment, "", "  ")
						t.Logf("Retrieved deployment '%s': %s", name, string(deploymentJSON))
					}
				}
			}
		}
	} else {
		t.Logf("No deployments found to test individual retrieval")
	}
}
