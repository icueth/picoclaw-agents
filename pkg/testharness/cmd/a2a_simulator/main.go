// A2A Simulator CLI
// เครื่องมือทดสอบจำลอง Agent-to-Agent (A2A) Orchestrator แบบสมบูรณ์แบบโดยไม่ต้องใช้ API Server
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"picoclaw/agent/pkg/agent"
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/providers"
	"picoclaw/agent/pkg/tools"
)

// ANSI Colors for output terminal
const (
	ColorReset  = "\033[0m"
	ColorBlue   = "\033[34m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
)

func main() {
	fmt.Println(ColorBold + ColorCyan + "╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║               🧠 PicoClaw A2A Simulator 🚀                   ║")
	fmt.Println("║  (100% Real LLM, Real Agents, Real A2A Orchestrator Engine)  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝" + ColorReset)
	fmt.Println()

	// 1. Load config
	configPath := os.ExpandEnv("${HOME}/.picoclaw/config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf(ColorRed+"Error: Config file not found at %s. Please run picoclaw configuration first.\n"+ColorReset, configPath)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf(ColorRed+"Error loading config: %v\n"+ColorReset, err)
		os.Exit(1)
	}

	modelName := cfg.Agents.Defaults.GetModelName()
	fmt.Printf("📦 Using Model: "+ColorCyan+"%s"+ColorReset+"\n", modelName)

	modelCfg, err := cfg.GetModelConfig(modelName)
	if err != nil {
		fmt.Printf(ColorRed+"Error getting model config: %v\n"+ColorReset, err)
		os.Exit(1)
	}

	// 2. Setup Provider
	provider, resolvedModelName, err := providers.CreateProviderFromConfig(modelCfg)
	if err != nil {
		fmt.Printf(ColorRed+"Failed to create LLM provider: %v\n"+ColorReset, err)
		os.Exit(1)
	}
	fmt.Printf("🚀 LLM Provider Initialized: "+ColorCyan+"%s"+ColorReset+"\n", resolvedModelName)

	// 3. Setup Bus & SharedContext
	msgBus := bus.NewMessageBus()

	// 4. Setup Agent Registry
	registry := agent.NewAgentRegistry(cfg, provider)
	agentIDs := registry.ListAgentIDs()
	fmt.Printf("👥 Agents Loaded (%d): "+ColorBlue+"%v"+ColorReset+"\n", len(agentIDs), agentIDs)

	// 5. Setup Agent Discovery
	discovery := agent.NewA2AAgentDiscovery(registry, provider, cfg.Agents.Defaults.GetModelName())
	// (Note: A2A discovery is usually dynamic, but we can seed it here if needed)

	// 6. Build A2A Orchestrator
	orchestrator := agent.NewA2AOrchestrator(registry, provider, cfg, msgBus)
	orchestrator.SetDiscovery(discovery)

	// 7. Register Shared Tools exactly like in main loop

	// Use the central registration logic from loop.go to ensure consistency
	// (Note: registerSharedTools is internal, but we can replicate it or make it accessible)
	// For now, let's manually call New tools or update loop.go to export it? 
	// Actually, I'll just register the missing A2A tools here manually as it's a test harness.
	
	for _, id := range agentIDs {
		instance, _ := registry.GetAgent(id)
		
		// Register core A2A tools
		instance.Tools.Register(agent.NewStartA2AProjectTool(orchestrator))
		instance.Tools.Register(agent.NewCheckA2AProjectStatusTool(orchestrator))
		instance.Tools.Register(agent.NewListA2AAgentsTool(orchestrator.GetDiscovery()))
		instance.Tools.Register(agent.NewSendA2AMessageTool(orchestrator))
		instance.Tools.Register(agent.NewGetA2AMessagesTool(orchestrator))
		instance.Tools.Register(agent.NewResumeA2AProjectTool(orchestrator))

		// Web search (duckduckgo)
		searchTool, err := tools.NewWebSearchTool(tools.WebSearchToolOptions{
			DuckDuckGoEnabled:    true,
			DuckDuckGoMaxResults: 5,
		})
		if err == nil {
			instance.Tools.Register(searchTool)
		}
		
		// Fetch tool
		fetchTool, err := tools.NewWebFetchToolWithProxy(50000, "", 1024*1024)
		if err == nil {
			instance.Tools.Register(fetchTool)
		}

		// ListDir tool
		instance.Tools.Register(tools.NewListDirTool(instance.Workspace, false))
		
		// Exec Tool
		execTool, err := tools.NewExecToolWithConfig(instance.Workspace, false, cfg)
		if err == nil {
			instance.Tools.Register(execTool)
		}
	}
	fmt.Println("🛠️  Agent Tools and Roles Successfully Bound")

	// 7. Initialize Orchestrator Workers
	orchestrator.Initialize()
	fmt.Println("⚙️  A2A Orchestrator Runtime Initialized")
	fmt.Println()

	// 8. Setup Callbacks to print live UI for user
	setupCallbacks(orchestrator)

	// Terminal interaction loop
	reader := bufio.NewReader(os.Stdin)

	// Handle Ctrl+C
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nExiting A2A Simulator...")
		os.Exit(0)
	}()

	fmt.Println(ColorGreen + "Ready! Type your instruction/task and press Enter." + ColorReset)
	fmt.Println(ColorYellow + "Type 'resume <project_id>' to continue a failed project." + ColorReset)
	fmt.Println(ColorYellow + "Example: 'Design and write a python script to ping 8.8.8.8'" + ColorReset)

	for {
		fmt.Printf("\n" + ColorBold + "JARVIS > " + ColorReset)
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			break
		}

		// Handle Resume Command
		if strings.HasPrefix(input, "resume ") {
			projectID := strings.TrimPrefix(input, "resume ")
			fmt.Printf("\n" + ColorBlue + "[System] Resuming A2A Project: %q" + ColorReset + "\n", projectID)
			err = orchestrator.ResumeProject(projectID)
			if err != nil {
				fmt.Printf(ColorRed + "Failed to resume project: %v\n" + ColorReset, err)
			} else {
				fmt.Println(ColorYellow + "Project resumed. Watch the progress below." + ColorReset)
			}
			continue
		}

		fmt.Printf("\n" + ColorBlue + "[System] Starting A2A Project for: %q" + ColorReset + "\n", input)
		
		// Create Project
		project := orchestrator.CreateProject("Simulated Project", input)
		// Run Project
		err = orchestrator.StartProject(project.ID)
		if err != nil {
			fmt.Printf(ColorRed + "Failed to start project: %v\n" + ColorReset, err)
		}

		// Next input will block until user types something, but A2A runs in background goroutine
		// The logs will interleave with the prompt, which is normal for a simulator.
		fmt.Println(ColorYellow + "Project submitted. Orchestrator and Agents are now communicating in the background." + ColorReset)
	}
}

func setupCallbacks(orchestrator *agent.A2AOrchestrator) {
	orchestrator.SetPhaseChangeCallback(func(projectID string, phase agent.Phase, status agent.PhaseStatus) {
		color := ColorCyan
		if status == agent.PhaseStatusCompleted {
			color = ColorGreen
		} else if status == agent.PhaseStatusFailed {
			color = ColorRed
		}
		fmt.Printf("\n%s▶️ PHASE CHANGE: [%s] -> %s%s\n", color, phase, status, ColorReset)
	})

	orchestrator.SetMessageCallback(func(projectID string, msg agent.A2AMessage) {
		senderColor := getAgentColor(msg.From)
		
		fmt.Printf("%s[%s]%s -> %s: %s\n", 
			senderColor, strings.ToUpper(msg.From), ColorReset, 
			msg.To, msg.Content)
	})

	orchestrator.SetAssignmentProgressCallback(func(projectID string, assignmentID string, progress int, message string) {
		fmt.Printf("%s⏳ PROGRESS: [%s] %d%% - %s%s\n", ColorYellow, assignmentID, progress, message, ColorReset)
	})

	orchestrator.SetAssignmentFailedCallback(func(projectID string, assignmentID string, agentID string, err error) {
		fmt.Printf("%s❌ ERROR: [%s] %s failed - %v%s\n", ColorRed, agentID, assignmentID, err, ColorReset)
	})
}

func getAgentColor(agentID string) string {
	switch strings.ToLower(agentID) {
	case "jarvis":
		return ColorCyan
	case "atlas":
		return ColorPurple
	case "clawed", "coder":
		return ColorGreen
	case "sentinel", "qa":
		return ColorYellow
	case "scribe", "writer":
		return ColorBlue
	default:
		return ColorReset
	}
}
