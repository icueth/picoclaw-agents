package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"picoclaw/agent/pkg/agentcomm"
	"picoclaw/agent/pkg/logger"
)

// Capability represents what an agent can do
type Capability struct {
	Name        string   // e.g., "code", "research", "creative"
	Description string   // human-readable description
	Keywords    []string // keywords that indicate this capability
	Models      []string // preferred models for this capability
}

// AgentCapability represents an agent's registered capabilities
type AgentCapability struct {
	AgentID      string
	Role         string
	Department   string
	Capabilities []Capability
	Metadata     map[string]any // additional metadata
}

// DiscoveryConfig configures the discovery system
type DiscoveryConfig struct {
	EnableAutoDiscover bool // auto-discover capabilities from agent behavior
	CacheTTLSeconds    int  // cache TTL in seconds
	MaxResults         int  // max results to return
}

// AgentDiscovery handles agent capability discovery and matching
type AgentDiscovery struct {
	mu           sync.RWMutex
	capabilities map[string]AgentCapability // agent ID -> capabilities
	registry     *AgentRegistry
	config       DiscoveryConfig
}

// DiscoveryInterface defines the interface for agent discovery (used by tools)
type DiscoveryInterface interface {
	FindAgents(ctx context.Context, query string) []agentcomm.AgentInfo
	FindByCapability(capabilityName string) []string
	GetCapabilities(agentID string) ([]Capability, bool)
	SuggestCapabilities(task string) []string
	GetAllCapabilities() map[string]AgentCapability
}

// NewAgentDiscovery creates a new AgentDiscovery instance
func NewAgentDiscovery(registry *AgentRegistry) *AgentDiscovery {
	return &AgentDiscovery{
		capabilities: make(map[string]AgentCapability),
		registry:     registry,
		config: DiscoveryConfig{
			EnableAutoDiscover: true,
			CacheTTLSeconds:    300,
			MaxResults:         10,
		},
	}
}

// SetConfig sets the discovery configuration
func (ad *AgentDiscovery) SetConfig(config DiscoveryConfig) {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	ad.config = config
}

// RegisterCapabilities registers capabilities for an agent
func (ad *AgentDiscovery) RegisterCapabilities(agentID, role, department string, caps []Capability, metadata map[string]any) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	ad.capabilities[agentID] = AgentCapability{
		AgentID:      agentID,
		Role:         role,
		Department:   department,
		Capabilities: caps,
		Metadata:     metadata,
	}

	logger.InfoCF("agent", "Registered agent capabilities",
		map[string]any{
			"agent_id":     agentID,
			"capabilities": len(caps),
		})
}

// GetCapabilities returns the capabilities for an agent
func (ad *AgentDiscovery) GetCapabilities(agentID string) ([]Capability, bool) {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	cap, ok := ad.capabilities[agentID]
	return cap.Capabilities, ok
}

// FindAgents finds agents matching the given query
func (ad *AgentDiscovery) FindAgents(ctx context.Context, query string) []agentcomm.AgentInfo {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	queryLower := strings.ToLower(query)
	var results []agentcomm.AgentInfo
	maxResults := ad.config.MaxResults

	// Score each agent
	type scoredAgent struct {
		agent agentcomm.AgentInfo
		score int
	}
	var scored []scoredAgent

	for agentID, caps := range ad.capabilities {
		score := 0

		// Check capability matches
		for _, cap := range caps.Capabilities {
			// Direct name match
			if strings.Contains(strings.ToLower(cap.Name), queryLower) {
				score += 10
			}
			// Description match
			if strings.Contains(strings.ToLower(cap.Description), queryLower) {
				score += 5
			}
			// Keyword match
			for _, kw := range cap.Keywords {
				if strings.Contains(strings.ToLower(kw), queryLower) {
					score += 3
				}
			}
		}

		// Check if query matches agent ID or name
		if strings.Contains(strings.ToLower(agentID), queryLower) {
			score += 8
		}

		// Get agent info from registry
		if info, ok := ad.registry.GetAgentInfo(agentID); ok {
			if strings.Contains(strings.ToLower(info.Name), queryLower) {
				score += 5
			}

			if score > 0 {
				scored = append(scored, scoredAgent{agent: info, score: score})
			}
		}
	}

	// Sort by score descending
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Return top results
	for i, s := range scored {
		if i >= maxResults {
			break
		}
		results = append(results, s.agent)
	}

	return results
}

// FindByCapability finds agents with a specific capability
func (ad *AgentDiscovery) FindByCapability(capabilityName string) []string {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	capLower := strings.ToLower(capabilityName)
	var matching []string

	for agentID, caps := range ad.capabilities {
		for _, cap := range caps.Capabilities {
			if strings.Contains(strings.ToLower(cap.Name), capLower) {
				matching = append(matching, agentID)
				break
			}
		}
	}

	return matching
}

// SuggestCapabilities suggests capabilities based on a task
func (ad *AgentDiscovery) SuggestCapabilities(task string) []string {
	taskLower := strings.ToLower(task)
	var suggestions []string

	// Define capability suggestions based on keywords
	capabilityKeywords := map[string][]string{
		"code":     {"code", "implement", "function", "debug", "refactor"},
		"research": {"research", "find", "search", "analyze", "investigate"},
		"creative": {"write", "story", "poem", "creative", "design"},
		"data":     {"analyze", "process", "transform", "visualize"},
		"planning": {"plan", "strategy", "roadmap", "organize"},
		"review":   {"review", "check", "validate", "audit"},
	}

	for capName, keywords := range capabilityKeywords {
		for _, kw := range keywords {
			if strings.Contains(taskLower, kw) {
				suggestions = append(suggestions, capName)
				break
			}
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, s := range suggestions {
		if !seen[s] {
			seen[s] = true
			unique = append(unique, s)
		}
	}

	return unique
}

// RegisterDefaultCapabilities registers default capabilities for known agents
func (ad *AgentDiscovery) RegisterDefaultCapabilities() {
	defaultCaps := map[string][]Capability{
		"main": {
			{
				Name:        "general",
				Description: "General purpose tasks and conversations",
				Keywords:    []string{"help", "question", "task"},
				Models:      []string{},
			},
		},
	}

	ad.mu.Lock()
	defer ad.mu.Unlock()

	for agentID, caps := range defaultCaps {
		ad.capabilities[agentID] = AgentCapability{
			AgentID:      agentID,
			Role:         "coordinator",
			Department:   "core",
			Capabilities: caps,
			Metadata:     nil,
		}
	}
}

// GetAllCapabilities returns all registered capabilities
func (ad *AgentDiscovery) GetAllCapabilities() map[string]AgentCapability {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	result := make(map[string]AgentCapability)
	for k, v := range ad.capabilities {
		result[k] = v
	}
	return result
}

// MatchScore represents how well an agent matches a query
type MatchScore struct {
	AgentID      string
	Score        int
	MatchedOn    []string
	Capabilities []Capability
}

// FindBestMatch finds the best matching agent for a task
func (ad *AgentDiscovery) FindBestMatch(task string) (string, []Capability, error) {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	taskLower := strings.ToLower(task)
	suggestions := ad.suggestCapabilitiesLocked(task)
	targetRole := ad.inferRoleLocked(taskLower)

	// Find agents with suggested capabilities and role match
	bestScore := -1
	var bestAgentID string
	var bestCaps []Capability

	for agentID, caps := range ad.capabilities {
		score := 0

		// Capability matching (highest weight)
		for _, capName := range suggestions {
			for _, cap := range caps.Capabilities {
				if strings.EqualFold(cap.Name, capName) {
					score += 15 // Increased weight for capability match
				}
			}
		}

		// Role matching (high weight)
		if targetRole != "" && strings.EqualFold(caps.Role, targetRole) {
			score += 10
		} else if targetRole != "" && strings.Contains(strings.ToLower(caps.Role), targetRole) {
			score += 5
		}

		// Department context (bonus weight)
		if strings.Contains(taskLower, "engineering") && caps.Department == "engineering" {
			score += 5
		}
		if strings.Contains(taskLower, "design") && caps.Department == "design" {
			score += 5
		}
		if strings.Contains(taskLower, "marketing") && caps.Department == "marketing" {
			score += 5
		}

		if score > bestScore {
			bestScore = score
			bestAgentID = agentID
			bestCaps = caps.Capabilities
		}
	}

	if bestAgentID == "" || bestScore <= 0 {
		// Fallback to main agent if nothing matches well
		if _, ok := ad.capabilities["main"]; ok {
			return "main", ad.capabilities["main"].Capabilities, nil
		}
		return "", nil, fmt.Errorf("no matching agent found for task")
	}

	return bestAgentID, bestCaps, nil
}

func (ad *AgentDiscovery) inferRoleLocked(task string) string {
	if strings.Contains(task, "architect") || strings.Contains(task, "design") && strings.Contains(task, "system") {
		return "architect"
	}
	if strings.Contains(task, "code") || strings.Contains(task, "implement") || strings.Contains(task, "developer") || strings.Contains(task, "backend") || strings.Contains(task, "frontend") {
		return "developer"
	}
	if strings.Contains(task, "test") || strings.Contains(task, "qa") || strings.Contains(task, "quality") || strings.Contains(task, "validate") {
		return "qa"
	}
	if strings.Contains(task, "data") || strings.Contains(task, "analyze") || strings.Contains(task, "analytics") {
		return "analyst"
	}
	if strings.Contains(task, "write") || strings.Contains(task, "doc") || strings.Contains(task, "content") {
		return "writer"
	}
	if strings.Contains(task, "research") || strings.Contains(task, "find") || strings.Contains(task, "investigate") {
		return "researcher"
	}
	if strings.Contains(task, "pm") || strings.Contains(task, "manage") || strings.Contains(task, "coordinate") {
		return "manager"
	}
	return ""
}

func (ad *AgentDiscovery) suggestCapabilitiesLocked(task string) []string {
	taskLower := strings.ToLower(task)
	var suggestions []string

	capabilityKeywords := map[string][]string{
		"code":     {"code", "implement", "function", "debug", "refactor", "program", "golang", "api", "backend", "frontend", "ui", "react", "css"},
		"research": {"research", "find", "search", "analyze", "investigate", "look up", "info", "discovery"},
		"creative": {"write", "story", "poem", "creative", "design", "compose", "ui", "ux", "avatar", "character"},
		"data":     {"analyze", "process", "transform", "visualize", "chart", "database", "sql", "migration", "schema"},
		"planning": {"plan", "strategy", "roadmap", "organize", "schedule", "design", "architect", "proposal"},
		"review":   {"review", "check", "validate", "audit", "test", "qa", "verification", "compliance"},
	}

	for capName, keywords := range capabilityKeywords {
		for _, kw := range keywords {
			if strings.Contains(taskLower, kw) {
				suggestions = append(suggestions, capName)
				break
			}
		}
	}

	return suggestions
}

// ToolGetAllCapabilities returns capabilities in tool-compatible format
func (ad *AgentDiscovery) ToolGetAllCapabilities() map[string]any {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	result := make(map[string]any)
	for agentID, ac := range ad.capabilities {
		caps := make([]map[string]any, len(ac.Capabilities))
		for i, c := range ac.Capabilities {
			caps[i] = map[string]any{
				"Name":        c.Name,
				"Description": c.Description,
				"Keywords":    c.Keywords,
				"Models":      c.Models,
			}
		}
		result[agentID] = map[string]any{
			"Capabilities": caps,
		}
	}
	return result
}

// ToolFindAgents returns agents in tool-compatible format
func (ad *AgentDiscovery) ToolFindAgents(ctx context.Context, query string) []map[string]string {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	queryLower := strings.ToLower(query)
	var results []map[string]string
	maxResults := ad.config.MaxResults

	type scoredAgent struct {
		agentID string
		score   int
	}
	var scored []scoredAgent

	for agentID, caps := range ad.capabilities {
		score := 0

		// Check capability matches
		for _, cap := range caps.Capabilities {
			if strings.Contains(strings.ToLower(cap.Name), queryLower) {
				score += 10
			}
			if strings.Contains(strings.ToLower(cap.Description), queryLower) {
				score += 5
			}
			for _, kw := range cap.Keywords {
				if strings.Contains(strings.ToLower(kw), queryLower) {
					score += 3
				}
			}
		}

		if strings.Contains(strings.ToLower(agentID), queryLower) {
			score += 8
		}

		if score > 0 {
			scored = append(scored, scoredAgent{agentID: agentID, score: score})
		}
	}

	// Sort
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	for i, s := range scored {
		if i >= maxResults {
			break
		}
		results = append(results, map[string]string{
			"ID":     s.agentID,
			"Name":   s.agentID,
			"Type":   "agent",
			"Status": "idle",
		})
	}

	return results
}

// ToolGetCapabilities returns capabilities for an agent in tool-compatible format
func (ad *AgentDiscovery) ToolGetCapabilities(agentID string) ([]map[string]any, bool) {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	cap, ok := ad.capabilities[agentID]
	if !ok {
		return nil, false
	}

	result := make([]map[string]any, len(cap.Capabilities))
	for i, c := range cap.Capabilities {
		result[i] = map[string]any{
			"Name":        c.Name,
			"Description": c.Description,
			"Keywords":    c.Keywords,
			"Models":      c.Models,
		}
	}
	return result, true
}
