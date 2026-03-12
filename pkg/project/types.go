// Package project provides multi-role project workflow management.
package project

import (
	"time"
)

// ProjectStatus represents the lifecycle status of a project.
type ProjectStatus string

const (
	// ProjectActive - Currently being worked on
	ProjectActive ProjectStatus = "active"
	// ProjectPaused - Temporarily stopped
	ProjectPaused ProjectStatus = "paused"
	// ProjectCompleted - Finished successfully
	ProjectCompleted ProjectStatus = "completed"
	// ProjectCancelled - Cancelled before completion
	ProjectCancelled ProjectStatus = "cancelled"
)

// ValidProjectStatuses returns all valid project statuses.
func ValidProjectStatuses() []string {
	return []string{
		string(ProjectActive),
		string(ProjectPaused),
		string(ProjectCompleted),
		string(ProjectCancelled),
	}
}

// IsValidProjectStatus checks if a status is valid.
func IsValidProjectStatus(status string) bool {
	for _, s := range ValidProjectStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// PhaseStatus represents the status of a project phase.
type PhaseStatus string

const (
	// PhasePending - Phase not yet started
	PhasePending PhaseStatus = "pending"
	// PhaseRunning - Phase currently executing
	PhaseRunning PhaseStatus = "running"
	// PhaseCompleted - Phase finished successfully
	PhaseCompleted PhaseStatus = "completed"
	// PhaseFailed - Phase encountered an error
	PhaseFailed PhaseStatus = "failed"
	// PhaseSkipped - Phase was skipped
	PhaseSkipped PhaseStatus = "skipped"
)

// ValidPhaseStatuses returns all valid phase statuses.
func ValidPhaseStatuses() []string {
	return []string{
		string(PhasePending),
		string(PhaseRunning),
		string(PhaseCompleted),
		string(PhaseFailed),
		string(PhaseSkipped),
	}
}

// IsValidPhaseStatus checks if a status is valid.
func IsValidPhaseStatus(status string) bool {
	for _, s := range ValidPhaseStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// Phase represents a project phase in a multi-role workflow.
type Phase struct {
	Name        string      `json:"name"`         // e.g., "planning"
	Role        string      `json:"role"`         // e.g., "planner"
	Status      PhaseStatus `json:"status"`       // pending, running, completed, failed, skipped
	JobID       string      `json:"job_id"`       // associated job ID
	Order       int         `json:"order"`        // phase order (0-indexed)
	Description string      `json:"description"`  // phase description
	Result      string      `json:"result"`       // phase result/output
	StartedAt   *time.Time  `json:"started_at"`   // when phase started
	CompletedAt *time.Time  `json:"completed_at"` // when phase completed
}

// Project represents a multi-role workflow project.
type Project struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	ConceptID    string      `json:"concept_id,omitempty"` // optional concept association
	CurrentPhase string      `json:"current_phase"`        // name of current phase
	CurrentOrder int         `json:"current_order"`        // index of current phase
	Phases       []Phase     `json:"phases"`
	Status       ProjectStatus `json:"status"`             // active, paused, completed, cancelled
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ProjectSummary provides a lightweight summary of a project.
type ProjectSummary struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	CurrentPhase string    `json:"current_phase"`
	Status       string    `json:"status"`
	Progress     float64   `json:"progress"` // percentage (0-100)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ProjectStatusDetail provides detailed status information.
type ProjectStatusDetail struct {
	Project      Project       `json:"project"`
	CurrentPhase *Phase        `json:"current_phase,omitempty"`
	Progress     float64       `json:"progress"`
	CompletedPhases int       `json:"completed_phases"`
	TotalPhases  int           `json:"total_phases"`
	NextPhase    *Phase        `json:"next_phase,omitempty"`
}

// CreateProjectRequest contains parameters for creating a new project.
type CreateProjectRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ConceptID   string   `json:"concept_id,omitempty"` // optional
	Phases      []string `json:"phases"`               // e.g., ["planning", "research", "coding", "review"]
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DefaultPhases returns the default phase sequence for a project.
func DefaultPhases() []string {
	return []string{"planning", "research", "coding", "review"}
}

// PhaseToRole maps phase names to their corresponding subagent roles.
var PhaseToRole = map[string]string{
	"planning":  "planner",
	"research":  "researcher",
	"coding":    "coder",
	"review":    "reviewer",
	"testing":   "tester",
	"deploy":    "executor",
	"document":  "documenter",
}

// RoleToPhase maps subagent roles to their corresponding phase names (reverse mapping).
var RoleToPhase = map[string]string{
	"planner":    "planning",
	"researcher": "research",
	"coder":      "coding",
	"reviewer":   "review",
	"tester":     "testing",
	"executor":   "deploy",
	"documenter": "document",
}

// GetRoleForPhase returns the subagent role for a given phase name.
func GetRoleForPhase(phaseName string) string {
	if role, ok := PhaseToRole[phaseName]; ok {
		return role
	}
	return "specialist" // default role for unknown phases
}

// GetPhaseForRole returns the phase name for a given subagent role.
func GetPhaseForRole(role string) string {
	if phase, ok := RoleToPhase[role]; ok {
		return phase
	}
	return ""
}

// BuildPhases constructs phase objects from phase names.
func BuildPhases(phaseNames []string) []Phase {
	phases := make([]Phase, len(phaseNames))
	for i, name := range phaseNames {
		phases[i] = Phase{
			Name:   name,
			Role:   GetRoleForPhase(name),
			Status: PhasePending,
			Order:  i,
		}
	}
	return phases
}

// CalculateProgress calculates the completion percentage of a project.
func CalculateProgress(project *Project) float64 {
	if len(project.Phases) == 0 {
		return 0.0
	}

	completed := 0
	for _, phase := range project.Phases {
		if phase.Status == PhaseCompleted || phase.Status == PhaseSkipped {
			completed++
		}
	}

	return float64(completed) / float64(len(project.Phases)) * 100.0
}

// GetCurrentPhase returns the current active phase of a project.
func GetCurrentPhase(project *Project) *Phase {
	for i := range project.Phases {
		if project.Phases[i].Name == project.CurrentPhase {
			return &project.Phases[i]
		}
	}
	return nil
}

// GetNextPhase returns the next pending phase after the current one.
func GetNextPhase(project *Project) *Phase {
	for i := range project.Phases {
		if project.Phases[i].Order > project.CurrentOrder &&
			(project.Phases[i].Status == PhasePending || project.Phases[i].Status == PhaseFailed) {
			return &project.Phases[i]
		}
	}
	return nil
}

// IsPhaseComplete checks if a phase is in a terminal state.
func IsPhaseComplete(status PhaseStatus) bool {
	return status == PhaseCompleted || status == PhaseSkipped || status == PhaseFailed
}

// CanAdvancePhase checks if the project can advance to the next phase.
func CanAdvancePhase(project *Project) bool {
	// Project must be active to advance
	if project.Status != ProjectActive {
		return false
	}

	currentPhase := GetCurrentPhase(project)
	if currentPhase == nil {
		return false
	}

	// Can advance if current phase is complete
	if !IsPhaseComplete(currentPhase.Status) {
		return false
	}

	// Check if there's a next phase
	return GetNextPhase(project) != nil
}
