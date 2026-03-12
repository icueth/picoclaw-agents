// Package persona manages agent personality, memory, and identity files
package persona

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"picoclaw/agent/pkg/config"
)

// PersonaFiles represents the three core files for each agent
type PersonaFiles struct {
	Identity string // IDENTITY.md - Who am I?
	Soul     string // SOUL.md - Personality, values, beliefs
	Memory   string // MEMORY.md - Experiences, learnings, relationships
}

// DefaultPersonaTemplates contains default templates for each agent type
type DefaultPersonaTemplates struct {
	Coordinator *PersonaFiles
	Researcher  *PersonaFiles
	Coder       *PersonaFiles
	Writer      *PersonaFiles
	QA          *PersonaFiles
	Analyst     *PersonaFiles
	Designer    *PersonaFiles
	Architect   *PersonaFiles
}

// GetDefaultPersona returns default persona for an agent role
func GetDefaultPersona(role, name, avatar string) *PersonaFiles {
	switch role {
	case "coordinator":
		return getCoordinatorPersona(name, avatar)
	case "researcher":
		return getResearcherPersona(name, avatar)
	case "coder":
		return getCoderPersona(name, avatar)
	case "writer":
		return getWriterPersona(name, avatar)
	case "qa":
		return getQAPersona(name, avatar)
	case "analyst":
		return getAnalystPersona(name, avatar)
	case "designer":
		return getDesignerPersona(name, avatar)
	case "architect":
		return getArchitectPersona(name, avatar)
	default:
		return getGenericPersona(name, avatar, role)
	}
}

// CreatePersonaFiles creates the three persona files for an agent
func CreatePersonaFiles(agentDir string, persona *PersonaFiles) error {
	// Create agent directory if not exists
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	// Write IDENTITY.md
	identityPath := filepath.Join(agentDir, "IDENTITY.md")
	if err := os.WriteFile(identityPath, []byte(persona.Identity), 0644); err != nil {
		return fmt.Errorf("failed to write IDENTITY.md: %w", err)
	}

	// Write SOUL.md
	soulPath := filepath.Join(agentDir, "SOUL.md")
	if err := os.WriteFile(soulPath, []byte(persona.Soul), 0644); err != nil {
		return fmt.Errorf("failed to write SOUL.md: %w", err)
	}

	// Write MEMORY.md
	memoryPath := filepath.Join(agentDir, "MEMORY.md")
	if err := os.WriteFile(memoryPath, []byte(persona.Memory), 0644); err != nil {
		return fmt.Errorf("failed to write MEMORY.md: %w", err)
	}

	return nil
}

// LoadPersonaFiles reads the persona files for an agent
func LoadPersonaFiles(agentDir string) (*PersonaFiles, error) {
	persona := &PersonaFiles{}

	// Read IDENTITY.md
	identityPath := filepath.Join(agentDir, "IDENTITY.md")
	if data, err := os.ReadFile(identityPath); err == nil {
		persona.Identity = string(data)
	}

	// Read SOUL.md
	soulPath := filepath.Join(agentDir, "SOUL.md")
	if data, err := os.ReadFile(soulPath); err == nil {
		persona.Soul = string(data)
	}

	// Read MEMORY.md
	memoryPath := filepath.Join(agentDir, "MEMORY.md")
	if data, err := os.ReadFile(memoryPath); err == nil {
		persona.Memory = string(data)
	}

	return persona, nil
}

// EnsureAgentPersona checks and creates persona files if missing
func EnsureAgentPersona(agentCfg *config.AgentConfig, baseDir string) error {
	agentDir := filepath.Join(baseDir, agentCfg.ID)
	
	// Check if files exist
	identityPath := filepath.Join(agentDir, "IDENTITY.md")
	if _, err := os.Stat(identityPath); err == nil {
		// Files exist
		return nil
	}

	// Create default persona
	persona := GetDefaultPersona(agentCfg.Role, agentCfg.Name, agentCfg.Avatar)
	return CreatePersonaFiles(agentDir, persona)
}

// InitializeAllAgentsPersona creates persona files for the given agents
func InitializeAllAgentsPersona(agents []*config.AgentConfig, baseDir string) error {
	agentsDir := filepath.Join(baseDir, "agents")
	
	for _, agent := range agents {
		if err := EnsureAgentPersona(agent, agentsDir); err != nil {
			return fmt.Errorf("failed to initialize persona for %s: %w", agent.ID, err)
		}
	}
	
	return nil
}

// UpdateMemory adds a new memory entry to MEMORY.md
func UpdateMemory(agentDir, entry string) error {
	memoryPath := filepath.Join(agentDir, "MEMORY.md")
	
	timestamp := time.Now().Format("2006-01-02 15:04")
	memoryEntry := fmt.Sprintf("\n## %s\n%s\n", timestamp, entry)
	
	// Append to file
	f, err := os.OpenFile(memoryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	
	_, err = f.WriteString(memoryEntry)
	return err
}

// GetPersonaSummary returns a summary of the agent's persona for prompts
func GetPersonaSummary(agentDir string) (string, error) {
	persona, err := LoadPersonaFiles(agentDir)
	if err != nil {
		return "", err
	}

	var summary strings.Builder
	
	if persona.Identity != "" {
		// Extract key identity points (first few lines)
		lines := strings.Split(persona.Identity, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				summary.WriteString(line + "\n")
				if summary.Len() > 500 {
					break
				}
			}
		}
	}
	
	if persona.Soul != "" {
		// Extract personality traits
		lines := strings.Split(persona.Soul, "\n")
		inTraits := false
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "personality") || 
			   strings.Contains(strings.ToLower(line), "traits") {
				inTraits = true
			}
			if inTraits && strings.HasPrefix(line, "-") {
				summary.WriteString(line + "\n")
			}
		}
	}
	
	return summary.String(), nil
}
