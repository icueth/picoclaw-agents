// Package office provides Office UI functionality for Picoclaw agent management
package office

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// NotificationPriority represents the priority of a notification
type NotificationPriority string

const (
	NotificationPriorityLow      NotificationPriority = "low"
	NotificationPriorityNormal   NotificationPriority = "normal"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityUrgent   NotificationPriority = "urgent"
	NotificationPriorityCritical NotificationPriority = "critical"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeTask       NotificationType = "task"
	NotificationTypeMeeting    NotificationType = "meeting"
	NotificationTypeMention    NotificationType = "mention"
	NotificationTypeSystem     NotificationType = "system"
	NotificationTypeAlert      NotificationType = "alert"
	NotificationTypeReminder   NotificationType = "reminder"
	NotificationTypeAchievement NotificationType = "achievement"
	NotificationTypeReport     NotificationType = "report"
	NotificationTypeBroadcast  NotificationType = "broadcast"
	NotificationTypeDirective  NotificationType = "directive"
)

// Notification represents a notification message
type Notification struct {
	ID          string               `json:"id"`
	Type        NotificationType     `json:"type"`
	Priority    NotificationPriority `json:"priority"`
	Title       string               `json:"title"`
	Message     string               `json:"message"`
	RecipientID string               `json:"recipient_id,omitempty"`
	SenderID    string               `json:"sender_id,omitempty"`
	Channel     string               `json:"channel,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	ExpiresAt   *time.Time           `json:"expires_at,omitempty"`
	ReadAt      *time.Time           `json:"read_at,omitempty"`
	ActionURL   string               `json:"action_url,omitempty"`
	ActionText  string               `json:"action_text,omitempty"`
	Tags        []string             `json:"tags,omitempty"`
	Metadata    map[string]any       `json:"metadata,omitempty"`

	// Delivery tracking
	Delivered   bool      `json:"delivered"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// IsRead returns true if the notification has been read
func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}

// IsExpired returns true if the notification has expired
func (n *Notification) IsExpired() bool {
	if n.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*n.ExpiresAt)
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.ReadAt = &now
}

// NotificationManager manages notifications and alerts
type NotificationManager struct {
	mu            sync.RWMutex
	notifications map[string]*Notification
	channels      map[string]NotificationChannel
	handlers      []NotificationHandler
	counter       int64

	// Configuration
	config NotificationConfig
}

// NotificationConfig contains configuration for the notification system
type NotificationConfig struct {
	// Enabled enables/disables notifications
	Enabled bool `json:"enabled"`

	// RetentionHours is how long to keep notifications
	RetentionHours int `json:"retention_hours"`

	// MaxNotificationsPerAgent limits notifications per agent
	MaxNotificationsPerAgent int `json:"max_notifications_per_agent"`

	// DefaultPriority is the default notification priority
	DefaultPriority NotificationPriority `json:"default_priority"`

	// RealTimeEnabled enables real-time notifications
	RealTimeEnabled bool `json:"real_time_enabled"`

	// BatchInterval is the interval for batching notifications
	BatchInterval time.Duration `json:"batch_interval"`

	// RateLimitPerMinute limits notifications per minute
	RateLimitPerMinute int `json:"rate_limit_per_minute"`

	// Email configuration
	Email EmailConfig `json:"email,omitempty"`

	// Slack configuration
	Slack SlackConfig `json:"slack,omitempty"`

	// Webhook configuration
	Webhook WebhookConfig `json:"webhook,omitempty"`
}

// EmailConfig contains email notification settings
type EmailConfig struct {
	Enabled    bool   `json:"enabled"`
	SMTPHost   string `json:"smtp_host,omitempty"`
	SMTPPort   int    `json:"smtp_port,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	From       string `json:"from,omitempty"`
	UseTLS     bool   `json:"use_tls"`
	MaxPerHour int    `json:"max_per_hour"`
}

// SlackConfig contains Slack notification settings
type SlackConfig struct {
	Enabled     bool              `json:"enabled"`
	WebhookURL  string            `json:"webhook_url,omitempty"`
	BotToken    string            `json:"bot_token,omitempty"`
	Channel     string            `json:"channel,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Mentions    map[string]string `json:"mentions,omitempty"`
	MaxPerHour  int               `json:"max_per_hour"`
}

// WebhookConfig contains generic webhook settings
type WebhookConfig struct {
	Enabled    bool              `json:"enabled"`
	URL        string            `json:"url,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Timeout    int               `json:"timeout_seconds"`
	MaxRetries int               `json:"max_retries"`
}

// DefaultNotificationConfig returns default notification configuration
func DefaultNotificationConfig() NotificationConfig {
	return NotificationConfig{
		Enabled:                  true,
		RetentionHours:           168, // 7 days
		MaxNotificationsPerAgent: 1000,
		DefaultPriority:          NotificationPriorityNormal,
		RealTimeEnabled:          true,
		BatchInterval:            5 * time.Second,
		RateLimitPerMinute:       60,
		Email: EmailConfig{
			Enabled:    false,
			SMTPPort:   587,
			UseTLS:     true,
			MaxPerHour: 100,
		},
		Slack: SlackConfig{
			Enabled:    false,
			Username:   "Picoclaw Office",
			IconEmoji:  ":robot_face:",
			MaxPerHour: 50,
		},
		Webhook: WebhookConfig{
			Enabled:    false,
			Timeout:    30,
			MaxRetries: 3,
		},
	}
}

// NotificationChannel represents a notification delivery channel
type NotificationChannel interface {
	Send(ctx context.Context, notification *Notification) error
	IsEnabled() bool
	GetName() string
}

// NotificationHandler is called when a notification is created
type NotificationHandler func(notification *Notification)

// NewNotificationManager creates a new notification manager
func NewNotificationManager(config NotificationConfig) *NotificationManager {
	return &NotificationManager{
		notifications: make(map[string]*Notification),
		channels:      make(map[string]NotificationChannel),
		handlers:      make([]NotificationHandler, 0),
		config:        config,
	}
}

// RegisterChannel registers a notification channel
func (m *NotificationManager) RegisterChannel(name string, channel NotificationChannel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[name] = channel
}

// RegisterHandler registers a notification handler
func (m *NotificationManager) RegisterHandler(handler NotificationHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
}

// CreateNotification creates a new notification
func (m *NotificationManager) CreateNotification(ctx context.Context,
	notifType NotificationType, priority NotificationPriority,
	title, message, recipientID string) (*Notification, error) {

	if !m.config.Enabled {
		return nil, fmt.Errorf("notifications are disabled")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter++
	id := fmt.Sprintf("notif-%d-%d", time.Now().Unix(), m.counter)

	if priority == "" {
		priority = m.config.DefaultPriority
	}

	notification := &Notification{
		ID:          id,
		Type:        notifType,
		Priority:    priority,
		Title:       title,
		Message:     message,
		RecipientID: recipientID,
		CreatedAt:   time.Now(),
		Tags:        make([]string, 0),
		Metadata:    make(map[string]any),
	}

	// Set expiration
	if m.config.RetentionHours > 0 {
		expiry := time.Now().Add(time.Duration(m.config.RetentionHours) * time.Hour)
		notification.ExpiresAt = &expiry
	}

	m.notifications[id] = notification

	// Trigger handlers
	for _, handler := range m.handlers {
		go handler(notification)
	}

	// Deliver to channels
	if m.config.RealTimeEnabled {
		go m.deliverNotification(ctx, notification)
	}

	return notification, nil
}

// deliverNotification sends notification to all enabled channels
func (m *NotificationManager) deliverNotification(ctx context.Context, notification *Notification) {
	m.mu.RLock()
	channels := make(map[string]NotificationChannel)
	for k, v := range m.channels {
		channels[k] = v
	}
	m.mu.RUnlock()

	for name, channel := range channels {
		if !channel.IsEnabled() {
			continue
		}

		err := channel.Send(ctx, notification)
		if err != nil {
			m.mu.Lock()
			notification.Error = fmt.Sprintf("%s: %v", name, err)
			m.mu.Unlock()
		}
	}

	// Mark as delivered
	now := time.Now()
	m.mu.Lock()
	notification.Delivered = true
	notification.DeliveredAt = &now
	m.mu.Unlock()
}

// GetNotification retrieves a notification by ID
func (m *NotificationManager) GetNotification(id string) (*Notification, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	notification, exists := m.notifications[id]
	return notification, exists
}

// GetNotificationsForAgent returns notifications for a specific agent
func (m *NotificationManager) GetNotificationsForAgent(agentID string, unreadOnly bool, limit int) []*Notification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var notifications []*Notification
	for _, n := range m.notifications {
		if n.RecipientID != agentID && agentID != "" {
			continue
		}

		if n.IsExpired() {
			continue
		}

		if unreadOnly && n.IsRead() {
			continue
		}

		notifications = append(notifications, n)
	}

	// Sort by created time (newest first)
	for i := 0; i < len(notifications); i++ {
		for j := i + 1; j < len(notifications); j++ {
			if notifications[j].CreatedAt.After(notifications[i].CreatedAt) {
				notifications[i], notifications[j] = notifications[j], notifications[i]
			}
		}
	}

	if limit > 0 && limit < len(notifications) {
		return notifications[:limit]
	}
	return notifications
}

// MarkAsRead marks a notification as read
func (m *NotificationManager) MarkAsRead(notificationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	notification, exists := m.notifications[notificationID]
	if !exists {
		return fmt.Errorf("notification not found: %s", notificationID)
	}

	notification.MarkAsRead()
	return nil
}

// MarkAllAsRead marks all notifications for an agent as read
func (m *NotificationManager) MarkAllAsRead(agentID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, n := range m.notifications {
		if n.RecipientID == agentID && !n.IsRead() {
			n.MarkAsRead()
			count++
		}
	}

	return count
}

// DeleteNotification deletes a notification
func (m *NotificationManager) DeleteNotification(notificationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.notifications[notificationID]; !exists {
		return fmt.Errorf("notification not found: %s", notificationID)
	}

	delete(m.notifications, notificationID)
	return nil
}

// CleanupExpired removes expired notifications
func (m *NotificationManager) CleanupExpired() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for id, n := range m.notifications {
		if n.IsExpired() {
			delete(m.notifications, id)
			count++
		}
	}

	return count
}

// SendAlert sends an urgent alert
func (m *NotificationManager) SendAlert(ctx context.Context, title, message, recipientID string) (*Notification, error) {
	return m.CreateNotification(ctx, NotificationTypeAlert, NotificationPriorityUrgent,
		title, message, recipientID)
}

// SendReminder sends a reminder notification
func (m *NotificationManager) SendReminder(ctx context.Context, title, message, recipientID string,
	priority NotificationPriority) (*Notification, error) {
	return m.CreateNotification(ctx, NotificationTypeReminder, priority,
		title, message, recipientID)
}

// Broadcast sends a notification to all agents
func (m *NotificationManager) Broadcast(ctx context.Context, notifType NotificationType,
	priority NotificationPriority, title, message string) ([]*Notification, error) {

	// This would typically get all agent IDs from a registry
	// For now, create a single broadcast notification
	notification, err := m.CreateNotification(ctx, notifType, priority, title, message, "")
	if err != nil {
		return nil, err
	}

	// Mark as broadcast
	notification.Metadata["broadcast"] = true

	return []*Notification{notification}, nil
}

// GetUnreadCount returns the number of unread notifications for an agent
func (m *NotificationManager) GetUnreadCount(agentID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, n := range m.notifications {
		if n.RecipientID == agentID && !n.IsRead() && !n.IsExpired() {
			count++
		}
	}

	return count
}

// GetStats returns notification statistics
func (m *NotificationManager) GetStats() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]any{
		"total_notifications": len(m.notifications),
		"unread_count":        0,
		"read_count":          0,
		"expired_count":       0,
		"by_type":             make(map[string]int),
		"by_priority":         make(map[string]int),
	}

	byType := stats["by_type"].(map[string]int)
	byPriority := stats["by_priority"].(map[string]int)

	for _, n := range m.notifications {
		if n.IsExpired() {
			stats["expired_count"] = stats["expired_count"].(int) + 1
		} else if n.IsRead() {
			stats["read_count"] = stats["read_count"].(int) + 1
		} else {
			stats["unread_count"] = stats["unread_count"].(int) + 1
		}

		byType[string(n.Type)]++
		byPriority[string(n.Priority)]++
	}

	return stats
}

// SlackChannel implements Slack notification delivery
type SlackChannel struct {
	config SlackConfig
	client *http.Client
}

// NewSlackChannel creates a new Slack notification channel
func NewSlackChannel(config SlackConfig) *SlackChannel {
	return &SlackChannel{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsEnabled returns true if Slack is enabled
func (s *SlackChannel) IsEnabled() bool {
	return s.config.Enabled && (s.config.WebhookURL != "" || s.config.BotToken != "")
}

// GetName returns the channel name
func (s *SlackChannel) GetName() string {
	return "slack"
}

// Send sends a notification to Slack
func (s *SlackChannel) Send(ctx context.Context, notification *Notification) error {
	if !s.IsEnabled() {
		return fmt.Errorf("slack channel is not enabled")
	}

	payload := map[string]any{
		"text": notification.Title,
		"attachments": []map[string]any{
			{
				"color":    s.getColorForPriority(notification.Priority),
				"title":    notification.Title,
				"text":     notification.Message,
				"footer":   "Picoclaw Office",
				"ts":       notification.CreatedAt.Unix(),
				"fallback": notification.Message,
			},
		},
	}

	if s.config.Username != "" {
		payload["username"] = s.config.Username
	}

	if s.config.IconEmoji != "" {
		payload["icon_emoji"] = s.config.IconEmoji
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *SlackChannel) getColorForPriority(priority NotificationPriority) string {
	switch priority {
	case NotificationPriorityCritical:
		return "#FF0000"
	case NotificationPriorityUrgent:
		return "#FF8C00"
	case NotificationPriorityHigh:
		return "#FFD700"
	case NotificationPriorityLow:
		return "#808080"
	default:
		return "#36A64F"
	}
}

// WebhookChannel implements generic webhook notification delivery
type WebhookChannel struct {
	config WebhookConfig
	client *http.Client
}

// NewWebhookChannel creates a new webhook notification channel
func NewWebhookChannel(config WebhookConfig) *WebhookChannel {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30
	}

	return &WebhookChannel{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// IsEnabled returns true if webhook is enabled
func (w *WebhookChannel) IsEnabled() bool {
	return w.config.Enabled && w.config.URL != ""
}

// GetName returns the channel name
func (w *WebhookChannel) GetName() string {
	return "webhook"
}

// Send sends a notification via webhook
func (w *WebhookChannel) Send(ctx context.Context, notification *Notification) error {
	if !w.IsEnabled() {
		return fmt.Errorf("webhook channel is not enabled")
	}

	payload := map[string]any{
		"id":          notification.ID,
		"type":        notification.Type,
		"priority":    notification.Priority,
		"title":       notification.Title,
		"message":     notification.Message,
		"recipient":   notification.RecipientID,
		"sender":      notification.SenderID,
		"created_at":  notification.CreatedAt,
		"tags":        notification.Tags,
		"metadata":    notification.Metadata,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= w.config.MaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", w.config.URL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create webhook request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		for key, value := range w.config.Headers {
			req.Header.Set(key, value)
		}

		resp, err := w.client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < w.config.MaxRetries {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		if attempt < w.config.MaxRetries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	return fmt.Errorf("webhook failed after %d retries: %w", w.config.MaxRetries, lastErr)
}

// EmailChannel implements email notification delivery (placeholder)
type EmailChannel struct {
	config EmailConfig
}

// NewEmailChannel creates a new email notification channel
func NewEmailChannel(config EmailConfig) *EmailChannel {
	return &EmailChannel{config: config}
}

// IsEnabled returns true if email is enabled
func (e *EmailChannel) IsEnabled() bool {
	return e.config.Enabled && e.config.SMTPHost != "" && e.config.From != ""
}

// GetName returns the channel name
func (e *EmailChannel) GetName() string {
	return "email"
}

// Send sends a notification via email
func (e *EmailChannel) Send(ctx context.Context, notification *Notification) error {
	if !e.IsEnabled() {
		return fmt.Errorf("email channel is not enabled")
	}

	// Email sending implementation would go here
	// This is a placeholder that would integrate with an email library
	return fmt.Errorf("email sending not yet implemented")
}

// RealTimeAlert represents a real-time alert
type RealTimeAlert struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	Data      map[string]any `json:"data,omitempty"`
}

// SendRealTimeAlert sends a real-time alert to all connected clients
func (m *NotificationManager) SendRealTimeAlert(ctx context.Context, alert *RealTimeAlert) error {
	if !m.config.RealTimeEnabled {
		return fmt.Errorf("real-time alerts are disabled")
	}

	// Create notification from alert
	priority := NotificationPriorityHigh
	switch alert.Severity {
	case "critical":
		priority = NotificationPriorityCritical
	case "warning":
		priority = NotificationPriorityHigh
	case "info":
		priority = NotificationPriorityNormal
	}

	_, err := m.CreateNotification(ctx, NotificationTypeAlert, priority,
		alert.Title, alert.Message, "")

	return err
}
