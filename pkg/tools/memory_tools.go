// Package tools provides tool handlers for memory operations.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"picoclaw/agent/pkg/memory"
)

// =============================================================================
// CreateConceptTool
// =============================================================================

// CreateConceptTool creates a new concept for work continuity.
type CreateConceptTool struct {
	conceptManager *memory.ConceptManager
}

// NewCreateConceptTool creates a new CreateConceptTool instance.
func NewCreateConceptTool(cm *memory.ConceptManager) *CreateConceptTool {
	return &CreateConceptTool{
		conceptManager: cm,
	}
}

// Name returns the tool name.
func (t *CreateConceptTool) Name() string {
	return "create_concept"
}

// Description returns the tool description.
func (t *CreateConceptTool) Description() string {
	return "Create a new concept for tracking work continuity. Concepts help maintain context across sessions and can be resumed later. Returns the concept ID for future reference."
}

// Parameters returns the tool parameters schema.
func (t *CreateConceptTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"concept": map[string]any{
				"type":        "string",
				"description": "A short description of the concept or work item to track",
			},
			"context": map[string]any{
				"type":        "string",
				"description": "Optional context information (working directory, files, notes) as JSON or plain text",
			},
		},
		"required": []string{"concept"},
	}
}

// Execute creates a new concept.
func (t *CreateConceptTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.conceptManager == nil {
		return ErrorResult("Concept manager not initialized").WithError(fmt.Errorf("concept manager is nil"))
	}

	conceptText, ok := args["concept"].(string)
	if !ok || conceptText == "" {
		return ErrorResult("concept is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid concept parameter"))
	}

	contextStr := ""
	if ctxArg, ok := args["context"].(string); ok {
		contextStr = ctxArg
	}

	id, err := t.conceptManager.CreateConcept(conceptText, contextStr)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to create concept: %v", err)).WithError(err)
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("Concept created successfully. ID: %s\nConcept: %s", id, conceptText),
		ForUser: fmt.Sprintf("Created concept: %s (ID: %s)", conceptText, id),
		Silent:  false,
		IsError: false,
		Async:   false,
	}
}

// =============================================================================
// ListConceptsTool
// =============================================================================

// ListConceptsTool lists all concepts, optionally filtered by status.
type ListConceptsTool struct {
	conceptManager *memory.ConceptManager
}

// NewListConceptsTool creates a new ListConceptsTool instance.
func NewListConceptsTool(cm *memory.ConceptManager) *ListConceptsTool {
	return &ListConceptsTool{
		conceptManager: cm,
	}
}

// Name returns the tool name.
func (t *ListConceptsTool) Name() string {
	return "list_concepts"
}

// Description returns the tool description.
func (t *ListConceptsTool) Description() string {
	return "List all concepts/work items, optionally filtered by status. Shows concept ID, description, status, and creation date. Use this to find concepts to continue working on."
}

// Parameters returns the tool parameters schema.
func (t *ListConceptsTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status": map[string]any{
				"type":        "string",
				"description": "Optional status filter: active, paused, completed, abandoned",
				"enum":        []string{"active", "paused", "completed", "abandoned", ""},
			},
			"limit": map[string]any{
				"type":        "number",
				"description": "Maximum number of concepts to return (default: 50)",
			},
		},
		"required": []string{},
	}
}

// Execute lists concepts.
func (t *ListConceptsTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.conceptManager == nil {
		return ErrorResult("Concept manager not initialized").WithError(fmt.Errorf("concept manager is nil"))
	}

	status := ""
	if s, ok := args["status"].(string); ok {
		status = s
	}

	limit := 50
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	concepts, err := t.conceptManager.ListConcepts(status)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to list concepts: %v", err)).WithError(err)
	}

	// Apply limit
	if limit > 0 && len(concepts) > limit {
		concepts = concepts[:limit]
	}

	if len(concepts) == 0 {
		return &ToolResult{
			ForLLM:  "No concepts found.",
			ForUser: "No concepts found.",
			Silent:  false,
			IsError: false,
		}
	}

	// Build response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d concept(s):\n\n", len(concepts)))

	for i, c := range concepts {
		sb.WriteString(fmt.Sprintf("%d. ID: %s\n", i+1, c.ID))
		sb.WriteString(fmt.Sprintf("   Concept: %s\n", c.Concept))
		sb.WriteString(fmt.Sprintf("   Status: %s\n", c.Status))
		sb.WriteString(fmt.Sprintf("   Created: %s\n", c.CreatedAt.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("   Updated: %s\n", c.UpdatedAt.Format(time.RFC3339)))
		if len(c.RelatedJobs) > 0 {
			sb.WriteString(fmt.Sprintf("   Related Jobs: %d\n", len(c.RelatedJobs)))
		}
		sb.WriteString("\n")
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Found %d concept(s)", len(concepts)),
		Silent:  false,
		IsError: false,
	}
}

// =============================================================================
// ContinueConceptTool
// =============================================================================

// ContinueConceptTool continues working on an existing concept.
type ContinueConceptTool struct {
	conceptManager *memory.ConceptManager
}

// NewContinueConceptTool creates a new ContinueConceptTool instance.
func NewContinueConceptTool(cm *memory.ConceptManager) *ContinueConceptTool {
	return &ContinueConceptTool{
		conceptManager: cm,
	}
}

// Name returns the tool name.
func (t *ContinueConceptTool) Name() string {
	return "continue_concept"
}

// Description returns the tool description.
func (t *ContinueConceptTool) Description() string {
	return "Continue working on an existing concept. Marks the concept as active and returns its full context including working directory, files, and conversation history. Use this to resume previous work."
}

// Parameters returns the tool parameters schema.
func (t *ContinueConceptTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "string",
				"description": "The ID of the concept to continue",
			},
		},
		"required": []string{"id"},
	}
}

// Execute continues a concept.
func (t *ContinueConceptTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.conceptManager == nil {
		return ErrorResult("Concept manager not initialized").WithError(fmt.Errorf("concept manager is nil"))
	}

	id, ok := args["id"].(string)
	if !ok || id == "" {
		return ErrorResult("id is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid id parameter"))
	}

	concept, err := t.conceptManager.ContinueConcept(id)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to continue concept: %v", err)).WithError(err)
	}

	// Build detailed context
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Continuing concept: %s\n", concept.Concept))
	sb.WriteString(fmt.Sprintf("ID: %s\n", concept.ID))
	sb.WriteString(fmt.Sprintf("Status: %s (now active)\n", concept.Status))
	sb.WriteString(fmt.Sprintf("Created: %s\n", concept.CreatedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Updated: %s\n\n", concept.UpdatedAt.Format(time.RFC3339)))

	// Context details
	if concept.Context.WorkingDirectory != "" {
		sb.WriteString(fmt.Sprintf("Working Directory: %s\n", concept.Context.WorkingDirectory))
	}
	if len(concept.Context.Files) > 0 {
		sb.WriteString(fmt.Sprintf("Files: %v\n", concept.Context.Files))
	}
	if len(concept.Context.Conversation) > 0 {
		sb.WriteString(fmt.Sprintf("\nConversation History (%d entries):\n", len(concept.Context.Conversation)))
		for i, entry := range concept.Context.Conversation {
			sb.WriteString(fmt.Sprintf("\n[%d] %s (%s):\n%s\n", i+1, entry.Role, entry.Timestamp.Format(time.RFC3339), entry.Content))
		}
	}
	if len(concept.Context.Metadata) > 0 {
		metadataJSON, _ := json.MarshalIndent(concept.Context.Metadata, "", "  ")
		sb.WriteString(fmt.Sprintf("\nMetadata:\n%s\n", string(metadataJSON)))
	}
	if len(concept.RelatedJobs) > 0 {
		sb.WriteString(fmt.Sprintf("\nRelated Jobs: %v\n", concept.RelatedJobs))
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Continuing concept: %s", concept.Concept),
		Silent:  false,
		IsError: false,
	}
}

// =============================================================================
// MemoryToolsRegistry
// =============================================================================

// MemoryToolsRegistry holds all memory-related tools.
type MemoryToolsRegistry struct {
	CreateConcept   *CreateConceptTool
	ListConcepts    *ListConceptsTool
	ContinueConcept *ContinueConceptTool
}

// NewMemoryToolsRegistry creates a new registry with all memory tools initialized.
func NewMemoryToolsRegistry(cm *memory.ConceptManager, jm *memory.JobManager) *MemoryToolsRegistry {
	return &MemoryToolsRegistry{
		CreateConcept:   NewCreateConceptTool(cm),
		ListConcepts:    NewListConceptsTool(cm),
		ContinueConcept: NewContinueConceptTool(cm),
	}
}

// RegisterAll registers all memory tools with the given tool registry.
func (r *MemoryToolsRegistry) RegisterAll(registry *ToolRegistry) {
	if r.CreateConcept != nil {
		registry.Register(r.CreateConcept)
	}
	if r.ListConcepts != nil {
		registry.Register(r.ListConcepts)
	}
	if r.ContinueConcept != nil {
		registry.Register(r.ContinueConcept)
	}
}

// GetAll returns all tools as a slice.
func (r *MemoryToolsRegistry) GetAll() []Tool {
	tools := make([]Tool, 0, 3)
	if r.CreateConcept != nil {
		tools = append(tools, r.CreateConcept)
	}
	if r.ListConcepts != nil {
		tools = append(tools, r.ListConcepts)
	}
	if r.ContinueConcept != nil {
		tools = append(tools, r.ContinueConcept)
	}
	return tools
}
