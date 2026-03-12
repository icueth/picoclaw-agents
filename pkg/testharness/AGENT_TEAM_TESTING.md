# Agent Team Testing Guide

เอกสารนี้อธิบายวิธีการทดสอบ Agent-to-Agent (A2A) communication, subagent spawning, meeting system, และ mailbox functionality ผ่าน Test Harness

## Overview

Test Harness ใน `pkg/testharness` ถูกออกแบบมาเพื่อจำลองการทำงานของ Agent Team โดยไม่ต้องเรียก LLM จริง ทำให้ทดสอบได้รวดเร็วและควบคุมผลลัพธ์ได้

## Core Concepts

### Spawn Subagent vs Agent-to-Agent

| Feature | Spawn Subagent | Agent-to-Agent (A2A) |
|---------|----------------|---------------------|
| ลักษณะ | สร้าง temporary helper | สื่อสารกับ existing agents |
| บุคลิก | ไม่มี persona เป็นพิเศษ | มี persona ใน `~/.picoclaw/agents/{id}/` |
| ใช้เมื่อ | ต้องการ help ชั่วคราว | ต้องการ agent เฉพาะทาง |
| Tools | `spawn_subagent`, `subagent_status` | `start_meeting`, `send_agent_message`, `check_agent_inbox` |

## Available Tools

### 1. Spawn Subagent
```go
{
    Name: "spawn_subagent",
    Parameters: {
        "role": string,     // architect, coder, researcher, qa, writer, reviewer, coordinator
        "task": string,     // คำอธิบายงาน
        "context": string,  // context เพิ่มเติม (optional)
    }
}
```

### 2. Subagent Status
```go
{
    Name: "subagent_status",
    Parameters: {
        "action": string,   // "get", "active", "list"
        "task_id": string,  // สำหรับ action "get"
    }
}
```

### 3. Start Meeting
```go
{
    Name: "start_meeting",
    Parameters: {
        "topic": string,        // หัวข้อการประชุม
        "participants": []string, // รายชื่อ agent IDs
        "duration": int,        // ระยะเวลา (นาที)
    }
}
```

### 4. Send Agent Message
```go
{
    Name: "send_agent_message",
    Parameters: {
        "to": string,       // ผู้รับ
        "subject": string,  // หัวข้อ
        "message": string,  // เนื้อหา
        "priority": string, // "low", "normal", "high", "urgent"
    }
}
```

### 5. Check Agent Inbox
```go
{
    Name: "check_agent_inbox",
    Parameters: {
        "unread_only": bool, // ดูเฉพาะ unread
        "limit": int,        // จำกัดจำนวน
    }
}
```

## Test Scenarios

### File Structure
```
pkg/testharness/
├── agent_team_scenarios.go    # 9 predefined scenarios
├── agent_team_test.go         # Test runners
├── example_agent_team_test.go # Example usage
└── AGENT_TEAM_TESTING.md      # This file
```

### Available Scenarios

1. **Agent to Agent - Simple Delegation**: Jarvis สั่งงาน Nova
2. **Jarvis Coordinator - Multi Agent**: เรียกหลาย agents พร้อมกัน
3. **Agent Meeting - Discussion**: จัดประชุมหลาย agents
4. **Agent Mailbox - Send Message**: ส่งข้อความระหว่าง agents
5. **Agent Mailbox - Check Inbox**: ตรวจสอบกล่องจดหมาย
6. **Agent Collaboration - Research then Code**: Atlas วิจัย → Clawed เขียนโค้ด
7. **Error - Agent Not Found**: ทดสอบ error handling
8. **Agent Status Check**: ตรวจสอบสถานะ agents
9. **Complex Workflow - Full Project**: Workflow ครบวงจร 5 agents

## Usage Examples

### Basic Test
```go
func TestSimpleDelegation(t *testing.T) {
    provider := NewMockProvider()
    provider.WithToolCallResponse(
        "Nova",
        "spawn_subagent",
        map[string]any{
            "role": "architect",
            "task": "ออกแบบระบบ",
        },
    )

    harness := New(provider)
    harness.Chat("ให้ Nova ออกแบบระบบหน่อย")
    
    if !provider.VerifyToolCall("spawn_subagent") {
        t.Error("Expected spawn_subagent to be called")
    }
}
```

### Multi-Agent Workflow
```go
func TestMultiAgent(t *testing.T) {
    provider := NewMockProvider()
    
    // Setup responses for each agent
    provider.WithToolCallResponse("Atlas", "spawn_subagent", 
        map[string]any{"role": "researcher"})
    provider.WithToolCallResponse("Nova", "spawn_subagent", 
        map[string]any{"role": "architect"})
    provider.WithToolCallResponse("Clawed", "spawn_subagent", 
        map[string]any{"role": "coder"})

    harness := New(provider)
    harness.Chat("ให้ Atlas วิจัย, Nova ออกแบบ, Clawed เขียนโค้ด")
    
    count := provider.GetToolCallCount("spawn_subagent")
    fmt.Printf("Spawned %d subagents\n", count)
}
```

### Using Predefined Scenarios
```go
func TestAgentTeamScenarios(t *testing.T) {
    for _, scenario := range AgentTeamScenarios {
        t.Run(scenario.Name, func(t *testing.T) {
            provider := NewMockProvider()
            scenario.Setup(provider)
            harness := New(provider)
            
            if err := scenario.Test(harness); err != nil {
                t.Errorf("Scenario %q failed: %v", scenario.Name, err)
            }
        })
    }
}
```

## Mock Provider API

### Setup Methods
- `WithResponsePattern(pattern, response)` - ตอบสนองตาม pattern
- `WithToolCallResponse(pattern, toolName, args)` - ตอบสนองด้วย tool call
- `WithSystemPrompt(prompt)` - ตั้งค่า system prompt
- `WithToolError(toolName, err)` - จำลอง tool error

### Verification Methods
- `VerifyToolCall(toolName)` - ตรวจสอบว่า tool ถูกเรียก
- `GetToolCallCount(toolName)` - นับจำนวนการเรียก tool
- `VerifyCall(pattern)` - ตรวจสอบข้อความที่ส่ง
- `AssertNoErrors()` - ตรวจสอบว่าไม่มี error

## Agent Roles Reference

| Agent | Role | Avatar | ความสามารถ |
|-------|------|--------|-----------|
| Jarvis | coordinator | 🤖 | ประสานงาน, กระจายงาน |
| Nova | architect | 🔮 | ออกแบบสถาปัตยกรรม |
| Atlas | researcher | 📚 | วิจัย, ค้นหาข้อมูล |
| Clawed | coder | 🔧 | เขียนโค้ด |
| Sentinel | qa | 🛡️ | ตรวจสอบคุณภาพ |
| Scribe | writer | 📝 | เขียนเอกสาร |
| Trendy | reviewer | 🔍 | Review code |
| Pixel | designer | 🎨 | ออกแบบ UI/UX |

## Running Tests

```bash
# Run all agent team tests
go test -v ./pkg/testharness/... -run "TestAgentTeam"

# Run specific test
go test -v ./pkg/testharness/... -run "TestSimpleDelegation"

# Run with coverage
go test -cover ./pkg/testharness/...

# Run benchmarks
go test -bench=. ./pkg/testharness/...
```

## Best Practices

1. **Isolate Tests**: แต่ละ test ควรมี provider ของตัวเอง
2. **Clear Patterns**: ใช้ pattern ที่เฉพาะเจาะจงใน `WithToolCallResponse`
3. **Verify Results**: ตรวจสอบทั้ง tool calls และ arguments
4. **Error Cases**: ทดสอบ error handling ด้วย `WithToolError`
5. **Documentation**: เพิ่ม example tests สำหรับ features ใหม่

## Integration with Real System

Test Harness ใช้ mock provider แต่สามารถเชื่อมต่อกับระบบจริงได้ผ่าน:

1. **Tool Definitions**: ใช้ tool definitions เดียวกับระบบจริง
2. **Provider Interface**: Implement `providers.LLMProvider` สำหรับ provider จริง
3. **Event Bus**: ใช้ `bus.MessageBus` เดียวกันกับระบบจริง

## Troubleshooting

### Pattern Not Matching
- ตรวจสอบว่า pattern ตรงกับข้อความจริง (case-insensitive)
- ใช้ `WithRegexPattern` สำหรับ matching ที่ซับซ้อน

### Tool Not Called
- ตรวจสอบ tool definitions ใน `buildToolDefinitions()`
- ใช้ `provider.Reset()` ก่อน setup

### Multiple Calls Not Working
- Mock provider รองรับหลาย rules แต่ต้องมี priority ที่เหมาะสม
- ใช้ `WithMaxCalls` เพื่อจำกัดจำนวนการใช้งาน

## Future Enhancements

- [ ] Async subagent execution simulation
- [ ] Meeting lifecycle (start, join, leave, end)
- [ ] Mailbox persistence simulation
- [ ] Agent capability advertisement
- [ ] Dynamic agent registration
