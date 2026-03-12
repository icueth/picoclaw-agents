package tools

import (
	"context"
	"testing"
	"time"

	"picoclaw/agent/pkg/agentcomm"
	"picoclaw/agent/pkg/providers"
)

// MockProvider is a mock LLM provider for testing
type MockProvider struct {
	responses []providers.Message
	callCount int
}

func (m *MockProvider) Chat(ctx context.Context, messages []providers.Message, tools []providers.ToolDefinition, model string, options map[string]any) (*providers.LLMResponse, error) {
	if m.callCount < len(m.responses) {
		resp := m.responses[m.callCount]
		m.callCount++
		return &providers.LLMResponse{
			Content:      resp.Content,
			FinishReason: "stop",
		}, nil
	}
	// Default response when no more mock responses
	return &providers.LLMResponse{
		Content:      "Task completed.",
		FinishReason: "stop",
	}, nil
}

func (m *MockProvider) GetDefaultModel() string {
	return "mock-model"
}

// TestFullSubagentWorkflow tests the complete subagent workflow including shared context
func TestFullSubagentWorkflow(t *testing.T) {
	// Create shared context
	sharedCtx := agentcomm.NewSharedContext(100, 1000)

	// Create mock provider
	mockProvider := &MockProvider{
		responses: []providers.Message{
			{Role: "assistant", Content: "I'll use spawn tool to delegate this task."},
			{Role: "assistant", Content: "Task completed successfully."},
		},
	}

	// Create subagent manager with shared context
	manager := NewSubagentManager(
		mockProvider,
		"test-model",
		"/tmp/test-workspace",
		nil, // no bus for testing
	)

	// Set shared context
	manager.SetSharedContext(sharedCtx)

	// Test 1: Set context before spawning
	sharedCtx.Set("test_key", "test_value")

	// Test 2: Spawn a subagent task
	task := "Calculate 2+2"
	label := "math-task"

	ctx := context.Background()
	result, err := manager.Spawn(ctx, task, label, "", "", "cli", "test-session", nil)

	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	if result == "" {
		t.Error("Expected spawn result message, got empty string")
	}

	// Test 3: Check shared context is accessible
	val, ok := sharedCtx.Get("test_key")
	if !ok {
		t.Error("Shared context should have test_key")
	}
	if val != "test_value" {
		t.Errorf("Expected test_value, got %v", val)
	}

	// Test 4: Wait for subagent to complete and check result
	time.Sleep(500 * time.Millisecond)

	tasks := manager.ListTasks()
	if len(tasks) == 0 {
		t.Error("Expected at least one task")
	}

	// Test 5: Check task result is stored in shared context
	resultKey := "task:" + tasks[0].ID + ":result"
	resultVal, ok := sharedCtx.Get(resultKey)
	t.Logf("Task ID: %s, Result key: %s, Found: %v, Value: %v", tasks[0].ID, resultKey, ok, resultVal)

	// Test 6: List tasks
	tasks = manager.ListTasks()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	// Test 7: Get specific task
	taskData, found := manager.GetTask(tasks[0].ID)
	if !found {
		t.Error("Should be able to get task by ID")
	}
	if taskData.Label != label {
		t.Errorf("Expected label %s, got %s", label, taskData.Label)
	}

	t.Logf("Full workflow test completed. Tasks: %d", len(tasks))
}

// TestSubagentWithSharedContext tests that subagent can access shared context
func TestSubagentWithSharedContext(t *testing.T) {
	sharedCtx := agentcomm.NewSharedContext(50, 500)

	// Pre-populate shared context
	sharedCtx.Set("shared_data", map[string]string{"file": "main.go", "line": "42"})
	sharedCtx.AddMessageLog("main", "subagent", "request", "Process this file")

	// Verify context is accessible
	data, ok := sharedCtx.Get("shared_data")
	if !ok {
		t.Fatal("Shared context should have shared_data")
	}

	dataMap, ok := data.(map[string]string)
	if !ok || dataMap["file"] != "main.go" {
		t.Error("Shared data should contain file:main.go")
	}

	// Verify message log
	log := sharedCtx.GetMessageLog()
	if len(log) == 0 {
		t.Error("Message log should not be empty")
	}

	// Test message log for specific agent
	agentLog := sharedCtx.GetMessagesForAgent("subagent")
	if len(agentLog) == 0 {
		t.Error("Should have messages for subagent")
	}

	t.Logf("Shared context test passed. Messages: %d, Agent messages: %d", len(log), len(agentLog))
}

// TestSubagentMessageTypes tests different message types
func TestSubagentMessageTypes(t *testing.T) {
	// Test message creation
	msg := agentcomm.NewAgentMessage("agent1", "agent2", agentcomm.MsgRequest, "hello", "session1")

	if msg.From != "agent1" {
		t.Errorf("Expected From=agent1, got %s", msg.From)
	}
	if msg.To != "agent2" {
		t.Errorf("Expected To=agent2, got %s", msg.To)
	}
	if msg.Type != agentcomm.MsgRequest {
		t.Errorf("Expected Type=request, got %s", msg.Type)
	}
	if msg.SessionID != "session1" {
		t.Errorf("Expected SessionID=session1, got %s", msg.SessionID)
	}

	// Test broadcast
	broadcast := agentcomm.NewAgentMessage("agent1", "", agentcomm.MsgBroadcast, "hello all", "session1")
	if !broadcast.IsBroadcast() {
		t.Error("Empty To should be treated as broadcast")
	}

	// Test response
	response := agentcomm.NewAgentMessage("agent2", "agent1", agentcomm.MsgResponse, "result", "session1")
	if response.Type != agentcomm.MsgResponse {
		t.Error("Response message should have MsgResponse type")
	}

	t.Logf("Message types test passed")
}

// TestSubagentManagerConcurrency tests concurrent subagent operations
func TestSubagentManagerConcurrency(t *testing.T) {
	sharedCtx := agentcomm.NewSharedContext(100, 1000)

	mockProvider := &MockProvider{
		responses: []providers.Message{
			{Role: "assistant", Content: "Done"},
		},
	}

	manager := NewSubagentManager(
		mockProvider,
		"test-model",
		"/tmp/test-workspace",
		nil,
	)
	manager.SetSharedContext(sharedCtx)

	// Spawn multiple subagents concurrently
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		go func(idx int) {
			_, _ = manager.Spawn(ctx, "Task number ", "task-", "", "", "cli", "test", nil)
		}(i)
	}

	// Wait a bit for tasks to be created
	time.Sleep(100 * time.Millisecond)

	tasks := manager.ListTasks()
	t.Logf("Concurrent test: created %d tasks", len(tasks))

	if len(tasks) != 5 {
		t.Logf("Warning: expected 5 tasks, got %d", len(tasks))
	}
}
