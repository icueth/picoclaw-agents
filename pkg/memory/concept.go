// Package memory provides concept and job tracking for work continuity.
package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/db"
)

// ConceptStatus represents the lifecycle status of a concept.
type ConceptStatus string

const (
	// ConceptActive - Currently being worked on
	ConceptActive ConceptStatus = "active"
	// ConceptPaused - Temporarily stopped
	ConceptPaused ConceptStatus = "paused"
	// ConceptCompleted - Finished successfully
	ConceptCompleted ConceptStatus = "completed"
	// ConceptAbandoned - No longer relevant
	ConceptAbandoned ConceptStatus = "abandoned"
)

// ValidConceptStatuses returns all valid concept statuses.
func ValidConceptStatuses() []string {
	return []string{
		string(ConceptActive),
		string(ConceptPaused),
		string(ConceptCompleted),
		string(ConceptAbandoned),
	}
}

// IsValidConceptStatus checks if a status is valid.
func IsValidConceptStatus(status string) bool {
	for _, s := range ValidConceptStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// ConceptContext stores the working context for a concept.
type ConceptContext struct {
	WorkingDirectory string                 `json:"working_directory,omitempty"`
	Files            []string               `json:"files,omitempty"`
	Conversation     []ConversationEntry    `json:"conversation,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ConversationEntry represents a single conversation entry.
type ConversationEntry struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Concept represents a high-level work concept with full context.
type Concept struct {
	ID          string         `json:"id"`
	Concept     string         `json:"concept"`
	Context     ConceptContext `json:"context"`
	RelatedJobs []string       `json:"related_jobs"`
	Status      ConceptStatus  `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ConceptManager handles high-level concept tracking for work continuity.
type ConceptManager struct {
	db     *db.DB
	config *config.MemoryConfig
}

// NewConceptManager creates a new ConceptManager instance.
func NewConceptManager(database *db.DB, cfg *config.MemoryConfig) *ConceptManager {
	if cfg == nil {
		cfg = &config.MemoryConfig{
			ConceptRetentionDays: 30,
		}
	}
	return &ConceptManager{
		db:     database,
		config: cfg,
	}
}

// CreateConcept creates a new concept with the given concept text and context.
func (cm *ConceptManager) CreateConcept(conceptText, contextStr string) (string, error) {
	if cm.db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	// Parse context string into ConceptContext
	ctx := ConceptContext{
		Metadata: make(map[string]interface{}),
	}
	if contextStr != "" {
		// Try to parse as JSON first
		if err := json.Unmarshal([]byte(contextStr), &ctx); err != nil {
			// If not JSON, treat as plain text in metadata
			ctx.Metadata["description"] = contextStr
		}
	}

	// Serialize context to JSON for storage
	contextJSON, err := json.Marshal(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to marshal context: %w", err)
	}

	// Create in database
	dbConcept, err := cm.db.CreateConcept(conceptText, string(contextJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create concept in database: %w", err)
	}

	return dbConcept.ID, nil
}

// CreateConceptWithContext creates a new concept with a structured context.
func (cm *ConceptManager) CreateConceptWithContext(conceptText string, ctx ConceptContext) (string, error) {
	if cm.db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	contextJSON, err := json.Marshal(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to marshal context: %w", err)
	}

	dbConcept, err := cm.db.CreateConcept(conceptText, string(contextJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create concept in database: %w", err)
	}

	return dbConcept.ID, nil
}

// ListConcepts returns all concepts, optionally filtered by status.
// If status is empty, returns all concepts.
func (cm *ConceptManager) ListConcepts(status string) ([]Concept, error) {
	if cm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	dbConcepts, err := cm.db.ListConcepts(status)
	if err != nil {
		return nil, fmt.Errorf("failed to list concepts: %w", err)
	}

	concepts := make([]Concept, len(dbConcepts))
	for i, dbConcept := range dbConcepts {
		concept, err := cm.dbConceptToConcept(dbConcept)
		if err != nil {
			return nil, err
		}
		concepts[i] = *concept
	}

	return concepts, nil
}

// GetConcept retrieves a concept by ID.
func (cm *ConceptManager) GetConcept(id string) (*Concept, error) {
	if cm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	dbConcept, err := cm.db.GetConcept(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get concept: %w", err)
	}

	return cm.dbConceptToConcept(dbConcept)
}

// ContinueConcept marks a concept as active and returns it for continuation.
func (cm *ConceptManager) ContinueConcept(id string) (*Concept, error) {
	if cm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Update status to active
	if err := cm.db.UpdateConceptStatus(id, string(ConceptActive)); err != nil {
		return nil, fmt.Errorf("failed to update concept status: %w", err)
	}

	// Return the updated concept
	return cm.GetConcept(id)
}

// UpdateConceptStatus updates the status of a concept.
func (cm *ConceptManager) UpdateConceptStatus(id, status string) error {
	if cm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if !IsValidConceptStatus(status) {
		return fmt.Errorf("invalid concept status: %s", status)
	}

	if err := cm.db.UpdateConceptStatus(id, status); err != nil {
		return fmt.Errorf("failed to update concept status: %w", err)
	}

	return nil
}

// UpdateConceptContext updates the context of a concept.
func (cm *ConceptManager) UpdateConceptContext(id string, ctx ConceptContext) error {
	if cm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	contextJSON, err := json.Marshal(ctx)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	if err := cm.db.UpdateConceptContext(id, string(contextJSON)); err != nil {
		return fmt.Errorf("failed to update concept context: %w", err)
	}

	return nil
}

// AddJobToConcept adds a job ID to a concept's related jobs list.
func (cm *ConceptManager) AddJobToConcept(conceptID, jobID string) error {
	if cm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := cm.db.AddRelatedJob(conceptID, jobID); err != nil {
		return fmt.Errorf("failed to add job to concept: %w", err)
	}

	return nil
}

// GetRelatedJobs retrieves all jobs related to a concept.
func (cm *ConceptManager) GetRelatedJobs(conceptID string) ([]db.Job, error) {
	if cm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	concept, err := cm.db.GetConcept(conceptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get concept: %w", err)
	}

	jobs := make([]db.Job, 0, len(concept.RelatedJobs))
	for _, jobID := range concept.RelatedJobs {
		job, err := cm.db.GetJob(jobID)
		if err != nil {
			// Job may have been deleted, skip it
			continue
		}
		jobs = append(jobs, *job)
	}

	return jobs, nil
}

// CleanupOldConcepts removes concepts older than the specified retention days.
// If retentionDays is 0, uses the value from config.
func (cm *ConceptManager) CleanupOldConcepts(retentionDays int) error {
	if cm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if retentionDays <= 0 {
		retentionDays = cm.config.ConceptRetentionDays
	}
	if retentionDays <= 0 {
		retentionDays = 30 // Default fallback
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	// Get all concepts
	concepts, err := cm.db.ListConcepts("")
	if err != nil {
		return fmt.Errorf("failed to list concepts for cleanup: %w", err)
	}

	// Delete old concepts
	for _, concept := range concepts {
		if concept.CreatedAt.Before(cutoff) && concept.Status != string(ConceptActive) {
			if err := cm.db.DeleteConcept(concept.ID); err != nil {
				// Log error but continue
				continue
			}
		}
	}

	return nil
}

// DeleteConcept permanently deletes a concept.
func (cm *ConceptManager) DeleteConcept(id string) error {
	if cm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := cm.db.DeleteConcept(id); err != nil {
		return fmt.Errorf("failed to delete concept: %w", err)
	}

	return nil
}

// dbConceptToConcept converts a database concept to a memory concept.
func (cm *ConceptManager) dbConceptToConcept(dbConcept *db.Concept) (*Concept, error) {
	concept := &Concept{
		ID:          dbConcept.ID,
		Concept:     dbConcept.Concept,
		RelatedJobs: dbConcept.RelatedJobs,
		Status:      ConceptStatus(dbConcept.Status),
		CreatedAt:   dbConcept.CreatedAt,
		UpdatedAt:   dbConcept.UpdatedAt,
		Context: ConceptContext{
			Metadata: make(map[string]interface{}),
		},
	}

	// Parse context JSON
	if dbConcept.Context != "" {
		if err := json.Unmarshal([]byte(dbConcept.Context), &concept.Context); err != nil {
			// If parsing fails, store raw context in metadata
			concept.Context.Metadata["raw_context"] = dbConcept.Context
		}
	}

	return concept, nil
}

// AddConversationEntry adds a conversation entry to a concept's context.
func (cm *ConceptManager) AddConversationEntry(conceptID string, entry ConversationEntry) error {
	concept, err := cm.GetConcept(conceptID)
	if err != nil {
		return err
	}

	concept.Context.Conversation = append(concept.Context.Conversation, entry)
	return cm.UpdateConceptContext(conceptID, concept.Context)
}

// AddFile adds a file reference to a concept's context.
func (cm *ConceptManager) AddFile(conceptID string, filePath string) error {
	concept, err := cm.GetConcept(conceptID)
	if err != nil {
		return err
	}

	// Check if file already exists
	for _, f := range concept.Context.Files {
		if f == filePath {
			return nil
		}
	}

	concept.Context.Files = append(concept.Context.Files, filePath)
	return cm.UpdateConceptContext(conceptID, concept.Context)
}

// SetWorkingDirectory sets the working directory for a concept.
func (cm *ConceptManager) SetWorkingDirectory(conceptID string, dir string) error {
	concept, err := cm.GetConcept(conceptID)
	if err != nil {
		return err
	}

	concept.Context.WorkingDirectory = dir
	return cm.UpdateConceptContext(conceptID, concept.Context)
}

// GenerateConceptID generates a new unique concept ID.
func GenerateConceptID() string {
	return uuid.New().String()
}

// StartCleanupRoutine starts a background routine that periodically cleans up old concepts.
func (cm *ConceptManager) StartCleanupRoutine(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 24 * time.Hour // Default daily cleanup
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = cm.CleanupOldConcepts(0) // Use config value
			}
		}
	}()
}
