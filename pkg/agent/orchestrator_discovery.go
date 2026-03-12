// A2A Agent Discovery
// ค้นหาและแสดงความสามารถของ Agents สำหรับ A2A (NO SUBAGENT!)

package agent

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
 
 	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/providers"
	"picoclaw/agent/pkg/providers/protocoltypes"
)

// A2AAgentCapability represents an agent's capability for A2A
type A2AAgentCapability struct {
	AgentID       string   `json:"agent_id"`
	AgentName     string   `json:"agent_name"`
	Avatar        string   `json:"avatar"`
	Role          string   `json:"role"`
	Department    string   `json:"department"`
	Capabilities  []string `json:"capabilities"`
	Responsibilities []string `json:"responsibilities"`
	Availability  string   `json:"availability"`
	Reputation    float64  `json:"reputation"`
}

// A2AAgentDiscovery discovers agents for A2A collaboration
type A2AAgentDiscovery struct {
	mu           sync.RWMutex
	registry     *AgentRegistry
	provider     providers.LLMProvider
	capabilities map[string]*A2AAgentCapability
}

// NewA2AAgentDiscovery creates a new A2A agent discovery
func NewA2AAgentDiscovery(registry *AgentRegistry, provider providers.LLMProvider) *A2AAgentDiscovery {
	return &A2AAgentDiscovery{
		registry:     registry,
		provider:     provider,
		capabilities: make(map[string]*A2AAgentCapability),
	}
}

// DiscoverAll discovers all agents and their capabilities for A2A
func (d *A2AAgentDiscovery) DiscoverAll() []*A2AAgentCapability {
	d.mu.Lock()
	defer d.mu.Unlock()

	agentIDs := d.registry.ListAgentIDs()
	capabilities := make([]*A2AAgentCapability, 0, len(agentIDs))

	for _, agentID := range agentIDs {
		cap := d.discoverAgent(agentID)
		if cap != nil {
			d.capabilities[agentID] = cap
			capabilities = append(capabilities, cap)
		}
	}

	logger.InfoCF("a2a_discovery", "A2A Agent discovery completed",
		map[string]any{
			"agents_found": len(capabilities),
		})

	return capabilities
}

// discoverAgent discovers a single agent's capabilities
func (d *A2AAgentDiscovery) discoverAgent(agentID string) *A2AAgentCapability {
	agent, ok := d.registry.GetAgent(agentID)
	if !ok {
		return nil
	}

	caps := &A2AAgentCapability{
		AgentID:      agentID,
		AgentName:    agent.Name,
		Availability: "available",
		Reputation:   4.5,
	}

	// Get from config
	if agent.Config != nil {
		caps.Role = agent.Config.Role
		caps.Avatar = agent.Config.Avatar
		caps.Department = agent.Config.Department
		caps.Capabilities = agent.Config.Capabilities
		caps.Responsibilities = agent.Config.Responsibilities
	}

	return caps
}

// FindAgentsByCapability finds agents with a specific capability
func (d *A2AAgentDiscovery) FindAgentsByCapability(capability string) []*A2AAgentCapability {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var results []*A2AAgentCapability
	for _, caps := range d.capabilities {
		for _, cap := range caps.Capabilities {
			if strings.EqualFold(cap, capability) || strings.Contains(strings.ToLower(cap), strings.ToLower(capability)) {
				results = append(results, caps)
				break
			}
		}
	}
	return results
}

// FindAgentsByRole finds agents by role
func (d *A2AAgentDiscovery) FindAgentsByRole(role string) []*A2AAgentCapability {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var results []*A2AAgentCapability
	for _, caps := range d.capabilities {
		if strings.EqualFold(caps.Role, role) {
			results = append(results, caps)
		}
	}
	return results
}

// GetAgentCapability gets capability for a specific agent
func (d *A2AAgentDiscovery) GetAgentCapability(agentID string) (*A2AAgentCapability, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	caps, ok := d.capabilities[agentID]
	return caps, ok
}

// SemanticScore scores an agent's suitability for a task using LLM (Real Thinking!)
func (d *A2AAgentDiscovery) SemanticScore(agentID string, task string) float64 {
	d.mu.RLock()
	caps, ok := d.capabilities[agentID]
	d.mu.RUnlock()

	if !ok || d.provider == nil {
		return d.ScoreAgentForTask(agentID, task)
	}

	// Build prompt for semantic matching
	prompt := fmt.Sprintf(`Analyze if this agent is suitable for the following task.
Agent Name: %s
Role: %s
Department: %s
Capabilities: %v
Responsibilities: %v

Task: %s

Respond ONLY with a score from 0.0 to 10.0, where 10.0 is a perfect match and 0.0 is completely irrelevant.`,
		caps.AgentName, caps.Role, caps.Department, caps.Capabilities, caps.Responsibilities, task)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []protocoltypes.Message{{Role: "user", Content: prompt}}
	resp, err := d.provider.Chat(ctx, messages, nil, "gpt-4o-mini", map[string]any{
		"temperature": 0.0,
		"max_tokens":  10,
	})

	if err != nil {
		return d.ScoreAgentForTask(agentID, task)
	}

	var score float64
	fmt.Sscanf(strings.TrimSpace(resp.Content), "%f", &score)

	// Multiply by 2 to match the aging 0-20 scale in legacy ScoreAgentForTask
	return score * 2.0
}

// ScoreAgentForTask scores an agent's suitability for a task (LEGACY KEYWORD SCORING - Deprecated)
func (d *A2AAgentDiscovery) ScoreAgentForTask(agentID string, task string) float64 {
	// ... (rest of the function remains the same for fallback)
	taskLower := strings.ToLower(task)
	score := 0.0
	// ...

	caps, ok := d.GetAgentCapability(agentID)
	if !ok {
		return 0.0
	}

	// Score based on role match
	roleScores := map[string][]string{
		"architect":   {"design", "architecture", "system", "plan", "structural", "pattern"},
		"coder":       {"code", "implement", "develop", "backend", "api", "golang", "java", "python", "javascript", "typescript"},
		"researcher":  {"research", "analyze", "study", "investigate", "find", "search", "lookup"},
		"writer":      {"write", "document", "readme", "content", "copy", "blog"},
		"designer":    {"ui", "ux", "frontend", "visual", "css", "html", "react", "vue"},
		"qa":          {"test", "quality", "review", "validate", "verify", "audit", "bug", "fix"},
		"analyst":     {"analyze", "data", "schema", "sql", "database", "migration"},
		"coordinator": {"coordinate", "manage", "lead", "orchestrate", "pm"},
	}

	if keywords, ok := roleScores[strings.ToLower(caps.Role)]; ok {
		for _, keyword := range keywords {
			if strings.Contains(taskLower, keyword) {
				score += 3.0 // High weight for role-task alignment
			}
		}
	}

	// Exact role match bonus
	if strings.Contains(taskLower, strings.ToLower(caps.Role)) {
		score += 5.0
	}

	// Score based on capabilities
	for _, cap := range caps.Capabilities {
		if strings.Contains(taskLower, strings.ToLower(cap)) {
			score += 2.0
		}
	}

	// Department bonus
	if strings.Contains(taskLower, "engineering") && strings.EqualFold(caps.Department, "engineering") {
		score += 4.0
	}
	if strings.Contains(taskLower, "design") && strings.EqualFold(caps.Department, "design") {
		score += 4.0
	}

	// Cap at 20 for more granularity
	if score > 20 {
		score = 20
	}

	return score
}

// GetBestAgentForTask returns the best agent for a task
func (d *A2AAgentDiscovery) GetBestAgentForTask(task string) (string, float64) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var bestAgent string
	bestScore := 0.0

	// We use the simpler ScoreAgentForTask first to shortlist
	// Then we use SemanticScore for the top candidates if needed
	// For now, let's use the LLM to analyze the top 3 keyword matches
	
	type scoredAgent struct {
		id    string
		score float64
	}
	var candidates []scoredAgent

	for agentID := range d.capabilities {
		score := d.ScoreAgentForTask(agentID, task)
		if score > 0 {
			candidates = append(candidates, scoredAgent{agentID, score})
		}
	}

	// Sort and take top 3
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	topN := 3
	if len(candidates) < topN {
		topN = len(candidates)
	}

	for i := 0; i < topN; i++ {
		// Calculate REAL semantic score
		semanticScore := d.SemanticScore(candidates[i].id, task)
		if semanticScore > bestScore {
			bestScore = semanticScore
			bestAgent = candidates[i].id
		}
	}

	// Fallback to main agent if score is low
	if bestScore < 4.0 { // Increased threshold because Semantic Score is more accurate
		if _, ok := d.capabilities["main"]; ok {
			return "main", 1.0
		}
	}

	return bestAgent, bestScore
}

// FormatAgentForDisplay formats agent info for display
func (d *A2AAgentDiscovery) FormatAgentForDisplay(agentID string) string {
	caps, ok := d.GetAgentCapability(agentID)
	if !ok {
		return fmt.Sprintf("Agent %s not found", agentID)
	}

	avatar := caps.Avatar
	if avatar == "" {
		avatar = "🤖"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s **%s** (%s)\n", avatar, caps.AgentName, caps.Role))
	sb.WriteString(fmt.Sprintf("   Department: %s\n", caps.Department))
	sb.WriteString(fmt.Sprintf("   Capabilities: %v\n", caps.Capabilities))
	if len(caps.Responsibilities) > 0 {
		sb.WriteString(fmt.Sprintf("   Responsibilities: %v\n", caps.Responsibilities))
	}

	return sb.String()
}

// ListAllAgents returns all agents with their capabilities
func (d *A2AAgentDiscovery) ListAllAgents() []*A2AAgentCapability {
	return d.DiscoverAll()
}
