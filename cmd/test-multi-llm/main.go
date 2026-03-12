// Test Multi-LLM Chat
// ทดสอบแชทกับหลาย LLM: antigravity/gemini-3-flash, moonshot/kimi-k2.5, kimi-coding/kimi-for-coding

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
		model      = flag.String("model", "", "Model to test (gemini, moonshot, kimi-coding, or 'all')")
		message    = flag.String("message", "", "Message to send (if empty, enters interactive mode)")
	)
	flag.Parse()

	if *configPath == "" {
		homeDir, _ := os.UserHomeDir()
		*configPath = homeDir + "/.picoclaw/config.json"
	}

	// ตรวจสอบ config มีโมเดลที่ต้องการหรือไม่
	if err := verifyModels(*configPath); err != nil {
		fmt.Printf("❌ Config error: %v\n", err)
		os.Exit(1)
	}

	// ทดสอบตามโมเดลที่เลือก
	models := []string{}
	switch *model {
	case "gemini":
		models = []string{"antigravity/gemini-3-flash"}
	case "moonshot", "kimi":
		models = []string{"moonshot/kimi-k2.5"}
	case "kimi-coding":
		models = []string{"kimi-coding/kimi-for-coding"}
	case "all", "":
		models = []string{"antigravity/gemini-3-flash", "moonshot/kimi-k2.5", "kimi-coding/kimi-for-coding"}
	default:
		// ถ้าไม่ใช่ชื่อที่รู้จัก ให้ถือว่าเป็น model ID โดยตรง (เช่น bailian/qwen3.5-plus)
		models = []string{*model}
	}

	// ถ้ามี message ให้ทดสอบแบบ one-shot
	if *message != "" {
		runOneShot(*configPath, models, *message)
		return
	}

	// ถ้าไม่มี message เข้า interactive mode
	if len(models) == 1 {
		runInteractive(*configPath, models[0])
	} else {
		fmt.Println("🚀 Multi-LLM Interactive Tester")
		fmt.Println("================================")
		fmt.Printf("Models: %v\n", models)
		fmt.Println("\nCommands:")
		fmt.Println("  /switch <gemini|moonshot|kimi-coding> - Switch model")
		fmt.Println("  /models                               - List available models")
		fmt.Println("  /compare <message>                    - Compare all models")
		fmt.Println("  /exit                                 - Exit")
		fmt.Println()
		runMultiInteractive(*configPath, models)
	}
}

func verifyModels(cfgPath string) error {
	// Quick check by loading harness
	tester, err := testharness.NewRealLLMInteractiveTester(cfgPath)
	if err != nil {
		return err
	}
	_ = tester
	return nil
}

func runOneShot(cfgPath string, models []string, message string) {
	fmt.Println("🧪 One-Shot Multi-LLM Test")
	fmt.Println(strings.Repeat("═", 60))

	for _, m := range models {
		fmt.Printf("\n🤖 Model: %s\n", m)
		fmt.Println(strings.Repeat("─", 40))

		tester, err := testharness.NewRealLLMInteractiveTesterWithModel(cfgPath, m)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		start := time.Now()
		response, _, err := tester.ProcessSingle(message)
		latency := time.Since(start)

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Printf("⏱️  Latency: %v\n", latency)
			fmt.Printf("💬 Response: %s\n", response)
		}
	}
}

func runInteractive(cfgPath, model string) {
	fmt.Printf("🚀 Interactive Mode: %s\n", model)
	fmt.Println("Type /exit to quit")
	fmt.Println()

	tester, err := testharness.NewRealLLMInteractiveTesterWithModel(cfgPath, model)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	tester.Start()
}

func runMultiInteractive(cfgPath string, models []string) {
	currentIdx := 0

	for {
		currentModel := models[currentIdx]
		fmt.Printf("\n[%s] 👤 You: ", currentModel)

		var input string
		fmt.Scanln(&input)

		if input == "/exit" {
			break
		}

		if strings.HasPrefix(input, "/switch ") {
			modelName := strings.TrimPrefix(input, "/switch ")
			switch modelName {
			case "gemini":
				currentIdx = 0
			case "moonshot", "kimi":
				currentIdx = 1
			case "kimi-coding":
				currentIdx = 2
			default:
				fmt.Println("Unknown model. Use: gemini, moonshot, kimi-coding")
			}
			continue
		}

		if input == "/models" {
			for i, m := range models {
				marker := " "
				if i == currentIdx {
					marker = "▶"
				}
				fmt.Printf("  %s %d. %s\n", marker, i+1, m)
			}
			continue
		}

		if strings.HasPrefix(input, "/compare ") {
			msg := strings.TrimPrefix(input, "/compare ")
			runOneShot(cfgPath, models, msg)
			continue
		}

		// Process with current model
		tester, err := testharness.NewRealLLMInteractiveTesterWithModel(cfgPath, currentModel)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		start := time.Now()
		response, _, err := tester.ProcessSingle(input)
		latency := time.Since(start)

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Printf("🤖 AI (%s, %v): %s\n", currentModel, latency, response)
		}
	}
}
