// Comprehensive Subagent Spawning Test
// ทดสอบครอบคลุมหลายบริบทเพื่อให้ได้ความแม่นยำ 95%+

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

	fmt.Println("🧪 Comprehensive Subagent Spawning Test")
	fmt.Println("========================================")
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
	fmt.Println("\n🧪 Running Comprehensive Tests...")

	testCategories := []struct {
		category string
		tests    []testCase
	}{
		{
			category: "🗣️ Pure Conversation (should NOT spawn)",
			tests: []testCase{
				{name: "Greeting", message: "สวัสดี", shouldSpawn: false},
				{name: "How are you", message: "สบายดีไหม", shouldSpawn: false},
				{name: "Thanks", message: "ขอบคุณมาก", shouldSpawn: false},
				{name: "Goodbye", message: "ลาก่อน", shouldSpawn: false},
				{name: "Tell joke", message: "เล่าเรื่องตลกให้ฟังหน่อย", shouldSpawn: false},
				{name: "Weather", message: "วันนี้อากาศดีไหม", shouldSpawn: false},
				{name: "Personal question", message: "คุณชื่ออะไร", shouldSpawn: false},
				{name: "Opinion", message: "คิดว่าไงกับ AI", shouldSpawn: false},
				{name: "Casual chat", message: "ว่างไหมคุยกันหน่อย", shouldSpawn: false},
				{name: "Compliment", message: "เก่งมากเลย", shouldSpawn: false},
			},
		},
		{
			category: "💭 Questions About Code (should NOT spawn - just asking)",
			tests: []testCase{
				{name: "Compare languages", message: "Python กับ Go อันไหนดีกว่า", shouldSpawn: false},
				{name: "Ask syntax", message: "Go มี generics ไหม", shouldSpawn: false},
				{name: "Best practice", message: "เขียน Go ยังไงให้ดี", shouldSpawn: false},
				{name: "Explain concept", message: "อธิบาย goroutine หน่อย", shouldSpawn: false},
				{name: "Learning advice", message: "ควรเรียน Python ยังไง", shouldSpawn: false},
				{name: "Framework choice", message: "ใช้ Django หรือ FastAPI ดี", shouldSpawn: false},
				{name: "Career advice", message: "ควรเป็น frontend หรือ backend", shouldSpawn: false},
				{name: "Tool recommendation", message: "แนะนำ IDE สำหรับ Go", shouldSpawn: false},
			},
		},
		{
			category: "💻 Coding Tasks (should spawn as coder)",
			tests: []testCase{
				{name: "Write function", message: "เขียนฟังก์ชั่น bubble sort ให้หน่อย", shouldSpawn: true, expectedRole: "coder"},
				{name: "Implement API", message: "สร้าง REST API ด้วย Gin", shouldSpawn: true, expectedRole: "coder"},
				{name: "Create script", message: "เขียน Python script อ่าน CSV", shouldSpawn: true, expectedRole: "coder"},
				{name: "Refactor code", message: "ช่วย refactor โค้ดนี้ให้สะอาดขึ้น", shouldSpawn: true, expectedRole: "coder"},
				{name: "Add feature", message: "เพิ่ม feature login ให้โปรเจคนี้", shouldSpawn: true, expectedRole: "coder"},
				{name: "Write test", message: "เขียน unit test สำหรับฟังก์ชั่นนี้", shouldSpawn: true, expectedRole: "coder"},
				{name: "Code review", message: "review โค้ดนี้ให้หน่อย", shouldSpawn: true, expectedRole: "coder"},
				{name: "Fix style", message: "แก้ให้ตาม PEP8 หน่อย", shouldSpawn: true, expectedRole: "coder"},
			},
		},
		{
			category: "🐛 Debugging Tasks (should spawn as debugger)",
			tests: []testCase{
				{name: "Find bug", message: "โค้ดนี้มี bug ช่วยหาหน่อย", shouldSpawn: true, expectedRole: "debugger"},
				{name: "Fix error", message: "มัน error ตรงนี้ แก้ยังไง", shouldSpawn: true, expectedRole: "debugger"},
				{name: "Crash issue", message: "โปรแกรม crash ตอนรัน", shouldSpawn: true, expectedRole: "debugger"},
				{name: "Performance issue", message: "ทำไมช้าจัง ช่วย optimize หน่อย", shouldSpawn: true, expectedRole: "debugger"},
				{name: "Memory leak", message: "memory leak ช่วยหาหน่อย", shouldSpawn: true, expectedRole: "debugger"},
				{name: "Race condition", message: "มี race condition ช่วย debug หน่อย", shouldSpawn: true, expectedRole: "debugger"},
			},
		},
		{
			category: "🏗️ Architecture Tasks (should spawn as architect)",
			tests: []testCase{
				{name: "System design", message: "ออกแบบระบบ e-commerce", shouldSpawn: true, expectedRole: "architect"},
				{name: "Database design", message: "design database สำหรับระบบจองห้อง", shouldSpawn: true, expectedRole: "architect"},
				{name: "API design", message: "ออกแบบ API สำหรับ mobile app", shouldSpawn: true, expectedRole: "architect"},
				{name: "Microservices", message: "แบ่ง microservices ยังไงดี", shouldSpawn: true, expectedRole: "architect"},
				{name: "Scalability", message: "ทำยังไงให้รองรับ 1M users", shouldSpawn: true, expectedRole: "architect"},
				{name: "Security design", message: "ออกแบบระบบ auth ยังไงดี", shouldSpawn: true, expectedRole: "architect"},
			},
		},
		{
			category: "📝 Documentation Tasks (should spawn as writer)",
			tests: []testCase{
				{name: "Write README", message: "เขียน README ให้โปรเจคนี้หน่อย", shouldSpawn: true, expectedRole: "writer"},
				{name: "API docs", message: "เขียน API documentation", shouldSpawn: true, expectedRole: "writer"},
				{name: "Tutorial", message: "เขียนวิธีใช้งานสำหรับ beginner", shouldSpawn: true, expectedRole: "writer"},
				{name: "Code comments", message: "เพิ่ม comment อธิบายโค้ดนี้", shouldSpawn: true, expectedRole: "writer"},
			},
		},
		{
			category: "🔍 Research Tasks (should spawn as researcher)",
			tests: []testCase{
				{name: "Find library", message: "หา library สำหรับทำ PDF", shouldSpawn: true, expectedRole: "researcher"},
				{name: "Compare tools", message: "เปรียบเทียบ Docker vs Podman", shouldSpawn: true, expectedRole: "researcher"},
				{name: "Best practices", message: "หา best practices สำหรับ Kubernetes", shouldSpawn: true, expectedRole: "researcher"},
				{name: "Latest trends", message: "AI 2026 มีอะไรใหม่", shouldSpawn: true, expectedRole: "researcher"},
			},
		},
		{
			category: "⚠️ Edge Cases - Mixed Intent",
			tests: []testCase{
				{name: "Greet + request", message: "สวัสดี ช่วยเขียนโค้ดหน่อย", shouldSpawn: true, expectedRole: "coder"},
				{name: "Thanks + request", message: "ขอบคุณครับ ช่วย debug หน่อย", shouldSpawn: true, expectedRole: "debugger"},
				{name: "Question + task", message: "Go มี generics ไหม ช่วยเขียนตัวอย่างหน่อย", shouldSpawn: true, expectedRole: "coder"},
				{name: "Follow up", message: "จากเมื่อกี้ ช่วยเพิ่ม feature นี้หน่อย", shouldSpawn: true, expectedRole: "coder"},
				{name: "Context aware", message: "ใช้ pattern เดิมที่คุยกันเมื่อกี้", shouldSpawn: true, expectedRole: "coder"},
			},
		},
	}

	totalTests := 0
	totalPassed := 0
	categoryResults := []categoryResult{}

	for _, cat := range testCategories {
		fmt.Printf("\n%s\n", cat.category)
		fmt.Println(strings.Repeat("-", 60))

		catPassed := 0
		catFailed := 0

		for _, test := range cat.tests {
			result := testSubagentSpawn(agentLoop, test.message)
			passed := checkResult(test, result)

			status := "✅"
			if !passed {
				status = "❌"
			}

			fmt.Printf("  %s %s: spawned=%v", status, test.name, result.spawned)
			if result.spawned {
				fmt.Println("Result:", result)
			}
			if !passed {
				fmt.Printf(" (expected spawn=%v", test.shouldSpawn)
				if test.expectedRole != "" {
					fmt.Printf(", role=%s", test.expectedRole)
				}
				fmt.Printf(")")
			}
			fmt.Println()

			if passed {
				catPassed++
			} else {
				catFailed++
			}

			totalTests++
			if passed {
				totalPassed++
			}

			time.Sleep(200 * time.Millisecond)
		}

		categoryResults = append(categoryResults, categoryResult{
			name:   cat.category,
			passed: catPassed,
			failed: catFailed,
		})

		fmt.Printf("  → %d/%d passed\n", catPassed, catPassed+catFailed)
	}

	// Print summary
	percentage := float64(totalPassed) / float64(totalTests) * 100

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📊 Final Test Summary")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total Tests: %d\n", totalTests)
	fmt.Printf("Passed: %d ✅\n", totalPassed)
	fmt.Printf("Failed: %d ❌\n", totalTests-totalPassed)
	fmt.Printf("Success Rate: %.1f%%\n", percentage)
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("\n📈 Category Breakdown:")
	for _, cr := range categoryResults {
		total := cr.passed + cr.failed
		pct := float64(cr.passed) / float64(total) * 100
		fmt.Printf("  %s: %d/%d (%.0f%%)\n", cr.name, cr.passed, total, pct)
	}

	if percentage >= 95 {
		fmt.Printf("\n🎉 Excellent! Achieved %.1f%% (target: 95%%+)\n", percentage)
	} else if percentage >= 90 {
		fmt.Printf("\n⚠️ Good progress: %.1f%% (need 95%%+)\n", percentage)
	} else {
		fmt.Printf("\n❌ Needs improvement: %.1f%% (target: 95%%+)\n", percentage)
	}
}

type testCase struct {
	name          string
	message       string
	shouldSpawn   bool
	expectedRole  string
}

type spawnResult struct {
	spawned bool
	role    string
}

type categoryResult struct {
	name   string
	passed int
	failed int
}

func testSubagentSpawn(agentLoop *agent.AgentLoop, message string) spawnResult {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := agentLoop.ProcessDirect(ctx, message, "test-session", nil)
	if err != nil {
		return spawnResult{spawned: false}
	}

	result := spawnResult{spawned: false}

	if strings.Contains(response, "Spawned subagent") || strings.Contains(response, "spawn_subagent") {
		result.spawned = true

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

func checkResult(test testCase, result spawnResult) bool {
	if test.shouldSpawn != result.spawned {
		return false
	}

	if test.shouldSpawn && test.expectedRole != "" {
		if result.role != test.expectedRole {
			return false
		}
	}

	return true
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
