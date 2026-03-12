# PicoClaw Agent Team Development Plan

> แผนพัฒนาระบบ Agent Team สำหรับ PicoClaw - จาก Generalist สู่ Specialist AI Team

## 🎯 วิสัยทัศน์

สร้าง "ทีม AI Agent" ที่แต่ละตัวมีหน้าที่เฉพาะทาง (Specialist) และทำงานร่วมกันเป็นระบบ เหมือนบริษัทจริงๆ โดยใช้ประโยชน์จากระบบที่มีอยู่แล้วใน PicoClaw (RAG, Kanban, Job, MCP, Skills)

## 📋 โครงสร้างทีมเป้าหมาย (The $400/month Team)

```
┌─────────────────────────────────────────────────────────────────┐
│                         JARVIS                                  │
│                    (Team Coordinator)                           │
│            Claude Sonnet - คอยแบ่งงานและประสานงาน              │
└──────────────────────┬──────────────────────────────────────────┘
                       │
       ┌───────────────┼───────────────┬───────────────┬───────────────┐
       ▼               ▼               ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│    ATLAS    │ │   SCRIBE    │ │   CLAWED    │ │  SENTINEL   │ │   TRENDY    │
│  Researcher │ │ Copywriter  │ │  Developer  │ │  QA/Monitor │ │ Trend Scout │
│   GPT-5     │ │   GPT-5     │ │Claude Sonnet│ │Claude Haiku │ │   GPT-4     │
│  (Research) │ │  (Marketing)│ │(Engineering)│ │    (QA)     │ │  (Research) │
└─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘
       │               │               │               │               │
       └───────────────┴───────────────┴───────────────┴───────────────┘
                       │
                       ▼
              ┌─────────────┐
              │   SHARED    │
              │   MEMORY    │
              │   + BOARD   │
              └─────────────┘
```

### Agent Configuration (ตัวอย่าง)

แต่ละ Agent สามารถตั้งค่า model และ parameters ผ่าน config ได้อย่างอิสระ:

**Default Model (ใช้งานได้ทันที):**
- ทุก Agent จะใช้ `kimi-coding/kimi-for-coding` เป็นค่าเริ่มต้น
- User สามารถเข้าไปแก้ไข config ของแต่ละ Agent ได้ภายหลัง
- รองรับการตั้งค่า model แยกตาม agent, role, หรือใช้ร่วมกัน

```json
{
  "agents": {
    "defaults": {
      "model": {
        "primary": "kimi-coding/kimi-for-coding",
        "params": {
          "temperature": 0.7,
          "max_tokens": 4096,
          "top_p": 0.9
        }
      }
    },
    "list": [
      {
        "id": "jarvis",
        "name": "Jarvis",
        "role": "coordinator",
        "department": "planning",
        "avatar": "jarvis_pixel",
        "model": {
          "primary": "kimi-coding/kimi-for-coding",
          "fallbacks": []
        },
        "model_params": {
          "temperature": 0.3,
          "max_tokens": 4096,
          "top_p": 0.9
        },
        "persona": {
          "soul": "คุณคือผู้จัดการโครงการ AI ที่เชี่ยวชาญในการวิเคราะห์และแบ่งงาน",
          "tone": "professional",
          "language": "th"
        },
        "ui_config": {
          "position": { "x": 400, "y": 300 },
          "room": "jarvis_office",
          "sprite_set": "coordinator_dark",
          "animation": "typing"
        }
      },
      {
        "id": "atlas",
        "name": "Atlas",
        "role": "researcher", 
        "department": "research",
        "avatar": "atlas_pixel",
        "model": {
          "primary": "openai/gpt-5.2",
          "fallbacks": ["google/gemini-pro"]
        },
        "model_params": {
          "temperature": 0.7,
          "max_tokens": 8192
        },
        "persona": {
          "soul": "คุณคือนักวิจัยข้อมูลที่ละเอียดและรอบคอบ",
          "specialties": ["web_research", "data_analysis", "trend_tracking"]
        },
        "schedule": "0 * * * *",
        "ui_config": {
          "position": { "x": 700, "y": 100 },
          "room": "research_corner",
          "sprite_set": "researcher_blue",
          "props": ["globe", "books", "computer"]
        }
      },
      {
        "id": "clawed",
        "name": "Clawed",
        "role": "developer",
        "department": "engineering",
        "avatar": "clawed_pixel",
        "model": {
          "primary": "anthropic/claude-sonnet-4.6",
          "fallbacks": ["openai/gpt-5.2", "deepseek/deepseek-chat"]
        },
        "model_params": {
          "temperature": 0.2,
          "max_tokens": 16384
        },
        "persona": {
          "soul": "คุณคือโปรแกรมเมอร์มือฉมังที่เขียนโค้ดสะอาด",
          "coding_style": "clean_code",
          "preferred_languages": ["go", "python", "typescript"]
        },
        "schedule": "0 2 * * *",
        "ui_config": {
          "position": { "x": 700, "y": 250 },
          "room": "dev_zone",
          "sprite_set": "developer_hoodie",
          "props": ["coffee", "multiple_monitors", "keyboard"]
        }
      }
    ]
  }
}
```

### Model Selection Strategy

**Default Model สำหรับทุก Agent:**
```json
{
  "model": {
    "primary": "kimi-coding/kimi-for-coding",
    "fallbacks": []
  }
}
```

**เหตุผลที่เลือก kimi-coding เป็นค่าเริ่มต้น:**
- รองรับทั้ง coding และ general tasks
- Performance ดีสำหรับทุก role (research, writing, coordination)
- ใช้งานได้ทันทีไม่ต้อง config เพิ่ม
- User สามารถเปลี่ยน model ภายหลังได้ตามต้องการ

**Model แนะนำสำหรับแต่ละ Role (optional config):**

| Task Type | แนะนำ Model | เหตุผล |
|-----------|------------|--------|
| **Default (All)** | **kimi-coding/kimi-for-coding** | ✅ ใช้ได้ทันที, ครอบจักรวาบ |
| **Coordination** | Claude Sonnet | เข้าใจ context ดี, ตัดสินใจแม่นยำ |
| **Research** | GPT-5 | ข้อมูลละเอียด, ครอบคลุม |
| **Coding** | Claude Sonnet | เขียนโค้ดสะอาด, debug เก่ง |
| **Content** | GPT-5 / Gemini | เร็ว, สร้างสรรค์ |
| **QA/Review** | Claude Haiku | เร็ว, ประหยัด, ตรวจสอบได้ |
| **Simple Tasks** | Local Models | ประหยัด cost, ไม่ต้องใช้ cloud |

### Agent Roles

| Agent | Department | Role | หน้าที่หลัก | Schedule |
|-------|-----------|------|-----------|----------|
| **Jarvis** | Planning | Coordinator | วิเคราะห์งาน แบ่งย่อย มอบหมาย | On-demand |
| **Atlas** | Research | Researcher | ค้นหาข้อมูล วิเคราะห์ | ทุกชั่วโมง |
| **Scribe** | Marketing | Copywriter | เขียน content บทความ | On-demand |
| **Clawed** | Engineering | Developer | เขียนโค้ด รีวิว PR | ตี 2 ทุกวัน |
| **Sentinel** | QA | Reviewer | ตรวจสอบคุณภาพ | ทุก 2 ชั่วโมง |
| **Trendy** | Research | Trend Scout | หาเทรนด์ใหม่ | ทุก 3 ชั่วโมง |
| **Pixel** | Design | Designer | สร้างภาพ UI/UX | On-demand |
| **Nova** | Architecture | Architect | ออกแบบระบบ | On-demand |

---

## 🏗️ Architecture ที่จะพัฒนา

```
┌─────────────────────────────────────────────────────────────────────┐
│                    AGENT TEAM ARCHITECTURE                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    COORDINATION LAYER                         │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │   │
│  │  │  Jarvis  │  │  Task    │  │  Load    │  │   Pipeline   │  │   │
│  │  │Coordinator│  │  Router  │  │ Balancer │  │   Manager    │  │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────────┘  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    COMMUNICATION LAYER                        │   │
│  │                     (Agent Mailbox)                           │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │   │
│  │  │ Messages │  │ Threads  │  │ Broadcast│  │  Notifications│  │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────────┘  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    AGENT LAYER (Specialists)                  │   │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐     │   │
│  │  │ Atlas  │ │ Scribe │ │ Clawed │ │Sentinel│ │ Trendy │ ... │   │
│  │  │Research│ │Content │ │  Code  │ │   QA   │ │ Trends │     │   │
│  │  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘     │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    MEMORY LAYER                               │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │   │
│  │  │Agent RAG │  │ Shared   │  │  Daily   │  │  Long-term   │  │   │
│  │  │Namespace │  │ Knowledge│  │   Log    │  │   Memory     │  │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────────┘  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    EXECUTION LAYER                            │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │   │
│  │  │  Kanban  │  │   Job    │  │  Skills  │  │    MCP       │  │   │
│  │  │  Board   │  │  Queue   │  │ Registry │  │   Tools      │  │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────────┘  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 📅 Development Phases

### Phase 1: Foundation (Weeks 1-3)

#### 1.1 Agent Persona Configuration System
**ไฟล์ที่ต้องสร้าง/แก้ไข:**
- `pkg/config/agent_persona.go` (🆕 ใหม่)
- `pkg/config/config.go` (📝 ขยาย)

**รายละเอียด:**
```go
// AgentPersona กำหนดตัวตนและพฤติกรรมของ Agent
type AgentPersona struct {
    Soul          string            // บุคลิก น้ำเสียง
    Boundaries    []string          // สิ่งที่ห้ามทำ
    AllowedTools  []string          // Tools ที่ใช้ได้
    DisallowedTools []string        // Tools ที่ห้ามใช้
    MemoryScope   []string          // RAG namespace ที่เข้าถึงได้
    ResponsePatterns []ResponsePattern // ตัวอย่างการตอบ
}

type ResponsePattern struct {
    When string
    Then string
}
```

**Config format:**
```json
{
  "agents": {
    "list": [
      {
        "id": "atlas",
        "name": "Atlas",
        "role": "researcher",
        "department": "research",
        "persona": {
          "soul": "คุณคือนักวิจัยข้อมูลที่ละเอียดและรอบคอบ",
          "boundaries": ["ไม่เขียนโค้ด", "ไม่ตัดสินใจแทนผู้ใช้"],
          "allowed_tools": ["web_search", "rag_search", "message"],
          "memory_scope": ["research", "web", "news"],
          "response_patterns": [
            {"when": "พบข้อมูลใหม่", "then": "สรุปพร้อมแหล่งที่มา"}
          ]
        }
      }
    ]
  }
}
```

#### 1.2 Agent-Specific Memory Namespace
**ไฟล์ที่ต้องสร้าง:**
- `pkg/rag/agent_collection.go` (🆕 ใหม่)
- `pkg/memory/agent_memory.go` (🆕 ใหม่)

**ฟีเจอร์:**
- แยก RAG collection ตาม agent/department
- Atlas เข้าถึง "research" namespace ทั้งหมด
- Scribe เข้าถึง "content" + รับข้อมูลจาก Atlas
- Memory isolation ป้องกัน context ปนกัน

#### 1.3 Basic Agent Mailbox System
**ไฟล์ที่ต้องสร้าง:**
- `pkg/agentcomm/mailbox.go` (🆕 ใหม่)
- `pkg/agentcomm/message.go` (🆕 ใหม่)
- `pkg/db/migrations/agent_comm.sql` (🆕 ใหม่)

**Schema:**
```sql
CREATE TABLE agent_messages (
    id TEXT PRIMARY KEY,
    from_agent TEXT NOT NULL,
    to_agent TEXT NOT NULL,  -- "broadcast" สำหรับส่งทั้งทีม
    type TEXT NOT NULL,      -- task, result, question, notify
    content TEXT NOT NULL,
    priority TEXT,           -- low, normal, high, critical
    thread_id TEXT,          -- สำหรับ conversation
    job_id TEXT,             -- เชื่อมกับ job system
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_to ON agent_messages(to_agent, read_at);
CREATE INDEX idx_messages_thread ON agent_messages(thread_id);
```

**Deliverables Phase 1:**
- [ ] Agent persona config system
- [ ] Memory namespace isolation
- [ ] Basic mailbox (send/receive/read)
- [ ] ตัวอย่าง config สำหรับ Atlas + Scribe
- [ ] Default model configuration (`kimi-coding/kimi-for-coding`)

---

### Phase 2: Coordination (Weeks 4-6)

#### 2.1 Jarvis - Team Coordinator
**ไฟล์ที่ต้องสร้าง:**
- `pkg/office/coordinator.go` (🆕 ใหม่)
- `pkg/office/execution_plan.go` (🆕 ใหม่)

**ฟีเจอร์:**
```go
type Coordinator struct {
    mailbox    *agentcomm.Mailbox
    kanban     *KanbanManager
    router     *TaskRouter
    registry   *agent.AgentRegistry
}

func (c *Coordinator) ProcessRequest(request string) (*ExecutionPlan, error) {
    // 1. วิเคราะห์ request
    // 2. สร้าง execution plan
    // 3. แบ่งงานให้ agent ต่างๆ
    // 4. สร้าง kanban board ติดตาม
}
```

#### 2.2 Task Dependencies & Pipeline
**ไฟล์ที่ต้องแก้ไข:**
- `pkg/office/kanban.go` (📝 เพิ่มฟิลด์)
- `pkg/office/pipeline.go` (🆕 ใหม่)

**ฟีเจอร์:**
```go
type KanbanTask struct {
    // ... existing fields ...
    Dependencies []string      // task IDs ที่ต้องทำก่อน
    Pipeline     string        // "content_creation", "code_review"
    Stages       []TaskStage   // Multi-agent workflow
}

type TaskStage struct {
    AgentID  string
    Role     string
    Status   StageStatus  // pending, in_progress, done, blocked
    Output   string       // ผลลัพธ์ส่งต่อ stage ถัดไป
}

// Pipeline Templates
var ContentPipeline = PipelineTemplate{
    Name: "content_creation",
    Stages: []StageTemplate{
        {AgentRole: "researcher", Task: "research_topic"},
        {AgentRole: "writer", Task: "write_content"},
        {AgentRole: "reviewer", Task: "review_content"},
    }
}
```

#### 2.3 Workload Balancing
**ไฟล์ที่ต้องสร้าง:**
- `pkg/office/workload.go` (🆕 ใหม่)

**ฟีเจอร์:**
- ตรวจสอบ workload ของแต่ละ agent
- Auto-assign งานให้ agent ที่ว่าง
- Overload detection และ redistribution

#### 2.4 Kanban + Job Integration
**ไฟล์ที่ต้องสร้าง:**
- `pkg/office/kanban_agent.go` (🆕 ใหม่)

**ฟีเจอร์:**
- Agent ดึงงานจาก Kanban เอง (Pull model)
- Sync status ระหว่าง Job <-> Kanban
- Agent workload dashboard

**Deliverables Phase 2:**
- [ ] Jarvis coordinator (basic)
- [ ] Task dependencies system
- [ ] Pipeline templates (content, code)
- [ ] Workload balancing
- [ ] Agent pull from kanban

---

### Phase 3: Intelligence (Weeks 7-10)

#### 3.1 Cross-Model Team Routing
**ไฟล์ที่ต้องสร้าง:**
- `pkg/routing/model_router.go` (🆕 ใหม่)

**ฟีเจอร์:**
```go
type ModelRouter struct {
    cfg *config.Config
}

func (mr *ModelRouter) SelectModelForTask(taskType string) string {
    switch taskType {
    case "coding", "debug", "refactor":
        return "anthropic/claude-sonnet-4"
    case "research", "analysis":
        return "openai/gpt-5"
    case "content", "writing":
        return "google/gemini-pro"
    case "design", "image":
        return "stability/sdxl"
    }
}
```

#### 3.2 Agent Skill Management
**ไฟล์ที่ต้องสร้าง:**
- `pkg/skills/agent_skills.go` (🆕 ใหม่)

**ฟีเจอร์:**
- Per-agent skill registry
- Auto skill discovery จาก ClawHub
- Skill dependency resolution

```go
type AgentSkillManager struct {
    registry *RegistryManager
    agentID  string
}

func (asm *AgentSkillManager) AutoInstallSkills(requirements []string) error {
    // ค้นหาและติดตั้ง skills ที่ขาดจาก ClawHub
}
```

#### 3.3 MCP Tool Delegation
**ไฟล์ที่ต้องสร้าง:**
- `pkg/mcp/agent_mcp.go` (🆕 ใหม่)

**ฟีเจอร์:**
- Agent-specific MCP tool access
- Permission control ตาม role
- Tool usage logging

```go
type AgentMCPManager struct {
    mcpMgr   *mcp.Manager
    agentID  string
    allowedTools []string
}

func (amm *AgentMCPManager) CanUseTool(server, tool string) bool {
    // ตรวจสอบ permission
}
```

#### 3.4 Smart Task Routing
**ไฟล์ที่ต้องสร้าง:**
- `pkg/office/smart_router.go` (🆕 ใหม่)

**ฟีเจอร์:**
- AI-based task classification
- Route งานไปยัง agent ที่เหมาะสมที่สุด
- Fallback chain เมื่อ agent ไม่ว่าง

**Deliverables Phase 3:**
- [ ] Cross-model routing
- [ ] Agent skill auto-install
- [ ] MCP tool delegation
- [ ] Smart task routing

---

### Phase 4: Autonomy (Weeks 11-15)

#### 4.1 Scheduled Agent Workflows
**ไฟล์ที่ต้องสร้าง:**
- `pkg/heartbeat/agent_scheduler.go` (🆕 ใหม่)

**ฟีเจอร์:**
```go
type AgentSchedule struct {
    AgentID      string
    Cron         string
    Task         string
    OutputTo     string    // ส่งผลลัพธ์ให้ agent ไหน
    RAGStore     bool      // บันทึกลง RAG
    KanbanCreate bool      // สร้าง kanban task
}
```

**ตัวอย่าง Scheduled Tasks:**
```json
{
  "agent_schedules": [
    {
      "agent_id": "atlas",
      "cron": "0 * * * *",
      "task": "research_trends",
      "output_to": "scribe",
      "rag_store": true
    },
    {
      "agent_id": "trendy",
      "cron": "0 */3 * * *",
      "task": "scan_social_trends",
      "output_to": "atlas"
    },
    {
      "agent_id": "clawed",
      "cron": "0 2 * * *",
      "task": "review_codebase",
      "create_pr": true,
      "notify": ["sentinel"]
    }
  ]
}
```

#### 4.2 Agent Learning & Feedback
**ไฟล์ที่ต้องสร้าง:**
- `pkg/agent/learning.go` (🆕 ใหม่)

**ฟีเจอร์:**
- Feedback loop จากผลงาน
- Auto-adjust persona จาก interaction
- Performance tracking per agent

#### 4.3 Visual Team Dashboard
**ไฟล์ที่ต้องสร้าง:**
- `pkg/api/ui/team_dashboard.go` (🆕 ใหม่)
- `ui/dashboard/` (🆕 ใหม่ - Frontend)

**ฟีเจอร์:**
```go
type TeamDashboard struct {
    Agents      []AgentStatus
    Tasks       []TaskStatus
    Pipelines   []PipelineStatus
    Messages    []MessagePreview
    Workload    WorkloadChart
}

type AgentStatus struct {
    ID          string
    Name        string
    Status      string  // idle, working, meeting, error
    CurrentTask string
    XP          int     // จาก pkg/office/xp.go
    Level       int
    Workload    float64 // 0-100%
    Avatar      string
}
```

**UI Components:**
- Team Overview (แบบภาพในแนวคิด)
- Kanban Board per Project
- Agent Activity Timeline
- Message Thread Viewer
- Pipeline Progress

#### 4.4 Agent Meeting System
**ไฟล์ที่ต้องสร้าง:**
- `pkg/office/meeting.go` (ขยายจาก existing)

**ฟีเจอร์:**
- Agent หลายตัวคุยกันใน meeting room
- Brainstorming session
- Decision making (voting)

**Deliverables Phase 4:**
- [ ] Scheduled workflows
- [ ] Agent learning system
- [ ] Visual team dashboard
- [ ] Agent meeting system

---

## 🔧 Technical Implementation Details

### Database Schema Extensions

```sql
-- Agent Persona
CREATE TABLE agent_personas (
    agent_id TEXT PRIMARY KEY,
    soul TEXT,
    boundaries TEXT,  -- JSON array
    allowed_tools TEXT, -- JSON array
    memory_scope TEXT,  -- JSON array
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Agent Messages (Mailbox)
CREATE TABLE agent_messages (
    id TEXT PRIMARY KEY,
    from_agent TEXT NOT NULL,
    to_agent TEXT NOT NULL,
    message_type TEXT NOT NULL,
    content TEXT NOT NULL,
    priority TEXT DEFAULT 'normal',
    thread_id TEXT,
    job_id TEXT,
    metadata TEXT,  -- JSON
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Task Dependencies
CREATE TABLE task_dependencies (
    task_id TEXT NOT NULL,
    depends_on TEXT NOT NULL,
    PRIMARY KEY (task_id, depends_on)
);

-- Pipeline Templates
CREATE TABLE pipeline_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    stages TEXT,  -- JSON array
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Agent Performance
CREATE TABLE agent_performance (
    agent_id TEXT NOT NULL,
    date TEXT NOT NULL,  -- YYYY-MM-DD
    tasks_completed INTEGER DEFAULT 0,
    tasks_failed INTEGER DEFAULT 0,
    avg_quality_score REAL,
    xp_earned INTEGER DEFAULT 0,
    PRIMARY KEY (agent_id, date)
);
```

### Configuration Structure

```json
{
  "agents": {
    "list": [
      {
        "id": "atlas",
        "name": "Atlas",
        "role": "researcher",
        "department": "research",
        "model": { "primary": "openai/gpt-5" },
        "persona": {
          "soul": "คุณคือนักวิจัยข้อมูลที่ละเอียดและรอบคอบ",
          "boundaries": ["ไม่เขียนโค้ด", "ไม่ตัดสินใจแทนผู้ใช้"],
          "allowed_tools": ["web_search", "rag_search", "message"],
          "memory_scope": ["research", "web", "news"]
        },
        "schedule": "0 * * * *",
        "skills": ["web_search", "data_analysis"],
        "mcp_servers": ["tavily", "brave_search"]
      }
    ]
  },
  "agent_team": {
    "coordinator": "jarvis",
    "mailbox_enabled": true,
    "auto_assign": true,
    "workload_threshold": 3,
    "pipelines": {
      "content_creation": [
        { "role": "researcher", "task": "research_topic" },
        { "role": "writer", "task": "write_content" },
        { "role": "reviewer", "task": "review_content" }
      ],
      "code_review": [
        { "role": "developer", "task": "review_code" },
        { "role": "qa", "task": "quality_check" }
      ]
    }
  },
  "agent_schedules": [
    {
      "agent_id": "atlas",
      "cron": "0 * * * *",
      "task": "research_trends",
      "output_to": "scribe"
    }
  ]
}
```

---

## 📊 Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Task Completion Rate** | >90% | งานที่เสร็จ / งานทั้งหมด |
| **Context Isolation** | 100% | Agent A ไม่เห็น memory ของ Agent B โดยไม่ได้รับอนุญาต |
| **Inter-Agent Communication** | <2s | Latency ของ mailbox system |
| **Auto-Assignment Accuracy** | >80% | งานที่ assign ถูก agent ที่เหมาะสม |
| **System Resource** | <50MB | RAM ทั้งหมดของ Agent Team |

---

## 🎨 UI/UX Design - Virtual Office Interface

### Design Concept: "Pixel Art Office Simulator"

UI จะออกแบบตามแนวคิดในภาพตัวอย่าง - **Pixel Art Virtual Office** ที่ผู้ใช้สามารถ:
- เห็น agent ทุกตัวใน office แบบ real-time
- คลิกที่ agent เพื่อดูรายละเอียด/ตั้งค่า
- ลาก agent เข้า conference room เพื่อให้คุยกัน
- ดูสถานะงานผ่าน visual cues (ไฟ, animation)

### Main Layout Structure

```
┌────────────────────────────────────────────────────────────────────────────┐
│  🏢 PICOCLAW TEAM v1.0                    [💬 Messages: 3]    [⚙️ Settings] │
├────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────────┐  ┌──────────────────────────────────────────────────┐│
│  │  📋 NAVIGATION   │  │         🏢 VIRTUAL OFFICE (Pixel Art)            ││
│  ├──────────────────┤  │                                                  ││
│  │ 🏠 Dashboard     │  │   ┌─────────────┐  ┌───────────────────────────┐ ││
│  │ 👥 Agents        │  │   │ CONFERENCE  │  │     JARVIS OFFICE         │ ││
│  │ 📊 Kanban        │  │   │    ROOM     │  │    [🖥️][🖥️][🖥️]          │ ││
│  │ 💬 Messages      │  │   │  [👤][👤]   │  │       👤 Jarvis           │ ││
│  │ 🔧 Workflows     │  │   │  Meeting    │  │      [Typing...]          │ ││
│  │ ⏰ Schedule      │  │   └─────────────┘  └───────────────────────────┘ ││
│  │ 🧠 Memory        │  │                                                  ││
│  │ ⚙️ Settings      │  │   ┌──────────┐ ┌──────────┐ ┌────────────────┐  ││
│  │ 📈 Analytics     │  │   │  ATLAS   │ │  SCRIBE  │ │     CLAWED     │  ││
│  └──────────────────┘   │   │   👤     │ │    👤    │ │       👤       │  ││
│                         │   │ [Research│ │ [Writing │ │   [Coding...]  │  ││
│  ┌──────────────────┐   │   │  80%]    │ │   50%]   │ │                │  ││
│  │ 🟢 TEAM STATUS   │   │   └──────────┘ └──────────┘ └────────────────┘  ││
│  ├──────────────────┤   │                                                  ││
│  │ ● Atlas (Busy)   │   │   ┌──────────┐ ┌──────────┐ ┌────────────────┐  ││
│  │ ● Scribe (Busy)  │   │   │ SENTINEL │ │  TRENDY  │ │     NOVA       │  ││
│  │ ○ Clawed (Idle)  │   │   │   👤     │ │    👤    │ │       👤       │  ││
│  │ ○ Sentinel       │   │   │ [Monitor │ │ [Scanning│ │ [Designing]    │  ││
│  └──────────────────┘   │   │   20%]   │ │   30%]   │ │                │  ││
│                         │   └──────────┘ └──────────┘ └────────────────┘  ││
│                         └──────────────────────────────────────────────────┘│
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │ 📋 ACTIVE TASKS        |  💬 RECENT MESSAGES  |  📊 PIPELINE STATUS │   │
│  │ [Kanban Preview]       |  [Message List]      |  [Progress Bars]    │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────────────────────┘
```

### Settings Modal Structure

```
┌─────────────────────────────────────────────────────────────────┐
│  ⚙️ AGENT CONFIGURATION - Atlas                              [X] │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────────────────────────────────┐ │
│  │  🎨 AVATAR   │  │  📋 GENERAL SETTINGS                    │ │
│  │              │  ├──────────────────────────────────────────┤ │
│  │  [👤 Atlas]  │  │  Name:        [ Atlas               ]   │ │
│  │  [Change]    │  │  Role:        [ Researcher ▼        ]   │ │
│  │              │  │  Department:  [ Research ▼          ]   │ │
│  └──────────────┘  └──────────────────────────────────────────┘ │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  🤖 MODEL CONFIGURATION                                     │ │
│  ├────────────────────────────────────────────────────────────┤ │
│  │  Primary Model:    [ openai/gpt-5.2 ▼ ]  [Test] [Verify]  │ │
│  │  Fallback Models:  [ anthropic/claude-sonnet-4.6 ✓ ]      │ │
│  │                    [ google/gemini-pro □ ]                │ │
│  │                    [ + Add Fallback ]                     │ │
│  │                                                            │ │
│  │  Temperature:      [====●====] 0.7                        │ │
│  │  Max Tokens:       [ 8192 ▼ ]                             │ │
│  │  Top P:            [====●====] 0.9                        │ │
│  │                                                            │ │
│  │  💰 Estimated Cost: $0.002 / 1K tokens                   │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  🎭 PERSONA CONFIGURATION                                   │ │
│  ├────────────────────────────────────────────────────────────┤ │
│  │  Soul/Personality:                                         │ │
│  │  ┌──────────────────────────────────────────────────────┐ │ │
│  │  │ คุณคือนักวิจัยข้อมูลที่ละเอียดและรอบคอบ...          │ │ │
│  │  │                                                      │ │ │
│  │  └──────────────────────────────────────────────────────┘ │ │
│  │                                                            │ │
│  │  Allowed Tools:  ☑ web_search  ☑ rag_search  ☐ exec      │ │
│  │  Disallowed:     ☐ write_file ☑ delete_file              │ │
│  │                                                            │ │
│  │  Memory Scope:   ☑ research  ☑ web  ☑ news  ☐ code       │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  📅 SCHEDULE & AUTOMATION                                   │ │
│  ├────────────────────────────────────────────────────────────┤ │
│  │  ☐ Enable Scheduled Tasks                                 │ │
│  │     Cron: [ 0 * * * * ] (Every hour)                      │ │
│  │     Task: [ research_trends ▼ ]                           │ │
│  │     Output to: [ scribe ▼ ]                               │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  [💾 Save]  [🔄 Reset to Default]  [🗑️ Delete Agent]           │
└─────────────────────────────────────────────────────────────────┘
```

### Visual Assets Requirements

#### 1. Pixel Art Style Assets

**Character Sprites (32x32 or 64x64 pixels):**
- Idle animation (2-4 frames)
- Working animation (typing, reading, thinking)
- Walking animation (สำหรับเดินไป meeting room)
- Status indicators (online dot, busy indicator)

**Office Furniture:**
- Desks (หลายแบบ: ทั่วไป, standing desk, executive)
- Chairs (หมุนได้, มี animation)
- Computers (monitor, laptop, multi-monitor setup)
- Decorations (ต้นไม้, นาฬิกา, whiteboard)

**Room Backgrounds:**
- Jarvis Office (high-tech, multiple screens)
- Conference Room (โต๊ะประชุม, projector)
- Research Corner (หนังสือ, globe)
- Dev Zone (multi-monitor, coffee cup)

#### 2. Asset Sources

**Free/Open Source:**
1. **itch.io** (pixel art asset packs)
   - "Office Pixel Art" packs
   - "Character Sprite Sheets"
   - License: ส่วนใหญ่ CC0 หรือ Commercial allowed

2. **OpenGameArt.org**
   - LPC (Liberated Pixel Cup) assets
   - 16x16 or 32x32 character bases

3. **Craftpix.net** (free section)
   - Office/Modern interior packs

4. **Generate with AI:**
   - DALL-E / Midjourney: "pixel art office worker, 32x32, sprite sheet"
   - Aseprite: สร้าง/แก้ไขเอง

**Asset Management Structure:**
```
assets/
├── agents/
│   ├── sprites/
│   │   ├── atlas/
│   │   │   ├── idle.png
│   │   │   ├── working.png
│   │   │   └── avatar.png
│   │   ├── jarvis/
│   │   ├── clawed/
│   │   └── ...
│   └── configs/
│       ├── atlas.json
│       └── ...
├── furniture/
│   ├── desks/
│   ├── chairs/
│   └── computers/
├── rooms/
│   ├── backgrounds/
│   └── overlays/
└── ui/
    ├── icons/
    ├── buttons/
    └── panels/
```

#### 3. Asset Download System

```go
// pkg/assets/downloader.go
type AssetDownloader struct {
    baseURL string  // CDN หรือ GitHub releases
}

func (ad *AssetDownloader) DownloadAgentPack(agentID string) error {
    // Download sprite pack สำหรับ agent
    // ถ้าไม่มี ใช้ default sprite
}

func (ad *AssetDownloader) GetDefaultAvatar() string {
    // คืนค่า default avatar (robot icon หรือ initials)
}
```

### UI Interaction Design

#### Agent Interactions:
| Action | Result |
|--------|--------|
| **Click Agent** | เปิด Settings Modal |
| **Double Click** | เปิด Chat Interface กับ Agent |
| **Drag to Conference Room** | สร้าง Meeting session |
| **Right Click** | Context Menu (Assign Task, Send Message, View Stats) |
| **Hover** | แสดง Tooltip (Status, Current Task, Model) |

#### Visual Status Cues:
| Status | Visual |
|--------|--------|
| 🟢 Online/Working | ไฟเขียว, Animation typing/reading |
| 🟡 Busy | ไฟเหลือง, Animation focused |
| 🔴 Idle/Offline | ไฟแดง, Static pose |
| 🔵 In Meeting | ไฟน้ำเงิน, อยู่ใน Conference Room |
| ⚡ High Priority | Border กระพริบสีแดง |

### Responsive Breakpoints

| Screen | Layout |
|--------|--------|
| **Desktop (1920x1080)** | Full Virtual Office + Sidebar |
| **Laptop (1366x768)** | Compact Office, Collapsible Sidebar |
| **Tablet (1024x768)** | List View + Modal Office |
| **Mobile (375x667)** | List View only, Simple status |

### Chat Sidebar Integration

Chat interface จะถูกรวมเข้ากับ Virtual Office UI โดยเป็น **Collapsible Sidebar** ทางด้านขวา:

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  🏢 PICOCLAW TEAM v1.0                                    [💬] [⚙️] [👤]     │
├──────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────────┐  ┌──────────────────────────────┐  ┌─────────────────┐ │
│  │  📋 NAVIGATION   │  │      🏢 VIRTUAL OFFICE       │  │   💬 CHAT       │ │
│  ├──────────────────┤  │                              │  ├─────────────────┤ │
│  │ 🏠 Dashboard     │  │   [Office Canvas Area]       │  │ ▼ Main Chat     │ │
│  │ 👥 Agents        │  │                              │  ├─────────────────┤ │
│  │ 📊 Kanban        │  │   👤 Atlas    👤 Scribe      │  │ User: สวัสดี    │ │
│  │ 💬 Messages      │  │   [Working]   [Writing]      │  │ Atlas: สวัสดีค่ะ│ │
│  │ 🔧 Workflows     │  │                              │  │ มีอะไรให้ช่วย   │ │
│  │ ⏰ Schedule      │  │   👤 Clawed                  │  │ ไหมคะ?          │ │
│  │ 🧠 Memory        │  │   [Coding...]                │  │                 │ │
│  └──────────────────┘  │                              │  │ User: ช่วยหา    │ │
│                        │   👤 Jarvis  [Coordinator]   │  │ ข้อมูล AI หน่อย  │ │
│  ┌──────────────────┐  │   [Monitoring]               │  │                 │ │
│  │ 🟢 TEAM STATUS   │  │                              │  │ Atlas: ได้ค่ะ   │ │
│  │ ● Atlas (Busy)   │  └──────────────────────────────┘  │ [กำลังค้นหา...] │ │
│  │ ● Scribe (Busy)  │                                    │                 │ │
│  │ ○ Clawed (Idle)  │                                    │ ┌─────────────┐ │ │
│  └──────────────────┘                                    │ │ ส่งข้อความ  │ │ │
│                                                          │ │ หรือ /คำสั่ง│ │ │
│                                                          │ └─────────────┘ │ │
│                                                          │ [📎] [😊] [📤]  │ │
│                                                          └─────────────────┘ │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Chat Interface Features

#### 1. Chat Selector (ใน Sidebar Header)
```
┌─────────────────┐
│ ▼ Chat: Atlas   │ ← Dropdown เลือก chat
├─────────────────┤
│ 👤 Main Chat    │ ← Default (คุยกับระบบหลัก)
│ ─────────────── │
│ 👤 Atlas        │ ← Agent แต่ละตัว
│ 👤 Scribe       │
│ 👤 Clawed       │
│ 👤 Jarvis       │
│ ─────────────── │
│ # General       │ ← Group chat (future)
└─────────────────┘
```

#### 2. Chat Modes

**Mode 1: Main Chat (Default)**
- คุยกับ Coordinator (Jarvis) หรือระบบหลัก
- Jarvis จะเป็นคนวิเคราะห์และมอบหมายงานให้ agent อื่น
- เหมือนใช้งาน PicoClaw ปกติ

**Mode 2: Direct Agent Chat**
- คลิกที่ agent ใน office → Chat sidebar เปลี่ยนเป็น agent นั้น
- คุยกับ agent โดยตรง ไม่ผ่าน coordinator
- Agent จะตอบตาม persona ของตัวเอง
- สามารถสั่งงานเฉพาะทางได้เลย (เช่น คุยกับ Clawed เรื่อง code โดยตรง)

**Mode 3: Multi-Agent Chat (Meeting)**
- ลากหลาย agent เข้า Conference Room
- Chat จะกลายเป็น Group Chat
- ทุก agent เห็นข้อความกัน

#### 3. Message Types & Commands

```typescript
interface ChatMessage {
  id: string;
  chatId: string;      // "main", "atlas", "meeting_123"
  sender: {
    type: 'user' | 'agent' | 'system';
    id: string;        // "user", "atlas", "system"
    name: string;
    avatar: string;
  };
  content: string;
  contentType: 'text' | 'markdown' | 'code' | 'image';
  metadata?: {
    taskId?: string;
    jobId?: string;
    toolCalls?: ToolCall[];
    isThinking?: boolean;
  };
  timestamp: Date;
  threadId?: string;   // สำหรับ reply/thread
}
```

**Slash Commands:**
```
/task เขียนบทความ AI      → สร้าง task ใหม่
/assign @atlas            → มอบหมายให้ agent
/status                   → ดูสถานะ agent
/meeting @atlas @scribe   → เรียกประชุม
/file                     → แนบไฟล์
/clear                    → ล้าง chat history
```

#### 4. Visual Feedback

**Agent Typing Indicator:**
```
Atlas กำลังพิมพ์...
[●   ]
[ ●  ]  ← Animation
[  ● ]
```

**Task Cards in Chat:**
```
┌─────────────────────────────────┐
│ 📝 Task Created                 │
│ เขียนบทความ AI                 │
│ สถานะ: ⏳ Pending               │
│ ผู้รับผิดชอบ: @Scribe          │
│ [ดูรายละเอียด] [ยกเลิก]        │
└─────────────────────────────────┘
```

**Tool Execution Preview:**
```
┌─────────────────────────────────┐
│ 🔧 Clawed กำลังใช้ tool:        │
│ write_file                      │
│                                 │
│ 📄 main.go                      │
│ ┌─────────────────────────────┐ │
│ │ package main                │ │
│ │ func main() { ... }         │ │
│ └─────────────────────────────┘ │
│ [✅ ยืนยัน] [❌ ยกเลิก]        │
└─────────────────────────────────┘
```

### Interaction Flows

#### Flow 1: คลิก Agent เพื่อแชท
1. User คลิกที่ Agent Atlas ใน Virtual Office
2. Chat Sidebar เปิด (ถ้ายังไม่เปิด)
3. Chat Selector เปลี่ยนเป็น "Atlas"
4. Load chat history กับ Atlas
5. User พิมพ์ข้อความ → ส่งไป Atlas โดยตรง

#### Flow 2: สั่งงานผ่าน Main Chat
1. User อยู่ใน Main Chat (Default)
2. User พิมพ์: "ช่วยหาข้อมูล AI Trends หน่อย"
3. Jarvis (Coordinator) รับ message
4. Jarvis สร้าง task → ส่งต่อให้ Atlas
5. Atlas ทำงาน → ส่งผลลัพธ์กลับมา
6. User เห็นผลลัพธ์ใน Main Chat

#### Flow 3: Direct Task Assignment
1. User คลิกที่ Clawed → Chat เปลี่ยนเป็น Clawed
2. User พิมพ์: "ช่วย refactor ไฟล์นี้หน่อย"
3. Clawed ทำงานทันที (ไม่ผ่าน coordinator)
4. Clawed ขออนุมัติก่อนแก้ไขไฟล์ (ถ้า setting เปิด)
5. User ยืนยัน → Clawed ทำงานต่อ

### Chat Sidebar States

```typescript
type ChatSidebarState = {
  isOpen: boolean;
  width: number;           // 300-500px (resizable)
  activeChatId: string;    // "main", "atlas", "meeting_123"
  recentChats: string[];   // ["main", "atlas", "scribe"]
  unreadCounts: Map<string, number>;
  isMinimized: boolean;    // true = แสดงแค่ icon
}
```

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + /` | Toggle Chat Sidebar |
| `Ctrl/Cmd + 1` | Switch to Main Chat |
| `Ctrl/Cmd + 2-9` | Switch to recent agent chats |
| `Ctrl/Cmd + Shift + M` | New Meeting with selected agents |
| `Esc` | Close current panel/modal |
| `↑` | Edit last message (like terminal) |
| `@` + Tab | Mention agent autocomplete |

### Mobile Responsive

**Mobile (< 768px):**
```
┌─────────────────────┐
│  🏢 PicoClaw Team   │
├─────────────────────┤
│  [Virtual Office    │
│   Scrollable]       │
│                     │
│   👤 Atlas          │
│   👤 Scribe         │
│                     │
│  [💬] [👥] [⚙️]    │ ← Bottom Nav
└─────────────────────┘

เมื่อกด 💬:
┌─────────────────────┐
│  ← Atlas            │
├─────────────────────┤
│  Chat History       │
│  ───────────────    │
│  User: Hi           │
│  Atlas: Hello!      │
│                     │
│  ┌───────────────┐  │
│  │ พิมพ์ข้อความ │  │
│  └───────────────┘  │
└─────────────────────┘
```

### Theme Customization

```json
{
  "ui_theme": {
    "name": "cyberpunk_office",
    "colors": {
      "background": "#0a0e27",
      "panel": "#1a1f3a",
      "accent": "#00d9ff",
      "text": "#e0e6ed",
      "status_online": "#00ff88",
      "status_busy": "#ffaa00",
      "status_offline": "#ff4444"
    },
    "pixel_scale": 2,
    "animation_speed": 1.0
  }
}
```

---

## 🖥️ UI Implementation Plan

### Phase UI-1: Core UI Framework (Week 1-2)

**Tech Stack:**
- **Frontend:** React + TypeScript + Vite
- **State Management:** Zustand
- **Styling:** TailwindCSS + CSS Modules
- **Canvas/Game:** PixiJS (สำหรับ Virtual Office rendering)
- **Icons:** Lucide React

**ไฟล์ที่ต้องสร้าง:**
```
ui/
├── src/
│   ├── components/
│   │   ├── Layout/
│   │   │   ├── MainLayout.tsx
│   │   │   ├── Sidebar.tsx              # Left navigation
│   │   │   ├── ChatSidebar.tsx          # NEW: Right chat sidebar
│   │   │   ├── Header.tsx
│   │   │   └── ResizablePanel.tsx       # NEW: Resizable container
│   │   ├── VirtualOffice/
│   │   │   ├── VirtualOfficeCanvas.tsx  # PixiJS canvas
│   │   │   ├── AgentSprite.tsx          # Agent sprite with click handler
│   │   │   ├── RoomBackground.tsx
│   │   │   └── InteractionLayer.tsx     # Click detection for agents
│   │   ├── Chat/
│   │   │   ├── ChatContainer.tsx        # NEW: Main chat container
│   │   │   ├── ChatHeader.tsx           # NEW: Chat selector + actions
│   │   │   ├── MessageList.tsx          # NEW: Scrollable message list
│   │   │   ├── MessageBubble.tsx        # NEW: Individual message
│   │   │   ├── MessageInput.tsx         # NEW: Input with commands
│   │   │   ├── TypingIndicator.tsx      # NEW: Agent typing animation
│   │   │   ├── TaskCard.tsx             # NEW: Inline task display
│   │   │   ├── ToolPreview.tsx          # NEW: Tool execution preview
│   │   │   └── ChatSelector.tsx         # NEW: Dropdown to switch chat
│   │   ├── AgentConfig/
│   │   │   ├── AgentConfigModal.tsx
│   │   │   ├── ModelSelector.tsx
│   │   │   ├── PersonaEditor.tsx
│   │   │   └── ScheduleConfig.tsx
│   │   └── Common/
│   │       ├── PixelButton.tsx
│   │       ├── StatusIndicator.tsx
│   │       ├── LoadingPixel.tsx
│   │       └── SlashCommandMenu.tsx     # NEW: Command autocomplete
│   ├── hooks/
│   │   ├── useAgentRegistry.ts
│   │   ├── useVirtualOffice.ts
│   │   ├── useModelConfig.ts
│   │   ├── useChat.ts                   # NEW: Chat management
│   │   ├── useWebSocket.ts              # NEW: Real-time connection
│   │   └── useAgentInteraction.ts       # NEW: Click agent → chat
│   ├── stores/
│   │   ├── agentStore.ts
│   │   ├── uiStore.ts
│   │   ├── officeStore.ts
│   │   └── chatStore.ts                 # NEW: Chat state management
│   ├── types/
│   │   ├── agent.ts
│   │   ├── office.ts
│   │   ├── ui.ts
│   │   └── chat.ts                      # NEW: Chat types
│   └── utils/
│       ├── api.ts
│       ├── assetLoader.ts
│       └── chatCommands.ts              # NEW: Slash command handlers
```

**Deliverables:**
- [ ] Layout พื้นฐาน (Sidebar + Main Area + Chat Sidebar)
- [ ] Resizable panels (Left sidebar + Chat sidebar)
- [ ] PixiJS Canvas setup
- [ ] Connection กับ Backend API

### Phase UI-2: Virtual Office Rendering (Week 3-4)

**Features:**
1. **Office Canvas:**
   - Render room backgrounds
   - Place agent sprites ตาม config position
   - Camera pan/zoom

2. **Agent Sprites:**
   - Load sprite sheets
   - Animation states (idle, working, walking)
   - Status indicators (online dot, progress bar)

3. **Interactions:**
   - Click to select
   - Drag to move (ใน config mode)
   - Right-click context menu

**Deliverables:**
- [ ] Render static office layout
- [ ] Agent sprite animation
- [ ] Basic interactions (click, hover)
- [ ] Chat Sidebar layout (เปิด/ปิด/resize ได้)
- [ ] Click agent → Open chat with that agent

### Phase UI-3: Agent Configuration UI (Week 5-6)

**Features:**
1. **Model Selection:**
   - Default: `kimi-coding/kimi-for-coding` (ใช้ได้ทันที)
   - Dropdown แสดง available models สำหรับเปลี่ยน
   - Test connection button
   - Cost estimation display
   - Fallback model chain
   - Reset to default button

2. **Persona Editor:**
   - Text editor สำหรับ soul/personality
   - Tool permission toggles
   - Memory scope selection

3. **Visual Config:**
   - Position editor (x,y coordinates)
   - Avatar selector: Use custom assets / Fallback (emoji/generated)
   - Room assignment
   - Animation preview

**Deliverables:**
- [ ] Agent config modal
- [ ] Model selector with test
- [ ] Persona editor
- [ ] Visual position editor
- [ ] Direct agent chat (คลิก agent → chat)
- [ ] Chat history persistence

### Phase UI-4: Advanced Features (Week 7-8)

**Features:**
1. **Real-time Updates:**
   - WebSocket connection
   - Agent status sync
   - Message notifications

2. **Kanban Integration:**
   - Drag-and-drop tasks
   - Column management
   - Agent workload view

3. **Meeting System:**
   - Drag agents to conference room
   - Meeting session UI
   - Shared context display

**Deliverables:**
- [ ] WebSocket real-time sync
- [ ] Kanban board integration
- [ ] Meeting room UI
- [ ] Slash commands (/task, /assign, /meeting)
- [ ] Task cards in chat
- [ ] Tool execution preview
- [ ] Multi-agent chat (meeting mode)

### Phase UI-5: Themes & Polish (Week 9-10)

**Features:**
1. **Theme System:**
   - Multiple themes (Cyberpunk, Corporate, Minimal)
   - Color customization
   - Font selection

2. **Fallback System:**
   - Auto-generate avatars (initials/gradient) ถ้าไม่มี custom sprites
   - Emoji mode สำหรับ minimal setup
   - SVG avatar generation

3. **Asset Mapping UI:**
   - แสดง assets ที่มีอยู่ในโฟลเดอร์
   - Map assets กับ agents
   - Preview animations

4. **Responsive:**
   - Tablet layout
   - Mobile-optimized view

**Deliverables:**
- [ ] Theme switcher
- [ ] Fallback avatar generation (SVG initials/emoji)
- [ ] Asset mapping UI (เลือกไฟล์ที่มีอยู่)
- [ ] Responsive layouts

---

## 📊 API Endpoints สำหรับ UI

```go
// pkg/api/ui/handlers.go เพิ่ม endpoints:

// Agent Management
GET    /api/agents                    // List all agents with status
GET    /api/agents/:id                // Get agent details
PUT    /api/agents/:id                // Update agent config
POST   /api/agents                    // Create new agent
DELETE /api/agents/:id                // Delete agent

// Model Management
GET    /api/models                    // List available models
GET    /api/models/default            // Get default model (kimi-coding)
POST   /api/models/test               // Test model connection
GET    /api/models/:id/cost           // Get cost estimation
POST   /api/models/reset-to-default   // Reset agent to use default model

// Virtual Office
GET    /api/office/state              // Get current office state
GET    /api/office/agents/positions   // Get agent positions
PUT    /api/office/agents/:id/position // Update agent position
GET    /api/office/rooms              // List rooms

// Chat System (NEW)
GET    /api/chats                     // List user's chats
POST   /api/chats                     // Create new chat (meeting)
GET    /api/chats/:id                 // Get chat details
GET    /api/chats/:id/messages        // Get chat messages (paginated)
POST   /api/chats/:id/messages        // Send message
PUT    /api/chats/:id/messages/:msgId // Edit message
DELETE /api/chats/:id/messages/:msgId // Delete message
POST   /api/chats/:id/read            // Mark as read
GET    /api/chats/:id/unread          // Get unread count
POST   /api/chats/:typing             // Typing indicator

// Direct Agent Chat
GET    /api/agents/:id/chat           // Get or create direct chat with agent
POST   /api/agents/:id/chat/message   // Send message to specific agent

// Real-time
WS     /ws/office                     // WebSocket for real-time updates
WS     /ws/chat/:id                   // WebSocket for specific chat (optional)

// Assets
GET    /api/assets/agents/:id/sprites // Get agent sprites
POST   /api/assets/upload             // Upload custom avatar
GET    /api/assets/themes             // List available themes
```

### Chat WebSocket Events

```typescript
// Client → Server
type ClientEvents = {
  'chat:message': { chatId: string; content: string; replyTo?: string }
  'chat:typing': { chatId: string; isTyping: boolean }
  'chat:read': { chatId: string; messageIds: string[] }
  'agent:click': { agentId: string }  // เมื่อ user คลิก agent
}

// Server → Client
type ServerEvents = {
  'chat:message': ChatMessage
  'chat:typing': { chatId: string; agentId: string; isTyping: boolean }
  'chat:status': { chatId: string; status: 'active' | 'offline' | 'busy' }
  'agent:status': { agentId: string; status: AgentStatus }
  'agent:thinking': { agentId: string; isThinking: boolean; currentTask?: string }
  'task:created': { taskId: string; title: string; assignee: string }
  'task:updated': { taskId: string; status: TaskStatus }
}
```

---

## 🎮 User Workflows

### Workflow 1: สร้าง Agent ใหม่
1. คลิก "+ Add Agent" ที่ sidebar
2. เลือก Template (Researcher, Developer, etc.)
3. ตั้งค่า Name, Role, Department
4. เลือก Model (primary + fallbacks)
5. ปรับ Persona (soul, tools, memory)
6. เลือก/Upload Avatar
7. วางตำแหน่งใน Virtual Office
8. Save

### Workflow 2: มอบหมายงาน
1. คลิกขวาที่ Agent หรือลากงานจาก Kanban
2. เลือก "Assign Task"
3. พิมพ์รายละเอียดงาน
4. (Optional) ตั้งเวลา Deadline
5. (Optional) เพิ่ม Dependencies
6. Confirm → Task ปรากฏใน Kanban

### Workflow 3: สร้าง Meeting
1. ลาก Agent หลายตัวเข้า Conference Room
2. ระบบสร้าง Meeting session
3. พิมพ์หัวข้อที่ต้องการหารือ
4. Agent ทั้งหมดได้รับ context พร้อมกัน
5. ดูผลลัพธ์แบบ real-time

### Workflow 4: เปลี่ยน Model
1. คลิกที่ Agent → เปิด Config
2. เลือก tab "Model"
3. เปลี่ยน Primary Model
4. คลิก "Test" เพื่อทดสอบ connection
5. ดู Cost estimation ที่เปลี่ยนไป
6. Save → Agent ใช้ model ใหม่ทันที

### Workflow 4b: เลือก Avatar (Manual Assets)
1. เตรียม assets ไว้ใน `~/.picoclaw/assets/agents/{agent_id}/`
2. คลิกที่ Agent → เปิด Config
3. เลือก tab "Appearance"
4. ระบบแสดง assets ที่มีอยู่:
   - Custom sprites (ถ้ามีในโฟลเดอร์)
   - Fallback options (emoji, generated)
5. เลือก:
   - "Use Custom Sprite" → เลือกไฟล์จาก assets folder
   - "Use Fallback" → ใช้ emoji หรือ initials
   - "Upload New" → อัพโหลดไฟล์ใหม่เข้า assets folder

### Workflow 5: แชทกับ Agent โดยตรง (ผ่าน Chat Sidebar)
1. User อยู่ใน Virtual Office
2. User คลิกที่ Agent Atlas ใน office
3. Chat Sidebar (ขวา) เปิดขึ้น (ถ้ายังไม่เปิด)
4. Chat selector เปลี่ยนเป็น "Atlas"
5. แสดง chat history กับ Atlas
6. User พิมพ์: "ช่วยหาข้อมูล Go Generics หน่อย"
7. Atlas ตอบกลับทันที (ใช้ persona ของ Atlas)
8. ผลลัพธ์แสดงใน Chat Sidebar

### Workflow 6: สั่งงานผ่าน Main Chat (ผ่าน Coordinator)
1. User อยู่ใน Main Chat (Default)
2. User พิมพ์: "เขียนบทความ AI Trends"
3. Jarvis (Coordinator) วิเคราะห์
4. Jarvis สร้าง task → ส่งต่อ Atlas (research) → Scribe (write)
5. User เห็น pipeline progress ใน chat
6. ผลลัพธ์สุดท้ายแสดงใน Main Chat

### Workflow 7: เรียกประชุม (Multi-Agent Chat)
1. User ลาก Atlas และ Scribe เข้า Conference Room
2. Chat Sidebar เปลี่ยนเป็น "Meeting: Atlas, Scribe"
3. User พิมพ์: "คุยกันเรื่องบทความ AI"
4. Atlas และ Scribe เห็นข้อความกัน
5. Atlas เสนอข้อมูล → Scribe สรุป → แสดงผลพร้อมกัน

---

## 🔧 Configuration Schema (Split)

> **แยก Config เป็น 2 ไฟล์:**
> - `~/.picoclaw/config.json` - ระบบหลัก (agents, models, persona)
> - `ui/src/config/agentUiConfig.ts` - UI/Visual (sprites, positions, colors)

### Default Model Configuration

ทุก Agent จะใช้ `kimi-coding/kimi-for-coding` เป็นค่าเริ่มต้น:

```json
{
  "agents": {
    "defaults": {
      "model": {
        "primary": "kimi-coding/kimi-for-coding",
        "params": {
          "temperature": 0.7,
          "max_tokens": 4096,
          "top_p": 0.9
        }
      }
    },
    "list": [
      {
        "id": "atlas",
        "model": {
          "primary": "kimi-coding/kimi-for-coding"
        }
      }
    ]
  }
}
```

**การทำงาน:**
1. ถ้าไม่ระบุ model ใน agent config → ใช้ defaults.model
2. ถ้าระบุ model.primary → ใช้ค่าที่ระบุ
3. ถ้าระบุ model.fallbacks → ใช้เมื่อ primary fail

**Override Model สำหรับ Agent ใด Agent หนึ่ง:**
```json
{
  "id": "clawed",
  "model": {
    "primary": "anthropic/claude-sonnet-4.6",
    "fallbacks": ["kimi-coding/kimi-for-coding"]
  }
}
```

---

```json
{
  "version": "2.0",
  "agents": {
    "defaults": {
      "model": {
        "primary": "kimi-coding/kimi-for-coding",
        "params": {
          "temperature": 0.7,
          "max_tokens": 4096
        }
      }
    },
    "list": [
      {
        "id": "atlas",
        "name": "Atlas",
        "role": "researcher",
        "department": "research",
        "is_active": true,
        
        "model": {
          "primary": "kimi-coding/kimi-for-coding",
          "fallbacks": [],
          "params": {
            "temperature": 0.7,
            "max_tokens": 8192,
            "top_p": 0.9
          }
        },
        "_model_note": "ค่าเริ่มต้นใช้ kimi-coding สำหรับทุก agent สามารถเปลี่ยนได้ใน UI หรือแก้ไข config",
        
        "persona": {
          "soul": "คุณคือนักวิจัยข้อมูลที่ละเอียดและรอบคอบ...",
          "boundaries": ["ไม่เขียนโค้ด", "ไม่ตัดสินใจแทนผู้ใช้"],
          "allowed_tools": ["web_search", "rag_search", "message", "memory_store"],
          "disallowed_tools": ["write_file", "exec", "shell"],
          "memory_scope": ["research", "web", "news", "shared"],
          "language": "th",
          "tone": "professional",
          "response_style": "detailed"
        },
        
        "capabilities": ["web_research", "data_analysis", "summarization"],
        
        "schedule": {
          "enabled": true,
          "cron": "0 * * * *",
          "tasks": ["research_trends", "update_knowledge_base"],
          "output_to": "scribe",
          "conditions": {
            "only_when_idle": true,
            "max_concurrent": 1
          }
        },
        
        "_note": "UI config แยกอยู่ที่ ui/src/config/agentUiConfig.ts",
        "_ui_config_location": "ui/src/config/agentUiConfig.ts",
        
        "created_at": "2026-03-09T00:00:00Z",
        "updated_at": "2026-03-09T12:00:00Z"
      }
    ]
  },
  
  "_office_ui_note": "Office/Visual config อยู่ที่ ui/src/config/agentUiConfig.ts",
  "_ui_files": [
    "ui/src/config/agentUiConfig.ts - Agent visual config",
    "ui/src/config/agentUiConfig.ts - Room layouts"
  ],
  
  "agent_team": {
    "coordinator": "jarvis",
    "auto_coordination": true,
    "mailbox_enabled": true,
    "auto_assign": true,
    "workload_threshold": 3,
    "load_balancing": "round_robin",
    
    "pipelines": {
      "content_creation": {
        "stages": [
          { "role": "researcher", "task": "research_topic", "output_as": "research_data" },
          { "role": "writer", "task": "write_content", "input_from": "research_data" },
          { "role": "reviewer", "task": "review_content" }
        ]
      }
    }
  }
}
```

---

## 🚀 Quick Start (สำหรับผู้ใช้)

```bash
# 1. สร้าง Agent Team
picoclaw team init --name "MyContentTeam" --template content

# 2. เริ่มทำงาน
picoclaw team start

# 3. มอบหมายงาน
picoclaw team assign "เขียนบทความ AI" --to "jarvis"

# 4. ดูสถานะ
picoclaw team status

# 5. ดู dashboard
picoclaw team dashboard
```

---

## 📝 Notes & Considerations

### Security
- Agent isolation ต้องเข้มงวด (workspace + memory + tools)
- MCP tool permission ต้องตรวจสอบก่อนใช้
- Message encryption สำหรับ sensitive data

### Performance
- Goroutine pool สำหรับ parallel agent execution
- RAG indexing แยกตาม namespace
- Message queue ใช้ SQLite + indexing

### Scalability
- รองรับ 10-50 agents บน hardware เดียวกัน
- External agent registration (ผ่าน API)
- Distributed agent team (future)

---

## 📦 Asset Management (ในโปรเจค)

Assets สำหรับ Web UI อยู่ในโปรเจค (`ui/public/`) ไม่ใช่ใน `~/.picoclaw`

### Asset Directory Structure

```
ui/public/
├── sprites/                          # Sprites สำหรับ Virtual Office
│   ├── Modern_Office_Revamped_v1.2/  # Asset pack ที่คุณมี
│   │   ├── 4_Modern_Office_singles/  # 339 sprites แยกชิ้น
│   │   │   ├── 32x32/
│   │   │   │   ├── Modern_Office_Singles_1.png
│   │   │   │   ├── Modern_Office_Singles_2.png
│   │   │   │   └── ... (339 files)
│   │   │   └── ...
│   │   └── 6_Office_Designs/         # Room backgrounds
│   │
│   └── agents/                       # (optional) แยกตาม agent
│       ├── atlas/
│       ├── clawed/
│       └── ...
│
└── assets/                           # อื่นๆ (logo, icons)
```

### วิธีใช้งาน

**1. Assets อยู่ในโปรเจคแล้ว:**
```
ui/public/sprites/Modern_Office_Revamped_v1.2/
```

**2. ใช้งานใน UI Component:**
```typescript
// ใช้ตรงๆจาก public folder
const spritePath = "/sprites/Modern_Office_Revamped_v1.2/4_Modern_Office_singles/32x32/Modern_Office_Singles_10.png";
```

**3. Config ใน Agent (ที่ ~/.picoclaw/config.json):**
```json
{
  "ui_config": {
    "props": [
      {
        "sprite_path": "/sprites/Modern_Office_Revamped_v1.2/4_Modern_Office_singles/32x32/Modern_Office_Singles_10.png",
        "x": 0,
        "y": 0
      }
    ]
  }
}
```

### ข้อแตกต่างระหว่างที่เก็บข้อมูล

| ข้อมูล | ที่อยู่ | เหตุผล |
|--------|--------|--------|
| **Sprites/Assets** | `ui/public/sprites/` (ในโปรเจค) | เป็น static files สำหรับ web UI |
| **Config** | `~/.picoclaw/config.json` | User configuration |
| **Database** | `~/.picoclaw/workspace/` | Runtime data |

### Asset Naming Convention

```
ui/public/
├── sprites/
│   ├── agents/              # (optional) แยกตาม agent
│   │   ├── atlas/
│   │   │   └── desk.png → symlink หรือ copy จาก singles
│   │   └── ...
│   │
│   └── Modern_Office_Revamped_v1.2/
│       └── 4_Modern_Office_singles/32x32/
│           ├── Modern_Office_Singles_1.png    # Desks
│           ├── Modern_Office_Singles_10.png   # Desk variant
│           ├── Modern_Office_Singles_101.png  # Computers
│           └── ... (339 files)
```

**Sprite Categories (จาก singles):**
- Singles 1-50: Desks
- Singles 51-100: Chairs
- Singles 101-150: Computers/Monitors
- Singles 151-200: Office supplies
- Singles 201-250: Decorations
- Singles 251-300: Storage/Cabinets
- Singles 301-339: Misc

### Recommended Sources

| Source | URL | หา assets อะไร |
|--------|-----|---------------|
| itch.io | itch.io/game-assets | Pixel art, sprites |
| OpenGameArt | opengameart.org | LPC, free sprites |
| Craftpix | craftpix.net/freebies | 2D game assets |

### Fallback: Auto-Generated Avatars

ถ้าไม่ต้องการใช้ sprites ระบบจะใช้ fallback (ไม่ต้องมีไฟล์อะไรเพิ่ม):

```typescript
// ui/src/config/agentSprites.ts

export const AGENT_FALLBACKS = {
  atlas:    { emoji: "🕵️", initials: "AT", color: "#3B82F6" },
  jarvis:   { emoji: "🤖", initials: "JV", color: "#8B5CF6" },
  clawed:   { emoji: "👨‍💻", initials: "CL", color: "#F59E0B" },
  scribe:   { emoji: "✍️", initials: "SC", color: "#10B981" },
  sentinel: { emoji: "🛡️", initials: "SE", color: "#EF4444" },
  trendy:   { emoji: "📈", initials: "TR", color: "#EC4899" },
  pixel:    { emoji: "🎨", initials: "PX", color: "#6366F1" },
  nova:     { emoji: "🌌", initials: "NV", color: "#06B6D4" }
};

// สร้าง SVG avatar อัตโนมัติ (ไม่ต้องมี assets)
export function generateAvatarSVG(initials: string, color: string): string {
  return `<svg viewBox="0 0 100 100">...</svg>`; // Gradient circle + initials
}
```

### Config Format (ใน ~/.picoclaw/config.json)

```json
{
  "ui_config": {
    "avatar": {
      "type": "fallback",        // "custom" | "fallback"
      "fallback": {
        "type": "emoji",         // "emoji" | "generated"
        "emoji": "🕵️",
        "initials": "AT",
        "color": "#3B82F6"
      }
    },
    "props": [
      {
        "name": "desk",
        "sprite_path": "/sprites/Modern_Office_Revamped_v1.2/4_Modern_Office_singles/32x32/Modern_Office_Singles_10.png",
        "x": 0,
        "y": 0,
        "z_index": 1
      },
      {
        "name": "bookshelf",
        "sprite_path": "/sprites/Modern_Office_Revamped_v1.2/4_Modern_Office_singles/32x32/Modern_Office_Singles_251.png",
        "x": -40,
        "y": -20,
        "z_index": 0
      }
    ]
  }
}
```

---

## 🧪 Testing Strategy

### Unit Tests
```go
// pkg/agentcomm/mailbox_test.go
func TestMailbox_SendMessage(t *testing.T) {
    // Test message routing
    // Test priority handling
    // Test thread creation
}

// pkg/office/coordinator_test.go
func TestCoordinator_ProcessRequest(t *testing.T) {
    // Test plan generation
    // Test task distribution
    // Test error handling
}
```

### Integration Tests
```go
// tests/integration/agent_team_test.go
func TestAgentTeam_FullWorkflow(t *testing.T) {
    // 1. Create team with 3 agents
    // 2. Send request to coordinator
    // 3. Verify task distribution
    // 4. Verify inter-agent messages
    // 5. Verify kanban updates
}
```

### E2E Tests (UI)
```typescript
// ui/e2e/agent-config.spec.ts
test('user can create and configure new agent', async () => {
  // 1. Click "Add Agent"
  // 2. Fill form
  // 3. Select model
  // 4. Verify in virtual office
})
```

---

## 📈 Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| **UI Load Time** | <2s | Time to interactive |
| **Agent Response** | <3s | จาก click จนถึง modal เปิด |
| **Real-time Sync** | <500ms | WebSocket latency |
| **Canvas FPS** | 60fps | Virtual office animation |
| **Model Switch** | <1s | เปลี่ยน model config |
| **Memory Usage** | <100MB | Browser tab |

---

## 🚀 Deployment Checklist

### Pre-release
- [ ] All unit tests pass
- [ ] Integration tests pass
- [ ] UI E2E tests pass
- [ ] Documentation complete
- [ ] Asset packs hosted
- [ ] Migration scripts ready

### Release
- [ ] Version bump
- [ ] Changelog update
- [ ] Git tag
- [ ] Binary build
- [ ] Docker image
- [ ] UI bundle build

### Post-release
- [ ] Monitor error rates
- [ ] Collect user feedback
- [ ] Performance monitoring
- [ ] Update documentation

---

## 🔗 Related Documents

- `README.md` - Project overview
- `ROADMAP.md` - Project roadmap
- `pkg/office/` - Office simulation system
- `pkg/rag/` - RAG implementation
- `pkg/mcp/` - MCP integration
- `ui/` - Frontend UI code
- `assets/` - Visual assets

---

## 👥 Team Roles (Development)

| Role | Responsibility | Skills |
|------|---------------|--------|
| **Backend Lead** | Core agent system, API | Go, SQLite, Architecture |
| **Frontend Lead** | UI/UX, Virtual Office | React, TypeScript, PixiJS |
| **Pixel Artist** | Character sprites, backgrounds | Aseprite, Pixel Art |
| **DevOps** | Deployment, CI/CD | Docker, GitHub Actions |

---

**Last Updated:** 2026-03-09  
**Version:** 1.0  
**Status:** Planning Phase  
**Next Review:** 2026-03-16
