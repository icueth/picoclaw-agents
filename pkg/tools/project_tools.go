// Package tools provides tool handlers for project management operations.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"picoclaw/agent/pkg/project"
)

// =============================================================================
// CreateProjectTool
// =============================================================================

// CreateProjectTool creates a new multi-role project.
type CreateProjectTool struct {
	projectManager *project.ProjectManager
}

// NewCreateProjectTool creates a new CreateProjectTool instance.
func NewCreateProjectTool(pm *project.ProjectManager) *CreateProjectTool {
	return &CreateProjectTool{
		projectManager: pm,
	}
}

// Name returns the tool name.
func (t *CreateProjectTool) Name() string {
	return "create_project"
}

// Description returns the tool description.
func (t *CreateProjectTool) Description() string {
	return "Create a new multi-role project with phases. Projects support workflows like planning → research → coding → review. Each phase spawns a subagent with the appropriate role. Returns the project ID for tracking."
}

// Parameters returns the tool parameters schema.
func (t *CreateProjectTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Name of the project",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "Description of what the project aims to accomplish",
			},
			"concept_id": map[string]any{
				"type":        "string",
				"description": "Optional concept ID to associate with this project",
			},
			"phases": map[string]any{
				"type":        "array",
				"description": "List of phase names (e.g., [\"planning\", \"research\", \"coding\", \"review\"]). If not provided, uses default phases.",
				"items": map[string]any{
					"type": "string",
					"enum": []string{"planning", "research", "coding", "review", "testing", "deploy", "document"},
				},
			},
			"metadata": map[string]any{
				"type":        "object",
				"description": "Optional metadata to store with the project",
			},
		},
		"required": []string{"name", "description"},
	}
}

// Execute creates a new project.
func (t *CreateProjectTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.projectManager == nil {
		return ErrorResult("Project manager not initialized").WithError(fmt.Errorf("project manager is nil"))
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return ErrorResult("name is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid name parameter"))
	}

	description, ok := args["description"].(string)
	if !ok || description == "" {
		return ErrorResult("description is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid description parameter"))
	}

	req := project.CreateProjectRequest{
		Name:        name,
		Description: description,
	}

	if conceptID, ok := args["concept_id"].(string); ok && conceptID != "" {
		req.ConceptID = conceptID
	}

	if phasesArg, ok := args["phases"].([]interface{}); ok && len(phasesArg) > 0 {
		phases := make([]string, len(phasesArg))
		for i, p := range phasesArg {
			if phaseStr, ok := p.(string); ok {
				phases[i] = phaseStr
			}
		}
		req.Phases = phases
	}

	if metadata, ok := args["metadata"].(map[string]interface{}); ok {
		req.Metadata = metadata
	}

	proj, err := t.projectManager.CreateProject(req)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to create project: %v", err)).WithError(err)
	}

	// Build response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Project created successfully.\n\n"))
	sb.WriteString(fmt.Sprintf("ID: %s\n", proj.ID))
	sb.WriteString(fmt.Sprintf("Name: %s\n", proj.Name))
	sb.WriteString(fmt.Sprintf("Description: %s\n", proj.Description))
	sb.WriteString(fmt.Sprintf("Status: %s\n", proj.Status))
	sb.WriteString(fmt.Sprintf("Current Phase: %s\n", proj.CurrentPhase))
	sb.WriteString(fmt.Sprintf("Phases (%d):\n", len(proj.Phases)))
	for i, phase := range proj.Phases {
		sb.WriteString(fmt.Sprintf("  %d. %s (%s) - %s\n", i+1, phase.Name, phase.Role, phase.Status))
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Created project: %s (ID: %s)", proj.Name, proj.ID),
		Silent:  false,
		IsError: false,
	}
}

// =============================================================================
// ListProjectsTool
// =============================================================================

// ListProjectsTool lists all projects, optionally filtered by status.
type ListProjectsTool struct {
	projectManager *project.ProjectManager
}

// NewListProjectsTool creates a new ListProjectsTool instance.
func NewListProjectsTool(pm *project.ProjectManager) *ListProjectsTool {
	return &ListProjectsTool{
		projectManager: pm,
	}
}

// Name returns the tool name.
func (t *ListProjectsTool) Name() string {
	return "list_projects"
}

// Description returns the tool description.
func (t *ListProjectsTool) Description() string {
	return "List all multi-role projects, optionally filtered by status. Shows project ID, name, current phase, status, and progress percentage. Use this to find projects to work on or monitor."
}

// Parameters returns the tool parameters schema.
func (t *ListProjectsTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status": map[string]any{
				"type":        "string",
				"description": "Optional status filter: active, paused, completed, cancelled",
				"enum":        []string{"active", "paused", "completed", "cancelled", ""},
			},
			"limit": map[string]any{
				"type":        "number",
				"description": "Maximum number of projects to return (default: 50)",
			},
		},
		"required": []string{},
	}
}

// Execute lists projects.
func (t *ListProjectsTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.projectManager == nil {
		return ErrorResult("Project manager not initialized").WithError(fmt.Errorf("project manager is nil"))
	}

	status := ""
	if s, ok := args["status"].(string); ok {
		status = s
	}

	limit := 50
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	summaries, err := t.projectManager.ListProjectSummaries(status)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to list projects: %v", err)).WithError(err)
	}

	// Apply limit
	if limit > 0 && len(summaries) > limit {
		summaries = summaries[:limit]
	}

	if len(summaries) == 0 {
		return &ToolResult{
			ForLLM:  "No projects found.",
			ForUser: "No projects found.",
			Silent:  false,
			IsError: false,
		}
	}

	// Build response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d project(s):\n\n", len(summaries)))

	for i, s := range summaries {
		sb.WriteString(fmt.Sprintf("%d. ID: %s\n", i+1, s.ID))
		sb.WriteString(fmt.Sprintf("   Name: %s\n", s.Name))
		sb.WriteString(fmt.Sprintf("   Current Phase: %s\n", s.CurrentPhase))
		sb.WriteString(fmt.Sprintf("   Status: %s\n", s.Status))
		sb.WriteString(fmt.Sprintf("   Progress: %.1f%%\n", s.Progress))
		sb.WriteString(fmt.Sprintf("   Created: %s\n", s.CreatedAt.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("   Updated: %s\n", s.UpdatedAt.Format(time.RFC3339)))
		sb.WriteString("\n")
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Found %d project(s)", len(summaries)),
		Silent:  false,
		IsError: false,
	}
}

// =============================================================================
// GetProjectStatusTool
// =============================================================================

// GetProjectStatusTool gets detailed status of a project.
type GetProjectStatusTool struct {
	projectManager *project.ProjectManager
}

// NewGetProjectStatusTool creates a new GetProjectStatusTool instance.
func NewGetProjectStatusTool(pm *project.ProjectManager) *GetProjectStatusTool {
	return &GetProjectStatusTool{
		projectManager: pm,
	}
}

// Name returns the tool name.
func (t *GetProjectStatusTool) Name() string {
	return "get_project_status"
}

// Description returns the tool description.
func (t *GetProjectStatusTool) Description() string {
	return "Get detailed status information for a project. Shows all phases, their statuses, progress percentage, and which phase is currently active. Use this to check project progress before advancing phases."
}

// Parameters returns the tool parameters schema.
func (t *GetProjectStatusTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_id": map[string]any{
				"type":        "string",
				"description": "The ID of the project to get status for",
			},
		},
		"required": []string{"project_id"},
	}
}

// Execute gets project status.
func (t *GetProjectStatusTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.projectManager == nil {
		return ErrorResult("Project manager not initialized").WithError(fmt.Errorf("project manager is nil"))
	}

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return ErrorResult("project_id is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid project_id parameter"))
	}

	status, err := t.projectManager.GetProjectStatus(projectID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get project status: %v", err)).WithError(err)
	}

	// Build detailed response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Project Status: %s\n\n", status.Project.Name))
	sb.WriteString(fmt.Sprintf("ID: %s\n", status.Project.ID))
	sb.WriteString(fmt.Sprintf("Description: %s\n", status.Project.Description))
	sb.WriteString(fmt.Sprintf("Status: %s\n", status.Project.Status))
	sb.WriteString(fmt.Sprintf("Progress: %.1f%% (%d/%d phases completed)\n",
		status.Progress, status.CompletedPhases, status.TotalPhases))
	sb.WriteString(fmt.Sprintf("Current Phase: %s\n", status.Project.CurrentPhase))

	if status.Project.ConceptID != "" {
		sb.WriteString(fmt.Sprintf("Concept ID: %s\n", status.Project.ConceptID))
	}

	sb.WriteString(fmt.Sprintf("\nPhases:\n"))
	for _, phase := range status.Project.Phases {
		marker := "  "
		if phase.Name == status.Project.CurrentPhase {
			marker = "→ "
		}
		sb.WriteString(fmt.Sprintf("%s%s (%s) - %s", marker, phase.Name, phase.Role, phase.Status))

		if phase.JobID != "" {
			sb.WriteString(fmt.Sprintf(" [Job: %s]", phase.JobID))
		}
		if phase.StartedAt != nil {
			sb.WriteString(fmt.Sprintf(" [Started: %s]", phase.StartedAt.Format(time.RFC3339)))
		}
		if phase.CompletedAt != nil {
			sb.WriteString(fmt.Sprintf(" [Completed: %s]", phase.CompletedAt.Format(time.RFC3339)))
		}
		sb.WriteString("\n")

		if phase.Result != "" {
			result := truncateString(phase.Result, 200)
			sb.WriteString(fmt.Sprintf("    Result: %s\n", result))
		}
	}

	if status.NextPhase != nil {
		sb.WriteString(fmt.Sprintf("\nNext Phase: %s (%s)\n", status.NextPhase.Name, status.NextPhase.Role))
	} else {
		sb.WriteString(fmt.Sprintf("\nNo next phase - project is complete or at final phase.\n"))
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Project '%s' status: %.1f%% complete (%s phase)", status.Project.Name, status.Progress, status.Project.CurrentPhase),
		Silent:  false,
		IsError: false,
	}
}

// =============================================================================
// NextPhaseTool
// =============================================================================

// NextPhaseTool advances a project to the next phase.
type NextPhaseTool struct {
	projectManager *project.ProjectManager
}

// NewNextPhaseTool creates a new NextPhaseTool instance.
func NewNextPhaseTool(pm *project.ProjectManager) *NextPhaseTool {
	return &NextPhaseTool{
		projectManager: pm,
	}
}

// Name returns the tool name.
func (t *NextPhaseTool) Name() string {
	return "next_phase"
}

// Description returns the tool description.
func (t *NextPhaseTool) Description() string {
	return "Advance a project to the next phase. The current phase must be completed before advancing. Returns the new current phase information. Use this after a phase has finished its work."
}

// Parameters returns the tool parameters schema.
func (t *NextPhaseTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_id": map[string]any{
				"type":        "string",
				"description": "The ID of the project to advance",
			},
		},
		"required": []string{"project_id"},
	}
}

// Execute advances to the next phase.
func (t *NextPhaseTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.projectManager == nil {
		return ErrorResult("Project manager not initialized").WithError(fmt.Errorf("project manager is nil"))
	}

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return ErrorResult("project_id is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid project_id parameter"))
	}

	newPhase, err := t.projectManager.NextPhase(projectID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to advance phase: %v", err)).WithError(err)
	}

	// Get updated project status
	status, err := t.projectManager.GetProjectStatus(projectID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Phase advanced but failed to get status: %v", err)).WithError(err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Project advanced to next phase successfully.\n\n"))
	sb.WriteString(fmt.Sprintf("Project: %s\n", status.Project.Name))
	sb.WriteString(fmt.Sprintf("Current Phase: %s (%s)\n", newPhase.Name, newPhase.Role))
	sb.WriteString(fmt.Sprintf("Phase Status: %s\n", newPhase.Status))
	sb.WriteString(fmt.Sprintf("Overall Progress: %.1f%% (%d/%d phases)\n",
		status.Progress, status.CompletedPhases, status.TotalPhases))

	if newPhase.StartedAt != nil {
		sb.WriteString(fmt.Sprintf("Phase Started: %s\n", newPhase.StartedAt.Format(time.RFC3339)))
	}

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Advanced '%s' to %s phase", status.Project.Name, newPhase.Name),
		Silent:  false,
		IsError: false,
	}
}

// =============================================================================
// StartPhaseTool
// =============================================================================

// StartPhaseTool starts execution of a specific project phase.
type StartPhaseTool struct {
	projectManager *project.ProjectManager
}

// NewStartPhaseTool creates a new StartPhaseTool instance.
func NewStartPhaseTool(pm *project.ProjectManager) *StartPhaseTool {
	return &StartPhaseTool{
		projectManager: pm,
	}
}

// Name returns the tool name.
func (t *StartPhaseTool) Name() string {
	return "start_phase"
}

// Description returns the tool description.
func (t *StartPhaseTool) Description() string {
	return "Start execution of a specific project phase. Creates a job for the phase's role and marks the phase as running. Use this to begin work on a phase."
}

// Parameters returns the tool parameters schema.
func (t *StartPhaseTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_id": map[string]any{
				"type":        "string",
				"description": "The ID of the project",
			},
			"phase_name": map[string]any{
				"type":        "string",
				"description": "The name of the phase to start",
			},
			"task": map[string]any{
				"type":        "string",
				"description": "Description of the task for this phase",
			},
			"context": map[string]any{
				"type":        "object",
				"description": "Optional context data to pass to the phase",
			},
		},
		"required": []string{"project_id", "phase_name", "task"},
	}
}

// Execute starts a phase.
func (t *StartPhaseTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.projectManager == nil {
		return ErrorResult("Project manager not initialized").WithError(fmt.Errorf("project manager is nil"))
	}

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return ErrorResult("project_id is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid project_id parameter"))
	}

	phaseName, ok := args["phase_name"].(string)
	if !ok || phaseName == "" {
		return ErrorResult("phase_name is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid phase_name parameter"))
	}

	task, ok := args["task"].(string)
	if !ok || task == "" {
		return ErrorResult("task is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid task parameter"))
	}

	var taskContext map[string]interface{}
	if ctxArg, ok := args["context"].(map[string]interface{}); ok {
		taskContext = ctxArg
	}

	phase, err := t.projectManager.StartPhase(projectID, phaseName, task, taskContext)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to start phase: %v", err)).WithError(err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Phase started successfully.\n\n"))
	sb.WriteString(fmt.Sprintf("Phase: %s (%s)\n", phase.Name, phase.Role))
	sb.WriteString(fmt.Sprintf("Status: %s\n", phase.Status))
	sb.WriteString(fmt.Sprintf("Job ID: %s\n", phase.JobID))
	sb.WriteString(fmt.Sprintf("Task: %s\n", task))

	return &ToolResult{
		ForLLM:  sb.String(),
		ForUser: fmt.Sprintf("Started %s phase (Job: %s)", phase.Name, phase.JobID),
		Silent:  false,
		IsError: false,
	}
}

// =============================================================================
// CompletePhaseTool
// =============================================================================

// CompletePhaseTool marks a project phase as completed.
type CompletePhaseTool struct {
	projectManager *project.ProjectManager
}

// NewCompletePhaseTool creates a new CompletePhaseTool instance.
func NewCompletePhaseTool(pm *project.ProjectManager) *CompletePhaseTool {
	return &CompletePhaseTool{
		projectManager: pm,
	}
}

// Name returns the tool name.
func (t *CompletePhaseTool) Name() string {
	return "complete_phase"
}

// Description returns the tool description.
func (t *CompletePhaseTool) Description() string {
	return "Mark a project phase as completed with results. This allows the project to advance to the next phase. Use this when a phase has finished its work successfully."
}

// Parameters returns the tool parameters schema.
func (t *CompletePhaseTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_id": map[string]any{
				"type":        "string",
				"description": "The ID of the project",
			},
			"phase_name": map[string]any{
				"type":        "string",
				"description": "The name of the phase to complete",
			},
			"result": map[string]any{
				"type":        "string",
				"description": "The result/output of the phase (e.g., architecture plan, research findings, code summary)",
			},
		},
		"required": []string{"project_id", "phase_name", "result"},
	}
}

// Execute completes a phase.
func (t *CompletePhaseTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.projectManager == nil {
		return ErrorResult("Project manager not initialized").WithError(fmt.Errorf("project manager is nil"))
	}

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return ErrorResult("project_id is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid project_id parameter"))
	}

	phaseName, ok := args["phase_name"].(string)
	if !ok || phaseName == "" {
		return ErrorResult("phase_name is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid phase_name parameter"))
	}

	result, ok := args["result"].(string)
	if !ok || result == "" {
		return ErrorResult("result is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid result parameter"))
	}

	if err := t.projectManager.CompletePhase(projectID, phaseName, result); err != nil {
		return ErrorResult(fmt.Sprintf("Failed to complete phase: %v", err)).WithError(err)
	}

	// Get updated status
	status, err := t.projectManager.GetProjectStatus(projectID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Phase completed but failed to get status: %v", err)).WithError(err)
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("Phase '%s' completed successfully.\n\nResult:\n%s\n\nProject Progress: %.1f%%", phaseName, result, status.Progress),
		ForUser: fmt.Sprintf("Completed %s phase. Project progress: %.1f%%", phaseName, status.Progress),
		Silent:  false,
		IsError: false,
	}
}

// =============================================================================
// UpdateProjectStatusTool
// =============================================================================

// UpdateProjectStatusTool updates the status of a project.
type UpdateProjectStatusTool struct {
	projectManager *project.ProjectManager
}

// NewUpdateProjectStatusTool creates a new UpdateProjectStatusTool instance.
func NewUpdateProjectStatusTool(pm *project.ProjectManager) *UpdateProjectStatusTool {
	return &UpdateProjectStatusTool{
		projectManager: pm,
	}
}

// Name returns the tool name.
func (t *UpdateProjectStatusTool) Name() string {
	return "update_project_status"
}

// Description returns the tool description.
func (t *UpdateProjectStatusTool) Description() string {
	return "Update the status of a project (active, paused, completed, cancelled). Use this to pause, resume, or complete projects."
}

// Parameters returns the tool parameters schema.
func (t *UpdateProjectStatusTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_id": map[string]any{
				"type":        "string",
				"description": "The ID of the project",
			},
			"status": map[string]any{
				"type":        "string",
				"description": "New status for the project",
				"enum":        []string{"active", "paused", "completed", "cancelled"},
			},
		},
		"required": []string{"project_id", "status"},
	}
}

// Execute updates project status.
func (t *UpdateProjectStatusTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.projectManager == nil {
		return ErrorResult("Project manager not initialized").WithError(fmt.Errorf("project manager is nil"))
	}

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return ErrorResult("project_id is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid project_id parameter"))
	}

	status, ok := args["status"].(string)
	if !ok || status == "" {
		return ErrorResult("status is required and must be a non-empty string").WithError(fmt.Errorf("missing or invalid status parameter"))
	}

	if err := t.projectManager.UpdateProjectStatus(projectID, status); err != nil {
		return ErrorResult(fmt.Sprintf("Failed to update project status: %v", err)).WithError(err)
	}

	// Get project info
	proj, err := t.projectManager.GetProject(projectID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Status updated but failed to get project: %v", err)).WithError(err)
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("Project '%s' status updated to '%s'.", proj.Name, status),
		ForUser: fmt.Sprintf("Project '%s' is now %s", proj.Name, status),
		Silent:  false,
		IsError: false,
	}
}

// =============================================================================
// ProjectToolsRegistry
// =============================================================================

// ProjectToolsRegistry holds all project-related tools.
type ProjectToolsRegistry struct {
	CreateProject      *CreateProjectTool
	ListProjects       *ListProjectsTool
	GetProjectStatus   *GetProjectStatusTool
	NextPhase          *NextPhaseTool
	StartPhase         *StartPhaseTool
	CompletePhase      *CompletePhaseTool
	UpdateProjectStatus *UpdateProjectStatusTool
}

// NewProjectToolsRegistry creates a new registry with all project tools initialized.
func NewProjectToolsRegistry(pm *project.ProjectManager) *ProjectToolsRegistry {
	return &ProjectToolsRegistry{
		CreateProject:      NewCreateProjectTool(pm),
		ListProjects:       NewListProjectsTool(pm),
		GetProjectStatus:   NewGetProjectStatusTool(pm),
		NextPhase:          NewNextPhaseTool(pm),
		StartPhase:         NewStartPhaseTool(pm),
		CompletePhase:      NewCompletePhaseTool(pm),
		UpdateProjectStatus: NewUpdateProjectStatusTool(pm),
	}
}

// RegisterAll registers all project tools with the given tool registry.
func (r *ProjectToolsRegistry) RegisterAll(registry *ToolRegistry) {
	if r.CreateProject != nil {
		registry.Register(r.CreateProject)
	}
	if r.ListProjects != nil {
		registry.Register(r.ListProjects)
	}
	if r.GetProjectStatus != nil {
		registry.Register(r.GetProjectStatus)
	}
	if r.NextPhase != nil {
		registry.Register(r.NextPhase)
	}
	if r.StartPhase != nil {
		registry.Register(r.StartPhase)
	}
	if r.CompletePhase != nil {
		registry.Register(r.CompletePhase)
	}
	if r.UpdateProjectStatus != nil {
		registry.Register(r.UpdateProjectStatus)
	}
}

// GetAll returns all tools as a slice.
func (r *ProjectToolsRegistry) GetAll() []Tool {
	tools := make([]Tool, 0, 7)
	if r.CreateProject != nil {
		tools = append(tools, r.CreateProject)
	}
	if r.ListProjects != nil {
		tools = append(tools, r.ListProjects)
	}
	if r.GetProjectStatus != nil {
		tools = append(tools, r.GetProjectStatus)
	}
	if r.NextPhase != nil {
		tools = append(tools, r.NextPhase)
	}
	if r.StartPhase != nil {
		tools = append(tools, r.StartPhase)
	}
	if r.CompletePhase != nil {
		tools = append(tools, r.CompletePhase)
	}
	if r.UpdateProjectStatus != nil {
		tools = append(tools, r.UpdateProjectStatus)
	}
	return tools
}

// ToJSON returns all tools as a JSON array for LLM function calling.
func (r *ProjectToolsRegistry) ToJSON() ([]byte, error) {
	tools := r.GetAll()
	schemas := make([]map[string]any, len(tools))
	for i, tool := range tools {
		schemas[i] = ToolToSchema(tool)
	}
	return json.MarshalIndent(schemas, "", "  ")
}
