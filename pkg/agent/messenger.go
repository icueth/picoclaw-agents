package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"picoclaw/agent/pkg/agentcomm"
	"picoclaw/agent/pkg/bus"
)

// Messenger handles message routing between agents using both central pub/sub
// and direct P2P delivery.
type Messenger struct {
	mu           sync.RWMutex
	agentID      string
	sharedCtx    *SharedContext
	msgBus       *bus.MessageBus
	subscribers  map[string][]MessageHandler
	directQueue  chan *agentcomm.AgentMessage
	handlers     map[string]MessageHandler
	agentRegistry map[string]*agentcomm.AgentInfo
	closed       bool
}

// MessageHandler is a function that handles incoming agent messages.
type MessageHandler func(ctx context.Context, msg *agentcomm.AgentMessage)

// NewMessenger creates a new Messenger with the given components.
func NewMessenger(agentID string, sharedCtx *SharedContext, msgBus *bus.MessageBus) *Messenger {
	m := &Messenger{
		agentID:       agentID,
		sharedCtx:     sharedCtx,
		msgBus:        msgBus,
		subscribers:   make(map[string][]MessageHandler),
		directQueue:   make(chan *agentcomm.AgentMessage, 100),
		handlers:      make(map[string]MessageHandler),
		agentRegistry: make(map[string]*agentcomm.AgentInfo),
	}

	// Start message processing loop
	go m.processMessages()

	return m
}

// SetBus sets the message bus for central pub/sub.
func (m *Messenger) SetBus(msgBus *bus.MessageBus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.msgBus = msgBus
}

// SetSharedContext sets the shared context.
func (m *Messenger) SetSharedContext(ctx *SharedContext) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sharedCtx = ctx
}

// Publish sends a message through the central message bus.
func (m *Messenger) Publish(ctx context.Context, msg agentcomm.AgentMessage) error {
	// Log to shared context
	if m.sharedCtx != nil {
		m.sharedCtx.AddMessageLog(msg.From, msg.To, string(msg.Type), msg.GetPayloadString())
	}

	m.mu.RLock()
	msgBus := m.msgBus
	m.mu.RUnlock()

	if msgBus == nil {
		return fmt.Errorf("message bus not configured")
	}

	// Convert AgentMessage to bus message
	content, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish to agent-specific channel
	channel := fmt.Sprintf("agent:%s", msg.To)
	if msg.To == "" {
		channel = "agent:broadcast"
	}

	busMsg := bus.InboundMessage{
		Channel:    channel,
		SenderID:   msg.From,
		ChatID:     msg.SessionID,
		Content:    string(content),
		SessionKey: msg.SessionID,
	}

	return msgBus.PublishInbound(ctx, busMsg)
}

// SendDirect sends a message directly to a specific agent (P2P).
func (m *Messenger) SendDirect(ctx context.Context, to string, msg agentcomm.AgentMessage) error {
	msg.To = to

	// Log to shared context
	if m.sharedCtx != nil {
		m.sharedCtx.AddMessageLog(msg.From, msg.To, string(msg.Type), msg.GetPayloadString())
	}

	// Try direct delivery first
	if handler := m.getHandler(to); handler != nil {
		select {
		case m.directQueue <- &msg:
			return nil
		default:
			// Queue full, try processing immediately
			handler(ctx, &msg)
			return nil
		}
	}

	// Fall back to bus if direct handler not found
	return m.Publish(ctx, msg)
}

// Broadcast sends a message to all agents.
func (m *Messenger) Broadcast(ctx context.Context, msg agentcomm.AgentMessage) error {
	msg.To = ""

	// Log to shared context
	if m.sharedCtx != nil {
		m.sharedCtx.AddMessageLog(msg.From, msg.To, string(msg.Type), msg.GetPayloadString())
	}

	// Notify all subscribers
	m.mu.RLock()
	var wg sync.WaitGroup
	for pattern, handlers := range m.subscribers {
		if matchesPattern(pattern, msg.To) || pattern == "*" {
			for _, handler := range handlers {
				wg.Add(1)
				go func(h MessageHandler) {
					defer wg.Done()
					h(ctx, &msg)
				}(handler)
			}
		}
	}
	m.mu.RUnlock()

	wg.Wait()

	return nil
}

// Subscribe registers a handler for messages matching the given pattern.
func (m *Messenger) Subscribe(pattern string, handler MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.subscribers[pattern] = append(m.subscribers[pattern], handler)
}

// RegisterHandler registers a direct handler for a specific agent ID.
func (m *Messenger) RegisterHandler(agentID string, handler MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers[agentID] = handler
}

// UnregisterHandler removes the handler for a specific agent.
func (m *Messenger) UnregisterHandler(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.handlers, agentID)
}

// getHandler retrieves the handler for a specific agent.
func (m *Messenger) getHandler(agentID string) MessageHandler {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.handlers[agentID]
}

// RegisterAgent registers an agent in the registry.
func (m *Messenger) RegisterAgent(info *agentcomm.AgentInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.agentRegistry[info.ID] = info
}

// UnregisterAgent removes an agent from the registry.
func (m *Messenger) UnregisterAgent(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.agentRegistry, agentID)
}

// GetAgent retrieves agent info from the registry.
func (m *Messenger) GetAgent(agentID string) (*agentcomm.AgentInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, ok := m.agentRegistry[agentID]
	return info, ok
}

// ListAgents returns all registered agents.
func (m *Messenger) ListAgents() []*agentcomm.AgentInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*agentcomm.AgentInfo, 0, len(m.agentRegistry))
	for _, info := range m.agentRegistry {
		agents = append(agents, info)
	}
	return agents
}

// SendContextUpdate broadcasts a context update to all interested agents.
func (m *Messenger) SendContextUpdate(ctx context.Context, key string, value any) error {
	msg := agentcomm.AgentMessage{
		From:   m.agentID,
		To:     "",
		Type:   agentcomm.MsgContextUpdate,
		Payload: MessagePayload{ContextKey: key, ContextVal: value},
	}
	return m.Broadcast(ctx, msg)
}

// ReadSharedContext reads from the shared context.
func (m *Messenger) ReadSharedContext(key string) (any, bool) {
	if m.sharedCtx == nil {
		return nil, false
	}
	return m.sharedCtx.Get(key)
}

// WriteSharedContext writes to the shared context.
func (m *Messenger) WriteSharedContext(key string, value any) {
	if m.sharedCtx == nil {
		return
	}
	m.sharedCtx.Set(key, value)
}

// ReadAllSharedContext reads all context.
func (m *Messenger) ReadAllSharedContext() map[string]any {
	if m.sharedCtx == nil {
		return nil
	}
	return m.sharedCtx.GetAll()
}

// GetMessageLog returns the message log from shared context.
func (m *Messenger) GetMessageLog() []agentcomm.MessageLogEntry {
	if m.sharedCtx == nil {
		return nil
	}
	return m.sharedCtx.GetMessageLog()
}

// processMessages handles incoming messages from the direct queue.
func (m *Messenger) processMessages() {
	for msg := range m.directQueue {
		m.mu.RLock()
		handler := m.handlers[m.agentID]
		m.mu.RUnlock()

		if handler != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
			handler(ctx, msg)
			cancel()
		}
	}
}

// Close shuts down the messenger.
func (m *Messenger) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return
	}

	m.closed = true
	close(m.directQueue)
}

// matchesPattern checks if a target matches a subscription pattern.
func matchesPattern(pattern, target string) bool {
	if pattern == "*" {
		return true
	}
	return pattern == target
}
