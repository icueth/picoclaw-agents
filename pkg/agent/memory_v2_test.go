package agent

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStructuredMemory_AddEntry(t *testing.T) {
	dir := t.TempDir()
	sm := NewStructuredMemory(dir)

	added := sm.AddEntry(MemoryEntry{
		Category:   MemCatProfile,
		Content:    "User speaks Thai",
		Importance: 4.0,
		Source:     "auto",
	})
	if !added {
		t.Fatal("expected entry to be added")
	}
	if sm.Count() != 1 {
		t.Fatalf("expected 1 entry, got %d", sm.Count())
	}

	// Exact dedup — same content should not be added again
	added2 := sm.AddEntry(MemoryEntry{
		Category:   MemCatProfile,
		Content:    "User speaks Thai",
		Importance: 4.0,
	})
	if added2 {
		t.Fatal("expected duplicate to be rejected")
	}
	if sm.Count() != 1 {
		t.Fatalf("expected 1 entry after dedup, got %d", sm.Count())
	}

	// Case-insensitive dedup
	added3 := sm.AddEntry(MemoryEntry{
		Category: MemCatProfile,
		Content:  "user speaks thai",
	})
	if added3 {
		t.Fatal("expected case-insensitive duplicate to be rejected")
	}
}

func TestStructuredMemory_AddEntries(t *testing.T) {
	dir := t.TempDir()
	sm := NewStructuredMemory(dir)

	entries := []MemoryEntry{
		{Category: MemCatCode, Content: "Uses Go", Importance: 3.0},
		{Category: MemCatCode, Content: "Prefers Vim", Importance: 2.0},
		{Category: MemCatCode, Content: "Uses Go", Importance: 3.0}, // duplicate
	}
	count := sm.AddEntries(entries)
	if count != 2 {
		t.Fatalf("expected 2 added, got %d", count)
	}
	if sm.Count() != 2 {
		t.Fatalf("expected 2 total entries, got %d", sm.Count())
	}
}

func TestStructuredMemory_GetByCategory(t *testing.T) {
	dir := t.TempDir()
	sm := NewStructuredMemory(dir)

	sm.AddEntries([]MemoryEntry{
		{Category: MemCatProfile, Content: "Name is John"},
		{Category: MemCatCode, Content: "Uses Python"},
		{Category: MemCatCode, Content: "Prefers TDD"},
		{Category: MemCatGeneral, Content: "Lives in Bangkok"},
	})

	codeEntries := sm.GetByCategory(MemCatCode)
	if len(codeEntries) != 2 {
		t.Fatalf("expected 2 code entries, got %d", len(codeEntries))
	}
}

func TestStructuredMemory_RemoveEntries(t *testing.T) {
	dir := t.TempDir()
	sm := NewStructuredMemory(dir)

	sm.AddEntries([]MemoryEntry{
		{ID: "a1", Category: MemCatProfile, Content: "Fact A"},
		{ID: "a2", Category: MemCatProfile, Content: "Fact B"},
		{ID: "a3", Category: MemCatCode, Content: "Fact C"},
	})

	removed := sm.RemoveEntries([]string{"a1", "a3"})
	if removed != 2 {
		t.Fatalf("expected 2 removed, got %d", removed)
	}
	if sm.Count() != 1 {
		t.Fatalf("expected 1 remaining, got %d", sm.Count())
	}

	remaining := sm.GetEntries()
	if remaining[0].ID != "a2" {
		t.Fatalf("expected a2 to remain, got %s", remaining[0].ID)
	}
}

func TestStructuredMemory_TouchEntries(t *testing.T) {
	dir := t.TempDir()
	sm := NewStructuredMemory(dir)

	sm.AddEntry(MemoryEntry{ID: "t1", Category: MemCatProfile, Content: "Touch test"})

	before := sm.GetEntries()[0]
	if before.AccessCount != 0 {
		t.Fatal("expected initial access count 0")
	}

	time.Sleep(10 * time.Millisecond)
	sm.TouchEntries([]string{"t1"})

	// Re-load to verify persistence
	sm2 := NewStructuredMemory(dir)
	after := sm2.GetEntries()[0]
	if after.AccessCount != 1 {
		t.Fatalf("expected access count 1 after touch, got %d", after.AccessCount)
	}
	if !after.LastAccess.After(before.LastAccess) {
		t.Fatal("expected LastAccess to be updated")
	}
}

func TestEffectiveImportance(t *testing.T) {
	// Fresh entry — no decay
	fresh := MemoryEntry{
		Importance:  4.0,
		LastAccess:  time.Now(),
		AccessCount: 0,
		Source:      "auto",
	}
	eff := EffectiveImportance(fresh)
	if eff < 3.9 || eff > 4.1 {
		t.Fatalf("expected ~4.0 for fresh entry, got %.2f", eff)
	}

	// Old entry (60 days) — should decay significantly
	old := MemoryEntry{
		Importance:  4.0,
		LastAccess:  time.Now().Add(-60 * 24 * time.Hour),
		AccessCount: 0,
		Source:      "auto",
	}
	effOld := EffectiveImportance(old)
	if effOld >= 2.0 {
		t.Fatalf("expected decayed importance < 2.0, got %.2f", effOld)
	}

	// User-sourced entry decays slower
	userOld := MemoryEntry{
		Importance:  4.0,
		LastAccess:  time.Now().Add(-60 * 24 * time.Hour),
		AccessCount: 0,
		Source:      "user",
	}
	effUser := EffectiveImportance(userOld)
	if effUser <= effOld {
		t.Fatalf("expected user entry to decay slower: user=%.2f, auto=%.2f", effUser, effOld)
	}

	// Frequently accessed entry gets bonus
	frequent := MemoryEntry{
		Importance:  2.0,
		LastAccess:  time.Now().Add(-30 * 24 * time.Hour),
		AccessCount: 10,
		Source:      "auto",
	}
	effFreq := EffectiveImportance(frequent)
	// Should have +1.0 frequency bonus
	if effFreq < 1.5 {
		t.Fatalf("expected frequency bonus to help, got %.2f", effFreq)
	}
}

func TestStructuredMemory_PruneStale(t *testing.T) {
	dir := t.TempDir()
	sm := NewStructuredMemory(dir)

	sm.AddEntries([]MemoryEntry{
		{ID: "fresh", Category: MemCatProfile, Content: "Fresh fact", Importance: 4.0},
		{ID: "stale", Category: MemCatGeneral, Content: "Stale fact", Importance: 1.0},
	})

	// Artificially age the "stale" entry
	sm.mu.Lock()
	for i := range sm.entries {
		if sm.entries[i].ID == "stale" {
			sm.entries[i].LastAccess = time.Now().Add(-365 * 24 * time.Hour)
			sm.entries[i].Importance = 0.5
		}
	}
	_ = sm.saveEntries()
	sm.mu.Unlock()

	pruned := sm.PruneStale(1.0)
	if len(pruned) != 1 || pruned[0] != "stale" {
		t.Fatalf("expected stale entry to be pruned, got %v", pruned)
	}
	if sm.Count() != 1 {
		t.Fatalf("expected 1 remaining after prune, got %d", sm.Count())
	}
}

func TestStructuredMemory_RetrieveRelevant(t *testing.T) {
	dir := t.TempDir()
	sm := NewStructuredMemory(dir)

	sm.AddEntries([]MemoryEntry{
		{Category: MemCatProfile, Content: "User name is John", Importance: 5.0},
		{Category: MemCatCode, Content: "Prefers Python for scripting", Importance: 3.0},
		{Category: MemCatCode, Content: "Uses Docker for deployment", Importance: 3.0},
		{Category: MemCatResearch, Content: "Interested in quantum computing", Importance: 2.0},
	})

	// Query about Python coding — should prioritize code + profile entries
	results := sm.RetrieveRelevant("code", "write a Python script", 3)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	// First result should be the Python one (keyword match + category match)
	foundPython := false
	for _, r := range results {
		if r.Content == "Prefers Python for scripting" {
			foundPython = true
			break
		}
	}
	if !foundPython {
		t.Fatal("expected Python entry in results")
	}
}

func TestFormatForPrompt(t *testing.T) {
	entries := []MemoryEntry{
		{Category: MemCatProfile, Content: "User speaks Thai"},
		{Category: MemCatCode, Content: "Uses Go"},
		{Category: MemCatProfile, Content: "Name is John"},
	}

	output := FormatForPrompt(entries)
	if output == "" {
		t.Fatal("expected non-empty output")
	}
	// Should have category headers
	if !contains(output, "### Code") || !contains(output, "### Profile") {
		t.Fatalf("expected category headers, got:\n%s", output)
	}
}

func TestStructuredMemory_SessionSummary(t *testing.T) {
	dir := t.TempDir()
	sm := NewStructuredMemory(dir)

	// Save session summaries
	err := sm.SaveSessionSummary(SessionSummary{
		Summary:   "Discussed Go coding patterns",
		KeyTopics: []string{"go", "patterns"},
		TurnCount: 5,
	})
	if err != nil {
		t.Fatalf("failed to save session summary: %v", err)
	}

	err = sm.SaveSessionSummary(SessionSummary{
		Summary:   "Fixed deployment issues",
		KeyTopics: []string{"docker", "deploy"},
		TurnCount: 3,
	})
	if err != nil {
		t.Fatalf("failed to save session summary: %v", err)
	}

	// Retrieve
	sessions := sm.GetRecentSessions(5)
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	// Format for prompt
	prompt := sm.GetRecentSessionsForPrompt(2)
	if prompt == "" {
		t.Fatal("expected non-empty session prompt")
	}
	if !contains(prompt, "Go coding") || !contains(prompt, "deployment") {
		t.Fatalf("expected session content in prompt, got:\n%s", prompt)
	}
}

func TestStructuredMemory_SyncToMemoryMD(t *testing.T) {
	dir := t.TempDir()
	sm := NewStructuredMemory(dir)

	sm.AddEntries([]MemoryEntry{
		{Category: MemCatProfile, Content: "Name is John"},
		{Category: MemCatCode, Content: "Uses Go"},
	})

	err := sm.SyncToMemoryMD(dir)
	if err != nil {
		t.Fatalf("SyncToMemoryMD failed: %v", err)
	}

	// Verify MEMORY.md was created
	data, err := os.ReadFile(filepath.Join(dir, "MEMORY.md"))
	if err != nil {
		t.Fatalf("failed to read MEMORY.md: %v", err)
	}
	content := string(data)
	if !contains(content, "[profile]") || !contains(content, "[code]") {
		t.Fatalf("expected sections in MEMORY.md, got:\n%s", content)
	}
	if !contains(content, "Name is John") || !contains(content, "Uses Go") {
		t.Fatalf("expected facts in MEMORY.md, got:\n%s", content)
	}
}

func TestStructuredMemory_MigrateFromMemoryMD(t *testing.T) {
	dir := t.TempDir()

	// Write legacy MEMORY.md
	content := `# Long-term Memory

## [profile]
- User name is Alice
- Speaks Thai and English

## [code]
- Prefers Go
- Uses VSCode
`
	os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte(content), 0o600)

	sm := NewStructuredMemory(dir)
	imported := sm.MigrateFromMemoryMD(dir)
	if imported != 4 {
		t.Fatalf("expected 4 imported entries, got %d", imported)
	}
	if sm.Count() != 4 {
		t.Fatalf("expected 4 total entries, got %d", sm.Count())
	}

	// Verify categories
	profileEntries := sm.GetByCategory(MemCatProfile)
	if len(profileEntries) != 2 {
		t.Fatalf("expected 2 profile entries, got %d", len(profileEntries))
	}
}

func TestStructuredMemory_Persistence(t *testing.T) {
	dir := t.TempDir()

	// Create and populate
	sm1 := NewStructuredMemory(dir)
	sm1.AddEntry(MemoryEntry{Category: MemCatProfile, Content: "Persistent fact", Importance: 4.0})

	// Re-open and verify
	sm2 := NewStructuredMemory(dir)
	if sm2.Count() != 1 {
		t.Fatalf("expected 1 entry after re-open, got %d", sm2.Count())
	}
	entries := sm2.GetEntries()
	if entries[0].Content != "Persistent fact" {
		t.Fatalf("expected 'Persistent fact', got %q", entries[0].Content)
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"key": "value"}`, `{"key": "value"}`},
		{"```json\n{\"key\": \"value\"}\n```", `{"key": "value"}`},
		{`Some text {"key": "value"} more text`, `{"key": "value"}`},
		{`[{"a": 1}]`, `[{"a": 1}]`},
		{`no json here`, `no json here`},
	}
	for _, tt := range tests {
		result := extractJSON(tt.input)
		if result != tt.expected {
			t.Errorf("extractJSON(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
