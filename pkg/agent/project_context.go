// Project Context Manager for A2A Token Optimization (Phase 4)
// Provides summarized project context instead of full message history

package agent

import (
	"fmt"
	"strings"
	"time"
)

// ProjectContextManager manages summarized project context
// Reduces token usage by providing condensed context to agents
type ProjectContextManager struct {
	project *A2AProject
}

// NewProjectContextManager creates a new context manager
func NewProjectContextManager(project *A2AProject) *ProjectContextManager {
	return &ProjectContextManager{project: project}
}

// BuildAgentContext creates optimized context for an agent
// Instead of sending all messages, sends summary + relevant context
func (pcm *ProjectContextManager) BuildAgentContext(agentID string, maxTokens int) string {
	pcm.project.mu.RLock()
	defer pcm.project.mu.RUnlock()

	var parts []string

	// 1. Project Summary (always include)
	parts = append(parts, pcm.buildProjectHeader())

	// 2. Current Goals (what we're working on now)
	if len(pcm.project.CurrentGoals) > 0 {
		parts = append(parts, pcm.buildCurrentGoals())
	}

	// 3. Recent relevant messages (filtered by agent)
	recentMessages := pcm.getRecentRelevantMessages(agentID, 5)
	if len(recentMessages) > 0 {
		parts = append(parts, pcm.buildRecentContext(recentMessages))
	}

	// 4. Key Decisions (if relevant to current phase)
	if len(pcm.project.KeyDecisions) > 0 {
		parts = append(parts, pcm.buildKeyDecisions())
	}

	// 5. Current task assignment (if any)
	assignment := pcm.getCurrentAssignment(agentID)
	if assignment != nil {
		parts = append(parts, pcm.buildTaskContext(assignment))
	}

	context := strings.Join(parts, "\n\n")

	// Ensure we don't exceed max tokens (rough estimate)
	if len(context)/4 > maxTokens {
		context = pcm.truncateContext(context, maxTokens)
	}

	return context
}

// buildProjectHeader creates a brief project summary
func (pcm *ProjectContextManager) buildProjectHeader() string {
	p := pcm.project

	duration := time.Since(p.StartTime)
	durationStr := fmt.Sprintf("%dm", int(duration.Minutes()))
	if duration.Hours() >= 1 {
		durationStr = fmt.Sprintf("%dh%dm", int(duration.Hours()), int(duration.Minutes())%60)
	}

	// Count completed tasks
	completedCount := 0
	for _, assignment := range p.Assignments {
		if assignment.Status == AssignmentStatusCompleted {
			completedCount++
		}
	}

	header := fmt.Sprintf(`# Project: %s
Status: %s | Phase: %s | Duration: %s
Progress: %d/%d tasks completed`,
		p.Name,
		p.Status,
		p.CurrentPhase,
		durationStr,
		completedCount,
		len(p.Assignments),
	)

	// Add brief description if available
	if p.ProjectSummary != "" {
		header += fmt.Sprintf("\nSummary: %s", p.ProjectSummary)
	}

	return header
}

// buildCurrentGoals shows what we're working on now
func (pcm *ProjectContextManager) buildCurrentGoals() string {
	var sb strings.Builder
	sb.WriteString("## Current Goals\n")
	for i, goal := range pcm.project.CurrentGoals {
		if i >= 3 { // Limit to 3 goals
			break
		}
		sb.WriteString(fmt.Sprintf("- %s\n", goal))
	}
	return sb.String()
}

// getRecentRelevantMessages gets messages relevant to an agent
func (pcm *ProjectContextManager) getRecentRelevantMessages(agentID string, count int) []A2AMessage {
	var relevant []A2AMessage

	// Get last N messages, filter by relevance
	start := len(pcm.project.Messages) - count
	if start < 0 {
		start = 0
	}

	for i := len(pcm.project.Messages) - 1; i >= start && len(relevant) < count; i-- {
		msg := pcm.project.Messages[i]

		// Include if: to this agent, from this agent, or broadcast
		if msg.To == agentID || msg.From == agentID || msg.To == "" || msg.To == "all" {
			relevant = append([]A2AMessage{msg}, relevant...) // Prepend to maintain order
		}
	}

	return relevant
}

// buildRecentContext formats recent messages
func (pcm *ProjectContextManager) buildRecentContext(messages []A2AMessage) string {
	var sb strings.Builder
	sb.WriteString("## Recent Activity\n")

	for _, msg := range messages {
		content := msg.Content
		// Truncate long messages
		if len(content) > 200 {
			content = content[:200] + "..."
		}

		switch msg.Type {
		case "task":
			sb.WriteString(fmt.Sprintf("- [%s] Assigned task: %s\n", msg.From, content))
		case "task_complete":
			sb.WriteString(fmt.Sprintf("- [%s] Completed: %s\n", msg.From, content))
		case "question":
			sb.WriteString(fmt.Sprintf("- [%s] Asked: %s\n", msg.From, content))
		default:
			sb.WriteString(fmt.Sprintf("- [%s → %s] %s\n", msg.From, msg.To, content))
		}
	}

	return sb.String()
}

// buildKeyDecisions shows important decisions
func (pcm *ProjectContextManager) buildKeyDecisions() string {
	var sb strings.Builder
	sb.WriteString("## Key Decisions\n")

	// Show last 3 decisions
	start := len(pcm.project.KeyDecisions) - 3
	if start < 0 {
		start = 0
	}

	for i := len(pcm.project.KeyDecisions) - 1; i >= start; i-- {
		decision := pcm.project.KeyDecisions[i]
		if len(decision) > 150 {
			decision = decision[:150] + "..."
		}
		sb.WriteString(fmt.Sprintf("- %s\n", decision))
	}

	return sb.String()
}

// getCurrentAssignment finds current task for an agent
func (pcm *ProjectContextManager) getCurrentAssignment(agentID string) *A2AAssignment {
	for i := range pcm.project.Assignments {
		assignment := &pcm.project.Assignments[i]
		if assignment.ToAgent == agentID &&
			(assignment.Status == AssignmentStatusRunning ||
				assignment.Status == AssignmentStatusAssigned) {
			return assignment
		}
	}
	return nil
}

// buildTaskContext creates context for current task
func (pcm *ProjectContextManager) buildTaskContext(assignment *A2AAssignment) string {
	return fmt.Sprintf(`## Your Current Task
**Status:** %s
**Task:** %s
**Progress:** %d%%`,
		assignment.Status,
		assignment.Task,
		assignment.Progress,
	)
}

// truncateContext truncates context to fit token limit
func (pcm *ProjectContextManager) truncateContext(context string, maxTokens int) string {
	// Rough estimate: 4 chars per token
	maxChars := maxTokens * 4

	if len(context) <= maxChars {
		return context
	}

	// Truncate and add notice
	truncated := context[:maxChars-100]
	truncated += "\n\n[Context truncated due to length...]"

	return truncated
}

// UpdateProjectSummary updates the running project summary
func (pcm *ProjectContextManager) UpdateProjectSummary(summary string) {
	pcm.project.mu.Lock()
	defer pcm.project.mu.Unlock()

	pcm.project.ProjectSummary = summary
	pcm.project.ContextVersion++
}

// AddKeyDecision adds a new key decision
func (pcm *ProjectContextManager) AddKeyDecision(decision string) {
	pcm.project.mu.Lock()
	defer pcm.project.mu.Unlock()

	pcm.project.KeyDecisions = append(pcm.project.KeyDecisions, decision)

	// Keep only last 10 decisions
	if len(pcm.project.KeyDecisions) > 10 {
		pcm.project.KeyDecisions = pcm.project.KeyDecisions[len(pcm.project.KeyDecisions)-10:]
	}

	pcm.project.ContextVersion++
}

// SetCurrentGoals updates current project goals
func (pcm *ProjectContextManager) SetCurrentGoals(goals []string) {
	pcm.project.mu.Lock()
	defer pcm.project.mu.Unlock()

	pcm.project.CurrentGoals = goals
	pcm.project.ContextVersion++
}

// GetContextStats returns statistics about context efficiency
func (pcm *ProjectContextManager) GetContextStats() map[string]any {
	pcm.project.mu.RLock()
	defer pcm.project.mu.RUnlock()

	totalMessages := len(pcm.project.Messages)
	summaryLength := len(pcm.project.ProjectSummary)

	// Calculate potential savings
	avgMessageLength := 200 // Estimate
	totalMessageChars := totalMessages * avgMessageLength
	summaryChars := summaryLength + 500 // Summary + context overhead

	savings := 0
	if totalMessageChars > summaryChars {
		savings = totalMessageChars - summaryChars
	}

	return map[string]any{
		"total_messages":       totalMessages,
		"summary_length":       summaryLength,
		"key_decisions":        len(pcm.project.KeyDecisions),
		"current_goals":        len(pcm.project.CurrentGoals),
		"context_version":      pcm.project.ContextVersion,
		"estimated_savings":    savings / 4, // Approximate tokens
		"compression_ratio":    float64(savings) / float64(totalMessageChars),
	}
}
