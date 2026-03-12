// PicoClaw - Test Harness Examples
// ตัวอย่างการใช้งานระบบทดสอบ

package testharness

import (
	"fmt"
	"log"
	"time"
)

// ExampleMockProvider ตัวอย่างการใช้งาน Mock Provider
func ExampleMockProvider() {
	// สร้าง mock provider ที่ตอบสนองตาม pattern
	provider := NewMockProvider().
		WithResponsePattern("สวัสดี", "สวัสดีครับ! มีอะไรให้ช่วยเหลือไหมครับ").
		WithResponsePattern("ชื่ออะไร", "ผมชื่อ PicoClaw ครับ").
		WithToolCallResponse("ค้นหา", "web_search", map[string]any{
			"query": "golang tutorial",
		})

	// สร้าง test harness
	harness := New(provider)

	// ทดสอบการสนทนา
	response, err := harness.Chat("สวัสดี")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response: %s\n", response)

	// ทดสอบการเรียก tool
	_, _ = harness.Chat("ช่วยค้นหา golang tutorial")
	if harness.AssertToolCalled("web_search") != nil {
		fmt.Println("Tool was called!")
	}
}

// ExampleRealLLMTestHarness ตัวอย่างการใช้งานกับ LLM จริง
func ExampleRealLLMTestHarness() {
	// สร้าง test harness จาก config
	harness, err := NewRealLLMTestHarness("/Users/icue/.picoclaw/config.json")
	if err != nil {
		log.Fatal(err)
	}

	// ตั้งค่า parameters
	harness.WithTemperature(0.7).
		WithMaxTokens(2048).
		WithTimeout(60 * time.Second)

	// ทดสอบการสนทนา
	response, err := harness.Chat("สวัสดี คุณชื่ออะไร")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response: %s\n", response)

	// แสดง metrics
	metrics := harness.GetMetrics()
	fmt.Printf("Total calls: %d, Avg latency: %v\n", metrics.TotalCalls, metrics.AvgLatency)

	// บันทึกการสนทนา
	harness.SaveConversation("test_conversation.json")
}

// ExampleRealLLMInteractiveTester ตัวอย่างการใช้งาน Interactive Tester
func ExampleRealLLMInteractiveTester() {
	// สร้าง interactive tester
	tester, err := NewRealLLMInteractiveTester("/Users/icue/.picoclaw/config.json")
	if err != nil {
		log.Fatal(err)
	}

	// เริ่มการทดสอบ interactive
	tester.Start()
}

// ExampleRunBatchTest ตัวอย่างการทดสอบแบบ batch
func ExampleRunBatchTest() {
	// สร้าง test cases
	testCases := []string{
		"สวัสดี",
		"คุณชื่ออะไร",
		"เขียนโค้ด Python สำหรับเรียงลำดับตัวเลข",
		"อธิบายเรื่อง Machine Learning",
	}

	// รัน batch test
	results, err := RunBatchTest("/Users/icue/.picoclaw/config.json", testCases)
	if err != nil {
		log.Fatal(err)
	}

	// แสดงผลลัพธ์
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("❌ %s: %v\n", result.Input, result.Error)
		} else {
			fmt.Printf("✅ %s: %v (%v)\n", result.Input, result.Response[:50], result.Latency)
		}
	}
}

// ExampleQuickTest ตัวอย่างการทดสอบอย่างรวดเร็ว
func ExampleQuickTest() {
	response, err := QuickTest("/Users/icue/.picoclaw/config.json", "สวัสดี")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response: %s\n", response)
}

// ExampleRunScenariosWithReport ตัวอย่างการใช้งาน Scenarios
func ExampleRunScenariosWithReport() {
	// รันสถานการณ์การทดสอบทั้งหมด
	results := RunScenariosWithReport()

	// แสดงรายงาน
	report := PrintReport(results)
	fmt.Println(report)
}

// ExampleCustomScenario ตัวอย่างการสร้างสถานการณ์ที่กำหนดเอง
func ExampleCustomScenario() {
	// สร้าง custom scenario
	scenario := CustomScenario(
		"My Custom Test",
		"ทดสอบการตอบสนองเฉพาะ",
		func(mp *MockProvider) {
			mp.Reset().
				WithExactMatch("test", "This is a test response").
				WithResponsePattern("error", "Something went wrong")
		},
		func(h *Harness) error {
			_, err := h.Chat("test")
			if err != nil {
				return err
			}
			return h.AssertResponseEquals("This is a test response")
		},
	)

	// รัน scenario
	err := RunScenario(scenario)
	if err != nil {
		fmt.Printf("Test failed: %v\n", err)
	}
}

// ExampleRunAllRealTestCases ตัวอย่างการใช้งาน Real Test Cases
func ExampleRunAllRealTestCases() {
	// สร้าง tester
	tester, err := NewRealLLMInteractiveTester("/Users/icue/.picoclaw/config.json")
	if err != nil {
		log.Fatal(err)
	}

	// รันชุดการทดสอบทั้งหมด
	results := RunAllRealTestCases(tester)

	// แสดงรายงาน
	report := PrintRealTestReport(results)
	fmt.Println(report)

	// แสดงสรุปตามหมวดหมู่
	summary := CategorySummary(results)
	fmt.Println("\nCategory Summary:")
	for cat, s := range summary {
		fmt.Printf("  %s: %d/%d passed (avg latency: %v)\n", cat, s.Passed, s.Total, s.AvgLatency)
	}
}

// ExampleRealLLMTestHarness_ChatStreaming ตัวอย่างการใช้งาน Streaming
func ExampleRealLLMTestHarness_ChatStreaming() {
	harness, err := NewRealLLMTestHarness("/Users/icue/.picoclaw/config.json")
	if err != nil {
		log.Fatal(err)
	}

	// ทดสอบการ streaming
	fmt.Print("Response: ")
	err = harness.ChatStreaming("เขียนเรื่องสั้น 3 ย่อหน้า", func(chunk string) {
		fmt.Print(chunk)
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()
}

// ExampleRealLLMTestHarness_MultiTurnChat ตัวอย่างการสนทนาหลาย turn
func ExampleRealLLMTestHarness_MultiTurnChat() {
	harness, err := NewRealLLMTestHarness("/Users/icue/.picoclaw/config.json")
	if err != nil {
		log.Fatal(err)
	}

	// สนทนาหลาย turn
	messages := []string{
		"ฉันชื่อ TestUser",
		"ชื่อฉันคืออะไร",
		"ฉันชอบสีน้ำเงิน",
		"ฉันชอบสีอะไร",
	}

	responses, err := harness.MultiTurnChat(messages)
	if err != nil {
		log.Fatal(err)
	}

	for i, resp := range responses {
		fmt.Printf("Turn %d: %s\n", i+1, resp)
	}
}

// ExampleResponseValidator ตัวอย่างการสร้าง validator ที่กำหนดเอง
func ExampleResponseValidator() {
	// สร้าง custom validator
	customValidator := func(response string) error {
		if len(response) < 10 {
			return fmt.Errorf("response too short")
		}
		if !containsAny(response, []string{"กรุงเทพ", "Bangkok", "Thailand"}) {
			return fmt.Errorf("response does not mention Thailand")
		}
		return nil
	}

	// สร้าง test case ที่ใช้ custom validator
	testCase := RealTestCase{
		Name:        "Thailand Test",
		Description: "ทดสอบการตอบสนองเกี่ยวกับประเทศไทย",
		Category:    "custom",
		Inputs:      []string{"อธิบายเกี่ยวกับประเทศไทย"},
		Validators:  []ResponseValidator{customValidator},
	}

	// รัน test
	tester, _ := NewRealLLMInteractiveTester("/Users/icue/.picoclaw/config.json")
	result, _ := RunRealTestCase(tester, testCase)

	if result.IsPassed() {
		fmt.Println("Test passed!")
	} else {
		fmt.Printf("Test failed: %d errors, %d validation failures\n",
			result.Errors, result.ValidationFailures)
	}
}

// Helper function
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if contains(s, substr) {
			return true
		}
	}
	return false
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
