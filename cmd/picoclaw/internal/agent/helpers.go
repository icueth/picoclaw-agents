package agent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"

	"picoclaw/agent/cmd/picoclaw/internal"
	"picoclaw/agent/pkg/agent"
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/memory"
	"picoclaw/agent/pkg/providers"
)

func agentCmd(message, sessionKey, model string, debug bool) error {
	if sessionKey == "" {
		sessionKey = "cli:default"
	}

	if debug {
		logger.SetLevel(logger.DEBUG)
		fmt.Println("🔍 Debug mode enabled")
	}

	cfg, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	if model != "" {
		cfg.Agents.Defaults.ModelName = model
	}

	provider, modelID, err := providers.CreateProvider(cfg)
	if err != nil {
		return fmt.Errorf("error creating provider: %w", err)
	}

	// Use the resolved model ID from provider creation
	if modelID != "" {
		cfg.Agents.Defaults.ModelName = modelID
	}

	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	// Initialize job manager (can be nil if not configured)
	var jobManager *memory.JobManager
	if cfg.Jobs.Persistence != "" {
		// Note: Database initialization would happen here in a full implementation
		// For now, we pass nil to use in-memory job tracking
		jobManager = memory.NewJobManager(nil, &cfg.Jobs)
	}

	// Memory manager is not initialized in agent mode without bootstrap
	// It will be nil, and the system will fall back to legacy memory
	var memoryManager *agent.MemoryManager

	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider, jobManager, memoryManager)

	// Print agent startup info (only for interactive mode)
	startupInfo := agentLoop.GetStartupInfo()
	logger.InfoCF("agent", "Agent initialized",
		map[string]any{
			"tools_count":      startupInfo["tools"].(map[string]any)["count"],
			"skills_total":     startupInfo["skills"].(map[string]any)["total"],
			"skills_available": startupInfo["skills"].(map[string]any)["available"],
		})

	if message != "" {
		ctx := context.Background()
		fmt.Printf("\n%s ", internal.Logo)
		response, err := agentLoop.ProcessDirect(ctx, message, sessionKey, func(chunk string) {
			fmt.Print(chunk)
		})
		if err != nil {
			return fmt.Errorf("error processing message: %w", err)
		}
		if response != "" {
			fmt.Println(response)
		}
		fmt.Println()

		agentLoop.WaitForBackgroundTasks()
		return nil
	}

	fmt.Printf("%s Interactive mode (Ctrl+C to exit)\n\n", internal.Logo)
	interactiveMode(agentLoop, sessionKey)

	return nil
}

func interactiveMode(agentLoop *agent.AgentLoop, sessionKey string) {
	prompt := fmt.Sprintf("%s You: ", internal.Logo)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		HistoryFile:     filepath.Join(os.TempDir(), ".picoclaw_history"),
		HistoryLimit:    100,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		fmt.Printf("Error initializing readline: %v\n", err)
		fmt.Println("Falling back to simple input mode...")
		simpleInteractiveMode(agentLoop, sessionKey)
		return
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				fmt.Println("\nGoodbye!")
				return
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			return
		}

		ctx := context.Background()
		fmt.Printf("\n%s ", internal.Logo)
		response, err := agentLoop.ProcessDirect(ctx, input, sessionKey, func(chunk string) {
			fmt.Print(chunk)
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		if response == "" {
			fmt.Println()
		} else {
			fmt.Println(response + "\n")
		}
	}
}

func simpleInteractiveMode(agentLoop *agent.AgentLoop, sessionKey string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(fmt.Sprintf("%s You: ", internal.Logo))
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nGoodbye!")
				return
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			return
		}

		ctx := context.Background()
		fmt.Printf("\n%s ", internal.Logo)
		response, err := agentLoop.ProcessDirect(ctx, input, sessionKey, func(chunk string) {
			fmt.Print(chunk)
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		if response != "" {
			fmt.Println(response)
		}
		fmt.Println()
	}
}
