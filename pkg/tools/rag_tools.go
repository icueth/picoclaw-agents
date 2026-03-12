// Package tools provides tool handlers for RAG (Retrieval-Augmented Generation) operations.
package tools

import (
	"context"
	"fmt"
	"strings"

	"picoclaw/agent/pkg/rag"
)

// =============================================================================
// QueryRAGTool - Search documents in RAG
// =============================================================================

// QueryRAGTool searches for relevant documents in the RAG system.
type QueryRAGTool struct {
	ragManager *rag.Manager
}

// NewQueryRAGTool creates a new QueryRAGTool instance.
func NewQueryRAGTool(rm *rag.Manager) *QueryRAGTool {
	return &QueryRAGTool{
		ragManager: rm,
	}
}

// Name returns the tool name.
func (t *QueryRAGTool) Name() string {
	return "query_rag"
}

// Description returns the tool description.
func (t *QueryRAGTool) Description() string {
	return "Search for relevant documents in the RAG (Retrieval-Augmented Generation) system. Returns documents ranked by relevance to the query."
}

// Parameters returns the tool parameters schema.
func (t *QueryRAGTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The search query to find relevant documents",
			},
			"max_results": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results to return (default: 5)",
			},
			"agent": map[string]any{
				"type":        "string",
				"description": "Optional filter: only return documents from this agent",
			},
			"role": map[string]any{
				"type":        "string",
				"description": "Optional filter: only return documents from this role",
			},
			"project": map[string]any{
				"type":        "string",
				"description": "Optional filter: only return documents from this project",
			},
		},
		"required": []string{"query"},
	}
}

// Execute performs the RAG search.
func (t *QueryRAGTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.ragManager == nil {
		return ErrorResult("RAG manager not initialized").WithError(fmt.Errorf("rag manager is nil"))
	}

	query, ok := args["query"].(string)
	if !ok || query == "" {
		return ErrorResult("query is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid query parameter"))
	}

	maxResults := 5
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// Check for metadata filters
	agent := ""
	role := ""
	project := ""
	if a, ok := args["agent"].(string); ok {
		agent = a
	}
	if r, ok := args["role"].(string); ok {
		role = r
	}
	if p, ok := args["project"].(string); ok {
		project = p
	}

	var results []rag.SearchResult
	var err error

	// If filters are specified, use metadata search
	if agent != "" || role != "" || project != "" {
		docs, err := t.ragManager.SearchByMetadata(agent, role, project, maxResults)
		if err != nil {
			return ErrorResult(fmt.Sprintf("Failed to search by metadata: %v", err)).WithError(err)
		}
		// Convert to SearchResult format
		for _, doc := range docs {
			results = append(results, rag.SearchResult{
				ID:      doc.ID,
				Content: doc.Content,
				Score:   1.0, // Metadata search doesn't have scores
				Metadata: rag.DocumentMetadata{
					Agent:   doc.Metadata.Agent,
					Role:    doc.Metadata.Role,
					Project: doc.Metadata.Project,
				},
			})
		}
	} else {
		// Use semantic search
		results, err = t.ragManager.Search(query, maxResults)
		if err != nil {
			return ErrorResult(fmt.Sprintf("Failed to search RAG: %v", err)).WithError(err)
		}
	}

	if len(results) == 0 {
	return &ToolResult{
		ForLLM: "No relevant documents found in the knowledge base.",
	}
	}

	// Format results for LLM
	var resultText strings.Builder
	resultText.WriteString(fmt.Sprintf("Found %d relevant documents:\n\n", len(results)))
	for i, r := range results {
		resultText.WriteString(fmt.Sprintf("[%d] Score: %.3f\n", i+1, r.Score))
		if r.Metadata.Agent != "" {
			resultText.WriteString(fmt.Sprintf("    Agent: %s\n", r.Metadata.Agent))
		}
		if r.Metadata.Role != "" {
			resultText.WriteString(fmt.Sprintf("    Role: %s\n", r.Metadata.Role))
		}
		resultText.WriteString(fmt.Sprintf("    Content: %s\n\n", truncateString(r.Content, 500)))
	}

	// Build data response
	resultData := make([]map[string]any, len(results))
	for i, r := range results {
		resultData[i] = map[string]any{
			"id":      r.ID,
			"content": r.Content,
			"score":   r.Score,
			"metadata": map[string]any{
				"agent":   r.Metadata.Agent,
				"role":    r.Metadata.Role,
				"project": r.Metadata.Project,
			},
		}
	}

	return &ToolResult{
		ForLLM:  resultText.String(),
		ForUser: fmt.Sprintf("Found %d relevant documents", len(results)),
	}
}

// =============================================================================
// SaveToRAGTool - Add document to RAG
// =============================================================================

// SaveToRAGTool adds a document to the RAG system.
type SaveToRAGTool struct {
	ragManager *rag.Manager
}

// NewSaveToRAGTool creates a new SaveToRAGTool instance.
func NewSaveToRAGTool(rm *rag.Manager) *SaveToRAGTool {
	return &SaveToRAGTool{
		ragManager: rm,
	}
}

// Name returns the tool name.
func (t *SaveToRAGTool) Name() string {
	return "save_to_rag"
}

// Description returns the tool description.
func (t *SaveToRAGTool) Description() string {
	return "Save a document to the RAG (Retrieval-Augmented Generation) system for future retrieval. The document will be chunked and indexed for semantic search."
}

// Parameters returns the tool parameters schema.
func (t *SaveToRAGTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"content": map[string]any{
				"type":        "string",
				"description": "The document content to save",
			},
			"agent": map[string]any{
				"type":        "string",
				"description": "Optional: agent ID that created this document",
			},
			"role": map[string]any{
				"type":        "string",
				"description": "Optional: role that created this document",
			},
			"project": map[string]any{
				"type":        "string",
				"description": "Optional: project ID this document belongs to",
			},
			"source": map[string]any{
				"type":        "string",
				"description": "Optional: source identifier (e.g., filename, URL)",
			},
		},
		"required": []string{"content"},
	}
}

// Execute saves the document to RAG.
func (t *SaveToRAGTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.ragManager == nil {
		return ErrorResult("RAG manager not initialized").WithError(fmt.Errorf("rag manager is nil"))
	}

	content, ok := args["content"].(string)
	if !ok || content == "" {
		return ErrorResult("content is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid content parameter"))
	}

	// Build metadata
	metadata := rag.DocumentMetadata{}
	if a, ok := args["agent"].(string); ok {
		metadata.Agent = a
	}
	if r, ok := args["role"].(string); ok {
		metadata.Role = r
	}
	if p, ok := args["project"].(string); ok {
		metadata.Project = p
	}
	if s, ok := args["source"].(string); ok {
		metadata.Source = s
	}

	ids, err := t.ragManager.AddDocument(content, metadata)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to save document to RAG: %v", err)).WithError(err)
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("Document saved successfully to RAG. Created %d chunk(s) with IDs: %v", len(ids), ids),
		ForUser: fmt.Sprintf("Saved document with %d chunks", len(ids)),
	}
}

// =============================================================================
// GetRAGContextTool - Get context for a query
// =============================================================================

// GetRAGContextTool retrieves relevant context for a query from RAG.
type GetRAGContextTool struct {
	ragManager *rag.Manager
}

// NewGetRAGContextTool creates a new GetRAGContextTool instance.
func NewGetRAGContextTool(rm *rag.Manager) *GetRAGContextTool {
	return &GetRAGContextTool{
		ragManager: rm,
	}
}

// Name returns the tool name.
func (t *GetRAGContextTool) Name() string {
	return "get_rag_context"
}

// Description returns the tool description.
func (t *GetRAGContextTool) Description() string {
	return "Retrieve relevant context from the RAG system for a given query. Returns concatenated context from the most relevant documents."
}

// Parameters returns the tool parameters schema.
func (t *GetRAGContextTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The query to retrieve context for",
			},
		},
		"required": []string{"query"},
	}
}

// Execute retrieves context from RAG.
func (t *GetRAGContextTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.ragManager == nil {
		return ErrorResult("RAG manager not initialized").WithError(fmt.Errorf("rag manager is nil"))
	}

	query, ok := args["query"].(string)
	if !ok || query == "" {
		return ErrorResult("query is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid query parameter"))
	}

	contextResult, err := t.ragManager.GetContextForQuery(query)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get RAG context: %v", err)).WithError(err)
	}

	if len(contextResult.Documents) == 0 {
		return &ToolResult{
			ForLLM: "No relevant context found in the knowledge base.",
		}
	}

	// Concatenate document contents for context
	var contextText strings.Builder
	for i, doc := range contextResult.Documents {
		if i > 0 {
			contextText.WriteString("\n\n---\n\n")
		}
		contextText.WriteString(doc.Content)
	}

	// Build document list for data
	docs := make([]map[string]any, len(contextResult.Documents))
	for i, d := range contextResult.Documents {
		docs[i] = map[string]any{
			"id":      d.ID,
			"content": d.Content,
			"score":   d.Score,
		}
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("Relevant context from %d documents:\n\n%s", len(contextResult.Documents), contextText.String()),
		ForUser: fmt.Sprintf("Retrieved context from %d documents", len(contextResult.Documents)),
	}
}

// =============================================================================
// RAGToolsRegistry - Helper to register all RAG tools
// =============================================================================

// RAGToolsRegistry holds all RAG tools for easy registration.
type RAGToolsRegistry struct {
	QueryRAG     *QueryRAGTool
	SaveToRAG    *SaveToRAGTool
	GetRAGContext *GetRAGContextTool
}

// NewRAGToolsRegistry creates a new RAGToolsRegistry with all tools initialized.
func NewRAGToolsRegistry(rm *rag.Manager) *RAGToolsRegistry {
	if rm == nil {
		return &RAGToolsRegistry{}
	}
	return &RAGToolsRegistry{
		QueryRAG:      NewQueryRAGTool(rm),
		SaveToRAG:     NewSaveToRAGTool(rm),
		GetRAGContext: NewGetRAGContextTool(rm),
	}
}

// RegisterAll registers all RAG tools to the provided registry.
func (r *RAGToolsRegistry) RegisterAll(registry ToolRegistry) {
	if r.QueryRAG != nil {
		registry.Register(r.QueryRAG)
	}
	if r.SaveToRAG != nil {
		registry.Register(r.SaveToRAG)
	}
	if r.GetRAGContext != nil {
		registry.Register(r.GetRAGContext)
	}
}

// GetTools returns all tools as a slice.
func (r *RAGToolsRegistry) GetTools() []Tool {
	var tools []Tool
	if r.QueryRAG != nil {
		tools = append(tools, r.QueryRAG)
	}
	if r.SaveToRAG != nil {
		tools = append(tools, r.SaveToRAG)
	}
	if r.GetRAGContext != nil {
		tools = append(tools, r.GetRAGContext)
	}
	return tools
}

