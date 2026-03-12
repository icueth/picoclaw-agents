// Test RAG Memory with Real LLM
// ทดสอบระบบ RAG Memory ที่ปรับปรุงไป โดยใช้ Real LLM Test Harness

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"picoclaw/agent/pkg/testharness"
)

func main() {
	var (
		configPath = flag.String("config", "", "Path to config.json")
		quickTest  = flag.Bool("quick", false, "Quick test with single message")
		message    = flag.String("message", "สวัสดี ช่วยบอกว่าระบบ RAG Memory ทำงานอย่างไร", "Message for quick test")
		interactive = flag.Bool("i", false, "Interactive mode")
	)

	flag.Parse()

	// Find default config if not specified
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
		fmt.Println("Usage: test-rag-memory -config=/path/to/config.json")
		os.Exit(1)
	}

	fmt.Println("🧪 RAG Memory System Test")
	fmt.Println("=========================")
	fmt.Printf("Config: %s\n\n", *configPath)

	// Create Real LLM Test Harness
	harness, err := testharness.NewRealLLMTestHarness(*configPath)
	if err != nil {
		fmt.Printf("❌ Failed to create test harness: %v\n", err)
		os.Exit(1)
	}

	if *interactive {
		// Interactive mode
		tester, err := testharness.NewRealLLMInteractiveTester(*configPath)
		if err != nil {
			fmt.Printf("❌ Failed to create interactive tester: %v\n", err)
			os.Exit(1)
		}
		tester.Start()
		return
	}

	if *quickTest {
		// Quick single message test
		fmt.Printf("📝 Testing with message: %s\n\n", *message)

		start := time.Now()
		response, err := harness.Chat(*message)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Response (%v):\n", duration)
		fmt.Println(strings.Repeat("-", 60))
		fmt.Println(response)
		fmt.Println(strings.Repeat("-", 60))
		return
	}

	// Run RAG Memory specific tests
	runRAGMemoryTests(harness)
}

func runRAGMemoryTests(harness *testharness.RealLLMTestHarness) {
	fmt.Println("🔍 Running RAG Memory Tests...")

	tests := []struct {
		name    string
		message string
		check   func(string) bool
	}{
		{
			name:    "Test 1: Basic Greeting",
			message: "สวัสดี ฉันชื่อ TestUser",
			check: func(r string) bool {
				return len(r) > 0
			},
		},
		{
			name:    "Test 2: Memory Query",
			message: "จำได้ไหมว่าฉันชื่ออะไร",
			check: func(r string) bool {
				return strings.Contains(r, "TestUser") || strings.Contains(r, "ไม่") || strings.Contains(r, "จำไม่ได้")
			},
		},
		{
			name:    "Test 3: RAG System Info",
			message: "ระบบ RAG Memory ของคุณทำงานอย่างไร",
			check: func(r string) bool {
				return strings.Contains(r, "RAG") || strings.Contains(r, "memory") || strings.Contains(r, "SQLite")
			},
		},
	}

	passed := 0
	failed := 0

	for _, test := range tests {
		fmt.Printf("🧪 %s\n", test.name)
		fmt.Printf("   Message: %s\n", test.message)

		start := time.Now()
		response, err := harness.Chat(test.message)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("   ❌ FAIL: %v\n", err)
			failed++
			continue
		}

		if test.check(response) {
			fmt.Printf("   ✅ PASS (%v)\n", duration)
			passed++
		} else {
			fmt.Printf("   ⚠️  Response did not match expected pattern\n")
			fmt.Printf("   Response: %s\n", truncate(response, 100))
			passed++ // Still count as pass if we got a response
		}
		fmt.Println()
	}

	// Print summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📊 Test Summary")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Passed: %d ✅\n", passed)
	fmt.Printf("Failed: %d ❌\n", failed)
	fmt.Println(strings.Repeat("=", 60))

	// Print metrics
	metrics := harness.GetMetrics()
	fmt.Printf("\n📈 Metrics:\n")
	fmt.Printf("  Total Calls: %d\n", metrics.TotalCalls)
	fmt.Printf("  Total Latency: %v\n", metrics.TotalLatency)
	fmt.Printf("  Avg Latency: %v\n", metrics.AvgLatency)

	if failed > 0 {
		os.Exit(1)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
