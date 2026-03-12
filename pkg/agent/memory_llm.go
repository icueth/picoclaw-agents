package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/providers"
)

// ———— Semantic Dedup ————

// SemanticDedupResult is the LLM response for dedup checking.
type SemanticDedupResult struct {
	DuplicateIDs []string `json:"duplicate_ids"` // IDs of existing entries that are semantically same
	MergedFact   string   `json:"merged_fact"`   // consolidated version if duplicates found
}

// CheckSemanticDuplicates uses LLM to find existing entries that are semantically
// equivalent to a new fact. Returns IDs of duplicates and a merged version.
func CheckSemanticDuplicates(
	ctx context.Context,
	provider providers.LLMProvider,
	model string,
	newFact string,
	existingEntries []MemoryEntry,
) (*SemanticDedupResult, error) {
	if len(existingEntries) == 0 {
		return &SemanticDedupResult{}, nil
	}

	// Build compact list of existing entries
	var existing strings.Builder
	for _, e := range existingEntries {
		existing.WriteString(fmt.Sprintf("[%s] %s\n", e.ID, e.Content))
	}

	systemPrompt := `You are a memory deduplication system. Given a NEW fact and a list of EXISTING memory entries, determine if the new fact is semantically equivalent to any existing entries.

Two facts are duplicates if they convey the same information, even if worded differently or in different languages.

Output JSON only:
{"duplicate_ids": ["id1", "id2"], "merged_fact": "best consolidated version of all duplicates + new fact"}

If NO duplicates found:
{"duplicate_ids": [], "merged_fact": ""}`

	userPrompt := fmt.Sprintf("NEW FACT:\n%s\n\nEXISTING ENTRIES:\n%s", newFact, existing.String())

	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	resp, err := provider.Chat(ctx, messages, nil, model, map[string]any{
		"temperature": 0.0,
		"max_tokens":  200,
	})
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(resp.Content)
	raw = extractJSON(raw)

	var result SemanticDedupResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return &SemanticDedupResult{}, nil // fail open: treat as no duplicates
	}
	return &result, nil
}

// ———— Importance Scoring ————

// ImportanceScoringResult is the LLM response for importance scoring.
type ImportanceScoringResult struct {
	Facts []ScoredFact `json:"facts"`
}

// ScoredFact pairs a fact with its importance score.
type ScoredFact struct {
	Content    string  `json:"content"`
	Category   string  `json:"category"`
	Importance float64 `json:"importance"`
}

// ScoreImportance uses LLM to assign importance scores (1.0-5.0) to extracted facts.
func ScoreImportance(
	ctx context.Context,
	provider providers.LLMProvider,
	model string,
	facts []string,
) ([]ScoredFact, error) {
	if len(facts) == 0 {
		return nil, nil
	}

	systemPrompt := `Score each memory fact by importance (1.0-5.0):
5.0 = Critical permanent info (name, primary language, key contacts)
4.0 = Important preference (preferred tools, coding style)
3.0 = Useful context (project details, current interests)
2.0 = Minor detail (one-time preference, temporary context)
1.0 = Trivial (greeting style, passing remark)

Output JSON array only:
{"facts": [{"content": "fact text", "category": "profile|code|research|writing|general", "importance": 4.0}]}`

	userPrompt := "Facts to score:\n"
	for _, f := range facts {
		userPrompt += "- " + f + "\n"
	}

	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	resp, err := provider.Chat(ctx, messages, nil, model, map[string]any{
		"temperature": 0.0,
		"max_tokens":  500,
	})
	if err != nil {
		return nil, err
	}

	raw := extractJSON(strings.TrimSpace(resp.Content))
	var result ImportanceScoringResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		// Fallback: return facts with default importance
		scored := make([]ScoredFact, len(facts))
		for i, f := range facts {
			scored[i] = ScoredFact{Content: f, Category: MemCatGeneral, Importance: 3.0}
		}
		return scored, nil
	}
	return result.Facts, nil
}

// ———— Consolidation ————

// ConsolidationResult is the LLM response for memory consolidation.
type ConsolidationResult struct {
	Merged  []MergedGroup `json:"merged"`
	Removed []string      `json:"removed"` // IDs to remove (redundant/outdated)
}

// MergedGroup represents a group of entries merged into one.
type MergedGroup struct {
	SourceIDs  []string `json:"source_ids"`
	Content    string   `json:"content"`
	Category   string   `json:"category"`
	Importance float64  `json:"importance"`
}

// ConsolidateMemories uses LLM to merge related/redundant entries and prune outdated ones.
// Should be called periodically (e.g., after every 10 new entries or daily).
func ConsolidateMemories(
	ctx context.Context,
	provider providers.LLMProvider,
	model string,
	entries []MemoryEntry,
) (*ConsolidationResult, error) {
	if len(entries) < 5 {
		return &ConsolidationResult{}, nil // too few to consolidate
	}

	// Build entry list for LLM
	var entryList strings.Builder
	for _, e := range entries {
		entryList.WriteString(fmt.Sprintf("[%s|%s|%.1f] %s\n", e.ID, e.Category, e.Importance, e.Content))
	}

	systemPrompt := `You are a memory consolidation system. Analyze the memory entries and:
1. MERGE entries that cover the same topic into a single better-worded entry
2. REMOVE entries that are outdated, superseded, or trivially redundant
3. Keep entries that are unique and valuable

Output JSON only:
{
  "merged": [
    {"source_ids": ["id1", "id2"], "content": "merged fact text", "category": "profile", "importance": 4.0}
  ],
  "removed": ["id3", "id4"]
}

If nothing to consolidate: {"merged": [], "removed": []}`

	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: "MEMORY ENTRIES:\n" + entryList.String()},
	}

	resp, err := provider.Chat(ctx, messages, nil, model, map[string]any{
		"temperature": 0.0,
		"max_tokens":  800,
	})
	if err != nil {
		return nil, err
	}

	raw := extractJSON(strings.TrimSpace(resp.Content))
	var result ConsolidationResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return &ConsolidationResult{}, nil
	}
	return &result, nil
}

// ApplyConsolidation applies consolidation results to the structured memory.
func ApplyConsolidation(sm *StructuredMemory, result *ConsolidationResult) {
	if result == nil {
		return
	}

	// Remove redundant entries
	if len(result.Removed) > 0 {
		removed := sm.RemoveEntries(result.Removed)
		logger.InfoCF("memory-v2", "Consolidation removed entries", map[string]any{"count": removed})
	}

	// Apply merges: remove source entries, add merged entry
	for _, mg := range result.Merged {
		if len(mg.SourceIDs) < 2 || mg.Content == "" {
			continue
		}
		sm.RemoveEntries(mg.SourceIDs)
		sm.AddEntry(MemoryEntry{
			Category:   mg.Category,
			Content:    mg.Content,
			Importance: mg.Importance,
			Source:     "consolidation",
		})
	}
}

// ———— Cross-session Summary ————

// GenerateSessionSummary uses LLM to create a concise session summary.
func GenerateSessionSummary(
	ctx context.Context,
	provider providers.LLMProvider,
	model string,
	agentID string,
	capability string,
	userMessages []string,
	assistantMessages []string,
) (*SessionSummary, error) {
	if len(userMessages) == 0 {
		return nil, nil
	}

	// Build a compact conversation excerpt
	var convo strings.Builder
	maxTurns := 10
	startIdx := 0
	if len(userMessages) > maxTurns {
		startIdx = len(userMessages) - maxTurns
	}
	for i := startIdx; i < len(userMessages); i++ {
		convo.WriteString("User: " + truncateStr(userMessages[i], 200) + "\n")
		if i < len(assistantMessages) {
			convo.WriteString("Assistant: " + truncateStr(assistantMessages[i], 300) + "\n")
		}
	}

	systemPrompt := `Summarize this conversation session in JSON:
{
  "summary": "1-2 sentence summary of what was discussed/accomplished",
  "key_topics": ["topic1", "topic2", "topic3"]
}
Output ONLY valid JSON.`

	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: convo.String()},
	}

	resp, err := provider.Chat(ctx, messages, nil, model, map[string]any{
		"temperature": 0.0,
		"max_tokens":  200,
	})
	if err != nil {
		return nil, err
	}

	raw := extractJSON(strings.TrimSpace(resp.Content))
	var result struct {
		Summary   string   `json:"summary"`
		KeyTopics []string `json:"key_topics"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil || result.Summary == "" {
		return nil, fmt.Errorf("failed to parse session summary: %w", err)
	}

	return &SessionSummary{
		ID:         fmt.Sprintf("s-%s", time.Now().Format("20060102-150405")),
		Date:       time.Now(),
		AgentID:    agentID,
		Capability: capability,
		Summary:    result.Summary,
		KeyTopics:  result.KeyTopics,
		TurnCount:  len(userMessages),
	}, nil
}

// ———— helpers ————

func extractJSON(s string) string {
	objStart := strings.Index(s, "{")
	arrStart := strings.Index(s, "[")

	// Pick whichever delimiter appears first
	if objStart == -1 && arrStart == -1 {
		return s
	}

	if arrStart != -1 && (objStart == -1 || arrStart < objStart) {
		end := strings.LastIndex(s, "]")
		if end > arrStart {
			return s[arrStart : end+1]
		}
	}

	if objStart != -1 {
		end := strings.LastIndex(s, "}")
		if end > objStart {
			return s[objStart : end+1]
		}
	}

	return s
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "…"
}
