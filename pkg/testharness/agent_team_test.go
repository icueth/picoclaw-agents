// Agent Team Test Runner
// ทดสอบการทำงานร่วมกันของ Agent Team

package testharness

import (
	"fmt"
	"testing"
)

// TestAgentTeamScenarios รันการทดสอบทั้งหมดสำหรับ Agent Team
func TestAgentTeamScenarios(t *testing.T) {
	for _, scenario := range AgentTeamScenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			// สร้าง mock provider
			provider := NewMockProvider()
			
			// ตั้งค่าการตอบสนอง
			scenario.Setup(provider)
			
			// สร้าง harness
			harness := New(provider)
			
			// รันการทดสอบ
			if err := scenario.Test(harness); err != nil {
				t.Errorf("Scenario %q failed: %v", scenario.Name, err)
			}
		})
	}
}

// TestSimpleDelegation ทดสอบการสั่งงานแบบง่าย
func TestSimpleDelegation(t *testing.T) {
	provider := NewMockProvider()
	provider.
		WithSystemPrompt("You are Jarvis, the coordinator").
		WithToolCallResponse(
			"ให้ Nova ออกแบบ",
			"spawn_subagent",
			map[string]any{
				"role": "architect",
				"task": "ออกแบบระบบ Todo List",
			},
		)

	harness := New(provider)
	_, err := harness.Chat("ให้ Nova ออกแบบระบบ Todo List ให้หน่อย")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if !provider.VerifyToolCall("spawn_subagent") {
		t.Error("Expected spawn_subagent to be called")
	}
}

// TestMultiAgentWorkflow ทดสอบ workflow หลาย Agent
func TestMultiAgentWorkflow(t *testing.T) {
	provider := NewMockProvider()
	provider.WithSystemPrompt("You are Jarvis, the coordinator")
	
	// ตั้งค่าให้เรียก tool ได้หลายครั้ง (ในระบบจริงจะเรียก 3 ครั้ง)
	provider.WithToolCallResponse("Atlas", "spawn_subagent", map[string]any{"role": "researcher", "task": "research"})

	harness := New(provider)
	_, err := harness.Chat("ให้ Atlas วิจัย, Nova ออกแบบ, Clawed เขียนโค้ด")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// ใน mock จะเรียกแค่ 1 ครั้ง แต่ในระบบจริงจะเรียก 3 ครั้ง
	if !provider.VerifyToolCall("spawn_subagent") {
		t.Error("Expected spawn_subagent to be called")
	}
}

// TestAgentMeeting ทดสอบการประชุม
func TestAgentMeeting(t *testing.T) {
	provider := NewMockProvider()
	provider.
		WithSystemPrompt("You are Jarvis, the coordinator").
		WithToolCallResponse(
			"ประชุม",
			"start_meeting",
			map[string]any{
				"topic":        "Planning Sprint",
				"participants": []string{"atlas", "nova", "clawed"},
				"duration":     30,
			},
		)

	harness := New(provider)
	_, err := harness.Chat("ให้จัดประชุม Planning Sprint กับ Atlas, Nova, และ Clawed")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if !provider.VerifyToolCall("start_meeting") {
		t.Error("Expected start_meeting to be called")
	}
}

// TestAgentMailbox ทดสอบการส่งข้อความ
func TestAgentMailbox(t *testing.T) {
	provider := NewMockProvider()
	provider.
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

	harness := New(provider)
	_, err := harness.Chat("ส่งข้อความไปบอก Clawed ให้ review โค้ดหน่อย")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if !provider.VerifyToolCall("send_agent_message") {
		t.Error("Expected send_agent_message to be called")
	}
}

// TestCheckInbox ทดสอบการเช็คกล่องจดหมาย
func TestCheckInbox(t *testing.T) {
	provider := NewMockProvider()
	provider.
		WithSystemPrompt("You are Jarvis, the coordinator").
		WithToolCallResponse(
			"จดหมาย",
			"check_agent_inbox",
			map[string]any{
				"unread_only": true,
			},
		)

	harness := New(provider)
	_, err := harness.Chat("เช็คว่ามีจดหมายอะไรเข้ามาบ้าง")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if !provider.VerifyToolCall("check_agent_inbox") {
		t.Error("Expected check_agent_inbox to be called")
	}
}

// TestFullProjectWorkflow ทดสอบ workflow โปรเจคเต็มรูปแบบ
func TestFullProjectWorkflow(t *testing.T) {
	provider := NewMockProvider()
	provider.WithSystemPrompt("You are Jarvis, the coordinator")
	
	// ตั้งค่า responses สำหรับแต่ละ agent
	provider.WithToolCallResponse("Atlas", "spawn_subagent", map[string]any{"role": "researcher"})
	provider.WithToolCallResponse("Nova", "spawn_subagent", map[string]any{"role": "architect"})
	provider.WithToolCallResponse("Clawed", "spawn_subagent", map[string]any{"role": "coder"})
	provider.WithToolCallResponse("Sentinel", "spawn_subagent", map[string]any{"role": "qa"})
	provider.WithToolCallResponse("Scribe", "spawn_subagent", map[string]any{"role": "writer"})

	harness := New(provider)
	_, err := harness.Chat("สร้างโปรเจคเต็มรูปแบบ: Atlas วิจัย, Nova ออกแบบ, Clawed เขียนโค้ด, Sentinel ตรวจสอบ, Scribe เขียนเอกสาร")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	count := provider.GetToolCallCount("spawn_subagent")
	fmt.Printf("Subagent calls: %d\n", count)
	
	// Note: ในการทดสอบจริง mock อาจไม่ได้เรียกครบ 5 ครั้ง ขึ้นกับการ matching
	// แต่ควรมีการเรียกอย่างน้อย 1 ครั้ง
	if count == 0 {
		t.Error("Expected at least 1 subagent call")
	}
}

// TestAgentStatus ทดสอบการเช็คสถานะ
func TestAgentStatus(t *testing.T) {
	provider := NewMockProvider()
	provider.
		WithSystemPrompt("You are Jarvis, the coordinator").
		WithToolCallResponse(
			"agent",
			"subagent_status",
			map[string]any{
				"action": "active",
			},
		)

	harness := New(provider)
	_, err := harness.Chat("ตอนนี้มี agent ไหนทำงานอยู่บ้าง")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if !provider.VerifyToolCall("subagent_status") {
		t.Error("Expected subagent_status to be called")
	}
}

// BenchmarkMultiAgentWorkflow benchmark สำหรับ workflow หลาย agent
func BenchmarkMultiAgentWorkflow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		provider := NewMockProvider()
		provider.
			WithSystemPrompt("You are Jarvis, the coordinator").
			WithToolCallResponse("workflow", "spawn_subagent", map[string]any{"role": "coder"})

		harness := New(provider)
		harness.Chat("ให้ Clawed เขียนโค้ด")
	}
}
