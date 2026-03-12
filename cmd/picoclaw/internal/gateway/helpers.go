package gateway

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"picoclaw/agent/cmd/picoclaw/internal"
	"picoclaw/agent/pkg/agent"
	"picoclaw/agent/pkg/agent/meeting"
	"picoclaw/agent/pkg/api/ui"
	"picoclaw/agent/pkg/bootstrap"
	"picoclaw/agent/pkg/bus"
	"picoclaw/agent/pkg/channels"
	_ "picoclaw/agent/pkg/channels/dingtalk"
	_ "picoclaw/agent/pkg/channels/discord"
	_ "picoclaw/agent/pkg/channels/feishu"
	_ "picoclaw/agent/pkg/channels/line"
	_ "picoclaw/agent/pkg/channels/maixcam"
	_ "picoclaw/agent/pkg/channels/onebot"
	_ "picoclaw/agent/pkg/channels/pico"
	_ "picoclaw/agent/pkg/channels/qq"
	_ "picoclaw/agent/pkg/channels/slack"
	_ "picoclaw/agent/pkg/channels/telegram"
	_ "picoclaw/agent/pkg/channels/wecom"
	_ "picoclaw/agent/pkg/channels/whatsapp"
	_ "picoclaw/agent/pkg/channels/whatsapp_native"
	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/cron"
	"picoclaw/agent/pkg/devices"
	"picoclaw/agent/pkg/health"
	"picoclaw/agent/pkg/heartbeat"
	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/media"
	"picoclaw/agent/pkg/office"
	"picoclaw/agent/pkg/providers"
	"picoclaw/agent/pkg/state"
	"picoclaw/agent/pkg/tools"
)

func gatewayCmd(debug bool, noUI bool) error {
	if debug {
		logger.SetLevel(logger.DEBUG)
		fmt.Println("🔍 Debug mode enabled")
	}

	cfg, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Disable UI if requested via CLI flag
	if noUI {
		cfg.Gateway.UIEnabled = false
	}

	if !cfg.Gateway.UIEnabled {
		fmt.Println("🚫 UI Office System components disabled")
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

	// Initialize system components using bootstrap
	fmt.Println("📦 Initializing system components...")
	sys, err := bootstrap.Bootstrap(cfg)
	if err != nil {
		return fmt.Errorf("failed to bootstrap system: %w", err)
	}
	defer sys.Close()

	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider, sys.JobMgr, sys.MemoryManager)

	// Print agent startup info
	fmt.Println("\n📦 Agent Status:")
	startupInfo := agentLoop.GetStartupInfo()
	toolsInfo := startupInfo["tools"].(map[string]any)
	skillsInfo := startupInfo["skills"].(map[string]any)
	fmt.Printf("  • Tools: %d loaded\n", toolsInfo["count"])
	fmt.Printf("  • Skills: %d/%d available\n",
		skillsInfo["available"],
		skillsInfo["total"])

	// Log to file as well
	logger.InfoCF("agent", "Agent initialized",
		map[string]any{
			"tools_count":      toolsInfo["count"],
			"skills_total":     skillsInfo["total"],
			"skills_available": skillsInfo["available"],
		})

	// Setup cron tool and service
	execTimeout := time.Duration(cfg.Tools.Cron.ExecTimeoutMinutes) * time.Minute
	cronService := setupCronTool(
		agentLoop,
		msgBus,
		cfg.WorkspacePath(),
		cfg.Agents.Defaults.RestrictToWorkspace,
		execTimeout,
		cfg,
	)

	heartbeatService := heartbeat.NewHeartbeatService(
		cfg.WorkspacePath(),
		cfg.Heartbeat.Interval,
		cfg.Heartbeat.Enabled,
	)
	heartbeatService.SetBus(msgBus)
	heartbeatService.SetHandler(func(prompt, channel, chatID string) *tools.ToolResult {
		// Use cli:direct as fallback if no valid channel
		if channel == "" || chatID == "" {
			channel, chatID = "cli", "direct"
		}
		// Use ProcessHeartbeat - no session history, each heartbeat is independent
		var response string
		response, err = agentLoop.ProcessHeartbeat(context.Background(), prompt, channel, chatID)
		if err != nil {
			return tools.ErrorResult(fmt.Sprintf("Heartbeat error: %v", err))
		}
		if response == "HEARTBEAT_OK" {
			return tools.SilentResult("Heartbeat OK")
		}
		// For heartbeat, always return silent - the subagent result will be
		// sent to user via processSystemMessage when the async task completes
		return tools.SilentResult(response)
	})

	// Create media store for file lifecycle management with TTL cleanup
	mediaStore := media.NewFileMediaStoreWithCleanup(media.MediaCleanerConfig{
		Enabled:  cfg.Tools.MediaCleanup.Enabled,
		MaxAge:   time.Duration(cfg.Tools.MediaCleanup.MaxAge) * time.Minute,
		Interval: time.Duration(cfg.Tools.MediaCleanup.Interval) * time.Minute,
	})
	mediaStore.Start()

	channelManager, err := channels.NewManager(cfg, msgBus, mediaStore)
	if err != nil {
		mediaStore.Stop()
		return fmt.Errorf("error creating channel manager: %w", err)
	}

	// Inject channel manager and media store into agent loop
	agentLoop.SetChannelManager(channelManager)
	agentLoop.SetMediaStore(mediaStore)

	enabledChannels := channelManager.GetEnabledChannels()
	if len(enabledChannels) > 0 {
		fmt.Printf("✓ Channels enabled: %s\n", enabledChannels)
	} else {
		fmt.Println("⚠ Warning: No channels enabled")
	}

	fmt.Printf("✓ Gateway started on %s:%d\n", cfg.Gateway.Host, cfg.Gateway.Port)
	fmt.Println("Press Ctrl+C to stop")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := cronService.Start(); err != nil {
		fmt.Printf("Error starting cron service: %v\n", err)
	}
	fmt.Println("✓ Cron service started")

	if err := heartbeatService.Start(); err != nil {
		fmt.Printf("Error starting heartbeat service: %v\n", err)
	}
	fmt.Println("✓ Heartbeat service started")

	stateManager := state.NewManager(cfg.WorkspacePath())
	deviceService := devices.NewService(devices.Config{
		Enabled:    cfg.Devices.Enabled,
		MonitorUSB: cfg.Devices.MonitorUSB,
	}, stateManager)
	deviceService.SetBus(msgBus)
	if err := deviceService.Start(ctx); err != nil {
		fmt.Printf("Error starting device service: %v\n", err)
	} else if cfg.Devices.Enabled {
		fmt.Println("✓ Device event service started")
	}

	// Setup shared HTTP server with health endpoints and webhook handlers
	healthServer := health.NewServer(cfg.Gateway.Host, cfg.Gateway.Port)
	addr := fmt.Sprintf("%s:%d", cfg.Gateway.Host, cfg.Gateway.Port)
	channelManager.SetupHTTPServer(addr, healthServer)

	// Setup meeting API
	fmt.Println("📦 Initializing meeting system...")
	builtinAgents := agent.GetBuiltinAgents()
	var apiAgents []config.AgentConfig
	for _, ba := range builtinAgents {
		apiAgents = append(apiAgents, ba.ToAgentConfig(cfg.GetDepartmentModel(ba.Department)))
	}
	meetingAPI := meeting.NewAPIHandler(cfg, provider, apiAgents)
	if mux := channelManager.GetMux(); mux != nil {
		meetingAPI.RegisterRoutes(mux)
		fmt.Println("✓ Meeting API registered at /api/meetings...")

		// Setup UI Office System event bridge
		if cfg.Gateway.UIEnabled {
			fmt.Println("📦 Initializing UI Office System...")
			uiHub := ui.NewHub()
			go uiHub.Run()

			officeManager := office.NewOfficeManager(office.DefaultOfficeConfig())
			if err := officeManager.Initialize(); err != nil {
				logger.ErrorCF("office", "Failed to initialize office manager", map[string]any{"error": err.Error()})
			}

			uiStore := ui.NewMemoryStore()
			uiStore.SetOfficeManager(officeManager)
			uiStore.InitializeWithDefaults()

			uiHandler := ui.NewHandler(uiStore, uiHub)
			uiHandler.RegisterRoutes(mux)

			// Sync actual agents to the UI store
			for _, agentID := range agentLoop.Registry().ListAgentIDs() {
				if agInfo, ok := agentLoop.Registry().GetAgentInfo(agentID); ok {
					uiStore.CreateAgent(ui.Agent{
						ID:           agInfo.ID,
						Name:         agInfo.Name,
						DepartmentID: "coding", // Use actual department from Initialize
						Role:         agInfo.Type,
						Status:       ui.AgentStatusIdle,
						Capabilities: agInfo.Capabilities,
						IsOnline:     true,
						LastActive:   time.Now(),
					})
					// Assign them to a valid room
					_ = officeManager.AssignAgentToRoom(agInfo.ID, "coding-main")
				}
			}

			mux.HandleFunc("/api/ui/ws", uiHub.WebSocketHandler)
			fmt.Println("✓ UI REST & WebSocket APIs registered at /api/ui and /api/ui/ws")

			// Bridge agent events → UI hub (WS broadcast) + uiStore (REST state)
			go func() {
				for {
					event, ok := msgBus.SubscribeEvent(ctx)
					if !ok {
						return
					}
					var newStatus ui.AgentStatus
					switch event.EventType {
					case "THINKING":
						newStatus = ui.AgentStatusBusy
					case "WORKING":
						newStatus = ui.AgentStatusWorking
					case "IDLE":
						newStatus = ui.AgentStatusIdle
					default:
						newStatus = ui.AgentStatusIdle
					}

					// 1. Update in-memory store so REST /agents reflects live status
					if existing, ok := uiStore.GetAgent(event.AgentID); ok {
						existing.Status = newStatus
						uiStore.UpdateAgent(existing)
					}

					// 2. Broadcast over WS so frontend gets it instantly
					uiHub.BroadcastAgentStatusChanged(event.AgentID, ui.AgentStatusIdle, newStatus, event.Details)
				}
			}()
		}
	}

	if err := channelManager.StartAll(ctx); err != nil {
		fmt.Printf("Error starting channels: %v\n", err)
		return err
	}

	fmt.Printf("✓ Health endpoints available at http://%s:%d/health and /ready\n", cfg.Gateway.Host, cfg.Gateway.Port)

	go agentLoop.Run(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	fmt.Println("\nShutting down...")
	if cp, ok := provider.(providers.StatefulProvider); ok {
		cp.Close()
	}
	cancel()
	msgBus.Close()

	// Use a fresh context with timeout for graceful shutdown,
	// since the original ctx is already canceled.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	channelManager.StopAll(shutdownCtx)
	deviceService.Stop()
	heartbeatService.Stop()
	cronService.Stop()
	mediaStore.Stop()
	agentLoop.Stop()
	fmt.Println("✓ Gateway stopped")

	return nil
}

func setupCronTool(
	agentLoop *agent.AgentLoop,
	msgBus *bus.MessageBus,
	workspace string,
	restrict bool,
	execTimeout time.Duration,
	cfg *config.Config,
) *cron.CronService {
	cronStorePath := filepath.Join(workspace, "cron", "jobs.json")

	// Create cron service
	cronService := cron.NewCronService(cronStorePath, nil)

	// Create and register CronTool
	cronTool, err := tools.NewCronTool(cronService, agentLoop, msgBus, workspace, restrict, execTimeout, cfg)
	if err != nil {
		log.Fatalf("Critical error during CronTool initialization: %v", err)
	}

	agentLoop.RegisterTool(cronTool)

	// Set the onJob handler
	cronService.SetOnJob(func(job *cron.CronJob) (string, error) {
		result := cronTool.ExecuteJob(context.Background(), job)
		return result, nil
	})

	return cronService
}
