# A2A (Agent-to-Agent) Orchestrator
## Pure A2A Collaboration - NO SUBAGENT!

---

## 🎯 Concept

ระบบนี้ใช้ **Agent-to-Agent (A2A)** communication เท่านั้น ไม่มีการ spawn subagent

### A2A vs Subagent

| A2A | Subagent |
|-----|----------|
| สื่อสารระหว่าง **Real Agents** (มี persona) | สร้าง **Temporary Helper** (ไม่มี persona) |
| Agents มี `IDENTITY.md`, `SOUL.md` | ใช้ role template อย่างเดียว |
| มีความทรงจำของตัวเอง | ไม่มี memory |
| อยู่ถาวรในระบบ | หายไปเมื่อ task เสร็จ |
| ใช้ **Messenger**, **Mailbox** | ใช้ **SubagentManager.Spawn()** |

---

## 🏗️ Architecture

```
User Request
    ↓
Jarvis (Coordinator)
    ↓
┌─────────────────────────────────────────────────────┐
│  A2A Orchestrator                                    │
│  ├── Discovery: Agents share capabilities           │
│  ├── Meeting: Real-time discussion via Messenger    │
│  ├── Planning: Task assignment via A2A messages     │
│  ├── Execution: Agents work in parallel             │
│  ├── Integration: Combine via A2A                   │
│  └── Validation: QA via A2A                         │
└─────────────────────────────────────────────────────┘
    ↓
Real Agents (8 agents)
├── Nova (Architect)      ←→  Messenger/Mailbox
├── Clawed (Coder)        ←→  Messenger/Mailbox  
├── Pixel (Designer)      ←→  Messenger/Mailbox
├── Sentinel (QA)         ←→  Messenger/Mailbox
└── ... etc
```

---

## 🔄 A2A Workflow

### Phase 1: Discovery
```
Jarvis → Broadcast → All Agents
       "Share your capabilities"

Nova → Jarvis
      "I can design architecture"

Clawed → Jarvis
        "I can code backend"
```

### Phase 2: Meeting
```
Jarvis → All Agents
       "Meeting: Expense Tracker Project"
       
Agents → Jarvis (introductions)
       "I'm Nova, I'll design..."
       "I'm Clawed, I'll implement..."
```

### Phase 3: Planning
```
Jarvis → Nova
       "Task: Design system architecture"

Jarvis → Clawed
       "Task: Implement backend API"

Agents → Jarvis (accept/reject)
       "Accepted"
```

### Phase 4: Execution
```
Nova → Working on design
     → Send progress updates via A2A

Clawed → Working on code
       → Send progress updates via A2A

(Parallel execution via real agents)
```

### Phase 5: Integration
```
Jarvis → All Agents
       "Submit your deliverables"

Agents → Jarvis
       "Here's my code..."
       "Here's my design..."
```

### Phase 6: Validation
```
Jarvis → Sentinel
       "Please validate the project"

Sentinel → Jarvis
          "Validation passed!"
```

---

## 🛠️ Components

### 1. A2AOrchestrator (`engine.go`)
- จัดการ workflow ทั้งหมด
- ใช้ **Messenger** สำหรับ real-time communication
- ใช้ **Mailbox** สำหรับ message queue
- ใช้ **SharedContext** สำหรับ shared state

### 2. A2AAgentDiscovery (`discovery.go`)
- ค้นหา capabilities ของ real agents
- ไม่มีการ spawn - ใช้ **AgentRegistry** อย่างเดียว

### 3. A2A Tools (`tools.go`)
```go
start_a2a_project       // เริ่มโปรเจค A2A
check_a2a_project_status // เช็คสถานะ
list_a2a_agents         // แสดง agents
send_a2a_message        // ส่งข้อความ A2A
get_a2a_messages        // ดูประวัติข้อความ
```

---

## 📋 Files

| File | Description |
|------|-------------|
| `engine.go` | A2A orchestrator engine - หัวใจหลัก |
| `discovery.go` | Agent discovery สำหรับ A2A |
| `tools.go` | Tools สำหรับ LLM เรียกใช้ |
| `README.md` | เอกสารนี้ |

---

## 🚀 Usage

### 1. Initialize
```go
orchestrator := NewA2AOrchestrator(registry, provider, config, msgBus)
orchestrator.Initialize() // Setup mailboxes & messengers
```

### 2. Create Project
```go
project := orchestrator.CreateProject(
    "expense-tracker",
    "Golang API + Next.js frontend",
)
```

### 3. Start A2A Workflow
```go
orchestrator.StartProject(project.ID)
// → Discovery → Meeting → Planning → Execution → Integration → Validation
```

### 4. Monitor (Optional)
```go
orchestrator.SetPhaseChangeCallback(func(id string, phase Phase, status PhaseStatus) {
    fmt.Printf("Phase %s: %s\n", phase, status)
})

orchestrator.SetMessageCallback(func(id string, msg A2AMessage) {
    fmt.Printf("[%s -> %s]: %s\n", msg.From, msg.To, msg.Content)
})
```

---

## ⚡ Key Differences from Subagent Approach

### ❌ ที่ผิด (Subagent)
```go
// สร้าง subagent ชั่วคราว
subagentManager.Spawn(ctx, task, label, agentID, model, channel, chatID, callback)
// → ไม่มี persona, ไม่มี memory, หายไปตอนเสร็จ
```

### ✅ ที่ถูก (A2A)
```go
// ส่งข้อความไปให้ real agent
messenger.SendDirect(ctx, toAgent, agentcomm.AgentMessage{...})
mailbox.Send(mailbox.Message{...})
// → มี persona, มี memory, อยู่ถาวร
```

---

## 🎭 A2A Agents (8 ตัว)

| Agent | Role | หน้าที่ |
|-------|------|--------|
| Jarvis | Coordinator | ประสานงาน |
| Nova | Architect | ออกแบบระบบ |
| Atlas | Researcher | วิจัย best practices |
| Clawed | Coder | เขียนโค้ด |
| Pixel | Designer | ออกแบบ UI/UX |
| Sentinel | QA | ตรวจสอบคุณภาพ |
| Scribe | Writer | เขียนเอกสาร |
| Trendy | Analyst | ออกแบบ database |

---

## 📊 A2A Communication Methods

### 1. Messenger (Real-time)
```go
messenger := agent.NewMessenger(agentID, sharedCtx, msgBus)
messenger.SendDirect(ctx, toAgent, message)
messenger.Publish(ctx, message) // Broadcast
```

### 2. Mailbox (Async)
```go
mailbox := mailbox.NewMailbox(agentID, capacity)
mailbox.Send(message)
msg, _ := mailbox.Receive()
```

### 3. Shared Context
```go
sharedCtx := agent.NewSharedContext(maxLogSize, maxContext)
sharedCtx.Set("key", value)
val, _ := sharedCtx.Get("key")
sharedCtx.AddMessageLog(from, to, type, content)
```

---

## ✅ Status

- [x] A2A Discovery (ใช้ AgentRegistry)
- [x] A2A Meeting (ใช้ Messenger/Mailbox)
- [x] A2A Planning (ใช้ A2A messages)
- [x] A2A Execution (Real agents ทำงาน)
- [x] A2A Integration (ผ่าน A2A)
- [x] A2A Validation (Agent ตรวจสอบกัน)
- [x] Tools สำหรับ LLM
- [ ] Integration เข้า bootstrap (ติดค้าง)
- [ ] Register tools เข้า ToolRegistry (ติดค้าง)

**สรุป: A2A Orchestrator พร้อมใช้งานแล้ว!** 🎉
แต่ต้อง integrate เข้าระบบหลัก (bootstrap + tool registry) อีกนิดหน่อย
