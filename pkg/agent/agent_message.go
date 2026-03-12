package agent

import (
	"time"

	"picoclaw/agent/pkg/agentcomm"
)

// Re-export types from agentcomm for convenience
type MessageType = agentcomm.MessageType
type AgentMessage = agentcomm.AgentMessage
type AgentInfo = agentcomm.AgentInfo
type AgentStatus = agentcomm.AgentStatus

const (
	MsgRequest       = agentcomm.MsgRequest
	MsgResponse      = agentcomm.MsgResponse
	MsgBroadcast     = agentcomm.MsgBroadcast
	MsgContextUpdate = agentcomm.MsgContextUpdate
	MsgHeartbeat     = agentcomm.MsgHeartbeat
	MsgTerminate     = agentcomm.MsgTerminate

	AgentStatusIdle      = agentcomm.AgentStatusIdle
	AgentStatusRunning   = agentcomm.AgentStatusRunning
	AgentStatusWaiting   = agentcomm.AgentStatusWaiting
	AgentStatusCompleted = agentcomm.AgentStatusCompleted
	AgentStatusFailed    = agentcomm.AgentStatusFailed
)

// NewAgentMessage creates a new AgentMessage with a generated ID and timestamp.
func NewAgentMessage(from, to string, msgType agentcomm.MessageType, payload any, sessionID string) agentcomm.AgentMessage {
	return agentcomm.NewAgentMessage(from, to, msgType, payload, sessionID)
}

// NewRequestMessage creates a new request message.
func NewRequestMessage(from, to, sessionID, content string) agentcomm.AgentMessage {
	return agentcomm.NewAgentMessage(from, to, agentcomm.MsgRequest, content, sessionID)
}

// NewResponseMessage creates a new response message.
func NewResponseMessage(from, to, sessionID, replyToID, content string) agentcomm.AgentMessage {
	msg := agentcomm.NewAgentMessage(from, to, agentcomm.MsgResponse, content, sessionID)
	msg.ReplyTo = replyToID
	return msg
}

// NewBroadcastMessage creates a new broadcast message.
func NewBroadcastMessage(from, sessionID, content string) agentcomm.AgentMessage {
	return agentcomm.NewAgentMessage(from, "", agentcomm.MsgBroadcast, content, sessionID)
}

// MessagePayload contains structured payload for agent messages.
type MessagePayload struct {
	Content    string         `json:"content,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
	ContextKey string         `json:"context_key,omitempty"`
	ContextVal any            `json:"context_val,omitempty"`
	Iterations int            `json:"iterations,omitempty"`
	Success    bool           `json:"success,omitempty"`
	Error      string         `json:"error,omitempty"`
}

// NewMessagePayload creates a new MessagePayload with content.
func NewMessagePayload(content string) MessagePayload {
	return MessagePayload{Content: content}
}

// WithData adds data to the payload.
func (p MessagePayload) WithData(key string, value any) MessagePayload {
	if p.Data == nil {
		p.Data = make(map[string]any)
	}
	p.Data[key] = value
	return p
}

// WithContext adds context key-value to the payload.
func (p MessagePayload) WithContext(key string, value any) MessagePayload {
	p.ContextKey = key
	p.ContextVal = value
	return p
}

// WithIterations adds iteration count to the payload.
func (p MessagePayload) WithIterations(iterations int) MessagePayload {
	p.Iterations = iterations
	return p
}

// WithSuccess adds success status to the payload.
func (p MessagePayload) WithSuccess(success bool) MessagePayload {
	p.Success = success
	return p
}

// WithError adds error message to the payload.
func (p MessagePayload) WithError(err string) MessagePayload {
	p.Error = err
	return p
}

// generateMessageID generates a unique message ID.
func generateMessageID(sender string) string {
	return sender + "-" + time.Now().Format("20060102150405.000")
}
