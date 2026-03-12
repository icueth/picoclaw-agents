// PicoClaw - LLM Chat Test Harness
// Interactive Chat Simulator สำหรับการทดสอบแบบโต้ตอบ

package testharness

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// InteractiveSimulator จำลองการแชทแบบ interactive
type InteractiveSimulator struct {
	harness     *Harness
	provider    *MockProvider
	isRunning   bool
	reader      *bufio.Reader
	printer     Printer
	history     []ChatMessage
	maxHistory  int
}

// ChatMessage ข้อความในการสนทนา
type ChatMessage struct {
	Role      string
	Content   string
	Timestamp time.Time
}

// Printer interface สำหรับการแสดงผล
type Printer interface {
	PrintUser(message string)
	PrintAssistant(message string)
	PrintSystem(message string)
	PrintToolCall(toolName string, args map[string]any)
	PrintToolResult(result string)
	PrintError(err error)
	PrintDivider()
}

// DefaultPrinter printer เริ่มต้น
type DefaultPrinter struct{}

// PrintUser แสดงข้อความจาก user
func (p DefaultPrinter) PrintUser(message string) {
	fmt.Printf("\n\033[36m[You]:\033[0m %s\n", message)
}

// PrintAssistant แสดงข้อความจาก assistant
func (p DefaultPrinter) PrintAssistant(message string) {
	fmt.Printf("\033[32m[AI]:\033[0m %s\n", message)
}

// PrintSystem แสดงข้อความระบบ
func (p DefaultPrinter) PrintSystem(message string) {
	fmt.Printf("\033[33m[System]:\033[0m %s\n", message)
}

// PrintToolCall แสดงการเรียก tool
func (p DefaultPrinter) PrintToolCall(toolName string, args map[string]any) {
	fmt.Printf("\033[35m[Tool]:\033[0m %s\n", toolName)
}

// PrintToolResult แสดงผลลัพธ์จาก tool
func (p DefaultPrinter) PrintToolResult(result string) {
	fmt.Printf("\033[35m[Result]:\033[0m %s\n", result)
}

// PrintError แสดง error
func (p DefaultPrinter) PrintError(err error) {
	fmt.Printf("\033[31m[Error]:\033[0m %v\n", err)
}

// PrintDivider แสดงเส้นแบ่ง
func (p DefaultPrinter) PrintDivider() {
	fmt.Println(strings.Repeat("-", 50))
}

// NewInteractiveSimulator สร้าง Interactive Simulator ใหม่
func NewInteractiveSimulator(provider *MockProvider) *InteractiveSimulator {
	return &InteractiveSimulator{
		harness:    New(provider),
		provider:   provider,
		reader:     bufio.NewReader(os.Stdin),
		printer:    DefaultPrinter{},
		history:    make([]ChatMessage, 0),
		maxHistory: 100,
	}
}

// WithPrinter กำหนด printer ที่กำหนดเอง
func (s *InteractiveSimulator) WithPrinter(printer Printer) *InteractiveSimulator {
	s.printer = printer
	return s
}

// WithMaxHistory กำหนดจำนวนประวัติสูงสุด
func (s *InteractiveSimulator) WithMaxHistory(max int) *InteractiveSimulator {
	s.maxHistory = max
	return s
}

// Start เริ่มการจำลองการแชท
func (s *InteractiveSimulator) Start() {
	s.isRunning = true
	s.showWelcome()

	for s.isRunning {
		s.printer.PrintDivider()
		fmt.Print("\033[36m[You]:\033[0m ")

		input, err := s.reader.ReadString('\n')
		if err != nil {
			s.printer.PrintError(err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// จัดการคำสั่งพิเศษ
		if s.handleCommand(input) {
			continue
		}

		// ประมวลผลข้อความ
		s.processMessage(input)
	}
}

// StartWithTimeout เริ่มการจำลองพร้อม timeout
func (s *InteractiveSimulator) StartWithTimeout(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan bool)
	go func() {
		s.Start()
		done <- true
	}()

	select {
	case <-done:
		return
	case <-ctx.Done():
		s.printer.PrintSystem("Session timeout. Goodbye!")
		s.Stop()
	}
}

// Stop หยุดการจำลอง
func (s *InteractiveSimulator) Stop() {
	s.isRunning = false
}

// ProcessSingle ประมวลผลข้อความเดียว (สำหรับการทดสอบ)
func (s *InteractiveSimulator) ProcessSingle(message string) (string, error) {
	return s.harness.Chat(message)
}

// ProcessBatch ประมวลผลหลายข้อความต่อเนื่อง
func (s *InteractiveSimulator) ProcessBatch(messages []string) ([]string, error) {
	return s.harness.MultiTurnChat(messages)
}

// GetHistory คืนค่าประวัติการสนทนา
func (s *InteractiveSimulator) GetHistory() []ChatMessage {
	return s.history
}

// ClearHistory ล้างประวัติการสนทนา
func (s *InteractiveSimulator) ClearHistory() {
	s.history = make([]ChatMessage, 0)
	s.harness.ClearConversation()
}

// showWelcome แสดงข้อความต้อนรับ
func (s *InteractiveSimulator) showWelcome() {
	s.printer.PrintSystem("Welcome to LLM Chat Simulator!")
	s.printer.PrintSystem("Type your message to chat with the AI.")
	s.printer.PrintSystem("Commands: /help, /history, /clear, /reset, /exit")
	s.printer.PrintDivider()
}

// handleCommand จัดการคำสั่งพิเศษ
func (s *InteractiveSimulator) handleCommand(input string) bool {
	switch strings.ToLower(input) {
	case "/help", "/h":
		s.showHelp()
		return true
	case "/history", "/hist":
		s.showHistory()
		return true
	case "/clear", "/c":
		s.ClearHistory()
		s.printer.PrintSystem("History cleared.")
		return true
	case "/reset", "/r":
		s.harness.Reset()
		s.ClearHistory()
		s.printer.PrintSystem("Simulator reset.")
		return true
	case "/exit", "/quit", "/q":
		s.printer.PrintSystem("Goodbye!")
		s.Stop()
		return true
	case "/status":
		s.showStatus()
		return true
	case "/rules":
		s.showRules()
		return true
	}
	return false
}

// processMessage ประมวลผลข้อความ
func (s *InteractiveSimulator) processMessage(message string) {
	// บันทึกข้อความจาก user
	s.addToHistory("user", message)

	// แสดงข้อความ (ถ้าไม่ได้แสดงไปแล้ว)
	// s.printer.PrintUser(message)

	// ประมวลผล
	start := time.Now()
	response, err := s.harness.Chat(message)
	elapsed := time.Since(start)

	if err != nil {
		s.printer.PrintError(err)
		return
	}

	// บันทึกข้อความตอบกลับ
	s.addToHistory("assistant", response)

	// แสดงผล
	s.printer.PrintAssistant(response)

	// แสดงเวลาประมวลผล (optional)
	if elapsed > 100*time.Millisecond {
		fmt.Printf("\033[90m(%v)\033[0m\n", elapsed)
	}

	// ตรวจสอบ tool calls
	if s.provider.VerifyToolCall("web_search") {
		s.printer.PrintToolCall("web_search", nil)
	}
}

// addToHistory เพิ่มข้อความลงในประวัติ
func (s *InteractiveSimulator) addToHistory(role, content string) {
	s.history = append(s.history, ChatMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})

	// จำกัดขนาดประวัติ
	if len(s.history) > s.maxHistory {
		s.history = s.history[len(s.history)-s.maxHistory:]
	}
}

// showHelp แสดงวิธีใช้
func (s *InteractiveSimulator) showHelp() {
	s.printer.PrintSystem("Available Commands:")
	fmt.Println("  /help, /h      - Show this help message")
	fmt.Println("  /history, /hist - Show conversation history")
	fmt.Println("  /clear, /c     - Clear conversation history")
	fmt.Println("  /reset, /r     - Reset simulator state")
	fmt.Println("  /status        - Show simulator status")
	fmt.Println("  /rules         - Show configured response rules")
	fmt.Println("  /exit, /quit   - Exit simulator")
}

// showHistory แสดงประวัติการสนทนา
func (s *InteractiveSimulator) showHistory() {
	if len(s.history) == 0 {
		s.printer.PrintSystem("No history yet.")
		return
	}

	s.printer.PrintSystem("Conversation History:")
	for _, msg := range s.history {
		timestamp := msg.Timestamp.Format("15:04:05")
		if msg.Role == "user" {
			fmt.Printf("  [%s] You: %s\n", timestamp, msg.Content)
		} else {
			fmt.Printf("  [%s] AI: %s\n", timestamp, msg.Content)
		}
	}
}

// showStatus แสดงสถานะ simulator
func (s *InteractiveSimulator) showStatus() {
	s.printer.PrintSystem("Simulator Status:")
	fmt.Printf("  Total messages: %d\n", len(s.history))
	fmt.Printf("  Provider calls: %d\n", s.provider.GetCallCount())
	fmt.Printf("  Max history: %d\n", s.maxHistory)
}

// showRules แสดงกฎการตอบสนองที่ตั้งไว้
func (s *InteractiveSimulator) showRules() {
	s.printer.PrintSystem("Response Rules:")
	// Note: ต้องเพิ่ม method ใน MockProvider เพื่อ expose rules
	fmt.Println("  (Rules are configured in the provider)")
}

// ScriptedSimulator จำลองการแชทตามสคริปต์ที่กำหนด
type ScriptedSimulator struct {
	harness  *Harness
	provider *MockProvider
	script   []ScriptStep
	results  []StepResult
}

// ScriptStep ขั้นตอนในสคริปต์
type ScriptStep struct {
	Input           string
	ExpectedOutput  string
	ExpectedTool    string
	ShouldError     bool
	Description     string
}

// StepResult ผลลัพธ์ของแต่ละขั้นตอน
type StepResult struct {
	Step      ScriptStep
	Actual    string
	Passed    bool
	Error     error
	Duration  time.Duration
}

// NewScriptedSimulator สร้าง Scripted Simulator ใหม่
func NewScriptedSimulator(provider *MockProvider, script []ScriptStep) *ScriptedSimulator {
	return &ScriptedSimulator{
		harness:  New(provider),
		provider: provider,
		script:   script,
		results:  make([]StepResult, 0),
	}
}

// Run รันสคริปต์
func (s *ScriptedSimulator) Run() []StepResult {
	s.results = make([]StepResult, 0, len(s.script))

	for _, step := range s.script {
		start := time.Now()
		actual, err := s.harness.Chat(step.Input)
		duration := time.Since(start)

		result := StepResult{
			Step:     step,
			Actual:   actual,
			Duration: duration,
		}

		if err != nil {
			result.Error = err
			if !step.ShouldError {
				result.Passed = false
			} else {
				result.Passed = true
			}
		} else {
			if step.ShouldError {
				result.Passed = false
				result.Error = fmt.Errorf("expected error but got success")
			} else if step.ExpectedOutput != "" && !strings.Contains(actual, step.ExpectedOutput) {
				result.Passed = false
				result.Error = fmt.Errorf("expected output containing %q, got %q", step.ExpectedOutput, actual)
			} else if step.ExpectedTool != "" && !s.provider.VerifyToolCall(step.ExpectedTool) {
				result.Passed = false
				result.Error = fmt.Errorf("expected tool %q to be called", step.ExpectedTool)
			} else {
				result.Passed = true
			}
		}

		s.results = append(s.results, result)
	}

	return s.results
}

// GetResults คืนค่าผลลัพธ์
func (s *ScriptedSimulator) GetResults() []StepResult {
	return s.results
}

// PrintResults พิมพ์ผลลัพธ์
func (s *ScriptedSimulator) PrintResults() string {
	var sb strings.Builder
	passed := 0
	failed := 0

	sb.WriteString("========================================\n")
	sb.WriteString("     Scripted Test Results\n")
	sb.WriteString("========================================\n\n")

	for i, r := range s.results {
		status := "✓ PASS"
		if !r.Passed {
			status = "✗ FAIL"
			failed++
		} else {
			passed++
		}

		sb.WriteString(fmt.Sprintf("Step %d: %s\n", i+1, status))
		sb.WriteString(fmt.Sprintf("  Input: %s\n", r.Step.Input))
		sb.WriteString(fmt.Sprintf("  Description: %s\n", r.Step.Description))
		sb.WriteString(fmt.Sprintf("  Duration: %v\n", r.Duration))
		if r.Error != nil {
			sb.WriteString(fmt.Sprintf("  Error: %v\n", r.Error))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("========================================\n")
	sb.WriteString(fmt.Sprintf("Total: %d | Passed: %d | Failed: %d\n", len(s.results), passed, failed))
	sb.WriteString("========================================\n")

	return sb.String()
}
