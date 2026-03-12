// PicoClaw - Real LLM Interactive Test
// ระบบทดสอบ interactive ที่ใช้งาน LLM จริง

package testharness

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// RealLLMInteractiveTester ระบบทดสอบ interactive กับ LLM จริง
type RealLLMInteractiveTester struct {
	harness    *RealLLMTestHarness
	reader     *bufio.Reader
	printer    RealLLMPrinter
	history    []InteractiveMessage
	isRunning  bool
	debugMode  bool
}

// InteractiveMessage ข้อความในการสนทนา interactive
type InteractiveMessage struct {
	Role      string
	Content   string
	Latency   time.Duration
	Timestamp time.Time
}

// RealLLMPrinter interface สำหรับการแสดงผล
type RealLLMPrinter interface {
	PrintWelcome(modelName string)
	PrintUser(message string)
	PrintAssistant(message string, latency time.Duration)
	PrintSystem(message string)
	PrintError(err error)
	PrintDivider()
	PrintMetrics(metrics TestMetrics)
	PrintDebug(info string)
}

// DefaultRealLLMPrinter printer เริ่มต้นสำหรับ Real LLM
type DefaultRealLLMPrinter struct{}

// PrintWelcome แสดงข้อความต้อนรับ
func (p DefaultRealLLMPrinter) PrintWelcome(modelName string) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           🚀 Real LLM Interactive Tester 🚀               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Printf("  Model: %s\n", modelName)
	fmt.Println("  Type /help for available commands")
	fmt.Println()
}

// PrintUser แสดงข้อความจาก user
func (p DefaultRealLLMPrinter) PrintUser(message string) {
	fmt.Printf("\n\033[36m👤 You:\033[0m %s\n", message)
}

// PrintAssistant แสดงข้อความจาก assistant
func (p DefaultRealLLMPrinter) PrintAssistant(message string, latency time.Duration) {
	fmt.Printf("\033[32m🤖 AI:\033[0m ")
	// พิมพ์ทีละบรรทัดเพื่อความสวยงาม
	lines := strings.Split(message, "\n")
	for i, line := range lines {
		if i > 0 {
			fmt.Print("      ")
		}
		fmt.Println(line)
	}
	if latency > 0 {
		fmt.Printf("\033[90m      (response time: %v)\033[0m\n", latency)
	}
}

// PrintSystem แสดงข้อความระบบ
func (p DefaultRealLLMPrinter) PrintSystem(message string) {
	fmt.Printf("\n\033[33mℹ️  %s\033[0m\n", message)
}

// PrintError แสดง error
func (p DefaultRealLLMPrinter) PrintError(err error) {
	fmt.Printf("\n\033[31m❌ Error: %v\033[0m\n", err)
}

// PrintDivider แสดงเส้นแบ่ง
func (p DefaultRealLLMPrinter) PrintDivider() {
	fmt.Println(strings.Repeat("─", 60))
}

// PrintMetrics แสดง metrics
func (p DefaultRealLLMPrinter) PrintMetrics(metrics TestMetrics) {
	fmt.Println()
	fmt.Println("┌───────────────── Test Metrics ─────────────────┐")
	fmt.Printf("│ Total Calls:    %d\n", metrics.TotalCalls)
	fmt.Printf("│ Total Latency:  %v\n", metrics.TotalLatency)
	fmt.Printf("│ Avg Latency:    %v\n", metrics.AvgLatency)
	fmt.Printf("│ Errors:         %d\n", metrics.Errors)
	fmt.Println("└────────────────────────────────────────────────┘")
}

// PrintDebug แสดงข้อมูล debug
func (p DefaultRealLLMPrinter) PrintDebug(info string) {
	fmt.Printf("\033[90m[DEBUG] %s\033[0m\n", info)
}

// NewRealLLMInteractiveTester สร้าง Interactive Tester ใหม่
func NewRealLLMInteractiveTester(cfgPath string) (*RealLLMInteractiveTester, error) {
	harness, err := NewRealLLMTestHarness(cfgPath)
	if err != nil {
		return nil, err
	}

	tester := &RealLLMInteractiveTester{
		harness:   harness,
		reader:    bufio.NewReader(os.Stdin),
		printer:   DefaultRealLLMPrinter{},
		history:   make([]InteractiveMessage, 0),
		debugMode: false,
	}

	return tester, nil
}

// NewRealLLMInteractiveTesterWithModel สร้าง tester ด้วย model ที่ระบุ
func NewRealLLMInteractiveTesterWithModel(cfgPath, modelName string) (*RealLLMInteractiveTester, error) {
	harness, err := NewRealLLMTestHarnessWithModel(cfgPath, modelName)
	if err != nil {
		return nil, err
	}

	tester := &RealLLMInteractiveTester{
		harness:   harness,
		reader:    bufio.NewReader(os.Stdin),
		printer:   DefaultRealLLMPrinter{},
		history:   make([]InteractiveMessage, 0),
		debugMode: false,
	}

	return tester, nil
}

// WithPrinter กำหนด printer ที่กำหนดเอง
func (t *RealLLMInteractiveTester) WithPrinter(printer RealLLMPrinter) *RealLLMInteractiveTester {
	t.printer = printer
	return t
}

// WithDebugMode เปิด/ปิด debug mode
func (t *RealLLMInteractiveTester) WithDebugMode(enabled bool) *RealLLMInteractiveTester {
	t.debugMode = enabled
	return t
}

// Start เริ่มการทดสอบ interactive
func (t *RealLLMInteractiveTester) Start() {
	t.isRunning = true
	t.printer.PrintWelcome(t.harness.GetModelName())

	for t.isRunning {
		t.printer.PrintDivider()
		fmt.Print("\033[36m👤 You:\033[0m ")

		input, err := t.reader.ReadString('\n')
		if err != nil {
			t.printer.PrintError(err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// จัดการคำสั่งพิเศษ
		if t.handleCommand(input) {
			continue
		}

		// ประมวลผลข้อความ
		t.processMessage(input)
	}
}

// StartWithTimeout เริ่มการทดสอบพร้อม timeout
func (t *RealLLMInteractiveTester) StartWithTimeout(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan bool)
	go func() {
		t.Start()
		done <- true
	}()

	select {
	case <-done:
		return
	case <-ctx.Done():
		t.printer.PrintSystem("Session timeout. Goodbye!")
		t.Stop()
	}
}

// Stop หยุดการทดสอบ
func (t *RealLLMInteractiveTester) Stop() {
	t.isRunning = false
}

// ProcessSingle ประมวลผลข้อความเดียว
func (t *RealLLMInteractiveTester) ProcessSingle(message string) (string, time.Duration, error) {
	start := time.Now()
	response, err := t.harness.Chat(message)
	latency := time.Since(start)

	if err != nil {
		return "", 0, err
	}

	// บันทึกประวัติ
	t.history = append(t.history, InteractiveMessage{
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	})
	t.history = append(t.history, InteractiveMessage{
		Role:      "assistant",
		Content:   response,
		Latency:   latency,
		Timestamp: time.Now(),
	})

	return response, latency, nil
}

// ProcessBatch ประมวลผลหลายข้อความต่อเนื่อง
func (t *RealLLMInteractiveTester) ProcessBatch(messages []string) ([]BatchResult, error) {
	results := make([]BatchResult, 0, len(messages))

	for _, msg := range messages {
		start := time.Now()
		response, err := t.harness.Chat(msg)
		latency := time.Since(start)

		result := BatchResult{
			Input:     msg,
			Response:  response,
			Latency:   latency,
			Timestamp: time.Now(),
		}

		if err != nil {
			result.Error = err
		}

		results = append(results, result)

		// แสดงผลแบบ real-time
		t.printer.PrintUser(msg)
		if err != nil {
			t.printer.PrintError(err)
		} else {
			t.printer.PrintAssistant(response, latency)
		}
	}

	return results, nil
}

// GetHarness คืนค่า harness
func (t *RealLLMInteractiveTester) GetHarness() *RealLLMTestHarness {
	return t.harness
}

// GetHistory คืนค่าประวัติการสนทนา
func (t *RealLLMInteractiveTester) GetHistory() []InteractiveMessage {
	return t.history
}

// handleCommand จัดการคำสั่งพิเศษ
func (t *RealLLMInteractiveTester) handleCommand(input string) bool {
	switch strings.ToLower(input) {
	case "/help", "/h":
		t.showHelp()
		return true
	case "/history", "/hist":
		t.showHistory()
		return true
	case "/clear", "/c":
		t.clearHistory()
		t.printer.PrintSystem("Conversation history cleared.")
		return true
	case "/reset", "/r":
		t.harness.Reset()
		t.history = make([]InteractiveMessage, 0)
		t.printer.PrintSystem("Tester reset.")
		return true
	case "/metrics", "/m":
		t.showMetrics()
		return true
	case "/model":
		t.printer.PrintSystem(fmt.Sprintf("Current model: %s", t.harness.GetModelName()))
		return true
	case "/config":
		t.showConfig()
		return true
	case "/save":
		t.saveConversation()
		return true
	case "/debug":
		t.debugMode = !t.debugMode
		if t.debugMode {
			t.printer.PrintSystem("Debug mode enabled.")
		} else {
			t.printer.PrintSystem("Debug mode disabled.")
		}
		return true
	case "/exit", "/quit", "/q":
		t.printer.PrintSystem("Goodbye! 👋")
		t.Stop()
		return true
	}

	// คำสั่งที่มี argument
	if strings.HasPrefix(strings.ToLower(input), "/temp ") {
		var temp float64
		if _, err := fmt.Sscanf(input, "/temp %f", &temp); err == nil {
			t.harness.WithTemperature(temp)
			t.printer.PrintSystem(fmt.Sprintf("Temperature set to %.2f", temp))
		} else {
			t.printer.PrintError(fmt.Errorf("invalid temperature value"))
		}
		return true
	}

	if strings.HasPrefix(strings.ToLower(input), "/tokens ") {
		var tokens int
		if _, err := fmt.Sscanf(input, "/tokens %d", &tokens); err == nil {
			t.harness.WithMaxTokens(tokens)
			t.printer.PrintSystem(fmt.Sprintf("Max tokens set to %d", tokens))
		} else {
			t.printer.PrintError(fmt.Errorf("invalid tokens value"))
		}
		return true
	}

	return false
}

// processMessage ประมวลผลข้อความ
func (t *RealLLMInteractiveTester) processMessage(message string) {
	if t.debugMode {
		t.printer.PrintDebug(fmt.Sprintf("Sending message to %s", t.harness.GetModelName()))
	}

	response, latency, err := t.ProcessSingle(message)
	if err != nil {
		t.printer.PrintError(err)
		return
	}

	t.printer.PrintAssistant(response, latency)
}

// showHelp แสดงวิธีใช้
func (t *RealLLMInteractiveTester) showHelp() {
	t.printer.PrintSystem("Available Commands:")
	fmt.Println("  General:")
	fmt.Println("    /help, /h        - Show this help message")
	fmt.Println("    /exit, /quit     - Exit the tester")
	fmt.Println()
	fmt.Println("  Conversation:")
	fmt.Println("    /history, /hist  - Show conversation history")
	fmt.Println("    /clear, /c       - Clear conversation history")
	fmt.Println("    /reset, /r       - Reset tester state")
	fmt.Println("    /save            - Save conversation to file")
	fmt.Println()
	fmt.Println("  Configuration:")
	fmt.Println("    /model           - Show current model")
	fmt.Println("    /config          - Show configuration")
	fmt.Println("    /temp <value>    - Set temperature (0.0-2.0)")
	fmt.Println("    /tokens <value>  - Set max tokens")
	fmt.Println()
	fmt.Println("  Debug:")
	fmt.Println("    /metrics, /m     - Show test metrics")
	fmt.Println("    /debug           - Toggle debug mode")
}

// showHistory แสดงประวัติการสนทนา
func (t *RealLLMInteractiveTester) showHistory() {
	if len(t.history) == 0 {
		t.printer.PrintSystem("No conversation history yet.")
		return
	}

	t.printer.PrintSystem("Conversation History:")
	for _, msg := range t.history {
		timestamp := msg.Timestamp.Format("15:04:05")
		if msg.Role == "user" {
			fmt.Printf("  [%s] You: %s\n", timestamp, msg.Content)
		} else {
			fmt.Printf("  [%s] AI: %s", timestamp, msg.Content)
			if msg.Latency > 0 {
				fmt.Printf(" (%v)", msg.Latency)
			}
			fmt.Println()
		}
	}
}

// clearHistory ล้างประวัติการสนทนา
func (t *RealLLMInteractiveTester) clearHistory() {
	t.history = make([]InteractiveMessage, 0)
	t.harness.ClearConversation()
}

// showMetrics แสดง metrics
func (t *RealLLMInteractiveTester) showMetrics() {
	metrics := t.harness.GetMetrics()
	t.printer.PrintMetrics(metrics)
}

// showConfig แสดง configuration
func (t *RealLLMInteractiveTester) showConfig() {
	t.printer.PrintSystem("Current Configuration:")
	fmt.Printf("  Model:       %s\n", t.harness.GetModelName())
	fmt.Printf("  Max Tokens:  %d\n", t.harness.maxTokens)
	fmt.Printf("  Temperature: %.2f\n", t.harness.temperature)
	fmt.Printf("  Timeout:     %v\n", t.harness.timeout)
}

// saveConversation บันทึกการสนทนา
func (t *RealLLMInteractiveTester) saveConversation() {
	filename := fmt.Sprintf("conversation_%s.json", time.Now().Format("20060102_150405"))
	if err := t.harness.SaveConversation(filename); err != nil {
		t.printer.PrintError(err)
	} else {
		t.printer.PrintSystem(fmt.Sprintf("Conversation saved to %s", filename))
	}
}

// BatchResult ผลลัพธ์การทดสอบแบบ batch
type BatchResult struct {
	Input     string
	Response  string
	Latency   time.Duration
	Timestamp time.Time
	Error     error
}

// RunBatchTest รันการทดสอบแบบ batch
func RunBatchTest(cfgPath string, testCases []string) ([]BatchResult, error) {
	tester, err := NewRealLLMInteractiveTester(cfgPath)
	if err != nil {
		return nil, err
	}

	return tester.ProcessBatch(testCases)
}

// QuickTest ทดสอบอย่างรวดเร็ว
func QuickTest(cfgPath, message string) (string, error) {
	tester, err := NewRealLLMInteractiveTester(cfgPath)
	if err != nil {
		return "", err
	}

	response, _, err := tester.ProcessSingle(message)
	return response, err
}
