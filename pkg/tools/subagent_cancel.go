package tools

import (
	"context"
	"fmt"
	"strings"
)

// SubagentCancelTool allows canceling running subagent tasks
type SubagentCancelTool struct {
	manager *SubagentManager
}

func NewSubagentCancelTool(manager *SubagentManager) *SubagentCancelTool {
	return &SubagentCancelTool{manager: manager}
}

func (t *SubagentCancelTool) Name() string {
	return "subagent_cancel"
}

func (t *SubagentCancelTool) Description() string {
	return "Cancel a running subagent task. Use this when a subagent is taking too long or needs to be stopped."
}

func (t *SubagentCancelTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"task_id": map[string]any{
				"type":        "string",
				"description": "The task ID to cancel",
			},
			"all": map[string]any{
				"type":        "bool",
				"description": "Cancel all active tasks",
			},
		},
	}
}

func (t *SubagentCancelTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.manager == nil {
		return ErrorResult("Subagent manager not configured")
	}

	taskID, _ := args["task_id"].(string)
	cancelAll, _ := args["all"].(bool)

	if cancelAll {
		return t.cancelAll()
	}

	if taskID == "" {
		return ErrorResult("task_id is required")
	}

	return t.cancelTask(taskID)
}

func (t *SubagentCancelTool) cancelTask(taskID string) *ToolResult {
	success, message := t.manager.CancelTask(taskID)

	if success {
		return &ToolResult{
			ForLLM:  fmt.Sprintf("Task %s has been cancelled", taskID),
			ForUser:  message,
			Silent:   false,
			IsError:  false,
		}
	}

	return ErrorResult(message)
}

func (t *SubagentCancelTool) cancelAll() *ToolResult {
	tasks := t.manager.ListActiveTasks()

	if len(tasks) == 0 {
		return &ToolResult{
			ForLLM:  "No active tasks to cancel",
			ForUser: "No active tasks to cancel",
			Silent:  false,
			IsError: false,
		}
	}

	var canceled []string
	for _, task := range tasks {
		success, _ := t.manager.CancelTask(task.ID)
		if success {
			canceled = append(canceled, task.ID)
		}
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("Cancelled %d tasks: %s", len(canceled), strings.Join(canceled, ", ")),
		ForUser: fmt.Sprintf("Cancelled %d tasks", len(canceled)),
		Silent:  false,
		IsError: false,
	}
}
