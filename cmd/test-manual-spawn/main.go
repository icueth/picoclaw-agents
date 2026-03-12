// Test Manual Spawn and Progress Tracking
// ทดสอบการ spawn ด้วยตนเองและติดตามความคืบหน้า

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
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config.json")
	flag.Parse()

	if configPath == "" {
		homeDir, _ := os.UserHomeDir()
		configPath = homeDir + "/.picoclaw/config.json"
	}

	fmt.Println("🧪 Manual Spawn & Progress Tracking Test")
	fmt.Println("=========================================")

	cfg, _ := config.LoadConfig(configPath)
	sys, _ := bootstrap.Bootstrap(cfg)
	defer sys.Close()

	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	provider, _, _ := createProviderFromConfig(cfg)
	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider, sys.JobMgr, sys.MemoryManager)

	fmt.Println("\n📋 Test Cases:")
	fmt.Println("1. Auto spawn should be DISABLED")
	fmt.Println("2. Manual spawn via tool should work")
	fmt.Println("3. Progress tracking should be visible")
	fmt.Println("4. Subagent result should return to parent")

	// Test 1: Verify auto spawn is disabled
	fmt.Println("\n🧪 Test 1: Verify auto spawn is disabled")
	testAutoSpawnDisabled(agentLoop)

	// Test 2: Manual spawn coder
	fmt.Println("\n🧪 Test 2: Manual spawn coder")
	testManualSpawn(agentLoop, "coder", "เขียน Python function คำนวณ factorial")

	// Test 3: Manual spawn debugger
	fmt.Println("\n🧪 Test 3: Manual spawn debugger")
	testManualSpawn(agentLoop, "debugger", "หา bug ในโค้ดนี้")

	// Test 4: Check progress tracking
	fmt.Println("\n🧪 Test 4: Progress tracking")
	testProgressTracking(agentLoop)

	fmt.Println("\n✅ All tests completed!")
}

func testAutoSpawnDisabled(al *agent.AgentLoop) {
	// Tasks that would previously trigger auto spawn
	tasks := []string{
		"เขียนโค้ด Python หน่อย",
		"ออกแบบระบบ e-commerce",
		"debug โค้ดนี้หน่อย",
	}

	for _, task := range tasks {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		resp, _ := al.ProcessDirect(ctx, task, "test", nil)
		cancel()

		if strings.Contains(resp, "Spawned subagent") {
			fmt.Printf("  ❌ FAIL: '%s' auto spawned (should not)\n", task)
		} else {
			fmt.Printf("  ✅ PASS: '%s' handled directly\n", task)
		}
	}
}

func testManualSpawn(al *agent.AgentLoop, role, task string) {
	// User explicitly requests spawn
	msg := fmt.Sprintf("spawn %s: %s", role, task)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	resp, err := al.ProcessDirect(ctx, msg, "test", nil)
	cancel()

	if err != nil {
		fmt.Printf("  ❌ ERROR: %v\n", err)
		return
	}

	if strings.Contains(resp, "Spawned subagent") || strings.Contains(resp, role) {
		fmt.Printf("  ✅ PASS: Manual spawn %s successful\n", role)
		fmt.Printf("     Response: %s\n", truncate(resp, 100))
	} else {
		fmt.Printf("  ⚠️  Response: %s\n", truncate(resp, 100))
	}
}

func testProgressTracking(al *agent.AgentLoop) {
	// Check if we can query subagent status
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	resp, _ := al.ProcessDirect(ctx, "check subagent status", "test", nil)
	cancel()

	fmt.Printf("  Status query response: %s\n", truncate(resp, 80))
}

func createProviderFromConfig(cfg *config.Config) (providers.LLMProvider, string, error) {
	if len(cfg.ModelList) == 0 {
		return nil, "", fmt.Errorf("no models")
	}
	mc := cfg.ModelList[0]
	p, m, e := providers.CreateProviderFromConfig(&mc)
	return p, m, e
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
