// Test RAG Memory with Full System
// ทดสอบระบบ RAG Memory โดยใช้ AgentLoop เต็มรูปแบบ

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

	// Find default config
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

	fmt.Println("🚀 Full RAG Memory System Test")
	fmt.Println("================================")
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

	// Check RAG status
	fmt.Println("\n📊 System Status:")
	if sys.MemoryManager != nil {
		fmt.Printf("  MemoryManager: ✅ Initialized\n")
		fmt.Printf("  RAG Enabled: %v\n", sys.MemoryManager.IsRAGEnabled())
		stats := sys.MemoryManager.GetStats()
		for k, v := range stats {
			fmt.Printf("  %s: %v\n", k, v)
		}
	} else {
		fmt.Println("  MemoryManager: ❌ Not available")
	}

	// Create message bus
	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	// Create provider
	fmt.Println("\n🔌 Creating LLM provider...")
	provider, _, err := createProviderFromConfig(cfg)
	if err != nil {
		fmt.Printf("❌ Failed to create provider: %v\n", err)
		os.Exit(1)
	}

	// Create AgentLoop with MemoryManager
	fmt.Println("🤖 Creating AgentLoop...")
	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider, sys.JobMgr, sys.MemoryManager)

	// Run tests
	fmt.Println("\n🧪 Running RAG Memory Tests...")
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "Test 1: Save User Info",
			message: "ฉันชื่อ TestUser กำลังพัฒนาโปรเจค PicoClaw",
		},
		{
			name:    "Test 2: Query User Info",
			message: "จำได้ไหมว่าฉันชื่ออะไร และทำโปรเจคอะไร",
		},
		{
			name:    "Test 3: Technical Info",
			message: "ระบบ RAG ใช้ SQLite และ embedding service ที่ port 18190",
		},
		{
			name:    "Test 4: Query Technical Info",
			message: "ระบบใช้ฐานข้อมูลอะไร และ port อะไร",
		},
	}

	passed := 0
	failed := 0

	for _, test := range tests {
		fmt.Printf("📝 %s\n", test.name)
		fmt.Printf("   Message: %s\n", test.message)

		start := time.Now()
		response, err := testMessage(agentLoop, test.message)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("   ❌ FAIL: %v\n", err)
			failed++
		} else {
			fmt.Printf("   ✅ PASS (%v)\n", duration)
			fmt.Printf("   Response: %s\n", truncate(response, 150))
			passed++
		}
		fmt.Println()

		// Wait for background tasks (memory extraction) to complete
		fmt.Println("   ⏳ Waiting for background tasks...")
		agentLoop.WaitForBackgroundTasks()
		fmt.Println("   ✅ Background tasks complete")
		fmt.Println()

		// Small delay between tests
		time.Sleep(1 * time.Second)
	}

	// Check RAG documents after tests
	fmt.Println("\n📊 Checking RAG Database...")
	if sys.MemoryManager != nil && sys.MemoryManager.IsRAGEnabled() {
		// Query to check if documents were saved
		result, err := sys.MemoryManager.QueryMemory("TestUser", 10)
		if err != nil {
			fmt.Printf("  Query error: %v\n", err)
		} else {
			fmt.Printf("  Documents found: %d\n", len(result.Documents))
			for i, doc := range result.Documents {
				fmt.Printf("    %d. [%s] %s\n", i+1, doc.Metadata.Source, truncate(doc.Content, 80))
			}
		}
	}

	// Print summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📊 Test Summary")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Passed: %d ✅\n", passed)
	fmt.Printf("Failed: %d ❌\n", failed)
	fmt.Println(strings.Repeat("=", 60))

	if failed > 0 {
		os.Exit(1)
	}

	fmt.Println("\n✅ All tests passed!")
}

func testMessage(agentLoop *agent.AgentLoop, message string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := agentLoop.ProcessDirect(ctx, message, "test-session", nil)
	if err != nil {
		return "", err
	}

	return response, nil
}

func createProviderFromConfig(cfg *config.Config) (providers.LLMProvider, string, error) {
	// Use first model from model_list
	if len(cfg.ModelList) == 0 {
		return nil, "", fmt.Errorf("no models configured")
	}

	modelCfg := cfg.ModelList[0]
	modelName := modelCfg.ModelName
	if modelName == "" {
		modelName = modelCfg.Model
	}

	// Create provider using factory
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
