// Message Summarizer for A2A Token Optimization (Phase 3)
// Summarizes older A2A messages to reduce context size

package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/fileutil"
	"picoclaw/agent/pkg/logger"
)

// MessageSummarizer manages message summarization for A2A projects
type MessageSummarizer struct {
	mu sync.RWMutex

	// Configuration
	RecentMessagesKeep    int           // Number of recent messages to keep in full (default: 5)
	SummarizeThreshold    int           // Min messages before summarization (default: 20)
	ArchiveDir            string        // Directory for archived full content
	MaxSummaryLength      int           // Max chars for summary (default: 500)

	// State
	messageOrder []string              // Message IDs in chronological order
	summaries    map[string]*A2AMessage // Summarized messages
}

// NewMessageSummarizer creates a new summarizer
func NewMessageSummarizer(archiveDir string) *MessageSummarizer {
	os.MkdirAll(archiveDir, 0o755)

	return &MessageSummarizer{
		RecentMessagesKeep: 5,
		SummarizeThreshold: 20,
		ArchiveDir:         archiveDir,
		MaxSummaryLength:   500,
		messageOrder:       make([]string, 0),
		summaries:          make(map[string]*A2AMessage),
	}
}

// SummarizeResult contains summarization results
type SummarizeResult struct {
	MessagesKept       int
	MessagesSummarized int
	MessagesArchived   int
	TotalSavings       int // Token savings
}

// ShouldSummarize checks if summarization is needed
func (ms *MessageSummarizer) ShouldSummarize(totalMessages int) bool {
	return totalMessages > ms.SummarizeThreshold
}

// SummarizeMessages processes messages and summarizes older ones
// Strategy:
// - Recent N messages: Keep full content
// - Older messages: Summarize to key points
// - Very old messages: Archive to disk, keep summary only
func (ms *MessageSummarizer) SummarizeMessages(messages []A2AMessage) ([]A2AMessage, SummarizeResult) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	result := SummarizeResult{}

	if len(messages) <= ms.RecentMessagesKeep {
		return messages, result
	}

	// Sort by timestamp (should already be sorted, but ensure)
	sorted := make([]A2AMessage, len(messages))
	copy(sorted, messages)

	// Keep recent messages in full
	recentCutoff := len(sorted) - ms.RecentMessagesKeep
	recentMessages := sorted[recentCutoff:]
	olderMessages := sorted[:recentCutoff]

	var processed []A2AMessage

	// Process older messages - summarize them
	for i := range olderMessages {
		msg := &olderMessages[i]

		// Skip if already a summary
		if msg.IsSummary {
			processed = append(processed, *msg)
			continue
		}

		// Create summary
		summary := ms.createSummary(msg)
		originalLength := len(msg.Content)

		// Archive full content if significant
		if originalLength > 1000 {
			archivePath := ms.archiveContent(msg.ID, msg.Content)
			msg.FullContentPath = archivePath
			result.MessagesArchived++
		}

		// Update message with summary
		msg.Content = summary
		msg.IsSummary = true
		msg.OriginalLength = originalLength

		processed = append(processed, *msg)
		result.MessagesSummarized++
		result.TotalSavings += (originalLength - len(summary)) / 4 // Approximate tokens
	}

	// Add recent messages
	processed = append(processed, recentMessages...)
	result.MessagesKept = len(recentMessages)

	return processed, result
}

// createSummary creates a brief summary of a message
func (ms *MessageSummarizer) createSummary(msg *A2AMessage) string {
	content := msg.Content

	// Different summarization based on message type
	switch msg.Type {
	case "task":
		return ms.summarizeTask(content)
	case "response", "task_complete":
		return ms.summarizeResponse(content)
	case "question":
		return ms.summarizeQuestion(content)
	case "decision":
		return ms.summarizeDecision(content)
	default:
		return ms.summarizeGeneric(content)
	}
}

func (ms *MessageSummarizer) summarizeTask(content string) string {
	// Extract task description (first sentence or first 100 chars)
	summary := extractFirstSentence(content, 100)
	return fmt.Sprintf("[Task assigned] %s", summary)
}

func (ms *MessageSummarizer) summarizeResponse(content string) string {
	// Extract key result
	summary := extractFirstSentence(content, 150)
	return fmt.Sprintf("[Response] %s", summary)
}

func (ms *MessageSummarizer) summarizeQuestion(content string) string {
	summary := extractFirstSentence(content, 100)
	return fmt.Sprintf("[Question] %s", summary)
}

func (ms *MessageSummarizer) summarizeDecision(content string) string {
	// Keep decision messages more complete
	summary := extractFirstSentence(content, 200)
	return fmt.Sprintf("[Decision] %s", summary)
}

func (ms *MessageSummarizer) summarizeGeneric(content string) string {
	summary := extractFirstSentence(content, 100)
	return fmt.Sprintf("[Message] %s", summary)
}

// archiveContent saves full content to disk
func (ms *MessageSummarizer) archiveContent(messageID, content string) string {
	filename := fmt.Sprintf("%s_%d.txt", messageID, time.Now().Unix())
	filepath := filepath.Join(ms.ArchiveDir, filename)

	err := fileutil.WriteFileAtomic(filepath, []byte(content), 0o600)
	if err != nil {
		logger.WarnCF("message_summarizer", "Failed to archive message",
			map[string]any{"message_id": messageID, "error": err.Error()})
		return ""
	}

	return filepath
}

// GetFullContent retrieves full content from archive
func (ms *MessageSummarizer) GetFullContent(archivePath string) (string, error) {
	if archivePath == "" {
		return "", fmt.Errorf("no archive path")
	}

	data, err := os.ReadFile(archivePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetSummarizationStats returns current stats
func (ms *MessageSummarizer) GetSummarizationStats() map[string]any {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return map[string]any{
		"messages_tracked": len(ms.messageOrder),
		"summaries_cached": len(ms.summaries),
		"archive_dir":      ms.ArchiveDir,
	}
}

// Helper functions

func extractFirstSentence(content string, maxLen int) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	// Find sentence ending
	sentenceEnders := []string{".\n", ". ", "!", "?", "\n\n"}
	for _, ender := range sentenceEnders {
		if idx := strings.Index(content, ender); idx > 0 && idx < maxLen {
			return strings.TrimSpace(content[:idx+1])
		}
	}

	// Truncate if needed
	if len(content) > maxLen {
		return content[:maxLen] + "..."
	}

	return content
}
