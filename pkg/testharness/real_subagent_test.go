// Real Subagent Integration Test
// ทดสอบ Subagent System กับ LLM จริง

package testharness

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/providers"
	"picoclaw/agent/pkg/tools"
)

// TestRealSubagentSpawn ทดสอบการ spawn subagent ด้วย LLM จริง
func TestRealSubagentSpawn(t *testing.T) {
	// ตรวจสอบว่ามี config จริง
	configPath := os.ExpandEnv("${HOME}/.picoclaw/config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Config file not found, skipping real LLM test")
	}

	// โหลด config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// สร้าง provider จริง
	provider, modelName, err := createRealProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	fmt.Printf("Using model: %s\n", modelName)

	// สร้าง subagent manager
	workspace := "/tmp/test-subagent"
	os.MkdirAll(workspace, 0755)

	msgBus := bus.NewMessageBus()
	manager := tools.NewSubagentManager(provider, modelName, workspace, msgBus)
	manager.SetLLMOptions(4096, 0.7)

	// สร้าง spawn tool
	spawnTool := tools.NewSpawnTool(manager)

	// ทดสอบ spawn subagent ด้วยงานง่ายๆ
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result := spawnTool.Execute(ctx, map[string]any{
		"task":  "Write a simple hello world program in Python",
		"label": "test-hello-world",
	})

	fmt.Printf("Spawn result: ForLLM=%s, Async=%v, IsError=%v\n", 
		result.ForLLM, result.Async, result.IsError)

	if result.IsError {
		t.Fatalf("Failed to spawn subagent: %s", result.ForLLM)
	}

	// รอให้ subagent ทำงานเสร็จ (สูงสุด 60 วินาที)
	// เนื่องจาก Spawn คืนค่าเป็น message ไม่ใช่ task ID โดยตรง
	// เราจะใช้ ListTasks เพื่อดึง task ล่าสุดแทน
	time.Sleep(500 * time.Millisecond) // รอให้ task ถูกสร้าง
	
	tasks := manager.ListTasks()
	if len(tasks) == 0 {
		t.Fatal("No tasks found after spawn")
	}
	
	// เอา task ล่าสุด (task ที่เพิ่งสร้าง)
	task := tasks[len(tasks)-1]
	taskID := task.ID
	
	fmt.Printf("Waiting for task %s to complete...\n", taskID)

	// รอและตรวจสอบสถานะ
	var finalStatus string
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		
		task, ok := manager.GetTask(taskID)
		if !ok {
			t.Fatalf("Task %s not found", taskID)
		}

		fmt.Printf("Status: %s (progress: %d%%)\n", task.Status, task.ProgressPercent)

		if task.Status == "completed" || task.Status == "failed" {
			finalStatus = task.Status
			fmt.Printf("Task result: %s\n", task.Result)
			if task.Error != "" {
				fmt.Printf("Task error: %s\n", task.Error)
			}
			break
		}
	}

	if finalStatus != "completed" {
		t.Fatalf("Task did not complete successfully, final status: %s", finalStatus)
	}

	fmt.Println("✅ Subagent test passed!")
}

// TestRealSubagentWithRole ทดสอบการ spawn subagent ตาม role
func TestRealSubagentWithRole(t *testing.T) {
	configPath := os.ExpandEnv("${HOME}/.picoclaw/config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Config file not found, skipping real LLM test")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	provider, modelName, err := createRealProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	workspace := "/tmp/test-subagent-role"
	os.MkdirAll(workspace, 0755)

	msgBus := bus.NewMessageBus()
	manager := tools.NewSubagentManager(provider, modelName, workspace, msgBus)
	
	// ตั้งค่า role config
	roleConfig := map[string]config.SubagentRoleConfig{
		"coder": {
			Model:             "",
			Description:       "Expert programmer",
			SystemPromptAddon: "You are an expert programmer. Write clean, efficient code.",
			MaxIterations:     30,
			TimeoutSeconds:    180,
			Extendable:        true,
			MaxExtensions:     2,
		},
	}
	manager.SetRoleConfig(roleConfig)

	// ทดสอบ spawn ด้วย role
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	task := "Create a simple function to calculate fibonacci numbers"
	result, err := manager.SpawnWithRole(ctx, 
		"coder", 
		task, 
		nil, // contextData
		"",  // conceptID
		180, // timeout
		"test", "test-chat", 
		nil, // callback
	)
	
	if err != nil {
		t.Fatalf("Failed to spawn subagent with role: %v", err)
	}

	fmt.Printf("Spawned task: %s\n", result)

	// รอผลลัพธ์ - แยก task ID จากข้อความ
	// Format: "Spawned subagent with role 'coder' for task: ... (task_id: subagent-1)"
	taskID := extractTaskIDFromRoleResult(result)
	if taskID == "" {
		t.Fatal("Could not extract task ID from result")
	}
	fmt.Printf("Extracted task ID: %s\n", taskID)
	
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		
		task, ok := manager.GetTask(taskID)
		if !ok {
			t.Fatalf("Task %s not found", taskID)
		}

		fmt.Printf("Status: %s (progress: %d%%) - %s\n", 
			task.Status, task.ProgressPercent, task.ProgressMessage)

		if task.Status == "completed" {
			fmt.Printf("✅ Task completed!\nResult: %s\n", task.Result)
			return
		}
		if task.Status == "failed" {
			t.Fatalf("Task failed: %s", task.Error)
		}
	}

	t.Fatal("Task timeout")
}

// TestRealSubagentListTasks ทดสอบการตรวจสอบรายการ tasks
func TestRealSubagentListTasks(t *testing.T) {
	configPath := os.ExpandEnv("${HOME}/.picoclaw/config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Config file not found, skipping real LLM test")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	provider, modelName, err := createRealProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	workspace := "/tmp/test-subagent-list"
	os.MkdirAll(workspace, 0755)

	msgBus := bus.NewMessageBus()
	manager := tools.NewSubagentManager(provider, modelName, workspace, msgBus)

	// Spawn หลาย tasks
	ctx := context.Background()
	
	for i := 0; i < 3; i++ {
		_, err := manager.Spawn(ctx, 
			fmt.Sprintf("Simple task %d", i+1),
			fmt.Sprintf("task-%d", i+1),
			"", "", "test", "test-chat", nil)
		if err != nil {
			t.Fatalf("Failed to spawn task %d: %v", i+1, err)
		}
	}

	// ตรวจสอบสถานะทั้งหมด
	time.Sleep(1 * time.Second)
	
	allTasks := manager.ListTasks()
	fmt.Printf("Total tasks: %d\n", len(allTasks))

	for _, task := range allTasks {
		fmt.Printf("- %s: %s (progress: %d%%)\n", task.ID, task.Status, task.ProgressPercent)
	}
}

// TestRealLLMWithAgentTools ทดสอบ LLM จริงพร้อม agent tools
func TestRealLLMWithAgentTools(t *testing.T) {
	configPath := os.ExpandEnv("${HOME}/.picoclaw/config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Config file not found, skipping real LLM test")
	}

	// สร้าง harness ด้วย LLM จริง
	harness, err := NewRealLLMTestHarness(configPath)
	if err != nil {
		t.Fatalf("Failed to create harness: %v", err)
	}

	harness.WithTimeout(60 * time.Second)

	// ทดสอบ chat ง่ายๆ
	response, err := harness.Chat("Hello, can you help me with a simple task?")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	fmt.Printf("Response: %s\n", response)

	// บันทึก conversation
	tmpDir := os.TempDir()
	savePath := filepath.Join(tmpDir, "test_conversation.json")
	harness.SaveConversation(savePath)
	fmt.Printf("Conversation saved to: %s\n", savePath)
}

// Helper functions

func createRealProvider(cfg *config.Config) (providers.LLMProvider, string, error) {
	// ใช้ default model จาก config
	modelName := cfg.Agents.Defaults.Model
	if modelName == "" {
		modelName = "kimi-coding"
	}

	modelCfg, err := cfg.GetModelConfig(modelName)
	if err != nil {
		return nil, "", fmt.Errorf("model %s not found: %w", modelName, err)
	}

	provider, resolvedModel, err := providers.CreateProviderFromConfig(modelCfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, resolvedModel, nil
}

func extractTaskID(result string) string {
	// แยก task ID จากผลลัพธ์
	parts := strings.Split(result, ":")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return strings.TrimSpace(result)
}

func extractTaskIDFromRoleResult(result string) string {
	// Format: "Spawned subagent with role 'coder' for task: ... (task_id: subagent-1)"
	// หา substring ที่อยู่ระหว่าง "(task_id: " และ ")"
	start := strings.Index(result, "(task_id: ")
	if start == -1 {
		return ""
	}
	start += len("(task_id: ")
	
	end := strings.Index(result[start:], ")")
	if end == -1 {
		return ""
	}
	
	return strings.TrimSpace(result[start : start+end])
}
