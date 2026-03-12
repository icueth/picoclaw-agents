// PicoClaw - Visual UI Test Scenarios
// Test scenarios for visual elements: Agent Movement, Status Change, Room Update, Task Board Updates, XP Changes

package testharness

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/providers"
)

// VisualEvent represents a visual event in the UI
type VisualEvent struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// VisualScenario represents a test scenario for visual UI elements
type VisualScenario struct {
	Name        string
	Description string
	EventType   string
	Setup       func(*MockProvider)
	Test        func(*Harness) error
	Validate    func(*Harness) error
}

// VisualScenarios contains all visual UI test scenarios
var VisualScenarios = []VisualScenario{
	// ============================================
	// Agent Movement Tests
	// ============================================
	{
		Name:        "Agent Movement - Move to Room",
		Description: "Test visual feedback when agent moves to a different room",
		EventType:   "agent_move",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("move to", "move_agent", map[string]any{
					"agent_id":    "agent-001",
					"agent_name":  "Engineer Bot",
					"from_room":   "lobby",
					"to_room":     "engineering-lab",
					"animation":   "walk",
					"duration_ms": 500,
				}).
				WithResponsePattern("moved", "Agent Engineer Bot moved to Engineering Lab")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Move Engineer Bot to the engineering lab")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("move_agent")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("moved")
		},
	},
	{
		Name:        "Agent Movement - Animated Path",
		Description: "Test agent movement with animated path visualization",
		EventType:   "agent_move_animated",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("animate", "move_agent_animated", map[string]any{
					"agent_id":   "agent-002",
					"path":       []string{"room-a", "hallway", "room-b"},
					"waypoints":  3,
					"speed":      "normal",
					"show_trail": true,
				}).
				WithResponsePattern("animating", "Agent movement animation started")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Animate agent movement from room A to room B through hallway")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("move_agent_animated")
		},
	},
	{
		Name:        "Agent Movement - Teleport",
		Description: "Test instant agent teleportation between rooms",
		EventType:   "agent_teleport",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("teleport", "teleport_agent", map[string]any{
					"agent_id": "agent-003",
					"from":     "engineering",
					"to":       "qa-lab",
					"instant":  true,
					"effect":   "fade",
				}).
				WithResponsePattern("teleported", "Agent teleported instantly")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Teleport the agent to QA lab instantly")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("teleport_agent")
		},
	},
	{
		Name:        "Agent Movement - Group Movement",
		Description: "Test multiple agents moving together",
		EventType:   "agent_group_move",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithMultiToolCallResponse("group move", []providers.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Name: "move_agent",
						Function: &providers.FunctionCall{
							Name:      "move_agent",
							Arguments: `{"agent_id": "agent-001", "to_room": "conference"}`,
						},
					},
					{
						ID:   "call_2",
						Type: "function",
						Name: "move_agent",
						Function: &providers.FunctionCall{
							Name:      "move_agent",
							Arguments: `{"agent_id": "agent-002", "to_room": "conference"}`,
						},
					},
				}).
				WithResponsePattern("meeting", "All agents moved to conference room")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Move all agents to the conference room for a meeting")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("move_agent")
		},
	},

	// ============================================
	// Status Change Tests
	// ============================================
	{
		Name:        "Status Change - Agent Online",
		Description: "Test visual indicator when agent comes online",
		EventType:   "status_online",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("online", "update_agent_status", map[string]any{
					"agent_id":  "agent-001",
					"status":    "online",
					"indicator": "green",
					"badge":     "active",
					"pulse":     true,
				}).
				WithResponsePattern("online", "Agent is now online")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Set agent status to online")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("update_agent_status")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("online")
		},
	},
	{
		Name:        "Status Change - Agent Busy",
		Description: "Test visual indicator when agent becomes busy",
		EventType:   "status_busy",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("busy", "update_agent_status", map[string]any{
					"agent_id":     "agent-002",
					"status":       "busy",
					"indicator":    "yellow",
					"badge":        "working",
					"progress_bar": true,
					"task_name":    "Code Review",
				}).
				WithResponsePattern("busy", "Agent is now busy with a task")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Set agent status to busy with code review")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("update_agent_status")
		},
	},
	{
		Name:        "Status Change - Agent Offline",
		Description: "Test visual indicator when agent goes offline",
		EventType:   "status_offline",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("offline", "update_agent_status", map[string]any{
					"agent_id":   "agent-003",
					"status":     "offline",
					"indicator":  "gray",
					"badge":      "disconnected",
					"dim_avatar": true,
				}).
				WithResponsePattern("offline", "Agent has gone offline")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Set agent status to offline")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("update_agent_status")
		},
	},
	{
		Name:        "Status Change - Task Progress",
		Description: "Test visual progress indicator for task completion",
		EventType:   "status_progress",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("progress", "update_task_progress", map[string]any{
					"task_id":      "task-001",
					"progress":     75,
					"status":       "in_progress",
					"progress_bar": true,
					"percentage":   "75%",
					"color":        "blue",
				}).
				WithResponsePattern("75%", "Task progress updated to 75%")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Update task progress to 75%")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("update_task_progress")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("75%")
		},
	},
	{
		Name:        "Status Change - Error State",
		Description: "Test visual error indicator",
		EventType:   "status_error",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("error", "update_agent_status", map[string]any{
					"agent_id":      "agent-004",
					"status":        "error",
					"indicator":     "red",
					"badge":         "error",
					"alert":         true,
					"error_message": "Connection timeout",
				}).
				WithResponsePattern("error", "Agent error state displayed")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Show error state for agent connection timeout")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("update_agent_status")
		},
	},

	// ============================================
	// Room Update Tests
	// ============================================
	{
		Name:        "Room Update - Create Room",
		Description: "Test visual creation of a new room",
		EventType:   "room_create",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("create room", "create_room", map[string]any{
					"room_id":    "room-new-001",
					"room_name":  "AI Research Lab",
					"department": "research",
					"capacity":   5,
					"theme":      "futuristic",
					"icon":       "microscope",
				}).
				WithResponsePattern("created", "Room 'AI Research Lab' created")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Create a new room called AI Research Lab")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("create_room")
		},
	},
	{
		Name:        "Room Update - Update Room Theme",
		Description: "Test updating room visual theme",
		EventType:   "room_theme_update",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("theme", "update_room_theme", map[string]any{
					"room_id":      "room-001",
					"theme":        "dark_mode",
					"color_scheme": "blue-gray",
					"background":   "gradient",
					"transition":   "fade",
				}).
				WithResponsePattern("theme", "Room theme updated to dark mode")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Update room theme to dark mode")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("update_room_theme")
		},
	},
	{
		Name:        "Room Update - Occupancy Change",
		Description: "Test visual update when room occupancy changes",
		EventType:   "room_occupancy",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("occupancy", "update_room_occupancy", map[string]any{
					"room_id":          "room-eng-001",
					"current_count":    4,
					"max_capacity":     5,
					"status":           "almost_full",
					"visual_indicator": "yellow",
				}).
				WithResponsePattern("occupancy", "Room occupancy updated")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Update room occupancy to 4 out of 5")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("update_room_occupancy")
		},
	},
	{
		Name:        "Room Update - Room Lock/Unlock",
		Description: "Test visual lock/unlock state of a room",
		EventType:   "room_lock",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("lock", "lock_room", map[string]any{
					"room_id":      "room-private-001",
					"locked":       true,
					"lock_icon":    "🔒",
					"access_level": "private",
				}).
				WithResponsePattern("locked", "Room is now locked")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Lock the private room")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("lock_room")
		},
	},

	// ============================================
	// Task Board Update Tests
	// ============================================
	{
		Name:        "Task Board - Add Task Card",
		Description: "Test adding a new task card to the board",
		EventType:   "board_add_task",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("add card", "add_task_card", map[string]any{
					"task_id":   "task-101",
					"title":     "Implement OAuth",
					"column":    "todo",
					"priority":  "high",
					"color":     "red",
					"assignee":  "agent-001",
					"animation": "slide_in",
				}).
				WithResponsePattern("added", "Task card added to board")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Add a new task card for OAuth implementation")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("add_task_card")
		},
	},
	{
		Name:        "Task Board - Move Task Card",
		Description: "Test moving a task card between columns",
		EventType:   "board_move_task",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("move card", "move_task_card", map[string]any{
					"task_id":     "task-101",
					"from_column": "todo",
					"to_column":   "in_progress",
					"animation":   "slide",
					"timestamp":   time.Now().Format(time.RFC3339),
				}).
				WithResponsePattern("moved", "Task card moved to in progress")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Move task card from todo to in progress")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("move_task_card")
		},
	},
	{
		Name:        "Task Board - Update Task Card",
		Description: "Test updating task card details",
		EventType:   "board_update_task",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("update card", "update_task_card", map[string]any{
					"task_id":   "task-101",
					"title":     "Implement OAuth (Updated)",
					"priority":  "critical",
					"color":     "purple",
					"highlight": true,
				}).
				WithResponsePattern("updated", "Task card updated")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Update task card priority to critical")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("update_task_card")
		},
	},
	{
		Name:        "Task Board - Archive Task Card",
		Description: "Test archiving a completed task card",
		EventType:   "board_archive_task",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("archive", "archive_task_card", map[string]any{
					"task_id":      "task-100",
					"from_column":  "done",
					"archive_date": time.Now().Format(time.RFC3339),
					"animation":    "fade_out",
				}).
				WithResponsePattern("archived", "Task card archived")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Archive the completed task")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("archive_task_card")
		},
	},
	{
		Name:        "Task Board - Column Reorder",
		Description: "Test reordering columns on the task board",
		EventType:   "board_reorder",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("reorder", "reorder_columns", map[string]any{
					"new_order": []string{"backlog", "todo", "in_progress", "review", "done"},
					"animation": "shuffle",
				}).
				WithResponsePattern("reordered", "Board columns reordered")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Reorder board columns with review before done")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("reorder_columns")
		},
	},

	// ============================================
	// XP Changes Tests
	// ============================================
	{
		Name:        "XP Change - Gain XP",
		Description: "Test visual XP gain animation",
		EventType:   "xp_gain",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("gain xp", "add_xp", map[string]any{
					"agent_id":  "agent-001",
					"xp_gained": 150,
					"total_xp":  1250,
					"level":     5,
					"animation": "float_up",
					"reason":    "Task completed",
				}).
				WithResponsePattern("gained", "Agent gained 150 XP")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Add 150 XP to agent for completing task")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("add_xp")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("150")
		},
	},
	{
		Name:        "XP Change - Level Up",
		Description: "Test level up animation and notification",
		EventType:   "xp_level_up",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("level up", "level_up", map[string]any{
					"agent_id":  "agent-001",
					"old_level": 4,
					"new_level": 5,
					"total_xp":  1000,
					"animation": "burst",
					"rewards":   []string{"New skill: Advanced Debugging", "Badge: Level 5 Master"},
				}).
				WithResponsePattern("level up", "🎉 Agent leveled up to level 5!")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Level up agent to level 5")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("level_up")
		},
		Validate: func(h *Harness) error {
			return h.AssertResponseContains("level 5")
		},
	},
	{
		Name:        "XP Change - XP Bar Update",
		Description: "Test XP progress bar visual update",
		EventType:   "xp_bar_update",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("xp bar", "update_xp_bar", map[string]any{
					"agent_id":     "agent-002",
					"current_xp":   750,
					"xp_to_next":   1000,
					"progress_pct": 75.0,
					"bar_color":    "gradient-blue",
				}).
				WithResponsePattern("75%", "XP bar updated to 75%")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Update XP bar to show 75% progress")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("update_xp_bar")
		},
	},
	{
		Name:        "XP Change - Achievement Unlocked",
		Description: "Test achievement unlock visual notification",
		EventType:   "xp_achievement",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("achievement", "unlock_achievement", map[string]any{
					"agent_id":       "agent-001",
					"achievement_id": "ach-first-100",
					"title":          "Century Club",
					"description":    "Complete 100 tasks",
					"icon":           "🏆",
					"rarity":         "gold",
					"animation":      "trophy_spin",
				}).
				WithResponsePattern("unlocked", "🏆 Achievement unlocked: Century Club")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Unlock Century Club achievement")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("unlock_achievement")
		},
	},
	{
		Name:        "XP Change - Leaderboard Update",
		Description: "Test leaderboard position change",
		EventType:   "xp_leaderboard",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithToolCallResponse("leaderboard", "update_leaderboard", map[string]any{
					"agent_id":     "agent-001",
					"old_position": 5,
					"new_position": 3,
					"total_xp":     2500,
					"animation":    "rank_up",
				}).
				WithResponsePattern("ranked", "Agent moved up to rank 3")
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("Update leaderboard - agent moved to rank 3")
			if err != nil {
				return err
			}
			return h.AssertToolCalled("update_leaderboard")
		},
	},
}

// RunVisualScenario executes a single visual scenario
func RunVisualScenario(scenario VisualScenario) (*ScenarioResult, error) {
	provider := NewMockProvider()
	scenario.Setup(provider)

	harness := New(provider)

	start := time.Now()
	err := scenario.Test(harness)
	duration := time.Since(start)

	result := &ScenarioResult{
		Name:     scenario.Name,
		Passed:   err == nil,
		Duration: duration,
		Error:    err,
	}

	if scenario.Validate != nil && err == nil {
		if valErr := scenario.Validate(harness); valErr != nil {
			result.Passed = false
			result.Error = valErr
		}
	}

	return result, nil
}

// RunVisualScenarios executes all visual scenarios
func RunVisualScenarios() []ScenarioResult {
	results := make([]ScenarioResult, 0, len(VisualScenarios))

	fmt.Printf("🎨 Running %d Visual UI Test Scenarios...\n\n", len(VisualScenarios))

	for _, scenario := range VisualScenarios {
		start := time.Now()
		result, _ := RunVisualScenario(scenario)
		duration := time.Since(start)
		result.Duration = duration

		results = append(results, *result)

		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
		}
		fmt.Printf("%s %s (%v)\n", status, scenario.Name, duration)
		if result.Error != nil {
			fmt.Printf("   Error: %v\n", result.Error)
		}
	}

	return results
}

// RunVisualScenariosByEventType executes scenarios filtered by event type
func RunVisualScenariosByEventType(eventType string) []ScenarioResult {
	var filtered []VisualScenario
	for _, s := range VisualScenarios {
		if s.EventType == eventType {
			filtered = append(filtered, s)
		}
	}

	results := make([]ScenarioResult, 0, len(filtered))
	fmt.Printf("🎨 Running %d visual scenarios for event type: %s\n\n", len(filtered), eventType)

	for _, scenario := range filtered {
		result, _ := RunVisualScenario(scenario)
		results = append(results, *result)

		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
		}
		fmt.Printf("%s %s\n", status, scenario.Name)
	}

	return results
}

// RunVisualScenariosByCategory executes scenarios by visual category
func RunVisualScenariosByCategory(category string) []ScenarioResult {
	categoryKeywords := map[string][]string{
		"movement": {"Agent Movement", "Move"},
		"status":   {"Status Change"},
		"room":     {"Room Update"},
		"board":    {"Task Board"},
		"xp":       {"XP Change"},
	}

	keywords, ok := categoryKeywords[category]
	if !ok {
		return []ScenarioResult{}
	}

	var filtered []VisualScenario
	for _, s := range VisualScenarios {
		for _, kw := range keywords {
			if strings.Contains(s.Name, kw) {
				filtered = append(filtered, s)
				break
			}
		}
	}

	results := make([]ScenarioResult, 0, len(filtered))
	fmt.Printf("🎨 Running %d visual scenarios for category: %s\n\n", len(filtered), category)

	for _, scenario := range filtered {
		result, _ := RunVisualScenario(scenario)
		results = append(results, *result)

		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
		}
		fmt.Printf("%s %s\n", status, scenario.Name)
	}

	return results
}

// PrintVisualReport prints a detailed report of visual scenario results
func PrintVisualReport(results []ScenarioResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎨 Visual UI Test Report")
	fmt.Println(strings.Repeat("=", 60))

	passed := 0
	failed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Printf("Total: %d | Passed: %d | Failed: %d\n", len(results), passed, failed)
	fmt.Println(strings.Repeat("=", 60))

	if failed > 0 {
		fmt.Println("\n❌ Failed Tests:")
		for _, r := range results {
			if !r.Passed {
				fmt.Printf("  • %s: %v\n", r.Name, r.Error)
			}
		}
	}
}

// VisualEventLog tracks visual events for testing
type VisualEventLog struct {
	events []VisualEvent
	mu     sync.RWMutex
}

// NewVisualEventLog creates a new visual event log
func NewVisualEventLog() *VisualEventLog {
	return &VisualEventLog{
		events: make([]VisualEvent, 0),
	}
}

// AddEvent adds a visual event to the log
func (l *VisualEventLog) AddEvent(eventType string, data map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.events = append(l.events, VisualEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	})
}

// GetEvents returns all events
func (l *VisualEventLog) GetEvents() []VisualEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	events := make([]VisualEvent, len(l.events))
	copy(events, l.events)
	return events
}

// GetEventsByType returns events of a specific type
func (l *VisualEventLog) GetEventsByType(eventType string) []VisualEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var filtered []VisualEvent
	for _, e := range l.events {
		if e.Type == eventType {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// Clear clears all events
func (l *VisualEventLog) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = make([]VisualEvent, 0)
}

// VisualTestHarness extends the base Harness with visual testing capabilities
type VisualTestHarness struct {
	*Harness
	EventLog *VisualEventLog
}

// NewVisualTestHarness creates a new visual test harness
func NewVisualTestHarness(provider *MockProvider) *VisualTestHarness {
	return &VisualTestHarness{
		Harness:  New(provider),
		EventLog: NewVisualEventLog(),
	}
}

// RecordEvent records a visual event
func (h *VisualTestHarness) RecordEvent(eventType string, data map[string]interface{}) {
	h.EventLog.AddEvent(eventType, data)
}

// AssertEventOccurred asserts that an event of a specific type occurred
func (h *VisualTestHarness) AssertEventOccurred(eventType string) error {
	events := h.EventLog.GetEventsByType(eventType)
	if len(events) == 0 {
		return fmt.Errorf("expected event of type %q to occur, but none found", eventType)
	}
	return nil
}

// AssertEventCount asserts the count of events of a specific type
func (h *VisualTestHarness) AssertEventCount(eventType string, expectedCount int) error {
	events := h.EventLog.GetEventsByType(eventType)
	if len(events) != expectedCount {
		return fmt.Errorf("expected %d events of type %q, got %d", expectedCount, eventType, len(events))
	}
	return nil
}

// VisualScenarioWithRealLLM represents a visual scenario for real LLM testing
type VisualScenarioWithRealLLM struct {
	Name        string
	Description string
	EventType   string
	Test        func(*RealLLMTestHarness) error
}

// RealLLMVisualScenarios contains visual scenarios for real LLM testing
var RealLLMVisualScenarios = []VisualScenarioWithRealLLM{
	{
		Name:        "Real LLM - Describe Visual Layout",
		Description: "Test real LLM can describe office visual layout",
		EventType:   "layout_description",
		Test: func(h *RealLLMTestHarness) error {
			response, err := h.Chat("Describe the visual layout of the engineering department room")
			if err != nil {
				return err
			}
			lower := strings.ToLower(response)
			if !strings.Contains(lower, "room") && !strings.Contains(lower, "layout") && !strings.Contains(lower, "visual") {
				return fmt.Errorf("expected visual layout description, got: %s", response)
			}
			return nil
		},
	},
	{
		Name:        "Real LLM - Status Change Description",
		Description: "Test real LLM can describe status change visuals",
		EventType:   "status_description",
		Test: func(h *RealLLMTestHarness) error {
			response, err := h.Chat("How would you visually indicate an agent is busy?")
			if err != nil {
				return err
			}
			lower := strings.ToLower(response)
			if !strings.Contains(lower, "color") && !strings.Contains(lower, "indicator") && !strings.Contains(lower, "status") {
				return fmt.Errorf("expected visual status indicators, got: %s", response)
			}
			return nil
		},
	},
	{
		Name:        "Real LLM - Task Board Layout",
		Description: "Test real LLM can describe task board layout",
		EventType:   "board_layout",
		Test: func(h *RealLLMTestHarness) error {
			response, err := h.Chat("Describe a kanban-style task board layout for the office")
			if err != nil {
				return err
			}
			lower := strings.ToLower(response)
			if !strings.Contains(lower, "column") && !strings.Contains(lower, "card") && !strings.Contains(lower, "board") {
				return fmt.Errorf("expected board layout description, got: %s", response)
			}
			return nil
		},
	},
}

// RunVisualScenarioWithRealLLM executes a visual scenario with real LLM
func RunVisualScenarioWithRealLLM(harness *RealLLMTestHarness, scenario VisualScenarioWithRealLLM) (*ScenarioResult, error) {
	start := time.Now()
	err := scenario.Test(harness)
	duration := time.Since(start)

	return &ScenarioResult{
		Name:        scenario.Name,
		Description: scenario.Description,
		Passed:      err == nil,
		Duration:    duration,
		Error:       err,
	}, nil
}

// GenerateVisualEventJSON generates JSON representation of a visual event
func GenerateVisualEventJSON(eventType string, data map[string]interface{}) (string, error) {
	event := VisualEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	jsonData, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// VisualEventTypes returns all available visual event types
func VisualEventTypes() []string {
	return []string{
		"agent_move",
		"agent_move_animated",
		"agent_teleport",
		"agent_group_move",
		"status_online",
		"status_busy",
		"status_offline",
		"status_progress",
		"status_error",
		"room_create",
		"room_theme_update",
		"room_occupancy",
		"room_lock",
		"board_add_task",
		"board_move_task",
		"board_update_task",
		"board_archive_task",
		"board_reorder",
		"xp_gain",
		"xp_level_up",
		"xp_bar_update",
		"xp_achievement",
		"xp_leaderboard",
	}
}
