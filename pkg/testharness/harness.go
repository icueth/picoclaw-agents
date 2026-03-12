// PicoClaw - LLM Chat Test Harness
// Test Harness หลักสำหรับจำลองการสนทนากับ LLM

package testharness

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/agent"
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/memory"
	"picoclaw/agent/pkg/providers"
)

// Harness เป็นตัวจัดการการทดสอบหลัก
type Harness struct {
	mu              sync.RWMutex
	provider        *MockProvider
	agentLoop       *agent.AgentLoop
	msgBus          *bus.MessageBus
	config          *config.Config
	memoryManager   *agent.MemoryManager
	jobManager      *memory.JobManager

	// สถานะการสนทนา
	sessionKey      string
	channel         string
	chatID          string
	conversation    []ConversationTurn
	toolResults     map[string]ToolResult

	// Callbacks
	onResponse      func(string)
	onToolCall      func(string, map[string]any)
	onError         func(error)

	// Configuration
	maxIterations   int
	timeout         time.Duration
	captureOutput   bool
	outputBuffer    strings.Builder
}

// ConversationTurn เก็บข้อมูลแต่ละ turn ของการสนทนา
type ConversationTurn struct {
	Timestamp   time.Time
	Role        string // "user", "assistant", "tool"
	Content     string
	ToolCalls   []providers.ToolCall
	ToolResults []ToolResult
}

// ToolResult ผลลัพธ์จากการเรียก tool
type ToolResult struct {
	ToolName string
	Args     map[string]any
	Result   string
	Error    error
}

func float64Ptr(f float64) *float64 {
	return &f
}

// New สร้าง Test Harness ใหม่
func New(provider *MockProvider) *Harness {
	msgBus := bus.NewMessageBus()

	// สร้าง config พื้นฐาน
	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Workspace:         "/tmp/test-harness",
				Model:             "mock-model",
				MaxToolIterations: 20,
				MaxTokens:         4096,
				Temperature:       float64Ptr(0.7),
				Provider:          "mock",
			},
		},
	}

	h := &Harness{
		provider:     provider,
		msgBus:       msgBus,
		config:       cfg,
		sessionKey:   "test-session",
		channel:      "test",
		chatID:       "test-chat",
		conversation: make([]ConversationTurn, 0),
		toolResults:  make(map[string]ToolResult),
		maxIterations: 20,
		timeout:      30 * time.Second,
		captureOutput: true,
	}

	return h
}

// WithConfig กำหนด config ที่กำหนดเอง
func (h *Harness) WithConfig(cfg *config.Config) *Harness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config = cfg
	return h
}

// WithSession กำหนด session key
func (h *Harness) WithSession(sessionKey string) *Harness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessionKey = sessionKey
	return h
}

// WithChannel กำหนด channel และ chat ID
func (h *Harness) WithChannel(channel, chatID string) *Harness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.channel = channel
	h.chatID = chatID
	return h
}

// WithTimeout กำหนด timeout
func (h *Harness) WithTimeout(timeout time.Duration) *Harness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.timeout = timeout
	return h
}

// WithMaxIterations กำหนดจำนวน iterations สูงสุด
func (h *Harness) WithMaxIterations(max int) *Harness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.maxIterations = max
	return h
}

// OnResponse กำหนด callback เมื่อได้รับ response
func (h *Harness) OnResponse(fn func(string)) *Harness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onResponse = fn
	return h
}

// OnToolCall กำหนด callback เมื่อมีการเรียก tool
func (h *Harness) OnToolCall(fn func(string, map[string]any)) *Harness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onToolCall = fn
	return h
}

// OnError กำหนด callback เมื่อเกิด error
func (h *Harness) OnError(fn func(error)) *Harness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onError = fn
	return h
}

// Chat ส่งข้อความและรับการตอบสนอง (แบบ synchronous)
func (h *Harness) Chat(message string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	return h.ChatWithContext(ctx, message)
}

// ChatWithContext ส่งข้อความพร้อม context
func (h *Harness) ChatWithContext(ctx context.Context, message string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// บันทึกข้อความจาก user
	h.conversation = append(h.conversation, ConversationTurn{
		Timestamp: time.Now(),
		Role:      "user",
		Content:   message,
	})

	// สร้าง messages สำหรับส่งไปยัง provider
	messages := h.buildMessages()

	// เรียก provider
	tools := h.buildToolDefinitions()
	resp, err := h.provider.Chat(ctx, messages, tools, h.config.Agents.Defaults.Model, map[string]any{
		"max_tokens":  h.config.Agents.Defaults.MaxTokens,
		"temperature": h.config.Agents.Defaults.Temperature,
	})

	if err != nil {
		if h.onError != nil {
			h.onError(err)
		}
		return "", fmt.Errorf("chat failed: %w", err)
	}

	// บันทึก response
	turn := ConversationTurn{
		Timestamp: time.Now(),
		Role:      "assistant",
		Content:   resp.Content,
		ToolCalls: resp.ToolCalls,
	}

	// จัดการ tool calls
	if len(resp.ToolCalls) > 0 {
		toolResults := h.executeToolCalls(ctx, resp.ToolCalls)
		turn.ToolResults = toolResults
	}

	h.conversation = append(h.conversation, turn)

	// เรียก callback
	if h.onResponse != nil {
		h.onResponse(resp.Content)
	}

	// เก็บ output
	if h.captureOutput {
		h.outputBuffer.WriteString(fmt.Sprintf("User: %s\n", message))
		h.outputBuffer.WriteString(fmt.Sprintf("Assistant: %s\n\n", resp.Content))
	}

	return resp.Content, nil
}

// ChatStreaming ส่งข้อความและรับการตอบสนองแบบ streaming
func (h *Harness) ChatStreaming(message string, onChunk func(string)) error {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	return h.ChatStreamingWithContext(ctx, message, onChunk)
}

// ChatStreamingWithContext ส่งข้อความแบบ streaming พร้อม context
func (h *Harness) ChatStreamingWithContext(ctx context.Context, message string, onChunk func(string)) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// บันทึกข้อความจาก user
	h.conversation = append(h.conversation, ConversationTurn{
		Timestamp: time.Now(),
		Role:      "user",
		Content:   message,
	})

	// สร้าง messages
	messages := h.buildMessages()
	tools := h.buildToolDefinitions()

	// เรียก provider พร้อม on_chunk callback
	resp, err := h.provider.Chat(ctx, messages, tools, h.config.Agents.Defaults.Model, map[string]any{
		"max_tokens":  h.config.Agents.Defaults.MaxTokens,
		"temperature": h.config.Agents.Defaults.Temperature,
		"on_chunk":    onChunk,
	})

	if err != nil {
		if h.onError != nil {
			h.onError(err)
		}
		return fmt.Errorf("streaming chat failed: %w", err)
	}

	// บันทึก response
	turn := ConversationTurn{
		Timestamp: time.Now(),
		Role:      "assistant",
		Content:   resp.Content,
		ToolCalls: resp.ToolCalls,
	}

	if len(resp.ToolCalls) > 0 {
		toolResults := h.executeToolCalls(ctx, resp.ToolCalls)
		turn.ToolResults = toolResults
	}

	h.conversation = append(h.conversation, turn)

	if h.onResponse != nil {
		h.onResponse(resp.Content)
	}

	return nil
}

// MultiTurnChat สนทนาหลาย turn ตามสคริปต์ที่กำหนด
func (h *Harness) MultiTurnChat(messages []string) ([]string, error) {
	responses := make([]string, 0, len(messages))

	for _, msg := range messages {
		resp, err := h.Chat(msg)
		if err != nil {
			return responses, err
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

// GetConversationHistory คืนค่าประวัติการสนทนา
func (h *Harness) GetConversationHistory() []ConversationTurn {
	h.mu.RLock()
	defer h.mu.RUnlock()

	history := make([]ConversationTurn, len(h.conversation))
	copy(history, h.conversation)
	return history
}

// GetLastResponse คืนค่าการตอบสนองล่าสุด
func (h *Harness) GetLastResponse() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for i := len(h.conversation) - 1; i >= 0; i-- {
		if h.conversation[i].Role == "assistant" {
			return h.conversation[i].Content
		}
	}
	return ""
}

// GetOutput คืนค่า output ทั้งหมดที่ถูก capture
func (h *Harness) GetOutput() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.outputBuffer.String()
}

// ClearConversation ล้างประวัติการสนทนา
func (h *Harness) ClearConversation() *Harness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conversation = make([]ConversationTurn, 0)
	h.outputBuffer.Reset()
	return h
}

// Reset รีเซ็ต harness ทั้งหมด
func (h *Harness) Reset() *Harness {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.conversation = make([]ConversationTurn, 0)
	h.toolResults = make(map[string]ToolResult)
	h.outputBuffer.Reset()
	h.provider.Reset()

	return h
}

// AssertResponseContains ตรวจสอบว่าการตอบสนองล่าสุดมีข้อความที่ต้องการ
func (h *Harness) AssertResponseContains(expected string) error {
	last := h.GetLastResponse()
	if !strings.Contains(strings.ToLower(last), strings.ToLower(expected)) {
		return fmt.Errorf("expected response to contain %q, got: %q", expected, last)
	}
	return nil
}

// AssertResponseEquals ตรวจสอบว่าการตอบสนองล่าสุดตรงกับที่ต้องการ
func (h *Harness) AssertResponseEquals(expected string) error {
	last := h.GetLastResponse()
	if strings.TrimSpace(last) != strings.TrimSpace(expected) {
		return fmt.Errorf("expected response %q, got: %q", expected, last)
	}
	return nil
}

// AssertToolCalled ตรวจสอบว่ามีการเรียก tool ที่ต้องการ
func (h *Harness) AssertToolCalled(toolName string) error {
	if !h.provider.VerifyToolCall(toolName) {
		return fmt.Errorf("expected tool %q to be called", toolName)
	}
	return nil
}

// AssertNoErrors ตรวจสอบว่าไม่มี error
func (h *Harness) AssertNoErrors() error {
	return h.provider.AssertNoErrors()
}

// GetCallCount คืนค่าจำนวนครั้งที่ provider ถูกเรียก
func (h *Harness) GetCallCount() int {
	return h.provider.GetCallCount()
}

// buildMessages สร้าง messages สำหรับส่งไปยัง provider
func (h *Harness) buildMessages() []providers.Message {
	messages := make([]providers.Message, 0)

	// System message
	messages = append(messages, providers.Message{
		Role:    "system",
		Content: "You are a helpful assistant.",
	})

	// Conversation history
	for _, turn := range h.conversation {
		switch turn.Role {
		case "user":
			messages = append(messages, providers.Message{
				Role:    "user",
				Content: turn.Content,
			})
		case "assistant":
			msg := providers.Message{
				Role:    "assistant",
				Content: turn.Content,
			}
			// เพิ่ม tool calls ถ้ามี
			if len(turn.ToolCalls) > 0 {
				msg.ToolCalls = turn.ToolCalls
			}
			messages = append(messages, msg)

			// เพิ่ม tool results
			for _, tr := range turn.ToolResults {
				messages = append(messages, providers.Message{
					Role:       "tool",
					Content:    tr.Result,
					ToolCallID: tr.ToolName,
				})
			}
		}
	}

	return messages
}

// buildToolDefinitions สร้าง tool definitions พื้นฐาน
func (h *Harness) buildToolDefinitions() []providers.ToolDefinition {
	// สร้าง tool definitions พื้นฐานสำหรับการทดสอบ
	tools := []providers.ToolDefinition{
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "web_search",
				Description: "Search the web for information",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "Search query",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "read_file",
				Description: "Read a file from the workspace",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "File path",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "write_file",
				Description: "Write content to a file",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "File path",
						},
						"content": map[string]any{
							"type":        "string",
							"description": "File content",
						},
					},
					"required": []string{"path", "content"},
				},
			},
		},
		// Agent Team Tools
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "spawn_subagent",
				Description: "Spawn a subagent with a specific role to complete a task",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"role": map[string]any{
							"type":        "string",
							"description": "Role of the subagent (architect, coder, researcher, qa, writer, reviewer, coordinator)",
							"enum":        []string{"architect", "coder", "researcher", "qa", "writer", "reviewer", "coordinator"},
						},
						"task": map[string]any{
							"type":        "string",
							"description": "Task description for the subagent",
						},
						"context": map[string]any{
							"type":        "string",
							"description": "Additional context for the subagent",
						},
					},
					"required": []string{"role", "task"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "subagent_status",
				Description: "Check status of subagents or get list of active subagents",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"action": map[string]any{
							"type":        "string",
							"description": "Action to perform",
							"enum":        []string{"get", "active", "list"},
						},
						"task_id": map[string]any{
							"type":        "string",
							"description": "Task ID to check status (required for 'get' action)",
						},
					},
					"required": []string{"action"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "start_meeting",
				Description: "Start a meeting with multiple agents",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"topic": map[string]any{
							"type":        "string",
							"description": "Meeting topic",
						},
						"participants": map[string]any{
							"type":        "array",
							"description": "List of agent IDs to invite",
							"items": map[string]any{
								"type": "string",
							},
						},
						"duration": map[string]any{
							"type":        "integer",
							"description": "Meeting duration in minutes",
						},
					},
					"required": []string{"topic", "participants"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "send_agent_message",
				Description: "Send a message to another agent",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"to": map[string]any{
							"type":        "string",
							"description": "Recipient agent ID",
						},
						"subject": map[string]any{
							"type":        "string",
							"description": "Message subject",
						},
						"message": map[string]any{
							"type":        "string",
							"description": "Message content",
						},
						"priority": map[string]any{
							"type":        "string",
							"description": "Message priority",
							"enum":        []string{"low", "normal", "high", "urgent"},
						},
					},
					"required": []string{"to", "subject", "message"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "check_agent_inbox",
				Description: "Check the agent's mailbox for messages",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"unread_only": map[string]any{
							"type":        "boolean",
							"description": "Only show unread messages",
						},
						"limit": map[string]any{
							"type":        "integer",
							"description": "Maximum number of messages to return",
						},
					},
					"required": []string{},
				},
			},
		},
	}

	return tools
}

// executeToolCalls จำลองการเรียกใช้งาน tools
func (h *Harness) executeToolCalls(ctx context.Context, toolCalls []providers.ToolCall) []ToolResult {
	results := make([]ToolResult, 0, len(toolCalls))

	for _, tc := range toolCalls {
		result := h.executeTool(ctx, tc)
		results = append(results, result)
	}

	return results
}

// executeTool จำลองการเรียกใช้งาน tool เดียว
func (h *Harness) executeTool(ctx context.Context, tc providers.ToolCall) ToolResult {
	// เรียก callback
	if h.onToolCall != nil {
		h.onToolCall(tc.Name, tc.Arguments)
	}

	// จำลองผลลัพธ์ตามชื่อ tool
	var result string
	switch tc.Name {
	case "web_search":
		query, _ := tc.Arguments["query"].(string)
		result = fmt.Sprintf(`{"results": [{"title": "Result for %s", "url": "https://example.com", "snippet": "This is a mock search result for %s"}]}`, query, query)
	case "read_file":
		path, _ := tc.Arguments["path"].(string)
		result = fmt.Sprintf(`{"content": "Mock content of file %s"}`, path)
	case "write_file":
		path, _ := tc.Arguments["path"].(string)
		result = fmt.Sprintf(`{"success": true, "path": "%s", "bytes_written": 100}`, path)
	default:
		result = fmt.Sprintf(`{"result": "Mock result for %s"}`, tc.Name)
	}

	return ToolResult{
		ToolName: tc.Name,
		Args:     tc.Arguments,
		Result:   result,
	}
}

// PrintConversation พิมพ์ประวัติการสนทนาออกมา
func (h *Harness) PrintConversation() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var sb strings.Builder
	for _, turn := range h.conversation {
		switch turn.Role {
		case "user":
			sb.WriteString(fmt.Sprintf("[User]: %s\n", turn.Content))
		case "assistant":
			sb.WriteString(fmt.Sprintf("[Assistant]: %s\n", turn.Content))
			if len(turn.ToolCalls) > 0 {
				for _, tc := range turn.ToolCalls {
					sb.WriteString(fmt.Sprintf("  [Tool Call]: %s\n", tc.Name))
				}
			}
		}
	}
	return sb.String()
}
