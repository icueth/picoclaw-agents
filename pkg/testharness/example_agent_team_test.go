// Example tests for Agent Team functionality
// ตัวอย่างการใช้งาน Test Harness สำหรับทดสอบ Agent Team

package testharness

import (
	"fmt"
	"testing"
)

// Example_agentDelegation ตัวอย่างการทดสอบการสั่งงานระหว่าง Agents
func Example_agentDelegation() {
	// สร้าง Mock Provider
	provider := NewMockProvider()
	
	// ตั้งค่าการตอบสนองเมื่อมีข้อความเกี่ยวกับ "Nova"
	provider.WithToolCallResponse(
		"Nova",
		"spawn_subagent",
		map[string]any{
			"role": "architect",
			"task": "ออกแบบระบบ",
		},
	)

	// สร้าง Test Harness
	harness := New(provider)
	
	// ส่งข้อความและรับผลลัพธ์
	response, _ := harness.Chat("ให้ Nova ออกแบบระบบหน่อย")
	
	// ตรวจสอบผลลัพธ์
	fmt.Println("Response:", response)
	fmt.Println("Tool called:", provider.VerifyToolCall("spawn_subagent"))
	
	// Output:
	// Response: I'll help you with that using spawn_subagent
	// Tool called: true
}

// Example_agentMeeting ตัวอย่างการทดสอบการประชุม
func Example_agentMeeting() {
	provider := NewMockProvider()
	
	// ตั้งค่าให้เรียก start_meeting เมื่อพูดถึงการประชุม
	provider.WithToolCallResponse(
		"ประชุม",
		"start_meeting",
		map[string]any{
			"topic":        "Sprint Planning",
			"participants": []string{"atlas", "nova", "clawed"},
			"duration":     30,
		},
	)

	harness := New(provider)
	harness.Chat("ให้จัดประชุม Sprint Planning กับ Atlas, Nova, และ Clawed")
	
	fmt.Println("Meeting scheduled:", provider.VerifyToolCall("start_meeting"))
	
	// Output:
	// Meeting scheduled: true
}

// Example_agentMailbox ตัวอย่างการทดสอบการส่งข้อความ
func Example_agentMailbox() {
	provider := NewMockProvider()
	
	provider.WithToolCallResponse(
		"ส่งข้อความ",
		"send_agent_message",
		map[string]any{
			"to":      "clawed",
			"subject": "Code Review",
			"message": "ช่วย review โค้ดหน่อย",
		},
	)

	harness := New(provider)
	harness.Chat("ส่งข้อความไปบอก Clawed ให้ review โค้ดหน่อย")
	
	fmt.Println("Message sent:", provider.VerifyToolCall("send_agent_message"))
	
	// Output:
	// Message sent: true
}

// Example_multiAgentWorkflow ตัวอย่างการทดสอบ workflow หลาย agents
func Example_multiAgentWorkflow() {
	provider := NewMockProvider()
	
	// ตั้งค่าให้เรียก spawn_subagent หลายครั้ง
	provider.WithToolCallResponse("Atlas", "spawn_subagent", map[string]any{"role": "researcher"})
	provider.WithToolCallResponse("Nova", "spawn_subagent", map[string]any{"role": "architect"})
	provider.WithToolCallResponse("Clawed", "spawn_subagent", map[string]any{"role": "coder"})

	harness := New(provider)
	harness.Chat("ให้ Atlas วิจัย, Nova ออกแบบ, Clawed เขียนโค้ด")
	
	count := provider.GetToolCallCount("spawn_subagent")
	fmt.Printf("Spawned %d subagents\n", count)
	
	// Output:
	// Spawned 1 subagents
}

// Example_agentTeamScenarios ตัวอย่างการรัน scenarios ทั้งหมด
func Example_agentTeamScenarios() {
	// รัน scenarios ทั้งหมด
	passed := 0
	failed := 0
	
	for _, scenario := range AgentTeamScenarios {
		provider := NewMockProvider()
		scenario.Setup(provider)
		harness := New(provider)
		
		if err := scenario.Test(harness); err != nil {
			fmt.Printf("❌ %s: %v\n", scenario.Name, err)
			failed++
		} else {
			fmt.Printf("✓ %s\n", scenario.Name)
			passed++
		}
	}
	
	fmt.Printf("\nResults: %d passed, %d failed\n", passed, failed)
}

// TestAgentTeamDocumentation ทดสอบเพื่อแสดงความสามารถของ Agent Team
func TestAgentTeamDocumentation(t *testing.T) {
	// สร้าง provider ที่ตอบสนองตามบทบาทต่างๆ
	provider := NewMockProvider()
	
	// แต่ละ agent มีบทบาทเฉพาะ
	roles := map[string]string{
		"Jarvis":   "coordinator",
		"Nova":     "architect",
		"Atlas":    "researcher",
		"Clawed":   "coder",
		"Sentinel": "qa",
		"Scribe":   "writer",
		"Trendy":   "reviewer",
		"Pixel":    "designer",
	}
	
	// ตั้งค่า provider ให้ตอบสนองตามชื่อ agent
	for name, role := range roles {
		provider.WithToolCallResponse(
			name,
			"spawn_subagent",
			map[string]any{"role": role},
		)
	}
	
	harness := New(provider)
	
	// ทดสอบเรียกแต่ละ agent
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{"Nova (architect)", "ให้ Nova ออกแบบ", "architect"},
		{"Atlas (researcher)", "ให้ Atlas วิจัย", "researcher"},
		{"Clawed (coder)", "ให้ Clawed เขียนโค้ด", "coder"},
		{"Sentinel (qa)", "ให้ Sentinel ตรวจสอบ", "qa"},
		{"Scribe (writer)", "ให้ Scribe เขียนเอกสาร", "writer"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := harness.Chat(tt.message)
			if err != nil {
				t.Errorf("Chat failed: %v", err)
			}
			
			// ตรวจสอบว่า spawn_subagent ถูกเรียก
			if !provider.VerifyToolCall("spawn_subagent") {
				t.Error("Expected spawn_subagent to be called")
			}
		})
	}
}

// TestAgentTeamCapabilities ทดสอบความสามารถหลักของ Agent Team
func TestAgentTeamCapabilities(t *testing.T) {
	// 1. Spawn Subagent
	t.Run("Spawn Subagent", func(t *testing.T) {
		provider := NewMockProvider()
		provider.WithToolCallResponse("spawn", "spawn_subagent", map[string]any{"role": "coder"})
		
		harness := New(provider)
		harness.Chat("spawn a coder")
		
		if !provider.VerifyToolCall("spawn_subagent") {
			t.Error("Failed to spawn subagent")
		}
	})
	
	// 2. Check Subagent Status
	t.Run("Check Subagent Status", func(t *testing.T) {
		provider := NewMockProvider()
		provider.WithToolCallResponse("status", "subagent_status", map[string]any{"action": "list"})
		
		harness := New(provider)
		harness.Chat("check subagent status")
		
		if !provider.VerifyToolCall("subagent_status") {
			t.Error("Failed to check subagent status")
		}
	})
	
	// 3. Start Meeting
	t.Run("Start Meeting", func(t *testing.T) {
		provider := NewMockProvider()
		provider.WithToolCallResponse("meeting", "start_meeting", map[string]any{
			"topic": "Test Meeting",
			"participants": []string{"atlas", "nova"},
		})
		
		harness := New(provider)
		harness.Chat("start a meeting")
		
		if !provider.VerifyToolCall("start_meeting") {
			t.Error("Failed to start meeting")
		}
	})
	
	// 4. Send Message
	t.Run("Send Message", func(t *testing.T) {
		provider := NewMockProvider()
		provider.WithToolCallResponse("message", "send_agent_message", map[string]any{
			"to":      "clawed",
			"subject": "Test",
			"message": "Hello",
		})
		
		harness := New(provider)
		harness.Chat("send message to clawed")
		
		if !provider.VerifyToolCall("send_agent_message") {
			t.Error("Failed to send message")
		}
	})
	
	// 5. Check Inbox
	t.Run("Check Inbox", func(t *testing.T) {
		provider := NewMockProvider()
		provider.WithToolCallResponse("inbox", "check_agent_inbox", map[string]any{"unread_only": true})
		
		harness := New(provider)
		harness.Chat("check inbox")
		
		if !provider.VerifyToolCall("check_agent_inbox") {
			t.Error("Failed to check inbox")
		}
	})
}
