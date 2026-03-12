// Quick Subagent Test - 20 test cases for rapid iteration
// ทดสอบเร็วเพื่อปรับปรุงให้ได้ 95%+

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

	fmt.Println("⚡ Quick Subagent Test (20 cases)")
	fmt.Println("==================================")

	cfg, _ := config.LoadConfig(configPath)
	sys, _ := bootstrap.Bootstrap(cfg)
	defer sys.Close()

	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	provider, _, _ := createProviderFromConfig(cfg)
	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider, sys.JobMgr, sys.MemoryManager)

	tests := []struct {
		name         string
		message      string
		shouldSpawn  bool
		expectedRole string
	}{
		// Conversation (NO spawn) - 5 tests
		{"Greeting", "สวัสดี", false, ""},
		{"Thanks", "ขอบคุณ", false, ""},
		{"How are you", "สบายดีไหม", false, ""},
		{"Tell joke", "เล่าเรื่องตลก", false, ""},
		{"Ask about Go", "Go มี generics ไหม", false, ""},

		// Coding (spawn coder) - 5 tests
		{"Write code", "เขียน bubble sort", true, "coder"},
		{"Code review", "review โค้ดนี้", true, "coder"},
		{"Refactor", "refactor ให้สะอาดขึ้น", true, "coder"},
		{"Add feature", "เพิ่ม feature login", true, "coder"},
		{"Write test", "เขียน unit test", true, "coder"},

		// Debugging (spawn debugger) - 3 tests
		{"Find bug", "มี bug ช่วยหาหน่อย", true, "debugger"},
		{"Fix error", "error ตรงนี้", true, "debugger"},
		{"Crash", "โปรแกรม crash", true, "debugger"},

		// Architecture (spawn architect) - 4 tests
		{"System design", "ออกแบบระบบ e-commerce", true, "architect"},
		{"Database design", "design database", true, "architect"},
		{"API design", "ออกแบบ API", true, "architect"},
		{"Microservices", "แบ่ง microservices", true, "architect"},

		// Edge cases - 3 tests
		{"Greet + code", "สวัสดี เขียนโค้ดหน่อย", true, "coder"},
		{"Mixed intent", "Go ดีไหม ช่วยเขียนตัวอย่าง", true, "coder"},
		{"Follow up", "จากเมื่อกี้ เพิ่มอีก", true, "coder"},
	}

	passed := 0
	for i, test := range tests {
		result := testSpawn(agentLoop, test.message)
		ok := (test.shouldSpawn == result.spawned) &&
			(!test.shouldSpawn || test.expectedRole == result.role)

		status := "✅"
		if !ok {
			status = "❌"
		} else {
			passed++
		}

		fmt.Printf("%s %2d. %s: spawn=%v", status, i+1, test.name, result.spawned)
		if result.spawned {
			fmt.Printf("(%s)", result.role)
		}
		if !ok {
			fmt.Printf(" [want=%v", test.shouldSpawn)
			if test.expectedRole != "" {
				fmt.Printf(",%s", test.expectedRole)
			}
			fmt.Printf("]")
		}
		fmt.Println()
		time.Sleep(100 * time.Millisecond)
	}

	pct := float64(passed) / float64(len(tests)) * 100
	fmt.Printf("\nResult: %d/%d (%.0f%%)\n", passed, len(tests), pct)
	if pct >= 95 {
		fmt.Println("🎉 Target achieved!")
	} else {
		fmt.Println("⚠️  Need improvement")
	}
}

type result struct {
	spawned bool
	role    string
}

func testSpawn(al *agent.AgentLoop, msg string) result {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	resp, _ := al.ProcessDirect(ctx, msg, "test", nil)
	r := result{}
	if strings.Contains(resp, "Spawned subagent") {
		r.spawned = true
		if i := strings.Index(resp, "role '"); i != -1 {
			start := i + 6
			if end := strings.Index(resp[start:], "'"); end != -1 {
				r.role = resp[start : start+end]
			}
		}
	}
	return r
}

func createProviderFromConfig(cfg *config.Config) (providers.LLMProvider, string, error) {
	if len(cfg.ModelList) == 0 {
		return nil, "", fmt.Errorf("no models")
	}
	mc := cfg.ModelList[0]
	p, m, e := providers.CreateProviderFromConfig(&mc)
	return p, m, e
}
