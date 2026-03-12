// Package bootstrap initializes all system components for the picoclaw agent system.
package bootstrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"picoclaw/agent/pkg/agent"
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/db"
	"picoclaw/agent/pkg/embeddingservice"
	"picoclaw/agent/pkg/memory"

	"picoclaw/agent/pkg/project"
	"picoclaw/agent/pkg/rag"
)

// System holds all initialized system components.
type System struct {
	DB               *db.DB
	RAG              *rag.Manager
	ConceptMgr       *memory.ConceptManager
	JobMgr           *memory.JobManager
	ProjectMgr       *project.ProjectManager
	MemoryManager    *agent.MemoryManager       // Unified memory manager (RAG + Concepts + Jobs)
	EmbeddingManager *embeddingservice.Manager  // Embedding service manager (DEPRECATED - optional)
	A2AOrchestrator  *agent.A2AOrchestrator // A2A Project Orchestrator
	Config           *config.Config
}

// Bootstrap initializes all components and returns a System.
// This is the main entry point for starting the picoclaw system.
//
// With the new keyword-only default (embedding_provider = "none"),
// no external services are required. The system works out of the box
// using SQLite FTS5 for keyword search.
func Bootstrap(cfg *config.Config) (*System, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	sys := &System{
		Config: cfg,
	}

	// 1. Initialize database
	if err := sys.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 2. Initialize RAG manager (if enabled)
	// With embedding_provider = "none" (default), this uses keyword-only search
	if err := sys.initRAG(); err != nil {
		log.Printf("Warning: RAG initialization failed: %v", err)
		// RAG is optional, continue without it
	}

	// 3. Initialize memory managers
	if err := sys.initMemoryManagers(); err != nil {
		return nil, fmt.Errorf("failed to initialize memory managers: %w", err)
	}

	// 4. Initialize project manager
	if err := sys.initProjectManager(); err != nil {
		return nil, fmt.Errorf("failed to initialize project manager: %w", err)
	}

	// 5. Start optional embedding service (DEPRECATED)
	// Only starts if explicitly configured with embedding_provider = "http"
	if err := sys.startEmbeddingService(); err != nil {
		log.Printf("Warning: Optional embedding service startup failed: %v", err)
		// Continue without embedding service - keyword search works without it
	}

	log.Printf("System bootstrap complete")
	log.Printf("Using %s search mode (embedding_provider = %s)",
		getSearchModeName(cfg.Memory.RAG.EmbeddingModel),
		cfg.Memory.RAG.EmbeddingModel)
	
	return sys, nil
}

// getSearchModeName returns a human-readable search mode name
func getSearchModeName(model string) string {
	switch model {
	case "none", "":
		return "keyword-only (FTS5)"
	case "local":
		return "hybrid (local embeddings + FTS5)"
	case "http":
		return "hybrid (HTTP embeddings + FTS5)"
	case "openai":
		return "hybrid (OpenAI embeddings + FTS5)"
	default:
		return "hybrid"
	}
}

// InitA2AOrchestrator initializes the A2A orchestrator with agent registry.
// This should be called after agent registry is created.
func (s *System) InitA2AOrchestrator(registry *agent.AgentRegistry, provider interface{}, msgBus *bus.MessageBus) error {
	if registry == nil {
		return fmt.Errorf("agent registry is required for A2A orchestrator")
	}

	// Type assert provider to LLMProvider
	llmProvider, ok := provider.(interface{ Chat(ctx context.Context, messages []interface{}, tools []interface{}, model string, options map[string]interface{}) (interface{}, error) })
	if !ok {
		// For now, just pass nil if type doesn't match - orchestrator can work without it
		llmProvider = nil
	}
	
	_ = llmProvider // Use when needed

	// Create A2A orchestrator
	a2aOrchestrator := agent.NewA2AOrchestrator(registry, nil, s.Config, msgBus)
	a2aOrchestrator.Initialize()

	s.A2AOrchestrator = a2aOrchestrator
	log.Printf("A2A Orchestrator initialized with %d agents", len(registry.ListAgentIDs()))
	return nil
}

// initDatabase initializes the SQLite database.
func (s *System) initDatabase() error {
	dbPath := s.Config.GetMemoryDatabasePath()
	if dbPath == "" {
		// Use default path in workspace
		workspace := s.Config.GetWorkspacePath()
		dbPath = filepath.Join(workspace, "picoclaw.db")
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	database, err := db.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := database.Init(); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	s.DB = database
	log.Printf("Database initialized at %s", dbPath)
	return nil
}

// initRAG initializes the RAG manager if enabled.
// With embedding_provider = "none" (default), uses keyword-only search.
func (s *System) initRAG() error {
	if !s.Config.Memory.RAG.Enabled {
		log.Printf("RAG is disabled")
		return nil
	}

	// Create a simple HTTP client for OpenAI embeddings if needed
	var httpClient rag.HTTPClient
	if s.Config.Memory.RAG.EmbeddingModel == "openai" {
		httpClient = &simpleHTTPClient{}
	}

	ragMgr, err := rag.NewManager(s.DB, &s.Config.Memory.RAG, httpClient)
	if err != nil {
		return fmt.Errorf("failed to create RAG manager: %w", err)
	}

	s.RAG = ragMgr
	
	// Log the search mode
	model := s.Config.Memory.RAG.EmbeddingModel
	if model == "none" || model == "" {
		log.Printf("RAG manager initialized with keyword-only search (FTS5)")
	} else {
		log.Printf("RAG manager initialized with hybrid search (model: %s)", model)
	}
	
	return nil
}

// initMemoryManagers initializes concept and job managers.
func (s *System) initMemoryManagers() error {
	// Concept manager
	s.ConceptMgr = memory.NewConceptManager(s.DB, &s.Config.Memory)
	log.Printf("Concept manager initialized")

	// Job manager
	s.JobMgr = memory.NewJobManager(s.DB, &s.Config.Jobs)
	log.Printf("Job manager initialized")

	// Unified MemoryManager (integrates RAG + Concepts + Jobs)
	// Pass existing RAG manager if it was initialized successfully
	memoryManager, err := agent.NewMemoryManager(s.DB, s.Config, s.RAG)
	if err != nil {
		log.Printf("Warning: Failed to initialize MemoryManager: %v", err)
		// MemoryManager is optional, continue without it
	} else {
		s.MemoryManager = memoryManager
		if memoryManager.IsRAGEnabled() {
			log.Printf("MemoryManager initialized (RAG enabled)")
		} else {
			log.Printf("MemoryManager initialized (keyword-only mode)")
		}
	}

	// Start cleanup routines
	ctx := context.Background()
	s.ConceptMgr.StartCleanupRoutine(ctx, 0) // Use default interval
	s.JobMgr.StartCleanupRoutine(ctx, 0)    // Use default interval

	return nil
}

// initProjectManager initializes the project manager.
func (s *System) initProjectManager() error {
	s.ProjectMgr = project.NewProjectManager(s.DB, s.ConceptMgr, s.JobMgr, s.Config)
	log.Printf("Project manager initialized")
	return nil
}

// Close gracefully shuts down all system components.
func (s *System) Close() error {
	var errs []error

	if s.EmbeddingManager != nil {
		if err := s.EmbeddingManager.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop embedding service: %w", err))
		}
	}

	if s.RAG != nil {
		if err := s.RAG.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close RAG manager: %w", err))
		}
	}

	if s.DB != nil {
		if err := s.DB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	return nil
}

// startEmbeddingService starts the Python embedding service if configured.
// DEPRECATED: This is only needed for embedding_provider = "http".
// Default is now embedding_provider = "none" (keyword-only search).
func (s *System) startEmbeddingService() error {
	// Skip if not using HTTP embedding model
	if !s.Config.Memory.RAG.Enabled {
		log.Printf("Embedding service: disabled in config")
		return nil
	}
	if s.Config.Memory.RAG.EmbeddingModel != "http" {
		log.Printf("Embedding service: not needed for model '%s' (using zero-config keyword search)", 
			s.Config.Memory.RAG.EmbeddingModel)
		return nil
	}

	log.Printf("Embedding service: DEPRECATED - HTTP embedding model configured")
	log.Printf("Embedding service: Consider switching to embedding_provider = \"none\" for zero-config operation")

	// Create and start embedding service manager
	s.EmbeddingManager = embeddingservice.NewManager(s.Config)
	
	if !s.EmbeddingManager.IsConfigured() {
		log.Printf("Embedding service: not configured, using keyword-only search")
		return nil
	}

	log.Printf("Embedding service: starting...")
	if err := s.EmbeddingManager.Start(); err != nil {
		return fmt.Errorf("failed to start embedding service: %w", err)
	}
	
	log.Printf("Embedding service: started successfully")
	return nil
}

// simpleHTTPClient is a basic HTTP client implementation for OpenAI embeddings.
type simpleHTTPClient struct{}

func (c *simpleHTTPClient) Post(url string, headers map[string]string, body []byte) ([]byte, error) {
	// This is a placeholder - in production, use a proper HTTP client
	// The actual implementation would use net/http to make the request
	return nil, fmt.Errorf("HTTP client not fully implemented")
}

// DefaultBootstrap creates a System with default configuration.
// Useful for testing and quick startup.
func DefaultBootstrap() (*System, error) {
	cfg := config.DefaultConfig()

	// Set default paths
	workspace := os.ExpandEnv("$HOME/.picoclaw/workspace")
	cfg.Agents.Defaults.Workspace = workspace
	cfg.Memory.Database = filepath.Join(workspace, "picoclaw.db")

	return Bootstrap(cfg)
}

// BootstrapWithConfigPath loads config from path and bootstraps the system.
func BootstrapWithConfigPath(configPath string) (*System, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return Bootstrap(cfg)
}
