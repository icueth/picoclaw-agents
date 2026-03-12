// Package meeting provides multi-agent conference and collaboration capabilities
package meeting

import (
	"context"
	"fmt"
	"sync"
	"time"

	"picoclaw/agent/pkg/agentcomm"
)

// MeetingStatus represents the current state of a meeting
type MeetingStatus string

const (
	MeetingStatusPending   MeetingStatus = "pending"   // Waiting to start
	MeetingStatusOngoing   MeetingStatus = "ongoing"   // In progress
	MeetingStatusPaused    MeetingStatus = "paused"    // Temporarily paused
	MeetingStatusCompleted MeetingStatus = "completed" // Finished
	MeetingStatusCancelled MeetingStatus = "cancelled" // Cancelled
)

// MeetingRole defines the role of a participant in a meeting
type MeetingRole string

const (
	MeetingRoleFacilitator MeetingRole = "facilitator" // ดำเนินการประชุม
	MeetingRoleParticipant MeetingRole = "participant" // ผู้เข้าร่วม
	MeetingRoleObserver    MeetingRole = "observer"    // สังเกตการณ์
)

// MeetingMessage represents a message in a meeting
type MeetingMessage struct {
	ID          string    `json:"id"`
	MeetingID   string    `json:"meeting_id"`
	FromAgent   string    `json:"from_agent"`
	Content     string    `json:"content"`
	Type        string    `json:"type"` // "statement", "question", "proposal", "vote", "summary"
	Timestamp   time.Time `json:"timestamp"`
	ReplyTo     string    `json:"reply_to,omitempty"` // ID of message being replied to
	Mentions    []string  `json:"mentions,omitempty"` // Agents mentioned
	Consensus   *bool     `json:"consensus,omitempty"` // For voting
}

// MeetingParticipant represents an agent participating in a meeting
type MeetingParticipant struct {
	AgentID   string      `json:"agent_id"`
	Name      string      `json:"name"`
	Avatar    string      `json:"avatar"`
	Role      MeetingRole `json:"role"`
	JoinedAt  time.Time   `json:"joined_at"`
	LastSeen  time.Time   `json:"last_seen"`
	IsOnline  bool        `json:"is_online"`
	Speaking  bool        `json:"speaking"` // Currently speaking
}

// Meeting represents a multi-agent conference
type Meeting struct {
	ID           string               `json:"id"`
	Topic        string               `json:"topic"`
	Description  string               `json:"description"`
	Status       MeetingStatus        `json:"status"`
	CreatedAt    time.Time            `json:"created_at"`
	StartedAt    *time.Time           `json:"started_at,omitempty"`
	EndedAt      *time.Time           `json:"ended_at,omitempty"`
	Facilitator  string               `json:"facilitator"` // Agent ID
	Participants map[string]*MeetingParticipant `json:"participants"`
	Messages     []MeetingMessage     `json:"messages"`
	Agenda       []AgendaItem         `json:"agenda"`
	CurrentItem  int                  `json:"current_item"` // Index of current agenda item
	Consensus    *bool                `json:"consensus,omitempty"` // Final decision
	Summary      string               `json:"summary,omitempty"`   // Meeting summary

	// Internal
	mu        sync.RWMutex
	msgChan   chan MeetingMessage
	ctx       context.Context
	cancel    context.CancelFunc
	onMessage func(MeetingMessage)
}

// AgendaItem represents an item on the meeting agenda
type AgendaItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"` // "pending", "discussing", "completed"
	AssignedTo  string `json:"assigned_to,omitempty"` // Agent responsible
}

// MeetingConfig provides configuration for creating a meeting
type MeetingConfig struct {
	Topic       string
	Description string
	Facilitator string              // Agent ID (default: jarvis)
	Agenda      []string            // List of agenda items
	Context     map[string]any      // Shared context/information
	Timeout     time.Duration       // Meeting timeout (default: 30 min)
}

// NewMeeting creates a new meeting instance
func NewMeeting(config MeetingConfig) *Meeting {
	ctx, cancel := context.WithCancel(context.Background())
	
	if config.Facilitator == "" {
		config.Facilitator = "jarvis" // Default facilitator
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Minute
	}

	meeting := &Meeting{
		ID:           generateMeetingID(),
		Topic:        config.Topic,
		Description:  config.Description,
		Status:       MeetingStatusPending,
		CreatedAt:    time.Now(),
		Facilitator:  config.Facilitator,
		Participants: make(map[string]*MeetingParticipant),
		Messages:     make([]MeetingMessage, 0),
		Agenda:       make([]AgendaItem, 0),
		CurrentItem:  0,
		ctx:          ctx,
		cancel:       cancel,
		msgChan:      make(chan MeetingMessage, 100),
	}

	// Convert agenda strings to items
	for i, title := range config.Agenda {
		meeting.Agenda = append(meeting.Agenda, AgendaItem{
			ID:     fmt.Sprintf("agenda-%d", i+1),
			Title:  title,
			Status: "pending",
		})
	}

	// Start message processing goroutine
	go meeting.processMessages()

	return meeting
}

// AddParticipant adds an agent to the meeting
func (m *Meeting) AddParticipant(agentID, name, avatar string, role MeetingRole) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Status != MeetingStatusPending && m.Status != MeetingStatusOngoing {
		return fmt.Errorf("cannot join meeting: status is %s", m.Status)
	}

	if _, exists := m.Participants[agentID]; exists {
		return fmt.Errorf("agent %s already in meeting", agentID)
	}

	m.Participants[agentID] = &MeetingParticipant{
		AgentID:  agentID,
		Name:     name,
		Avatar:   avatar,
		Role:     role,
		JoinedAt: time.Now(),
		LastSeen: time.Now(),
		IsOnline: true,
	}

	return nil
}

// RemoveParticipant removes an agent from the meeting
func (m *Meeting) RemoveParticipant(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if participant, exists := m.Participants[agentID]; exists {
		participant.IsOnline = false
		// Keep in list but mark offline
	}
}

// PostMessage adds a message to the meeting
func (m *Meeting) PostMessage(fromAgent, content, msgType string, mentions []string) (*MeetingMessage, error) {
	m.mu.RLock()
	if _, exists := m.Participants[fromAgent]; !exists {
		m.mu.RUnlock()
		return nil, fmt.Errorf("agent %s not in meeting", fromAgent)
	}
	if m.Status != MeetingStatusOngoing {
		m.mu.RUnlock()
		return nil, fmt.Errorf("meeting not ongoing")
	}
	m.mu.RUnlock()

	msg := MeetingMessage{
		ID:        generateMessageID(),
		MeetingID: m.ID,
		FromAgent: fromAgent,
		Content:   content,
		Type:      msgType,
		Timestamp: time.Now(),
		Mentions:  mentions,
	}

	select {
	case m.msgChan <- msg:
		return &msg, nil
	case <-time.After(time.Second):
		return nil, fmt.Errorf("message queue full")
	}
}

// Start begins the meeting
func (m *Meeting) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Status != MeetingStatusPending {
		return fmt.Errorf("meeting already started or ended")
	}

	now := time.Now()
	m.Status = MeetingStatusOngoing
	m.StartedAt = &now

	return nil
}

// End finishes the meeting
func (m *Meeting) End(summary string, consensus *bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.Status = MeetingStatusCompleted
	m.EndedAt = &now
	m.Summary = summary
	m.Consensus = consensus
	
	m.cancel() // Stop message processing
}

// GetMessages returns all messages since a given time
func (m *Meeting) GetMessages(since time.Time) []MeetingMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []MeetingMessage
	for _, msg := range m.Messages {
		if msg.Timestamp.After(since) {
			result = append(result, msg)
		}
	}
	return result
}

// GetParticipantList returns all participants
func (m *Meeting) GetParticipantList() []*MeetingParticipant {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*MeetingParticipant, 0, len(m.Participants))
	for _, p := range m.Participants {
		result = append(result, p)
	}
	return result
}

// SetOnMessage sets a callback for new messages
func (m *Meeting) SetOnMessage(fn func(MeetingMessage)) {
	m.onMessage = fn
}

// Internal message processing
func (m *Meeting) processMessages() {
	for {
		select {
		case msg := <-m.msgChan:
			m.mu.Lock()
			m.Messages = append(m.Messages, msg)
			// Update participant last seen
			if p, exists := m.Participants[msg.FromAgent]; exists {
				p.LastSeen = time.Now()
			}
			m.mu.Unlock()

			if m.onMessage != nil {
				m.onMessage(msg)
			}

		case <-m.ctx.Done():
			return
		}
	}
}

// Helper functions
func generateMeetingID() string {
	return fmt.Sprintf("meeting-%d", time.Now().UnixNano())
}

func generateMessageID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}

// ToAgentMessage converts MeetingMessage to AgentMessage for agent communication
func (m *MeetingMessage) ToAgentMessage() agentcomm.AgentMessage {
	return agentcomm.NewAgentMessage(
		m.FromAgent,
		"", // Broadcast to all
		agentcomm.MsgBroadcast,
		m.Content,
		m.MeetingID,
	)
}
