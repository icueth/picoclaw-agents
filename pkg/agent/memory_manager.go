// Package agent provides memory management that integrates SQLite, RAG, and Embedding services.
package agent

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/db"
	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/memory"
	"picoclaw/agent/pkg/rag"
)

// simpleHTTPClient implements rag.HTTPClient interface
type simpleHTTPClient struct {
	client *http.Client
}

func (c *simpleHTTPClient) Post(url string, headers map[string]string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// MemoryManager integrates ConceptManager, RAG, and Embedding services
// to provide unified memory functionality for agents.
type MemoryManager struct {
	// Core components
	db             *db.DB
	conceptManager *memory.ConceptManager
	jobManager     *memory.JobManager
	ragManager     *rag.Manager

	// Configuration
	cfg *config.Config
}

// NewMemoryManager creates a new MemoryManager with all components initialized.
// If existingRAG is provided (from bootstrap), it will be used instead of creating a new one.
func NewMemoryManager(database *db.DB, cfg *config.Config, existingRAG *rag.Manager) (*MemoryManager, error) {
	if database == nil {
		return nil, fmt.Errorf("database is required")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	mm := &MemoryManager{
		db:  database,
		cfg: cfg,
	}

	// Initialize ConceptManager
	mm.conceptManager = memory.NewConceptManager(database, &cfg.Memory)

	// Initialize JobManager
	mm.jobManager = memory.NewJobManager(database, &cfg.Jobs)

	// Initialize RAG Manager if enabled
	if cfg.Memory.RAG.Enabled {
		if existingRAG != nil {
			// Use existing RAG manager from bootstrap
			mm.ragManager = existingRAG
			logger.InfoCF("memory_manager", "Using existing RAG manager from bootstrap", nil)
		} else {
			// Create new RAG manager
			httpClient := &simpleHTTPClient{
				client: &http.Client{Timeout: 30 * time.Second},
			}
			ragMgr, err := rag.NewManager(database, &cfg.Memory.RAG, httpClient)
			if err != nil {
				logger.WarnCF("memory_manager", "Failed to initialize RAG manager, continuing without RAG",
					map[string]any{"error": err.Error()})
			} else {
				mm.ragManager = ragMgr
				logger.InfoCF("memory_manager", "RAG manager initialized successfully", nil)
			}
		}
	}

	logger.InfoCF("memory_manager", "MemoryManager initialized",
		map[string]any{
			"rag_enabled": mm.ragManager != nil,
		})

	return mm, nil
}

// SaveFact saves a fact to memory using RAG if available, otherwise falls back to concepts.
func (mm *MemoryManager) SaveFact(content string, tags []string, source string) error {
	// Try RAG first if available
	if mm.ragManager != nil {
		metadata := rag.DocumentMetadata{
			Source:    source,
			Timestamp: time.Now(),
			Extra: map[string]string{
				"tags": fmt.Sprintf("%v", tags),
			},
		}

		docIDs, err := mm.ragManager.AddDocument(content, metadata)
		if err != nil {
			logger.WarnCF("memory_manager", "Failed to save to RAG, falling back to concept",
				map[string]any{"error": err.Error()})
		} else {
			logger.DebugCF("memory_manager", "Saved fact to RAG",
				map[string]any{"doc_ids": docIDs, "content_len": len(content)})
			return nil
		}
	}

	// Fallback: Create a concept (conceptText, contextStr)
	contextStr := fmt.Sprintf(`{"source": "%s", "tags": "%v"}`, source, tags)
	conceptID, err := mm.conceptManager.CreateConcept(content, contextStr)
	if err != nil {
		return fmt.Errorf("failed to save fact: %w", err)
	}

	logger.DebugCF("memory_manager", "Saved fact as concept",
		map[string]any{"concept_id": conceptID})

	return nil
}

// QueryMemory queries memory for relevant information.
func (mm *MemoryManager) QueryMemory(query string, topK int) (*rag.ContextResult, error) {
	// Try RAG first if available
	if mm.ragManager != nil {
		results, err := mm.ragManager.Search(query, topK)
		if err != nil {
			logger.WarnCF("memory_manager", "RAG search failed, falling back to concepts",
				map[string]any{"error": err.Error()})
		} else if len(results) > 0 {
			return &rag.ContextResult{
				Documents: results,
				Query:     query,
			}, nil
		}
	}

	// Fallback: List all concepts and convert to search results
	// Note: ConceptManager doesn't have ListConcepts with status/tag filter in current API
	// So we'll return empty result for now
	logger.DebugCF("memory_manager", "No RAG available, returning empty context", nil)

	return &rag.ContextResult{
		Documents: []rag.SearchResult{},
		Query:     query,
	}, nil
}

// GetRelevantContext gets relevant context for a given topic or query.
func (mm *MemoryManager) GetRelevantContext(topic string, maxTokens int) (string, error) {
	result, err := mm.QueryMemory(topic, 5)
	if err != nil {
		return "", err
	}

	if len(result.Documents) == 0 {
		return "", nil
	}

	// Build context string from results
	var context string
	for _, doc := range result.Documents {
		context += fmt.Sprintf("\n[%s] %s\n", doc.Metadata.Source, doc.Content)
	}

	return context, nil
}

// CreateConcept creates a new concept.
func (mm *MemoryManager) CreateConcept(title, description string) (string, error) {
	// ConceptManager.CreateConcept takes (conceptText, contextStr)
	contextStr := fmt.Sprintf(`{"description": "%s"}`, description)
	return mm.conceptManager.CreateConcept(title, contextStr)
}

// CreateJob creates a new job for tracking.
// Returns jobID string
func (mm *MemoryManager) CreateJob(role, task string, data map[string]interface{}) (string, error) {
	return mm.jobManager.CreateJob(role, task, data)
}

// GetJob retrieves a job by ID.
func (mm *MemoryManager) GetJob(id string) (*memory.Job, error) {
	return mm.jobManager.GetJob(id)
}

// UpdateJobStatus updates the status of a job.
func (mm *MemoryManager) UpdateJobStatus(id, status string) error {
	return mm.jobManager.UpdateJobStatus(id, status)
}

// ListJobs lists jobs with optional status filter.
func (mm *MemoryManager) ListJobs(status string) ([]memory.Job, error) {
	return mm.jobManager.ListJobs(status)
}

// AddDocument adds a document to RAG.
func (mm *MemoryManager) AddDocument(content string, metadata rag.DocumentMetadata) ([]string, error) {
	if mm.ragManager == nil {
		return nil, fmt.Errorf("RAG is not enabled")
	}
	return mm.ragManager.AddDocument(content, metadata)
}

// SearchRAG searches RAG documents.
func (mm *MemoryManager) SearchRAG(query string, topK int) ([]rag.SearchResult, error) {
	if mm.ragManager == nil {
		return nil, fmt.Errorf("RAG is not enabled")
	}
	return mm.ragManager.Search(query, topK)
}

// IsRAGEnabled returns true if RAG is enabled and initialized.
func (mm *MemoryManager) IsRAGEnabled() bool {
	return mm.ragManager != nil
}

// GetStats returns statistics about the memory system.
func (mm *MemoryManager) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"rag_enabled": mm.ragManager != nil,
	}

	// Get RAG document count if available
	if mm.ragManager != nil {
		count, _ := mm.ragManager.Count()
		stats["rag_document_count"] = count
	}

	return stats
}
