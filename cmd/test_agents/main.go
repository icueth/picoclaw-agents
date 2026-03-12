// Complete Agent System Test with Persona, AI Discussion, and Scheduler
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"picoclaw/agent/pkg/agent"
	"picoclaw/agent/pkg/agent/meeting"
	"picoclaw/agent/pkg/agent/persona"
	"picoclaw/agent/pkg/bootstrap"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/providers"
)

func main() {
	fmt.Println("🤖 Complete Agent System Test")
	fmt.Println("=====================================")

	// Load config
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".picoclaw", "config.json")
	
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return
	}
	
	builtinAgents := agent.GetBuiltinAgents()
	var testAgents []config.AgentConfig
	for _, ba := range builtinAgents {
		testAgents = append(testAgents, ba.ToAgentConfig(cfg.GetDepartmentModel(ba.Department)))
	}
	fmt.Printf("✅ Config loaded: %d built-in agents\n\n", len(testAgents))

	// Bootstrap system
	fmt.Println("📦 Bootstrapping system...")
	sys, err := bootstrap.Bootstrap(cfg)
	if err != nil {
		fmt.Printf("❌ Failed to bootstrap: %v\n", err)
		return
	}
	defer sys.Close()
	fmt.Printf("✅ System bootstrapped\n")

	// Create provider
	provider, _, _ := providers.CreateProvider(cfg)

	// Create AgentRegistry
	fmt.Println("\n🔧 Creating AgentRegistry...")
	_ = agent.NewAgentRegistry(cfg, provider)
	fmt.Println("   ✅ AgentRegistry created")

	// Test 1: Initialize Persona Files
	fmt.Println("\n📋 Test 1: Initialize Agent Persona Files")
	agentsDir := filepath.Join(homeDir, ".picoclaw", "agents")
	
	for i := range testAgents {
		agentCfg := testAgents[i]
		agentDir := filepath.Join(agentsDir, agentCfg.ID)
		if err := persona.EnsureAgentPersona(&agentCfg, agentsDir); err != nil {
			fmt.Printf("   ❌ %s: %v\n", agentCfg.ID, err)
		} else {
			// Check what files were created
			identityPath := filepath.Join(agentDir, "IDENTITY.md")
			soulPath := filepath.Join(agentDir, "SOUL.md")
			memoryPath := filepath.Join(agentDir, "MEMORY.md")
			
			identityExists := fileExists(identityPath)
			soulExists := fileExists(soulPath)
			memoryExists := fileExists(memoryPath)
			
			if identityExists && soulExists && memoryExists {
				fmt.Printf("   ✅ %s %s: IDENTITY.md, SOUL.md, MEMORY.md\n", agentCfg.Avatar, agentCfg.Name)
			}
		}
	}

	// Show sample persona file
	fmt.Println("\n📄 Sample: Jarvis IDENTITY.md")
	jarvisPersona, _ := persona.LoadPersonaFiles(filepath.Join(agentsDir, "jarvis"))
	if jarvisPersona != nil {
		lines := splitLines(jarvisPersona.Identity)
		for i, line := range lines {
			if i < 15 { // Show first 15 lines
				fmt.Printf("   %s\n", line)
			}
		}
		if len(lines) > 15 {
			fmt.Println("   ...")
		}
	}

	// Test 2: Conference Manager
	fmt.Println("\n📋 Test 2: Conference Manager Setup")
	conference := meeting.NewConferenceManager()
	conference.PopulateFromRegistry(testAgents)
	fmt.Printf("   ✅ ConferenceManager with %d agents\n", len(testAgents))

	// Test 3: Meeting Scheduler
	fmt.Println("\n📋 Test 3: Meeting Scheduler")
	scheduler := meeting.NewScheduler()
	defer scheduler.Stop()

	// Schedule a meeting in 5 minutes
	futureTime := time.Now().Add(5 * time.Minute)
	schedule, err := scheduler.ScheduleMeeting(meeting.ScheduleConfig{
		Topic:        "Weekly Team Sync",
		Description:  "Regular team synchronization meeting",
		ScheduledAt:  futureTime,
		Participants: []string{"atlas", "clawed", "nova", "pixel"},
		Facilitator:  "jarvis",
		Agenda:       []string{"Progress updates", "Blockers discussion", "Next week planning"},
		Reminder:     2 * time.Minute,
		AutoStart:    true,
	})
	
	if err != nil {
		fmt.Printf("   ❌ Failed to schedule: %v\n", err)
	} else {
		fmt.Printf("   ✅ Scheduled meeting: %s\n", schedule.ID)
		fmt.Printf("   📌 Topic: %s\n", schedule.Topic)
		fmt.Printf("   🕐 When: %s (in %s)\n", schedule.ScheduledAt.Format("15:04"), time.Until(schedule.ScheduledAt).Round(time.Second))
	}

	// Schedule a daily standup
	dailySchedule, _ := scheduler.ScheduleDaily("Daily Standup", 9, 0, []string{"jarvis", "atlas", "clawed"})
	fmt.Printf("   ✅ Daily Standup scheduled: %s at 09:00\n", dailySchedule.ID)

	// Show upcoming meetings
	fmt.Println("\n📅 Upcoming Meetings (next 24h):")
	upcoming := scheduler.GetUpcomingMeetings(24 * time.Hour)
	for _, m := range upcoming {
		fmt.Printf("   • %s at %s\n", m.Topic, m.ScheduledAt.Format("15:04"))
	}

	// Test 4: AI-Powered Discussion
	fmt.Println("\n📋 Test 4: AI-Powered Agent Discussion")
	
	if provider == nil {
		fmt.Println("   ⚠️  Skipping (no LLM provider available)")
	} else {
		aiManager := meeting.NewAIDiscussionManager(provider, conference, agentsDir)
		
		fmt.Println("   🎬 Starting AI discussion: 'System Architecture Review'")
		
		discCtx := context.Background()
		disc, err := aiManager.StartAIDiscussion(
			discCtx,
			"System Architecture Review",
			"We're designing a new microservices architecture. Need to decide on service boundaries, communication patterns, and database strategy.",
			[]string{"jarvis", "nova", "clawed"},
			2, // 2 rounds
		)
		
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
		} else {
			// Set callback to show messages
			disc.SetCallbacks(
				func(msg meeting.AIDiscussionMessage) {
					fmt.Printf("   %s %s: %s\n", msg.Avatar, msg.Name, truncate(msg.Content, 60))
				},
				func(consensus string) {
					fmt.Printf("\n   ✅ Discussion completed!\n")
					fmt.Printf("   📊 Consensus: %s\n", consensus)
				},
			)
			
			// Wait for discussion to complete
			time.Sleep(3 * time.Second)
			
			// Show final stats
			if fullDisc, ok := aiManager.GetDiscussion(disc.ID); ok {
				fmt.Printf("\n   📊 Total messages: %d\n", len(fullDisc.Messages))
				fmt.Printf("   🔄 Turns completed: %d/%d\n", fullDisc.CurrentTurn+1, fullDisc.MaxTurns)
			}
		}
	}

	// Test 5: Manual Meeting with AI
	fmt.Println("\n📋 Test 5: Facilitated Meeting")
	
	meetCtx := context.Background()
	m, _ := conference.FacilitateMeeting(meetCtx, "Feature Planning", 
		[]string{"Requirements", "Design", "Implementation", "Testing"},
		[]string{"atlas", "clawed", "pixel"})
	
	// Simulate discussion
	discussion := []struct {
		agent   string
		content string
		msgType string
	}{
		{"atlas", "จาก research ผู้ใช้ต้องการ feature ที่ใช้งานง่ายและเร็ว", "statement"},
		{"clawed", " technically feasible ครับ ใช้ Go + Redis ได้", "agreement"},
		{"pixel", "UI ผมเสนอแบบ minimalist เน้น speed", "proposal"},
		{"jarvis", "สรุป: ไปตามแผนนี้ได้ครับ", "summary"},
	}
	
	for _, msg := range discussion {
		m.PostMessage(msg.agent, msg.content, msg.msgType, nil)
		time.Sleep(100 * time.Millisecond)
	}
	
	m.End("เห็นชอบดำเนินการตามแผน", boolPtr(true))
	
	fmt.Printf("   ✅ Meeting completed: %s\n", m.ID)
	fmt.Printf("   📊 Messages: %d\n", len(m.Messages))

	// Summary
	fmt.Println("\n=====================================")
	fmt.Println("✅ Complete Agent System Test Finished!")
	fmt.Println("\n🎉 Features demonstrated:")
	fmt.Println("   • Agent Persona Files (IDENTITY.md, SOUL.md, MEMORY.md)")
	fmt.Println("   • Auto-creation of persona files")
	fmt.Println("   • Meeting Scheduler with recurring meetings")
	fmt.Println("   • AI-Powered Agent Discussions (LLM-driven)")
	fmt.Println("   • Facilitated Meetings")
	fmt.Println("\n📁 Persona files created at:")
	fmt.Printf("   %s/agents/{agent_id}/\n", agentsDir)
	fmt.Println("\n💡 Next steps:")
	fmt.Println("   - Agents now have unique personalities")
	fmt.Println("   - Meetings can be scheduled in advance")
	fmt.Println("   - AI can drive multi-agent discussions")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func boolPtr(b bool) *bool {
	return &b
}
