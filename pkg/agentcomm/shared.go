package agentcomm

import (
	"sync"
	"time"
)

// MessageLogEntry represents a single message in the shared log.
type MessageLogEntry struct {
	From      string
	To        string
	Type      string
	Content   string
	Timestamp int64
}

// SharedContext provides a thread-safe shared memory layer for agents.
type SharedContext struct {
	mu         sync.RWMutex
	context    map[string]any
	messageLog []MessageLogEntry
	maxLogSize int
	maxContext int
}

// NewSharedContext creates a new SharedContext with configurable limits.
func NewSharedContext(maxLogSize, maxContext int) *SharedContext {
	if maxLogSize <= 0 {
		maxLogSize = 100
	}
	if maxContext <= 0 {
		maxContext = 1000
	}
	return &SharedContext{
		context:    make(map[string]any),
		messageLog: make([]MessageLogEntry, 0, maxLogSize),
		maxLogSize: maxLogSize,
		maxContext: maxContext,
	}
}

// Set sets a key-value pair in the shared context.
func (sc *SharedContext) Set(key string, value any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if len(sc.context) >= sc.maxContext {
		for k := range sc.context {
			delete(sc.context, k)
			break
		}
	}
	sc.context[key] = value
}

// Get retrieves a value from the shared context.
func (sc *SharedContext) Get(key string) (any, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	val, ok := sc.context[key]
	return val, ok
}

// GetAll returns a copy of the entire context.
func (sc *SharedContext) GetAll() map[string]any {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	result := make(map[string]any, len(sc.context))
	for k, v := range sc.context {
		result[k] = v
	}
	return result
}

// Delete removes a key from the shared context.
func (sc *SharedContext) Delete(key string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	delete(sc.context, key)
}

// Clear removes all entries from the shared context.
func (sc *SharedContext) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.context = make(map[string]any)
	sc.messageLog = make([]MessageLogEntry, 0, sc.maxLogSize)
}

// AddMessageLog adds a message to the shared message log.
func (sc *SharedContext) AddMessageLog(from, to, msgType, content string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	entry := MessageLogEntry{
		From:      from,
		To:        to,
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now().UnixMilli(),
	}

	if len(sc.messageLog) >= sc.maxLogSize {
		sc.messageLog = sc.messageLog[1:]
	}

	sc.messageLog = append(sc.messageLog, entry)
}

// GetMessageLog returns the message log entries.
func (sc *SharedContext) GetMessageLog() []MessageLogEntry {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	result := make([]MessageLogEntry, len(sc.messageLog))
	copy(result, sc.messageLog)
	return result
}

// GetMessageLogSince returns message log entries since the given timestamp.
func (sc *SharedContext) GetMessageLogSince(since int64) []MessageLogEntry {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var result []MessageLogEntry
	for _, entry := range sc.messageLog {
		if entry.Timestamp > since {
			result = append(result, entry)
		}
	}
	return result
}

// GetMessagesForAgent returns messages for a specific agent.
func (sc *SharedContext) GetMessagesForAgent(agentID string) []MessageLogEntry {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var result []MessageLogEntry
	for _, entry := range sc.messageLog {
		if entry.From == agentID || entry.To == agentID || entry.To == "" {
			result = append(result, entry)
		}
	}
	return result
}

// ContextSize returns the number of entries in the context.
func (sc *SharedContext) ContextSize() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return len(sc.context)
}

// LogSize returns the number of entries in the message log.
func (sc *SharedContext) LogSize() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return len(sc.messageLog)
}

// Keys returns all context keys.
func (sc *SharedContext) Keys() []string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	keys := make([]string, 0, len(sc.context))
	for k := range sc.context {
		keys = append(keys, k)
	}
	return keys
}

// MessageType defines the type of agent-to-agent message.
type MessageType string

const (
	MsgRequest       MessageType = "request"
	MsgResponse      MessageType = "response"
	MsgBroadcast     MessageType = "broadcast"
	MsgContextUpdate MessageType = "context_update"
	MsgHeartbeat     MessageType = "heartbeat"
	MsgTerminate     MessageType = "terminate"
)

// AgentMessage represents a message sent between agents.
type AgentMessage struct {
	From      string      // sender agent ID
	To        string      // target agent ID ("" = broadcast)
	Type      MessageType // message type
	Payload   any         // message content
	SessionID string      // session identifier
	Timestamp int64       // Unix timestamp in milliseconds
	ReplyTo   string      // message ID being replied to
	ID        string      // unique message ID
}

// NewAgentMessage creates a new AgentMessage.
func NewAgentMessage(from, to string, msgType MessageType, payload any, sessionID string) AgentMessage {
	return AgentMessage{
		From:      from,
		To:        to,
		Type:      msgType,
		Payload:   payload,
		SessionID: sessionID,
		Timestamp: time.Now().UnixMilli(),
		ID:        from + "-" + time.Now().Format("20060102150405.000"),
	}
}

// IsBroadcast checks if the message is a broadcast.
func (m *AgentMessage) IsBroadcast() bool {
	return m.Type == MsgBroadcast || m.To == ""
}

// GetPayloadString returns the payload as a string.
func (m *AgentMessage) GetPayloadString() string {
	if m.Payload == nil {
		return ""
	}
	if s, ok := m.Payload.(string); ok {
		return s
	}
	return ""
}

// AgentStatus represents the status of an agent.
type AgentStatus string

const (
	AgentStatusIdle      AgentStatus = "idle"
	AgentStatusRunning   AgentStatus = "running"
	AgentStatusWaiting   AgentStatus = "waiting"
	AgentStatusCompleted AgentStatus = "completed"
	AgentStatusFailed    AgentStatus = "failed"
)

// AgentInfo represents information about an agent.
type AgentInfo struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	ParentID     string      `json:"parent_id,omitempty"`
	Model        string      `json:"model,omitempty"`
	Status       AgentStatus `json:"status"`
	CreatedAt    int64       `json:"created_at"`
	LastActive   int64       `json:"last_active"`
	Capabilities []string    `json:"capabilities,omitempty"`
}
