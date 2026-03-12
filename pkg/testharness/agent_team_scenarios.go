// Agent Team Test Scenarios
// ทดสอบการทำงานร่วมกันของ Agent Team

package testharness

import (
	"fmt"
)

// AgentTeamScenarios สถานการณ์การทดสอบ Agent Team
var AgentTeamScenarios = []Scenario{
	{
		Name:        "Agent to Agent - Simple Delegation",
		Description: "ทดสอบการสั่งงานจาก Jarvis ไปยัง Nova",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithSystemPrompt("You are Jarvis, the coordinator").
				WithToolCallResponse(
					"Nova",
					"spawn_subagent",
					map[string]any{
						"role": "architect",
						"task": "ออกแบบระบบ Todo List",
					},
				)
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("ให้ Nova ออกแบบระบบ Todo List ให้หน่อย")
			if err != nil {
				return err
			}
			
			// ตรวจสอบว่า spawn_subagent ถูกเรียก
			if err := h.AssertToolCalled("spawn_subagent"); err != nil {
				return fmt.Errorf("spawn_subagent not called: %w", err)
			}
			
			return nil
		},
	},
	{
		Name:        "Jarvis Coordinator - Multi Agent",
		Description: "ทดสอบ Jarvis เรียกหลาย Agent ทำงานแล้วรวมผล",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithSystemPrompt("You are Jarvis, the coordinator").
				WithToolCallResponse(
					"Todo",
					"spawn_subagent",
					map[string]any{
						"role": "architect",
						"task": "ออกแบบและสร้างโปรเจค",
					},
				)
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("ช่วยสร้างโปรเจค Todo List โดยให้ Nova ออกแบบ, Clawed เขียนโค้ด, และ Sentinel ตรวจสอบ")
			if err != nil {
				return err
			}
			
			// ตรวจสอบว่า spawn_subagent ถูกเรียก (ในระบบจริงจะเรียก 3 ครั้ง)
			if err := h.AssertToolCalled("spawn_subagent"); err != nil {
				return err
			}
			
			return nil
		},
	},
	{
		Name:        "Agent Meeting - Discussion",
		Description: "ทดสอบการประชุมระหว่าง Agents",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithSystemPrompt("You are Jarvis, the coordinator").
				WithToolCallResponse(
					"ประชุม",
					"start_meeting",
					map[string]any{
						"topic":    "Planning Sprint",
						"participants": []string{"atlas", "nova", "clawed"},
						"duration": 30,
					},
				)
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("ให้จัดประชุม Planning Sprint กับ Atlas, Nova, และ Clawed")
			if err != nil {
				return err
			}
			
			return h.AssertToolCalled("start_meeting")
		},
	},
	{
		Name:        "Agent Mailbox - Send Message",
		Description: "ทดสอบการส่งข้อความระหว่าง Agents",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithSystemPrompt("You are Jarvis, the coordinator").
				WithToolCallResponse(
					"ส่งข้อความ",
					"send_agent_message",
					map[string]any{
						"to":      "clawed",
						"subject": "Review Code",
						"message": "ช่วย review โค้ดส่วนนี้ด้วย",
					},
				)
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("ส่งข้อความไปบอก Clawed ให้ review โค้ดหน่อย")
			if err != nil {
				return err
			}
			
			return h.AssertToolCalled("send_agent_message")
		},
	},
	{
		Name:        "Agent Mailbox - Check Inbox",
		Description: "ทดสอบการตรวจสอบกล่องจดหมาย",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithSystemPrompt("You are Jarvis, the coordinator").
				WithToolCallResponse(
					"จดหมาย",
					"check_agent_inbox",
					map[string]any{
						"unread_only": true,
					},
				)
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("เช็คว่ามีจดหมายอะไรเข้ามาบ้าง")
			if err != nil {
				return err
			}
			
			return h.AssertToolCalled("check_agent_inbox")
		},
	},
	{
		Name:        "Agent Collaboration - Research then Code",
		Description: "ทดสอบ Atlas วิจัยแล้วส่งต่อให้ Clawed เขียนโค้ด",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithSystemPrompt("You are Jarvis, the coordinator").
				WithToolCallResponse(
					"Atlas",
					"spawn_subagent",
					map[string]any{
						"role": "researcher",
						"task": "วิจัย API ที่ดีที่สุด",
					},
				)
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("ให้ Atlas วิจัย API ที่ดีที่สุดก่อน แล้วส่งต่อให้ Clawed เขียนโค้ด")
			if err != nil {
				return err
			}
			
			// ตรวจสอบว่า spawn_subagent ถูกเรียก
			return h.AssertToolCalled("spawn_subagent")
		},
	},
	{
		Name:        "Error - Agent Not Found",
		Description: "ทดสอบกรณีเรียก Agent ที่ไม่มีอยู่",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithSystemPrompt("You are Jarvis, the coordinator").
				WithToolError(
					"spawn_subagent",
					fmt.Errorf("agent with role 'unknown' not found"),
				)
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("ให้ UnknownAgent ทำงาน")
			
			// ควรจะได้ error
			if err == nil {
				return fmt.Errorf("expected error for unknown agent")
			}
			
			return nil
		},
	},
	{
		Name:        "Agent Status Check",
		Description: "ทดสอบการตรวจสอบสถานะ Agent",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithSystemPrompt("You are Jarvis, the coordinator").
				WithToolCallResponse(
					"agent",
					"subagent_status",
					map[string]any{
						"action": "active",
					},
				)
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("ตอนนี้มี agent ไหนทำงานอยู่บ้าง")
			if err != nil {
				return err
			}
			
			return h.AssertToolCalled("subagent_status")
		},
	},
	{
		Name:        "Complex Workflow - Full Project",
		Description: "ทดสอบ workflow ซับซ้อน: วิจัย->ออกแบบ->เขียนโค้ด->ตรวจสอบ->เขียนเอกสาร",
		Setup: func(mp *MockProvider) {
			mp.Reset().
				WithSystemPrompt("You are Jarvis, the coordinator").
				WithToolCallResponse(
					"โปรเจค",
					"spawn_subagent",
					map[string]any{
						"role": "coordinator",
						"task": "manage full project workflow",
					},
				)
		},
		Test: func(h *Harness) error {
			_, err := h.Chat("สร้างโปรเจคเต็มรูปแบบ: Atlas วิจัย, Nova ออกแบบ, Clawed เขียนโค้ด, Sentinel ตรวจสอบ, Scribe เขียนเอกสาร")
			if err != nil {
				return err
			}
			
			// ตรวจสอบว่า spawn_subagent ถูกเรียก
			return h.AssertToolCalled("spawn_subagent")
		},
	},
}

// Helper methods for the Harness

// GetToolCallCount returns the number of times a tool was called
func (h *Harness) GetToolCallCount(toolName string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	count := 0
	for _, turn := range h.conversation {
		for _, tc := range turn.ToolCalls {
			if tc.Name == toolName {
				count++
			}
		}
	}
	return count
}

// GetToolCalls returns all calls to a specific tool
func (h *Harness) GetToolCalls(toolName string) []ToolResult {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	var results []ToolResult
	for _, turn := range h.conversation {
		for _, tc := range turn.ToolCalls {
			if tc.Name == toolName {
				if result, ok := h.toolResults[tc.ID]; ok {
					results = append(results, result)
				}
			}
		}
	}
	return results
}

// AssertToolCallArgs checks if a tool was called with specific arguments
func (h *Harness) AssertToolCallArgs(toolName string, expectedArgs map[string]any) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for _, turn := range h.conversation {
		for _, tc := range turn.ToolCalls {
			if tc.Name == toolName {
				// Check if all expected args match
				for key, expectedVal := range expectedArgs {
					if actualVal, ok := tc.Arguments[key]; !ok {
						return fmt.Errorf("missing argument %s", key)
					} else if actualVal != expectedVal {
						return fmt.Errorf("argument %s mismatch: expected %v, got %v", key, expectedVal, actualVal)
					}
				}
				return nil
			}
		}
	}
	
	return fmt.Errorf("tool %s was not called", toolName)
}

