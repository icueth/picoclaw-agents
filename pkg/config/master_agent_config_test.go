package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSubagentRoleConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		roles   map[string]SubagentRoleConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid role config",
			roles: map[string]SubagentRoleConfig{
				"coder": {
					Model:             "openai/gpt-5.2",
					Description:       "Writes code",
					SystemPromptAddon: "You are a coder",
					MaxIterations:     50,
					TimeoutSeconds:    300,
				},
			},
			wantErr: false,
		},
		{
			name: "negative max_iterations",
			roles: map[string]SubagentRoleConfig{
				"coder": {
					MaxIterations: -1,
				},
			},
			wantErr: true,
			errMsg:  "max_iterations cannot be negative",
		},
		{
			name: "negative timeout",
			roles: map[string]SubagentRoleConfig{
				"coder": {
					TimeoutSeconds: -1,
				},
			},
			wantErr: true,
			errMsg:  "timeout_seconds cannot be negative",
		},
		{
			name:    "empty role name not allowed in map",
			roles:   map[string]SubagentRoleConfig{},
			wantErr: false, // empty map is fine
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{SubagentRoles: tt.roles}
			err := cfg.ValidateSubagentRoles()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSubagentRoles() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() == "" || !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateSubagentRoles() error message = %v, should contain %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestMemoryConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		memory  MemoryConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid sqlite config",
			memory: MemoryConfig{
				Type:                 "sqlite",
				Database:             "~/.picoclaw/memory.db",
				ConceptRetentionDays: 30,
				MaxContextMemories:   10,
				RAG: RAGConfig{
					Enabled:             true,
					EmbeddingModel:      "local",
					ChunkSize:           512,
					Overlap:             128,
					SimilarityThreshold: 0.7,
				},
			},
			wantErr: false,
		},
		{
			name: "valid postgres config",
			memory: MemoryConfig{
				Type:     "postgres",
				Database: "postgres://user:pass@localhost/db",
			},
			wantErr: false,
		},
		{
			name: "invalid memory type",
			memory: MemoryConfig{
				Type: "invalid",
			},
			wantErr: true,
			errMsg:  "memory.type must be one of",
		},
		{
			name: "invalid embedding model",
			memory: MemoryConfig{
				Type: "sqlite",
				RAG: RAGConfig{
					Enabled:        true,
					EmbeddingModel: "invalid",
				},
			},
			wantErr: true,
			errMsg:  "memory.rag.embedding_model must be one of",
		},
		{
			name: "negative chunk size",
			memory: MemoryConfig{
				Type: "sqlite",
				RAG: RAGConfig{
					Enabled:   true,
					ChunkSize: -1,
				},
			},
			wantErr: true,
			errMsg:  "chunk_size cannot be negative",
		},
		{
			name: "negative overlap",
			memory: MemoryConfig{
				Type: "sqlite",
				RAG: RAGConfig{
					Enabled: true,
					Overlap: -1,
				},
			},
			wantErr: true,
			errMsg:  "overlap cannot be negative",
		},
		{
			name: "similarity threshold too high",
			memory: MemoryConfig{
				Type: "sqlite",
				RAG: RAGConfig{
					Enabled:             true,
					SimilarityThreshold: 1.5,
				},
			},
			wantErr: true,
			errMsg:  "similarity_threshold must be between 0 and 1",
		},
		{
			name: "negative concept retention",
			memory: MemoryConfig{
				Type:                 "sqlite",
				ConceptRetentionDays: -1,
			},
			wantErr: true,
			errMsg:  "concept_retention_days cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Memory: tt.memory}
			err := cfg.ValidateMemory()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMemory() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateMemory() error message = %v, should contain %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestJobConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		jobs    JobConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid sqlite config",
			jobs: JobConfig{
				Persistence:          "sqlite",
				DatabasePath:         "~/.picoclaw/jobs.db",
				DefaultTimeout:       300,
				MaxConcurrent:        10,
				CleanupIntervalHours: 24,
				RetentionDays:        7,
			},
			wantErr: false,
		},
		{
			name: "invalid persistence type",
			jobs: JobConfig{
				Persistence: "invalid",
			},
			wantErr: true,
			errMsg:  "jobs.persistence must be one of",
		},
		{
			name: "negative timeout",
			jobs: JobConfig{
				Persistence:    "sqlite",
				DefaultTimeout: -1,
			},
			wantErr: true,
			errMsg:  "default_timeout cannot be negative",
		},
		{
			name: "negative max concurrent",
			jobs: JobConfig{
				Persistence:   "sqlite",
				MaxConcurrent: -1,
			},
			wantErr: true,
			errMsg:  "max_concurrent cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Jobs: tt.jobs}
			err := cfg.ValidateJobs()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJobs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateJobs() error message = %v, should contain %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestWorkspaceConfig_Validation(t *testing.T) {
	// Test that ValidateWorkspace always passes now (deprecated config)
	cfg := &Config{
		Workspace: WorkspaceConfig{
			Path:        "~/.picoclaw/workspace",
			DeniedPaths: []string{"~/.ssh"},
		},
	}
	err := cfg.ValidateWorkspace()
	if err != nil {
		t.Errorf("ValidateWorkspace() should not return error for deprecated config, got %v", err)
	}
}

func TestDefaultConfig_MasterAgentSettings(t *testing.T) {
	cfg := DefaultConfig()

	// Test Memory defaults
	if cfg.Memory.Type != "sqlite" {
		t.Errorf("DefaultConfig() Memory.Type = %v, want sqlite", cfg.Memory.Type)
	}
	if cfg.Memory.ConceptRetentionDays != 30 {
		t.Errorf("DefaultConfig() Memory.ConceptRetentionDays = %v, want 30", cfg.Memory.ConceptRetentionDays)
	}
	if !cfg.Memory.AutoSave {
		t.Error("DefaultConfig() Memory.AutoSave should be true")
	}
	if !cfg.Memory.RAG.Enabled {
		t.Error("DefaultConfig() Memory.RAG.Enabled should be true")
	}
	if cfg.Memory.RAG.ChunkSize != 512 {
		t.Errorf("DefaultConfig() Memory.RAG.ChunkSize = %v, want 512", cfg.Memory.RAG.ChunkSize)
	}

	// Test Jobs defaults
	if cfg.Jobs.Persistence != "sqlite" {
		t.Errorf("DefaultConfig() Jobs.Persistence = %v, want sqlite", cfg.Jobs.Persistence)
	}
	if cfg.Jobs.DefaultTimeout != 3600 {
		t.Errorf("DefaultConfig() Jobs.DefaultTimeout = %v, want 3600", cfg.Jobs.DefaultTimeout)
	}
	if cfg.Jobs.MaxConcurrent != 10 {
		t.Errorf("DefaultConfig() Jobs.MaxConcurrent = %v, want 10", cfg.Jobs.MaxConcurrent)
	}

	// Test Agent Defaults Workspace
	if cfg.Agents.Defaults.Workspace == "" {
		t.Error("DefaultConfig() Agents.Defaults.Workspace should not be empty")
	}
}

func TestConfig_JSONSerialization(t *testing.T) {
	cfg := DefaultConfig()

	// Serialize to JSON
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Deserialize
	var restored Config
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify subagent roles
	if len(restored.SubagentRoles) != len(cfg.SubagentRoles) {
		t.Errorf("SubagentRoles count mismatch: got %d, want %d", len(restored.SubagentRoles), len(cfg.SubagentRoles))
	}

	// Verify memory config
	if restored.Memory.Type != cfg.Memory.Type {
		t.Errorf("Memory.Type mismatch: got %v, want %v", restored.Memory.Type, cfg.Memory.Type)
	}

	// Verify jobs config
	if restored.Jobs.Persistence != cfg.Jobs.Persistence {
		t.Errorf("Jobs.Persistence mismatch: got %v, want %v", restored.Jobs.Persistence, cfg.Jobs.Persistence)
	}

	// Verify agent defaults workspace
	if restored.Agents.Defaults.Workspace != cfg.Agents.Defaults.Workspace {
		t.Errorf("Agents.Defaults.Workspace mismatch: got %v, want %v", restored.Agents.Defaults.Workspace, cfg.Agents.Defaults.Workspace)
	}
}

func TestConfig_GetSubagentRole(t *testing.T) {
	cfg := &Config{
		SubagentRoles: map[string]SubagentRoleConfig{
			"coder": {Model: "openai/gpt-5.2"},
		},
	}

	// Test getting existing role
	role, exists := cfg.GetSubagentRole("coder")
	if !exists {
		t.Error("GetSubagentRole('coder') should return exists=true")
	}
	if role.Model != "openai/gpt-5.2" {
		t.Errorf("GetSubagentRole('coder').Model = %v, want openai/gpt-5.2", role.Model)
	}

	// Test getting non-existent role
	_, exists = cfg.GetSubagentRole("nonexistent")
	if exists {
		t.Error("GetSubagentRole('nonexistent') should return exists=false")
	}
}

func TestConfig_GetMemoryDatabasePath(t *testing.T) {
	cfg := &Config{
		Memory: MemoryConfig{
			Database: "~/.picoclaw/memory.db",
		},
	}

	path := cfg.GetMemoryDatabasePath()
	if path == "" {
		t.Error("GetMemoryDatabasePath() should not return empty string")
	}
	if path == "~/.picoclaw/memory.db" {
		t.Error("GetMemoryDatabasePath() should expand ~ to home directory")
	}
}

func TestConfig_GetJobsDatabasePath(t *testing.T) {
	cfg := &Config{
		Jobs: JobConfig{
			DatabasePath: "~/.picoclaw/jobs.db",
		},
	}

	path := cfg.GetJobsDatabasePath()
	if path == "" {
		t.Error("GetJobsDatabasePath() should not return empty string")
	}
	if path == "~/.picoclaw/jobs.db" {
		t.Error("GetJobsDatabasePath() should expand ~ to home directory")
	}
}

func TestConfig_GetWorkspacePath(t *testing.T) {
	cfg := DefaultConfig()

	// Test Agents.Defaults.Workspace is used and path is expanded
	path := cfg.GetWorkspacePath()
	if path == "" {
		t.Error("GetWorkspacePath() should not return empty string")
	}
	// Path should be expanded (no ~ at start)
	if len(path) > 0 && path[0] == '~' {
		t.Error("GetWorkspacePath() should expand ~ to home directory")
	}
	// Path should contain "workspace" in it
	if !strings.Contains(path, "workspace") {
		t.Error("GetWorkspacePath() should contain 'workspace' in path")
	}
}

func TestConfig_Validate(t *testing.T) {
	cfg := DefaultConfig()

	// Should pass validation
	if err := cfg.Validate(); err != nil {
		t.Errorf("DefaultConfig().Validate() error = %v", err)
	}

	// Should fail with invalid memory config
	cfg.Memory.Type = "invalid"
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error for invalid memory type")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
