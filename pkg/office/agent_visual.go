package office

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// AgentStatus represents the visual status of an agent.
type AgentStatus string

const (
	// AgentStatusIdle means the agent is idle and waiting.
	AgentStatusIdle AgentStatus = "idle"
	// AgentStatusWorking means the agent is actively working.
	AgentStatusWorking AgentStatus = "working"
	// AgentStatusThinking means the agent is processing/thinking.
	AgentStatusThinking AgentStatus = "thinking"
	// AgentStatusTalking means the agent is communicating.
	AgentStatusTalking AgentStatus = "talking"
	// AgentStatusAway means the agent is temporarily away.
	AgentStatusAway AgentStatus = "away"
	// AgentStatusOffline means the agent is offline.
	AgentStatusOffline AgentStatus = "offline"
	// AgentStatusError means the agent has encountered an error.
	AgentStatusError AgentStatus = "error"
)

// IsValid checks if the agent status is valid.
func (as AgentStatus) IsValid() bool {
	switch as {
	case AgentStatusIdle, AgentStatusWorking, AgentStatusThinking,
		AgentStatusTalking, AgentStatusAway, AgentStatusOffline, AgentStatusError:
		return true
	}
	return false
}

// Color returns the color associated with the status.
func (as AgentStatus) Color() string {
	switch as {
	case AgentStatusIdle:
		return "#10B981" // Green
	case AgentStatusWorking:
		return "#3B82F6" // Blue
	case AgentStatusThinking:
		return "#8B5CF6" // Purple
	case AgentStatusTalking:
		return "#F59E0B" // Amber
	case AgentStatusAway:
		return "#6B7280" // Gray
	case AgentStatusOffline:
		return "#374151" // Dark Gray
	case AgentStatusError:
		return "#EF4444" // Red
	default:
		return "#6B7280" // Gray
	}
}

// Animation returns the animation type for the status.
func (as AgentStatus) Animation() string {
	switch as {
	case AgentStatusIdle:
		return "breathe"
	case AgentStatusWorking:
		return "pulse"
	case AgentStatusThinking:
		return "ripple"
	case AgentStatusTalking:
		return "bounce"
	case AgentStatusAway:
		return "fade"
	case AgentStatusOffline:
		return "static"
	case AgentStatusError:
		return "shake"
	default:
		return "static"
	}
}

// AvatarType represents the type of avatar for an agent.
type AvatarType string

const (
	// AvatarTypeEmoji uses an emoji as the avatar.
	AvatarTypeEmoji AvatarType = "emoji"
	// AvatarTypeInitials uses initials as the avatar.
	AvatarTypeInitials AvatarType = "initials"
	// AvatarTypeRobot uses a robot-style avatar.
	AvatarTypeRobot AvatarType = "robot"
	// AvatarTypeCustom uses a custom avatar.
	AvatarTypeCustom AvatarType = "custom"
)

// AgentVisualState represents the visual state of an agent in the office.
type AgentVisualState struct {
	AgentID      string            `json:"agent_id"`
	AgentName    string            `json:"agent_name"`
	Avatar       Avatar            `json:"avatar"`
	Status       AgentStatus       `json:"status"`
	Position     Position          `json:"position"`
	RoomID       string            `json:"room_id"`
	Direction    Direction         `json:"direction"`
	Activity     string            `json:"activity,omitempty"`
	Progress     float64           `json:"progress,omitempty"` // 0.0 to 1.0
	Metadata     map[string]string `json:"metadata,omitempty"`
	LastActivity time.Time         `json:"last_activity"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// Avatar represents an agent's avatar.
type Avatar struct {
	Type     AvatarType `json:"type"`
	Value    string     `json:"value"`              // emoji, initials, or URL
	Color    string     `json:"color"`              // background color
	BorderColor string  `json:"border_color,omitempty"`
}

// Direction represents the facing direction of an agent.
type Direction string

const (
	DirectionNorth Direction = "north"
	DirectionEast  Direction = "east"
	DirectionSouth Direction = "south"
	DirectionWest  Direction = "west"
)

// DefaultAvatars provides default avatar options.
var DefaultAvatars = []Avatar{
	{Type: AvatarTypeEmoji, Value: "🤖", Color: "#3B82F6"},
	{Type: AvatarTypeEmoji, Value: "👨‍💻", Color: "#10B981"},
	{Type: AvatarTypeEmoji, Value: "👩‍💻", Color: "#8B5CF6"},
	{Type: AvatarTypeEmoji, Value: "🦊", Color: "#F59E0B"},
	{Type: AvatarTypeEmoji, Value: "🐱", Color: "#EF4444"},
	{Type: AvatarTypeEmoji, Value: "🐶", Color: "#6366F1"},
	{Type: AvatarTypeEmoji, Value: "🐼", Color: "#374151"},
	{Type: AvatarTypeEmoji, Value: "🐨", Color: "#6B7280"},
	{Type: AvatarTypeRobot, Value: "R1", Color: "#3B82F6"},
	{Type: AvatarTypeRobot, Value: "R2", Color: "#10B981"},
	{Type: AvatarTypeRobot, Value: "R3", Color: "#8B5CF6"},
}

// NewAgentVisualState creates a new visual state for an agent.
func NewAgentVisualState(agentID, agentName string, avatar Avatar) *AgentVisualState {
	now := time.Now()
	return &AgentVisualState{
		AgentID:      agentID,
		AgentName:    agentName,
		Avatar:       avatar,
		Status:       AgentStatusIdle,
		Position:     Position{X: 0, Y: 0},
		Direction:    DirectionSouth,
		Metadata:     make(map[string]string),
		LastActivity: now,
		UpdatedAt:    now,
	}
}

// SetStatus updates the agent's status.
func (avs *AgentVisualState) SetStatus(status AgentStatus) error {
	if !status.IsValid() {
		return fmt.Errorf("invalid agent status: %s", status)
	}
	avs.Status = status
	avs.LastActivity = time.Now()
	avs.UpdatedAt = time.Now()
	return nil
}

// SetPosition updates the agent's position.
func (avs *AgentVisualState) SetPosition(position Position) {
	avs.Position = position
	avs.RoomID = position.RoomID
	avs.UpdatedAt = time.Now()
}

// SetActivity updates the agent's current activity.
func (avs *AgentVisualState) SetActivity(activity string) {
	avs.Activity = activity
	avs.LastActivity = time.Now()
	avs.UpdatedAt = time.Now()
}

// SetProgress updates the agent's progress (0.0 to 1.0).
func (avs *AgentVisualState) SetProgress(progress float64) {
	if progress < 0 {
		progress = 0
	} else if progress > 1 {
		progress = 1
	}
	avs.Progress = progress
	avs.UpdatedAt = time.Now()
}

// SetDirection updates the agent's facing direction.
func (avs *AgentVisualState) SetDirection(direction Direction) {
	avs.Direction = direction
	avs.UpdatedAt = time.Now()
}

// IsActive returns true if the agent is currently active.
func (avs *AgentVisualState) IsActive() bool {
	return avs.Status != AgentStatusOffline && avs.Status != AgentStatusAway
}

// GetIdleDuration returns how long the agent has been idle.
func (avs *AgentVisualState) GetIdleDuration() time.Duration {
	return time.Since(avs.LastActivity)
}

// AgentVisualizer manages the visual states of all agents.
type AgentVisualizer struct {
	mu          sync.RWMutex
	states      map[string]*AgentVisualState
	officeMgr   *OfficeManager
	handlers    []VisualStateHandler
}

// VisualStateHandler is a function that handles visual state changes.
type VisualStateHandler func(event VisualStateEvent)

// VisualStateEvent represents a visual state change event.
type VisualStateEvent struct {
	Type      VisualEventType    `json:"type"`
	Timestamp time.Time          `json:"timestamp"`
	AgentID   string             `json:"agent_id"`
	OldState  *AgentVisualState  `json:"old_state,omitempty"`
	NewState  *AgentVisualState  `json:"new_state"`
}

// VisualEventType represents the type of visual event.
type VisualEventType string

const (
	// VisualEventStateCreated is fired when a new visual state is created.
	VisualEventStateCreated VisualEventType = "state_created"
	// VisualEventStateUpdated is fired when a visual state is updated.
	VisualEventStateUpdated VisualEventType = "state_updated"
	// VisualEventStateDeleted is fired when a visual state is deleted.
	VisualEventStateDeleted VisualEventType = "state_deleted"
	// VisualEventStatusChanged is fired when an agent's status changes.
	VisualEventStatusChanged VisualEventType = "status_changed"
	// VisualEventPositionChanged is fired when an agent moves.
	VisualEventPositionChanged VisualEventType = "position_changed"
	// VisualEventActivityChanged is fired when an agent's activity changes.
	VisualEventActivityChanged VisualEventType = "activity_changed"
)

// NewAgentVisualizer creates a new agent visualizer.
func NewAgentVisualizer(officeMgr *OfficeManager) *AgentVisualizer {
	av := &AgentVisualizer{
		states:    make(map[string]*AgentVisualState),
		officeMgr: officeMgr,
		handlers:  make([]VisualStateHandler, 0),
	}

	// Register for office events
	if officeMgr != nil {
		officeMgr.RegisterEventHandler(av.handleOfficeEvent)
	}

	return av
}

// RegisterVisualStateHandler registers a visual state handler.
func (av *AgentVisualizer) RegisterVisualStateHandler(handler VisualStateHandler) {
	av.mu.Lock()
	defer av.mu.Unlock()
	av.handlers = append(av.handlers, handler)
}

// CreateState creates a new visual state for an agent.
func (av *AgentVisualizer) CreateState(agentID, agentName string, avatar Avatar) (*AgentVisualState, error) {
	av.mu.Lock()
	defer av.mu.Unlock()

	if _, exists := av.states[agentID]; exists {
		return nil, fmt.Errorf("visual state already exists for agent %s", agentID)
	}

	state := NewAgentVisualState(agentID, agentName, avatar)
	av.states[agentID] = state

	av.emitEvent(VisualStateEvent{
		Type:      VisualEventStateCreated,
		Timestamp: time.Now(),
		AgentID:   agentID,
		NewState:  state,
	})

	return state, nil
}

// GetState retrieves an agent's visual state.
func (av *AgentVisualizer) GetState(agentID string) (*AgentVisualState, error) {
	av.mu.RLock()
	defer av.mu.RUnlock()

	state, exists := av.states[agentID]
	if !exists {
		return nil, fmt.Errorf("visual state not found for agent %s", agentID)
	}

	return state, nil
}

// UpdateState updates an agent's visual state.
func (av *AgentVisualizer) UpdateState(agentID string, updates map[string]interface{}) (*AgentVisualState, error) {
	av.mu.Lock()
	defer av.mu.Unlock()

	state, exists := av.states[agentID]
	if !exists {
		return nil, fmt.Errorf("visual state not found for agent %s", agentID)
	}

	oldState := *state // Copy for event

	if status, ok := updates["status"].(AgentStatus); ok {
		state.Status = status
	}
	if activity, ok := updates["activity"].(string); ok {
		state.Activity = activity
	}
	if progress, ok := updates["progress"].(float64); ok {
		state.Progress = progress
	}
	if direction, ok := updates["direction"].(Direction); ok {
		state.Direction = direction
	}
	if metadata, ok := updates["metadata"].(map[string]string); ok {
		for k, v := range metadata {
			state.Metadata[k] = v
		}
	}

	state.UpdatedAt = time.Now()

	av.emitEvent(VisualStateEvent{
		Type:      VisualEventStateUpdated,
		Timestamp: time.Now(),
		AgentID:   agentID,
		OldState:  &oldState,
		NewState:  state,
	})

	return state, nil
}

// DeleteState removes an agent's visual state.
func (av *AgentVisualizer) DeleteState(agentID string) error {
	av.mu.Lock()
	defer av.mu.Unlock()

	state, exists := av.states[agentID]
	if !exists {
		return fmt.Errorf("visual state not found for agent %s", agentID)
	}

	delete(av.states, agentID)

	av.emitEvent(VisualStateEvent{
		Type:      VisualEventStateDeleted,
		Timestamp: time.Now(),
		AgentID:   agentID,
		OldState:  state,
	})

	return nil
}

// SetStatus updates an agent's status.
func (av *AgentVisualizer) SetStatus(agentID string, status AgentStatus) error {
	av.mu.Lock()
	defer av.mu.Unlock()

	state, exists := av.states[agentID]
	if !exists {
		return fmt.Errorf("visual state not found for agent %s", agentID)
	}

	oldStatus := state.Status
	if err := state.SetStatus(status); err != nil {
		return err
	}

	if oldStatus != status {
		av.emitEvent(VisualStateEvent{
			Type:      VisualEventStatusChanged,
			Timestamp: time.Now(),
			AgentID:   agentID,
			OldState:  &AgentVisualState{Status: oldStatus},
			NewState:  state,
		})
	}

	return nil
}

// SetPosition updates an agent's position.
func (av *AgentVisualizer) SetPosition(agentID string, position Position) error {
	av.mu.Lock()
	defer av.mu.Unlock()

	state, exists := av.states[agentID]
	if !exists {
		return fmt.Errorf("visual state not found for agent %s", agentID)
	}

	oldPosition := state.Position
	state.SetPosition(position)

	if oldPosition != position {
		av.emitEvent(VisualStateEvent{
			Type:      VisualEventPositionChanged,
			Timestamp: time.Now(),
			AgentID:   agentID,
			OldState:  &AgentVisualState{Position: oldPosition},
			NewState:  state,
		})
	}

	return nil
}

// SetActivity updates an agent's activity.
func (av *AgentVisualizer) SetActivity(agentID, activity string) error {
	av.mu.Lock()
	defer av.mu.Unlock()

	state, exists := av.states[agentID]
	if !exists {
		return fmt.Errorf("visual state not found for agent %s", agentID)
	}

	oldActivity := state.Activity
	state.SetActivity(activity)

	if oldActivity != activity {
		av.emitEvent(VisualStateEvent{
			Type:      VisualEventActivityChanged,
			Timestamp: time.Now(),
			AgentID:   agentID,
			OldState:  &AgentVisualState{Activity: oldActivity},
			NewState:  state,
		})
	}

	return nil
}

// SetProgress updates an agent's progress.
func (av *AgentVisualizer) SetProgress(agentID string, progress float64) error {
	av.mu.Lock()
	defer av.mu.Unlock()

	state, exists := av.states[agentID]
	if !exists {
		return fmt.Errorf("visual state not found for agent %s", agentID)
	}

	state.SetProgress(progress)
	return nil
}

// ListStates returns all agent visual states.
func (av *AgentVisualizer) ListStates() []*AgentVisualState {
	av.mu.RLock()
	defer av.mu.RUnlock()

	result := make([]*AgentVisualState, 0, len(av.states))
	for _, state := range av.states {
		result = append(result, state)
	}

	return result
}

// GetStatesByRoom returns all visual states for agents in a room.
func (av *AgentVisualizer) GetStatesByRoom(roomID string) []*AgentVisualState {
	av.mu.RLock()
	defer av.mu.RUnlock()

	result := make([]*AgentVisualState, 0)
	for _, state := range av.states {
		if state.RoomID == roomID {
			result = append(result, state)
		}
	}

	return result
}

// GetStatesByStatus returns all visual states with a specific status.
func (av *AgentVisualizer) GetStatesByStatus(status AgentStatus) []*AgentVisualState {
	av.mu.RLock()
	defer av.mu.RUnlock()

	result := make([]*AgentVisualState, 0)
	for _, state := range av.states {
		if state.Status == status {
			result = append(result, state)
		}
	}

	return result
}

// GetActiveAgents returns all active agents.
func (av *AgentVisualizer) GetActiveAgents() []*AgentVisualState {
	av.mu.RLock()
	defer av.mu.RUnlock()

	result := make([]*AgentVisualState, 0)
	for _, state := range av.states {
		if state.IsActive() {
			result = append(result, state)
		}
	}

	return result
}

// GetAvatarForAgent generates an avatar for an agent based on their ID.
func (av *AgentVisualizer) GetAvatarForAgent(agentID, agentName string) Avatar {
	// Use hash of agentID to pick a consistent avatar
	hash := 0
	for _, c := range agentID {
		hash = (hash*31 + int(c)) % len(DefaultAvatars)
	}

	avatar := DefaultAvatars[hash]

	// If using initials, generate from agent name
	if avatar.Type == AvatarTypeInitials && agentName != "" {
		initials := ""
		for _, word := range strings.Fields(agentName) {
			if len(word) > 0 {
				initials += strings.ToUpper(string(word[0]))
			}
		}
		if initials == "" {
			initials = "A"
		}
		if len(initials) > 2 {
			initials = initials[:2]
		}
		avatar.Value = initials
	}

	return avatar
}

// handleOfficeEvent handles office events to sync visual states.
func (av *AgentVisualizer) handleOfficeEvent(event OfficeEvent) {
	switch event.Type {
	case EventAgentAssigned:
		// Update position when agent is assigned to a room
		if state, exists := av.states[event.AgentID]; exists {
			if assignment, err := av.officeMgr.GetAgentAssignment(event.AgentID); err == nil {
				state.SetPosition(Position{
					X:      assignment.PositionX,
					Y:      assignment.PositionY,
					RoomID: assignment.RoomID,
				})
				state.SetStatus(AgentStatusIdle)
			}
		}

	case EventAgentUnassigned:
		// Mark as offline when unassigned
		if state, exists := av.states[event.AgentID]; exists {
			state.SetStatus(AgentStatusOffline)
			state.RoomID = ""
		}

	case EventAgentMoved:
		// Update position when agent moves
		if state, exists := av.states[event.AgentID]; exists {
			if newX, ok := event.Data["new_x"].(int); ok {
				if newY, ok := event.Data["new_y"].(int); ok {
					state.SetPosition(Position{
						X:      newX,
						Y:      newY,
						RoomID: state.RoomID,
					})
				}
			}
		}
	}
}

// emitEvent emits a visual state event to all handlers.
func (av *AgentVisualizer) emitEvent(event VisualStateEvent) {
	for _, handler := range av.handlers {
		handler(event)
	}
}

// SyncWithOffice syncs visual states with the office manager.
func (av *AgentVisualizer) SyncWithOffice() error {
	av.mu.Lock()
	defer av.mu.Unlock()

	if av.officeMgr == nil {
		return fmt.Errorf("office manager not set")
	}

	// Get all assignments from office manager
	for agentID, assignment := range av.officeMgr.assignments {
		if state, exists := av.states[agentID]; exists {
			state.SetPosition(Position{
				X:      assignment.PositionX,
				Y:      assignment.PositionY,
				RoomID: assignment.RoomID,
			})
		}
	}

	return nil
}

// GetVisualSnapshot returns a complete visual snapshot of all agents.
func (av *AgentVisualizer) GetVisualSnapshot() VisualSnapshot {
	av.mu.RLock()
	defer av.mu.RUnlock()

	states := make([]*AgentVisualState, 0, len(av.states))
	for _, state := range av.states {
		states = append(states, state)
	}

	return VisualSnapshot{
		States:    states,
		Count:     len(states),
		Active:    len(av.GetActiveAgents()),
		Timestamp: time.Now(),
	}
}

// VisualSnapshot represents a snapshot of all visual states.
type VisualSnapshot struct {
	States    []*AgentVisualState `json:"states"`
	Count     int                 `json:"count"`
	Active    int                 `json:"active"`
	Timestamp time.Time           `json:"timestamp"`
}
