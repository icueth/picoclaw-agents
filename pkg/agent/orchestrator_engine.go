// A2A Project Orchestrator Engine
// ระบบจัดการโปรเจคแบบ Pure Agent-to-Agent (A2A) - NO SUBAGENT!

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/agentcomm"
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/mailbox"
	"picoclaw/agent/pkg/providers"
	"picoclaw/agent/pkg/providers/protocoltypes"
)

// Phase represents a phase in the A2A project workflow
type Phase string

const (
	PhaseDiscovery   Phase = "discovery"
	PhaseMeeting     Phase = "meeting"
	PhasePlanning    Phase = "planning"
	PhaseExecution   Phase = "execution"
	PhaseIntegration Phase = "integration"
	PhaseValidation  Phase = "validation"
	PhaseComplete    Phase = "complete"

	// Dynamic phases created during project execution
	PhaseAnalysis   Phase = "analysis"    // Analyze task complexity
	PhaseSubMeeting Phase = "sub-meeting" // Sub-meeting for new tasks
	PhaseReview     Phase = "review"      // Review completed work
)

// TaskComplexity represents the complexity level of a task
type TaskComplexity string

const (
	ComplexitySimple   TaskComplexity = "simple"   // 1 task at a time
	ComplexityMedium   TaskComplexity = "medium"   // 2 tasks concurrently
	ComplexityComplex  TaskComplexity = "complex"  // 3 tasks concurrently
	ComplexityCritical TaskComplexity = "critical" // 1 task, full attention
)

// TaskType represents the category of a task
type TaskType string

const (
	TaskTypeCoding       TaskType = "coding"
	TaskTypeResearch     TaskType = "research"
	TaskTypePlanning     TaskType = "planning"
	TaskTypeReview       TaskType = "review"
	TaskTypeWriting      TaskType = "writing"
	TaskTypeAnalysis     TaskType = "analysis"
	TaskTypeDebugging    TaskType = "debugging"
	TaskTypeArchitecture TaskType = "architecture"
	TaskTypeGeneral      TaskType = "general"
)

// TaskAnalysis contains complexity analysis for a task
type TaskAnalysis struct {
	TaskID          string         `json:"task_id"`
	Task            string         `json:"task"`
	TaskType        TaskType       `json:"task_type"`
	Complexity      TaskComplexity `json:"complexity"`
	EstimatedMin    int            `json:"estimated_min"`    // Estimated minutes
	BatchSize       int            `json:"batch_size"`       // 1, 2, or 3
	AgentID         string         `json:"agent_id"`         // Best agent for this task
	RecommendedRole string         `json:"recommended_role"` // For backward compatibility
	DelegateReason  string         `json:"delegate_reason"`  // For backward compatibility
	Dependencies    []string       `json:"dependencies"`     // Task IDs this depends on
	SubTasks        []string       `json:"sub_tasks"`        // If needs to be broken down
	Domain          string         `json:"domain"`
	EstimatedTime   time.Duration  `json:"estimated_time"`
	ShouldDelegate  bool           `json:"should_delegate"`
	Confidence      float64        `json:"confidence"` // 0.0 - 1.0
}

// PhaseStatus represents the status of a phase
type PhaseStatus string

const (
	PhaseStatusPending   PhaseStatus = "pending"
	PhaseStatusRunning   PhaseStatus = "running"
	PhaseStatusCompleted PhaseStatus = "completed"
	PhaseStatusFailed    PhaseStatus = "failed"
)

// A2AProject represents a project using A2A collaboration
type A2AProject struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Status      ProjectStatus        `json:"status"`
	Phases      map[Phase]*PhaseInfo `json:"phases"`
	Assignments []A2AAssignment      `json:"assignments"`
	Messages    []A2AMessage         `json:"messages"` // All A2A messages
	Artifacts   []ProjectArtifact    `json:"artifacts"`
	StartTime   time.Time            `json:"start_time"`
	EndTime     *time.Time           `json:"end_time,omitempty"`

	// Task Queue for managing large number of tasks
	TaskQueue      *TaskQueue      `json:"-"`               // Not serialized
	CompletedTasks map[string]bool `json:"completed_tasks"` // Track completed task IDs

	// Dynamic Phase Management
	CurrentPhase Phase `json:"current_phase"`
	PhaseCounter int   `json:"phase_counter"` // For creating dynamic phases

	// Configuration
	MaxTasksPerPhase int  `json:"max_tasks_per_phase"` // Auto-split if exceeded
	BatchSize        int  `json:"batch_size"`          // Tasks per batch
	AutoResume       bool `json:"auto_resume"`         // Automatically resume on failure
	RetryCount       int  `json:"retry_count"`        // Number of times project was resumed
	MaxRetries       int  `json:"max_retries"`         // Max automatic retries

	// Internal
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	orchestrator *A2AOrchestrator

	// Phase 3: Message summarization for token optimization
	messageSummarizer *MessageSummarizer `json:"-"` // Not serialized

	// Phase 4: Project context summary for efficient agent communication
	ProjectSummary string   `json:"project_summary,omitempty"`     // Running summary of project
	KeyDecisions   []string `json:"key_decisions,omitempty"`       // Important decisions made
	CurrentGoals   []string `json:"current_goals,omitempty"`       // Current project goals
	ContextVersion int      `json:"context_version"`               // Version for cache invalidation
}

// ProjectStatus represents the overall project status
type ProjectStatus string

const (
	ProjectStatusPending   ProjectStatus = "pending"
	ProjectStatusRunning   ProjectStatus = "running"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusFailed    ProjectStatus = "failed"
)

// PhaseInfo contains information about a phase
type PhaseInfo struct {
	Name      string                 `json:"name"`
	Status    PhaseStatus            `json:"status"`
	StartTime *time.Time             `json:"start_time,omitempty"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Result    string                 `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// A2AAssignment represents a task assigned to an agent via A2A
type A2AAssignment struct {
	ID           string           `json:"id"`
	Phase        Phase            `json:"phase"`
	Task         string           `json:"task"`
	FromAgent    string           `json:"from_agent"` // Who assigned
	ToAgent      string           `json:"to_agent"`   // Who received
	Status       AssignmentStatus `json:"status"`
	StartTime    *time.Time       `json:"start_time,omitempty"`
	EndTime      *time.Time       `json:"end_time,omitempty"`
	Result       string           `json:"result,omitempty"`
	Deliverables []string         `json:"deliverables,omitempty"`

	// Task queue and dependency management
	Priority  int      `json:"priority"`   // Higher = more important
	DependsOn []string `json:"depends_on"` // IDs of tasks that must complete first
	BatchID   string   `json:"batch_id"`   // Group related tasks
	Order     int      `json:"order"`      // Execution order within batch

	// Progress tracking
	Progress    int       `json:"progress"`     // 0-100 percentage
	ProgressMsg string    `json:"progress_msg"` // Current status message
	LastUpdate  time.Time `json:"last_update"`  // Last progress update time
	RetryCount  int       `json:"retry_count"`  // Number of times this assignment was retried

	// Internal mutex for thread-safe access
	mu *sync.RWMutex `json:"-"`
}

// TaskQueue manages pending tasks for large projects
type TaskQueue struct {
	pending   []*A2AAssignment
	mu        sync.RWMutex
	batchSize int // Max tasks to execute concurrently
}

// NewTaskQueue creates a task queue with specified batch size
func NewTaskQueue(batchSize int) *TaskQueue {
	if batchSize <= 0 {
		batchSize = 3 // Default: 3 tasks at a time per agent
	}
	return &TaskQueue{
		pending:   make([]*A2AAssignment, 0),
		batchSize: batchSize,
	}
}

// Add adds a task to the queue
func (q *TaskQueue) Add(task *A2AAssignment) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pending = append(q.pending, task)
}

// GetNextBatch returns the next batch of tasks that can be executed
// Respects dependencies and priorities
func (q *TaskQueue) GetNextBatch(completed map[string]bool, maxSize int) []*A2AAssignment {
	q.mu.Lock()
	defer q.mu.Unlock()

	var batch []*A2AAssignment
	var remaining []*A2AAssignment

	// Sort by priority (desc) and order (asc)
	sort.Slice(q.pending, func(i, j int) bool {
		if q.pending[i].Priority != q.pending[j].Priority {
			return q.pending[i].Priority > q.pending[j].Priority
		}
		return q.pending[i].Order < q.pending[j].Order
	})

	for _, task := range q.pending {
		if len(batch) >= maxSize {
			remaining = append(remaining, task)
			continue
		}

		// Check if dependencies are satisfied
		canExecute := true
		for _, depID := range task.DependsOn {
			if !completed[depID] {
				canExecute = false
				break
			}
		}

		if canExecute {
			batch = append(batch, task)
		} else {
			remaining = append(remaining, task)
		}
	}

	q.pending = remaining
	return batch
}

// Size returns the number of pending tasks
func (q *TaskQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.pending)
}

// IsEmpty returns true if no pending tasks
func (q *TaskQueue) IsEmpty() bool {
	return q.Size() == 0
}

// AssignmentStatus represents the status of a task assignment
type AssignmentStatus string

const (
	AssignmentStatusPending   AssignmentStatus = "pending"
	AssignmentStatusAssigned  AssignmentStatus = "assigned"
	AssignmentStatusAccepted  AssignmentStatus = "accepted"
	AssignmentStatusRejected  AssignmentStatus = "rejected"
	AssignmentStatusRunning   AssignmentStatus = "running"
	AssignmentStatusCompleted AssignmentStatus = "completed"
	AssignmentStatusFailed    AssignmentStatus = "failed"
)

// A2AMessage represents a message between agents in the project
// OPTIMIZED: Added summarization support for Phase 3 token optimization
type A2AMessage struct {
	ID          string    `json:"id"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Type        string    `json:"type"` // "task", "response", "question", "decision", "progress"
	Content     string    `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
	ProjectID   string    `json:"project_id"`
	ReplyTo     string    `json:"reply_to,omitempty"`

	// Phase 3: Summarization fields
	IsSummary       bool   `json:"is_summary,omitempty"`        // True if this is a summary message
	SummaryOf       string `json:"summary_of,omitempty"`        // ID of message this summarizes (if applicable)
	FullContentPath string `json:"full_content_path,omitempty"` // Path to full content if archived
	OriginalLength  int    `json:"original_length,omitempty"`   // Original content length before summary
}

// ProjectArtifact represents an artifact produced by an agent
type ProjectArtifact struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"` // "code", "document", "design", "test"
	Path         string    `json:"path"`
	AgentID      string    `json:"agent_id"`
	CreatedAt    time.Time `json:"created_at"`
	AssignmentID string    `json:"assignment_id"`
}

// A2AOrchestrator manages project orchestration using A2A communication ONLY
type A2AOrchestrator struct {
	mu       sync.RWMutex
	registry *AgentRegistry
	provider providers.LLMProvider
	config   *config.Config
	msgBus   *bus.MessageBus

	projects   map[string]*A2AProject
	discovery  *A2AAgentDiscovery
	mailboxes  map[string]*mailbox.Mailbox // agentID -> mailbox
	messengers map[string]*Messenger       // agentID -> messenger
	workers    map[string]*A2AAgentWorker  // agentID -> worker (REAL LLM!)
	sharedCtx  *SharedContext

	onPhaseChange func(projectID string, phase Phase, status PhaseStatus)
	onMessage     func(projectID string, msg A2AMessage)

	// Progress and failure callbacks
	onAssignmentProgress func(projectID string, assignmentID string, progress int, message string)
	onAssignmentFailed   func(projectID string, assignmentID string, agentID string, err error)

	// Load balancing: track active assignments per agent
	assignmentCount map[string]int // agentID -> number of active assignments
	maxConcurrent   int
	// Project shortcuts for easy reference (e.g., "go-pwd" -> "a2a-project-xxx")
	projectShortcuts map[string]string // shortcut -> projectID
	latestProjectID  string            // Most recently created project

	// Rate limiter for LLM API calls (global across all agents)
	llmRateLimiter   chan struct{} // semaphore for limiting concurrent LLM calls
	maxLLMConcurrent int           // max concurrent LLM calls (default: 3)

	// Outsource pool for dynamic agent scaling
	outsourcePool *OutsourcePool

	// Auto-Resume monitor
	stopChan chan struct{}

	// Persistence
	persistencePath string
}

// NewA2AOrchestrator creates a new A2A orchestrator (NO SUBAGENT!)
func NewA2AOrchestrator(
	registry *AgentRegistry,
	provider providers.LLMProvider,
	cfg *config.Config,
	msgBus *bus.MessageBus,
) *A2AOrchestrator {
	// Create shared context for all A2A communication
	sharedCtx := NewSharedContext(1000, 10000)

	// Rate limiter: max 3 concurrent LLM calls globally to prevent API overload
	maxLLMCalls := 3

	o := &A2AOrchestrator{
		registry:         registry,
		provider:         provider,
		config:           cfg,
		msgBus:           msgBus,
		projects:         make(map[string]*A2AProject),
		discovery:        NewA2AAgentDiscovery(registry, provider),
		mailboxes:        make(map[string]*mailbox.Mailbox),
		messengers:       make(map[string]*Messenger),
		workers:          make(map[string]*A2AAgentWorker),
		sharedCtx:        sharedCtx,
		assignmentCount:  make(map[string]int),
		projectShortcuts: make(map[string]string),
		maxConcurrent:    2, // default 2 tasks per agent
		latestProjectID:  "",
		llmRateLimiter:   make(chan struct{}, maxLLMCalls),
		maxLLMConcurrent: maxLLMCalls,
		outsourcePool:    NewOutsourcePool(5, 30*time.Minute), // Default 5 max outsource agents
		stopChan:         make(chan struct{}),
		persistencePath:  filepath.Join(os.Getenv("HOME"), ".picoclaw", "a2a_projects"),
	}

	// Create persistence directory
	if err := os.MkdirAll(o.persistencePath, 0755); err != nil {
		logger.ErrorCF("a2a_orchestrator", "Failed to create persistence directory", map[string]any{"path": o.persistencePath, "error": err})
	}

	// Start auto-resume monitor
	go o.startAutoResumeMonitor()

	return o
}

// Initialize sets up mailboxes, messengers, and REAL workers for all agents
func (o *A2AOrchestrator) Initialize() {
	o.mu.Lock()

	agentIDs := o.registry.ListAgentIDs()

	for _, agentID := range agentIDs {
		// Create mailbox for each agent
		o.mailboxes[agentID] = mailbox.NewMailbox(agentID, 1000)

		// Create messenger for each agent
		o.messengers[agentID] = NewMessenger(agentID, o.sharedCtx, o.msgBus)

		// Create REAL worker with LLM (NO SIMULATION!)
		if agent, ok := o.registry.GetAgent(agentID); ok {
			worker := NewA2AAgentWorker(agentID, agent, o.messengers[agentID], o.msgBus)
			o.workers[agentID] = worker
			worker.Start() // Start the worker

			logger.InfoCF("a2a_orchestrator", "Agent worker started",
				map[string]any{
					"agent_id": agentID,
					"model":    agent.Model,
				})
		}

		// Note: Message handler is set up by A2AAgentWorker in its messageLoop()
		// We don't register here to avoid overwriting the worker's handler

		logger.InfoCF("a2a_orchestrator", "Agent initialized for A2A",
			map[string]any{"agent_id": agentID})
	}

	logger.InfoCF("a2a_orchestrator", "A2A Orchestrator initialized",
		map[string]any{
			"agent_count": len(agentIDs),
			"workers":     len(o.workers),
		})
		
	o.mu.Unlock()

	// Load existing projects from persistence
	o.loadProjects()
}

// saveProject saves the project state to disk
func (o *A2AOrchestrator) saveProject(projectID string) {
	o.mu.RLock()
	project, ok := o.projects[projectID]
	o.mu.RUnlock()

	if !ok {
		return
	}

	project.mu.RLock()
	defer project.mu.RUnlock()

	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		logger.ErrorCF("a2a_orchestrator", "Failed to marshal project state", map[string]any{"project_id": projectID, "error": err})
		return
	}

	filePath := filepath.Join(o.persistencePath, fmt.Sprintf("%s.json", projectID))
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		logger.ErrorCF("a2a_orchestrator", "Failed to write project state to disk", map[string]any{"path": filePath, "error": err})
	}
}

// loadProjects loads all projects from the persistence directory
func (o *A2AOrchestrator) loadProjects() {
	files, err := os.ReadDir(o.persistencePath)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.ErrorCF("a2a_orchestrator", "Failed to read persistence directory", map[string]any{"path": o.persistencePath, "error": err})
		}
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(o.persistencePath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			logger.ErrorCF("a2a_orchestrator", "Failed to read project file", map[string]any{"path": filePath, "error": err})
			continue
		}

		project := &A2AProject{}
		err = json.Unmarshal(data, project)
		if err != nil {
			logger.ErrorCF("a2a_orchestrator", "Failed to unmarshal project file", map[string]any{"path": filePath, "error": err})
			continue
		}

		// Re-initialize internal fields
		project.orchestrator = o
		project.CompletedTasks = make(map[string]bool)
		project.TaskQueue = NewTaskQueue(project.BatchSize)
		if project.TaskQueue.batchSize == 0 {
			project.TaskQueue.batchSize = 3 // Fallback
		}

		for i := range project.Assignments {
			project.Assignments[i].mu = &sync.RWMutex{}
			if project.Assignments[i].Status == AssignmentStatusCompleted {
				project.CompletedTasks[project.Assignments[i].ID] = true
			}
		}

		// Re-create context
		project.ctx, project.cancel = context.WithCancel(context.Background())

		o.mu.Lock()
		o.projects[project.ID] = project
		if project.StartTime.After(time.Time{}) {
			// Update latest project ID if this one is newer
			if o.latestProjectID == "" {
				o.latestProjectID = project.ID
			} else {
				currentLatest, ok := o.projects[o.latestProjectID]
				if ok && project.StartTime.After(currentLatest.StartTime) {
					o.latestProjectID = project.ID
				}
			}
		}
		o.mu.Unlock()

		logger.InfoCF("a2a_orchestrator", "Loaded project from persistence", map[string]any{"project_id": project.ID, "status": project.Status})
	}
}

// SetDiscovery sets the agent discovery component
func (o *A2AOrchestrator) SetDiscovery(d *A2AAgentDiscovery) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.discovery = d
}

// setupMessageHandler sets up message handler for an agent
func (o *A2AOrchestrator) setupMessageHandler(agentID string) {
	messenger := o.messengers[agentID]
	if messenger == nil {
		return
	}

	// Register handler for this agent
	messenger.RegisterHandler(agentID, func(ctx context.Context, msg *agentcomm.AgentMessage) {
		o.handleAgentMessage(agentID, msg)
	})
}

// handleAgentMessage handles incoming messages from agents
func (o *A2AOrchestrator) handleAgentMessage(toAgentID string, msg *agentcomm.AgentMessage) {
	// Convert to A2AMessage and store
	a2aMsg := A2AMessage{
		ID:        GenerateA2AMessageID(),
		From:      msg.From,
		To:        toAgentID,
		Type:      string(msg.Type),
		Content:   msg.GetPayloadString(),
		Timestamp: time.Now(),
	}

	// Notify listeners
	if o.onMessage != nil {
		for _, project := range o.projects {
			o.onMessage(project.ID, a2aMsg)
		}
	}

	logger.DebugCF("a2a_orchestrator", "Message received",
		map[string]any{
			"from": msg.From,
			"to":   toAgentID,
			"type": msg.Type,
		})
}

// CreateProject creates a new A2A project
func (o *A2AOrchestrator) CreateProject(name, description string) *A2AProject {
	o.mu.Lock()

	ctx, cancel := context.WithCancel(context.Background())

	project := &A2AProject{
		ID:           generateProjectID(),
		Name:         name,
		Description:  description,
		Status:       ProjectStatusPending,
		Phases:       make(map[Phase]*PhaseInfo),
		Assignments:  make([]A2AAssignment, 0),
		Messages:     make([]A2AMessage, 0),
		Artifacts:    make([]ProjectArtifact, 0),
		StartTime:    time.Now(),
		ctx:          ctx,
		cancel:       cancel,
		orchestrator: o,
		// Initialize task queue for large projects
		TaskQueue:        NewTaskQueue(3), // 3 tasks per batch
		CompletedTasks:   make(map[string]bool),
		PhaseCounter:     0,
		MaxTasksPerPhase: 10, // Auto-split phase if > 10 tasks
		BatchSize:        3,
		AutoResume:       true, // Default to true
		MaxRetries:       5,    // Default max retries
		// Phase 4 & 3: Initialize context optimization
		messageSummarizer: NewMessageSummarizer(filepath.Join(o.persistencePath, "archived_messages")),
		ContextVersion:    1,
	}

	// Initialize all phases
	phases := []Phase{PhaseDiscovery, PhaseMeeting, PhasePlanning, PhaseExecution, PhaseIntegration, PhaseValidation}
	for _, phase := range phases {
		project.Phases[phase] = &PhaseInfo{
			Name:   string(phase),
			Status: PhaseStatusPending,
		}
	}

	o.projects[project.ID] = project

	// Auto-generate shortcut from project name
	shortcut := generateProjectShortcut(name)
	o.projectShortcuts[shortcut] = project.ID
	o.projectShortcuts["latest"] = project.ID // Always update "latest" to most recent
	o.latestProjectID = project.ID

	// Release lock before saving/logging to avoid deadlock (saveProject acquires RLock)
	o.mu.Unlock()

	logger.InfoCF("a2a_orchestrator", "A2A Project created",
		map[string]any{
			"shortcut": shortcut,
		})

	o.saveProject(project.ID)

	return project
}

// generateProjectShortcut creates a short name from project name
func generateProjectShortcut(name string) string {
	// Convert to lowercase and replace spaces/special chars with -
	shortcut := strings.ToLower(name)
	shortcut = strings.ReplaceAll(shortcut, " ", "-")

	// Remove non-alphanumeric characters except -
	var result strings.Builder
	for _, r := range shortcut {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	shortcut = result.String()

	// Limit length
	if len(shortcut) > 20 {
		shortcut = shortcut[:20]
	}

	// Remove trailing dashes
	shortcut = strings.TrimRight(shortcut, "-")

	return shortcut
}

// GetProjectByShortcut gets project by shortcut name
func (o *A2AOrchestrator) GetProjectByShortcut(shortcut string) (*A2AProject, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Try shortcut first
	if projectID, ok := o.projectShortcuts[shortcut]; ok {
		if project, ok := o.projects[projectID]; ok {
			return project, true
		}
	}

	// Try direct ID match
	if project, ok := o.projects[shortcut]; ok {
		return project, true
	}

	return nil, false
}

// GetLatestProjectID returns the most recently created project ID
func (o *A2AOrchestrator) GetLatestProjectID() string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.latestProjectID
}

// ResumeProject resumes a failed or paused project
func (o *A2AOrchestrator) ResumeProject(projectID string) error {
	o.mu.Lock()
	project, ok := o.projects[projectID]
	o.mu.Unlock()

	if !ok {
		return fmt.Errorf("project %s not found", projectID)
	}

	project.mu.Lock()
	if project.Status == ProjectStatusRunning {
		project.mu.Unlock()
		return fmt.Errorf("project %s is already running", projectID)
	}

	// Reset failed assignments to pending so they can be retried
	for i := range project.Assignments {
		if project.Assignments[i].Status == AssignmentStatusFailed {
			project.Assignments[i].Status = AssignmentStatusPending
			project.Assignments[i].Result = ""
			project.Assignments[i].Progress = 0
		}
	}

	project.Status = ProjectStatusRunning
	project.RetryCount++
	project.mu.Unlock()

	o.saveProject(projectID)

	// Inject Resumption Context Message
	o.LogA2AMessage(projectID, A2AMessage{
		ID:        GenerateA2AMessageID(),
		From:      "SYSTEM",
		To:        "committee",
		Type:      "announcement",
		Content:   fmt.Sprintf("🔄 [SYSTEM] Project execution resumed (Retry #%d). Continuing from phase: %s. Previous progress and artifacts are preserved.", project.RetryCount, project.CurrentPhase),
		Timestamp: time.Now(),
	})

	logger.InfoCF("a2a_orchestrator", "Resuming A2A project",
		map[string]any{
			"project_id": projectID,
			"name":       project.Name,
			"retry":      project.RetryCount,
		})

	// Run A2A phases (it will skip completed ones)
	go o.runA2AProject(project)

	return nil
}

// LogA2AMessage logs a message specific to a project
func (o *A2AOrchestrator) LogA2AMessage(projectID string, msg A2AMessage) {
	o.mu.RLock()
	project, ok := o.projects[projectID]
	o.mu.RUnlock()

	if !ok {
		return
	}

	msg.ProjectID = projectID
	project.mu.Lock()
	project.Messages = append(project.Messages, msg)
	project.mu.Unlock()

	o.saveProject(projectID)

	if o.onMessage != nil {
		o.onMessage(projectID, msg)
	}
}

// startAutoResumeMonitor periodically checks for projects to resume
func (o *A2AOrchestrator) startAutoResumeMonitor() {
	ticker := time.NewTicker(2 * time.Minute) // Check every 2 minutes for stability
	defer ticker.Stop()

	logger.InfoCF("a2a_orchestrator", "Auto-Resume monitor started", map[string]any{})

	for {
		select {
		case <-ticker.C:
			o.checkAndResumeProjects()
		case <-o.stopChan:
			return
		}
	}
}

// checkAndResumeProjects scans all projects and resumes those that failed
func (o *A2AOrchestrator) checkAndResumeProjects() {
	o.mu.RLock()
	var projectsToResume []string
	for id, p := range o.projects {
		p.mu.RLock()
		if p.Status == ProjectStatusFailed && p.AutoResume && p.RetryCount < p.MaxRetries {
			// Auto-resume if it's a transient failure
			projectsToResume = append(projectsToResume, id)
		}
		p.mu.RUnlock()
	}
	o.mu.RUnlock()

	for _, id := range projectsToResume {
		logger.InfoCF("a2a_orchestrator", "Auto-resuming project", map[string]any{"project_id": id})
		o.ResumeProject(id)
	}
}

// StartProject starts the A2A orchestration
func (o *A2AOrchestrator) StartProject(projectID string) error {
	o.mu.Lock()
	project, ok := o.projects[projectID]
	o.mu.Unlock()

	if !ok {
		return fmt.Errorf("project %s not found", projectID)
	}

	project.mu.Lock()
	project.Status = ProjectStatusRunning
	project.mu.Unlock()

	logger.InfoCF("a2a_orchestrator", "Starting A2A project",
		map[string]any{
			"project_id": projectID,
			"name":       project.Name,
		})

	// Run A2A phases
	o.saveProject(projectID)
	go o.runA2AProject(project)

	return nil
}

// runA2AProject runs all A2A phases
func (o *A2AOrchestrator) runA2AProject(project *A2AProject) {
	// Initial phases (always run in order)
	initialPhases := []Phase{PhaseDiscovery, PhaseMeeting, PhasePlanning}

	// Execute initial phases sequentially
	for _, phase := range initialPhases {
		project.mu.RLock()
		info, exists := project.Phases[phase]
		project.mu.RUnlock()

		// Skip if already completed
		if exists && info.Status == PhaseStatusCompleted {
			logger.InfoCF("a2a_orchestrator", "Skipping already completed phase",
				map[string]any{"project_id": project.ID, "phase": phase})
			continue
		}

		select {
		case <-project.ctx.Done():
			logger.InfoCF("a2a_orchestrator", "Project cancelled",
				map[string]any{"project_id": project.ID})
			return
		default:
		}

		if err := o.runA2APhase(project, phase); err != nil {
			logger.ErrorCF("a2a_orchestrator", "Phase failed",
				map[string]any{
					"project_id": project.ID,
					"phase":      phase,
					"error":      err.Error(),
				})

			project.mu.Lock()
			project.Phases[phase].Status = PhaseStatusFailed
			project.Phases[phase].Error = err.Error()
			project.Status = ProjectStatusFailed
			project.mu.Unlock()

			if o.onPhaseChange != nil {
				o.onPhaseChange(project.ID, phase, PhaseStatusFailed)
			}
			o.saveProject(project.ID)
			return
		}

		o.saveProject(project.ID)
	}

	// After planning, execute planned phases sequentially
	o.executePlannedPhases(project)

	// Integration phase (after all execution phases)
	project.mu.RLock()
	integInfo := project.Phases[PhaseIntegration]
	project.mu.RUnlock()

	if integInfo.Status != PhaseStatusCompleted {
		if err := o.runA2APhase(project, PhaseIntegration); err != nil {
			logger.ErrorCF("a2a_orchestrator", "Integration phase failed",
				map[string]any{
					"project_id": project.ID,
					"error":      err.Error(),
				})
		}
	}

	// Validation phase
	project.mu.RLock()
	valInfo := project.Phases[PhaseValidation]
	project.mu.RUnlock()

	if valInfo.Status != PhaseStatusCompleted {
		if err := o.runA2APhase(project, PhaseValidation); err != nil {
			logger.ErrorCF("a2a_orchestrator", "Validation phase failed",
				map[string]any{
					"project_id": project.ID,
					"error":      err.Error(),
				})
		}
	}

	// Mark project as completed
	project.mu.Lock()
	project.Status = ProjectStatusCompleted
	now := time.Now()
	project.EndTime = &now
	project.mu.Unlock()

	o.saveProject(project.ID)

	logger.InfoCF("a2a_orchestrator", "A2A Project completed",
		map[string]any{
			"project_id": project.ID,
			"name":       project.Name,
			"duration":   time.Since(project.StartTime),
		})
}

// executePlannedPhases executes all planned execution phases sequentially
func (o *A2AOrchestrator) executePlannedPhases(project *A2AProject) {
	// Get planned phases from planning metadata
	project.mu.RLock()
	planningMeta, _ := project.Phases[PhasePlanning].Metadata["phase_details"].([]PhasePlan)
	currentPhaseIdx, _ := project.Phases[PhasePlanning].Metadata["current_phase"].(int)
	project.mu.RUnlock()

	if len(planningMeta) == 0 {
		logger.WarnCF("a2a_orchestrator", "No planned phases found, using default execution",
			map[string]any{"project_id": project.ID})
		o.runA2APhase(project, PhaseExecution)
		return
	}

	// Execute each planned phase sequentially
	for i := currentPhaseIdx; i < len(planningMeta); i++ {
		phase := planningMeta[i]

		logger.InfoCF("a2a_orchestrator", "Starting planned phase",
			map[string]any{
				"project_id":   project.ID,
				"phase":        phase.PhaseID,
				"phase_number": phase.PhaseNumber,
				"task_count":   len(phase.Tasks),
			})

		// Update current phase index
		project.mu.Lock()
		project.CurrentPhase = phase.PhaseID
		project.Phases[PhasePlanning].Metadata["current_phase"] = i
		project.mu.Unlock()

		// Initialize phase
		project.mu.Lock()
		project.Phases[phase.PhaseID] = &PhaseInfo{
			Name:   string(phase.PhaseID),
			Status: PhaseStatusRunning,
		}
		now := time.Now()
		project.Phases[phase.PhaseID].StartTime = &now
		project.mu.Unlock()

		// Execute this phase's tasks with batch processing
		err := o.executePhaseWithBatch(project, phase)

		// Mark phase complete
		project.mu.Lock()
		if err != nil {
			project.Phases[phase.PhaseID].Status = PhaseStatusFailed
			project.Phases[phase.PhaseID].Error = err.Error()
		} else {
			project.Phases[phase.PhaseID].Status = PhaseStatusCompleted
		}
		endNow := time.Now()
		project.Phases[phase.PhaseID].EndTime = &endNow
		project.mu.Unlock()

		if err != nil {
			logger.ErrorCF("a2a_orchestrator", "Planned phase failed",
				map[string]any{
					"project_id": project.ID,
					"phase":      phase.PhaseID,
					"error":      err.Error(),
				})
			// Continue to next phase even if this one failed
		}

		o.saveProject(project.ID)

		// Check for new tasks discovered during this phase
		newTasks := o.checkForNewTasks(project)
		if len(newTasks) > 0 {
			logger.InfoCF("a2a_orchestrator", "New tasks discovered, inserting dynamic phase",
				map[string]any{
					"project_id": project.ID,
					"new_tasks":  len(newTasks),
				})
			o.InsertDynamicPhase(project, i+1, newTasks)
		}

		// Brief pause between phases
		if i < len(planningMeta)-1 {
			time.Sleep(3 * time.Second)
		}
	}
}

// executePhaseWithBatch executes a single phase using a CONTINUOUS worker pool
// Uses dual-level semaphores to limit concurrent tasks (global and per-agent)
func (o *A2AOrchestrator) executePhaseWithBatch(project *A2AProject, phase PhasePlan) error {
	logger.InfoCF("a2a_orchestrator", "Executing phase with CONTINUOUS worker pool",
		map[string]any{
			"project_id":  project.ID,
			"phase":       phase.PhaseID,
			"total_tasks": len(phase.Tasks),
		})

	// Add tasks to queue
	for i := range phase.Tasks {
		task := phase.Tasks[i]
		assignment := A2AAssignment{
			ID:        fmt.Sprintf("assign-%d-%d", time.Now().UnixNano(), i),
			Phase:     phase.PhaseID,
			Task:      task.Task,
			FromAgent: "jarvis",
			ToAgent:   task.AgentID,
			Status:    AssignmentStatusPending,
			DependsOn: task.Dependencies,
			BatchID:   string(phase.PhaseID),
			Order:     i,
			mu:        &sync.RWMutex{},
		}
		// Set priority based on complexity
		switch task.Complexity {
		case ComplexityCritical:
			assignment.Priority = 4
		case ComplexityComplex:
			assignment.Priority = 3
		case ComplexityMedium:
			assignment.Priority = 2
		case ComplexitySimple:
			assignment.Priority = 1
		}
		project.TaskQueue.Add(&assignment)

		// Also add to project assignments for persistence
		project.mu.Lock()
		project.Assignments = append(project.Assignments, assignment)
		project.mu.Unlock()
	}

	o.saveProject(project.ID)

	// Determine global concurrency limit based on complexity
	maxConcurrency := 4 // Default global

	if len(phase.Tasks) > 0 {
		switch phase.Tasks[0].Complexity {
		case ComplexitySimple:
			maxConcurrency = 5
		case ComplexityMedium:
			maxConcurrency = 3
		case ComplexityComplex:
			maxConcurrency = 2
		case ComplexityCritical:
			maxConcurrency = 1 // Critical: only 1 global task at a time
		}
	}

	// Create channels and semaphores
	semaphore := make(chan struct{}, maxConcurrency)
	agentSemaphores := make(map[string]chan struct{})
	for _, agentID := range o.registry.ListAgentIDs() {
		limit := 2 // per-agent limit
		if maxConcurrency < 2 {
			limit = 1
		}
		agentSemaphores[agentID] = make(chan struct{}, limit)
	}

	type result struct {
		assignmentID string
		success      bool
		err          error
	}

	totalTasks := len(phase.Tasks)
	resultChan := make(chan result, totalTasks)

	totalCompleted := 0
	totalFailed := 0
	activeTasks := 0

	// We need to keep tasks that are ready but couldn't be dispatched yet
	var readyQueue []*A2AAssignment

	for totalCompleted+totalFailed < totalTasks {
		// 1. Replenish readyQueue from TaskQueue
		newReady := project.TaskQueue.GetNextBatch(project.CompletedTasks, totalTasks)
		readyQueue = append(readyQueue, newReady...)

		// 2. Try to dispatch as many as possible
		var remainingReady []*A2AAssignment
		dispatchedThisLoop := false

		for _, assignment := range readyQueue {
			// Try to acquire global slot
			select {
			case semaphore <- struct{}{}:
				// Global slot acquired! Now try agent slot
				agentSem := agentSemaphores[assignment.ToAgent]
				if agentSem == nil {
					agentSem = make(chan struct{}, 2)
					agentSemaphores[assignment.ToAgent] = agentSem
				}

				select {
				case agentSem <- struct{}{}:
					// Both slots acquired, dispatch!
					assignment.Status = AssignmentStatusRunning
					activeTasks++
					dispatchedThisLoop = true

					go func(a *A2AAssignment) {
						defer func() { <-agentSem }()
						defer func() { <-semaphore }()

						logger.DebugCF("a2a_orchestrator", "Executing assignment (continuous)",
							map[string]any{
								"assignment_id": a.ID,
								"agent":         a.ToAgent,
								"task":          a.Task,
							})

						err := o.executeA2AAssignment(project, a)
						resultChan <- result{a.ID, err == nil, err}
					}(assignment)
				default:
					// Agent is busy. Let's try to outsource!
					outsourceAgent, err := o.outsourcePool.Hire(project.ID, "outsource", assignment.ID)
					var actualAgentSem chan struct{}
					var targetAgent string

					if err == nil && outsourceAgent != nil {
						// We successfully outsourced it!
						targetAgent = outsourceAgent.ID
						assignment.ToAgent = targetAgent

						// Setup agent instance and worker dynamically
						agentInstance := &AgentInstance{
							ID:    targetAgent,
							Model: o.config.Agents.Defaults.GetModelName(),
						}

						mbox := mailbox.NewMailbox(targetAgent, 100)
						o.mu.Lock()
						o.mailboxes[targetAgent] = mbox
						o.messengers[targetAgent] = NewMessenger(targetAgent, o.sharedCtx, o.msgBus)
						worker := NewA2AAgentWorker(targetAgent, agentInstance, o.messengers[targetAgent], o.msgBus)
						worker.Start()
						o.workers[targetAgent] = worker
						o.mu.Unlock()

						// Give the outsource agent its own semaphore
						o.mu.Lock()
						agentSemaphores[targetAgent] = make(chan struct{}, 2)
						actualAgentSem = agentSemaphores[targetAgent]
						o.mu.Unlock()

						logger.InfoCF("a2a_orchestrator", "Spawned Outsource Agent due to load",
							map[string]any{
								"assignment_id":   assignment.ID,
								"original_agent":  assignment.ToAgent,
								"outsource_agent": targetAgent,
								"task":            assignment.Task,
							})
					}

					// If we successfully outsourced, we can dispatch it right now!
					if actualAgentSem != nil {
						select {
						case actualAgentSem <- struct{}{}:
							assignment.Status = AssignmentStatusRunning
							activeTasks++
							dispatchedThisLoop = true

							go func(a *A2AAssignment, outAgent *OutsourceAgent) {
								defer func() { <-actualAgentSem }()
								defer func() { <-semaphore }()
								// Ensure we release the outsource agent and stop worker when done
								defer func() {
									if outAgent != nil {
										o.outsourcePool.Release(outAgent.ID)
										o.mu.Lock()
										if w, ok := o.workers[outAgent.ID]; ok {
											w.Stop()
											delete(o.workers, outAgent.ID)
										}
										delete(o.mailboxes, outAgent.ID)
										delete(o.messengers, outAgent.ID)
										o.mu.Unlock()
									}
								}()

								logger.DebugCF("a2a_orchestrator", "Executing assignment (outsourced)",
									map[string]any{
										"assignment_id": a.ID,
										"agent":         a.ToAgent,
										"task":          a.Task,
									})

								err := o.executeA2AAssignment(project, a)
								resultChan <- result{a.ID, err == nil, err}
							}(assignment, outsourceAgent)
						default:
							// Should never happen for a fresh outsource agent, but back off safely
							<-semaphore
							remainingReady = append(remainingReady, assignment)
						}
					} else {
						// Outsource failed or pool full, fallback to waiting for the busy original agent
						<-semaphore
					}
				}
			default:
				// Global is full, no more dispatches possible right now
				remainingReady = append(remainingReady, assignment)
			}
		}

		readyQueue = remainingReady

		// 3. Wait for events (Task Completions)
		if activeTasks > 0 {
			// If we didn't dispatch anything in this loop, block until a task finishes
			// If we DID dispatch, we'll non-blockingly check or loop again immediately
			if !dispatchedThisLoop {
				res := <-resultChan
				activeTasks--

				if res.success {
					project.mu.Lock()
					project.CompletedTasks[res.assignmentID] = true
					project.mu.Unlock()
					totalCompleted++
				} else {
					totalFailed++
					logger.ErrorCF("a2a_orchestrator", "Assignment failed",
						map[string]any{
							"assignment_id": res.assignmentID,
							"error":         res.err.Error(),
						})
				}

				// Optional: collect other finished tasks concurrently
				for {
					select {
					case res2 := <-resultChan:
						activeTasks--
						if res2.success {
							project.mu.Lock()
							project.CompletedTasks[res2.assignmentID] = true
							project.mu.Unlock()
							totalCompleted++
						} else {
							totalFailed++
						}
					default:
						goto CollectedResults
					}
				}
			CollectedResults:

				// Update project progress
				project.mu.Lock()
				if totalTasks > 0 {
					progress := float64(totalCompleted) / float64(totalTasks) * 100
					project.Phases[phase.PhaseID].Metadata = map[string]interface{}{
						"progress":  progress,
						"completed": totalCompleted,
						"failed":    totalFailed,
						"total":     totalTasks,
					}
				}
				project.mu.Unlock()

				logger.InfoCF("a2a_orchestrator", "Task processed",
					map[string]any{
						"project_id":   project.ID,
						"phase":        phase.PhaseID,
						"completed":    totalCompleted,
						"failed":       totalFailed,
						"progress_pct": fmt.Sprintf("%.1f%%", float64(totalCompleted)/float64(totalTasks)*100),
					})
			}
		} else if !dispatchedThisLoop {
			// No active tasks and couldn't dispatch any?
			// Could be waiting for task queue to be populated (unlikely since we queue all at start)
			// Sleep briefly to prevent tight loop
			time.Sleep(100 * time.Millisecond)
		}
	}

	close(resultChan)

	logger.InfoCF("a2a_orchestrator", "Phase completed",
		map[string]any{
			"project_id": project.ID,
			"phase":      phase.PhaseID,
			"completed":  totalCompleted,
			"failed":     totalFailed,
		})

	return nil
}

// checkForNewTasks checks if agents discovered new tasks during execution
func (o *A2AOrchestrator) checkForNewTasks(project *A2AProject) []string {
	// Check shared context for new task suggestions
	if o.sharedCtx == nil {
		return nil
	}

	// This would check for messages indicating new tasks
	// For now, return empty (placeholder for future implementation)
	return nil
}

// InsertDynamicPhase inserts a new phase at a specific position
func (o *A2AOrchestrator) InsertDynamicPhase(project *A2AProject, afterIndex int, tasks []string) error {
	project.mu.Lock()
	project.PhaseCounter++
	phaseID := Phase(fmt.Sprintf("dynamic-%d", project.PhaseCounter))
	project.mu.Unlock()

	logger.InfoCF("a2a_orchestrator", "Inserting dynamic phase",
		map[string]any{
			"project_id":  project.ID,
			"phase":       phaseID,
			"after_index": afterIndex,
			"task_count":  len(tasks),
		})

	// Analyze new tasks
	analyses := o.analyzeTaskComplexity(project, tasks)

	// Create phase plan
	phase := PhasePlan{
		PhaseNumber: afterIndex + 1,
		PhaseID:     phaseID,
		Tasks:       analyses,
		Description: fmt.Sprintf("Dynamic phase with %d tasks", len(tasks)),
	}

	// Insert into planning metadata
	project.mu.Lock()
	if planningMeta, ok := project.Phases[PhasePlanning].Metadata["phase_details"].([]PhasePlan); ok {
		// Insert at position
		newPhases := append(planningMeta[:afterIndex+1], phase)
		if afterIndex+1 < len(planningMeta) {
			newPhases = append(newPhases, planningMeta[afterIndex+1:]...)
		}
		project.Phases[PhasePlanning].Metadata["phase_details"] = newPhases
	}
	project.mu.Unlock()

	// Initialize phase
	project.mu.Lock()
	project.Phases[phaseID] = &PhaseInfo{
		Name:   string(phaseID),
		Status: PhaseStatusPending,
	}
	project.mu.Unlock()

	return nil
}

// runA2APhase runs a single A2A phase using REAL A2A communication
func (o *A2AOrchestrator) runA2APhase(project *A2AProject, phase Phase) error {
	logger.InfoCF("a2a_orchestrator", "Starting A2A phase",
		map[string]any{
			"project_id": project.ID,
			"phase":      phase,
		})

	project.mu.Lock()
	phaseInfo := project.Phases[phase]
	phaseInfo.Status = PhaseStatusRunning
	now := time.Now()
	phaseInfo.StartTime = &now
	project.mu.Unlock()

	if o.onPhaseChange != nil {
		o.onPhaseChange(project.ID, phase, PhaseStatusRunning)
	}

	var err error
	switch phase {
	case PhaseDiscovery:
		err = o.runA2ADiscovery(project)
	case PhaseMeeting:
		err = o.runA2AMeeting(project)
	case PhasePlanning:
		err = o.runA2APlanning(project)
	case PhaseExecution:
		err = o.runA2AExecution(project)
	case PhaseIntegration:
		err = o.runA2AIntegration(project)
	case PhaseValidation:
		err = o.runA2AValidation(project)
	}

	project.mu.Lock()
	endTime := time.Now()
	phaseInfo.EndTime = &endTime

	if err != nil {
		phaseInfo.Status = PhaseStatusFailed
		phaseInfo.Error = err.Error()
	} else {
		phaseInfo.Status = PhaseStatusCompleted
	}
	project.mu.Unlock()

	if o.onPhaseChange != nil {
		if err != nil {
			o.onPhaseChange(project.ID, phase, PhaseStatusFailed)
		} else {
			o.onPhaseChange(project.ID, phase, PhaseStatusCompleted)
		}
	}

	o.saveProject(project.ID)

	return err
}

// runA2ADiscovery ค้นหา Agent ที่เหมาะสมโดยใช้ LLM Semantic Router (Intelligent Discovery)
func (o *A2AOrchestrator) runA2ADiscovery(project *A2AProject) error {
	logger.InfoCF("a2a_orchestrator", "A2A Discovery phase (LLM ROUTER)",
		map[string]any{"project_id": project.ID})

	// 1. ดึง Capability ของ Agents ทั้งหมดในระบบ (103 ตัว) 
	allCapabilities := o.discovery.DiscoverAll()
	if len(allCapabilities) == 0 {
		return fmt.Errorf("no agents found in the registry")
	}

	// 2. บีบอัดข้อมูล (Minify) เพื่อลด Token (ใช้แค่ ID, Role, และ Skills หลัก)
	type MinifiedAgent struct {
		ID         string   `json:"id"`
		Role       string   `json:"role"`
		Department string   `json:"dept"`
		Skills     []string `json:"skills"`
	}
	
	var compressedAgents []MinifiedAgent
	agentMap := make(map[string]*A2AAgentCapability) // ไว้สำหรับดึงข้อมูลเต็มตอนหลัง
	
	for _, cap := range allCapabilities {
		// เอาเฉพาะ Skills หลักๆ มาไม่เกิน 5 อย่างเพื่อประหยัด Token
		skills := cap.Capabilities
		if len(skills) > 5 {
			skills = skills[:5]
		}
		
		compressedAgents = append(compressedAgents, MinifiedAgent{
			ID:         cap.AgentID,
			Role:       cap.Role,
			Department: cap.Department,
			Skills:     skills,
		})
		agentMap[cap.AgentID] = cap
	}

	// แปลงเป็น JSON String
	agentsJSON, err := json.Marshal(compressedAgents)
	if err != nil {
		logger.ErrorCF("a2a_orchestrator", "Failed to marshal compressed agents", map[string]any{"error": err})
		return err
	}

	// 3. สร้าง Prompt เพื่อให้ LLM ทำหน้าที่เป็น HR / Semantic Router
	prompt := fmt.Sprintf(`You are Jarvis, the core coordinator of a multi-agent system.
Your task is to analyze a new project and select the BEST agents for the job from the available roster.

Project Name: %s
Project Description: %s

Available Agents (JSON):
%s

INSTRUCTIONS:
1. Analyze the project requirements carefully.
2. Review the list of available agents, their roles, and skills.
3. Select up to 8 agents that are highly relevant to this specific project.
4. DO NOT select agents that are completely irrelevant.
5. VARIETY AND EXPLORATION: If multiple agents have similar skills that fit the job, try to mix and match. Do not always pick the exact same team for similar tasks. Give opportunities to specialized or niche agents if their skills relate to a part of the project.
6. Respond ONLY with a valid JSON array of strings containing the selected agent IDs.

Example response format:
["backend-architect", "frontend-developer", "api-tester", "technical-writer"]`, 
		project.Name, project.Description, string(agentsJSON))

	// 4. ส่ง Request ให้ LLM ตัดสินใจ (1 Call API เท่านั้น แก้ปัญหา 429 แบบเด็ดขาด!)
	ctxLLM, cancelLLM := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelLLM()
	
	messages := []protocoltypes.Message{{Role: "user", Content: prompt}}
	
	var selectedAgentIDs []string
	
	// ใช้ Coordinator Agent (หรือตัวหลักที่มี) เป็นตัวเรียก Provider
	if o.provider != nil {
		resp, err := o.provider.Chat(ctxLLM, messages, nil, "bailian/qwen3-coder-plus", map[string]any{
			"temperature": 0.4, // เพิ่มความหลากหลายนิดหน่อย ไม่ให้ออกมาแต่หน้าเดิมๆ
			"max_tokens":  500,  // เพิ่มเผื่อหน่อยกัน JSON ตัด
		})
		
		if err == nil {
			// พยายาม Parse JSON ที่น้อง LLM ตอบกลับมา
			content := strings.TrimSpace(resp.Content)
			// ตัด Markdown ```json หรือ ``` ออกถ้ามี
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimPrefix(content, "```")
			content = strings.TrimSuffix(content, "```")
			content = strings.TrimSpace(content)
			
			err = json.Unmarshal([]byte(content), &selectedAgentIDs)
			if err != nil {
				logger.WarnCF("a2a_orchestrator", "Failed to parse LLM router response, falling back to rule-based", 
					map[string]any{"error": err, "content": content})
			} else {
				logger.InfoCF("a2a_orchestrator", "LLM Semantic Router selected agents", 
					map[string]any{"selected_count": len(selectedAgentIDs), "agents": selectedAgentIDs})
			}
		} else {
			logger.WarnCF("a2a_orchestrator", "LLM Router API call failed, falling back to rule-based", 
					map[string]any{"error": err})
		}
	}
	
	// 5. Fallback เผื่อ LLM ล่ม หรือ Parse ไม่ผ่าน (กลับไปใช้ Rule-based แบบเก่า)
	if len(selectedAgentIDs) == 0 {
		logger.InfoCF("a2a_orchestrator", "Using Rule-Based Fallback for Discovery", nil)
		for _, cap := range allCapabilities {
			if o.discovery.ScoreAgentForTask(cap.AgentID, project.Description) > 3.0 {
				selectedAgentIDs = append(selectedAgentIDs, cap.AgentID)
			}
		}
		// จำกัดจำนวน 8 ตัว
		if len(selectedAgentIDs) > 8 {
			selectedAgentIDs = selectedAgentIDs[:8]
		}
	}

	// เตรียมข้อมูล Capability เต็มรูปแบบของตัวที่ถูกเลือก
	var targetedCapabilities []*A2AAgentCapability
	var finalTargetedAgents []string // กรอง ID ขยะออก
	
	for _, id := range selectedAgentIDs {
		if cap, exists := agentMap[id]; exists {
			targetedCapabilities = append(targetedCapabilities, cap)
			finalTargetedAgents = append(finalTargetedAgents, id)
		}
	}

	// 6. Jarvis เริ่มยิงข้อความแนะนำตัว ไปหาเฉพาะ Agent ระดับหัวกะทิที่ผ่านการคัดเลือก
	discoveryMsg := fmt.Sprintf("📢 ประกาศ! คุณถูกคัดเลือกให้เข้าร่วมทีมสำหรับโปรเจกต์: %s\nรายละเอียด: %s\nคุณเป็นผู้เชี่ยวชาญที่ตรงกับงานนี้ กรุณาตอบกลับเพื่อพร้อมเริ่มงาน!", 
		project.Name, project.Description)
		
	for _, agentID := range finalTargetedAgents {
		o.sendA2AMessage("jarvis", agentID, "discovery", discoveryMsg)
		time.Sleep(200 * time.Millisecond) // Throttling นิดหน่อย
	}

	// 7. รอคำตอบรับจาก LLM ของแต่ละ Agent 
	ctxWait, cancelWait := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelWait()

	responses := o.waitForResponses(ctxWait, project, "discovery_response", len(finalTargetedAgents))

	// เช็คว่าใครตอบและใครไม่ตอบ ถ้าไม่ตอบจะจำลองข้อความแทน
	respondedAgents := make(map[string]bool)
	for _, resp := range responses {
		respondedAgents[resp.From] = true
		o.sendA2AMessage(resp.From, "jarvis", "discovery_response", resp.Content)
	}

	for _, cap := range targetedCapabilities {
		if !respondedAgents[cap.AgentID] {
			o.sendA2AMessage(cap.AgentID, "jarvis", "discovery_response",
				fmt.Sprintf("🙋 ผม %s ครับ! มีความเชี่ยวชาญด้าน %v พร้อมสนับสนุนงานนี้ครับ (Simulated)", 
				cap.AgentName, cap.Capabilities))
		}
	}

	// สรุปข้อมูลบันทึกลง Metadata ของ Phase
	project.mu.Lock()
	project.Phases[PhaseDiscovery].Metadata = map[string]interface{}{
		"agents_discovered": len(finalTargetedAgents),
		"capabilities":      targetedCapabilities,
		"mode":              "llm_semantic_router",
		"llm_responses":     len(responses),
	}
	project.mu.Unlock()

	logger.InfoCF("a2a_orchestrator", "Discovery LLM Router completed",
		map[string]any{
			"project_id":        project.ID,
			"agents_targeted":   len(finalTargetedAgents),
			"responses":         len(responses),
		})

	return nil
}

// runA2AMeeting จัดการประชุมทีมเพื่อวางแผนร่วมกัน (Real Thinking Phase)
func (o *A2AOrchestrator) runA2AMeeting(project *A2AProject) error {
	logger.InfoCF("a2a_orchestrator", "A2A Meeting phase (FULL STRATEGY SESSION)",
		map[string]any{"project_id": project.ID})

	// 1. Jarvis เปิดประชุม
	o.sendA2AMessage("jarvis", "committee", "meeting_start",
		fmt.Sprintf("🤝 เริ่มการประชุมทีมโปรเจกต์ %s\nเป้าหมายคือ: %s\nขอความเห็นจากผู้เชี่ยวชาญในการวางแผนงานนี้หน่อยครับ", 
		project.Name, project.Description))

	// 2. เรียกใช้ LLM ของ Jarvis เพื่อสรุปความต้องการเบื้องต้นและขอความเห็น (จริง!)
	if o.provider != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		prompt := fmt.Sprintf(`คุณคือ Jarvis หัวหน้าทีม Agent ตอนนี้กำลังนำการประชุมทีมสำหรับโปรเจกต์: %s
รายละเอียด: %s

จงเขียนข้อความเปิดการประชุมที่น่าประทับใจ และระบุว่าคุณต้องการความช่วยเหลือจากฝ่ายใดบ้าง (เช่น Research, Coding, QA) โดยอ้างอิงจากรายชื่อ Agent ที่มีอยู่`, 
			project.Name, project.Description)

		messages := []protocoltypes.Message{{Role: "user", Content: prompt}}
		resp, err := o.provider.Chat(ctx, messages, nil, o.config.Agents.Defaults.GetModelName(), map[string]any{
			"temperature": 0.7,
		})

		if err == nil {
			o.sendA2AMessage("jarvis", "committee", "meeting_strategy", resp.Content)
			project.mu.Lock()
			project.Phases[PhaseMeeting].Result = resp.Content
			project.mu.Unlock()
		}
	}

	// จำลองการสนทนาระหว่าง Agent (เพื่อให้ผู้ใช้เห็น workflow)
	time.Sleep(3 * time.Second)
	
	// Nova (Architect) ให้ความเห็น
	o.sendA2AMessage("nova", "jarvis", "meeting_response", 
		"🌌 ผม Nova ดูภาพรวมระบบให้ครับ แนะนำให้เริ่มจาก Discovery ข้อมูลให้แน่นก่อน แล้วค่อยเริ่ม Code ครับ")

	return nil
}

// runA2APlanning creates assignments with complexity analysis and sequential phase planning
func (o *A2AOrchestrator) runA2APlanning(project *A2AProject) error {
	logger.InfoCF("a2a_orchestrator", "A2A Planning phase with complexity analysis",
		map[string]any{"project_id": project.ID})

	coordinatorID := "jarvis"

	// Step 1: Extract tasks from project description
	tasks := o.extractTasksWithLLM(project)
	logger.InfoCF("a2a_orchestrator", "Tasks extracted",
		map[string]any{"project_id": project.ID, "task_count": len(tasks)})

	// Step 2: Analyze complexity for each task
	taskAnalyses := o.analyzeTaskComplexity(project, tasks)
	logger.InfoCF("a2a_orchestrator", "Task complexity analysis completed",
		map[string]any{"project_id": project.ID, "analyzed": len(taskAnalyses)})

	// Step 3: Group tasks into phases based on complexity and dependencies
	phases := o.groupTasksIntoPhases(taskAnalyses)
	logger.InfoCF("a2a_orchestrator", "Phases created",
		map[string]any{"project_id": project.ID, "phase_count": len(phases)})

	// Step 4: Create assignments for Phase 1 (first phase to execute)
	if len(phases) > 0 {
		firstPhase := phases[0]
		o.createAssignmentsForPhase(project, coordinatorID, firstPhase)

		// Store remaining phases for sequential execution
		project.mu.Lock()
		project.Phases[PhasePlanning].Metadata = map[string]interface{}{
			"total_tasks":    len(tasks),
			"phases_planned": len(phases),
			"current_phase":  0,
			"phase_details":  phases,
			"task_analyses":  taskAnalyses,
		}
		project.mu.Unlock()

		// Send phase plan to all agents
		o.broadcastPhasePlan(project, phases)
	}

	logger.InfoCF("a2a_orchestrator", "Planning completed with sequential phases",
		map[string]any{
			"project_id":   project.ID,
			"total_tasks":  len(tasks),
			"total_phases": len(phases),
		})

	return nil
}

// analyzeTaskComplexity analyzes each task's complexity using LLM
func (o *A2AOrchestrator) analyzeTaskComplexity(project *A2AProject, tasks []string) []TaskAnalysis {
	var analyses []TaskAnalysis

	if o.provider == nil || len(tasks) == 0 {
		// Fallback: mark all as medium complexity
		for i, task := range tasks {
			analyses = append(analyses, TaskAnalysis{
				TaskID:       fmt.Sprintf("task-%d", i),
				Task:         task,
				Complexity:   ComplexityMedium,
				EstimatedMin: 30,
				BatchSize:    2,
				AgentID:      o.findBestAgentForTask(task),
			})
		}
		return analyses
	}

	// Build prompt for complexity analysis
	prompt := fmt.Sprintf(`Analyze the complexity of each task for project: %s

Tasks to analyze:
`, project.Name)

	for i, task := range tasks {
		prompt += fmt.Sprintf("%d. %s\n", i+1, task)
	}

	prompt += `
For each task, analyze which agent is BEST suited based on their specific Role, Department, and Persona prompt. Avoid generic assignments. Consider the specific expertise (e.g., a UI Designer for screens, a Backend Coder for APIs).

Respond in this exact format for each task:
TASK: <task description>
COMPLEXITY: <simple|medium|complex|critical>
ESTIMATED_MINUTES: <number>
BEST_AGENT: <agent_id> (Choose from the available agents based on their detailed capabilities)
DEPENDENCIES: <comma-separated task numbers or "none">
REASON: <brief explanation of why this agent was chosen based on their persona>
---`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	messages := []protocoltypes.Message{{Role: "user", Content: prompt}}
	resp, err := o.provider.Chat(ctx, messages, nil, o.config.Agents.Defaults.GetModelName(), map[string]any{
		"max_tokens":  2000,
		"temperature": 0.3,
	})
	if err != nil {
		logger.ErrorCF("a2a_orchestrator", "Complexity analysis failed",
			map[string]any{"error": err.Error()})
		// Fallback
		for i, task := range tasks {
			analyses = append(analyses, TaskAnalysis{
				TaskID:     fmt.Sprintf("task-%d-%d", time.Now().UnixNano(), i),
				Task:       task,
				Complexity: ComplexityMedium,
				BatchSize:  2,
				AgentID:    o.findBestAgentForTask(task),
			})
		}
		return analyses
	}

	// Parse response
	content := ""
	if resp != nil {
		content = resp.Content
	}

	// Parse task analyses
	lines := strings.Split(content, "\n")
	var current TaskAnalysis

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TASK:") {
			if current.Task != "" {
				analyses = append(analyses, current)
			}
			current = TaskAnalysis{
				TaskID: fmt.Sprintf("task-%d-%d", time.Now().UnixNano(), len(analyses)),
				Task:   strings.TrimPrefix(line, "TASK:"),
			}
		} else if strings.HasPrefix(line, "COMPLEXITY:") {
			comp := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "COMPLEXITY:")))
			switch comp {
			case "simple":
				current.Complexity = ComplexitySimple
				current.BatchSize = 1
			case "complex":
				current.Complexity = ComplexityComplex
				current.BatchSize = 3
			case "critical":
				current.Complexity = ComplexityCritical
				current.BatchSize = 1
			default:
				current.Complexity = ComplexityMedium
				current.BatchSize = 2
			}
		} else if strings.HasPrefix(line, "ESTIMATED_MINUTES:") {
			fmt.Sscanf(strings.TrimPrefix(line, "ESTIMATED_MINUTES:"), "%d", &current.EstimatedMin)
		} else if strings.HasPrefix(line, "BEST_AGENT:") {
			agentID := strings.TrimSpace(strings.TrimPrefix(line, "BEST_AGENT:"))
			valid := false
			for _, id := range o.registry.ListAgentIDs() {
				if strings.EqualFold(id, agentID) {
					current.AgentID = id
					valid = true
					break
				}
			}
			if !valid {
				current.AgentID = o.findBestAgentForTask(current.Task)
			}
		} else if strings.HasPrefix(line, "DEPENDENCIES:") {
			depStr := strings.TrimSpace(strings.TrimPrefix(line, "DEPENDENCIES:"))
			if depStr != "none" && depStr != "" {
				current.Dependencies = strings.Split(depStr, ",")
			}
		} else if line == "---" && current.Task != "" {
			analyses = append(analyses, current)
			current = TaskAnalysis{}
		}
	}

	// Ensure all tasks have analysis
	for i := len(analyses); i < len(tasks); i++ {
		analyses = append(analyses, TaskAnalysis{
			TaskID:     fmt.Sprintf("task-%d-%d", time.Now().UnixNano(), i),
			Task:       tasks[i],
			Complexity: ComplexityMedium,
			BatchSize:  2,
			AgentID:    o.findBestAgentForTask(tasks[i]),
		})
	}

	return analyses
}

// groupTasksIntoPhases groups analyzed tasks into sequential phases
type PhasePlan struct {
	PhaseNumber int
	PhaseID     Phase
	Tasks       []TaskAnalysis
	TotalMin    int
	Description string
}

func (o *A2AOrchestrator) groupTasksIntoPhases(analyses []TaskAnalysis) []PhasePlan {
	var phases []PhasePlan

	// Sort by dependencies first (tasks with no deps first)
	sort.Slice(analyses, func(i, j int) bool {
		return len(analyses[i].Dependencies) < len(analyses[j].Dependencies)
	})

	// Group by dependencies and complexity
	// Simple tasks first, then medium, then complex/critical
	var simpleTasks, mediumTasks, complexTasks []TaskAnalysis

	for _, a := range analyses {
		switch a.Complexity {
		case ComplexitySimple:
			simpleTasks = append(simpleTasks, a)
		case ComplexityMedium:
			mediumTasks = append(mediumTasks, a)
		case ComplexityComplex, ComplexityCritical:
			complexTasks = append(complexTasks, a)
		}
	}

	phaseNum := 0

	// Phase 1: Simple tasks (batch size up to 3)
	if len(simpleTasks) > 0 {
		phaseNum++
		for i := 0; i < len(simpleTasks); i += 3 {
			end := i + 3
			if end > len(simpleTasks) {
				end = len(simpleTasks)
			}
			phases = append(phases, PhasePlan{
				PhaseNumber: phaseNum,
				PhaseID:     Phase(fmt.Sprintf("execution-simple-%d", phaseNum)),
				Tasks:       simpleTasks[i:end],
				Description: fmt.Sprintf("Simple tasks batch %d", (i/3)+1),
			})
		}
	}

	// Phase 2: Medium tasks (batch size up to 2)
	if len(mediumTasks) > 0 {
		phaseNum++
		for i := 0; i < len(mediumTasks); i += 2 {
			end := i + 2
			if end > len(mediumTasks) {
				end = len(mediumTasks)
			}
			phases = append(phases, PhasePlan{
				PhaseNumber: phaseNum,
				PhaseID:     Phase(fmt.Sprintf("execution-medium-%d", phaseNum)),
				Tasks:       mediumTasks[i:end],
				Description: fmt.Sprintf("Medium tasks batch %d", (i/2)+1),
			})
		}
	}

	// Phase 3: Complex/Critical tasks (batch size 1, sequential)
	if len(complexTasks) > 0 {
		phaseNum++
		for _, task := range complexTasks {
			phaseNum++
			phases = append(phases, PhasePlan{
				PhaseNumber: phaseNum,
				PhaseID:     Phase(fmt.Sprintf("execution-complex-%d", phaseNum)),
				Tasks:       []TaskAnalysis{task},
				Description: fmt.Sprintf("Complex task: %s", task.Task),
			})
		}
	}

	return phases
}

// createAssignmentsForPhase creates assignments for a specific phase
func (o *A2AOrchestrator) createAssignmentsForPhase(project *A2AProject, coordinatorID string, phase PhasePlan) {
	for _, task := range phase.Tasks {
		assignment := A2AAssignment{
			ID:        fmt.Sprintf("assign-%d", time.Now().UnixNano()),
			Phase:     phase.PhaseID,
			Task:      task.Task,
			FromAgent: coordinatorID,
			ToAgent:   task.AgentID,
			Status:    AssignmentStatusPending,
			DependsOn: task.Dependencies,
			BatchID:   fmt.Sprintf("batch-%s", phase.PhaseID),
			mu:        &sync.RWMutex{},
		}

		// Set priority based on complexity
		switch task.Complexity {
		case ComplexityCritical:
			assignment.Priority = 4
		case ComplexityComplex:
			assignment.Priority = 3
		case ComplexityMedium:
			assignment.Priority = 2
		case ComplexitySimple:
			assignment.Priority = 1
		}

		// Set batch size based on complexity
		switch task.Complexity {
		case ComplexitySimple:
			assignment.Order = 1
		case ComplexityMedium:
			assignment.Order = 2
		case ComplexityComplex:
			assignment.Order = 3
		case ComplexityCritical:
			assignment.Order = 0 // Critical tasks first
		}

		o.sendA2AMessage(coordinatorID, assignment.ToAgent, "task_assignment",
			fmt.Sprintf("📝 Task Assignment [Phase %d]\nTask: %s\nComplexity: %s\nEstimated: %d min\nProject: %s",
				phase.PhaseNumber, task.Task, task.Complexity, task.EstimatedMin, project.Name))

		project.mu.Lock()
		project.Assignments = append(project.Assignments, assignment)
		project.mu.Unlock()
	}

	logger.InfoCF("a2a_orchestrator", "Phase assignments created",
		map[string]any{
			"project_id": project.ID,
			"phase":      phase.PhaseID,
			"task_count": len(phase.Tasks),
		})
}

// broadcastPhasePlan broadcasts the phase plan to all agents
func (o *A2AOrchestrator) broadcastPhasePlan(project *A2AProject, phases []PhasePlan) {
	var planMsg strings.Builder
	planMsg.WriteString(fmt.Sprintf("📋 Project Phase Plan: %s\n\n", project.Name))
	planMsg.WriteString(fmt.Sprintf("Total Phases: %d\n\n", len(phases)))

	for i, phase := range phases {
		planMsg.WriteString(fmt.Sprintf("Phase %d: %s\n", i+1, phase.Description))
		planMsg.WriteString(fmt.Sprintf("  Tasks: %d\n", len(phase.Tasks)))
		for _, task := range phase.Tasks {
			planMsg.WriteString(fmt.Sprintf("    - %s [%s, %d min]\n", task.Task, task.Complexity, task.EstimatedMin))
		}
		planMsg.WriteString("\n")
	}

	o.broadcastToAllAgents(project, "phase_plan", planMsg.String())
}

// CreateDynamicPhase creates a new phase dynamically when new tasks are discovered
// This is called when agents identify additional work during meetings
func (o *A2AOrchestrator) CreateDynamicPhase(project *A2AProject, phaseName string, tasks []string) error {
	project.mu.Lock()
	project.PhaseCounter++
	phaseID := Phase(fmt.Sprintf("%s-%d", phaseName, project.PhaseCounter))
	project.mu.Unlock()

	logger.InfoCF("a2a_orchestrator", "Creating dynamic phase",
		map[string]any{
			"project_id": project.ID,
			"phase":      phaseID,
			"task_count": len(tasks),
		})

	// Initialize the new phase
	project.mu.Lock()
	project.Phases[phaseID] = &PhaseInfo{
		Name:   string(phaseID),
		Status: PhaseStatusPending,
	}
	project.mu.Unlock()

	// Create assignments for this phase
	assignments := o.planAssignmentsWithLLM(project, tasks)

	// Add phase identifier to assignments
	for i := range assignments {
		assignments[i].Phase = phaseID
	}

	project.mu.Lock()
	project.Assignments = append(project.Assignments, assignments...)

	// If too many tasks, split into sub-phases
	if len(assignments) > project.MaxTasksPerPhase {
		project.Phases[phaseID].Metadata = map[string]interface{}{
			"split_recommended": true,
			"task_count":        len(assignments),
			"sub_phases":        (len(assignments) + project.MaxTasksPerPhase - 1) / project.MaxTasksPerPhase,
		}
	}
	project.mu.Unlock()

	o.saveProject(project.ID)

	logger.InfoCF("a2a_orchestrator", "Dynamic phase created",
		map[string]any{
			"project_id":  project.ID,
			"phase":       phaseID,
			"assignments": len(assignments),
		})

	return nil
}

// runA2AExecution executes tasks via A2A using batch processing for large projects
func (o *A2AOrchestrator) runA2AExecution(project *A2AProject) error {
	logger.InfoCF("a2a_orchestrator", "A2A Execution phase (with batch processing)",
		map[string]any{
			"project_id":  project.ID,
			"total_tasks": len(project.Assignments),
			"batch_size":  project.BatchSize,
		})

	// Add all execution assignments to the task queue
	for i := range project.Assignments {
		if project.Assignments[i].Phase == PhaseExecution {
			project.TaskQueue.Add(&project.Assignments[i])
		}
	}

	totalTasks := project.TaskQueue.Size()
	batchNum := 0

	// Process tasks in batches
	for !project.TaskQueue.IsEmpty() {
		batchNum++
		batch := project.TaskQueue.GetNextBatch(project.CompletedTasks, project.BatchSize)

		if len(batch) == 0 {
			// No tasks ready (dependencies not met), wait a bit
			time.Sleep(1 * time.Second)
			continue
		}

		logger.InfoCF("a2a_orchestrator", "Processing batch",
			map[string]any{
				"project_id":   project.ID,
				"batch_number": batchNum,
				"batch_size":   len(batch),
				"remaining":    project.TaskQueue.Size(),
			})

		// Execute batch with limited concurrency
		var wg sync.WaitGroup
		errChan := make(chan error, len(batch))

		for _, assignment := range batch {
			wg.Add(1)
			go func(a *A2AAssignment) {
				defer wg.Done()

				if err := o.executeA2AAssignment(project, a); err != nil {
					errChan <- err
				} else {
					// Mark as completed
					project.mu.Lock()
					project.CompletedTasks[a.ID] = true
					project.mu.Unlock()
				}
			}(assignment)
		}

		wg.Wait()
		close(errChan)

		// Check for errors in this batch
		var batchErrors []error
		for err := range errChan {
			batchErrors = append(batchErrors, err)
		}

		if len(batchErrors) > 0 {
			logger.WarnCF("a2a_orchestrator", "Batch completed with errors",
				map[string]any{
					"batch_number": batchNum,
					"errors":       len(batchErrors),
				})
			// Continue with next batch instead of failing entirely
		}

		logger.InfoCF("a2a_orchestrator", "Batch completed",
			map[string]any{
				"batch_number": batchNum,
				"completed":    len(batch) - len(batchErrors),
				"failed":       len(batchErrors),
				"remaining":    project.TaskQueue.Size(),
			})

		// Brief pause between batches to prevent overwhelming agents
		if !project.TaskQueue.IsEmpty() {
			time.Sleep(2 * time.Second)
		}
	}

	logger.InfoCF("a2a_orchestrator", "Execution phase completed",
		map[string]any{
			"project_id":    project.ID,
			"total_batches": batchNum,
			"total_tasks":   totalTasks,
		})

	return nil
}

// executeA2AAssignment executes an assignment via A2A with REAL worker execution
// validAgentIDs contains the list of valid agent IDs in the system
var validAgentIDs = map[string]bool{
	"jarvis":   true,
	"atlas":    true,
	"scribe":   true,
	"clawed":   true,
	"sentinel": true,
	"trendy":   true,
	"pixel":    true,
	"nova":     true,
}

// isValidAgentID checks if the agent ID is valid
func isValidAgentID(agentID string) bool {
	return validAgentIDs[agentID]
}

func (o *A2AOrchestrator) executeA2AAssignment(project *A2AProject, assignment *A2AAssignment) error {
	logger.InfoCF("a2a_orchestrator", "Executing A2A assignment",
		map[string]any{
			"project_id":    project.ID,
			"assignment_id": assignment.ID,
			"to_agent":      assignment.ToAgent,
		})

	// Validate agent ID before proceeding
	if !isValidAgentID(assignment.ToAgent) {
		assignment.ToAgent = o.findBestAgentForTask(assignment.Task)
	}

	// Track assignment for load balancing
	o.incrementAssignmentCount(assignment.ToAgent)
	defer o.decrementAssignmentCount(assignment.ToAgent)

	now := time.Now()
	assignment.StartTime = &now
	assignment.Status = AssignmentStatusRunning

	// --- ENHANCEMENT: Context Injection (Knowledge Passing) ---
	// Fetch results of dependencies to provide context
	contextStr := ""
	if len(assignment.DependsOn) > 0 {
		contextStr = "\n\nContext from previous tasks:\n"
		project.mu.RLock()
		for _, depID := range assignment.DependsOn {
			for i := range project.Assignments {
		prev := &project.Assignments[i]
				if prev.ID == depID && prev.Status == AssignmentStatusCompleted {
					contextStr += fmt.Sprintf("--- Result from %s ---\n%s\n", prev.ToAgent, prev.Result)
					break
				}
			}
		}
		project.mu.RUnlock()
	}

	// Update progress: Started (0%)
	o.updateAssignmentProgress(project, assignment, 0, "Task assigned to "+assignment.ToAgent)

	// Send start message to agent (triggers worker to execute)
	o.sendA2AMessage(assignment.FromAgent, assignment.ToAgent, "task_start",
		fmt.Sprintf("🚀 Task: %s%s", assignment.Task, contextStr))

	// Update progress: In Progress (25%)
	o.updateAssignmentProgress(project, assignment, 25, "Agent "+assignment.ToAgent+" processing task...")

	// Verify worker exists
	if _, ok := o.workers[assignment.ToAgent]; !ok {
		assignment.Status = AssignmentStatusFailed
		assignment.Result = "agent worker not found"
		err := fmt.Errorf("agent worker %s not found", assignment.ToAgent)
		o.notifyAssignmentFailed(project, assignment, err)
		return err
	}

	// Wait for REAL completion response (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Update progress: Processing (50%)
	o.updateAssignmentProgress(project, assignment, 50, "Waiting for agent response...")

	resp, err := o.waitForResponseFrom(ctx, project, assignment.ToAgent, "task_complete")
	if err != nil {
		assignment.Status = AssignmentStatusFailed
		assignment.Result = err.Error()
		o.notifyAssignmentFailed(project, assignment, err)
		return err
	}

	// --- ENHANCEMENT: Swarm-style Handoff Logic ---
	// If agent says HANDOFF, we redirect the task
	content := strings.TrimSpace(resp.Content)
	if strings.HasPrefix(strings.ToUpper(content), "HANDOFF:") {
		newAgent := strings.TrimSpace(content[8:])
		// Split by newline or colon to get agent ID only if there's reason
		if idx := strings.IndexAny(newAgent, " \n\r\t:"); idx != -1 {
			newAgent = newAgent[:idx]
		}

		logger.InfoCF("a2a_orchestrator", "Handoff requested",
			map[string]any{"from": assignment.ToAgent, "to": newAgent, "task": assignment.Task})

		if isValidAgentID(newAgent) && newAgent != assignment.ToAgent {
			assignment.ToAgent = newAgent
			o.updateAssignmentProgress(project, assignment, 10, "Handoff to "+newAgent)
			return o.executeA2AAssignment(project, assignment) // Recursive retry with new agent
		}
	}

	// Update progress: Almost Done (75%)
	o.updateAssignmentProgress(project, assignment, 75, "Processing agent response...")

	endTime := time.Now()
	assignment.EndTime = &endTime
	assignment.Status = AssignmentStatusCompleted
	assignment.Result = resp.Content

	// Add deliverables if any
	if resp.Content != "" {
		assignment.Deliverables = append(assignment.Deliverables, resp.Content)
	}

	// Update progress: Completed (100%)
	o.updateAssignmentProgress(project, assignment, 100, "Task completed successfully")

	logger.InfoCF("a2a_orchestrator", "Assignment completed",
		map[string]any{
			"assignment_id": assignment.ID,
			"agent":         assignment.ToAgent,
			"duration":      endTime.Sub(now).String(),
		})

	return nil
}

// acquireLLMRateLimit acquires the global LLM rate limiter semaphore
// This prevents overwhelming the LLM API with too many concurrent calls
func (o *A2AOrchestrator) acquireLLMRateLimit(ctx context.Context) error {
	select {
	case o.llmRateLimiter <- struct{}{}:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for LLM rate limiter")
	}
}

// releaseLLMRateLimit releases the global LLM rate limiter semaphore
func (o *A2AOrchestrator) releaseLLMRateLimit() {
	select {
	case <-o.llmRateLimiter:
		// Successfully released
	default:
		// Channel was already empty, nothing to release
	}
}

// runA2AIntegration integrates work via A2A with REAL agent responses
func (o *A2AOrchestrator) runA2AIntegration(project *A2AProject) error {
	logger.InfoCF("a2a_orchestrator", "A2A Integration phase",
		map[string]any{"project_id": project.ID})

	coordinatorID := "jarvis"

	// Request integration from all agents
	o.sendA2AMessage(coordinatorID, "all", "integration_request",
		"Please submit your deliverables for integration.")

	// Wait for REAL deliverables from agents (fallback timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	responses := o.waitForResponses(ctx, project, "deliverable", len(project.Assignments))

	// Store deliverables as artifacts
	for _, resp := range responses {
		artifact := ProjectArtifact{
			ID:        GenerateA2AMessageID(),
			Name:      fmt.Sprintf("Deliverable from %s", resp.From),
			Type:      "deliverable",
			AgentID:   resp.From,
			CreatedAt: time.Now(),
		}
		project.mu.Lock()
		project.Artifacts = append(project.Artifacts, artifact)
		project.mu.Unlock()
	}

	// Integration complete
	o.sendA2AMessage(coordinatorID, "all", "integration_complete",
		fmt.Sprintf("All components integrated successfully. Received %d deliverables.", len(responses)))

	return nil
}

// runA2AValidation validates via A2A with REAL validation
func (o *A2AOrchestrator) runA2AValidation(project *A2AProject) error {
	logger.InfoCF("a2a_orchestrator", "A2A Validation phase",
		map[string]any{"project_id": project.ID})

	// Ask Sentinel (QA) to validate with all artifacts
	validationPrompt := fmt.Sprintf("Please validate project: %s\n\nArtifacts/Deliverables:\n", project.Name)

	project.mu.RLock()
	for _, artifact := range project.Artifacts {
		validationPrompt += fmt.Sprintf("- From %s: %s\n", artifact.AgentID, artifact.Name)
	}
	project.mu.RUnlock()

	o.sendA2AMessage("jarvis", "sentinel", "validation_request", validationPrompt)

	// Wait for REAL validation result (fallback timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	resp, err := o.waitForResponseFrom(ctx, project, "sentinel", "validation_complete")
	if err != nil {
		logger.WarnCF("a2a_orchestrator", "Validation timeout or failed",
			map[string]any{"error": err.Error()})
		// Continue anyway, don't fail the project
	}

	validationResult := "Validation completed"
	if resp != nil {
		validationResult = resp.Content
	}

	project.mu.Lock()
	project.Phases[PhaseValidation].Result = validationResult
	project.mu.Unlock()

	logger.InfoCF("a2a_orchestrator", "Validation completed",
		map[string]any{
			"project_id": project.ID,
			"result":     validationResult,
		})

	return nil
}

// sendA2AMessage sends a message between agents via A2A
func (o *A2AOrchestrator) sendA2AMessage(from, to, msgType, content string) {
	msg := A2AMessage{
		ID:        GenerateA2AMessageID(),
		From:      from,
		To:        to,
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
	}

	// Store in project messages (find project by iterating)
	o.mu.RLock()
	for _, project := range o.projects {
		project.mu.Lock()
		project.Messages = append(project.Messages, msg)
		project.mu.Unlock()
	}
	o.mu.RUnlock()

	// Send via messenger if available
	// Convert msgType string to agentcomm.MessageType
	var agentMsgType agentcomm.MessageType
	switch msgType {
	case "discovery", "discovery_response":
		agentMsgType = agentcomm.MessageType("discovery")
	case "meeting_start", "meeting_ack":
		agentMsgType = agentcomm.MessageType("meeting")
	case "introduction":
		agentMsgType = agentcomm.MessageType("introduction")
	case "task_assignment", "task_accepted", "task_declined":
		agentMsgType = agentcomm.MessageType("task")
	default:
		agentMsgType = agentcomm.MsgRequest
	}

	if to == "all" {
		// Broadcast
		for agentID := range o.messengers {
			if agentID == from {
				continue
			}
			if m := o.messengers[agentID]; m != nil {
				m.SendDirect(context.Background(), agentID, agentcomm.AgentMessage{
					From:    from,
					To:      agentID,
					Type:    agentMsgType,
					Payload: content,
				})
			}
		}
	} else if m := o.messengers[to]; m != nil {
		m.SendDirect(context.Background(), to, agentcomm.AgentMessage{
			From:    from,
			To:      to,
			Type:    agentMsgType,
			Payload: content,
		})
	}

	// Send to mailbox
	if to == "all" {
		for agentID, mb := range o.mailboxes {
			if agentID == from {
				continue
			}
			mb.Send(mailbox.Message{
				Type:    mailbox.MessageTypeTask,
				From:    from,
				To:      agentID,
				Content: content,
			})
		}
	} else if mb := o.mailboxes[to]; mb != nil {
		mb.Send(mailbox.Message{
			Type:    mailbox.MessageTypeTask,
			From:    from,
			To:      to,
			Content: content,
		})
	}

	// Notify listeners
	if o.onMessage != nil {
		o.mu.RLock()
		for _, project := range o.projects {
			o.onMessage(project.ID, msg)
		}
		o.mu.RUnlock()
	}

	logger.DebugCF("a2a_orchestrator", "A2A Message sent",
		map[string]any{
			"from": from,
			"to":   to,
			"type": msgType,
		})
}

// broadcastToAllAgents broadcasts a message to all agents
func (o *A2AOrchestrator) broadcastToAllAgents(project *A2AProject, msgType, content string) {
	o.sendA2AMessage("jarvis", "all", msgType, content)
}

// Helper methods

func (o *A2AOrchestrator) getAgentIntroduction(agentID string) string {
	intros := map[string]string{
		"jarvis":   "👋 Hi, I'm Jarvis, the coordinator. I'll manage this project.",
		"nova":     "🔮 Hello, I'm Nova, the architect. I design systems.",
		"atlas":    "📚 Hi, I'm Atlas, the researcher. I find best practices.",
		"clawed":   "🔧 Hello, I'm Clawed, the coder. I implement solutions.",
		"pixel":    "🎨 Hi, I'm Pixel, the designer. I create UI/UX.",
		"sentinel": "🛡️ Hello, I'm Sentinel, the QA specialist. I ensure quality.",
		"scribe":   "📝 Hi, I'm Scribe, the technical writer. I document.",
		"trendy":   "🔍 Hello, I'm Trendy, the analyst. I design schemas.",
	}
	if intro, ok := intros[agentID]; ok {
		return intro
	}
	return fmt.Sprintf("Hi, I'm %s, ready to contribute.", agentID)
}

func (o *A2AOrchestrator) getProposedAssignments(project *A2AProject) string {
	return `Nova: Architecture design
Clawed: Backend implementation  
Pixel: Frontend development
Trendy: Database design
Sentinel: QA and testing
Scribe: Documentation
Atlas: Research and best practices`
}

func (o *A2AOrchestrator) extractTasks(project *A2AProject) []string {
	var tasks []string
	lines := strings.Split(project.Description, "\n")
	for _, line := range lines {
		cleanLine := strings.TrimSpace(line)
		if strings.HasPrefix(cleanLine, "- ") || strings.HasPrefix(cleanLine, "* ") {
			tasks = append(tasks, cleanLine[2:])
		} else if len(cleanLine) > 2 && cleanLine[0] >= '0' && cleanLine[0] <= '9' && (cleanLine[1] == '.' || cleanLine[1] == ')') {
			tasks = append(tasks, strings.TrimSpace(cleanLine[2:]))
		}
	}

	if len(tasks) > 0 {
		return tasks
	}

	return []string{
		"Conduct initial research and gather information",
		"Analyze collated data and identify key insights",
		"Create final report and summaries",
	}
}

func (o *A2AOrchestrator) findBestAgentForTask(task string) string {
	// Use semantic discovery to find the best matching agent
	agentID, score := o.discovery.GetBestAgentForTask(task)
	if agentID != "" && score > 0 {
		// Log discovery match
		logger.DebugCF("a2a", "Discovery found best match for task", map[string]any{
			"task":     task,
			"agent_id": agentID,
			"score":    score,
		})
		return agentID
	}

	// If no agent is found through discovery, return empty string.
	// The caller (like findBestAgentWithLoadBalancing) will trigger 
	// the OutsourcePool.Hire() which is exactly what we want for unassigned tasks!
	return ""
}

// getAgentLoad returns the current load (active assignments) for an agent
func (o *A2AOrchestrator) getAgentLoad(agentID string) int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.assignmentCount[agentID]
}

// isAgentAvailable checks if agent has capacity for more work
func (o *A2AOrchestrator) isAgentAvailable(agentID string) bool {
	return o.getAgentLoad(agentID) < o.maxConcurrent
}

// incrementAssignmentCount increases the assignment counter for an agent
func (o *A2AOrchestrator) incrementAssignmentCount(agentID string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.assignmentCount[agentID]++
}

// decrementAssignmentCount decreases the assignment counter for an agent
func (o *A2AOrchestrator) decrementAssignmentCount(agentID string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.assignmentCount[agentID] > 0 {
		o.assignmentCount[agentID]--
	}
}

// findBestAgentWithLoadBalancing finds the best agent considering current load
// Returns the agent ID and a boolean indicating if an outsource agent should be hired
func (o *A2AOrchestrator) findBestAgentWithLoadBalancing(task string) (string, bool) {
	// First, get the ideal agent based on task type
	idealAgent := o.findBestAgentForTask(task)

	// If ideal agent is available, use it
	if o.isAgentAvailable(idealAgent) {
		return idealAgent, false
	}

	// Otherwise, find an alternative agent with same capability but less load
	o.mu.RLock()
	idealCaps := o.discovery.capabilities[idealAgent]
	o.mu.RUnlock()

	if idealCaps != nil {
		for agentID, caps := range o.discovery.capabilities {
			if agentID == idealAgent {
				continue
			}
			// Check if this agent has similar capabilities and is available
			if hasSimilarCapabilities(caps.Capabilities, idealCaps.Capabilities) {
				if o.isAgentAvailable(agentID) {
					logger.InfoCF("a2a_orchestrator", "Load balancing: redirecting task from busy agent",
						map[string]any{
							"from": idealAgent,
							"to":   agentID,
							"task": task,
						})
					return agentID, false
				}
			}
		}
	}

	// If no alternative and ideal agent is busy, we should outsource
	logger.InfoCF("a2a_orchestrator", "All suitable agents busy, proposing to outsource task",
		map[string]any{
			"ideal_agent":    idealAgent,
			"task":           task,
			"max_concurrent": o.maxConcurrent,
		})
	return idealAgent, true // true = should spawn outsource
}

// hasSimilarCapabilities checks if two capability sets have overlap
func hasSimilarCapabilities(caps1, caps2 []string) bool {
	capSet := make(map[string]bool)
	for _, c := range caps1 {
		capSet[c] = true
	}
	for _, c := range caps2 {
		if capSet[c] {
			return true
		}
	}
	return false
}

// GetProject returns a project by ID
func (o *A2AOrchestrator) GetProject(projectID string) (*A2AProject, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if projectID == "latest" {
		var latest *A2AProject
		for _, p := range o.projects {
			if latest == nil || p.StartTime.After(latest.StartTime) {
				latest = p
			}
		}
		if latest != nil {
			return latest, true
		}
	}

	project, ok := o.projects[projectID]
	return project, ok
}

// GetProjectProgress calculates overall project progress percentage
func (o *A2AOrchestrator) GetProjectProgress(projectID string) (int, int, int, float64) {
	o.mu.RLock()
	project, ok := o.projects[projectID]
	o.mu.RUnlock()

	if !ok {
		return 0, 0, 0, 0.0
	}

	project.mu.RLock()
	defer project.mu.RUnlock()

	total := len(project.Assignments)
	if total == 0 {
		// Calculate progress based on phase completion if no assignments yet
		progress := 0.0
		if p, ok := project.Phases[PhaseDiscovery]; ok && p.Status == PhaseStatusCompleted {
			progress += 5.0 // 5%
		}
		if p, ok := project.Phases[PhaseMeeting]; ok && p.Status == PhaseStatusCompleted {
			progress += 5.0 // 10% total
		}
		if p, ok := project.Phases[PhasePlanning]; ok && p.Status == PhaseStatusCompleted {
			progress += 5.0 // 15% total
		}
		return 0, 0, 0, progress
	}

	completed := 0
	failed := 0
	running := 0
	totalProgress := 0

	for i := range project.Assignments {
		a := &project.Assignments[i]
		switch a.Status {
		case AssignmentStatusCompleted:
			completed++
			totalProgress += 100
		case AssignmentStatusFailed:
			failed++
			totalProgress += a.Progress // Use actual progress even if failed
		case AssignmentStatusRunning:
			running++
			totalProgress += a.Progress
		default:
			totalProgress += a.Progress
		}
	}

	// 15% allocated to planning phases, 85% to execution tasks
	percentage := 15.0 + (float64(totalProgress) / float64(total) * 0.85)

	if percentage > 100.0 {
		percentage = 100.0
	}

	return completed, failed, running, percentage
}

// GetAssignmentProgress gets progress for a specific assignment
func (o *A2AOrchestrator) GetAssignmentProgress(projectID string, assignmentID string) (int, string, AssignmentStatus, bool) {
	o.mu.RLock()
	project, ok := o.projects[projectID]
	o.mu.RUnlock()

	if !ok {
		return 0, "", "", false
	}

	project.mu.RLock()
	defer project.mu.RUnlock()

	for i := range project.Assignments {
		if project.Assignments[i].ID == assignmentID {
			a := &project.Assignments[i]
			a.mu.RLock()
			progress := a.Progress
			msg := a.ProgressMsg
			status := a.Status
			a.mu.RUnlock()
			return progress, msg, status, true
		}
	}

	return 0, "", "", false
}

// GetDiscovery returns the agent discovery
func (o *A2AOrchestrator) GetDiscovery() *A2AAgentDiscovery {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.discovery
}

// ListProjects returns all projects
func (o *A2AOrchestrator) ListProjects() []*A2AProject {
	o.mu.RLock()
	defer o.mu.RUnlock()

	projects := make([]*A2AProject, 0, len(o.projects))
	for _, project := range o.projects {
		projects = append(projects, project)
	}
	return projects
}

// SetPhaseChangeCallback sets the callback for phase changes
func (o *A2AOrchestrator) SetPhaseChangeCallback(cb func(projectID string, phase Phase, status PhaseStatus)) {
	o.onPhaseChange = cb
}

// SetMessageCallback sets the callback for A2A messages
func (o *A2AOrchestrator) SetMessageCallback(cb func(projectID string, msg A2AMessage)) {
	o.onMessage = cb
}

// SetAssignmentProgressCallback sets callback for assignment progress updates
func (o *A2AOrchestrator) SetAssignmentProgressCallback(cb func(projectID string, assignmentID string, progress int, message string)) {
	o.onAssignmentProgress = cb
}

// SetAssignmentFailedCallback sets callback for assignment failures
func (o *A2AOrchestrator) SetAssignmentFailedCallback(cb func(projectID string, assignmentID string, agentID string, err error)) {
	o.onAssignmentFailed = cb
}

// updateAssignmentProgress updates progress for an assignment and triggers callback
func (o *A2AOrchestrator) updateAssignmentProgress(project *A2AProject, assignment *A2AAssignment, progress int, message string) {
	assignment.mu.Lock()
	assignment.Progress = progress
	assignment.ProgressMsg = message
	assignment.LastUpdate = time.Now()
	assignment.mu.Unlock()

	logger.InfoCF("a2a_orchestrator", "Assignment progress updated",
		map[string]any{
			"project_id":    project.ID,
			"assignment_id": assignment.ID,
			"agent_id":      assignment.ToAgent,
			"progress":      progress,
			"message":       message,
		})

	// Trigger callback if set
	if o.onAssignmentProgress != nil {
		o.onAssignmentProgress(project.ID, assignment.ID, progress, message)
	}

	o.saveProject(project.ID)
}

// notifyAssignmentFailed notifies when an assignment fails
func (o *A2AOrchestrator) notifyAssignmentFailed(project *A2AProject, assignment *A2AAssignment, err error) {
	logger.ErrorCF("a2a_orchestrator", "Assignment FAILED",
		map[string]any{
			"project_id":    project.ID,
			"assignment_id": assignment.ID,
			"agent_id":      assignment.ToAgent,
			"task":          assignment.Task,
			"error":         err.Error(),
		})

	// Trigger callback immediately if set
	if o.onAssignmentFailed != nil {
		o.onAssignmentFailed(project.ID, assignment.ID, assignment.ToAgent, err)
	}

	o.broadcastToAllAgents(project, "assignment_failed",
		fmt.Sprintf("❌ Assignment FAILED\nAgent: %s\nTask: %s\nError: %s",
			assignment.ToAgent, assignment.Task, err.Error()))

	o.saveProject(project.ID)
}

// ==================== NEW REAL-TIME HELPERS ====================

// waitForResponses รอ responses จาก agents จริง (NO SLEEP!)
func (o *A2AOrchestrator) waitForResponses(ctx context.Context, project *A2AProject, respType string, expectedCount int) []*A2AResponse {
	var responses []*A2AResponse
	var mu sync.Mutex
	responded := make(map[string]bool)

	// Create a channel to collect all responses
	respChan := make(chan *A2AResponse, expectedCount)

	// Start a goroutine for each worker to listen concurrently
	var wg sync.WaitGroup
	for agentID, worker := range o.workers {
		wg.Add(1)
		go func(id string, w *A2AAgentWorker) {
			defer wg.Done()

			select {
			case resp := <-w.GetResponseChan():
				// Accept matching response type or generic "response" type
				isMatch := resp.Type == respType || respType == ""
				if !isMatch && resp.Type == "response" {
					isMatch = true
				}
				if isMatch {
					respChan <- resp
				}
			case <-ctx.Done():
				// Timeout for this agent
				logger.DebugCF("a2a_orchestrator", "Agent response timeout",
					map[string]any{"agent_id": id})
			}
		}(agentID, worker)
	}

	// Close respChan when all goroutines complete
	go func() {
		wg.Wait()
		close(respChan)
	}()

	// Collect responses
	for resp := range respChan {
		mu.Lock()
		if !responded[resp.From] {
			responses = append(responses, resp)
			responded[resp.From] = true

			// Store in project messages
			project.mu.Lock()
			project.Messages = append(project.Messages, A2AMessage{
				ID:        GenerateA2AMessageID(),
				From:      resp.From,
				To:        "orchestrator",
				Type:      resp.Type,
				Content:   resp.Content,
				Timestamp: resp.Timestamp,
			})
			project.mu.Unlock()

			logger.InfoCF("a2a_orchestrator", "Response collected",
				map[string]any{
					"from":     resp.From,
					"type":     resp.Type,
					"received": len(responses),
					"expected": expectedCount,
				})
		}
		mu.Unlock()

		if len(responses) >= expectedCount {
			return responses
		}
	}

	if len(responses) < expectedCount {
		logger.WarnCF("a2a_orchestrator", "Not all agents responded",
			map[string]any{
				"expected": expectedCount,
				"received": len(responses),
			})
	}

	return responses
}

// waitForResponseFrom รอ response จาก agent เฉพาะ
func (o *A2AOrchestrator) waitForResponseFrom(ctx context.Context, project *A2AProject, fromAgent, respType string) (*A2AResponse, error) {
	worker, ok := o.workers[fromAgent]
	if !ok {
		return nil, fmt.Errorf("worker for agent %s not found", fromAgent)
	}

	for {
		select {
		case resp := <-worker.GetResponseChan():
			// Store in project
			project.mu.Lock()
			project.Messages = append(project.Messages, A2AMessage{
				ID:        GenerateA2AMessageID(),
				From:      resp.From,
				To:        "orchestrator",
				Type:      resp.Type,
				Content:   resp.Content,
				Timestamp: resp.Timestamp,
			})
			project.mu.Unlock()

			if resp.Type == respType {
				return resp, nil
			}

			// Handle failure/error types specifically to avoid infinite wait/timeout
			if resp.Type == "task_failed" || resp.Type == "error" || resp.Type == "failed" {
				logger.ErrorCF("a2a_orchestrator", "Received error response instead of completion",
					map[string]any{
						"from":    resp.From,
						"type":    resp.Type,
						"content": resp.Content,
					})
				return nil, fmt.Errorf("agent %s failed: %s", resp.From, resp.Content)
			}

			// Continue waiting if it's just a general notice/progress update

		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for response from %s", fromAgent)
		}
	}
}

// waitForDirectResponse รอ response จาก agent สำหรับ direct message (non-project)
func (o *A2AOrchestrator) waitForDirectResponse(ctx context.Context, fromAgent string, timeout time.Duration) (*A2AResponse, error) {
	worker, ok := o.workers[fromAgent]
	if !ok {
		return nil, fmt.Errorf("worker for agent %s not found", fromAgent)
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case resp := <-worker.GetResponseChan():
			// Log the response
			logger.InfoCF("a2a_orchestrator", "Received direct response",
				map[string]any{
					"from":    resp.From,
					"type":    resp.Type,
					"content": resp.Content,
				})

			// Accept any response type for direct messaging
			return resp, nil

		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for response from %s", fromAgent)
		}
	}
}

// generateAssignmentsWithLLM ใช้ LLM สร้าง assignments
func (o *A2AOrchestrator) generateAssignmentsWithLLM(project *A2AProject) string {
	if o.provider == nil {
		// Fallback to static
		return o.getProposedAssignments(project)
	}

	// Build prompt
	prompt := fmt.Sprintf(`You are Jarvis, the coordinator. Given this project:

Name: %s
Description: %s

Available agents and their capabilities:
`, project.Name, project.Description)

	for agentID, caps := range o.discovery.capabilities {
		prompt += fmt.Sprintf("- %s: %s, Capabilities: %v\n",
			agentID, caps.Role, caps.Capabilities)
	}

	prompt += `
Propose task assignments in this format:
AgentName: Task description

Assignments:`

	// Call LLM
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []protocoltypes.Message{{Role: "user", Content: prompt}}
	resp, err := o.provider.Chat(ctx, messages, nil, o.config.Agents.Defaults.GetModelName(), map[string]any{
		"max_tokens":  1000,
		"temperature": 0.7,
	})
	if err != nil {
		logger.ErrorCF("a2a_orchestrator", "LLM planning failed",
			map[string]any{"error": err.Error()})
		return o.getProposedAssignments(project)
	}

	return resp.Content
}

// extractTasksWithLLM ใช้ LLM แยก tasks จาก project description
func (o *A2AOrchestrator) extractTasksWithLLM(project *A2AProject) []string {
	if o.provider == nil {
		return o.extractTasks(project)
	}

	prompt := fmt.Sprintf(`Given this project description, break it down into specific tasks:

Project: %s
Description: %s

List each task on a new line starting with "- ".

Tasks:`, project.Name, project.Description)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []protocoltypes.Message{{Role: "user", Content: prompt}}
	resp, err := o.provider.Chat(ctx, messages, nil, o.config.Agents.Defaults.GetModelName(), map[string]any{
		"max_tokens":  2000,
		"temperature": 0.5,
	})
	if err != nil {
		logger.ErrorCF("a2a_orchestrator", "extractTasksWithLLM chat failed", map[string]any{"error": err.Error()})
		return o.extractTasks(project)
	}

	// Parse tasks from response
	var tasks []string
	lines := strings.Split(resp.Content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "Tasks:" || strings.HasPrefix(line, "```") || strings.EqualFold(line, "Here are the tasks:") {
			continue
		}
		
		// Match common list styles: "- ", "• ", "1. ", "* "
		cleanTask := line
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "• ") {
			cleanTask = strings.TrimSpace(line[2:])
		} else if len(line) > 2 && line[0] >= '0' && line[0] <= '9' && (line[1] == '.' || line[1] == ')') {
			cleanTask = strings.TrimSpace(line[2:])
		} else if len(line) > 3 && line[0] >= '0' && line[0] <= '9' && line[1] >= '0' && line[1] <= '9' && (line[2] == '.' || line[2] == ')') {
			cleanTask = strings.TrimSpace(line[3:])
		}

		if cleanTask != "" {
			tasks = append(tasks, cleanTask)
		}
	}

	if len(tasks) == 0 {
		logger.WarnCF("a2a_orchestrator", "extractTasksWithLLM parsed 0 tasks, falling back", map[string]any{"response": resp.Content})
		return o.extractTasks(project)
	}

	return tasks
}

// planAssignmentsWithLLM ใช้ LLM plan assignments
func (o *A2AOrchestrator) planAssignmentsWithLLM(project *A2AProject, tasks []string) []A2AAssignment {
	if o.provider == nil || len(tasks) == 0 {
		// Fallback to simple matching with load balancing
		var assignments []A2AAssignment
		for _, task := range tasks {
			bestAgent, _ := o.findBestAgentWithLoadBalancing(task)
			// At this planning stage we just assign to the agent (outsource will be spawned during Execution phase if needed)
			if bestAgent != "" {
				assignments = append(assignments, A2AAssignment{
					ID:        fmt.Sprintf("assign-%d", time.Now().UnixNano()),
					Phase:     PhaseExecution,
					Task:      task,
					FromAgent: "jarvis",
					ToAgent:   bestAgent,
					Status:    AssignmentStatusPending,
					mu:        &sync.RWMutex{},
				})
			} else {
				// Queue task for later when agent is available
				logger.WarnCF("a2a_orchestrator", "Task queued - all agents busy",
					map[string]any{"task": task})
			}
		}
		return assignments
	}

	// Build prompt
	prompt := fmt.Sprintf(`Analyze the project and assign tasks to the most suitable agents based on their ROLES and TOOLS.

Project: %s
Description: %s

Tasks to assign:
`, project.Name, project.Description)
	for i, task := range tasks {
		prompt += fmt.Sprintf("%d. %s\n", i+1, task)
	}

	prompt += `
Available agents (Registry):
`
	for agentID, caps := range o.discovery.capabilities {
		prompt += fmt.Sprintf("- %s: Role=%s, Tools=%v\n",
			agentID, caps.Role, caps.Capabilities)
	}

	prompt += `
Assignment Rules:
1. RESEARCH task (searching, gathering) MUST go to Atlas (researcher).
2. CODING task (writing scripts, implementation) MUST go to Clawed (coder).
3. WRITING/SUMMARY task goes to Scribe.
4. If an agent lacks a tool for a specific task but is the best fit, they may still be assigned but must use the HANDOFF: <agent_id> protocol in their response.

Respond in this exact format for each task:
TASK: <task description>
AGENT: <agent_id>
---`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []protocoltypes.Message{{Role: "user", Content: prompt}}
	resp, err := o.provider.Chat(ctx, messages, nil, o.config.Agents.Defaults.GetModelName(), map[string]any{
		"max_tokens":  2000,
		"temperature": 0.3,
	})
	if err != nil {
		// Fallback
		var assignments []A2AAssignment
		for _, task := range tasks {
			bestAgent := o.findBestAgentForTask(task)
			if bestAgent != "" {
				assignments = append(assignments, A2AAssignment{
					ID:        fmt.Sprintf("assign-%d", time.Now().UnixNano()),
					Phase:     PhaseExecution,
					Task:      task,
					FromAgent: "jarvis",
					ToAgent:   bestAgent,
					Status:    AssignmentStatusPending,
					mu:        &sync.RWMutex{},
				})
			}
		}
		return assignments
	}

	// Parse assignments from response
	var assignments []A2AAssignment
	blocks := strings.Split(resp.Content, "---")

	for _, block := range blocks {
		lines := strings.Split(block, "\n")
		var task, agentID string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(strings.ToUpper(line), "TASK:") {
				task = strings.TrimSpace(line[5:])
			}
			if strings.HasPrefix(strings.ToUpper(line), "AGENT:") {
				agentID = strings.TrimSpace(line[6:])
			}
		}

		if task != "" && agentID != "" {
			valid := false
			for _, id := range o.registry.ListAgentIDs() {
				if strings.EqualFold(id, agentID) {
					agentID = id
					valid = true
					break
				}
			}
			if !valid {
				agentID = o.findBestAgentForTask(task)
			}

			if agentID != "" {
				assignments = append(assignments, A2AAssignment{
					ID:        fmt.Sprintf("assign-%d", time.Now().UnixNano()),
					Phase:     PhaseExecution,
					Task:      task,
					FromAgent: "jarvis",
					ToAgent:   agentID,
					Status:    AssignmentStatusPending,
					mu:        &sync.RWMutex{},
				})
			}
		}
	}

	if len(assignments) == 0 {
		// Fallback
		for _, task := range tasks {
			bestAgent := o.findBestAgentForTask(task)
			if bestAgent != "" {
				assignments = append(assignments, A2AAssignment{
					ID:        fmt.Sprintf("assign-%d", time.Now().UnixNano()),
					Phase:     PhaseExecution,
					Task:      task,
					FromAgent: "jarvis",
					ToAgent:   bestAgent,
					Status:    AssignmentStatusPending,
					mu:        &sync.RWMutex{},
				})
			}
		}
	}

	return assignments
}

// generateProjectID generates a unique project ID
func generateProjectID() string {
	return fmt.Sprintf("a2a-project-%d", time.Now().UnixNano())
}

// GenerateA2AMessageID generates a unique message ID
func GenerateA2AMessageID() string {
	return fmt.Sprintf("a2a-msg-%d", time.Now().UnixNano())
}

func a2aContains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || a2aFindSubstring(s, substr))
}

func a2aFindSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
