// Test RAG Memory System
// ใช้ระบบ testharness ที่มีอยู่เพื่อทดสอบ RAG Memory

package main

import (
	"flag"
	"fmt"
	"os"

	"picoclaw/agent/pkg/testharness"
)

func main() {
	var (
		listTests = flag.Bool("list", false, "List all RAG test scenarios")
		runAll    = flag.Bool("all", true, "Run all RAG memory tests")
	)

	flag.Parse()

	if *listTests {
		fmt.Println("📋 RAG Memory Test Scenarios:")
		fmt.Println()

		// แสดงรายการ scenarios ที่เกี่ยวกับ RAG
		allScenarios := append(testharness.CommonScenarios, testharness.RAGMemoryScenarios...)
		for _, s := range allScenarios {
			if contains(s.Name, "RAG") || contains(s.Name, "Memory") {
				fmt.Printf("  📌 %s\n", s.Name)
				fmt.Printf("     %s\n", s.Description)
				fmt.Println()
			}
		}
		return
	}

	if *runAll {
		fmt.Println("🚀 RAG Memory System Test")
		fmt.Println("==========================")
		fmt.Println()

		// รันการทดสอบ RAG Memory
		results := testharness.RunRAGMemoryScenarios()

		// พิมพ์รายงาน
		testharness.PrintRAGReport(results)

		// ตรวจสอบผล
		for _, r := range results {
			if !r.Passed {
				os.Exit(1)
			}
		}

		fmt.Println("\n✅ All RAG Memory tests passed!")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
