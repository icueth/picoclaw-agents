// Package office provides Office UI functionality for Picoclaw agent management
package office

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// CEOPriority represents the priority level for CEO directives
type CEOPriority int

const (
	CEOPriorityLow CEOPriority = iota
	CEOPriorityNormal
	CEOPriorityHigh
	CEOPriorityCritical
	CEOPriorityEmergency
)

func (p CEOPriority) String() string {
	switch p {
	case CEOPriorityLow:
		return "low"
	case CEOPriorityNormal:
		return "normal"
	case CEOPriorityHigh:
		return "high"
	case CEOPriorityCritical:
		return "critical"
	case CEOPriorityEmergency:
		return "emergency"
	default:
		return "unknown"
	}
}

// CEODirectiveType represents the type of CEO directive
type CEODirectiveType string

const (
	DirectiveTypeTask        CEODirectiveType = "task"
	DirectiveTypeMeeting     CEODirectiveType = "meeting"
	DirectiveTypeReport      CEODirectiveType = "report"
	DirectiveTypeReassign    CEODirectiveType = "reassign"
	DirectiveTypePriority    CEODirectiveType = "priority"
	DirectiveTypeBroadcast   CEODirectiveType = "broadcast"
	DirectiveTypeSkipMeeting CEODirectiveType = "skip_meeting"
)

// CEODirective represents a directive issued by the CEO (user with $ prefix)
type CEODirective struct {
	ID          string           `json:"id"`
	Type        CEODirectiveType `json:"type"`
	Priority    CEOPriority      `json:"priority"`
	Content     string           `json:"content"`
	TargetAgent string           `json:"target_agent,omitempty"`
	Department  string           `json:"department,omitempty"`
	IssuedAt    time.Time        `json:"issued_at"`
	IssuedBy    string           `json:"issued_by"`
	ExpiresAt   *time.Time       `json:"expires_at,omitempty"`
	SkipMeeting bool             `json:"skip_meeting"`
	Metadata    map[string]any   `json:"metadata,omitempty"`
}

// IsExpired checks if the directive has expired
func (d *CEODirective) IsExpired() bool {
	if d.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*d.ExpiresAt)
}

// IsHighPriority returns true if directive is high priority or above
func (d *CEODirective) IsHighPriority() bool {
	return d.Priority >= CEOPriorityHigh
}

// CEOCommand represents a parsed CEO command
type CEOCommand struct {
	Raw         string
	Type        CEODirectiveType
	Priority    CEOPriority
	Target      string
	Content     string
	Options     map[string]string
	SkipMeeting bool
}

// CEOManager handles CEO directives and command routing
type CEOManager struct {
	mu         sync.RWMutex
	directives map[string]*CEODirective
	handlers   map[CEODirectiveType]DirectiveHandler
	counter    int64

	// Configuration
	config CEOConfig
}

// CEOConfig contains configuration for CEO functionality
type CEOConfig struct {
	// RequireConfirmation requires confirmation for critical/emergency directives
	RequireConfirmation bool `json:"require_confirmation"`

	// AutoRouteHighPriority automatically routes high priority directives
	AutoRouteHighPriority bool `json:"auto_route_high_priority"`

	// SkipMeetingEnabled allows skipping meetings for urgent directives
	SkipMeetingEnabled bool `json:"skip_meeting_enabled"`

	// DefaultDirectiveExpiry is the default expiration time for directives
	DefaultDirectiveExpiry time.Duration `json:"default_directive_expiry"`

	// PriorityKeywords maps keywords to priority levels
	PriorityKeywords map[string]CEOPriority `json:"priority_keywords"`

	// DepartmentMapping maps departments to agent lists
	DepartmentMapping map[string][]string `json:"department_mapping"`
}

// DefaultCEOConfig returns default CEO configuration
func DefaultCEOConfig() CEOConfig {
	return CEOConfig{
		RequireConfirmation:    true,
		AutoRouteHighPriority:  true,
		SkipMeetingEnabled:     true,
		DefaultDirectiveExpiry: 24 * time.Hour,
		PriorityKeywords: map[string]CEOPriority{
			"urgent":    CEOPriorityHigh,
			"emergency": CEOPriorityEmergency,
			"critical":  CEOPriorityCritical,
			"asap":      CEOPriorityHigh,
			"low":       CEOPriorityLow,
		},
		DepartmentMapping: make(map[string][]string),
	}
}

// DirectiveHandler is a function that handles a specific directive type
type DirectiveHandler func(ctx context.Context, directive *CEODirective) (*DirectiveResult, error)

// DirectiveResult represents the result of executing a directive
type DirectiveResult struct {
	Success     bool           `json:"success"`
	Message     string         `json:"message"`
	DirectiveID string         `json:"directive_id"`
	Data        map[string]any `json:"data,omitempty"`
	Actions     []string       `json:"actions,omitempty"`
}

// NewCEOManager creates a new CEO manager
func NewCEOManager(config CEOConfig) *CEOManager {
	return &CEOManager{
		directives: make(map[string]*CEODirective),
		handlers:   make(map[CEODirectiveType]DirectiveHandler),
		config:     config,
	}
}

// RegisterHandler registers a handler for a directive type
func (m *CEOManager) RegisterHandler(directiveType CEODirectiveType, handler DirectiveHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[directiveType] = handler
}

// ParseCEOCommand parses a CEO command (prefixed with $)
func (m *CEOManager) ParseCEOCommand(input string) (*CEOCommand, error) {
	// Remove $ prefix and trim
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "$") {
		return nil, fmt.Errorf("not a CEO command: missing $ prefix")
	}

	content := strings.TrimSpace(input[1:])
	if content == "" {
		return nil, fmt.Errorf("empty CEO command")
	}

	cmd := &CEOCommand{
		Raw:     input,
		Options: make(map[string]string),
		Priority: CEOPriorityNormal,
	}

	// Parse priority indicators
	content = m.parsePriority(content, cmd)

	// Parse skip meeting flag
	content = m.parseSkipMeeting(content, cmd)

	// Parse command type and content
	m.parseCommandType(content, cmd)

	return cmd, nil
}

// parsePriority extracts priority from content
func (m *CEOManager) parsePriority(content string, cmd *CEOCommand) string {
	lower := strings.ToLower(content)

	for keyword, priority := range m.config.PriorityKeywords {
		pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(keyword) + `\b`)
		if pattern.MatchString(lower) {
			cmd.Priority = priority
			// Remove keyword from content
			content = pattern.ReplaceAllString(content, "")
			break
		}
	}

	return strings.TrimSpace(content)
}

// parseSkipMeeting extracts skip meeting flag
func (m *CEOManager) parseSkipMeeting(content string, cmd *CEOCommand) string {
	patterns := []string{
		`(?i)\bskip\s+meeting\b`,
		`(?i)\bno\s+meeting\b`,
		`(?i)\b bypass\s+meeting\b`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(content) {
			cmd.SkipMeeting = true
			content = re.ReplaceAllString(content, "")
		}
	}

	return strings.TrimSpace(content)
}

// parseCommandType determines the command type and extracts target/content
func (m *CEOManager) parseCommandType(content string, cmd *CEOCommand) {
	lower := strings.ToLower(content)

	// Check for directive type keywords
	switch {
	case strings.HasPrefix(lower, "meeting"):
		cmd.Type = DirectiveTypeMeeting
		cmd.Content = strings.TrimSpace(content[7:])

	case strings.HasPrefix(lower, "report"):
		cmd.Type = DirectiveTypeReport
		cmd.Content = strings.TrimSpace(content[6:])

	case strings.HasPrefix(lower, "reassign"):
		cmd.Type = DirectiveTypeReassign
		cmd.Content = strings.TrimSpace(content[8:])

	case strings.HasPrefix(lower, "broadcast"):
		cmd.Type = DirectiveTypeBroadcast
		cmd.Content = strings.TrimSpace(content[9:])

	case strings.HasPrefix(lower, "priority"):
		cmd.Type = DirectiveTypePriority
		cmd.Content = strings.TrimSpace(content[8:])

	case strings.HasPrefix(lower, "skip meeting"):
		cmd.Type = DirectiveTypeSkipMeeting
		cmd.Content = strings.TrimSpace(content[12:])
		cmd.SkipMeeting = true

	default:
		// Default to task directive
		cmd.Type = DirectiveTypeTask
		cmd.Content = content
	}

	// Extract target agent (@agentname)
	targetPattern := regexp.MustCompile(`@(\w+)`)
	if matches := targetPattern.FindStringSubmatch(cmd.Content); len(matches) > 1 {
		cmd.Target = matches[1]
		cmd.Content = targetPattern.ReplaceAllString(cmd.Content, "")
		cmd.Content = strings.TrimSpace(cmd.Content)
	}

	// Extract department (#department)
	deptPattern := regexp.MustCompile(`#(\w+)`)
	if matches := deptPattern.FindStringSubmatch(content); len(matches) > 1 {
		cmd.Options["department"] = matches[1]
	}
}

// CreateDirective creates a CEODirective from a parsed command
func (m *CEOManager) CreateDirective(cmd *CEOCommand, issuedBy string) *CEODirective {
	m.mu.Lock()
	m.counter++
	id := fmt.Sprintf("ceo-%d-%d", time.Now().Unix(), m.counter)
	m.mu.Unlock()

	directive := &CEODirective{
		ID:          id,
		Type:        cmd.Type,
		Priority:    cmd.Priority,
		Content:     cmd.Content,
		TargetAgent: cmd.Target,
		IssuedAt:    time.Now(),
		IssuedBy:    issuedBy,
		SkipMeeting: cmd.SkipMeeting,
		Metadata:    make(map[string]any),
	}

	if dept, ok := cmd.Options["department"]; ok {
		directive.Department = dept
	}

	// Set expiration
	if m.config.DefaultDirectiveExpiry > 0 {
		expiry := time.Now().Add(m.config.DefaultDirectiveExpiry)
		directive.ExpiresAt = &expiry
	}

	return directive
}

// ExecuteDirective executes a directive using the registered handler
func (m *CEOManager) ExecuteDirective(ctx context.Context, directive *CEODirective) (*DirectiveResult, error) {
	m.mu.RLock()
	handler, exists := m.handlers[directive.Type]
	m.mu.RUnlock()

	if !exists {
		return &DirectiveResult{
			Success:     false,
			Message:     fmt.Sprintf("no handler registered for directive type: %s", directive.Type),
			DirectiveID: directive.ID,
		}, fmt.Errorf("no handler for directive type: %s", directive.Type)
	}

	// Store directive
	m.mu.Lock()
	m.directives[directive.ID] = directive
	m.mu.Unlock()

	// Execute handler
	result, err := handler(ctx, directive)
	if err != nil {
		return &DirectiveResult{
			Success:     false,
			Message:     err.Error(),
			DirectiveID: directive.ID,
		}, err
	}

	return result, nil
}

// GetDirective retrieves a directive by ID
func (m *CEOManager) GetDirective(id string) (*CEODirective, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	directive, exists := m.directives[id]
	return directive, exists
}

// GetActiveDirectives returns all non-expired directives
func (m *CEOManager) GetActiveDirectives() []*CEODirective {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []*CEODirective
	for _, d := range m.directives {
		if !d.IsExpired() {
			active = append(active, d)
		}
	}

	return active
}

// GetDirectivesByPriority returns directives filtered by minimum priority
func (m *CEOManager) GetDirectivesByPriority(minPriority CEOPriority) []*CEODirective {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []*CEODirective
	for _, d := range m.directives {
		if !d.IsExpired() && d.Priority >= minPriority {
			filtered = append(filtered, d)
		}
	}

	return filtered
}

// GetDirectivesForAgent returns directives targeting a specific agent
func (m *CEOManager) GetDirectivesForAgent(agentID string) []*CEODirective {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var agentDirectives []*CEODirective
	for _, d := range m.directives {
		if !d.IsExpired() && (d.TargetAgent == agentID || d.TargetAgent == "") {
			agentDirectives = append(agentDirectives, d)
		}
	}

	return agentDirectives
}

// CancelDirective cancels a directive by ID
func (m *CEOManager) CancelDirective(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if directive, exists := m.directives[id]; exists {
		now := time.Now()
		directive.ExpiresAt = &now
		return true
	}
	return false
}

// CleanupExpired removes expired directives
func (m *CEOManager) CleanupExpired() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for id, d := range m.directives {
		if d.IsExpired() {
			delete(m.directives, id)
			count++
		}
	}

	return count
}

// RoutePriorityTask routes a high-priority task to appropriate agents
func (m *CEOManager) RoutePriorityTask(directive *CEODirective) ([]string, error) {
	if !m.config.AutoRouteHighPriority {
		return nil, fmt.Errorf("auto-routing is disabled")
	}

	if directive.Priority < CEOPriorityHigh {
		return nil, fmt.Errorf("directive is not high priority")
	}

	// Get target agents
	var targets []string

	if directive.TargetAgent != "" {
		// Specific agent targeted
		targets = []string{directive.TargetAgent}
	} else if directive.Department != "" {
		// Department targeted
		if agents, ok := m.config.DepartmentMapping[directive.Department]; ok {
			targets = agents
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no target agents found for directive")
	}

	return targets, nil
}

// ShouldSkipMeeting determines if a meeting should be skipped for a directive
func (m *CEOManager) ShouldSkipMeeting(directive *CEODirective) bool {
	if !m.config.SkipMeetingEnabled {
		return false
	}

	// Skip meeting if explicitly requested
	if directive.SkipMeeting {
		return true
	}

	// Skip meeting for emergency/critical priority
	if directive.Priority >= CEOPriorityCritical {
		return true
	}

	return false
}

// GetStats returns statistics about directives
func (m *CEOManager) GetStats() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]any{
		"total_directives":   len(m.directives),
		"active_directives":  0,
		"expired_directives": 0,
		"by_type":            make(map[string]int),
		"by_priority":        make(map[string]int),
	}

	byType := stats["by_type"].(map[string]int)
	byPriority := stats["by_priority"].(map[string]int)

	for _, d := range m.directives {
		if d.IsExpired() {
			stats["expired_directives"] = stats["expired_directives"].(int) + 1
		} else {
			stats["active_directives"] = stats["active_directives"].(int) + 1
		}

		byType[string(d.Type)]++
		byPriority[d.Priority.String()]++
	}

	return stats
}
