package ui

import (
	"sync"
	"time"

	"picoclaw/agent/pkg/office"
)

// MemoryStore provides in-memory storage for UI data
// This can be replaced with a persistent store in production
type MemoryStore struct {
	agents      map[string]Agent
	departments map[string]Department
	rooms       map[string]Room
	tasks       map[string]Task
	boards      map[string]KanbanBoard
	mu          sync.RWMutex

	// OfficeManager integration
	officeManager *office.OfficeManager
}

// NewMemoryStore creates a new memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		agents:      make(map[string]Agent),
		departments: make(map[string]Department),
		rooms:       make(map[string]Room),
		tasks:       make(map[string]Task),
		boards:      make(map[string]KanbanBoard),
	}
}

// SetOfficeManager sets the office manager for integration
func (s *MemoryStore) SetOfficeManager(om *office.OfficeManager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.officeManager = om
}

// InitializeWithDefaults populates the store with default data
func (s *MemoryStore) InitializeWithDefaults() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create default departments
	depts := []Department{
		{
			ID:          "engineering",
			Name:        "Engineering",
			Description: "Software development and technical operations",
			Color:       "#3b82f6",
			AgentIDs:    []string{},
			RoomIDs:     []string{"eng-main", "eng-meeting"},
		},
		{
			ID:          "design",
			Name:        "Design",
			Description: "UI/UX design and creative work",
			Color:       "#8b5cf6",
			AgentIDs:    []string{},
			RoomIDs:     []string{"design-studio"},
		},
		{
			ID:          "product",
			Name:        "Product",
			Description: "Product management and strategy",
			Color:       "#10b981",
			AgentIDs:    []string{},
			RoomIDs:     []string{"product-room"},
		},
	}

	for _, dept := range depts {
		s.departments[dept.ID] = dept
	}

	// Create default rooms
	rooms := []Room{
		{
			ID:           "eng-main",
			Name:         "Engineering Main",
			Type:         RoomTypeOpenOffice,
			DepartmentID: "engineering",
			Capacity:     20,
			Occupants:    []string{},
			Position:     Position{X: 100, Y: 100},
			Size:         Size{Width: 400, Height: 300},
			Metadata: RoomMetadata{
				Icon:     "code",
				Color:    "#3b82f6",
				Features: []string{"desks", "whiteboard", "monitors"},
			},
		},
		{
			ID:           "eng-meeting",
			Name:         "Engineering Meeting Room",
			Type:         RoomTypeMeeting,
			DepartmentID: "engineering",
			Capacity:     8,
			Occupants:    []string{},
			Position:     Position{X: 520, Y: 100},
			Size:         Size{Width: 200, Height: 150},
			Metadata: RoomMetadata{
				Icon:     "users",
				Color:    "#3b82f6",
				Features: []string{"conference_table", "projector", "video_call"},
			},
		},
		{
			ID:           "design-studio",
			Name:         "Design Studio",
			Type:         RoomTypeFocus,
			DepartmentID: "design",
			Capacity:     10,
			Occupants:    []string{},
			Position:     Position{X: 100, Y: 420},
			Size:         Size{Width: 300, Height: 200},
			Metadata: RoomMetadata{
				Icon:     "palette",
				Color:    "#8b5cf6",
				Features: []string{"drawing_tablets", "color_printer", "mood_boards"},
			},
		},
		{
			ID:           "product-room",
			Name:         "Product Room",
			Type:         RoomTypeMeeting,
			DepartmentID: "product",
			Capacity:     6,
			Occupants:    []string{},
			Position:     Position{X: 420, Y: 420},
			Size:         Size{Width: 180, Height: 150},
			Metadata: RoomMetadata{
				Icon:     "briefcase",
				Color:    "#10b981",
				Features: []string{"whiteboard", "analytics_screens"},
			},
		},
		{
			ID:        "reception",
			Name:      "Reception",
			Type:      RoomTypeReception,
			Capacity:  5,
			Occupants: []string{},
			Position:  Position{X: 10, Y: 10},
			Size:      Size{Width: 80, Height: 80},
			Metadata: RoomMetadata{
				Icon:     "home",
				Color:    "#f59e0b",
				Features: []string{"reception_desk", "waiting_area"},
			},
		},
	}

	for _, room := range rooms {
		s.rooms[room.ID] = room
	}

	// Create default kanban board
	board := KanbanBoard{
		ID:           "main-board",
		Name:         "Main Project Board",
		Description:  "Primary kanban board for tracking all tasks",
		DepartmentID: "",
		RoomID:       "reception",
		Columns: []Column{
			{ID: "col-todo", Name: "To Do", Status: TaskStatusTodo, Order: 0, TaskIDs: []string{}, WIPLimit: 0},
			{ID: "col-inprogress", Name: "In Progress", Status: TaskStatusInProgress, Order: 1, TaskIDs: []string{}, WIPLimit: 5},
			{ID: "col-review", Name: "Review", Status: TaskStatusReview, Order: 2, TaskIDs: []string{}, WIPLimit: 3},
			{ID: "col-done", Name: "Done", Status: TaskStatusDone, Order: 3, TaskIDs: []string{}, WIPLimit: 0},
		},
		Tasks:     []Task{},
		CreatedAt: time.Now(),
	}

	s.boards[board.ID] = board
}

// Agent methods

func (s *MemoryStore) GetAgent(id string) (Agent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	agent, ok := s.agents[id]
	return agent, ok
}

func (s *MemoryStore) ListAgents() []Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	agents := make([]Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}
	return agents
}

func (s *MemoryStore) CreateAgent(agent Agent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agents[agent.ID] = agent

	// Update department's agent list
	if agent.DepartmentID != "" {
		if dept, ok := s.departments[agent.DepartmentID]; ok {
			if !containsString(dept.AgentIDs, agent.ID) {
				dept.AgentIDs = append(dept.AgentIDs, agent.ID)
				s.departments[agent.DepartmentID] = dept
			}
		}
	}

	// Update room's occupants
	if agent.CurrentRoom != "" {
		if room, ok := s.rooms[agent.CurrentRoom]; ok {
			if !containsString(room.Occupants, agent.ID) {
				room.Occupants = append(room.Occupants, agent.ID)
				s.rooms[agent.CurrentRoom] = room
			}
		}
	}
}

func (s *MemoryStore) UpdateAgent(agent Agent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldAgent, exists := s.agents[agent.ID]
	s.agents[agent.ID] = agent

	// Handle room change
	if exists && oldAgent.CurrentRoom != agent.CurrentRoom {
		// Remove from old room
		if oldAgent.CurrentRoom != "" {
			if room, ok := s.rooms[oldAgent.CurrentRoom]; ok {
				room.Occupants = removeString(room.Occupants, agent.ID)
				s.rooms[oldAgent.CurrentRoom] = room
			}
		}
		// Add to new room
		if agent.CurrentRoom != "" {
			if room, ok := s.rooms[agent.CurrentRoom]; ok {
				if !containsString(room.Occupants, agent.ID) {
					room.Occupants = append(room.Occupants, agent.ID)
					s.rooms[agent.CurrentRoom] = room
				}
			}
		}
	}

	// Handle department change
	if exists && oldAgent.DepartmentID != agent.DepartmentID {
		// Remove from old department
		if oldAgent.DepartmentID != "" {
			if dept, ok := s.departments[oldAgent.DepartmentID]; ok {
				dept.AgentIDs = removeString(dept.AgentIDs, agent.ID)
				s.departments[oldAgent.DepartmentID] = dept
			}
		}
		// Add to new department
		if agent.DepartmentID != "" {
			if dept, ok := s.departments[agent.DepartmentID]; ok {
				if !containsString(dept.AgentIDs, agent.ID) {
					dept.AgentIDs = append(dept.AgentIDs, agent.ID)
					s.departments[agent.DepartmentID] = dept
				}
			}
		}
	}
}

func (s *MemoryStore) DeleteAgent(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.agents[id]
	if !exists {
		return
	}

	// Remove from department
	if agent.DepartmentID != "" {
		if dept, ok := s.departments[agent.DepartmentID]; ok {
			dept.AgentIDs = removeString(dept.AgentIDs, id)
			s.departments[agent.DepartmentID] = dept
		}
	}

	// Remove from room
	if agent.CurrentRoom != "" {
		if room, ok := s.rooms[agent.CurrentRoom]; ok {
			room.Occupants = removeString(room.Occupants, id)
			s.rooms[agent.CurrentRoom] = room
		}
	}

	delete(s.agents, id)
}

// Department methods

func (s *MemoryStore) GetDepartment(id string) (Department, bool) {
	// Try OfficeManager first
	if s.officeManager != nil {
		deptData, err := s.officeManager.GetDepartmentManager().GetDepartment(id)
		if err == nil && deptData != nil {
			return Department{
				ID:          deptData.ID,
				Name:        deptData.Name,
				Description: deptData.Description,
				Color:       deptData.Color,
				AgentIDs:    []string{}, // Will be populated from agent data
				RoomIDs:     []string{}, // Will be populated from room data
			}, true
		}
	}

	// Fallback to memory store
	s.mu.RLock()
	defer s.mu.RUnlock()
	dept, ok := s.departments[id]
	return dept, ok
}

func (s *MemoryStore) ListDepartments() []Department {
	// Try OfficeManager first - currently using memory store fallback
	// TODO: Enhance DepartmentManager to return full department data with IDs
	_ = s.officeManager // Reference to avoid unused field warning

	// Fallback to memory store
	s.mu.RLock()
	defer s.mu.RUnlock()
	depts := make([]Department, 0, len(s.departments))
	for _, dept := range s.departments {
		depts = append(depts, dept)
	}
	return depts
}

func (s *MemoryStore) CreateDepartment(dept Department) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.departments[dept.ID] = dept
}

func (s *MemoryStore) UpdateDepartment(dept Department) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.departments[dept.ID] = dept
}

func (s *MemoryStore) DeleteDepartment(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.departments, id)
}

// Room methods

func (s *MemoryStore) GetRoom(id string) (Room, bool) {
	// Try OfficeManager first
	if s.officeManager != nil {
		roomData, err := s.officeManager.GetRoomManager().GetRoom(id)
		if err == nil && roomData != nil {
			s.mu.Lock()
			defer s.mu.Unlock()
			return s.convertOfficeRoomToUIRoom(roomData), true
		}
	}

	// Fallback to memory store
	s.mu.RLock()
	defer s.mu.RUnlock()
	room, ok := s.rooms[id]
	return room, ok
}

func (s *MemoryStore) ListRooms() []Room {
	// Try OfficeManager first
	if s.officeManager != nil {
		s.mu.Lock()
		defer s.mu.Unlock()

		roomDataList := s.officeManager.GetRoomManager().ListRooms()
		rooms := make([]Room, 0, len(roomDataList))
		for _, roomData := range roomDataList {
			rooms = append(rooms, s.convertOfficeRoomToUIRoom(roomData))
		}
		return rooms
	}

	// Fallback to memory store
	s.mu.RLock()
	defer s.mu.RUnlock()
	rooms := make([]Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// roomPositionCounter is used to assign unique positions to rooms
var roomPositionCounter = 0

// resetRoomPositionCounter resets the counter (call when reinitializing)
func resetRoomPositionCounter() {
	roomPositionCounter = 0
}

// getNextRoomPosition returns the next available room position in a grid layout
func getNextRoomPosition() (x, y float64) {
	const (
		startX   = 50
		startY   = 50
		spacingX = 220
		spacingY = 180
		cols     = 4
	)

	col := roomPositionCounter % cols
	row := roomPositionCounter / cols

	x = float64(startX + col*spacingX)
	y = float64(startY + row*spacingY)

	roomPositionCounter++
	return x, y
}

// convertOfficeRoomToUIRoom converts office.Room to ui.Room
func (s *MemoryStore) convertOfficeRoomToUIRoom(roomData *office.Room) Room {
	// Convert room type
	var roomType RoomType
	switch roomData.Type {
	case office.RoomTypeWorkspace:
		roomType = RoomTypeOpenOffice
	case office.RoomTypeMeeting:
		roomType = RoomTypeMeeting
	case office.RoomTypeFocus:
		roomType = RoomTypeFocus
	case office.RoomTypeCollaboration:
		roomType = RoomTypeOpenOffice
	default:
		roomType = RoomTypeOpenOffice
	}

	// Check memory store for existing UI data
	var position Position
	existingRoom, exists := s.rooms[roomData.ID]
	if exists {
		position = existingRoom.Position
	} else {
		posX, posY := getNextRoomPosition()
		position = Position{
			X: posX,
			Y: posY,
		}
	}

	const gridSize = 32
	size := Size{
		Width:  float64(roomData.Layout.Width * gridSize),
		Height: float64(roomData.Layout.Height * gridSize),
	}

	// Get theme from metadata if available
	theme := "default"
	if roomData.Metadata != nil {
		if t, ok := roomData.Metadata["theme"].(string); ok {
			theme = t
		}
	}

	// Determine color based on department
	color := "#3b82f6" // Default blue
	switch roomData.DepartmentID {
	case "planning":
		color = "#8b5cf6" // Purple
	case "coding":
		color = "#3b82f6" // Blue
	case "design":
		color = "#ec4899" // Pink
	case "marketing":
		color = "#f59e0b" // Amber
	case "quality":
		color = "#10b981" // Green
	case "legal":
		color = "#ef4444" // Red
	}

	newRoom := Room{
		ID:           roomData.ID,
		Name:         roomData.Name,
		Type:         roomType,
		DepartmentID: roomData.DepartmentID,
		Capacity:     roomData.Capacity,
		Occupants:    roomData.Occupants,
		Position:     position,
		Size:         size,
		Metadata: RoomMetadata{
			Icon:     theme,
			Color:    color,
			Features: []string{},
		},
	}

	// Save back to memory store to persist position
	if !exists {
		s.rooms[roomData.ID] = newRoom
	}

	return newRoom
}

func (s *MemoryStore) CreateRoom(room Room) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms[room.ID] = room

	// Update department's room list
	if room.DepartmentID != "" {
		if dept, ok := s.departments[room.DepartmentID]; ok {
			if !containsString(dept.RoomIDs, room.ID) {
				dept.RoomIDs = append(dept.RoomIDs, room.ID)
				s.departments[room.DepartmentID] = dept
			}
		}
	}
}

func (s *MemoryStore) UpdateRoom(room Room) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldRoom, exists := s.rooms[room.ID]
	s.rooms[room.ID] = room

	// Handle department change
	if exists && oldRoom.DepartmentID != room.DepartmentID {
		// Remove from old department
		if oldRoom.DepartmentID != "" {
			if dept, ok := s.departments[oldRoom.DepartmentID]; ok {
				dept.RoomIDs = removeString(dept.RoomIDs, room.ID)
				s.departments[oldRoom.DepartmentID] = dept
			}
		}
		// Add to new department
		if room.DepartmentID != "" {
			if dept, ok := s.departments[room.DepartmentID]; ok {
				if !containsString(dept.RoomIDs, room.ID) {
					dept.RoomIDs = append(dept.RoomIDs, room.ID)
					s.departments[room.DepartmentID] = dept
				}
			}
		}
	}
}

func (s *MemoryStore) DeleteRoom(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, exists := s.rooms[id]
	if !exists {
		return
	}

	// Remove from department
	if room.DepartmentID != "" {
		if dept, ok := s.departments[room.DepartmentID]; ok {
			dept.RoomIDs = removeString(dept.RoomIDs, id)
			s.departments[room.DepartmentID] = dept
		}
	}

	// Clear agents from this room
	for agentID := range s.agents {
		if agent := s.agents[agentID]; agent.CurrentRoom == id {
			agent.CurrentRoom = ""
			s.agents[agentID] = agent
		}
	}

	delete(s.rooms, id)
}

// Task methods

func (s *MemoryStore) GetTask(id string) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	return task, ok
}

func (s *MemoryStore) ListTasks() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (s *MemoryStore) CreateTask(task Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
}

func (s *MemoryStore) UpdateTask(task Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
}

func (s *MemoryStore) DeleteTask(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
}

// Kanban Board methods

func (s *MemoryStore) GetKanbanBoard(id string) (KanbanBoard, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	board, ok := s.boards[id]
	return board, ok
}

func (s *MemoryStore) GetKanbanBoardByTaskID(taskID string) (*KanbanBoard, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, board := range s.boards {
		for _, task := range board.Tasks {
			if task.ID == taskID {
				return &board, true
			}
		}
	}
	return nil, false
}

func (s *MemoryStore) ListKanbanBoards() []KanbanBoard {
	s.mu.RLock()
	defer s.mu.RUnlock()
	boards := make([]KanbanBoard, 0, len(s.boards))
	for _, board := range s.boards {
		// Load tasks for each board
		board.Tasks = make([]Task, 0)
		for _, task := range s.tasks {
			// In a real implementation, tasks would reference their board
			// For now, include all tasks
			board.Tasks = append(board.Tasks, task)
		}
		boards = append(boards, board)
	}
	return boards
}

func (s *MemoryStore) CreateKanbanBoard(board KanbanBoard) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.boards[board.ID] = board
}

func (s *MemoryStore) UpdateKanbanBoard(board KanbanBoard) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.boards[board.ID] = board
}

func (s *MemoryStore) DeleteKanbanBoard(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.boards, id)
}

// Office Status

func (s *MemoryStore) GetOfficeStatus() OfficeStatus {
	// Use OfficeManager data if available
	if s.officeManager != nil {
		s.mu.Lock()
		defer s.mu.Unlock()

		// Get rooms from OfficeManager
		roomDataList := s.officeManager.GetRoomManager().ListRooms()
		rooms := make([]Room, 0, len(roomDataList))
		for _, roomData := range roomDataList {
			rooms = append(rooms, s.convertOfficeRoomToUIRoom(roomData))
		}

		// For departments, agents, and tasks, still use memory store
		// TODO: Integrate with OfficeManager for departments and agents

		agents := make([]Agent, 0, len(s.agents))
		onlineCount := 0
		for _, agent := range s.agents {
			if assign, err := s.officeManager.GetAgentAssignment(agent.ID); err == nil && assign != nil {
				agent.CurrentRoom = assign.RoomID
				// Compute global pixel position based on the room's UI bounds
				for _, r := range rooms {
					if r.ID == assign.RoomID {
						agent.Position.X = r.Position.X + float64(assign.PositionX*32)
						agent.Position.Y = r.Position.Y + float64(assign.PositionY*32)
						break
					}
				}
			}

			agents = append(agents, agent)
			if agent.IsOnline {
				onlineCount++
			}
		}

		departments := make([]Department, 0, len(s.departments))
		for _, dept := range s.departments {
			departments = append(departments, dept)
		}

		activeTasks := 0
		for _, task := range s.tasks {
			if task.Status != TaskStatusDone {
				activeTasks++
			}
		}

		return OfficeStatus{
			Agents:       agents,
			Departments:  departments,
			Rooms:        rooms,
			ActiveTasks:  activeTasks,
			TotalTasks:   len(s.tasks),
			OnlineAgents: onlineCount,
			TotalAgents:  len(s.agents),
			Timestamp:    time.Now(),
		}
	}

	// Fallback to memory store only
	s.mu.RLock()
	defer s.mu.RUnlock()

	agents := make([]Agent, 0, len(s.agents))
	onlineCount := 0
	for _, agent := range s.agents {
		agents = append(agents, agent)
		if agent.IsOnline {
			onlineCount++
		}
	}

	departments := make([]Department, 0, len(s.departments))
	for _, dept := range s.departments {
		departments = append(departments, dept)
	}

	rooms := make([]Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}

	activeTasks := 0
	for _, task := range s.tasks {
		if task.Status != TaskStatusDone {
			activeTasks++
		}
	}

	return OfficeStatus{
		Agents:       agents,
		Departments:  departments,
		Rooms:        rooms,
		ActiveTasks:  activeTasks,
		TotalTasks:   len(s.tasks),
		OnlineAgents: onlineCount,
		TotalAgents:  len(s.agents),
		Timestamp:    time.Now(),
	}
}

// Helper functions

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeString(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
