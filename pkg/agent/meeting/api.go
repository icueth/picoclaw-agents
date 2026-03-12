// Package meeting provides HTTP API for agent meetings
package meeting

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"picoclaw/agent/pkg/agent/persona"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/providers"
)

// APIHandler provides HTTP handlers for meeting management
type APIHandler struct {
	conference  *ConferenceManager
	scheduler   *Scheduler
	aiManager   *AIDiscussionManager
	cfg         *config.Config
	agents      []config.AgentConfig
	personaDir  string
}

// NewAPIHandler creates a new meeting API handler
func NewAPIHandler(cfg *config.Config, provider providers.LLMProvider, agents []config.AgentConfig) *APIHandler {
	conference := NewConferenceManager()
	conference.PopulateFromRegistry(agents)
	
	scheduler := NewScheduler()
	
	homeDir := "."
	if h, err := os.UserHomeDir(); err == nil {
		homeDir = h
	}
	personaDir := filepath.Join(homeDir, ".picoclaw", "agents")
	
	var aiManager *AIDiscussionManager
	if provider != nil {
		aiManager = NewAIDiscussionManager(provider, conference, personaDir)
	}
	
	return &APIHandler{
		conference: conference,
		scheduler:  scheduler,
		aiManager:  aiManager,
		cfg:        cfg,
		agents:     agents,
		personaDir: personaDir,
	}
}

// RegisterRoutes registers all meeting API routes
func (h *APIHandler) RegisterRoutes(mux *http.ServeMux) {
	// Meeting management
	mux.HandleFunc("/api/meetings", h.handleMeetings)
	mux.HandleFunc("/api/meetings/", h.handleMeetingDetail)
	
	// Scheduler
	mux.HandleFunc("/api/schedule", h.handleSchedule)
	mux.HandleFunc("/api/schedule/", h.handleScheduleDetail)
	mux.HandleFunc("/api/schedule/upcoming", h.handleUpcoming)
	
	// AI Discussions
	mux.HandleFunc("/api/discussions", h.handleDiscussions)
	mux.HandleFunc("/api/discussions/", h.handleDiscussionDetail)
	
	// Agents
	mux.HandleFunc("/api/agents", h.handleAgents)
	mux.HandleFunc("/api/agents/", h.handleAgentDetail)
	
	// WebSocket for real-time updates
	mux.HandleFunc("/ws/meetings", h.handleMeetingWebSocket)
}

// MeetingRequest represents a request to create a meeting
type MeetingRequest struct {
	Topic        string   `json:"topic"`
	Description  string   `json:"description"`
	Participants []string `json:"participants"`
	Facilitator  string   `json:"facilitator"`
	Agenda       []string `json:"agenda"`
	AutoStart    bool     `json:"auto_start"`
}

// MeetingResponse represents a meeting response
type MeetingResponse struct {
	ID           string            `json:"id"`
	Topic        string            `json:"topic"`
	Description  string            `json:"description"`
	Status       string            `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	Facilitator  string            `json:"facilitator"`
	Participants []ParticipantInfo `json:"participants"`
	Agenda       []AgendaItem      `json:"agenda"`
	MessageCount int               `json:"message_count"`
}

// ParticipantInfo represents a meeting participant
type ParticipantInfo struct {
	AgentID  string `json:"agent_id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
	IsOnline bool   `json:"is_online"`
}

func (h *APIHandler) handleMeetings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listMeetings(w, r)
	case http.MethodPost:
		h.createMeeting(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) listMeetings(w http.ResponseWriter, r *http.Request) {
	meetings := h.conference.ListMeetings()
	
	response := make([]MeetingResponse, 0, len(meetings))
	for _, m := range meetings {
		participants := make([]ParticipantInfo, 0)
		for _, p := range m.GetParticipantList() {
			participants = append(participants, ParticipantInfo{
				AgentID:  p.AgentID,
				Name:     p.Name,
				Avatar:   p.Avatar,
				Role:     string(p.Role),
				IsOnline: p.IsOnline,
			})
		}
		
		response = append(response, MeetingResponse{
			ID:           m.ID,
			Topic:        m.Topic,
			Description:  m.Description,
			Status:       string(m.Status),
			CreatedAt:    m.CreatedAt,
			StartedAt:    m.StartedAt,
			Facilitator:  m.Facilitator,
			Participants: participants,
			Agenda:       m.Agenda,
			MessageCount: len(m.Messages),
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandler) createMeeting(w http.ResponseWriter, r *http.Request) {
	var req MeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}
	
	if req.Facilitator == "" {
		req.Facilitator = "jarvis"
	}
	
	config := MeetingConfig{
		Topic:       req.Topic,
		Description: req.Description,
		Facilitator: req.Facilitator,
		Agenda:      req.Agenda,
		Timeout:     30 * time.Minute,
	}
	
	meeting, err := h.conference.CreateMeeting(config, req.Participants)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create meeting: %v", err), http.StatusInternalServerError)
		return
	}
	
	if req.AutoStart {
		meeting.Start()
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      meeting.ID,
		"topic":   meeting.Topic,
		"status":  string(meeting.Status),
		"message": "Meeting created successfully",
	})
}

func (h *APIHandler) handleMeetingDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/meetings/")
	if id == "" {
		http.Error(w, "Meeting ID required", http.StatusBadRequest)
		return
	}
	
	meeting, ok := h.conference.GetMeeting(id)
	if !ok {
		http.Error(w, "Meeting not found", http.StatusNotFound)
		return
	}
	
	switch r.Method {
	case http.MethodGet:
		h.getMeeting(w, meeting)
	case http.MethodPost:
		h.handleMeetingAction(w, r, meeting)
	case http.MethodDelete:
		h.deleteMeeting(w, meeting)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) getMeeting(w http.ResponseWriter, meeting *Meeting) {
	participants := make([]ParticipantInfo, 0)
	for _, p := range meeting.GetParticipantList() {
		participants = append(participants, ParticipantInfo{
			AgentID:  p.AgentID,
			Name:     p.Name,
			Avatar:   p.Avatar,
			Role:     string(p.Role),
			IsOnline: p.IsOnline,
		})
	}
	
	messages := make([]map[string]interface{}, 0)
	for _, m := range meeting.Messages {
		messages = append(messages, map[string]interface{}{
			"id":        m.ID,
			"from":      m.FromAgent,
			"content":   m.Content,
			"type":      m.Type,
			"timestamp": m.Timestamp,
		})
	}
	
	response := map[string]interface{}{
		"id":           meeting.ID,
		"topic":        meeting.Topic,
		"description":  meeting.Description,
		"status":       meeting.Status,
		"created_at":   meeting.CreatedAt,
		"started_at":   meeting.StartedAt,
		"ended_at":     meeting.EndedAt,
		"facilitator":  meeting.Facilitator,
		"participants": participants,
		"agenda":       meeting.Agenda,
		"messages":     messages,
		"summary":      meeting.Summary,
		"consensus":    meeting.Consensus,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandler) handleMeetingAction(w http.ResponseWriter, r *http.Request, meeting *Meeting) {
	var action struct {
		Action  string `json:"action"`
		Content string `json:"content,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	switch action.Action {
	case "start":
		if err := meeting.Start(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
	case "post":
		var req struct {
			Agent   string   `json:"agent"`
			Content string   `json:"content"`
			Type    string   `json:"type"`
			Mentions []string `json:"mentions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if _, err := meeting.PostMessage(req.Agent, req.Content, req.Type, req.Mentions); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
	case "end":
		var req struct {
			Summary   string `json:"summary"`
			Consensus *bool  `json:"consensus"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			req.Summary = "Meeting concluded"
		}
		meeting.End(req.Summary, req.Consensus)
		
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Action completed",
	})
}

func (h *APIHandler) deleteMeeting(w http.ResponseWriter, meeting *Meeting) {
	// Mark as cancelled
	meeting.Status = MeetingStatusCancelled
	w.WriteHeader(http.StatusNoContent)
}

// ScheduleRequest represents a schedule creation request
type ScheduleRequest struct {
	Topic        string   `json:"topic"`
	Description  string   `json:"description"`
	ScheduledAt  string   `json:"scheduled_at"` // ISO 8601
	Participants []string `json:"participants"`
	Facilitator  string   `json:"facilitator"`
	Agenda       []string `json:"agenda"`
	Reminder     string   `json:"reminder"` // e.g., "15m", "1h"
	AutoStart    bool     `json:"auto_start"`
}

func (h *APIHandler) handleSchedule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listSchedules(w, r)
	case http.MethodPost:
		h.createSchedule(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) listSchedules(w http.ResponseWriter, r *http.Request) {
	schedules := h.scheduler.ListSchedules()
	
	response := make([]map[string]interface{}, 0, len(schedules))
	for _, s := range schedules {
		response = append(response, map[string]interface{}{
			"id":           s.ID,
			"topic":        s.Topic,
			"description":  s.Description,
			"scheduled_at": s.ScheduledAt,
			"status":       s.Status,
			"participants": s.Participants,
			"facilitator":  s.Facilitator,
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandler) createSchedule(w http.ResponseWriter, r *http.Request) {
	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		http.Error(w, "Invalid scheduled_at format (use ISO 8601)", http.StatusBadRequest)
		return
	}
	
	var reminder time.Duration
	if req.Reminder != "" {
		reminder, _ = time.ParseDuration(req.Reminder)
	}
	
	if req.Facilitator == "" {
		req.Facilitator = "jarvis"
	}
	
	config := ScheduleConfig{
		Topic:        req.Topic,
		Description:  req.Description,
		ScheduledAt:  scheduledAt,
		Participants: req.Participants,
		Facilitator:  req.Facilitator,
		Agenda:       req.Agenda,
		Reminder:     reminder,
		AutoStart:    req.AutoStart,
	}
	
	schedule, err := h.scheduler.ScheduleMeeting(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      schedule.ID,
		"topic":   schedule.Topic,
		"status":  schedule.Status,
		"message": "Meeting scheduled successfully",
	})
}

func (h *APIHandler) handleScheduleDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/schedule/")
	id = strings.TrimSuffix(id, "/upcoming")
	if id == "" || id == "upcoming" {
		return
	}
	
	schedule, ok := h.scheduler.GetSchedule(id)
	if !ok {
		http.Error(w, "Schedule not found", http.StatusNotFound)
		return
	}
	
	if r.Method == http.MethodDelete {
		if err := h.scheduler.CancelSchedule(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

func (h *APIHandler) handleUpcoming(w http.ResponseWriter, r *http.Request) {
	within := 24 * time.Hour
	if d := r.URL.Query().Get("within"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			within = parsed
		}
	}
	
	upcoming := h.scheduler.GetUpcomingMeetings(within)
	
	response := make([]map[string]interface{}, 0, len(upcoming))
	for _, s := range upcoming {
		response = append(response, map[string]interface{}{
			"id":           s.ID,
			"topic":        s.Topic,
			"scheduled_at": s.ScheduledAt,
			"status":       s.Status,
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DiscussionRequest represents a request to start AI discussion
type DiscussionRequest struct {
	Topic        string   `json:"topic"`
	Context      string   `json:"context"`
	Participants []string `json:"participants"`
	MaxTurns     int      `json:"max_turns"`
}

func (h *APIHandler) handleDiscussions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if h.aiManager == nil {
		http.Error(w, "AI discussion not available (no LLM provider)", http.StatusServiceUnavailable)
		return
	}
	
	var req DiscussionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	if req.MaxTurns == 0 {
		req.MaxTurns = 3
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	disc, err := h.aiManager.StartAIDiscussion(ctx, req.Topic, req.Context, req.Participants, req.MaxTurns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       disc.ID,
		"topic":    disc.Topic,
		"status":   disc.Status,
		"max_turns": disc.MaxTurns,
		"message":  "AI discussion started",
	})
}

func (h *APIHandler) handleDiscussionDetail(w http.ResponseWriter, r *http.Request) {
	if h.aiManager == nil {
		http.Error(w, "AI discussion not available", http.StatusServiceUnavailable)
		return
	}
	
	id := strings.TrimPrefix(r.URL.Path, "/api/discussions/")
	
	disc, ok := h.aiManager.GetDiscussion(id)
	if !ok {
		http.Error(w, "Discussion not found", http.StatusNotFound)
		return
	}
	
	messages := make([]map[string]interface{}, 0, len(disc.Messages))
	for _, m := range disc.Messages {
		messages = append(messages, map[string]interface{}{
			"turn":      m.Turn,
			"agent_id":  m.AgentID,
			"name":      m.Name,
			"avatar":    m.Avatar,
			"content":   m.Content,
			"type":      m.Type,
			"timestamp": m.Timestamp,
		})
	}
	
	response := map[string]interface{}{
		"id":           disc.ID,
		"topic":        disc.Topic,
		"context":      disc.Context,
		"status":       disc.Status,
		"participants": disc.Participants,
		"max_turns":    disc.MaxTurns,
		"current_turn": disc.CurrentTurn,
		"messages":     messages,
		"consensus":    disc.Consensus,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandler) handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	agents := make([]map[string]interface{}, 0, len(h.agents))
	for _, agent := range h.agents {
		// Load persona summary
		agentDir := filepath.Join(h.personaDir, agent.ID)
		summary, _ := persona.GetPersonaSummary(agentDir)
		
		agents = append(agents, map[string]interface{}{
			"id":           agent.ID,
			"name":         agent.Name,
			"avatar":       agent.Avatar,
			"role":         agent.Role,
			"department":   agent.Department,
			"capabilities": agent.Capabilities,
			"is_coordinator": agent.IsCoordinator,
			"persona_summary": summary,
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func (h *APIHandler) handleAgentDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/agents/")
	
	var agent *config.AgentConfig
	for i := range h.agents {
		if h.agents[i].ID == id {
			agent = &h.agents[i]
			break
		}
	}
	
	if agent == nil {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}
	
	// Load full persona
	agentDir := filepath.Join(h.personaDir, agent.ID)
	personaFiles, _ := persona.LoadPersonaFiles(agentDir)
	
	response := map[string]interface{}{
		"id":             agent.ID,
		"name":           agent.Name,
		"avatar":         agent.Avatar,
		"role":           agent.Role,
		"department":     agent.Department,
		"capabilities":   agent.Capabilities,
		"responsibilities": agent.Responsibilities,
		"is_coordinator": agent.IsCoordinator,
		"is_permanent":   agent.IsPermanent,
		"identity":       personaFiles.Identity,
		"soul":           personaFiles.Soul,
		"memory":         personaFiles.Memory,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// WebSocket handling for real-time meeting updates
func (h *APIHandler) handleMeetingWebSocket(w http.ResponseWriter, r *http.Request) {
	// WebSocket upgrade would go here
	// For now, return info about the endpoint
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "info",
		"message": "WebSocket endpoint for real-time meeting updates (upgrade required)",
	})
}
