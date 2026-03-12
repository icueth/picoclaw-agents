package office

import (
	"testing"
)

func TestNewTaskRouter(t *testing.T) {
	router := NewTaskRouter()
	if router == nil {
		t.Fatal("NewTaskRouter returned nil")
	}

	if len(router.mappings) == 0 {
		t.Error("TaskRouter should have default mappings")
	}
}

func TestDetermineDepartment(t *testing.T) {
	router := NewTaskRouter()

	tests := []struct {
		taskType string
		expected Department
	}{
		{"code", DeptEngineering},
		{"debug", DeptEngineering},
		{"research", DeptResearch},
		{"planning", DeptPlanning},
		{"content", DeptMarketing},
		{"documentation", DeptWriting},
		{"review", DeptQA},
		{"execution", DeptOperations},
		{"unknown", DeptOperations}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.taskType, func(t *testing.T) {
			dept := router.DetermineDepartment(tt.taskType)
			if dept != tt.expected {
				t.Errorf("DetermineDepartment(%q) = %v, want %v", tt.taskType, dept, tt.expected)
			}
		})
	}
}

func TestDetermineDepartmentFromDescription(t *testing.T) {
	router := NewTaskRouter()

	tests := []struct {
		description string
		expected    Department
	}{
		{"Write a function to parse JSON", DeptEngineering},
		{"Debug this error in the code", DeptEngineering},
		{"Research the best practices for Go", DeptResearch},
		{"Create a plan for the project", DeptPlanning},
		{"Write a blog post about AI", DeptMarketing},
		{"Document the API endpoints", DeptWriting},
		{"Review this code for bugs", DeptEngineering}, // "code" keyword scores higher
		{"Execute the deployment script", DeptOperations},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			dept := router.DetermineDepartmentFromDescription(tt.description)
			if dept != tt.expected {
				t.Errorf("DetermineDepartmentFromDescription(%q) = %v, want %v", tt.description, dept, tt.expected)
			}
		})
	}
}

func TestDetermineRole(t *testing.T) {
	router := NewTaskRouter()

	tests := []struct {
		taskType    string
		description string
		expected    string
	}{
		{"code", "", "coder"},
		{"debug", "", "debugger"},
		{"research", "", "researcher"},
		{"planning", "", "planner"},
		{"review", "", "reviewer"},
		{"", "debug this issue", "debugger"},
		{"", "research the topic", "researcher"},
		{"", "unknown task", "executor"},
	}

	for _, tt := range tests {
		t.Run(tt.taskType+"_"+tt.description, func(t *testing.T) {
			role := router.DetermineRole(tt.taskType, tt.description)
			if role != tt.expected {
				t.Errorf("DetermineRole(%q, %q) = %v, want %v", tt.taskType, tt.description, role, tt.expected)
			}
		})
	}
}

func TestRoute(t *testing.T) {
	router := NewTaskRouter()

	dept, role := router.Route("code", "implement a function")
	if dept != DeptEngineering {
		t.Errorf("Route code: expected dept %v, got %v", DeptEngineering, dept)
	}
	if role != "coder" {
		t.Errorf("Route code: expected role coder, got %v", role)
	}
}

func TestCustomRoute(t *testing.T) {
	router := NewTaskRouter()

	// Add custom route
	router.AddCustomRoute("custom_task", DeptResearch)

	dept := router.DetermineDepartment("custom_task")
	if dept != DeptResearch {
		t.Errorf("Custom route: expected %v, got %v", DeptResearch, dept)
	}

	// Remove custom route
	router.RemoveCustomRoute("custom_task")
	dept = router.DetermineDepartment("custom_task")
	if dept != DeptOperations {
		t.Errorf("After removing custom route: expected %v, got %v", DeptOperations, dept)
	}
}

func TestRoleOverride(t *testing.T) {
	router := NewTaskRouter()

	// Set role override
	router.SetRoleOverride("code", "senior_coder")

	role := router.DetermineRole("code", "")
	if role != "senior_coder" {
		t.Errorf("Role override: expected senior_coder, got %v", role)
	}

	// Remove override
	router.RemoveRoleOverride("code")
	role = router.DetermineRole("code", "")
	if role != "coder" {
		t.Errorf("After removing override: expected coder, got %v", role)
	}
}

func TestGetDepartmentInfo(t *testing.T) {
	info := GetDepartmentInfo(DeptEngineering)
	if info.Name != "Engineering" {
		t.Errorf("Expected Engineering, got %v", info.Name)
	}

	if len(info.Roles) == 0 {
		t.Error("Engineering department should have roles")
	}
}

func TestAllDepartments(t *testing.T) {
	depts := AllDepartments()
	if len(depts) == 0 {
		t.Error("AllDepartments should return departments")
	}

	expected := []Department{DeptPlanning, DeptEngineering, DeptResearch, DeptMarketing, DeptQA, DeptOperations, DeptArchitecture, DeptWriting}
	if len(depts) != len(expected) {
		t.Errorf("Expected %d departments, got %d", len(expected), len(depts))
	}
}

func TestIsValidDepartment(t *testing.T) {
	if !IsValidDepartment("engineering") {
		t.Error("engineering should be a valid department")
	}
	if !IsValidDepartment("planning") {
		t.Error("planning should be a valid department")
	}
	if IsValidDepartment("invalid") {
		t.Error("invalid should not be a valid department")
	}
}

func TestValidateRoute(t *testing.T) {
	router := NewTaskRouter()

	// Valid route
	err := router.ValidateRoute("code", DeptEngineering, "coder")
	if err != nil {
		t.Errorf("Valid route should not return error: %v", err)
	}

	// Invalid department
	err = router.ValidateRoute("code", "invalid_dept", "coder")
	if err == nil {
		t.Error("Invalid department should return error")
	}

	// Invalid role for department
	err = router.ValidateRoute("code", DeptEngineering, "planner")
	if err == nil {
		t.Error("Invalid role for department should return error")
	}
}

func TestRouteRequest(t *testing.T) {
	router := NewTaskRouter()

	request := WorkflowRequest{
		Type:        "code",
		Description: "Write a function to debug the issue",
	}

	dept, role, additionalRoles := router.RouteRequest(request)

	if dept != DeptEngineering {
		t.Errorf("Expected dept %v, got %v", DeptEngineering, dept)
	}
	// Role is inferred from description containing "debug"
	if role != "debugger" {
		t.Errorf("Expected role debugger (from description), got %v", role)
	}
	// Debug keyword doesn't add additional roles in current implementation
	_ = additionalRoles
}
