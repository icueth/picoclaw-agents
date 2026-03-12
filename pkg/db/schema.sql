-- Picoclaw Agent System Database Schema
-- Phase 1: Master Agent System

-- Enable foreign key support
PRAGMA foreign_keys = ON;

-- =============================================================================
-- RAG Documents Table (Vector Search Storage)
-- =============================================================================
CREATE TABLE IF NOT EXISTS rag_documents (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    embedding BLOB NOT NULL,
    metadata TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for rag_documents
CREATE INDEX IF NOT EXISTS idx_rag_documents_created_at ON rag_documents(created_at);

-- =============================================================================
-- FTS5 Virtual Table for Keyword Search (Hybrid RAG)
-- =============================================================================
CREATE VIRTUAL TABLE IF NOT EXISTS rag_documents_fts USING fts5(
    content,
    metadata,
    content_rowid=rowid,
    content='rag_documents'
);

-- Triggers to sync rag_documents with FTS5 index
CREATE TRIGGER IF NOT EXISTS trg_rag_documents_insert_fts
AFTER INSERT ON rag_documents
BEGIN
    INSERT INTO rag_documents_fts(rowid, content, metadata)
    VALUES (NEW.rowid, NEW.content, NEW.metadata);
END;

CREATE TRIGGER IF NOT EXISTS trg_rag_documents_delete_fts
AFTER DELETE ON rag_documents
BEGIN
    INSERT INTO rag_documents_fts(rag_documents_fts, rowid, content, metadata)
    VALUES ('delete', OLD.rowid, OLD.content, OLD.metadata);
END;

CREATE TRIGGER IF NOT EXISTS trg_rag_documents_update_fts
AFTER UPDATE ON rag_documents
BEGIN
    INSERT INTO rag_documents_fts(rag_documents_fts, rowid, content, metadata)
    VALUES ('delete', OLD.rowid, OLD.content, OLD.metadata);
    INSERT INTO rag_documents_fts(rowid, content, metadata)
    VALUES (NEW.rowid, NEW.content, NEW.metadata);
END;

-- =============================================================================
-- Concepts Table (Work Continuity)
-- =============================================================================
CREATE TABLE IF NOT EXISTS concepts (
    id TEXT PRIMARY KEY,
    concept TEXT NOT NULL,
    context TEXT NOT NULL,
    related_jobs TEXT NOT NULL DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_concepts_status CHECK (status IN ('active', 'completed', 'paused'))
);

-- Indexes for concepts
CREATE INDEX IF NOT EXISTS idx_concepts_status ON concepts(status);
CREATE INDEX IF NOT EXISTS idx_concepts_created_at ON concepts(created_at);
CREATE INDEX IF NOT EXISTS idx_concepts_updated_at ON concepts(updated_at);

-- =============================================================================
-- Jobs Table (Task Persistence)
-- =============================================================================
CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    concept_id TEXT,
    role TEXT NOT NULL,
    task TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    context TEXT NOT NULL DEFAULT '{}',
    result TEXT,
    parent_job_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_jobs_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    CONSTRAINT fk_jobs_concept_id FOREIGN KEY (concept_id) REFERENCES concepts(id) ON DELETE SET NULL,
    CONSTRAINT fk_jobs_parent_job_id FOREIGN KEY (parent_job_id) REFERENCES jobs(id) ON DELETE SET NULL
);

-- Indexes for jobs
CREATE INDEX IF NOT EXISTS idx_jobs_concept_id ON jobs(concept_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_role ON jobs(role);
CREATE INDEX IF NOT EXISTS idx_jobs_parent_job_id ON jobs(parent_job_id);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at);
CREATE INDEX IF NOT EXISTS idx_jobs_updated_at ON jobs(updated_at);

-- =============================================================================
-- Projects Table (Multi-Role Workflow Tracking)
-- =============================================================================
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    concept_id TEXT,
    current_phase TEXT NOT NULL DEFAULT 'planning',
    phases TEXT NOT NULL DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'active',
    metadata TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_projects_current_phase CHECK (current_phase IN ('planning', 'research', 'coding', 'review')),
    CONSTRAINT chk_projects_status CHECK (status IN ('active', 'completed', 'paused', 'cancelled')),
    CONSTRAINT fk_projects_concept_id FOREIGN KEY (concept_id) REFERENCES concepts(id) ON DELETE SET NULL
);

-- Indexes for projects
CREATE INDEX IF NOT EXISTS idx_projects_concept_id ON projects(concept_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_current_phase ON projects(current_phase);
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at);
CREATE INDEX IF NOT EXISTS idx_projects_updated_at ON projects(updated_at);

-- =============================================================================
-- Triggers for Automatic updated_at Management
-- =============================================================================

-- Trigger to update concepts.updated_at
CREATE TRIGGER IF NOT EXISTS trg_concepts_updated_at
AFTER UPDATE ON concepts
FOR EACH ROW
BEGIN
    UPDATE concepts SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger to update jobs.updated_at
CREATE TRIGGER IF NOT EXISTS trg_jobs_updated_at
AFTER UPDATE ON jobs
FOR EACH ROW
BEGIN
    UPDATE jobs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger to update projects.updated_at
CREATE TRIGGER IF NOT EXISTS trg_projects_updated_at
AFTER UPDATE ON projects
FOR EACH ROW
BEGIN
    UPDATE projects SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
