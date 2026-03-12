# LLM Chat Test Harness - วิธีใช้งาน

ระบบทดสอบการสนทนากับ LLM สำหรับ PicoClaw ที่รองรับทั้ง Mock Provider และ Real LLM

## โครงสร้างไฟล์

```
pkg/testharness/
├── README.md                    # คำอธิบายโครงการ
├── USAGE.md                     # ไฟล์นี้ - วิธีใช้งาน
├── mock_provider.go             # Mock LLM Provider สำหรับการทดสอบ
├── harness.go                   # Test Harness พื้นฐาน
├── scenarios.go                 # สถานการณ์การทดสอบสำเร็จรูป
├── interactive.go               # Interactive Chat Simulator (Mock)
├── real_llm_harness.go          # Real LLM Test Harness
├── real_llm_interactive.go      # Interactive Tester สำหรับ LLM จริง
├── real_llm_test_cases.go       # Test Cases สำหรับ LLM จริง
├── example_test.go              # ตัวอย่างการใช้งาน
└── cmd/tester/main.go           # CLI tool สำหรับทดสอบ
```

## การใช้งานกับ LLM จริง

### 1. ทดสอบอย่างรวดเร็ว (Quick Test)

```go
package main

import (
    "fmt"
    "log"
    "picoclaw/agent/pkg/testharness"
)

func main() {
    // ทดสอบด้วยข้อความเดียว
    response, err := testharness.QuickTest(
        "/Users/icue/.picoclaw/config.json",
        "สวัสดี คุณชื่ออะไร",
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(response)
}
```

### 2. ทดสอบแบบ Batch

```go
func main() {
    testCases := []string{
        "สวัสดี",
        "คุณชื่ออะไร",
        "เขียนโค้ด Python สำหรับเรียงลำดับตัวเลข",
        "อธิบายเรื่อง Machine Learning",
    }

    results, err := testharness.RunBatchTest(
        "/Users/icue/.picoclaw/config.json",
        testCases,
    )
    if err != nil {
        log.Fatal(err)
    }

    for _, result := range results {
        fmt.Printf("Input: %s\n", result.Input)
        fmt.Printf("Response: %s\n", result.Response)
        fmt.Printf("Latency: %v\n\n", result.Latency)
    }
}
```

### 3. ทดสอบแบบ Interactive

```go
func main() {
    tester, err := testharness.NewRealLLMInteractiveTester(
        "/Users/icue/.picoclaw/config.json",
    )
    if err != nil {
        log.Fatal(err)
    }

    // เริ่มการทดสอบ interactive
    tester.Start()
}
```

หรือใช้งานผ่าน command line:

```bash
# Interactive mode
picoclaw-tester -mode=interactive

# Quick test
picoclaw-tester -mode=quick -message="สวัสดี"

# Batch test
picoclaw-tester -mode=batch "ข้อความที่ 1" "ข้อความที่ 2" "ข้อความที่ 3"

# Run scenarios
picoclaw-tester -mode=scenario
```

### 4. ทดสอบด้วย Test Harness โดยตรง

```go
func main() {
    // สร้าง harness
    harness, err := testharness.NewRealLLMTestHarness(
        "/Users/icue/.picoclaw/config.json",
    )
    if err != nil {
        log.Fatal(err)
    }

    // ตั้งค่า parameters
    harness.WithTemperature(0.7).
        WithMaxTokens(2048).
        WithTimeout(60 * time.Second)

    // ทดสอบการสนทนา
    response, err := harness.Chat("สวัสดี")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(response)

    // แสดง metrics
    metrics := harness.GetMetrics()
    fmt.Printf("Total calls: %d, Avg latency: %v\n",
        metrics.TotalCalls, metrics.AvgLatency)

    // บันทึกการสนทนา
    harness.SaveConversation("test_conversation.json")
}
```

### 5. ทดสอบแบบ Streaming

```go
func main() {
    harness, err := testharness.NewRealLLMTestHarness(
        "/Users/icue/.picoclaw/config.json",
    )
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
```

### 6. ทดสอบหลาย Turn

```go
func main() {
    harness, err := testharness.NewRealLLMTestHarness(
        "/Users/icue/.picoclaw/config.json",
    )
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
```

### 7. รันชุดการทดสอบสำเร็จรูป

```go
func main() {
    tester, err := testharness.NewRealLLMInteractiveTester(
        "/Users/icue/.picoclaw/config.json",
    )
    if err != nil {
        log.Fatal(err)
    }

    // รันชุดการทดสอบทั้งหมด
    results := testharness.RunAllRealTestCases(tester)

    // แสดงรายงาน
    report := testharness.PrintRealTestReport(results)
    fmt.Println(report)

    // แสดงสรุปตามหมวดหมู่
    summary := testharness.CategorySummary(results)
    for cat, s := range summary {
        fmt.Printf("%s: %d/%d passed (avg latency: %v)\n",
            cat, s.Passed, s.Total, s.AvgLatency)
    }
}
```

## การใช้งานกับ Mock Provider (สำหรับ Unit Test)

```go
func TestMyFeature(t *testing.T) {
    // สร้าง mock provider
    provider := testharness.NewMockProvider().
        WithResponsePattern("สวัสดี", "สวัสดีครับ!").
        WithToolCallResponse("ค้นหา", "web_search", map[string]any{
            "query": "test",
        })

    // สร้าง harness
    harness := testharness.New(provider)

    // ทดสอบ
    response, err := harness.Chat("สวัสดี")
    if err != nil {
        t.Fatal(err)
    }

    // ตรวจสอบผลลัพธ์
    if err := harness.AssertResponseContains("สวัสดี"); err != nil {
        t.Error(err)
    }

    // ตรวจสอบว่ามีการเรียก tool
    if err := harness.AssertToolCalled("web_search"); err != nil {
        t.Error(err)
    }
}
```

## คำสั่งใน Interactive Mode

เมื่ออยู่ใน interactive mode คุณสามารถใช้คำสั่งต่อไปนี้:

- `/help` หรือ `/h` - แสดงวิธีใช้
- `/history` หรือ `/hist` - แสดงประวัติการสนทนา
- `/clear` หรือ `/c` - ล้างประวัติการสนทนา
- `/reset` หรือ `/r` - รีเซ็ตสถานะ
- `/metrics` หรือ `/m` - แสดง metrics
- `/model` - แสดงชื่อ model ปัจจุบัน
- `/config` - แสดง configuration
- `/temp <value>` - ตั้งค่า temperature
- `/tokens <value>` - ตั้งค่า max tokens
- `/save` - บันทึกการสนทนา
- `/debug` - เปิด/ปิด debug mode
- `/exit` หรือ `/quit` - ออกจากโปรแกรม

## การสร้าง Test Case ที่กำหนดเอง

```go
// สร้าง custom test case
testCase := testharness.RealTestCase{
    Name:        "My Custom Test",
    Description: "ทดสอบการตอบสนองเฉพาะ",
    Category:    "custom",
    Inputs:      []string{"คำถามของฉัน"},
    Validators: []testharness.ResponseValidator{
        testharness.NotEmpty(),
        testharness.Contains("คำตอบที่ต้องการ"),
        testharness.MinLength(50),
    },
}

// รัน test
tester, _ := testharness.NewRealLLMInteractiveTester(configPath)
result, _ := testharness.RunRealTestCase(tester, testCase)

if result.IsPassed() {
    fmt.Println("Test passed!")
}
```

## หมายเหตุ

- ระบบนี้ใช้ config จาก `~/.picoclaw/config.json` ตามค่าเริ่มต้น
- สามารถระบุ path ของ config ได้เอง
- รองรับการทดสอบกับ model ต่างๆ ที่ configured ไว้ใน config
- บันทึกผลการทดสอบได้ในรูปแบบ JSON
