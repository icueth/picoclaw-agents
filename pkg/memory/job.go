// Package memory provides concept and job tracking for work continuity.
package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/db"
)

// JobStatus represents the lifecycle status of a job.
type JobStatus string

const (
	// JobPending - Created but not started
	JobPending JobStatus = "pending"
	// JobRunning - Currently executing
	JobRunning JobStatus = "running"
	// JobCompleted - Finished successfully
	JobCompleted JobStatus = "completed"
	// JobFailed - Error occurred
	JobFailed JobStatus = "failed"
	// JobCancelled - Manually stopped
	JobCancelled JobStatus = "cancelled"
)

// ValidJobStatuses returns all valid job statuses.
func ValidJobStatuses() []string {
	return []string{
		string(JobPending),
		string(JobRunning),
		string(JobCompleted),
		string(JobFailed),
		string(JobCancelled),
	}
}

// IsValidJobStatus checks if a status is valid.
func IsValidJobStatus(status string) bool {
	for _, s := range ValidJobStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// JobContext stores the execution context for a job.
type JobContext struct {
	WorkingDirectory string                 `json:"working_directory,omitempty"`
	Files            []string               `json:"files,omitempty"`
	Variables        map[string]interface{} `json:"variables,omitempty"`
	ParentContext    map[string]interface{} `json:"parent_context,omitempty"`
}

// ActiveJob represents a job that is currently executing in memory.
type ActiveJob struct {
	ID      string
	Role    string
	Task    string
	Status  JobStatus
	Context JobContext
	Cancel  context.CancelFunc
	ctx     context.Context
}

// IsActive returns true if the job is currently running.
func (aj *ActiveJob) IsActive() bool {
	return aj.Status == JobRunning
}

// CancelJob cancels the job execution.
func (aj *ActiveJob) CancelJob() {
	if aj.Cancel != nil {
		aj.Cancel()
	}
	aj.Status = JobCancelled
}

// Job represents a job with full context for resumption.
type Job struct {
	ID          string                 `json:"id"`
	ConceptID   *string                `json:"concept_id,omitempty"`
	Role        string                 `json:"role"`
	Task        string                 `json:"task"`
	Status      JobStatus              `json:"status"`
	Context     JobContext             `json:"context"`
	Result      *string                `json:"result,omitempty"`
	ParentJobID *string                `json:"parent_job_id,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// JobManager handles job tracking with persistence.
type JobManager struct {
	db         *db.DB
	config     *config.JobConfig
	activeJobs map[string]*ActiveJob
	mu         sync.RWMutex
}

// NewJobManager creates a new JobManager instance.
func NewJobManager(database *db.DB, cfg *config.JobConfig) *JobManager {
	if cfg == nil {
		cfg = &config.JobConfig{
			DefaultTimeout:       3600,
			MaxConcurrent:        10,
			RetentionDays:        30,
			CleanupIntervalHours: 24,
		}
	}

	return &JobManager{
		db:         database,
		config:     cfg,
		activeJobs: make(map[string]*ActiveJob),
	}
}

// CreateJob creates a new job with the given role, task, and context.
func (jm *JobManager) CreateJob(role, task string, context map[string]interface{}) (string, error) {
	if jm.db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	// Build job context
	jobCtx := JobContext{
		Variables:     make(map[string]interface{}),
		ParentContext: make(map[string]interface{}),
	}

	if context != nil {
		// Extract known fields
		if wd, ok := context["working_directory"].(string); ok {
			jobCtx.WorkingDirectory = wd
		}
		if files, ok := context["files"].([]string); ok {
			jobCtx.Files = files
		}
		if vars, ok := context["variables"].(map[string]interface{}); ok {
			jobCtx.Variables = vars
		}
		if parent, ok := context["parent_context"].(map[string]interface{}); ok {
			jobCtx.ParentContext = parent
		}

		// Store any remaining fields in variables
		for k, v := range context {
			if k != "working_directory" && k != "files" && k != "variables" && k != "parent_context" {
				jobCtx.Variables[k] = v
			}
		}
	}

	// Serialize context
	contextJSON, err := json.Marshal(jobCtx)
	if err != nil {
		return "", fmt.Errorf("failed to marshal context: %w", err)
	}

	// Create in database
	var conceptID, parentJobID *string
	if cid, ok := context["concept_id"].(string); ok && cid != "" {
		conceptID = &cid
	}
	if pid, ok := context["parent_job_id"].(string); ok && pid != "" {
		parentJobID = &pid
	}

	dbJob, err := jm.db.CreateJob(conceptID, role, task, map[string]interface{}{
		"context": string(contextJSON),
	}, parentJobID)
	if err != nil {
		return "", fmt.Errorf("failed to create job in database: %w", err)
	}

	return dbJob.ID, nil
}

// CreateJobWithContext creates a new job with a structured context.
func (jm *JobManager) CreateJobWithContext(role, task string, ctx JobContext, conceptID, parentJobID *string) (string, error) {
	if jm.db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	contextJSON, err := json.Marshal(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to marshal context: %w", err)
	}

	dbJob, err := jm.db.CreateJob(conceptID, role, task, map[string]interface{}{
		"context": string(contextJSON),
	}, parentJobID)
	if err != nil {
		return "", fmt.Errorf("failed to create job in database: %w", err)
	}

	return dbJob.ID, nil
}

// ListJobs returns all jobs, optionally filtered by status.
// If status is empty, returns all jobs.
func (jm *JobManager) ListJobs(status string) ([]Job, error) {
	if jm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	dbJobs, err := jm.db.ListJobs("", status, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	jobs := make([]Job, len(dbJobs))
	for i, dbJob := range dbJobs {
		job, err := jm.dbJobToJob(dbJob)
		if err != nil {
			return nil, err
		}
		jobs[i] = *job
	}

	return jobs, nil
}

// ListJobsByConcept returns all jobs for a specific concept.
func (jm *JobManager) ListJobsByConcept(conceptID string) ([]Job, error) {
	if jm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	dbJobs, err := jm.db.ListJobs(conceptID, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	jobs := make([]Job, len(dbJobs))
	for i, dbJob := range dbJobs {
		job, err := jm.dbJobToJob(dbJob)
		if err != nil {
			return nil, err
		}
		jobs[i] = *job
	}

	return jobs, nil
}

// GetJob retrieves a job by ID.
func (jm *JobManager) GetJob(id string) (*Job, error) {
	if jm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	dbJob, err := jm.db.GetJob(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return jm.dbJobToJob(dbJob)
}

// UpdateJobStatus updates the status of a job.
func (jm *JobManager) UpdateJobStatus(id, status string) error {
	if jm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if !IsValidJobStatus(status) {
		return fmt.Errorf("invalid job status: %s", status)
	}

	// Update in active jobs if present
	jm.mu.Lock()
	if activeJob, ok := jm.activeJobs[id]; ok {
		activeJob.Status = JobStatus(status)
		if status == string(JobCompleted) || status == string(JobFailed) || status == string(JobCancelled) {
			delete(jm.activeJobs, id)
		}
	}
	jm.mu.Unlock()

	if err := jm.db.UpdateJobStatus(id, status); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// UpdateJobResult updates the result of a job.
func (jm *JobManager) UpdateJobResult(id, result string) error {
	if jm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := jm.db.UpdateJobResult(id, result); err != nil {
		return fmt.Errorf("failed to update job result: %w", err)
	}

	return nil
}

// UpdateJobContext updates the context of a job.
func (jm *JobManager) UpdateJobContext(id string, ctx JobContext) error {
	if jm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	contextJSON, err := json.Marshal(ctx)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	if err := jm.db.UpdateJobContext(id, map[string]interface{}{
		"context": string(contextJSON),
	}); err != nil {
		return fmt.Errorf("failed to update job context: %w", err)
	}

	return nil
}

// ResumeJob prepares a job for resumption and returns it.
func (jm *JobManager) ResumeJob(id string) (*Job, error) {
	if jm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	job, err := jm.GetJob(id)
	if err != nil {
		return nil, err
	}

	// Only pending or failed jobs can be resumed
	if job.Status != JobPending && job.Status != JobFailed && job.Status != JobCancelled {
		return nil, fmt.Errorf("cannot resume job with status: %s", job.Status)
	}

	// Update status to pending
	if err := jm.UpdateJobStatus(id, string(JobPending)); err != nil {
		return nil, err
	}

	job.Status = JobPending
	return job, nil
}

// StartJob marks a job as running and creates an active job entry.
func (jm *JobManager) StartJob(id string) (*ActiveJob, error) {
	if jm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	job, err := jm.GetJob(id)
	if err != nil {
		return nil, err
	}

	// Check max concurrent
	jm.mu.Lock()
	if jm.config.MaxConcurrent > 0 && len(jm.activeJobs) >= jm.config.MaxConcurrent {
		jm.mu.Unlock()
		return nil, fmt.Errorf("max concurrent jobs reached: %d", jm.config.MaxConcurrent)
	}

	// Create context with timeout
	timeout := time.Duration(jm.config.DefaultTimeout) * time.Second
	if timeout <= 0 {
		timeout = 3600 * time.Second // Default 1 hour
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	activeJob := &ActiveJob{
		ID:      id,
		Role:    job.Role,
		Task:    job.Task,
		Status:  JobRunning,
		Context: job.Context,
		Cancel:  cancel,
		ctx:     ctx,
	}

	jm.activeJobs[id] = activeJob
	jm.mu.Unlock()

	// Update status in database
	if err := jm.db.UpdateJobStatus(id, string(JobRunning)); err != nil {
		jm.mu.Lock()
		delete(jm.activeJobs, id)
		jm.mu.Unlock()
		cancel()
		return nil, fmt.Errorf("failed to update job status: %w", err)
	}

	return activeJob, nil
}

// KillJob cancels a running job.
func (jm *JobManager) KillJob(id string) error {
	jm.mu.Lock()
	activeJob, ok := jm.activeJobs[id]
	if ok {
		activeJob.CancelJob()
		delete(jm.activeJobs, id)
	}
	jm.mu.Unlock()

	// Update status in database
	if jm.db != nil {
		if err := jm.db.UpdateJobStatus(id, string(JobCancelled)); err != nil {
			return fmt.Errorf("failed to update job status: %w", err)
		}
	}

	if !ok {
		return fmt.Errorf("job not found or not running: %s", id)
	}

	return nil
}

// GetActiveJob returns an active job by ID.
func (jm *JobManager) GetActiveJob(id string) (*ActiveJob, bool) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	job, ok := jm.activeJobs[id]
	return job, ok
}

// ListActiveJobs returns all currently active jobs.
func (jm *JobManager) ListActiveJobs() []*ActiveJob {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	jobs := make([]*ActiveJob, 0, len(jm.activeJobs))
	for _, job := range jm.activeJobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// CleanupOldJobs removes jobs older than the specified retention days.
// If retentionDays is 0, uses the value from config.
func (jm *JobManager) CleanupOldJobs(retentionDays int) error {
	if jm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if retentionDays <= 0 {
		retentionDays = jm.config.RetentionDays
	}
	if retentionDays <= 0 {
		retentionDays = 30 // Default fallback
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	// Get all completed/failed/cancelled jobs
	statuses := []string{string(JobCompleted), string(JobFailed), string(JobCancelled)}

	for _, status := range statuses {
		jobs, err := jm.db.ListJobs("", status, "")
		if err != nil {
			return fmt.Errorf("failed to list jobs for cleanup: %w", err)
		}

		for _, job := range jobs {
			if job.CreatedAt.Before(cutoff) {
				if err := jm.db.DeleteJob(job.ID); err != nil {
					// Log error but continue
					continue
				}
			}
		}
	}

	return nil
}

// DeleteJob permanently deletes a job.
func (jm *JobManager) DeleteJob(id string) error {
	if jm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Remove from active jobs if present
	jm.mu.Lock()
	if activeJob, ok := jm.activeJobs[id]; ok {
		activeJob.CancelJob()
		delete(jm.activeJobs, id)
	}
	jm.mu.Unlock()

	if err := jm.db.DeleteJob(id); err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	return nil
}

// dbJobToJob converts a database job to a memory job.
func (jm *JobManager) dbJobToJob(dbJob *db.Job) (*Job, error) {
	job := &Job{
		ID:          dbJob.ID,
		ConceptID:   dbJob.ConceptID,
		Role:        dbJob.Role,
		Task:        dbJob.Task,
		Status:      JobStatus(dbJob.Status),
		Result:      dbJob.Result,
		ParentJobID: dbJob.ParentJobID,
		CreatedAt:   dbJob.CreatedAt,
		UpdatedAt:   dbJob.UpdatedAt,
		Metadata:    make(map[string]interface{}),
		Context: JobContext{
			Variables:     make(map[string]interface{}),
			ParentContext: make(map[string]interface{}),
		},
	}

	// Parse context from database
	if dbJob.Context != nil {
		if ctxStr, ok := dbJob.Context["context"].(string); ok && ctxStr != "" {
			if err := json.Unmarshal([]byte(ctxStr), &job.Context); err != nil {
				// If parsing fails, store in metadata
				job.Metadata["raw_context"] = ctxStr
			}
		} else {
			// Try to parse the whole context as JobContext
			ctxJSON, _ := json.Marshal(dbJob.Context)
			if err := json.Unmarshal(ctxJSON, &job.Context); err != nil {
				job.Metadata["raw_context"] = string(ctxJSON)
			}
		}
	}

	return job, nil
}

// GenerateJobID generates a new unique job ID.
func GenerateJobID() string {
	return uuid.New().String()
}

// StartCleanupRoutine starts a background routine that periodically cleans up old jobs.
func (jm *JobManager) StartCleanupRoutine(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Duration(jm.config.CleanupIntervalHours) * time.Hour
		if interval <= 0 {
			interval = 24 * time.Hour // Default daily
		}
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = jm.CleanupOldJobs(0) // Use config value
			}
		}
	}()
}

// GetJobCount returns the total number of jobs, optionally filtered by status.
func (jm *JobManager) GetJobCount(status string) (int, error) {
	if jm.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	jobs, err := jm.db.ListJobs("", status, "")
	if err != nil {
		return 0, fmt.Errorf("failed to list jobs: %w", err)
	}

	return len(jobs), nil
}

// GetActiveJobCount returns the number of currently running jobs.
func (jm *JobManager) GetActiveJobCount() int {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	return len(jm.activeJobs)
}

// SetJobVariable sets a variable in a job's context.
func (jm *JobManager) SetJobVariable(jobID string, key string, value interface{}) error {
	job, err := jm.GetJob(jobID)
	if err != nil {
		return err
	}

	job.Context.Variables[key] = value
	return jm.UpdateJobContext(jobID, job.Context)
}

// GetJobVariable gets a variable from a job's context.
func (jm *JobManager) GetJobVariable(jobID string, key string) (interface{}, error) {
	job, err := jm.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	value, ok := job.Context.Variables[key]
	if !ok {
		return nil, fmt.Errorf("variable not found: %s", key)
	}

	return value, nil
}
