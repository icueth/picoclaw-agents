# LLM Chat Test Harness

ระบบทดสอบการสนทนากับ LLM แบบจำลองสมจริง สำหรับ PicoClaw

## คุณสมบัติ

- **Mock LLM Provider** - จำลองการตอบสนองของ LLM ได้อย่างยืดหยุ่น
- **Conversation Simulator** - จำลองการสนทนาแบบ multi-turn ได้
- **Tool Call Simulation** - ทดสอบการเรียกใช้งาน tools ได้
- **Response Scenarios** - กำหนดสถานการณ์การตอบสนองได้หลากหลาย
- **Real-time Streaming** - จำลองการสตรีมข้อความแบบ realtime

## การใช้งาน

```go
// สร้าง mock provider ที่ตอบสนองตามสถานการณ์ที่กำหนด
provider := testharness.NewMockProvider().
    WithResponsePattern("สวัสดี", "สวัสดีครับ! มีอะไรให้ช่วยไหมครับ").
    WithToolCallPattern("ค้นหา", "web_search", map[string]any{"query": "..."})

// สร้าง test harness
harness := testharness.New(provider)

// เริ่มการสนทนา
response, err := harness.Chat("สวัสดี")
```

## โครงสร้างไฟล์

- `mock_provider.go` - Mock LLM Provider
- `harness.go` - Test Harness หลัก
- `scenarios.go` - สถานการณ์การทดสอบต่างๆ
- `conversation.go` - ตัวจัดการการสนทนา
- `assertions.go` - ตัวช่วยตรวจสอบผลลัพธ์
