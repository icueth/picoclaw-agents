// Package office provides Office UI functionality for Picoclaw agent management
package office

import (
	"fmt"
	"sync"
	"time"
)

// DepartmentConfig contains configuration for creating a department
type DepartmentConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// DepartmentData represents a department in the office (internal structure)
type DepartmentData struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Color       string            `json:"color,omitempty"`
	Icon        string            `json:"icon,omitempty"`
	RoomCount   int               `json:"room_count"`
	AgentCount  int               `json:"agent_count"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

// DepartmentManager manages departments in the office
type DepartmentManager struct {
	mu          sync.RWMutex
	departments map[string]*DepartmentData
}

// NewDepartmentManager creates a new department manager
func NewDepartmentManager() *DepartmentManager {
	return &DepartmentManager{
		departments: make(map[string]*DepartmentData),
	}
}

// CreateDepartment creates a new department
func (dm *DepartmentManager) CreateDepartment(config DepartmentConfig) (*DepartmentData, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if config.ID == "" {
		return nil, fmt.Errorf("department ID is required")
	}

	if _, exists := dm.departments[config.ID]; exists {
		return nil, fmt.Errorf("department %s already exists", config.ID)
	}

	dept := &DepartmentData{
		ID:          config.ID,
		Name:        config.Name,
		Description: config.Description,
		Color:       config.Color,
		Icon:        config.Icon,
		RoomCount:   0,
		AgentCount:  0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]any),
	}

	dm.departments[config.ID] = dept
	return dept, nil
}

// GetDepartment retrieves a department by ID
func (dm *DepartmentManager) GetDepartment(id string) (*DepartmentData, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	dept, exists := dm.departments[id]
	if !exists {
		return nil, fmt.Errorf("department not found: %s", id)
	}

	return dept, nil
}

// UpdateDepartment updates a department
func (dm *DepartmentManager) UpdateDepartment(id string, updates map[string]any) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dept, exists := dm.departments[id]
	if !exists {
		return fmt.Errorf("department not found: %s", id)
	}

	if name, ok := updates["name"].(string); ok {
		dept.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		dept.Description = desc
	}
	if color, ok := updates["color"].(string); ok {
		dept.Color = color
	}
	if icon, ok := updates["icon"].(string); ok {
		dept.Icon = icon
	}

	dept.UpdatedAt = time.Now()
	return nil
}

// DeleteDepartment deletes a department
func (dm *DepartmentManager) DeleteDepartment(id string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if _, exists := dm.departments[id]; !exists {
		return fmt.Errorf("department not found: %s", id)
	}

	delete(dm.departments, id)
	return nil
}

// ListDepartments returns all departments
func (dm *DepartmentManager) ListDepartments() []*DepartmentInfo {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	list := make([]*DepartmentInfo, 0, len(dm.departments))
	for _, dept := range dm.departments {
		list = append(list, &DepartmentInfo{
			Name:         dept.Name,
			Description:  dept.Description,
			Roles:        []string{},
			Capabilities: []string{},
		})
	}

	return list
}

// UpdateRoomCount updates the room count for a department
func (dm *DepartmentManager) UpdateRoomCount(deptID string, delta int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dept, exists := dm.departments[deptID]; exists {
		dept.RoomCount += delta
		if dept.RoomCount < 0 {
			dept.RoomCount = 0
		}
		dept.UpdatedAt = time.Now()
	}
}

// UpdateAgentCount updates the agent count for a department
func (dm *DepartmentManager) UpdateAgentCount(deptID string, delta int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dept, exists := dm.departments[deptID]; exists {
		dept.AgentCount += delta
		if dept.AgentCount < 0 {
			dept.AgentCount = 0
		}
		dept.UpdatedAt = time.Now()
	}
}

// GetDefaultDepartments returns the default department configurations
func GetDefaultDepartments() []DepartmentConfig {
	return []DepartmentConfig{
		{
			ID:          "planning",
			Name:        "Planning",
			Description: "Strategic planning and project management",
			Color:       "#3B82F6",
			Icon:        "📋",
		},
		{
			ID:          "coding",
			Name:        "Development",
			Description: "Software development and engineering",
			Color:       "#10B981",
			Icon:        "💻",
		},
		{
			ID:          "design",
			Name:        "Design",
			Description: "UI/UX design and creative work",
			Color:       "#8B5CF6",
			Icon:        "🎨",
		},
		{
			ID:          "marketing",
			Name:        "Marketing",
			Description: "Marketing and communications",
			Color:       "#F59E0B",
			Icon:        "📢",
		},
		{
			ID:          "quality",
			Name:        "Quality Assurance",
			Description: "Testing and quality assurance",
			Color:       "#EC4899",
			Icon:        "🔍",
		},
		{
			ID:          "legal",
			Name:        "Legal",
			Description: "Legal and compliance",
			Color:       "#6366F1",
			Icon:        "⚖️",
		},
	}
}
