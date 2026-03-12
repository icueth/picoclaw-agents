package tools

import (
	"context"
	"fmt"

	"picoclaw/agent/pkg/logger"
)

// ReportProgressTool allows subagents to report their progress
type ReportProgressTool struct {
	manager *SubagentManager
}

// NewReportProgressTool creates a new ReportProgressTool
func NewReportProgressTool(manager *SubagentManager) *ReportProgressTool {
	return &ReportProgressTool{
		manager: manager,
	}
}

// Name returns the tool name
func (t *ReportProgressTool) Name() string {
	return "report_progress"
}

// Description returns the tool description
func (t *ReportProgressTool) Description() string {
	return `Report progress for the current subagent task.

This tool should be called periodically by subagents to:
1. Update progress percentage (0-100)
2. Report current activity/status message
3. Request timeout extension if needed
4. Provide estimated completion time

IMPORTANT: Call this tool every 30-60 seconds during long-running tasks
to prevent the system from thinking your task is stuck.

Examples:
- report_progress({"percent": 25, "message": "Analyzing requirements..."})
- report_progress({"percent": 50, "message": "Writing code...", "request_extend": 300})
- report_progress({"percent": 75, "message": "Testing implementation...", "eta_seconds": 120})`
}

// Parameters returns the tool parameters
func (t *ReportProgressTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"percent": map[string]any{
				"type":        "integer",
				"description": "Progress percentage (0-100)",
				"minimum":     0,
				"maximum":     100,
			},
			"message": map[string]any{
				"type":        "string",
				"description": "Current activity or status message",
			},
			"request_extend": map[string]any{
				"type":        "integer",
				"description": "Request additional seconds (only if task is extendable)",
			},
			"reason": map[string]any{
				"type":        "string",
				"description": "Reason for extension request",
			},
			"eta_seconds": map[string]any{
				"type":        "integer",
				"description": "Estimated time remaining in seconds",
			},
		},
		"required": []string{},
	}
}

// Execute runs the tool
func (t *ReportProgressTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.manager == nil {
		return ErrorResult("Subagent manager not configured")
	}

	// Get task ID from context (should be set by the subagent runner)
	taskID, ok := ctx.Value("subagent_task_id").(string)
	if !ok || taskID == "" {
		return ErrorResult("No task ID found in context - this tool can only be called from within a subagent")
	}

	// Get progress parameters
	percent := 0
	if p, ok := args["percent"].(float64); ok {
		percent = int(p)
		if percent < 0 {
			percent = 0
		}
		if percent > 100 {
			percent = 100
		}
	}

	message, _ := args["message"].(string)

	// Update progress
	success := t.manager.UpdateTaskProgress(taskID, percent, message)
	if !success {
		return ErrorResult(fmt.Sprintf("Failed to update progress for task %s - task may not be running", taskID))
	}

	// Handle extension request
	extensionResult := ""
	if extendSec, ok := args["request_extend"].(float64); ok && extendSec > 0 {
		reason, _ := args["reason"].(string)
		if reason == "" {
			reason = "Progress report requested extension"
		}

		extSuccess, extMsg := t.manager.RequestTaskExtension(taskID, int(extendSec), reason)
		if extSuccess {
			extensionResult = fmt.Sprintf(" ✅ Timeout extended by %d seconds", int(extendSec))
		} else {
			extensionResult = fmt.Sprintf(" ⚠️ Extension failed: %s", extMsg)
		}
	}

	// Handle ETA update
	if etaSec, ok := args["eta_seconds"].(float64); ok && etaSec > 0 {
		// Store ETA in task (would need to add this field)
		logger.DebugCF("report_progress", "ETA updated", map[string]any{
			"task_id":    taskID,
			"eta_seconds": etaSec,
		})
	}

	// Build response
	response := fmt.Sprintf("Progress updated: %d%%", percent)
	if message != "" {
		response += fmt.Sprintf(" - %s", message)
	}
	if extensionResult != "" {
		response += extensionResult
	}

	logger.InfoCF("report_progress", "Progress reported", map[string]any{
		"task_id": taskID,
		"percent": percent,
		"message": message,
	})

	return &ToolResult{
		ForLLM:  response,
		ForUser: "",
		Silent:  true, // Silent - don't show to user, just for LLM
		IsError: false,
	}
}

// SetTaskContext sets the task ID in context for subagent use
func SetTaskContext(ctx context.Context, taskID string) context.Context {
	return context.WithValue(ctx, "subagent_task_id", taskID)
}
