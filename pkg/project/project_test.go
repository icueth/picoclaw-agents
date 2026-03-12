// Package project provides tests for multi-role project workflow management.
package project

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Types Tests
// =============================================================================

func TestValidProjectStatuses(t *testing.T) {
	statuses := ValidProjectStatuses()
	expected := []string{"active", "paused", "completed", "cancelled"}

	if len(statuses) != len(expected) {
		t.Errorf("Expected %d statuses, got %d", len(expected), len(statuses))
	}

	for i, s := range expected {
		if statuses[i] != s {
			t.Errorf("Expected status %s at index %d, got %s", s, i, statuses[i])
		}
	}
}

func TestIsValidProjectStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"active", true},
		{"paused", true},
		{"completed", true},
		{"cancelled", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := IsValidProjectStatus(tt.status)
			if result != tt.expected {
				t.Errorf("IsValidProjectStatus(%q) = %v, expected %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestValidPhaseStatuses(t *testing.T) {
	statuses := ValidPhaseStatuses()
	expected := []string{"pending", "running", "completed", "failed", "skipped"}

	if len(statuses) != len(expected) {
		t.Errorf("Expected %d statuses, got %d", len(expected), len(statuses))
	}
}

func TestIsValidPhaseStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"pending", true},
		{"running", true},
		{"completed", true},
		{"failed", true},
		{"skipped", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := IsValidPhaseStatus(tt.status)
			if result != tt.expected {
				t.Errorf("IsValidPhaseStatus(%q) = %v, expected %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestDefaultPhases(t *testing.T) {
	phases := DefaultPhases()
	expected := []string{"planning", "research", "coding", "review"}

	if len(phases) != len(expected) {
		t.Errorf("Expected %d phases, got %d", len(expected), len(phases))
	}

	for i, p := range expected {
		if phases[i] != p {
			t.Errorf("Expected phase %s at index %d, got %s", p, i, phases[i])
		}
	}
}

func TestGetRoleForPhase(t *testing.T) {
	tests := []struct {
		phase    string
		expected string
	}{
		{"planning", "planner"},
		{"research", "researcher"},
		{"coding", "coder"},
		{"review", "reviewer"},
		{"testing", "tester"},
		{"deploy", "executor"},
		{"document", "documenter"},
		{"unknown", "specialist"},
	}

	for _, tt := range tests {
		t.Run(tt.phase, func(t *testing.T) {
			result := GetRoleForPhase(tt.phase)
			if result != tt.expected {
				t.Errorf("GetRoleForPhase(%q) = %q, expected %q", tt.phase, result, tt.expected)
			}
		})
	}
}

func TestGetPhaseForRole(t *testing.T) {
	tests := []struct {
		role     string
		expected string
	}{
		{"planner", "planning"},
		{"researcher", "research"},
		{"coder", "coding"},
		{"reviewer", "review"},
		{"tester", "testing"},
		{"executor", "deploy"},
		{"documenter", "document"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			result := GetPhaseForRole(tt.role)
			if result != tt.expected {
				t.Errorf("GetPhaseForRole(%q) = %q, expected %q", tt.role, result, tt.expected)
			}
		})
	}
}

func TestBuildPhases(t *testing.T) {
	phaseNames := []string{"planning", "research", "coding"}
	phases := BuildPhases(phaseNames)

	if len(phases) != len(phaseNames) {
		t.Fatalf("Expected %d phases, got %d", len(phaseNames), len(phases))
	}

	for i, phase := range phases {
		if phase.Name != phaseNames[i] {
			t.Errorf("Expected phase name %q at index %d, got %q", phaseNames[i], i, phase.Name)
		}
		if phase.Order != i {
			t.Errorf("Expected order %d at index %d, got %d", i, i, phase.Order)
		}
		if phase.Status != PhasePending {
			t.Errorf("Expected status %q at index %d, got %q", PhasePending, i, phase.Status)
		}
		expectedRole := GetRoleForPhase(phaseNames[i])
		if phase.Role != expectedRole {
			t.Errorf("Expected role %q at index %d, got %q", expectedRole, i, phase.Role)
		}
	}
}

func TestCalculateProgress(t *testing.T) {
	tests := []struct {
		name     string
		project  Project
		expected float64
	}{
		{
			name:     "empty project",
			project:  Project{Phases: []Phase{}},
			expected: 0.0,
		},
		{
			name: "all pending",
			project: Project{
				Phases: []Phase{
					{Name: "planning", Status: PhasePending},
					{Name: "coding", Status: PhasePending},
				},
			},
			expected: 0.0,
		},
		{
			name: "half completed",
			project: Project{
				Phases: []Phase{
					{Name: "planning", Status: PhaseCompleted},
					{Name: "coding", Status: PhasePending},
				},
			},
			expected: 50.0,
		},
		{
			name: "all completed",
			project: Project{
				Phases: []Phase{
					{Name: "planning", Status: PhaseCompleted},
					{Name: "coding", Status: PhaseCompleted},
				},
			},
			expected: 100.0,
		},
		{
			name: "with skipped",
			project: Project{
				Phases: []Phase{
					{Name: "planning", Status: PhaseCompleted},
					{Name: "research", Status: PhaseSkipped},
					{Name: "coding", Status: PhasePending},
				},
			},
			expected: 66.66666666666666,
		},
		{
			name: "with failed",
			project: Project{
				Phases: []Phase{
					{Name: "planning", Status: PhaseCompleted},
					{Name: "coding", Status: PhaseFailed},
				},
			},
			expected: 50.0, // failed counts as terminal but not complete
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateProgress(&tt.project)
			if result != tt.expected {
				t.Errorf("CalculateProgress() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetCurrentPhase(t *testing.T) {
	project := Project{
		CurrentPhase: "research",
		Phases: []Phase{
			{Name: "planning", Order: 0},
			{Name: "research", Order: 1},
			{Name: "coding", Order: 2},
		},
	}

	current := GetCurrentPhase(&project)
	if current == nil {
		t.Fatal("Expected current phase, got nil")
	}
	if current.Name != "research" {
		t.Errorf("Expected current phase 'research', got '%s'", current.Name)
	}
	if current.Order != 1 {
		t.Errorf("Expected order 1, got %d", current.Order)
	}
}

func TestGetCurrentPhase_NotFound(t *testing.T) {
	project := Project{
		CurrentPhase: "unknown",
		Phases: []Phase{
			{Name: "planning", Order: 0},
		},
	}

	current := GetCurrentPhase(&project)
	if current != nil {
		t.Errorf("Expected nil for unknown phase, got %v", current)
	}
}

func TestGetNextPhase(t *testing.T) {
	project := Project{
		CurrentPhase: "planning",
		CurrentOrder: 0,
		Phases: []Phase{
			{Name: "planning", Order: 0, Status: PhaseCompleted},
			{Name: "research", Order: 1, Status: PhasePending},
			{Name: "coding", Order: 2, Status: PhasePending},
		},
	}

	next := GetNextPhase(&project)
	if next == nil {
		t.Fatal("Expected next phase, got nil")
	}
	if next.Name != "research" {
		t.Errorf("Expected next phase 'research', got '%s'", next.Name)
	}
}

func TestGetNextPhase_NoNext(t *testing.T) {
	project := Project{
		CurrentPhase: "coding",
		CurrentOrder: 2,
		Phases: []Phase{
			{Name: "planning", Order: 0, Status: PhaseCompleted},
			{Name: "research", Order: 1, Status: PhaseCompleted},
			{Name: "coding", Order: 2, Status: PhaseRunning},
		},
	}

	next := GetNextPhase(&project)
	if next != nil {
		t.Errorf("Expected nil for no next phase, got %v", next)
	}
}

func TestIsPhaseComplete(t *testing.T) {
	tests := []struct {
		status   PhaseStatus
		expected bool
	}{
		{PhaseCompleted, true},
		{PhaseSkipped, true},
		{PhaseFailed, true},
		{PhasePending, false},
		{PhaseRunning, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := IsPhaseComplete(tt.status)
			if result != tt.expected {
				t.Errorf("IsPhaseComplete(%q) = %v, expected %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestCanAdvancePhase(t *testing.T) {
	tests := []struct {
		name     string
		project  Project
		expected bool
	}{
		{
			name: "can advance - current complete, has next",
			project: Project{
				Status:       ProjectActive,
				CurrentPhase: "planning",
				CurrentOrder: 0,
				Phases: []Phase{
					{Name: "planning", Order: 0, Status: PhaseCompleted},
					{Name: "research", Order: 1, Status: PhasePending},
				},
			},
			expected: true,
		},
		{
			name: "cannot advance - current not complete",
			project: Project{
				Status:       ProjectActive,
				CurrentPhase: "planning",
				CurrentOrder: 0,
				Phases: []Phase{
					{Name: "planning", Order: 0, Status: PhaseRunning},
					{Name: "research", Order: 1, Status: PhasePending},
				},
			},
			expected: false,
		},
		{
			name: "cannot advance - no next phase",
			project: Project{
				Status:       ProjectActive,
				CurrentPhase: "coding",
				CurrentOrder: 1,
				Phases: []Phase{
					{Name: "planning", Order: 0, Status: PhaseCompleted},
					{Name: "coding", Order: 1, Status: PhaseCompleted},
				},
			},
			expected: false,
		},
		{
			name: "cannot advance - project not active",
			project: Project{
				Status:       ProjectPaused,
				CurrentPhase: "planning",
				CurrentOrder: 0,
				Phases: []Phase{
					{Name: "planning", Order: 0, Status: PhaseCompleted},
					{Name: "research", Order: 1, Status: PhasePending},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanAdvancePhase(&tt.project)
			if result != tt.expected {
				t.Errorf("CanAdvancePhase() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Project Struct Tests
// =============================================================================

func TestProjectStruct(t *testing.T) {
	id := uuid.New().String()

	project := Project{
		ID:           id,
		Name:         "Test Project",
		Description:  "A test project",
		ConceptID:    "concept-123",
		CurrentPhase: "planning",
		CurrentOrder: 0,
		Phases: []Phase{
			{Name: "planning", Role: "planner", Status: PhaseRunning, Order: 0},
			{Name: "coding", Role: "coder", Status: PhasePending, Order: 1},
		},
		Status: ProjectActive,
		Metadata: map[string]interface{}{
			"priority": "high",
		},
	}

	if project.ID != id {
		t.Errorf("Expected ID %s, got %s", id, project.ID)
	}
	if project.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", project.Name)
	}
	if project.Description != "A test project" {
		t.Errorf("Expected description 'A test project', got '%s'", project.Description)
	}
	if project.ConceptID != "concept-123" {
		t.Errorf("Expected concept ID 'concept-123', got '%s'", project.ConceptID)
	}
	if project.CurrentPhase != "planning" {
		t.Errorf("Expected current phase 'planning', got '%s'", project.CurrentPhase)
	}
	if project.Status != ProjectActive {
		t.Errorf("Expected status 'active', got '%s'", project.Status)
	}
	if len(project.Phases) != 2 {
		t.Errorf("Expected 2 phases, got %d", len(project.Phases))
	}
}

func TestPhaseStruct(t *testing.T) {
	now := time.Now()
	phase := Phase{
		Name:        "planning",
		Role:        "planner",
		Status:      PhaseRunning,
		JobID:       "job-123",
		Order:       0,
		Description: "Planning phase",
		Result:      "Plan created",
		StartedAt:   &now,
	}

	if phase.Name != "planning" {
		t.Errorf("Expected name 'planning', got '%s'", phase.Name)
	}
	if phase.Role != "planner" {
		t.Errorf("Expected role 'planner', got '%s'", phase.Role)
	}
	if phase.Status != PhaseRunning {
		t.Errorf("Expected status 'running', got '%s'", phase.Status)
	}
	if phase.JobID != "job-123" {
		t.Errorf("Expected job ID 'job-123', got '%s'", phase.JobID)
	}
	if phase.Order != 0 {
		t.Errorf("Expected order 0, got %d", phase.Order)
	}
	if phase.StartedAt == nil {
		t.Error("Expected started at to be set")
	}
}

// =============================================================================
// CreateProjectRequest Tests
// =============================================================================

func TestCreateProjectRequest(t *testing.T) {
	req := CreateProjectRequest{
		Name:        "Test Project",
		Description: "A test project",
		ConceptID:   "concept-123",
		Phases:      []string{"planning", "coding"},
		Metadata: map[string]interface{}{
			"priority": "high",
		},
	}

	if req.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", req.Name)
	}
	if req.Description != "A test project" {
		t.Errorf("Expected description 'A test project', got '%s'", req.Description)
	}
	if req.ConceptID != "concept-123" {
		t.Errorf("Expected concept ID 'concept-123', got '%s'", req.ConceptID)
	}
	if len(req.Phases) != 2 {
		t.Errorf("Expected 2 phases, got %d", len(req.Phases))
	}
	if req.Metadata["priority"] != "high" {
		t.Errorf("Expected priority 'high', got '%v'", req.Metadata["priority"])
	}
}

// =============================================================================
// ProjectSummary Tests
// =============================================================================

func TestProjectSummary(t *testing.T) {
	now := time.Now()

	summary := ProjectSummary{
		ID:           "proj-123",
		Name:         "Test Project",
		CurrentPhase: "planning",
		Status:       "active",
		Progress:     25.5,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if summary.ID != "proj-123" {
		t.Errorf("Expected ID 'proj-123', got '%s'", summary.ID)
	}
	if summary.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", summary.Name)
	}
	if summary.CurrentPhase != "planning" {
		t.Errorf("Expected current phase 'planning', got '%s'", summary.CurrentPhase)
	}
	if summary.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", summary.Status)
	}
	if summary.Progress != 25.5 {
		t.Errorf("Expected progress 25.5, got %f", summary.Progress)
	}
}

// =============================================================================
// ProjectStatusDetail Tests
// =============================================================================

func TestProjectStatusDetail(t *testing.T) {
	detail := ProjectStatusDetail{
		Project: Project{
			ID:   "proj-123",
			Name: "Test Project",
		},
		CurrentPhase: &Phase{
			Name:   "planning",
			Status: PhaseRunning,
		},
		Progress:        50.0,
		CompletedPhases: 1,
		TotalPhases:     2,
		NextPhase: &Phase{
			Name:   "coding",
			Status: PhasePending,
		},
	}

	if detail.Project.ID != "proj-123" {
		t.Errorf("Expected project ID 'proj-123', got '%s'", detail.Project.ID)
	}
	if detail.CurrentPhase == nil || detail.CurrentPhase.Name != "planning" {
		t.Error("Expected current phase 'planning'")
	}
	if detail.Progress != 50.0 {
		t.Errorf("Expected progress 50.0, got %f", detail.Progress)
	}
	if detail.CompletedPhases != 1 {
		t.Errorf("Expected 1 completed phase, got %d", detail.CompletedPhases)
	}
	if detail.TotalPhases != 2 {
		t.Errorf("Expected 2 total phases, got %d", detail.TotalPhases)
	}
	if detail.NextPhase == nil || detail.NextPhase.Name != "coding" {
		t.Error("Expected next phase 'coding'")
	}
}

// =============================================================================
// GenerateProjectID Tests
// =============================================================================

func TestGenerateProjectID(t *testing.T) {
	id1 := GenerateProjectID()
	id2 := GenerateProjectID()

	if id1 == "" {
		t.Error("Expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("Expected unique IDs")
	}

	// Should be valid UUID format
	if _, err := uuid.Parse(id1); err != nil {
		t.Errorf("Expected valid UUID, got error: %v", err)
	}
}
