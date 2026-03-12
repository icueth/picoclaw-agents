// Package project provides multi-role project workflow management.
package project

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/db"
	"picoclaw/agent/pkg/memory"
)

// ProjectManager handles multi-role project workflows.
type ProjectManager struct {
	db         *db.DB
	conceptMgr *memory.ConceptManager
	jobMgr     *memory.JobManager
	config     *config.Config
}

// NewProjectManager creates a new ProjectManager instance.
func NewProjectManager(database *db.DB, conceptMgr *memory.ConceptManager, jobMgr *memory.JobManager, cfg *config.Config) *ProjectManager {
	return &ProjectManager{
		db:         database,
		conceptMgr: conceptMgr,
		jobMgr:     jobMgr,
		config:     cfg,
	}
}

// CreateProject creates a new project with phases.
func (pm *ProjectManager) CreateProject(req CreateProjectRequest) (*Project, error) {
	if pm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	// Use default phases if none provided
	phaseNames := req.Phases
	if len(phaseNames) == 0 {
		phaseNames = DefaultPhases()
	}

	// Build phases
	phases := BuildPhases(phaseNames)

	// Create project in database
	var conceptIDPtr *string
	if req.ConceptID != "" {
		conceptIDPtr = &req.ConceptID
	}

	dbProject, err := pm.db.CreateProject(req.Name, conceptIDPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to create project in database: %w", err)
	}

	// Update with description and phases
	if err := pm.db.UpdateProjectDescription(dbProject.ID, req.Description); err != nil {
		return nil, fmt.Errorf("failed to update project description: %w", err)
	}

	// Convert phases to db.Phase format
	dbPhases := make([]db.Phase, len(phases))
	for i, p := range phases {
		dbPhases[i] = db.Phase{
			Role:   p.Role,
			Status: string(p.Status),
			JobID:  p.JobID,
		}
	}

	if err := pm.db.UpdateProjectPhases(dbProject.ID, dbPhases); err != nil {
		return nil, fmt.Errorf("failed to update project phases: %w", err)
	}

	// Store metadata if provided
	if req.Metadata != nil && len(req.Metadata) > 0 {
		metadataJSON, _ := json.Marshal(req.Metadata)
		if err := pm.db.UpdateProjectMetadata(dbProject.ID, string(metadataJSON)); err != nil {
			return nil, fmt.Errorf("failed to update project metadata: %w", err)
		}
	}

	// Convert to project type
	return pm.dbProjectToProject(dbProject)
}

// GetProject retrieves a project by ID.
func (pm *ProjectManager) GetProject(id string) (*Project, error) {
	if pm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	dbProject, err := pm.db.GetProject(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return pm.dbProjectToProject(dbProject)
}

// ListProjects returns all projects, optionally filtered by status.
func (pm *ProjectManager) ListProjects(status string) ([]Project, error) {
	if pm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	dbProjects, err := pm.db.ListProjects(status)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	projects := make([]Project, len(dbProjects))
	for i, dbProject := range dbProjects {
		project, err := pm.dbProjectToProject(dbProject)
		if err != nil {
			return nil, err
		}
		projects[i] = *project
	}

	return projects, nil
}

// ListProjectSummaries returns lightweight summaries of all projects.
func (pm *ProjectManager) ListProjectSummaries(status string) ([]ProjectSummary, error) {
	projects, err := pm.ListProjects(status)
	if err != nil {
		return nil, err
	}

	summaries := make([]ProjectSummary, len(projects))
	for i, p := range projects {
		summaries[i] = ProjectSummary{
			ID:           p.ID,
			Name:         p.Name,
			CurrentPhase: p.CurrentPhase,
			Status:       string(p.Status),
			Progress:     CalculateProgress(&p),
			CreatedAt:    p.CreatedAt,
			UpdatedAt:    p.UpdatedAt,
		}
	}

	return summaries, nil
}

// UpdateProjectStatus updates the status of a project.
func (pm *ProjectManager) UpdateProjectStatus(id, status string) error {
	if pm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if !IsValidProjectStatus(status) {
		return fmt.Errorf("invalid project status: %s", status)
	}

	if err := pm.db.UpdateProjectStatus(id, status); err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	return nil
}

// GetProjectStatus returns detailed status information for a project.
func (pm *ProjectManager) GetProjectStatus(id string) (*ProjectStatusDetail, error) {
	project, err := pm.GetProject(id)
	if err != nil {
		return nil, err
	}

	currentPhase := GetCurrentPhase(project)
	nextPhase := GetNextPhase(project)

	completedPhases := 0
	for _, phase := range project.Phases {
		if phase.Status == PhaseCompleted || phase.Status == PhaseSkipped {
			completedPhases++
		}
	}

	return &ProjectStatusDetail{
		Project:         *project,
		CurrentPhase:    currentPhase,
		Progress:        CalculateProgress(project),
		CompletedPhases: completedPhases,
		TotalPhases:     len(project.Phases),
		NextPhase:       nextPhase,
	}, nil
}

// NextPhase advances the project to the next phase.
// Returns the new current phase or an error if advancement is not possible.
func (pm *ProjectManager) NextPhase(projectID string) (*Phase, error) {
	if pm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	project, err := pm.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	// Check if project is active
	if project.Status != ProjectActive {
		return nil, fmt.Errorf("cannot advance phase: project is not active (status: %s)", project.Status)
	}

	// Get current phase
	currentPhase := GetCurrentPhase(project)
	if currentPhase == nil {
		return nil, fmt.Errorf("current phase not found: %s", project.CurrentPhase)
	}

	// Check if current phase is complete
	if !IsPhaseComplete(currentPhase.Status) {
		return nil, fmt.Errorf("cannot advance phase: current phase '%s' is not complete (status: %s)",
			currentPhase.Name, currentPhase.Status)
	}

	// Find next phase
	nextPhase := GetNextPhase(project)
	if nextPhase == nil {
		return nil, fmt.Errorf("no next phase available")
	}

	// Update current phase status to completed if not already
	if currentPhase.Status != PhaseCompleted && currentPhase.Status != PhaseSkipped {
		if err := pm.UpdatePhaseStatus(projectID, currentPhase.Name, PhaseCompleted); err != nil {
			return nil, fmt.Errorf("failed to update current phase status: %w", err)
		}
	}

	// Update project to next phase
	if err := pm.db.UpdateProjectPhase(projectID, nextPhase.Name); err != nil {
		return nil, fmt.Errorf("failed to update project phase: %w", err)
	}

	// Update next phase status to running
	if err := pm.UpdatePhaseStatus(projectID, nextPhase.Name, PhaseRunning); err != nil {
		return nil, fmt.Errorf("failed to update next phase status: %w", err)
	}

	// Set phase start time
	now := time.Now()
	if err := pm.UpdatePhaseStartedAt(projectID, nextPhase.Name, &now); err != nil {
		return nil, fmt.Errorf("failed to update phase start time: %w", err)
	}

	// Refresh project data
	updatedProject, err := pm.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	return GetCurrentPhase(updatedProject), nil
}

// StartPhase starts execution of a specific phase by creating a job.
func (pm *ProjectManager) StartPhase(projectID, phaseName, task string, context map[string]interface{}) (*Phase, error) {
	if pm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if pm.jobMgr == nil {
		return nil, fmt.Errorf("job manager not initialized")
	}

	project, err := pm.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	// Find the phase
	var targetPhase *Phase
	for i := range project.Phases {
		if project.Phases[i].Name == phaseName {
			targetPhase = &project.Phases[i]
			break
		}
	}
	if targetPhase == nil {
		return nil, fmt.Errorf("phase not found: %s", phaseName)
	}

	// Check if phase can be started
	if targetPhase.Status != PhasePending && targetPhase.Status != PhaseFailed {
		return nil, fmt.Errorf("phase '%s' cannot be started (status: %s)", phaseName, targetPhase.Status)
	}

	// Prepare job context
	jobContext := make(map[string]interface{})
	if context != nil {
		for k, v := range context {
			jobContext[k] = v
		}
	}
	jobContext["project_id"] = projectID
	jobContext["phase_name"] = phaseName
	jobContext["phase_role"] = targetPhase.Role
	if project.ConceptID != "" {
		jobContext["concept_id"] = project.ConceptID
	}

	// Create job for the phase
	jobID, err := pm.jobMgr.CreateJob(targetPhase.Role, task, jobContext)
	if err != nil {
		return nil, fmt.Errorf("failed to create job for phase: %w", err)
	}

	// Update phase with job ID and status
	if err := pm.UpdatePhaseJobID(projectID, phaseName, jobID); err != nil {
		return nil, fmt.Errorf("failed to update phase job ID: %w", err)
	}

	if err := pm.UpdatePhaseStatus(projectID, phaseName, PhaseRunning); err != nil {
		return nil, fmt.Errorf("failed to update phase status: %w", err)
	}

	// Set phase start time
	now := time.Now()
	if err := pm.UpdatePhaseStartedAt(projectID, phaseName, &now); err != nil {
		return nil, fmt.Errorf("failed to update phase start time: %w", err)
	}

	// Refresh and return
	updatedProject, err := pm.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	for i := range updatedProject.Phases {
		if updatedProject.Phases[i].Name == phaseName {
			return &updatedProject.Phases[i], nil
		}
	}

	return nil, fmt.Errorf("phase not found after update")
}

// CompletePhase marks a phase as completed with a result.
func (pm *ProjectManager) CompletePhase(projectID, phaseName, result string) error {
	if pm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	// Find the phase
	var targetPhase *Phase
	for i := range project.Phases {
		if project.Phases[i].Name == phaseName {
			targetPhase = &project.Phases[i]
			break
		}
	}
	if targetPhase == nil {
		return fmt.Errorf("phase not found: %s", phaseName)
	}

	// Update phase status and result
	if err := pm.UpdatePhaseStatus(projectID, phaseName, PhaseCompleted); err != nil {
		return fmt.Errorf("failed to update phase status: %w", err)
	}

	if err := pm.UpdatePhaseResult(projectID, phaseName, result); err != nil {
		return fmt.Errorf("failed to update phase result: %w", err)
	}

	// Set phase completion time
	now := time.Now()
	if err := pm.UpdatePhaseCompletedAt(projectID, phaseName, &now); err != nil {
		return fmt.Errorf("failed to update phase completion time: %w", err)
	}

	// Update job result if job exists
	if targetPhase.JobID != "" && pm.jobMgr != nil {
		if err := pm.jobMgr.UpdateJobResult(targetPhase.JobID, result); err != nil {
			return fmt.Errorf("failed to update job result: %w", err)
		}
		if err := pm.jobMgr.UpdateJobStatus(targetPhase.JobID, string(memory.JobCompleted)); err != nil {
			return fmt.Errorf("failed to update job status: %w", err)
		}
	}

	return nil
}

// FailPhase marks a phase as failed with an error message.
func (pm *ProjectManager) FailPhase(projectID, phaseName, errorMsg string) error {
	if pm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := pm.UpdatePhaseStatus(projectID, phaseName, PhaseFailed); err != nil {
		return fmt.Errorf("failed to update phase status: %w", err)
	}

	if err := pm.UpdatePhaseResult(projectID, phaseName, errorMsg); err != nil {
		return fmt.Errorf("failed to update phase result: %w", err)
	}

	// Update job status if job exists
	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	for _, phase := range project.Phases {
		if phase.Name == phaseName && phase.JobID != "" && pm.jobMgr != nil {
			if err := pm.jobMgr.UpdateJobResult(phase.JobID, errorMsg); err != nil {
				return fmt.Errorf("failed to update job result: %w", err)
			}
			if err := pm.jobMgr.UpdateJobStatus(phase.JobID, string(memory.JobFailed)); err != nil {
				return fmt.Errorf("failed to update job status: %w", err)
			}
			break
		}
	}

	return nil
}

// SkipPhase skips a phase (marks it as skipped).
func (pm *ProjectManager) SkipPhase(projectID, phaseName string) error {
	if pm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	return pm.UpdatePhaseStatus(projectID, phaseName, PhaseSkipped)
}

// UpdatePhaseStatus updates the status of a specific phase.
func (pm *ProjectManager) UpdatePhaseStatus(projectID, phaseName string, status PhaseStatus) error {
	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	// Find and update the phase
	for i := range project.Phases {
		if project.Phases[i].Name == phaseName {
			project.Phases[i].Status = status
			return pm.updateProjectPhases(projectID, project.Phases)
		}
	}

	return fmt.Errorf("phase not found: %s", phaseName)
}

// UpdatePhaseJobID updates the job ID for a specific phase.
func (pm *ProjectManager) UpdatePhaseJobID(projectID, phaseName, jobID string) error {
	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	for i := range project.Phases {
		if project.Phases[i].Name == phaseName {
			project.Phases[i].JobID = jobID
			return pm.updateProjectPhases(projectID, project.Phases)
		}
	}

	return fmt.Errorf("phase not found: %s", phaseName)
}

// UpdatePhaseResult updates the result for a specific phase.
func (pm *ProjectManager) UpdatePhaseResult(projectID, phaseName, result string) error {
	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	for i := range project.Phases {
		if project.Phases[i].Name == phaseName {
			project.Phases[i].Result = result
			return pm.updateProjectPhases(projectID, project.Phases)
		}
	}

	return fmt.Errorf("phase not found: %s", phaseName)
}

// UpdatePhaseStartedAt updates the start time for a specific phase.
func (pm *ProjectManager) UpdatePhaseStartedAt(projectID, phaseName string, startedAt *time.Time) error {
	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	for i := range project.Phases {
		if project.Phases[i].Name == phaseName {
			project.Phases[i].StartedAt = startedAt
			return pm.updateProjectPhases(projectID, project.Phases)
		}
	}

	return fmt.Errorf("phase not found: %s", phaseName)
}

// UpdatePhaseCompletedAt updates the completion time for a specific phase.
func (pm *ProjectManager) UpdatePhaseCompletedAt(projectID, phaseName string, completedAt *time.Time) error {
	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	for i := range project.Phases {
		if project.Phases[i].Name == phaseName {
			project.Phases[i].CompletedAt = completedAt
			return pm.updateProjectPhases(projectID, project.Phases)
		}
	}

	return fmt.Errorf("phase not found: %s", phaseName)
}

// updateProjectPhases updates all phases for a project in the database.
func (pm *ProjectManager) updateProjectPhases(projectID string, phases []Phase) error {
	dbPhases := make([]db.Phase, len(phases))
	for i, p := range phases {
		dbPhases[i] = db.Phase{
			Role:   p.Role,
			Status: string(p.Status),
			JobID:  p.JobID,
		}
	}

	return pm.db.UpdateProjectPhases(projectID, dbPhases)
}

// DeleteProject permanently deletes a project.
func (pm *ProjectManager) DeleteProject(id string) error {
	if pm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := pm.db.DeleteProject(id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// PauseProject pauses a project.
func (pm *ProjectManager) PauseProject(id string) error {
	return pm.UpdateProjectStatus(id, string(ProjectPaused))
}

// ResumeProject resumes a paused project.
func (pm *ProjectManager) ResumeProject(id string) error {
	return pm.UpdateProjectStatus(id, string(ProjectActive))
}

// CompleteProject marks a project as completed.
func (pm *ProjectManager) CompleteProject(id string) error {
	return pm.UpdateProjectStatus(id, string(ProjectCompleted))
}

// CancelProject cancels a project.
func (pm *ProjectManager) CancelProject(id string) error {
	return pm.UpdateProjectStatus(id, string(ProjectCancelled))
}

// GetPhaseJob retrieves the job associated with a phase.
func (pm *ProjectManager) GetPhaseJob(projectID, phaseName string) (*memory.Job, error) {
	if pm.jobMgr == nil {
		return nil, fmt.Errorf("job manager not initialized")
	}

	project, err := pm.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	for _, phase := range project.Phases {
		if phase.Name == phaseName && phase.JobID != "" {
			return pm.jobMgr.GetJob(phase.JobID)
		}
	}

	return nil, fmt.Errorf("no job found for phase: %s", phaseName)
}

// GetProjectConcept retrieves the concept associated with a project.
func (pm *ProjectManager) GetProjectConcept(projectID string) (*memory.Concept, error) {
	if pm.conceptMgr == nil {
		return nil, fmt.Errorf("concept manager not initialized")
	}

	project, err := pm.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	if project.ConceptID == "" {
		return nil, fmt.Errorf("project has no associated concept")
	}

	return pm.conceptMgr.GetConcept(project.ConceptID)
}

// LinkConcept links a concept to a project.
func (pm *ProjectManager) LinkConcept(projectID, conceptID string) error {
	if pm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := pm.db.UpdateProjectConceptID(projectID, &conceptID); err != nil {
		return fmt.Errorf("failed to link concept to project: %w", err)
	}

	return nil
}

// UnlinkConcept removes the concept association from a project.
func (pm *ProjectManager) UnlinkConcept(projectID string) error {
	if pm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := pm.db.UpdateProjectConceptID(projectID, nil); err != nil {
		return fmt.Errorf("failed to unlink concept from project: %w", err)
	}

	return nil
}

// dbProjectToProject converts a database project to a project.Project.
func (pm *ProjectManager) dbProjectToProject(dbProject *db.Project) (*Project, error) {
	project := &Project{
		ID:           dbProject.ID,
		Name:         dbProject.Name,
		CurrentPhase: dbProject.CurrentPhase,
		Status:       ProjectStatus(dbProject.Status),
		CreatedAt:    dbProject.CreatedAt,
		UpdatedAt:    dbProject.UpdatedAt,
		Metadata:     make(map[string]interface{}),
	}

	if dbProject.ConceptID != nil {
		project.ConceptID = *dbProject.ConceptID
	}

	// Convert phases
	project.Phases = make([]Phase, len(dbProject.Phases))
	for i, dbPhase := range dbProject.Phases {
		project.Phases[i] = Phase{
			Name:   pm.getPhaseNameByOrder(i),
			Role:   dbPhase.Role,
			Status: PhaseStatus(dbPhase.Status),
			JobID:  dbPhase.JobID,
			Order:  i,
		}
	}

	// Set current order based on current phase
	for i, phase := range project.Phases {
		if phase.Name == project.CurrentPhase {
			project.CurrentOrder = i
			break
		}
	}

	// Try to get description and metadata from database
	if pm.db != nil {
		// These would need to be added to the db package
		// For now, we'll leave them empty
		project.Description = ""
	}

	return project, nil
}

// getPhaseNameByOrder returns a phase name based on its order.
// This is a helper since db.Phase doesn't store the name.
func (pm *ProjectManager) getPhaseNameByOrder(order int) string {
	defaultPhases := DefaultPhases()
	if order < len(defaultPhases) {
		return defaultPhases[order]
	}
	return fmt.Sprintf("phase_%d", order)
}

// GenerateProjectID generates a new unique project ID.
func GenerateProjectID() string {
	return uuid.New().String()
}
