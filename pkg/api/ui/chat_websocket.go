package ui

import (
	"encoding/json"
	"time"

	"picoclaw/agent/pkg/logger"
)

// ChatEventType represents chat-specific event types
type ChatEventType string

const (
	ChatEventMessage       ChatEventType = "chat_message"
	ChatEventTypingStart   ChatEventType = "typing_start"
	ChatEventTypingStop    ChatEventType = "typing_stop"
	ChatEventSessionJoined ChatEventType = "session_joined"
	ChatEventSessionLeft   ChatEventType = "session_left"
	ChatEventAgentStatus   ChatEventType = "agent_status"
	ChatEventSettingsSaved ChatEventType = "settings_saved"
)

// ChatMessage represents a chat message
type ChatMessage struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Sender    string                 `json:"sender"` // "user", "agent", "system"
	AgentID   string                 `json:"agent_id,omitempty"`
	AgentName string                 `json:"agent_name,omitempty"`
	SessionID string                 `json:"session_id"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ChatSession represents a chat session
type ChatSession struct {
	ID           string    `json:"id"`
	Mode         string    `json:"mode"` // "main", "direct", "meeting"
	Name         string    `json:"name"`
	AgentID      string    `json:"agent_id,omitempty"`
	Participants []string  `json:"participants"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ChatTypingPayload represents typing indicator payload
type ChatTypingPayload struct {
	AgentID   string    `json:"agent_id"`
	SessionID string    `json:"session_id"`
	IsTyping  bool      `json:"is_typing"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatAgentStatusPayload represents agent status in chat
type ChatAgentStatusPayload struct {
	AgentID     string    `json:"agent_id"`
	Status      string    `json:"status"` // "online", "away", "offline"
	CurrentTask string    `json:"current_task,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// ChatSettingsPayload represents settings update
type ChatSettingsPayload struct {
	AgentID   string                 `json:"agent_id"`
	Settings  map[string]interface{} `json:"settings"`
	Timestamp time.Time              `json:"timestamp"`
}

// handleChatMessage processes incoming chat messages from clients
func (c *Client) handleChatMessage(data []byte) {
	var msg ChatMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		c.sendError("invalid_chat_message", "Failed to parse chat message")
		return
	}

	// Validate message
	if msg.Content == "" || msg.SessionID == "" {
		c.sendError("invalid_chat_message", "Content and session_id are required")
		return
	}

	// Set timestamp and ID if not provided
	if msg.ID == "" {
		msg.ID = generateMessageID()
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	msg.Sender = "user" // Always from user for now

	logger.DebugF("ui_chat", map[string]any{
		"action":     "chat_message_received",
		"session_id": msg.SessionID,
		"client_id":  c.id,
	})

	// Broadcast to all clients in the same session
	c.hub.broadcastChatEvent(&ChatEvent{
		Type:      ChatEventMessage,
		Timestamp: time.Now(),
		Payload:   msg,
	})

	// TODO: Route to appropriate agent via mailbox system
	// This would integrate with pkg/agents/coordinator.go
	c.routeToAgent(msg)
}

// ChatEvent wraps chat events for broadcasting
type ChatEvent struct {
	Type      ChatEventType `json:"type"`
	Timestamp time.Time     `json:"timestamp"`
	Payload   interface{}   `json:"payload"`
}

// broadcastChatEvent broadcasts a chat event to subscribed clients
func (h *Hub) broadcastChatEvent(event *ChatEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(event)
	if err != nil {
		logger.DebugF("ui_chat", map[string]any{
			"error": err.Error(),
		})
		return
	}

	for client := range h.clients {
		// For chat events, broadcast to all clients (or filter by session)
		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// BroadcastChatMessage broadcasts a chat message from an agent
func (h *Hub) BroadcastChatMessage(msg ChatMessage) {
	h.broadcastChatEvent(&ChatEvent{
		Type:      ChatEventMessage,
		Timestamp: time.Now(),
		Payload:   msg,
	})
}

// BroadcastAgentTyping broadcasts typing indicator
func (h *Hub) BroadcastAgentTyping(agentID, sessionID string, isTyping bool) {
	h.broadcastChatEvent(&ChatEvent{
		Type:    ChatEventTypingStart,
		Payload: ChatTypingPayload{
			AgentID:   agentID,
			SessionID: sessionID,
			IsTyping:  isTyping,
			Timestamp: time.Now(),
		},
	})
}

// BroadcastChatAgentStatus broadcasts agent status change
func (h *Hub) BroadcastChatAgentStatus(agentID, status, currentTask string) {
	h.broadcastChatEvent(&ChatEvent{
		Type:    ChatEventAgentStatus,
		Payload: ChatAgentStatusPayload{
			AgentID:     agentID,
			Status:      status,
			CurrentTask: currentTask,
			Timestamp:   time.Now(),
		},
	})
}

// routeToAgent routes a user message to the appropriate agent
func (c *Client) routeToAgent(msg ChatMessage) {
	// This is where we integrate with the agent system
	// For now, just log and echo back a system message
	
	logger.DebugF("ui_chat", map[string]any{
		"action":     "route_to_agent",
		"session_id": msg.SessionID,
		"content":    msg.Content,
	})

	// Send acknowledgment
	ack := ChatMessage{
		ID:        generateMessageID(),
		Content:   "Message received. Routing to agent...",
		Sender:    "system",
		SessionID: msg.SessionID,
		Timestamp: time.Now(),
		Type:      "system_notification",
	}

	c.hub.BroadcastChatMessage(ack)

	// TODO: Integrate with:
	// 1. pkg/agents/coordinator.go - to analyze and route task
	// 2. pkg/mailbox/mailbox.go - to send message to agent's mailbox
	// 3. Agent LLM processing - to generate response
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// Extend handleMessage to support chat events
func (c *Client) handleChatEvent(msgType string, payload []byte) {
	switch msgType {
	case "chat_message":
		c.handleChatMessage(payload)
	case "typing_start", "typing_stop":
		var typingData struct {
			SessionID string `json:"session_id"`
			IsTyping  bool   `json:"is_typing"`
		}
		if err := json.Unmarshal(payload, &typingData); err == nil {
			// Broadcast typing indicator
			eventType := ChatEventTypingStart
			if !typingData.IsTyping {
				eventType = ChatEventTypingStop
			}
			c.hub.broadcastChatEvent(&ChatEvent{
				Type: eventType,
				Payload: ChatTypingPayload{
					SessionID: typingData.SessionID,
					IsTyping:  typingData.IsTyping,
					Timestamp: time.Now(),
				},
			})
		}
	case "session_joined":
		var sessionData struct {
			SessionID string `json:"session_id"`
			Mode      string `json:"mode"`
			AgentID   string `json:"agent_id,omitempty"`
		}
		if err := json.Unmarshal(payload, &sessionData); err == nil {
			c.hub.broadcastChatEvent(&ChatEvent{
				Type: ChatEventSessionJoined,
				Payload: map[string]interface{}{
					"session_id": sessionData.SessionID,
					"mode":       sessionData.Mode,
					"agent_id":   sessionData.AgentID,
				},
			})
		}
	default:
		logger.DebugF("ui_chat", map[string]any{
			"message": "Unknown chat event type",
			"type":    msgType,
		})
	}
}

// Add to Hub initialization
func (h *Hub) StartChatHandlers() {
	// Start any background goroutines for chat processing
	// This could include:
	// - Agent response processing loop
	// - Message persistence
	// - Session cleanup
	
	logger.DebugF("ui_chat", map[string]any{
		"message": "Chat handlers started",
	})
}
