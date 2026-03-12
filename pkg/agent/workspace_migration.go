package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"picoclaw/agent/pkg/fileutil"
	"picoclaw/agent/pkg/logger"
)

// MigrationResult tracks the result of migrating data for one agent
type MigrationResult struct {
	AgentID       string   `json:"agent_id"`
	SessionsMoved int      `json:"sessions_moved"`
	MemoryCopied  bool     `json:"memory_copied"`
	Errors        []string `json:"errors,omitempty"`
}

// WorkspaceMigration handles migrating from shared workspace to per-agent workspaces
type WorkspaceMigration struct {
	workspaceRoot string
	results       []MigrationResult
}

// NewWorkspaceMigration creates a new migration helper
func NewWorkspaceMigration(workspaceRoot string) *WorkspaceMigration {
	return &WorkspaceMigration{
		workspaceRoot: workspaceRoot,
		results:       make([]MigrationResult, 0),
	}
}

// DetectLegacyData checks if there's legacy shared data that needs migration
func (wm *WorkspaceMigration) DetectLegacyData() map[string]any {
	result := map[string]any{
		"has_legacy_sessions": false,
		"has_legacy_memory":   false,
		"legacy_session_count": 0,
		"legacy_agents_found": []string{},
	}

	// Check for shared sessions
	sessionsDir := filepath.Join(wm.workspaceRoot, "sessions")
	if entries, err := os.ReadDir(sessionsDir); err == nil {
		count := 0
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				count++
			}
		}
		if count > 0 {
			result["has_legacy_sessions"] = true
			result["legacy_session_count"] = count
		}
	}

	// Check for legacy memory
	memoryFile := filepath.Join(wm.workspaceRoot, "memory", "MEMORY.md")
	if _, err := os.Stat(memoryFile); err == nil {
		result["has_legacy_memory"] = true
	}

	// Check for agents that have been used (from session keys)
	agentIDs := wm.detectAgentsFromSessions()
	result["legacy_agents_found"] = agentIDs

	return result
}

// detectAgentsFromSessions extracts agent IDs from session filenames
// Session format: {channel}_{chatID}.json or agent-specific formats
func (wm *WorkspaceMigration) detectAgentsFromSessions() []string {
	agentSet := make(map[string]bool)
	
	sessionsDir := filepath.Join(wm.workspaceRoot, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return []string{}
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Try to read session and detect agent
		sessionPath := filepath.Join(sessionsDir, entry.Name())
		if data, err := os.ReadFile(sessionPath); err == nil {
			var session struct {
				Messages []struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						Function struct {
							Name string `json:"name"`
						} `json:"function"`
					} `json:"tool_calls,omitempty"`
				} `json:"messages"`
			}
			if err := json.Unmarshal(data, &session); err == nil {
				// Look for agent mentions in messages
				for _, msg := range session.Messages {
					content := strings.ToLower(msg.Content)
					// Simple heuristic: look for @agent-name patterns
					if strings.Contains(content, "@") {
						parts := strings.Split(content, "@")
						for i := 1; i < len(parts); i++ {
							agentName := strings.Fields(parts[i])[0]
							agentName = strings.Trim(agentName, "!?.:,;\"")
							// Validate: agent name should be alphanumeric with hyphens/underscores only
							// and reasonable length (not URLs or emails)
							if isValidAgentName(agentName) {
								agentSet[agentName] = true
							}
						}
					}
				}
			}
		}
	}

	// Convert set to slice
	agents := make([]string, 0, len(agentSet))
	for agent := range agentSet {
		agents = append(agents, agent)
	}
	return agents
}

// isValidAgentName checks if a string looks like a valid agent identifier
func isValidAgentName(name string) bool {
	if name == "" {
		return false
	}
	// Check length (shouldn't be too long or too short)
	if len(name) < 2 || len(name) > 50 {
		return false
	}
	// Should not contain URL-specific characters
	if strings.ContainsAny(name, "/.?&=+%") {
		return false
	}
	// Should not look like an email or URL
	if strings.Contains(name, ".com") || strings.Contains(name, ".org") ||
	   strings.Contains(name, ".net") || strings.Contains(name, ".io") ||
	   strings.Contains(name, "http") || strings.Contains(name, "www") {
		return false
	}
	// Valid characters: alphanumeric, hyphen, underscore
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
		     (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// MigrateAgent migrates all data for a specific agent to its per-agent workspace
func (wm *WorkspaceMigration) MigrateAgent(agentID string) (*MigrationResult, error) {
	result := &MigrationResult{
		AgentID:       agentID,
		SessionsMoved: 0,
		MemoryCopied:  false,
		Errors:        make([]string, 0),
	}

	// Create agent workspace
	aw := NewAgentWorkspace(wm.workspaceRoot, agentID)

	// Migrate sessions that belong to this agent
	sessionsMoved, errs := wm.migrateSessionsForAgent(agentID, aw)
	result.SessionsMoved = sessionsMoved
	result.Errors = append(result.Errors, errs...)

	// Copy shared memory to agent's memory (as starting point)
	if err := wm.migrateMemoryForAgent(aw); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("memory migration: %v", err))
	} else {
		result.MemoryCopied = true
	}

	wm.results = append(wm.results, *result)
	return result, nil
}

// migrateSessionsForAgent moves sessions from shared to per-agent directory
func (wm *WorkspaceMigration) migrateSessionsForAgent(agentID string, aw *AgentWorkspace) (int, []string) {
	moved := 0
	errs := make([]string, 0)

	sourceDir := filepath.Join(wm.workspaceRoot, "sessions")
	targetDir := aw.SessionDir

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return 0, []string{fmt.Sprintf("cannot read sessions dir: %v", err)}
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		sourcePath := filepath.Join(sourceDir, entry.Name())
		targetPath := filepath.Join(targetDir, entry.Name())

		// Check if this session belongs to the agent (simple heuristic)
		belongsToAgent := wm.sessionBelongsToAgent(sourcePath, agentID)
		
		if belongsToAgent {
			// Copy session file
			if data, err := os.ReadFile(sourcePath); err == nil {
				if err := fileutil.WriteFileAtomic(targetPath, data, 0o600); err != nil {
					errs = append(errs, fmt.Sprintf("failed to copy %s: %v", entry.Name(), err))
				} else {
					moved++
					logger.InfoCF("migration", "Migrated session",
						map[string]any{
							"agent_id": agentID,
							"session":  entry.Name(),
						})
				}
			}
		}
	}

	return moved, errs
}

// sessionBelongsToAgent checks if a session file belongs to a specific agent
func (wm *WorkspaceMigration) sessionBelongsToAgent(sessionPath, agentID string) bool {
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return false
	}

	var session struct {
		Key      string `json:"key"`
		Messages []struct {
			Content string `json:"content"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(data, &session); err != nil {
		return false
	}

	// Check if agent is mentioned in any message
	// Look for @agent-id pattern with word boundaries
	for _, msg := range session.Messages {
		content := strings.ToLower(msg.Content)
		// Match @agentID where it's a whole word (not part of URL/email)
		patterns := []string{
			"@" + strings.ToLower(agentID) + " ",
			"@" + strings.ToLower(agentID) + "\n",
			"@" + strings.ToLower(agentID) + "\t",
			"@" + strings.ToLower(agentID) + ",",
			"@" + strings.ToLower(agentID) + ".",
			"@" + strings.ToLower(agentID) + "!",
			"@" + strings.ToLower(agentID) + "?",
			"@" + strings.ToLower(agentID) + ":",
		}
		for _, pattern := range patterns {
			if strings.Contains(content, pattern) {
				return true
			}
		}
		// Also check end of string
		if strings.HasSuffix(content, "@"+strings.ToLower(agentID)) {
			return true
		}
	}

	return false
}

// migrateMemoryForAgent copies shared MEMORY.md to agent's memory as starting point
func (wm *WorkspaceMigration) migrateMemoryForAgent(aw *AgentWorkspace) error {
	sourceMemory := filepath.Join(wm.workspaceRoot, "memory", "MEMORY.md")
	targetMemory := aw.GetMemoryPath()

	// If agent already has memory, don't overwrite
	if _, err := os.Stat(targetMemory); err == nil {
		logger.DebugCF("migration", "Agent already has MEMORY.md, skipping",
			map[string]any{"agent_id": aw.AgentID})
		return nil
	}

	// Read shared memory
	data, err := os.ReadFile(sourceMemory)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No shared memory to migrate
		}
		return err
	}

	// Add migration note
	migrationNote := fmt.Sprintf(`

---

## Migration Note
*Migrated from shared workspace on %s*
This memory was copied from the shared workspace as a starting point.
Agent-specific memories will be added here.
`, time.Now().Format("2006-01-02"))

	content := string(data) + migrationNote

	// Write to agent's memory
	if err := os.WriteFile(targetMemory, []byte(content), 0o644); err != nil {
		return err
	}

	logger.InfoCF("migration", "Migrated shared memory to agent",
		map[string]any{"agent_id": aw.AgentID})

	return nil
}

// MigrateAll runs migration for all detected agents
func (wm *WorkspaceMigration) MigrateAll() ([]MigrationResult, error) {
	agents := wm.detectAgentsFromSessions()
	
	logger.InfoCF("migration", "Starting migration",
		map[string]any{"agents_detected": len(agents)})

	for _, agentID := range agents {
		if _, err := wm.MigrateAgent(agentID); err != nil {
			logger.ErrorCF("migration", "Failed to migrate agent",
				map[string]any{"agent_id": agentID, "error": err.Error()})
		}
	}

	return wm.results, nil
}

// GenerateReport creates a migration report
func (wm *WorkspaceMigration) GenerateReport() string {
	var sb strings.Builder
	
	sb.WriteString("# Workspace Migration Report\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	
	totalSessions := 0
	totalMemory := 0
	
	for _, r := range wm.results {
		sb.WriteString(fmt.Sprintf("## Agent: %s\n", r.AgentID))
		sb.WriteString(fmt.Sprintf("- Sessions moved: %d\n", r.SessionsMoved))
		sb.WriteString(fmt.Sprintf("- Memory copied: %v\n", r.MemoryCopied))
		if len(r.Errors) > 0 {
			sb.WriteString("- Errors:\n")
			for _, err := range r.Errors {
				sb.WriteString(fmt.Sprintf("  - %s\n", err))
			}
		}
		sb.WriteString("\n")
		
		totalSessions += r.SessionsMoved
		if r.MemoryCopied {
			totalMemory++
		}
	}
	
	sb.WriteString("## Summary\n")
	sb.WriteString(fmt.Sprintf("- Total agents migrated: %d\n", len(wm.results)))
	sb.WriteString(fmt.Sprintf("- Total sessions moved: %d\n", totalSessions))
	sb.WriteString(fmt.Sprintf("- Total memories copied: %d\n", totalMemory))
	
	return sb.String()
}

// SaveReport saves the migration report to a file
func (wm *WorkspaceMigration) SaveReport() error {
	reportPath := filepath.Join(wm.workspaceRoot, "MIGRATION_REPORT.md")
	report := wm.GenerateReport()
	return os.WriteFile(reportPath, []byte(report), 0o644)
}
