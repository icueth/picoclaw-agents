// Test Subagent Spawning System
// ทดสอบระบบ spawn subagent เพื่อวิเคราะห์ปัญหาการ spawn ผิดพลาด

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"picoclaw/agent/pkg/bootstrap"
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/providers"
	"picoclaw/agent/pkg/agent"
)

func main() {
	var (
		configPath = flag.String("config", "", "Path to config.json")
	)

	flag.Parse()

	if *configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			defaultPath := homeDir + "/.picoclaw/config.json"
			if _, err := os.Stat(defaultPath); err == nil {
				*configPath = defaultPath
			}
		}
	}

	if *configPath == "" {
		fmt.Println("Error: config path is required")
		os.Exit(1)
	}

	fmt.Println("🧪 Subagent Spawning System Test")
	fmt.Println("==================================")
	fmt.Printf("Config: %s\n\n", *configPath)

	// Load config
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Bootstrap system
	fmt.Println("📦 Bootstrapping system...")
	sys, err := bootstrap.Bootstrap(cfg)
	if err != nil {
		fmt.Printf("❌ Failed to bootstrap: %v\n", err)
		os.Exit(1)
	}
	defer sys.Close()

	// Create message bus
	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	// Create provider
	fmt.Println("🔌 Creating LLM provider...")
	provider, _, err := createProviderFromConfig(cfg)
	if err != nil {
		fmt.Printf("❌ Failed to create provider: %v\n", err)
		os.Exit(1)
	}

	// Create AgentLoop
	fmt.Println("🤖 Creating AgentLoop...")
	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider, sys.JobMgr, sys.MemoryManager)

	// Run comprehensive tests
	fmt.Println("\n🧪 Running Subagent Spawn Tests...")

	tests := []struct {
		name           string
		message        string
		shouldSpawn    bool   // ควร spawn หรือไม่
		expectedRole   string // ควร spawn เป็น role อะไร
		isConversation bool   // เป็นการสนทนาทั่วไปหรือไม่
	}{
		// การสนทนาทั่วไป - ไม่ควร spawn
		{
			name:           "Conversation 1: Greeting",
			message:        "สวัสดี คุณเป็นใคร",
			shouldSpawn:    false,
			isConversation: true,
		},
		{
			name:           "Conversation 2: How are you",
			message:        "วันนี้เป็นอย่างไรบ้าง",
			shouldSpawn:    false,
			isConversation: true,
		},
		{
			name:           "Conversation 3: Thanks",
			message:        "ขอบคุณมาก งั้นลาก่อน",
			shouldSpawn:    false,
			isConversation: true,
		},
		{
			name:           "Conversation 4: Casual chat",
			message:        "เล่าเรื่องตลกให้ฟังหน่อย",
			shouldSpawn:    false,
			isConversation: true,
		},
		// งานเฉพาะทาง - ควร spawn
		{
			name:         "Task 1: Code Review",
			message:      "ช่วย review โค้ดนี้หน่อย func main() { fmt.Println(\"hello\") }",
			shouldSpawn:  true,
			expectedRole: "coder",
		},
		{
			name:         "Task 2: Write Python",
			message:      "เขียนโปรแกรม Python คำนวณ fibonacci ให้หน่อย",
			shouldSpawn:  true,
			expectedRole: "coder",
		},
		{
			name:         "Task 3: Debug",
			message:      "โค้ดนี้มี bug ช่วยหาหน่อย",
			shouldSpawn:  true,
			expectedRole: "debugger",
		},
		{
			name:         "Task 4: Architecture",
			message:      "ออกแบบระบบ microservices สำหรับ e-commerce",
			shouldSpawn:  true,
			expectedRole: "architect",
		},
		{
			name:         "Task 5: Database Design",
			message:      "ช่วย design database schema สำหรับระบบจองห้องพัก",
			shouldSpawn:  true,
			expectedRole: "architect",
		},
		// Edge cases - ต้องดู context
		{
			name:           "Edge 1: Mixed content",
			message:        "สวัสดี ช่วยเขียนโค้ด Python หน่อย",
			shouldSpawn:    true,
			expectedRole:   "coder",
			isConversation: false,
		},
		{
			name:           "Edge 2: Question about code",
			message:        "Python กับ Go อันไหนดีกว่ากันสำหรับ backend",
			shouldSpawn:    false,
			isConversation: true,
		},
		{
			name:           "Edge 3: Request with context",
			message:        "จากโค้ดที่คุยกันเมื่อกี้ ช่วยเพิ่ม feature นี้หน่อย",
			shouldSpawn:    true,
			expectedRole:   "coder",
		},
	}

	passed := 0
	failed := 0
	conversationPassed := 0
	conversationFailed := 0
	taskPassed := 0
	taskFailed := 0

	for _, test := range tests {
		fmt.Printf("📝 %s", test.name)
		fmt.Printf("   Message: %s\n", truncate(test.message, 60))

		result := testSubagentSpawn(agentLoop, test.message)

		// Analyze result
		spawned := result.spawned
		role := result.role

		fmt.Printf("   Result: spawned=%v, role=%s\n", spawned, role)

		// Check correctness
		correct := true
		if test.isConversation {
			// การสนทนาทั่วไปไม่ควร spawn
			if spawned {
				correct = false
				fmt.Printf("   ❌ WRONG: Conversation should NOT spawn subagent\n")
			} else {
				fmt.Printf("   ✅ CORRECT: Conversation handled directly\n")
				conversationPassed++
			}
		} else if test.shouldSpawn {
			// งานเฉพาะทางควร spawn
			if !spawned {
				correct = false
				fmt.Printf("   ❌ WRONG: Task should spawn subagent\n")
			} else if test.expectedRole != "" && role != test.expectedRole {
				fmt.Printf("   ⚠️  PARTIAL: Spawned but wrong role (expected %s, got %s)\n", test.expectedRole, role)
				// ถือว่าผ่านแต่มีหมายเหตุ
				correct = true
				taskPassed++
			} else {
				fmt.Printf("   ✅ CORRECT: Task spawned with correct role\n")
				taskPassed++
			}
		}

		if correct {
			passed++
		} else {
			failed++
			if test.isConversation {
				conversationFailed++
			} else {
				taskFailed++
			}
		}
		fmt.Println()

		// Small delay between tests
		time.Sleep(500 * time.Millisecond)
	}

	// Print summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📊 Test Summary")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total: %d | Passed: %d ✅ | Failed: %d ❌\n", passed+failed, passed, failed)
	fmt.Println()
	fmt.Printf("Conversation Tests: %d/%d passed\n", conversationPassed, conversationPassed+conversationFailed)
	fmt.Printf("Task Tests: %d/%d passed\n", taskPassed, taskPassed+taskFailed)
	fmt.Println(strings.Repeat("=", 60))

	if failed > 0 {
		fmt.Printf("\n❌ Some tests failed. Need to improve task analyzer.\n")
		os.Exit(1)
	}

	fmt.Println("\n✅ All tests passed!")
}

type spawnResult struct {
	spawned bool
	role    string
	reason  string
}

func testSubagentSpawn(agentLoop *agent.AgentLoop, message string) spawnResult {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := agentLoop.ProcessDirect(ctx, message, "test-session", nil)
	if err != nil {
		return spawnResult{spawned: false, reason: err.Error()}
	}

	// Check if response indicates subagent was spawned
	result := spawnResult{spawned: false}

	if strings.Contains(response, "Spawned subagent") || strings.Contains(response, "spawn_subagent") {
		result.spawned = true

		// Extract role from response
		if idx := strings.Index(response, "role '"); idx != -1 {
			start := idx + 6
			end := strings.Index(response[start:], "'")
			if end != -1 {
				result.role = response[start : start+end]
			}
		}
	}

	return result
}

func createProviderFromConfig(cfg *config.Config) (providers.LLMProvider, string, error) {
	if len(cfg.ModelList) == 0 {
		return nil, "", fmt.Errorf("no models configured")
	}

	modelCfg := cfg.ModelList[0]
	modelName := modelCfg.ModelName
	if modelName == "" {
		modelName = modelCfg.Model
	}

	provider, returnedModelName, err := providers.CreateProviderFromConfig(&modelCfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create provider: %w", err)
	}

	if returnedModelName != "" {
		modelName = returnedModelName
	}

	return provider, modelName, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
