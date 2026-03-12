// Package ui provides the API layer for Picoclaw Office UI.
// This package handles REST API endpoints and WebSocket connections
// for real-time visualization of the office environment.
package ui

import (
	"time"
)

// AgentStatus represents the current state of an agent
type AgentStatus string

const (
	AgentStatusIdle       AgentStatus = "idle"
	AgentStatusWorking    AgentStatus = "working"
	AgentStatusBusy       AgentStatus = "busy"
	AgentStatusOffline    AgentStatus = "offline"
	AgentStatusError      AgentStatus = "error"
	AgentStatusMigrating  AgentStatus = "migrating"
)

// TaskStatus represents the state of a kanban task
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusBlocked    TaskStatus = "blocked"
)

// RoomType represents the type of office room
type RoomType string

const (
	RoomTypeOpenOffice   RoomType = "open_office"
	RoomTypeMeeting      RoomType = "meeting"
	RoomTypeFocus        RoomType = "focus"
	RoomTypeBreakout     RoomType = "breakout"
	RoomTypeReception    RoomType = "reception"
)

// Department represents an organizational department
type Department struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Color       string   `json:"color,omitempty"`
	AgentIDs    []string `json:"agent_ids"`
	RoomIDs     []string `json:"room_ids"`
}

// Room represents a physical or virtual space in the office
type Room struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Type         RoomType    `json:"type"`
	DepartmentID string      `json:"department_id,omitempty"`
	Capacity     int         `json:"capacity"`
	Occupants    []string    `json:"occupants"` // Agent IDs currently in room
	Position     Position    `json:"position"`  // 2D coordinates for UI
	Size         Size        `json:"size"`      // Width/height for UI
	Metadata     RoomMetadata `json:"metadata,omitempty"`
}

// Position represents 2D coordinates
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Size represents dimensions
type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// RoomMetadata contains additional room information
type RoomMetadata struct {
	Icon        string            `json:"icon,omitempty"`
	Color       string            `json:"color,omitempty"`
	Features    []string          `json:"features,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// Agent represents an agent in the office
type Agent struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	DepartmentID string      `json:"department_id,omitempty"`
	Role         string      `json:"role,omitempty"`
	Status       AgentStatus `json:"status"`
	CurrentRoom  string      `json:"current_room,omitempty"`
	Position     Position    `json:"position"`
	Avatar       string      `json:"avatar,omitempty"`
	Capabilities []string    `json:"capabilities"`
	Model        string      `json:"model,omitempty"`
	IsOnline     bool        `json:"is_online"`
	LastActive   time.Time   `json:"last_active"`
	CurrentTask  *Task       `json:"current_task,omitempty"`
}

// Task represents a kanban board task
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	AssigneeID  string     `json:"assignee_id,omitempty"`
	CreatorID   string     `json:"creator_id"`
	Priority    int        `json:"priority"` // 1-5, higher is more important
	Tags        []string   `json:"tags,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	RoomID      string     `json:"room_id,omitempty"`
}

// KanbanBoard represents a project board with tasks
type KanbanBoard struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	DepartmentID string    `json:"department_id,omitempty"`
	RoomID      string     `json:"room_id,omitempty"`
	Columns     []Column   `json:"columns"`
	Tasks       []Task     `json:"tasks"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Column represents a kanban column
type Column struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Status   TaskStatus `json:"status"`
	Order    int        `json:"order"`
	TaskIDs  []string   `json:"task_ids"`
	WIPLimit int        `json:"wip_limit,omitempty"` // Work in progress limit
}

// OfficeStatus represents the overall office state
type OfficeStatus struct {
	Agents       []Agent       `json:"agents"`
	Departments  []Department  `json:"departments"`
	Rooms        []Room        `json:"rooms"`
	ActiveTasks  int           `json:"active_tasks"`
	TotalTasks   int           `json:"total_tasks"`
	OnlineAgents int           `json:"online_agents"`
	TotalAgents  int           `json:"total_agents"`
	Timestamp    time.Time     `json:"timestamp"`
}

// EventType represents WebSocket event types
type EventType string

const (
	// Agent events
	EventAgentMoved           EventType = "agent_moved"
	EventAgentStatusChanged   EventType = "agent_status_changed"
	EventAgentJoined          EventType = "agent_joined"
	EventAgentLeft            EventType = "agent_left"
	EventAgentRoomChanged     EventType = "agent_room_changed"

	// Task events
	EventTaskCreated          EventType = "task_created"
	EventTaskUpdated          EventType = "task_updated"
	EventTaskDeleted          EventType = "task_deleted"
	EventTaskAssigned         EventType = "task_assigned"
	EventTaskStatusChanged    EventType = "task_status_changed"

	// Room events
	EventRoomUpdated          EventType = "room_updated"
	EventRoomOccupantsChanged EventType = "room_occupants_changed"

	// Department events
	EventDepartmentUpdated    EventType = "department_updated"

	// System events
	EventSystemStatus         EventType = "system_status"
	EventError                EventType = "error"
	EventPing                 EventType = "ping"
	EventPong                 EventType = "pong"
)

// Event represents a WebSocket event
type Event struct {
	Type      EventType       `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   interface{}     `json:"payload"`
}

// EventPayloads for specific event types

// AgentMovedPayload represents agent movement
type AgentMovedPayload struct {
	AgentID   string   `json:"agent_id"`
	RoomID    string   `json:"room_id"`
	From      Position `json:"from"`
	To        Position `json:"to"`
	Timestamp time.Time `json:"timestamp"`
}

// AgentStatusChangedPayload represents status change
type AgentStatusChangedPayload struct {
	AgentID   string      `json:"agent_id"`
	OldStatus AgentStatus `json:"old_status"`
	NewStatus AgentStatus `json:"new_status"`
	Reason    string      `json:"reason,omitempty"`
}

// TaskCreatedPayload represents new task
type TaskCreatedPayload struct {
	Task      Task   `json:"task"`
	CreatorID string `json:"creator_id"`
}

// TaskUpdatedPayload represents task update
type TaskUpdatedPayload struct {
	Task      Task      `json:"task"`
	Changes   []string  `json:"changes"` // List of changed fields
	UpdatedBy string    `json:"updated_by"`
}

// TaskAssignedPayload represents task assignment
type TaskAssignedPayload struct {
	TaskID     string `json:"task_id"`
	AssigneeID string `json:"assignee_id"`
	AssignerID string `json:"assigner_id"`
}

// RoomOccupantsChangedPayload represents room occupancy change
type RoomOccupantsChangedPayload struct {
	RoomID      string   `json:"room_id"`
	Occupants   []string `json:"occupants"`
	Joined      []string `json:"joined,omitempty"`
	Left        []string `json:"left,omitempty"`
}

// SystemStatusPayload represents system status update
type SystemStatusPayload struct {
	OnlineAgents  int       `json:"online_agents"`
	ActiveTasks   int       `json:"active_tasks"`
	MemoryUsage   uint64    `json:"memory_usage"`
	Uptime        string    `json:"uptime"`
}

// ErrorPayload represents an error event
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// REST API Request/Response Types

// ListAgentsResponse represents the response for listing agents
type ListAgentsResponse struct {
	Agents []Agent `json:"agents"`
	Total  int     `json:"total"`
}

// GetAgentResponse represents the response for getting a single agent
type GetAgentResponse struct {
	Agent Agent `json:"agent"`
}

// ListDepartmentsResponse represents the response for listing departments
type ListDepartmentsResponse struct {
	Departments []Department `json:"departments"`
	Total       int          `json:"total"`
}

// ListRoomsResponse represents the response for listing rooms
type ListRoomsResponse struct {
	Rooms []Room `json:"rooms"`
	Total int    `json:"total"`
}

// ListKanbanBoardsResponse represents the response for listing boards
type ListKanbanBoardsResponse struct {
	Boards []KanbanBoard `json:"boards"`
	Total  int           `json:"total"`
}

// GetKanbanBoardResponse represents the response for getting a board
type GetKanbanBoardResponse struct {
	Board KanbanBoard `json:"board"`
}

// CreateTaskRequest represents the request to create a task
type CreateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	AssigneeID  string     `json:"assignee_id,omitempty"`
	Priority    int        `json:"priority,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	BoardID     string     `json:"board_id"`
}

// UpdateTaskRequest represents the request to update a task
type UpdateTaskRequest struct {
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	Status      TaskStatus `json:"status,omitempty"`
	AssigneeID  string     `json:"assignee_id,omitempty"`
	Priority    int        `json:"priority,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// MoveTaskRequest represents the request to move a task between columns
type MoveTaskRequest struct {
	TaskID     string `json:"task_id"`
	ToColumnID string `json:"to_column_id"`
	Position   int    `json:"position,omitempty"` // Position within column
}

// MoveAgentRequest represents the request to move an agent
type MoveAgentRequest struct {
	AgentID  string   `json:"agent_id"`
	RoomID   string   `json:"room_id,omitempty"`
	Position Position `json:"position,omitempty"`
}

// APIError represents an API error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface
func (e APIError) Error() string {
	return e.Message
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// SubscriptionRequest represents a request to subscribe to events
type SubscriptionRequest struct {
	Events []EventType `json:"events"`
	Filter Filter      `json:"filter,omitempty"`
}

// Filter represents event filtering options
type Filter struct {
	AgentIDs     []string `json:"agent_ids,omitempty"`
	RoomIDs      []string `json:"room_ids,omitempty"`
	DepartmentIDs []string `json:"department_ids,omitempty"`
	TaskStatuses []string `json:"task_statuses,omitempty"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data   interface{} `json:"data"`
	Total  int         `json:"total"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
}
