package agent

import (
	"sync"

	"picoclaw/agent/pkg/agentcomm"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/providers"
	"picoclaw/agent/pkg/routing"
)

// AgentRegistry manages multiple agent instances and routes messages to them.
// Supports fix agent list from config with department and role indexing for Office UI.
type AgentRegistry struct {
	agents        map[string]*AgentInstance
	resolver      *routing.RouteResolver
	mu            sync.RWMutex
	fixAgentList  []string                    // Fixed list of agent IDs from config
	byDepartment  map[string][]string         // department -> agent IDs
	byRole        map[string][]string         // role -> agent IDs
	modelResolver *ModelResolver
}

// NewAgentRegistry creates a registry from config, instantiating all agents.
// Builds department and role indexes for Office UI agent discovery.
func NewAgentRegistry(
	cfg *config.Config,
	provider providers.LLMProvider,
) *AgentRegistry {
	registry := &AgentRegistry{
		agents:        make(map[string]*AgentInstance),
		resolver:      routing.NewRouteResolver(cfg),
		fixAgentList:  make([]string, 0),
		byDepartment:  make(map[string][]string),
		byRole:        make(map[string][]string),
		modelResolver: NewModelResolver(cfg, provider),
	}

	// 1. Load built-in core agents
	builtinAgents := GetBuiltinAgents()
	for _, ba := range builtinAgents {
		// Resolve model and provider for this agent
		model := cfg.GetDepartmentModel(ba.Department)
		ac := ba.ToAgentConfig(model)
		id := routing.NormalizeAgentID(ac.ID)

		// Resolve model and provider for this agent
		agentModel := registry.modelResolver.ResolveModel(&ac)
		agentProvider, _ := registry.modelResolver.GetProviderForAgent(id)

		instance := NewAgentInstance(&ac, &cfg.Agents.Defaults, cfg, agentProvider)
		instance.Model = agentModel // Set the resolved model
		registry.agents[id] = instance
		registry.fixAgentList = append(registry.fixAgentList, id)

		// Build department index
		if ac.Department != "" {
			registry.byDepartment[ac.Department] = append(registry.byDepartment[ac.Department], id)
		}

		// Build role index
		if ac.Role != "" {
			registry.byRole[ac.Role] = append(registry.byRole[ac.Role], id)
		}
	}

	// 2. Load custom agents from config (agents.list) - backward compatibility
	agentConfigs := cfg.Agents.List
	for i := range agentConfigs {
		ac := &agentConfigs[i]
		id := routing.NormalizeAgentID(ac.ID)

		// Resolve model and provider for this agent
		agentModel := registry.modelResolver.ResolveModel(ac)
		agentProvider, _ := registry.modelResolver.GetProviderForAgent(id)

		instance := NewAgentInstance(ac, &cfg.Agents.Defaults, cfg, agentProvider)
		instance.Model = agentModel // Set the resolved model

		if _, exists := registry.agents[id]; !exists {
			registry.fixAgentList = append(registry.fixAgentList, id)

			// Build department index
			if ac.Department != "" {
				registry.byDepartment[ac.Department] = append(registry.byDepartment[ac.Department], id)
			}

			// Build role index
			if ac.Role != "" {
				registry.byRole[ac.Role] = append(registry.byRole[ac.Role], id)
			}
		}

		// Overwrite or add
		registry.agents[id] = instance

		logger.InfoCF("agent", "Registered custom agent from config",
			map[string]any{
				"agent_id":   id,
				"name":       ac.Name,
				"department": ac.Department,
				"role":       ac.Role,
				"workspace":  instance.Workspace,
				"model":      instance.Model,
			})
	}

	logger.InfoCF("agent", "Agent registry initialized", map[string]any{
		"total_agents": len(registry.agents),
		"builtin":      len(builtinAgents),
		"custom":       len(agentConfigs),
	})

	return registry
}

// GetAgent returns the agent instance for a given ID.
func (r *AgentRegistry) GetAgent(agentID string) (*AgentInstance, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id := routing.NormalizeAgentID(agentID)
	agent, ok := r.agents[id]
	return agent, ok
}

// ResolveRoute determines which agent handles the message.
func (r *AgentRegistry) ResolveRoute(input routing.RouteInput) routing.ResolvedRoute {
	return r.resolver.ResolveRoute(input)
}

// ListAgentIDs returns all registered agent IDs.
func (r *AgentRegistry) ListAgentIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.agents))
	for id := range r.agents {
		ids = append(ids, id)
	}
	return ids
}

// SetMemoryManager injects MemoryManager into all agents' ContextBuilders.
// This should be called after the registry is created and MemoryManager is initialized.
func (r *AgentRegistry) SetMemoryManager(mm *MemoryManager) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, agent := range r.agents {
		if agent.ContextBuilder != nil {
			agent.ContextBuilder.SetMemoryManager(mm)
		}
	}

	logger.InfoCF("agent", "MemoryManager injected into all agents",
		map[string]any{
			"agent_count": len(r.agents),
			"rag_enabled": mm != nil && mm.IsRAGEnabled(),
		})
}

// CanSpawnSubagent checks if parentAgentID is allowed to spawn targetAgentID.
func (r *AgentRegistry) CanSpawnSubagent(parentAgentID, targetAgentID string) bool {
	parent, ok := r.GetAgent(parentAgentID)
	if !ok {
		return false
	}
	if parent.Subagents == nil || parent.Subagents.AllowAgents == nil {
		return false
	}
	targetNorm := routing.NormalizeAgentID(targetAgentID)
	for _, allowed := range parent.Subagents.AllowAgents {
		if allowed == "*" {
			return true
		}
		if routing.NormalizeAgentID(allowed) == targetNorm {
			return true
		}
	}
	return false
}

// GetDefaultAgent returns the system's default coordinator agent
func (r *AgentRegistry) GetDefaultAgent() *AgentInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. New built-in coordinator
	if agent, ok := r.agents["coordinator"]; ok && agent != nil {
		return agent
	}

	// 2. Legacy jarvis or main
	if agent, ok := r.agents["jarvis"]; ok && agent != nil {
		return agent
	}
	if agent, ok := r.agents["main"]; ok && agent != nil {
		return agent
	}

	// 3. Any default agent
	for _, agent := range r.agents {
		if agent.Config.Default {
			return agent
		}
	}

	// 4. Fallback to any agent
	for _, agent := range r.agents {
		return agent
	}
	return nil
}

// GetAllAgents returns public info about all registered agents.
func (r *AgentRegistry) GetAllAgents() []agentcomm.AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]agentcomm.AgentInfo, 0, len(r.agents))
	for id, instance := range r.agents {
		agents = append(agents, agentcomm.AgentInfo{
			ID:           id,
			Name:         instance.Name,
			Type:         "agent",
			Status:       agentcomm.AgentStatusIdle,
			Capabilities: []string{},
		})
	}
	return agents
}

// GetAgentInfo returns public info about a specific agent.
func (r *AgentRegistry) GetAgentInfo(agentID string) (agentcomm.AgentInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id := routing.NormalizeAgentID(agentID)
	instance, ok := r.agents[id]
	if !ok {
		return agentcomm.AgentInfo{}, false
	}

	caps := []string{}
	if instance.Config != nil {
		caps = instance.Config.Capabilities
	}
	return agentcomm.AgentInfo{
		ID:           id,
		Name:         instance.Name,
		Type:         "agent",
		Model:        instance.Model,
		Status:       agentcomm.AgentStatusIdle,
		Capabilities: caps,
	}, true
}

// CanDelegateTo checks if this agent is allowed to delegate to target agent.
func (r *AgentRegistry) CanDelegateTo(fromAgentID, toAgentID string) bool {
	from, ok := r.GetAgent(fromAgentID)
	if !ok {
		return false
	}

	// Check if target agent exists
	_, toOk := r.GetAgent(toAgentID)
	if !toOk {
		return false
	}

	// If no restrictions configured, allow by default
	if from.Subagents == nil || from.Subagents.AllowAgents == nil {
		return true
	}

	// Check allowlist
	targetNorm := routing.NormalizeAgentID(toAgentID)
	for _, allowed := range from.Subagents.AllowAgents {
		if allowed == "*" {
			return true
		}
		if routing.NormalizeAgentID(allowed) == targetNorm {
			return true
		}
	}
	return false
}

// GetAgentsByDepartment returns all agent IDs in a specific department.
func (r *AgentRegistry) GetAgentsByDepartment(department string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if department == "" {
		return nil
	}

	// Return a copy to prevent external modification
	agents := r.byDepartment[department]
	result := make([]string, len(agents))
	copy(result, agents)
	return result
}

// GetAgentsByRole returns all agent IDs with a specific role.
func (r *AgentRegistry) GetAgentsByRole(role string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if role == "" {
		return nil
	}

	// Return a copy to prevent external modification
	agents := r.byRole[role]
	result := make([]string, len(agents))
	copy(result, agents)
	return result
}

// GetAvailableAgent returns an available agent for a given department and role.
// If department is empty, only role is used. If role is empty, only department is used.
// Returns the first matching agent or empty string if none found.
func (r *AgentRegistry) GetAvailableAgent(department, role string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// If both filters specified, find intersection
	if department != "" && role != "" {
		deptAgents := r.byDepartment[department]
		roleAgents := r.byRole[role]

		// Find first agent that matches both
		for _, deptAgent := range deptAgents {
			for _, roleAgent := range roleAgents {
				if deptAgent == roleAgent {
					return deptAgent
				}
			}
		}
		return ""
	}

	// If only department specified
	if department != "" {
		if agents := r.byDepartment[department]; len(agents) > 0 {
			return agents[0]
		}
		return ""
	}

	// If only role specified
	if role != "" {
		if agents := r.byRole[role]; len(agents) > 0 {
			return agents[0]
		}
		return ""
	}

	// If no filters specified, return first agent from fix list
	if len(r.fixAgentList) > 0 {
		return r.fixAgentList[0]
	}

	return ""
}

// GetFixAgentList returns the fixed list of agent IDs from config.
func (r *AgentRegistry) GetFixAgentList() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, len(r.fixAgentList))
	copy(result, r.fixAgentList)
	return result
}

// GetAllDepartments returns all unique department names.
func (r *AgentRegistry) GetAllDepartments() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	departments := make([]string, 0, len(r.byDepartment))
	for dept := range r.byDepartment {
		departments = append(departments, dept)
	}
	return departments
}

// GetAllRoles returns all unique role names.
func (r *AgentRegistry) GetAllRoles() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roles := make([]string, 0, len(r.byRole))
	for role := range r.byRole {
		roles = append(roles, role)
	}
	return roles
}
