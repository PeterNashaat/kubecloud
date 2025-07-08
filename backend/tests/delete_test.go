package main

import (
	"log"
	"testing"
)

func TestDeleteDeployment(t *testing.T) {
	log.Printf("Starting delete deployment test")

	client := NewClient()

	if err := client.Login("alaamahmoud.1223@gmail.com", "Password@22"); err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	err := client.DeleteDeployment("test2")
	if err != nil {
		t.Fatalf("Failed to delete deployment: %v", err)
	}

	log.Printf("Deployment deleted successfully")

}
