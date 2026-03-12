// Meeting Scheduler for scheduling meetings in advance
package meeting

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Scheduler manages scheduled meetings
type Scheduler struct {
	mu          sync.RWMutex
	schedules   map[string]*ScheduleEntry
	onDue       func(*ScheduleEntry)
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// ScheduleEntry represents a scheduled meeting
type ScheduleEntry struct {
	ID           string
	Topic        string
	Description  string
	ScheduledAt  time.Time
	Participants []string
	Facilitator  string
	Agenda       []string
	Context      map[string]interface{}
	
	// Scheduling options
	Recurring    *RecurringSchedule
	Reminder     time.Duration // Send reminder before meeting
	AutoStart    bool          // Start automatically at scheduled time
	
	// Status
	Status       ScheduleStatus
	CreatedAt    time.Time
	MeetingID    string // ID of created meeting (if started)
}

// ScheduleStatus represents the status of a scheduled entry
type ScheduleStatus string

const (
	ScheduleStatusPending   ScheduleStatus = "pending"
	ScheduleStatusReminded  ScheduleStatus = "reminded"
	ScheduleStatusStarted   ScheduleStatus = "started"
	ScheduleStatusCompleted ScheduleStatus = "completed"
	ScheduleStatusCancelled ScheduleStatus = "cancelled"
)

// RecurringSchedule defines recurring meeting pattern
type RecurringSchedule struct {
	Type      string // "daily", "weekly", "monthly"
	Interval  int    // Every N days/weeks/months
	Weekday   int    // For weekly: 0=Sunday, 1=Monday, etc.
	DayOfMonth int   // For monthly: 1-31
	EndDate   *time.Time
	MaxOccurrences int
}

// NewScheduler creates a new meeting scheduler
func NewScheduler() *Scheduler {
	s := &Scheduler{
		schedules: make(map[string]*ScheduleEntry),
		stopChan:  make(chan struct{}),
	}
	
	// Start the scheduler loop
	s.wg.Add(1)
	go s.schedulerLoop()
	
	return s
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

// ScheduleMeeting schedules a new meeting
func (s *Scheduler) ScheduleMeeting(config ScheduleConfig) (*ScheduleEntry, error) {
	if config.ScheduledAt.IsZero() {
		return nil, fmt.Errorf("scheduled time is required")
	}
	
	if config.ScheduledAt.Before(time.Now()) {
		return nil, fmt.Errorf("scheduled time must be in the future")
	}

	entry := &ScheduleEntry{
		ID:           generateScheduleID(),
		Topic:        config.Topic,
		Description:  config.Description,
		ScheduledAt:  config.ScheduledAt,
		Participants: config.Participants,
		Facilitator:  config.Facilitator,
		Agenda:       config.Agenda,
		Context:      config.Context,
		Recurring:    config.Recurring,
		Reminder:     config.Reminder,
		AutoStart:    config.AutoStart,
		Status:       ScheduleStatusPending,
		CreatedAt:    time.Now(),
	}

	s.mu.Lock()
	s.schedules[entry.ID] = entry
	s.mu.Unlock()

	return entry, nil
}

// GetSchedule returns a schedule entry by ID
func (s *Scheduler) GetSchedule(id string) (*ScheduleEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.schedules[id]
	return entry, ok
}

// ListSchedules returns all schedules (optionally filtered by status)
func (s *Scheduler) ListSchedules(status ...ScheduleStatus) []*ScheduleEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*ScheduleEntry
	for _, entry := range s.schedules {
		if len(status) == 0 {
			result = append(result, entry)
		} else {
			for _, s := range status {
				if entry.Status == s {
					result = append(result, entry)
					break
				}
			}
		}
	}

	// Sort by scheduled time
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].ScheduledAt.After(result[j].ScheduledAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// CancelSchedule cancels a scheduled meeting
func (s *Scheduler) CancelSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.schedules[id]
	if !ok {
		return fmt.Errorf("schedule not found")
	}

	if entry.Status == ScheduleStatusStarted || entry.Status == ScheduleStatusCompleted {
		return fmt.Errorf("cannot cancel meeting that has already started or completed")
	}

	entry.Status = ScheduleStatusCancelled
	return nil
}

// UpdateSchedule updates a schedule entry
func (s *Scheduler) UpdateSchedule(id string, updates ScheduleUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.schedules[id]
	if !ok {
		return fmt.Errorf("schedule not found")
	}

	if entry.Status == ScheduleStatusStarted || entry.Status == ScheduleStatusCompleted {
		return fmt.Errorf("cannot update meeting that has already started or completed")
	}

	if updates.ScheduledAt != nil {
		entry.ScheduledAt = *updates.ScheduledAt
	}
	if updates.Topic != "" {
		entry.Topic = updates.Topic
	}
	if updates.Description != "" {
		entry.Description = updates.Description
	}
	if len(updates.Participants) > 0 {
		entry.Participants = updates.Participants
	}
	if len(updates.Agenda) > 0 {
		entry.Agenda = updates.Agenda
	}

	return nil
}

// SetOnDue sets the callback for when a schedule is due
func (s *Scheduler) SetOnDue(fn func(*ScheduleEntry)) {
	s.onDue = fn
}

// schedulerLoop checks for due schedules
func (s *Scheduler) schedulerLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkSchedules()
		case <-s.stopChan:
			return
		}
	}
}

// checkSchedules checks for due or reminder-due schedules
func (s *Scheduler) checkSchedules() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entry := range s.schedules {
		// Skip cancelled or completed
		if entry.Status == ScheduleStatusCancelled || entry.Status == ScheduleStatusCompleted {
			continue
		}

		// Check for reminder
		if entry.Status == ScheduleStatusPending && entry.Reminder > 0 {
			reminderTime := entry.ScheduledAt.Add(-entry.Reminder)
			if now.After(reminderTime) && now.Before(entry.ScheduledAt) {
				entry.Status = ScheduleStatusReminded
				if s.onDue != nil {
					go s.onDue(entry)
				}
			}
		}

		// Check if meeting is due
		if entry.Status == ScheduleStatusPending || entry.Status == ScheduleStatusReminded {
			if now.After(entry.ScheduledAt) || now.Equal(entry.ScheduledAt) {
				entry.Status = ScheduleStatusStarted
				if s.onDue != nil {
					go s.onDue(entry)
				}
			}
		}
	}
}

// GetUpcomingMeetings returns meetings scheduled within the next duration
func (s *Scheduler) GetUpcomingMeetings(within time.Duration) []*ScheduleEntry {
	now := time.Now()
	cutoff := now.Add(within)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*ScheduleEntry
	for _, entry := range s.schedules {
		if entry.Status == ScheduleStatusPending || entry.Status == ScheduleStatusReminded {
			if entry.ScheduledAt.After(now) && entry.ScheduledAt.Before(cutoff) {
				result = append(result, entry)
			}
		}
	}

	return result
}

// ScheduleConfig provides configuration for scheduling a meeting
type ScheduleConfig struct {
	Topic        string
	Description  string
	ScheduledAt  time.Time
	Participants []string
	Facilitator  string
	Agenda       []string
	Context      map[string]interface{}
	Recurring    *RecurringSchedule
	Reminder     time.Duration
	AutoStart    bool
}

// ScheduleUpdate provides fields for updating a schedule
type ScheduleUpdate struct {
	ScheduledAt  *time.Time
	Topic        string
	Description  string
	Participants []string
	Agenda       []string
}

// Helper functions
func generateScheduleID() string {
	return fmt.Sprintf("sch-%d", time.Now().UnixNano())
}

// Quick schedule helpers

// ScheduleDaily creates a daily recurring meeting
func (s *Scheduler) ScheduleDaily(topic string, hour, minute int, participants []string) (*ScheduleEntry, error) {
	// Calculate next occurrence
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}

	return s.ScheduleMeeting(ScheduleConfig{
		Topic:        topic,
		ScheduledAt:  next,
		Participants: participants,
		Facilitator:  "jarvis",
		Recurring: &RecurringSchedule{
			Type:     "daily",
			Interval: 1,
		},
		Reminder:  15 * time.Minute,
		AutoStart: true,
	})
}

// ScheduleWeekly creates a weekly recurring meeting
func (s *Scheduler) ScheduleWeekly(topic string, weekday time.Weekday, hour, minute int, participants []string) (*ScheduleEntry, error) {
	now := time.Now()
	
	// Calculate days until target weekday
	daysUntil := int(weekday - now.Weekday())
	if daysUntil < 0 {
		daysUntil += 7
	}
	
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	next = next.Add(time.Duration(daysUntil) * 24 * time.Hour)
	
	if next.Before(now) {
		next = next.Add(7 * 24 * time.Hour)
	}

	return s.ScheduleMeeting(ScheduleConfig{
		Topic:        topic,
		ScheduledAt:  next,
		Participants: participants,
		Facilitator:  "jarvis",
		Recurring: &RecurringSchedule{
			Type:    "weekly",
			Weekday: int(weekday),
		},
		Reminder:  30 * time.Minute,
		AutoStart: true,
	})
}

// ScheduleNow schedules a meeting to start immediately (within 1 minute)
func (s *Scheduler) ScheduleNow(topic string, participants []string) (*ScheduleEntry, error) {
	return s.ScheduleMeeting(ScheduleConfig{
		Topic:        topic,
		ScheduledAt:  time.Now().Add(1 * time.Minute),
		Participants: participants,
		Facilitator:  "jarvis",
		AutoStart:    true,
	})
}

// FormatSchedule formats a schedule entry for display
func FormatSchedule(entry *ScheduleEntry) string {
	var builder strings.Builder
	
	builder.WriteString(fmt.Sprintf("📅 **%s**\n", entry.Topic))
	builder.WriteString(fmt.Sprintf("🆔 ID: %s\n", entry.ID))
	builder.WriteString(fmt.Sprintf("🕐 Scheduled: %s\n", entry.ScheduledAt.Format("2006-01-02 15:04")))
	builder.WriteString(fmt.Sprintf("📊 Status: %s\n", entry.Status))
	
	if entry.Description != "" {
		builder.WriteString(fmt.Sprintf("📝 %s\n", entry.Description))
	}
	
	builder.WriteString(fmt.Sprintf("👥 Participants (%d): %v\n", len(entry.Participants), entry.Participants))
	
	if len(entry.Agenda) > 0 {
		builder.WriteString("📋 Agenda:\n")
		for i, item := range entry.Agenda {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, item))
		}
	}
	
	if entry.Recurring != nil {
		builder.WriteString(fmt.Sprintf("🔄 Recurring: %s\n", entry.Recurring.Type))
	}
	
	if entry.Reminder > 0 {
		builder.WriteString(fmt.Sprintf("⏰ Reminder: %s before\n", entry.Reminder))
	}
	
	return builder.String()
}
