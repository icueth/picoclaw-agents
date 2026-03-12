// Package office provides Office UI functionality for Picoclaw agent management
package office

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// ReportType represents the type of report
type ReportType string

const (
	ReportTypeAgentPerformance   ReportType = "agent_performance"
	ReportTypeDepartmentStats    ReportType = "department_stats"
	ReportTypeProjectSummary     ReportType = "project_summary"
	ReportTypeMeetingAnalysis    ReportType = "meeting_analysis"
	ReportTypeXPDistribution     ReportType = "xp_distribution"
	ReportTypeTaskCompletion     ReportType = "task_completion"
	ReportTypeSystemHealth       ReportType = "system_health"
	ReportTypeCustom             ReportType = "custom"
)

// ReportFormat represents the output format for reports
type ReportFormat string

const (
	ReportFormatJSON    ReportFormat = "json"
	ReportFormatMarkdown ReportFormat = "markdown"
	ReportFormatHTML    ReportFormat = "html"
	ReportFormatCSV     ReportFormat = "csv"
)

// Report represents a generated report
type Report struct {
	ID          string         `json:"id"`
	Type        ReportType     `json:"type"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Format      ReportFormat   `json:"format"`
	GeneratedAt time.Time      `json:"generated_at"`
	GeneratedBy string         `json:"generated_by"`
	PeriodStart time.Time      `json:"period_start"`
	PeriodEnd   time.Time      `json:"period_end"`
	Data        map[string]any `json:"data"`
	Summary     ReportSummary  `json:"summary"`
	Content     string         `json:"content,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
}

// ReportSummary contains summary statistics for a report
type ReportSummary struct {
	TotalAgents      int     `json:"total_agents"`
	TotalTasks       int     `json:"total_tasks"`
	TotalMeetings    int     `json:"total_meetings"`
	TotalXP          int     `json:"total_xp"`
	CompletionRate   float64 `json:"completion_rate"`
	AverageQuality   float64 `json:"average_quality"`
	ActiveProjects   int     `json:"active_projects"`
	PendingDirectives int    `json:"pending_directives"`
}

// AgentPerformanceMetrics contains performance metrics for an agent
type AgentPerformanceMetrics struct {
	AgentID           string         `json:"agent_id"`
	AgentName         string         `json:"agent_name,omitempty"`
	Department        string         `json:"department,omitempty"`
	Level             int            `json:"level"`
	Title             string         `json:"title"`

	// Task metrics
	TasksAssigned     int            `json:"tasks_assigned"`
	TasksCompleted    int            `json:"tasks_completed"`
	TasksInProgress   int            `json:"tasks_in_progress"`
	TasksOverdue      int            `json:"tasks_overdue"`
	CompletionRate    float64        `json:"completion_rate"`

	// Quality metrics
	AverageQuality    float64        `json:"average_quality"`
	HighQualityTasks  int            `json:"high_quality_tasks"` // Quality >= 4
	LowQualityTasks   int            `json:"low_quality_tasks"`  // Quality <= 2

	// XP metrics
	XPEarned          int            `json:"xp_earned"`
	XPGrowth          float64        `json:"xp_growth"` // Percentage growth
	CurrentStreak     int            `json:"current_streak"`
	MaxStreak         int            `json:"max_streak"`

	// Time metrics
	TotalTimeSpent    time.Duration  `json:"total_time_spent"`
	AverageTaskTime   time.Duration  `json:"average_task_time"`
	ResponseTime      time.Duration  `json:"response_time"` // Time to first response

	// Meeting metrics
	MeetingsAttended  int            `json:"meetings_attended"`
	MeetingsOrganized int            `json:"meetings_organized"`
	AttendanceRate    float64        `json:"attendance_rate"`

	// Collaboration metrics
	TasksCollaborated int            `json:"tasks_collaborated"`
	PeersWorkedWith   []string       `json:"peers_worked_with,omitempty"`

	// Efficiency score (0-100)
	EfficiencyScore   float64        `json:"efficiency_score"`

	// Trends (compared to previous period)
	Trends            MetricTrends   `json:"trends"`
}

// MetricTrends shows how metrics changed compared to previous period
type MetricTrends struct {
	TasksCompletedChange   float64 `json:"tasks_completed_change"`   // Percentage change
	QualityChange          float64 `json:"quality_change"`
	XPEarnedChange         float64 `json:"xp_earned_change"`
	EfficiencyScoreChange  float64 `json:"efficiency_score_change"`
}

// DepartmentStatistics contains statistics for a department
type DepartmentStatistics struct {
	Department        string                  `json:"department"`
	AgentCount        int                     `json:"agent_count"`
	Agents            []string                `json:"agents,omitempty"`

	// Aggregated metrics
	TotalTasks        int                     `json:"total_tasks"`
	CompletedTasks    int                     `json:"completed_tasks"`
	CompletionRate    float64                 `json:"completion_rate"`
	TotalXP           int                     `json:"total_xp"`
	AverageLevel      float64                 `json:"average_level"`

	// Workload distribution
	WorkloadDistribution map[string]int       `json:"workload_distribution,omitempty"`

	// Top performers
	TopPerformers     []string                `json:"top_performers,omitempty"`

	// Department efficiency
	EfficiencyScore   float64                 `json:"efficiency_score"`

	// Meetings
	MeetingsHeld      int                     `json:"meetings_held"`
	AverageAttendance float64                 `json:"average_attendance"`
}

// ProjectStatistics contains statistics for a project
type ProjectStatistics struct {
	ProjectID         string         `json:"project_id"`
	ProjectName       string         `json:"project_name"`
	Status            string         `json:"status"`

	// Timeline
	StartDate         *time.Time     `json:"start_date,omitempty"`
	EndDate           *time.Time     `json:"end_date,omitempty"`
	Duration          time.Duration  `json:"duration"`

	// Task statistics
	TotalTasks        int            `json:"total_tasks"`
	CompletedTasks    int            `json:"completed_tasks"`
	InProgressTasks   int            `json:"in_progress_tasks"`
	OverdueTasks      int            `json:"overdue_tasks"`
	CompletionRate    float64        `json:"completion_rate"`

	// Agent statistics
	AssignedAgents    []string       `json:"assigned_agents,omitempty"`
	AgentCount        int            `json:"agent_count"`

	// XP earned on project
	TotalXP           int            `json:"total_xp"`

	// Milestones
	Milestones        []Milestone    `json:"milestones,omitempty"`
	CompletedMilestones int          `json:"completed_milestones"`
}

// Milestone represents a project milestone
type Milestone struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Status      string     `json:"status"`
}

// MeetingStatistics contains statistics about meetings
type MeetingStatistics struct {
	TotalMeetings     int            `json:"total_meetings"`
	CompletedMeetings int            `json:"completed_meetings"`
	CancelledMeetings int            `json:"cancelled_meetings"`
	SkippedMeetings   int            `json:"skipped_meetings"`

	// Type breakdown
	MeetingsByType    map[string]int `json:"meetings_by_type"`

	// Participation
	TotalParticipants int            `json:"total_participants"`
	AverageAttendance float64        `json:"average_attendance"`
	AverageDuration   time.Duration  `json:"average_duration"`

	// Efficiency
	MeetingsWithMinutes int          `json:"meetings_with_minutes"`
	ActionItemsCreated  int          `json:"action_items_created"`
	ActionItemsCompleted int         `json:"action_items_completed"`

	// Trends
	MeetingsTrend     float64        `json:"meetings_trend"` // vs previous period
}

// ReportManager manages report generation
type ReportManager struct {
	mu        sync.RWMutex
	reports   map[string]*Report
	counter   int64

	// Data sources (would be interfaces in real implementation)
	xpManager       *XPManager
	meetingManager  *MeetingManager
	ceoManager      *CEOManager

	// Configuration
	config ReportConfig
}

// ReportConfig contains configuration for report generation
type ReportConfig struct {
	// DefaultFormat is the default report format
	DefaultFormat ReportFormat `json:"default_format"`

	// DefaultPeriod is the default time period for reports
	DefaultPeriod time.Duration `json:"default_period"`

	// MaxReportHistory is how many reports to keep
	MaxReportHistory int `json:"max_report_history"`

	// AutoGenerate enables automatic report generation
	AutoGenerate bool `json:"auto_generate"`

	// AutoGenerateSchedule is the cron schedule for auto-generation
	AutoGenerateSchedule string `json:"auto_generate_schedule"`

	// IncludeRawData includes raw data in reports
	IncludeRawData bool `json:"include_raw_data"`

	// MetricsRetentionDays is how long to keep metrics data
	MetricsRetentionDays int `json:"metrics_retention_days"`
}

// DefaultReportConfig returns default report configuration
func DefaultReportConfig() ReportConfig {
	return ReportConfig{
		DefaultFormat:        ReportFormatMarkdown,
		DefaultPeriod:        7 * 24 * time.Hour, // 1 week
		MaxReportHistory:     100,
		AutoGenerate:         false,
		AutoGenerateSchedule: "0 0 * * 0", // Weekly on Sunday
		IncludeRawData:       false,
		MetricsRetentionDays: 90,
	}
}

// NewReportManager creates a new report manager
func NewReportManager(config ReportConfig, xpManager *XPManager,
	meetingManager *MeetingManager, ceoManager *CEOManager) *ReportManager {

	return &ReportManager{
		reports:        make(map[string]*Report),
		xpManager:      xpManager,
		meetingManager: meetingManager,
		ceoManager:     ceoManager,
		config:         config,
	}
}

// GenerateAgentPerformanceReport generates a performance report for agents
func (m *ReportManager) GenerateAgentPerformanceReport(ctx context.Context,
	agentIDs []string, periodStart, periodEnd time.Time) (*Report, error) {

	m.mu.Lock()
	m.counter++
	reportID := fmt.Sprintf("rpt-%d-%d", time.Now().Unix(), m.counter)
	m.mu.Unlock()

	metrics := make([]AgentPerformanceMetrics, 0)
	totalTasks := 0
	totalXP := 0
	sumCompletionRate := 0.0
	sumQuality := 0.0

	// If no specific agents requested, get all agents
	if len(agentIDs) == 0 && m.xpManager != nil {
		// Get leaderboard to get all agents
		leaderboard := m.xpManager.GetLeaderboard(0)
		for _, agent := range leaderboard {
			agentIDs = append(agentIDs, agent.AgentID)
		}
	}

	for _, agentID := range agentIDs {
		metric := m.calculateAgentMetrics(agentID, periodStart, periodEnd)
		metrics = append(metrics, metric)

		totalTasks += metric.TasksCompleted
		totalXP += metric.XPEarned
		sumCompletionRate += metric.CompletionRate
		sumQuality += metric.AverageQuality
	}

	// Calculate averages
	agentCount := len(metrics)
	avgCompletionRate := 0.0
	avgQuality := 0.0
	if agentCount > 0 {
		avgCompletionRate = sumCompletionRate / float64(agentCount)
		avgQuality = sumQuality / float64(agentCount)
	}

	report := &Report{
		ID:          reportID,
		Type:        ReportTypeAgentPerformance,
		Title:       fmt.Sprintf("Agent Performance Report (%s to %s)",
			periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02")),
		Format:      m.config.DefaultFormat,
		GeneratedAt: time.Now(),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Data: map[string]any{
			"metrics":    metrics,
			"agent_count": agentCount,
		},
		Summary: ReportSummary{
			TotalAgents:    agentCount,
			TotalTasks:     totalTasks,
			TotalXP:        totalXP,
			CompletionRate: avgCompletionRate,
			AverageQuality: avgQuality,
		},
		Tags: []string{"performance", "agents"},
	}

	// Generate content based on format
	report.Content = m.formatReport(report)

	m.mu.Lock()
	m.reports[reportID] = report
	m.mu.Unlock()

	return report, nil
}

// calculateAgentMetrics calculates metrics for a single agent
func (m *ReportManager) calculateAgentMetrics(agentID string,
	periodStart, periodEnd time.Time) AgentPerformanceMetrics {

	metrics := AgentPerformanceMetrics{
		AgentID: agentID,
	}

	// Get XP data
	if m.xpManager != nil {
		if agentXP, exists := m.xpManager.GetAgentXP(agentID); exists {
			metrics.Level = agentXP.Level
			metrics.Title = agentXP.Title
			metrics.XPEarned = agentXP.TotalXPEarned
			metrics.CurrentStreak = agentXP.StreakDays
			metrics.TasksCompleted = agentXP.TasksCompleted
		}

		// Calculate completion rate from records
		records := m.xpManager.GetAgentRecords(agentID)
		totalQuality := 0
		qualityCount := 0

		for _, record := range records {
			if record.Timestamp.After(periodStart) && record.Timestamp.Before(periodEnd) {
				totalQuality += record.Quality
				qualityCount++
			}
		}

		if qualityCount > 0 {
			metrics.AverageQuality = float64(totalQuality) / float64(qualityCount)
		}
	}

	// Calculate efficiency score
	metrics.EfficiencyScore = m.calculateEfficiencyScore(&metrics)

	return metrics
}

// calculateEfficiencyScore calculates an overall efficiency score (0-100)
func (m *ReportManager) calculateEfficiencyScore(metrics *AgentPerformanceMetrics) float64 {
	score := 0.0

	// Completion rate contributes up to 40 points
	score += metrics.CompletionRate * 0.4

	// Quality contributes up to 30 points (assuming quality is 1-5)
	score += (metrics.AverageQuality / 5.0) * 30

	// XP growth contributes up to 20 points
	if metrics.XPGrowth > 0 {
		score += math.Min(metrics.XPGrowth, 100) * 0.2
	}

	// Streak contributes up to 10 points
	streakScore := math.Min(float64(metrics.CurrentStreak), 30) / 3
	score += streakScore

	return math.Min(score, 100)
}

// GenerateDepartmentReport generates statistics for departments
func (m *ReportManager) GenerateDepartmentReport(ctx context.Context,
	departments []string, periodStart, periodEnd time.Time) (*Report, error) {

	m.mu.Lock()
	m.counter++
	reportID := fmt.Sprintf("rpt-%d-%d", time.Now().Unix(), m.counter)
	m.mu.Unlock()

	// Group agents by department
	deptStats := make(map[string]*DepartmentStatistics)

	if m.xpManager != nil {
		leaderboard := m.xpManager.GetLeaderboard(0)

		for _, agent := range leaderboard {
			// In real implementation, would get department from agent config
			dept := "general" // Default department

			if _, exists := deptStats[dept]; !exists {
				deptStats[dept] = &DepartmentStatistics{
					Department:           dept,
					WorkloadDistribution: make(map[string]int),
				}
			}

			stats := deptStats[dept]
			stats.AgentCount++
			stats.Agents = append(stats.Agents, agent.AgentID)
			stats.TotalTasks += agent.TasksCompleted
			stats.TotalXP += agent.TotalXPEarned
			stats.AverageLevel += float64(agent.Level)
		}

		// Calculate averages and completion rates
		for _, stats := range deptStats {
			if stats.AgentCount > 0 {
				stats.AverageLevel = stats.AverageLevel / float64(stats.AgentCount)
			}
			if stats.TotalTasks > 0 {
				stats.CompletedTasks = stats.TotalTasks // Simplified
				stats.CompletionRate = 100.0
			}
		}
	}

	// Convert map to slice
	var deptList []DepartmentStatistics
	for _, stats := range deptStats {
		if len(departments) == 0 || contains(departments, stats.Department) {
			deptList = append(deptList, *stats)
		}
	}

	report := &Report{
		ID:          reportID,
		Type:        ReportTypeDepartmentStats,
		Title:       fmt.Sprintf("Department Statistics (%s to %s)",
			periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02")),
		Format:      m.config.DefaultFormat,
		GeneratedAt: time.Now(),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Data: map[string]any{
			"departments": deptList,
		},
		Summary: ReportSummary{
			TotalAgents: len(deptStats),
		},
		Tags: []string{"departments", "statistics"},
	}

	report.Content = m.formatReport(report)

	m.mu.Lock()
	m.reports[reportID] = report
	m.mu.Unlock()

	return report, nil
}

// GenerateMeetingReport generates meeting statistics
func (m *ReportManager) GenerateMeetingReport(ctx context.Context,
	periodStart, periodEnd time.Time) (*Report, error) {

	m.mu.Lock()
	m.counter++
	reportID := fmt.Sprintf("rpt-%d-%d", time.Now().Unix(), m.counter)
	m.mu.Unlock()

	stats := MeetingStatistics{
		MeetingsByType: make(map[string]int),
	}

	if m.meetingManager != nil {
		// Get all meetings
		allMeetings := m.meetingManager.GetMeetingsByStatus(MeetingStatusCompleted)
		_ = m.meetingManager.GetMeetingsByStatus(MeetingStatusScheduled) // For future use
		cancelled := m.meetingManager.GetMeetingsByStatus(MeetingStatusCancelled)
		skipped := m.meetingManager.GetMeetingsByStatus(MeetingStatusSkipped)

		// Filter by period
		for _, m := range allMeetings {
			if m.ScheduledAt.After(periodStart) && m.ScheduledAt.Before(periodEnd) {
				stats.TotalMeetings++
				if m.Status == MeetingStatusCompleted {
					stats.CompletedMeetings++
				}
				stats.MeetingsByType[string(m.Type)]++
				stats.TotalParticipants += len(m.Participants)

				if m.Minutes != nil {
					stats.MeetingsWithMinutes++
					stats.ActionItemsCreated += len(m.Minutes.ActionItems)
				}
			}
		}

		stats.CancelledMeetings = len(cancelled)
		stats.SkippedMeetings = len(skipped)

		// Calculate averages
		if stats.TotalMeetings > 0 {
			stats.AverageAttendance = float64(stats.TotalParticipants) / float64(stats.TotalMeetings)
		}
	}

	report := &Report{
		ID:          reportID,
		Type:        ReportTypeMeetingAnalysis,
		Title:       fmt.Sprintf("Meeting Analysis Report (%s to %s)",
			periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02")),
		Format:      m.config.DefaultFormat,
		GeneratedAt: time.Now(),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Data: map[string]any{
			"statistics": stats,
		},
		Summary: ReportSummary{
			TotalMeetings: stats.TotalMeetings,
		},
		Tags: []string{"meetings", "analysis"},
	}

	report.Content = m.formatReport(report)

	m.mu.Lock()
	m.reports[reportID] = report
	m.mu.Unlock()

	return report, nil
}

// GenerateSystemHealthReport generates overall system health report
func (m *ReportManager) GenerateSystemHealthReport(ctx context.Context) (*Report, error) {
	m.mu.Lock()
	m.counter++
	reportID := fmt.Sprintf("rpt-%d-%d", time.Now().Unix(), m.counter)
	m.mu.Unlock()

	health := map[string]any{
		"status":            "healthy",
		"timestamp":         time.Now(),
		"components":        make(map[string]any),
	}

	// Gather health data from all managers
	if m.xpManager != nil {
		health["xp_stats"] = m.xpManager.GetStats()
	}

	if m.meetingManager != nil {
		health["meeting_stats"] = m.meetingManager.GetStats()
	}

	if m.ceoManager != nil {
		health["directive_stats"] = m.ceoManager.GetStats()
	}

	report := &Report{
		ID:          reportID,
		Type:        ReportTypeSystemHealth,
		Title:       "System Health Report",
		Format:      m.config.DefaultFormat,
		GeneratedAt: time.Now(),
		PeriodStart: time.Now().Add(-24 * time.Hour),
		PeriodEnd:   time.Now(),
		Data:        health,
		Summary:     ReportSummary{},
		Tags:        []string{"health", "system"},
	}

	report.Content = m.formatReport(report)

	m.mu.Lock()
	m.reports[reportID] = report
	m.mu.Unlock()

	return report, nil
}

// formatReport formats the report content based on format
func (m *ReportManager) formatReport(report *Report) string {
	switch report.Format {
	case ReportFormatMarkdown:
		return m.formatMarkdown(report)
	case ReportFormatJSON:
		return m.formatJSON(report)
	case ReportFormatHTML:
		return m.formatHTML(report)
	default:
		return m.formatMarkdown(report)
	}
}

// formatMarkdown formats report as markdown
func (m *ReportManager) formatMarkdown(report *Report) string {
	var buf string

	buf += fmt.Sprintf("# %s\n\n", report.Title)
	buf += fmt.Sprintf("**Generated:** %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	buf += fmt.Sprintf("**Period:** %s to %s\n\n",
		report.PeriodStart.Format("2006-01-02"),
		report.PeriodEnd.Format("2006-01-02"))

	// Summary
	buf += "## Summary\n\n"
	buf += fmt.Sprintf("- **Total Agents:** %d\n", report.Summary.TotalAgents)
	buf += fmt.Sprintf("- **Total Tasks:** %d\n", report.Summary.TotalTasks)
	buf += fmt.Sprintf("- **Total XP:** %d\n", report.Summary.TotalXP)
	buf += fmt.Sprintf("- **Completion Rate:** %.1f%%\n", report.Summary.CompletionRate)
	buf += fmt.Sprintf("- **Average Quality:** %.2f\n\n", report.Summary.AverageQuality)

	return buf
}

// formatJSON formats report as JSON (placeholder)
func (m *ReportManager) formatJSON(report *Report) string {
	// In real implementation, would use json.Marshal
	return "{\"report\": \"json_format\"}"
}

// formatHTML formats report as HTML (placeholder)
func (m *ReportManager) formatHTML(report *Report) string {
	return fmt.Sprintf("<h1>%s</h1>", report.Title)
}

// GetReport retrieves a report by ID
func (m *ReportManager) GetReport(reportID string) (*Report, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	report, exists := m.reports[reportID]
	return report, exists
}

// GetReports returns all generated reports
func (m *ReportManager) GetReports(reportType ReportType, limit int) []*Report {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var reports []*Report
	for _, r := range m.reports {
		if reportType == "" || r.Type == reportType {
			reports = append(reports, r)
		}
	}

	// Sort by generated time (newest first)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].GeneratedAt.After(reports[j].GeneratedAt)
	})

	if limit > 0 && limit < len(reports) {
		return reports[:limit]
	}
	return reports
}

// DeleteReport deletes a report
func (m *ReportManager) DeleteReport(reportID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.reports[reportID]; !exists {
		return fmt.Errorf("report not found: %s", reportID)
	}

	delete(m.reports, reportID)
	return nil
}

// CleanupOldReports removes old reports beyond retention limit
func (m *ReportManager) CleanupOldReports() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.reports) <= m.config.MaxReportHistory {
		return 0
	}

	// Get all reports sorted by time
	var reports []*Report
	for _, r := range m.reports {
		reports = append(reports, r)
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].GeneratedAt.Before(reports[j].GeneratedAt)
	})

	// Remove oldest
	toRemove := len(reports) - m.config.MaxReportHistory
	for i := 0; i < toRemove; i++ {
		delete(m.reports, reports[i].ID)
	}

	return toRemove
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
