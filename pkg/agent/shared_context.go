package agent

import (
	"picoclaw/agent/pkg/agentcomm"
)

// SharedContext provides a thread-safe shared memory layer for agents.
// It wraps the agentcomm.SharedContext for internal use.
type SharedContext struct {
	inner *agentcomm.SharedContext
}

// NewSharedContext creates a new SharedContext with configurable limits.
func NewSharedContext(maxLogSize, maxContext int) *SharedContext {
	return &SharedContext{
		inner: agentcomm.NewSharedContext(maxLogSize, maxContext),
	}
}

// Set sets a key-value pair in the shared context.
func (sc *SharedContext) Set(key string, value any) {
	sc.inner.Set(key, value)
}

// Get retrieves a value from the shared context.
func (sc *SharedContext) Get(key string) (any, bool) {
	return sc.inner.Get(key)
}

// GetAll returns a copy of the entire context.
func (sc *SharedContext) GetAll() map[string]any {
	return sc.inner.GetAll()
}

// Delete removes a key from the shared context.
func (sc *SharedContext) Delete(key string) {
	sc.inner.Delete(key)
}

// Clear removes all entries from the shared context.
func (sc *SharedContext) Clear() {
	sc.inner.Clear()
}

// AddMessageLog adds a message to the shared message log.
func (sc *SharedContext) AddMessageLog(from, to, msgType, content string) {
	sc.inner.AddMessageLog(from, to, msgType, content)
}

// GetMessageLog returns the message log entries.
func (sc *SharedContext) GetMessageLog() []agentcomm.MessageLogEntry {
	return sc.inner.GetMessageLog()
}

// GetMessageLogSince returns message log entries since the given timestamp.
func (sc *SharedContext) GetMessageLogSince(since int64) []agentcomm.MessageLogEntry {
	return sc.inner.GetMessageLogSince(since)
}

// GetMessagesForAgent returns messages for a specific agent.
func (sc *SharedContext) GetMessagesForAgent(agentID string) []agentcomm.MessageLogEntry {
	return sc.inner.GetMessagesForAgent(agentID)
}

// ContextSize returns the number of entries in the context.
func (sc *SharedContext) ContextSize() int {
	return sc.inner.ContextSize()
}

// LogSize returns the number of entries in the message log.
func (sc *SharedContext) LogSize() int {
	return sc.inner.LogSize()
}

// Keys returns all context keys.
func (sc *SharedContext) Keys() []string {
	return sc.inner.Keys()
}
