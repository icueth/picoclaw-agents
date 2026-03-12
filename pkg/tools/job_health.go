package tools

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// JobHealthStatus represents the health status of a job
type JobHealthStatus string

const (
	JobHealthHealthy   JobHealthStatus = "healthy"   // Job is running normally
	JobHealthWarning   JobHealthStatus = "warning"   // Job is slow but progressing
	JobHealthStuck     JobHealthStatus = "stuck"     // Job appears stuck (no progress)
	JobHealthCritical  JobHealthStatus = "critical"  // Job near timeout
	JobHealthTimeout   JobHealthStatus = "timeout"   // Job exceeded timeout
	JobHealthCompleted JobHealthStatus = "completed" // Job finished successfully
)

// JobHealthTool provides detailed health monitoring for jobs/subagent tasks
type JobHealthTool struct {
	manager *SubagentManager
}

// NewJobHealthTool creates a new JobHealthTool
func NewJobHealthTool(manager *SubagentManager) *JobHealthTool {
	return &JobHealthTool{
		manager: manager,
	}
}

// Name returns the tool name
func (t *JobHealthTool) Name() string {
	return "job_health"
}

// Description returns the tool description
func (t *JobHealthTool) Description() string {
	return `Get detailed health information about subagent jobs/tasks.

This tool provides comprehensive health monitoring including:
- Current status and progress
- Health assessment (healthy, warning, stuck, critical)
- Time elapsed and remaining
- Timeout information and extension status
- Recommendations for actions

Use this when you need to understand why a job is taking long or if it's stuck.`
}

// Parameters returns the tool parameters
func (t *JobHealthTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type": "string",
				"enum": []string{"check", "list", "extend", "kill"},
				"description": "Action to perform: check (specific job), list (all jobs), extend (add time), kill (cancel)",
			},
			"task_id": map[string]any{
				"type":        "string",
				"description": "Task ID to check/extend/kill (required for check, extend, kill actions)",
			},
			"additional_seconds": map[string]any{
				"type":        "integer",
				"description": "Additional seconds to extend (for extend action)",
			},
			"reason": map[string]any{
				"type":        "string",
				"description": "Reason for extension (for extend action)",
			},
		},
		"required": []string{"action"},
	}
}

// Execute runs the tool
func (t *JobHealthTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.manager == nil {
		return ErrorResult("Subagent manager not configured")
	}

	action, _ := args["action"].(string)

	switch action {
	case "check":
		return t.checkJob(args)
	case "list":
		return t.listJobs()
	case "extend":
		return t.extendJob(args)
	case "kill":
		return t.killJob(args)
	default:
		return ErrorResult(fmt.Sprintf("Unknown action: %s", action))
	}
}

func (t *JobHealthTool) checkJob(args map[string]any) *ToolResult {
	taskID, _ := args["task_id"].(string)
	if taskID == "" {
		return ErrorResult("task_id is required for 'check' action")
	}

	// Get task status
	status := t.manager.GetTaskStatus(taskID)
	if exists, ok := status["exists"].(bool); !ok || !exists {
		return ErrorResult(fmt.Sprintf("Task not found: %s", taskID))
	}

	// Get detailed health info
	healthInfo := t.formatHealthStatus(status)

	// Build detailed report
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Job Health Report: %s\n\n", taskID))

	// Basic info
	sb.WriteString("### Basic Information\n")
	sb.WriteString(fmt.Sprintf("- **Status**: %v\n", status["status"]))
	sb.WriteString(fmt.Sprintf("- **Role**: %v\n", status["role"]))
	sb.WriteString(fmt.Sprintf("- **Label**: %v\n", status["label"]))

	// Progress info
	if progress, ok := status["progress_percent"].(int); ok && progress > 0 {
		sb.WriteString(fmt.Sprintf("- **Progress**: %d%%\n", progress))
	}
	if msg, ok := status["progress_message"].(string); ok && msg != "" {
		sb.WriteString(fmt.Sprintf("- **Current Activity**: %s\n", msg))
	}

	// Timing info
	sb.WriteString("\n### Timing\n")
	if started, ok := status["started"].(int64); ok && started > 0 {
		elapsed := time.Since(time.UnixMilli(started))
		sb.WriteString(fmt.Sprintf("- **Elapsed**: %s\n", formatDuration(elapsed)))
	}
	if timeoutAt, ok := status["timeout_at"].(int64); ok && timeoutAt > 0 {
		remaining := time.Until(time.UnixMilli(timeoutAt))
		if remaining > 0 {
			sb.WriteString(fmt.Sprintf("- **Remaining**: %s\n", formatDuration(remaining)))
		} else {
			sb.WriteString("- **Remaining**: ⚠️ EXCEEDED\n")
		}
	}

	// Extension info
	if isExtendable, ok := status["is_extendable"].(bool); ok && isExtendable {
		extensions := status["extensions_used"].(int)
		maxExtensions := status["max_extensions"].(int)
		sb.WriteString(fmt.Sprintf("- **Extensions**: %d/%d used\n", extensions, maxExtensions))
	}

	// Health assessment
	if healthInfo != "" {
		sb.WriteString("\n### Health Assessment\n")
		sb.WriteString(healthInfo)
	}

	// Recommendations
	sb.WriteString("\n### Recommendations\n")
	sb.WriteString(t.getRecommendations(status))

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Job health check for %s completed", taskID),
		Silent:  false,
		IsError: false,
	}
}

func (t *JobHealthTool) listJobs() *ToolResult {
	tasks := t.manager.ListTasks()

	if len(tasks) == 0 {
		return &ToolResult{
			ForLLM:  "No jobs found",
			ForUser: "No active jobs",
			Silent:  false,
			IsError: false,
		}
	}

	var sb strings.Builder
	sb.WriteString("## All Jobs\n\n")
	sb.WriteString("| ID | Role | Status | Progress | Extensions |\n")
	sb.WriteString("|----|------|--------|----------|------------|\n")

	for _, task := range tasks {
		progress := fmt.Sprintf("%d%%", task.ProgressPercent)
		extensions := "-"
		if task.IsExtendable {
			extensions = fmt.Sprintf("%d/%d", task.ExtensionsUsed, task.MaxExtensions)
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			task.ID, task.Role, task.Status, progress, extensions))
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Found %d jobs", len(tasks)),
		Silent:  false,
		IsError: false,
	}
}

func (t *JobHealthTool) extendJob(args map[string]any) *ToolResult {
	taskID, _ := args["task_id"].(string)
	if taskID == "" {
		return ErrorResult("task_id is required for 'extend' action")
	}

	additionalSeconds := 0
	if sec, ok := args["additional_seconds"].(float64); ok {
		additionalSeconds = int(sec)
	}
	if additionalSeconds <= 0 {
		additionalSeconds = 300 // Default 5 minutes
	}

	reason, _ := args["reason"].(string)
	if reason == "" {
		reason = "Manual extension via job_health tool"
	}

	success, msg := t.manager.RequestTaskExtension(taskID, additionalSeconds, reason)
	if !success {
		return ErrorResult(msg)
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("✅ %s", msg),
		ForUser: msg,
		Silent:  false,
		IsError: false,
	}
}

func (t *JobHealthTool) killJob(args map[string]any) *ToolResult {
	taskID, _ := args["task_id"].(string)
	if taskID == "" {
		return ErrorResult("task_id is required for 'kill' action")
	}

	success, msg := t.manager.CancelTask(taskID)
	if !success {
		return ErrorResult(msg)
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("✅ Job %s cancelled: %s", taskID, msg),
		ForUser: fmt.Sprintf("Job %s cancelled", taskID),
		Silent:  false,
		IsError: false,
	}
}

func (t *JobHealthTool) getRecommendations(status map[string]any) string {
	var recommendations []string

	taskStatus, _ := status["status"].(string)
	isExtendable, _ := status["is_extendable"].(bool)

	switch taskStatus {
	case "running":
		// Check if near timeout
		if timeoutAt, ok := status["timeout_at"].(int64); ok && timeoutAt > 0 {
			remaining := time.Until(time.UnixMilli(timeoutAt))
			if remaining < 2*time.Minute {
				recommendations = append(recommendations, "⚠️ Job is near timeout. Consider extending if more time is needed.")
			}
		}

		// Check progress
		if lastProgress, ok := status["last_progress"].(int64); ok && lastProgress > 0 {
			timeSinceProgress := time.Since(time.UnixMilli(lastProgress))
			if timeSinceProgress > 2*time.Minute {
				recommendations = append(recommendations, "🔄 No progress reported for 2+ minutes. Job may be stuck.")
			}
		}

		if isExtendable {
			recommendations = append(recommendations, "💡 Use 'extend' action to add more time if needed.")
		}

		recommendations = append(recommendations, "⏱️ Wait and check again in 30 seconds, or use 'check' action to monitor.")

	case "completed":
		recommendations = append(recommendations, "✅ Job completed successfully. Check the result.")

	case "failed":
		recommendations = append(recommendations, "❌ Job failed. Check error message and consider retrying.")
		if err, ok := status["error"].(string); ok && err != "" {
			recommendations = append(recommendations, fmt.Sprintf("   Error: %s", err))
		}

	case "cancelled":
		recommendations = append(recommendations, "🛑 Job was cancelled. You can spawn a new one if needed.")

	default:
		recommendations = append(recommendations, "ℹ️ Unknown status. Use 'check' action for more details.")
	}

	return strings.Join(recommendations, "\n")
}

// formatHealthStatus formats the health status for display
func (t *JobHealthTool) formatHealthStatus(status map[string]any) string {
	taskStatus, _ := status["status"].(string)

	var statusEmoji string
	switch taskStatus {
	case "running":
		statusEmoji = "🔄"
	case "completed":
		statusEmoji = "✅"
	case "failed":
		statusEmoji = "❌"
	case "cancelled":
		statusEmoji = "🛑"
	default:
		statusEmoji = "❓"
	}

	return fmt.Sprintf("%s Status: %s", statusEmoji, taskStatus)
}

// formatDuration formats a duration for human reading
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}
