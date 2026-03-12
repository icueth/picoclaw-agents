// Package mailbox provides inter-agent communication via message queue
package mailbox

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Message types
const (
	MessageTypeTask      = "task"      // Task assignment
	MessageTypeStatus    = "status"    // Status update
	MessageTypeQuestion  = "question"  // Ask for information
	MessageTypeAnswer    = "answer"    // Response to question
	MessageTypeMeeting   = "meeting"   // Meeting invitation
	MessageTypeBroadcast = "broadcast" // Broadcast to all
)

// Priority levels
const (
	PriorityCritical = 1
	PriorityHigh     = 2
	PriorityNormal   = 3
	PriorityLow      = 4
)

// Message represents a message between agents
type Message struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	From        string                 `json:"from"`
	To          string                 `json:"to"`           // "all" for broadcast
	Priority    int                    `json:"priority"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Read        bool                   `json:"read"`
	Delivered   bool                   `json:"delivered"`
	DeliveredAt *time.Time             `json:"delivered_at,omitempty"`
	ReplyTo     string                 `json:"reply_to,omitempty"` // Reference to original message
}

// Mailbox manages message queue for an agent
type Mailbox struct {
	agentID      string
	capacity     int
	messages     []Message
	unreadCount  int
	mu           sync.RWMutex
	subscribers  []chan Message
	subMu        sync.RWMutex
}

// NewMailbox creates a new mailbox for an agent
func NewMailbox(agentID string, capacity int) *Mailbox {
	if capacity <= 0 {
		capacity = 1000
	}
	return &Mailbox{
		agentID:     agentID,
		capacity:    capacity,
		messages:    make([]Message, 0, capacity),
		subscribers: make([]chan Message, 0),
	}
}

// Send adds a message to the mailbox
func (m *Mailbox) Send(msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.messages) >= m.capacity {
		// Remove oldest non-critical message
		removed := false
		for i := len(m.messages) - 1; i >= 0; i-- {
			if m.messages[i].Priority > PriorityHigh {
				// Check unread before removing
				if !m.messages[i].Read {
					m.unreadCount--
				}
				m.messages = append(m.messages[:i], m.messages[i+1:]...)
				removed = true
				break
			}
		}
		if !removed {
			return fmt.Errorf("mailbox full, cannot add message")
		}
	}

	msg.Delivered = true
	now := time.Now()
	msg.DeliveredAt = &now

	// Insert by priority (lower number = higher priority)
	insertIdx := len(m.messages)
	for i, existing := range m.messages {
		if msg.Priority < existing.Priority {
			insertIdx = i
			break
		}
	}

	m.messages = append(m.messages, Message{})
	copy(m.messages[insertIdx+1:], m.messages[insertIdx:])
	m.messages[insertIdx] = msg
	m.unreadCount++

	// Notify subscribers
	m.subMu.RLock()
	for _, ch := range m.subscribers {
		select {
		case ch <- msg:
		default:
		}
	}
	m.subMu.RUnlock()

	return nil
}

// Receive gets next unread message (highest priority first)
func (m *Mailbox) Receive() (*Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.messages {
		if !m.messages[i].Read {
			m.messages[i].Read = true
			m.unreadCount--
			return &m.messages[i], nil
		}
	}
	return nil, fmt.Errorf("no unread messages")
}

// Peek returns next unread message without marking as read
func (m *Mailbox) Peek() (*Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := range m.messages {
		if !m.messages[i].Read {
			return &m.messages[i], nil
		}
	}
	return nil, fmt.Errorf("no unread messages")
}

// Subscribe returns a channel for real-time message notifications
func (m *Mailbox) Subscribe() <-chan Message {
	m.subMu.Lock()
	defer m.subMu.Unlock()

	ch := make(chan Message, 10)
	m.subscribers = append(m.subscribers, ch)
	return ch
}

// Unsubscribe removes a subscriber
func (m *Mailbox) Unsubscribe(ch <-chan Message) {
	m.subMu.Lock()
	defer m.subMu.Unlock()

	for i, sub := range m.subscribers {
		if sub == ch {
			close(sub)
			m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
			break
		}
	}
}

// GetMessages returns messages with optional filters
func (m *Mailbox) GetMessages(unreadOnly bool, msgType string, limit int) []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Message, 0)
	count := 0

	for i := len(m.messages) - 1; i >= 0 && count < limit; i-- {
		msg := m.messages[i]

		if unreadOnly && msg.Read {
			continue
		}
		if msgType != "" && msg.Type != msgType {
			continue
		}

		result = append(result, msg)
		count++
	}

	return result
}

// GetUnreadCount returns number of unread messages
func (m *Mailbox) GetUnreadCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.unreadCount
}

// MarkAsRead marks specific message as read
func (m *Mailbox) MarkAsRead(msgID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.messages {
		if m.messages[i].ID == msgID && !m.messages[i].Read {
			m.messages[i].Read = true
			m.unreadCount--
			return true
		}
	}
	return false
}

// Clear removes all messages
func (m *Mailbox) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = make([]Message, 0, m.capacity)
	m.unreadCount = 0
}

// Size returns total message count
func (m *Mailbox) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.messages)
}

// CleanupExpired removes expired messages
func (m *Mailbox) CleanupExpired() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	removed := 0
	newMessages := make([]Message, 0, m.capacity)

	for _, msg := range m.messages {
		if msg.ExpiresAt != nil && msg.ExpiresAt.Before(now) {
			if !msg.Read {
				m.unreadCount--
			}
			removed++
			continue
		}
		newMessages = append(newMessages, msg)
	}

	m.messages = newMessages
	return removed
}

// ToJSON serializes mailbox state
func (m *Mailbox) ToJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return json.Marshal(map[string]interface{}{
		"agent_id":     m.agentID,
		"size":         len(m.messages),
		"unread_count": m.unreadCount,
		"capacity":     m.capacity,
		"messages":     m.messages,
	})
}

// Hub manages all agent mailboxes
type Hub struct {
	mailboxes map[string]*Mailbox
	mu        sync.RWMutex
	capacity  int
}

// NewHub creates a new mailbox hub
func NewHub(mailboxCapacity int) *Hub {
	return &Hub{
		mailboxes: make(map[string]*Mailbox),
		capacity:  mailboxCapacity,
	}
}

// Register creates mailbox for an agent
func (h *Hub) Register(agentID string) *Mailbox {
	h.mu.Lock()
	defer h.mu.Unlock()

	if mb, exists := h.mailboxes[agentID]; exists {
		return mb
	}

	mb := NewMailbox(agentID, h.capacity)
	h.mailboxes[agentID] = mb
	return mb
}

// Get retrieves mailbox for an agent
func (h *Hub) Get(agentID string) (*Mailbox, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	mb, ok := h.mailboxes[agentID]
	return mb, ok
}

// Unregister removes an agent's mailbox
func (h *Hub) Unregister(agentID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.mailboxes, agentID)
}

// SendTo sends message to specific agent
func (h *Hub) SendTo(agentID string, msg Message) error {
	mb, ok := h.Get(agentID)
	if !ok {
		return fmt.Errorf("agent %s not registered", agentID)
	}
	return mb.Send(msg)
}

// Broadcast sends message to all agents
func (h *Hub) Broadcast(msg Message) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	failed := make([]string, 0)
	for id, mb := range h.mailboxes {
		if id == msg.From {
			continue
		}
		if err := mb.Send(msg); err != nil {
			failed = append(failed, id)
		}
	}
	return failed
}

// GetAllMailboxes returns all registered mailboxes
func (h *Hub) GetAllMailboxes() map[string]*Mailbox {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]*Mailbox)
	for id, mb := range h.mailboxes {
		result[id] = mb
	}
	return result
}

// GetStats returns hub statistics
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := map[string]interface{}{
		"total_agents": len(h.mailboxes),
		"agents":       make(map[string]interface{}),
	}

	agentStats := stats["agents"].(map[string]interface{})
	for id, mb := range h.mailboxes {
		agentStats[id] = map[string]interface{}{
			"size":         mb.Size(),
			"unread_count": mb.GetUnreadCount(),
		}
	}

	return stats
}

// StartCleanup starts periodic cleanup of expired messages
func (h *Hub) StartCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				h.mu.RLock()
				mailboxes := make([]*Mailbox, 0, len(h.mailboxes))
				for _, mb := range h.mailboxes {
					mailboxes = append(mailboxes, mb)
				}
				h.mu.RUnlock()

				for _, mb := range mailboxes {
					mb.CleanupExpired()
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// GenerateMessageID creates a unique message ID
func GenerateMessageID(from string) string {
	return fmt.Sprintf("%s-%d", from, time.Now().UnixNano())
}
