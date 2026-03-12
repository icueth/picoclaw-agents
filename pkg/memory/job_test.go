// Package memory provides concept and job tracking for work continuity.
package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/db"
)

func setupTestDBForJobs(t *testing.T) (*db.DB, func()) {
	t.Helper()

	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "job_test_*")
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

func TestJobManager_CreateJob(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	tests := []struct {
		name    string
		role    string
		task    string
		context map[string]interface{}
		wantErr bool
	}{
		{
			name:    "create simple job",
			role:    "executor",
			task:    "Run tests",
			context: nil,
			wantErr: false,
		},
		{
			name: "create job with context",
			role: "coder",
			task: "Write function",
			context: map[string]interface{}{
				"working_directory": "/tmp",
				"files":             []string{"test.go"},
				"priority":          "high",
			},
			wantErr: false,
		},
		{
			name: "create job with concept_id",
			role: "specialist",
			task: "Analyze data",
			context: map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := jm.CreateJob(tt.role, tt.task, tt.context)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if id == "" {
					t.Error("CreateJob() returned empty ID")
				}

				// Verify we can retrieve the job
				job, err := jm.GetJob(id)
				if err != nil {
					t.Errorf("GetJob() failed after CreateJob: %v", err)
					return
				}

				if job.Role != tt.role {
					t.Errorf("Role mismatch: got %v, want %v", job.Role, tt.role)
				}

				if job.Task != tt.task {
					t.Errorf("Task mismatch: got %v, want %v", job.Task, tt.task)
				}

				if job.Status != JobPending {
					t.Errorf("Job status should be pending, got %v", job.Status)
				}
			}
		})
	}
}

func TestJobManager_CreateJobWithContext(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	ctx := JobContext{
		WorkingDirectory: "/workspace/project",
		Files:            []string{"main.go", "test.go"},
		Variables: map[string]interface{}{
			"timeout": 60,
		},
		ParentContext: map[string]interface{}{
			"parent_id": "parent-123",
		},
	}

	id, err := jm.CreateJobWithContext("executor", "Test task", ctx, nil, nil)
	if err != nil {
		t.Fatalf("CreateJobWithContext() error = %v", err)
	}

	if id == "" {
		t.Error("CreateJobWithContext() returned empty ID")
	}

	// Verify context was stored correctly
	job, err := jm.GetJob(id)
	if err != nil {
		t.Fatalf("GetJob() error = %v", err)
	}

	if job.Context.WorkingDirectory != ctx.WorkingDirectory {
		t.Errorf("WorkingDirectory mismatch: got %v, want %v", job.Context.WorkingDirectory, ctx.WorkingDirectory)
	}

	if len(job.Context.Files) != len(ctx.Files) {
		t.Errorf("Files length mismatch: got %d, want %d", len(job.Context.Files), len(ctx.Files))
	}
}

func TestJobManager_ListJobs(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	// Create test jobs
	ids := make([]string, 4)
	for i := 0; i < 4; i++ {
		id, err := jm.CreateJob("executor", fmt.Sprintf("Task %d", i+1), nil)
		if err != nil {
			t.Fatalf("Failed to create job: %v", err)
		}
		ids[i] = id
	}

	// Update statuses
	jm.UpdateJobStatus(ids[0], string(JobCompleted))
	jm.UpdateJobStatus(ids[1], string(JobFailed))
	jm.UpdateJobStatus(ids[2], string(JobRunning))
	// ids[3] stays pending

	tests := []struct {
		name          string
		status        string
		expectedCount int
	}{
		{
			name:          "list all jobs",
			status:        "",
			expectedCount: 4,
		},
		{
			name:          "list pending jobs",
			status:        "pending",
			expectedCount: 1,
		},
		{
			name:          "list running jobs",
			status:        "running",
			expectedCount: 1,
		},
		{
			name:          "list completed jobs",
			status:        "completed",
			expectedCount: 1,
		},
		{
			name:          "list failed jobs",
			status:        "failed",
			expectedCount: 1,
		},
		{
			name:          "list cancelled jobs",
			status:        "cancelled",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs, err := jm.ListJobs(tt.status)
			if err != nil {
				t.Errorf("ListJobs() error = %v", err)
				return
			}

			if len(jobs) != tt.expectedCount {
				t.Errorf("ListJobs() returned %d jobs, want %d", len(jobs), tt.expectedCount)
			}
		})
	}
}

func TestJobManager_ListJobsByConcept(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	// First create a concept
	cm := NewConceptManager(database, &config.MemoryConfig{})
	conceptID, err := cm.CreateConcept("Test Concept", "")
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}

	// Create jobs with and without concept
	_, err = jm.CreateJob("executor", "Task with concept", map[string]interface{}{
		"concept_id": conceptID,
	})
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	_, err = jm.CreateJob("executor", "Task with concept 2", map[string]interface{}{
		"concept_id": conceptID,
	})
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	_, err = jm.CreateJob("executor", "Task without concept", nil)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	jobs, err := jm.ListJobsByConcept(conceptID)
	if err != nil {
		t.Errorf("ListJobsByConcept() error = %v", err)
		return
	}

	if len(jobs) != 2 {
		t.Errorf("ListJobsByConcept() returned %d jobs, want 2", len(jobs))
	}
}

func TestJobManager_UpdateJobStatus(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	id, err := jm.CreateJob("executor", "Test task", nil)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{
			name:    "update to running",
			status:  "running",
			wantErr: false,
		},
		{
			name:    "update to completed",
			status:  "completed",
			wantErr: false,
		},
		{
			name:    "update to failed",
			status:  "failed",
			wantErr: false,
		},
		{
			name:    "update to cancelled",
			status:  "cancelled",
			wantErr: false,
		},
		{
			name:    "update to pending",
			status:  "pending",
			wantErr: false,
		},
		{
			name:    "update to invalid status",
			status:  "invalid_status",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jm.UpdateJobStatus(id, tt.status)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateJobStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				job, err := jm.GetJob(id)
				if err != nil {
					t.Errorf("GetJob() failed: %v", err)
					return
				}

				if string(job.Status) != tt.status {
					t.Errorf("Status not updated: got %v, want %v", job.Status, tt.status)
				}
			}
		})
	}
}

func TestJobManager_UpdateJobResult(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	id, err := jm.CreateJob("executor", "Test task", nil)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	result := "Task completed successfully with output: All tests passed"
	if err := jm.UpdateJobResult(id, result); err != nil {
		t.Errorf("UpdateJobResult() error = %v", err)
	}

	job, err := jm.GetJob(id)
	if err != nil {
		t.Fatalf("GetJob() error = %v", err)
	}

	if job.Result == nil {
		t.Fatal("Job result is nil")
	}

	if *job.Result != result {
		t.Errorf("Result mismatch: got %v, want %v", *job.Result, result)
	}
}

func TestJobManager_ResumeJob(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	// Create and fail a job
	id, err := jm.CreateJob("executor", "Test task", nil)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	if err := jm.UpdateJobStatus(id, string(JobFailed)); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Resume the job
	job, err := jm.ResumeJob(id)
	if err != nil {
		t.Errorf("ResumeJob() error = %v", err)
		return
	}

	if job.Status != JobPending {
		t.Errorf("ResumeJob() should set status to pending, got %v", job.Status)
	}

	// Test resuming a running job (should fail)
	id2, _ := jm.CreateJob("executor", "Another task", nil)
	jm.UpdateJobStatus(id2, string(JobRunning))

	_, err = jm.ResumeJob(id2)
	if err == nil {
		t.Error("ResumeJob() should fail for running jobs")
	}
}

func TestJobManager_StartJob(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	id, err := jm.CreateJob("executor", "Test task", nil)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	activeJob, err := jm.StartJob(id)
	if err != nil {
		t.Errorf("StartJob() error = %v", err)
		return
	}

	if activeJob.ID != id {
		t.Errorf("ActiveJob ID mismatch: got %v, want %v", activeJob.ID, id)
	}

	if activeJob.Status != JobRunning {
		t.Errorf("ActiveJob status should be running, got %v", activeJob.Status)
	}

	if activeJob.Cancel == nil {
		t.Error("ActiveJob Cancel function should not be nil")
	}

	// Verify job is in active jobs map
	aj, ok := jm.GetActiveJob(id)
	if !ok {
		t.Error("Job should be in active jobs map")
	}

	if aj.ID != id {
		t.Error("Active job ID mismatch")
	}
}

func TestJobManager_KillJob(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	id, err := jm.CreateJob("executor", "Test task", nil)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Start the job
	_, err = jm.StartJob(id)
	if err != nil {
		t.Fatalf("Failed to start job: %v", err)
	}

	// Kill the job
	if err := jm.KillJob(id); err != nil {
		t.Errorf("KillJob() error = %v", err)
	}

	// Verify job is no longer in active jobs
	_, ok := jm.GetActiveJob(id)
	if ok {
		t.Error("Job should not be in active jobs map after kill")
	}

	// Verify status is cancelled
	job, err := jm.GetJob(id)
	if err != nil {
		t.Fatalf("GetJob() error = %v", err)
	}

	if job.Status != JobCancelled {
		t.Errorf("Job status should be cancelled, got %v", job.Status)
	}

	// Test killing non-existent job
	err = jm.KillJob("non-existent-id")
	if err == nil {
		t.Error("KillJob() should return error for non-existent job")
	}
}

func TestJobManager_MaxConcurrent(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  2,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	// Create and start 2 jobs (max concurrent)
	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		id, err := jm.CreateJob("executor", fmt.Sprintf("Task %d", i+1), nil)
		if err != nil {
			t.Fatalf("Failed to create job: %v", err)
		}
		ids[i] = id
	}

	// Start first 2 jobs
	_, err := jm.StartJob(ids[0])
	if err != nil {
		t.Fatalf("Failed to start job 1: %v", err)
	}

	_, err = jm.StartJob(ids[1])
	if err != nil {
		t.Fatalf("Failed to start job 2: %v", err)
	}

	// Third job should fail due to max concurrent
	_, err = jm.StartJob(ids[2])
	if err == nil {
		t.Error("StartJob() should fail when max concurrent reached")
	}

	// Kill one job
	jm.KillJob(ids[0])

	// Now third job should start
	_, err = jm.StartJob(ids[2])
	if err != nil {
		t.Errorf("StartJob() should succeed after killing a job: %v", err)
	}
}

func TestJobManager_ListActiveJobs(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	// Create and start some jobs
	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		id, err := jm.CreateJob("executor", fmt.Sprintf("Task %d", i+1), nil)
		if err != nil {
			t.Fatalf("Failed to create job: %v", err)
		}
		ids[i] = id
	}

	// Start 2 jobs
	jm.StartJob(ids[0])
	jm.StartJob(ids[1])

	activeJobs := jm.ListActiveJobs()
	if len(activeJobs) != 2 {
		t.Errorf("ListActiveJobs() returned %d jobs, want 2", len(activeJobs))
	}
}

func TestJobManager_SetJobVariable(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	id, err := jm.CreateJob("executor", "Test task", nil)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Set variables
	if err := jm.SetJobVariable(id, "timeout", 60); err != nil {
		t.Errorf("SetJobVariable() error = %v", err)
	}

	if err := jm.SetJobVariable(id, "retries", 3); err != nil {
		t.Errorf("SetJobVariable() error = %v", err)
	}

	// Get variables
	timeout, err := jm.GetJobVariable(id, "timeout")
	if err != nil {
		t.Errorf("GetJobVariable() error = %v", err)
	}

	// JSON unmarshaling converts numbers to float64
	timeoutFloat, ok := timeout.(float64)
	if !ok {
		t.Errorf("Variable type mismatch: got %T, want float64", timeout)
	} else if timeoutFloat != 60 {
		t.Errorf("Variable value mismatch: got %v, want 60", timeoutFloat)
	}

	// Get non-existent variable
	_, err = jm.GetJobVariable(id, "nonexistent")
	if err == nil {
		t.Error("GetJobVariable() should return error for non-existent variable")
	}
}

func TestJobManager_DeleteJob(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	id, err := jm.CreateJob("executor", "Test task", nil)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	if err := jm.DeleteJob(id); err != nil {
		t.Errorf("DeleteJob() error = %v", err)
	}

	// Verify job was deleted
	_, err = jm.GetJob(id)
	if err == nil {
		t.Error("Job should have been deleted")
	}
}

func TestJobManager_CleanupOldJobs(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout:       300,
		MaxConcurrent:        10,
		RetentionDays:        1,
		CleanupIntervalHours: 24,
	}
	jm := NewJobManager(database, cfg)

	// Create a job and mark as completed
	id, err := jm.CreateJob("executor", "Test task", nil)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	if err := jm.UpdateJobStatus(id, string(JobCompleted)); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Cleanup with 0 days (should use config value of 1)
	if err := jm.CleanupOldJobs(0); err != nil {
		t.Errorf("CleanupOldJobs() error = %v", err)
	}

	// Job should still exist (just created)
	_, err = jm.GetJob(id)
	if err != nil {
		t.Error("Job was incorrectly cleaned up")
	}
}

func TestJobManager_GetJobCount(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	// Create jobs
	for i := 0; i < 3; i++ {
		_, err := jm.CreateJob("executor", fmt.Sprintf("Task %d", i+1), nil)
		if err != nil {
			t.Fatalf("Failed to create job: %v", err)
		}
	}

	count, err := jm.GetJobCount("")
	if err != nil {
		t.Errorf("GetJobCount() error = %v", err)
	}

	if count != 3 {
		t.Errorf("GetJobCount() = %d, want 3", count)
	}
}

func TestJobManager_GetActiveJobCount(t *testing.T) {
	database, cleanup := setupTestDBForJobs(t)
	defer cleanup()

	cfg := &config.JobConfig{
		DefaultTimeout: 300,
		MaxConcurrent:  10,
		RetentionDays:  30,
	}
	jm := NewJobManager(database, cfg)

	// Create and start jobs
	for i := 0; i < 2; i++ {
		id, err := jm.CreateJob("executor", fmt.Sprintf("Task %d", i+1), nil)
		if err != nil {
			t.Fatalf("Failed to create job: %v", err)
		}
		jm.StartJob(id)
	}

	count := jm.GetActiveJobCount()
	if count != 2 {
		t.Errorf("GetActiveJobCount() = %d, want 2", count)
	}
}

func TestActiveJob_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		active bool
	}{
		{"running is active", JobRunning, true},
		{"pending is not active", JobPending, false},
		{"completed is not active", JobCompleted, false},
		{"failed is not active", JobFailed, false},
		{"cancelled is not active", JobCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aj := &ActiveJob{
				ID:     "test-id",
				Status: tt.status,
			}

			if got := aj.IsActive(); got != tt.active {
				t.Errorf("IsActive() = %v, want %v", got, tt.active)
			}
		})
	}
}

func TestActiveJob_CancelJob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	aj := &ActiveJob{
		ID:     "test-id",
		Status: JobRunning,
		Cancel: cancel,
		ctx:    ctx,
	}

	aj.CancelJob()

	if aj.Status != JobCancelled {
		t.Errorf("Status should be cancelled, got %v", aj.Status)
	}

	// Context should be cancelled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled")
	}
}

func TestJobStatus_Validation(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{"pending", true},
		{"running", true},
		{"completed", true},
		{"failed", true},
		{"cancelled", true},
		{"invalid", false},
		{"", false},
		{"PENDING", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := IsValidJobStatus(tt.status); got != tt.valid {
				t.Errorf("IsValidJobStatus(%q) = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}

func TestValidJobStatuses(t *testing.T) {
	statuses := ValidJobStatuses()
	expected := []string{"pending", "running", "completed", "failed", "cancelled"}

	if len(statuses) != len(expected) {
		t.Errorf("ValidJobStatuses() returned %d statuses, want %d", len(statuses), len(expected))
	}

	for i, status := range expected {
		if i >= len(statuses) || statuses[i] != status {
			t.Errorf("ValidJobStatuses()[%d] = %v, want %v", i, statuses[i], status)
		}
	}
}

func TestGenerateJobID(t *testing.T) {
	id1 := GenerateJobID()
	id2 := GenerateJobID()

	if id1 == "" {
		t.Error("GenerateJobID() returned empty string")
	}

	if id1 == id2 {
		t.Error("GenerateJobID() returned duplicate IDs")
	}
}

func TestJobManager_NilDB(t *testing.T) {
	jm := NewJobManager(nil, nil)

	_, err := jm.CreateJob("executor", "test", nil)
	if err == nil || err.Error() != "database not initialized" {
		t.Errorf("CreateJob() with nil db should return 'database not initialized' error, got: %v", err)
	}

	_, err = jm.ListJobs("")
	if err == nil || err.Error() != "database not initialized" {
		t.Errorf("ListJobs() with nil db should return 'database not initialized' error, got: %v", err)
	}

	_, err = jm.GetJob("test-id")
	if err == nil || err.Error() != "database not initialized" {
		t.Errorf("GetJob() with nil db should return 'database not initialized' error, got: %v", err)
	}
}
