// Package office provides company-style workflow management for Picoclaw.
package office

import (
	"fmt"
	"strings"
	"sync"
)

// Department represents a company department/team
type Department string

const (
	// DeptPlanning handles task analysis and planning
	DeptPlanning Department = "planning"
	// DeptEngineering handles code development and technical implementation
	DeptEngineering Department = "engineering"
	// DeptResearch handles information gathering and analysis
	DeptResearch Department = "research"
	// DeptMarketing handles content creation and marketing materials
	DeptMarketing Department = "marketing"
	// DeptQA handles quality assurance and review
	DeptQA Department = "qa"
	// DeptOperations handles execution and operational tasks
	DeptOperations Department = "operations"
	// DeptArchitecture handles system design and architecture
	DeptArchitecture Department = "architecture"
	// DeptWriting handles documentation and technical writing
	DeptWriting Department = "writing"
)

// AllDepartments returns all available departments
func AllDepartments() []Department {
	return []Department{
		DeptPlanning,
		DeptEngineering,
		DeptResearch,
		DeptMarketing,
		DeptQA,
		DeptOperations,
		DeptArchitecture,
		DeptWriting,
	}
}

// IsValidDepartment checks if a department is valid
func IsValidDepartment(dept string) bool {
	for _, d := range AllDepartments() {
		if string(d) == dept {
			return true
		}
	}
	return false
}

// DepartmentInfo contains information about a department
type DepartmentInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Roles       []string `json:"roles"`
	Capabilities []string `json:"capabilities"`
}

// GetDepartmentInfo returns information about a department
func GetDepartmentInfo(dept Department) DepartmentInfo {
	info := map[Department]DepartmentInfo{
		DeptPlanning: {
			Name:        "Planning",
			Description: "Analyzes requests and creates detailed execution plans",
			Roles:       []string{"planner", "analyst"},
			Capabilities: []string{"task_breakdown", "dependency_analysis", "estimation"},
		},
		DeptEngineering: {
			Name:        "Engineering",
			Description: "Implements code solutions and technical tasks",
			Roles:       []string{"coder", "developer", "debugger", "engineer"},
			Capabilities: []string{"coding", "debugging", "refactoring", "testing"},
		},
		DeptResearch: {
			Name:        "Research",
			Description: "Gathers information and performs analysis",
			Roles:       []string{"researcher", "analyst", "investigator"},
			Capabilities: []string{"web_search", "data_analysis", "investigation"},
		},
		DeptMarketing: {
			Name:        "Marketing",
			Description: "Creates content and marketing materials",
			Roles:       []string{"writer", "content_creator", "marketer"},
			Capabilities: []string{"content_creation", "copywriting", "blog_writing"},
		},
		DeptQA: {
			Name:        "Quality Assurance",
			Description: "Reviews and validates work for quality",
			Roles:       []string{"reviewer", "auditor", "tester"},
			Capabilities: []string{"code_review", "quality_check", "validation"},
		},
		DeptOperations: {
			Name:        "Operations",
			Description: "Executes operational tasks and coordinates work",
			Roles:       []string{"executor", "operator", "coordinator"},
			Capabilities: []string{"task_execution", "coordination", "deployment"},
		},
		DeptArchitecture: {
			Name:        "Architecture",
			Description: "Designs system architecture and high-level solutions",
			Roles:       []string{"architect", "designer"},
			Capabilities: []string{"system_design", "architecture", "pattern_selection"},
		},
		DeptWriting: {
			Name:        "Technical Writing",
			Description: "Creates documentation and technical content",
			Roles:       []string{"writer", "documentarian"},
			Capabilities: []string{"documentation", "technical_writing", "api_docs"},
		},
	}

	if i, ok := info[dept]; ok {
		return i
	}
	return DepartmentInfo{Name: string(dept)}
}

// TaskTypeMapping maps task types to departments and roles
type TaskTypeMapping struct {
	TaskType   string     `json:"task_type"`
	Department Department `json:"department"`
	DefaultRole string    `json:"default_role"`
	Keywords   []string   `json:"keywords"`
	Priority   int        `json:"priority"` // Higher priority = preferred mapping when multiple match
}

// TaskRouter routes tasks to appropriate departments and agents
type TaskRouter struct {
	mappings      []TaskTypeMapping
	customRoutes  map[string]Department
	roleOverrides map[string]string
	mu            sync.RWMutex
}

// NewTaskRouter creates a new task router with default mappings
func NewTaskRouter() *TaskRouter {
	router := &TaskRouter{
		mappings:      make([]TaskTypeMapping, 0),
		customRoutes:  make(map[string]Department),
		roleOverrides: make(map[string]string),
	}

	router.initializeDefaultMappings()
	return router
}

// initializeDefaultMappings sets up the default task type mappings
func (tr *TaskRouter) initializeDefaultMappings() {
	tr.mappings = []TaskTypeMapping{
		// Code-related tasks -> Engineering
		{
			TaskType:    "code",
			Department:  DeptEngineering,
			DefaultRole: "coder",
			Keywords:    []string{"code", "implement", "function", "class", "method", "programming", "development"},
			Priority:    100,
		},
		{
			TaskType:    "debug",
			Department:  DeptEngineering,
			DefaultRole: "debugger",
			Keywords:    []string{"debug", "fix", "bug", "error", "issue", "troubleshoot"},
			Priority:    100,
		},
		{
			TaskType:    "refactor",
			Department:  DeptEngineering,
			DefaultRole: "coder",
			Keywords:    []string{"refactor", "rewrite", "restructure", "optimize", "clean up"},
			Priority:    90,
		},
		{
			TaskType:    "test",
			Department:  DeptEngineering,
			DefaultRole: "coder",
			Keywords:    []string{"test", "unit test", "integration test", "test case"},
			Priority:    90,
		},

		// Research tasks -> Research
		{
			TaskType:    "research",
			Department:  DeptResearch,
			DefaultRole: "researcher",
			Keywords:    []string{"research", "investigate", "analyze", "study", "explore"},
			Priority:    100,
		},
		{
			TaskType:    "search",
			Department:  DeptResearch,
			DefaultRole: "researcher",
			Keywords:    []string{"search", "find", "lookup", "discover"},
			Priority:    90,
		},

		// Planning tasks -> Planning
		{
			TaskType:    "planning",
			Department:  DeptPlanning,
			DefaultRole: "planner",
			Keywords:    []string{"plan", "design", "strategy", "roadmap", "blueprint"},
			Priority:    100,
		},
		{
			TaskType:    "architecture",
			Department:  DeptArchitecture,
			DefaultRole: "architect",
			Keywords:    []string{"architecture", "system design", "infrastructure", "scalability"},
			Priority:    100,
		},

		// Content/Marketing tasks -> Marketing
		{
			TaskType:    "content",
			Department:  DeptMarketing,
			DefaultRole: "writer",
			Keywords:    []string{"content", "blog", "article", "post", "marketing", "social media"},
			Priority:    100,
		},
		{
			TaskType:    "copywriting",
			Department:  DeptMarketing,
			DefaultRole: "writer",
			Keywords:    []string{"copy", "headline", "tagline", "promotional", "advertising"},
			Priority:    90,
		},

		// Documentation tasks -> Writing
		{
			TaskType:    "documentation",
			Department:  DeptWriting,
			DefaultRole: "writer",
			Keywords:    []string{"document", "documentation", "readme", "guide", "manual"},
			Priority:    100,
		},
		{
			TaskType:    "technical_writing",
			Department:  DeptWriting,
			DefaultRole: "writer",
			Keywords:    []string{"technical writing", "api doc", "specification", "reference"},
			Priority:    90,
		},

		// Review/QA tasks -> QA
		{
			TaskType:    "review",
			Department:  DeptQA,
			DefaultRole: "reviewer",
			Keywords:    []string{"review", "audit", "check", "verify", "validate", "inspect"},
			Priority:    100,
		},
		{
			TaskType:    "code_review",
			Department:  DeptQA,
			DefaultRole: "reviewer",
			Keywords:    []string{"code review", "pr review", "pull request", "peer review"},
			Priority:    100,
		},

		// Execution tasks -> Operations
		{
			TaskType:    "execution",
			Department:  DeptOperations,
			DefaultRole: "executor",
			Keywords:    []string{"execute", "run", "perform", "carry out", "implement"},
			Priority:    80,
		},
	}
}

// DetermineDepartment determines the appropriate department for a task type
func (tr *TaskRouter) DetermineDepartment(taskType string) Department {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	// Check custom routes first
	if dept, ok := tr.customRoutes[taskType]; ok {
		return dept
	}

	// Find matching mapping
	for _, mapping := range tr.mappings {
		if mapping.TaskType == taskType {
			return mapping.Department
		}
	}

	// Default to operations if no match found
	return DeptOperations
}

// DetermineDepartmentFromDescription determines department based on task description
func (tr *TaskRouter) DetermineDepartmentFromDescription(description string) Department {
	descLower := strings.ToLower(description)
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	// Score each department based on keyword matches
	deptScores := make(map[Department]int)

	for _, mapping := range tr.mappings {
		for _, keyword := range mapping.Keywords {
			if strings.Contains(descLower, strings.ToLower(keyword)) {
				deptScores[mapping.Department] += mapping.Priority
			}
		}
	}

	// Find department with highest score
	var bestDept Department = DeptOperations
	bestScore := 0

	for dept, score := range deptScores {
		if score > bestScore {
			bestScore = score
			bestDept = dept
		}
	}

	return bestDept
}

// DetermineRole determines the appropriate role for a task
func (tr *TaskRouter) DetermineRole(taskType string, description string) string {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	// Check role overrides first
	if role, ok := tr.roleOverrides[taskType]; ok {
		return role
	}

	// Find matching mapping
	for _, mapping := range tr.mappings {
		if mapping.TaskType == taskType {
			// Check if description suggests a more specific role
			specificRole := tr.inferSpecificRole(mapping.DefaultRole, description)
			if specificRole != "" {
				return specificRole
			}
			return mapping.DefaultRole
		}
	}

	// Infer role from description
	return tr.inferRoleFromDescription(description)
}

// inferSpecificRole infers a more specific role based on description
func (tr *TaskRouter) inferSpecificRole(baseRole string, description string) string {
	descLower := strings.ToLower(description)

	switch baseRole {
	case "coder":
		if strings.Contains(descLower, "debug") || strings.Contains(descLower, "fix") || strings.Contains(descLower, "bug") {
			return "debugger"
		}
		if strings.Contains(descLower, "test") {
			return "coder" // Could be specific test role in future
		}
	case "writer":
		if strings.Contains(descLower, "document") || strings.Contains(descLower, "readme") {
			return "writer"
		}
		if strings.Contains(descLower, "content") || strings.Contains(descLower, "blog") {
			return "writer"
		}
	}

	return ""
}

// inferRoleFromDescription infers a role from task description
func (tr *TaskRouter) inferRoleFromDescription(description string) string {
	descLower := strings.ToLower(description)

	// Check for role-specific keywords
	roleKeywords := map[string][]string{
		"coder":      {"code", "implement", "function", "class", "programming"},
		"debugger":   {"debug", "fix", "bug", "error", "troubleshoot"},
		"researcher": {"research", "investigate", "analyze", "study"},
		"planner":    {"plan", "design", "architecture", "strategy"},
		"reviewer":   {"review", "audit", "check", "verify"},
		"writer":     {"write", "document", "content", "blog"},
		"executor":   {"execute", "run", "perform"},
		"architect":  {"system design", "architecture", "infrastructure"},
	}

	for role, keywords := range roleKeywords {
		for _, keyword := range keywords {
			if strings.Contains(descLower, keyword) {
				return role
			}
		}
	}

	return "executor" // Default role
}

// Route finds the best department and role for a task
func (tr *TaskRouter) Route(taskType string, description string) (Department, string) {
	dept := tr.DetermineDepartment(taskType)
	if dept == DeptOperations && description != "" {
		// Try to infer from description
		dept = tr.DetermineDepartmentFromDescription(description)
	}

	role := tr.DetermineRole(taskType, description)

	return dept, role
}

// AddCustomRoute adds a custom route for a task type
func (tr *TaskRouter) AddCustomRoute(taskType string, dept Department) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.customRoutes[taskType] = dept
}

// RemoveCustomRoute removes a custom route
func (tr *TaskRouter) RemoveCustomRoute(taskType string) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	delete(tr.customRoutes, taskType)
}

// SetRoleOverride sets a role override for a task type
func (tr *TaskRouter) SetRoleOverride(taskType string, role string) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.roleOverrides[taskType] = role
}

// RemoveRoleOverride removes a role override
func (tr *TaskRouter) RemoveRoleOverride(taskType string) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	delete(tr.roleOverrides, taskType)
}

// AddMapping adds a new task type mapping
func (tr *TaskRouter) AddMapping(mapping TaskTypeMapping) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.mappings = append(tr.mappings, mapping)
}

// GetMappings returns all task type mappings
func (tr *TaskRouter) GetMappings() []TaskTypeMapping {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	mappings := make([]TaskTypeMapping, len(tr.mappings))
	copy(mappings, tr.mappings)
	return mappings
}

// GetDepartmentMappings returns all mappings for a department
func (tr *TaskRouter) GetDepartmentMappings(dept Department) []TaskTypeMapping {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	var mappings []TaskTypeMapping
	for _, mapping := range tr.mappings {
		if mapping.Department == dept {
			mappings = append(mappings, mapping)
		}
	}
	return mappings
}

// RouteRequest is a convenience method that routes a workflow request
func (tr *TaskRouter) RouteRequest(request WorkflowRequest) (Department, string, []string) {
	// Determine primary department and role
	dept, role := tr.Route(request.Type, request.Description)

	// Determine additional roles needed based on request analysis
	additionalRoles := tr.determineAdditionalRoles(request)

	return dept, role, additionalRoles
}

// determineAdditionalRoles determines if additional roles are needed
func (tr *TaskRouter) determineAdditionalRoles(request WorkflowRequest) []string {
	roles := make([]string, 0)
	descLower := strings.ToLower(request.Description)

	// Check if review is needed
	if containsAny(descLower, []string{"review", "check quality", "validate"}) {
		roles = append(roles, "reviewer")
	}

	// Check if research is needed
	if containsAny(descLower, []string{"research", "find information", "look up"}) {
		roles = append(roles, "researcher")
	}

	// Check if planning is needed
	if containsAny(descLower, []string{"plan", "design", "architecture"}) {
		roles = append(roles, "planner")
	}

	return roles
}

// GetBestAgentForTask returns the best agent/role for a specific task
func (tr *TaskRouter) GetBestAgentForTask(taskType string, complexity string) string {
	// Map complexity to role variants
	complexityModifiers := map[string]string{
		"simple":   "junior",
		"moderate": "",
		"complex":  "senior",
	}

	baseRole := tr.DetermineRole(taskType, "")
	modifier := complexityModifiers[complexity]

	if modifier != "" {
		// In a real implementation, this could map to different role configs
		return fmt.Sprintf("%s_%s", modifier, baseRole)
	}

	return baseRole
}

// ValidateRoute validates if a route is valid
func (tr *TaskRouter) ValidateRoute(taskType string, dept Department, role string) error {
	if !IsValidDepartment(string(dept)) {
		return fmt.Errorf("invalid department: %s", dept)
	}

	// Check if the role is appropriate for the department
	info := GetDepartmentInfo(dept)
	roleValid := false
	for _, r := range info.Roles {
		if r == role {
			roleValid = true
			break
		}
	}

	if !roleValid {
		return fmt.Errorf("role %s is not valid for department %s", role, dept)
	}

	return nil
}
