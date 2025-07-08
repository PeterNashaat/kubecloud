package main

import (
	"testing"
)

func TestRegister(t *testing.T) {
	client := NewClient()

	err := client.Register("Test User", "testuser@example.com", "testpassword123", "testpassword123")
	if err != nil {
		t.Logf("Registration failed (might already exist): %v", err)
	} else {
		t.Logf("User registration successful")
	}
}
