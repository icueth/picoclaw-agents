// Package config provides default configurations for agents
package config

// DefaultAgentModelConfig returns the default model configuration for agents
// ทุก Agent ใช้ kimi-coding/kimi-for-coding เป็นค่าเริ่มต้น
func DefaultAgentModelConfig() AgentModelConfig {
	return AgentModelConfig{
		Primary:   "kimi-coding/kimi-for-coding",
		Fallbacks: []string{},
	}
}

// DefaultAgentConfig returns complete default agent configuration
type DefaultAgentConfig struct {
	Model   AgentModelConfig `json:"model"`
	Persona AgentPersona     `json:"persona"`
}

// GetDefaultAgentConfig returns default config for a role
func GetDefaultAgentConfig(role string) DefaultAgentConfig {
	return DefaultAgentConfig{
		Model:   DefaultAgentModelConfig(),
		Persona: DefaultAgentPersona(role),
	}
}

// AgentTeamDefaults contains default team structure
var AgentTeamDefaults = []struct {
	ID         string
	Name       string
	Role       string
	Department string
	Default    bool
}{
	{"jarvis", "Jarvis", "coordinator", "planning", true},
	{"atlas", "Atlas", "researcher", "research", false},
	{"scribe", "Scribe", "copywriter", "marketing", false},
	{"clawed", "Clawed", "developer", "engineering", false},
	{"sentinel", "Sentinel", "qa", "qa", false},
	{"trendy", "Trendy", "trend_scout", "research", false},
	{"pixel", "Pixel", "designer", "design", false},
	{"nova", "Nova", "architect", "architecture", false},
}

// DefaultTeamCoordinator is the default coordinator agent ID
const DefaultTeamCoordinator = "jarvis"
