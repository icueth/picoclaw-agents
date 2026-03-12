package tools

import (
	"context"
	"fmt"
	"strings"

	"picoclaw/agent/pkg/config"
)

// ListSubagentRolesTool lists available subagent roles
type ListSubagentRolesTool struct {
	cfg *config.Config
}

// NewListSubagentRolesTool creates a new ListSubagentRolesTool
func NewListSubagentRolesTool(cfg *config.Config) *ListSubagentRolesTool {
	return &ListSubagentRolesTool{cfg: cfg}
}

// Name returns the tool name
func (t *ListSubagentRolesTool) Name() string {
	return "list_subagent_roles"
}

// Description returns the tool description
func (t *ListSubagentRolesTool) Description() string {
	return `List all available subagent roles that can be used with spawn_subagent.
Each role has specific configuration (model, temperature, max iterations) optimized for different task types.
Use this to select the appropriate role for a task.`
}

// Parameters returns the tool parameters
func (t *ListSubagentRolesTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

// Execute runs the tool
func (t *ListSubagentRolesTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.cfg == nil {
		return ErrorResult("Configuration not available")
	}

	if len(t.cfg.SubagentRoles) == 0 {
		return &ToolResult{
			ForLLM:  "No subagent roles configured. Add roles to config.json under 'subagent_roles'.",
			ForUser: "No subagent roles configured",
			Silent:  false,
			IsError: false,
		}
	}

	var sb strings.Builder
	sb.WriteString("## Available Subagent Roles\n\n")
	sb.WriteString("| Role | Description | Model | Max Iterations | Temperature |\n")
	sb.WriteString("|------|-------------|-------|----------------|-------------|\n")

	for roleName, role := range t.cfg.SubagentRoles {
		description := role.Description
		if description == "" {
			description = "(no description)"
		}

		model := role.Model
		if model == "" {
			model = "(default)"
		}

		maxIter := "(default)"
		if role.MaxIterations > 0 {
			maxIter = fmt.Sprintf("%d", role.MaxIterations)
		}

		temp := "(default)"
		if role.Temperature != nil {
			temp = fmt.Sprintf("%.2f", *role.Temperature)
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			roleName, description, model, maxIter, temp))
	}

	sb.WriteString("\n### Usage\n")
	sb.WriteString("Use `spawn_subagent` with the role name to spawn a subagent with that configuration.\n")
	sb.WriteString("All subagents have equal capabilities - the role determines configuration, not permissions.\n")

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Found %d subagent roles", len(t.cfg.SubagentRoles)),
		Silent:  false,
		IsError: false,
	}
}

// GetSubagentRoleTool gets details about a specific role
type GetSubagentRoleTool struct {
	cfg *config.Config
}

// NewGetSubagentRoleTool creates a new GetSubagentRoleTool
func NewGetSubagentRoleTool(cfg *config.Config) *GetSubagentRoleTool {
	return &GetSubagentRoleTool{cfg: cfg}
}

// Name returns the tool name
func (t *GetSubagentRoleTool) Name() string {
	return "get_subagent_role"
}

// Description returns the tool description
func (t *GetSubagentRoleTool) Description() string {
	return "Get detailed information about a specific subagent role"
}

// Parameters returns the tool parameters
func (t *GetSubagentRoleTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"role": map[string]any{
				"type":        "string",
				"description": "The role name to get details for",
			},
		},
		"required": []string{"role"},
	}
}

// Execute runs the tool
func (t *GetSubagentRoleTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	roleName, ok := args["role"].(string)
	if !ok || roleName == "" {
		return ErrorResult("role is required")
	}

	if t.cfg == nil {
		return ErrorResult("Configuration not available")
	}

	role, exists := t.cfg.SubagentRoles[roleName]
	if !exists {
		return ErrorResult(fmt.Sprintf("Role '%s' not found. Use list_subagent_roles to see available roles.", roleName))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Role: %s\n\n", roleName))

	if role.Description != "" {
		sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", role.Description))
	}

	if role.Model != "" {
		sb.WriteString(fmt.Sprintf("**Model**: %s\n\n", role.Model))
	}

	if role.MaxIterations > 0 {
		sb.WriteString(fmt.Sprintf("**Max Iterations**: %d\n\n", role.MaxIterations))
	}

	if role.Temperature != nil {
		sb.WriteString(fmt.Sprintf("**Temperature**: %.2f\n\n", *role.Temperature))
	}

	if role.MaxTokens > 0 {
		sb.WriteString(fmt.Sprintf("**Max Tokens**: %d\n\n", role.MaxTokens))
	}

	if role.TimeoutSeconds > 0 {
		sb.WriteString(fmt.Sprintf("**Timeout**: %d seconds\n\n", role.TimeoutSeconds))
	}

	if role.SystemPromptAddon != "" {
		sb.WriteString(fmt.Sprintf("**System Prompt Addon**: %s\n\n", role.SystemPromptAddon))
	}

	if len(role.AllowedTools) > 0 {
		sb.WriteString(fmt.Sprintf("**Allowed Tools**: %s\n\n", strings.Join(role.AllowedTools, ", ")))
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Role '%s' details retrieved", roleName),
		Silent:  false,
		IsError: false,
	}
}
