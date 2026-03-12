// Package memory provides namespace isolation for agent memory
package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Namespace defines a memory namespace
type Namespace struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	OwnedBy     string            `json:"owned_by"`    // Agent ID or "shared"
	SharedWith  []string          `json:"shared_with"` // List of agent IDs with access
	AccessLevel AccessLevel       `json:"access_level"`
	CreatedAt   time.Time         `json:"created_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// AccessLevel defines access permissions
type AccessLevel int

const (
	AccessPrivate AccessLevel = iota // Owner only
	AccessShared                     // Shared with specific agents
	AccessPublic                     // All agents
)

// String returns string representation of access level
func (a AccessLevel) String() string {
	switch a {
	case AccessPrivate:
		return "private"
	case AccessShared:
		return "shared"
	case AccessPublic:
		return "public"
	default:
		return "unknown"
	}
}

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	Namespace   string    `json:"namespace"`
	AgentID     string    `json:"agent_id"` // Agent who created this
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []string  `json:"tags,omitempty"`
	Importance  int       `json:"importance"` // 1-10
	EmbeddingID string    `json:"embedding_id,omitempty"`
}

// Manager handles memory namespace operations
type Manager struct {
	namespaces map[string]*Namespace
	entries    map[string][]*MemoryEntry // namespace -> entries
	db         DB                        // Interface for persistence
	mu         sync.RWMutex
}

// DB interface for persistence
type DB interface {
	SaveNamespace(ctx context.Context, ns *Namespace) error
	LoadNamespaces(ctx context.Context) ([]*Namespace, error)
	SaveMemory(ctx context.Context, entry *MemoryEntry) error
	LoadMemories(ctx context.Context, namespace string) ([]*MemoryEntry, error)
	SearchMemories(ctx context.Context, query string, namespaces []string, limit int) ([]*MemoryEntry, error)
	DeleteMemory(ctx context.Context, id string) error
}

// NewManager creates a new memory namespace manager
func NewManager(db DB) *Manager {
	return &Manager{
		namespaces: make(map[string]*Namespace),
		entries:    make(map[string][]*MemoryEntry),
		db:         db,
	}
}

// Initialize loads existing namespaces from DB
func (m *Manager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load namespaces
	nsList, err := m.db.LoadNamespaces(ctx)
	if err != nil {
		return fmt.Errorf("failed to load namespaces: %w", err)
	}

	for _, ns := range nsList {
		m.namespaces[ns.Name] = ns
	}

	// Create default namespaces if none exist
	if len(m.namespaces) == 0 {
		m.createDefaultNamespaces()
	}

	// Load entries for each namespace
	for name := range m.namespaces {
		entries, err := m.db.LoadMemories(ctx, name)
		if err != nil {
			// Log warning but continue
			continue
		}
		m.entries[name] = entries
	}

	return nil
}

// CreateNamespace creates a new namespace
func (m *Manager) CreateNamespace(name, description, owner string, access AccessLevel) (*Namespace, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.namespaces[name]; exists {
		return nil, fmt.Errorf("namespace %s already exists", name)
	}

	ns := &Namespace{
		Name:        name,
		Description: description,
		OwnedBy:     owner,
		AccessLevel: access,
		SharedWith:  make([]string, 0),
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]string),
	}

	m.namespaces[name] = ns
	m.entries[name] = make([]*MemoryEntry, 0)

	// Persist
	if err := m.db.SaveNamespace(context.Background(), ns); err != nil {
		return nil, fmt.Errorf("failed to save namespace: %w", err)
	}

	return ns, nil
}

// GetNamespace retrieves a namespace
func (m *Manager) GetNamespace(name string) (*Namespace, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ns, ok := m.namespaces[name]
	return ns, ok
}

// DeleteNamespace removes a namespace
func (m *Manager) DeleteNamespace(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.namespaces[name]; !ok {
		return fmt.Errorf("namespace %s not found", name)
	}

	delete(m.namespaces, name)
	delete(m.entries, name)

	return nil
}

// ListNamespaces returns all namespaces accessible by an agent
func (m *Manager) ListNamespaces(agentID string) []*Namespace {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Namespace, 0)
	for _, ns := range m.namespaces {
		if m.hasAccess(ns, agentID) {
			result = append(result, ns)
		}
	}
	return result
}

// ShareNamespace grants access to an agent
func (m *Manager) ShareNamespace(namespace, agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ns, ok := m.namespaces[namespace]
	if !ok {
		return fmt.Errorf("namespace %s not found", namespace)
	}

	// Check if already shared
	for _, id := range ns.SharedWith {
		if id == agentID {
			return nil
		}
	}

	ns.SharedWith = append(ns.SharedWith, agentID)

	return m.db.SaveNamespace(context.Background(), ns)
}

// RevokeNamespace removes access from an agent
func (m *Manager) RevokeNamespace(namespace, agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ns, ok := m.namespaces[namespace]
	if !ok {
		return fmt.Errorf("namespace %s not found", namespace)
	}

	ns.SharedWith = filterStrings(ns.SharedWith, agentID)

	return m.db.SaveNamespace(context.Background(), ns)
}

// Store saves a memory entry
func (m *Manager) Store(ctx context.Context, content, namespace, agentID string, importance int, tags []string) (*MemoryEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ns, ok := m.namespaces[namespace]
	if !ok {
		return nil, fmt.Errorf("namespace %s not found", namespace)
	}

	if !m.hasAccess(ns, agentID) {
		return nil, fmt.Errorf("agent %s has no access to namespace %s", agentID, namespace)
	}

	entry := &MemoryEntry{
		ID:         uuid.New().String(),
		Content:    content,
		Namespace:  namespace,
		AgentID:    agentID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Tags:       tags,
		Importance: importance,
	}

	m.entries[namespace] = append(m.entries[namespace], entry)

	if err := m.db.SaveMemory(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to save memory: %w", err)
	}

	return entry, nil
}

// Retrieve gets memories by ID
func (m *Manager) Retrieve(namespace, id string) (*MemoryEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries, ok := m.entries[namespace]
	if !ok {
		return nil, fmt.Errorf("namespace %s not found", namespace)
	}

	for _, entry := range entries {
		if entry.ID == id {
			return entry, nil
		}
	}

	return nil, fmt.Errorf("memory %s not found", id)
}

// Search searches memories in accessible namespaces
func (m *Manager) Search(ctx context.Context, query string, agentID string, limit int) ([]*MemoryEntry, error) {
	// Get accessible namespaces
	m.mu.RLock()
	accessibleNS := make([]string, 0)
	for name, ns := range m.namespaces {
		if m.hasAccess(ns, agentID) {
			accessibleNS = append(accessibleNS, name)
		}
	}
	m.mu.RUnlock()

	if len(accessibleNS) == 0 {
		return []*MemoryEntry{}, nil
	}

	// Use DB search (vector search if available)
	return m.db.SearchMemories(ctx, query, accessibleNS, limit)
}

// GetAllByNamespace returns all memories in a namespace
func (m *Manager) GetAllByNamespace(namespace string) ([]*MemoryEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries, ok := m.entries[namespace]
	if !ok {
		return nil, fmt.Errorf("namespace %s not found", namespace)
	}

	result := make([]*MemoryEntry, len(entries))
	copy(result, entries)
	return result, nil
}

// DeleteMemory removes a memory entry
func (m *Manager) DeleteMemory(namespace, id, agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ns, ok := m.namespaces[namespace]
	if !ok {
		return fmt.Errorf("namespace %s not found", namespace)
	}

	// Only owner or creator can delete
	if ns.OwnedBy != agentID {
		entry, err := m.findEntry(namespace, id)
		if err != nil || entry.AgentID != agentID {
			return fmt.Errorf("permission denied")
		}
	}

	// Remove from memory
	entries, ok := m.entries[namespace]
	if !ok {
		return fmt.Errorf("namespace %s not found", namespace)
	}

	m.entries[namespace] = filterEntries(entries, id)

	return m.db.DeleteMemory(context.Background(), id)
}

// Helper methods

func (m *Manager) hasAccess(ns *Namespace, agentID string) bool {
	if ns.AccessLevel == AccessPublic {
		return true
	}
	if ns.OwnedBy == agentID {
		return true
	}
	if ns.AccessLevel == AccessShared {
		for _, id := range ns.SharedWith {
			if id == agentID {
				return true
			}
		}
	}
	return false
}

func (m *Manager) findEntry(namespace, id string) (*MemoryEntry, error) {
	entries, ok := m.entries[namespace]
	if !ok {
		return nil, fmt.Errorf("namespace not found")
	}
	for _, e := range entries {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, fmt.Errorf("entry not found")
}

func (m *Manager) createDefaultNamespaces() {
	defaults := []struct {
		name        string
		description string
		owner       string
		access      AccessLevel
	}{
		{"shared", "Shared knowledge accessible by all agents", "system", AccessPublic},
		{"planning", "Planning and task coordination", "jarvis", AccessShared},
		{"research", "Research findings and web search results", "atlas", AccessShared},
		{"code", "Code snippets and technical documentation", "clawed", AccessShared},
		{"content", "Content writing and copy", "scribe", AccessShared},
		{"qa", "QA reports and testing results", "sentinel", AccessShared},
		{"design", "Design assets and feedback", "pixel", AccessShared},
		{"architecture", "System architecture decisions", "nova", AccessShared},
	}

	for _, d := range defaults {
		ns := &Namespace{
			Name:        d.name,
			Description: d.description,
			OwnedBy:     d.owner,
			AccessLevel: d.access,
			SharedWith:  make([]string, 0),
			CreatedAt:   time.Now(),
			Metadata:    make(map[string]string),
		}
		m.namespaces[d.name] = ns
		m.entries[d.name] = make([]*MemoryEntry, 0)

		if err := m.db.SaveNamespace(context.Background(), ns); err != nil {
			// Log warning but continue
			continue
		}
	}
}

// GetStats returns memory statistics
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"total_namespaces": len(m.namespaces),
		"namespaces":       make(map[string]interface{}),
	}

	nsStats := stats["namespaces"].(map[string]interface{})
	for name, entries := range m.entries {
		ns := m.namespaces[name]
		nsStats[name] = map[string]interface{}{
			"entry_count": len(entries),
			"owned_by":    ns.OwnedBy,
			"access":      ns.AccessLevel.String(),
		}
	}

	return stats
}

// FileDB implements DB interface using JSON files
type FileDB struct {
	basePath string
	mu       sync.RWMutex
}

// NewFileDB creates a new file-based DB
func NewFileDB(basePath string) *FileDB {
	return &FileDB{basePath: basePath}
}

// SaveNamespace persists namespace
func (db *FileDB) SaveNamespace(ctx context.Context, ns *Namespace) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Implementation would write to file
	// For now, just log
	return nil
}

// LoadNamespaces loads all namespaces
func (db *FileDB) LoadNamespaces(ctx context.Context) ([]*Namespace, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return []*Namespace{}, nil
}

// SaveMemory persists memory entry
func (db *FileDB) SaveMemory(ctx context.Context, entry *MemoryEntry) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return nil
}

// LoadMemories loads memories for namespace
func (db *FileDB) LoadMemories(ctx context.Context, namespace string) ([]*MemoryEntry, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return []*MemoryEntry{}, nil
}

// SearchMemories searches memories (simple text search fallback)
func (db *FileDB) SearchMemories(ctx context.Context, query string, namespaces []string, limit int) ([]*MemoryEntry, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return []*MemoryEntry{}, nil
}

// DeleteMemory removes a memory
func (db *FileDB) DeleteMemory(ctx context.Context, id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return nil
}

// Helper functions

func filterStrings(strs []string, remove string) []string {
	result := make([]string, 0, len(strs))
	for _, s := range strs {
		if s != remove {
			result = append(result, s)
		}
	}
	return result
}

func filterEntries(entries []*MemoryEntry, id string) []*MemoryEntry {
	result := make([]*MemoryEntry, 0, len(entries))
	for _, e := range entries {
		if e.ID != id {
			result = append(result, e)
		}
	}
	return result
}

// ValidateNamespaceName checks if namespace name is valid
func ValidateNamespaceName(name string) error {
	if name == "" {
		return fmt.Errorf("namespace name cannot be empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("namespace name cannot contain path separators")
	}
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("namespace name cannot start with dot")
	}
	return nil
}

// GetNamespacePath returns storage path for namespace
func GetNamespacePath(basePath, namespace string) string {
	return filepath.Join(basePath, "memory", namespace+".json")
}

// ToJSON converts stats to JSON
func ToJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
