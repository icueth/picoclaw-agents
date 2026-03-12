// Package agent provides job monitoring and adaptive timeout management
// for subagent tasks with automatic health checking and self-healing capabilities.
package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/tools"
)

// JobHealthStatus represents the health status of a job
type JobHealthStatus string

const (
	JobHealthHealthy   JobHealthStatus = "healthy"   // Job is running normally
	JobHealthWarning   JobHealthStatus = "warning"   // Job is slow but progressing
	JobHealthStuck     JobHealthStatus = "stuck"     // Job appears stuck (no progress)
	JobHealthCritical  JobHealthStatus = "critical"  // Job near timeout
	JobHealthTimeout   JobHealthStatus = "timeout"   // Job exceeded timeout
	JobHealthCompleted JobHealthStatus = "completed" // Job finished successfully
)

// JobMonitorConfig configures the job monitoring behavior
type JobMonitorConfig struct {
	CheckInterval         time.Duration // How often to check jobs (default: 30s)
	WarningThreshold      float64       // Percentage of timeout to trigger warning (default: 0.8)
	StuckThreshold        time.Duration // Time without progress to consider stuck (default: 2m)
	GracePeriod           time.Duration // Extra time after warning before timeout (default: 1m)
	AutoExtendEnabled     bool          // Whether to auto-extend timeout (default: true)
	AutoExtendThreshold   float64       // Percentage to trigger auto-extend (default: 0.75)
	MaxAutoExtensions     int           // Max number of auto-extensions (default: 3)
	AutoExtendDuration    time.Duration // How much to extend (default: 5m)
	ProgressCheckInterval time.Duration // How often subagent should report progress (default: 30s)
}

// DefaultJobMonitorConfig returns sensible defaults
func DefaultJobMonitorConfig() *JobMonitorConfig {
	return &JobMonitorConfig{
		CheckInterval:         30 * time.Second,
		WarningThreshold:      0.8, // 80% of timeout
		StuckThreshold:        2 * time.Minute,
		GracePeriod:           1 * time.Minute,
		AutoExtendEnabled:     true,
		AutoExtendThreshold:   0.75, // 75% of timeout
		MaxAutoExtensions:     3,
		AutoExtendDuration:    5 * time.Minute,
		ProgressCheckInterval: 30 * time.Second,
	}
}

// JobInfo contains monitoring information about a job
type JobInfo struct {
	TaskID              string
	Status              string
	Health              JobHealthStatus
	ProgressPercent     int
	ProgressMessage     string
	StartedAt           time.Time
	TimeoutAt           time.Time
	LastProgressAt      time.Time
	TimeElapsed         time.Duration
	TimeRemaining       time.Duration
	TimeoutPercent      float64
	ExtensionsUsed      int
	MaxExtensions       int
	IsExtendable        bool
	EstimatedCompletion time.Duration
}

// JobMonitor monitors subagent jobs and manages adaptive timeouts
type JobMonitor struct {
	config          *JobMonitorConfig
	subagentManager *tools.SubagentManager
	stopCh          chan struct{}
	wg              sync.WaitGroup
	mu              sync.RWMutex
	// Callbacks for notifications
	onWarning     func(taskID string, info *JobInfo)
	onStuck       func(taskID string, info *JobInfo)
	onTimeout     func(taskID string, info *JobInfo)
	onAutoExtend  func(taskID string, info *JobInfo, newTimeout time.Time)
	onProgress    func(taskID string, info *JobInfo)
}

// NewJobMonitor creates a new job monitor
func NewJobMonitor(config *JobMonitorConfig, sm *tools.SubagentManager) *JobMonitor {
	if config == nil {
		config = DefaultJobMonitorConfig()
	}
	return &JobMonitor{
		config:          config,
		subagentManager: sm,
		stopCh:          make(chan struct{}),
	}
}

// SetCallbacks sets notification callbacks
func (jm *JobMonitor) SetCallbacks(
	onWarning func(taskID string, info *JobInfo),
	onStuck func(taskID string, info *JobInfo),
	onTimeout func(taskID string, info *JobInfo),
	onAutoExtend func(taskID string, info *JobInfo, newTimeout time.Time),
	onProgress func(taskID string, info *JobInfo),
) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	jm.onWarning = onWarning
	jm.onStuck = onStuck
	jm.onTimeout = onTimeout
	jm.onAutoExtend = onAutoExtend
	jm.onProgress = onProgress
}

// Start begins the monitoring loop
func (jm *JobMonitor) Start(ctx context.Context) {
	jm.wg.Add(1)
	go jm.monitorLoop(ctx)
	logger.InfoCF("job_monitor", "Job monitor started", map[string]any{
		"check_interval": jm.config.CheckInterval.Seconds(),
	})
}

// Stop stops the monitoring loop
func (jm *JobMonitor) Stop() {
	close(jm.stopCh)
	jm.wg.Wait()
	logger.InfoCF("job_monitor", "Job monitor stopped", nil)
}

// monitorLoop runs the periodic health checks
func (jm *JobMonitor) monitorLoop(ctx context.Context) {
	defer jm.wg.Done()

	ticker := time.NewTicker(jm.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-jm.stopCh:
			return
		case <-ticker.C:
			jm.checkAllJobs()
		}
	}
}

// checkAllJobs checks the health of all running jobs
func (jm *JobMonitor) checkAllJobs() {
	if jm.subagentManager == nil {
		return
	}

	tasks := jm.subagentManager.ListActiveTasks()
	for _, task := range tasks {
		jm.checkJobHealth(task)
	}
}

// checkJobHealth checks the health of a specific job
func (jm *JobMonitor) checkJobHealth(task *tools.SubagentTask) {
	info := jm.GetJobInfo(task.ID)
	if info == nil {
		return
	}

	// Skip if already completed or failed
	if info.Status == "completed" || info.Status == "failed" || info.Status == "cancelled" {
		return
	}

	// Check for timeout
	if info.TimeRemaining <= 0 {
		jm.handleTimeout(task, info)
		return
	}

	// Check if job is stuck (no progress for too long)
	if jm.isJobStuck(info) {
		jm.handleStuck(task, info)
		return
	}

	// Check if job needs warning (near timeout)
	if jm.needsWarning(info) {
		jm.handleWarning(task, info)
		return
	}

	// Check if job should be auto-extended
	if jm.shouldAutoExtend(info) {
		jm.handleAutoExtend(task, info)
		return
	}

	// Normal progress check
	if jm.onProgress != nil {
		jm.onProgress(task.ID, info)
	}
}

// isJobStuck checks if a job hasn't made progress for too long
func (jm *JobMonitor) isJobStuck(info *JobInfo) bool {
	if info.Status != "running" {
		return false
	}
	timeSinceProgress := time.Since(info.LastProgressAt)
	return timeSinceProgress > jm.config.StuckThreshold
}

// needsWarning checks if job is near timeout
func (jm *JobMonitor) needsWarning(info *JobInfo) bool {
	if info.Status != "running" || !info.IsExtendable {
		return false
	}
	return info.TimeoutPercent >= jm.config.WarningThreshold*100
}

// shouldAutoExtend checks if job should be auto-extended
func (jm *JobMonitor) shouldAutoExtend(info *JobInfo) bool {
	if !jm.config.AutoExtendEnabled || !info.IsExtendable {
		return false
	}
	if info.ExtensionsUsed >= info.MaxExtensions {
		return false
	}
	return info.TimeoutPercent >= jm.config.AutoExtendThreshold*100
}

// handleStuck handles a stuck job
func (jm *JobMonitor) handleStuck(task *tools.SubagentTask, info *JobInfo) {
	info.Health = JobHealthStuck
	timeSinceProgress := time.Since(info.LastProgressAt)

	logger.WarnCF("job_monitor", "Job appears stuck", map[string]any{
		"task_id":               task.ID,
		"time_since_progress":   timeSinceProgress.Seconds(),
		"progress_percent":      info.ProgressPercent,
		"progress_message":      info.ProgressMessage,
	})

	if jm.onStuck != nil {
		jm.onStuck(task.ID, info)
	}

	// If stuck for too long (grace period), cancel the job
	if timeSinceProgress > jm.config.StuckThreshold+jm.config.GracePeriod {
		logger.ErrorCF("job_monitor", "Cancelling stuck job", map[string]any{
			"task_id": task.ID,
			"reason":  "no progress for extended period",
		})
		jm.subagentManager.CancelTask(task.ID)
	}
}

// handleWarning handles a job near timeout
func (jm *JobMonitor) handleWarning(task *tools.SubagentTask, info *JobInfo) {
	info.Health = JobHealthWarning

	logger.WarnCF("job_monitor", "Job near timeout", map[string]any{
		"task_id":          task.ID,
		"time_remaining":   info.TimeRemaining.Seconds(),
		"timeout_percent":  info.TimeoutPercent,
		"progress_percent": info.ProgressPercent,
	})

	if jm.onWarning != nil {
		jm.onWarning(task.ID, info)
	}
}

// handleTimeout handles a timed-out job
func (jm *JobMonitor) handleTimeout(task *tools.SubagentTask, info *JobInfo) {
	info.Health = JobHealthTimeout

	logger.ErrorCF("job_monitor", "Job timed out", map[string]any{
		"task_id":          task.ID,
		"time_elapsed":     info.TimeElapsed.Seconds(),
		"progress_percent": info.ProgressPercent,
	})

	if jm.onTimeout != nil {
		jm.onTimeout(task.ID, info)
	}

	// Cancel the task
	jm.subagentManager.CancelTask(task.ID)
}

// handleAutoExtend handles auto-extending a job's timeout
func (jm *JobMonitor) handleAutoExtend(task *tools.SubagentTask, info *JobInfo) {
	// Extend the timeout
	newTimeout := jm.subagentManager.ExtendTaskTimeout(task.ID, jm.config.AutoExtendDuration)
	if newTimeout.IsZero() {
		return
	}

	info.ExtensionsUsed++
	info.TimeoutAt = newTimeout
	info.TimeRemaining = time.Until(newTimeout)
	info.Health = JobHealthHealthy

	logger.InfoCF("job_monitor", "Auto-extended job timeout", map[string]any{
		"task_id":         task.ID,
		"extension":       jm.config.AutoExtendDuration.Minutes(),
		"new_timeout":     newTimeout.Format(time.RFC3339),
		"extensions_used": info.ExtensionsUsed,
		"max_extensions":  info.MaxExtensions,
	})

	if jm.onAutoExtend != nil {
		jm.onAutoExtend(task.ID, info, newTimeout)
	}
}

// GetJobInfo retrieves monitoring information for a job
func (jm *JobMonitor) GetJobInfo(taskID string) *JobInfo {
	if jm.subagentManager == nil {
		return nil
	}

	status := jm.subagentManager.GetTaskStatus(taskID)
	if status == nil {
		return nil
	}

	exists, _ := status["exists"].(bool)
	if !exists {
		return nil
	}

	info := &JobInfo{
		TaskID:         taskID,
		Status:         getString(status, "status"),
		ProgressPercent: getInt(status, "progress_percent"),
		ProgressMessage: getString(status, "progress_message"),
		ExtensionsUsed:  getInt(status, "extensions_used"),
		MaxExtensions:   getInt(status, "max_extensions"),
		IsExtendable:    getBool(status, "is_extendable"),
	}

	// Parse timestamps
	if started, ok := status["started"].(int64); ok && started > 0 {
		info.StartedAt = time.UnixMilli(started)
	}
	if lastProgress, ok := status["last_progress"].(int64); ok && lastProgress > 0 {
		info.LastProgressAt = time.UnixMilli(lastProgress)
	} else {
		info.LastProgressAt = info.StartedAt
	}

	// Calculate timing
	info.TimeElapsed = time.Since(info.StartedAt)

	// Get timeout information
	if timeoutAt, ok := status["timeout_at"].(int64); ok && timeoutAt > 0 {
		info.TimeoutAt = time.UnixMilli(timeoutAt)
		info.TimeRemaining = time.Until(info.TimeoutAt)
	}

	// Calculate timeout percentage
	if maxDuration, ok := status["max_duration_ms"].(int64); ok && maxDuration > 0 {
		info.TimeoutPercent = (float64(info.TimeElapsed.Milliseconds()) / float64(maxDuration)) * 100
	}

	// Determine health status
	info.Health = jm.determineHealth(info)

	return info
}

// determineHealth determines the health status of a job
func (jm *JobMonitor) determineHealth(info *JobInfo) JobHealthStatus {
	switch info.Status {
	case "completed":
		return JobHealthCompleted
	case "failed", "cancelled":
		return JobHealthCritical
	case "running":
		if info.TimeRemaining <= 0 {
			return JobHealthTimeout
		}
		timeSinceProgress := time.Since(info.LastProgressAt)
		if timeSinceProgress > jm.config.StuckThreshold {
			return JobHealthStuck
		}
		if info.TimeoutPercent >= jm.config.WarningThreshold*100 {
			return JobHealthWarning
		}
		return JobHealthHealthy
	default:
		return JobHealthHealthy
	}
}

// GetAllJobInfos returns monitoring info for all jobs
func (jm *JobMonitor) GetAllJobInfos() []*JobInfo {
	if jm.subagentManager == nil {
		return nil
	}

	tasks := jm.subagentManager.ListTasks()
	infos := make([]*JobInfo, 0, len(tasks))
	for _, task := range tasks {
		if info := jm.GetJobInfo(task.ID); info != nil {
			infos = append(infos, info)
		}
	}
	return infos
}

// GetActiveJobInfos returns monitoring info for active jobs only
func (jm *JobMonitor) GetActiveJobInfos() []*JobInfo {
	if jm.subagentManager == nil {
		return nil
	}

	tasks := jm.subagentManager.ListActiveTasks()
	infos := make([]*JobInfo, 0, len(tasks))
	for _, task := range tasks {
		if info := jm.GetJobInfo(task.ID); info != nil {
			infos = append(infos, info)
		}
	}
	return infos
}

// Helper functions for type conversion
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	if v, ok := m[key].(int64); ok {
		return int(v)
	}
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

// FormatDuration formats a duration for human reading
func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

// FormatJobInfo formats job info for display
func FormatJobInfo(info *JobInfo) string {
	var statusEmoji string
	switch info.Health {
	case JobHealthHealthy:
		statusEmoji = "✅"
	case JobHealthWarning:
		statusEmoji = "⚠️"
	case JobHealthStuck:
		statusEmoji = "🔄"
	case JobHealthCritical, JobHealthTimeout:
		statusEmoji = "❌"
	case JobHealthCompleted:
		statusEmoji = "✨"
	default:
		statusEmoji = "❓"
	}

	result := fmt.Sprintf("%s **%s** (%s)\n", statusEmoji, info.TaskID, info.Status)
	result += fmt.Sprintf("   Health: %s | Progress: %d%%\n", info.Health, info.ProgressPercent)
	result += fmt.Sprintf("   Elapsed: %s | Remaining: %s (%.0f%%)\n",
		FormatDuration(info.TimeElapsed),
		FormatDuration(info.TimeRemaining),
		info.TimeoutPercent)

	if info.ProgressMessage != "" {
		result += fmt.Sprintf("   Message: %s\n", info.ProgressMessage)
	}

	if info.IsExtendable {
		result += fmt.Sprintf("   Extensions: %d/%d\n", info.ExtensionsUsed, info.MaxExtensions)
	}

	return result
}
