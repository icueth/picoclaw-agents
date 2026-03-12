# PicoClaw Changes & Enhancements

เอกสารนี้สรุปการเปลี่ยนแปลงและการปรับปรุงที่ทำในระบบ PicoClaw เมื่อเทียบกับ [sipeed/picoclaw](https://github.com/sipeed/picoclaw) ต้นฉบับ

**Version:** Enhanced Edition  
**Last Updated:** March 2026

---

## Table of Contents

1. [Overview](#overview)
2. [New Systems Added](#new-systems-added)
3. [Detailed Feature Comparison](#detailed-feature-comparison)
4. [Configuration Changes](#configuration-changes)
5. [New Tools Added](#new-tools-added)
6. [Database Schema](#database-schema)
7. [File Structure Changes](#file-structure-changes)
8. [Summary](#summary)

---

## Overview

ระบบ PicoClaw ที่ปรับปรุงนี้ได้เพิ่มฟีเจอร์ enterprise-grade หลายอย่างเข้าไป รวมถึง:

- **Multi-Agent Orchestration** - ระบบ subagent ที่มี role-based spawning
- **Persistent Memory** - ระบบจดจำแบบ persistent ผ่าน SQLite
- **RAG System** - Retrieval-Augmented Generation สำหรับ knowledge retrieval
- **Job Monitoring** - ระบบติดตามและจัดการงานแบบ adaptive timeout
- **Local Embedding** - ระบบสร้าง embedding แบบ Go-native (ไม่ต้องใช้ Python) สำหรับ semantic search เบื้องต้น

---

## New Systems Added

### 1. Subagent System (Multi-Agent Orchestration)

| Component | Original | Enhanced |
|-----------|----------|----------|
| **Subagent Spawning** | ❌ ไม่มี | ✅ Role-based spawning |
| **Async Execution** | ❌ ไม่มี | ✅ Async + Sync modes |
| **Progress Tracking** | ❌ ไม่มี | ✅ Progress percent + messages |
| **Adaptive Timeout** | ❌ ไม่มี | ✅ Auto-extend timeout |
| **Job Persistence** | ❌ ไม่มี | ✅ SQLite persistence |
| **Inter-Agent Messaging** | ❌ ไม่มี | ✅ P2P + Pub/Sub messaging |

**ไฟล์ที่เพิ่ม:**
- `pkg/tools/subagent.go` - Core subagent management
- `pkg/tools/subagent_roles.go` - Role configuration
- `pkg/tools/subagent_status.go` - Status checking
- `pkg/tools/subagent_cancel.go` - Task cancellation
- `pkg/agent/job_monitor.go` - Job monitoring service
- `pkg/agent/hierarchy.go` - Agent hierarchy management
- `pkg/agent/messenger.go` - Inter-agent messaging

---

### 2. RAG System (Retrieval-Augmented Generation)

| Feature | Original | Enhanced |
|---------|----------|----------|
| **Document Storage** | ❌ ไม่มี | ✅ Vector storage |
| **Semantic Search** | ❌ ไม่มี | ✅ Cosine similarity |
| **Text Chunking** | ❌ ไม่มี | ✅ Configurable chunking |
| **Multiple Embeddings** | ❌ ไม่มี | ✅ Local/OpenAI/HTTP/GGUF |
| **Fallback Support** | ❌ ไม่มี | ✅ Auto-fallback |

**ไฟล์ที่เพิ่ม:**
- `pkg/rag/manager.go` - Core RAG manager
- `pkg/rag/vectorstore.go` - Vector storage
- `pkg/rag/embeddings.go` - Embedding generators
- `pkg/rag/chunker.go` - Text chunking
- `pkg/rag/similarity.go` - Similarity calculations
- `pkg/tools/rag_tools.go` - RAG tools (query_rag, save_to_rag)

---

### 3. Memory System

| Feature | Original | Enhanced |
|---------|----------|----------|
| **Concept Tracking** | ❌ ไม่มี | ✅ Work continuity concepts |
| **Job Tracking** | ❌ ไม่มี | ✅ Persistent job management |
| **Session Memory** | ✅ มี | ✅ Enhanced with importance scoring |
| **Cross-Session Recall** | ❌ ไม่มี | ✅ Semantic search |

**ไฟล์ที่เพิ่ม:**
- `pkg/memory/concept.go` - Concept manager
- `pkg/memory/job.go` - Job manager
- `pkg/agent/memory_v2.go` - Enhanced memory system
- `pkg/agent/memory_llm.go` - LLM-based memory processing
- `pkg/agent/memory_tasks.go` - Memory task management
- `pkg/agent/auto_memory.go` - Automatic memory management
- `pkg/tools/memory_tools.go` - Memory tools

---

### 4. Embedding Support

| Feature | Original | Enhanced |
|---------|----------|----------|
| **Go-Native Local** | ❌ ไม่มี | ✅ Built-in hash-based embedding |
| **OpenAI Support** | ❌ ไม่มี | ✅ Integrated OpenAI embeddings |
| **No-op Mode** | ❌ ไม่มี | ✅ Keyword-only search fallback |
| **HTTP Fallback** | ❌ ไม่มี | ✅ Supports any OpenAI-compatible API |

**ไฟล์ที่เปลี่ยนแปลง:**
- `pkg/rag/embeddings.go` - Multimodal embedding providers
- `pkg/rag/vectorstore.go` - SQLite vector storage implementation

---

### 5. Job Monitoring System

| Feature | Original | Enhanced |
|---------|----------|----------|
| **Health Monitoring** | ❌ ไม่มี | ✅ Health status tracking |
| **Stuck Detection** | ❌ ไม่มี | ✅ Auto-detect stuck jobs |
| **Auto-Extension** | ❌ ไม่มี | ✅ Extend timeout automatically |
| **Progress Reporting** | ❌ ไม่มี | ✅ Percent + ETA |
| **Intelligent Retry** | ❌ ไม่มี | ✅ Exponential backoff |

**ไฟล์ที่เพิ่ม:**
- `pkg/agent/job_monitor.go` - Job monitoring service
- `pkg/tools/job_health.go` - Job health tool
- `pkg/tools/report_progress.go` - Progress reporting tool
- `pkg/tools/job_management.go` - Job management tools

---

### 6. Project Management

| Feature | Original | Enhanced |
|---------|----------|----------|
| **Multi-Role Projects** | ❌ ไม่มี | ✅ Phase-based workflows |
| **Role Assignment** | ❌ ไม่มี | ✅ Planner/Coder/Reviewer |
| **Project Phases** | ❌ ไม่มี | ✅ Planning/Research/Coding/Review |
| **Progress Tracking** | ❌ ไม่มี | ✅ Phase completion tracking |

**ไฟล์ที่เพิ่ม:**
- `pkg/project/manager.go` - Project manager
- `pkg/tools/project_tools.go` - Project tools

---

## Detailed Feature Comparison

### Agent Capabilities

| Capability | Original | Enhanced |
|------------|----------|----------|
| Basic Agent Loop | ✅ | ✅ |
| Tool Execution | ✅ | ✅ Enhanced |
| Memory (Session) | ✅ | ✅ Enhanced |
| Subagent Spawning | ❌ | ✅ |
| Role-Based Agents | ❌ | ✅ |
| Agent Hierarchy | ❌ | ✅ |
| Inter-Agent Messaging | ❌ | ✅ |
| Agent Discovery | ❌ | ✅ |
| Capability Matching | ❌ | ✅ |

### Tool System

| Tool Category | Original | Enhanced |
|---------------|----------|----------|
| File Operations | ✅ | ✅ |
| Shell Execution | ✅ | ✅ Enhanced with safety |
| Web Search | ✅ | ✅ |
| Memory Tools | ❌ | ✅ create_concept, list_concepts |
| RAG Tools | ❌ | ✅ query_rag, save_to_rag |
| Subagent Tools | ❌ | ✅ spawn_subagent, subagent_status |
| Job Tools | ❌ | ✅ job_health, list_jobs |
| Project Tools | ❌ | ✅ create_project, assign_phase |

### Configuration System

```go
// Original Configuration
{
  "agents": { "defaults": {...} },
  "model_list": [...],
  "channels": {...},
  "tools": {...}
}

// Enhanced Configuration
{
  "agents": { "defaults": {...}, "roles": {...} },
  "model_list": [...],
  "channels": {...},
  "tools": {...},
  "subagent_roles": {         // NEW
    "planner": { "model": "...", "timeout_seconds": 300, "extendable": true },
    "coder": { "model": "...", "timeout_seconds": 600, "max_extensions": 3 },
    "reviewer": {...}
  },
  "memory": {                 // NEW
    "type": "sqlite",
    "rag": { "enabled": true, "embedding_model": "http" }
  },
  "jobs": {                   // NEW
    "persistence": "sqlite",
    "default_timeout": 300
  },
  "rag": {                    // NEW
    "enabled": true,
    "embedding_model": "http",
    "api_base": "http://localhost:18190"
  }
}
```

---

## Configuration Changes

### New Configuration Sections

#### 1. Subagent Roles (`subagent_roles`)

```json
{
  "subagent_roles": {
    "planner": {
      "model": "gpt-4",
      "description": "Creates detailed plans and architecture",
      "system_prompt_addon": "You are a planning expert...",
      "max_iterations": 30,
      "timeout_seconds": 300,
      "extendable": true,
      "max_extensions": 3,
      "allowed_tools": ["read_file", "write_file", "web_search"]
    },
    "coder": {
      "model": "claude-sonnet",
      "timeout_seconds": 600,
      "extendable": true,
      "max_extensions": 5
    },
    "reviewer": {
      "model": "gpt-4",
      "timeout_seconds": 180,
      "extendable": false
    }
  }
}
```

#### 2. RAG Configuration (`rag`)

```json
{
  "rag": {
    "enabled": true,
    "embedding_model": "http",
    "chunk_size": 1000,
    "overlap": 200,
    "max_results": 5,
    "similarity_threshold": 0.7,
    "api_base": "http://localhost:18190",
    "model_path": "./models"
  }
}
```

#### 3. Memory Configuration (`memory`)

```json
{
  "memory": {
    "type": "sqlite",
    "database_path": "~/.picoclaw/workspace/memory.db",
    "rag": {
      "enabled": true,
      "embedding_model": "http"
    },
    "concept_retention_days": 90,
    "max_context_memories": 10,
    "auto_save": true
  }
}
```

#### 4. Job Configuration (`jobs`)

```json
{
  "jobs": {
    "persistence": "sqlite",
    "database_path": "~/.picoclaw/workspace/jobs.db",
    "default_timeout": 300,
    "max_concurrent": 10,
    "cleanup_interval_hours": 24,
    "retention_days": 30
  }
}
```

---

## New Tools Added

### Subagent Management Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `spawn_subagent` | สร้าง subagent ใหม่ | `role`, `task`, `concept_id`, `label`, `async` |
| `subagent_status` | ตรวจสอบสถานะ subagent | `action` (get/list), `task_id` |
| `cancel_subagent` | ยกเลิก subagent | `task_id` |
| `list_subagent_roles` | แสดงรายการ roles | - |

### Job Management Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `job_health` | ตรวจสอบสุขภาพงาน | `action` (check/list/extend/kill), `task_id` |
| `list_jobs` | แสดงรายการงาน | `status`, `limit` |
| `report_progress` | รายงานความคืบหน้า (subagent) | `percent`, `message`, `eta`, `request_extension` |

### RAG Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `query_rag` | ค้นหาใน RAG | `query`, `top_k`, `threshold`, `filter` |
| `save_to_rag` | บันทึกลง RAG | `content`, `metadata`, `source` |
| `get_rag_context` | ดึง context จาก RAG | `query`, `max_tokens` |

### Memory Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `create_concept` | สร้าง concept ใหม่ | `title`, `description`, `tags` |
| `list_concepts` | แสดงรายการ concepts | `status`, `tag` |
| `continue_concept` | ทำงานต่อบน concept | `concept_id` |

### Project Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `create_project` | สร้างโปรเจคใหม่ | `name`, `description`, `phases` |
| `assign_phase` | มอบหมาย phase | `project_id`, `phase`, `role` |
| `get_project_status` | ดูสถานะโปรเจค | `project_id` |

---

## Database Schema

### New Tables

#### 1. rag_documents
```sql
CREATE TABLE rag_documents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    embedding BLOB,
    metadata JSON,
    source TEXT,
    chunk_index INTEGER,
    total_chunks INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2. concepts
```sql
CREATE TABLE concepts (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'active',
    tags JSON,
    context JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 3. jobs
```sql
CREATE TABLE jobs (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    data JSON,
    result JSON,
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);
```

#### 4. projects
```sql
CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'active',
    phases JSON,
    current_phase INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## File Structure Changes

### Directory Tree (New Additions)

```
picoclaw/
├── pkg/
│   ├── agent/
│   │   ├── job_monitor.go          # NEW: Job monitoring
│   │   ├── hierarchy.go            # NEW: Agent hierarchy
│   │   ├── messenger.go            # NEW: Inter-agent messaging
│   │   ├── discovery.go            # NEW: Agent discovery
│   │   ├── memory_v2.go            # NEW: Enhanced memory
│   │   ├── memory_llm.go           # NEW: LLM memory processing
│   │   ├── memory_tasks.go         # NEW: Memory tasks
│   │   ├── auto_memory.go          # NEW: Auto memory
│   │   ├── shared_context.go       # NEW: Shared context
│   │   └── agent_message.go        # NEW: Message types
│   ├── agentcomm/                  # NEW: Agent communication
│   ├── db/
│   │   ├── manager.go              # NEW: SQLite manager
│   │   └── schema.sql              # NEW: Database schema
│   ├── memory/
│   │   ├── concept.go              # NEW: Concept manager
│   │   └── job.go                  # NEW: Job manager
│   ├── project/
│   │   └── manager.go              # NEW: Project manager
│   ├── rag/
│   │   ├── manager.go              # NEW: RAG manager
│   │   ├── vectorstore.go          # NEW: Vector storage
│   │   ├── embeddings.go           # NEW: Embeddings (Go-native)
│   │   ├── chunker.go              # NEW: Text chunking
│   │   └── similarity.go           # NEW: Similarity
│   └── tools/
│       ├── subagent.go             # NEW: Subagent tool
│       ├── subagent_roles.go       # NEW: Role management
│       ├── subagent_status.go      # NEW: Status tool
│       ├── subagent_cancel.go      # NEW: Cancel tool
│       ├── job_health.go           # NEW: Job health tool
│       ├── job_management.go       # NEW: Job management
│       ├── report_progress.go      # NEW: Progress reporting
│       ├── rag_tools.go            # NEW: RAG tools
│       ├── memory_tools.go         # NEW: Memory tools
│       ├── project_tools.go        # NEW: Project tools
│       ├── discovery_tools.go      # NEW: Discovery tools
│       └── agent_delegate.go       # NEW: Delegation
```

---

## Summary

### สรุปการเปลี่ยนแปลงหลัก

| หมวดหมู่ | จำนวน Features ใหม่ | ความสำคัญ |
|----------|---------------------|-----------|
| **Subagent System** | 6+ components | สูง - ระบบ multi-agent |
| **RAG System** | 5+ components | สูง - knowledge retrieval |
| **Memory System** | 6+ components | สูง - work continuity |
| **Job Monitoring** | 4+ components | สูง - self-healing |
| **Embedding Service** | 4+ components | ปานกลาง - vector search |
| **Project Management** | 2+ components | ปานกลาง - workflow |

### จำนวนไฟล์ที่เพิ่ม

- **Go Files:** ~35 ไฟล์
- **Total:** ~35 ไฟล์ใหม่

### ความสามารถที่เพิ่มขึ้น

1. **Multi-Agent Orchestration** - สามารถสร้างและจัดการ subagents ได้
2. **Persistent Memory** - จดจำงานและ context ข้าม session
3. **Semantic Search** - ค้นหาแบบ semantic ผ่าน RAG
4. **Self-Healing** - ระบบตรวจสอบและซ่อมแซมตัวเอง
5. **Adaptive Timeouts** - ปรับ timeout ตามความเหมาะสม
6. **Project Workflows** - จัดการโปรเจคแบบ multi-phase

---

## Migration Notes

สำหรับผู้ที่ต้องการ migrate จาก picoclaw ต้นฉบับ:

1. **Config Migration:** เพิ่ม sections ใหม่ (`subagent_roles`, `rag`, `memory`, `jobs`)
2. **Database Setup:** รัน `picoclaw onboard` เพื่อสร้าง database
3. **Backward Compatibility:** Config เดิมยังทำงานได้ (fallback to defaults)

---

*End of Document*
