// Agent Team Test Runner
// รันการทดสอบ Agent Team ทั้งหมดและแสดงผลสรุป

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"picoclaw/agent/pkg/testharness"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║         PicoClaw - Agent Team Test Suite                  ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Run scenarios
	fmt.Println("📋 Running Agent Team Scenarios...")
	fmt.Println(strings.Repeat("─", 60))
	
	passed := 0
	failed := 0
	
	for i, scenario := range testharness.AgentTeamScenarios {
		provider := testharness.NewMockProvider()
		scenario.Setup(provider)
		harness := testharness.New(provider)
		
		start := time.Now()
		err := scenario.Test(harness)
		duration := time.Since(start)
		
		status := "✅ PASS"
		if err != nil {
			status = "❌ FAIL"
			failed++
		} else {
			passed++
		}
		
		fmt.Printf("%d. %s (%v)\n", i+1, scenario.Name, duration)
		fmt.Printf("   %s %s\n", status, scenario.Description)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
		}
		fmt.Println()
	}

	// Summary
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println("📊 Test Summary")
	fmt.Println(strings.Repeat("─", 60))
	total := passed + failed
	passRate := float64(passed) / float64(total) * 100
	
	fmt.Printf("Total:  %d\n", total)
	fmt.Printf("Passed: %d\n", passed)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Rate:   %.1f%%\n", passRate)
	fmt.Println(strings.Repeat("═", 60))

	// Capability demo
	fmt.Println()
	fmt.Println("🎯 Agent Team Capabilities")
	fmt.Println(strings.Repeat("─", 60))
	
	capabilities := []string{
		"✅ Spawn Subagent (spawn_subagent)",
		"✅ Check Subagent Status (subagent_status)",
		"✅ Start Meeting (start_meeting)",
		"✅ Send Message (send_agent_message)",
		"✅ Check Inbox (check_agent_inbox)",
	}
	
	for _, cap := range capabilities {
		fmt.Println(cap)
	}

	// Agent roles
	fmt.Println()
	fmt.Println("👥 Agent Roles")
	fmt.Println(strings.Repeat("─", 60))
	
	roles := map[string]string{
		"Jarvis":   "🤖 coordinator",
		"Nova":     "🔮 architect",
		"Atlas":    "📚 researcher",
		"Clawed":   "🔧 coder",
		"Sentinel": "🛡️  qa",
		"Scribe":   "📝 writer",
		"Trendy":   "🔍 reviewer",
		"Pixel":    "🎨 designer",
	}
	
	for name, role := range roles {
		fmt.Printf("  %s: %s\n", name, role)
	}

	// Exit code
	if failed > 0 {
		fmt.Println()
		fmt.Println("❌ Some tests failed")
		os.Exit(1)
	}
	
	fmt.Println()
	fmt.Println("✅ All tests passed!")
	os.Exit(0)
}
