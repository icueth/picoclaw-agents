package agent

import (
	"strings"

	"picoclaw/agent/pkg/config"
)


// BuiltinAgent represents a built-in agent definition embedded in core.
// These agents are always available without config.json agents.list configuration.
type BuiltinAgent struct {
	ID             string   // unique slug, e.g. "frontend-developer"
	Name           string   // human-readable display name
	Department     string   // one of the standard departments
	Role           string   // descriptor used for role-based lookup
	Avatar         string   // emoji for display
	Description    string   // short description of capabilities
	Capabilities   []string // list of capability tags
	IsPermanent    bool     // cannot be removed from registry
	Prompt         string   // Embedded persona prompt markdown
}

// ToAgentConfig converts a BuiltinAgent to config.AgentConfig for registry use.
func (ba *BuiltinAgent) ToAgentConfig(model string) config.AgentConfig {
	var modelCfg *config.AgentModelConfig
	if model != "" {
		modelCfg = &config.AgentModelConfig{Primary: model}
	}
	return config.AgentConfig{
		ID:             ba.ID,
		Name:           ba.Name,
		Department:     ba.Department,
		Role:           ba.Role,
		Avatar:         ba.Avatar,
		Model:          modelCfg,
		Capabilities:   ba.Capabilities,
		Default:        ba.ID == "coordinator",
		Prompt:         ba.Prompt,
	}
}

// GetBuiltinAgents returns all built-in agents across all departments.
func GetBuiltinAgents() []BuiltinAgent {
	agents := make([]BuiltinAgent, 0, 120)
	agents = append(agents, coreAgents()...)
	agents = append(agents, engineeringAgents()...)
	agents = append(agents, designAgents()...)
	agents = append(agents, marketingAgents()...)
	agents = append(agents, testingAgents()...)
	agents = append(agents, productAgents()...)
	agents = append(agents, projectManagementAgents()...)
	agents = append(agents, supportAgents()...)
	agents = append(agents, specializedAgents()...)
	agents = append(agents, gameDevelopmentAgents()...)
	agents = append(agents, spatialComputingAgents()...)
	agents = append(agents, paidMediaAgents()...)
	return agents
}

// GetBuiltinAgentsByDepartment returns built-in agents filtered by department.
func GetBuiltinAgentsByDepartment(dept string) []BuiltinAgent {
	all := GetBuiltinAgents()
	result := make([]BuiltinAgent, 0)
	for _, a := range all {
		if a.Department == dept {
			result = append(result, a)
		}
	}
	return result
}

// GetAllDepartmentNames returns all standard department names.
func GetAllDepartmentNames() []string {
	return []string{
		"core",
		"engineering",
		"design",
		"marketing",
		"testing",
		"product",
		"project-management",
		"support",
		"specialized",
		"game-development",
		"spatial-computing",
		"paid-media",
	}
}

// BuildAgentRosterForSystemPrompt returns a compact listing of all built-in
// agents grouped by department. Intentionally minimal — just IDs — so the
// coordinator's system prompt stays small and LLM can process it efficiently.
func BuildAgentRosterForSystemPrompt() string {
	agents := GetBuiltinAgents()

	departmentOrder := GetAllDepartmentNames()
	byDept := make(map[string][]string, len(departmentOrder))
	for _, a := range agents {
		byDept[a.Department] = append(byDept[a.Department], a.ID)
	}

	var sb strings.Builder
	for _, dept := range departmentOrder {
		ids, ok := byDept[dept]
		if !ok {
			continue
		}
		sb.WriteString("**" + dept + "**: ")
		sb.WriteString(strings.Join(ids, ", "))
		sb.WriteString("\n")
	}
	return sb.String()
}

// itoa converts an int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
