package tools

import (
	"context"
	"fmt"
	"strings"

	"picoclaw/agent/pkg/memory"
)

// ListJobsTool lists jobs from the job manager
type ListJobsTool struct {
	jobManager *memory.JobManager
}

// NewListJobsTool creates a new ListJobsTool
func NewListJobsTool(jm *memory.JobManager) *ListJobsTool {
	return &ListJobsTool{jobManager: jm}
}

// Name returns the tool name
func (t *ListJobsTool) Name() string {
	return "list_jobs"
}

// Description returns the tool description
func (t *ListJobsTool) Description() string {
	return `List jobs tracked by the job manager.
Can filter by status to show only pending, running, completed, failed, or cancelled jobs.`
}

// Parameters returns the tool parameters
func (t *ListJobsTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status": map[string]any{
				"type":        "string",
				"enum":        []string{"", "pending", "running", "completed", "failed", "cancelled"},
				"description": "Filter by status (empty for all jobs)",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of jobs to return (default: 50)",
			},
		},
	}
}

// Execute runs the tool
func (t *ListJobsTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.jobManager == nil {
		return ErrorResult("Job manager not configured")
	}

	status, _ := args["status"].(string)
	limit := 50
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	jobs, err := t.jobManager.ListJobs(status)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to list jobs: %v", err))
	}

	if len(jobs) == 0 {
		return &ToolResult{
			ForLLM:  "No jobs found",
			ForUser: "No jobs found",
			Silent:  false,
			IsError: false,
		}
	}

	// Apply limit
	if len(jobs) > limit {
		jobs = jobs[:limit]
	}

	var sb strings.Builder
	if status == "" {
		sb.WriteString("## All Jobs\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("## Jobs (%s)\n\n", status))
	}

	sb.WriteString("| ID | Role | Status | Created | Result |\n")
	sb.WriteString("|----|------|--------|---------|--------|\n")

	for _, job := range jobs {
		result := "-"
		if job.Result != nil {
			result = truncateString(*job.Result, 30)
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			truncateString(job.ID, 8),
			job.Role,
			job.Status,
			job.CreatedAt.Format("2006-01-02 15:04"),
			result))
	}

	if len(jobs) >= limit {
		sb.WriteString(fmt.Sprintf("\n*Showing first %d jobs*\n", limit))
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Found %d jobs", len(jobs)),
		Silent:  false,
		IsError: false,
	}
}

// GetJobStatusTool gets detailed status of a specific job
type GetJobStatusTool struct {
	jobManager *memory.JobManager
}

// NewGetJobStatusTool creates a new GetJobStatusTool
func NewGetJobStatusTool(jm *memory.JobManager) *GetJobStatusTool {
	return &GetJobStatusTool{jobManager: jm}
}

// Name returns the tool name
func (t *GetJobStatusTool) Name() string {
	return "get_job_status"
}

// Description returns the tool description
func (t *GetJobStatusTool) Description() string {
	return "Get detailed status and information about a specific job by ID"
}

// Parameters returns the tool parameters
func (t *GetJobStatusTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"job_id": map[string]any{
				"type":        "string",
				"description": "The job ID to get status for",
			},
		},
		"required": []string{"job_id"},
	}
}

// Execute runs the tool
func (t *GetJobStatusTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	jobID, ok := args["job_id"].(string)
	if !ok || jobID == "" {
		return ErrorResult("job_id is required")
	}

	if t.jobManager == nil {
		return ErrorResult("Job manager not configured")
	}

	job, err := t.jobManager.GetJob(jobID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get job: %v", err))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Job: %s\n\n", job.ID))
	sb.WriteString(fmt.Sprintf("**Role**: %s\n\n", job.Role))
	sb.WriteString(fmt.Sprintf("**Task**: %s\n\n", job.Task))
	sb.WriteString(fmt.Sprintf("**Status**: %s\n\n", job.Status))
	sb.WriteString(fmt.Sprintf("**Created**: %s\n\n", job.CreatedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Updated**: %s\n\n", job.UpdatedAt.Format("2006-01-02 15:04:05")))

	if job.ConceptID != nil {
		sb.WriteString(fmt.Sprintf("**Concept ID**: %s\n\n", *job.ConceptID))
	}

	if job.ParentJobID != nil {
		sb.WriteString(fmt.Sprintf("**Parent Job**: %s\n\n", *job.ParentJobID))
	}

	if job.Result != nil && *job.Result != "" {
		sb.WriteString(fmt.Sprintf("**Result**:\n```\n%s\n```\n\n", truncateString(*job.Result, 500)))
	}

	// Show context if available
	if len(job.Context.Variables) > 0 {
		sb.WriteString("**Context Variables**:\n")
		for k, v := range job.Context.Variables {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
		}
		sb.WriteString("\n")
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Job %s status: %s", truncateString(job.ID, 8), job.Status),
		Silent:  false,
		IsError: false,
	}
}

// KillJobTool cancels/kills a running job
type KillJobTool struct {
	jobManager *memory.JobManager
}

// NewKillJobTool creates a new KillJobTool
func NewKillJobTool(jm *memory.JobManager) *KillJobTool {
	return &KillJobTool{jobManager: jm}
}

// Name returns the tool name
func (t *KillJobTool) Name() string {
	return "kill_job"
}

// Description returns the tool description
func (t *KillJobTool) Description() string {
	return "Cancel/kill a running or pending job by ID"
}

// Parameters returns the tool parameters
func (t *KillJobTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"job_id": map[string]any{
				"type":        "string",
				"description": "The job ID to cancel",
			},
		},
		"required": []string{"job_id"},
	}
}

// Execute runs the tool
func (t *KillJobTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	jobID, ok := args["job_id"].(string)
	if !ok || jobID == "" {
		return ErrorResult("job_id is required")
	}

	if t.jobManager == nil {
		return ErrorResult("Job manager not configured")
	}

	// First check if job exists and is active
	job, err := t.jobManager.GetJob(jobID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get job: %v", err))
	}

	if job.Status != memory.JobRunning && job.Status != memory.JobPending {
		return ErrorResult(fmt.Sprintf("Job is not active (status: %s)", job.Status))
	}

	// Kill the job
	if err := t.jobManager.KillJob(jobID); err != nil {
		return ErrorResult(fmt.Sprintf("Failed to kill job: %v", err))
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("Job %s has been cancelled", jobID),
		ForUser: fmt.Sprintf("Job %s cancelled", truncateString(jobID, 8)),
		Silent:  false,
		IsError: false,
	}
}

// ResumeJobTool resumes a suspended/failed job
type ResumeJobTool struct {
	jobManager *memory.JobManager
}

// NewResumeJobTool creates a new ResumeJobTool
func NewResumeJobTool(jm *memory.JobManager) *ResumeJobTool {
	return &ResumeJobTool{jobManager: jm}
}

// Name returns the tool name
func (t *ResumeJobTool) Name() string {
	return "resume_job"
}

// Description returns the tool description
func (t *ResumeJobTool) Description() string {
	return `Resume a pending, failed, or cancelled job by ID.
The job will be reset to pending status and can be restarted.`
}

// Parameters returns the tool parameters
func (t *ResumeJobTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"job_id": map[string]any{
				"type":        "string",
				"description": "The job ID to resume",
			},
		},
		"required": []string{"job_id"},
	}
}

// Execute runs the tool
func (t *ResumeJobTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	jobID, ok := args["job_id"].(string)
	if !ok || jobID == "" {
		return ErrorResult("job_id is required")
	}

	if t.jobManager == nil {
		return ErrorResult("Job manager not configured")
	}

	// Resume the job
	job, err := t.jobManager.ResumeJob(jobID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to resume job: %v", err))
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("Job %s has been resumed and is now %s", jobID, job.Status),
		ForUser: fmt.Sprintf("Job %s resumed", truncateString(jobID, 8)),
		Silent:  false,
		IsError: false,
	}
}

// Helper function to truncate strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
