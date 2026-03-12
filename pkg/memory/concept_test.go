// Package memory provides concept and job tracking for work continuity.
package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/db"
)

func setupTestDB(t *testing.T) (*db.DB, func()) {
	t.Helper()

	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "memory_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.New(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := database.Init(); err != nil {
		database.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to initialize database: %v", err)
	}

	cleanup := func() {
		database.Close()
		os.RemoveAll(tmpDir)
	}

	return database, cleanup
}

func TestConceptManager_CreateConcept(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.MemoryConfig{
		ConceptRetentionDays: 30,
	}
	cm := NewConceptManager(database, cfg)

	tests := []struct {
		name        string
		concept     string
		context     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "create concept with text context",
			concept: "Test Concept",
			context: "This is a test context",
			wantErr: false,
		},
		{
			name:    "create concept with empty context",
			concept: "Empty Context Concept",
			context: "",
			wantErr: false,
		},
		{
			name:    "create concept with JSON context",
			concept: "JSON Context Concept",
			context: `{"working_directory": "/tmp", "files": ["test.go"] }`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := cm.CreateConcept(tt.concept, tt.context)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateConcept() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if id == "" {
					t.Error("CreateConcept() returned empty ID")
				}

				// Verify we can retrieve the concept
				concept, err := cm.GetConcept(id)
				if err != nil {
					t.Errorf("GetConcept() failed after CreateConcept: %v", err)
					return
				}

				if concept.Concept != tt.concept {
					t.Errorf("Concept text mismatch: got %v, want %v", concept.Concept, tt.concept)
				}

				if concept.Status != ConceptActive {
					t.Errorf("Concept status should be active, got %v", concept.Status)
				}
			}
		})
	}
}

func TestConceptManager_CreateConceptWithContext(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.MemoryConfig{
		ConceptRetentionDays: 30,
	}
	cm := NewConceptManager(database, cfg)

	ctx := ConceptContext{
		WorkingDirectory: "/workspace/project",
		Files:            []string{"main.go", "test.go"},
		Conversation: []ConversationEntry{
			{Role: "user", Content: "Hello", Timestamp: time.Now()},
		},
		Metadata: map[string]interface{}{
			"priority": "high",
		},
	}

	id, err := cm.CreateConceptWithContext("Structured Context Concept", ctx)
	if err != nil {
		t.Fatalf("CreateConceptWithContext() error = %v", err)
	}

	if id == "" {
		t.Error("CreateConceptWithContext() returned empty ID")
	}

	// Verify context was stored correctly
	concept, err := cm.GetConcept(id)
	if err != nil {
		t.Fatalf("GetConcept() error = %v", err)
	}

	if concept.Context.WorkingDirectory != ctx.WorkingDirectory {
		t.Errorf("WorkingDirectory mismatch: got %v, want %v", concept.Context.WorkingDirectory, ctx.WorkingDirectory)
	}

	if len(concept.Context.Files) != len(ctx.Files) {
		t.Errorf("Files length mismatch: got %d, want %d", len(concept.Context.Files), len(ctx.Files))
	}

	if len(concept.Context.Conversation) != len(ctx.Conversation) {
		t.Errorf("Conversation length mismatch: got %d, want %d", len(concept.Context.Conversation), len(ctx.Conversation))
	}
}

func TestConceptManager_ListConcepts(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.MemoryConfig{
		ConceptRetentionDays: 30,
	}
	cm := NewConceptManager(database, cfg)

	// Create test concepts
	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		id, err := cm.CreateConcept(fmt.Sprintf("Concept %d", i+1), "")
		if err != nil {
			t.Fatalf("Failed to create concept: %v", err)
		}
		ids[i] = id
	}

	// Update statuses
	cm.UpdateConceptStatus(ids[0], string(ConceptCompleted))
	cm.UpdateConceptStatus(ids[1], string(ConceptPaused))
	// ids[2] stays active

	tests := []struct {
		name          string
		status        string
		expectedCount int
	}{
		{
			name:          "list all concepts",
			status:        "",
			expectedCount: 3,
		},
		{
			name:          "list active concepts",
			status:        "active",
			expectedCount: 1,
		},
		{
			name:          "list completed concepts",
			status:        "completed",
			expectedCount: 1,
		},
		{
			name:          "list paused concepts",
			status:        "paused",
			expectedCount: 1,
		},
		{
			name:          "list abandoned concepts",
			status:        "abandoned",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			concepts, err := cm.ListConcepts(tt.status)
			if err != nil {
				t.Errorf("ListConcepts() error = %v", err)
				return
			}

			if len(concepts) != tt.expectedCount {
				t.Errorf("ListConcepts() returned %d concepts, want %d", len(concepts), tt.expectedCount)
			}
		})
	}
}

func TestConceptManager_ContinueConcept(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.MemoryConfig{
		ConceptRetentionDays: 30,
	}
	cm := NewConceptManager(database, cfg)

	// Create a concept and mark it completed
	id, err := cm.CreateConcept("Test Concept", "")
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}

	if err := cm.UpdateConceptStatus(id, string(ConceptCompleted)); err != nil {
		t.Fatalf("Failed to update concept status: %v", err)
	}

	// Continue the concept
	concept, err := cm.ContinueConcept(id)
	if err != nil {
		t.Errorf("ContinueConcept() error = %v", err)
		return
	}

	if concept.Status != ConceptActive {
		t.Errorf("ContinueConcept() should set status to active, got %v", concept.Status)
	}
}

func TestConceptManager_UpdateConceptStatus(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.MemoryConfig{
		ConceptRetentionDays: 30,
	}
	cm := NewConceptManager(database, cfg)

	id, err := cm.CreateConcept("Test Concept", "")
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}

	tests := []struct {
		name        string
		status      string
		wantErr     bool
		errContains string
	}{
		{
			name:    "update to paused",
			status:  "paused",
			wantErr: false,
		},
		{
			name:    "update to completed",
			status:  "completed",
			wantErr: false,
		},
		{
			name:    "update to active",
			status:  "active",
			wantErr: false,
		},
		{
			name:        "update to invalid status",
			status:      "invalid_status",
			wantErr:     true,
			errContains: "invalid concept status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cm.UpdateConceptStatus(id, tt.status)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateConceptStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				concept, err := cm.GetConcept(id)
				if err != nil {
					t.Errorf("GetConcept() failed: %v", err)
					return
				}

				if string(concept.Status) != tt.status {
					t.Errorf("Status not updated: got %v, want %v", concept.Status, tt.status)
				}
			}
		})
	}
}

func TestConceptManager_AddJobToConcept(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.MemoryConfig{
		ConceptRetentionDays: 30,
	}
	cm := NewConceptManager(database, cfg)

	// Create a concept
	conceptID, err := cm.CreateConcept("Test Concept", "")
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}

	// Create a job
	jobID, err := database.CreateJob(&conceptID, "executor", "Test task", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Add job to concept
	if err := cm.AddJobToConcept(conceptID, jobID.ID); err != nil {
		t.Errorf("AddJobToConcept() error = %v", err)
	}

	// Verify job was added
	concept, err := cm.GetConcept(conceptID)
	if err != nil {
		t.Fatalf("GetConcept() error = %v", err)
	}

	found := false
	for _, jid := range concept.RelatedJobs {
		if jid == jobID.ID {
			found = true
			break
		}
	}

	if !found {
		t.Error("Job ID not found in concept's related jobs")
	}
}

func TestConceptManager_AddFile(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.MemoryConfig{
		ConceptRetentionDays: 30,
	}
	cm := NewConceptManager(database, cfg)

	id, err := cm.CreateConcept("Test Concept", "")
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}

	// Add files
	files := []string{"/path/to/file1.go", "/path/to/file2.go"}
	for _, file := range files {
		if err := cm.AddFile(id, file); err != nil {
			t.Errorf("AddFile() error = %v", err)
		}
	}

	// Verify files were added
	concept, err := cm.GetConcept(id)
	if err != nil {
		t.Fatalf("GetConcept() error = %v", err)
	}

	if len(concept.Context.Files) != len(files) {
		t.Errorf("Files count mismatch: got %d, want %d", len(concept.Context.Files), len(files))
	}

	// Test adding duplicate file (should not add again)
	if err := cm.AddFile(id, files[0]); err != nil {
		t.Errorf("AddFile() duplicate error = %v", err)
	}

	concept, _ = cm.GetConcept(id)
	if len(concept.Context.Files) != len(files) {
		t.Error("Duplicate file was added")
	}
}

func TestConceptManager_SetWorkingDirectory(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.MemoryConfig{
		ConceptRetentionDays: 30,
	}
	cm := NewConceptManager(database, cfg)

	id, err := cm.CreateConcept("Test Concept", "")
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}

	wd := "/workspace/my-project"
	if err := cm.SetWorkingDirectory(id, wd); err != nil {
		t.Errorf("SetWorkingDirectory() error = %v", err)
	}

	concept, err := cm.GetConcept(id)
	if err != nil {
		t.Fatalf("GetConcept() error = %v", err)
	}

	if concept.Context.WorkingDirectory != wd {
		t.Errorf("WorkingDirectory mismatch: got %v, want %v", concept.Context.WorkingDirectory, wd)
	}
}

func TestConceptManager_AddConversationEntry(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.MemoryConfig{
		ConceptRetentionDays: 30,
	}
	cm := NewConceptManager(database, cfg)

	id, err := cm.CreateConcept("Test Concept", "")
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}

	entry := ConversationEntry{
		Role:      "user",
		Content:   "Hello, can you help me?",
		Timestamp: time.Now(),
	}

	if err := cm.AddConversationEntry(id, entry); err != nil {
		t.Errorf("AddConversationEntry() error = %v", err)
	}

	concept, err := cm.GetConcept(id)
	if err != nil {
		t.Fatalf("GetConcept() error = %v", err)
	}

	if len(concept.Context.Conversation) != 1 {
		t.Errorf("Conversation length mismatch: got %d, want 1", len(concept.Context.Conversation))
	}

	if concept.Context.Conversation[0].Content != entry.Content {
		t.Errorf("Conversation content mismatch: got %v, want %v", concept.Context.Conversation[0].Content, entry.Content)
	}
}

func TestConceptManager_CleanupOldConcepts(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.MemoryConfig{
		ConceptRetentionDays: 1, // 1 day retention
	}
	cm := NewConceptManager(database, cfg)

	// Create a concept
	id, err := cm.CreateConcept("Old Concept", "")
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}

	// Mark as completed
	if err := cm.UpdateConceptStatus(id, string(ConceptCompleted)); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Cleanup with 0 days (should use config value of 1)
	if err := cm.CleanupOldConcepts(0); err != nil {
		t.Errorf("CleanupOldConcepts() error = %v", err)
	}

	// Concept should still exist (just created)
	_, err = cm.GetConcept(id)
	if err != nil {
		t.Error("Concept was incorrectly cleaned up")
	}
}

func TestConceptManager_DeleteConcept(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.MemoryConfig{
		ConceptRetentionDays: 30,
	}
	cm := NewConceptManager(database, cfg)

	id, err := cm.CreateConcept("Test Concept", "")
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}

	if err := cm.DeleteConcept(id); err != nil {
		t.Errorf("DeleteConcept() error = %v", err)
	}

	// Verify concept was deleted
	_, err = cm.GetConcept(id)
	if err == nil {
		t.Error("Concept should have been deleted")
	}
}

func TestConceptStatus_Validation(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{"active", true},
		{"paused", true},
		{"completed", true},
		{"abandoned", true},
		{"invalid", false},
		{"", false},
		{"ACTIVE", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := IsValidConceptStatus(tt.status); got != tt.valid {
				t.Errorf("IsValidConceptStatus(%q) = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}

func TestValidConceptStatuses(t *testing.T) {
	statuses := ValidConceptStatuses()
	expected := []string{"active", "paused", "completed", "abandoned"}

	if len(statuses) != len(expected) {
		t.Errorf("ValidConceptStatuses() returned %d statuses, want %d", len(statuses), len(expected))
	}

	for i, status := range expected {
		if i >= len(statuses) || statuses[i] != status {
			t.Errorf("ValidConceptStatuses()[%d] = %v, want %v", i, statuses[i], status)
		}
	}
}

func TestGenerateConceptID(t *testing.T) {
	id1 := GenerateConceptID()
	id2 := GenerateConceptID()

	if id1 == "" {
		t.Error("GenerateConceptID() returned empty string")
	}

	if id1 == id2 {
		t.Error("GenerateConceptID() returned duplicate IDs")
	}
}

func TestConceptManager_NilDB(t *testing.T) {
	cm := NewConceptManager(nil, nil)

	_, err := cm.CreateConcept("test", "")
	if err == nil || err.Error() != "database not initialized" {
		t.Errorf("CreateConcept() with nil db should return 'database not initialized' error, got: %v", err)
	}

	_, err = cm.ListConcepts("")
	if err == nil || err.Error() != "database not initialized" {
		t.Errorf("ListConcepts() with nil db should return 'database not initialized' error, got: %v", err)
	}

	_, err = cm.GetConcept("test-id")
	if err == nil || err.Error() != "database not initialized" {
		t.Errorf("GetConcept() with nil db should return 'database not initialized' error, got: %v", err)
	}
}

