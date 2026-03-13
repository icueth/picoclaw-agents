// PicoClaw - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package config

import (
	"os"
	"path/filepath"
)

// floatPtr returns a pointer to a float64 value
func floatPtr(f float64) *float64 {
	return &f
}

// DefaultConfig returns the default configuration for PicoClaw with A2A 8-Agent System.
func DefaultConfig() *Config {
	// Determine the base path for the workspace.
	// Priority: $PICOCLAW_HOME > ~/.picoclaw
	var homePath string
	if picoclawHome := os.Getenv("PICOCLAW_HOME"); picoclawHome != "" {
		homePath = picoclawHome
	} else {
		userHome, _ := os.UserHomeDir()
		homePath = filepath.Join(userHome, ".picoclaw")
	}
	workspacePath := filepath.Join(homePath, "workspace")

	return &Config{
		// Workspace is the shared workspace directory for all agents
		Workspace: WorkspaceConfig{
			Path: workspacePath,
		},
		// A2A defines the Agent-to-Agent collaboration system with token optimization
		A2A: A2AConfig{
			Enabled: true,
			TokenOptimization: A2ATokenOptimizationConfig{
				Phase1RedundantIdentity:    true,
				Phase2ContextCompression:   true,
				Phase3MessageSummarization: true,
				Phase4ProjectSummary:       true,
				Phase5LightweightPrompt:    true,
			},
			Compression: A2ACompressionConfig{
				FullIterationsKeep:      2,
				CompressThresholdTokens: 4000,
				MaxContextTokens:        8000,
			},
			Messaging: A2AMessagingConfig{
				RecentMessagesKeep: 5,
				SummarizeThreshold: 20,
				ArchiveThreshold:   50,
			},
			Orchestrator: A2AOrchestratorConfig{
				MaxConcurrentTasksPerAgent: 2,
				MaxConcurrentLLMCalls:      3,
				OutsourcePoolSize:          5,
				OutsourceAgentTTLMinutes:   30,
				SharedContextMaxLogSize:    1000,
				SharedContextMaxContext:    10000,
				PersistencePath:            filepath.Join(homePath, "a2a_projects"),
				DefaultLanguage:            "en",
			},
		},
		Agents: AgentsConfig{
			Defaults: AgentDefaults{
				// Workspace is now centrally defined in Workspace.Path.
				// Kept as fallback for old agent configs.
				Workspace:                 workspacePath,
				RestrictToWorkspace:       false,
				AllowReadOutsideWorkspace: true,
				Model:                     "kimi-coding/kimi-for-coding",
				MaxTokens:                 32768,
				MaxToolIterations:         50,
				SummarizeMessageThreshold: 20,
				SummarizeTokenPercent:     75,
			},
			// Built-in agents inherit models from their department
			DepartmentModels: map[string]DepartmentModelConfig{
				"engineering":        {Model: "kimi-coding/kimi-for-coding"},
				"marketing":          {Model: "kimi-coding/kimi-for-coding"},
				"testing":            {Model: "kimi-coding/kimi-for-coding"},
				"design":             {Model: "kimi-coding/kimi-for-coding"},
				"product":            {Model: "kimi-coding/kimi-for-coding"},
				"project-management": {Model: "kimi-coding/kimi-for-coding"},
				"support":            {Model: "kimi-coding/kimi-for-coding"},
				"specialized":        {Model: "kimi-coding/kimi-for-coding"},
				"game-development":   {Model: "kimi-coding/kimi-for-coding"},
				"spatial-computing":  {Model: "kimi-coding/kimi-for-coding"},
				"paid-media":         {Model: "kimi-coding/kimi-for-coding"},
				"core":               {Model: "kimi-coding/kimi-for-coding"},
			},
		},
		Bindings: []AgentBinding{},
		Session: SessionConfig{
			DMScope: "per-channel-peer",
		},
		Channels: ChannelsConfig{
			WhatsApp: WhatsAppConfig{
				Enabled:          false,
				BridgeURL:        "ws://localhost:3001",
				UseNative:        false,
				SessionStorePath: "",
				AllowFrom:        FlexibleStringSlice{},
			},
			Telegram: TelegramConfig{
				Enabled:   false,
				Token:     "",
				AllowFrom: FlexibleStringSlice{},
				Typing:    TypingConfig{Enabled: true},
				Placeholder: PlaceholderConfig{
					Enabled: true,
					Text:    "Thinking... 💭",
				},
			},
			Feishu: FeishuConfig{
				Enabled:           false,
				AppID:             "",
				AppSecret:         "",
				EncryptKey:        "",
				VerificationToken: "",
				AllowFrom:         FlexibleStringSlice{},
			},
			Discord: DiscordConfig{
				Enabled:     false,
				Token:       "",
				AllowFrom:   FlexibleStringSlice{},
				MentionOnly: false,
			},
			MaixCam: MaixCamConfig{
				Enabled:   false,
				Host:      "0.0.0.0",
				Port:      18790,
				AllowFrom: FlexibleStringSlice{},
			},
			QQ: QQConfig{
				Enabled:   false,
				AppID:     "",
				AppSecret: "",
				AllowFrom: FlexibleStringSlice{},
			},
			DingTalk: DingTalkConfig{
				Enabled:      false,
				ClientID:     "",
				ClientSecret: "",
				AllowFrom:    FlexibleStringSlice{},
			},
			Slack: SlackConfig{
				Enabled:   false,
				BotToken:  "",
				AppToken:  "",
				AllowFrom: FlexibleStringSlice{},
			},
			LINE: LINEConfig{
				Enabled:            false,
				ChannelSecret:      "",
				ChannelAccessToken: "",
				WebhookHost:        "0.0.0.0",
				WebhookPort:        18791,
				WebhookPath:        "/webhook/line",
				AllowFrom:          FlexibleStringSlice{},
				GroupTrigger:       GroupTriggerConfig{MentionOnly: true},
			},
			OneBot: OneBotConfig{
				Enabled:            false,
				WSUrl:              "ws://127.0.0.1:3001",
				AccessToken:        "",
				ReconnectInterval:  5,
				GroupTriggerPrefix: []string{},
				AllowFrom:          FlexibleStringSlice{},
			},
			WeCom: WeComConfig{
				Enabled:        false,
				Token:          "",
				EncodingAESKey: "",
				WebhookURL:     "",
				WebhookHost:    "0.0.0.0",
				WebhookPort:    18793,
				WebhookPath:    "/webhook/wecom",
				AllowFrom:      FlexibleStringSlice{},
				ReplyTimeout:   5,
			},
			WeComApp: WeComAppConfig{
				Enabled:        false,
				CorpID:         "",
				CorpSecret:     "",
				AgentID:        0,
				Token:          "",
				EncodingAESKey: "",
				WebhookHost:    "0.0.0.0",
				WebhookPort:    18792,
				WebhookPath:    "/webhook/wecom-app",
				AllowFrom:      FlexibleStringSlice{},
				ReplyTimeout:   5,
			},
			WeComAIBot: WeComAIBotConfig{
				Enabled:        false,
				Token:          "",
				EncodingAESKey: "",
				WebhookPath:    "/webhook/wecom-aibot",
				AllowFrom:      FlexibleStringSlice{},
				ReplyTimeout:   5,
				MaxSteps:       10,
				WelcomeMessage: "Hello! I'm your AI assistant. How can I help you today?",
			},
			Pico: PicoConfig{
				Enabled:        false,
				Token:          "",
				PingInterval:   30,
				ReadTimeout:    60,
				WriteTimeout:   10,
				MaxConnections: 100,
				AllowFrom:      FlexibleStringSlice{},
			},
		},
		Providers: ProvidersConfig{
			OpenAI: OpenAIProviderConfig{WebSearch: true},
		},
		ModelList: []ModelConfig{
			{
				ModelName: "kimi-for-coding",
				Model:     "kimi-coding/kimi-for-coding",
				APIBase:   "https://api.kimi.com/coding/v1",
				APIKey:    "",
			},
			{
				ModelName: "qwen3.5-plus",
				Model:     "bailian/qwen3.5-plus",
				APIBase:   "https://coding-intl.dashscope.aliyuncs.com/v1",
				APIKey:    "",
			},
			{
				ModelName: "kimi-k2.5",
				Model:     "moonshot/kimi-k2.5",
				APIBase:   "https://api.moonshot.ai/v1",
				APIKey:    "",
			},
			{
				ModelName: "gpt-4o",
				Model:     "openai/gpt-4o",
				APIBase:   "https://api.openai.com/v1",
				APIKey:    "",
			},
			{
				ModelName: "claude-3-5-sonnet",
				Model:     "anthropic/claude-3-5-sonnet-20241022",
				APIBase:   "https://api.anthropic.com/v1",
				APIKey:    "",
			},
			{
				ModelName: "deepseek-chat",
				Model:     "deepseek/deepseek-chat",
				APIBase:   "https://api.deepseek.com/v1",
				APIKey:    "",
			},
			{
				ModelName: "gemini-2.0-flash",
				Model:     "google/gemini-2.0-flash",
				APIBase:   "https://generativelanguage.googleapis.com/v1beta",
				APIKey:    "",
			},
		},
		Gateway: GatewayConfig{
			Host:      "127.0.0.1",
			Port:      18790,
			UIEnabled: false,
		},
		Tools: ToolsConfig{
			MediaCleanup: MediaCleanupConfig{
				Enabled:  true,
				MaxAge:   30,
				Interval: 5,
			},
			Web: WebToolsConfig{
				Proxy:           "",
				FetchLimitBytes: 10 * 1024 * 1024,
				Brave: BraveConfig{
					Enabled:    false,
					APIKey:     "",
					MaxResults: 5,
				},
				DuckDuckGo: DuckDuckGoConfig{
					Enabled:    true,
					MaxResults: 5,
				},
				Perplexity: PerplexityConfig{
					Enabled:    false,
					APIKey:     "",
					MaxResults: 5,
				},
				GLMSearch: GLMSearchConfig{
					Enabled:      false,
					APIKey:       "",
					BaseURL:      "https://open.bigmodel.cn/api/paas/v4/web_search",
					SearchEngine: "search_std",
					MaxResults:   5,
				},
			},
			Cron: CronToolsConfig{
				ExecTimeoutMinutes: 5,
			},
			Exec: ExecConfig{
				EnableDenyPatterns: true,
				SafetyLevel:        "permissive",
			},
			Skills: SkillsToolsConfig{
				Registries: SkillsRegistriesConfig{
					ClawHub: ClawHubRegistryConfig{
						Enabled: true,
						BaseURL: "https://clawhub.ai",
					},
				},
				MaxConcurrentSearches: 2,
				SearchCache: SearchCacheConfig{
					MaxSize:    50,
					TTLSeconds: 300,
				},
			},
			MCP: MCPConfig{
				Enabled: false,
				Servers: map[string]MCPServerConfig{
					"postgres": {
						Enabled: false,
						Command: "npx",
						Args:    []string{"-y", "@modelcontextprotocol/server-postgres", "postgresql://user:password@localhost/dbname"},
					},
					"mysql": {
						Enabled: false,
						Command: "npx",
						Args:    []string{"-y", "@benborla29/mcp-server-mysql"},
						Env: map[string]string{
							"MYSQL_HOST":     "localhost",
							"MYSQL_PORT":     "3306",
							"MYSQL_USER":     "root",
							"MYSQL_PASSWORD": "",
							"MYSQL_DATABASE": "mydb",
						},
					},
					"mongodb": {
						Enabled: false,
						Command: "npx",
						Args:    []string{"-y", "mongodb-mcp-server"},
						Env:     map[string]string{"MDB_MCP_CONNECTION_STRING": "mongodb://localhost:27017/mydb"},
					},
					"sqlite": {
						Enabled: false,
						Command: "npx",
						Args:    []string{"-y", "@anthropic/mcp-server-sqlite", "/path/to/database.db"},
					},
					"redis": {
						Enabled: false,
						Command: "npx",
						Args:    []string{"-y", "@anthropic/mcp-server-redis", "redis://localhost:6379"},
					},
					"playwright": {
						Enabled: true,
						Command: "npx",
						Args:    []string{"-y", "@executeautomation/playwright-mcp-server"},
					},
					"sequential-thinking": {
						Enabled: false,
						Command: "npx",
						Args:    []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
					},
				},
			},
		},
		Heartbeat: HeartbeatConfig{
			Enabled:  true,
			Interval: 30,
		},
		Devices: DevicesConfig{
			Enabled:    false,
			MonitorUSB: true,
		},
		Memory: MemoryConfig{
			Type:                 "sqlite",
			Database:             filepath.Join(homePath, "picoclaw.db"),
			ConceptRetentionDays: 30,
			MaxContextMemories:   10,
			AutoSave:             true,
			RAG: RAGConfig{
				Enabled:             true,
				EmbeddingModel:      "none", // Default to keyword-only search (zero-config)
				ChunkSize:           512,
				Overlap:             128,
				MaxResults:          5,
				SimilarityThreshold: 0.7,
				APIBase:             "http://localhost:18190",
				APIKey:              "",
				ModelPath:           filepath.Join(homePath, "models"),
				Dimension:           384,
				VectorWeight:        0.7, // Used when embeddings are enabled
				KeywordWeight:       0.3, // Always used for hybrid search
			},
		},
		Jobs: JobConfig{
			Persistence:          "sqlite",
			DatabasePath:         filepath.Join(homePath, "picoclaw.db"),
			DefaultTimeout:       3600,
			MaxConcurrent:        10,
			CleanupIntervalHours: 24,
			RetentionDays:        30,
		},
	}
}
