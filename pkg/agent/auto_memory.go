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

// extractBackgroundTasks runs task archival, memory extraction, and session summary
// sequentially in a single goroutine. This prevents concurrent API calls that compete
// for provider semaphore slots and cause timeouts on rate-limited APIs like DashScope.
func (al *AgentLoop) extractBackgroundTasks(
	agent *AgentInstance,
	userMessage string,
	assistantMessage string,
	capability string,
	sessionKey string,
) {
	if userMessage == "" || assistantMessage == "" {
		return
	}

	al.bgTasks.Add(1)
	go func() {
		defer al.bgTasks.Done()

		// Brief delay to avoid rate-limiting — main response just finished using the API
		time.Sleep(2 * time.Second)

		// Step 1: Archive task summary (skip trivial exchanges)
		al.doExtractTask(agent, userMessage, assistantMessage, capability)

		// Step 2: Extract memory facts (runs after task archival finishes)
		al.doExtractMemory(agent, userMessage, assistantMessage, capability)

		// Step 3: Generate session summary (every 5 turns for cross-session context)
		al.maybeGenerateSessionSummary(agent, sessionKey, capability)
	}()
}

// maybeGenerateSessionSummary generates a cross-session summary when the conversation
// has accumulated enough turns. Summaries are saved for future session context injection.
func (al *AgentLoop) maybeGenerateSessionSummary(
	agent *AgentInstance,
	sessionKey string,
	capability string,
) {
	if sessionKey == "" {
		return
	}

	history := agent.Sessions.GetHistory(sessionKey)
	if len(history) < 6 { // at least 3 user+assistant turn pairs
		return
	}

	// Only generate every 5 turns (message count / 2 = turn count)
	turnCount := len(history) / 2
	if turnCount%5 != 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Collect user and assistant messages
	var userMsgs, assistantMsgs []string
	for _, m := range history {
		switch m.Role {
		case "user":
			userMsgs = append(userMsgs, m.Content)
		case "assistant":
			assistantMsgs = append(assistantMsgs, m.Content)
		}
	}

	summary, err := GenerateSessionSummary(ctx, agent.Provider, agent.Model, agent.ID, capability, userMsgs, assistantMsgs)
	if err != nil || summary == nil {
		logger.DebugCF("memory", "Session summary generation skipped", map[string]any{"error": fmt.Sprintf("%v", err)})
		return
	}

	sm := NewStructuredMemory(agent.Workspace)
	if err := sm.SaveSessionSummary(*summary); err != nil {
		logger.WarnCF("memory", "Failed to save session summary", map[string]any{"error": err.Error()})
		return
	}

	logger.InfoCF("memory", "Session summary saved", map[string]any{
		"session_key": sessionKey,
		"agent_id":    agent.ID,
		"turns":       turnCount,
		"topics":      summary.KeyTopics,
	})
}

// doExtractMemory extracts categorized facts from the conversation, scores their
// importance, checks for semantic duplicates, and stores them in StructuredMemory.
// Falls back to legacy MemoryStore if structured operations fail.
// Called sequentially from extractBackgroundTasks.
func (al *AgentLoop) doExtractMemory(
	agent *AgentInstance,
	userMessage string,
	assistantMessage string,
	capability string,
) {
	logger.InfoCF("memory", "doExtractMemory STARTED", map[string]any{
		"agent_id":   agent.ID,
		"capability": capability,
		"user_msg_len": len(userMessage),
		"assistant_msg_len": len(assistantMessage),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	logger.DebugCF("memory", "Starting automatic memory extraction (v2)", map[string]any{
		"agent_id":   agent.ID,
		"capability": capability,
	})

	// Step 1: Extract raw facts via LLM (same prompt as before)
	systemPrompt := `You are an automatic memory extractor. Analyze the conversation.
Extract ONLY permanent facts about the user to remember across sessions.
DO NOT extract temporary details, code content, or conversation summaries.

Output format — one fact per line, prefixed with its category:
[profile] <fact>    ← name, language, contacts, general permanent preferences
[code] <fact>       ← coding languages, tools, style preferences
[research] <fact>   ← research interests, knowledge domains
[writing] <fact>    ← writing style, content type preferences
[general] <fact>    ← other permanent facts that don't fit above

If NO new permanent facts exist, output exactly: NONE`

	userPrompt := "User Message:\n" + userMessage + "\n\nAssistant Reply:\n" + assistantMessage

	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	resp, err := agent.Provider.Chat(ctx, messages, nil, agent.Model, map[string]any{
		"temperature": 0.0,
		"max_tokens":  300,
	})
	if err != nil {
		logger.WarnCF("memory", "Failed to extract memory via LLM", map[string]any{"error": err.Error()})
		logger.InfoCF("memory", "doExtractMemory EXIT: LLM error", map[string]any{"agent_id": agent.ID})
		return
	}

	logger.InfoCF("memory", "LLM raw response", map[string]any{
		"agent_id":     agent.ID,
		"resp_nil":     resp == nil,
		"content_len":  len(resp.Content),
		"tool_calls":   len(resp.ToolCalls),
	})

	extracted := strings.TrimSpace(resp.Content)
	logger.InfoCF("memory", "Memory extraction LLM response", map[string]any{
		"agent_id":   agent.ID,
		"extracted":  extracted,
		"raw_content": resp.Content,
		"has_mm":     agent.ContextBuilder != nil && agent.ContextBuilder.memoryManager != nil,
		"rag_enabled": agent.ContextBuilder != nil && agent.ContextBuilder.memoryManager != nil && agent.ContextBuilder.memoryManager.IsRAGEnabled(),
	})
	if extracted == "" || strings.EqualFold(extracted, "NONE") {
		logger.InfoCF("memory", "doExtractMemory EXIT: No facts (NONE or empty)", map[string]any{
			"agent_id":  agent.ID,
			"extracted": extracted,
		})
		return
	}

	// Parse "[category] fact" lines
	var rawFacts []string
	byCategory := make(map[string][]string)
	for _, line := range strings.Split(extracted, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cat, fact := parseCategoryLine(line)
		if fact != "" {
			byCategory[cat] = append(byCategory[cat], fact)
			rawFacts = append(rawFacts, fact)
		}
	}

	if len(byCategory) == 0 {
		logger.InfoCF("memory", "doExtractMemory EXIT: No parseable facts", map[string]any{
			"agent_id": agent.ID,
			"raw":      extracted,
		})
		return
	}

	// Step 2: Score importance via LLM
	sm := NewStructuredMemory(agent.Workspace)
	scoredFacts, scoreErr := ScoreImportance(ctx, agent.Provider, agent.Model, rawFacts)

	// Build entries from scored facts (or fallback to raw facts)
	var newEntries []MemoryEntry
	if scoreErr == nil && len(scoredFacts) > 0 {
		for _, sf := range scoredFacts {
			cat := sf.Category
			if cat == "" {
				cat = MemCatGeneral
			}
			newEntries = append(newEntries, MemoryEntry{
				Category:   cat,
				Content:    sf.Content,
				Importance: sf.Importance,
				Source:     "auto",
			})
		}
	} else {
		// Fallback: use parsed categories with default importance
		for cat, facts := range byCategory {
			for _, f := range facts {
				newEntries = append(newEntries, MemoryEntry{
					Category:   cat,
					Content:    f,
					Importance: 3.0,
					Source:     "auto",
				})
			}
		}
	}

	// Step 3: Semantic dedup — check each new entry against existing same-category entries
	existing := sm.GetEntries()
	var finalEntries []MemoryEntry
	for _, ne := range newEntries {
		// Filter existing entries in same category for dedup check
		var sameCat []MemoryEntry
		for _, e := range existing {
			if e.Category == ne.Category {
				sameCat = append(sameCat, e)
			}
		}

		// Skip semantic dedup if few existing entries (faster, less API calls)
		if len(sameCat) < 3 {
			finalEntries = append(finalEntries, ne)
			continue
		}

		dedupResult, dedupErr := CheckSemanticDuplicates(ctx, agent.Provider, agent.Model, ne.Content, sameCat)
		if dedupErr != nil || len(dedupResult.DuplicateIDs) == 0 {
			// No duplicates found or error — add as new
			finalEntries = append(finalEntries, ne)
			continue
		}

		// Merge: replace oldest duplicate with merged version, remove others
		if dedupResult.MergedFact != "" {
			ne.Content = dedupResult.MergedFact
			// Boost importance slightly for merged facts
			ne.Importance = ne.Importance + 0.5
			if ne.Importance > 5.0 {
				ne.Importance = 5.0
			}
		}
		sm.RemoveEntries(dedupResult.DuplicateIDs)
		finalEntries = append(finalEntries, ne)

		logger.DebugCF("memory", "Semantic dedup merged", map[string]any{
			"merged_ids": dedupResult.DuplicateIDs,
			"new_fact":   ne.Content,
		})
	}

	// Step 4: Add final entries to structured memory
	added := sm.AddEntries(finalEntries)

	// Step 5: Sync to MEMORY.md for backward compatibility
	if added > 0 {
		if err := sm.SyncToMemoryMD(agent.Workspace); err != nil {
			logger.WarnCF("memory", "Failed to sync to MEMORY.md", map[string]any{"error": err.Error()})
		}
	}

	// Step 6: Save to RAG for semantic search (if MemoryManager is available)
	logger.DebugCF("memory", "Checking RAG save conditions",
		map[string]any{
			"added":       added,
			"has_cb":      agent.ContextBuilder != nil,
			"has_mm":      agent.ContextBuilder != nil && agent.ContextBuilder.memoryManager != nil,
		})
	if added > 0 && agent.ContextBuilder != nil && agent.ContextBuilder.memoryManager != nil {
		if agent.ContextBuilder.memoryManager.IsRAGEnabled() {
			ragAdded := 0
			for _, entry := range finalEntries {
				content := fmt.Sprintf("[%s] %s", entry.Category, entry.Content)
				tags := []string{entry.Category, "memory", "auto-extracted"}
				source := fmt.Sprintf("agent:%s:memory:%s", agent.ID, entry.ID)
				contentPreview := content
			if len(contentPreview) > 50 {
				contentPreview = contentPreview[:50] + "..."
			}
			logger.DebugCF("memory", "Saving fact to RAG",
					map[string]any{"entry_id": entry.ID, "content": contentPreview})
				if err := agent.ContextBuilder.memoryManager.SaveFact(content, tags, source); err != nil {
					logger.WarnCF("memory", "Failed to save to RAG",
						map[string]any{"error": err.Error(), "entry_id": entry.ID})
				} else {
					ragAdded++
				}
			}
			if ragAdded > 0 {
				logger.InfoCF("memory", "Saved memories to RAG",
					map[string]any{"count": ragAdded, "agent_id": agent.ID})
			}
		} else {
			logger.DebugCF("memory", "RAG not enabled, skipping RAG save", nil)
		}
	} else {
		logger.DebugCF("memory", "Skipping RAG save - conditions not met",
			map[string]any{
				"added":       added,
				"has_cb":      agent.ContextBuilder != nil,
				"has_mm":      agent.ContextBuilder != nil && agent.ContextBuilder.memoryManager != nil,
			})
	}

	logger.InfoCF("memory", "Memory updated (v2)", map[string]any{
		"agent_id":   agent.ID,
		"capability": capability,
		"extracted":  len(rawFacts),
		"added":      added,
		"total":      sm.Count(),
	})

	logger.InfoCF("memory", "doExtractMemory COMPLETED", map[string]any{
		"agent_id":   agent.ID,
		"added":      added,
	})

	// Step 7: Trigger consolidation every 10 new entries
	if sm.Count() > 0 && sm.Count()%10 < added {
		al.doConsolidate(agent, sm)
	}
}

// doConsolidate runs LLM-based memory consolidation to merge related entries
// and prune redundant ones. Called periodically from doExtractMemory.
func (al *AgentLoop) doConsolidate(agent *AgentInstance, sm *StructuredMemory) {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	entries := sm.GetEntries()
	if len(entries) < 10 {
		return
	}

	logger.InfoCF("memory", "Starting memory consolidation", map[string]any{
		"entry_count": len(entries),
	})

	result, err := ConsolidateMemories(ctx, agent.Provider, agent.Model, entries)
	if err != nil {
		logger.WarnCF("memory", "Consolidation failed", map[string]any{"error": err.Error()})
		return
	}

	ApplyConsolidation(sm, result)

	// Also prune stale entries (effective importance < 1.0)
	pruned := sm.PruneStale(1.0)

	// Sync changes to MEMORY.md
	if len(result.Merged) > 0 || len(result.Removed) > 0 || len(pruned) > 0 {
		_ = sm.SyncToMemoryMD(agent.Workspace)
	}

	logger.InfoCF("memory", "Consolidation complete", map[string]any{
		"merged":  len(result.Merged),
		"removed": len(result.Removed),
		"pruned":  len(pruned),
		"total":   sm.Count(),
	})
}

// doExtractTask summarizes the completed task and saves it to the task archive.
// Called sequentially from extractBackgroundTasks.
func (al *AgentLoop) doExtractTask(
	agent *AgentInstance,
	userMessage string,
	assistantMessage string,
	capability string,
) {
	// Only archive substantive exchanges (skip greetings, trivial replies)
	if len(assistantMessage) < 150 || userMessage == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// Truncate assistant message to keep prompt short
	excerpt := assistantMessage
	if len(excerpt) > 600 {
		excerpt = excerpt[:600] + "…"
	}

	systemPrompt := `Summarize the completed task in JSON.
Fields:
- title: short task title (5-10 words)
- summary: 1-2 sentences of what was requested and what was produced
- tags: 3-8 lowercase keywords relevant to retrieval

Output ONLY valid JSON, no markdown fences. Example:
{"title":"Python sort function","summary":"User requested sorted_list() with type hints. Provided implementation using built-in sorted().","tags":["python","sort","type-hints","function"]}`

	userPrompt := "User: " + userMessage + "\n\nAssistant: " + excerpt

	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	resp, err := agent.Provider.Chat(ctx, messages, nil, agent.Model, map[string]any{
		"temperature": 0.0,
		"max_tokens":  250,
	})
	if err != nil || strings.TrimSpace(resp.Content) == "" {
		return
	}

	raw := strings.TrimSpace(resp.Content)
	// Strip markdown fences if present
	if start := strings.Index(raw, "{"); start >= 0 {
		if end := strings.LastIndex(raw, "}"); end > start {
			raw = raw[start : end+1]
		}
	}

	var result struct {
		Title   string   `json:"title"`
		Summary string   `json:"summary"`
		Tags    []string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil || result.Title == "" {
		return
	}

	// Ensure capability is in tags
	if capability != "" {
		hasTag := false
		for _, t := range result.Tags {
			if strings.EqualFold(t, capability) {
				hasTag = true
				break
			}
		}
		if !hasTag {
			result.Tags = append(result.Tags, capability)
		}
	}

	entry := TaskEntry{
		ID:         fmt.Sprintf("%s-%s", time.Now().Format("20060102-150405"), capability),
		Date:       time.Now(),
		Capability: capability,
		Title:      result.Title,
		Tags:       result.Tags,
		Summary:    result.Summary,
	}

	archive := NewTaskArchive(agent.Workspace)
	if err := archive.SaveEntry(entry); err != nil {
		logger.WarnCF("memory", "Failed to save task entry", map[string]any{"error": err.Error()})
		return
	}

	logger.InfoCF("memory", "Task archived", map[string]any{
		"agent_id":   agent.ID,
		"capability": capability,
		"title":      entry.Title,
		"tags":       entry.Tags,
	})
}

// parseCategoryLine parses a line of the form "[category] fact text".
// Returns (category, fact). Falls back to (MemCatGeneral, line) if no tag found.
func parseCategoryLine(line string) (string, string) {
	if !strings.HasPrefix(line, "[") {
		// No tag — treat as general fact (strip leading "- " if present)
		return MemCatGeneral, strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
	}
	end := strings.Index(line, "]")
	if end == -1 {
		return MemCatGeneral, line
	}
	cat := strings.ToLower(strings.TrimSpace(line[1:end]))
	fact := strings.TrimSpace(line[end+1:])
	fact = strings.TrimPrefix(fact, " ")
	// Validate category; default to general for unknown tags.
	switch cat {
	case MemCatProfile, MemCatCode, MemCatResearch, MemCatWriting, MemCatGeneral:
	default:
		cat = MemCatGeneral
	}
	if fact == "" {
		return cat, ""
	}
	// Normalise: prepend "- " bullet if not already present.
	if !strings.HasPrefix(fact, "- ") {
		fact = "- " + fact
	}
	return cat, fact
}
