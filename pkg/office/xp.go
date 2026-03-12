// Package office provides Office UI functionality for Picoclaw agent management
package office

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// TaskType represents the type of task for XP calculation
type TaskType string

const (
	TaskTypeCodeReview    TaskType = "code_review"
	TaskTypeBugFix        TaskType = "bug_fix"
	TaskTypeFeature       TaskType = "feature"
	TaskTypeDocumentation TaskType = "documentation"
	TaskTypeTesting       TaskType = "testing"
	TaskTypeResearch      TaskType = "research"
	TaskTypeMeeting       TaskType = "meeting"
	TaskTypePlanning      TaskType = "planning"
	TaskTypeSupport       TaskType = "support"
	TaskTypeRefactoring   TaskType = "refactoring"
	TaskTypeDeployment    TaskType = "deployment"
	TaskTypeAnalysis      TaskType = "analysis"
)

// TaskComplexity represents the complexity level of a task
type TaskComplexity int

const (
	ComplexityTrivial TaskComplexity = iota
	ComplexitySimple
	ComplexityModerate
	ComplexityComplex
	ComplexityVeryComplex
	ComplexityEpic
)

func (c TaskComplexity) String() string {
	switch c {
	case ComplexityTrivial:
		return "trivial"
	case ComplexitySimple:
		return "simple"
	case ComplexityModerate:
		return "moderate"
	case ComplexityComplex:
		return "complex"
	case ComplexityVeryComplex:
		return "very_complex"
	case ComplexityEpic:
		return "epic"
	default:
		return "unknown"
	}
}

// XPBaseValues contains base XP values for different task types and complexities
type XPBaseValues struct {
	// TypeMultipliers are multipliers for different task types
	TypeMultipliers map[TaskType]float64

	// ComplexityBaseXP is the base XP for each complexity level
	ComplexityBaseXP map[TaskComplexity]int

	// QualityBonus is the bonus for high quality work
	QualityBonus map[int]float64 // quality score (1-5) -> multiplier
}

// DefaultXPBaseValues returns default XP calculation values
func DefaultXPBaseValues() XPBaseValues {
	return XPBaseValues{
		TypeMultipliers: map[TaskType]float64{
			TaskTypeCodeReview:    1.2,
			TaskTypeBugFix:        1.5,
			TaskTypeFeature:       1.8,
			TaskTypeDocumentation: 0.8,
			TaskTypeTesting:       1.0,
			TaskTypeResearch:      1.3,
			TaskTypeMeeting:       0.5,
			TaskTypePlanning:      1.1,
			TaskTypeSupport:       0.9,
			TaskTypeRefactoring:   1.4,
			TaskTypeDeployment:    1.2,
			TaskTypeAnalysis:      1.3,
		},
		ComplexityBaseXP: map[TaskComplexity]int{
			ComplexityTrivial:     5,
			ComplexitySimple:      15,
			ComplexityModerate:    30,
			ComplexityComplex:     60,
			ComplexityVeryComplex: 100,
			ComplexityEpic:        200,
		},
		QualityBonus: map[int]float64{
			1: 0.0,  // Poor
			2: 0.1,  // Below average
			3: 0.25, // Average
			4: 0.5,  // Good
			5: 1.0,  // Excellent
		},
	}
}

// XPRecord represents an XP gain record
type XPRecord struct {
	ID          string         `json:"id"`
	AgentID     string         `json:"agent_id"`
	TaskType    TaskType       `json:"task_type"`
	Complexity  TaskComplexity `json:"complexity"`
	BaseXP      int            `json:"base_xp"`
	BonusXP     int            `json:"bonus_xp"`
	TotalXP     int            `json:"total_xp"`
	Quality     int            `json:"quality"` // 1-5
	Description string         `json:"description"`
	Timestamp   time.Time      `json:"timestamp"`
	ProjectID   string         `json:"project_id,omitempty"`
	TaskID      string         `json:"task_id,omitempty"`
}

// LevelDefinition defines a level in the progression system
type LevelDefinition struct {
	Level        int    `json:"level"`
	Title        string `json:"title"`
	MinXP        int    `json:"min_xp"`
	MaxXP        int    `json:"max_xp"`
	XPToNext     int    `json:"xp_to_next"`
	BonusPerks   []string `json:"bonus_perks,omitempty"`
}

// XPManager manages XP, levels, and progression for agents
type XPManager struct {
	mu        sync.RWMutex
	agents    map[string]*AgentXP
	records   []*XPRecord
	levels    []LevelDefinition
	baseValues XPBaseValues
	config    XPConfig

	// Award handlers
	awardHandlers []AwardHandler
}

// AgentXP represents an agent's XP and level information
type AgentXP struct {
	AgentID          string         `json:"agent_id"`
	CurrentXP        int            `json:"current_xp"`
	TotalXPEarned    int            `json:"total_xp_earned"`
	Level            int            `json:"level"`
	Title            string         `json:"title"`
	XPToNextLevel    int            `json:"xp_to_next_level"`
	ProgressPercent  float64        `json:"progress_percent"`
	TasksCompleted   int            `json:"tasks_completed"`
	StreakDays       int            `json:"streak_days"`
	LastActive       time.Time      `json:"last_active"`
	Achievements     []string       `json:"achievements"`
	BonusesEarned    int            `json:"bonuses_earned"`
}

// XPConfig contains configuration for XP system
type XPConfig struct {
	// EnableStreaks enables daily streak bonuses
	EnableStreaks bool `json:"enable_streaks"`

	// StreakBonusXP is the bonus XP for maintaining a streak
	StreakBonusXP int `json:"streak_bonus_xp"`

	// MaxStreakBonus is the maximum streak bonus multiplier
	MaxStreakBonus float64 `json:"max_streak_bonus"`

	// FirstTaskBonus is XP awarded for first task of the day
	FirstTaskBonus int `json:"first_task_bonus"`

	// TeamworkBonus is the multiplier for collaborative tasks
	TeamworkBonus float64 `json:"teamwork_bonus"`

	// WeeklyBonusThreshold is the number of tasks for weekly bonus
	WeeklyBonusThreshold int `json:"weekly_bonus_threshold"`

	// WeeklyBonusXP is the bonus for meeting weekly threshold
	WeeklyBonusXP int `json:"weekly_bonus_xp"`
}

// DefaultXPConfig returns default XP configuration
func DefaultXPConfig() XPConfig {
	return XPConfig{
		EnableStreaks:        true,
		StreakBonusXP:        5,
		MaxStreakBonus:       2.0,
		FirstTaskBonus:       10,
		TeamworkBonus:        1.25,
		WeeklyBonusThreshold: 20,
		WeeklyBonusXP:        100,
	}
}

// AwardHandler is called when an award is granted
type AwardHandler func(agentID string, award *XPAward)

// XPAward represents an XP award or bonus
type XPAward struct {
	ID          string    `json:"id"`
	AgentID     string    `json:"agent_id"`
	Type        string    `json:"type"`
	XPAmount    int       `json:"xp_amount"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// AwardType represents different types of awards
type AwardType string

const (
	AwardTypeDailyStreak    AwardType = "daily_streak"
	AwardTypeFirstTask      AwardType = "first_task"
	AwardTypeTeamwork       AwardType = "teamwork"
	AwardTypeWeeklyBonus    AwardType = "weekly_bonus"
	AwardTypeMilestone      AwardType = "milestone"
	AwardTypePerfectScore   AwardType = "perfect_score"
	AwardTypeSpeedBonus     AwardType = "speed_bonus"
	AwardTypeConsistency    AwardType = "consistency"
)

// NewXPManager creates a new XP manager with default levels
func NewXPManager(config XPConfig) *XPManager {
	return &XPManager{
		agents:        make(map[string]*AgentXP),
		records:       make([]*XPRecord, 0),
		levels:        generateDefaultLevels(),
		baseValues:    DefaultXPBaseValues(),
		config:        config,
		awardHandlers: make([]AwardHandler, 0),
	}
}

// generateDefaultLevels creates the default level progression
func generateDefaultLevels() []LevelDefinition {
	levels := []LevelDefinition{
		{Level: 1, Title: "Novice", MinXP: 0, MaxXP: 99, XPToNext: 100},
		{Level: 2, Title: "Apprentice", MinXP: 100, MaxXP: 249, XPToNext: 150},
		{Level: 3, Title: "Junior", MinXP: 250, MaxXP: 499, XPToNext: 250},
		{Level: 4, Title: "Associate", MinXP: 500, MaxXP: 999, XPToNext: 500},
		{Level: 5, Title: "Professional", MinXP: 1000, MaxXP: 1999, XPToNext: 1000},
		{Level: 6, Title: "Senior", MinXP: 2000, MaxXP: 3999, XPToNext: 2000},
		{Level: 7, Title: "Lead", MinXP: 4000, MaxXP: 6999, XPToNext: 3000},
		{Level: 8, Title: "Expert", MinXP: 7000, MaxXP: 11999, XPToNext: 5000},
		{Level: 9, Title: "Master", MinXP: 12000, MaxXP: 19999, XPToNext: 8000},
		{Level: 10, Title: "Grandmaster", MinXP: 20000, MaxXP: 29999, XPToNext: 10000},
		{Level: 11, Title: "Legend", MinXP: 30000, MaxXP: 49999, XPToNext: 20000},
		{Level: 12, Title: "Mythic", MinXP: 50000, MaxXP: 99999, XPToNext: 50000},
		{Level: 13, Title: "Immortal", MinXP: 100000, MaxXP: 199999, XPToNext: 100000},
		{Level: 14, Title: "Transcendent", MinXP: 200000, MaxXP: 499999, XPToNext: 300000},
		{Level: 15, Title: "Omniscient", MinXP: 500000, MaxXP: math.MaxInt32, XPToNext: 0},
	}

	// Add perks for higher levels
	for i := range levels {
		switch levels[i].Level {
		case 5:
			levels[i].BonusPerks = []string{"priority_support", "custom_avatar"}
		case 10:
			levels[i].BonusPerks = []string{"team_lead", "mentor_status", "special_title_color"}
		case 15:
			levels[i].BonusPerks = []string{"hall_of_fame", "legacy_status", "all_perks"}
		}
	}

	return levels
}

// GetOrCreateAgentXP gets or creates XP record for an agent
func (m *XPManager) GetOrCreateAgentXP(agentID string) *AgentXP {
	m.mu.Lock()
	defer m.mu.Unlock()

	if agent, exists := m.agents[agentID]; exists {
		return agent
	}

	agent := &AgentXP{
		AgentID:       agentID,
		CurrentXP:     0,
		TotalXPEarned: 0,
		Level:         1,
		Title:         m.levels[0].Title,
		XPToNextLevel: m.levels[0].XPToNext,
		Achievements:  make([]string, 0),
		LastActive:    time.Now(),
	}

	m.agents[agentID] = agent
	return agent
}

// CalculateXP calculates XP for a task based on type and complexity
func (m *XPManager) CalculateXP(taskType TaskType, complexity TaskComplexity, quality int) int {
	// Get base XP for complexity
	baseXP, ok := m.baseValues.ComplexityBaseXP[complexity]
	if !ok {
		baseXP = m.baseValues.ComplexityBaseXP[ComplexitySimple]
	}

	// Apply type multiplier
	multiplier := 1.0
	if typeMult, ok := m.baseValues.TypeMultipliers[taskType]; ok {
		multiplier = typeMult
	}

	// Calculate quality bonus
	qualityMult := 0.0
	if qm, ok := m.baseValues.QualityBonus[quality]; ok {
		qualityMult = qm
	}

	// Calculate total XP
	totalXP := float64(baseXP) * multiplier
	bonusXP := totalXP * qualityMult

	return int(totalXP + bonusXP)
}

// AwardXP awards XP to an agent for completing a task
func (m *XPManager) AwardXP(agentID string, taskType TaskType, complexity TaskComplexity, quality int, description string) (*XPRecord, error) {
	if quality < 1 || quality > 5 {
		return nil, fmt.Errorf("quality must be between 1 and 5")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	agent := m.GetOrCreateAgentXP(agentID)

	// Calculate XP
	totalXP := m.CalculateXP(taskType, complexity, quality)
	baseXP := m.baseValues.ComplexityBaseXP[complexity]
	bonusXP := totalXP - baseXP

	// Create record
	record := &XPRecord{
		ID:          fmt.Sprintf("xp-%d-%s", time.Now().UnixNano(), agentID),
		AgentID:     agentID,
		TaskType:    taskType,
		Complexity:  complexity,
		BaseXP:      baseXP,
		BonusXP:     bonusXP,
		TotalXP:     totalXP,
		Quality:     quality,
		Description: description,
		Timestamp:   time.Now(),
	}

	m.records = append(m.records, record)

	// Update agent XP
	agent.CurrentXP += totalXP
	agent.TotalXPEarned += totalXP
	agent.TasksCompleted++
	agent.LastActive = time.Now()

	// Check for level up
	m.checkLevelUp(agent)

	return record, nil
}

// checkLevelUp checks and handles level progression
func (m *XPManager) checkLevelUp(agent *AgentXP) {
	for _, level := range m.levels {
		if agent.CurrentXP >= level.MinXP && agent.CurrentXP <= level.MaxXP && agent.Level < level.Level {
			// Level up!
			oldLevel := agent.Level
			agent.Level = level.Level
			agent.Title = level.Title
			agent.XPToNextLevel = level.XPToNext

			// Calculate progress to next level
			if level.XPToNext > 0 {
				progress := float64(agent.CurrentXP-level.MinXP) / float64(level.XPToNext)
				agent.ProgressPercent = math.Min(progress*100, 100)
			} else {
				agent.ProgressPercent = 100
			}

			// Trigger level up award
			m.triggerAward(agent.AgentID, &XPAward{
				ID:       fmt.Sprintf("award-%d", time.Now().UnixNano()),
				AgentID:  agent.AgentID,
				Type:     string(AwardTypeMilestone),
				XPAmount: 0,
				Reason:   fmt.Sprintf("Leveled up from %d to %d", oldLevel, agent.Level),
				Timestamp: time.Now(),
				Metadata: map[string]any{
					"old_level": oldLevel,
					"new_level": agent.Level,
					"new_title": agent.Title,
				},
			})

			break
		}
	}

	// Update progress for current level
	for _, level := range m.levels {
		if agent.Level == level.Level && level.XPToNext > 0 {
			progress := float64(agent.CurrentXP-level.MinXP) / float64(level.XPToNext)
			agent.ProgressPercent = math.Min(progress*100, 100)
			break
		}
	}
}

// AwardBonus awards a bonus to an agent
func (m *XPManager) AwardBonus(agentID string, awardType AwardType, xpAmount int, reason string, metadata map[string]any) *XPAward {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent := m.GetOrCreateAgentXP(agentID)

	award := &XPAward{
		ID:        fmt.Sprintf("bonus-%d", time.Now().UnixNano()),
		AgentID:   agentID,
		Type:      string(awardType),
		XPAmount:  xpAmount,
		Reason:    reason,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	if xpAmount > 0 {
		agent.CurrentXP += xpAmount
		agent.TotalXPEarned += xpAmount
		agent.BonusesEarned++
		m.checkLevelUp(agent)
	}

	m.triggerAward(agentID, award)

	return award
}

// triggerAward calls all registered award handlers
func (m *XPManager) triggerAward(agentID string, award *XPAward) {
	for _, handler := range m.awardHandlers {
		go handler(agentID, award)
	}
}

// RegisterAwardHandler registers a handler for award events
func (m *XPManager) RegisterAwardHandler(handler AwardHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.awardHandlers = append(m.awardHandlers, handler)
}

// GetAgentXP gets an agent's XP information
func (m *XPManager) GetAgentXP(agentID string) (*AgentXP, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, exists := m.agents[agentID]
	return agent, exists
}

// GetAgentRecords gets all XP records for an agent
func (m *XPManager) GetAgentRecords(agentID string) []*XPRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var records []*XPRecord
	for _, r := range m.records {
		if r.AgentID == agentID {
			records = append(records, r)
		}
	}

	return records
}

// GetLeaderboard returns the top agents by XP
func (m *XPManager) GetLeaderboard(limit int) []*AgentXP {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Convert map to slice
	agents := make([]*AgentXP, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}

	// Sort by total XP (bubble sort for simplicity)
	for i := 0; i < len(agents); i++ {
		for j := i + 1; j < len(agents); j++ {
			if agents[j].TotalXPEarned > agents[i].TotalXPEarned {
				agents[i], agents[j] = agents[j], agents[i]
			}
		}
	}

	// Return top N
	if limit > 0 && limit < len(agents) {
		return agents[:limit]
	}
	return agents
}

// GetLevelDefinition gets the definition for a specific level
func (m *XPManager) GetLevelDefinition(level int) (*LevelDefinition, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, l := range m.levels {
		if l.Level == level {
			return &l, true
		}
	}
	return nil, false
}

// GetAllLevels returns all level definitions
func (m *XPManager) GetAllLevels() []LevelDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]LevelDefinition, len(m.levels))
	copy(result, m.levels)
	return result
}

// UpdateStreak updates an agent's daily streak
func (m *XPManager) UpdateStreak(agentID string) (*XPAward, error) {
	if !m.config.EnableStreaks {
		return nil, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	agent := m.GetOrCreateAgentXP(agentID)

	now := time.Now()
	lastActive := agent.LastActive.Truncate(24 * time.Hour)
	today := now.Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)

	if lastActive.Equal(yesterday) {
		// Continued streak
		agent.StreakDays++

		// Calculate streak bonus
		streakMultiplier := math.Min(float64(agent.StreakDays)*0.1+1.0, m.config.MaxStreakBonus)
		bonusXP := int(float64(m.config.StreakBonusXP) * streakMultiplier)

		return m.AwardBonus(agentID, AwardTypeDailyStreak, bonusXP,
			fmt.Sprintf("%d day streak!", agent.StreakDays),
			map[string]any{"streak_days": agent.StreakDays}), nil

	} else if lastActive.Before(yesterday) {
		// Streak broken
		agent.StreakDays = 1
		return nil, nil
	}

	// Same day, no streak update
	return nil, nil
}

// AwardFirstTaskBonus awards bonus for first task of the day
func (m *XPManager) AwardFirstTaskBonus(agentID string) *XPAward {
	return m.AwardBonus(agentID, AwardTypeFirstTask, m.config.FirstTaskBonus,
		"First task of the day!", nil)
}

// AwardTeamworkBonus awards bonus for collaborative work
func (m *XPManager) AwardTeamworkBonus(agentID string, collaborators int) *XPAward {
	bonus := int(float64(10*collaborators) * m.config.TeamworkBonus)
	return m.AwardBonus(agentID, AwardTypeTeamwork, bonus,
		fmt.Sprintf("Teamwork with %d collaborators", collaborators),
		map[string]any{"collaborators": collaborators})
}

// GetStats returns global XP statistics
func (m *XPManager) GetStats() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalXP := 0
	totalTasks := 0
	levelDistribution := make(map[int]int)

	for _, agent := range m.agents {
		totalXP += agent.TotalXPEarned
		totalTasks += agent.TasksCompleted
		levelDistribution[agent.Level]++
	}

	return map[string]any{
		"total_agents":       len(m.agents),
		"total_xp_earned":    totalXP,
		"total_tasks":        totalTasks,
		"total_records":      len(m.records),
		"level_distribution": levelDistribution,
		"avg_xp_per_agent":   float64(totalXP) / float64(len(m.agents)),
	}
}

// ResetAgentXP resets an agent's XP (for testing or penalties)
func (m *XPManager) ResetAgentXP(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.agents, agentID)
}

// GetXPForNextLevel returns the XP needed for the next level
func (m *XPManager) GetXPForNextLevel(agentID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, exists := m.agents[agentID]
	if !exists {
		return 0, fmt.Errorf("agent not found: %s", agentID)
	}

	return agent.XPToNextLevel, nil
}
