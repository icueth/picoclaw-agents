package tools

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SubagentStatusTool provides visibility into subagent tasks
type SubagentStatusTool struct {
	manager *SubagentManager
}

func NewSubagentStatusTool(manager *SubagentManager) *SubagentStatusTool {
	return &SubagentStatusTool{manager: manager}
}

func (t *SubagentStatusTool) Name() string {
	return "subagent_status"
}

func (t *SubagentStatusTool) Description() string {
	return `Get the status of subagent tasks. Use this to check if subagents are running, completed, or failed.

BEST PRACTICES:
- After spawning a subagent, check status once immediately to confirm it started
- If status is 'running', wait 15-20 seconds before checking again
- Do NOT check status more frequently than every 10 seconds
- Do NOT use 'sleep' tool between status checks - just respond to the user that you're waiting
- Typical tasks complete in 30-120 seconds
- If a task runs longer than 5 minutes, inform the user about the delay`
}

func (t *SubagentStatusTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type": "string",
				"enum":   []string{"list", "get", "active", "failed"},
				"description": "Action: list (all tasks), get (specific task), active (running only), failed (failed only)",
			},
			"task_id": map[string]any{
				"type":        "string",
				"description": "Task ID to get status (for 'get' action)",
			},
		},
		"required": []string{"action"},
	}
}

func (t *SubagentStatusTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.manager == nil {
		return ErrorResult("Subagent manager not configured")
	}

	action, _ := args["action"].(string)
	taskID, _ := args["task_id"].(string)

	switch action {
	case "list":
		return t.listTasks()
	case "get":
		return t.getTask(taskID)
	case "active":
		return t.listActiveTasks()
	case "failed":
		return t.listFailedTasks()
	default:
		return ErrorResult(fmt.Sprintf("Unknown action: %s", action))
	}
}

func (t *SubagentStatusTool) listTasks() *ToolResult {
	tasks := t.manager.ListTasks()

	if len(tasks) == 0 {
		return &ToolResult{
			ForLLM:  "No subagent tasks found",
			ForUser: "No active subagent tasks",
			Silent:  false,
			IsError: false,
		}
	}

	var sb strings.Builder
	sb.WriteString("## Subagent Tasks\n\n")
	sb.WriteString("| ID | Label | Status | Progress | Duration | Extensions |\n")
	sb.WriteString("|----|-------|--------|----------|----------|------------|\n")

	for _, task := range tasks {
		duration := ""
		if task.Started > 0 {
			elapsedMs := time.Now().UnixMilli() - task.Started
			if task.Finished > 0 {
				elapsedMs = task.Finished - task.Started
			}
			elapsedSec := elapsedMs / 1000
			if elapsedSec < 60 {
				duration = fmt.Sprintf("%ds", elapsedSec)
			} else {
				duration = fmt.Sprintf("%dm%ds", elapsedSec/60, elapsedSec%60)
			}
		}

		progress := fmt.Sprintf("%d%%", task.ProgressPercent)
		extensions := "-"
		if task.IsExtendable {
			extensions = fmt.Sprintf("%d/%d", task.ExtensionsUsed, task.MaxExtensions)
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			task.ID, task.Label, task.Status, progress, duration, extensions))
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Found %d subagent tasks", len(tasks)),
		Silent:  false,
		IsError: false,
	}
}

func (t *SubagentStatusTool) getTask(taskID string) *ToolResult {
	if taskID == "" {
		return ErrorResult("task_id is required for 'get' action")
	}

	status := t.manager.GetTaskStatus(taskID)

	if exists, ok := status["exists"].(bool); !ok || !exists {
		return ErrorResult(fmt.Sprintf("Task not found: %s", taskID))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Task: %s\n\n", taskID))
	sb.WriteString(fmt.Sprintf("- **Label**: %v\n", status["label"]))
	sb.WriteString(fmt.Sprintf("- **Role**: %v\n", status["role"]))
	sb.WriteString(fmt.Sprintf("- **Status**: %v\n", status["status"]))

	// Progress information
	if progress, ok := status["progress_percent"].(int); ok && progress > 0 {
		sb.WriteString(fmt.Sprintf("- **Progress**: %d%%\n", progress))
	}
	if msg, ok := status["progress_message"].(string); ok && msg != "" {
		sb.WriteString(fmt.Sprintf("- **Current Activity**: %s\n", msg))
	}

	// Duration and timing
	sb.WriteString(fmt.Sprintf("- **Duration**: %v\n", status["duration"]))

	// Timeout information
	if timeoutAt, ok := status["timeout_at"].(int64); ok && timeoutAt > 0 {
		remaining := time.Until(time.UnixMilli(timeoutAt))
		if remaining > 0 {
			sb.WriteString(fmt.Sprintf("- **Time Remaining**: %s\n", formatDuration(remaining)))
		} else {
			sb.WriteString("- **Time Remaining**: ⚠️ EXCEEDED\n")
		}
	}

	// Extension information
	if isExtendable, ok := status["is_extendable"].(bool); ok && isExtendable {
		extensions := status["extensions_used"].(int)
		maxExtensions := status["max_extensions"].(int)
		sb.WriteString(fmt.Sprintf("- **Extensions**: %d/%d used\n", extensions, maxExtensions))
	}

	if errorMsg, ok := status["error"].(string); ok && errorMsg != "" {
		sb.WriteString(fmt.Sprintf("- **Error**: %s\n", errorMsg))
	}

	// Check if task is still active
	if isActive, ok := status["is_active"].(bool); ok && isActive {
		sb.WriteString("\n⚠️ Task is still running\n")

		// Add elapsed time guidance
		if elapsed, ok := status["elapsed_ms"].(int64); ok && elapsed > 0 {
			elapsedSec := elapsed / 1000
			if elapsedSec < 60 {
				sb.WriteString(fmt.Sprintf("⏱️ Running for %d seconds\n", elapsedSec))
			} else {
				sb.WriteString(fmt.Sprintf("⏱️ Running for %d minutes %d seconds\n", elapsedSec/60, elapsedSec%60))
			}

			// Add guidance based on elapsed time
			if elapsedSec > 300 {
				sb.WriteString("\n📝 This task is taking longer than expected. Consider:\n")
				sb.WriteString("- Using 'job_health' tool for detailed analysis\n")
				sb.WriteString("- Extending timeout if the task is still making progress\n")
				sb.WriteString("- Informing the user about the delay\n")
			} else if elapsedSec > 60 {
				sb.WriteString("\n⏳ Task is taking some time. Typical completion: 30-120 seconds.\n")
			}
		}
		sb.WriteString("\n💡 Next step: Wait 15-20 seconds, then check status again.\n")
	}

	// Check if task failed
	if isFailed, ok := status["is_failed"].(bool); ok && isFailed {
		sb.WriteString("\n❌ Task failed\n")
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Task %s: %s", taskID, status["status"]),
		Silent:  false,
		IsError: false,
	}
}


func (t *SubagentStatusTool) listActiveTasks() *ToolResult {
	tasks := t.manager.ListActiveTasks()

	if len(tasks) == 0 {
		return &ToolResult{
			ForLLM:  "No active subagent tasks",
			ForUser: "No active subagent tasks",
			Silent:  false,
			IsError: false,
		}
	}

	var sb strings.Builder
	sb.WriteString("## Active Subagent Tasks\n\n")
	sb.WriteString("| ID | Label | Status |\n")
	sb.WriteString("|----|-------|--------|\n")

	for _, task := range tasks {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", task.ID, task.Label, task.Status))
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Found %d active tasks", len(tasks)),
		Silent:  false,
		IsError: false,
	}
}

func (t *SubagentStatusTool) listFailedTasks() *ToolResult {
	allTasks := t.manager.ListTasks()
	
	var failedTasks []*SubagentTask
	for _, task := range allTasks {
		if task.Status == "failed" || task.Status == "canceled" {
			failedTasks = append(failedTasks, task)
		}
	}

	if len(failedTasks) == 0 {
		return &ToolResult{
			ForLLM:  "No failed subagent tasks",
			ForUser: "No failed subagent tasks",
			Silent:  false,
			IsError: false,
		}
	}

	var sb strings.Builder
	sb.WriteString("## Failed/Canceled Subagent Tasks\n\n")
	sb.WriteString("| ID | Label | Status | Error |\n")
	sb.WriteString("|----|-------|--------|-------|\n")

	for _, task := range failedTasks {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", 
			task.ID, task.Label, task.Status, task.Error))
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Found %d failed tasks", len(failedTasks)),
		Silent:  false,
		IsError: false,
	}
}
