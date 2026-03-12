package agent

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/fileutil"
	"picoclaw/agent/pkg/logger"
)

// MemoryEntry represents a single fact stored in the structured memory system.
type MemoryEntry struct {
	ID          string    `json:"id"`
	Category    string    `json:"category"`
	Content     string    `json:"content"`
	Importance  float64   `json:"importance"`  // 1.0 (trivial) to 5.0 (critical)
	CreatedAt   time.Time `json:"created_at"`
	LastAccess  time.Time `json:"last_access"`
	AccessCount int       `json:"access_count"`
	Source      string    `json:"source,omitempty"` // "auto" | "user" | "consolidation"
}

// SessionSummary represents a conversation session summary for cross-session context.
type SessionSummary struct {
	ID         string    `json:"id"`
	Date       time.Time `json:"date"`
	AgentID    string    `json:"agent_id"`
	Capability string    `json:"capability,omitempty"`
	Summary    string    `json:"summary"`
	KeyTopics  []string  `json:"key_topics"`
	TurnCount  int       `json:"turn_count"`
}

// StructuredMemory manages the v2 memory system with importance scoring,
// semantic dedup, consolidation, and cross-session summaries.
// It coexists with MemoryStore (MEMORY.md) for backward compatibility.
type StructuredMemory struct {
	mu          sync.RWMutex
	workspace   string
	entriesPath string
	sessionsDir string
	entries     []MemoryEntry
}

// NewStructuredMemory creates or loads the structured memory store.
// Stores data in baseDir directly (following OpenClaw agent pattern).
func NewStructuredMemory(baseDir string) *StructuredMemory {
	os.MkdirAll(baseDir, 0o755)

	sessionsDir := filepath.Join(baseDir, "sessions")
	os.MkdirAll(sessionsDir, 0o755)

	sm := &StructuredMemory{
		workspace:   baseDir,
		entriesPath: filepath.Join(baseDir, "entries.json"),
		sessionsDir: sessionsDir,
	}
	sm.entries = sm.loadEntries()
	return sm
}

// ———— Entry CRUD ————

func (sm *StructuredMemory) loadEntries() []MemoryEntry {
	data, err := os.ReadFile(sm.entriesPath)
	if err != nil {
		return nil
	}
	var entries []MemoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		logger.WarnCF("memory-v2", "Failed to parse entries.json", map[string]any{"error": err.Error()})
		return nil
	}
	return entries
}

func (sm *StructuredMemory) saveEntries() error {
	data, err := json.MarshalIndent(sm.entries, "", "  ")
	if err != nil {
		return err
	}
	return fileutil.WriteFileAtomic(sm.entriesPath, data, 0o600)
}

// AddEntry adds a new memory entry with dedup check (exact match).
// Returns true if the entry was actually added (not a duplicate).
func (sm *StructuredMemory) AddEntry(entry MemoryEntry) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Exact dedup
	normalized := strings.ToLower(strings.TrimSpace(entry.Content))
	for _, e := range sm.entries {
		if strings.ToLower(strings.TrimSpace(e.Content)) == normalized {
			return false
		}
	}

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("m-%d", time.Now().UnixNano())
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	if entry.LastAccess.IsZero() {
		entry.LastAccess = entry.CreatedAt
	}
	if entry.Importance == 0 {
		entry.Importance = 3.0 // default mid-importance
	}

	sm.entries = append(sm.entries, entry)
	if err := sm.saveEntries(); err != nil {
		logger.WarnCF("memory-v2", "Failed to save entry", map[string]any{"error": err.Error()})
	}
	return true
}

// AddEntries adds multiple entries, returning count of actually added.
func (sm *StructuredMemory) AddEntries(entries []MemoryEntry) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	existingSet := make(map[string]bool, len(sm.entries))
	for _, e := range sm.entries {
		existingSet[strings.ToLower(strings.TrimSpace(e.Content))] = true
	}

	added := 0
	for _, entry := range entries {
		normalized := strings.ToLower(strings.TrimSpace(entry.Content))
		if existingSet[normalized] {
			continue
		}
		if entry.ID == "" {
			entry.ID = fmt.Sprintf("m-%d-%d", time.Now().UnixNano(), added)
		}
		if entry.CreatedAt.IsZero() {
			entry.CreatedAt = time.Now()
		}
		if entry.LastAccess.IsZero() {
			entry.LastAccess = entry.CreatedAt
		}
		if entry.Importance == 0 {
			entry.Importance = 3.0
		}
		sm.entries = append(sm.entries, entry)
		existingSet[normalized] = true
		added++
	}

	if added > 0 {
		if err := sm.saveEntries(); err != nil {
			logger.WarnCF("memory-v2", "Failed to save entries", map[string]any{"error": err.Error()})
		}
	}
	return added
}

// GetEntries returns all entries (read-only snapshot).
func (sm *StructuredMemory) GetEntries() []MemoryEntry {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	result := make([]MemoryEntry, len(sm.entries))
	copy(result, sm.entries)
	return result
}

// GetByCategory returns entries for a specific category.
func (sm *StructuredMemory) GetByCategory(category string) []MemoryEntry {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	var result []MemoryEntry
	for _, e := range sm.entries {
		if e.Category == category {
			result = append(result, e)
		}
	}
	return result
}

// TouchEntries bumps LastAccess and AccessCount for the given IDs.
func (sm *StructuredMemory) TouchEntries(ids []string) {
	if len(ids) == 0 {
		return
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()

	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}

	changed := false
	now := time.Now()
	for i := range sm.entries {
		if idSet[sm.entries[i].ID] {
			sm.entries[i].LastAccess = now
			sm.entries[i].AccessCount++
			changed = true
		}
	}
	if changed {
		_ = sm.saveEntries()
	}
}

// RemoveEntries removes entries by ID. Returns count removed.
func (sm *StructuredMemory) RemoveEntries(ids []string) int {
	if len(ids) == 0 {
		return 0
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()

	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}

	kept := sm.entries[:0]
	removed := 0
	for _, e := range sm.entries {
		if idSet[e.ID] {
			removed++
		} else {
			kept = append(kept, e)
		}
	}
	sm.entries = kept
	if removed > 0 {
		_ = sm.saveEntries()
	}
	return removed
}

// ReplaceEntry replaces an entry by ID (used by consolidation).
func (sm *StructuredMemory) ReplaceEntry(id string, updated MemoryEntry) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for i, e := range sm.entries {
		if e.ID == id {
			updated.ID = id
			sm.entries[i] = updated
			_ = sm.saveEntries()
			return true
		}
	}
	return false
}

// Count returns total entry count.
func (sm *StructuredMemory) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.entries)
}

// ———— Importance + Decay ————

// EffectiveImportance returns the decayed importance of an entry.
// Formula: base_importance * decay_factor
// decay_factor = 1.0 for entries accessed in the last day,
// halving every 30 days of inactivity. User-sourced entries decay slower.
func EffectiveImportance(e MemoryEntry) float64 {
	daysSinceAccess := time.Since(e.LastAccess).Hours() / 24.0
	halfLife := 30.0
	if e.Source == "user" {
		halfLife = 90.0 // user facts decay 3x slower
	}
	decay := math.Pow(0.5, daysSinceAccess/halfLife)

	// Access frequency bonus: +0.1 per access, capped at +1.0
	freqBonus := math.Min(float64(e.AccessCount)*0.1, 1.0)

	return e.Importance*decay + freqBonus
}

// PruneStale removes entries whose effective importance drops below threshold.
// Returns IDs of removed entries.
func (sm *StructuredMemory) PruneStale(threshold float64) []string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var pruned []string
	kept := sm.entries[:0]
	for _, e := range sm.entries {
		if EffectiveImportance(e) < threshold {
			pruned = append(pruned, e.ID)
		} else {
			kept = append(kept, e)
		}
	}
	sm.entries = kept
	if len(pruned) > 0 {
		_ = sm.saveEntries()
		logger.InfoCF("memory-v2", "Pruned stale memories", map[string]any{
			"count":    len(pruned),
			"threshold": threshold,
		})
	}
	return pruned
}

// ———— Retrieval (keyword-based, capability-aware) ————

// RetrieveRelevant returns the top entries relevant to a capability and message,
// sorted by a combined score of keyword overlap + effective importance.
func (sm *StructuredMemory) RetrieveRelevant(capability, message string, limit int) []MemoryEntry {
	sm.mu.RLock()
	entries := make([]MemoryEntry, len(sm.entries))
	copy(entries, sm.entries)
	sm.mu.RUnlock()

	if len(entries) == 0 || limit <= 0 {
		return nil
	}

	targetCat := CapabilityToMemCategory(capability)
	msgTokens := tokenizeForSearch(message)

	type scored struct {
		entry MemoryEntry
		score float64
	}

	var candidates []scored
	for _, e := range entries {
		score := EffectiveImportance(e) * 0.3 // base: importance contributes 30%

		// Category match
		if e.Category == targetCat {
			score += 2.0
		}
		if e.Category == MemCatProfile {
			score += 1.5 // profile always relevant
		}

		// Keyword overlap
		contentTokens := tokenizeForSearch(e.Content)
		for _, ct := range contentTokens {
			for _, mt := range msgTokens {
				if ct == mt || (len(ct) > 3 && strings.Contains(ct, mt)) || (len(mt) > 3 && strings.Contains(mt, ct)) {
					score += 1.0
					break
				}
			}
		}

		if score > 1.0 {
			candidates = append(candidates, scored{e, score})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	result := make([]MemoryEntry, 0, limit)
	for i, c := range candidates {
		if i >= limit {
			break
		}
		result = append(result, c.entry)
	}

	// Touch accessed entries
	if len(result) > 0 {
		ids := make([]string, len(result))
		for i, r := range result {
			ids[i] = r.ID
		}
		sm.TouchEntries(ids)
	}

	return result
}

// FormatForPrompt formats retrieved entries into a prompt-friendly string.
func FormatForPrompt(entries []MemoryEntry) string {
	if len(entries) == 0 {
		return ""
	}
	var sb strings.Builder
	// Group by category for readability
	byCategory := make(map[string][]MemoryEntry)
	for _, e := range entries {
		byCategory[e.Category] = append(byCategory[e.Category], e)
	}

	// Sort categories for deterministic output
	cats := make([]string, 0, len(byCategory))
	for c := range byCategory {
		cats = append(cats, c)
	}
	sort.Strings(cats)

	for _, cat := range cats {
		label := strings.ToUpper(cat[:1]) + cat[1:]
		sb.WriteString("### " + label + "\n")
		for _, e := range byCategory[cat] {
			sb.WriteString(e.Content + "\n")
		}
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n")
}

// ———— Session Summaries ————

// SaveSessionSummary saves a conversation session summary.
func (sm *StructuredMemory) SaveSessionSummary(summary SessionSummary) error {
	if summary.ID == "" {
		summary.ID = fmt.Sprintf("s-%s", time.Now().Format("20060102-150405"))
	}
	if summary.Date.IsZero() {
		summary.Date = time.Now()
	}

	entries := sm.loadSessionIndex()
	entries = append(entries, summary)

	// Keep last 100 session summaries
	if len(entries) > 100 {
		entries = entries[len(entries)-100:]
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	indexPath := filepath.Join(sm.sessionsDir, "index.json")
	return fileutil.WriteFileAtomic(indexPath, data, 0o600)
}

func (sm *StructuredMemory) loadSessionIndex() []SessionSummary {
	indexPath := filepath.Join(sm.sessionsDir, "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil
	}
	var entries []SessionSummary
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}
	return entries
}

// GetRecentSessions returns the N most recent session summaries.
func (sm *StructuredMemory) GetRecentSessions(limit int) []SessionSummary {
	entries := sm.loadSessionIndex()
	if len(entries) == 0 {
		return nil
	}
	if limit >= len(entries) {
		return entries
	}
	return entries[len(entries)-limit:]
}

// GetRecentSessionsForPrompt returns formatted session summaries for prompt injection.
func (sm *StructuredMemory) GetRecentSessionsForPrompt(limit int) string {
	sessions := sm.GetRecentSessions(limit)
	if len(sessions) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("### Recent Sessions\n")
	for _, s := range sessions {
		age := formatAge(s.Date)
		sb.WriteString(fmt.Sprintf("- **%s** (%s", s.Summary, age))
		if len(s.KeyTopics) > 0 {
			sb.WriteString(", topics: " + strings.Join(s.KeyTopics, ", "))
		}
		sb.WriteString(")\n")
	}
	return sb.String()
}

// ———— Migration: import existing MEMORY.md into structured entries ————

// MigrateFromMemoryMD reads existing MEMORY.md sections and imports them
// as structured entries. Skips entries that already exist (exact dedup).
// Returns count of imported entries.
func (sm *StructuredMemory) MigrateFromMemoryMD(workspace string) int {
	ms := NewMemoryStoreWithOptions(workspace, false)
	content := ms.ReadLongTerm()
	if content == "" {
		return 0
	}

	categories := []string{MemCatProfile, MemCatCode, MemCatResearch, MemCatWriting, MemCatGeneral}
	var toAdd []MemoryEntry

	for _, cat := range categories {
		section := extractSection(content, cat)
		if section == "" {
			continue
		}
		for _, line := range strings.Split(section, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Strip common markdown bullet points
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimPrefix(line, "* ")
			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			toAdd = append(toAdd, MemoryEntry{
				Category:   cat,
				Content:    line,
				Importance: 3.0,
				Source:     "migration",
			})
		}
	}

	if len(toAdd) == 0 {
		return 0
	}
	return sm.AddEntries(toAdd)
}

// ———— Sync: render entries back to MEMORY.md for backward compat ————

// SyncToMemoryMD renders all structured entries into MEMORY.md format.
// This keeps the flat file in sync for tools/agents that read it directly.
func (sm *StructuredMemory) SyncToMemoryMD(workspace string) error {
	sm.mu.RLock()
	entries := make([]MemoryEntry, len(sm.entries))
	copy(entries, sm.entries)
	sm.mu.RUnlock()

	// Group by category
	byCategory := make(map[string][]string)
	for _, e := range entries {
		byCategory[e.Category] = append(byCategory[e.Category], e.Content)
	}

	var sb strings.Builder
	sb.WriteString("# Long-term Memory\n")

	// Deterministic order
	cats := []string{MemCatProfile, MemCatCode, MemCatResearch, MemCatWriting, MemCatGeneral}
	for _, cat := range cats {
		facts, ok := byCategory[cat]
		if !ok || len(facts) == 0 {
			continue
		}
		sb.WriteString("\n## [" + cat + "]\n")
		for _, f := range facts {
			sb.WriteString(f + "\n")
		}
	}

	ms := NewMemoryStoreWithOptions(workspace, false)
	return ms.WriteLongTerm(sb.String())
}
