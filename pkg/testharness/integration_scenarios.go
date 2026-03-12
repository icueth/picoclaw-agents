// PicoClaw - Integration Test Scenarios
// End-to-end integration tests, full workflow tests, multi-agent collaboration tests

package testharness

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/office"
	"picoclaw/agent/pkg/providers"
)

// IntegrationScenario represents an end-to-end integration test scenario
type IntegrationScenario struct {
	Name        string
	Description string
	Category    string // "e2e", "workflow", "collaboration"
	Setup       func(*MockProvider)
	Test        func(*Harness) error
	Validate    func(*Harness) error
	Cleanup     func() // Optional cleanup after test
}

// IntegrationScenarios contains all integration test scenarios
var IntegrationScenarios = []IntegrationScenario{
	// ============================================
	// End-to-End Integration Tests
	// ============================================
	{
		Name:        "E2E - Complete Task Lifecycle",
		Description: "End-to-end test of a task from creation to completion",
		Category:    "e2e",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("create task", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "create_task",
						Function: &providers.FunctionCall{
							Name:      "create_task",
							Arguments: `{"title": "Build API", "status": "created"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "route_task",
						Function: &providers.FunctionCall{
							Name:      "route_task",
							Arguments: `{"task_id": "task-001", "department": "engineering"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"role": "coder", "task": "Build API"}`,
						},
					},
					{
						ID:   "call_4",
						Type: "function",
						Name: "update_task_status",
						Function: &providers.FunctionCall{
							Name:      "update_task_status",
							Arguments: `{"task_id": "task-001", "status": "completed"}`,
						},
					},
				}).
				WithResponsePattern("lifecycle", "Task completed full lifecycle")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Create, route, execute, and complete a task to build an API")
			if err != nil {
				return err
			}
			// Verify all tools were called
			if err := h.AssertToolCalled("create_task"); err != nil {
				return err
			}
			if err := h.AssertToolCalled("route_task"); err != nil {
				return err
			}
			if err := h.AssertToolCalled("spawn_subagent"); err != nil {
				return err
			}
			return h.AssertToolCalled("update_task_status")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("completed")
		},
	},
	{
		Name:        "E2E - Multi-Department Request",
		Description: "End-to-end test of a request spanning multiple departments",
		Category:    "e2e",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("multi dept", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"role": "planner", "department": "planning", "task": "Design system"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"role": "architect", "department": "architecture", "task": "Create architecture"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"role": "coder", "department": "engineering", "task": "Implement"}`,
						},
					},
					{
						ID:   "call_4",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"role": "reviewer", "department": "qa", "task": "Review"}`,
						},
					},
				}).
				WithResponsePattern("multi-department", "Multi-department request processed")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Process a request through planning, architecture, engineering, and QA")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("spawn_subagent")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("multi-department")
		},
	},
	{
		Name:        "E2E - Error Recovery Flow",
		Description: "End-to-end test of error detection and recovery",
		Category:    "e2e",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("error recovery", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"task": "Process data"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "detect_error",
						Function: &providers.FunctionCall{
							Name:      "detect_error",
							Arguments: `{"error_type": "timeout"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "retry_task",
						Function: &providers.FunctionCall{
							Name:      "retry_task",
							Arguments: `{"retry_count": 1}`,
						},
					},
					{
						ID:   "call_4",
						Type: "function",
						Name: "update_task_status",
						Function: &providers.FunctionCall{
							Name:      "update_task_status",
							Arguments: `{"status": "completed"}`,
						},
					},
				}).
				WithResponsePattern("recovered", "Error recovered and task completed")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Process a task with error detection and retry")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("retry_task")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("recovered")
		},
	},
	{
		Name:        "E2E - Priority Queue Handling",
		Description: "End-to-end test of priority-based task queue",
		Category:    "e2e",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("priority", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "create_task",
						Function: &providers.FunctionCall{
							Name:      "create_task",
							Arguments: `{"priority": "low", "title": "Low priority task"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "create_task",
						Function: &providers.FunctionCall{
							Name:      "create_task",
							Arguments: `{"priority": "critical", "title": "Critical task"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "process_queue",
						Function: &providers.FunctionCall{
							Name:      "process_queue",
							Arguments: `{"order": "priority"}`,
						},
					},
				}).
				WithResponsePattern("priority", "Tasks processed by priority")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Create low and critical priority tasks, then process by priority")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("process_queue")
		},
	},

	// ============================================
	// Full Workflow Tests
	// ============================================
	{
		Name:        "Workflow - Complete Project Workflow",
		Description: "Full workflow test for a complete project",
		Category:    "workflow",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("project workflow", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "execute_workflow",
						Function: &providers.FunctionCall{
							Name:      "execute_workflow",
							Arguments: `{"stage": "input", "status": "completed"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "execute_workflow",
						Function: &providers.FunctionCall{
							Name:      "execute_workflow",
							Arguments: `{"stage": "planning", "status": "completed"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "execute_workflow",
						Function: &providers.FunctionCall{
							Name:      "execute_workflow",
							Arguments: `{"stage": "assignment", "status": "completed"}`,
						},
					},
					{
						ID:   "call_4",
						Type: "function",
						Name: "execute_workflow",
						Function: &providers.FunctionCall{
							Name:      "execute_workflow",
							Arguments: `{"stage": "execution", "status": "completed"}`,
						},
					},
					{
						ID:   "call_5",
						Type: "function",
						Name: "execute_workflow",
						Function: &providers.FunctionCall{
							Name:      "execute_workflow",
							Arguments: `{"stage": "review", "status": "completed"}`,
						},
					},
					{
						ID:   "call_6",
						Type: "function",
						Name: "execute_workflow",
						Function: &providers.FunctionCall{
							Name:      "execute_workflow",
							Arguments: `{"stage": "output", "status": "completed"}`,
						},
					},
				}).
				WithResponsePattern("all stages", "All workflow stages completed successfully")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Execute complete project workflow through all stages")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("execute_workflow")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("all stages")
		},
	},
	{
		Name:        "Workflow - Parallel Execution Workflow",
		Description: "Full workflow test with parallel task execution",
		Category:    "workflow",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("parallel workflow", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"task": "Frontend", "parallel": true}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"task": "Backend", "parallel": true}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"task": "Database", "parallel": true}`,
						},
					},
					{
						ID:   "call_4",
						Type: "function",
						Name: "join_results",
						Function: &providers.FunctionCall{
							Name:      "join_results",
							Arguments: `{"tasks": ["Frontend", "Backend", "Database"]}`,
						},
					},
				}).
				WithResponsePattern("parallel", "Parallel workflow completed")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Execute parallel workflow for frontend, backend, and database")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("join_results")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("parallel")
		},
	},
	{
		Name:        "Workflow - Conditional Branching",
		Description: "Full workflow test with conditional branches",
		Category:    "workflow",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("conditional", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "evaluate_condition",
						Function: &providers.FunctionCall{
							Name:      "evaluate_condition",
							Arguments: `{"condition": "complexity > 5", "result": true}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"role": "architect", "branch": "complex"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "merge_branches",
						Function: &providers.FunctionCall{
							Name:      "merge_branches",
							Arguments: `{"branches": ["complex"]}`,
						},
					},
				}).
				WithResponsePattern("branched", "Conditional workflow completed")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Execute conditional workflow with complexity check")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("evaluate_condition")
		},
	},
	{
		Name:        "Workflow - Loop and Iteration",
		Description: "Full workflow test with iterative processing",
		Category:    "workflow",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("iteration", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "create_iteration",
						Function: &providers.FunctionCall{
							Name:      "create_iteration",
							Arguments: `{"items": ["file1", "file2", "file3"]}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "process_item",
						Function: &providers.FunctionCall{
							Name:      "process_item",
							Arguments: `{"item": "file1"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "process_item",
						Function: &providers.FunctionCall{
							Name:      "process_item",
							Arguments: `{"item": "file2"}`,
						},
					},
					{
						ID:   "call_4",
						Type: "function",
						Name: "process_item",
						Function: &providers.FunctionCall{
							Name:      "process_item",
							Arguments: `{"item": "file3"}`,
						},
					},
					{
						ID:   "call_5",
						Type: "function",
						Name: "aggregate_results",
						Function: &providers.FunctionCall{
							Name:      "aggregate_results",
							Arguments: `{"count": 3}`,
						},
					},
				}).
				WithResponsePattern("iteration", "Iterative workflow completed")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Execute iterative workflow processing 3 files")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("aggregate_results")
		},
	},

	// ============================================
	// Multi-Agent Collaboration Tests
	// ============================================
	{
		Name:        "Collaboration - Pair Programming",
		Description: "Test two agents collaborating on code",
		Category:    "collaboration",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("pair programming", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"role": "coder", "agent_id": "coder-1", "mode": "driver"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"role": "coder", "agent_id": "coder-2", "mode": "navigator"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "collaborate",
						Function: &providers.FunctionCall{
							Name:      "collaborate",
							Arguments: `{"agents": ["coder-1", "coder-2"], "pattern": "pair_programming"}`,
						},
					},
				}).
				WithResponsePattern("collaboration", "Pair programming collaboration completed")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Start pair programming with two coders collaborating")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("collaborate")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("collaboration")
		},
	},
	{
		Name:        "Collaboration - Code Review",
		Description: "Test code review collaboration between author and reviewer",
		Category:    "collaboration",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("code review", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "submit_code",
						Function: &providers.FunctionCall{
							Name:      "submit_code",
							Arguments: `{"author": "coder-1", "code": "function example() {}"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"role": "reviewer", "task": "review code"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "provide_feedback",
						Function: &providers.FunctionCall{
							Name:      "provide_feedback",
							Arguments: `{"reviewer": "reviewer-1", "feedback": "LGTM"}`,
						},
					},
					{
						ID:   "call_4",
						Type: "function",
						Name: "approve_code",
						Function: &providers.FunctionCall{
							Name:      "approve_code",
							Arguments: `{"approved": true}`,
						},
					},
				}).
				WithResponsePattern("reviewed", "Code review collaboration completed")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Submit code for review and get approval")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("approve_code")
		},
	},
	{
		Name:        "Collaboration - Team Discussion",
		Description: "Test multi-agent team discussion",
		Category:    "collaboration",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("discussion", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "start_discussion",
						Function: &providers.FunctionCall{
							Name:      "start_discussion",
							Arguments: `{"topic": "Architecture decision", "participants": 3}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "add_comment",
						Function: &providers.FunctionCall{
							Name:      "add_comment",
							Arguments: `{"agent": "architect", "comment": "Use microservices"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "add_comment",
						Function: &providers.FunctionCall{
							Name:      "add_comment",
							Arguments: `{"agent": "engineer", "comment": "Agreed"}`,
						},
					},
					{
						ID:   "call_4",
						Type: "function",
						Name: "reach_consensus",
						Function: &providers.FunctionCall{
							Name:      "reach_consensus",
							Arguments: `{"decision": "Use microservices"}`,
						},
					},
				}).
				WithResponsePattern("consensus", "Team reached consensus")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Start team discussion on architecture and reach consensus")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("reach_consensus")
		},
	},
	{
		Name:        "Collaboration - Handoff Between Agents",
		Description: "Test task handoff from one agent to another",
		Category:    "collaboration",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("handoff", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"role": "planner", "agent_id": "planner-1"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "handoff_task",
						Function: &providers.FunctionCall{
							Name:      "handoff_task",
							Arguments: `{"from": "planner-1", "to": "coder-1", "context": "full"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "spawn_subagent",
						Function: &providers.FunctionCall{
							Name:      "spawn_subagent",
							Arguments: `{"role": "coder", "agent_id": "coder-1"}`,
						},
					},
					{
						ID:   "call_4",
						Type: "function",
						Name: "acknowledge_handoff",
						Function: &providers.FunctionCall{
							Name:      "acknowledge_handoff",
							Arguments: `{"agent": "coder-1", "received": true}`,
						},
					},
				}).
				WithResponsePattern("handed off", "Task successfully handed off")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Handoff task from planner to coder with full context")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("handoff_task")
		},
	},
	{
		Name:        "Collaboration - Conflict Resolution",
		Description: "Test conflict resolution between agents",
		Category:    "collaboration",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("conflict", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "detect_conflict",
						Function: &providers.FunctionCall{
							Name:      "detect_conflict",
							Arguments: `{"agents": ["agent-1", "agent-2"], "conflict_type": "approach"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "mediate",
						Function: &providers.FunctionCall{
							Name:      "mediate",
							Arguments: `{"mediator": "manager", "approach": "compromise"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "resolve_conflict",
						Function: &providers.FunctionCall{
							Name:      "resolve_conflict",
							Arguments: `{"resolution": "hybrid_approach"}`,
						},
					},
				}).
				WithResponsePattern("resolved", "Conflict resolved successfully")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Detect and resolve conflict between two agents")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("resolve_conflict")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("resolved")
		},
	},
	{
		Name:        "Collaboration - Knowledge Sharing",
		Description: "Test knowledge sharing between agents",
		Category:    "collaboration",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("knowledge", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "share_knowledge",
						Function: &providers.FunctionCall{
							Name:      "share_knowledge",
							Arguments: `{"from": "senior-1", "topic": "design_patterns"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "receive_knowledge",
						Function: &providers.FunctionCall{
							Name:      "receive_knowledge",
							Arguments: `{"to": "junior-1", "topic": "design_patterns"}`,
						},
					},
					{
						ID:   "call_3",
						Type: "function",
						Name: "update_agent_skills",
						Function: &providers.FunctionCall{
							Name:      "update_agent_skills",
							Arguments: `{"agent": "junior-1", "new_skill": "design_patterns"}`,
						},
					},
				}).
				WithResponsePattern("shared", "Knowledge shared successfully")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Share design patterns knowledge from senior to junior agent")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("update_agent_skills")
		},
	},
}

// RunIntegrationScenario executes a single integration scenario
func RunIntegrationScenario(scenario IntegrationScenario) (*ScenarioResult, error) {
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

	if scenario.Cleanup != nil {
		scenario.Cleanup()
	}

	return result, nil
}

// RunIntegrationScenarios executes all integration scenarios
func RunIntegrationScenarios() []ScenarioResult {
	results := make([]ScenarioResult, 0, len(IntegrationScenarios))

	fmt.Printf("🔗 Running %d Integration Test Scenarios...\n\n", len(IntegrationScenarios))

	for _, scenario := range IntegrationScenarios {
		start := time.Now()
		result, _ := RunIntegrationScenario(scenario)
		duration := time.Since(start)
		result.Duration = duration

		results = append(results, *result)

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

// RunIntegrationScenariosByCategory executes scenarios filtered by category
func RunIntegrationScenariosByCategory(category string) []ScenarioResult {
	var filtered []IntegrationScenario
	for _, s := range IntegrationScenarios {
		if s.Category == category {
			filtered = append(filtered, s)
		}
	}

	results := make([]ScenarioResult, 0, len(filtered))
	fmt.Printf("🔗 Running %d integration scenarios for category: %s\n\n", len(filtered), category)

	for _, scenario := range filtered {
		result, _ := RunIntegrationScenario(scenario)
		results = append(results, *result)

		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
		}
		fmt.Printf("%s %s\n", status, scenario.Name)
	}

	return results
}

// PrintIntegrationReport prints a detailed report of integration scenario results
func PrintIntegrationReport(results []ScenarioResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🔗 Integration Test Report")
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

// IntegrationTestHarness extends the base Harness with integration testing capabilities
type IntegrationTestHarness struct {
	*Harness
	Router         *office.TaskRouter
	mu             sync.RWMutex
	collaborations []CollaborationRecord
}

// CollaborationRecord records a collaboration between agents
type CollaborationRecord struct {
	Timestamp time.Time
	AgentIDs  []string
	Pattern   string
	TaskID    string
	Result    string
}

// NewIntegrationTestHarness creates a new integration test harness
func NewIntegrationTestHarness(provider *MockProvider) *IntegrationTestHarness {
	return &IntegrationTestHarness{
		Harness:        New(provider),
		Router:         office.NewTaskRouter(),
		collaborations: make([]CollaborationRecord, 0),
	}
}

// RecordCollaboration records a collaboration event
func (h *IntegrationTestHarness) RecordCollaboration(agentIDs []string, pattern, taskID, result string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.collaborations = append(h.collaborations, CollaborationRecord{
		Timestamp: time.Now(),
		AgentIDs:  agentIDs,
		Pattern:   pattern,
		TaskID:    taskID,
		Result:    result,
	})
}

// GetCollaborations returns all recorded collaborations
func (h *IntegrationTestHarness) GetCollaborations() []CollaborationRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	records := make([]CollaborationRecord, len(h.collaborations))
	copy(records, h.collaborations)
	return records
}

// AssertCollaborationOccurred asserts that a collaboration pattern occurred
func (h *IntegrationTestHarness) AssertCollaborationOccurred(pattern string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, c := range h.collaborations {
		if c.Pattern == pattern {
			return nil
		}
	}
	return fmt.Errorf("expected collaboration pattern %q to occur, but none found", pattern)
}

// CreateWorkflowRequest creates a workflow request for integration testing
func (h *IntegrationTestHarness) CreateWorkflowRequest(taskType, description string, priority office.TaskPriority) office.WorkflowRequest {
	return office.WorkflowRequest{
		ID:          fmt.Sprintf("integ-req-%d", time.Now().Unix()),
		Type:        taskType,
		Description: description,
		Priority:    priority,
		CreatedAt:   time.Now(),
		Context:     make(map[string]interface{}),
	}
}

// IntegrationScenarioWithRealLLM represents an integration scenario for real LLM testing
type IntegrationScenarioWithRealLLM struct {
	Name        string
	Description string
	Category    string
	Test        func(*RealLLMTestHarness) error
}

// RealLLMIntegrationScenarios contains integration scenarios for real LLM testing
var RealLLMIntegrationScenarios = []IntegrationScenarioWithRealLLM{
	{
		Name:        "Real LLM - Multi-Step Workflow",
		Description: "Test real LLM handling a multi-step workflow",
		Category:    "e2e",
		Test: func(h *RealLLMTestHarness) error {
			messages := []string{
				"I need to build a web application. Start by planning the architecture.",
				"Now create the database schema based on that plan.",
				"Implement the API endpoints following the schema.",
				"Finally, review the implementation for best practices.",
			}
			responses, err := h.MultiTurnChat(messages)
			if err != nil {
				return err
			}
			if len(responses) != 4 {
				return fmt.Errorf("expected 4 responses, got %d", len(responses))
			}
			// Check that responses show progression
			for i, resp := range responses {
				if len(resp) < 10 {
					return fmt.Errorf("response %d seems too short: %s", i, resp)
				}
			}
			return nil
		},
	},
	{
		Name:        "Real LLM - Department Coordination",
		Description: "Test real LLM coordinating between departments",
		Category:    "collaboration",
		Test: func(h *RealLLMTestHarness) error {
			response, err := h.Chat("I have a complex project that needs planning, coding, and review. How would you coordinate these departments?")
			if err != nil {
				return err
			}
			lower := strings.ToLower(response)
			if !strings.Contains(lower, "plan") && !strings.Contains(lower, "code") && !strings.Contains(lower, "review") {
				return fmt.Errorf("expected coordination plan, got: %s", response)
			}
			return nil
		},
	},
	{
		Name:        "Real LLM - Error Handling Strategy",
		Description: "Test real LLM proposing error handling strategy",
		Category:    "e2e",
		Test: func(h *RealLLMTestHarness) error {
			response, err := h.Chat("What strategy would you use to handle errors in a distributed multi-agent system?")
			if err != nil {
				return err
			}
			lower := strings.ToLower(response)
			if !strings.Contains(lower, "error") && !strings.Contains(lower, "retry") && !strings.Contains(lower, "fallback") {
				return fmt.Errorf("expected error handling strategy, got: %s", response)
			}
			return nil
		},
	},
}

// RunIntegrationScenarioWithRealLLM executes an integration scenario with real LLM
func RunIntegrationScenarioWithRealLLM(harness *RealLLMTestHarness, scenario IntegrationScenarioWithRealLLM) (*ScenarioResult, error) {
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

// RunAllIntegrationTestsWithRealLLM runs all integration tests with real LLM
func RunAllIntegrationTestsWithRealLLM(harness *RealLLMTestHarness) []ScenarioResult {
	results := make([]ScenarioResult, 0, len(RealLLMIntegrationScenarios))

	fmt.Printf("🔗 Running %d Real LLM Integration Tests...\n\n", len(RealLLMIntegrationScenarios))

	for _, scenario := range RealLLMIntegrationScenarios {
		result, _ := RunIntegrationScenarioWithRealLLM(harness, scenario)
		results = append(results, *result)

		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
		}
		fmt.Printf("%s %s\n", status, scenario.Name)
		if result.Error != nil {
			fmt.Printf("   Error: %v\n", result.Error)
		}
	}

	return results
}

// ParallelIntegrationTest runs multiple integration tests in parallel
func ParallelIntegrationTest(scenarios []IntegrationScenario) []ScenarioResult {
	results := make([]ScenarioResult, len(scenarios))
	var wg sync.WaitGroup

	for i, scenario := range scenarios {
		wg.Add(1)
		go func(index int, s IntegrationScenario) {
			defer wg.Done()
			result, _ := RunIntegrationScenario(s)
			results[index] = *result
		}(i, scenario)
	}

	wg.Wait()
	return results
}

// IntegrationTestMetrics tracks metrics for integration tests
type IntegrationTestMetrics struct {
	TotalTests     int
	PassedTests    int
	FailedTests    int
	TotalDuration  time.Duration
	AvgDuration    time.Duration
	CategoryCounts map[string]int
	CategoryPassed map[string]int
}

// CalculateIntegrationMetrics calculates metrics from results
func CalculateIntegrationMetrics(results []ScenarioResult, scenarios []IntegrationScenario) IntegrationTestMetrics {
	metrics := IntegrationTestMetrics{
		TotalTests:     len(results),
		CategoryCounts: make(map[string]int),
		CategoryPassed: make(map[string]int),
	}

	var totalDuration time.Duration

	for i, result := range results {
		totalDuration += result.Duration
		if result.Passed {
			metrics.PassedTests++
		} else {
			metrics.FailedTests++
		}

		if i < len(scenarios) {
			cat := scenarios[i].Category
			metrics.CategoryCounts[cat]++
			if result.Passed {
				metrics.CategoryPassed[cat]++
			}
		}
	}

	metrics.TotalDuration = totalDuration
	if metrics.TotalTests > 0 {
		metrics.AvgDuration = totalDuration / time.Duration(metrics.TotalTests)
	}

	return metrics
}
