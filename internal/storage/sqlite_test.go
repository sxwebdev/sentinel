package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
)

func TestSQLiteStorage(t *testing.T) {
	// Create temporary database file
	tmpfile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Create storage
	storage, err := NewSQLiteStorage(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Test 1: Save incident
	incident := &config.Incident{
		ID:          "test-1",
		ServiceName: "test-service",
		StartTime:   time.Now(),
		Error:       "test error",
		Resolved:    false,
	}

	err = storage.SaveIncident(ctx, incident)
	if err != nil {
		t.Errorf("Failed to save incident: %v", err)
	}

	// Test 2: Get incident
	retrieved, err := storage.GetIncident(ctx, "test-service", "test-1")
	if err != nil {
		t.Errorf("Failed to get incident: %v", err)
	}

	if retrieved.ID != incident.ID {
		t.Errorf("Expected ID %s, got %s", incident.ID, retrieved.ID)
	}

	// Test 3: Update incident
	now := time.Now()
	duration := time.Minute * 5
	incident.EndTime = &now
	incident.Duration = &duration
	incident.Resolved = true

	err = storage.UpdateIncident(ctx, incident)
	if err != nil {
		t.Errorf("Failed to update incident: %v", err)
	}

	// Check update
	updated, err := storage.GetIncident(ctx, "test-service", "test-1")
	if err != nil {
		t.Errorf("Failed to get updated incident: %v", err)
	}

	if !updated.Resolved {
		t.Error("Expected incident to be resolved")
	}

	// Test 4: Get incidents by service
	incidents, err := storage.GetIncidentsByService(ctx, "test-service")
	if err != nil {
		t.Errorf("Failed to get service incidents: %v", err)
	}

	if len(incidents) != 1 {
		t.Errorf("Expected 1 incident, got %d", len(incidents))
	}

	// Test 5: Get active incidents
	active, err := storage.GetActiveIncidents(ctx)
	if err != nil {
		t.Errorf("Failed to get active incidents: %v", err)
	}

	if len(active) != 0 {
		t.Errorf("Expected 0 active incidents, got %d", len(active))
	}

	// Test 6: Get statistics
	stats, err := storage.GetServiceStats(ctx, "test-service", time.Now().Add(-time.Hour))
	if err != nil {
		t.Errorf("Failed to get service stats: %v", err)
	}

	if stats.TotalIncidents != 1 {
		t.Errorf("Expected 1 total incident, got %d", stats.TotalIncidents)
	}
}

func TestSQLiteORM(t *testing.T) {
	// Create temporary database file
	tmpfile, err := os.CreateTemp("", "test-orm-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Create storage
	storage, err := NewSQLiteStorage(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Test ORM functionality
	orm := storage.orm

	// Create incident through ORM
	incident := &config.Incident{
		ID:          "orm-test-1",
		ServiceName: "orm-service",
		StartTime:   time.Now(),
		Error:       "ORM test error",
		Resolved:    false,
	}

	err = orm.CreateIncident(ctx, incident)
	if err != nil {
		t.Errorf("Failed to create incident through ORM: %v", err)
	}

	// Get incident through ORM
	retrieved, err := orm.FindIncidentByID(ctx, "orm-service", "orm-test-1")
	if err != nil {
		t.Errorf("Failed to get incident through ORM: %v", err)
	}

	if retrieved.ServiceName != "orm-service" {
		t.Errorf("Expected service name 'orm-service', got '%s'", retrieved.ServiceName)
	}

	// Test QueryIncidents
	query := orm.QueryIncidents()
	query = query.Where(query.Equal("service_name", "orm-service"))

	sql, args := query.Build()
	if sql == "" {
		t.Error("Expected SQL query to be generated")
	}

	if len(args) == 0 {
		t.Error("Expected query arguments")
	}

	t.Logf("Generated SQL: %s", sql)
	t.Logf("Arguments: %v", args)
}

func TestStorageFactory(t *testing.T) {
	// Test factory
	tmpfile, err := os.CreateTemp("", "test-factory-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	storage, err := NewStorage(StorageTypeSQLite, tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	// Check that this is SQLiteStorage
	_, ok := storage.(*SQLiteStorage)
	if !ok {
		t.Error("Expected SQLiteStorage type")
	}
}
