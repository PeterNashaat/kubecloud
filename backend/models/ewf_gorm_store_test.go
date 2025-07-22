package models

import (
	"context"
	"github.com/xmonader/ewf"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"testing"
)

// TestSQLiteStore_SaveAndLoad tests saving and loading a workflow in SQLiteStore.
func TestGormStore_SaveAndLoad(t *testing.T) {
	dbFile := "test.db"

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open dbFile: %v", err)
	}
	defer func() {
		if err := os.Remove(dbFile); err != nil {
			t.Fatalf("failed to remove dbFile: %v", err)
		}
	}()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get DB %v", err)
	}
	defer sqlDB.Close()

	store := NewGormStore(db)
	if err != nil {
		t.Fatalf("NewGormStore() error = %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("failed to close store: %v", err)
		}
	}()
	if err := store.Setup(); err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	wfName := "test-gorm-workflow"
	wf := ewf.NewWorkflow(wfName)
	wf.Steps = []ewf.Step{{Name: "dummy_activity"}}
	wf.State["key"] = "value"
	wf.CurrentStep = 2
	wf.Status = ewf.StatusCompleted

	err = store.SaveWorkflow(context.Background(), wf)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loadedWf, err := store.LoadWorkflowByUUID(context.Background(), wf.UUID)
	if err != nil {
		t.Fatalf("LoadWorkflowByUUID() error = %v", err)
	}

	// Also test loading by name
	loadedByName, err := store.LoadWorkflowByName(context.Background(), wf.Name)
	if err != nil {
		t.Fatalf("LoadWorkflowByName() error = %v", err)
	}

	if loadedByName.UUID != wf.UUID {
		t.Errorf("Expected workflow UUID %s, got %s", wf.UUID, loadedByName.UUID)
	}

	if loadedWf.Name != wfName {
		t.Errorf("Expected workflow ID %s, got %s", wfName, loadedWf.Name)
	}
	if loadedWf.CurrentStep != 2 {
		t.Errorf("Expected CurrentStep to be 2, got %d", loadedWf.CurrentStep)
	}
	if loadedWf.Status != ewf.StatusCompleted {
		t.Errorf("Expected Status to be COMPLETED, got %s", loadedWf.Status)
	}
	if loadedWf.State["key"] != "value" {
		t.Errorf("Expected state['key'] to be 'value', got '%v'", loadedWf.State["key"])
	}
}

func TestGormStore_LoadNotFound(t *testing.T) {
	dbFile := "test.db"

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open dbFile: %v", err)
	}
	defer func() {
		if err := os.Remove(dbFile); err != nil {
			t.Fatalf("failed to remove dbFile: %v", err)
		}
	}()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get DB %v", err)
	}
	defer sqlDB.Close()

	store := NewGormStore(db)
	if err != nil {
		t.Fatalf("NewGormStore() error = %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("failed to close store: %v", err)
		}
	}()
	// Test LoadWorkflowByUUID with non-existent UUID
	_, err = store.LoadWorkflowByUUID(context.Background(), "non-existent-id")
	if err == nil {
		t.Fatal("Expected an error when loading a non-existent workflow by UUID, but got nil")
	}

	// Test LoadWorkflowByName with non-existent name
	_, err = store.LoadWorkflowByName(context.Background(), "non-existent-name")
	if err == nil {
		t.Fatal("Expected an error when loading a non-existent workflow by name, but got nil")
	}
}
