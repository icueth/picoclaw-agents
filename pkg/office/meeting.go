// Package office provides Office UI functionality for Picoclaw agent management
package office

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MeetingStatus represents the current status of a meeting
type MeetingStatus string

const (
	MeetingStatusScheduled MeetingStatus = "scheduled"
	MeetingStatusActive    MeetingStatus = "active"
	MeetingStatusPaused    MeetingStatus = "paused"
	MeetingStatusCompleted MeetingStatus = "completed"
	MeetingStatusCancelled MeetingStatus = "cancelled"
	MeetingStatusSkipped   MeetingStatus = "skipped"
)

// MeetingType represents the type of meeting
type MeetingType string

const (
	MeetingTypeStandup         MeetingType = "standup"
	MeetingTypePlanning        MeetingType = "planning"
	MeetingTypeReview          MeetingType = "review"
	MeetingTypeRetrospective   MeetingType = "retrospective"
	MeetingTypeOneOnOne        MeetingType = "one_on_one"
	MeetingTypeSprint          MeetingType = "sprint"
	MeetingTypeEmergency       MeetingType = "emergency"
	MeetingTypeDemo            MeetingType = "demo"
	MeetingTypeBrainstorm      MeetingType = "brainstorm"
)

// Meeting represents a scheduled meeting between agents
type Meeting struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description,omitempty"`
	Type         MeetingType       `json:"type"`
	Status       MeetingStatus     `json:"status"`
	ScheduledAt  time.Time         `json:"scheduled_at"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	EndedAt      *time.Time        `json:"ended_at,omitempty"`
	Duration     time.Duration     `json:"duration"`
	OrganizerID  string            `json:"organizer_id"`
	Participants []MeetingParticipant `json:"participants"`
	Agenda       []AgendaItem      `json:"agenda,omitempty"`
	Minutes      *MeetingMinutes   `json:"minutes,omitempty"`
	Location     string            `json:"location,omitempty"`
	IsRecurring  bool              `json:"is_recurring"`
	Recurrence   *RecurrenceRule   `json:"recurrence,omitempty"`
	ParentID     string            `json:"parent_id,omitempty"`
	ProjectID    string            `json:"project_id,omitempty"`
	Priority     CEOPriority       `json:"priority"`
	Tags         []string          `json:"tags,omitempty"`
	Metadata     map[string]any    `json:"metadata,omitempty"`

	// Internal tracking
	createdAt time.Time
	updatedAt time.Time
}

// MeetingParticipant represents a meeting participant
type MeetingParticipant struct {
	AgentID      string                 `json:"agent_id"`
	Name         string                 `json:"name"`
	Role         MeetingParticipantRole `json:"role"`
	Status       ParticipantStatus      `json:"status"`
	JoinedAt     *time.Time             `json:"joined_at,omitempty"`
	LeftAt       *time.Time             `json:"left_at,omitempty"`
	RSVP         RSVPStatus             `json:"rsvp"`
	Notes        string                 `json:"notes,omitempty"`
}

// MeetingParticipantRole represents the role of a participant
type MeetingParticipantRole string

const (
	MeetingRoleOrganizer  MeetingParticipantRole = "organizer"
	MeetingRolePresenter  MeetingParticipantRole = "presenter"
	MeetingRoleAttendee   MeetingParticipantRole = "attendee"
	MeetingRoleObserver   MeetingParticipantRole = "observer"
	MeetingRoleScribe     MeetingParticipantRole = "scribe"
	MeetingRoleFacilitator MeetingParticipantRole = "facilitator"
)

// ParticipantStatus represents the participation status
type ParticipantStatus string

const (
	ParticipantStatusInvited   ParticipantStatus = "invited"
	ParticipantStatusConfirmed ParticipantStatus = "confirmed"
	ParticipantStatusDeclined  ParticipantStatus = "declined"
	ParticipantStatusPresent   ParticipantStatus = "present"
	ParticipantStatusAbsent    ParticipantStatus = "absent"
	ParticipantStatusLate      ParticipantStatus = "late"
)

// RSVPStatus represents the response to a meeting invitation
type RSVPStatus string

const (
	RSVPYes     RSVPStatus = "yes"
	RSVPNo      RSVPStatus = "no"
	RSVPMaybe   RSVPStatus = "maybe"
	RSVPPending RSVPStatus = "pending"
)

// AgendaItem represents an item on the meeting agenda
type AgendaItem struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Duration    time.Duration `json:"duration"`
	PresenterID string        `json:"presenter_id,omitempty"`
	Status      AgendaStatus  `json:"status"`
	Order       int           `json:"order"`
	Notes       string        `json:"notes,omitempty"`
}

// AgendaStatus represents the status of an agenda item
type AgendaStatus string

const (
	AgendaStatusPending    AgendaStatus = "pending"
	AgendaStatusInProgress AgendaStatus = "in_progress"
	AgendaStatusCompleted  AgendaStatus = "completed"
	AgendaStatusSkipped    AgendaStatus = "skipped"
)

// MeetingMinutes represents the minutes/notes from a meeting
type MeetingMinutes struct {
	MeetingID      string            `json:"meeting_id"`
	Summary        string            `json:"summary"`
	KeyPoints      []string          `json:"key_points,omitempty"`
	Decisions      []MeetingDecision `json:"decisions,omitempty"`
	ActionItems    []MeetingActionItem `json:"action_items,omitempty"`
	Notes          string            `json:"notes,omitempty"`
	RecordedBy     string            `json:"recorded_by"`
	RecordedAt     time.Time         `json:"recorded_at"`
	NextMeeting    *NextMeetingInfo  `json:"next_meeting,omitempty"`
	Attachments    []MeetingAttachment `json:"attachments,omitempty"`
}

// MeetingDecision represents a decision made during a meeting
type MeetingDecision struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	ProposedBy  string    `json:"proposed_by"`
	ApprovedBy  []string  `json:"approved_by,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// MeetingActionItem represents a task assigned during a meeting
type MeetingActionItem struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	AssigneeID  string     `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Priority    CEOPriority `json:"priority"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
}

// NextMeetingInfo contains information about the next meeting
type NextMeetingInfo struct {
	ScheduledAt time.Time `json:"scheduled_at"`
	Title       string    `json:"title,omitempty"`
	AgendaDraft []string  `json:"agenda_draft,omitempty"`
}

// MeetingAttachment represents a file attached to meeting minutes
type MeetingAttachment struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Type       string `json:"type"`
	Size       int64  `json:"size"`
	UploadedBy string `json:"uploaded_by"`
}

// RecurrenceRule defines how a meeting recurs
type RecurrenceRule struct {
	Frequency   RecurrenceFrequency `json:"frequency"`
	Interval    int                 `json:"interval"`
	DaysOfWeek  []int               `json:"days_of_week,omitempty"` // 0 = Sunday
	EndDate     *time.Time          `json:"end_date,omitempty"`
	Occurrences int                 `json:"occurrences,omitempty"`
}

// RecurrenceFrequency represents how often a meeting recurs
type RecurrenceFrequency string

const (
	RecurrenceDaily    RecurrenceFrequency = "daily"
	RecurrenceWeekly   RecurrenceFrequency = "weekly"
	RecurrenceBiweekly RecurrenceFrequency = "biweekly"
	RecurrenceMonthly  RecurrenceFrequency = "monthly"
)

// MeetingManager manages meetings and scheduling
type MeetingManager struct {
	mu       sync.RWMutex
	meetings map[string]*Meeting
	counter  int64

	// Configuration
	config MeetingConfig

	// Event handlers
	handlers []MeetingEventHandler
}

// MeetingConfig contains configuration for meetings
type MeetingConfig struct {
	// DefaultMeetingDuration is the default duration for new meetings
	DefaultMeetingDuration time.Duration `json:"default_meeting_duration"`

	// ReminderTimes are the times before a meeting to send reminders
	ReminderTimes []time.Duration `json:"reminder_times"`

	// AutoGenerateMinutes enables automatic minutes generation
	AutoGenerateMinutes bool `json:"auto_generate_minutes"`

	// RequireMinutesApproval requires approval before publishing minutes
	RequireMinutesApproval bool `json:"require_minutes_approval"`

	// MaxParticipants is the maximum number of participants per meeting
	MaxParticipants int `json:"max_participants"`

	// AllowOverlappingMeetings allows agents to attend multiple simultaneous meetings
	AllowOverlappingMeetings bool `json:"allow_overlapping_meetings"`

	// StandupTime is the default time for daily standups
	StandupTime string `json:"standup_time"`

	// Timezone for scheduling
	Timezone string `json:"timezone"`
}

// DefaultMeetingConfig returns default meeting configuration
func DefaultMeetingConfig() MeetingConfig {
	return MeetingConfig{
		DefaultMeetingDuration:   30 * time.Minute,
		ReminderTimes:            []time.Duration{15 * time.Minute, 5 * time.Minute},
		AutoGenerateMinutes:      true,
		RequireMinutesApproval:   false,
		MaxParticipants:          20,
		AllowOverlappingMeetings: false,
		StandupTime:              "09:00",
		Timezone:                 "UTC",
	}
}

// MeetingEvent represents a meeting-related event
type MeetingEvent struct {
	Type      MeetingEventType `json:"type"`
	MeetingID string           `json:"meeting_id"`
	AgentID   string           `json:"agent_id,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
	Data      map[string]any   `json:"data,omitempty"`
}

// MeetingEventType represents the type of meeting event
type MeetingEventType string

const (
	MeetingEventCreated          MeetingEventType = "created"
	MeetingEventStarted          MeetingEventType = "started"
	MeetingEventEnded            MeetingEventType = "ended"
	MeetingEventCancelled        MeetingEventType = "cancelled"
	MeetingEventParticipantJoined MeetingEventType = "participant_joined"
	MeetingEventParticipantLeft   MeetingEventType = "participant_left"
	MeetingEventRSVP             MeetingEventType = "rsvp"
	MeetingEventAgendaUpdated    MeetingEventType = "agenda_updated"
	MeetingEventMinutesPublished MeetingEventType = "minutes_published"
)

// MeetingEventHandler is a function that handles meeting events
type MeetingEventHandler func(event MeetingEvent)

// NewMeetingManager creates a new meeting manager
func NewMeetingManager(config MeetingConfig) *MeetingManager {
	return &MeetingManager{
		meetings: make(map[string]*Meeting),
		config:   config,
		handlers: make([]MeetingEventHandler, 0),
	}
}

// ScheduleMeeting creates and schedules a new meeting
func (m *MeetingManager) ScheduleMeeting(ctx context.Context, title string, meetingType MeetingType,
	organizerID string, scheduledAt time.Time, duration time.Duration) (*Meeting, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter++
	meetingID := fmt.Sprintf("mtg-%d-%d", time.Now().Unix(), m.counter)

	if duration == 0 {
		duration = m.config.DefaultMeetingDuration
	}

	meeting := &Meeting{
		ID:           meetingID,
		Title:        title,
		Type:         meetingType,
		Status:       MeetingStatusScheduled,
		ScheduledAt:  scheduledAt,
		Duration:     duration,
		OrganizerID:  organizerID,
		Participants: make([]MeetingParticipant, 0),
		Agenda:       make([]AgendaItem, 0),
		Priority:     CEOPriorityNormal,
		Tags:         make([]string, 0),
		Metadata:     make(map[string]any),
		createdAt:    time.Now(),
		updatedAt:    time.Now(),
	}

	// Add organizer as participant
	organizer := MeetingParticipant{
		AgentID:  organizerID,
		Role:     MeetingRoleOrganizer,
		Status:   ParticipantStatusConfirmed,
		RSVP:     RSVPYes,
	}
	meeting.Participants = append(meeting.Participants, organizer)

	m.meetings[meetingID] = meeting

	m.emitEvent(MeetingEvent{
		Type:      MeetingEventCreated,
		MeetingID: meetingID,
		AgentID:   organizerID,
		Timestamp: time.Now(),
	})

	return meeting, nil
}

// GetMeeting retrieves a meeting by ID
func (m *MeetingManager) GetMeeting(meetingID string) (*Meeting, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	meeting, exists := m.meetings[meetingID]
	return meeting, exists
}

// AddParticipant adds a participant to a meeting
func (m *MeetingManager) AddParticipant(meetingID string, agentID string, role MeetingParticipantRole) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, exists := m.meetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting not found: %s", meetingID)
	}

	if meeting.Status != MeetingStatusScheduled {
		return fmt.Errorf("cannot add participants to %s meeting", meeting.Status)
	}

	if len(meeting.Participants) >= m.config.MaxParticipants {
		return fmt.Errorf("meeting has reached maximum participants (%d)", m.config.MaxParticipants)
	}

	// Check if already participating
	for _, p := range meeting.Participants {
		if p.AgentID == agentID {
			return fmt.Errorf("agent %s is already a participant", agentID)
		}
	}

	participant := MeetingParticipant{
		AgentID: agentID,
		Role:    role,
		Status:  ParticipantStatusInvited,
		RSVP:    RSVPPending,
	}

	meeting.Participants = append(meeting.Participants, participant)
	meeting.updatedAt = time.Now()

	return nil
}

// RemoveParticipant removes a participant from a meeting
func (m *MeetingManager) RemoveParticipant(meetingID string, agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, exists := m.meetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting not found: %s", meetingID)
	}

	for i, p := range meeting.Participants {
		if p.AgentID == agentID {
			meeting.Participants = append(meeting.Participants[:i], meeting.Participants[i+1:]...)
			meeting.updatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("agent %s is not a participant", agentID)
}

// UpdateRSVP updates a participant's RSVP status
func (m *MeetingManager) UpdateRSVP(meetingID string, agentID string, rsvp RSVPStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, exists := m.meetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting not found: %s", meetingID)
	}

	for i := range meeting.Participants {
		if meeting.Participants[i].AgentID == agentID {
			meeting.Participants[i].RSVP = rsvp

			switch rsvp {
			case RSVPYes:
				meeting.Participants[i].Status = ParticipantStatusConfirmed
			case RSVPNo:
				meeting.Participants[i].Status = ParticipantStatusDeclined
			default:
				meeting.Participants[i].Status = ParticipantStatusInvited
			}

			meeting.updatedAt = time.Now()

			m.emitEvent(MeetingEvent{
				Type:      MeetingEventRSVP,
				MeetingID: meetingID,
				AgentID:   agentID,
				Timestamp: time.Now(),
				Data:      map[string]any{"rsvp": rsvp},
			})

			return nil
		}
	}

	return fmt.Errorf("agent %s is not a participant", agentID)
}

// AddAgendaItem adds an item to the meeting agenda
func (m *MeetingManager) AddAgendaItem(meetingID string, title string, description string,
	duration time.Duration, presenterID string) (*AgendaItem, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, exists := m.meetings[meetingID]
	if !exists {
		return nil, fmt.Errorf("meeting not found: %s", meetingID)
	}

	item := AgendaItem{
		ID:          fmt.Sprintf("agenda-%d", len(meeting.Agenda)+1),
		Title:       title,
		Description: description,
		Duration:    duration,
		PresenterID: presenterID,
		Status:      AgendaStatusPending,
		Order:       len(meeting.Agenda),
	}

	meeting.Agenda = append(meeting.Agenda, item)
	meeting.updatedAt = time.Now()

	return &item, nil
}

// StartMeeting starts a scheduled meeting
func (m *MeetingManager) StartMeeting(meetingID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, exists := m.meetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting not found: %s", meetingID)
	}

	if meeting.Status != MeetingStatusScheduled {
		return fmt.Errorf("cannot start meeting with status: %s", meeting.Status)
	}

	now := time.Now()
	meeting.Status = MeetingStatusActive
	meeting.StartedAt = &now
	meeting.updatedAt = now

	// Update participant statuses
	for i := range meeting.Participants {
		if meeting.Participants[i].RSVP == RSVPYes {
			meeting.Participants[i].Status = ParticipantStatusPresent
			meeting.Participants[i].JoinedAt = &now
		}
	}

	m.emitEvent(MeetingEvent{
		Type:      MeetingEventStarted,
		MeetingID: meetingID,
		Timestamp: now,
	})

	return nil
}

// EndMeeting ends an active meeting
func (m *MeetingManager) EndMeeting(meetingID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, exists := m.meetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting not found: %s", meetingID)
	}

	if meeting.Status != MeetingStatusActive && meeting.Status != MeetingStatusPaused {
		return fmt.Errorf("cannot end meeting with status: %s", meeting.Status)
	}

	now := time.Now()
	meeting.Status = MeetingStatusCompleted
	meeting.EndedAt = &now
	meeting.updatedAt = now

	// Update participant statuses
	for i := range meeting.Participants {
		if meeting.Participants[i].Status == ParticipantStatusPresent {
			meeting.Participants[i].LeftAt = &now
		}
	}

	// Auto-generate minutes if enabled
	if m.config.AutoGenerateMinutes {
		m.generateMinutes(meeting)
	}

	m.emitEvent(MeetingEvent{
		Type:      MeetingEventEnded,
		MeetingID: meetingID,
		Timestamp: now,
	})

	return nil
}

// CancelMeeting cancels a scheduled meeting
func (m *MeetingManager) CancelMeeting(meetingID string, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, exists := m.meetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting not found: %s", meetingID)
	}

	if meeting.Status == MeetingStatusCompleted || meeting.Status == MeetingStatusCancelled {
		return fmt.Errorf("meeting already %s", meeting.Status)
	}

	meeting.Status = MeetingStatusCancelled
	meeting.updatedAt = time.Now()
	meeting.Metadata["cancellation_reason"] = reason

	m.emitEvent(MeetingEvent{
		Type:      MeetingEventCancelled,
		MeetingID: meetingID,
		Timestamp: time.Now(),
		Data:      map[string]any{"reason": reason},
	})

	return nil
}

// SkipMeeting marks a meeting as skipped (for CEO directives)
func (m *MeetingManager) SkipMeeting(meetingID string, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, exists := m.meetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting not found: %s", meetingID)
	}

	if meeting.Status != MeetingStatusScheduled {
		return fmt.Errorf("can only skip scheduled meetings")
	}

	meeting.Status = MeetingStatusSkipped
	meeting.updatedAt = time.Now()
	meeting.Metadata["skip_reason"] = reason
	meeting.Metadata["skipped_at"] = time.Now()

	return nil
}

// generateMinutes creates meeting minutes automatically
func (m *MeetingManager) generateMinutes(meeting *Meeting) {
	minutes := &MeetingMinutes{
		MeetingID:   meeting.ID,
		Summary:     fmt.Sprintf("%s - %s meeting", meeting.Title, meeting.Type),
		KeyPoints:   make([]string, 0),
		Decisions:   make([]MeetingDecision, 0),
		ActionItems: make([]MeetingActionItem, 0),
		RecordedBy:  meeting.OrganizerID,
		RecordedAt:  time.Now(),
	}

	// Extract key points from agenda
	for _, item := range meeting.Agenda {
		if item.Status == AgendaStatusCompleted {
			minutes.KeyPoints = append(minutes.KeyPoints, item.Title)
		}
	}

	// List participants
	var attendees []string
	for _, p := range meeting.Participants {
		if p.Status == ParticipantStatusPresent {
			attendees = append(attendees, p.AgentID)
		}
	}

	minutes.Notes = fmt.Sprintf("Attendees: %s", strings.Join(attendees, ", "))

	meeting.Minutes = minutes
}

// UpdateMinutes updates the meeting minutes
func (m *MeetingManager) UpdateMinutes(meetingID string, minutes *MeetingMinutes) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, exists := m.meetings[meetingID]
	if !exists {
		return fmt.Errorf("meeting not found: %s", meetingID)
	}

	meeting.Minutes = minutes
	meeting.updatedAt = time.Now()

	m.emitEvent(MeetingEvent{
		Type:      MeetingEventMinutesPublished,
		MeetingID: meetingID,
		Timestamp: time.Now(),
	})

	return nil
}

// AddActionItem adds an action item to meeting minutes
func (m *MeetingManager) AddActionItem(meetingID string, description string,
	assigneeID string, dueDate *time.Time, priority CEOPriority) (*MeetingActionItem, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, exists := m.meetings[meetingID]
	if !exists {
		return nil, fmt.Errorf("meeting not found: %s", meetingID)
	}

	if meeting.Minutes == nil {
		return nil, fmt.Errorf("meeting minutes not yet created")
	}

	actionItem := MeetingActionItem{
		ID:          fmt.Sprintf("action-%d", len(meeting.Minutes.ActionItems)+1),
		Description: description,
		AssigneeID:  assigneeID,
		DueDate:     dueDate,
		Priority:    priority,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	meeting.Minutes.ActionItems = append(meeting.Minutes.ActionItems, actionItem)
	meeting.updatedAt = time.Now()

	return &actionItem, nil
}

// GetUpcomingMeetings returns meetings scheduled in the future
func (m *MeetingManager) GetUpcomingMeetings(agentID string, limit int) []*Meeting {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	var upcoming []*Meeting

	for _, meeting := range m.meetings {
		if meeting.ScheduledAt.After(now) &&
			(meeting.Status == MeetingStatusScheduled || meeting.Status == MeetingStatusActive) {

			// Filter by agent participation if specified
			if agentID != "" {
				isParticipant := false
				for _, p := range meeting.Participants {
					if p.AgentID == agentID && p.RSVP != RSVPNo {
						isParticipant = true
						break
					}
				}
				if !isParticipant {
					continue
				}
			}

			upcoming = append(upcoming, meeting)
		}
	}

	// Sort by scheduled time
	for i := 0; i < len(upcoming); i++ {
		for j := i + 1; j < len(upcoming); j++ {
			if upcoming[j].ScheduledAt.Before(upcoming[i].ScheduledAt) {
				upcoming[i], upcoming[j] = upcoming[j], upcoming[i]
			}
		}
	}

	if limit > 0 && limit < len(upcoming) {
		return upcoming[:limit]
	}
	return upcoming
}

// GetMeetingsByStatus returns meetings filtered by status
func (m *MeetingManager) GetMeetingsByStatus(status MeetingStatus) []*Meeting {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var meetings []*Meeting
	for _, meeting := range m.meetings {
		if meeting.Status == status {
			meetings = append(meetings, meeting)
		}
	}
	return meetings
}

// RegisterEventHandler registers a handler for meeting events
func (m *MeetingManager) RegisterEventHandler(handler MeetingEventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
}

// emitEvent emits a meeting event to all handlers
func (m *MeetingManager) emitEvent(event MeetingEvent) {
	for _, handler := range m.handlers {
		go handler(event)
	}
}

// GetStats returns meeting statistics
func (m *MeetingManager) GetStats() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]any{
		"total_meetings":     len(m.meetings),
		"by_status":          make(map[string]int),
		"by_type":            make(map[string]int),
		"total_participants": 0,
		"total_minutes":      0,
	}

	byStatus := stats["by_status"].(map[string]int)
	byType := stats["by_type"].(map[string]int)

	for _, meeting := range m.meetings {
		byStatus[string(meeting.Status)]++
		byType[string(meeting.Type)]++
		stats["total_participants"] = stats["total_participants"].(int) + len(meeting.Participants)

		if meeting.Minutes != nil {
			stats["total_minutes"] = stats["total_minutes"].(int) + 1
		}
	}

	return stats
}

// CheckConflicts checks for scheduling conflicts for an agent
func (m *MeetingManager) CheckConflicts(agentID string, startTime time.Time, duration time.Duration) []*Meeting {
	m.mu.RLock()
	defer m.mu.RUnlock()

	endTime := startTime.Add(duration)
	var conflicts []*Meeting

	for _, meeting := range m.meetings {
		if meeting.Status != MeetingStatusScheduled && meeting.Status != MeetingStatusActive {
			continue
		}

		// Check if agent is participating
		isParticipant := false
		for _, p := range meeting.Participants {
			if p.AgentID == agentID && p.RSVP == RSVPYes {
				isParticipant = true
				break
			}
		}

		if !isParticipant {
			continue
		}

		// Check for time overlap
		meetingEnd := meeting.ScheduledAt.Add(meeting.Duration)
		if startTime.Before(meetingEnd) && endTime.After(meeting.ScheduledAt) {
			conflicts = append(conflicts, meeting)
		}
	}

	return conflicts
}
