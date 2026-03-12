// Package office provides company-style workflow management for Picoclaw.
// It implements a multi-stage workflow engine that processes tasks through
// various departments similar to a company organizational structure.
package office

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/memory"
	"picoclaw/agent/pkg/tools"
)

// WorkflowStage represents a stage in the workflow pipeline
type WorkflowStage string

const (
	// StageInput is where requests are received and validated
	StageInput WorkflowStage = "input"
	// StagePlanning is where tasks are analyzed and planned
	StagePlanning WorkflowStage = "planning"
	// StageAssignment is where tasks are routed to departments
	StageAssignment WorkflowStage = "assignment"
	// StageExecution is where tasks are executed
	StageExecution WorkflowStage = "execution"
	// StageReview is where completed tasks are reviewed for quality
	StageReview WorkflowStage = "review"
	// StageOutput is where final results are delivered
	StageOutput WorkflowStage = "output"
)

// AllWorkflowStages returns all workflow stages in order
func AllWorkflowStages() []WorkflowStage {
	return []WorkflowStage{
		StageInput,
		StagePlanning,
		StageAssignment,
		StageExecution,
		StageReview,
		StageOutput,
	}
}

// TaskStatus represents the current status of a task in the workflow
type TaskStatus string

const (
	// TaskPending - Task is waiting to be processed
	TaskPending TaskStatus = "pending"
	// TaskInProgress - Task is currently being processed
	TaskInProgress TaskStatus = "in_progress"
	// TaskCompleted - Task has been completed successfully
	TaskCompleted TaskStatus = "completed"
	// TaskFailed - Task has failed
	TaskFailed TaskStatus = "failed"
	// TaskCancelled - Task was cancelled
	TaskCancelled TaskStatus = "cancelled"
	// TaskWaitingReview - Task is waiting for review
	TaskWaitingReview TaskStatus = "waiting_review"
)

// TaskPriority represents the priority level of a task
type TaskPriority int

const (
	TaskPriorityLow TaskPriority = iota
	TaskPriorityNormal
	TaskPriorityHigh
	TaskPriorityCritical
)

// WorkflowRequest represents an incoming request to the workflow
type WorkflowRequest struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Priority    TaskPriority           `json:"priority"`
	Context     map[string]interface{} `json:"context"`
	RequesterID string                 `json:"requester_id"`
	CreatedAt   time.Time              `json:"created_at"`
	Deadline    *time.Time             `json:"deadline,omitempty"`
}

// WorkflowPlan represents the output of the planning stage
type WorkflowPlan struct {
	RequestID       string                 `json:"request_id"`
	Tasks           []PlannedTask          `json:"tasks"`
	EstimatedTime   time.Duration          `json:"estimated_time"`
	RequiredRoles   []string               `json:"required_roles"`
	Dependencies    map[string][]string    `json:"dependencies"`
	Metadata        map[string]interface{} `json:"metadata"`
	PlannedAt       time.Time              `json:"planned_at"`
}

// PlannedTask represents a single task within a workflow plan
type PlannedTask struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Role         string                 `json:"role"`
	Department   Department             `json:"department"`
	EstimatedTime time.Duration         `json:"estimated_time"`
	Dependencies []string               `json:"dependencies"`
	Context      map[string]interface{} `json:"context"`
}

// TaskAssignment represents a task ready for execution
type TaskAssignment struct {
	TaskID       string                 `json:"task_id"`
	PlanID       string                 `json:"plan_id"`
	Role         string                 `json:"role"`
	Department   Department             `json:"department"`
	Description  string                 `json:"description"`
	Context      map[string]interface{} `json:"context"`
	AssignedAt   time.Time              `json:"assigned_at"`
	JobID        string                 `json:"job_id,omitempty"`
}

// TaskExecution represents the execution of a task
type TaskExecution struct {
	Assignment  TaskAssignment         `json:"assignment"`
	Status      TaskStatus             `json:"status"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Result      string                 `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	SubagentID  string                 `json:"subagent_id,omitempty"`
}

// TaskReview represents a review of a completed task
type TaskReview struct {
	ExecutionID   string                 `json:"execution_id"`
	ReviewerRole  string                 `json:"reviewer_role"`
	Status        TaskStatus             `json:"status"`
	Feedback      string                 `json:"feedback"`
	Approved      bool                   `json:"approved"`
	NeedsRevision bool                   `json:"needs_revision"`
	ReviewedAt    time.Time              `json:"reviewed_at"`
}

// TaskFlow represents a task's complete journey through the workflow
type TaskFlow struct {
	ID          string                 `json:"id"`
	Request     WorkflowRequest        `json:"request"`
	Plan        *WorkflowPlan          `json:"plan,omitempty"`
	Assignments []TaskAssignment       `json:"assignments"`
	Executions  []TaskExecution        `json:"executions"`
	Reviews     []TaskReview           `json:"reviews"`
	CurrentStage WorkflowStage          `json:"current_stage"`
	Status      TaskStatus             `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewTaskFlow creates a new task flow from a request
func NewTaskFlow(request WorkflowRequest) *TaskFlow {
	now := time.Now()
	return &TaskFlow{
		ID:           uuid.New().String(),
		Request:      request,
		Assignments:  make([]TaskAssignment, 0),
		Executions:   make([]TaskExecution, 0),
		Reviews:      make([]TaskReview, 0),
		CurrentStage: StageInput,
		Status:       TaskPending,
		CreatedAt:    now,
		UpdatedAt:    now,
		Metadata:     make(map[string]interface{}),
	}
}

// WorkflowEngine manages the company-style workflow
type WorkflowEngine struct {
	router        *TaskRouter
	jobManager    *memory.JobManager
	subagentMgr   *tools.SubagentManager
	workloads     map[Department]int
	maxWorkload   map[Department]int
	flows         map[string]*TaskFlow
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	// Configuration
	autoOutsource bool
	outsourceThreshold int
}

// WorkflowEngineConfig contains configuration for the workflow engine
type WorkflowEngineConfig struct {
	AutoOutsource      bool
	OutsourceThreshold int
	DepartmentLimits   map[Department]int
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(
	router *TaskRouter,
	jobManager *memory.JobManager,
	subagentMgr *tools.SubagentManager,
	config *WorkflowEngineConfig,
) *WorkflowEngine {
	ctx, cancel := context.WithCancel(context.Background())

	engine := &WorkflowEngine{
		router:        router,
		jobManager:    jobManager,
		subagentMgr:   subagentMgr,
		workloads:     make(map[Department]int),
		maxWorkload:   make(map[Department]int),
		flows:         make(map[string]*TaskFlow),
		ctx:           ctx,
		cancel:        cancel,
		autoOutsource: config.AutoOutsource,
		outsourceThreshold: config.OutsourceThreshold,
	}

	// Set default department limits
	for _, dept := range AllDepartments() {
		if limit, ok := config.DepartmentLimits[dept]; ok {
			engine.maxWorkload[dept] = limit
		} else {
			engine.maxWorkload[dept] = 5 // Default limit
		}
	}

	return engine
}

// Stop stops the workflow engine
func (we *WorkflowEngine) Stop() {
	we.cancel()
}

// ProcessRequest is the main entry point for processing a workflow request
func (we *WorkflowEngine) ProcessRequest(request WorkflowRequest) (*TaskFlow, error) {
	// Create task flow
	flow := NewTaskFlow(request)

	we.mu.Lock()
	we.flows[flow.ID] = flow
	we.mu.Unlock()

	logger.InfoCF("workflow", "Processing new request", map[string]any{
		"flow_id": flow.ID,
		"type":    request.Type,
	})

	// Stage 1: Input validation
	if err := we.validateRequest(request); err != nil {
		flow.Status = TaskFailed
		flow.Metadata["error"] = err.Error()
		return flow, err
	}

	// Stage 2: Planning
	flow.CurrentStage = StagePlanning
	plan, err := we.planningStage(request)
	if err != nil {
		flow.Status = TaskFailed
		flow.Metadata["error"] = fmt.Sprintf("planning failed: %v", err)
		return flow, err
	}
	flow.Plan = plan
	flow.UpdatedAt = time.Now()

	logger.InfoCF("workflow", "Planning complete", map[string]any{
		"flow_id":     flow.ID,
		"tasks_count": len(plan.Tasks),
	})

	// Stage 3: Assignment
	flow.CurrentStage = StageAssignment
	assignments := we.assignmentStage(plan)
	flow.Assignments = assignments
	flow.UpdatedAt = time.Now()

	logger.InfoCF("workflow", "Assignment complete", map[string]any{
		"flow_id":         flow.ID,
		"assignments_count": len(assignments),
	})

	// Stage 4: Execution
	flow.CurrentStage = StageExecution
	flow.Status = TaskInProgress

	// Check if we need to outsource due to high workload
	shouldOutsource := we.shouldOutsource(assignments)

	if shouldOutsource {
		logger.InfoCF("workflow", "Outsourcing tasks due to high workload", map[string]any{
			"flow_id": flow.ID,
		})
		flow.Metadata["outsourced"] = true
	}

	// Execute tasks (async for multiple assignments)
	if len(assignments) == 1 {
		// Single task - execute synchronously
		execution := we.executionStage(assignments[0], shouldOutsource)
		flow.Executions = append(flow.Executions, execution)
	} else {
		// Multiple tasks - execute asynchronously
		executions := we.executeAsync(assignments, shouldOutsource)
		flow.Executions = executions
	}

	flow.UpdatedAt = time.Now()

	// Check if any executions failed
	allCompleted := true
	hasFailures := false
	for _, exec := range flow.Executions {
		if exec.Status == TaskFailed {
			hasFailures = true
		}
		if exec.Status != TaskCompleted && exec.Status != TaskFailed {
			allCompleted = false
		}
	}

	// Stage 5: Review (if completed and not failed)
	if allCompleted && !hasFailures {
		flow.CurrentStage = StageReview
		for _, execution := range flow.Executions {
			review := we.reviewStage(execution)
			flow.Reviews = append(flow.Reviews, review)
		}
	}

	// Stage 6: Output
	flow.CurrentStage = StageOutput
	if hasFailures {
		flow.Status = TaskFailed
	} else {
		flow.Status = TaskCompleted
	}
	now := time.Now()
	flow.CompletedAt = &now
	flow.UpdatedAt = now

	logger.InfoCF("workflow", "Request processing complete", map[string]any{
		"flow_id": flow.ID,
		"status":  flow.Status,
	})

	return flow, nil
}

// validateRequest validates the incoming request
func (we *WorkflowEngine) validateRequest(request WorkflowRequest) error {
	if request.Description == "" {
		return fmt.Errorf("request description is required")
	}
	if request.Type == "" {
		return fmt.Errorf("request type is required")
	}
	return nil
}

// planningStage handles the planning department work
func (we *WorkflowEngine) planningStage(request WorkflowRequest) (*WorkflowPlan, error) {
	// Use the planning role to break down the request
	plan := &WorkflowPlan{
		RequestID:    request.ID,
		Tasks:        make([]PlannedTask, 0),
		Dependencies: make(map[string][]string),
		Metadata:     make(map[string]interface{}),
		PlannedAt:    time.Now(),
	}

	// Determine the complexity and break down into tasks
	// For simple requests, create a single task
	// For complex requests, break down into multiple tasks

	taskTypes := we.analyzeRequestType(request)

	for i, taskType := range taskTypes {
		department := we.router.DetermineDepartment(taskType)
		role := we.router.DetermineRole(taskType, request.Description)

		task := PlannedTask{
			ID:           fmt.Sprintf("%s-task-%d", request.ID, i),
			Type:         taskType,
			Description:  we.generateTaskDescription(request, taskType, i),
			Role:         role,
			Department:   department,
			EstimatedTime: we.estimateTaskTime(taskType, request.Description),
			Dependencies: []string{},
			Context:      request.Context,
		}

		plan.Tasks = append(plan.Tasks, task)
		plan.RequiredRoles = append(plan.RequiredRoles, role)
	}

	// Calculate total estimated time
	var totalTime time.Duration
	for _, task := range plan.Tasks {
		totalTime += task.EstimatedTime
	}
	plan.EstimatedTime = totalTime

	// Store plan metadata
	plan.Metadata["task_count"] = len(plan.Tasks)
	plan.Metadata["complexity"] = we.assessComplexity(request)

	return plan, nil
}

// analyzeRequestType analyzes the request and determines task types needed
func (we *WorkflowEngine) analyzeRequestType(request WorkflowRequest) []string {
	// Simple heuristic-based analysis
	// In a real implementation, this could use LLM to analyze

	desc := request.Description
	taskTypes := []string{}

	// Check for code-related keywords
	if containsAny(desc, []string{"code", "implement", "function", "class", "refactor", "debug"}) {
		taskTypes = append(taskTypes, "code")
	}

	// Check for research-related keywords
	if containsAny(desc, []string{"research", "find", "search", "analyze", "investigate"}) {
		taskTypes = append(taskTypes, "research")
	}

	// Check for planning-related keywords
	if containsAny(desc, []string{"plan", "design", "architecture", "strategy"}) {
		taskTypes = append(taskTypes, "planning")
	}

	// Check for writing-related keywords
	if containsAny(desc, []string{"write", "document", "create content", "blog"}) {
		taskTypes = append(taskTypes, "content")
	}

	// Check for review-related keywords
	if containsAny(desc, []string{"review", "audit", "check", "verify"}) {
		taskTypes = append(taskTypes, "review")
	}

	// If no specific type detected, default to execution
	if len(taskTypes) == 0 {
		taskTypes = append(taskTypes, "execution")
	}

	return taskTypes
}

// generateTaskDescription generates a specific description for a task
func (we *WorkflowEngine) generateTaskDescription(request WorkflowRequest, taskType string, index int) string {
	return fmt.Sprintf("[%s] %s (part %d of request %s)",
		taskType,
		request.Description,
		index+1,
		request.ID,
	)
}

// estimateTaskTime estimates the time needed for a task
func (we *WorkflowEngine) estimateTaskTime(taskType string, description string) time.Duration {
	// Base estimates by task type
	baseTimes := map[string]time.Duration{
		"code":      10 * time.Minute,
		"research":  5 * time.Minute,
		"planning":  5 * time.Minute,
		"content":   10 * time.Minute,
		"review":    5 * time.Minute,
		"execution": 5 * time.Minute,
	}

	base := baseTimes[taskType]
	if base == 0 {
		base = 5 * time.Minute
	}

	// Adjust based on description length (heuristic for complexity)
	if len(description) > 500 {
		base *= 2
	}
	if len(description) > 1000 {
		base *= 2
	}

	return base
}

// assessComplexity assesses the complexity of a request
func (we *WorkflowEngine) assessComplexity(request WorkflowRequest) string {
	descLen := len(request.Description)
	if descLen < 100 {
		return "simple"
	}
	if descLen < 500 {
		return "moderate"
	}
	if descLen < 1000 {
		return "complex"
	}
	return "very_complex"
}

// assignmentStage routes tasks to departments and creates assignments
func (we *WorkflowEngine) assignmentStage(plan *WorkflowPlan) []TaskAssignment {
	assignments := make([]TaskAssignment, 0, len(plan.Tasks))

	for _, task := range plan.Tasks {
		// Check department workload
		we.mu.Lock()
		currentLoad := we.workloads[task.Department]
		we.workloads[task.Department] = currentLoad + 1
		we.mu.Unlock()

		assignment := TaskAssignment{
			TaskID:      task.ID,
			PlanID:      plan.RequestID,
			Role:        task.Role,
			Department:  task.Department,
			Description: task.Description,
			Context:     task.Context,
			AssignedAt:  time.Now(),
		}

		// Create job in job manager if available
		if we.jobManager != nil {
			jobCtx := map[string]interface{}{
				"task_id":     task.ID,
				"department":  string(task.Department),
				"role":        task.Role,
				"plan_id":     plan.RequestID,
			}
			if task.Context != nil {
				for k, v := range task.Context {
					jobCtx[k] = v
				}
			}

			jobID, err := we.jobManager.CreateJob(task.Role, task.Description, jobCtx)
			if err == nil {
				assignment.JobID = jobID
			}
		}

		assignments = append(assignments, assignment)
	}

	return assignments
}

// shouldOutsource determines if tasks should be outsourced due to high workload
func (we *WorkflowEngine) shouldOutsource(assignments []TaskAssignment) bool {
	if !we.autoOutsource {
		return false
	}

	we.mu.RLock()
	defer we.mu.RUnlock()

	// Check if any department is over threshold
	for _, assignment := range assignments {
		dept := assignment.Department
		currentLoad := we.workloads[dept]
		maxLoad := we.maxWorkload[dept]

		if maxLoad > 0 && currentLoad > maxLoad {
			return true
		}

		// Check against outsource threshold
		if we.outsourceThreshold > 0 && currentLoad > we.outsourceThreshold {
			return true
		}
	}

	return false
}

// executionStage executes a single task assignment
func (we *WorkflowEngine) executionStage(assignment TaskAssignment, outsource bool) TaskExecution {
	execution := TaskExecution{
		Assignment: assignment,
		Status:     TaskInProgress,
	}

	now := time.Now()
	execution.StartedAt = &now

	// Update job status if available
	if we.jobManager != nil && assignment.JobID != "" {
		we.jobManager.UpdateJobStatus(assignment.JobID, string(memory.JobRunning))
	}

	// Execute via subagent manager
	if we.subagentMgr != nil {
		ctx, cancel := context.WithTimeout(we.ctx, 30*time.Minute)
		defer cancel()

		// Prepare context data
		contextData := map[string]interface{}{
			"department": string(assignment.Department),
			"task_type":  assignment.Role,
		}
		if assignment.Context != nil {
			for k, v := range assignment.Context {
				contextData[k] = v
			}
		}

		// Add outsource flag if applicable
		if outsource {
			contextData["outsourced"] = true
		}

		// Spawn subagent with role
		result, err := we.subagentMgr.SpawnWithRole(
			ctx,
			assignment.Role,
			assignment.Description,
			contextData,
			"", // conceptID
			0,  // timeout (use role default)
			"workflow",
			assignment.PlanID,
			nil, // callback
		)

		if err != nil {
			execution.Status = TaskFailed
			execution.Error = err.Error()
			if we.jobManager != nil && assignment.JobID != "" {
				we.jobManager.UpdateJobStatus(assignment.JobID, string(memory.JobFailed))
				we.jobManager.UpdateJobResult(assignment.JobID, err.Error())
			}
		} else {
			execution.Status = TaskCompleted
			execution.Result = result
			// Extract subagent ID from result if possible
			execution.SubagentID = extractSubagentID(result)
			if we.jobManager != nil && assignment.JobID != "" {
				we.jobManager.UpdateJobStatus(assignment.JobID, string(memory.JobCompleted))
				we.jobManager.UpdateJobResult(assignment.JobID, result)
			}
		}
	} else {
		execution.Status = TaskFailed
		execution.Error = "subagent manager not available"
	}

	completedAt := time.Now()
	execution.CompletedAt = &completedAt

	// Decrement workload
	we.mu.Lock()
	we.workloads[assignment.Department]--
	if we.workloads[assignment.Department] < 0 {
		we.workloads[assignment.Department] = 0
	}
	we.mu.Unlock()

	return execution
}

// executeAsync executes multiple assignments asynchronously
func (we *WorkflowEngine) executeAsync(assignments []TaskAssignment, outsource bool) []TaskExecution {
	var wg sync.WaitGroup
	executions := make([]TaskExecution, len(assignments))
	var mu sync.Mutex

	for i, assignment := range assignments {
		wg.Add(1)
		go func(index int, assign TaskAssignment) {
			defer wg.Done()
			execution := we.executionStage(assign, outsource)
			mu.Lock()
			executions[index] = execution
			mu.Unlock()
		}(i, assignment)
	}

	wg.Wait()
	return executions
}

// reviewStage performs QA review on a completed task
func (we *WorkflowEngine) reviewStage(execution TaskExecution) TaskReview {
	review := TaskReview{
		ExecutionID:  fmt.Sprintf("%s-review", execution.Assignment.TaskID),
		ReviewerRole: "reviewer",
		ReviewedAt:   time.Now(),
	}

	// Simple automated review logic
	// In a real implementation, this could use a reviewer subagent

	if execution.Status == TaskFailed {
		review.Status = TaskFailed
		review.Approved = false
		review.Feedback = fmt.Sprintf("Task failed with error: %s", execution.Error)
		review.NeedsRevision = true
		return review
	}

	// Check if result is empty or too short
	if execution.Result == "" || len(execution.Result) < 10 {
		review.Status = TaskFailed
		review.Approved = false
		review.Feedback = "Task result is empty or insufficient"
		review.NeedsRevision = true
		return review
	}

	// Basic approval
	review.Status = TaskCompleted
	review.Approved = true
	review.Feedback = "Task completed successfully and meets basic quality criteria"
	review.NeedsRevision = false

	return review
}

// GetFlow retrieves a task flow by ID
func (we *WorkflowEngine) GetFlow(flowID string) (*TaskFlow, bool) {
	we.mu.RLock()
	defer we.mu.RUnlock()
	flow, ok := we.flows[flowID]
	return flow, ok
}

// ListFlows returns all task flows
func (we *WorkflowEngine) ListFlows() []*TaskFlow {
	we.mu.RLock()
	defer we.mu.RUnlock()

	flows := make([]*TaskFlow, 0, len(we.flows))
	for _, flow := range we.flows {
		flows = append(flows, flow)
	}
	return flows
}

// GetWorkload returns current workload for all departments
func (we *WorkflowEngine) GetWorkload() map[Department]int {
	we.mu.RLock()
	defer we.mu.RUnlock()

	workload := make(map[Department]int)
	for dept, count := range we.workloads {
		workload[dept] = count
	}
	return workload
}

// Helper functions

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if containsSubstring(s, substr) {
			return true
		}
	}
	return false
}

func containsSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func extractSubagentID(result string) string {
	// Simple extraction - look for patterns like "task_id: subagent-X"
	// This is a placeholder - real implementation would parse properly
	return ""
}
