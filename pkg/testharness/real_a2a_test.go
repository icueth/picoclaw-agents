// Real Agent-to-Agent (A2A) Integration Test
// ทดสอบ A2A Communication กับระบบจริง (ใช้ Agent Registry, Mailbox, Messenger)

package testharness

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"picoclaw/agent/pkg/agent"
	"picoclaw/agent/pkg/agentcomm"
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/mailbox"
	"picoclaw/agent/pkg/tools"
)

// TestRealA2ADelegate ทดสอบการ delegate งานจาก Jarvis ไปยัง agent อื่นจริงๆ
func TestRealA2ADelegate(t *testing.T) {
	configPath := os.ExpandEnv("${HOME}/.picoclaw/config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Config file not found, skipping real A2A test")
	}

	// โหลด config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// สร้าง provider จริง
	provider, modelName, err := createRealProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	fmt.Printf("Using model: %s\n", modelName)

	// สร้าง Agent Registry (โหลด agents จาก config จริง)
	registry := agent.NewAgentRegistry(cfg, provider)
	
	// ตรวจสอบว่ามี agents ถูกโหลด
	agentIDs := registry.ListAgentIDs()
	fmt.Printf("Registered agents: %v\n", agentIDs)
	
	if len(agentIDs) == 0 {
		t.Fatal("No agents registered")
	}

	// หา default agent (Jarvis)
	jarvis, ok := registry.GetAgent("jarvis")
	if !ok {
		// ถ้าไม่มี jarvis ใช้ agent แรกที่พบ
		jarvis, _ = registry.GetAgent(agentIDs[0])
		fmt.Printf("Using fallback agent: %s\n", agentIDs[0])
	} else {
		fmt.Println("Found Jarvis agent")
	}

	// ตรวจสอบว่า Jarvis สามารถ spawn subagent ได้หรือไม่
	for _, targetID := range agentIDs {
		canSpawn := registry.CanSpawnSubagent(jarvis.ID, targetID)
		fmt.Printf("Jarvis can spawn %s: %v\n", targetID, canSpawn)
	}

	fmt.Println("✅ A2A Delegate test setup passed!")
}

// TestRealMailboxCommunication ทดสอบการส่งข้อความผ่าน Mailbox
func TestRealMailboxCommunication(t *testing.T) {
	// สร้าง mailbox สำหรับสอง agents
	_ = mailbox.NewMailbox("jarvis", 100) // jarvis mailbox (not used directly)
	clawedMailbox := mailbox.NewMailbox("clawed", 100)

	// สร้างข้อความจาก Jarvis ถึง Clawed
	msg := mailbox.Message{
		ID:        "msg-001",
		Type:      mailbox.MessageTypeTask,
		From:      "jarvis",
		To:        "clawed",
		Priority:  mailbox.PriorityHigh,
		Content:   "Please write a function to calculate factorial",
		CreatedAt: time.Now(),
	}

	// ส่งข้อความไปยัง Clawed's mailbox
	err := clawedMailbox.Send(msg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// ตรวจสอบว่า Clawed ได้รับข้อความ
	received, err := clawedMailbox.Receive()
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if received.From != "jarvis" {
		t.Errorf("Expected sender 'jarvis', got '%s'", received.From)
	}

	if received.Content != msg.Content {
		t.Errorf("Expected content '%s', got '%s'", msg.Content, received.Content)
	}

	fmt.Printf("✅ Message received: From=%s, Content=%s\n", received.From, received.Content)

	// ทดสอบ unread count
	count := clawedMailbox.GetUnreadCount()
	if count != 0 {
		t.Errorf("Expected 0 unread after receive, got %d", count)
	}
}

// TestRealMessengerCommunication ทดสอบการส่งข้อความผ่าน Messenger
func TestRealMessengerCommunication(t *testing.T) {
	// สร้าง shared context และ message bus (ใช้ agent.SharedContext ไม่ใช่ agentcomm)
	sharedCtx := agent.NewSharedContext(100, 1000)
	msgBus := bus.NewMessageBus()

	// สร้าง messenger สองตัว
	jarvisMessenger := agent.NewMessenger("jarvis", sharedCtx, msgBus)
	clawedMessenger := agent.NewMessenger("clawed", sharedCtx, msgBus)

	// ลงทะเบียน handler สำหรับ Clawed
	messageReceived := make(chan string, 1)
	clawedMessenger.RegisterHandler("clawed", func(ctx context.Context, msg *agentcomm.AgentMessage) {
		if payload, ok := msg.Payload.(string); ok {
			messageReceived <- payload
		} else {
			messageReceived <- fmt.Sprintf("%v", msg.Payload)
		}
	})

	// สร้างข้อความจาก Jarvis (ใช้ agentcomm.AgentMessage โดยตรง)
	msg := agentcomm.AgentMessage{
		From:      "jarvis",
		To:        "clawed",
		Type:      agentcomm.MsgRequest,
		Payload:   "Please help me review this code",
		SessionID: "test-session",
		Timestamp: time.Now().UnixMilli(),
	}

	// ส่งข้อความ
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// เริ่ม goroutine จำลองการดึงข้อความจาก MessageBus (เหมือนที่ Agent Loop ทำ)
	go func() {
		for {
			busMsg, ok := msgBus.ConsumeInbound(ctx)
			if !ok {
				break
			}
			var receivedMsg agentcomm.AgentMessage
			if err := json.Unmarshal([]byte(busMsg.Content), &receivedMsg); err == nil {
				if receivedMsg.To == "clawed" {
					if payload, ok := receivedMsg.Payload.(string); ok {
						messageReceived <- payload
					} else {
						messageReceived <- fmt.Sprintf("%v", receivedMsg.Payload)
					}
				}
			}
		}
	}()

	err := jarvisMessenger.SendDirect(ctx, "clawed", msg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// รอรับข้อความ
	select {
	case content := <-messageReceived:
		fmt.Printf("✅ Message received by Clawed: %s\n", content)
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

// TestRealA2AWithDelegationTool ทดสอบการใช้ Delegate Tool จริง
func TestRealA2AWithDelegationTool(t *testing.T) {
	configPath := os.ExpandEnv("${HOME}/.picoclaw/config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Config file not found, skipping real A2A test")
	}

	// โหลด config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// สร้าง provider จริง
	provider, modelName, err := createRealProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	fmt.Printf("Using model: %s\n", modelName)

	// สร้าง Agent Registry
	registry := agent.NewAgentRegistry(cfg, provider)
	
	agentIDs := registry.ListAgentIDs()
	fmt.Printf("Available agents: %v\n", agentIDs)

	// สร้าง subagent manager สำหรับรันงาน
	workspace := "/tmp/test-a2a-delegate"
	os.MkdirAll(workspace, 0755)
	
	msgBus := bus.NewMessageBus()
	manager := tools.NewSubagentManager(provider, modelName, workspace, msgBus)

	// สร้าง delegate tool
	delegateTool := tools.NewDelegateTool(registry, manager)

	// ทดสอบ list agents ก่อน
	listResult := delegateTool.Execute(context.Background(), map[string]any{
		"verbose": true,
	})
	
	fmt.Printf("List agents result:\n%s\n", listResult.ForLLM)

	// ถ้ามี agents หลายตัว ให้ลอง delegate
	if len(agentIDs) >= 2 {
		targetAgent := agentIDs[1] // agent ตัวที่สอง
		
		fmt.Printf("Delegating task to %s...\n", targetAgent)
		
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		
		result := delegateTool.Execute(ctx, map[string]any{
			"task":         "Write a simple hello world program in Go",
			"target_agent": targetAgent,
		})

		fmt.Printf("Delegate result:\n%s\n", result.ForLLM)
		
		if result.IsError {
			t.Logf("Delegation failed (expected for some agents): %s", result.ForLLM)
		} else {
			fmt.Println("✅ Delegation successful!")
		}
	}
}

// TestRealAgentRegistry ทดสอบ Agent Registry จริง
func TestRealAgentRegistry(t *testing.T) {
	configPath := os.ExpandEnv("${HOME}/.picoclaw/config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Config file not found, skipping real A2A test")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	provider, _, err := createRealProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// สร้าง registry
	registry := agent.NewAgentRegistry(cfg, provider)

	// ทดสอบ ListAgentIDs
	agentIDs := registry.ListAgentIDs()
	fmt.Printf("Total agents: %d\n", len(agentIDs))
	
	for _, id := range agentIDs {
		agent, ok := registry.GetAgent(id)
		if !ok {
			t.Errorf("Failed to get agent %s", id)
			continue
		}
		
		fmt.Printf("- %s: model=%s, workspace=%s\n", 
			id, agent.Model, agent.Workspace)
	}

	// ทดสอบ GetDefaultAgent
	defaultAgent := registry.GetDefaultAgent()
	if defaultAgent == nil {
		t.Log("No default agent found")
	} else {
		fmt.Printf("Default agent: %s\n", defaultAgent.ID)
	}

	// ทดสอบ CanSpawnSubagent
	if len(agentIDs) >= 2 {
		canSpawn := registry.CanSpawnSubagent(agentIDs[0], agentIDs[1])
		fmt.Printf("Can %s spawn %s: %v\n", agentIDs[0], agentIDs[1], canSpawn)
	}
}

// TestRealSharedContext ทดสอบ Shared Context สำหรับ A2A
func TestRealSharedContext(t *testing.T) {
	// สร้าง shared context
	sharedCtx := agentcomm.NewSharedContext(100, 1000)

	// เซ็ตค่า
	sharedCtx.Set("project_name", "PicoClaw")
	sharedCtx.Set("current_task", "A2A Testing")

	// อ่านค่า
	if val, ok := sharedCtx.Get("project_name"); ok {
		fmt.Printf("Project name: %s\n", val)
	}

	// เพิ่ม message log
	sharedCtx.AddMessageLog("jarvis", "clawed", "task", "Please write a function")
	sharedCtx.AddMessageLog("clawed", "jarvis", "answer", "Function completed")

	// อ่าน message log
	logs := sharedCtx.GetMessageLog()
	fmt.Printf("Message logs (%d entries):\n", len(logs))
	
	for _, log := range logs {
		fmt.Printf("- %s -> %s: %s\n", log.From, log.To, log.Content)
	}

	// ทดสอบ GetAll
	all := sharedCtx.GetAll()
	fmt.Printf("All context keys: %d\n", len(all))
}

// TestFullA2AWorkflow ทดสอบ A2A workflow เต็มรูปแบบ
func TestFullA2AWorkflow(t *testing.T) {
	configPath := os.ExpandEnv("${HOME}/.picoclaw/config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Config file not found, skipping real A2A test")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	provider, modelName, err := createRealProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	fmt.Printf("=== Full A2A Workflow Test ===\n")
	fmt.Printf("Model: %s\n\n", modelName)

	// 1. สร้าง Agent Registry
	fmt.Println("1. Creating Agent Registry...")
	registry := agent.NewAgentRegistry(cfg, provider)
	agentIDs := registry.ListAgentIDs()
	fmt.Printf("   Agents: %v\n\n", agentIDs)

	// 2. สร้าง Shared Context และ Message Bus
	fmt.Println("2. Setting up Shared Context and Message Bus...")
	sharedCtx := agent.NewSharedContext(100, 1000)
	msgBus := bus.NewMessageBus()
	fmt.Println("   Setup complete")

	// 3. สร้าง Mailboxes สำหรับทุก agents
	fmt.Println("3. Creating Mailboxes...")
	mailboxes := make(map[string]*mailbox.Mailbox)
	for _, id := range agentIDs {
		mailboxes[id] = mailbox.NewMailbox(id, 100)
		fmt.Printf("   Mailbox for %s created\n", id)
	}
	fmt.Println()

	// 4. สร้าง Messengers
	fmt.Println("4. Creating Messengers...")
	messengers := make(map[string]*agent.Messenger)
	for _, id := range agentIDs {
		messengers[id] = agent.NewMessenger(id, sharedCtx, msgBus)
		fmt.Printf("   Messenger for %s created\n", id)
	}
	fmt.Println()

	// 5. ทดสอบส่งข้อความระหว่าง agents
	if len(agentIDs) >= 2 {
		from := agentIDs[0]
		to := agentIDs[1]
		
		fmt.Printf("5. Testing message from %s to %s...\n", from, to)
		
		msg := mailbox.Message{
			ID:        "test-msg-001",
			Type:      mailbox.MessageTypeTask,
			From:      from,
			To:        to,
			Priority:  mailbox.PriorityNormal,
			Content:   "Please help me test the A2A system",
			CreatedAt: time.Now(),
		}
		
		err := mailboxes[to].Send(msg)
		if err != nil {
			t.Errorf("Failed to send message: %v", err)
		} else {
			fmt.Println("   Message sent successfully")
		}
		
		// ตรวจสอบว่าได้รับ
		if received, err := mailboxes[to].Receive(); err == nil {
			fmt.Printf("   Message received: %s\n", received.Content)
		}
		fmt.Println()
	}

	// 6. ทดสอบ Shared Context
	fmt.Println("6. Testing Shared Context...")
	sharedCtx.Set("test_key", "test_value")
	if val, ok := sharedCtx.Get("test_key"); ok {
		fmt.Printf("   Shared context value: %s\n", val)
	}
	fmt.Println()

	fmt.Println("✅ Full A2A Workflow Test completed!")
}
