// Package db provides SQLite database management for the picoclaw agent system.
package db

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// DB wraps the SQLite database connection with helper methods.
type DB struct {
	conn *sql.DB
	path string
}

// New creates a new DB instance with the given database file path.
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return &DB{
		conn: conn,
		path: dbPath,
	}, nil
}

// Init creates all tables and indexes from the embedded schema.
func (d *DB) Init() error {
	if _, err := d.conn.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}
	return nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

// Conn returns the underlying sql.DB connection for advanced queries.
func (d *DB) Conn() *sql.DB {
	return d.conn
}

// =============================================================================
// RAG Document Operations
// =============================================================================

// RAGDocument represents a document stored for vector search.
type RAGDocument struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Embedding []byte    `json:"embedding"`
	Metadata  Metadata  `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

// Metadata stores additional information about a RAG document.
type Metadata struct {
	Agent     string            `json:"agent,omitempty"`
	Role      string            `json:"role,omitempty"`
	Project   string            `json:"project,omitempty"`
	Source    string            `json:"source,omitempty"`
	Timestamp time.Time         `json:"timestamp,omitempty"`
	Extra     map[string]string `json:"extra,omitempty"`
}

// CreateRAGDocument inserts a new RAG document into the database.
func (d *DB) CreateRAGDocument(content string, embedding []byte, metadata Metadata) (*RAGDocument, error) {
	id := uuid.New().String()
	if metadata.Timestamp.IsZero() {
		metadata.Timestamp = time.Now()
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = d.conn.Exec(
		`INSERT INTO rag_documents (id, content, embedding, metadata) VALUES (?, ?, ?, ?)`,
		id, content, embedding, string(metadataJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create RAG document: %w", err)
	}

	return d.GetRAGDocument(id)
}

// GetRAGDocument retrieves a RAG document by ID.
func (d *DB) GetRAGDocument(id string) (*RAGDocument, error) {
	var doc RAGDocument
	var metadataJSON string

	err := d.conn.QueryRow(
		`SELECT id, content, embedding, metadata, created_at FROM rag_documents WHERE id = ?`,
		id,
	).Scan(&doc.ID, &doc.Content, &doc.Embedding, &metadataJSON, &doc.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("RAG document not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get RAG document: %w", err)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &doc.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &doc, nil
}

// ListRAGDocuments returns all RAG documents, optionally filtered by metadata.
// ListRAGEmbeddings returns only IDs and embeddings for memory-efficient similarity search.
func (d *DB) ListRAGEmbeddings(limit int) ([]*RAGDocument, error) {
	// Only fetch id and embedding to save RAM
	query := `SELECT id, embedding FROM rag_documents ORDER BY created_at DESC`
	if limit > 0 {
		query = fmt.Sprintf("%s LIMIT %d", query, limit)
	}

	rows, err := d.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list RAG embeddings: %w", err)
	}
	defer rows.Close()

	var docs []*RAGDocument
	for rows.Next() {
		var doc RAGDocument
		if err := rows.Scan(&doc.ID, &doc.Embedding); err != nil {
			continue
		}
		docs = append(docs, &doc)
	}
	return docs, nil
}

// ListRAGDocuments returns all RAG documents, optionally filtered by metadata.
func (d *DB) ListRAGDocuments(limit int) ([]*RAGDocument, error) {
	query := `SELECT id, content, embedding, metadata, created_at FROM rag_documents ORDER BY created_at DESC`
	if limit > 0 {
		query = fmt.Sprintf("%s LIMIT %d", query, limit)
	}

	rows, err := d.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list RAG documents: %w", err)
	}
	defer rows.Close()

	return scanRAGDocuments(rows)
}
// DeleteRAGDocument removes a RAG document by ID.
func (d *DB) DeleteRAGDocument(id string) error {
	_, err := d.conn.Exec(`DELETE FROM rag_documents WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete RAG document: %w", err)
	}
	return nil
}

func scanRAGDocuments(rows *sql.Rows) ([]*RAGDocument, error) {
	var docs []*RAGDocument
	for rows.Next() {
		var doc RAGDocument
		var metadataJSON string

		if err := rows.Scan(&doc.ID, &doc.Content, &doc.Embedding, &metadataJSON, &doc.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan RAG document: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataJSON), &doc.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		docs = append(docs, &doc)
	}
	return docs, rows.Err()
}

// =============================================================================
// Concept Operations
// =============================================================================

// Concept represents a high-level work concept for continuity.
type Concept struct {
	ID          string    `json:"id"`
	Concept     string    `json:"concept"`
	Context     string    `json:"context"`
	RelatedJobs []string  `json:"related_jobs"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateConcept inserts a new concept into the database.
func (d *DB) CreateConcept(concept, context string) (*Concept, error) {
	id := uuid.New().String()
	relatedJobsJSON, _ := json.Marshal([]string{})

	_, err := d.conn.Exec(
		`INSERT INTO concepts (id, concept, context, related_jobs, status) VALUES (?, ?, ?, ?, ?)`,
		id, concept, context, string(relatedJobsJSON), "active",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create concept: %w", err)
	}

	return d.GetConcept(id)
}

// GetConcept retrieves a concept by ID.
func (d *DB) GetConcept(id string) (*Concept, error) {
	var c Concept
	var relatedJobsJSON string

	err := d.conn.QueryRow(
		`SELECT id, concept, context, related_jobs, status, created_at, updated_at FROM concepts WHERE id = ?`,
		id,
	).Scan(&c.ID, &c.Concept, &c.Context, &relatedJobsJSON, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("concept not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get concept: %w", err)
	}

	if err := json.Unmarshal([]byte(relatedJobsJSON), &c.RelatedJobs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal related_jobs: %w", err)
	}

	return &c, nil
}

// ListConcepts returns all concepts, optionally filtered by status.
func (d *DB) ListConcepts(status string) ([]*Concept, error) {
	query := `SELECT id, concept, context, related_jobs, status, created_at, updated_at FROM concepts`
	var args []interface{}

	if status != "" {
		query += ` WHERE status = ?`
		args = append(args, status)
	}

	query += ` ORDER BY updated_at DESC`

	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list concepts: %w", err)
	}
	defer rows.Close()

	return scanConcepts(rows)
}

// UpdateConceptStatus updates the status of a concept.
func (d *DB) UpdateConceptStatus(id, status string) error {
	if status != "active" && status != "completed" && status != "paused" {
		return fmt.Errorf("invalid status: %s", status)
	}

	_, err := d.conn.Exec(`UPDATE concepts SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("failed to update concept status: %w", err)
	}
	return nil
}

// UpdateConceptContext updates the context of a concept.
func (d *DB) UpdateConceptContext(id, context string) error {
	_, err := d.conn.Exec(`UPDATE concepts SET context = ? WHERE id = ?`, context, id)
	if err != nil {
		return fmt.Errorf("failed to update concept context: %w", err)
	}
	return nil
}

// AddRelatedJob adds a job ID to a concept's related_jobs list.
func (d *DB) AddRelatedJob(conceptID, jobID string) error {
	c, err := d.GetConcept(conceptID)
	if err != nil {
		return err
	}

	for _, existingID := range c.RelatedJobs {
		if existingID == jobID {
			return nil // Already exists
		}
	}

	c.RelatedJobs = append(c.RelatedJobs, jobID)
	relatedJobsJSON, _ := json.Marshal(c.RelatedJobs)

	_, err = d.conn.Exec(`UPDATE concepts SET related_jobs = ? WHERE id = ?`, string(relatedJobsJSON), conceptID)
	if err != nil {
		return fmt.Errorf("failed to add related job: %w", err)
	}
	return nil
}

// DeleteConcept removes a concept by ID.
func (d *DB) DeleteConcept(id string) error {
	_, err := d.conn.Exec(`DELETE FROM concepts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete concept: %w", err)
	}
	return nil
}

func scanConcepts(rows *sql.Rows) ([]*Concept, error) {
	var concepts []*Concept
	for rows.Next() {
		var c Concept
		var relatedJobsJSON string

		if err := rows.Scan(&c.ID, &c.Concept, &c.Context, &relatedJobsJSON, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan concept: %w", err)
		}

		if err := json.Unmarshal([]byte(relatedJobsJSON), &c.RelatedJobs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal related_jobs: %w", err)
		}

		concepts = append(concepts, &c)
	}
	return concepts, rows.Err()
}

// =============================================================================
// Job Operations
// =============================================================================

// Job represents a task/job in the system.
type Job struct {
	ID          string                 `json:"id"`
	ConceptID   *string                `json:"concept_id,omitempty"`
	Role        string                 `json:"role"`
	Task        string                 `json:"task"`
	Status      string                 `json:"status"`
	Context     map[string]interface{} `json:"context"`
	Result      *string                `json:"result,omitempty"`
	ParentJobID *string                `json:"parent_job_id,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CreateJob inserts a new job into the database.
func (d *DB) CreateJob(conceptID *string, role, task string, context map[string]interface{}, parentJobID *string) (*Job, error) {
	id := uuid.New().String()

	contextJSON, _ := json.Marshal(context)
	if context == nil {
		contextJSON = []byte("{}")
	}

	_, err := d.conn.Exec(
		`INSERT INTO jobs (id, concept_id, role, task, status, context, parent_job_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, conceptID, role, task, "pending", string(contextJSON), parentJobID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return d.GetJob(id)
}

// GetJob retrieves a job by ID.
func (d *DB) GetJob(id string) (*Job, error) {
	var j Job
	var contextJSON string
	var result sql.NullString
	var conceptID sql.NullString
	var parentJobID sql.NullString

	err := d.conn.QueryRow(
		`SELECT id, concept_id, role, task, status, context, result, parent_job_id, created_at, updated_at FROM jobs WHERE id = ?`,
		id,
	).Scan(&j.ID, &conceptID, &j.Role, &j.Task, &j.Status, &contextJSON, &result, &parentJobID, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	if conceptID.Valid {
		j.ConceptID = &conceptID.String
	}
	if result.Valid {
		j.Result = &result.String
	}
	if parentJobID.Valid {
		j.ParentJobID = &parentJobID.String
	}

	if err := json.Unmarshal([]byte(contextJSON), &j.Context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	return &j, nil
}

// ListJobs returns jobs, optionally filtered by concept_id, status, or role.
func (d *DB) ListJobs(conceptID, status, role string) ([]*Job, error) {
	query := `SELECT id, concept_id, role, task, status, context, result, parent_job_id, created_at, updated_at FROM jobs WHERE 1=1`
	var args []interface{}

	if conceptID != "" {
		query += ` AND concept_id = ?`
		args = append(args, conceptID)
	}
	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	if role != "" {
		query += ` AND role = ?`
		args = append(args, role)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()

	return scanJobs(rows)
}

// UpdateJobStatus updates the status of a job.
func (d *DB) UpdateJobStatus(id, status string) error {
	if status != "pending" && status != "running" && status != "completed" && status != "failed" && status != "cancelled" {
		return fmt.Errorf("invalid status: %s", status)
	}

	_, err := d.conn.Exec(`UPDATE jobs SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}
	return nil
}

// UpdateJobResult updates the result of a job.
func (d *DB) UpdateJobResult(id, result string) error {
	_, err := d.conn.Exec(`UPDATE jobs SET result = ? WHERE id = ?`, result, id)
	if err != nil {
		return fmt.Errorf("failed to update job result: %w", err)
	}
	return nil
}

// UpdateJobContext updates the context of a job.
func (d *DB) UpdateJobContext(id string, context map[string]interface{}) error {
	contextJSON, _ := json.Marshal(context)
	_, err := d.conn.Exec(`UPDATE jobs SET context = ? WHERE id = ?`, string(contextJSON), id)
	if err != nil {
		return fmt.Errorf("failed to update job context: %w", err)
	}
	return nil
}

// GetChildJobs retrieves all child jobs for a given parent job ID.
func (d *DB) GetChildJobs(parentJobID string) ([]*Job, error) {
	rows, err := d.conn.Query(
		`SELECT id, concept_id, role, task, status, context, result, parent_job_id, created_at, updated_at FROM jobs WHERE parent_job_id = ? ORDER BY created_at ASC`,
		parentJobID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get child jobs: %w", err)
	}
	defer rows.Close()

	return scanJobs(rows)
}

// DeleteJob removes a job by ID.
func (d *DB) DeleteJob(id string) error {
	_, err := d.conn.Exec(`DELETE FROM jobs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}
	return nil
}

func scanJobs(rows *sql.Rows) ([]*Job, error) {
	var jobs []*Job
	for rows.Next() {
		var j Job
		var contextJSON string
		var result sql.NullString
		var conceptID sql.NullString
		var parentJobID sql.NullString

		if err := rows.Scan(&j.ID, &conceptID, &j.Role, &j.Task, &j.Status, &contextJSON, &result, &parentJobID, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if conceptID.Valid {
			j.ConceptID = &conceptID.String
		}
		if result.Valid {
			j.Result = &result.String
		}
		if parentJobID.Valid {
			j.ParentJobID = &parentJobID.String
		}

		if err := json.Unmarshal([]byte(contextJSON), &j.Context); err != nil {
			return nil, fmt.Errorf("failed to unmarshal context: %w", err)
		}

		jobs = append(jobs, &j)
	}
	return jobs, rows.Err()
}

// =============================================================================
// Project Operations
// =============================================================================

// Phase represents a phase in a project workflow.
type Phase struct {
	Role   string `json:"role"`
	Status string `json:"status"`
	JobID  string `json:"job_id"`
}

// Project represents a multi-role workflow project.
type Project struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ConceptID    *string   `json:"concept_id,omitempty"`
	CurrentPhase string    `json:"current_phase"`
	Phases       []Phase   `json:"phases"`
	Status       string    `json:"status"`
	Metadata     string    `json:"metadata"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateProject inserts a new project into the database.
func (d *DB) CreateProject(name string, conceptID *string) (*Project, error) {
	id := uuid.New().String()
	phasesJSON, _ := json.Marshal([]Phase{})
	metadataJSON, _ := json.Marshal(map[string]interface{}{})

	_, err := d.conn.Exec(
		`INSERT INTO projects (id, name, description, concept_id, current_phase, phases, status, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, name, "", conceptID, "planning", string(phasesJSON), "active", string(metadataJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return d.GetProject(id)
}

// GetProject retrieves a project by ID.
func (d *DB) GetProject(id string) (*Project, error) {
	var p Project
	var phasesJSON string
	var metadataJSON string
	var conceptID sql.NullString

	err := d.conn.QueryRow(
		`SELECT id, name, description, concept_id, current_phase, phases, status, metadata, created_at, updated_at FROM projects WHERE id = ?`,
		id,
	).Scan(&p.ID, &p.Name, &p.Description, &conceptID, &p.CurrentPhase, &phasesJSON, &p.Status, &metadataJSON, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if conceptID.Valid {
		p.ConceptID = &conceptID.String
	}

	if err := json.Unmarshal([]byte(phasesJSON), &p.Phases); err != nil {
		return nil, fmt.Errorf("failed to unmarshal phases: %w", err)
	}

	p.Metadata = metadataJSON

	return &p, nil
}

// ListProjects returns all projects, optionally filtered by status.
func (d *DB) ListProjects(status string) ([]*Project, error) {
	query := `SELECT id, name, description, concept_id, current_phase, phases, status, metadata, created_at, updated_at FROM projects`
	var args []interface{}

	if status != "" {
		query += ` WHERE status = ?`
		args = append(args, status)
	}

	query += ` ORDER BY updated_at DESC`

	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	return scanProjects(rows)
}

// UpdateProjectStatus updates the status of a project.
func (d *DB) UpdateProjectStatus(id, status string) error {
	if status != "active" && status != "completed" && status != "paused" && status != "cancelled" {
		return fmt.Errorf("invalid status: %s", status)
	}

	_, err := d.conn.Exec(`UPDATE projects SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}
	return nil
}

// UpdateProjectPhase updates the current phase of a project.
func (d *DB) UpdateProjectPhase(id, phase string) error {
	if phase != "planning" && phase != "research" && phase != "coding" && phase != "review" {
		return fmt.Errorf("invalid phase: %s", phase)
	}

	_, err := d.conn.Exec(`UPDATE projects SET current_phase = ? WHERE id = ?`, phase, id)
	if err != nil {
		return fmt.Errorf("failed to update project phase: %w", err)
	}
	return nil
}

// AddProjectPhase adds a phase to a project's phases list.
func (d *DB) AddProjectPhase(projectID string, phase Phase) error {
	p, err := d.GetProject(projectID)
	if err != nil {
		return err
	}

	p.Phases = append(p.Phases, phase)
	phasesJSON, _ := json.Marshal(p.Phases)

	_, err = d.conn.Exec(`UPDATE projects SET phases = ? WHERE id = ?`, string(phasesJSON), projectID)
	if err != nil {
		return fmt.Errorf("failed to add project phase: %w", err)
	}
	return nil
}

// UpdateProjectPhases replaces all phases for a project.
func (d *DB) UpdateProjectPhases(projectID string, phases []Phase) error {
	phasesJSON, _ := json.Marshal(phases)

	_, err := d.conn.Exec(`UPDATE projects SET phases = ? WHERE id = ?`, string(phasesJSON), projectID)
	if err != nil {
		return fmt.Errorf("failed to update project phases: %w", err)
	}
	return nil
}

// DeleteProject removes a project by ID.
func (d *DB) DeleteProject(id string) error {
	_, err := d.conn.Exec(`DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// UpdateProjectDescription updates the description of a project.
func (d *DB) UpdateProjectDescription(id, description string) error {
	_, err := d.conn.Exec(`UPDATE projects SET description = ? WHERE id = ?`, description, id)
	if err != nil {
		return fmt.Errorf("failed to update project description: %w", err)
	}
	return nil
}

// UpdateProjectMetadata updates the metadata of a project.
func (d *DB) UpdateProjectMetadata(id, metadata string) error {
	_, err := d.conn.Exec(`UPDATE projects SET metadata = ? WHERE id = ?`, metadata, id)
	if err != nil {
		return fmt.Errorf("failed to update project metadata: %w", err)
	}
	return nil
}

// UpdateProjectConceptID updates the concept ID of a project.
func (d *DB) UpdateProjectConceptID(id string, conceptID *string) error {
	_, err := d.conn.Exec(`UPDATE projects SET concept_id = ? WHERE id = ?`, conceptID, id)
	if err != nil {
		return fmt.Errorf("failed to update project concept_id: %w", err)
	}
	return nil
}

func scanProjects(rows *sql.Rows) ([]*Project, error) {
	var projects []*Project
	for rows.Next() {
		var p Project
		var phasesJSON string
		var metadataJSON string
		var conceptID sql.NullString

		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &conceptID, &p.CurrentPhase, &phasesJSON, &p.Status, &metadataJSON, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		if conceptID.Valid {
			p.ConceptID = &conceptID.String
		}

		if err := json.Unmarshal([]byte(phasesJSON), &p.Phases); err != nil {
			return nil, fmt.Errorf("failed to unmarshal phases: %w", err)
		}

		p.Metadata = metadataJSON

		projects = append(projects, &p)
	}
	return projects, rows.Err()
}
