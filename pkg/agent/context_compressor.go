// Context Compressor for A2A Token Optimization
// Reduces context size by intelligently summarizing older messages

package agent

import (
	"fmt"
	"strings"

	"picoclaw/agent/pkg/providers/protocoltypes"
)

// ContextCompressor provides intelligent context compression for tool loops
type ContextCompressor struct {
	// Configuration
	FullIterationsKeep    int   // Number of recent iterations to keep in full (default: 2)
	MaxContextTokens      int   // Max tokens before compression triggers (default: 8000)
	SummarizeThreshold    int   // Min tokens to trigger summarization (default: 4000)
}

// NewContextCompressor creates a new compressor with default settings
func NewContextCompressor() *ContextCompressor {
	return &ContextCompressor{
		FullIterationsKeep: 2,
		MaxContextTokens:   8000,
		SummarizeThreshold: 4000,
	}
}

// CompressionResult contains compressed messages and metadata
type CompressionResult struct {
	Messages          []protocoltypes.Message
	OriginalTokens    int
	CompressedTokens  int
	IterationsKept    int
	IterationsSummarized int
}

// CompressToolLoop compresses tool loop messages intelligently
// Strategy:
// - Always keep: System prompt + Original task
// - Keep full: Last N iterations (assistant + tool results)
// - Summarize: Older iterations -> key findings only
func (cc *ContextCompressor) CompressToolLoop(
	messages []protocoltypes.Message,
	currentIteration int,
) CompressionResult {
	if len(messages) < 3 {
		// Not enough messages to compress
		return CompressionResult{
			Messages:       messages,
			OriginalTokens: estimateMessageTokens(messages),
		}
	}

	originalTokens := estimateMessageTokens(messages)

	// Identify message roles
	var systemMsg, taskMsg *protocoltypes.Message
	var iterations []IterationBlock

	for i := range messages {
		msg := &messages[i]
		switch msg.Role {
		case "system":
			systemMsg = msg
		case "user":
			if taskMsg == nil {
				taskMsg = msg // First user message is the task
			}
		case "assistant":
			// Start of a new iteration
			iterations = append(iterations, IterationBlock{
				AssistantMsg: msg,
			})
		case "tool":
			// Add to last iteration
			if len(iterations) > 0 {
				iterations[len(iterations)-1].ToolMsgs = append(
					iterations[len(iterations)-1].ToolMsgs,
					msg,
				)
			}
		}
	}

	// Build compressed messages
	var compressed []protocoltypes.Message

	// Always keep system and task
	if systemMsg != nil {
		compressed = append(compressed, *systemMsg)
	}
	if taskMsg != nil {
		compressed = append(compressed, *taskMsg)
	}

	// Decide which iterations to keep vs summarize
	totalIterations := len(iterations)
	iterationsToKeep := cc.FullIterationsKeep
	if iterationsToKeep > totalIterations {
		iterationsToKeep = totalIterations
	}

	// Summarize older iterations (if any)
	if totalIterations > iterationsToKeep {
		olderIterations := iterations[:totalIterations-iterationsToKeep]
		summary := cc.summarizeIterations(olderIterations)
		if summary != "" {
			compressed = append(compressed, protocoltypes.Message{
				Role:    "user",
				Content: fmt.Sprintf("[Previous work summary]: %s", summary),
			})
		}
	}

	// Keep recent iterations in full
	recentIterations := iterations[totalIterations-iterationsToKeep:]
	for _, iter := range recentIterations {
		compressed = append(compressed, *iter.AssistantMsg)
		for _, toolMsg := range iter.ToolMsgs {
			compressed = append(compressed, *toolMsg)
		}
	}

	compressedTokens := estimateMessageTokens(compressed)

	return CompressionResult{
		Messages:             compressed,
		OriginalTokens:       originalTokens,
		CompressedTokens:     compressedTokens,
		IterationsKept:       iterationsToKeep,
		IterationsSummarized: totalIterations - iterationsToKeep,
	}
}

// IterationBlock represents one tool loop iteration
type IterationBlock struct {
	AssistantMsg *protocoltypes.Message
	ToolMsgs     []*protocoltypes.Message
}

// summarizeIterations creates a summary of older iterations
// Focus on: key findings, decisions, failures
func (cc *ContextCompressor) summarizeIterations(iterations []IterationBlock) string {
	var findings []string

	for i, iter := range iterations {
		iterationNum := i + 1

		// Extract tool calls from assistant message
		if iter.AssistantMsg != nil {
			toolNames := extractToolNames(iter.AssistantMsg)
			if len(toolNames) > 0 {
				findings = append(findings, fmt.Sprintf("Step %d: Used %v", iterationNum, toolNames))
			}
		}

		// Extract key results from tool messages
		for _, toolMsg := range iter.ToolMsgs {
			result := extractKeyResult(toolMsg.Content)
			if result != "" {
				findings = append(findings, result)
			}
		}
	}

	if len(findings) == 0 {
		return ""
	}

	// Limit summary length
	summary := strings.Join(findings, "; ")
	maxLen := 500
	if len(summary) > maxLen {
		summary = summary[:maxLen] + "..."
	}

	return summary
}

// ShouldCompress determines if compression is needed
func (cc *ContextCompressor) ShouldCompress(messages []protocoltypes.Message) bool {
	tokens := estimateMessageTokens(messages)
	return tokens > cc.SummarizeThreshold
}

// GetCompressionStats returns compression statistics
func (cc *ContextCompressor) GetCompressionStats(result CompressionResult) string {
	if result.OriginalTokens == 0 {
		return "No compression applied"
	}

	savings := result.OriginalTokens - result.CompressedTokens
	percent := float64(savings) * 100 / float64(result.OriginalTokens)

	return fmt.Sprintf(
		"Compressed: %d -> %d tokens (%.1f%% savings). Kept %d iterations, summarized %d.",
		result.OriginalTokens,
		result.CompressedTokens,
		percent,
		result.IterationsKept,
		result.IterationsSummarized,
	)
}

// Helper functions

func estimateMessageTokens(messages []protocoltypes.Message) int {
	total := 0
	for _, msg := range messages {
		total += EstimateTokens(msg.Content)
		// Also count tool calls
		for _, tc := range msg.ToolCalls {
			total += EstimateTokens(tc.Name)
			if tc.Function != nil {
				total += EstimateTokens(tc.Function.Arguments)
			}
		}
	}
	return total
}

func extractToolNames(msg *protocoltypes.Message) []string {
	var names []string
	for _, tc := range msg.ToolCalls {
		if tc.Name != "" {
			names = append(names, tc.Name)
		}
	}
	return names
}

func extractKeyResult(content string) string {
	// Extract first sentence or first 100 chars
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	// Try to find first sentence
	if idx := strings.IndexAny(content, ".!?\n"); idx > 0 && idx < 100 {
		return content[:idx+1]
	}

	// Otherwise truncate
	if len(content) > 100 {
		return content[:100] + "..."
	}
	return content
}
