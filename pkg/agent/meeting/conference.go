// Conference provides high-level meeting management for agent teams
package meeting

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/config"
)

// ConferenceManager manages multiple meetings and agent participation
type ConferenceManager struct {
	mu       sync.RWMutex
	meetings map[string]*Meeting
	agents   map[string]*AgentInfo // Available agents

	// Active discussions
	discussions map[string]*AgentDiscussion

	// Callbacks
	onMeetingCreated func(*Meeting)
	onMeetingEnded   func(*Meeting)
	onAgentSpeak     func(agentID, content string)
}

// AgentInfo holds information about an available agent
type AgentInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Avatar       string   `json:"avatar"`
	Role         string   `json:"role"`
	Department   string   `json:"department"`
	Capabilities []string `json:"capabilities"`
	IsAvailable  bool     `json:"is_available"`
}

// AgentDiscussion represents an active discussion between agents
type AgentDiscussion struct {
	ID          string            `json:"id"`
	Topic       string            `json:"topic"`
	Context     string            `json:"context"`      // Background information
	Participants []string         `json:"participants"` // Agent IDs
	Messages    []DiscussionTurn  `json:"messages"`
	Status      string            `json:"status"` // "ongoing", "completed"
	StartedAt   time.Time         `json:"started_at"`
	Consensus   string            `json:"consensus,omitempty"`
}

// DiscussionTurn represents one turn in a discussion
type DiscussionTurn struct {
	AgentID   string    `json:"agent_id"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "thought", "proposal", "question", "agreement", "objection"
}

// NewConferenceManager creates a new conference manager
func NewConferenceManager() *ConferenceManager {
	return &ConferenceManager{
		meetings:    make(map[string]*Meeting),
		agents:      make(map[string]*AgentInfo),
		discussions: make(map[string]*AgentDiscussion),
	}
}

// RegisterAgent adds an agent to the conference system
func (cm *ConferenceManager) RegisterAgent(info *AgentInfo) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.agents[info.ID] = info
}

// GetAgent returns information about a registered agent
func (cm *ConferenceManager) GetAgent(agentID string) (*AgentInfo, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	agent, ok := cm.agents[agentID]
	return agent, ok
}

// GetAgentsByDepartment returns all agents in a department
func (cm *ConferenceManager) GetAgentsByDepartment(dept string) []*AgentInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var result []*AgentInfo
	for _, agent := range cm.agents {
		if agent.Department == dept {
			result = append(result, agent)
		}
	}
	return result
}

// CreateMeeting creates a new meeting with the given participants
func (cm *ConferenceManager) CreateMeeting(config MeetingConfig, participantIDs []string) (*Meeting, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate facilitator
	if config.Facilitator == "" {
		config.Facilitator = "jarvis"
	}
	if _, ok := cm.agents[config.Facilitator]; !ok {
		return nil, fmt.Errorf("facilitator %s not found", config.Facilitator)
	}

	// Create meeting
	meeting := NewMeeting(config)

	// Add facilitator
	facilitator := cm.agents[config.Facilitator]
	meeting.AddParticipant(facilitator.ID, facilitator.Name, facilitator.Avatar, MeetingRoleFacilitator)

	// Add participants
	for _, agentID := range participantIDs {
		if agentID == config.Facilitator {
			continue // Skip facilitator (already added)
		}
		if agent, ok := cm.agents[agentID]; ok {
			meeting.AddParticipant(agent.ID, agent.Name, agent.Avatar, MeetingRoleParticipant)
		}
	}

	cm.meetings[meeting.ID] = meeting

	if cm.onMeetingCreated != nil {
		cm.onMeetingCreated(meeting)
	}

	return meeting, nil
}

// GetMeeting returns a meeting by ID
func (cm *ConferenceManager) GetMeeting(meetingID string) (*Meeting, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	meeting, ok := cm.meetings[meetingID]
	return meeting, ok
}

// ListMeetings returns all meetings (optionally filtered by status)
func (cm *ConferenceManager) ListMeetings(status ...MeetingStatus) []*Meeting {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var result []*Meeting
	for _, meeting := range cm.meetings {
		if len(status) == 0 {
			result = append(result, meeting)
		} else {
			for _, s := range status {
				if meeting.Status == s {
					result = append(result, meeting)
					break
				}
			}
		}
	}

	// Sort by created time (newest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result
}

// StartAgentDiscussion initiates a discussion between multiple agents
func (cm *ConferenceManager) StartAgentDiscussion(topic, context string, participantIDs []string) (*AgentDiscussion, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate all participants exist
	for _, id := range participantIDs {
		if _, ok := cm.agents[id]; !ok {
			return nil, fmt.Errorf("agent %s not found", id)
		}
	}

	discussion := &AgentDiscussion{
		ID:           generateDiscussionID(),
		Topic:        topic,
		Context:      context,
		Participants: participantIDs,
		Messages:     make([]DiscussionTurn, 0),
		Status:       "ongoing",
		StartedAt:    time.Now(),
	}

	cm.discussions[discussion.ID] = discussion

	return discussion, nil
}

// AddDiscussionTurn adds a turn to a discussion
func (cm *ConferenceManager) AddDiscussionTurn(discussionID, agentID, content, turnType string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	discussion, ok := cm.discussions[discussionID]
	if !ok {
		return fmt.Errorf("discussion not found")
	}

	agent, ok := cm.agents[agentID]
	if !ok {
		return fmt.Errorf("agent not found")
	}

	turn := DiscussionTurn{
		AgentID:   agentID,
		Name:      agent.Name,
		Avatar:    agent.Avatar,
		Content:   content,
		Type:      turnType,
		Timestamp: time.Now(),
	}

	discussion.Messages = append(discussion.Messages, turn)

	if cm.onAgentSpeak != nil {
		cm.onAgentSpeak(agentID, content)
	}

	return nil
}

// CompleteDiscussion marks a discussion as completed with consensus
func (cm *ConferenceManager) CompleteDiscussion(discussionID, consensus string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	discussion, ok := cm.discussions[discussionID]
	if !ok {
		return fmt.Errorf("discussion not found")
	}

	discussion.Status = "completed"
	discussion.Consensus = consensus

	return nil
}

// GetDiscussion returns a discussion by ID
func (cm *ConferenceManager) GetDiscussion(discussionID string) (*AgentDiscussion, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	discussion, ok := cm.discussions[discussionID]
	return discussion, ok
}

// FacilitateMeeting simulates a facilitated meeting where Jarvis coordinates
func (cm *ConferenceManager) FacilitateMeeting(ctx context.Context, topic string, agenda []string, agentIDs []string) (*Meeting, error) {
	// Create meeting with Jarvis as facilitator
	config := MeetingConfig{
		Topic:       topic,
		Description: fmt.Sprintf("Facilitated discussion on: %s", topic),
		Facilitator: "jarvis",
		Agenda:      agenda,
		Timeout:     30 * time.Minute,
	}

	meeting, err := cm.CreateMeeting(config, agentIDs)
	if err != nil {
		return nil, err
	}

	// Start the meeting
	if err := meeting.Start(); err != nil {
		return nil, err
	}

	// Post welcome message from Jarvis
	welcomeMsg := fmt.Sprintf("👋 สวัสดีทุกคน! ยินดีต้อนรับเข้าสู่การประชุมเรื่อง **%s**\n\n", topic)
	welcomeMsg += "📋 **วาระการประชุม:**\n"
	for i, item := range agenda {
		welcomeMsg += fmt.Sprintf("%d. %s\n", i+1, item)
	}
	welcomeMsg += "\n💡 เรามาช่วยกันคิดและตัดสินใจกันนะครับ"

	meeting.PostMessage("jarvis", welcomeMsg, "statement", nil)

	return meeting, nil
}

// GenerateMeetingSummary creates a summary of the meeting
func (cm *ConferenceManager) GenerateMeetingSummary(meetingID string) (string, error) {
	meeting, ok := cm.GetMeeting(meetingID)
	if !ok {
		return "", fmt.Errorf("meeting not found")
	}

	meeting.mu.RLock()
	defer meeting.mu.RUnlock()

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("# สรุปการประชุม: %s\n\n", meeting.Topic))
	summary.WriteString(fmt.Sprintf("**รหัสการประชุม:** %s\n", meeting.ID))
	summary.WriteString(fmt.Sprintf("**วันที่:** %s\n", meeting.CreatedAt.Format("2006-01-02 15:04")))
	summary.WriteString(fmt.Sprintf("**ผู้ดำเนินการ:** %s\n\n", meeting.Facilitator))

	// Participants
	summary.WriteString("## ผู้เข้าร่วม\n")
	for _, p := range meeting.Participants {
		role := ""
		if p.Role == MeetingRoleFacilitator {
			role = " (ดำเนินการ)"
		}
		summary.WriteString(fmt.Sprintf("- %s %s%s\n", p.Avatar, p.Name, role))
	}
	summary.WriteString("\n")

	// Key discussion points
	summary.WriteString("## ประเด็นที่หารือ\n")
	for _, agenda := range meeting.Agenda {
		summary.WriteString(fmt.Sprintf("- %s\n", agenda.Title))
	}
	summary.WriteString("\n")

	// Messages by type
	proposals := []MeetingMessage{}
	votes := []MeetingMessage{}
	for _, msg := range meeting.Messages {
		if msg.Type == "proposal" {
			proposals = append(proposals, msg)
		} else if msg.Type == "vote" {
			votes = append(votes, msg)
		}
	}

	if len(proposals) > 0 {
		summary.WriteString("## ข้อเสนอที่หารือ\n")
		for _, p := range proposals {
			summary.WriteString(fmt.Sprintf("- **%s**: %s\n", p.FromAgent, p.Content))
		}
		summary.WriteString("\n")
	}

	// Consensus
	if meeting.Consensus != nil {
		if *meeting.Consensus {
			summary.WriteString("## ✅ มติที่ประชุม\nเห็นชอบตามข้อเสนอ\n")
		} else {
			summary.WriteString("## ❌ มติที่ประชุม\nไม่เห็นชอบ\n")
		}
	}

	if meeting.Summary != "" {
		summary.WriteString("\n## สรุปผล\n")
		summary.WriteString(meeting.Summary)
	}

	return summary.String(), nil
}

// SetCallbacks sets event callbacks
func (cm *ConferenceManager) SetCallbacks(
	onMeetingCreated func(*Meeting),
	onMeetingEnded func(*Meeting),
	onAgentSpeak func(agentID, content string),
) {
	cm.onMeetingCreated = onMeetingCreated
	cm.onMeetingEnded = onMeetingEnded
	cm.onAgentSpeak = onAgentSpeak
}

// Helper function
func generateDiscussionID() string {
	return fmt.Sprintf("discussion-%d", time.Now().UnixNano())
}

// PopulateFromRegistry registers all agents from config (since registry uses AgentInstance)
func (cm *ConferenceManager) PopulateFromRegistry(agents []config.AgentConfig) {
	for _, ac := range agents {
		cm.RegisterAgent(&AgentInfo{
			ID:           ac.ID,
			Name:         ac.Name,
			Avatar:       ac.Avatar,
			Role:         ac.Role,
			Department:   ac.Department,
			Capabilities: ac.Capabilities,
			IsAvailable:  true,
		})
	}
}
