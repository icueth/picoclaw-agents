package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"picoclaw/agent/pkg/fileutil"
)

// TaskEntry represents a single completed task stored in the task archive.
type TaskEntry struct {
	ID         string    `json:"id"`
	Date       time.Time `json:"date"`
	Capability string    `json:"capability"`
	Title      string    `json:"title"`
	Tags       []string  `json:"tags"`
	Summary    string    `json:"summary"`
}

// TaskArchive manages an append-only index of completed task summaries.
// Tasks are stored in memory/tasks/index.json for quick retrieval.
type TaskArchive struct {
	indexPath string
}

// NewTaskArchive creates (or opens) the task archive for the given workspace.
func NewTaskArchive(workspace string) *TaskArchive {
	tasksDir := filepath.Join(workspace, "memory", "tasks")
	os.MkdirAll(tasksDir, 0o755)
	return &TaskArchive{
		indexPath: filepath.Join(tasksDir, "index.json"),
	}
}

// LoadIndex reads all task entries from disk.
func (ta *TaskArchive) LoadIndex() []TaskEntry {
	data, err := os.ReadFile(ta.indexPath)
	if err != nil {
		return nil
	}
	var entries []TaskEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}
	return entries
}

// SaveEntry appends a task entry to the index (capped at 500 entries).
func (ta *TaskArchive) SaveEntry(entry TaskEntry) error {
	entries := ta.LoadIndex()
	entries = append(entries, entry)
	if len(entries) > 500 {
		entries = entries[len(entries)-500:]
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return fileutil.WriteFileAtomic(ta.indexPath, data, 0o600)
}

// RetrieveRelevant returns up to limit task entries most relevant to
// the given capability and user message, scored by keyword overlap + recency.
func (ta *TaskArchive) RetrieveRelevant(capability, userMessage string, limit int) []TaskEntry {
	entries := ta.LoadIndex()
	if len(entries) == 0 {
		return nil
	}

	msgTokens := tokenizeForSearch(userMessage)
	now := time.Now()

	type scored struct {
		entry TaskEntry
		score float64
	}

	var candidates []scored
	for _, e := range entries {
		score := 0.0

		// Same capability: strong match
		if capability != "" && e.Capability == capability {
			score += 3.0
		}

		// Tag overlap with current message tokens
		for _, tag := range e.Tags {
			tagLower := strings.ToLower(tag)
			for _, word := range msgTokens {
				if tagLower == word || strings.Contains(tagLower, word) || strings.Contains(word, tagLower) {
					score += 1.5
					break
				}
			}
		}

		// Title word overlap
		titleTokens := tokenizeForSearch(e.Title)
		for _, tt := range titleTokens {
			for _, word := range msgTokens {
				if tt == word {
					score += 0.5
					break
				}
			}
		}

		// Recency bonus
		age := now.Sub(e.Date)
		switch {
		case age < 24*time.Hour:
			score += 1.0
		case age < 7*24*time.Hour:
			score += 0.5
		case age < 30*24*time.Hour:
			score += 0.2
		}

		if score > 0.8 {
			candidates = append(candidates, scored{e, score})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	result := make([]TaskEntry, 0, limit)
	for i, c := range candidates {
		if i >= limit {
			break
		}
		result = append(result, c.entry)
	}
	return result
}

// GetRelevantContext returns a formatted snippet of relevant past tasks
// suitable for injection into the LLM system prompt.
func (ta *TaskArchive) GetRelevantContext(capability, userMessage string, limit int) string {
	entries := ta.RetrieveRelevant(capability, userMessage, limit)
	if len(entries) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, e := range entries {
		age := formatAge(e.Date)
		sb.WriteString(fmt.Sprintf("- **%s** (%s", e.Title, age))
		if e.Capability != "" {
			sb.WriteString(", " + e.Capability)
		}
		sb.WriteString("): ")
		sb.WriteString(e.Summary)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// tokenizeForSearch splits text into lowercase tokens for keyword matching.
// Keeps Unicode characters (Thai) intact by splitting only on ASCII punctuation/spaces.
func tokenizeForSearch(text string) []string {
	var tokens []string
	var cur strings.Builder

	for _, r := range text {
		switch {
		case r == ' ' || r == '\t' || r == '\n' || r == '\r' ||
			r == ',' || r == '.' || r == ':' || r == ';' ||
			r == '(' || r == ')' || r == '[' || r == ']' ||
			r == '"' || r == '\'' || r == '/' || r == '\\' ||
			r == '-' || r == '_':
			if tok := strings.ToLower(cur.String()); runeLen(tok) >= 2 {
				tokens = append(tokens, tok)
			}
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	if tok := strings.ToLower(cur.String()); runeLen(tok) >= 2 {
		tokens = append(tokens, tok)
	}
	return tokens
}

func runeLen(s string) int { return len([]rune(s)) }

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < 2*24*time.Hour:
		return "today"
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d weeks ago", int(d.Hours()/(24*7)))
	default:
		return t.Format("Jan 2")
	}
}
