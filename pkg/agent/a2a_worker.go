// A2A Agent Worker
// Agent ที่ทำงานจริงผ่าน LLM สำหรับระบบ A2A (NO SIMULATION!)

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/agentcomm"
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/providers"
	"picoclaw/agent/pkg/providers/protocoltypes"
)

// A2AAgentWorker ทำงานจริงผ่าน LLM ไม่ใช่ simulation
type A2AAgentWorker struct {
	agentID      string
	agent        *AgentInstance
	messenger    *Messenger
	msgBus       *bus.MessageBus
	mailbox      interface{} // *mailbox.Mailbox
	responseChan chan *A2AResponse
	handlers     map[string]A2AMessageHandler
	mu           sync.RWMutex
	running      bool
	ctx          context.Context
	cancel       context.CancelFunc
}

// A2AResponse เก็บ response จาก agent
type A2AResponse struct {
	From      string
	Type      string
	Content   string
	Timestamp time.Time
	Error     error
}

// TokenMetrics tracks token usage for A2A operations
// Used for optimization analysis and cost monitoring
type TokenMetrics struct {
	AgentID            string        `json:"agent_id"`
	TaskType           string        `json:"task_type"`
	SystemPromptTokens int           `json:"system_prompt_tokens"`
	TaskPromptTokens   int           `json:"task_prompt_tokens"`
	ToolLoopTokens     int           `json:"tool_loop_tokens"`
	TotalTokens        int           `json:"total_tokens"`
	IterationCount     int           `json:"iteration_count"`
	ToolCallCount      int           `json:"tool_call_count"`
	Duration           time.Duration `json:"duration"`
	Timestamp          time.Time     `json:"timestamp"`
}

// EstimateTokens provides a rough estimate of tokens in a string
// This is a simple approximation: ~4 chars per token on average
func EstimateTokens(text string) int {
	return len(text) / 4
}

// LogTokenUsage logs token metrics for analysis
func (w *A2AAgentWorker) LogTokenUsage(metrics TokenMetrics) {
	metrics.AgentID = w.agentID
	metrics.Timestamp = time.Now()
	
	logger.InfoCF("a2a_tokens", "Token usage metrics",
		map[string]any{
			"agent_id":        metrics.AgentID,
			"task_type":       metrics.TaskType,
			"system_tokens":   metrics.SystemPromptTokens,
			"task_tokens":     metrics.TaskPromptTokens,
			"tool_tokens":     metrics.ToolLoopTokens,
			"total_tokens":    metrics.TotalTokens,
			"iterations":      metrics.IterationCount,
			"tool_calls":      metrics.ToolCallCount,
			"duration_ms":     metrics.Duration.Milliseconds(),
		})
}

// A2AMessageHandler จัดการข้อความประเภทต่างๆ
type A2AMessageHandler func(ctx context.Context, msg *A2AMessage) (*A2AResponse, error)

// NewA2AAgentWorker สร้าง worker ใหม่
func NewA2AAgentWorker(agentID string, agent *AgentInstance, messenger *Messenger, msgBus *bus.MessageBus) *A2AAgentWorker {
	ctx, cancel := context.WithCancel(context.Background())

	w := &A2AAgentWorker{
		agentID:      agentID,
		agent:        agent,
		messenger:    messenger,
		msgBus:       msgBus,
		responseChan: make(chan *A2AResponse, 100),
		handlers:     make(map[string]A2AMessageHandler),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Register default handlers
	w.registerDefaultHandlers()

	return w
}

// Start เริ่ม worker
func (w *A2AAgentWorker) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return
	}

	w.running = true

	// เริ่ม goroutine ฟัง messages
	go w.messageLoop()

	logger.InfoCF("a2a_worker", "Agent worker started",
		map[string]any{"agent_id": w.agentID})
}

// Stop หยุด worker
func (w *A2AAgentWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	w.running = false
	w.cancel()

	logger.InfoCF("a2a_worker", "Agent worker stopped",
		map[string]any{"agent_id": w.agentID})
}

// messageLoop ฟังและจัดการ messages
func (w *A2AAgentWorker) messageLoop() {
	// Register handler กับ messenger
	if w.messenger != nil {
		w.messenger.RegisterHandler(w.agentID, func(ctx context.Context, msg *agentcomm.AgentMessage) {
			w.handleAgentMessage(ctx, msg)
		})
	}

	<-w.ctx.Done()
}

// handleAgentMessage จัดการ message ที่ได้รับ
func (w *A2AAgentWorker) handleAgentMessage(ctx context.Context, msg *agentcomm.AgentMessage) {
	a2aMsg := &A2AMessage{
		ID:        GenerateA2AMessageID(),
		From:      msg.From,
		To:        w.agentID,
		Type:      string(msg.Type),
		Content:   msg.GetPayloadString(),
		Timestamp: time.Now(),
	}

	logger.InfoCF("a2a_worker", "Received message",
		map[string]any{
			"agent_id": w.agentID,
			"from":     msg.From,
			"type":     msg.Type,
		})

	// หา handler ที่เหมาะสม
	handler := w.getHandler(a2aMsg.Type)
	if handler == nil {
		handler = w.handleGenericMessage
	}

	// ประมวลผลด้วย LLM จริง (async with longer timeout)
	go func() {
		// Create a longer timeout context for LLM operations
		llmCtx, cancel := context.WithTimeout(w.ctx, 30*time.Minute)
		defer cancel()

		w.emitStatus("THINKING", fmt.Sprintf("Handling message: %s", a2aMsg.Type))
		defer w.emitStatus("IDLE", "Finished processing")

		resp, err := handler(llmCtx, a2aMsg)
		if err != nil {
			logger.ErrorCF("a2a_worker", "Handler error",
				map[string]any{
					"agent_id": w.agentID,
					"error":    err.Error(),
				})
			// Send error response so the caller knows
			w.responseChan <- &A2AResponse{
				From:      w.agentID,
				Type:      "error",
				Content:   fmt.Sprintf("Error processing message: %v", err),
				Timestamp: time.Now(),
				Error:     err,
			}
			return
		}

		if resp != nil {
			w.responseChan <- resp
		}
	}()
}

// registerDefaultHandlers ลงทะเบียน handlers พื้นฐาน
func (w *A2AAgentWorker) registerDefaultHandlers() {
	w.handlers["discovery"] = w.handleDiscovery
	w.handlers["meeting_start"] = w.handleMeetingStart
	w.handlers["introduction"] = w.handleIntroduction
	w.handlers["task"] = w.handleTaskStart
	w.handlers["task_assignment"] = w.handleTaskAssignment
	w.handlers["task_start"] = w.handleTaskStart
	w.handlers["task_complete"] = w.handleTaskComplete
	w.handlers["question"] = w.handleQuestion
}

// getHandler ดึง handler ตาม type
func (w *A2AAgentWorker) getHandler(msgType string) A2AMessageHandler {
	w.mu.RLock()
	defer w.mu.RUnlock()
	handler := w.handlers[msgType]
	if handler == nil {
		logger.DebugCF("a2a_worker", "No handler found for message type",
			map[string]any{
				"agent_id":  w.agentID,
				"msg_type":  msgType,
				"available": len(w.handlers),
			})
	} else {
		logger.DebugCF("a2a_worker", "Handler found for message type",
			map[string]any{
				"agent_id": w.agentID,
				"msg_type": msgType,
			})
	}
	return handler
}

// emitStatus ส่งสถานะของ agent ให้ UI อัพเดท (เช่น THINKING, WORKING, IDLE)
func (w *A2AAgentWorker) emitStatus(status string, details string) {
	if w.msgBus == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	_ = w.msgBus.PublishEvent(ctx, bus.AgentEvent{
		AgentID:   w.agentID,
		EventType: status,
		Details:   details,
	})
}

// ==================== HANDLERS ====================

// handleDiscovery ตอบกลับด้วย capabilities จริง
// OPTIMIZED: Removed redundant agent identity - already in system prompt
func (w *A2AAgentWorker) handleDiscovery(ctx context.Context, msg *A2AMessage) (*A2AResponse, error) {
	logger.InfoCF("a2a_worker", "handleDiscovery called",
		map[string]any{
			"agent_id": w.agentID,
			"from":     msg.From,
		})

	// OPTIMIZED: Minimal prompt - agent identity already in system prompt
	// Project info only - saves ~150-250 tokens per discovery
	prompt := fmt.Sprintf(`A project is starting: "%s"

Please introduce yourself briefly for this project.`,
		msg.Content,
	)

	content, err := w.callLLM(ctx, prompt)
	if err != nil {
		logger.ErrorCF("a2a_worker", "handleDiscovery LLM error",
			map[string]any{
				"agent_id": w.agentID,
				"error":    err.Error(),
			})
		return nil, err
	}

	logger.InfoCF("a2a_worker", "handleDiscovery completed",
		map[string]any{
			"agent_id":      w.agentID,
			"content_chars": len(content),
		})

	return &A2AResponse{
		From:      w.agentID,
		Type:      "discovery_response",
		Content:   content,
		Timestamp: time.Now(),
	}, nil
}

// handleMeetingStart ตอบกลับการประชุม
func (w *A2AAgentWorker) handleMeetingStart(ctx context.Context, msg *A2AMessage) (*A2AResponse, error) {
	// ตอบรับการประชุม
	intro := w.getDefaultIntroduction()

	return &A2AResponse{
		From:      w.agentID,
		Type:      "meeting_ack",
		Content:   intro,
		Timestamp: time.Now(),
	}, nil
}

// handleIntroduction ตอบกลับการแนะนำตัว
func (w *A2AAgentWorker) handleIntroduction(ctx context.Context, msg *A2AMessage) (*A2AResponse, error) {
	// ไม่ต้องตอบกลับ introduction
	return nil, nil
}

// handleTaskAssignment จัดการการมอบหมายงาน
// OPTIMIZED: Removed redundant agent identity - already in system prompt
func (w *A2AAgentWorker) handleTaskAssignment(ctx context.Context, msg *A2AMessage) (*A2AResponse, error) {
	// OPTIMIZED: Minimal prompt - agent identity and capabilities already in system prompt
	// Saves ~100-200 tokens per assignment
	prompt := fmt.Sprintf(`Task assignment:

%s

Do you accept this task? Respond with ACCEPT or DECLINE and briefly explain why.`,
		msg.Content,
	)

	content, err := w.callLLM(ctx, prompt)
	if err != nil {
		return nil, err
	}

	respType := "task_accepted"
	if strings.Contains(strings.ToUpper(content), "DECLINE") {
		respType = "task_declined"
	}

	return &A2AResponse{
		From:      w.agentID,
		Type:      respType,
		Content:   content,
		Timestamp: time.Now(),
	}, nil
}

// handleTaskStart เริ่มทำงานจริง
func (w *A2AAgentWorker) handleTaskStart(ctx context.Context, msg *A2AMessage) (*A2AResponse, error) {
	// ดึง task จาก message
	task := w.extractTaskFromMessage(msg.Content)

	logger.InfoCF("a2a_worker", "Starting task execution",
		map[string]any{
			"agent_id": w.agentID,
			"task":     task,
		})

	// Use agent loopจริงทำงาน
	result, err := w.executeTaskWithAgent(ctx, task)
	if err != nil {
		logger.ErrorCF("a2a_worker", "Task failed",
			map[string]any{
				"agent_id": w.agentID,
				"task":     task,
				"error":    err.Error(),
			})
		return &A2AResponse{
			From:      w.agentID,
			Type:      "task_failed",
			Content:   fmt.Sprintf("❌ Task failed: %v", err),
			Timestamp: time.Now(),
			Error:     err,
		}, nil
	}

	return &A2AResponse{
		From:      w.agentID,
		Type:      "task_complete",
		Content:   result,
		Timestamp: time.Now(),
	}, nil
}

// handleTaskComplete จัดการเมื่องานเสร็จ
func (w *A2AAgentWorker) handleTaskComplete(ctx context.Context, msg *A2AMessage) (*A2AResponse, error) {
	// ไม่ต้องทำอะไร
	return nil, nil
}

// handleQuestion ตอบคำถามด้วย LLM (with tools if available)
func (w *A2AAgentWorker) handleQuestion(ctx context.Context, msg *A2AMessage) (*A2AResponse, error) {
	var content string
	var err error
	// Use tool loop if agent has tools registered
	if w.agent.Tools != nil && len(w.agent.Tools.List()) > 0 {
		content, err = w.executeWithTools(ctx, msg.Content)
	} else {
		content, err = w.callLLM(ctx, msg.Content)
	}
	if err != nil {
		return nil, err
	}

	return &A2AResponse{
		From:      w.agentID,
		Type:      "response",
		Content:   content,
		Timestamp: time.Now(),
	}, nil
}

// handleGenericMessage จัดการ message ทั่วไป (with tools if available)
func (w *A2AAgentWorker) handleGenericMessage(ctx context.Context, msg *A2AMessage) (*A2AResponse, error) {
	var content string
	var err error
	// Use tool loop if agent has tools registered
	if w.agent.Tools != nil && len(w.agent.Tools.List()) > 0 {
		content, err = w.executeWithTools(ctx, msg.Content)
	} else {
		content, err = w.callLLM(ctx, msg.Content)
	}
	if err != nil {
		return nil, err
	}

	return &A2AResponse{
		From:      w.agentID,
		Type:      "response",
		Content:   content,
		Timestamp: time.Now(),
	}, nil
}

// ==================== HELPERS ====================

// callLLM เรียก LLM จริง
// callLLM เรียก LLM จริง พร้อม Retry ในกรณี 429
func (w *A2AAgentWorker) callLLM(ctx context.Context, prompt string) (string, error) {
	if w.agent.Provider == nil {
		return "", fmt.Errorf("agent %s has no LLM provider", w.agentID)
	}

	// สร้าง messages
	messages := []protocoltypes.Message{
		{Role: "system", Content: w.getSystemPrompt()},
		{Role: "user", Content: prompt},
	}

	var lastErr error
	backoff := 1 * time.Second

	// ลองสูงสุด 5 ครั้ง
	for i := 0; i < 5; i++ {
		// เรียก LLM ผ่าน Chat
		resp, err := w.agent.Provider.Chat(ctx, messages, nil, w.agent.Model, map[string]any{
			"max_tokens":  w.agent.MaxTokens,
			"temperature": w.agent.Temperature,
		})
		
		if err == nil {
			return resp.Content, nil
		}
		
		lastErr = err
		
		// เช็คว่าเป็น 429 หรือ Rate Limit หรือไม่
		if strings.Contains(strings.ToLower(err.Error()), "429") || 
		   strings.Contains(strings.ToLower(err.Error()), "rate_limit") ||
		   strings.Contains(strings.ToLower(err.Error()), "too many") {
			
			logger.WarnCF("a2a_worker", "Rate limit hit, backing off", map[string]any{
				"agent":   w.agentID,
				"retry":   i + 1,
				"backoff": backoff.String(),
			})
			
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
				backoff *= 2 // Exponential backoff (1s, 2s, 4s, 8s...)
				continue
			}
		}

		// ถ้า Error อื่นๆ ก็ return เลย
		return "", err
	}

	return "", fmt.Errorf("failed after retries, last error: %v", lastErr)
}

// executeTaskWithAgent ใช้ agent ทำงานจริงกับ tools
// OPTIMIZED: Minimal prompt - agent identity already in system prompt
func (w *A2AAgentWorker) executeTaskWithAgent(ctx context.Context, task string) (string, error) {
	// OPTIMIZED: Minimal task prompt - saves ~150-300 tokens per task execution
	prompt := fmt.Sprintf(`Task:
%s

Use your available tools proactively. Provide a clear, actionable result.`,
		task,
	)

	// ถ้ามี tools ให้ใช้ tool loop
	if w.agent.Tools != nil && len(w.agent.Tools.List()) > 0 {
		return w.executeWithTools(ctx, prompt)
	}

	// ไม่มี tools เรียก LLM ตรงๆ
	return w.callLLM(ctx, prompt)
}

// executeWithTools ทำงานพร้อมใช้ tools
// OPTIMIZED: Added token metrics tracking and context compression
func (w *A2AAgentWorker) executeWithTools(ctx context.Context, prompt string) (string, error) {
	startTime := time.Now()
	maxIterations := w.agent.MaxIterations
	if maxIterations == 0 {
		maxIterations = 10
	}

	systemPrompt := w.getSystemPrompt()
	messages := []protocoltypes.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	// Initialize token metrics
	metrics := TokenMetrics{
		TaskType:           "tool_execution",
		SystemPromptTokens: EstimateTokens(systemPrompt),
		TaskPromptTokens:   EstimateTokens(prompt),
	}

	// Initialize context compressor for Phase 2 optimization
	compressor := NewContextCompressor()

	// ดึง tool definitions
	var toolDefs []providers.ToolDefinition
	if w.agent.Tools != nil {
		allDefs := w.agent.Tools.ToProviderDefs()
		for _, def := range allDefs {
			// Prevent infinite recursion: workers shouldn't start new A2A projects while executing a task
			if def.Function.Name != "start_a2a_project" {
				toolDefs = append(toolDefs, def)
			}
		}
	}

	totalToolTokens := 0
	totalToolCalls := 0

	for i := 0; i < maxIterations; i++ {
		metrics.IterationCount = i + 1

		// PHASE 2: Compress context if needed (after iteration 3)
		if i >= 3 && compressor.ShouldCompress(messages) {
			compressed := compressor.CompressToolLoop(messages, i)
			if compressed.IterationsSummarized > 0 {
				logger.InfoCF("a2a_worker", "Context compression applied",
					map[string]any{
						"agent_id":              w.agentID,
						"iteration":             i,
						"compression":           compressor.GetCompressionStats(compressed),
					})
				messages = compressed.Messages
			}
		}
		var resp *providers.LLMResponse
		var err error
		var lastErr error
		backoff := 1 * time.Second

		// เรียก LLM ลองสูงสุด 5 ครั้งกรณีติด Rate Limit
		for j := 0; j < 5; j++ {
			resp, err = w.agent.Provider.Chat(ctx, messages, toolDefs, w.agent.Model, map[string]any{
				"max_tokens":  w.agent.MaxTokens,
				"temperature": w.agent.Temperature,
			})
			
			if err == nil {
				break
			}
			
			lastErr = err
			
			// เช็คว่าเป็น 429 หรือ Rate Limit หรือไม่
			if strings.Contains(strings.ToLower(err.Error()), "429") || 
			   strings.Contains(strings.ToLower(err.Error()), "rate_limit") ||
			   strings.Contains(strings.ToLower(err.Error()), "too many") {
				
				logger.WarnCF("a2a_worker", "Rate limit hit in tools, backing off", map[string]any{
					"agent":   w.agentID,
					"retry":   j + 1,
					"backoff": backoff.String(),
				})
				
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(backoff):
					backoff *= 2
					continue
				}
			}
			
			break // ถ้าเป็น error อื่นๆ ให้หลุด loop ทันที
		}

		if err != nil {
			return "", fmt.Errorf("provider chat error (last error: %v)", lastErr)
		}

		// ถ้าไม่มี tool calls ให้ return ผลลัพธ์
		if len(resp.ToolCalls) == 0 {
			// Log token metrics before return
			metrics.Duration = time.Since(startTime)
			metrics.ToolCallCount = totalToolCalls
			for _, msg := range messages {
				totalToolTokens += EstimateTokens(msg.Content)
			}
			metrics.ToolLoopTokens = totalToolTokens
			metrics.TotalTokens = metrics.SystemPromptTokens + metrics.TaskPromptTokens + metrics.ToolLoopTokens
			w.LogTokenUsage(metrics)
			return resp.Content, nil
		}

		// Normalize tool calls
		normalizedToolCalls := make([]protocoltypes.ToolCall, 0, len(resp.ToolCalls))
		for _, tc := range resp.ToolCalls {
			normalizedToolCalls = append(normalizedToolCalls, providers.NormalizeToolCall(tc))
		}

		// ประมวลผล tool calls เพื่อใส่ใน history
		assistantMsg := protocoltypes.Message{
			Role:             "assistant",
			Content:          resp.Content,
			ReasoningContent: resp.ReasoningContent,
		}
		for _, tc := range normalizedToolCalls {
			argumentsJSON, _ := json.Marshal(tc.Arguments)
			assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, protocoltypes.ToolCall{
				ID:        tc.ID,
				Type:      "function",
				Name:      tc.Name,
				Arguments: tc.Arguments,
				Function: &protocoltypes.FunctionCall{
					Name:      tc.Name,
					Arguments: string(argumentsJSON),
				},
			})
		}
		messages = append(messages, assistantMsg)

		for _, tc := range normalizedToolCalls {
			// แปลง arguments
			args := make(map[string]any)
			if tc.Arguments != nil {
				args = tc.Arguments
			}

			w.emitStatus("WORKING", fmt.Sprintf("Using tool: %s", tc.Name))
			result := w.agent.Tools.Execute(ctx, tc.Name, args)
			w.emitStatus("THINKING", "Analyzing tool results")

			messages = append(messages, protocoltypes.Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    result.ForLLM,
			})
			totalToolCalls++
		}
	}

	// Log token metrics on max iterations
	metrics.Duration = time.Since(startTime)
	metrics.ToolCallCount = totalToolCalls
	for _, msg := range messages {
		totalToolTokens += EstimateTokens(msg.Content)
	}
	metrics.ToolLoopTokens = totalToolTokens
	metrics.TotalTokens = metrics.SystemPromptTokens + metrics.TaskPromptTokens + metrics.ToolLoopTokens
	w.LogTokenUsage(metrics)

	return "", fmt.Errorf("max iterations reached")
}

// extractTaskFromMessage ดึง task จาก message
func (w *A2AAgentWorker) extractTaskFromMessage(content string) string {
	// หา Task: หรือ Task
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "task:") {
			return strings.TrimSpace(line[5:])
		}
		if strings.HasPrefix(strings.ToLower(line), "task") && len(line) > 5 {
			return strings.TrimSpace(line[5:])
		}
	}
	return content
}

// getDefaultIntroduction คำแนะนำตัว default
func (w *A2AAgentWorker) getDefaultIntroduction() string {
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

	if intro, ok := intros[w.agentID]; ok {
		return intro
	}
	return fmt.Sprintf("Hi, I'm %s, ready to contribute.", w.agent.Name)
}

// getSystemPrompt ดึง system prompt จาก agent (Identity + Persona + Soul + Tools)
// Phase 5: Uses lightweight A2A mode for reduced token usage
func (w *A2AAgentWorker) getSystemPrompt() string {
	// Enable A2A lightweight mode for token optimization
	if w.agent.ContextBuilder != nil {
		// Phase 5: Enable A2A mode for lightweight system prompt
		w.agent.ContextBuilder.SetA2AMode(true)
		
		// Use the lightweight A2A system prompt
		prompt := w.agent.ContextBuilder.BuildSystemPromptWithCache()

		// Append comprehensive tool usage instructions to ensure they are used
		toolInstructions := "\n\n## Tools Capabilities & Equal Rights\n"
		if w.agent.Tools != nil && len(w.agent.Tools.List()) > 0 {
			toolInstructions += "You have FULL ACCESS to the entire suite of tools. You are empowered with the same capabilities as the primary coordinator (Jarvis). Use these tools proactively to ensure accuracy and complete your tasks:\n\n"
			
			var webTools, fsTools, shellTools, skillTools []string
			for _, toolName := range w.agent.Tools.List() {
				if toolName == "start_a2a_project" {
					continue
				}
				lower := strings.ToLower(toolName)
				if strings.Contains(lower, "search") || strings.Contains(lower, "fetch") || strings.Contains(lower, "web") {
					webTools = append(webTools, toolName)
				} else if strings.Contains(lower, "file") || strings.Contains(lower, "dir") {
					fsTools = append(fsTools, toolName)
				} else if strings.Contains(lower, "exec") || strings.Contains(lower, "shell") {
					shellTools = append(shellTools, toolName)
				} else {
					skillTools = append(skillTools, toolName)
				}
			}

			if len(webTools) > 0 {
				toolInstructions += fmt.Sprintf("- **Web Access**: You can browse and search the internet using: %s\n", strings.Join(webTools, ", "))
			}
			if len(fsTools) > 0 {
				toolInstructions += fmt.Sprintf("- **Filesystem**: You can read, write, and manage files using: %s\n", strings.Join(fsTools, ", "))
			}
			if len(shellTools) > 0 {
				toolInstructions += fmt.Sprintf("- **Shell/Execution**: You can execute terminal commands using: %s\n", strings.Join(shellTools, ", "))
			}
			if len(skillTools) > 0 {
				toolInstructions += fmt.Sprintf("- **Skills & MCP**: You have access to specialized functions and MCP servers: %s\n", strings.Join(skillTools, ", "))
			}
			
			toolInstructions += "\nDon't hesitate to use these tools if your task requires external information, file modification, or technical execution.\n"
		} else {
			toolInstructions += "Currently, you do not have external tools assigned. Rely on your internal knowledge and logic.\n"
		}

		return prompt + toolInstructions
	}

	// Fallback if ContextBuilder is missing
	if w.agent.Config != nil {
		return fmt.Sprintf("You are %s, %s. Capabilities: %v",
			w.agent.Name,
			w.agent.Config.Role,
			w.agent.Config.Capabilities)
	}
	return fmt.Sprintf("You are %s, a helpful AI assistant.", w.agent.Name)
}

// GetResponseChan ดึง channel สำหรับรอ response
func (w *A2AAgentWorker) GetResponseChan() <-chan *A2AResponse {
	return w.responseChan
}

// SendResponse ส่ง response กลับ
func (w *A2AAgentWorker) SendResponse(resp *A2AResponse) {
	w.responseChan <- resp
}
