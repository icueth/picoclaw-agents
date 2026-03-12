// PicoClaw - Office UI Test Scenarios
// Test scenarios for Office UI functionality: Agent Assignment, Task Creation, Department Routing

package testharness

import (
	"context"
	"fmt"
	"strings"
	"time"

	"picoclaw/agent/pkg/office"
	"picoclaw/agent/pkg/providers"
)

// OfficeScenario represents a test scenario for Office UI workflows
type OfficeScenario struct {
	Name        string
	Description string
	Department  office.Department
	Role        string
	Setup       func(*MockProvider)
	Test        func(*Harness) error
	Validate    func(*Harness) error // Optional additional validation
}

// OfficeScenarios contains all Office UI test scenarios
var OfficeScenarios = []OfficeScenario{
	// ============================================
	// Agent Assignment Tests
	// ============================================
	{
		Name:        "Agent Assignment - Engineering Department",
		Description: "Test assigning a coding task to an engineering agent",
		Department:  office.DeptEngineering,
		Role:        "coder",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("write code", "spawn_subagent", map[string]any{
					"role":        "coder",
					"task":        "Write a Go function to calculate factorial",
					"department":  "engineering",
					"description": "Create a recursive factorial function with error handling",
				}).
				WithResponsePattern("assignment", "Task assigned to engineering department")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Help me write a Go function to calculate factorial")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("spawn_subagent")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("assigned")
		},
	},
	{
		Name:        "Agent Assignment - Research Department",
		Description: "Test assigning a research task to a research agent",
		Department:  office.DeptResearch,
		Role:        "researcher",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("research", "spawn_subagent", map[string]any{
					"role":       "researcher",
					"task":       "Research best practices for Go error handling",
					"department": "research",
				}).
				WithResponsePattern("research", "I'll research that for you")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Research best practices for Go error handling")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("spawn_subagent")
		},
	},
	{
		Name:        "Agent Assignment - Planning Department",
		Description: "Test assigning a planning task to a planning agent",
		Department:  office.DeptPlanning,
		Role:        "planner",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("plan", "spawn_subagent", map[string]any{
					"role":       "planner",
					"task":       "Create a project plan for a microservices architecture",
					"department": "planning",
				}).
				WithResponsePattern("plan", "I'll create a comprehensive plan")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Create a project plan for a microservices architecture")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("spawn_subagent")
		},
	},
	{
		Name:        "Agent Assignment - QA Department",
		Description: "Test assigning a review task to a QA agent",
		Department:  office.DeptQA,
		Role:        "reviewer",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("review", "spawn_subagent", map[string]any{
					"role":       "reviewer",
					"task":       "Review this code for best practices",
					"department": "qa",
				}).
				WithResponsePattern("review", "I'll review the code")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Review this code for best practices")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("spawn_subagent")
		},
	},

	// ============================================
	// Task Creation Tests
	// ============================================
	{
		Name:        "Task Creation - Simple Task",
		Description: "Test creating a simple single task",
		Department:  office.DeptEngineering,
		Role:        "coder",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("create task", "create_task", map[string]any{
					"title":       "Implement user authentication",
					"description": "Create JWT-based authentication system",
					"department":  "engineering",
					"role":        "coder",
					"priority":    "high",
				}).
				WithResponsePattern("task", "Task created successfully")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Create a task to implement user authentication")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("create_task")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("created")
		},
	},
	{
		Name:        "Task Creation - Complex Multi-Step Task",
		Description: "Test creating a complex task with multiple subtasks",
		Department:  office.DeptPlanning,
		Role:        "planner",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("complex project", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "create_task",
						Function: &providers.FunctionCall{
							Name:      "create_task",
							Arguments: `{"title": "Design database schema", "department": "architecture"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "create_task",
						Function: &providers.FunctionCall{
							Name:      "create_task",
							Arguments: `{"title": "Implement API endpoints", "department": "engineering"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "create_task",
						Function: &providers.FunctionCall{
							Name:      "create_task",
							Arguments: `{"title": "Write tests", "department": "qa"}`,
						},
					},
				}).
				WithResponsePattern("project", "I've broken down the project into tasks")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Create a complex project with multiple tasks: design database, implement API, and write tests")
			if err != nil {
				return err
			}
			// Should have created multiple tasks
			return h.AssertToolCalled("create_task")
		},
	},
	{
		Name:        "Task Creation - With Priority",
		Description: "Test creating a task with priority level",
		Department:  office.DeptOperations,
		Role:        "executor",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("urgent", "create_task", map[string]any{
					"title":       "Fix critical bug",
					"description": "Fix the memory leak in production",
					"priority":    "critical",
					"department":  "engineering",
				}).
				WithResponsePattern("critical", "Critical task created with high priority")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Create an urgent critical task to fix the memory leak")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("create_task")
		},
	},
	{
		Name:        "Task Creation - With Deadline",
		Description: "Test creating a task with a deadline",
		Department:  office.DeptPlanning,
		Role:        "planner",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("deadline", "create_task", map[string]any{
					"title":       "Complete documentation",
					"description": "Write API documentation",
					"deadline":    "2024-12-31T23:59:59Z",
					"department":  "writing",
				}).
				WithResponsePattern("deadline", "Task created with deadline")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Create a task to complete documentation by end of year")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("create_task")
		},
	},

	// ============================================
	// Department Routing Tests
	// ============================================
	{
		Name:        "Department Routing - Code to Engineering",
		Description: "Test routing a coding task to engineering department",
		Department:  office.DeptEngineering,
		Role:        "coder",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("route code", "route_task", map[string]any{
					"task_type":   "code",
					"department":  "engineering",
					"role":        "coder",
					"description": "Implement a REST API endpoint",
				}).
				WithResponsePattern("engineering", "Task routed to Engineering department")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Route a coding task to implement a REST API")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("route_task")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("Engineering")
		},
	},
	{
		Name:        "Department Routing - Research to Research",
		Description: "Test routing a research task to research department",
		Department:  office.DeptResearch,
		Role:        "researcher",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("route research", "route_task", map[string]any{
					"task_type":   "research",
					"department":  "research",
					"role":        "researcher",
					"description": "Research competitor analysis",
				}).
				WithResponsePattern("research", "Task routed to Research department")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Route a research task for competitor analysis")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("route_task")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("Research")
		},
	},
	{
		Name:        "Department Routing - Content to Marketing",
		Description: "Test routing a content task to marketing department",
		Department:  office.DeptMarketing,
		Role:        "writer",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("route content", "route_task", map[string]any{
					"task_type":   "content",
					"department":  "marketing",
					"role":        "writer",
					"description": "Write a blog post about our product",
				}).
				WithResponsePattern("marketing", "Task routed to Marketing department")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Route a content task to write a blog post")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("route_task")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("Marketing")
		},
	},
	{
		Name:        "Department Routing - Auto-Detect from Description",
		Description: "Test automatic department detection from task description",
		Department:  office.DeptEngineering,
		Role:        "coder",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("auto route", "route_task", map[string]any{
					"task_type":     "auto",
					"department":    "engineering",
					"detected_from": "description keywords: code, implement, function",
					"confidence":    0.95,
				}).
				WithResponsePattern("auto-detected", "Department auto-detected as Engineering")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Route this task: implement a sorting function in Python")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("route_task")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("auto-detected")
		},
	},

	// ============================================
	// Room Assignment Tests
	// ============================================
	{
		Name:        "Room Assignment - Default Room",
		Description: "Test assigning an agent to the default room",
		Department:  office.DeptEngineering,
		Role:        "coder",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("assign room", "assign_room", map[string]any{
					"agent_id":   "agent-001",
					"room_id":    "main-office",
					"room_name":  "Main Office",
					"department": "engineering",
				}).
				WithResponsePattern("room", "Agent assigned to Main Office")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Assign the engineering agent to the main office")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("assign_room")
		},
	},
	{
		Name:        "Room Assignment - Department Room",
		Description: "Test assigning an agent to a department-specific room",
		Department:  office.DeptQA,
		Role:        "reviewer",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("qa room", "assign_room", map[string]any{
					"agent_id":   "agent-qa-001",
					"room_id":    "qa-lab",
					"room_name":  "QA Testing Lab",
					"department": "qa",
				}).
				WithResponsePattern("lab", "Agent assigned to QA Testing Lab")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Assign the QA agent to the testing lab")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("assign_room")
		},
	},
	{
		Name:        "Room Assignment - Move Between Rooms",
		Description: "Test moving an agent between rooms",
		Department:  office.DeptEngineering,
		Role:        "coder",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("move room", "move_agent", map[string]any{
					"agent_id":  "agent-001",
					"from_room": "planning-room",
					"to_room":   "engineering-lab",
					"reason":    "task reassignment",
				}).
				WithResponsePattern("moved", "Agent moved to Engineering Lab")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Move the agent from planning to engineering lab")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("move_agent")
		},
	},

	// ============================================
	// Workflow Execution Tests
	// ============================================
	{
		Name:        "Workflow Execution - Simple Workflow",
		Description: "Test executing a simple single-stage workflow",
		Department:  office.DeptOperations,
		Role:        "executor",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("execute workflow", "execute_workflow", map[string]any{
					"workflow_id": "wf-001",
					"stages":      []string{"input", "execution", "output"},
					"status":      "completed",
				}).
				WithResponsePattern("completed", "Workflow executed successfully")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Execute the simple workflow")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("execute_workflow")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("completed")
		},
	},
	{
		Name:        "Workflow Execution - Multi-Stage Workflow",
		Description: "Test executing a multi-stage workflow with planning and review",
		Department:  office.DeptPlanning,
		Role:        "planner",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("complex workflow", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "execute_workflow",
						Function: &providers.FunctionCall{
							Name:      "execute_workflow",
							Arguments: `{"stage": "planning", "status": "completed"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "execute_workflow",
						Function: &providers.FunctionCall{
							Name:      "execute_workflow",
							Arguments: `{"stage": "execution", "status": "completed"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "execute_workflow",
						Function: &providers.FunctionCall{
							Name:      "execute_workflow",
							Arguments: `{"stage": "review", "status": "completed"}`,
						},
					},
				}).
				WithResponsePattern("all stages", "All workflow stages completed")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Execute the full workflow with planning, execution, and review")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("execute_workflow")
		},
	},
	{
		Name:        "Workflow Execution - With Error Handling",
		Description: "Test workflow execution with error recovery",
		Department:  office.DeptOperations,
		Role:        "executor",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("error workflow", "execute_workflow", map[string]any{
					"workflow_id": "wf-error-001",
					"status":      "recovered",
					"error":       "temporary network issue",
					"recovery":    "retry succeeded",
				}).
				WithResponsePattern("recovered", "Workflow recovered from error")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Execute workflow with error handling")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("execute_workflow")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("recovered")
		},
	},
	{
		Name:        "Workflow Execution - Parallel Tasks",
		Description: "Test executing parallel tasks in a workflow",
		Department:  office.DeptEngineering,
		Role:        "coder",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("parallel", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"task": "Implement frontend", "parallel": true}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"task": "Implement backend", "parallel": true}`,
						},
					},
				}).
				WithResponsePattern("parallel", "Parallel tasks executed")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Execute parallel tasks for frontend and backend")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("spawn_subagent")
		},
	},
}

// RunOfficeScenario executes a single office scenario
func RunOfficeScenario(scenario OfficeScenario) (*ScenarioResult, error) {
	provider := NewMockProvider()
	scenario.Setup(provider)

	harness := New(provider)

	start := time.Now()
	err := scenario.Test(harness)
	duration := time.Since(start)

	result := &ScenarioResult{
		Name:     scenario.Name,
		Passed:   err == nil,
		Duration: duration,
		Error:    err,
	}

	if scenario.Validate != nil && err == nil {
		if valErr := scenario.Validate(harness); valErr != nil {
			result.Passed = false
			result.Error = valErr
		}
	}

	return result, nil
}

// RunOfficeScenarios executes all office scenarios
func RunOfficeScenarios() []ScenarioResult {
	results := make([]ScenarioResult, 0, len(OfficeScenarios))

	fmt.Printf("🏢 Running %d Office UI Test Scenarios...\n\n", len(OfficeScenarios))

	for _, scenario := range OfficeScenarios {
		start := time.Now()
		result, _ := RunOfficeScenario(scenario)
		duration := time.Since(start)
		result.Duration = duration

		results = append(results, *result)

		// Print result immediately
		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
		}
		fmt.Printf("%s %s (%v)\n", status, scenario.Name, duration)
		if result.Error != nil {
			fmt.Printf("   Error: %v\n", result.Error)
		}
	}

	return results
}

// RunOfficeScenariosByDepartment executes scenarios filtered by department
func RunOfficeScenariosByDepartment(dept office.Department) []ScenarioResult {
	var filtered []OfficeScenario
	for _, s := range OfficeScenarios {
		if s.Department == dept {
			filtered = append(filtered, s)
		}
	}

	results := make([]ScenarioResult, 0, len(filtered))
	fmt.Printf("🏢 Running %d scenarios for department: %s\n\n", len(filtered), dept)

	for _, scenario := range filtered {
		result, _ := RunOfficeScenario(scenario)
		results = append(results, *result)

		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
		}
		fmt.Printf("%s %s\n", status, scenario.Name)
	}

	return results
}

// RunOfficeScenariosByCategory executes scenarios filtered by category
func RunOfficeScenariosByCategory(category string) []ScenarioResult {
	categoryKeywords := map[string][]string{
		"assignment": {"Agent Assignment"},
		"task":       {"Task Creation"},
		"routing":    {"Department Routing"},
		"room":       {"Room Assignment"},
		"workflow":   {"Workflow Execution"},
	}

	keywords, ok := categoryKeywords[category]
	if !ok {
		return []ScenarioResult{}
	}

	var filtered []OfficeScenario
	for _, s := range OfficeScenarios {
		for _, kw := range keywords {
			if strings.Contains(s.Name, kw) {
				filtered = append(filtered, s)
				break
			}
		}
	}

	results := make([]ScenarioResult, 0, len(filtered))
	fmt.Printf("🏢 Running %d scenarios for category: %s\n\n", len(filtered), category)

	for _, scenario := range filtered {
		result, _ := RunOfficeScenario(scenario)
		results = append(results, *result)

		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
		}
		fmt.Printf("%s %s\n", status, scenario.Name)
	}

	return results
}

// PrintOfficeReport prints a detailed report of office scenario results
func PrintOfficeReport(results []ScenarioResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🏢 Office UI Test Report")
	fmt.Println(strings.Repeat("=", 60))

	passed := 0
	failed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Printf("Total: %d | Passed: %d | Failed: %d\n", len(results), passed, failed)
	fmt.Println(strings.Repeat("=", 60))

	if failed > 0 {
		fmt.Println("\n❌ Failed Tests:")
		for _, r := range results {
			if !r.Passed {
				fmt.Printf("  • %s: %v\n", r.Name, r.Error)
			}
		}
	}
}

// OfficeTestHarness extends the base Harness with Office-specific functionality
type OfficeTestHarness struct {
	*Harness
	Router     *office.TaskRouter
	Department office.Department
	Role       string
}

// NewOfficeTestHarness creates a new Office test harness
func NewOfficeTestHarness(provider *MockProvider, dept office.Department, role string) *OfficeTestHarness {
	return &OfficeTestHarness{
		Harness:    New(provider),
		Router:     office.NewTaskRouter(),
		Department: dept,
		Role:       role,
	}
}

// RouteTask routes a task using the office task router
func (h *OfficeTestHarness) RouteTask(taskType, description string) (office.Department, string, error) {
	dept, role := h.Router.Route(taskType, description)
	return dept, role, nil
}

// CreateWorkflowRequest creates a workflow request for testing
func (h *OfficeTestHarness) CreateWorkflowRequest(taskType, description string, priority office.TaskPriority) office.WorkflowRequest {
	return office.WorkflowRequest{
		ID:          fmt.Sprintf("test-req-%d", time.Now().Unix()),
		Type:        taskType,
		Description: description,
		Priority:    priority,
		CreatedAt:   time.Now(),
		Context:     make(map[string]interface{}),
	}
}

// ValidateDepartment validates that a department is valid
func (h *OfficeTestHarness) ValidateDepartment(dept office.Department) error {
	if !office.IsValidDepartment(string(dept)) {
		return fmt.Errorf("invalid department: %s", dept)
	}
	return nil
}

// AssertDepartmentEquals asserts that the department matches expected
func (h *OfficeTestHarness) AssertDepartmentEquals(expected office.Department) error {
	if h.Department != expected {
		return fmt.Errorf("expected department %s, got %s", expected, h.Department)
	}
	return nil
}

// AssertRoleEquals asserts that the role matches expected
func (h *OfficeTestHarness) AssertRoleEquals(expected string) error {
	if h.Role != expected {
		return fmt.Errorf("expected role %s, got %s", expected, h.Role)
	}
	return nil
}

// OfficeScenarioWithRealLLM runs an office scenario with real LLM
type OfficeScenarioWithRealLLM struct {
	Name        string
	Description string
	Department  office.Department
	Role        string
	Test        func(*RealLLMTestHarness) error
}

// RunOfficeScenarioWithRealLLM executes an office scenario with real LLM
func RunOfficeScenarioWithRealLLM(harness *RealLLMTestHarness, scenario OfficeScenarioWithRealLLM) (*ScenarioResult, error) {
	start := time.Now()
	err := scenario.Test(harness)
	duration := time.Since(start)

	return &ScenarioResult{
		Name:        scenario.Name,
		Description: scenario.Description,
		Passed:      err == nil,
		Duration:    duration,
		Error:       err,
	}, nil
}

// RealLLMOfficeScenarios contains scenarios designed for real LLM testing
var RealLLMOfficeScenarios = []OfficeScenarioWithRealLLM{
	{
		Name:        "Real LLM - Department Routing",
		Description: "Test real LLM can route tasks to correct departments",
		Department:  office.DeptEngineering,
		Role:        "coder",
		Test: func(h *RealLLMTestHarness) error {
			response, err := h.Chat("I need to implement a new API endpoint in Go. Which department should handle this?")
			if err != nil {
				return err
			}
			// Check if response mentions engineering or coding
			lower := strings.ToLower(response)
			if !strings.Contains(lower, "engineer") && !strings.Contains(lower, "code") && !strings.Contains(lower, "develop") {
				return fmt.Errorf("expected response about engineering, got: %s", response)
			}
			return nil
		},
	},
	{
		Name:        "Real LLM - Task Planning",
		Description: "Test real LLM can create task plans",
		Department:  office.DeptPlanning,
		Role:        "planner",
		Test: func(h *RealLLMTestHarness) error {
			response, err := h.Chat("Create a plan for building a web application with user authentication")
			if err != nil {
				return err
			}
			// Check if response contains planning elements
			lower := strings.ToLower(response)
			if !strings.Contains(lower, "step") && !strings.Contains(lower, "plan") && !strings.Contains(lower, "phase") {
				return fmt.Errorf("expected a plan with steps, got: %s", response)
			}
			return nil
		},
	},
	{
		Name:        "Real LLM - Multi-turn Office Conversation",
		Description: "Test real LLM maintains context in office scenarios",
		Department:  office.DeptEngineering,
		Role:        "coder",
		Test: func(h *RealLLMTestHarness) error {
			messages := []string{
				"Assign a coding task to the engineering department",
				"What department did I just assign that task to?",
				"Can you route another task to the same department?",
			}
			responses, err := h.MultiTurnChat(messages)
			if err != nil {
				return err
			}
			if len(responses) != 3 {
				return fmt.Errorf("expected 3 responses, got %d", len(responses))
			}
			// Check second response remembers the department
			lower := strings.ToLower(responses[1])
			if !strings.Contains(lower, "engineer") {
				return fmt.Errorf("expected context awareness about engineering, got: %s", responses[1])
			}
			return nil
		},
	},
}

// Context key type for office test harness
type officeHarnessKey struct{}

// WithOfficeContext adds office test harness to context
func WithOfficeContext(ctx context.Context, harness *OfficeTestHarness) context.Context {
	return context.WithValue(ctx, officeHarnessKey{}, harness)
}

// GetOfficeContext retrieves office test harness from context
func GetOfficeContext(ctx context.Context) (*OfficeTestHarness, bool) {
	harness, ok := ctx.Value(officeHarnessKey{}).(*OfficeTestHarness)
	return harness, ok
}
