package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"picoclaw/agent/pkg/logger"
)

// AgentWorkspace manages per-agent workspace directories and files.
// Each agent gets its own isolated storage within the shared workspace.
type AgentWorkspace struct {
	AgentID   string
	BasePath  string // ~/.picoclaw/workspace/agents/{agent_id}/
	MemoryDir string
	SessionDir string
}

// NewAgentWorkspace creates or initializes an agent's workspace.
// Structure:
//   ~/.picoclaw/workspace/
//     agents/
//       {agent_id}/
//         memory/
//           MEMORY.md
//           entries/
//         sessions/
//         tasks/
func NewAgentWorkspace(workspaceRoot, agentID string) *AgentWorkspace {
	if agentID == "" {
		agentID = "default"
	}

	basePath := filepath.Join(workspaceRoot, "agents", agentID)
	
	aw := &AgentWorkspace{
		AgentID:    agentID,
		BasePath:   basePath,
		MemoryDir:  filepath.Join(basePath, "memory"),
		SessionDir: filepath.Join(basePath, "sessions"),
	}

	// Auto-create directory structure
	aw.initialize()

	return aw
}

// initialize creates the directory structure if it doesn't exist.
func (aw *AgentWorkspace) initialize() {
	dirs := []string{
		aw.BasePath,
		aw.MemoryDir,
		filepath.Join(aw.MemoryDir, "entries"),
		aw.SessionDir,
		filepath.Join(aw.BasePath, "tasks"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			logger.ErrorCF("agent_workspace", "Failed to create directory",
				map[string]any{"dir": dir, "error": err.Error()})
		}
	}

	// Create default MEMORY.md if it doesn't exist
	memoryFile := filepath.Join(aw.MemoryDir, "MEMORY.md")
	if _, err := os.Stat(memoryFile); os.IsNotExist(err) {
		defaultMemory := fmt.Sprintf(`# Memory: %s

## About Me
I am %s, an AI assistant.

## Key Information
<!-- Important facts and context about the user and projects -->

## Preferences
<!-- User preferences and settings -->

## Ongoing Tasks
<!-- Current projects and their status -->
`, aw.AgentID, aw.AgentID)
		
		if err := os.WriteFile(memoryFile, []byte(defaultMemory), 0o644); err != nil {
			logger.ErrorCF("agent_workspace", "Failed to create MEMORY.md",
				map[string]any{"error": err.Error()})
		} else {
			logger.InfoCF("agent_workspace", "Created default MEMORY.md",
				map[string]any{"agent_id": aw.AgentID})
		}
	}

	logger.InfoCF("agent_workspace", "Agent workspace initialized",
		map[string]any{
			"agent_id": aw.AgentID,
			"path":     aw.BasePath,
		})
}

// GetMemoryPath returns the path to the agent's MEMORY.md
func (aw *AgentWorkspace) GetMemoryPath() string {
	return filepath.Join(aw.MemoryDir, "MEMORY.md")
}

// GetMemoryEntriesPath returns the path to the agent's memory entries directory
func (aw *AgentWorkspace) GetMemoryEntriesPath() string {
	return filepath.Join(aw.MemoryDir, "entries")
}

// GetSessionPath returns the path to the agent's sessions directory
func (aw *AgentWorkspace) GetSessionPath() string {
	return aw.SessionDir
}

// GetTasksPath returns the path to the agent's tasks directory
func (aw *AgentWorkspace) GetTasksPath() string {
	return filepath.Join(aw.BasePath, "tasks")
}

// EnsureDir ensures a subdirectory exists within the agent workspace
func (aw *AgentWorkspace) EnsureDir(subdir string) string {
	path := filepath.Join(aw.BasePath, subdir)
	os.MkdirAll(path, 0o755)
	return path
}

// ListAgentWorkspaces returns all agent IDs that have workspaces
func ListAgentWorkspaces(workspaceRoot string) ([]string, error) {
	agentsDir := filepath.Join(workspaceRoot, "agents")
	
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var agents []string
	for _, entry := range entries {
		if entry.IsDir() {
			agents = append(agents, entry.Name())
		}
	}
	
	return agents, nil
}

// DeleteAgentWorkspace removes an agent's workspace completely
func DeleteAgentWorkspace(workspaceRoot, agentID string) error {
	basePath := filepath.Join(workspaceRoot, "agents", agentID)
	logger.InfoCF("agent_workspace", "Deleting agent workspace",
		map[string]any{"agent_id": agentID, "path": basePath})
	return os.RemoveAll(basePath)
}
