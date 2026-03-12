// PicoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"picoclaw/agent/pkg/fileutil"
)

// MemoryStore manages persistent memory for the agent.
// - Long-term memory: memory/MEMORY.md
// - Daily notes: memory/YYYYMM/YYYYMMDD.md
type MemoryStore struct {
	workspace  string
	memoryDir  string
	memoryFile string
}

// NewMemoryStore creates a new MemoryStore with the given workspace path.
// It ensures the memory directory exists.
func NewMemoryStore(workspace string) *MemoryStore {
	return NewMemoryStoreWithOptions(workspace, true)
}

// NewMemoryStoreWithOptions creates a MemoryStore with optional memory subdirectory.
// If useSubdir is true: workspace/memory/MEMORY.md
// If useSubdir is false: workspace/MEMORY.md (for agent directory pattern)
func NewMemoryStoreWithOptions(workspace string, useSubdir bool) *MemoryStore {
	var memoryDir, memoryFile string
	
	if useSubdir {
		memoryDir = filepath.Join(workspace, "memory")
		memoryFile = filepath.Join(memoryDir, "MEMORY.md")
	} else {
		memoryDir = workspace
		memoryFile = filepath.Join(workspace, "MEMORY.md")
	}

	// Ensure memory directory exists
	os.MkdirAll(memoryDir, 0o755)

	return &MemoryStore{
		workspace:  workspace,
		memoryDir:  memoryDir,
		memoryFile: memoryFile,
	}
}

// getTodayFile returns the path to today's daily note file.
// For workspace/memory pattern: memory/YYYYMM/YYYYMMDD.md
// For agentDir pattern: notes/YYYYMM/YYYYMMDD.md
func (ms *MemoryStore) getTodayFile() string {
	today := time.Now().Format("20060102") // YYYYMMDD
	monthDir := today[:6]                  // YYYYMM
	
	// Check if this is agentDir pattern (MEMORY.md is directly in memoryDir)
	memoryMdPath := filepath.Join(ms.memoryDir, "MEMORY.md")
	if _, err := os.Stat(memoryMdPath); err == nil {
		// This is agentDir pattern, use notes/ subdir for daily notes
		filePath := filepath.Join(ms.memoryDir, "notes", monthDir, today+".md")
		return filePath
	}
	
	// Standard pattern: memory/YYYYMM/YYYYMMDD.md
	filePath := filepath.Join(ms.memoryDir, monthDir, today+".md")
	return filePath
}

// ReadLongTerm reads the long-term memory (MEMORY.md).
// Returns empty string if the file doesn't exist.
func (ms *MemoryStore) ReadLongTerm() string {
	if data, err := os.ReadFile(ms.memoryFile); err == nil {
		return string(data)
	}
	return ""
}

// WriteLongTerm writes content to the long-term memory file (MEMORY.md).
func (ms *MemoryStore) WriteLongTerm(content string) error {
	// Use unified atomic write utility with explicit sync for flash storage reliability.
	// Using 0o600 (owner read/write only) for secure default permissions.
	return fileutil.WriteFileAtomic(ms.memoryFile, []byte(content), 0o600)
}

// ReadToday reads today's daily note.
// Returns empty string if the file doesn't exist.
func (ms *MemoryStore) ReadToday() string {
	todayFile := ms.getTodayFile()
	if data, err := os.ReadFile(todayFile); err == nil {
		return string(data)
	}
	return ""
}

// AppendToday appends content to today's daily note.
// If the file doesn't exist, it creates a new file with a date header.
func (ms *MemoryStore) AppendToday(content string) error {
	todayFile := ms.getTodayFile()

	// Ensure month directory exists
	monthDir := filepath.Dir(todayFile)
	if err := os.MkdirAll(monthDir, 0o755); err != nil {
		return err
	}

	var existingContent string
	if data, err := os.ReadFile(todayFile); err == nil {
		existingContent = string(data)
	}

	var newContent string
	if existingContent == "" {
		// Add header for new day
		header := fmt.Sprintf("# %s\n\n", time.Now().Format("2006-01-02"))
		newContent = header + content
	} else {
		// Append to existing content
		newContent = existingContent + "\n" + content
	}

	// Use unified atomic write utility with explicit sync for flash storage reliability.
	return fileutil.WriteFileAtomic(todayFile, []byte(newContent), 0o600)
}

// GetRecentDailyNotes returns daily notes from the last N days.
// Contents are joined with "---" separator.
func (ms *MemoryStore) GetRecentDailyNotes(days int) string {
	var sb strings.Builder
	first := true

	// Determine if this is agentDir pattern (MEMORY.md directly in memoryDir)
	isAgentDir := false
	memoryMdPath := filepath.Join(ms.memoryDir, "MEMORY.md")
	if _, err := os.Stat(memoryMdPath); err == nil {
		isAgentDir = true
	}

	for i := range days {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("20060102") // YYYYMMDD
		monthDir := dateStr[:6]            // YYYYMM
		
		var filePath string
		if isAgentDir {
			// AgentDir pattern: notes/YYYYMM/YYYYMMDD.md
			filePath = filepath.Join(ms.memoryDir, "notes", monthDir, dateStr+".md")
		} else {
			// Standard pattern: memory/YYYYMM/YYYYMMDD.md
			filePath = filepath.Join(ms.memoryDir, monthDir, dateStr+".md")
		}

		if data, err := os.ReadFile(filePath); err == nil {
			if !first {
				sb.WriteString("\n\n---\n\n")
			}
			sb.Write(data)
			first = false
		}
	}

	return sb.String()
}

// ——— Capability-aware section-based memory ———

// Memory category tags used as section headers "## [tag]" in MEMORY.md.
const (
	MemCatProfile  = "profile"  // permanent user facts: name, language, contacts
	MemCatCode     = "code"     // coding languages, tools, patterns
	MemCatResearch = "research" // research interests, knowledge domains
	MemCatWriting  = "writing"  // writing style, content type preferences
	MemCatGeneral  = "general"  // other cross-cutting facts
)

// capabilityCategories maps agent capability names to memory categories.
var capabilityCategories = map[string]string{
	"code":        MemCatCode,
	"debug":       MemCatCode,
	"refactor":    MemCatCode,
	"research":    MemCatResearch,
	"analysis":    MemCatResearch,
	"information": MemCatResearch,
	"writing":     MemCatWriting,
	"creative":    MemCatWriting,
	"content":     MemCatWriting,
}

// CapabilityToMemCategory maps an agent capability name to its memory category.
// Returns MemCatGeneral if no specific mapping exists.
func CapabilityToMemCategory(capability string) string {
	if cat, ok := capabilityCategories[strings.ToLower(capability)]; ok {
		return cat
	}
	return MemCatGeneral
}

// extractSection returns the trimmed body of a "## [tag]" section.
// Returns "" if the section is not present.
func extractSection(content, tag string) string {
	header := "## [" + tag + "]"
	idx := strings.Index(content, header)
	if idx == -1 {
		return ""
	}
	start := idx + len(header)
	if start < len(content) && content[start] == '\n' {
		start++
	}
	rest := content[start:]
	if next := strings.Index(rest, "\n## ["); next != -1 {
		return strings.TrimSpace(rest[:next])
	}
	return strings.TrimSpace(rest)
}

// ReadSection returns the body of the "## [tag]" section in MEMORY.md.
func (ms *MemoryStore) ReadSection(tag string) string {
	return extractSection(ms.ReadLongTerm(), tag)
}

// WriteSection deduplicates and appends new facts into the tagged section
// of MEMORY.md, creating the section if it doesn't exist.
func (ms *MemoryStore) WriteSection(tag string, newFacts []string) error {
	if len(newFacts) == 0 {
		return nil
	}
	content := ms.ReadLongTerm()
	if content == "" {
		content = "# Long-term Memory\n"
	}

	// Build dedup set from existing section content.
	existing := extractSection(content, tag)
	seen := make(map[string]bool)
	for _, l := range strings.Split(existing, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			seen[strings.ToLower(t)] = true
		}
	}

	var unique []string
	for _, f := range newFacts {
		if t := strings.TrimSpace(f); t != "" && !seen[strings.ToLower(t)] {
			unique = append(unique, t)
			seen[strings.ToLower(t)] = true
		}
	}
	if len(unique) == 0 {
		return nil
	}

	header := "## [" + tag + "]"
	idx := strings.Index(content, header)

	var nb strings.Builder
	if idx == -1 {
		// Append a new section at the end.
		nb.WriteString(strings.TrimRight(content, "\n"))
		nb.WriteString("\n\n")
		nb.WriteString(header)
		nb.WriteString("\n")
		for _, f := range unique {
			nb.WriteString(f)
			nb.WriteString("\n")
		}
	} else {
		// Insert into existing section before the next "## [" header.
		afterHeader := content[idx+len(header):]
		newline := ""
		if strings.HasPrefix(afterHeader, "\n") {
			newline = "\n"
			afterHeader = afterHeader[1:]
		}
		nextSec := strings.Index(afterHeader, "\n## [")
		var body, tail string
		if nextSec == -1 {
			body, tail = afterHeader, ""
		} else {
			body, tail = afterHeader[:nextSec], afterHeader[nextSec:]
		}
		nb.WriteString(content[:idx+len(header)])
		nb.WriteString(newline)
		nb.WriteString(body)
		if body != "" && !strings.HasSuffix(body, "\n") {
			nb.WriteByte('\n')
		}
		for _, f := range unique {
			nb.WriteString(f)
			nb.WriteByte('\n')
		}
		nb.WriteString(tail)
	}
	return ms.WriteLongTerm(nb.String())
}

// GetMemoryForCapability returns memory relevant to the given capability.
// Always includes the [profile] section plus the capability-mapped category.
// Falls back to full GetMemoryContext() when no tagged sections exist (backward compat).
func (ms *MemoryStore) GetMemoryForCapability(capability string) string {
	profile := ms.ReadSection(MemCatProfile)
	cat := CapabilityToMemCategory(capability)
	specific := ""
	if cat != MemCatProfile {
		specific = ms.ReadSection(cat)
	}

	// Backward compat: no tagged sections yet → return full memory.
	if profile == "" && specific == "" {
		return ms.GetMemoryContext()
	}

	var sb strings.Builder
	if profile != "" {
		sb.WriteString("### User Profile\n")
		sb.WriteString(profile)
	}
	if specific != "" {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		label := strings.ToUpper(cat[:1]) + cat[1:]
		sb.WriteString("### ")
		sb.WriteString(label)
		sb.WriteString("\n")
		sb.WriteString(specific)
	}
	return sb.String()
}

// GetMemoryContext returns formatted memory context for the agent prompt.
// Includes long-term memory and recent daily notes.
func (ms *MemoryStore) GetMemoryContext() string {
	longTerm := ms.ReadLongTerm()
	recentNotes := ms.GetRecentDailyNotes(3)

	if longTerm == "" && recentNotes == "" {
		return ""
	}

	var sb strings.Builder

	if longTerm != "" {
		sb.WriteString("## Long-term Memory\n\n")
		sb.WriteString(longTerm)
	}

	if recentNotes != "" {
		if longTerm != "" {
			sb.WriteString("\n\n---\n\n")
		}
		sb.WriteString("## Recent Daily Notes\n\n")
		sb.WriteString(recentNotes)
	}

	return sb.String()
}
