package main

import (
	"log"
	"testing"
	"time"
)

func TestDeleteDeployment(t *testing.T) {
	log.Printf("Starting delete deployment test")

	client := NewClient()

	// First register and login
	if err := client.Register("Delete Test User", "delete-test@example.com", "password123", "password123"); err != nil {
		log.Printf("Registration failed (might already exist): %v", err)
	}

	log.Printf("User registration completed")

	if err := client.Login("delete-test@example.com", "password123"); err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	log.Printf("User logged in successfully")

	log.Printf("Starting deployment of test cluster")
	taskID, err := client.DeployCluster("delete-test-cluster")
	if err != nil {
		t.Fatalf("Failed to start deployment: %v", err)
	}

	log.Printf("Deployment started with task ID: %s", taskID)

	log.Printf("Waiting for deployment to complete...")

	if err := client.ListenToSSEWithLogger(taskID, log.Printf); err != nil {
		log.Printf("Warning: SSE listening failed (deployment might still be in progress): %v", err)
	}

	time.Sleep(5 * time.Second)

	log.Printf("Checking if deployment exists before deletion")
	deployments, err := client.ListDeployments()
	if err != nil {
		t.Fatalf("Failed to list deployments: %v", err)
	}

	found := false
	for _, deployment := range deployments {
		if deploymentMap, ok := deployment.(map[string]interface{}); ok {
			if name, exists := deploymentMap["project_name"]; exists && name == "delete-test-cluster" {
				found = true
				break
			}
		}
	}

	if !found {
		log.Printf("Deployment not found in list, but proceeding with delete test anyway")
	} else {
		log.Printf("Deployment found in list, proceeding with delete")
	}

	// Now test the delete functionality
	log.Printf("Attempting to delete deployment: delete-test-cluster")
	err = client.DeleteDeployment("delete-test-cluster")
	if err != nil {
		t.Fatalf("Failed to delete deployment: %v", err)
	}

	log.Printf("Deployment deleted successfully")

	// Verify the deployment is no longer in the list
	log.Printf("Verifying deployment was removed from the list")
	deployments, err = client.ListDeployments()
	if err != nil {
		t.Fatalf("Failed to list deployments after deletion: %v", err)
	}

	found = false
	for _, deployment := range deployments {
		if deploymentMap, ok := deployment.(map[string]interface{}); ok {
			if name, exists := deploymentMap["project_name"]; exists && name == "delete-test-cluster" {
				found = true
				break
			}
		}
	}

	if found {
		t.Fatalf("Deployment still found in list after deletion")
	}

	log.Printf("Confirmed: deployment was successfully removed from the list")

	// Test deleting a non-existent deployment
	log.Printf("Testing deletion of non-existent deployment")
	err = client.DeleteDeployment("non-existent-deployment")
	if err == nil {
		t.Fatalf("Expected error when deleting non-existent deployment, but got none")
	}

	log.Printf("Confirmed: deleting non-existent deployment returns error as expected: %v", err)

	log.Printf("Delete deployment test completed successfully")
}

func TestDeleteDeploymentUnauthorized(t *testing.T) {
	log.Printf("Starting unauthorized delete test")

	client := NewClient()

	// Try to delete without authentication
	err := client.DeleteDeployment("some-deployment")
	if err == nil {
		t.Fatalf("Expected error when deleting without authentication, but got none")
	}

	log.Printf("Confirmed: unauthorized delete returns error as expected: %v", err)
}
