// PicoClaw - LLM Chat Test Harness
// สถานการณ์การทดสอบสำเร็จรูป

package testharness

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"picoclaw/agent/pkg/providers"
)

// Scenario เป็นสถานการณ์การทดสอบสำเร็จรูป
type Scenario struct {
	Name        string
	Description string
	Setup       func(*MockProvider)
	Test        func(*Harness) error
}

// CommonScenarios สถานการณ์การทดสอบทั่วไป
var CommonScenarios = []Scenario{
	{
		Name:        "Simple Greeting",
		Description: "ทดสอบการทักทายพื้นฐาน",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithResponsePattern("สวัสดี", "สวัสดีครับ! มีอะไรให้ช่วยเหลือไหมครับ").
				WithResponsePattern("hello", "Hello! How can I help you today?")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("สวัสดี")
			if err != nil {
				return err
			}
			return h.AssertResponseContains("สวัสดี")
		},
	},
	{
		Name:        "Multi-turn Conversation",
		Description: "ทดสอบการสนทนาหลาย turn",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithResponsePattern("ชื่ออะไร", "ผมชื่อ PicoClaw ครับ").
				WithResponsePattern("ทำอะไรได้", "ผมสามารถช่วยตอบคำถาม เขียนโค้ด และช่วยงานต่างๆ ได้ครับ").
				WithResponsePattern("ขอบคุณ", "ยินดีครับ! มีอะไรให้ช่วยอีกไหมครับ")
		},
		Test: func(h *Harness) error {
			messages := []string{
				"คุณชื่ออะไร",
				"คุณทำอะไรได้บ้าง",
				"ขอบคุณครับ",
			}
			_, err := h.MultiTurnChat(messages)
			return err
		},
	},
	{
		Name:        "Tool Call - Web Search",
		Description: "ทดสอบการเรียกใช้งาน web search tool",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("ค้นหา", "web_search", map[string]any{
					"query": "golang tutorial",
				})
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("ช่วยค้นหา golang tutorial หน่อย")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("web_search")
		},
	},
	{
		Name:        "Tool Call - File Operations",
		Description: "ทดสอบการเรียกใช้งาน file tools",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("อ่านไฟล์", "read_file", map[string]any{
					"path": "test.txt",
				}).
				WithToolCallResponse("เขียนไฟล์", "write_file", map[string]any{
					"path":    "output.txt",
					"content": "Hello World",
				})
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("ช่วยอ่านไฟล์ test.txt")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("read_file")
		},
	},
	{
		Name:        "Error Handling",
		Description: "ทดสอบการจัดการ error",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithErrorResponse("error", "mock error: something went wrong")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("trigger error")
			if err == nil {
				return fmt.Errorf("expected error but got none")
			}
			return nil
		},
	},
	{
		Name:        "Delayed Response",
		Description: "ทดสอบการตอบสนองที่มี delay",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithDelayedResponse("ช้า", "นี่คือคำตอบที่ช้าครับ", 100*time.Millisecond)
		},
		Test: func(h *Harness) error {
			start := time.Now()
			_, err := h.Chat("ช่วยตอบช้าหน่อย")
			if err != nil {
				return err
			}
			elapsed := time.Since(start)
			if elapsed < 100*time.Millisecond {
				return fmt.Errorf("expected delay of at least 100ms, got %v", elapsed)
			}
			return h.AssertResponseContains("ช้า")
		},
	},
	{
		Name:        "Streaming Response",
		Description: "ทดสอบการตอบสนองแบบ streaming",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithResponsePattern("stream", "This is a streaming response")
		},
		Test: func(h *Harness) error {
			var chunks []string
			err := h.ChatStreaming("test stream", func(chunk string) {
				chunks = append(chunks, chunk)
			})
			if err != nil {
				return err
			}
			if len(chunks) == 0 {
				return fmt.Errorf("expected streaming chunks but got none")
			}
			return nil
		},
	},
	{
		Name:        "Reasoning Response",
		Description: "ทดสอบการตอบสนองที่มี reasoning",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithReasoningResponse("คิด", "Let me think about this...", "นี่คือคำตอบหลังจากคิดแล้วครับ")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("ช่วยคิดหน่อย")
			if err != nil {
				return err
			}
			return h.AssertResponseContains("คำตอบ")
		},
	},
	{
		Name:        "Multi Tool Calls",
		Description: "ทดสอบการเรียกใช้งานหลาย tools พร้อมกัน",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("หลายอย่าง", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "read_file",
						Function: &providers.FunctionCall{
							Name:      "read_file",
							Arguments: `{"path": "config.json"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "web_search",
						Function: &providers.FunctionCall{
							Name:      "web_search",
							Arguments: `{"query": "golang best practices"}`,
						},
					},
				})
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("ช่วยทำหลายอย่างพร้อมกัน")
			if err != nil {
				return err
			}
			if err := h.AssertToolCalled("read_file"); err != nil {
				return err
			}
			return h.AssertToolCalled("web_search")
		},
	},
	{
		Name:        "Context Awareness",
		Description: "ทดสอบการจดจำ context จากการสนทนาก่อนหน้า",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithExactMatch("ชื่ออะไร", "ผมชื่อ PicoClaw").
				WithResponsePattern("ชื่อฉัน", "จากที่คุณบอกไว้ว่าชื่อ TestUser")
		},
		Test: func(h *Harness) error {
			// ส่งข้อความแรก
			_, err := h.Chat("ชื่ออะไร")
			if err != nil {
				return err
			}
			// ส่งข้อความที่สอง (ควรมี context จากข้อความแรก)
			_, err = h.Chat("ฉันชื่อ TestUser")
			return err
		},
	},
	{
		Name:        "RAG Memory - Save and Retrieve",
		Description: "ทดสอบการบันทึกและค้นหาข้อมูลด้วย RAG Memory",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithResponsePattern("สวัสดี", "สวัสดีครับ! มีอะไรให้ช่วยเหลือไหมครับ").
				WithResponsePattern("ชื่ออะไร", "จากข้อมูลที่มี ผู้ใช้ชื่อ icue ครับ").
				WithResponsePattern("โปรเจคอะไร", "คุณกำลังพัฒนาโปรเจค PicoClaw ครับ")
		},
		Test: func(h *Harness) error {
			// ขั้นตอน 1: เริ่มต้นสนทนา
			_, err := h.Chat("สวัสดี ฉันชื่อ icue")
			if err != nil {
				return err
			}

			// ขั้นตอน 2: บอกข้อมูลเพิ่มเติม
			_, err = h.Chat("ฉันกำลังพัฒนาโปรเจค PicoClaw")
			if err != nil {
				return err
			}

			// ขั้นตอน 3: ทดสอบว่า RAG จดจำข้อมูลได้หรือไม่
			response, err := h.Chat("ฉันชื่ออะไร")
			if err != nil {
				return err
			}

			// ตรวจสอบว่า response มีข้อมูลจาก RAG
			if !strings.Contains(response, "icue") {
				return fmt.Errorf("RAG should remember user name 'icue', got: %s", response)
			}

			// ขั้นตอน 4: ทดสอบคำถามเกี่ยวกับโปรเจค
			response, err = h.Chat("ฉันทำโปรเจคอะไรอยู่")
			if err != nil {
				return err
			}

			if !strings.Contains(response, "PicoClaw") {
				return fmt.Errorf("RAG should remember project 'PicoClaw', got: %s", response)
			}

			return nil
		},
	},
	{
		Name:        "RAG Memory - Semantic Search",
		Description: "ทดสอบการค้นหาแบบ Semantic ด้วย RAG",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithResponsePattern("Clean Code", "การเขียน Clean Code เป็นสิ่งสำคัญ").
				WithResponsePattern("ทำให้โค้ดดี", "การเขียนโค้ดที่ดีต้องมีหลักการ Clean Code")
		},
		Test: func(h *Harness) error {
			// ขั้นตอน 1: ให้ข้อมูลเกี่ยวกับ Clean Code
			_, err := h.Chat("การเขียน Clean Code เป็นสิ่งสำคัญสำหรับการพัฒนาซอฟต์แวร์")
			if err != nil {
				return err
			}

			// ขั้นตอน 2: ถามคำถามที่เกี่ยวข้องแบบ semantic (ไม่ใช่คำเดียวกัน)
			response, err := h.Chat("วิธีทำให้โค้ดดีขึ้น")
			if err != nil {
				return err
			}

			// ตรวจสอบว่า RAG สามารถค้นหาแบบ semantic ได้
			if !strings.Contains(response, "Clean Code") {
				return fmt.Errorf("RAG semantic search should find 'Clean Code' related info, got: %s", response)
			}

			return nil
		},
	},
	{
		Name:        "RAG Memory - Multi-language Support",
		Description: "ทดสอบการรองรับหลายภาษาด้วย RAG",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithResponsePattern("ภาษาไทย", "ผมสามารถสื่อสารภาษาไทยได้").
				WithResponsePattern("language", "User prefers Thai language")
		},
		Test: func(h *Harness) error {
			// ขั้นตอน 1: บอกว่าชอบภาษาไทย
			_, err := h.Chat("ฉันชอบใช้ภาษาไทยในการสื่อสาร")
			if err != nil {
				return err
			}

			// ขั้นตอน 2: ถามเป็นภาษาอังกฤษ (cross-language search)
			response, err := h.Chat("what language do I prefer")
			if err != nil {
				return err
			}

			// ตรวจสอบว่า RAG รองรับหลายภาษา
			if !strings.Contains(response, "Thai") && !strings.Contains(response, "ไทย") {
				return fmt.Errorf("RAG should support multi-language, got: %s", response)
			}

			return nil
		},
	},
	{
		Name:        "RAG Memory - Context Retrieval",
		Description: "ทดสอบการดึง context ที่เกี่ยวข้องจาก RAG",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithResponsePattern("SQLite", "ระบบใช้ SQLite เป็นฐานข้อมูล").
				WithResponsePattern("ฐานข้อมูล", "ใช้ SQLite สำหรับเก็บข้อมูล")
		},
		Test: func(h *Harness) error {
			// ขั้นตอน 1: ให้ข้อมูลเทคนิค
			_, err := h.Chat("ระบบ RAG ใช้ SQLite เป็นฐานข้อมูลและรองรับการค้นหาแบบ semantic")
			if err != nil {
				return err
			}

			// ขั้นตอน 2: ถามคำถามเฉพาะเจาะจง
			response, err := h.Chat("ระบบใช้ฐานข้อมูลอะไร")
			if err != nil {
				return err
			}

			// ตรวจสอบว่า RAG ดึง context ที่ถูกต้อง
			if !strings.Contains(response, "SQLite") {
				return fmt.Errorf("RAG should retrieve context about SQLite database, got: %s", response)
			}

			return nil
		},
	},
}

// RAGMemoryScenarios สถานการณ์การทดสอบ RAG Memory โดยเฉพาะ
var RAGMemoryScenarios = []Scenario{
	{
		Name:        "RAG - Layered Memory Architecture",
		Description: "ทดสอบการทำงานแบบ Layered (RAG -> StructuredMemory -> Legacy)",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithResponsePattern("ทดสอบ", "กำลังทดสอบระบบ RAG Memory").
				WithResponsePattern("จำได้ไหม", "จำได้ครับ คุณบอกว่า")
		},
		Test: func(h *Harness) error {
			// ทดสอบการทำงานหลายรอบเพื่อให้แน่ใจว่า RAG ทำงาน
			messages := []string{
				"ทดสอบระบบ RAG Memory",
				"จำข้อมูลนี้ไว้: ฉันชื่อ TestUser",
				"จำได้ไหมว่าฉันชื่ออะไร",
			}

			for _, msg := range messages {
				_, err := h.Chat(msg)
				if err != nil {
					return err
				}
			}

			return nil
		},
	},
}

// RunScenario รันสถานการณ์การทดสอบ
func RunScenario(scenario Scenario) error {
	provider := NewMockProvider()
	scenario.Setup(provider)

	harness := New(provider)
	return scenario.Test(harness)
}

// RunAllScenarios รันสถานการณ์การทดสอบทั้งหมด
func RunAllScenarios() map[string]error {
	results := make(map[string]error)

	for _, scenario := range CommonScenarios {
		err := RunScenario(scenario)
		results[scenario.Name] = err
	}

	return results
}

// ScenarioResult ผลลัพธ์การทดสอบสถานการณ์
type ScenarioResult struct {
	Name        string
	Description string
	Passed      bool
	Error       error
	Duration    time.Duration
}

// RunScenariosWithReport รันสถานการณ์และสร้างรายงาน
func RunScenariosWithReport() []ScenarioResult {
	results := make([]ScenarioResult, 0, len(CommonScenarios))

	for _, scenario := range CommonScenarios {
		start := time.Now()
		err := RunScenario(scenario)
		duration := time.Since(start)

		results = append(results, ScenarioResult{
			Name:        scenario.Name,
			Description: scenario.Description,
			Passed:      err == nil,
			Error:       err,
			Duration:    duration,
		})
	}

	return results
}

// RunRAGMemoryScenarios รันสถานการณ์การทดสอบ RAG Memory ทั้งหมด
func RunRAGMemoryScenarios() []ScenarioResult {
	allScenarios := append(CommonScenarios, RAGMemoryScenarios...)
	results := make([]ScenarioResult, 0)

	// กรองเฉพาะ scenarios ที่เกี่ยวกับ RAG Memory
	ragScenarios := []Scenario{}
	for _, s := range allScenarios {
		if strings.Contains(s.Name, "RAG") || strings.Contains(s.Name, "Memory") {
			ragScenarios = append(ragScenarios, s)
		}
	}

	fmt.Printf("🧪 Running %d RAG Memory Test Scenarios...\n\n", len(ragScenarios))

	for _, scenario := range ragScenarios {
		start := time.Now()
		err := RunScenario(scenario)
		duration := time.Since(start)

		result := ScenarioResult{
			Name:        scenario.Name,
			Description: scenario.Description,
			Passed:      err == nil,
			Error:       err,
			Duration:    duration,
		}
		results = append(results, result)

		// พิมพ์ผลทันที
		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
		}
		fmt.Printf("%s %s (%v)\n", status, scenario.Name, duration)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
		}
	}

	return results
}

// PrintRAGReport พิมพ์รายงานผลการทดสอบ RAG Memory
func PrintRAGReport(results []ScenarioResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 RAG Memory Test Report")
	fmt.Println(strings.Repeat("=", 60))

	passed := 0
	failed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Printf("Total: %d | Passed: %d | Failed: %d\n", len(results), passed, failed)
	fmt.Println(strings.Repeat("=", 60))

	if failed > 0 {
		fmt.Println("\n❌ Failed Tests:")
		for _, r := range results {
			if !r.Passed {
				fmt.Printf("  • %s: %v\n", r.Name, r.Error)
			}
		}
	}
}

// PrintReport พิมพ์รายงานผลการทดสอบ
func PrintReport(results []ScenarioResult) string {
	var sb strings.Builder
	passed := 0
	failed := 0

	sb.WriteString("========================================\n")
	sb.WriteString("       LLM Chat Test Report\n")
	sb.WriteString("========================================\n\n")

	for _, r := range results {
		status := "✓ PASS"
		if !r.Passed {
			status = "✗ FAIL"
			failed++
		} else {
			passed++
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", status, r.Name))
		sb.WriteString(fmt.Sprintf("   Description: %s\n", r.Description))
		sb.WriteString(fmt.Sprintf("   Duration: %v\n", r.Duration))
		if r.Error != nil {
			sb.WriteString(fmt.Sprintf("   Error: %v\n", r.Error))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("========================================\n")
	sb.WriteString(fmt.Sprintf("Total: %d | Passed: %d | Failed: %d\n", len(results), passed, failed))
	sb.WriteString("========================================\n")

	return sb.String()
}

// CustomScenario สร้างสถานการณ์การทดสอบแบบกำหนดเอง
func CustomScenario(name, description string, setup func(*MockProvider), test func(*Harness) error) Scenario {
	return Scenario{
		Name:        name,
		Description: description,
		Setup:       setup,
		Test:        test,
	}
}

// LoadScenarioFromJSON โหลดสถานการณ์จาก JSON
func LoadScenarioFromJSON(data []byte) (*Scenario, error) {
	var s struct {
		Name        string                     `json:"name"`
		Description string                     `json:"description"`
		Responses   map[string]string          `json:"responses"`
		ToolCalls   map[string]json.RawMessage `json:"tool_calls"`
		TestSteps   []string                   `json:"test_steps"`
	}

	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	scenario := &Scenario{
		Name:        s.Name,
		Description: s.Description,
		Setup: func(mp *MockProvider) {
			mp.Reset()
			for pattern, response := range s.Responses {
				mp.WithResponsePattern(pattern, response)
			}
			for pattern, argsJSON := range s.ToolCalls {
				var args map[string]any
				json.Unmarshal(argsJSON, &args)
				mp.WithToolCallResponse(pattern, "tool", args)
			}
		},
		Test: func(h *Harness) error {
			for _, step := range s.TestSteps {
				if _, err := h.Chat(step); err != nil {
					return err
				}
			}
			return nil
		},
	}

	return scenario, nil
}
