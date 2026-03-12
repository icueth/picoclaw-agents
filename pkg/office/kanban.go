package office

import (
	"fmt"
	"sync"
	"time"

	"picoclaw/agent/pkg/memory"
)

// KanbanColumn represents a column in a Kanban board.
type KanbanColumn string

const (
	// ColumnBacklog is for tasks that are not yet started.
	ColumnBacklog KanbanColumn = "backlog"
	// ColumnTodo is for tasks ready to be worked on.
	ColumnTodo KanbanColumn = "todo"
	// ColumnInProgress is for tasks currently being worked on.
	ColumnInProgress KanbanColumn = "in_progress"
	// ColumnReview is for tasks awaiting review.
	ColumnReview KanbanColumn = "review"
	// ColumnTesting is for tasks in testing.
	ColumnTesting KanbanColumn = "testing"
	// ColumnDone is for completed tasks.
	ColumnDone KanbanColumn = "done"
	// ColumnBlocked is for blocked tasks.
	ColumnBlocked KanbanColumn = "blocked"
)

// IsValid checks if the column is valid.
func (kc KanbanColumn) IsValid() bool {
	switch kc {
	case ColumnBacklog, ColumnTodo, ColumnInProgress, ColumnReview,
		ColumnTesting, ColumnDone, ColumnBlocked:
		return true
	}
	return false
}

// DisplayName returns a human-readable name for the column.
func (kc KanbanColumn) DisplayName() string {
	switch kc {
	case ColumnBacklog:
		return "Backlog"
	case ColumnTodo:
		return "To Do"
	case ColumnInProgress:
		return "In Progress"
	case ColumnReview:
		return "Review"
	case ColumnTesting:
		return "Testing"
	case ColumnDone:
		return "Done"
	case ColumnBlocked:
		return "Blocked"
	default:
		return string(kc)
	}
}

// Color returns the color associated with the column.
func (kc KanbanColumn) Color() string {
	switch kc {
	case ColumnBacklog:
		return "#6B7280" // Gray
	case ColumnTodo:
		return "#3B82F6" // Blue
	case ColumnInProgress:
		return "#F59E0B" // Amber
	case ColumnReview:
		return "#8B5CF6" // Purple
	case ColumnTesting:
		return "#EC4899" // Pink
	case ColumnDone:
		return "#10B981" // Green
	case ColumnBlocked:
		return "#EF4444" // Red
	default:
		return "#6B7280" // Gray
	}
}

// KanbanPriority represents the priority of a task.
type KanbanPriority string

const (
	// PriorityLow is low priority.
	PriorityLow KanbanPriority = "low"
	// PriorityMedium is medium priority.
	PriorityMedium KanbanPriority = "medium"
	// PriorityHigh is high priority.
	PriorityHigh KanbanPriority = "high"
	// PriorityCritical is critical priority.
	PriorityCritical KanbanPriority = "critical"
)

// IsValid checks if the priority is valid.
func (kp KanbanPriority) IsValid() bool {
	switch kp {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		return true
	}
	return false
}

// Weight returns the numeric weight of the priority (higher = more important).
func (kp KanbanPriority) Weight() int {
	switch kp {
	case PriorityLow:
		return 1
	case PriorityMedium:
		return 2
	case PriorityHigh:
		return 3
	case PriorityCritical:
		return 4
	default:
		return 0
	}
}

// KanbanTask represents a task on a Kanban board.
type KanbanTask struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Column      KanbanColumn      `json:"column"`
	Priority    KanbanPriority    `json:"priority"`
	Assignee    string            `json:"assignee,omitempty"` // Agent ID
	Reporter    string            `json:"reporter,omitempty"` // Agent ID
	Tags        []string          `json:"tags,omitempty"`
	JobID       string            `json:"job_id,omitempty"`   // Associated job ID
	Department  string            `json:"department,omitempty"` // Department ID
	RoomID      string            `json:"room_id,omitempty"`    // Room ID
	DueDate     *time.Time        `json:"due_date,omitempty"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Position    int               `json:"position"` // Position within column
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// IsInFinalColumn returns true if the task is in a final column.
func (kt *KanbanTask) IsInFinalColumn() bool {
	return kt.Column == ColumnDone
}

// IsBlocked returns true if the task is blocked.
func (kt *KanbanTask) IsBlocked() bool {
	return kt.Column == ColumnBlocked
}

// CanStart returns true if the task can be started.
func (kt *KanbanTask) CanStart() bool {
	return kt.Column == ColumnTodo || kt.Column == ColumnBacklog
}

// Start moves the task to in-progress.
func (kt *KanbanTask) Start() error {
	if !kt.CanStart() {
		return fmt.Errorf("task cannot be started from column %s", kt.Column)
	}
	now := time.Now()
	kt.Column = ColumnInProgress
	kt.StartedAt = &now
	kt.UpdatedAt = now
	return nil
}

// Complete moves the task to done.
func (kt *KanbanTask) Complete() error {
	if kt.Column != ColumnInProgress && kt.Column != ColumnReview && kt.Column != ColumnTesting {
		return fmt.Errorf("task cannot be completed from column %s", kt.Column)
	}
	now := time.Now()
	kt.Column = ColumnDone
	kt.CompletedAt = &now
	kt.UpdatedAt = now
	return nil
}

// Block moves the task to blocked.
func (kt *KanbanTask) Block(reason string) {
	kt.Column = ColumnBlocked
	if kt.Metadata == nil {
		kt.Metadata = make(map[string]string)
	}
	kt.Metadata["block_reason"] = reason
	kt.UpdatedAt = time.Now()
}

// Unblock moves the task from blocked to its previous column.
func (kt *KanbanTask) Unblock() error {
	if kt.Column != ColumnBlocked {
		return fmt.Errorf("task is not blocked")
	}
	// Move to todo by default, or could track previous column
	kt.Column = ColumnTodo
	delete(kt.Metadata, "block_reason")
	kt.UpdatedAt = time.Now()
	return nil
}

// GetDuration returns how long the task has been in progress.
func (kt *KanbanTask) GetDuration() time.Duration {
	if kt.StartedAt == nil {
		return 0
	}
	if kt.CompletedAt != nil {
		return kt.CompletedAt.Sub(*kt.StartedAt)
	}
	return time.Since(*kt.StartedAt)
}

// KanbanBoard represents a Kanban board.
type KanbanBoard struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	DepartmentID string            `json:"department_id,omitempty"`
	RoomID       string            `json:"room_id,omitempty"`
	Columns      []KanbanColumn    `json:"columns"`
	TaskIDs      []string          `json:"task_ids"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// DefaultColumns returns the default columns for a new board.
func DefaultColumns() []KanbanColumn {
	return []KanbanColumn{
		ColumnBacklog,
		ColumnTodo,
		ColumnInProgress,
		ColumnReview,
		ColumnDone,
	}
}

// HasColumn checks if the board has a specific column.
func (kb *KanbanBoard) HasColumn(column KanbanColumn) bool {
	for _, c := range kb.Columns {
		if c == column {
			return true
		}
	}
	return false
}

// AddColumn adds a column to the board.
func (kb *KanbanBoard) AddColumn(column KanbanColumn) error {
	if !column.IsValid() {
		return fmt.Errorf("invalid column: %s", column)
	}
	if kb.HasColumn(column) {
		return fmt.Errorf("column %s already exists", column)
	}
	kb.Columns = append(kb.Columns, column)
	kb.UpdatedAt = time.Now()
	return nil
}

// RemoveColumn removes a column from the board.
func (kb *KanbanBoard) RemoveColumn(column KanbanColumn) error {
	for i, c := range kb.Columns {
		if c == column {
			kb.Columns = append(kb.Columns[:i], kb.Columns[i+1:]...)
			kb.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("column %s not found", column)
}

// KanbanManager manages Kanban boards and tasks.
type KanbanManager struct {
	mu        sync.RWMutex
	boards    map[string]*KanbanBoard
	tasks     map[string]*KanbanTask
	byBoard   map[string][]string // boardID -> taskIDs
	byAssignee map[string][]string // agentID -> taskIDs
	byJob     map[string]string   // jobID -> taskID
	jobMgr    *memory.JobManager
}

// NewKanbanManager creates a new Kanban manager.
func NewKanbanManager(jobMgr *memory.JobManager) *KanbanManager {
	return &KanbanManager{
		boards:     make(map[string]*KanbanBoard),
		tasks:      make(map[string]*KanbanTask),
		byBoard:    make(map[string][]string),
		byAssignee: make(map[string][]string),
		byJob:      make(map[string]string),
		jobMgr:     jobMgr,
	}
}

// CreateBoard creates a new Kanban board.
func (km *KanbanManager) CreateBoard(id, name, description string, columns []KanbanColumn) (*KanbanBoard, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	if id == "" {
		return nil, fmt.Errorf("board ID is required")
	}

	if _, exists := km.boards[id]; exists {
		return nil, fmt.Errorf("board %s already exists", id)
	}

	if len(columns) == 0 {
		columns = DefaultColumns()
	}

	// Validate columns
	for _, col := range columns {
		if !col.IsValid() {
			return nil, fmt.Errorf("invalid column: %s", col)
		}
	}

	now := time.Now()
	board := &KanbanBoard{
		ID:          id,
		Name:        name,
		Description: description,
		Columns:     columns,
		TaskIDs:     make([]string, 0),
		Metadata:    make(map[string]string),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	km.boards[id] = board

	return board, nil
}

// GetBoard retrieves a board by ID.
func (km *KanbanManager) GetBoard(id string) (*KanbanBoard, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	board, exists := km.boards[id]
	if !exists {
		return nil, fmt.Errorf("board %s not found", id)
	}

	return board, nil
}

// ListBoards returns all boards.
func (km *KanbanManager) ListBoards() []*KanbanBoard {
	km.mu.RLock()
	defer km.mu.RUnlock()

	result := make([]*KanbanBoard, 0, len(km.boards))
	for _, board := range km.boards {
		result = append(result, board)
	}

	return result
}

// UpdateBoard updates a board's configuration.
func (km *KanbanManager) UpdateBoard(id string, updates map[string]interface{}) (*KanbanBoard, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	board, exists := km.boards[id]
	if !exists {
		return nil, fmt.Errorf("board %s not found", id)
	}

	if name, ok := updates["name"].(string); ok && name != "" {
		board.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		board.Description = description
	}
	if deptID, ok := updates["department_id"].(string); ok {
		board.DepartmentID = deptID
	}
	if roomID, ok := updates["room_id"].(string); ok {
		board.RoomID = roomID
	}
	if metadata, ok := updates["metadata"].(map[string]string); ok {
		for k, v := range metadata {
			board.Metadata[k] = v
		}
	}

	board.UpdatedAt = time.Now()

	return board, nil
}

// DeleteBoard removes a board.
func (km *KanbanManager) DeleteBoard(id string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	board, exists := km.boards[id]
	if !exists {
		return fmt.Errorf("board %s not found", id)
	}

	// Remove all tasks associated with this board
	for _, taskID := range board.TaskIDs {
		delete(km.tasks, taskID)
		delete(km.byJob, taskID)
	}
	delete(km.byBoard, id)

	delete(km.boards, id)

	return nil
}

// CreateTask creates a new task on a board.
func (km *KanbanManager) CreateTask(boardID string, task KanbanTask) (*KanbanTask, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	board, exists := km.boards[boardID]
	if !exists {
		return nil, fmt.Errorf("board %s not found", boardID)
	}

	if task.ID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	if _, exists := km.tasks[task.ID]; exists {
		return nil, fmt.Errorf("task %s already exists", task.ID)
	}

	if task.Title == "" {
		return nil, fmt.Errorf("task title is required")
	}

	// Set default column if not specified
	if task.Column == "" {
		task.Column = ColumnBacklog
	}

	if !task.Column.IsValid() {
		return nil, fmt.Errorf("invalid column: %s", task.Column)
	}

	if !board.HasColumn(task.Column) {
		return nil, fmt.Errorf("column %s not found in board", task.Column)
	}

	// Set default priority if not specified
	if task.Priority == "" {
		task.Priority = PriorityMedium
	}

	if !task.Priority.IsValid() {
		return nil, fmt.Errorf("invalid priority: %s", task.Priority)
	}

	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	task.Position = len(km.byBoard[boardID])

	km.tasks[task.ID] = &task
	km.byBoard[boardID] = append(km.byBoard[boardID], task.ID)

	if task.Assignee != "" {
		km.byAssignee[task.Assignee] = append(km.byAssignee[task.Assignee], task.ID)
	}

	if task.JobID != "" {
		km.byJob[task.JobID] = task.ID
	}

	board.TaskIDs = append(board.TaskIDs, task.ID)
	board.UpdatedAt = now

	return &task, nil
}

// GetTask retrieves a task by ID.
func (km *KanbanManager) GetTask(id string) (*KanbanTask, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	task, exists := km.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task %s not found", id)
	}

	return task, nil
}

// GetTasksByBoard returns all tasks on a board.
func (km *KanbanManager) GetTasksByBoard(boardID string) ([]*KanbanTask, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	board, exists := km.boards[boardID]
	if !exists {
		return nil, fmt.Errorf("board %s not found", boardID)
	}

	result := make([]*KanbanTask, 0, len(board.TaskIDs))
	for _, taskID := range board.TaskIDs {
		if task, exists := km.tasks[taskID]; exists {
			result = append(result, task)
		}
	}

	return result, nil
}

// GetTasksByColumn returns all tasks in a specific column.
func (km *KanbanManager) GetTasksByColumn(boardID string, column KanbanColumn) ([]*KanbanTask, error) {
	tasks, err := km.GetTasksByBoard(boardID)
	if err != nil {
		return nil, err
	}

	result := make([]*KanbanTask, 0)
	for _, task := range tasks {
		if task.Column == column {
			result = append(result, task)
		}
	}

	return result, nil
}

// GetTasksByAssignee returns all tasks assigned to an agent.
func (km *KanbanManager) GetTasksByAssignee(agentID string) []*KanbanTask {
	km.mu.RLock()
	defer km.mu.RUnlock()

	taskIDs := km.byAssignee[agentID]
	result := make([]*KanbanTask, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		if task, exists := km.tasks[taskID]; exists {
			result = append(result, task)
		}
	}

	return result
}

// GetTaskByJob returns the task associated with a job.
func (km *KanbanManager) GetTaskByJob(jobID string) (*KanbanTask, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	taskID, exists := km.byJob[jobID]
	if !exists {
		return nil, fmt.Errorf("no task found for job %s", jobID)
	}

	task, exists := km.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return task, nil
}

// UpdateTask updates a task.
func (km *KanbanManager) UpdateTask(id string, updates map[string]interface{}) (*KanbanTask, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	task, exists := km.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task %s not found", id)
	}

	oldAssignee := task.Assignee

	if title, ok := updates["title"].(string); ok && title != "" {
		task.Title = title
	}
	if description, ok := updates["description"].(string); ok {
		task.Description = description
	}
	if column, ok := updates["column"].(KanbanColumn); ok && column.IsValid() {
		task.Column = column
	}
	if priority, ok := updates["priority"].(KanbanPriority); ok && priority.IsValid() {
		task.Priority = priority
	}
	if assignee, ok := updates["assignee"].(string); ok {
		task.Assignee = assignee
	}
	if tags, ok := updates["tags"].([]string); ok {
		task.Tags = tags
	}
	if dueDate, ok := updates["due_date"].(*time.Time); ok {
		task.DueDate = dueDate
	}
	if metadata, ok := updates["metadata"].(map[string]string); ok {
		for k, v := range metadata {
			task.Metadata[k] = v
		}
	}

	// Update assignee index
	if oldAssignee != task.Assignee {
		// Remove from old assignee
		if oldAssignee != "" {
			ids := km.byAssignee[oldAssignee]
			for i, taskID := range ids {
				if taskID == id {
					km.byAssignee[oldAssignee] = append(ids[:i], ids[i+1:]...)
					break
				}
			}
		}
		// Add to new assignee
		if task.Assignee != "" {
			km.byAssignee[task.Assignee] = append(km.byAssignee[task.Assignee], id)
		}
	}

	task.UpdatedAt = time.Now()

	return task, nil
}

// MoveTask moves a task to a different column.
func (km *KanbanManager) MoveTask(taskID string, toColumn KanbanColumn, position int) (*KanbanTask, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	task, exists := km.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	if !toColumn.IsValid() {
		return nil, fmt.Errorf("invalid column: %s", toColumn)
	}

	// Find the board
	var board *KanbanBoard
	for _, b := range km.boards {
		for _, tid := range b.TaskIDs {
			if tid == taskID {
				board = b
				break
			}
		}
		if board != nil {
			break
		}
	}

	if board == nil {
		return nil, fmt.Errorf("board not found for task %s", taskID)
	}

	if !board.HasColumn(toColumn) {
		return nil, fmt.Errorf("column %s not found in board", toColumn)
	}

	// Update task
	task.Column = toColumn
	task.Position = position
	task.UpdatedAt = time.Now()

	// Update timestamps based on column
	now := time.Now()
	switch toColumn {
	case ColumnInProgress:
		if task.StartedAt == nil {
			task.StartedAt = &now
		}
	case ColumnDone:
		if task.CompletedAt == nil {
			task.CompletedAt = &now
		}
	}

	return task, nil
}

// DeleteTask removes a task.
func (km *KanbanManager) DeleteTask(id string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	task, exists := km.tasks[id]
	if !exists {
		return fmt.Errorf("task %s not found", id)
	}

	// Remove from board
	for _, board := range km.boards {
		for i, taskID := range board.TaskIDs {
			if taskID == id {
				board.TaskIDs = append(board.TaskIDs[:i], board.TaskIDs[i+1:]...)
				board.UpdatedAt = time.Now()
				break
			}
		}
	}

	// Remove from assignee index
	if task.Assignee != "" {
		ids := km.byAssignee[task.Assignee]
		for i, taskID := range ids {
			if taskID == id {
				km.byAssignee[task.Assignee] = append(ids[:i], ids[i+1:]...)
				break
			}
		}
	}

	// Remove from job index
	if task.JobID != "" {
		delete(km.byJob, task.JobID)
	}

	delete(km.tasks, id)

	return nil
}

// CreateTaskFromJob creates a Kanban task from a job.
func (km *KanbanManager) CreateTaskFromJob(boardID string, job memory.Job) (*KanbanTask, error) {
	task := KanbanTask{
		ID:         fmt.Sprintf("task-%s", job.ID),
		Title:      job.Task,
		Column:     ColumnTodo,
		Priority:   PriorityMedium,
		JobID:      job.ID,
		Department: "", // Could be derived from job context
		Metadata:   make(map[string]string),
	}

	// Derive priority from job metadata if available
	if priority, ok := job.Metadata["priority"].(string); ok {
		switch priority {
		case "low":
			task.Priority = PriorityLow
		case "high":
			task.Priority = PriorityHigh
		case "critical":
			task.Priority = PriorityCritical
		default:
			task.Priority = PriorityMedium
		}
	}

	// Derive tags from job role
	if job.Role != "" {
		task.Tags = []string{job.Role}
	}

	return km.CreateTask(boardID, task)
}

// SyncTaskWithJob synchronizes a task's status with its associated job.
func (km *KanbanManager) SyncTaskWithJob(taskID string) (*KanbanTask, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	task, exists := km.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	if task.JobID == "" {
		return task, nil // No job to sync with
	}

	if km.jobMgr == nil {
		return nil, fmt.Errorf("job manager not available")
	}

	job, err := km.jobMgr.GetJob(task.JobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Sync column based on job status
	newColumn := task.Column
	switch job.Status {
	case memory.JobPending:
		if task.Column == ColumnInProgress {
			newColumn = ColumnTodo
		}
	case memory.JobRunning:
		newColumn = ColumnInProgress
	case memory.JobCompleted:
		newColumn = ColumnDone
		now := time.Now()
		task.CompletedAt = &now
	case memory.JobFailed, memory.JobCancelled:
		newColumn = ColumnBlocked
		task.Metadata["block_reason"] = string(job.Status)
	}

	if newColumn != task.Column {
		task.Column = newColumn
		task.UpdatedAt = time.Now()
	}

	return task, nil
}

// GetBoardStats returns statistics for a board.
func (km *KanbanManager) GetBoardStats(boardID string) (BoardStats, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	board, exists := km.boards[boardID]
	if !exists {
		return BoardStats{}, fmt.Errorf("board %s not found", boardID)
	}

	stats := BoardStats{
		BoardID:      boardID,
		TotalTasks:   len(board.TaskIDs),
		ColumnCounts: make(map[KanbanColumn]int),
	}

	for _, taskID := range board.TaskIDs {
		if task, exists := km.tasks[taskID]; exists {
			stats.ColumnCounts[task.Column]++

			if task.Column == ColumnDone {
				stats.CompletedTasks++
			} else if task.Column == ColumnBlocked {
				stats.BlockedTasks++
			}

			// Calculate total time spent
			if task.StartedAt != nil {
				if task.CompletedAt != nil {
					stats.TotalTimeSpent += task.CompletedAt.Sub(*task.StartedAt)
				} else {
					stats.TotalTimeSpent += time.Since(*task.StartedAt)
				}
			}
		}
	}

	if stats.TotalTasks > 0 {
		stats.CompletionRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks)
	}

	return stats, nil
}

// BoardStats holds statistics for a Kanban board.
type BoardStats struct {
	BoardID        string                 `json:"board_id"`
	TotalTasks     int                    `json:"total_tasks"`
	CompletedTasks int                    `json:"completed_tasks"`
	BlockedTasks   int                    `json:"blocked_tasks"`
	ColumnCounts   map[KanbanColumn]int   `json:"column_counts"`
	CompletionRate float64                `json:"completion_rate"`
	TotalTimeSpent time.Duration          `json:"total_time_spent"`
}

// GetWorkload returns the workload for an agent.
func (km *KanbanManager) GetWorkload(agentID string) AgentWorkload {
	km.mu.RLock()
	defer km.mu.RUnlock()

	tasks := km.byAssignee[agentID]
	workload := AgentWorkload{
		AgentID:      agentID,
		TotalTasks:   len(tasks),
		ColumnCounts: make(map[KanbanColumn]int),
	}

	for _, taskID := range tasks {
		if task, exists := km.tasks[taskID]; exists {
			workload.ColumnCounts[task.Column]++

			if task.Column == ColumnInProgress {
				workload.ActiveTasks++
			}

			// Calculate workload weight based on priority
			if task.Column != ColumnDone {
				workload.Weight += task.Priority.Weight()
			}
		}
	}

	return workload
}

// AgentWorkload represents an agent's workload.
type AgentWorkload struct {
	AgentID      string              `json:"agent_id"`
	TotalTasks   int                 `json:"total_tasks"`
	ActiveTasks  int                 `json:"active_tasks"`
	Weight       int                 `json:"weight"`
	ColumnCounts map[KanbanColumn]int `json:"column_counts"`
}

// IsOverloaded returns true if the agent has too many active tasks.
func (aw *AgentWorkload) IsOverloaded(threshold int) bool {
	return aw.ActiveTasks >= threshold
}
