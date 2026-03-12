package office

import (
	"testing"
	"time"
)

func TestNewTaskFlow(t *testing.T) {
	request := WorkflowRequest{
		ID:          "test-123",
		Type:        "code",
		Description: "Test task",
		Priority:    TaskPriorityNormal,
	}

	flow := NewTaskFlow(request)

	if flow.ID == "" {
		t.Error("TaskFlow should have an ID")
	}
	if flow.Request.ID != request.ID {
		t.Error("TaskFlow should store the request")
	}
	if flow.CurrentStage != StageInput {
		t.Errorf("Initial stage should be %v, got %v", StageInput, flow.CurrentStage)
	}
	if flow.Status != TaskPending {
		t.Errorf("Initial status should be %v, got %v", TaskPending, flow.Status)
	}
}

func TestAllWorkflowStages(t *testing.T) {
	stages := AllWorkflowStages()
	expected := []WorkflowStage{StageInput, StagePlanning, StageAssignment, StageExecution, StageReview, StageOutput}

	if len(stages) != len(expected) {
		t.Errorf("Expected %d stages, got %d", len(expected), len(stages))
	}

	for i, stage := range stages {
		if stage != expected[i] {
			t.Errorf("Stage %d: expected %v, got %v", i, expected[i], stage)
		}
	}
}

func TestNewWorkflowEngine(t *testing.T) {
	router := NewTaskRouter()
	config := &WorkflowEngineConfig{
		AutoOutsource:      true,
		OutsourceThreshold: 3,
		DepartmentLimits: map[Department]int{
			DeptEngineering: 5,
		},
	}

	engine := NewWorkflowEngine(router, nil, nil, config)
	if engine == nil {
		t.Fatal("NewWorkflowEngine returned nil")
	}

	if !engine.autoOutsource {
		t.Error("AutoOutsource should be true")
	}
	if engine.outsourceThreshold != 3 {
		t.Errorf("OutsourceThreshold should be 3, got %d", engine.outsourceThreshold)
	}

	engine.Stop()
}

func TestWorkflowEngineValidateRequest(t *testing.T) {
	router := NewTaskRouter()
	engine := NewWorkflowEngine(router, nil, nil, &WorkflowEngineConfig{})
	defer engine.Stop()

	// Valid request
	validReq := WorkflowRequest{
		Type:        "code",
		Description: "Write a function",
	}
	err := engine.validateRequest(validReq)
	if err != nil {
		t.Errorf("Valid request should not return error: %v", err)
	}

	// Missing description
	invalidReq := WorkflowRequest{
		Type:        "code",
		Description: "",
	}
	err = engine.validateRequest(invalidReq)
	if err == nil {
		t.Error("Request without description should return error")
	}

	// Missing type
	invalidReq2 := WorkflowRequest{
		Type:        "",
		Description: "Write a function",
	}
	err = engine.validateRequest(invalidReq2)
	if err == nil {
		t.Error("Request without type should return error")
	}
}

func TestAnalyzeRequestType(t *testing.T) {
	router := NewTaskRouter()
	engine := NewWorkflowEngine(router, nil, nil, &WorkflowEngineConfig{})
	defer engine.Stop()

	tests := []struct {
		description string
		expectCode  bool
		expectResearch bool
	}{
		{"Write code to implement a function", true, false},
		{"Debug this error", false, false}, // Debug is a separate task type, not code
		{"Research best practices", false, true},
		{"Plan the architecture", false, false},
		{"Write documentation", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			req := WorkflowRequest{
				Description: tt.description,
			}
			types := engine.analyzeRequestType(req)

			hasCode := false
			hasResearch := false
			for _, taskType := range types {
				if taskType == "code" {
					hasCode = true
				}
				if taskType == "research" {
					hasResearch = true
				}
			}

			if hasCode != tt.expectCode {
				t.Errorf("Expected code=%v, got %v", tt.expectCode, hasCode)
			}
			if hasResearch != tt.expectResearch {
				t.Errorf("Expected research=%v, got %v", tt.expectResearch, hasResearch)
			}
		})
	}
}

func TestEstimateTaskTime(t *testing.T) {
	router := NewTaskRouter()
	engine := NewWorkflowEngine(router, nil, nil, &WorkflowEngineConfig{})
	defer engine.Stop()

	// Test base times
	codeTime := engine.estimateTaskTime("code", "simple")
	if codeTime <= 0 {
		t.Error("Code task should have positive time estimate")
	}

	// Test complexity adjustment
	simpleTime := engine.estimateTaskTime("code", "simple")
	longDesc := make([]byte, 600)
	for i := range longDesc {
		longDesc[i] = 'a'
	}
	complexTime := engine.estimateTaskTime("code", string(longDesc))

	if complexTime <= simpleTime {
		t.Error("Complex task should have longer estimate than simple task")
	}
}

func TestAssessComplexity(t *testing.T) {
	router := NewTaskRouter()
	engine := NewWorkflowEngine(router, nil, nil, &WorkflowEngineConfig{})
	defer engine.Stop()

	tests := []struct {
		description string
		expected    string
	}{
		{"short", "simple"},
		{"this is a moderately long description with some details about what needs to be done", "simple"}, // Under 100 chars threshold
		{"this is a very long description " + string(make([]byte, 500)), "complex"},
		{"this is an extremely long description " + string(make([]byte, 1000)), "very_complex"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			req := WorkflowRequest{
				Description: tt.description,
			}
			complexity := engine.assessComplexity(req)
			if complexity != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, complexity)
			}
		})
	}
}

func TestShouldOutsource(t *testing.T) {
	router := NewTaskRouter()
	config := &WorkflowEngineConfig{
		AutoOutsource:      true,
		OutsourceThreshold: 2,
		DepartmentLimits: map[Department]int{
			DeptEngineering: 3,
		},
	}
	engine := NewWorkflowEngine(router, nil, nil, config)
	defer engine.Stop()

	// Initially should not outsource
	assignments := []TaskAssignment{
		{Department: DeptEngineering},
	}

	if engine.shouldOutsource(assignments) {
		t.Error("Should not outsource with low workload")
	}

	// Simulate high workload
	engine.workloads[DeptEngineering] = 5

	if !engine.shouldOutsource(assignments) {
		t.Error("Should outsource with high workload")
	}
}

func TestReviewStage(t *testing.T) {
	router := NewTaskRouter()
	engine := NewWorkflowEngine(router, nil, nil, &WorkflowEngineConfig{})
	defer engine.Stop()

	// Test failed execution
	failedExec := TaskExecution{
		Status: TaskFailed,
		Error:  "something went wrong",
	}
	review := engine.reviewStage(failedExec)
	if review.Approved {
		t.Error("Failed execution should not be approved")
	}
	if !review.NeedsRevision {
		t.Error("Failed execution should need revision")
	}

	// Test empty result
	emptyExec := TaskExecution{
		Status: TaskCompleted,
		Result: "",
	}
	review = engine.reviewStage(emptyExec)
	if review.Approved {
		t.Error("Empty result should not be approved")
	}

	// Test successful execution
	successExec := TaskExecution{
		Status: TaskCompleted,
		Result: "This is a valid result with sufficient content",
	}
	review = engine.reviewStage(successExec)
	if !review.Approved {
		t.Error("Successful execution should be approved")
	}
}

func TestGetWorkload(t *testing.T) {
	router := NewTaskRouter()
	engine := NewWorkflowEngine(router, nil, nil, &WorkflowEngineConfig{})
	defer engine.Stop()

	// Set some workloads
	engine.workloads[DeptEngineering] = 3
	engine.workloads[DeptResearch] = 1

	workload := engine.GetWorkload()

	if workload[DeptEngineering] != 3 {
		t.Errorf("Expected engineering workload 3, got %d", workload[DeptEngineering])
	}
	if workload[DeptResearch] != 1 {
		t.Errorf("Expected research workload 1, got %d", workload[DeptResearch])
	}
}

func TestGetFlow(t *testing.T) {
	router := NewTaskRouter()
	engine := NewWorkflowEngine(router, nil, nil, &WorkflowEngineConfig{})
	defer engine.Stop()

	request := WorkflowRequest{
		Type:        "code",
		Description: "Test",
	}

	flow := NewTaskFlow(request)
	engine.flows[flow.ID] = flow

	// Get existing flow
	retrieved, ok := engine.GetFlow(flow.ID)
	if !ok {
		t.Error("Should find existing flow")
	}
	if retrieved.ID != flow.ID {
		t.Error("Retrieved flow should match")
	}

	// Get non-existent flow
	_, ok = engine.GetFlow("non-existent")
	if ok {
		t.Error("Should not find non-existent flow")
	}
}

func TestListFlows(t *testing.T) {
	router := NewTaskRouter()
	engine := NewWorkflowEngine(router, nil, nil, &WorkflowEngineConfig{})
	defer engine.Stop()

	// Create some flows
	for i := 0; i < 3; i++ {
		request := WorkflowRequest{
			Type:        "code",
			Description: "Test",
		}
		flow := NewTaskFlow(request)
		engine.flows[flow.ID] = flow
	}

	flows := engine.ListFlows()
	if len(flows) != 3 {
		t.Errorf("Expected 3 flows, got %d", len(flows))
	}
}

func TestPlanningStage(t *testing.T) {
	router := NewTaskRouter()
	engine := NewWorkflowEngine(router, nil, nil, &WorkflowEngineConfig{})
	defer engine.Stop()

	request := WorkflowRequest{
		ID:          "test-123",
		Type:        "code",
		Description: "Write a function to parse JSON and debug any issues",
		Context:     map[string]interface{}{"key": "value"},
	}

	plan, err := engine.planningStage(request)
	if err != nil {
		t.Errorf("Planning should not error: %v", err)
	}

	if plan == nil {
		t.Fatal("Plan should not be nil")
	}

	if plan.RequestID != request.ID {
		t.Error("Plan should reference request ID")
	}

	if len(plan.Tasks) == 0 {
		t.Error("Plan should have tasks")
	}

	if plan.EstimatedTime <= 0 {
		t.Error("Plan should have estimated time")
	}
}

func TestAssignmentStage(t *testing.T) {
	router := NewTaskRouter()
	engine := NewWorkflowEngine(router, nil, nil, &WorkflowEngineConfig{})
	defer engine.Stop()

	plan := &WorkflowPlan{
		RequestID: "test-123",
		Tasks: []PlannedTask{
			{
				ID:          "task-1",
				Type:        "code",
				Description: "Write code",
				Role:        "coder",
				Department:  DeptEngineering,
			},
			{
				ID:          "task-2",
				Type:        "research",
				Description: "Research",
				Role:        "researcher",
				Department:  DeptResearch,
			},
		},
	}

	assignments := engine.assignmentStage(plan)

	if len(assignments) != 2 {
		t.Errorf("Expected 2 assignments, got %d", len(assignments))
	}

	// Check workload tracking
	if engine.workloads[DeptEngineering] != 1 {
		t.Errorf("Expected engineering workload 1, got %d", engine.workloads[DeptEngineering])
	}
}

func TestWorkflowRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request WorkflowRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: WorkflowRequest{
				ID:          "test-1",
				Type:        "code",
				Description: "Write a function",
				Priority:    TaskPriorityNormal,
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing description",
			request: WorkflowRequest{
				ID:        "test-2",
				Type:      "code",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing type",
			request: WorkflowRequest{
				ID:          "test-3",
				Description: "Write a function",
				CreatedAt:   time.Now(),
			},
			wantErr: true,
		},
	}

	router := NewTaskRouter()
	engine := NewWorkflowEngine(router, nil, nil, &WorkflowEngineConfig{})
	defer engine.Stop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.validateRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
