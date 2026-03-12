// PicoClaw - Real LLM Test Harness
// ระบบทดสอบที่ใช้งาน LLM จริงตาม config

package testharness

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/providers"
)

// RealLLMTestHarness ระบบทดสอบที่ใช้งาน LLM จริง
type RealLLMTestHarness struct {
	mu           sync.RWMutex
	config       *config.Config
	provider     providers.LLMProvider
	modelName    string
	sessionKey   string
	channel      string
	chatID       string

	// Test state
	conversation []TestMessage
	toolCalls    []ToolCallRecord
	metrics      TestMetrics

	// Configuration
	timeout      time.Duration
	maxTokens    int
	temperature  float64
	maxIterations int
}

// TestMessage ข้อความในการทดสอบ
type TestMessage struct {
	Role        string                 `json:"role"`
	Content     string                 `json:"content"`
	ToolCalls   []providers.ToolCall   `json:"tool_calls,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Latency     time.Duration          `json:"latency"`
}

// ToolCallRecord บันทึกการเรียกใช้งาน tool
type ToolCallRecord struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]any         `json:"arguments"`
	Result    string                 `json:"result"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// TestMetrics metrics การทดสอบ
type TestMetrics struct {
	TotalCalls      int           `json:"total_calls"`
	TotalTokens     int           `json:"total_tokens"`
	TotalLatency    time.Duration `json:"total_latency"`
	AvgLatency      time.Duration `json:"avg_latency"`
	Errors          int           `json:"errors"`
	ToolCalls       int           `json:"tool_calls"`
	StartTime       time.Time     `json:"start_time"`
}

// NewRealLLMTestHarness สร้าง Test Harness ที่ใช้ LLM จริง
func NewRealLLMTestHarness(cfgPath string) (*RealLLMTestHarness, error) {
	// โหลด config
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// สร้าง provider จาก config
	provider, modelName, err := createProviderFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	harness := &RealLLMTestHarness{
		config:        cfg,
		provider:      provider,
		modelName:     modelName,
		sessionKey:    "test-session",
		channel:       "test",
		chatID:        "test-chat",
		conversation:  make([]TestMessage, 0),
		toolCalls:     make([]ToolCallRecord, 0),
		timeout:       60 * time.Second,
		maxTokens:     cfg.Agents.Defaults.MaxTokens,
		maxIterations: cfg.Agents.Defaults.MaxToolIterations,
	}

	if harness.maxTokens == 0 {
		harness.maxTokens = 4096
	}
	if harness.maxIterations == 0 {
		harness.maxIterations = 20
	}

	// Set default temperature if not specified
	if cfg.Agents.Defaults.Temperature != nil {
		harness.temperature = *cfg.Agents.Defaults.Temperature
	} else {
		harness.temperature = 0.7
	}

	return harness, nil
}

// NewRealLLMTestHarnessWithModel สร้าง Test Harness ด้วย model ที่ระบุ
func NewRealLLMTestHarnessWithModel(cfgPath, modelName string) (*RealLLMTestHarness, error) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// หา model config
	modelCfg, err := cfg.GetModelConfig(modelName)
	if err != nil {
		return nil, fmt.Errorf("model %s not found: %w", modelName, err)
	}

	provider, resolvedModel, err := providers.CreateProviderFromConfig(modelCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	harness := &RealLLMTestHarness{
		config:        cfg,
		provider:      provider,
		modelName:     resolvedModel,
		sessionKey:    "test-session",
		channel:       "test",
		chatID:        "test-chat",
		conversation:  make([]TestMessage, 0),
		toolCalls:     make([]ToolCallRecord, 0),
		timeout:       60 * time.Second,
		maxTokens:     cfg.Agents.Defaults.MaxTokens,
		temperature:   0.7,
		maxIterations: 20,
	}

	if harness.maxTokens == 0 {
		harness.maxTokens = 4096
	}

	return harness, nil
}

// WithTimeout กำหนด timeout
func (h *RealLLMTestHarness) WithTimeout(timeout time.Duration) *RealLLMTestHarness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.timeout = timeout
	return h
}

// WithMaxTokens กำหนด max tokens
func (h *RealLLMTestHarness) WithMaxTokens(maxTokens int) *RealLLMTestHarness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.maxTokens = maxTokens
	return h
}

// WithTemperature กำหนด temperature
func (h *RealLLMTestHarness) WithTemperature(temp float64) *RealLLMTestHarness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.temperature = temp
	return h
}

// Chat ส่งข้อความไปยัง LLM และรับการตอบสนอง
func (h *RealLLMTestHarness) Chat(message string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	return h.ChatWithContext(ctx, message)
}

// ChatWithContext ส่งข้อความพร้อม context
func (h *RealLLMTestHarness) ChatWithContext(ctx context.Context, message string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// บันทึกข้อความจาก user
	h.conversation = append(h.conversation, TestMessage{
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	})

	// สร้าง messages สำหรับส่งไปยัง LLM
	messages := h.buildMessages()

	// เรียก LLM
	start := time.Now()
	resp, err := h.provider.Chat(ctx, messages, nil, h.modelName, map[string]any{
		"max_tokens":  h.maxTokens,
		"temperature": h.temperature,
	})
	latency := time.Since(start)

	if err != nil {
		h.metrics.Errors++
		return "", fmt.Errorf("LLM call failed: %w", err)
	}

	// อัปเดต metrics
	h.metrics.TotalCalls++
	h.metrics.TotalLatency += latency
	h.metrics.AvgLatency = h.metrics.TotalLatency / time.Duration(h.metrics.TotalCalls)

	// บันทึก response
	h.conversation = append(h.conversation, TestMessage{
		Role:      "assistant",
		Content:   resp.Content,
		ToolCalls: resp.ToolCalls,
		Timestamp: time.Now(),
		Latency:   latency,
	})

	return resp.Content, nil
}

// ChatStreaming ส่งข้อความแบบ streaming
func (h *RealLLMTestHarness) ChatStreaming(message string, onChunk func(string)) error {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	return h.ChatStreamingWithContext(ctx, message, onChunk)
}

// ChatStreamingWithContext ส่งข้อความแบบ streaming พร้อม context
func (h *RealLLMTestHarness) ChatStreamingWithContext(ctx context.Context, message string, onChunk func(string)) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// บันทึกข้อความจาก user
	h.conversation = append(h.conversation, TestMessage{
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	})

	// สร้าง messages
	messages := h.buildMessages()

	// เรียก LLM พร้อม streaming
	start := time.Now()
	resp, err := h.provider.Chat(ctx, messages, nil, h.modelName, map[string]any{
		"max_tokens":  h.maxTokens,
		"temperature": h.temperature,
		"on_chunk":    onChunk,
	})
	latency := time.Since(start)

	if err != nil {
		h.metrics.Errors++
		return fmt.Errorf("LLM streaming call failed: %w", err)
	}

	// อัปเดต metrics
	h.metrics.TotalCalls++
	h.metrics.TotalLatency += latency
	h.metrics.AvgLatency = h.metrics.TotalLatency / time.Duration(h.metrics.TotalCalls)

	// บันทึก response
	h.conversation = append(h.conversation, TestMessage{
		Role:      "assistant",
		Content:   resp.Content,
		ToolCalls: resp.ToolCalls,
		Timestamp: time.Now(),
		Latency:   latency,
	})

	return nil
}

// MultiTurnChat สนทนาหลาย turn
func (h *RealLLMTestHarness) MultiTurnChat(messages []string) ([]string, error) {
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
func (h *RealLLMTestHarness) GetConversationHistory() []TestMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()

	history := make([]TestMessage, len(h.conversation))
	copy(history, h.conversation)
	return history
}

// GetLastResponse คืนค่าการตอบสนองล่าสุด
func (h *RealLLMTestHarness) GetLastResponse() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for i := len(h.conversation) - 1; i >= 0; i-- {
		if h.conversation[i].Role == "assistant" {
			return h.conversation[i].Content
		}
	}
	return ""
}

// GetMetrics คืนค่า metrics
func (h *RealLLMTestHarness) GetMetrics() TestMetrics {
	h.mu.RLock()
	defer h.mu.RUnlock()

	metrics := h.metrics
	if metrics.TotalCalls > 0 {
		metrics.AvgLatency = metrics.TotalLatency / time.Duration(metrics.TotalCalls)
	}
	return metrics
}

// ClearConversation ล้างประวัติการสนทนา
func (h *RealLLMTestHarness) ClearConversation() *RealLLMTestHarness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conversation = make([]TestMessage, 0)
	return h
}

// ResetMetrics รีเซ็ต metrics
func (h *RealLLMTestHarness) ResetMetrics() *RealLLMTestHarness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.metrics = TestMetrics{}
	return h
}

// Reset รีเซ็ตทั้งหมด
func (h *RealLLMTestHarness) Reset() *RealLLMTestHarness {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conversation = make([]TestMessage, 0)
	h.toolCalls = make([]ToolCallRecord, 0)
	h.metrics = TestMetrics{}
	return h
}

// SaveConversation บันทึกประวัติการสนทนาลงไฟล์
func (h *RealLLMTestHarness) SaveConversation(path string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data := struct {
		Model        string        `json:"model"`
		Messages     []TestMessage `json:"messages"`
		ToolCalls    []ToolCallRecord `json:"tool_calls"`
		Metrics      TestMetrics   `json:"metrics"`
		SavedAt      time.Time     `json:"saved_at"`
	}{
		Model:     h.modelName,
		Messages:  h.conversation,
		ToolCalls: h.toolCalls,
		Metrics:   h.metrics,
		SavedAt:   time.Now(),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonData, 0644)
}

// LoadConversation โหลดประวัติการสนทนาจากไฟล์
func (h *RealLLMTestHarness) LoadConversation(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var saved struct {
		Messages  []TestMessage    `json:"messages"`
		ToolCalls []ToolCallRecord `json:"tool_calls"`
		Metrics   TestMetrics      `json:"metrics"`
	}

	if err := json.Unmarshal(data, &saved); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.conversation = saved.Messages
	h.toolCalls = saved.ToolCalls
	h.metrics = saved.Metrics

	return nil
}

// PrintConversation พิมพ์ประวัติการสนทนา
func (h *RealLLMTestHarness) PrintConversation() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Model: %s\n", h.modelName))
	sb.WriteString(strings.Repeat("=", 50) + "\n")

	for _, msg := range h.conversation {
		switch msg.Role {
		case "user":
			sb.WriteString(fmt.Sprintf("\n[User]: %s\n", msg.Content))
		case "assistant":
			sb.WriteString(fmt.Sprintf("[Assistant]: %s\n", msg.Content))
			if msg.Latency > 0 {
				sb.WriteString(fmt.Sprintf("  (latency: %v)\n", msg.Latency))
			}
		}
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Total calls: %d | Avg latency: %v\n",
		h.metrics.TotalCalls, h.metrics.AvgLatency))

	return sb.String()
}

// buildMessages สร้าง messages สำหรับส่งไปยัง LLM
func (h *RealLLMTestHarness) buildMessages() []providers.Message {
	messages := make([]providers.Message, 0)

	// System message
	messages = append(messages, providers.Message{
		Role:    "system",
		Content: "You are a helpful assistant.",
	})

	// Conversation history
	for _, msg := range h.conversation {
		messages = append(messages, providers.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return messages
}

// createProviderFromConfig สร้าง provider จาก config
func createProviderFromConfig(cfg *config.Config) (providers.LLMProvider, string, error) {
	// ใช้ default model
	modelName := cfg.Agents.Defaults.GetModelName()
	if modelName == "" {
		return nil, "", fmt.Errorf("no default model configured")
	}

	// หา model config
	modelCfg, err := cfg.GetModelConfig(modelName)
	if err != nil {
		return nil, "", fmt.Errorf("model %s not found: %w", modelName, err)
	}

	return providers.CreateProviderFromConfig(modelCfg)
}

// GetConfig คืนค่า config
func (h *RealLLMTestHarness) GetConfig() *config.Config {
	return h.config
}

// GetProvider คืนค่า provider
func (h *RealLLMTestHarness) GetProvider() providers.LLMProvider {
	return h.provider
}

// GetModelName คืนค่าชื่อ model
func (h *RealLLMTestHarness) GetModelName() string {
	return h.modelName
}
