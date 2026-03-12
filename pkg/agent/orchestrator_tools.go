// A2A Orchestrator Tools
// Tools สำหรับเรียกใช้ A2A Orchestrator จาก LLM (NO SUBAGENT!)

package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"picoclaw/agent/pkg/tools"
)

// StartA2AProjectTool starts a new A2A project
type StartA2AProjectTool struct {
	orchestrator *A2AOrchestrator
}

// NewStartA2AProjectTool creates a new start A2A project tool
func NewStartA2AProjectTool(orchestrator *A2AOrchestrator) *StartA2AProjectTool {
	return &StartA2AProjectTool{orchestrator: orchestrator}
}

// Name returns the tool name
func (t *StartA2AProjectTool) Name() string {
	return "start_a2a_project"
}

// Description returns the tool description
func (t *StartA2AProjectTool) Description() string {
	return "Start a new project with A2A (Agent-to-Agent) collaboration. Provides a way to break down any complex goal (research, analysis, software, content creation, etc.) among multiple specialized agents. NO subagents are used - only real agents with their own personas."
}

// Parameters returns the tool parameters
func (t *StartA2AProjectTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Project name",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "Detailed project description outlining the problem to solve, tasks to do, or goal to achieve (e.g. 'Research US-Iran conflict' or 'Build a web app'). Do not hallucinate a tech stack unless explicitly requested.",
			},
			"priority": map[string]any{
				"type":        "string",
				"description": "Project priority",
				"enum":        []string{"low", "medium", "high", "critical"},
			},
		},
		"required": []string{"name", "description"},
	}
}

// Execute executes the tool
func (t *StartA2AProjectTool) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
	name, ok := args["name"].(string)
	if !ok || strings.TrimSpace(name) == "" {
		return tools.ErrorResult("project name is required")
	}

	description, _ := args["description"].(string)

	// Create A2A project
	project := t.orchestrator.CreateProject(name, description)

	// Start A2A orchestration
	if err := t.orchestrator.StartProject(project.ID); err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to start A2A project: %v", err))
	}

	// Get shortcut for the project
	shortcut := generateProjectShortcut(name)

	result := fmt.Sprintf("🚀 A2A Project '%s' started!\n", name)
	result += fmt.Sprintf("Project ID: %s\n", project.ID)
	result += fmt.Sprintf("Shortcut: %s (use this for quick reference)\n", shortcut)
	result += "\nA2A Workflow:\n"
	result += "1. Discovery - Agents share capabilities\n"
	result += "2. Meeting - Agents discuss and agree on tasks\n"
	result += "3. Planning - Task assignments via A2A\n"
	result += "4. Execution - Agents work sequentially (max 3 concurrent)\n"
	result += "5. Integration - Combine all work\n"
	result += "6. Validation - QA review\n"
	result += "\nAll communication is A2A (Agent-to-Agent) - NO subagents!"
	result += fmt.Sprintf("\n\nTo check status, use: check_a2a_project_status with project_id=\"%s\" or project_id=\"%s\" or project_id=\"latest\"", project.ID, shortcut)

	return &tools.ToolResult{
		ForLLM:  result,
		ForUser: fmt.Sprintf("🚀 A2A Project '%s' started!\n📋 ID: %s\n🔖 Shortcut: %s", name, project.ID, shortcut),
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CheckA2AProjectStatusTool checks A2A project status
type CheckA2AProjectStatusTool struct {
	orchestrator *A2AOrchestrator
}

// NewCheckA2AProjectStatusTool creates a new check A2A project status tool
func NewCheckA2AProjectStatusTool(orchestrator *A2AOrchestrator) *CheckA2AProjectStatusTool {
	return &CheckA2AProjectStatusTool{orchestrator: orchestrator}
}

// Name returns the tool name
func (t *CheckA2AProjectStatusTool) Name() string {
	return "check_a2a_project_status"
}

// Description returns the tool description
func (t *CheckA2AProjectStatusTool) Description() string {
	return "Check the status of an A2A project. Shows current phase, agent assignments, A2A messages, and overall progress. You can use project ID, shortcut (e.g., 'go-pwd'), or 'latest' for the most recent project."
}

// Parameters returns the tool parameters
func (t *CheckA2AProjectStatusTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_id": map[string]any{
				"type":        "string",
				"description": "Project ID",
			},
		},
		"required": []string{"project_id"},
	}
}

// Execute executes the tool
func (t *CheckA2AProjectStatusTool) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
	projectID, ok := args["project_id"].(string)
	if !ok {
		return tools.ErrorResult("project_id is required")
	}

	// Try to get project by shortcut or ID
	project, ok := t.orchestrator.GetProjectByShortcut(projectID)
	if !ok {
		return tools.ErrorResult(fmt.Sprintf("project %s not found (try 'latest' for most recent project)", projectID))
	}

	project.mu.RLock()
	defer project.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("A2A Project: %s\n", project.Name))
	sb.WriteString(fmt.Sprintf("Project ID: %s\n", project.ID))
	sb.WriteString(fmt.Sprintf("Status: %s\n\n", project.Status))

	// Calculate progress
	completed, failed, running, progressPercentage := t.orchestrator.GetProjectProgress(project.ID)
	totalAssignments := len(project.Assignments)

	sb.WriteString(fmt.Sprintf("Overall Progress: %.1f%%\n", progressPercentage))
	sb.WriteString(fmt.Sprintf("Tasks: %d completed, %d failed, %d running, %d pending\n\n",
		completed, failed, running, totalAssignments-completed-failed-running))

	sb.WriteString("Phases:\n")
	phases := []Phase{PhaseDiscovery, PhaseMeeting, PhasePlanning, PhaseExecution, PhaseIntegration, PhaseValidation}
	for _, phase := range phases {
		info := project.Phases[phase]
		status := "⏳"
		if info.Status == PhaseStatusCompleted {
			status = "✅"
		} else if info.Status == PhaseStatusRunning {
			status = "🔄"
		} else if info.Status == PhaseStatusFailed {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", status, phase))
	}

	sb.WriteString(fmt.Sprintf("\nA2A Messages: %d\n", len(project.Messages)))

	// Show recent running/failed assignments with progress
	if running > 0 || failed > 0 {
		sb.WriteString("\nActive Tasks:\n")
		for i := range project.Assignments {
			a := &project.Assignments[i]
			if a.Status == AssignmentStatusRunning {
				sb.WriteString(fmt.Sprintf("  🔄 [%s] %s: %d%% - %s\n", a.ToAgent, a.Task[:min(30, len(a.Task))], a.Progress, a.ProgressMsg))
			} else if a.Status == AssignmentStatusFailed {
				sb.WriteString(fmt.Sprintf("  ❌ [%s] %s: FAILED - %s\n", a.ToAgent, a.Task[:min(30, len(a.Task))], a.Result))
			}
		}
	}

	return &tools.ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("📊 Project '%s': %.1f%% complete (%d/%d tasks)", project.Name, progressPercentage, completed, totalAssignments),
	}
}

// ListA2AAgentsTool lists all A2A capable agents
type ListA2AAgentsTool struct {
	discovery *A2AAgentDiscovery
}

// NewListA2AAgentsTool creates a new list A2A agents tool
func NewListA2AAgentsTool(discovery *A2AAgentDiscovery) *ListA2AAgentsTool {
	return &ListA2AAgentsTool{discovery: discovery}
}

// Name returns the tool name
func (t *ListA2AAgentsTool) Name() string {
	return "list_a2a_agents"
}

// Description returns the tool description
func (t *ListA2AAgentsTool) Description() string {
	return "List all agents available for A2A collaboration with their capabilities, roles, and departments."
}

// Parameters returns the tool parameters
func (t *ListA2AAgentsTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"role": map[string]any{
				"type":        "string",
				"description": "Filter by role",
			},
		},
	}
}

// Execute executes the tool
func (t *ListA2AAgentsTool) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
	role, _ := args["role"].(string)

	var agents []*A2AAgentCapability
	if role != "" {
		agents = t.discovery.FindAgentsByRole(role)
	} else {
		agents = t.discovery.DiscoverAll()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d A2A agents:\n\n", len(agents)))

	for _, agent := range agents {
		sb.WriteString(t.discovery.FormatAgentForDisplay(agent.AgentID))
		sb.WriteString("\n")
	}

	return &tools.ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("🔍 Found %d A2A agents", len(agents)),
	}
}

// SendA2AMessageTool sends a message between agents and waits for response
type SendA2AMessageTool struct {
	orchestrator *A2AOrchestrator
}

// NewSendA2AMessageTool creates a new send A2A message tool
func NewSendA2AMessageTool(orchestrator *A2AOrchestrator) *SendA2AMessageTool {
	return &SendA2AMessageTool{orchestrator: orchestrator}
}

// Name returns the tool name
func (t *SendA2AMessageTool) Name() string {
	return "send_a2a_message"
}

// Description returns the tool description
func (t *SendA2AMessageTool) Description() string {
	return "Send a message from one agent to another via A2A and wait for their response. Use this for direct agent-to-agent conversation."
}

// Parameters returns the tool parameters
func (t *SendA2AMessageTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"from": map[string]any{
				"type":        "string",
				"description": "Sender agent ID (usually 'jarvis')",
			},
			"to": map[string]any{
				"type":        "string",
				"description": "Recipient agent ID (e.g., 'nova', 'clawed', 'pixel')",
			},
			"message": map[string]any{
				"type":        "string",
				"description": "Message content to send",
			},
			"type": map[string]any{
				"type":        "string",
				"description": "Message type",
				"enum":        []string{"task", "response", "question", "decision", "progress"},
			},
			"wait_for_response": map[string]any{
				"type":        "boolean",
				"description": "Whether to wait for agent's response (default: true for one-to-one chat)",
			},
		},
		"required": []string{"from", "to", "message"},
	}
}

// Execute executes the tool
func (t *SendA2AMessageTool) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
	from, _ := args["from"].(string)
	to, _ := args["to"].(string)
	message, _ := args["message"].(string)
	msgType, _ := args["type"].(string)
	waitForResponse := false // Default: async mode (don't wait for real-time response)

	if v, ok := args["wait_for_response"].(bool); ok {
		waitForResponse = v
	}

	if msgType == "" {
		msgType = "question"
	}

	// Send the message
	t.orchestrator.sendA2AMessage(from, to, msgType, message)

	// If not waiting for response, return immediately
	if !waitForResponse {
		return &tools.ToolResult{
			ForLLM:  fmt.Sprintf("A2A message sent from %s to %s", from, to),
			ForUser: fmt.Sprintf("📨 Message sent: %s -> %s", from, to),
		}
	}

	// Wait for response with timeout
	resp, err := t.orchestrator.waitForDirectResponse(ctx, to, 30*time.Second)
	if err != nil {
		// Return sent status but note that response timed out
		return &tools.ToolResult{
			ForLLM: fmt.Sprintf(`Message sent from %s to %s, but no response received within timeout.
The agent may be processing or unavailable.

Original message: %s`, from, to, message),
			ForUser: fmt.Sprintf("📨 Message sent to %s (no response yet)", to),
		}
	}

	// Return the conversation
	return &tools.ToolResult{
		ForLLM: fmt.Sprintf(`Conversation with %s:

You: %s
%s: %s`, to, message, to, resp.Content),
		ForUser: fmt.Sprintf("💬 %s replied: %s", to, resp.Content),
	}
}

// GetA2AMessagesTool gets A2A messages for a project or direct messages
type GetA2AMessagesTool struct {
	orchestrator *A2AOrchestrator
}

// NewGetA2AMessagesTool creates a new get A2A messages tool
func NewGetA2AMessagesTool(orchestrator *A2AOrchestrator) *GetA2AMessagesTool {
	return &GetA2AMessagesTool{orchestrator: orchestrator}
}

// Name returns the tool name
func (t *GetA2AMessagesTool) Name() string {
	return "get_a2a_messages"
}

// Description returns the tool description
func (t *GetA2AMessagesTool) Description() string {
	return "Get all A2A messages for a project or direct conversation. Shows the conversation history between agents."
}

// Parameters returns the tool parameters
func (t *GetA2AMessagesTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_id": map[string]any{
				"type":        "string",
				"description": "Project ID (use 'direct' for direct messages not in a project)",
			},
			"agent_id": map[string]any{
				"type":        "string",
				"description": "Filter by agent ID (e.g., 'nova' to see messages with Nova)",
			},
			"conversation_with": map[string]any{
				"type":        "string",
				"description": "Show conversation with specific agent (e.g., 'nova' for Jarvis-Nova conversation)",
			},
		},
		"required": []string{"project_id"},
	}
}

// Execute executes the tool
func (t *GetA2AMessagesTool) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
	projectID, _ := args["project_id"].(string)
	agentID, _ := args["agent_id"].(string)
	conversationWith, _ := args["conversation_with"].(string)

	// Handle direct messages (non-project)
	if projectID == "direct" || projectID == "" {
		return t.getDirectMessages(conversationWith)
	}

	project, ok := t.orchestrator.GetProject(projectID)
	if !ok {
		return tools.ErrorResult(fmt.Sprintf("project %s not found", projectID))
	}

	project.mu.RLock()
	defer project.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("A2A Messages for Project: %s\n\n", project.Name))

	for _, msg := range project.Messages {
		if agentID != "" && msg.From != agentID && msg.To != agentID {
			continue
		}
		sb.WriteString(fmt.Sprintf("[%s] %s -> %s: %s\n\n",
			msg.Type, msg.From, msg.To, msg.Content))
	}

	return &tools.ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("📨 Found %d messages", len(project.Messages)),
	}
}

// getDirectMessages gets direct (non-project) messages
func (t *GetA2AMessagesTool) getDirectMessages(conversationWith string) *tools.ToolResult {
	// For now, return info about how to use direct messaging
	if conversationWith == "" {
		return &tools.ToolResult{
			ForLLM: `Direct Messages (Non-Project):
To see conversation with a specific agent, use conversation_with parameter.
Example: conversation_with="nova"`,
			ForUser: "💡 Use conversation_with parameter to see messages with a specific agent",
		}
	}

	// Check if agent exists
	if _, ok := t.orchestrator.workers[conversationWith]; !ok {
		return tools.ErrorResult(fmt.Sprintf("agent %s not found", conversationWith))
	}

	return &tools.ToolResult{
		ForLLM: fmt.Sprintf(`Direct conversation with %s:
Use send_a2a_message with wait_for_response=true to have a conversation.
Example: send_a2a_message(from="jarvis", to="%s", message="Hello!")`,
			conversationWith, conversationWith),
		ForUser: fmt.Sprintf("💬 Ready to chat with %s", conversationWith),
	}
}

// ResumeA2AProjectTool resumes a failed or paused A2A project
type ResumeA2AProjectTool struct {
	orchestrator *A2AOrchestrator
}

// NewResumeA2AProjectTool creates a new resume A2A project tool
func NewResumeA2AProjectTool(orchestrator *A2AOrchestrator) *ResumeA2AProjectTool {
	return &ResumeA2AProjectTool{orchestrator: orchestrator}
}

// Name returns the tool name
func (t *ResumeA2AProjectTool) Name() string {
	return "resume_a2a_project"
}

// Description returns the tool description
func (t *ResumeA2AProjectTool) Description() string {
	return "Resume a failed or interrupted A2A project using project_id or shortcut. All completed phases and tasks are skipped, and failed tasks are retried. Does not restart gathered data - only resumes work."
}

// Parameters returns the tool parameters
func (t *ResumeA2AProjectTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_id": map[string]any{
				"type":        "string",
				"description": "Project ID or shortcut (can be 'latest')",
			},
		},
		"required": []string{"project_id"},
	}
}

// Execute executes the tool
func (t *ResumeA2AProjectTool) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return tools.ErrorResult("project_id is required")
	}

	// Resolve latest
	if projectID == "latest" {
		projectID = t.orchestrator.GetLatestProjectID()
	}

	// Find project (by ID or shortcut)
	project, ok := t.orchestrator.GetProjectByShortcut(projectID)
	if !ok {
		return tools.ErrorResult(fmt.Sprintf("project %s not found", projectID))
	}

	if err := t.orchestrator.ResumeProject(project.ID); err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to resume project: %v", err))
	}

	return &tools.ToolResult{
		ForLLM:  fmt.Sprintf("✅ A2A Project '%s' resumed!", project.Name),
		ForUser: fmt.Sprintf("✅ Project '%s' (%s) resumed and retrying failed tasks.", project.Name, project.ID),
	}
}
