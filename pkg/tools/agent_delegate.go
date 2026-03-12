package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"picoclaw/agent/pkg/agentcomm"
)

// MessengerSender is an interface for sending messages between agents
type MessengerSender interface {
	SendDirect(ctx context.Context, to string, msg agentcomm.AgentMessage) error
	Broadcast(ctx context.Context, msg agentcomm.AgentMessage) error
}

// SyncTaskRunner can run tasks synchronously, blocking until result.
type SyncTaskRunner interface {
	SpawnAndWait(ctx context.Context, task, label, agentID, model, originChannel, originChatID string) (string, error)
}

// DelegateTool allows agents to delegate tasks to other agents.
// It runs synchronously: blocks until the subagent completes and returns the result.
type DelegateTool struct {
	registry interface {
		GetAgentInfo(agentID string) (agentcomm.AgentInfo, bool)
		GetAllAgents() []agentcomm.AgentInfo
	}
	runner        SyncTaskRunner
	originChannel string
	originChatID  string
}

// NewDelegateTool creates a new DelegateTool
func NewDelegateTool(registry interface {
	GetAgentInfo(agentID string) (agentcomm.AgentInfo, bool)
	GetAllAgents() []agentcomm.AgentInfo
}, runner SyncTaskRunner) *DelegateTool {
	return &DelegateTool{
		registry:      registry,
		runner:        runner,
		originChannel: "cli",
		originChatID:  "direct",
	}
}

func (t *DelegateTool) Name() string {
	return "delegate"
}

func (t *DelegateTool) Description() string {
	return "Delegate a task to another agent. Use this when you want another agent to handle a specific task. The target agent will execute the task and return results."
}

func (t *DelegateTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"task": map[string]any{
				"type":        "string",
				"description": "The task to delegate to the target agent",
			},
			"target_agent": map[string]any{
				"type":        "string",
				"description": "The target agent ID to delegate the task to",
			},
		},
		"required": []string{"task", "target_agent"},
	}
}

func (t *DelegateTool) SetContext(channel, chatID string) {
	t.originChannel = channel
	t.originChatID = chatID
}

func (t *DelegateTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	task, ok := args["task"].(string)
	if !ok || strings.TrimSpace(task) == "" {
		return ErrorResult("task is required and must be a non-empty string")
	}

	targetAgent, ok := args["target_agent"].(string)
	if !ok || strings.TrimSpace(targetAgent) == "" {
		return ErrorResult("target_agent is required")
	}

	// Check if target agent exists and get its model
	agentInfo, ok := t.registry.GetAgentInfo(targetAgent)
	if !ok {
		return ErrorResult(fmt.Sprintf("Agent '%s' not found. Use list_agents to see available agents.", targetAgent))
	}

	if t.runner == nil {
		return ErrorResult("delegation not available: task runner not configured")
	}

	// Use a generous timeout so the subagent has enough time to complete.
	delegateCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Run synchronously — block until subagent finishes.
	// Use the target agent's model so the correct LLM handles the task.
	result, err := t.runner.SpawnAndWait(delegateCtx, task, targetAgent, targetAgent, agentInfo.Model, t.originChannel, t.originChatID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Delegation to '%s' failed: %v", targetAgent, err))
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("Agent '%s' completed the task.\n\nResult:\n%s", targetAgent, result),
		ForUser: result,
		Silent:  false,
		IsError: false,
	}
}

// ListAgentsTool allows agents to see other available agents
type ListAgentsTool struct {
	registry interface {
		GetAllAgents() []agentcomm.AgentInfo
	}
}

func NewListAgentsTool(registry interface {
	GetAllAgents() []agentcomm.AgentInfo
}) *ListAgentsTool {
	return &ListAgentsTool{registry: registry}
}

func (t *ListAgentsTool) Name() string {
	return "list_agents"
}

func (t *ListAgentsTool) Description() string {
	return "List all available agents. Use this to see which agents you can delegate tasks to."
}

func (t *ListAgentsTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"verbose": map[string]any{
				"type":        "bool",
				"description": "Show detailed information about each agent",
			},
		},
	}
}

func (t *ListAgentsTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	agents := t.registry.GetAllAgents()

	if len(agents) == 0 {
		return &ToolResult{
			ForLLM:  "No agents found",
			ForUser: "No agents available",
			Silent:  false,
			IsError: false,
		}
	}

	verbose, _ := args["verbose"].(bool)

	var sb strings.Builder
	sb.WriteString("## Available Agents\n\n")
	sb.WriteString("| ID | Name | Type | Status |\n")
	sb.WriteString("|----|------|------|--------|\n")

	for _, a := range agents {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			a.ID, a.Name, a.Type, a.Status))
	}

	if verbose {
		sb.WriteString("\n### Detailed Info\n")
		for _, a := range agents {
			sb.WriteString(fmt.Sprintf("\n**%s** (%s)\n", a.Name, a.ID))
			sb.WriteString(fmt.Sprintf("- Type: %s\n", a.Type))
			sb.WriteString(fmt.Sprintf("- Status: %s\n", a.Status))
			if len(a.Capabilities) > 0 {
				sb.WriteString(fmt.Sprintf("- Capabilities: %s\n", strings.Join(a.Capabilities, ", ")))
			}
		}
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Found %d agents", len(agents)),
		Silent:  false,
		IsError: false,
	}
}

// AskAgentTool allows agents to ask questions to other agents
type AskAgentTool struct {
	registry interface {
		GetAgentInfo(agentID string) (agentcomm.AgentInfo, bool)
	}
	messenger     MessengerSender
	runner        SyncTaskRunner
	originChannel string
	originChatID  string
}

func NewAskAgentTool(registry interface {
	GetAgentInfo(agentID string) (agentcomm.AgentInfo, bool)
}, messenger MessengerSender, runner SyncTaskRunner) *AskAgentTool {
	return &AskAgentTool{
		registry:      registry,
		messenger:     messenger,
		runner:        runner,
		originChannel: "cli",
		originChatID:  "direct",
	}
}

func (t *AskAgentTool) Name() string {
	return "ask_agent"
}

func (t *AskAgentTool) Description() string {
	return "Ask a question to another agent and get a response. Use this for consultation or when you need expertise from another agent."
}

func (t *AskAgentTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"question": map[string]any{
				"type":        "string",
				"description": "The question to ask the agent",
			},
			"agent_id": map[string]any{
				"type":        "string",
				"description": "The agent ID to ask",
			},
		},
		"required": []string{"question", "agent_id"},
	}
}

func (t *AskAgentTool) SetContext(channel, chatID string) {
	t.originChannel = channel
	t.originChatID = chatID
}

func (t *AskAgentTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	question, ok := args["question"].(string)
	if !ok || strings.TrimSpace(question) == "" {
		return ErrorResult("question is required")
	}

	agentID, ok := args["agent_id"].(string)
	if !ok || strings.TrimSpace(agentID) == "" {
		return ErrorResult("agent_id is required")
	}

	// Check if agent exists
	_, ok = t.registry.GetAgentInfo(agentID)
	if !ok {
		return ErrorResult(fmt.Sprintf("Agent '%s' not found", agentID))
	}

	// Primary: run synchronously via SubagentManager
	if t.runner != nil {
		agentInfo, _ := t.registry.GetAgentInfo(agentID)
		result, err := t.runner.SpawnAndWait(ctx, question, agentID, agentID, agentInfo.Model, t.originChannel, t.originChatID)
		if err != nil {
			return ErrorResult(fmt.Sprintf("Failed to ask '%s': %v", agentID, err))
		}
		return &ToolResult{
			ForLLM:  fmt.Sprintf("Response from %s:\n%s", agentID, result),
			ForUser: fmt.Sprintf("Got response from %s", agentID),
			Silent:  false,
			IsError: false,
		}
	}

	// Fallback: send question via messenger (fire-and-forget)
	if t.messenger != nil {
		msg := agentcomm.NewAgentMessage(
			"main",
			agentID,
			agentcomm.MsgRequest,
			fmt.Sprintf("[QUESTION]\n\n%s", question),
			"ask-"+time.Now().Format("20060102150405"),
		)
		t.messenger.SendDirect(ctx, agentID, msg)
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("Question sent to %s: %s", agentID, question),
		ForUser: fmt.Sprintf("Question sent to %s", agentID),
		Silent:  false,
		IsError: false,
	}
}
