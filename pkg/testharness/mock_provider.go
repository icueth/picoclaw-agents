// PicoClaw - LLM Chat Test Harness
// Mock Provider สำหรับจำลองการตอบสนองของ LLM

package testharness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/providers"
)

// MockResponse เก็บข้อมูลการตอบสนองที่จำลอง
type MockResponse struct {
	Content      string
	ToolCalls    []providers.ToolCall
	Reasoning    string
	Delay        time.Duration
	ShouldError  bool
	ErrorMessage string
}

// ResponseMatcher เป็นฟังก์ชันสำหรับตรวจสอบว่าข้อความตรงกับเงื่อนไขหรือไม่
type ResponseMatcher func(message string) bool

// ResponseRule กฎสำหรับกำหนดการตอบสนอง
type ResponseRule struct {
	Matcher      ResponseMatcher
	Response     MockResponse
	Description  string
	Priority     int
	MaxCalls     int
	CallCount    int
}

// MockProvider จำลองการทำงานของ LLM Provider
type MockProvider struct {
	mu            sync.RWMutex
	rules         []ResponseRule
	defaultResp   MockResponse
	callHistory   []CallRecord
	modelName     string
	onChunkDelay  time.Duration
	captureInputs bool
	inputs        []providers.Message
}

// CallRecord บันทึกการเรียกใช้งาน
type CallRecord struct {
	Timestamp time.Time
	Messages  []providers.Message
	Tools     []providers.ToolDefinition
	Model     string
	Response  *providers.LLMResponse
	Error     error
}

// NewMockProvider สร้าง Mock Provider ใหม่
func NewMockProvider() *MockProvider {
	return &MockProvider{
		rules:         make([]ResponseRule, 0),
		callHistory:   make([]CallRecord, 0),
		modelName:     "mock-model",
		onChunkDelay:  10 * time.Millisecond,
		captureInputs: true,
		inputs:        make([]providers.Message, 0),
		defaultResp: MockResponse{
			Content: "This is a default mock response.",
		},
	}
}

// Chat จำลองการเรียก LLM
func (m *MockProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// บันทึก inputs
	if m.captureInputs {
		m.inputs = append(m.inputs, messages...)
	}

	// หาข้อความล่าสุดจาก user
	lastMessage := m.extractLastUserMessage(messages)

	// หากฎที่ตรงกัน
	var matchedRule *ResponseRule
	for i := range m.rules {
		if m.rules[i].MaxCalls > 0 && m.rules[i].CallCount >= m.rules[i].MaxCalls {
			continue
		}
		if m.rules[i].Matcher(lastMessage) {
			if matchedRule == nil || m.rules[i].Priority > matchedRule.Priority {
				matchedRule = &m.rules[i]
			}
		}
	}

	var response MockResponse
	if matchedRule != nil {
		matchedRule.CallCount++
		response = matchedRule.Response
	} else {
		response = m.defaultResp
	}

	// จำลอง delay
	if response.Delay > 0 {
		select {
		case <-time.After(response.Delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// จำลอง streaming
	if onChunk, ok := opts["on_chunk"].(func(string)); ok && onChunk != nil {
		m.simulateStreaming(response.Content, onChunk)
	}

	// สร้าง response
	resp := &providers.LLMResponse{
		Content:   response.Content,
		ToolCalls: response.ToolCalls,
		Reasoning: response.Reasoning,
	}

	// บันทึกประวัติ
	record := CallRecord{
		Timestamp: time.Now(),
		Messages:  messages,
		Tools:     tools,
		Model:     model,
		Response:  resp,
	}
	if response.ShouldError {
		record.Error = errors.New(response.ErrorMessage)
	}
	m.callHistory = append(m.callHistory, record)

	if response.ShouldError {
		return nil, errors.New(response.ErrorMessage)
	}

	return resp, nil
}

// GetDefaultModel คืนค่าชื่อ model
func (m *MockProvider) GetDefaultModel() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.modelName
}

// WithModelName กำหนดชื่อ model
func (m *MockProvider) WithModelName(name string) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelName = name
	return m
}

// WithResponsePattern กำหนดการตอบสนองตาม pattern (contains)
func (m *MockProvider) WithResponsePattern(pattern, response string) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rules = append(m.rules, ResponseRule{
		Matcher: func(msg string) bool {
			return strings.Contains(strings.ToLower(msg), strings.ToLower(pattern))
		},
		Response: MockResponse{
			Content: response,
		},
		Description: fmt.Sprintf("Contains: %s", pattern),
		Priority:    1,
	})
	return m
}

// WithRegexPattern กำหนดการตอบสนองตาม regex
func (m *MockProvider) WithRegexPattern(regex string, response string) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()

	re := regexp.MustCompile(regex)
	m.rules = append(m.rules, ResponseRule{
		Matcher: func(msg string) bool {
			return re.MatchString(msg)
		},
		Response: MockResponse{
			Content: response,
		},
		Description: fmt.Sprintf("Regex: %s", regex),
		Priority:    2,
	})
	return m
}

// WithExactMatch กำหนดการตอบสนองตามข้อความตรงตัว
func (m *MockProvider) WithExactMatch(exactText, response string) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rules = append(m.rules, ResponseRule{
		Matcher: func(msg string) bool {
			return strings.TrimSpace(msg) == exactText
		},
		Response: MockResponse{
			Content: response,
		},
		Description: fmt.Sprintf("Exact: %s", exactText),
		Priority:    10,
	})
	return m
}

// WithToolCallResponse กำหนดการตอบสนองที่มี tool call
func (m *MockProvider) WithToolCallResponse(pattern, toolName string, args map[string]any) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()

	argsJSON, _ := json.Marshal(args)
	toolCall := providers.ToolCall{
		ID:   fmt.Sprintf("call_%d", len(m.rules)),
		Type: "function",
		Name: toolName,
		Function: &providers.FunctionCall{
			Name:      toolName,
			Arguments: string(argsJSON),
		},
	}

	m.rules = append(m.rules, ResponseRule{
		Matcher: func(msg string) bool {
			return strings.Contains(strings.ToLower(msg), strings.ToLower(pattern))
		},
		Response: MockResponse{
			Content:   fmt.Sprintf("I'll help you with that using %s", toolName),
			ToolCalls: []providers.ToolCall{toolCall},
		},
		Description: fmt.Sprintf("Tool call: %s", toolName),
		Priority:    5,
	})
	return m
}

// WithMultiToolCallResponse กำหนดการตอบสนองที่มีหลาย tool calls
func (m *MockProvider) WithMultiToolCallResponse(pattern string, toolCalls []providers.ToolCall) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rules = append(m.rules, ResponseRule{
		Matcher: func(msg string) bool {
			return strings.Contains(strings.ToLower(msg), strings.ToLower(pattern))
		},
		Response: MockResponse{
			Content:   "I'll execute multiple actions for you.",
			ToolCalls: toolCalls,
		},
		Description: fmt.Sprintf("Multi tool calls: %d", len(toolCalls)),
		Priority:    5,
	})
	return m
}

// WithErrorResponse กำหนดการตอบสนองที่เป็น error
func (m *MockProvider) WithErrorResponse(pattern, errorMsg string) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rules = append(m.rules, ResponseRule{
		Matcher: func(msg string) bool {
			return strings.Contains(strings.ToLower(msg), strings.ToLower(pattern))
		},
		Response: MockResponse{
			ShouldError:  true,
			ErrorMessage: errorMsg,
		},
		Description: fmt.Sprintf("Error: %s", errorMsg),
		Priority:    100,
	})
	return m
}

// WithReasoningResponse กำหนดการตอบสนองที่มี reasoning
func (m *MockProvider) WithReasoningResponse(pattern, reasoning, content string) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rules = append(m.rules, ResponseRule{
		Matcher: func(msg string) bool {
			return strings.Contains(strings.ToLower(msg), strings.ToLower(pattern))
		},
		Response: MockResponse{
			Content:   content,
			Reasoning: reasoning,
		},
		Description: fmt.Sprintf("Reasoning: %s", pattern),
		Priority:    3,
	})
	return m
}

// WithDelayedResponse กำหนดการตอบสนองที่มี delay
func (m *MockProvider) WithDelayedResponse(pattern, response string, delay time.Duration) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rules = append(m.rules, ResponseRule{
		Matcher: func(msg string) bool {
			return strings.Contains(strings.ToLower(msg), strings.ToLower(pattern))
		},
		Response: MockResponse{
			Content: response,
			Delay:   delay,
		},
		Description: fmt.Sprintf("Delayed (%v): %s", delay, pattern),
		Priority:    2,
	})
	return m
}

// WithDefaultResponse กำหนดค่า default response
func (m *MockProvider) WithDefaultResponse(response string) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultResp = MockResponse{
		Content: response,
	}
	return m
}

// WithMaxCalls จำกัดจำนวนครั้งที่กฎนี้จะถูกใช้
func (m *MockProvider) WithMaxCalls(pattern string, maxCalls int) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.rules {
		if m.rules[i].Description == fmt.Sprintf("Contains: %s", pattern) {
			m.rules[i].MaxCalls = maxCalls
			break
		}
	}
	return m
}

// Reset รีเซ็ตสถานะทั้งหมด
func (m *MockProvider) Reset() *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rules = make([]ResponseRule, 0)
	m.callHistory = make([]CallRecord, 0)
	m.inputs = make([]providers.Message, 0)
	m.defaultResp = MockResponse{
		Content: "This is a default mock response.",
	}
	return m
}

// ClearHistory ล้างประวัติการเรียก
func (m *MockProvider) ClearHistory() *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callHistory = make([]CallRecord, 0)
	return m
}

// GetCallHistory คืนค่าประวัติการเรียก
func (m *MockProvider) GetCallHistory() []CallRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history := make([]CallRecord, len(m.callHistory))
	copy(history, m.callHistory)
	return history
}

// GetCallCount คืนค่าจำนวนครั้งที่ถูกเรียก
func (m *MockProvider) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.callHistory)
}

// GetInputs คืนค่า inputs ทั้งหมดที่ถูกส่งมา
func (m *MockProvider) GetInputs() []providers.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	inputs := make([]providers.Message, len(m.inputs))
	copy(inputs, m.inputs)
	return inputs
}

// GetLastInput คืนค่า input ล่าสุด
func (m *MockProvider) GetLastInput() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := len(m.inputs) - 1; i >= 0; i-- {
		if m.inputs[i].Role == "user" {
			return m.inputs[i].Content
		}
	}
	return ""
}

// SetOnChunkDelay กำหนด delay สำหรับ streaming
func (m *MockProvider) SetOnChunkDelay(delay time.Duration) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onChunkDelay = delay
	return m
}

// simulateStreaming จำลองการสตรีมข้อความ
func (m *MockProvider) simulateStreaming(content string, onChunk func(string)) {
	// แบ่งข้อความเป็นคำๆ
	words := strings.Fields(content)
	for _, word := range words {
		onChunk(word + " ")
		time.Sleep(m.onChunkDelay)
	}
}

// extractLastUserMessage ดึงข้อความล่าสุดจาก user
func (m *MockProvider) extractLastUserMessage(messages []providers.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	return ""
}

// VerifyCall ตรวจสอบว่ามีการเรียกด้วยข้อความที่ตรงกับ pattern
func (m *MockProvider) VerifyCall(pattern string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, record := range m.callHistory {
		msg := m.extractLastUserMessage(record.Messages)
		if strings.Contains(strings.ToLower(msg), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// VerifyToolCall ตรวจสอบว่ามีการเรียก tool ที่ตรงกับชื่อ
func (m *MockProvider) VerifyToolCall(toolName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, record := range m.callHistory {
		if record.Response != nil {
			for _, tc := range record.Response.ToolCalls {
				if tc.Name == toolName {
					return true
				}
			}
		}
	}
	return false
}

// AssertNoErrors ตรวจสอบว่าไม่มี error
func (m *MockProvider) AssertNoErrors() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, record := range m.callHistory {
		if record.Error != nil {
			return fmt.Errorf("unexpected error: %v", record.Error)
		}
	}
	return nil
}

// WithSystemPrompt กำหนด system prompt สำหรับทดสอบ
func (m *MockProvider) WithSystemPrompt(prompt string) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	// เก็บ system prompt ใน default response เพื่อให้เข้าถึงได้
	m.defaultResp.Content = fmt.Sprintf("[%s] ", prompt)
	return m
}

// WithToolError กำหนดการตอบสนองที่เกิด error จาก tool
func (m *MockProvider) WithToolError(toolName string, err error) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()

	argsJSON, _ := json.Marshal(map[string]any{})
	toolCall := providers.ToolCall{
		ID:   fmt.Sprintf("call_error_%s", toolName),
		Type: "function",
		Name: toolName,
		Function: &providers.FunctionCall{
			Name:      toolName,
			Arguments: string(argsJSON),
		},
	}

	m.rules = append(m.rules, ResponseRule{
		Matcher: func(msg string) bool {
			return true // จับทุกข้อความ
		},
		Response: MockResponse{
			Content:   fmt.Sprintf("Tool %s failed", toolName),
			ToolCalls: []providers.ToolCall{toolCall},
			ShouldError: true,
			ErrorMessage: err.Error(),
		},
		Description: fmt.Sprintf("Tool error: %s", toolName),
		Priority:    1,
	})
	return m
}

// GetToolCallCount คืนค่าจำนวนครั้งที่ tool ถูกเรียก
func (m *MockProvider) GetToolCallCount(toolName string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, record := range m.callHistory {
		if record.Response != nil {
			for _, tc := range record.Response.ToolCalls {
				if tc.Name == toolName {
					count++
				}
			}
		}
	}
	return count
}
