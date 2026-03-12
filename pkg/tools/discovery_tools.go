package tools

import (
	"context"
	"fmt"
	"strings"
)

// DiscoveryProvider defines the interface for agent discovery
type DiscoveryProvider interface {
	FindAgents(ctx context.Context, query string) []AgentInfoResult
	FindByCapability(capabilityName string) []string
	GetCapabilities(agentID string) ([]CapabilityResult, bool)
	SuggestCapabilities(task string) []string
	GetAllCapabilities() map[string]AgentCapabilityResult
}

// AgentInfoResult represents agent info from discovery
type AgentInfoResult struct {
	ID     string
	Name   string
	Type   string
	Status string
}

// AgentCapabilityResult represents capability info
type AgentCapabilityResult struct {
	AgentID      string
	Capabilities []CapabilityResult
	Metadata     map[string]any
}

// CapabilityResult represents what an agent can do
type CapabilityResult struct {
	Name        string
	Description string
	Keywords    []string
	Models      []string
}

// DiscoveryAdapter wraps agent.AgentDiscovery to implement DiscoveryProvider
type DiscoveryAdapter struct {
	capabilities    map[string][]CapabilityResult
	findAgentsFunc  func(ctx context.Context, query string) []map[string]string
	findByCapFunc   func(capabilityName string) []string
	suggestFunc    func(task string) []string
}

// NewDiscoveryAdapter creates a DiscoveryAdapter
func NewDiscoveryAdapter(
	capabilities map[string][]CapabilityResult,
	findAgents func(ctx context.Context, query string) []map[string]string,
	findByCap func(capabilityName string) []string,
	suggest func(task string) []string,
) *DiscoveryAdapter {
	return &DiscoveryAdapter{
		capabilities:   capabilities,
		findAgentsFunc: findAgents,
		findByCapFunc:  findByCap,
		suggestFunc:    suggest,
	}
}

func (da *DiscoveryAdapter) FindAgents(ctx context.Context, query string) []AgentInfoResult {
	if da.findAgentsFunc != nil {
		result := da.findAgentsFunc(ctx, query)
		converted := make([]AgentInfoResult, len(result))
		for i, r := range result {
			converted[i] = AgentInfoResult{
				ID:     r["ID"],
				Name:   r["Name"],
				Type:   r["Type"],
				Status: r["Status"],
			}
		}
		return converted
	}
	return nil
}

func (da *DiscoveryAdapter) FindByCapability(capabilityName string) []string {
	if da.findByCapFunc != nil {
		return da.findByCapFunc(capabilityName)
	}
	return nil
}

func (da *DiscoveryAdapter) GetCapabilities(agentID string) ([]CapabilityResult, bool) {
	if da.capabilities != nil {
		caps, ok := da.capabilities[agentID]
		return caps, ok
	}
	return nil, false
}

func (da *DiscoveryAdapter) SuggestCapabilities(task string) []string {
	if da.suggestFunc != nil {
		return da.suggestFunc(task)
	}
	return nil
}

func (da *DiscoveryAdapter) GetAllCapabilities() map[string]AgentCapabilityResult {
	result := make(map[string]AgentCapabilityResult)
	if da.capabilities != nil {
		for agentID, caps := range da.capabilities {
			result[agentID] = AgentCapabilityResult{
				AgentID:      agentID,
				Capabilities: caps,
			}
		}
	}
	return result
}

// FindAgentTool allows agents to discover other agents by capability
type FindAgentTool struct {
	discovery DiscoveryProvider
}

// NewFindAgentTool creates a new FindAgentTool
func NewFindAgentTool(discovery DiscoveryProvider) *FindAgentTool {
	return &FindAgentTool{discovery: discovery}
}

func (t *FindAgentTool) Name() string {
	return "find_agent"
}

func (t *FindAgentTool) Description() string {
	return "Find agents that match a specific capability or task. Use this to discover the best agent for a job."
}

func (t *FindAgentTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Search query (capability name, task description, or keyword)",
			},
			"capability": map[string]any{
				"type":        "string",
				"description": "Optional specific capability to search for (e.g., 'code', 'research')",
			},
		},
		"required": []string{"query"},
	}
}

func (t *FindAgentTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	query, ok := args["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return ErrorResult("query is required")
	}

	if t.discovery == nil {
		return ErrorResult("discovery system not configured")
	}

	// Check for specific capability
	capability, hasCap := args["capability"].(string)
	if hasCap && strings.TrimSpace(capability) != "" {
		// Find by specific capability
		agents := t.discovery.FindByCapability(capability)
		if len(agents) == 0 {
			return &ToolResult{
				ForLLM:  fmt.Sprintf("No agents found with capability '%s'", capability),
				ForUser: fmt.Sprintf("No agents found with capability '%s'", capability),
				Silent:  false,
				IsError:  false,
			}
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Agents with capability '%s':\n\n", capability))
		for _, agentID := range agents {
			sb.WriteString(fmt.Sprintf("- %s\n", agentID))
		}

		return &ToolResult{
			ForLLM:  sb.String(),
			ForUser: fmt.Sprintf("Found %d agent(s) with capability '%s'", len(agents), capability),
			Silent:  false,
			IsError: false,
		}
	}

	// Search by query
	agents := t.discovery.FindAgents(ctx, query)
	if len(agents) == 0 {
		// Suggest capabilities
		suggestions := t.discovery.SuggestCapabilities(query)
		var sb strings.Builder
		sb.WriteString("No agents found for query.\n\n")

		if len(suggestions) > 0 {
			sb.WriteString("You might want to try these capabilities:\n")
			for _, s := range suggestions {
				sb.WriteString(fmt.Sprintf("- %s\n", s))
			}
		} else {
			sb.WriteString("Try using the 'capability' parameter to search specifically.")
		}

		return &ToolResult{
			ForLLM:  sb.String(),
			ForUser: "No agents found for query",
			Silent:  false,
			IsError: false,
		}
	}

	// Format results
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d agent(s) matching '%s':\n\n", len(agents), query))
	sb.WriteString("| Agent ID | Name | Type | Status |\n")
	sb.WriteString("|----------|------|------|--------|\n")

	for _, a := range agents {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			a.ID, a.Name, a.Type, a.Status))
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Found %d agent(s)", len(agents)),
		Silent:  false,
		IsError: false,
	}
}

// CapabilitiesTool shows available capabilities
type CapabilitiesTool struct {
	discovery DiscoveryProvider
}

// NewCapabilitiesTool creates a new CapabilitiesTool
func NewCapabilitiesTool(discovery DiscoveryProvider) *CapabilitiesTool {
	return &CapabilitiesTool{discovery: discovery}
}

func (t *CapabilitiesTool) Name() string {
	return "list_capabilities"
}

func (t *CapabilitiesTool) Description() string {
	return "List all available agent capabilities in the system."
}

func (t *CapabilitiesTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"agent_id": map[string]any{
				"type":        "string",
				"description": "Optional agent ID to filter capabilities",
			},
		},
	}
}

func (t *CapabilitiesTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.discovery == nil {
		return ErrorResult("discovery system not configured")
	}

	agentID, hasAgentID := args["agent_id"].(string)

	var sb strings.Builder

	if hasAgentID && strings.TrimSpace(agentID) != "" {
		// Show capabilities for specific agent
		caps, ok := t.discovery.GetCapabilities(agentID)
		if !ok {
			return ErrorResult(fmt.Sprintf("Agent '%s' not found", agentID))
		}

		sb.WriteString(fmt.Sprintf("Capabilities for agent '%s':\n\n", agentID))
		if len(caps) == 0 {
			sb.WriteString("No capabilities registered.")
		} else {
			for _, cap := range caps {
				sb.WriteString(fmt.Sprintf("### %s\n", cap.Name))
				sb.WriteString(fmt.Sprintf("%s\n\n", cap.Description))
				if len(cap.Keywords) > 0 {
					sb.WriteString(fmt.Sprintf("Keywords: %s\n", strings.Join(cap.Keywords, ", ")))
				}
			}
		}
	} else {
		// Show all capabilities
		allCaps := t.discovery.GetAllCapabilities()

		sb.WriteString("## All Agent Capabilities\n\n")
		if len(allCaps) == 0 {
			sb.WriteString("No capabilities registered.")
		} else {
			for agentID, caps := range allCaps {
				sb.WriteString(fmt.Sprintf("### %s\n", agentID))
				if len(caps.Capabilities) == 0 {
					sb.WriteString("No capabilities registered.\n")
				} else {
					for _, cap := range caps.Capabilities {
						sb.WriteString(fmt.Sprintf("- **%s**: %s\n", cap.Name, cap.Description))
					}
				}
				sb.WriteString("\n")
			}
		}
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: "Capabilities listed",
		Silent:  false,
		IsError: false,
	}
}
