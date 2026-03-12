package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"picoclaw/agent/pkg/logger"
)

// Handler contains all HTTP handlers for the UI API
type Handler struct {
	store *MemoryStore
	hub   *Hub
}

// NewHandler creates a new Handler instance
func NewHandler(store *MemoryStore, hub *Hub) *Handler {
	return &Handler{
		store: store,
		hub:   hub,
	}
}

// RegisterRoutes registers all UI API endpoints on the given ServeMux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/ui/office/status", h.GetOfficeStatus)
	mux.HandleFunc("/api/ui/office/stats", h.GetOfficeStats)
	mux.HandleFunc("/api/ui/departments", h.ListDepartments)
	mux.HandleFunc("/api/ui/departments/{id}", h.GetDepartment)
	mux.HandleFunc("/api/ui/rooms", h.ListRooms)
	mux.HandleFunc("/api/ui/rooms/{id}", h.GetRoom)
	mux.HandleFunc("/api/ui/agents", h.ListAgents)
	mux.HandleFunc("/api/ui/agents/{id}", h.GetAgent)
	mux.HandleFunc("PUT /api/ui/agents/{id}/move", h.MoveAgent)
	mux.HandleFunc("PUT /api/ui/agents/{id}/status", h.UpdateAgentStatus)
	mux.HandleFunc("/api/ui/kanban", h.ListKanbanBoards)
	mux.HandleFunc("/api/ui/kanban/board", h.GetKanbanBoard)
	mux.HandleFunc("/api/ui/kanban/tasks", h.ListTasks)
	mux.HandleFunc("POST /api/ui/kanban/tasks", h.CreateTask)
	mux.HandleFunc("PUT /api/ui/kanban/tasks", h.UpdateTask)
	mux.HandleFunc("DELETE /api/ui/kanban/tasks", h.DeleteTask)
	mux.HandleFunc("POST /api/ui/kanban/tasks/move", h.MoveTask)
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.DebugF("ui_handler", map[string]any{
			"error": err.Error(),
		})
	}
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, APIError{
		Code:    code,
		Message: message,
	})
}

// parsePagination parses pagination parameters from the request
func parsePagination(r *http.Request) PaginationParams {
	params := PaginationParams{
		Offset: 0,
		Limit:  50,
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil && val >= 0 {
			params.Offset = val
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil && val > 0 && val <= 100 {
			params.Limit = val
		}
	}

	return params
}

// parseJSON parses JSON from request body
func parseJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

// Office Status Handlers

// GetOfficeStatus returns the overall office status
func (h *Handler) GetOfficeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	status := h.store.GetOfficeStatus()
	writeJSON(w, http.StatusOK, status)
}

// GetOfficeStats returns statistics about the office
func (h *Handler) GetOfficeStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	agents := h.store.ListAgents()
	tasks := h.store.ListTasks()

	activeAgents := 0
	for _, a := range agents {
		if a.Status != AgentStatusIdle && a.Status != AgentStatusOffline {
			activeAgents++
		}
	}

	tasksByStatus := make(map[TaskStatus]int)
	for _, t := range tasks {
		tasksByStatus[t.Status]++
	}

	stats := map[string]any{
		"totalAgents":   len(agents),
		"activeAgents":  activeAgents,
		"totalTasks":    len(tasks),
		"tasksByStatus": tasksByStatus,
	}

	writeJSON(w, http.StatusOK, stats)
}

// Department Handlers

// ListDepartments returns all departments
func (h *Handler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	params := parsePagination(r)
	departments := h.store.ListDepartments()

	// Apply pagination
	total := len(departments)
	start := params.Offset
	if start > total {
		start = total
	}
	end := start + params.Limit
	if end > total {
		end = total
	}

	response := ListDepartmentsResponse{
		Departments: departments[start:end],
		Total:       total,
	}

	writeJSON(w, http.StatusOK, response)
}

// GetDepartment returns a specific department
func (h *Handler) GetDepartment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "Department ID is required")
		return
	}

	dept, ok := h.store.GetDepartment(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Department not found")
		return
	}

	writeJSON(w, http.StatusOK, dept)
}

// Room Handlers

// ListRooms returns all rooms
func (h *Handler) ListRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	params := parsePagination(r)
	rooms := h.store.ListRooms()

	// Filter by department if specified
	deptID := r.URL.Query().Get("department_id")
	if deptID != "" {
		filtered := make([]Room, 0)
		for _, room := range rooms {
			if room.DepartmentID == deptID {
				filtered = append(filtered, room)
			}
		}
		rooms = filtered
	}

	// Filter by type if specified
	roomType := r.URL.Query().Get("type")
	if roomType != "" {
		filtered := make([]Room, 0)
		for _, room := range rooms {
			if string(room.Type) == roomType {
				filtered = append(filtered, room)
			}
		}
		rooms = filtered
	}

	// Apply pagination
	total := len(rooms)
	start := params.Offset
	if start > total {
		start = total
	}
	end := start + params.Limit
	if end > total {
		end = total
	}

	response := ListRoomsResponse{
		Rooms: rooms[start:end],
		Total: total,
	}

	writeJSON(w, http.StatusOK, response)
}

// GetRoom returns a specific room
func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "Room ID is required")
		return
	}

	room, ok := h.store.GetRoom(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Room not found")
		return
	}

	writeJSON(w, http.StatusOK, room)
}

// Agent Handlers

// ListAgents returns all agents
func (h *Handler) ListAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	params := parsePagination(r)
	agents := h.store.ListAgents()

	// Filter by department if specified
	deptID := r.URL.Query().Get("department_id")
	if deptID != "" {
		filtered := make([]Agent, 0)
		for _, agent := range agents {
			if agent.DepartmentID == deptID {
				filtered = append(filtered, agent)
			}
		}
		agents = filtered
	}

	// Filter by status if specified
	status := r.URL.Query().Get("status")
	if status != "" {
		filtered := make([]Agent, 0)
		for _, agent := range agents {
			if string(agent.Status) == status {
				filtered = append(filtered, agent)
			}
		}
		agents = filtered
	}

	// Filter by room if specified
	roomID := r.URL.Query().Get("room_id")
	if roomID != "" {
		filtered := make([]Agent, 0)
		for _, agent := range agents {
			if agent.CurrentRoom == roomID {
				filtered = append(filtered, agent)
			}
		}
		agents = filtered
	}

	// Apply pagination
	total := len(agents)
	start := params.Offset
	if start > total {
		start = total
	}
	end := start + params.Limit
	if end > total {
		end = total
	}

	response := ListAgentsResponse{
		Agents: agents[start:end],
		Total:  total,
	}

	writeJSON(w, http.StatusOK, response)
}

// GetAgent returns a specific agent
func (h *Handler) GetAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "Agent ID is required")
		return
	}

	agent, ok := h.store.GetAgent(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Agent not found")
		return
	}

	writeJSON(w, http.StatusOK, GetAgentResponse{Agent: agent})
}

// MoveAgent moves an agent to a new position or room
func (h *Handler) MoveAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "Agent ID is required")
		return
	}

	var req MoveAgentRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	agent, ok := h.store.GetAgent(id) // Use id from PathValue
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Agent not found")
		return
	}

	oldRoom := agent.CurrentRoom
	oldPosition := agent.Position

	// Update agent position/room
	if req.RoomID != "" {
		agent.CurrentRoom = req.RoomID
	}
	if req.Position.X != 0 || req.Position.Y != 0 {
		agent.Position = req.Position
	}
	agent.LastActive = time.Now()

	// Save to store
	h.store.UpdateAgent(agent)

	// Broadcast event
	if h.hub != nil {
		h.hub.BroadcastAgentMoved(agent.ID, agent.CurrentRoom, oldPosition, agent.Position)
		if oldRoom != agent.CurrentRoom {
			h.hub.BroadcastRoomOccupantsChanged(agent.CurrentRoom, []string{agent.ID}, []string{agent.ID}, nil)
			if oldRoom != "" {
				h.hub.BroadcastRoomOccupantsChanged(oldRoom, []string{}, nil, []string{agent.ID})
			}
		}
	}

	writeJSON(w, http.StatusOK, agent)
}

// UpdateAgentStatus updates an agent's status
func (h *Handler) UpdateAgentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "Agent ID is required")
		return
	}

	var payload struct {
		Status AgentStatus `json:"status"`
		Reason string      `json:"reason,omitempty"`
	}
	if err := parseJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	agent, ok := h.store.GetAgent(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Agent not found")
		return
	}

	oldStatus := agent.Status
	agent.Status = payload.Status
	agent.LastActive = time.Now()

	// Save to store
	h.store.UpdateAgent(agent)

	// Broadcast event
	if h.hub != nil {
		h.hub.BroadcastAgentStatusChanged(agent.ID, oldStatus, agent.Status, payload.Reason)
	}

	writeJSON(w, http.StatusOK, agent)
}

// Kanban Handlers

// ListKanbanBoards returns all kanban boards
func (h *Handler) ListKanbanBoards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	params := parsePagination(r)
	boards := h.store.ListKanbanBoards()

	// Filter by department if specified
	deptID := r.URL.Query().Get("department_id")
	if deptID != "" {
		filtered := make([]KanbanBoard, 0)
		for _, board := range boards {
			if board.DepartmentID == deptID {
				filtered = append(filtered, board)
			}
		}
		boards = filtered
	}

	// Filter by room if specified
	roomID := r.URL.Query().Get("room_id")
	if roomID != "" {
		filtered := make([]KanbanBoard, 0)
		for _, board := range boards {
			if board.RoomID == roomID {
				filtered = append(filtered, board)
			}
		}
		boards = filtered
	}

	// Apply pagination
	total := len(boards)
	start := params.Offset
	if start > total {
		start = total
	}
	end := start + params.Limit
	if end > total {
		end = total
	}

	response := ListKanbanBoardsResponse{
		Boards: boards[start:end],
		Total:  total,
	}

	writeJSON(w, http.StatusOK, response)
}

// GetKanbanBoard returns a specific kanban board
func (h *Handler) GetKanbanBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "Board ID is required")
		return
	}

	board, ok := h.store.GetKanbanBoard(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Board not found")
		return
	}

	writeJSON(w, http.StatusOK, GetKanbanBoardResponse{Board: board})
}

// CreateTask creates a new task
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	var req CreateTaskRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "missing_title", "Task title is required")
		return
	}

	if req.BoardID == "" {
		writeError(w, http.StatusBadRequest, "missing_board_id", "Board ID is required")
		return
	}

	// Validate board exists
	_, ok := h.store.GetKanbanBoard(req.BoardID)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Board not found")
		return
	}

	// Create task
	task := Task{
		ID:          generateTaskID(),
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		AssigneeID:  req.AssigneeID,
		CreatorID:   "system", // TODO: Get from authenticated user
		Priority:    req.Priority,
		Tags:        req.Tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DueDate:     req.DueDate,
	}

	if task.Status == "" {
		task.Status = TaskStatusTodo
	}

	// Save task
	h.store.CreateTask(task)

	// Broadcast event
	if h.hub != nil {
		h.hub.BroadcastTaskCreated(task, task.CreatorID)
	}

	writeJSON(w, http.StatusCreated, task)
}

// UpdateTask updates an existing task
func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "Task ID is required")
		return
	}

	task, ok := h.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Task not found")
		return
	}

	var req UpdateTaskRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	changes := []string{}

	if req.Title != "" && req.Title != task.Title {
		task.Title = req.Title
		changes = append(changes, "title")
	}
	if req.Description != "" && req.Description != task.Description {
		task.Description = req.Description
		changes = append(changes, "description")
	}
	if req.Status != "" && req.Status != task.Status {
		oldStatus := task.Status
		task.Status = req.Status
		changes = append(changes, "status")

		// Broadcast status change separately
		if h.hub != nil {
			h.hub.Broadcast(&Event{
				Type:      EventTaskStatusChanged,
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"task_id":    task.ID,
					"old_status": oldStatus,
					"new_status": task.Status,
				},
			})
		}
	}
	if req.AssigneeID != "" && req.AssigneeID != task.AssigneeID {
		oldAssignee := task.AssigneeID
		task.AssigneeID = req.AssigneeID
		changes = append(changes, "assignee")

		// Broadcast assignment
		if h.hub != nil {
			h.hub.BroadcastTaskAssigned(task.ID, task.AssigneeID, oldAssignee)
		}
	}
	if req.Priority != 0 && req.Priority != task.Priority {
		task.Priority = req.Priority
		changes = append(changes, "priority")
	}
	if len(req.Tags) > 0 {
		task.Tags = req.Tags
		changes = append(changes, "tags")
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
		changes = append(changes, "due_date")
	}

	task.UpdatedAt = time.Now()

	// Save task
	h.store.UpdateTask(task)

	// Broadcast event
	if h.hub != nil && len(changes) > 0 {
		h.hub.BroadcastTaskUpdated(task, changes, "system")
	}

	writeJSON(w, http.StatusOK, task)
}

// DeleteTask deletes a task
func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "Task ID is required")
		return
	}

	_, ok := h.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Task not found")
		return
	}

	// Delete task
	h.store.DeleteTask(id)

	// Broadcast event
	if h.hub != nil {
		h.hub.Broadcast(&Event{
			Type:      EventTaskDeleted,
			Timestamp: time.Now(),
			Payload: map[string]string{
				"task_id": id,
			},
		})
	}

	writeJSON(w, http.StatusNoContent, nil)
}

// MoveTask moves a task to a different column
func (h *Handler) MoveTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	var req MoveTaskRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if req.TaskID == "" {
		writeError(w, http.StatusBadRequest, "missing_task_id", "Task ID is required")
		return
	}

	task, ok := h.store.GetTask(req.TaskID)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Task not found")
		return
	}

	// Update task status based on column
	if req.ToColumnID != "" {
		// Find column and get its status
		board, _ := h.store.GetKanbanBoardByTaskID(req.TaskID)
		if board != nil {
			for _, col := range board.Columns {
				if col.ID == req.ToColumnID {
					oldStatus := task.Status
					task.Status = col.Status
					task.UpdatedAt = time.Now()

					h.store.UpdateTask(task)

					// Broadcast events
					if h.hub != nil {
						h.hub.BroadcastTaskUpdated(task, []string{"status"}, "system")
						h.hub.Broadcast(&Event{
							Type:      EventTaskStatusChanged,
							Timestamp: time.Now(),
							Payload: map[string]interface{}{
								"task_id":    task.ID,
								"old_status": oldStatus,
								"new_status": task.Status,
							},
						})
					}
					break
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, task)
}

// ListTasks returns all tasks
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	params := parsePagination(r)
	tasks := h.store.ListTasks()

	// Filter by board if specified
	boardID := r.URL.Query().Get("board_id")
	if boardID != "" {
		filtered := make([]Task, 0)
		for _, task := range tasks {
			board, _ := h.store.GetKanbanBoardByTaskID(task.ID)
			if board != nil && board.ID == boardID {
				filtered = append(filtered, task)
			}
		}
		tasks = filtered
	}

	// Filter by assignee if specified
	assigneeID := r.URL.Query().Get("assignee_id")
	if assigneeID != "" {
		filtered := make([]Task, 0)
		for _, task := range tasks {
			if task.AssigneeID == assigneeID {
				filtered = append(filtered, task)
			}
		}
		tasks = filtered
	}

	// Filter by status if specified
	status := r.URL.Query().Get("status")
	if status != "" {
		filtered := make([]Task, 0)
		for _, task := range tasks {
			if string(task.Status) == status {
				filtered = append(filtered, task)
			}
		}
		tasks = filtered
	}

	// Apply pagination
	total := len(tasks)
	start := params.Offset
	if start > total {
		start = total
	}
	end := start + params.Limit
	if end > total {
		end = total
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:   tasks[start:end],
		Total:  total,
		Offset: params.Offset,
		Limit:  params.Limit,
	})
}

// generateTaskID generates a unique task ID
func generateTaskID() string {
	return fmt.Sprintf("task_%d_%s", time.Now().Unix(), randomString(6))
}
