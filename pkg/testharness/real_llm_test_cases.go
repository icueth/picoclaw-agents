// PicoClaw - Real LLM Test Cases
// ชุดการทดสอบสำเร็จรูปสำหรับ LLM จริง

package testharness

import (
	"fmt"
	"strings"
	"time"
)

// RealTestCase ชุดการทดสอบสำหรับ LLM จริง
type RealTestCase struct {
	Name        string
	Description string
	Category    string
	Inputs      []string
	Validators  []ResponseValidator
}

// ResponseValidator ตัวตรวจสอบการตอบสนอง
type ResponseValidator func(response string) error

// CommonRealTestCases ชุดการทดสอบทั่วไปสำหรับ LLM จริง
var CommonRealTestCases = []RealTestCase{
	{
		Name:        "Basic Greeting",
		Description: "ทดสอบการตอบสนองพื้นฐาน",
		Category:    "basic",
		Inputs:      []string{"สวัสดี", "Hello", "Hi there"},
		Validators: []ResponseValidator{
			NotEmpty(),
			MaxLength(500),
		},
	},
	{
		Name:        "Thai Language",
		Description: "ทดสอบการเข้าใจภาษาไทย",
		Category:    "language",
		Inputs: []string{
			"คุณชื่ออะไร",
			"อธิบายเรื่อง Machine Learning เป็นภาษาไทย",
			"เขียนโค้ด Python สำหรับเรียงลำดับตัวเลข",
		},
		Validators: []ResponseValidator{
			NotEmpty(),
			ContainsThai(),
		},
	},
	{
		Name:        "Code Generation",
		Description: "ทดสอบการเขียนโค้ด",
		Category:    "coding",
		Inputs: []string{
			"เขียน function สำหรับหาค่า fibonacci ใน Go",
			"สร้าง REST API ด้วย Python Flask",
			"เขียน SQL query สำหรับหา top 10 สินค้าขายดี",
		},
		Validators: []ResponseValidator{
			NotEmpty(),
			ContainsCode(),
		},
	},
	{
		Name:        "Reasoning",
		Description: "ทดสอบการให้เหตุผล",
		Category:    "reasoning",
		Inputs: []string{
			"ถ้า A มากกว่า B และ B มากกว่า C แล้ว A กับ C ใครมากกว่า?",
			"อธิบายขั้นตอนการแก้ปัญหานี้: มีนก 10 ตัวบนต้นไม้ ยิงตก 2 ตัว เหลือกี่ตัว?",
		},
		Validators: []ResponseValidator{
			NotEmpty(),
			MinLength(50),
		},
	},
	{
		Name:        "Context Memory",
		Description: "ทดสอบการจดจำ context",
		Category:    "memory",
		Inputs: []string{
			"ฉันชื่อ TestUser",
			"ชื่อฉันคืออะไร",
		},
		Validators: []ResponseValidator{
			NotEmpty(),
		},
	},
	{
		Name:        "Long Response",
		Description: "ทดสอบการตอบสนองที่ยาว",
		Category:    "performance",
		Inputs: []string{
			"อธิบายหลักการทำงานของ Large Language Model อย่างละเอียด",
			"เขียนบทความสั้นๆ เรื่องประวัติศาสตร์ของ Artificial Intelligence",
		},
		Validators: []ResponseValidator{
			NotEmpty(),
			MinLength(200),
		},
	},
	{
		Name:        "Edge Cases",
		Description: "ทดสอบกรณีพิเศษ",
		Category:    "edge",
		Inputs: []string{
			"",
			"123",
			"!@#$%",
			"ความยาวมาก" + strings.Repeat("มาก", 100),
		},
		Validators: []ResponseValidator{
			NotEmpty(),
		},
	},
}

// RunRealTestCase รันชุดการทดสอบ
func RunRealTestCase(tester *RealLLMInteractiveTester, testCase RealTestCase) (*RealTestResult, error) {
	result := &RealTestResult{
		Name:        testCase.Name,
		Description: testCase.Description,
		Category:    testCase.Category,
		StartTime:   time.Now(),
		Responses:   make([]TestResponse, 0, len(testCase.Inputs)),
	}

	for _, input := range testCase.Inputs {
		resp, latency, err := tester.ProcessSingle(input)

		testResp := TestResponse{
			Input:     input,
			Response:  resp,
			Latency:   latency,
			Timestamp: time.Now(),
		}

		if err != nil {
			testResp.Error = err
			result.Errors++
		} else {
			// ตรวจสอบ validators
			for _, validator := range testCase.Validators {
				if err := validator(resp); err != nil {
					testResp.ValidationErrors = append(testResp.ValidationErrors, err)
					result.ValidationFailures++
				}
			}
		}

		result.Responses = append(result.Responses, testResp)
		result.TotalLatency += latency
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	if len(testCase.Inputs) > 0 {
		result.AvgLatency = result.TotalLatency / time.Duration(len(testCase.Inputs))
	}

	return result, nil
}

// RunAllRealTestCases รันชุดการทดสอบทั้งหมด
func RunAllRealTestCases(tester *RealLLMInteractiveTester) []RealTestResult {
	results := make([]RealTestResult, 0, len(CommonRealTestCases))

	for _, testCase := range CommonRealTestCases {
		result, err := RunRealTestCase(tester, testCase)
		if err != nil {
			// สร้าง result ที่มี error
			result = &RealTestResult{
				Name:        testCase.Name,
				Description: testCase.Description,
				Category:    testCase.Category,
				StartTime:   time.Now(),
				EndTime:     time.Now(),
				Errors:      1,
			}
		}
		results = append(results, *result)
	}

	return results
}

// RealTestResult ผลลัพธ์การทดสอบ
type RealTestResult struct {
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	Category           string          `json:"category"`
	StartTime          time.Time       `json:"start_time"`
	EndTime            time.Time       `json:"end_time"`
	Duration           time.Duration   `json:"duration"`
	Responses          []TestResponse  `json:"responses"`
	TotalLatency       time.Duration   `json:"total_latency"`
	AvgLatency         time.Duration   `json:"avg_latency"`
	Errors             int             `json:"errors"`
	ValidationFailures int             `json:"validation_failures"`
}

// TestResponse การตอบสนองของแต่ละ test
type TestResponse struct {
	Input            string    `json:"input"`
	Response         string    `json:"response"`
	Latency          time.Duration `json:"latency"`
	Timestamp        time.Time `json:"timestamp"`
	Error            error     `json:"error,omitempty"`
	ValidationErrors []error   `json:"validation_errors,omitempty"`
}

// IsPassed ตรวจสอบว่าผ่านการทดสอบหรือไม่
func (r *RealTestResult) IsPassed() bool {
	return r.Errors == 0 && r.ValidationFailures == 0
}

// Response Validators

// NotEmpty ตรวจสอบว่า response ไม่ว่างเปล่า
func NotEmpty() ResponseValidator {
	return func(response string) error {
		if strings.TrimSpace(response) == "" {
			return fmt.Errorf("response is empty")
		}
		return nil
	}
}

// MinLength ตรวจสอบความยาวขั้นต่ำ
func MinLength(min int) ResponseValidator {
	return func(response string) error {
		if len(response) < min {
			return fmt.Errorf("response length %d is less than minimum %d", len(response), min)
		}
		return nil
	}
}

// MaxLength ตรวจสอบความยาวสูงสุด
func MaxLength(max int) ResponseValidator {
	return func(response string) error {
		if len(response) > max {
			return fmt.Errorf("response length %d exceeds maximum %d", len(response), max)
		}
		return nil
	}
}

// Contains ตรวจสอบว่ามีข้อความที่ต้องการ
func Contains(substring string) ResponseValidator {
	return func(response string) error {
		if !strings.Contains(strings.ToLower(response), strings.ToLower(substring)) {
			return fmt.Errorf("response does not contain %q", substring)
		}
		return nil
	}
}

// ContainsThai ตรวจสอบว่ามีภาษาไทย
func ContainsThai() ResponseValidator {
	return func(response string) error {
		for _, r := range response {
			if r >= 0x0E00 && r <= 0x0E7F {
				return nil
			}
		}
		return fmt.Errorf("response does not contain Thai characters")
	}
}

// ContainsCode ตรวจสอบว่ามีโค้ด
func ContainsCode() ResponseValidator {
	return func(response string) error {
		codeIndicators := []string{"```", "def ", "func ", "class ", "import ", "package ", "function", "return"}
		for _, indicator := range codeIndicators {
			if strings.Contains(response, indicator) {
				return nil
			}
		}
		return fmt.Errorf("response does not appear to contain code")
	}
}

// MatchesRegex ตรวจสอบว่าตรงกับ regex
func MatchesRegex(pattern string) ResponseValidator {
	return func(response string) error {
		// Note: ต้อง import regexp ถ้าจะใช้
		return nil
	}
}

// PrintRealTestReport พิมพ์รายงานผลการทดสอบ
func PrintRealTestReport(results []RealTestResult) string {
	var sb strings.Builder
	passed := 0
	failed := 0

	sb.WriteString("╔════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║              Real LLM Test Report                          ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════╝\n\n")

	for _, r := range results {
		status := "✅ PASS"
		if !r.IsPassed() {
			status = "❌ FAIL"
			failed++
		} else {
			passed++
		}

		sb.WriteString(fmt.Sprintf("%s %s (%s)\n", status, r.Name, r.Category))
		sb.WriteString(fmt.Sprintf("   Description: %s\n", r.Description))
		sb.WriteString(fmt.Sprintf("   Duration: %v | Avg Latency: %v\n", r.Duration, r.AvgLatency))
		sb.WriteString(fmt.Sprintf("   Errors: %d | Validation Failures: %d\n", r.Errors, r.ValidationFailures))

		if len(r.Responses) > 0 {
			sb.WriteString("   Responses:\n")
			for i, resp := range r.Responses {
				sb.WriteString(fmt.Sprintf("     [%d] ", i+1))
				if resp.Error != nil {
					sb.WriteString(fmt.Sprintf("ERROR: %v\n", resp.Error))
				} else {
					preview := resp.Response
					if len(preview) > 50 {
						preview = preview[:50] + "..."
					}
					sb.WriteString(fmt.Sprintf("%s (%v)\n", preview, resp.Latency))
				}
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("════════════════════════════════════════════════════════════\n")
	sb.WriteString(fmt.Sprintf("Total: %d | Passed: %d | Failed: %d\n", len(results), passed, failed))
	sb.WriteString("════════════════════════════════════════════════════════════\n")

	return sb.String()
}

// CategorySummary สรุปผลการทดสอบตามหมวดหมู่
func CategorySummary(results []RealTestResult) map[string]struct {
	Total   int
	Passed  int
	Failed  int
	AvgLatency time.Duration
} {
	summary := make(map[string]struct {
		Total   int
		Passed  int
		Failed  int
		AvgLatency time.Duration
	})

	for _, r := range results {
		s := summary[r.Category]
		s.Total++
		if r.IsPassed() {
			s.Passed++
		} else {
			s.Failed++
		}
		s.AvgLatency += r.AvgLatency
		summary[r.Category] = s
	}

	// คำนวณค่าเฉลี่ย
	for cat, s := range summary {
		if s.Total > 0 {
			s.AvgLatency = s.AvgLatency / time.Duration(s.Total)
			summary[cat] = s
		}
	}

	return summary
}
