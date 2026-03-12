# A2A Project Execution Flow

## Overview

A2A (Agent-to-Agent) Project Execution Flow คือ กระบวนการทำงานแบบ Multi-Agent ที่ Agent หลายตัวทำงานร่วมกันเพื่อสร้าง deliverables โดยไม่มีการใช้ Subagent แต่ใช้การสื่อสาร A2A แทน

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         A2A PROJECT LIFECYCLE                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Discovery        Meeting         Planning        Execution               │
│  ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌──────────┐             │
│  │  3 min  │────>│  3 min  │────>│  5 min  │────>│ (varies) │             │
│  └─────────┘     └─────────┘     └─────────┘     └──────────┘             │
│                                                        │                    │
│                                                        v                    │
│   ┌──────────┐     ┌──────────┐                  Integration               │
│   │ Complete │<────│Validation│<──────────────────── 2 min                  │
│   └──────────┘     └──────────┘                                             │
│                        3 min                                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Phase 1: Discovery (ค้นหาความสามารถ Agents)

### Objective

ค้นหาและบันทึกความสามารถของแต่ละ Agent ในระบบ เพื่อใช้ประโยชน์ในการมอบหมายงาน

### Participants

- **Jarvis** (Coordinator) - ผู้ดำเนินการ
- **All Agents** (8 agents) - ผู้ตอบกลับ

### Workflow

```
┌─────────┐                  ┌─────────┐
│ Jarvis  │─── broadcast ───>│ Agent 1 │──┐
└─────────┘   "discovery"    └─────────┘  │
      │                                    │
      │                  ┌─────────┐       │    ┌─────────────┐
      ├─────────────────>│ Agent 2 │───────┼───>│  Responses  │
      │   "discovery"    └─────────┘       │    │  Collector  │
      │                                    │    └─────────────┘
      │                  ┌─────────┐       │           │
      └─────────────────>│ Agent N │───────┘           │
          "discovery"    └─────────┘                   │
                                                      v
                                               ┌─────────────┐
                                               │  Store to   │
                                               │  Project.   │
                                               │  Metadata   │
                                               └─────────────┘
```

### Steps

1. **Broadcast Discovery Request**

   ```go
   Jarvis broadcast "discovery" to all agents
   Message: "Project 'X' starting. Please share your capabilities"
   ```

2. **Agents Respond with Capabilities**

   ```go
   Each agent responds with:
   {
     "agent_id": "clawed",
     "role": "coder",
     "capabilities": ["go", "python", "debugging", "testing"],
     "responsibilities": ["write code", "fix bugs", "review PRs"]
   }
   ```

3. **Collect and Store**
   - Wait: **Until all active agents respond (Event-driven) OR fallback timeout 3 minutes**
   - Store in: `project.Phases[PhaseDiscovery].Metadata`
   - Key: `capabilities`

### Output

- รู้ว่าใครถนัดอะไร
- ใช้เป็นข้อมูลสำหรับ task assignment ใน phase ถัดไป

---

## Phase 2: Meeting (ประชุมวางแผน)

### Objective

สร้างความเข้าใจร่วมกันในทีม และกำหนดงานเริ่มต้นให้แต่ละคน

### Participants

- **Jarvis** (Coordinator)
- **All Agents** (8 agents)

### Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                        MEETING FLOW                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Step 1: Meeting Invitation                                       │
│  ┌─────────┐              ┌─────────┐                          │
│  │ Jarvis  │──broadcast──>│ Agents  │                          │
│  │         │ "meeting_start"                                      │
│  └─────────┘              └────┬────┘                          │
│                                │                                │
│  Step 2: Acknowledgment (1 min)                                 │
│  │<──────── meeting_ack ────────┤                                │
│  │                from each agent                                │
│                                                                 │
│  Step 3: Introduction Request                                     │
│  ┌─────────┐              ┌─────────┐                          │
│  │ Jarvis  │──direct─────>│ Agent 1 │                          │
│  │         │"introduction"│         │                          │
│  │         ├─────────────>│ Agent 2 │                          │
│  │         │              │  ...    │                          │
│  └─────────┘              └─────────┘                          │
│                                                                 │
│  Step 4: Introductions (1 min)                                    │
│  │<──────── introduction ────────┤                                │
│  │                from each agent                                │
│                                                                 │
│  Step 5: Task Assignments                                         │
│  ┌─────────┐    LLM     ┌─────────────┐                        │
│  │ Project │───────────>│ Assignments │                        │
│  │ Details │  Analysis  │   Text      │                        │
│  └─────────┘            └──────┬──────┘                        │
│                                │                                │
│  Step 6: Broadcast Assignments                                  │
│  ┌─────────┐              ┌─────────┐                          │
│  │ Jarvis  │──broadcast──>│ Agents  │                          │
│  │         │"task_discussion"                                    │
│  └─────────┘              └────┬────┘                          │
│                                │                                │
│  Step 7: Acceptance (1 min)                                     │
│  │<──────── task_accepted ───────┤                                │
│  │                from each agent                                │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Steps Detail

| Step | Action                        | Wait Condition                        | Details                                |
| ---- | ----------------------------- | ------------------------------------- | -------------------------------------- |
| 1    | Broadcast meeting invitation  | -                                     | Agenda: Introductions, Task Discussion |
| 2    | Wait for meeting_ack          | **All agents ACK or 1 min timeout**   | Agents confirm attendance              |
| 3    | Request introductions         | -                                     | Direct message to each agent           |
| 4    | Wait for introductions        | **All agents intro or 1 min timeout** | Agents introduce themselves            |
| 5    | Generate assignments with LLM | -                                     | Analyze project → assign tasks         |
| 6    | Broadcast assignments         | -                                     | Share with all agents                  |
| 7    | Wait for task_accepted        | **All accepted or 1 min timeout**     | Agents confirm task acceptance         |

### LLM Task Assignment

```go
// Prompt ที่ใช้สร้าง assignments
prompt := `You are Jarvis, the coordinator. Given this project:

Name: {project_name}
Description: {project_description}

Available agents and their capabilities:
- nova: architect, Capabilities: [system design, api design]
- clawed: coder, Capabilities: [go, python, testing]
...

Propose task assignments in this format:
AgentName: Task description`
```

### Output

- ทุก agent รู้ว่าตัวเองต้องทำอะไร
- มี list ของ proposed assignments
- เก็บใน: `project.Phases[PhaseMeeting].Result`

---

## Phase 3: Planning (วางแผนการทำงาน)

### Objective

วิเคราะห์งานละเอียด แบ่งเป็น phases และจัดลำดับความสำคัญ

### Participants

- **Jarvis** (ใช้ LLM วิเคราะห์)

### Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                      PLANNING PIPELINE                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Input: Project Description                                       │
│       │                                                          │
│       v                                                          │
│  ┌─────────────────┐                                             │
│  │ Step 1: Extract │──LLM──> List of tasks                       │
│  │ Tasks           │     [Task1, Task2, Task3, ...]               │
│  └─────────────────┘                                             │
│       │                                                          │
│       v                                                          │
│  ┌─────────────────┐                                             │
│  │ Step 2: Analyze │──LLM──> TaskAnalysis[]                       │
│  │ Complexity      │     [{task, complexity, agent, deps}, ...]   │
│  └─────────────────┘                                             │
│       │                                                          │
│       v                                                          │
│  ┌─────────────────┐                                             │
│  │ Step 3: Group   │────────> PhasePlan[]                         │
│  │ into Phases     │     [Phase1, Phase2, ...]                    │
│  └─────────────────┘                                             │
│       │                                                          │
│       v                                                          │
│  ┌─────────────────┐                                             │
│  │ Step 4: Create  │────────> TaskQueue (ready for execution)     │
│  │ Task Queue      │                                             │
│  └─────────────────┘                                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Step 1: Extract Tasks

```go
// ใช้ LLM แยก project เป็นรายการ task ย่อย
Input:  "สร้าง REST API สำหรับระบบ e-commerce"
Output: [
    "Design API schema and endpoints",
    "Setup database with migrations",
    "Implement authentication middleware",
    "Create product endpoints",
    "Create order endpoints",
    "Write unit tests",
    "Write integration tests"
]
```

### Step 2: Analyze Complexity

```go
// วิเคราะห์แต่ละ task ด้วย LLM
For each task, determine:
- COMPLEXITY: simple | medium | complex | critical
- ESTIMATED_MINUTES: เวลาที่ใช้
- BEST_AGENT: ใครเหมาะสมที่สุด
- DEPENDENCIES: ต้องรอ task ไหนเสร็จก่อน
```

**Complexity Levels:**

| Level    | Max Global Concurrent | Max Per-Agent | Description                |
| -------- | --------------------- | ------------- | -------------------------- |
| Simple   | 3-5                   | 2             | งานง่าย ไม่มี dependencies |
| Medium   | 2-3                   | 1             | งานปานกลาง                 |
| Complex  | 1-2                   | 1             | งานซับซ้อน ต้องใช้ LLM นาน |
| Critical | 1                     | 1             | งานสำคัญ ทำทีละงานเท่านั้น |

### Step 3: Group into Phases

```go
// จัดกลุ่มตาม complexity + dependencies
Phases:
├── Phase 1: Simple tasks (batch up to 3)
│   ├── Task A
│   ├── Task B
│   └── Task C
├── Phase 2: Medium tasks (batch up to 2)
│   ├── Task D
│   └── Task E
└── Phase 3: Complex/Critical (sequential)
    ├── Task F
    └── Task G
```

### Step 4: Create Task Queue

```go
TaskQueue {
    pending: [
        {id: "task-1", priority: 1, deps: [], agent: "nova"},
        {id: "task-2", priority: 2, deps: ["task-1"], agent: "clawed"},
        ...
    ]
    batchSize: 3 (default)
}
```

### Output

- `TaskQueue` พร้อมสำหรับ execution
- `PhasePlan[]` แผนการทำงานแบ่งเป็น phases
- `project.Phases[PhasePlanning].Metadata`

---

## Phase 4: Execution (เริ่มทำงานจริง)

### Objective

Agents ทำงานจริงตามที่วางแผนไว้ โดยทำงานแบบ Parallel แต่ควบคุมไม่ให้ API โดน overwhelm

### Participants

- **All Agents** (ทำงานตามที่ได้รับมอบหมาย)

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│        CONTINUOUS WORKER POOL & DUAL-LEVEL RATE LIMITING         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │     GLOBAL RATE LIMITER (e.g. Max 3-5 LLM calls)        │   │
│  │         ┌─────┐    ┌─────┐    ┌─────┐                   │   │
│  │         │ 🔴  │    │ 🟢  │    │ 🟢  │  ← Semaphore      │   │
│  └─────────┴──┬──┴────┴──┬──┴────┴──┬──┴───────────────────┘   │
│               │          │          │                          │
│               v          v          v                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │    PER-AGENT CONCURRENCY LIMITER (e.g. Max 1-2 tasks)   │   │
│  │    Agent A [1/1]🔴   Agent B [0/1]🟢   Agent C [0/1]🟢   │   │
│  └─────────────────────────────────────────────────────────┘   │
│               │                     │                          │
│               v                     v                          │
│         ┌──────────┐          ┌──────────┐                     │
│         │ LLM Call │          │ LLM Call │                     │
│         └──────────┘          └──────────┘                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Continuous Processing Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    CONTINUOUS TASK DISPATCHER                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Event Trigger (On Next Task Added or Task Completed):           │
│  │                                                               │
│  ├── 1. Get Ready Tasks                                          │
│  │      ├── Filter: dependencies completed                       │
│  │      └── Sort: by priority (high first)                       │
│  │                                                               │
│  ├── 2. Dispatch Loop (Iterate ready tasks)                      │
│  │      │                                                        │
│  │      ├── Global Slot Available? No ──> Break Loop             │
│  │      ├── Agent Slot Available? No ──> See On-Demand Spawning  │
│  │      │                                                        │
│  │      └── Yes ──> Acquire Slots ──> Execute Task Async         │
│  │                                                               │
│  ├── 3. On-Demand Worker Spawning (If Target Agent is Full)      │
│  │      ├── Check Role Match for available slot                  │
│  │      └── Trigger OutsourcePool.Hire() if mismatched or busy   │
│  │          ├── Create temp OutsourceAgent + worker loop         │
│  │          └── Execute Task Async with temp semaphore           │
│  │                                                               │
│  └── 4. Handle Completion                                        │
│         ├── Release Agent/Outsource Slot                         │
│         ├── Release Global Slot                                  │
│         └── Re-trigger Event Loop to process queue without delay │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Assignment Execution Detail

```go
func executeA2AAssignment(project, assignment) {
    // 1. Update status
    assignment.Status = AssignmentStatusRunning
    assignment.Progress = 0

    // 2. Send task_start message to agent
    sendA2AMessage("jarvis", assignment.ToAgent, "task_start", task)
    assignment.Progress = 25

    // 3. Agent processes with LLM (async)
    //    - Can use tools
    //    - Can make decisions
    //    - Returns result

    // 4. Wait for task_complete response
    //    Timeout: 30 minutes
    assignment.Progress = 75

    // 5. Store result
    assignment.Status = AssignmentStatusCompleted
    assignment.Progress = 100
    assignment.Result = response.Content
}
```

### Concurrency by Complexity

| Complexity | Max Global Concurrent | Max Per-Agent | Execution Mode  |
| ---------- | --------------------- | ------------- | --------------- |
| Simple     | 3-5                   | 2             | Parallel        |
| Medium     | 2-3                   | 1             | Parallel        |
| Complex    | 1-2                   | 1             | Parallel        |
| Critical   | 1                     | 1             | Sequential only |

### Progress Tracking

```go
// Real-time progress updates
onAssignmentProgress(projectID, assignmentID, progress, message)

// Example updates:
Progress 0%:  "Task assigned to clawed"
Progress 25%: "Agent clawed started working"
Progress 50%: "Processing..."
Progress 75%: "Almost done..."
Progress 100%: "Task completed successfully"
```

### Output

- งานที่กำหนดเสร็จสมบูรณ์
- Artifacts (code, documents, etc.)
- Project.Progress อัพเดต

---

## Phase 5: Integration (รวมผลงาน)

### Objective

รวมผลงานจากทุก agent เข้าด้วยกันเป็นชิ้นงานสมบูรณ์

### Participants

- **Jarvis** (Coordinator)
- **All Agents** (ส่ง deliverables)

### Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                     INTEGRATION FLOW                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Request Deliverables                                          │
│     ┌─────────┐              ┌─────────┐                        │
│     │ Jarvis  │──broadcast──>│ Agents  │                        │
│     │         │"integration_request"                              │
│     └─────────┘              └────┬────┘                        │
│                                   │                             │
│  2. Collect Deliverables (10 min)                                 │
│     │<──────── deliverable ──────┤                                │
│     │                from each agent                              │
│     │                                                             │
│     v                                                             │
│  ┌─────────────────────────────────────────┐                      │
│  │           PROJECT ARTIFACTS              │                      │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐   │                      │
│  │  │ Code    │ │ Design  │ │ Tests   │   │                      │
│  │  │ (Clawed)│ │ (Nova)  │ │(Sentinel│   │                      │
│  │  └─────────┘ └─────────┘ └─────────┘   │                      │
│  └─────────────────────────────────────────┘                      │
│                                   │                             │
│  3. Integration Complete                                          │
│     ┌─────────┐              ┌─────────┐                        │
│     │ Jarvis  │──broadcast──>│ Agents  │                        │
│     │         │"integration_complete"                            │
│     └─────────┘              └─────────┘                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Steps

1. **Request Integration**
   - Jarvis broadcast "integration_request"
   - Message: "Please submit your deliverables for integration"

2. **Collect Deliverables**
   - Wait: **Until all active agents submit (Event-driven) OR 3 minutes timeout**
   - Each agent sends their work
   - Store as `Project.Artifacts[]`

3. **Integration Complete**
   - Jarvis รวมผลงาน
   - Broadcast แจ้งทุกคน

### Output

- `Project.Artifacts[]` รวมผลงานทั้งหมด
- สรุปจำนวน deliverables

---

## Phase 6: Validation (ตรวจสอบคุณภาพ)

### Objective

ตรวจสอบคุณภาพของผลงานโดย QA Agent

### Participants

- **Jarvis** (ส่งตรวจ)
- **Sentinel** (QA Agent) (ตรวจสอบ)

### Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                    VALIDATION FLOW                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Request Validation                                            │
│     ┌─────────┐              ┌──────────┐                       │
│     │ Jarvis  │──direct─────>│ Sentinel │                       │
│     │         │"validation_  │  (QA)    │                       │
│     │         │  request"    │          │                       │
│     └─────────┘              └────┬─────┘                       │
│                                   │                             │
│  2. QA Analysis (by LLM)                                          │
│     ┌──────────────────────────────────────────┐                  │
│     │  Sentinel ใช้ LLM ตรวจสอบ:              │                  │
│     │  ├── Code quality                       │                  │
│     │  ├── Best practices                     │                  │
│     │  ├── Completeness                       │                  │
│     │  └── Potential issues                   │                  │
│     └──────────────────────────────────────────┘                  │
│                                   │                             │
│  3. Validation Result                                             │
│     │<──────── validation_complete ──────────┤                    │
│     v                                                             │
│  ┌──────────────────────────────────────────┐                     │
│  │  Result: "All tests passed. Code looks   │                     │
│  │  good with minor suggestions..."         │                     │
│  └──────────────────────────────────────────┘                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Steps

1. **Request Validation**
   - Jarvis ส่งให้ Sentinel
   - Include: Project details + All artifacts

2. **QA Analysis**
   - Sentinel ใช้ LLM วิเคราะห์
   - ตรวจสอบหลายด้าน

3. **Store Result**
   - Wait: **Until QA responds OR 5 minutes timeout**
   - Store in: `project.Phases[PhaseValidation].Result`

### Output

- Validation report
- Quality score/feedback
- หรือ "Validation timeout" ถ้าไม่สำเร็จ (ไม่ fail project)

---

## Completion

```
┌─────────────────────────────────────────────────────────────────┐
│                     PROJECT COMPLETION                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Project Status: COMPLETED ✓                                     │
│  Duration: 31 minutes                                            │
│                                                                 │
│  Summary:                                                        │
│  ├── Phases: 6/6 completed                                       │
│  ├── Artifacts: 5 items                                          │
│  ├── Tasks: 8 completed, 0 failed                                │
│  └── Quality: Passed validation                                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Key Mechanisms

### 1. Semaphore-based Concurrency Control

```go
// Global limiter จำกัด concurrent LLM calls
type A2AOrchestrator struct {
    llmRateLimiter chan struct{}  // Buffer = 3
    maxLLMConcurrent int          // Max 3 concurrent
}

func (o *A2AOrchestrator) executeWithLimit(ctx, fn) {
    select {
    case o.llmRateLimiter <- struct{}{}:  // Acquire
        defer func() { <-o.llmRateLimiter }()  // Release
        return fn()
    case <-ctx.Done():
        return timeoutError
    }
}
```

### 2. Task Queue with Dependencies

```go
type TaskQueue struct {
    pending []*A2AAssignment
    batchSize int
}

func (q *TaskQueue) GetNextBatch(completed map[string]bool, maxSize int) []*A2AAssignment {
    var batch []*A2AAssignment

    for _, task := range q.pending {
        // Check dependencies
        canExecute := true
        for _, depID := range task.DependsOn {
            if !completed[depID] {
                canExecute = false
                break
            }
        }

        if canExecute && len(batch) < maxSize {
            batch = append(batch, task)
        }
    }

    return batch
}
```

### 3. Progress Callback System

```go
type A2AOrchestrator struct {
    onAssignmentProgress func(projectID, assignmentID string, progress int, message string)
    onAssignmentFailed   func(projectID, assignmentID string, agentID string, err error)
}

// Called during execution
o.updateAssignmentProgress(project, assignment, 50, "Processing...")
```

### 4. Timeout Management

| Phase       | Component      | Timeout                                       |
| ----------- | -------------- | --------------------------------------------- |
| Discovery   | Responses      | **Event-driven (wait all) or 3 min fallback** |
| Meeting     | meeting_ack    | **Event-driven (wait all) or 1 min fallback** |
| Meeting     | introduction   | **Event-driven (wait all) or 1 min fallback** |
| Meeting     | task_accepted  | **Event-driven (wait all) or 1 min fallback** |
| Planning    | LLM calls      | 5 min                                         |
| Execution   | Per assignment | 30 min                                        |
| Integration | Deliverables   | **Event-driven (wait all) or 3 min fallback** |
| Validation  | QA result      | 5 min                                         |

---

## Error Handling

### Graceful Degradation

```go
// ถ้า agent ไม่ตอบ → ไม่ fail ทั้ง project
responses := o.waitForResponses(ctx, project, "meeting_ack", expectedCount)
if len(responses) < expectedCount {
    log.Warn("Not all agents responded, continuing...")
    // Continue with available agents
}

// ถ้า task fail → ทำต่อ
if err := o.executeA2AAssignment(project, assignment); err != nil {
    log.Error("Assignment failed:", err)
    // Mark as failed but continue with next tasks
}
```

### Retry Mechanism (Future)

```go
// Task ที่ fail สามารถ retry ได้
if assignment.Status == AssignmentStatusFailed {
    if assignment.RetryCount < maxRetries {
        assignment.RetryCount++
        project.TaskQueue.Add(assignment)  // Re-queue
    }
}
```

---

## Performance Comparison

| Mode                | Meeting | Execution | Total   |
| ------------------- | ------- | --------- | ------- |
| **Sequential**      | 30 min  | 60 min    | 90+ min |
| **Parallel (this)** | 3 min   | 15 min    | ~31 min |
| **Speedup**         | **10x** | **4x**    | **3x**  |

---

## Related Files

- `pkg/agent/orchestrator_engine.go` - Main orchestration logic
- `pkg/agent/a2a_worker.go` - Agent worker implementation
- `pkg/agent/orchestrator_tools.go` - A2A tools (start_project, check_status)
