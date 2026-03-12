package office

import (
	"fmt"
	"sync"
	"time"
)

// OfficeConfig holds configuration for the office manager.
type OfficeConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxRooms    int    `json:"max_rooms"`
	MaxAgents   int    `json:"max_agents"`
}

// DefaultOfficeConfig returns the default office configuration.
func DefaultOfficeConfig() OfficeConfig {
	return OfficeConfig{
		Name:        "Picoclaw Office",
		Description: "Virtual office for agent collaboration",
		MaxRooms:    100,
		MaxAgents:   500,
	}
}

// AgentAssignment tracks an agent's assignment to a room.
type AgentAssignment struct {
	AgentID    string    `json:"agent_id"`
	RoomID     string    `json:"room_id"`
	PositionX  int       `json:"position_x"`
	PositionY  int       `json:"position_y"`
	AssignedAt time.Time `json:"assigned_at"`
}

// OfficeManager coordinates departments, rooms, and agent assignments.
type OfficeManager struct {
	mu              sync.RWMutex
	config          OfficeConfig
	departments     *DepartmentManager
	rooms           *RoomManager
	assignments     map[string]*AgentAssignment // agentID -> assignment
	agentPositions  map[string]Position         // agentID -> position
	eventHandlers   []OfficeEventHandler
}

// Position represents a 2D position in the office.
type Position struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	RoomID string `json:"room_id"`
}

// OfficeEvent represents an event in the office.
type OfficeEvent struct {
	Type      OfficeEventType `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	AgentID   string          `json:"agent_id,omitempty"`
	RoomID    string          `json:"room_id,omitempty"`
	DeptID    string          `json:"dept_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// OfficeEventType represents the type of office event.
type OfficeEventType string

const (
	// EventAgentAssigned is fired when an agent is assigned to a room.
	EventAgentAssigned OfficeEventType = "agent_assigned"
	// EventAgentUnassigned is fired when an agent is unassigned from a room.
	EventAgentUnassigned OfficeEventType = "agent_unassigned"
	// EventAgentMoved is fired when an agent moves within a room.
	EventAgentMoved OfficeEventType = "agent_moved"
	// EventRoomCreated is fired when a room is created.
	EventRoomCreated OfficeEventType = "room_created"
	// EventRoomDeleted is fired when a room is deleted.
	EventRoomDeleted OfficeEventType = "room_deleted"
	// EventDepartmentCreated is fired when a department is created.
	EventDepartmentCreated OfficeEventType = "department_created"
	// EventDepartmentDeleted is fired when a department is deleted.
	EventDepartmentDeleted OfficeEventType = "department_deleted"
)

// OfficeEventHandler is a function that handles office events.
type OfficeEventHandler func(event OfficeEvent)

// NewOfficeManager creates a new office manager.
func NewOfficeManager(config OfficeConfig) *OfficeManager {
	return &OfficeManager{
		config:         config,
		departments:    NewDepartmentManager(),
		rooms:          NewRoomManager(),
		assignments:    make(map[string]*AgentAssignment),
		agentPositions: make(map[string]Position),
		eventHandlers:  make([]OfficeEventHandler, 0),
	}
}

// Initialize sets up the office with default departments and rooms.
func (om *OfficeManager) Initialize() error {
	om.mu.Lock()
	defer om.mu.Unlock()

	// Create default departments
	for _, deptConfig := range GetDefaultDepartments() {
		if _, err := om.departments.CreateDepartment(deptConfig); err != nil {
			return fmt.Errorf("failed to create department %s: %w", deptConfig.ID, err)
		}
	}

	// Create default rooms for each department
	defaultRooms := []struct {
		ID       string
		Name     string
		DeptID   string
		Type     RoomType
		Capacity int
		GridSize GridSize
	}{
		{"planning-main", "Planning Central", "planning", RoomTypeCollaboration, 8, GridSizeMedium},
		{"planning-focus", "Focus Room Alpha", "planning", RoomTypeFocus, 2, GridSizeSmall},
		{"coding-main", "Dev Hub", "coding", RoomTypeWorkspace, 12, GridSizeLarge},
		{"coding-focus", "Deep Work Zone", "coding", RoomTypeFocus, 4, GridSizeMedium},
		{"design-main", "Design Studio", "design", RoomTypeCollaboration, 6, GridSizeMedium},
		{"design-review", "Design Review", "design", RoomTypeMeeting, 4, GridSizeSmall},
		{"marketing-main", "Marketing Hub", "marketing", RoomTypeWorkspace, 6, GridSizeMedium},
		{"marketing-meeting", "Campaign Room", "marketing", RoomTypeMeeting, 8, GridSizeMedium},
		{"quality-main", "QA Lab", "quality", RoomTypeWorkspace, 8, GridSizeMedium},
		{"quality-testing", "Testing Chamber", "quality", RoomTypeFocus, 4, GridSizeSmall},
		{"legal-main", "Legal Office", "legal", RoomTypeWorkspace, 4, GridSizeMedium},
		{"legal-meeting", "Conference Room", "legal", RoomTypeMeeting, 6, GridSizeMedium},
	}

	for _, roomDef := range defaultRooms {
		if _, err := om.rooms.CreateRoom(
			roomDef.ID,
			roomDef.Name,
			roomDef.DeptID,
			roomDef.Type,
			roomDef.Capacity,
			roomDef.GridSize,
		); err != nil {
			return fmt.Errorf("failed to create room %s: %w", roomDef.ID, err)
		}

		// Update department room count
		om.departments.UpdateRoomCount(roomDef.DeptID, 1)
	}

	return nil
}

// GetConfig returns the office configuration.
func (om *OfficeManager) GetConfig() OfficeConfig {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.config
}

// GetDepartmentManager returns the department manager.
func (om *OfficeManager) GetDepartmentManager() *DepartmentManager {
	return om.departments
}

// GetRoomManager returns the room manager.
func (om *OfficeManager) GetRoomManager() *RoomManager {
	return om.rooms
}

// AssignAgentToRoom assigns an agent to a room.
func (om *OfficeManager) AssignAgentToRoom(agentID, roomID string) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	// Check if agent is already assigned
	if existing, exists := om.assignments[agentID]; exists {
		if existing.RoomID == roomID {
			return fmt.Errorf("agent %s is already assigned to room %s", agentID, roomID)
		}
		// Unassign from current room first
		if err := om.unassignAgentLocked(agentID); err != nil {
			return fmt.Errorf("failed to unassign agent from current room: %w", err)
		}
	}

	// Get the room
	room, err := om.rooms.GetRoom(roomID)
	if err != nil {
		return err
	}

	if room.IsFull() {
		return fmt.Errorf("room %s is at capacity", roomID)
	}

	// Find empty position
	x, y, found := room.Layout.FindEmptyPosition()
	if !found {
		return fmt.Errorf("no empty position available in room %s", roomID)
	}

	// Update room
	room.Occupants = append(room.Occupants, agentID)
	room.Layout.Grid[y][x].Occupant = agentID
	room.Layout.Version++
	room.UpdatedAt = time.Now()

	// Create assignment
	assignment := &AgentAssignment{
		AgentID:    agentID,
		RoomID:     roomID,
		PositionX:  x,
		PositionY:  y,
		AssignedAt: time.Now(),
	}
	om.assignments[agentID] = assignment

	// Update position tracking
	om.agentPositions[agentID] = Position{
		X:      x,
		Y:      y,
		RoomID: roomID,
	}

	// Update department agent count
	om.departments.UpdateAgentCount(room.DepartmentID, 1)

	// Emit event
	om.emitEvent(OfficeEvent{
		Type:      EventAgentAssigned,
		Timestamp: time.Now(),
		AgentID:   agentID,
		RoomID:    roomID,
		DeptID:    room.DepartmentID,
	})

	return nil
}

// UnassignAgent removes an agent from their assigned room.
func (om *OfficeManager) UnassignAgent(agentID string) error {
	om.mu.Lock()
	defer om.mu.Unlock()
	return om.unassignAgentLocked(agentID)
}

// unassignAgentLocked removes an agent from their assigned room (must hold lock).
func (om *OfficeManager) unassignAgentLocked(agentID string) error {
	assignment, exists := om.assignments[agentID]
	if !exists {
		return fmt.Errorf("agent %s is not assigned to any room", agentID)
	}

	// Get the room
	room, err := om.rooms.GetRoom(assignment.RoomID)
	if err != nil {
		return err
	}

	// Remove from room occupants
	for i, id := range room.Occupants {
		if id == agentID {
			room.Occupants = append(room.Occupants[:i], room.Occupants[i+1:]...)
			break
		}
	}

	// Clear from layout
	room.Layout.ClearOccupant(agentID)
	room.UpdatedAt = time.Now()

	// Update department agent count
	om.departments.UpdateAgentCount(room.DepartmentID, -1)

	// Remove assignment
	delete(om.assignments, agentID)
	delete(om.agentPositions, agentID)

	// Emit event
	om.emitEvent(OfficeEvent{
		Type:      EventAgentUnassigned,
		Timestamp: time.Now(),
		AgentID:   agentID,
		RoomID:    assignment.RoomID,
		DeptID:    room.DepartmentID,
	})

	return nil
}

// MoveAgent moves an agent to a new position within their room.
func (om *OfficeManager) MoveAgent(agentID string, newX, newY int) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	assignment, exists := om.assignments[agentID]
	if !exists {
		return fmt.Errorf("agent %s is not assigned to any room", agentID)
	}

	room, err := om.rooms.GetRoom(assignment.RoomID)
	if err != nil {
		return err
	}

	// Check if new position is valid and empty
	cell, err := room.Layout.GetCell(newX, newY)
	if err != nil {
		return err
	}

	if cell.Type != CellTypeEmpty || cell.Occupant != "" {
		return fmt.Errorf("position (%d, %d) is not available", newX, newY)
	}

	// Clear old position
	room.Layout.Grid[assignment.PositionY][assignment.PositionX].Occupant = ""

	// Set new position
	room.Layout.Grid[newY][newX].Occupant = agentID
	room.Layout.Version++

	// Update assignment
	assignment.PositionX = newX
	assignment.PositionY = newY

	// Update position tracking
	om.agentPositions[agentID] = Position{
		X:      newX,
		Y:      newY,
		RoomID: assignment.RoomID,
	}

	// Emit event
	om.emitEvent(OfficeEvent{
		Type:      EventAgentMoved,
		Timestamp: time.Now(),
		AgentID:   agentID,
		RoomID:    assignment.RoomID,
		Data: map[string]interface{}{
			"old_x": assignment.PositionX,
			"old_y": assignment.PositionY,
			"new_x": newX,
			"new_y": newY,
		},
	})

	return nil
}

// GetAgentAssignment returns an agent's current assignment.
func (om *OfficeManager) GetAgentAssignment(agentID string) (*AgentAssignment, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	assignment, exists := om.assignments[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s is not assigned to any room", agentID)
	}

	return assignment, nil
}

// GetAgentPosition returns an agent's current position.
func (om *OfficeManager) GetAgentPosition(agentID string) (Position, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	position, exists := om.agentPositions[agentID]
	if !exists {
		return Position{}, fmt.Errorf("agent %s has no position", agentID)
	}

	return position, nil
}

// GetAgentsInRoom returns all agents assigned to a room.
func (om *OfficeManager) GetAgentsInRoom(roomID string) ([]string, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	room, err := om.rooms.GetRoom(roomID)
	if err != nil {
		return nil, err
	}

	agents := make([]string, len(room.Occupants))
	copy(agents, room.Occupants)

	return agents, nil
}

// GetRoomOccupancyStats returns occupancy statistics for all rooms.
func (om *OfficeManager) GetRoomOccupancyStats() map[string]RoomOccupancyStats {
	om.mu.RLock()
	defer om.mu.RUnlock()

	stats := make(map[string]RoomOccupancyStats)
	for _, room := range om.rooms.ListRooms() {
		occupancy := 0.0
		if room.Capacity > 0 {
			occupancy = float64(len(room.Occupants)) / float64(room.Capacity)
		}
		stats[room.ID] = RoomOccupancyStats{
			RoomID:      room.ID,
			Capacity:    room.Capacity,
			Occupied:    len(room.Occupants),
			Occupancy:   occupancy,
			Available:   room.Capacity - len(room.Occupants),
			DepartmentID: room.DepartmentID,
		}
	}

	return stats
}

// RoomOccupancyStats holds occupancy statistics for a room.
type RoomOccupancyStats struct {
	RoomID       string  `json:"room_id"`
	Capacity     int     `json:"capacity"`
	Occupied     int     `json:"occupied"`
	Occupancy    float64 `json:"occupancy"`
	Available    int     `json:"available"`
	DepartmentID string  `json:"department_id"`
}

// FindBestRoomForAgent finds the most suitable room for an agent based on department.
func (om *OfficeManager) FindBestRoomForAgent(agentID string, deptID string) (*Room, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	// Get rooms in the department
	rooms := om.rooms.GetRoomsByDepartment(deptID)
	if len(rooms) == 0 {
		return nil, fmt.Errorf("no rooms found in department %s", deptID)
	}

	// Find a room with available space
	for _, room := range rooms {
		if !room.IsFull() {
			// Check if there's an empty position
			if _, _, found := room.Layout.FindEmptyPosition(); found {
				return room, nil
			}
		}
	}

	return nil, fmt.Errorf("no available rooms in department %s", deptID)
}

// AutoAssignAgent automatically assigns an agent to the best available room.
func (om *OfficeManager) AutoAssignAgent(agentID string, deptID string) error {
	room, err := om.FindBestRoomForAgent(agentID, deptID)
	if err != nil {
		return err
	}

	return om.AssignAgentToRoom(agentID, room.ID)
}

// RegisterEventHandler registers an event handler.
func (om *OfficeManager) RegisterEventHandler(handler OfficeEventHandler) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.eventHandlers = append(om.eventHandlers, handler)
}

// UnregisterEventHandler removes an event handler.
func (om *OfficeManager) UnregisterEventHandler(handler OfficeEventHandler) {
	om.mu.Lock()
	defer om.mu.Unlock()

	for i, h := range om.eventHandlers {
		// Compare function pointers (this is a simplification)
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			om.eventHandlers = append(om.eventHandlers[:i], om.eventHandlers[i+1:]...)
			break
		}
	}
}

// emitEvent emits an event to all registered handlers.
func (om *OfficeManager) emitEvent(event OfficeEvent) {
	for _, handler := range om.eventHandlers {
		handler(event)
	}
}

// GetOfficeSnapshot returns a complete snapshot of the office state.
func (om *OfficeManager) GetOfficeSnapshot() OfficeSnapshot {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return OfficeSnapshot{
		Config:       om.config,
		Departments:  om.departments.ListDepartments(),
		Rooms:        om.rooms.ListRooms(),
		Assignments:  om.assignments,
		Positions:    om.agentPositions,
		Timestamp:    time.Now(),
	}
}

// OfficeSnapshot represents a complete snapshot of the office state.
type OfficeSnapshot struct {
	Config      OfficeConfig                 `json:"config"`
	Departments []*DepartmentInfo            `json:"departments"`
	Rooms       []*Room                      `json:"rooms"`
	Assignments map[string]*AgentAssignment  `json:"assignments"`
	Positions   map[string]Position          `json:"positions"`
	Timestamp   time.Time                    `json:"timestamp"`
}

// GetTotalAgentCount returns the total number of agents in the office.
func (om *OfficeManager) GetTotalAgentCount() int {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return len(om.assignments)
}

// GetTotalRoomCount returns the total number of rooms in the office.
func (om *OfficeManager) GetTotalRoomCount() int {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return len(om.rooms.ListRooms())
}

// GetTotalDepartmentCount returns the total number of departments.
func (om *OfficeManager) GetTotalDepartmentCount() int {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return len(om.departments.ListDepartments())
}
