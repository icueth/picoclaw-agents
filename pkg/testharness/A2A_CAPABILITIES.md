# A2A (Agent-to-Agent) System Capabilities

## ภาพรวม

A2A System ใน PicoClaw ถูกออกแบบให้ agents สามารถสื่อสารและทำงานร่วมกันได้หลายรูปแบบ ตั้งแต่การส่งข้อความง่ายๆ ไปจนถึงการประชุมหลาย agents และการจัดลำดับชั้น (hierarchy)

---

## 🔧 Core A2A Components

### 1. Agent Registry
**ไฟล์:** `pkg/agent/registry.go`

**ความสามารถ:**
- จัดการ agents ทั้งหมดในระบบ (8 agents)
- ค้นหา agent ตาม ID, Department, Role
- ตรวจสอบสิทธิ์การ spawn subagent (`CanSpawnSubagent`)
- แยกแยะ default agent
- Department indexing และ Role indexing

**ตัวอย่าง:**
```go
registry := agent.NewAgentRegistry(cfg, provider)
agent, ok := registry.GetAgent("jarvis")
canSpawn := registry.CanSpawnSubagent("jarvis", "clawed") // true
```

---

### 2. Mailbox System
**ไฟล์:** `pkg/mailbox/mailbox.go`

**ความสามารถ:**
- **Send Message:** ส่งข้อความไปยัง agent อื่น
- **Receive Message:** รับข้อความ (FIFO by priority)
- **Priority Queue:** ข้อความเรียงตาม priority (Critical > High > Normal > Low)
- **Unread Count:** นับจำนวนข้อความที่ยังไม่ได้อ่าน
- **Broadcast:** ส่งข้อความถึงทุก agents
- **Message Types:** Task, Status, Question, Answer, Meeting, Broadcast

**Message Types:**
```go
MessageTypeTask      // มอบหมายงาน
MessageTypeStatus    // อัปเดตสถานะ
MessageTypeQuestion  // ถามคำถาม
MessageTypeAnswer    // ตอบคำถาม
MessageTypeMeeting   // เชิญประชุม
MessageTypeBroadcast // ประกาศทั่วไป
```

**ตัวอย่าง:**
```go
mailbox := mailbox.NewMailbox("jarvis", 100)
mailbox.Send(mailbox.Message{
    Type: mailbox.MessageTypeTask,
    From: "jarvis",
    To: "clawed",
    Priority: mailbox.PriorityHigh,
    Content: "Please write a function...",
})
msg, _ := mailbox.Receive()
```

---

### 3. Messenger System
**ไฟล์:** `pkg/agent/messenger.go`

**ความสามารถ:**
- **Publish:** ส่งข้อความผ่าน message bus (pub/sub)
- **SendDirect:** ส่งข้อความแบบ P2P (point-to-point)
- **Broadcast:** ส่งข้อความถึงทุก agents
- **Subscribe:** ลงทะเบียนรับข้อความตาม pattern
- **Message Handler:** จัดการข้อความที่เข้ามา

**ตัวอย่าง:**
```go
messenger := agent.NewMessenger("jarvis", sharedCtx, msgBus)
messenger.Publish(ctx, agentcomm.AgentMessage{
    From: "jarvis",
    To: "clawed",
    Type: agentcomm.MsgRequest,
    Payload: "Please help...",
})
```

---

### 4. Shared Context
**ไฟล์:** `pkg/agentcomm/shared.go`, `pkg/agent/shared_context.go`

**ความสามารถ:**
- **Key-Value Store:** แชร์ข้อมูลระหว่าง agents
- **Message Log:** บันทึกประวัติการสื่อสาร
- **Thread-Safe:** ปลอดภัยในการเข้าถึงจากหลาย goroutines
- **Size Limits:** จำกัดขนาด context และ log

**ตัวอย่าง:**
```go
sharedCtx := agentcomm.NewSharedContext(100, 1000)
sharedCtx.Set("project_status", "in_progress")
val, _ := sharedCtx.Get("project_status")
sharedCtx.AddMessageLog("jarvis", "clawed", "task", "Write code")
logs := sharedCtx.GetMessageLog()
```

---

## 🤝 Collaboration Tools

### 5. Delegate Tool
**ไฟล์:** `pkg/tools/agent_delegate.go`

**ความสามารถ:**
- **Delegate Task:** มอบหมายงานให้ agent อื่น
- **Synchronous Execution:** รอผลลัพธ์จนกว่าจะเสร็จ
- **Agent Discovery:** ค้นหา agents ที่ใช้งานได้
- **Permission Check:** ตรวจสอบสิทธิ์ก่อน delegate

**ตัวอย่าง:**
```go
delegateTool := tools.NewDelegateTool(registry, runner)
result := delegateTool.Execute(ctx, map[string]any{
    "task": "Write hello world in Go",
    "target_agent": "clawed",
})
// Clawed จะทำงานและตอบกลับด้วย code
```

---

### 6. Subagent Spawning
**ไฟล์:** `pkg/tools/subagent.go`, `pkg/tools/spawn.go`

**ความสามารถ:**
- **Spawn Subagent:** สร้าง subagent ชั่วคราว
- **Role-Based Spawning:** Spawn ตาม role (coder, researcher, etc.)
- **Async Execution:** ทำงานแบบ asynchronous
- **Progress Tracking:** ติดตามความคืบหน้า
- **Task Status:** ตรวจสอบสถานะงาน (pending/running/completed/failed)
- **Cancel Task:** ยกเลิกงานที่กำลังทำ
- **Timeout Management:** จัดการ timeout และ extensions

**ตัวอย่าง:**
```go
manager := tools.NewSubagentManager(provider, model, workspace, msgBus)
taskID, _ := manager.Spawn(ctx, "Write Python code", "task-1", "", "", "cli", "chat", nil)
task := manager.GetTask(taskID)
// task.Status, task.ProgressPercent, task.Result
```

---

## 📅 Meeting System

### 7. Multi-Agent Meeting
**ไฟล์:** `pkg/agent/meeting/types.go`, `pkg/agent/meeting/scheduler.go`

**ความสามารถ:**
- **Create Meeting:** สร้างการประชุมหลาย agents
- **Add Participants:** เพิ่มผู้เข้าร่วม (Facilitator, Participant, Observer)
- **Agenda Management:** จัดการวาระการประชุม
- **Message Thread:** ข้อความในการประชุม (statement, question, proposal, vote, summary)
- **Consensus Tracking:** ติดตามการตัดสินใจร่วมกัน
- **Meeting Summary:** สรุปผลการประชุม
- **Real-time Updates:** อัปเดตสถานะแบบ real-time

**Meeting Status:**
```go
MeetingStatusPending   // รอเริ่ม
MeetingStatusOngoing   // กำลังดำเนินการ
MeetingStatusPaused    // หยุดชั่วคราว
MeetingStatusCompleted // เสร็จสิ้น
MeetingStatusCancelled // ยกเลิก
```

**ตัวอย่าง:**
```go
meeting := meeting.NewMeeting(meeting.MeetingConfig{
    Topic: "Sprint Planning",
    Facilitator: "jarvis",
    Agenda: []string{"Review last sprint", "Plan next sprint"},
    Timeout: 30 * time.Minute,
})
meeting.AddParticipant("clawed", "Clawed", "🔧", meeting.MeetingRoleParticipant)
meeting.Start()
meeting.AddMessage("clawed", "I propose we focus on API development", "proposal")
```

---

### 8. Meeting Scheduler
**ไฟล์:** `pkg/agent/meeting/scheduler.go`

**ความสามารถ:**
- **Schedule Meeting:** กำหนดเวลาประชุมล่วงหน้า
- **Recurring Meetings:** ประชุมประจำ (daily, weekly, monthly)
- **Reminders:** แจ้งเตือนก่อนประชุม
- **Auto-Start:** เริ่มประชุมอัตโนมัติตามเวลา
- **Context Preservation:** เก็บ context สำหรับประชุม

**ตัวอย่าง:**
```go
scheduler := meeting.NewScheduler()
entry, _ := scheduler.ScheduleMeeting(meeting.ScheduleConfig{
    Topic: "Weekly Standup",
    ScheduledAt: time.Now().Add(24 * time.Hour),
    Participants: []string{"jarvis", "clawed", "nova"},
    Recurring: &meeting.RecurringSchedule{
        Type: "weekly",
        Interval: 1,
    },
})
```

---

## 🏢 Hierarchy System

### 9. Agent Hierarchy
**ไฟล์:** `pkg/agent/hierarchy.go`

**ความสามารถ:**
- **Hierarchical Structure:** จัดลำดับชั้น agents (Manager → Planner → Executor)
- **Parent-Child Relations:** ความสัมพันธ์ parent-child
- **Role Assignment:** กำหนดบทบาท (Manager, Planner, Executor, Specialist, Worker)
- **Capability Management:** จัดการความสามารถของแต่ละ agent
- **Descendant Query:** ค้นหา agents ที่อยู่ใต้ hierarchy
- **Ancestor Query:** ค้นหา ancestors ของ agent
- **Hierarchical Task Execution:** สั่งงานตาม hierarchy

**Agent Roles:**
```go
RoleManager    // จัดการและประสานงาน
RolePlanner    // วางแผนและแบ่งงาน
RoleExecutor   // ปฏิบัติงาน
RoleSpecialist // ผู้เชี่ยวชาญเฉพาะทาง
RoleWorker     // ทำงานทั่วไป
```

**ตัวอย่าง:**
```go
hm := agent.NewHierarchyManager(registry)
hm.RegisterNode("clawed", "jarvis", agent.RoleExecutor, []string{"coding"})
hm.RegisterNode("sentinel", "jarvis", agent.RoleSpecialist, []string{"qa"})

children := hm.GetChildren("jarvis") // [clawed, sentinel]
descendants := hm.GetDescendants("jarvis") // all under jarvis
result, _ := hm.ExecuteHierarchicalTask(ctx, "jarvis", "Build API")
```

---

## 🎯 A2A Use Cases

### Use Case 1: Simple Task Delegation
```
User -> Jarvis: "Help me write a Python function"
Jarvis -> DelegateTool -> Clawed (coder)
Clawed -> Execute -> Return code
Jarvis -> User: Code result
```

### Use Case 2: Multi-Agent Collaboration
```
User -> Jarvis: "Build a web app"
Jarvis -> Nova (architect): Design system
Nova -> Atlas (researcher): Research best practices  
Atlas -> Clawed (coder): Implement
Clawed -> Sentinel (qa): Test
Sentinel -> Scribe (writer): Document
Jarvis -> User: Complete system
```

### Use Case 3: Meeting Discussion
```
Jarvis -> Create Meeting: "Sprint Planning"
Jarvis -> Invite: Atlas, Nova, Clawed, Sentinel
Meeting -> Discussion: Architecture decisions
Meeting -> Vote: Technology choices
Meeting -> Summary: Action items
```

### Use Case 4: Hierarchical Task Execution
```
Jarvis (Manager) -> Analyze task
  -> Nova (Planner) -> Create plan
    -> Clawed (Executor) -> Write code
    -> Pixel (Specialist) -> Design UI
  -> Sentinel (Specialist) -> Review
Jarvis -> Consolidate results
```

---

## 📊 A2A System Status

| Component | Status | Notes |
|-----------|--------|-------|
| Agent Registry | ✅ Ready | 8 agents loaded from config |
| Mailbox | ✅ Ready | Send/receive working |
| Messenger | ⚠️ Partial | P2P needs work, mailbox works |
| Shared Context | ✅ Ready | Key-value + logs working |
| Delegate Tool | ✅ Ready | Task delegation working |
| Subagent Spawn | ✅ Ready | Basic spawn working |
| Meeting System | ⚠️ Partial | Core ready, API needs integration |
| Hierarchy | ⚠️ Partial | Structure ready, execution needs work |

---

## 🔮 Future A2A Enhancements

- [ ] **Agent Discovery Service:** Agents can advertise capabilities
- [ ] **Negotiation Protocol:** Agents negotiate task assignments
- [ ] **Conflict Resolution:** Handle disagreements between agents
- [ ] **Shared Memory RAG:** Agents share knowledge via RAG
- [ ] **Workflow Engine:** Visual workflow builder for A2A
- [ ] **Performance Metrics:** Track A2A collaboration efficiency
- [ ] **Agent Reputation:** Score agents based on task completion
