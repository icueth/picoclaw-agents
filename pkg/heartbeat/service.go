// PicoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package heartbeat

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/constants"
	"picoclaw/agent/pkg/fileutil"
	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/state"
	"picoclaw/agent/pkg/tools"
)

const (
	minIntervalMinutes     = 5
	defaultIntervalMinutes = 30
	startupDelay           = 30 * time.Second // Wait for services to initialize
	execTimeout            = 2 * time.Minute  // Max time for a single heartbeat execution
	maxLogBytes            = 256 * 1024       // 256 KB log rotation threshold
	maxConsecutiveErrors   = 3                // After this many errors, apply backoff
)

// HeartbeatHandler is the function type for handling heartbeat.
// It returns a ToolResult that can indicate async operations.
// channel and chatID are derived from the last active user channel.
type HeartbeatHandler func(prompt, channel, chatID string) *tools.ToolResult

// HeartbeatService manages periodic heartbeat checks
type HeartbeatService struct {
	workspace string
	bus       *bus.MessageBus
	state     *state.Manager
	handler   HeartbeatHandler
	interval  time.Duration
	enabled   bool
	mu        sync.RWMutex
	stopChan  chan struct{}

	// Concurrency guard — prevents overlapping heartbeats
	executing atomic.Bool

	// Error backoff — increases interval after consecutive failures
	consecutiveErrors atomic.Int32
}

// NewHeartbeatService creates a new heartbeat service
func NewHeartbeatService(workspace string, intervalMinutes int, enabled bool) *HeartbeatService {
	// Apply minimum interval
	if intervalMinutes < minIntervalMinutes && intervalMinutes != 0 {
		intervalMinutes = minIntervalMinutes
	}

	if intervalMinutes == 0 {
		intervalMinutes = defaultIntervalMinutes
	}

	return &HeartbeatService{
		workspace: workspace,
		interval:  time.Duration(intervalMinutes) * time.Minute,
		enabled:   enabled,
		state:     state.NewManager(workspace),
	}
}

// SetBus sets the message bus for delivering heartbeat results.
func (hs *HeartbeatService) SetBus(msgBus *bus.MessageBus) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.bus = msgBus
}

// SetHandler sets the heartbeat handler.
func (hs *HeartbeatService) SetHandler(handler HeartbeatHandler) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.handler = handler
}

// Start begins the heartbeat service
func (hs *HeartbeatService) Start() error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if hs.stopChan != nil {
		logger.InfoC("heartbeat", "Heartbeat service already running")
		return nil
	}

	if !hs.enabled {
		logger.InfoC("heartbeat", "Heartbeat service disabled")
		return nil
	}

	hs.stopChan = make(chan struct{})
	go hs.runLoop(hs.stopChan)

	logger.InfoCF("heartbeat", "Heartbeat service started", map[string]any{
		"interval_minutes": hs.interval.Minutes(),
	})

	return nil
}

// Stop gracefully stops the heartbeat service
func (hs *HeartbeatService) Stop() {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if hs.stopChan == nil {
		return
	}

	logger.InfoC("heartbeat", "Stopping heartbeat service")
	close(hs.stopChan)
	hs.stopChan = nil
}

// IsRunning returns whether the service is running
func (hs *HeartbeatService) IsRunning() bool {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.stopChan != nil
}

// runLoop runs the heartbeat ticker
func (hs *HeartbeatService) runLoop(stopChan chan struct{}) {
	// Wait for services to fully initialize before first heartbeat
	select {
	case <-stopChan:
		return
	case <-time.After(startupDelay):
	}

	// Run first heartbeat immediately after startup delay
	hs.executeHeartbeat()

	ticker := time.NewTicker(hs.interval)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			hs.executeHeartbeat()
		}
	}
}

// executeHeartbeat performs a single heartbeat check with concurrency guard,
// execution timeout, and error backoff.
func (hs *HeartbeatService) executeHeartbeat() {
	// Concurrency guard — skip if previous heartbeat is still running
	if !hs.executing.CompareAndSwap(false, true) {
		logger.InfoC("heartbeat", "Skipping heartbeat — previous execution still running")
		return
	}
	defer hs.executing.Store(false)

	hs.mu.RLock()
	enabled := hs.enabled
	handler := hs.handler
	if !hs.enabled || hs.stopChan == nil {
		hs.mu.RUnlock()
		return
	}
	hs.mu.RUnlock()

	if !enabled {
		return
	}

	// Error backoff — if too many consecutive errors, skip this cycle
	errCount := hs.consecutiveErrors.Load()
	if errCount >= maxConsecutiveErrors {
		// Apply exponential backoff: skip 2^(errCount - maxConsecutiveErrors) cycles
		skipCycles := int32(1) << (errCount - maxConsecutiveErrors)
		if skipCycles > 8 {
			skipCycles = 8 // Cap at 8x interval (~4 hours at 30-min interval)
		}
		// Decrement to eventually retry
		hs.consecutiveErrors.Add(-1)
		logger.InfoCF("heartbeat", "Backing off due to consecutive errors", map[string]any{
			"consecutive_errors": errCount,
			"skip_remaining":     skipCycles - 1,
		})
		return
	}

	// Rotate log if needed
	hs.rotateLogIfNeeded()

	logger.DebugC("heartbeat", "Executing heartbeat")

	prompt := hs.buildPrompt()
	if prompt == "" {
		logger.DebugC("heartbeat", "No heartbeat tasks — skipping LLM call")
		return
	}

	if handler == nil {
		hs.logErrorf("Heartbeat handler not configured")
		return
	}

	// Get last channel info for context
	lastChannel := hs.state.GetLastChannel()
	channel, chatID := hs.parseLastChannel(lastChannel)

	hs.logInfof("Resolved channel: %s, chatID: %s (from lastChannel: %s)", channel, chatID, lastChannel)

	// Execute with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	type handlerResult struct {
		result *tools.ToolResult
	}
	resultCh := make(chan handlerResult, 1)

	go func() {
		r := handler(prompt, channel, chatID)
		resultCh <- handlerResult{result: r}
	}()

	var result *tools.ToolResult
	select {
	case <-ctx.Done():
		hs.consecutiveErrors.Add(1)
		hs.logErrorf("Heartbeat execution timed out after %v", execTimeout)
		return
	case hr := <-resultCh:
		result = hr.result
	}

	if result == nil {
		hs.logInfof("Heartbeat handler returned nil result")
		return
	}

	// Handle different result types
	if result.IsError {
		hs.consecutiveErrors.Add(1)
		hs.logErrorf("Heartbeat error: %s", result.ForLLM)
		return
	}

	// Success — reset error counter
	hs.consecutiveErrors.Store(0)

	if result.Async {
		hs.logInfof("Async task started: %s", result.ForLLM)
		logger.InfoCF("heartbeat", "Async heartbeat task started",
			map[string]any{
				"message": result.ForLLM,
			})
		return
	}

	// Check if silent
	if result.Silent {
		hs.logInfof("Heartbeat OK - silent")
		return
	}

	// Send result to user
	if result.ForUser != "" {
		hs.sendResponse(result.ForUser)
	} else if result.ForLLM != "" {
		hs.sendResponse(result.ForLLM)
	}

	hs.logInfof("Heartbeat completed: %s", result.ForLLM)
}

// rotateLogIfNeeded checks the heartbeat log file size and rotates if needed
func (hs *HeartbeatService) rotateLogIfNeeded() {
	logPath := filepath.Join(hs.workspace, "heartbeat.log")
	info, err := os.Stat(logPath)
	if err != nil {
		return
	}
	if info.Size() < maxLogBytes {
		return
	}

	// Read current log, keep last half
	data, err := os.ReadFile(logPath)
	if err != nil {
		return
	}
	half := len(data) / 2
	// Find next newline after halfway point to avoid splitting a line
	for i := half; i < len(data); i++ {
		if data[i] == '\n' {
			half = i + 1
			break
		}
	}
	if err := os.WriteFile(logPath, data[half:], 0o644); err != nil {
		logger.WarnCF("heartbeat", "Failed to rotate heartbeat log", map[string]any{"error": err.Error()})
	} else {
		logger.DebugCF("heartbeat", "Rotated heartbeat log", map[string]any{
			"old_size": info.Size(),
			"new_size": int64(len(data) - half),
		})
	}
}

// buildPrompt builds the heartbeat prompt from HEARTBEAT.md.
// Returns empty string if no actual user tasks are found (saves LLM calls).
func (hs *HeartbeatService) buildPrompt() string {
	heartbeatPath := filepath.Join(hs.workspace, "HEARTBEAT.md")

	data, err := os.ReadFile(heartbeatPath)
	if err != nil {
		if os.IsNotExist(err) {
			hs.createDefaultHeartbeatTemplate()
			return ""
		}
		hs.logErrorf("Error reading HEARTBEAT.md: %v", err)
		return ""
	}

	content := string(data)
	if len(content) == 0 {
		return ""
	}

	// Extract only user tasks (content below the "---" separator).
	// This avoids sending template/instructions to LLM, saving tokens.
	userTasks := extractUserTasks(content)
	if userTasks == "" {
		return "" // No actual tasks — skip LLM call entirely
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf(`# Heartbeat Check

Current time: %s

You are a proactive AI assistant. This is a scheduled heartbeat check.
Review the following tasks and execute any necessary actions using available skills.
- For simple tasks (e.g., report current time), respond directly.
- For complex tasks, use the spawn tool to create a subagent.
- After spawning, CONTINUE to process remaining tasks.
If there is nothing that requires attention, respond ONLY with: HEARTBEAT_OK

## Tasks
%s
`, now, userTasks)
}

// extractUserTasks extracts user-defined tasks from HEARTBEAT.md.
// Tasks are expected below a "---" horizontal rule separator.
// Returns empty string if no tasks are found.
func extractUserTasks(content string) string {
	// Find the last "---" separator
	sepIdx := strings.LastIndex(content, "---")
	if sepIdx < 0 {
		// No separator — treat entire content as tasks (backward compat)
		tasks := strings.TrimSpace(content)
		if tasks == "" {
			return ""
		}
		return tasks
	}

	// Extract content after the separator
	after := content[sepIdx+3:]
	tasks := strings.TrimSpace(after)

	// Check if it's just the template placeholder or truly empty
	if tasks == "" || tasks == "Add your heartbeat tasks below this line:" {
		return ""
	}

	// Filter out lines that are only whitespace or the placeholder
	lines := strings.Split(tasks, "\n")
	var realTasks []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "Add your heartbeat tasks below this line:" {
			continue
		}
		realTasks = append(realTasks, line)
	}

	if len(realTasks) == 0 {
		return ""
	}

	return strings.Join(realTasks, "\n")
}

// createDefaultHeartbeatTemplate creates the default HEARTBEAT.md file
func (hs *HeartbeatService) createDefaultHeartbeatTemplate() {
	heartbeatPath := filepath.Join(hs.workspace, "HEARTBEAT.md")

	defaultContent := `# Heartbeat Check List

This file contains tasks for the heartbeat service to check periodically.

## Examples

- Check for unread messages
- Review upcoming calendar events
- Check device status (e.g., MaixCam)

## Instructions

- Execute ALL tasks listed below. Do NOT skip any task.
- For simple tasks (e.g., report current time), respond directly.
- For complex tasks that may take time, use the spawn tool to create a subagent.
- The spawn tool is async - subagent results will be sent to the user automatically.
- After spawning a subagent, CONTINUE to process remaining tasks.
- Only respond with HEARTBEAT_OK when ALL tasks are done AND nothing needs attention.

---

Add your heartbeat tasks below this line:
`

	if err := fileutil.WriteFileAtomic(heartbeatPath, []byte(defaultContent), 0o644); err != nil {
		hs.logErrorf("Failed to create default HEARTBEAT.md: %v", err)
	} else {
		hs.logInfof("Created default HEARTBEAT.md template")
	}
}

// sendResponse sends the heartbeat response to the last channel
func (hs *HeartbeatService) sendResponse(response string) {
	hs.mu.RLock()
	msgBus := hs.bus
	hs.mu.RUnlock()

	if msgBus == nil {
		hs.logInfof("No message bus configured, heartbeat result not sent")
		return
	}

	// Get last channel from state
	lastChannel := hs.state.GetLastChannel()
	if lastChannel == "" {
		hs.logInfof("No last channel recorded, heartbeat result not sent")
		return
	}

	platform, userID := hs.parseLastChannel(lastChannel)

	// Skip internal channels that can't receive messages
	if platform == "" || userID == "" {
		return
	}

	pubCtx, pubCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pubCancel()
	msgBus.PublishOutbound(pubCtx, bus.OutboundMessage{
		Channel: platform,
		ChatID:  userID,
		Content: response,
	})

	hs.logInfof("Heartbeat result sent to %s", platform)
}

// parseLastChannel parses the last channel string into platform and userID.
// Returns empty strings for invalid or internal channels.
func (hs *HeartbeatService) parseLastChannel(lastChannel string) (platform, userID string) {
	if lastChannel == "" {
		return "", ""
	}

	// Parse channel format: "platform:user_id" (e.g., "telegram:123456")
	parts := strings.SplitN(lastChannel, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		hs.logErrorf("Invalid last channel format: %s", lastChannel)
		return "", ""
	}

	platform, userID = parts[0], parts[1]

	// Skip internal channels
	if constants.IsInternalChannel(platform) {
		hs.logInfof("Skipping internal channel: %s", platform)
		return "", ""
	}

	return platform, userID
}

// logInfof logs an informational message to the heartbeat log
func (hs *HeartbeatService) logInfof(format string, args ...any) {
	hs.logf("INFO", format, args...)
}

// logErrorf logs an error message to the heartbeat log
func (hs *HeartbeatService) logErrorf(format string, args ...any) {
	hs.logf("ERROR", format, args...)
}

// logf writes a message to the heartbeat log file
func (hs *HeartbeatService) logf(level, format string, args ...any) {
	logFile := filepath.Join(hs.workspace, "heartbeat.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "[%s] [%s] %s\n", timestamp, level, fmt.Sprintf(format, args...))
}
