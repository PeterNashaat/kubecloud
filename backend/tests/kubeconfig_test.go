package main

import (
	"strings"
	"testing"
)

func TestKubeconfig(t *testing.T) {
	client := NewClient()

	err := client.Login("testuser@example.com", "testpassword123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	t.Logf("Login successful")

	// List deployments to find an available cluster
	deployments, err := client.ListDeployments()
	if err != nil {
		t.Fatalf("Failed to list deployments: %v", err)
	}

	if len(deployments) == 0 {
		t.Skip("No deployments found to test kubeconfig retrieval")
		return
	}

	// Get the first deployment's project name
	var projectName string
	if deploymentMap, ok := deployments[0].(map[string]interface{}); ok {
		if name, exists := deploymentMap["project_name"]; exists {
			if nameStr, ok := name.(string); ok {
				projectName = nameStr
			}
		}
	}

	if projectName == "" {
		t.Skip("Could not extract project name from deployment")
		return
	}

	t.Logf("Testing kubeconfig retrieval for deployment: %s", projectName)

	// Get kubeconfig
	kubeconfig, err := client.GetKubeconfig(projectName)
	if err != nil {
		t.Fatalf("Failed to get kubeconfig for '%s': %v", projectName, err)
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

	// Log first 5 lines for verification (don't log sensitive data)
	lines := getFirstLines(kubeconfig, 5)
	for i, line := range lines {
		if len(line) > 100 {
			t.Logf("Line %d: %s...", i+1, line[:100])
		} else {
			t.Logf("Line %d: %s", i+1, line)
		}
	}
}

func contains(text, substr string) bool {
	return strings.Contains(text, substr)
}

func getFirstLines(text string, n int) []string {
	lines := []string{}
	currentLine := ""
	lineCount := 0

	for _, char := range text {
		if char == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
			lineCount++
			if lineCount >= n {
				break
			}
		} else {
			currentLine += string(char)
		}
	}

	// Add the last line if it doesn't end with newline and we haven't reached limit
	if currentLine != "" && lineCount < n {
		lines = append(lines, currentLine)
	}

	return lines
}
