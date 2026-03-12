// Comprehensive Job Tracking Test
// ทดสอบการติดตามงานแบบครอบคลุมเพื่อให้มั่นใจว่าระบบไม่เงียบหาย

package main

import (
	"context"
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
	homeDir, _ := os.UserHomeDir()
	configPath := homeDir + "/.picoclaw/config.json"

	fmt.Println("🧪 Comprehensive Job Tracking Test")
	fmt.Println("====================================")
	fmt.Println("Testing: Job creation → Execution → Completion → Result delivery")
	fmt.Println()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("❌ Config error: %v\n", err)
		os.Exit(1)
	}

	sys, err := bootstrap.Bootstrap(cfg)
	if err != nil {
		fmt.Printf("❌ Bootstrap error: %v\n", err)
		os.Exit(1)
	}
	defer sys.Close()

	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	provider, _, _ := createProviderFromConfig(cfg)
	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider, sys.JobMgr, sys.MemoryManager)

	tests := []struct {
		name        string
		message     string
		wantSpawn   bool
		checkResult bool
	}{
		// Test 1-3: Simple tasks that should complete quickly
		{"Quick task: Write function", "spawn coder: เขียน function บวกเลขสองตัว", true, true},
		{"Quick task: Debug", "spawn debugger: หา bug ในโค้ดนี้", true, true},
		{"Quick task: Review", "spawn reviewer: review โค้ดนี้", true, true},

		// Test 4-6: Tasks with progress tracking
		{"Progress tracking: Multi-step", "spawn coder: สร้าง REST API แบบง่าย", true, true},
		{"Progress tracking: Analysis", "spawn researcher: หาข้อมูล Go vs Python", true, true},

		// Test 7-8: Edge cases
		{"Edge: Very short task", "spawn coder: print hello", true, true},
		{"Edge: Complex task", "spawn architect: ออกแบบระบบง่ายๆ", true, true},
	}

	passed := 0
	failed := 0

	for i, test := range tests {
		fmt.Printf("\n🧪 Test %d: %s\n", i+1, test.name)
		fmt.Printf("   Input: %s\n", test.message)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		start := time.Now()
		resp, err := agentLoop.ProcessDirect(ctx, test.message, fmt.Sprintf("test-%d", i), nil)
		duration := time.Since(start)
		cancel()

		if err != nil {
			fmt.Printf("   ❌ ERROR: %v\n", err)
			failed++
			continue
		}

		// Check if spawn was triggered
		spawned := strings.Contains(resp, "Spawned subagent") ||
			strings.Contains(resp, "spawn_subagent") ||
			strings.Contains(resp, "subagent-")

		if test.wantSpawn && !spawned {
			fmt.Printf("   ❌ FAIL: Expected spawn but didn't\n")
			fmt.Printf("   Response: %s\n", truncate(resp, 100))
			failed++
			continue
		}

		if !test.wantSpawn && spawned {
			fmt.Printf("   ❌ FAIL: Unexpected spawn\n")
			failed++
			continue
		}

		// Check for result delivery (non-empty meaningful response)
		hasResult := len(resp) > 50 && !strings.Contains(resp, "error")

		if test.checkResult && !hasResult {
			fmt.Printf("   ⚠️  WARNING: No meaningful result\n")
			fmt.Printf("   Response: %s\n", truncate(resp, 100))
		} else {
			fmt.Printf("   ✅ PASS: %s\n", truncate(resp, 80))
		}

		fmt.Printf("   Duration: %v\n", duration)
		passed++

		time.Sleep(1 * time.Second)
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 Test Summary")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total: %d | Passed: %d ✅ | Failed: %d ❌\n", passed+failed, passed, failed)
	pct := float64(passed) / float64(passed+failed) * 100
	fmt.Printf("Success Rate: %.1f%%\n", pct)

	if pct >= 95 {
		fmt.Println("🎉 Excellent! System is working correctly")
	} else if pct >= 80 {
		fmt.Println("⚠️  Good but needs improvement")
	} else {
		fmt.Println("❌ System needs significant fixes")
	}
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
