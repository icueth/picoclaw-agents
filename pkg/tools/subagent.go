package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"picoclaw/agent/pkg/agentcomm"
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/memory"
	"picoclaw/agent/pkg/providers"
)

// SubagentTask represents a task executed by a subagent
type SubagentTask struct {
	ID            string
	Task          string
	Label         string
	AgentID       string
	Role          string // Role-based spawning (e.g., "coder", "researcher")
	Model         string // Override model (empty = use SubagentManager default)
	OriginChannel string
	OriginChatID  string
	Status        string // "pending", "running", "completed", "failed", "canceled"
	Result        string
	Error         string // Error message if failed
	Created       int64
	Started       int64         // When task started executing
	Finished      int64         // When task finished
	MaxDuration   time.Duration // Maximum allowed duration
	// Progress tracking fields
	ProgressPercent     int           // Current progress (0-100)
	ProgressMessage     string        // Current progress description
	LastProgressAt      int64         // When progress was last updated
	ExtensionsUsed      int           // Number of timeout extensions used
	MaxExtensions       int           // Maximum allowed extensions (from role config)
	IsExtendable        bool          // Whether this task can be extended
	WarningSent         bool          // Whether warning has been sent
	EstimatedCompletion int64         // Estimated completion time (Unix ms)
	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// SubagentManager manages subagent tasks
type SubagentManager struct {
	tasks          map[string]*SubagentTask
	mu             sync.RWMutex
	provider       providers.LLMProvider
	defaultModel   string
	bus            *bus.MessageBus
	workspace      string
	tools          *ToolRegistry
	maxIterations  int
	maxTokens      int
	temperature    float64
	hasMaxTokens   bool
	hasTemperature bool
	nextID         int
	// New fields for shared context and messenger
	sharedContext *agentcomm.SharedContext
	messenger     MessengerHandler
	// Auto-retry settings
	maxRetries int // 0 = no retry
	retryDelay time.Duration
	// Context builder for inheriting parent's system prompt with skills
	contextBuilder ContextBuilderInterface
	// Role configuration from config
	roleConfig map[string]config.SubagentRoleConfig
	// Job manager for persistence
	jobManager *memory.JobManager
}

// NewSubagentManager creates a new SubagentManager
func NewSubagentManager(
	provider providers.LLMProvider,
	defaultModel, workspace string,
	bus *bus.MessageBus,
) *SubagentManager {
	return &SubagentManager{
		tasks:         make(map[string]*SubagentTask),
		provider:      provider,
		defaultModel:  defaultModel,
		bus:           bus,
		workspace:     workspace,
		tools:         NewToolRegistry(),
		maxIterations: 10,
		nextID:        1,
		roleConfig:    make(map[string]config.SubagentRoleConfig),
	}
}

// SetRoleConfig sets the role configuration from the main config
func (sm *SubagentManager) SetRoleConfig(roles map[string]config.SubagentRoleConfig) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.roleConfig = roles
}

// SetJobManager sets the job manager for persistence
func (sm *SubagentManager) SetJobManager(jm *memory.JobManager) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.jobManager = jm
}

// SetLLMOptions sets max tokens and temperature for subagent LLM calls.
func (sm *SubagentManager) SetLLMOptions(maxTokens int, temperature float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.maxTokens = maxTokens
	sm.hasMaxTokens = true
	sm.temperature = temperature
	sm.hasTemperature = true
}

// SetAutoRetry sets the auto-retry configuration for failed subagent tasks.
func (sm *SubagentManager) SetAutoRetry(maxRetries int, retryDelay time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.maxRetries = maxRetries
	sm.retryDelay = retryDelay
}

// SetTools sets the tool registry for subagent execution.
// If not set, subagent will have access to the provided tools.
func (sm *SubagentManager) SetTools(tools *ToolRegistry) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.tools = tools
}

// RegisterTool registers a tool for subagent execution.
func (sm *SubagentManager) RegisterTool(tool Tool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.tools.Register(tool)
}

// SetSharedContext sets the shared context for subagent communication.
func (sm *SubagentManager) SetSharedContext(ctx *agentcomm.SharedContext) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sharedContext = ctx
}

// ContextBuilderInterface defines the minimal interface needed from ContextBuilder
// to avoid import cycles.
type ContextBuilderInterface interface {
	BuildSystemPrompt() string
}

// SetContextBuilder sets the context builder for creating system prompts with skills.
// This allows subagent to inherit parent's skills and context.
func (sm *SubagentManager) SetContextBuilder(cb ContextBuilderInterface) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.contextBuilder = cb
}

// MessengerHandler defines the interface for sending messages between agents.
// This is an interface to avoid import cycles between agent and tools packages.
type MessengerHandler interface {
	Publish(ctx context.Context, msg agentcomm.AgentMessage) error
	SendDirect(ctx context.Context, to string, msg agentcomm.AgentMessage) error
	Broadcast(ctx context.Context, msg agentcomm.AgentMessage) error
	RegisterAgent(info *agentcomm.AgentInfo)
	UnregisterAgent(agentID string)
	GetAgent(agentID string) (*agentcomm.AgentInfo, bool)
	ListAgents() []*agentcomm.AgentInfo
	ReadSharedContext(key string) (any, bool)
	WriteSharedContext(key string, value any)
	ReadAllSharedContext() map[string]any
	GetMessageLog() []agentcomm.MessageLogEntry
}

// SetMessenger sets the messenger for inter-agent communication.
func (sm *SubagentManager) SetMessenger(m MessengerHandler) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.messenger = m
}

// GetSharedContext returns the shared context.
func (sm *SubagentManager) GetSharedContext() *agentcomm.SharedContext {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sharedContext
}

// GetMessenger returns the messenger.
func (sm *SubagentManager) GetMessenger() MessengerHandler {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.messenger
}

// Spawn starts a subagent task with the given parameters
func (sm *SubagentManager) Spawn(
	ctx context.Context,
	task, label, agentID, model, originChannel, originChatID string,
	callback AsyncCallback,
) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	taskID := fmt.Sprintf("subagent-%d", sm.nextID)
	sm.nextID++

	subagentTask := &SubagentTask{
		ID:            taskID,
		Task:          task,
		Label:         label,
		AgentID:       agentID,
		Model:         model,
		OriginChannel: originChannel,
		OriginChatID:  originChatID,
		Status:        "running",
		Created:       time.Now().UnixMilli(),
	}
	sm.tasks[taskID] = subagentTask

	// Start task in background with context cancellation support
	go sm.runTask(ctx, subagentTask, callback)

	if label != "" {
		return fmt.Sprintf("Spawned subagent '%s' for task: %s", label, task), nil
	}
	return fmt.Sprintf("Spawned subagent for task: %s", task), nil
}

// SpawnWithRole starts a subagent task with a specific role
func (sm *SubagentManager) SpawnWithRole(
	ctx context.Context,
	role, task string,
	contextData map[string]interface{},
	conceptID string,
	timeout int,
	originChannel, originChatID string,
	callback AsyncCallback,
) (string, error) {
	sm.mu.Lock()
	roleConfig, roleExists := sm.roleConfig[role]
	jm := sm.jobManager
	sm.mu.Unlock()

	if !roleExists {
		return "", fmt.Errorf("unknown role: %s", role)
	}

	// Create job in database if job manager is available
	var jobID string
	if jm != nil {
		ctxData := make(map[string]interface{})
		if contextData != nil {
			for k, v := range contextData {
				ctxData[k] = v
			}
		}
		if conceptID != "" {
			ctxData["concept_id"] = conceptID
		}

		var err error
		jobID, err = jm.CreateJob(role, task, ctxData)
		if err != nil {
			return "", fmt.Errorf("failed to create job: %w", err)
		}
	}

	sm.mu.Lock()
	taskID := fmt.Sprintf("subagent-%d", sm.nextID)
	sm.nextID++

	// Determine model from role config or use default
	model := sm.defaultModel
	if roleConfig.Model != "" {
		model = roleConfig.Model
	}

	subagentTask := &SubagentTask{
		ID:            taskID,
		Task:          task,
		Label:         role,
		Role:          role,
		AgentID:       "", // Role-based subagents don't have a specific agent ID
		Model:         model,
		OriginChannel: originChannel,
		OriginChatID:  originChatID,
		Status:        "running",
		Created:       time.Now().UnixMilli(),
	}

	// Apply timeout from role config if not specified
	if timeout <= 0 && roleConfig.TimeoutSeconds > 0 {
		timeout = roleConfig.TimeoutSeconds
	}
	if timeout > 0 {
		subagentTask.MaxDuration = time.Duration(timeout) * time.Second
	}

	sm.tasks[taskID] = subagentTask
	sm.mu.Unlock()

	// Start task in background
	go sm.runTaskWithRole(ctx, subagentTask, roleConfig, jobID, callback)

	result := fmt.Sprintf("Spawned subagent with role '%s' for task: %s (task_id: %s)", role, task, taskID)
	if jobID != "" {
		result += fmt.Sprintf(" [job_id: %s]", jobID)
	}
	return result, nil
}

// SpawnAndWait runs a task synchronously, blocking until completion or context cancellation.
func (sm *SubagentManager) SpawnAndWait(ctx context.Context, task, label, agentID, model, originChannel, originChatID string) (string, error) {
	var resultContent string
	var resultErr error
	done := make(chan struct{})

	_, err := sm.Spawn(ctx, task, label, agentID, model, originChannel, originChatID, func(_ context.Context, result *ToolResult) {
		if result.IsError {
			resultErr = result.Err
			if resultErr == nil {
				resultErr = fmt.Errorf("%s", result.ForLLM)
			}
		} else {
			resultContent = result.ForLLM
		}
		close(done)
	})
	if err != nil {
		return "", err
	}

	select {
	case <-done:
		if resultErr != nil {
			return "", resultErr
		}
		return resultContent, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// runTask executes the subagent task
func (sm *SubagentManager) runTask(ctx context.Context, task *SubagentTask, callback AsyncCallback) {
	task.Status = "running"
	task.Created = time.Now().UnixMilli()
	task.Started = time.Now().UnixMilli()

	// Get settings
	sm.mu.RLock()
	sharedCtx := sm.sharedContext
	messenger := sm.messenger
	maxRetries := sm.maxRetries
	retryDelay := sm.retryDelay
	cb := sm.contextBuilder
	_ = sm.tools // tools are used in RunToolLoop later
	sm.mu.RUnlock()

	// Build system prompt for subagent
	var systemPrompt string

	// Use parent's context builder if available (inherits skills and full context)
	if cb != nil {
		systemPrompt = cb.BuildSystemPrompt()
		// Add subagent-specific instructions
		systemPrompt += `

## Subagent Mode
You are running as a subagent to complete a specific task. Focus on the task at hand and report back when done.
You have access to all tools and skills from the parent agent. Use them as needed.
After completing the task, provide a clear summary of what was done.`
	} else {
		// Fallback to basic system prompt if no context builder
		systemPrompt = `You are a subagent. Complete the given task independently and report the result.
You have access to tools - use them as needed to complete your task.
After completing the task, provide a clear summary of what was done.
IMPORTANT: If you encounter an error, try again with a different approach. Do not give up easily.`
	}

	// Add shared context information if available
	if sharedCtx != nil {
		contextData := sharedCtx.GetAll()
		if len(contextData) > 0 {
			systemPrompt += "\n\nShared context available:\n"
			for k, v := range contextData {
				systemPrompt += fmt.Sprintf("- %s: %v\n", k, v)
			}
		}
	}

	// Register this subagent in the messenger if available
	if messenger != nil {
		agentInfo := &agentcomm.AgentInfo{
			ID:        task.ID,
			Name:      task.Label,
			Type:      "subagent",
			Status:    agentcomm.AgentStatusRunning,
			CreatedAt: task.Created,
		}
		messenger.RegisterAgent(agentInfo)
		defer messenger.UnregisterAgent(task.ID)
	}

	messages := []providers.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: task.Task,
		},
	}

	// Check if context is already canceled before starting
	select {
	case <-ctx.Done():
		sm.mu.Lock()
		task.Status = "canceled"
		task.Result = "Task canceled before execution"
		sm.mu.Unlock()
		return
	default:
	}

	// Get tool settings
	sm.mu.RLock()
	tools := sm.tools
	maxIter := sm.maxIterations
	maxTokens := sm.maxTokens
	temperature := sm.temperature
	hasMaxTokens := sm.hasMaxTokens
	hasTemperature := sm.hasTemperature
	sm.mu.RUnlock()

	var llmOptions map[string]any
	if hasMaxTokens || hasTemperature {
		llmOptions = map[string]any{}
		if hasMaxTokens {
			llmOptions["max_tokens"] = maxTokens
		}
		if hasTemperature {
			llmOptions["temperature"] = temperature
		}
	}

	// Auto-retry loop
	var loopResult *ToolLoopResult
	var err error
	retryCount := 0

	// Use task-specific model if set, otherwise default
	taskModel := sm.defaultModel
	if task.Model != "" {
		taskModel = task.Model
	}

	for {
		loopResult, err = RunToolLoop(ctx, ToolLoopConfig{
			Provider:      sm.provider,
			Model:         taskModel,
			Tools:         tools,
			MaxIterations: maxIter,
			LLMOptions:    llmOptions,
		}, messages, task.OriginChannel, task.OriginChatID)

		// Check if canceled
		if ctx.Err() != nil {
			break
		}

		// If success or no retries configured, break
		if err == nil || maxRetries <= 0 {
			break
		}

		// If failed and retries remaining, wait and retry
		if retryCount < maxRetries {
			retryCount++
			task.Status = "retrying"
			task.Error = fmt.Sprintf("Attempt %d failed: %v. Retrying...", retryCount, err)

			// Notify about retry
			if messenger != nil {
				announceMsg := agentcomm.NewAgentMessage(
					task.ID,
					"main",
					agentcomm.MsgResponse,
					fmt.Sprintf("Task '%s' failed (attempt %d/%d), retrying...\nError: %v",
						task.Label, retryCount, maxRetries, err),
					task.OriginChatID,
				)
				messenger.SendDirect(context.Background(), "main", announceMsg)
			}

			// Wait before retry
			select {
			case <-time.After(retryDelay):
			case <-ctx.Done():
				break
			}
			continue
		}

		// No more retries
		break
	}

	sm.mu.Lock()
	var result *ToolResult
	defer func() {
		sm.mu.Unlock()
		// Call callback if provided and result is set
		if callback != nil && result != nil {
			callback(ctx, result)
		}
	}()

	if err != nil {
		task.Status = "failed"
		task.Result = fmt.Sprintf("Error: %v", err)
		task.Error = err.Error()
		// Check if it was canceled
		if ctx.Err() != nil {
			task.Status = "canceled"
			task.Result = "Task canceled during execution"
			task.Error = "Cancelled by user"
		}
		if retryCount > 0 {
			task.Error = fmt.Sprintf("Failed after %d retries: %v", retryCount, err)
		}
		task.Finished = time.Now().UnixMilli()
		result = &ToolResult{
			ForLLM:  task.Result,
			ForUser: "",
			Silent:  false,
			IsError: true,
			Async:   false,
			Err:     err,
		}
	} else {
		task.Status = "completed"
		task.Result = loopResult.Content
		task.Error = ""
		task.Finished = time.Now().UnixMilli()

		// Update shared context with task result if available
		if sharedCtx != nil {
			sharedCtx.Set(fmt.Sprintf("task:%s:result", task.ID), loopResult.Content)
			sharedCtx.AddMessageLog(task.ID, "", "completed",
				fmt.Sprintf("Task '%s' completed with %d iterations", task.Label, loopResult.Iterations))
		}

		result = &ToolResult{
			ForLLM: fmt.Sprintf(
				"Subagent '%s' completed (iterations: %d): %s",
				task.Label,
				loopResult.Iterations,
				loopResult.Content,
			),
			ForUser: loopResult.Content,
			Silent:  false,
			IsError: false,
			Async:   false,
		}
	}

	// Send announce message back to main agent via messenger or bus.
	// Skip if callback is provided (sync callers handle the result directly).
	if callback == nil {
		if messenger != nil {
			// Use messenger for inter-agent communication
			announceMsg := agentcomm.NewAgentMessage(
				task.ID,
				"main",
				agentcomm.MsgResponse,
				fmt.Sprintf("Task '%s' completed.\n\nResult:\n%s", task.Label, task.Result),
				task.OriginChatID,
			)
			messenger.SendDirect(context.Background(), "main", announceMsg)
		} else if sm.bus != nil {
			// Fallback to bus
			announceContent := fmt.Sprintf("Task '%s' completed.\n\nResult:\n%s", task.Label, task.Result)
			pubCtx, pubCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer pubCancel()
			sm.bus.PublishInbound(pubCtx, bus.InboundMessage{
				Channel:  "system",
				SenderID: fmt.Sprintf("subagent:%s", task.ID),
				// Format: "original_channel:original_chat_id" for routing back
				ChatID:  fmt.Sprintf("%s:%s", task.OriginChannel, task.OriginChatID),
				Content: announceContent,
			})
		}
	}
}

// runTaskWithRole executes a subagent task with role-specific configuration
func (sm *SubagentManager) runTaskWithRole(ctx context.Context, task *SubagentTask, roleConfig config.SubagentRoleConfig, jobID string, callback AsyncCallback) {
	logger.InfoCF("subagent", "Subagent task starting",
		map[string]any{
			"task_id": task.ID,
			"role":    task.Role,
			"task":    task.Task,
		})
	
	task.Status = "running"
	task.Started = time.Now().UnixMilli()
	task.LastProgressAt = task.Started

	// Initialize extendable settings from role config
	task.IsExtendable = roleConfig.Extendable
	task.MaxExtensions = roleConfig.MaxExtensions
	if task.MaxExtensions == 0 && task.IsExtendable {
		task.MaxExtensions = 3 // Default max extensions
	}

	// Store context for cancellation and extension
	task.ctx = ctx

	// Apply timeout if set - use a cancellable context that can be extended
	var cancel context.CancelFunc
	if task.MaxDuration > 0 {
		ctx, cancel = context.WithTimeout(ctx, task.MaxDuration)
		task.cancel = cancel
		defer func() {
			if cancel != nil {
				cancel()
			}
		}()
	}
	
	defer func() {
		logger.InfoCF("subagent", "Subagent task completed",
			map[string]any{
				"task_id": task.ID,
				"status":  task.Status,
				"result":  task.Result,
			})
	}()

	// Get settings
	sm.mu.RLock()
	sharedCtx := sm.sharedContext
	messenger := sm.messenger
	maxRetries := sm.maxRetries
	retryDelay := sm.retryDelay
	cb := sm.contextBuilder
	tools := sm.tools
	maxIter := sm.maxIterations
	maxTokens := sm.maxTokens
	temperature := sm.temperature
	hasMaxTokens := sm.hasMaxTokens
	hasTemperature := sm.hasTemperature
	jm := sm.jobManager
	sm.mu.RUnlock()

	// Register report_progress tool for this specific task
	if tools != nil {
		progressTool := NewReportProgressTool(sm)
		tools.Register(progressTool)
	}

	// Update job status if job manager is available
	if jm != nil && jobID != "" {
		if err := jm.UpdateJobStatus(jobID, string(memory.JobRunning)); err != nil {
			// Log but continue
		}
	}

	// Progress update channel for adaptive timeout
	progressCh := make(chan struct{}, 1)
	defer close(progressCh)

	// Build system prompt for subagent with role-specific additions
	var systemPrompt string

	// Use parent's context builder if available (inherits skills and full context)
	if cb != nil {
		systemPrompt = cb.BuildSystemPrompt()
	} else {
		// Fallback to basic system prompt
		systemPrompt = `You are a subagent. Complete the given task independently and report the result.
You have access to tools - use them as needed to complete your task.
After completing the task, provide a clear summary of what was done.`
	}

	// Add role-specific system prompt addon
	if roleConfig.SystemPromptAddon != "" {
		systemPrompt += "\n\n" + roleConfig.SystemPromptAddon
	}

	// Add role identification
	systemPrompt += fmt.Sprintf("\n\n## Role: %s\nYou are running as a '%s' subagent. Focus on the task at hand and report back when done.", task.Role, task.Role)

	// Add shared context information if available
	if sharedCtx != nil {
		contextData := sharedCtx.GetAll()
		if len(contextData) > 0 {
			systemPrompt += "\n\nShared context available:\n"
			for k, v := range contextData {
				systemPrompt += fmt.Sprintf("- %s: %v\n", k, v)
			}
		}
	}

	// Register this subagent in the messenger if available
	if messenger != nil {
		agentInfo := &agentcomm.AgentInfo{
			ID:        task.ID,
			Name:      task.Label,
			Type:      "subagent",
			Status:    agentcomm.AgentStatusRunning,
			CreatedAt: task.Created,
		}
		messenger.RegisterAgent(agentInfo)
		defer messenger.UnregisterAgent(task.ID)
	}

	messages := []providers.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: task.Task,
		},
	}

	// Check if context is already canceled before starting
	select {
	case <-ctx.Done():
		sm.mu.Lock()
		task.Status = "canceled"
		task.Result = "Task canceled before execution"
		sm.mu.Unlock()
		if jm != nil && jobID != "" {
			jm.UpdateJobStatus(jobID, string(memory.JobCancelled))
		}
		return
	default:
	}

	// Build LLM options with role-specific overrides
	var llmOptions map[string]any
	if hasMaxTokens || hasTemperature || roleConfig.MaxTokens > 0 || roleConfig.Temperature != nil {
		llmOptions = map[string]any{}
		if roleConfig.MaxTokens > 0 {
			llmOptions["max_tokens"] = roleConfig.MaxTokens
		} else if hasMaxTokens {
			llmOptions["max_tokens"] = maxTokens
		}
		if roleConfig.Temperature != nil {
			llmOptions["temperature"] = *roleConfig.Temperature
		} else if hasTemperature {
			llmOptions["temperature"] = temperature
		}
	}

	// Apply role-specific max iterations
	if roleConfig.MaxIterations > 0 {
		maxIter = roleConfig.MaxIterations
	}

	// Auto-retry loop with intelligent retry logic
	var loopResult *ToolLoopResult
	var err error
	retryCount := 0
	maxRetryAttempts := maxRetries
	if maxRetryAttempts <= 0 {
		maxRetryAttempts = 2 // Default to 2 retries if not configured
	}

	// Use task-specific model if set, otherwise default
	taskModel := sm.defaultModel
	if task.Model != "" {
		taskModel = task.Model
	}

	// Retryable error detection
	isRetryableError := func(err error) bool {
		if err == nil {
			return false
		}
		errStr := err.Error()
		// List of retryable errors
		retryablePatterns := []string{
			"timeout",
			"deadline exceeded",
			"rate limit",
			"too many requests",
			"connection refused",
			"connection reset",
			"temporary",
			"unavailable",
			"internal error",
		}
		for _, pattern := range retryablePatterns {
			if containsSubstring(errStr, pattern) {
				return true
			}
		}
		return false
	}

	for {
		// Update progress before starting
		task.ProgressMessage = fmt.Sprintf("Running (attempt %d/%d)", retryCount+1, maxRetryAttempts+1)
		task.LastProgressAt = time.Now().UnixMilli()

		loopResult, err = RunToolLoop(ctx, ToolLoopConfig{
			Provider:      sm.provider,
			Model:         taskModel,
			Tools:         tools,
			MaxIterations: maxIter,
			LLMOptions:    llmOptions,
		}, messages, task.OriginChannel, task.OriginChatID)

		// Check if canceled
		if ctx.Err() != nil {
			break
		}

		// If success, break
		if err == nil {
			break
		}

		// Check if error is retryable
		if !isRetryableError(err) {
			logger.WarnCF("subagent", "Non-retryable error, stopping retries", map[string]any{
				"task_id": task.ID,
				"error":   err.Error(),
			})
			break
		}

		// If retries remaining, wait and retry
		if retryCount < maxRetryAttempts {
			retryCount++
			task.Status = "retrying"
			task.Error = fmt.Sprintf("Attempt %d failed: %v. Retrying...", retryCount, err)
			task.ProgressMessage = fmt.Sprintf("Retrying (attempt %d/%d)", retryCount, maxRetryAttempts)
			task.LastProgressAt = time.Now().UnixMilli()

			logger.InfoCF("subagent", "Retrying task", map[string]any{
				"task_id":     task.ID,
				"attempt":     retryCount,
				"max_retries": maxRetryAttempts,
				"error":       err.Error(),
			})

			// Notify about retry
			if messenger != nil {
				announceMsg := agentcomm.NewAgentMessage(
					task.ID,
					"main",
					agentcomm.MsgResponse,
					fmt.Sprintf("Task '%s' failed (attempt %d/%d), retrying...\nError: %v",
						task.Label, retryCount, maxRetryAttempts, err),
					task.OriginChatID,
				)
				messenger.SendDirect(context.Background(), "main", announceMsg)
			}

			// Calculate backoff delay with exponential backoff
			backoffDelay := retryDelay * time.Duration(retryCount)
			if backoffDelay > 30*time.Second {
				backoffDelay = 30 * time.Second // Cap at 30 seconds
			}

			// Wait before retry
			select {
			case <-time.After(backoffDelay):
			case <-ctx.Done():
				break
			}
			continue
		}

		// No more retries
		break
	}

	sm.mu.Lock()
	var result *ToolResult
	defer func() {
		sm.mu.Unlock()
		// Call callback if provided and result is set
		if callback != nil && result != nil {
			callback(ctx, result)
		}
	}()

	if err != nil {
		task.Status = "failed"
		task.Result = fmt.Sprintf("Error: %v", err)
		task.Error = err.Error()
		// Check if it was canceled
		if ctx.Err() != nil {
			task.Status = "canceled"
			task.Result = "Task canceled during execution"
			task.Error = "Cancelled by user"
		}
		if retryCount > 0 {
			task.Error = fmt.Sprintf("Failed after %d retries: %v", retryCount, err)
		}
		task.Finished = time.Now().UnixMilli()

		// Update job status
		if jm != nil && jobID != "" {
			jm.UpdateJobStatus(jobID, string(memory.JobFailed))
			jm.UpdateJobResult(jobID, task.Error)
		}

		result = &ToolResult{
			ForLLM:  task.Result,
			ForUser: "",
			Silent:  false,
			IsError: true,
			Async:   false,
			Err:     err,
		}
	} else {
		task.Status = "completed"
		task.Result = loopResult.Content
		task.Error = ""
		task.Finished = time.Now().UnixMilli()
		task.ProgressPercent = 100
		task.ProgressMessage = "Completed successfully"

		// Update shared context with task result if available
		if sharedCtx != nil {
			sharedCtx.Set(fmt.Sprintf("task:%s:result", task.ID), loopResult.Content)
			sharedCtx.AddMessageLog(task.ID, "", "completed",
				fmt.Sprintf("Task '%s' completed with %d iterations", task.Label, loopResult.Iterations))
		}

		// Update job status
		if jm != nil && jobID != "" {
			jm.UpdateJobStatus(jobID, string(memory.JobCompleted))
			jm.UpdateJobResult(jobID, loopResult.Content)
		}

		result = &ToolResult{
			ForLLM: fmt.Sprintf(
				"Subagent '%s' completed (iterations: %d): %s",
				task.Label,
				loopResult.Iterations,
				loopResult.Content,
			),
			ForUser: loopResult.Content,
			Silent:  false,
			IsError: false,
			Async:   false,
		}
	}

	// Send announce message back to main agent via messenger or bus
	if callback == nil {
		if messenger != nil {
			announceMsg := agentcomm.NewAgentMessage(
				task.ID,
				"main",
				agentcomm.MsgResponse,
				fmt.Sprintf("Task '%s' completed.\n\nResult:\n%s", task.Label, task.Result),
				task.OriginChatID,
			)
			messenger.SendDirect(context.Background(), "main", announceMsg)
		} else if sm.bus != nil {
			announceContent := fmt.Sprintf("Task '%s' completed.\n\nResult:\n%s", task.Label, task.Result)
			pubCtx, pubCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer pubCancel()
			sm.bus.PublishInbound(pubCtx, bus.InboundMessage{
				Channel:  "system",
				SenderID: fmt.Sprintf("subagent:%s", task.ID),
				ChatID:   fmt.Sprintf("%s:%s", task.OriginChannel, task.OriginChatID),
				Content:  announceContent,
			})
		}
	}
}

// GetTask returns a task by ID
func (sm *SubagentManager) GetTask(taskID string) (*SubagentTask, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	task, ok := sm.tasks[taskID]
	return task, ok
}

// ListTasks returns all tasks
func (sm *SubagentManager) ListTasks() []*SubagentTask {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	tasks := make([]*SubagentTask, 0, len(sm.tasks))
	for _, task := range sm.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// ListActiveTasks returns only running/pending tasks
func (sm *SubagentManager) ListActiveTasks() []*SubagentTask {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var tasks []*SubagentTask
	for _, task := range sm.tasks {
		if task.Status == "running" || task.Status == "pending" {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// GetTaskStatus returns detailed status of a task
func (sm *SubagentManager) GetTaskStatus(taskID string) map[string]any {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	task, ok := sm.tasks[taskID]
	if !ok {
		return map[string]any{
			"exists": false,
			"error":  "Task not found",
		}
	}

	// Calculate duration if running
	var duration string
	var elapsedMs int64
	if task.Started > 0 {
		elapsedMs = time.Now().UnixMilli() - task.Started
		duration = fmt.Sprintf("%dms", elapsedMs)
		if task.Finished > 0 {
			elapsedMs = task.Finished - task.Started
			duration = fmt.Sprintf("%dms", elapsedMs)
		}
	}

	// Calculate timeout information
	var timeoutAt int64
	var maxDurationMs int64
	if task.MaxDuration > 0 {
		maxDurationMs = task.MaxDuration.Milliseconds()
		if task.Started > 0 {
			timeoutAt = task.Started + maxDurationMs
		}
	}

	return map[string]any{
		"exists":               true,
		"id":                     task.ID,
		"label":                  task.Label,
		"role":                   task.Role,
		"status":                 task.Status,
		"created":                task.Created,
		"started":                task.Started,
		"finished":               task.Finished,
		"duration":               duration,
		"elapsed_ms":             elapsedMs,
		"timeout_at":             timeoutAt,
		"max_duration_ms":        maxDurationMs,
		"error":                  task.Error,
		"has_error":              task.Error != "",
		"is_active":              task.Status == "running" || task.Status == "pending",
		"is_complete":            task.Status == "completed",
		"is_failed":              task.Status == "failed",
		// Progress tracking fields
		"progress_percent":       task.ProgressPercent,
		"progress_message":       task.ProgressMessage,
		"last_progress":          task.LastProgressAt,
		"extensions_used":        task.ExtensionsUsed,
		"max_extensions":         task.MaxExtensions,
		"is_extendable":          task.IsExtendable,
		"estimated_completion":   task.EstimatedCompletion,
	}
}

// UpdateTaskProgress updates the progress of a task
func (sm *SubagentManager) UpdateTaskProgress(taskID string, percent int, message string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	task, ok := sm.tasks[taskID]
	if !ok {
		return false
	}

	// Only update if task is still running
	if task.Status != "running" {
		return false
	}

	task.ProgressPercent = percent
	if message != "" {
		task.ProgressMessage = message
	}
	task.LastProgressAt = time.Now().UnixMilli()

	return true
}

// ExtendTaskTimeout extends the timeout of a task
func (sm *SubagentManager) ExtendTaskTimeout(taskID string, extension time.Duration) time.Time {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	task, ok := sm.tasks[taskID]
	if !ok {
		return time.Time{}
	}

	// Check if task can be extended
	if !task.IsExtendable {
		return time.Time{}
	}

	// Check if max extensions reached
	if task.MaxExtensions > 0 && task.ExtensionsUsed >= task.MaxExtensions {
		return time.Time{}
	}

	// Extend the timeout
	task.MaxDuration += extension
	task.ExtensionsUsed++

	// Calculate new timeout time
	var newTimeout time.Time
	if task.Started > 0 {
		newTimeout = time.UnixMilli(task.Started + task.MaxDuration.Milliseconds())
	}

	return newTimeout
}

// RequestTaskExtension allows a subagent to request more time
func (sm *SubagentManager) RequestTaskExtension(taskID string, additionalSeconds int, reason string) (bool, string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	task, ok := sm.tasks[taskID]
	if !ok {
		return false, "Task not found"
	}

	if task.Status != "running" {
		return false, "Task is not running"
	}

	if !task.IsExtendable {
		return false, "Task is not extendable"
	}

	if task.MaxExtensions > 0 && task.ExtensionsUsed >= task.MaxExtensions {
		return false, fmt.Sprintf("Maximum extensions reached (%d/%d)", task.ExtensionsUsed, task.MaxExtensions)
	}

	// Extend the timeout
	extension := time.Duration(additionalSeconds) * time.Second
	task.MaxDuration += extension
	task.ExtensionsUsed++
	task.ProgressMessage = fmt.Sprintf("Extended: %s (%d/%d)", reason, task.ExtensionsUsed, task.MaxExtensions)

	return true, fmt.Sprintf("Timeout extended by %d seconds. Total extensions: %d/%d",
		additionalSeconds, task.ExtensionsUsed, task.MaxExtensions)
}

// CancelTask attempts to cancel a running task
func (sm *SubagentManager) CancelTask(taskID string) (bool, string) {
	sm.mu.RLock()
	task, ok := sm.tasks[taskID]
	sm.mu.RUnlock()

	if !ok {
		return false, "Task not found"
	}

	if task.Status != "running" && task.Status != "pending" {
		return false, "Task is not running (status: " + task.Status + ")"
	}

	// Mark as canceled - the actual cancellation happens in runTask via context
	task.Status = "canceled"
	task.Error = "Cancelled by user"

	// Call the cancel function if available
	if task.cancel != nil {
		task.cancel()
	}

	return true, "Task cancellation requested"
}

// containsSubstring checks if a string contains a substring (case-insensitive)
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(s[:len(substr)] == substr) ||
		(s[len(s)-len(substr):] == substr) ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SpawnSubagentRequest represents a request to spawn a subagent with a role
type SpawnSubagentRequest struct {
	Role      string                 `json:"role"`       // e.g., "coder", "researcher"
	Task      string                 `json:"task"`       // task description
	Context   map[string]interface{} `json:"context"`    // optional context
	ConceptID string                 `json:"concept_id"` // optional concept link
	Timeout   int                    `json:"timeout"`    // timeout in seconds
}

// SpawnSubagentTool spawns a subagent with a specific role
type SpawnSubagentTool struct {
	manager       *SubagentManager
	originChannel string
	originChatID  string
	callback      AsyncCallback
}

// NewSpawnSubagentTool creates a new SpawnSubagentTool
func NewSpawnSubagentTool(manager *SubagentManager) *SpawnSubagentTool {
	return &SpawnSubagentTool{
		manager:       manager,
		originChannel: "cli",
		originChatID:  "direct",
	}
}

// SetCallback implements AsyncTool interface
func (t *SpawnSubagentTool) SetCallback(cb AsyncCallback) {
	t.callback = cb
}

// SetContext implements ContextualTool interface
func (t *SpawnSubagentTool) SetContext(channel, chatID string) {
	t.originChannel = channel
	t.originChatID = chatID
}

// Name returns the tool name
func (t *SpawnSubagentTool) Name() string {
	return "spawn_subagent"
}

// Description returns the tool description
func (t *SpawnSubagentTool) Description() string {
	return `Spawn a specialized subagent with a specific role to handle a task asynchronously.

## When to Use This Tool

You MUST spawn a subagent when:
1. **Coding Tasks** - Writing, debugging, reviewing, or refactoring code
2. **Research Tasks** - Web search, data analysis, investigation
3. **Planning Tasks** - Architecture design, system planning, strategy
4. **Complex Tasks** - Multi-step tasks requiring deep expertise
5. **Tasks Beyond Your Capability** - When the task exceeds your model's strengths

## Available Roles (use list_subagent_roles to see all)

Common roles include:
- **coder**: Code generation, debugging, refactoring, code review
  - Use for: Python, Go, JavaScript, Java, C++, SQL, etc.
  - Use for: Writing functions, classes, algorithms, tests
  - Use for: Fixing bugs, optimizing code, adding features

- **researcher**: Research, analysis, investigation
  - Use for: Web search, data gathering, comparative analysis
  - Use for: Technology research, best practices investigation

- **planner**: Architecture and planning
  - Use for: System design, API design, database schema
  - Use for: Project planning, roadmap creation

- **reviewer**: Code review and quality assurance
  - Use for: Security audits, performance review, best practices check

- **architect**: High-level system design
  - Use for: Microservices architecture, infrastructure design

## How to Select the Right Role

Ask yourself:
1. What type of task is this? (coding/research/planning/review)
2. What expertise is required?
3. How long will this take? (>5 minutes usually needs a subagent)

## After Spawning - CRITICAL

The subagent runs ASYNCHRONOUSLY in the background. You MUST:

1. **Check status immediately**: Use subagent_status with action="get" and task_id
2. **Wait if running**: If status is "running", inform user and wait 15-20 seconds
3. **Check again**: Call subagent_status again (do NOT use sleep tool)
4. **Repeat**: Keep checking every 15-20 seconds until complete
5. **MAX WAIT**: Do NOT exceed 10 status checks (approximately 2-3 minutes total wait)
6. **Report result**: Once completed or max wait reached, present the result to the user

Example workflow:
  1. spawn_subagent(role="coder", task="Write a Python function to...")
     Returns: "Spawned subagent... task_id: subagent-1"
  
  2. subagent_status(action="get", task_id="subagent-1")
     If running: "Task is running... (check 1/10)"
  
  3. (Wait 15-20 seconds, inform user)
  
  4. subagent_status(action="get", task_id="subagent-1")
     If still running: "Still processing... (check 2/10)"
  
  5. Repeat until complete OR max 10 checks reached
  
  6. If max checks reached: "Task is still running. Current progress: X%. I'll check again later."

## Task Duration Guidelines

- Simple tasks (< 2 min): May handle yourself if within your capability
- Medium tasks (2-5 min): Consider spawning based on complexity
- Complex tasks (> 5 min): ALWAYS spawn a subagent
- Coding tasks: ALWAYS spawn a "coder" subagent unless trivial

## Important Notes

- Subagents have access to the same tools and skills as you
- Each role has optimized settings (model, temperature, timeout)
- Subagents can report progress and request timeout extensions
- Failed subagents can be retried or cancelled
- When in doubt, DELEGATE rather than attempt yourself`
}

// Parameters returns the tool parameters
func (t *SpawnSubagentTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"role": map[string]any{
				"type":        "string",
				"description": "The role for the subagent (e.g., 'coder', 'researcher', 'reviewer')",
			},
			"task": map[string]any{
				"type":        "string",
				"description": "The task description for the subagent",
			},
			"context": map[string]any{
				"type":        "object",
				"description": "Optional context data to pass to the subagent",
			},
			"concept_id": map[string]any{
				"type":        "string",
				"description": "Optional concept ID to link this task to",
			},
			"timeout": map[string]any{
				"type":        "integer",
				"description": "Optional timeout in seconds (overrides role default)",
			},
		},
		"required": []string{"role", "task"},
	}
}

// Execute runs the tool
func (t *SpawnSubagentTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	role, ok := args["role"].(string)
	if !ok || role == "" {
		return ErrorResult("role is required")
	}

	task, ok := args["task"].(string)
	if !ok || task == "" {
		return ErrorResult("task is required")
	}

	var contextData map[string]interface{}
	if ctxArg, ok := args["context"].(map[string]interface{}); ok {
		contextData = ctxArg
	}

	conceptID, _ := args["concept_id"].(string)

	timeout := 0
	if t, ok := args["timeout"].(float64); ok {
		timeout = int(t)
	}

	if t.manager == nil {
		return ErrorResult("Subagent manager not configured")
	}

	result, err := t.manager.SpawnWithRole(ctx, role, task, contextData, conceptID, timeout, t.originChannel, t.originChatID, t.callback)
	if err != nil {
		return ErrorResult(fmt.Sprintf("failed to spawn subagent: %v", err))
	}

	return AsyncResult(result)
}

// SubagentTool executes a subagent task synchronously and returns the result.
// Unlike SpawnTool which runs tasks asynchronously, SubagentTool waits for completion
// and returns the result directly in the ToolResult.
type SubagentTool struct {
	manager       *SubagentManager
	originChannel string
	originChatID  string
}

// NewSubagentTool creates a new SubagentTool
func NewSubagentTool(manager *SubagentManager) *SubagentTool {
	return &SubagentTool{
		manager:       manager,
		originChannel: "cli",
		originChatID:  "direct",
	}
}

// Name returns the tool name
func (t *SubagentTool) Name() string {
	return "subagent"
}

// Description returns the tool description
func (t *SubagentTool) Description() string {
	return "Execute a subagent task synchronously and return the result. Use this for delegating specific tasks to an independent agent instance. Returns execution summary to user and full details to LLM."
}

// Parameters returns the tool parameters
func (t *SubagentTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"task": map[string]any{
				"type":        "string",
				"description": "The task for subagent to complete",
			},
			"label": map[string]any{
				"type":        "string",
				"description": "Optional short label for the task (for display)",
			},
		},
		"required": []string{"task"},
	}
}

// SetContext sets the context for the tool
func (t *SubagentTool) SetContext(channel, chatID string) {
	t.originChannel = channel
	t.originChatID = chatID
}

// Execute runs the tool
func (t *SubagentTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	task, ok := args["task"].(string)
	if !ok {
		return ErrorResult("task is required").WithError(fmt.Errorf("task parameter is required"))
	}

	label, _ := args["label"].(string)

	if t.manager == nil {
		return ErrorResult("Subagent manager not configured").WithError(fmt.Errorf("manager is nil"))
	}

	// Build messages for subagent
	messages := []providers.Message{
		{
			Role:    "system",
			Content: "You are a subagent. Complete the given task independently and provide a clear, concise result.",
		},
		{
			Role:    "user",
			Content: task,
		},
	}

	// Use RunToolLoop to execute with tools (same as async SpawnTool)
	sm := t.manager
	sm.mu.RLock()
	tools := sm.tools
	maxIter := sm.maxIterations
	maxTokens := sm.maxTokens
	temperature := sm.temperature
	hasMaxTokens := sm.hasMaxTokens
	hasTemperature := sm.hasTemperature
	sm.mu.RUnlock()

	var llmOptions map[string]any
	if hasMaxTokens || hasTemperature {
		llmOptions = map[string]any{}
		if hasMaxTokens {
			llmOptions["max_tokens"] = maxTokens
		}
		if hasTemperature {
			llmOptions["temperature"] = temperature
		}
	}

	loopResult, err := RunToolLoop(ctx, ToolLoopConfig{
		Provider:      sm.provider,
		Model:         sm.defaultModel,
		Tools:         tools,
		MaxIterations: maxIter,
		LLMOptions:    llmOptions,
	}, messages, t.originChannel, t.originChatID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Subagent execution failed: %v", err)).WithError(err)
	}

	// ForUser: Brief summary for user (truncated if too long)
	userContent := loopResult.Content
	maxUserLen := 500
	if len(userContent) > maxUserLen {
		userContent = userContent[:maxUserLen] + "..."
	}

	// ForLLM: Full execution details
	labelStr := label
	if labelStr == "" {
		labelStr = "(unnamed)"
	}
	llmContent := fmt.Sprintf("Subagent task completed:\nLabel: %s\nIterations: %d\nResult: %s",
		labelStr, loopResult.Iterations, loopResult.Content)

	return &ToolResult{
		ForLLM:  llmContent,
		ForUser: userContent,
		Silent:  false,
		IsError: false,
		Async:   false,
	}
}

// JSONString returns a JSON string representation of the request
func (r *SpawnSubagentRequest) JSONString() string {
	data, _ := json.Marshal(r)
	return string(data)
}
