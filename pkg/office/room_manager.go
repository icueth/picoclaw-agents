// Package office provides Office UI functionality for Picoclaw agent management
package office

import (
	"fmt"
	"sync"
	"time"
)

// RoomType represents the type of room
type RoomType string

const (
	RoomTypeWorkspace       RoomType = "workspace"
	RoomTypeMeeting         RoomType = "meeting"
	RoomTypeFocus           RoomType = "focus"
	RoomTypeCollaboration   RoomType = "collaboration"
	RoomTypeBreak           RoomType = "break"
)

// GridSize represents the size of a room grid
type GridSize string

const (
	GridSizeSmall  GridSize = "small"
	GridSizeMedium GridSize = "medium"
	GridSizeLarge  GridSize = "large"
)

// GridCell represents a cell in the room grid
type GridCell struct {
	Type     CellType `json:"type"`
	Occupant string   `json:"occupant,omitempty"`
}

// CellType represents the type of cell in the grid
type CellType string

const (
	CellTypeEmpty    CellType = "empty"
	CellTypeWall     CellType = "wall"
	CellTypeDesk     CellType = "desk"
	CellTypeDoor     CellType = "door"
	CellTypeWindow   CellType = "window"
)

// RoomLayout represents the spatial layout of a room
type RoomLayout struct {
	Width   int         `json:"width"`
	Height  int         `json:"height"`
	Grid    [][]GridCell `json:"grid"`
	Version int         `json:"version"`
}

// FindEmptyPosition finds an empty position in the room
func (rl *RoomLayout) FindEmptyPosition() (x, y int, found bool) {
	if rl == nil || rl.Grid == nil {
		return 0, 0, false
	}

	for y := 0; y < rl.Height; y++ {
		for x := 0; x < rl.Width; x++ {
			if y < len(rl.Grid) && x < len(rl.Grid[y]) &&
				rl.Grid[y][x].Type == CellTypeEmpty && rl.Grid[y][x].Occupant == "" {
				return x, y, true
			}
		}
	}

	return 0, 0, false
}

// GetCell returns the cell at the specified coordinates
func (rl *RoomLayout) GetCell(x, y int) (*GridCell, error) {
	if rl == nil || rl.Grid == nil {
		return nil, fmt.Errorf("layout is nil")
	}

	if x < 0 || x >= rl.Width || y < 0 || y >= rl.Height {
		return nil, fmt.Errorf("coordinates (%d, %d) out of bounds", x, y)
	}

	if y >= len(rl.Grid) || x >= len(rl.Grid[y]) {
		return nil, fmt.Errorf("coordinates (%d, %d) out of grid bounds", x, y)
	}

	return &rl.Grid[y][x], nil
}

// ClearOccupant removes an occupant from the grid
func (rl *RoomLayout) ClearOccupant(agentID string) {
	if rl == nil || rl.Grid == nil {
		return
	}

	for y := 0; y < rl.Height; y++ {
		for x := 0; x < rl.Width; x++ {
			if y < len(rl.Grid) && x < len(rl.Grid[y]) && rl.Grid[y][x].Occupant == agentID {
				rl.Grid[y][x].Occupant = ""
				rl.Version++
				return
			}
		}
	}
}

// Room represents a room in the office
type Room struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	DepartmentID string         `json:"department_id"`
	Type         RoomType       `json:"type"`
	Capacity     int            `json:"capacity"`
	GridSize     GridSize       `json:"grid_size"`
	Layout       *RoomLayout    `json:"layout"`
	Occupants    []string       `json:"occupants"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// IsFull returns true if the room is at capacity
func (r *Room) IsFull() bool {
	return len(r.Occupants) >= r.Capacity
}

// GetOccupancy returns the current occupancy count
func (r *Room) GetOccupancy() int {
	return len(r.Occupants)
}

// RoomManager manages rooms in the office
type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

// NewRoomManager creates a new room manager
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

// CreateRoom creates a new room
func (rm *RoomManager) CreateRoom(id, name, deptID string, roomType RoomType,
	capacity int, gridSize GridSize) (*Room, error) {

	rm.mu.Lock()
	defer rm.mu.Unlock()

	if id == "" {
		return nil, fmt.Errorf("room ID is required")
	}

	if _, exists := rm.rooms[id]; exists {
		return nil, fmt.Errorf("room %s already exists", id)
	}

	// Determine grid dimensions based on size
	width, height := getGridDimensions(gridSize)

	// Initialize grid
	grid := make([][]GridCell, height)
	for i := range grid {
		grid[i] = make([]GridCell, width)
		for j := range grid[i] {
			grid[i][j] = GridCell{Type: CellTypeEmpty}
		}
	}

	room := &Room{
		ID:           id,
		Name:         name,
		DepartmentID: deptID,
		Type:         roomType,
		Capacity:     capacity,
		GridSize:     gridSize,
		Layout: &RoomLayout{
			Width:   width,
			Height:  height,
			Grid:    grid,
			Version: 1,
		},
		Occupants: make([]string, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]any),
	}

	rm.rooms[id] = room
	return room, nil
}

// GetRoom retrieves a room by ID
func (rm *RoomManager) GetRoom(id string) (*Room, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[id]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", id)
	}

	return room, nil
}

// UpdateRoom updates a room
func (rm *RoomManager) UpdateRoom(id string, updates map[string]any) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[id]
	if !exists {
		return fmt.Errorf("room not found: %s", id)
	}

	if name, ok := updates["name"].(string); ok {
		room.Name = name
	}
	if capacity, ok := updates["capacity"].(int); ok {
		room.Capacity = capacity
	}

	room.UpdatedAt = time.Now()
	return nil
}

// DeleteRoom deletes a room
func (rm *RoomManager) DeleteRoom(id string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.rooms[id]; !exists {
		return fmt.Errorf("room not found: %s", id)
	}

	delete(rm.rooms, id)
	return nil
}

// ListRooms returns all rooms
func (rm *RoomManager) ListRooms() []*Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	list := make([]*Room, 0, len(rm.rooms))
	for _, room := range rm.rooms {
		list = append(list, room)
	}

	return list
}

// ListRoomsByDepartment returns rooms for a specific department
func (rm *RoomManager) ListRoomsByDepartment(deptID string) []*Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	list := make([]*Room, 0)
	for _, room := range rm.rooms {
		if room.DepartmentID == deptID {
			list = append(list, room)
		}
	}

	return list
}

// GetRoomsByDepartment returns rooms for a specific department (alias for ListRoomsByDepartment)
func (rm *RoomManager) GetRoomsByDepartment(deptID string) []*Room {
	return rm.ListRoomsByDepartment(deptID)
}

// getGridDimensions returns width and height for a grid size
func getGridDimensions(size GridSize) (width, height int) {
	switch size {
	case GridSizeSmall:
		return 4, 4
	case GridSizeMedium:
		return 6, 6
	case GridSizeLarge:
		return 8, 8
	default:
		return 6, 6
	}
}
