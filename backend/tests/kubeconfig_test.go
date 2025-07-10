package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestKubeconfig(t *testing.T) {
	client := NewClient()

	err := client.Login("alaamahmoud.1223@gmail.com", "Password@22")

	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	t.Logf("Login successful")

	// Get kubeconfig
	kubeconfig, err := client.GetKubeconfig(clusterName)
	if err != nil {
		t.Fatalf("Failed to get kubeconfig for '%s': %v", clusterName, err)
	}

	// Validate kubeconfig content
	if len(kubeconfig) == 0 {
		t.Fatal("Received empty kubeconfig")
	}

	// Check for essential kubeconfig components
	essentialKeys := []string{"apiVersion", "clusters", "contexts", "users"}
	for _, key := range essentialKeys {
		if !contains(kubeconfig, key) {
			t.Errorf("Kubeconfig missing essential key: %s", key)
		}
	}

	t.Logf("Successfully retrieved kubeconfig (size: %d bytes)", len(kubeconfig))

	fmt.Println(kubeconfig)
}

func contains(text, substr string) bool {
	return strings.Contains(text, substr)
}
