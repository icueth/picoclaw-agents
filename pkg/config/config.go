package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/caarlos0/env/v11"

	"picoclaw/agent/pkg/fileutil"
)

// rrCounter is a global counter for round-robin load balancing across models.
var rrCounter atomic.Uint64

// FlexibleStringSlice is a []string that also accepts JSON numbers,
// so allow_from can contain both "123" and 123.
type FlexibleStringSlice []string

func (f *FlexibleStringSlice) UnmarshalJSON(data []byte) error {
	// Try []string first
	var ss []string
	if err := json.Unmarshal(data, &ss); err == nil {
		*f = ss
		return nil
	}

	// Try []interface{} to handle mixed types
	var raw []any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	result := make([]string, 0, len(raw))
	for _, v := range raw {
		switch val := v.(type) {
		case string:
			result = append(result, val)
		case float64:
			result = append(result, fmt.Sprintf("%.0f", val))
		default:
			result = append(result, fmt.Sprintf("%v", val))
		}
	}
	*f = result
	return nil
}

type Config struct {
	Agents    AgentsConfig    `json:"agents"`
	Bindings  []AgentBinding  `json:"bindings,omitempty"`
	Session   SessionConfig   `json:"session,omitempty"`
	Channels  ChannelsConfig  `json:"channels"`
	// Providers is DEPRECATED. Use model_list instead.
	// Kept for backward compatibility only.
	Providers ProvidersConfig `json:"providers,omitempty"`
	// ModelList is the new model-centric provider configuration.
	// All provider settings should be configured here.
	ModelList []ModelConfig `json:"model_list"`
	Gateway   GatewayConfig `json:"gateway"`
	Tools     ToolsConfig   `json:"tools"`
	Heartbeat HeartbeatConfig `json:"heartbeat"`
	Devices   DevicesConfig   `json:"devices"`
	// SubagentRoles is DEPRECATED. A2A (Agent-to-Agent) system is used instead.
	// Kept for backward compatibility only.
	SubagentRoles map[string]SubagentRoleConfig `json:"subagent_roles,omitempty"`
	Memory        MemoryConfig                  `json:"memory,omitempty"`
	Jobs          JobConfig                     `json:"jobs,omitempty"`
	// Workspace defines the shared workspace directory for all agents.
	// All agents use the same workspace specified by workspace.path.
	Workspace WorkspaceConfig `json:"workspace,omitempty"`
	// A2A defines the Agent-to-Agent collaboration configuration with token optimization.
	// This is the new multi-agent system that replaced subagent_roles.
	A2A A2AConfig `json:"a2a,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for Config
// to omit providers section when empty and session when empty
func (c Config) MarshalJSON() ([]byte, error) {
	type Alias Config
	aux := &struct {
		Providers *ProvidersConfig `json:"providers,omitempty"`
		Session   *SessionConfig   `json:"session,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(&c),
	}

	// Only include providers if not empty
	if !c.Providers.IsEmpty() {
		aux.Providers = &c.Providers
	}

	// Only include session if not empty
	if c.Session.DMScope != "" || len(c.Session.IdentityLinks) > 0 {
		aux.Session = &c.Session
	}

	return json.Marshal(aux)
}

type AgentsConfig struct {
	Defaults AgentDefaults `json:"defaults"`
	// List contains the fixed agent list for the Office UI system.
	// DEPRECATED: Built-in agents are now embedded in core via builtin_agents.go.
	// This field is kept for backward compatibility — if set, these agents
	// are registered in ADDITION to built-in agents.
	List []AgentConfig `json:"list,omitempty"`
	// DepartmentModels maps department name -> model configuration.
	// Supports both legacy string format and new structured format.
	// Built-in agents inherit their model from this map based on their department.
	// Falls back to Defaults.Model if a department is not configured here.
	DepartmentModels map[string]DepartmentModelConfig `json:"department_models,omitempty"`
}

// UnmarshalJSON clears the default agents List before unmarshaling user config.
// If user config doesn't have agents.list, it stays nil (built-in agents are used).
func (a *AgentsConfig) UnmarshalJSON(data []byte) error {
	// Save defaults before unmarshaling
	defaultList := a.List
	defaultDeptModels := a.DepartmentModels

	// Clear defaults so they don't leak into user config
	a.List = nil
	a.DepartmentModels = nil
	type raw AgentsConfig
	if err := json.Unmarshal(data, (*raw)(a)); err != nil {
		return err
	}

	// If user config doesn't have agents.list, use defaults
	if a.List == nil {
		a.List = defaultList
	}

	// Merge department models: user overrides defaults
	if a.DepartmentModels == nil {
		a.DepartmentModels = defaultDeptModels
	} else {
		for dept, model := range defaultDeptModels {
			if _, exists := a.DepartmentModels[dept]; !exists {
				a.DepartmentModels[dept] = model
			}
		}
	}

	return nil
}

// GetDepartmentModel returns the model for a specific department.
// Falls back to Agents.Defaults.Model if the department is not configured.
func (c *Config) GetDepartmentModel(department string) string {
	if c.Agents.DepartmentModels != nil {
		if modelConfig, ok := c.Agents.DepartmentModels[department]; ok {
			effectiveModel := modelConfig.GetEffectiveModelName()
			if effectiveModel != "" {
				return effectiveModel
			}
		}
	}
	return c.Agents.Defaults.GetModelName()
}

// AgentModelConfig supports both string and structured model config.
// String format: "gpt-4" (just primary, no fallbacks)
// Object format: {"primary": "gpt-4", "fallbacks": ["claude-haiku"]}
type AgentModelConfig struct {
	Primary   string   `json:"primary,omitempty"`
	Fallbacks []string `json:"fallbacks,omitempty"`
}

// DepartmentModelConfig supports both string and structured model config for departments.
// String format: "gpt-4" or "moonshotai/kimi-k2.5" (legacy)
// Object format: {"provider": "moonshotai", "model": "kimi-k2.5"} (new)
type DepartmentModelConfig struct {
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}

func (m *AgentModelConfig) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		m.Primary = s
		m.Fallbacks = nil
		return nil
	}
	type raw struct {
		Primary   string   `json:"primary"`
		Fallbacks []string `json:"fallbacks"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	m.Primary = r.Primary
	m.Fallbacks = r.Fallbacks
	return nil
}

func (m AgentModelConfig) MarshalJSON() ([]byte, error) {
	if len(m.Fallbacks) == 0 && m.Primary != "" {
		return json.Marshal(m.Primary)
	}
	type raw struct {
		Primary   string   `json:"primary,omitempty"`
		Fallbacks []string `json:"fallbacks,omitempty"`
	}
	return json.Marshal(raw{Primary: m.Primary, Fallbacks: m.Fallbacks})
}

func (d *DepartmentModelConfig) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		// String format - could be just model name or provider/model
		if strings.Contains(s, "/") {
			// Legacy provider/model format
			parts := strings.SplitN(s, "/", 2)
			d.Provider = parts[0]
			d.Model = parts[1]
		} else {
			// Just model name - provider will be determined from context
			d.Model = s
		}
		return nil
	}

	// Object format
	type raw struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	d.Provider = r.Provider
	d.Model = r.Model
	return nil
}

func (d DepartmentModelConfig) MarshalJSON() ([]byte, error) {
	if d.Provider == "" {
		// If no provider, marshal as string (just the model)
		return json.Marshal(d.Model)
	}
	// Marshal as object
	type raw struct {
		Provider string `json:"provider,omitempty"`
		Model    string `json:"model,omitempty"`
	}
	return json.Marshal(raw{Provider: d.Provider, Model: d.Model})
}

// GetEffectiveModelName returns the effective model name for this department model config.
func (d *DepartmentModelConfig) GetEffectiveModelName() string {
	if d.Provider != "" && d.Model != "" {
		return d.Provider + "/" + d.Model
	}
	return d.Model
}

type AgentConfig struct {
	ID                string            `json:"id"`
	Default           bool              `json:"default,omitempty"`
	Name              string            `json:"name,omitempty"`
	Workspace         string            `json:"workspace,omitempty"`
	AgentDir          string            `json:"agentDir,omitempty"`          // Agent config directory (like OpenClaw)
	Model             *AgentModelConfig `json:"model,omitempty"`
	Skills            []string          `json:"skills,omitempty"`
	Subagents         *SubagentsConfig  `json:"subagents,omitempty"`
	Role              string            `json:"role,omitempty"`               // "manager", "planner", "executor", "specialist", "worker"
	ParentID          string            `json:"parent_id,omitempty"`          // parent agent ID for hierarchy
	Children          []string          `json:"children,omitempty"`           // child agent IDs
	Capabilities      []string          `json:"capabilities,omitempty"`       // agent capabilities for discovery
	CapabilityPrompts map[string]string `json:"capability_prompts,omitempty"` // dynamic system prompts per capability

	// Office UI specific fields (Phase 0.1/0.2)
	Department      string   `json:"department,omitempty"`       // Department the agent belongs to
	IsPermanent     bool     `json:"is_permanent,omitempty"`     // Whether this is a permanent agent
	Avatar          string   `json:"avatar,omitempty"`           // Avatar identifier or URL
	Responsibilities []string `json:"responsibilities,omitempty"` // List of responsibilities
	
	// Agent Team fields
	IsCoordinator   bool             `json:"is_coordinator,omitempty"` // Is this the team coordinator
	Prompt          string           `json:"prompt,omitempty"`         // Embedded persona prompt
}

// GetModel returns the effective model for the agent.
// It checks both Model.Primary and legacy string model field.
func (a *AgentConfig) GetModel() string {
	if a.Model != nil && strings.TrimSpace(a.Model.Primary) != "" {
		return strings.TrimSpace(a.Model.Primary)
	}
	return ""
}

// UnmarshalJSON implements custom JSON unmarshaling for AgentConfig
// to support both string and object formats for the "model" field.
func (a *AgentConfig) UnmarshalJSON(data []byte) error {
	// Create a type alias to avoid infinite recursion
	type Alias AgentConfig
	aux := &struct {
		*Alias
		Model interface{} `json:"model"`
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle model field - can be string or object
	if aux.Model != nil {
		switch v := aux.Model.(type) {
		case string:
			// String format: "model": "kimi-coding/kimi-for-coding"
			a.Model = &AgentModelConfig{Primary: v}
		case map[string]interface{}:
			// Object format: "model": {"primary": "...", "fallbacks": [...]}
			modelJSON, _ := json.Marshal(v)
			var modelCfg AgentModelConfig
			if err := json.Unmarshal(modelJSON, &modelCfg); err == nil {
				a.Model = &modelCfg
			}
		}
	}

	return nil
}

type SubagentsConfig struct {
	AllowAgents []string          `json:"allow_agents,omitempty"`
	Model       *AgentModelConfig `json:"model,omitempty"`
}

type PeerMatch struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type BindingMatch struct {
	Channel   string     `json:"channel"`
	AccountID string     `json:"account_id,omitempty"`
	Peer      *PeerMatch `json:"peer,omitempty"`
	GuildID   string     `json:"guild_id,omitempty"`
	TeamID    string     `json:"team_id,omitempty"`
}

type AgentBinding struct {
	AgentID string       `json:"agent_id"`
	Match   BindingMatch `json:"match"`
}

type SessionConfig struct {
	DMScope       string              `json:"dm_scope,omitempty"`
	IdentityLinks map[string][]string `json:"identity_links,omitempty"`
}

type AgentDefaults struct {
	// DEPRECATED: Workspace is now defined in WorkspaceConfig.Path.
	// All agents share the same workspace. Kept for backward compatibility.
	Workspace                 string   `json:"workspace,omitempty"             env:"PICOCLAW_AGENTS_DEFAULTS_WORKSPACE"`
	RestrictToWorkspace       bool     `json:"restrict_to_workspace"           env:"PICOCLAW_AGENTS_DEFAULTS_RESTRICT_TO_WORKSPACE"`
	AllowReadOutsideWorkspace bool     `json:"allow_read_outside_workspace"    env:"PICOCLAW_AGENTS_DEFAULTS_ALLOW_READ_OUTSIDE_WORKSPACE"`
	// Provider field for new format (provider + model).
	// When used with Model field, forms "provider/model" effective model name.
	// Also supports legacy format where Model contains full "provider/model" string.
	Provider                  string   `json:"provider,omitempty"              env:"PICOCLAW_AGENTS_DEFAULTS_PROVIDER"`
	// Model is either:
	// - New format: just the model identifier (e.g., "kimi-for-coding")
	// - Legacy format: full provider/model string (e.g., "kimi-coding/kimi-for-coding")
	Model                     string   `json:"model"                           env:"PICOCLAW_AGENTS_DEFAULTS_MODEL"`
	ModelName                 string   `json:"model_name,omitempty"            env:"PICOCLAW_AGENTS_DEFAULTS_MODEL_NAME"`
	ModelFallbacks            []string `json:"model_fallbacks,omitempty"`
	ImageModel                string   `json:"image_model,omitempty"           env:"PICOCLAW_AGENTS_DEFAULTS_IMAGE_MODEL"`
	ImageModelFallbacks       []string `json:"image_model_fallbacks,omitempty"`
	MaxTokens                 int      `json:"max_tokens"                      env:"PICOCLAW_AGENTS_DEFAULTS_MAX_TOKENS"`
	Temperature               *float64 `json:"temperature,omitempty"           env:"PICOCLAW_AGENTS_DEFAULTS_TEMPERATURE"`
	MaxToolIterations         int      `json:"max_tool_iterations"             env:"PICOCLAW_AGENTS_DEFAULTS_MAX_TOOL_ITERATIONS"`
	SummarizeMessageThreshold int      `json:"summarize_message_threshold"     env:"PICOCLAW_AGENTS_DEFAULTS_SUMMARIZE_MESSAGE_THRESHOLD"`
	SummarizeTokenPercent     int      `json:"summarize_token_percent"         env:"PICOCLAW_AGENTS_DEFAULTS_SUMMARIZE_TOKEN_PERCENT"`
	MaxMediaSize              int      `json:"max_media_size,omitempty"        env:"PICOCLAW_AGENTS_DEFAULTS_MAX_MEDIA_SIZE"`
}

const DefaultMaxMediaSize = 20 * 1024 * 1024 // 20 MB

func (d *AgentDefaults) GetMaxMediaSize() int {
	if d.MaxMediaSize > 0 {
		return d.MaxMediaSize
	}
	return DefaultMaxMediaSize
}

// GetModelName returns the effective model name for the agent defaults.
// It handles both new format (provider + model) and legacy format (full model string).
func (d *AgentDefaults) GetModelName() string {
	if d.ModelName != "" {
		return d.ModelName
	}

	// New format: provider + model
	if d.Provider != "" && d.Model != "" {
		return d.Provider + "/" + d.Model
	}

	// Legacy format: full model string
	if d.Model != "" {
		return d.Model
	}

	// Fallback to empty string
	return ""
}

// GetEffectiveModelID returns the effective model identifier for the agent defaults.
// It handles both new format (provider + model) and legacy format (full model string).
func (d *AgentDefaults) GetEffectiveModelID() string {
	// New format: provider + model
	if d.Provider != "" && d.Model != "" {
		return d.Model
	}

	// Legacy format: full model string
	if strings.Contains(d.Model, "/") {
		parts := strings.SplitN(d.Model, "/", 2)
		if len(parts) > 1 {
			return parts[1]
		}
	}

	// Just model name or empty
	return d.Model
}

// GetEffectiveProvider returns the effective provider for the agent defaults.
// It handles both new format (provider + model) and legacy format (full model string).
func (d *AgentDefaults) GetEffectiveProvider() string {
	// New format: explicit provider
	if d.Provider != "" {
		return d.Provider
	}

	// Legacy format: extract from model string
	if strings.Contains(d.Model, "/") {
		parts := strings.SplitN(d.Model, "/", 2)
		return parts[0]
	}

	// Default to "openai" for models without provider prefix
	return "openai"
}

// InitializeAgentDirs initializes agentDir for all agents if not set.
// Similar to OpenClaw's agent directory structure.
func (cfg *Config) InitializeAgentDirs(homePath string) error {
	agentsDir := filepath.Join(homePath, "agents")
	
	for i := range cfg.Agents.List {
		agent := &cfg.Agents.List[i]
		
		// Set default agentDir if not specified
		// FORCED: All agents use 'main' directory
		if agent.AgentDir == "" {
			agent.AgentDir = filepath.Join(agentsDir, "main")
		}
		
		// Ensure agent directory exists
		if err := os.MkdirAll(agent.AgentDir, 0755); err != nil {
			return fmt.Errorf("failed to create agent directory for %s: %w", agent.ID, err)
		}
	}
	
	return nil
}

type ChannelsConfig struct {
	WhatsApp   WhatsAppConfig   `json:"whatsapp"`
	Telegram   TelegramConfig   `json:"telegram"`
	Feishu     FeishuConfig     `json:"feishu"`
	Discord    DiscordConfig    `json:"discord"`
	MaixCam    MaixCamConfig    `json:"maixcam"`
	QQ         QQConfig         `json:"qq"`
	DingTalk   DingTalkConfig   `json:"dingtalk"`
	Slack      SlackConfig      `json:"slack"`
	LINE       LINEConfig       `json:"line"`
	OneBot     OneBotConfig     `json:"onebot"`
	WeCom      WeComConfig      `json:"wecom"`
	WeComApp   WeComAppConfig   `json:"wecom_app"`
	WeComAIBot WeComAIBotConfig `json:"wecom_aibot"`
	Pico       PicoConfig       `json:"pico"`
}

// GroupTriggerConfig controls when the bot responds in group chats.
type GroupTriggerConfig struct {
	MentionOnly bool     `json:"mention_only,omitempty"`
	Prefixes    []string `json:"prefixes,omitempty"`
}

// TypingConfig controls typing indicator behavior (Phase 10).
type TypingConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

// PlaceholderConfig controls placeholder message behavior (Phase 10).
type PlaceholderConfig struct {
	Enabled bool   `json:"enabled,omitempty"`
	Text    string `json:"text,omitempty"`
}

type WhatsAppConfig struct {
	Enabled            bool                `json:"enabled"              env:"PICOCLAW_CHANNELS_WHATSAPP_ENABLED"`
	BridgeURL          string              `json:"bridge_url"           env:"PICOCLAW_CHANNELS_WHATSAPP_BRIDGE_URL"`
	UseNative          bool                `json:"use_native"           env:"PICOCLAW_CHANNELS_WHATSAPP_USE_NATIVE"`
	SessionStorePath   string              `json:"session_store_path"   env:"PICOCLAW_CHANNELS_WHATSAPP_SESSION_STORE_PATH"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"           env:"PICOCLAW_CHANNELS_WHATSAPP_ALLOW_FROM"`
	ReasoningChannelID string              `json:"reasoning_channel_id" env:"PICOCLAW_CHANNELS_WHATSAPP_REASONING_CHANNEL_ID"`
}

type TelegramConfig struct {
	Enabled            bool                `json:"enabled"                 env:"PICOCLAW_CHANNELS_TELEGRAM_ENABLED"`
	Token              string              `json:"token"                   env:"PICOCLAW_CHANNELS_TELEGRAM_TOKEN"`
	BaseURL            string              `json:"base_url"                env:"PICOCLAW_CHANNELS_TELEGRAM_BASE_URL"`
	Proxy              string              `json:"proxy"                   env:"PICOCLAW_CHANNELS_TELEGRAM_PROXY"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"              env:"PICOCLAW_CHANNELS_TELEGRAM_ALLOW_FROM"`
	GroupTrigger       GroupTriggerConfig  `json:"group_trigger,omitempty"`
	Typing             TypingConfig        `json:"typing,omitempty"`
	Placeholder        PlaceholderConfig   `json:"placeholder,omitempty"`
	ReasoningChannelID string              `json:"reasoning_channel_id"    env:"PICOCLAW_CHANNELS_TELEGRAM_REASONING_CHANNEL_ID"`
}

type FeishuConfig struct {
	Enabled            bool                `json:"enabled"                 env:"PICOCLAW_CHANNELS_FEISHU_ENABLED"`
	AppID              string              `json:"app_id"                  env:"PICOCLAW_CHANNELS_FEISHU_APP_ID"`
	AppSecret          string              `json:"app_secret"              env:"PICOCLAW_CHANNELS_FEISHU_APP_SECRET"`
	EncryptKey         string              `json:"encrypt_key"             env:"PICOCLAW_CHANNELS_FEISHU_ENCRYPT_KEY"`
	VerificationToken  string              `json:"verification_token"      env:"PICOCLAW_CHANNELS_FEISHU_VERIFICATION_TOKEN"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"              env:"PICOCLAW_CHANNELS_FEISHU_ALLOW_FROM"`
	GroupTrigger       GroupTriggerConfig  `json:"group_trigger,omitempty"`
	Placeholder        PlaceholderConfig   `json:"placeholder,omitempty"`
	ReasoningChannelID string              `json:"reasoning_channel_id"    env:"PICOCLAW_CHANNELS_FEISHU_REASONING_CHANNEL_ID"`
}

type DiscordConfig struct {
	Enabled            bool                `json:"enabled"                 env:"PICOCLAW_CHANNELS_DISCORD_ENABLED"`
	Token              string              `json:"token"                   env:"PICOCLAW_CHANNELS_DISCORD_TOKEN"`
	Proxy              string              `json:"proxy"                   env:"PICOCLAW_CHANNELS_DISCORD_PROXY"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"              env:"PICOCLAW_CHANNELS_DISCORD_ALLOW_FROM"`
	MentionOnly        bool                `json:"mention_only"            env:"PICOCLAW_CHANNELS_DISCORD_MENTION_ONLY"`
	GroupTrigger       GroupTriggerConfig  `json:"group_trigger,omitempty"`
	Typing             TypingConfig        `json:"typing,omitempty"`
	Placeholder        PlaceholderConfig   `json:"placeholder,omitempty"`
	ReasoningChannelID string              `json:"reasoning_channel_id"    env:"PICOCLAW_CHANNELS_DISCORD_REASONING_CHANNEL_ID"`
}

type MaixCamConfig struct {
	Enabled            bool                `json:"enabled"              env:"PICOCLAW_CHANNELS_MAIXCAM_ENABLED"`
	Host               string              `json:"host"                 env:"PICOCLAW_CHANNELS_MAIXCAM_HOST"`
	Port               int                 `json:"port"                 env:"PICOCLAW_CHANNELS_MAIXCAM_PORT"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"           env:"PICOCLAW_CHANNELS_MAIXCAM_ALLOW_FROM"`
	ReasoningChannelID string              `json:"reasoning_channel_id" env:"PICOCLAW_CHANNELS_MAIXCAM_REASONING_CHANNEL_ID"`
}

type QQConfig struct {
	Enabled            bool                `json:"enabled"                 env:"PICOCLAW_CHANNELS_QQ_ENABLED"`
	AppID              string              `json:"app_id"                  env:"PICOCLAW_CHANNELS_QQ_APP_ID"`
	AppSecret          string              `json:"app_secret"              env:"PICOCLAW_CHANNELS_QQ_APP_SECRET"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"              env:"PICOCLAW_CHANNELS_QQ_ALLOW_FROM"`
	GroupTrigger       GroupTriggerConfig  `json:"group_trigger,omitempty"`
	ReasoningChannelID string              `json:"reasoning_channel_id"    env:"PICOCLAW_CHANNELS_QQ_REASONING_CHANNEL_ID"`
}

type DingTalkConfig struct {
	Enabled            bool                `json:"enabled"                 env:"PICOCLAW_CHANNELS_DINGTALK_ENABLED"`
	ClientID           string              `json:"client_id"               env:"PICOCLAW_CHANNELS_DINGTALK_CLIENT_ID"`
	ClientSecret       string              `json:"client_secret"           env:"PICOCLAW_CHANNELS_DINGTALK_CLIENT_SECRET"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"              env:"PICOCLAW_CHANNELS_DINGTALK_ALLOW_FROM"`
	GroupTrigger       GroupTriggerConfig  `json:"group_trigger,omitempty"`
	ReasoningChannelID string              `json:"reasoning_channel_id"    env:"PICOCLAW_CHANNELS_DINGTALK_REASONING_CHANNEL_ID"`
}

type SlackConfig struct {
	Enabled            bool                `json:"enabled"                 env:"PICOCLAW_CHANNELS_SLACK_ENABLED"`
	BotToken           string              `json:"bot_token"               env:"PICOCLAW_CHANNELS_SLACK_BOT_TOKEN"`
	AppToken           string              `json:"app_token"               env:"PICOCLAW_CHANNELS_SLACK_APP_TOKEN"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"              env:"PICOCLAW_CHANNELS_SLACK_ALLOW_FROM"`
	GroupTrigger       GroupTriggerConfig  `json:"group_trigger,omitempty"`
	Typing             TypingConfig        `json:"typing,omitempty"`
	Placeholder        PlaceholderConfig   `json:"placeholder,omitempty"`
	ReasoningChannelID string              `json:"reasoning_channel_id"    env:"PICOCLAW_CHANNELS_SLACK_REASONING_CHANNEL_ID"`
}

type LINEConfig struct {
	Enabled            bool                `json:"enabled"                 env:"PICOCLAW_CHANNELS_LINE_ENABLED"`
	ChannelSecret      string              `json:"channel_secret"          env:"PICOCLAW_CHANNELS_LINE_CHANNEL_SECRET"`
	ChannelAccessToken string              `json:"channel_access_token"    env:"PICOCLAW_CHANNELS_LINE_CHANNEL_ACCESS_TOKEN"`
	WebhookHost        string              `json:"webhook_host"            env:"PICOCLAW_CHANNELS_LINE_WEBHOOK_HOST"`
	WebhookPort        int                 `json:"webhook_port"            env:"PICOCLAW_CHANNELS_LINE_WEBHOOK_PORT"`
	WebhookPath        string              `json:"webhook_path"            env:"PICOCLAW_CHANNELS_LINE_WEBHOOK_PATH"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"              env:"PICOCLAW_CHANNELS_LINE_ALLOW_FROM"`
	GroupTrigger       GroupTriggerConfig  `json:"group_trigger,omitempty"`
	Typing             TypingConfig        `json:"typing,omitempty"`
	Placeholder        PlaceholderConfig   `json:"placeholder,omitempty"`
	ReasoningChannelID string              `json:"reasoning_channel_id"    env:"PICOCLAW_CHANNELS_LINE_REASONING_CHANNEL_ID"`
}

type OneBotConfig struct {
	Enabled            bool                `json:"enabled"                 env:"PICOCLAW_CHANNELS_ONEBOT_ENABLED"`
	WSUrl              string              `json:"ws_url"                  env:"PICOCLAW_CHANNELS_ONEBOT_WS_URL"`
	AccessToken        string              `json:"access_token"            env:"PICOCLAW_CHANNELS_ONEBOT_ACCESS_TOKEN"`
	ReconnectInterval  int                 `json:"reconnect_interval"      env:"PICOCLAW_CHANNELS_ONEBOT_RECONNECT_INTERVAL"`
	GroupTriggerPrefix []string            `json:"group_trigger_prefix"    env:"PICOCLAW_CHANNELS_ONEBOT_GROUP_TRIGGER_PREFIX"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"              env:"PICOCLAW_CHANNELS_ONEBOT_ALLOW_FROM"`
	GroupTrigger       GroupTriggerConfig  `json:"group_trigger,omitempty"`
	Typing             TypingConfig        `json:"typing,omitempty"`
	Placeholder        PlaceholderConfig   `json:"placeholder,omitempty"`
	ReasoningChannelID string              `json:"reasoning_channel_id"    env:"PICOCLAW_CHANNELS_ONEBOT_REASONING_CHANNEL_ID"`
}

type WeComConfig struct {
	Enabled            bool                `json:"enabled"                 env:"PICOCLAW_CHANNELS_WECOM_ENABLED"`
	Token              string              `json:"token"                   env:"PICOCLAW_CHANNELS_WECOM_TOKEN"`
	EncodingAESKey     string              `json:"encoding_aes_key"        env:"PICOCLAW_CHANNELS_WECOM_ENCODING_AES_KEY"`
	WebhookURL         string              `json:"webhook_url"             env:"PICOCLAW_CHANNELS_WECOM_WEBHOOK_URL"`
	WebhookHost        string              `json:"webhook_host"            env:"PICOCLAW_CHANNELS_WECOM_WEBHOOK_HOST"`
	WebhookPort        int                 `json:"webhook_port"            env:"PICOCLAW_CHANNELS_WECOM_WEBHOOK_PORT"`
	WebhookPath        string              `json:"webhook_path"            env:"PICOCLAW_CHANNELS_WECOM_WEBHOOK_PATH"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"              env:"PICOCLAW_CHANNELS_WECOM_ALLOW_FROM"`
	ReplyTimeout       int                 `json:"reply_timeout"           env:"PICOCLAW_CHANNELS_WECOM_REPLY_TIMEOUT"`
	GroupTrigger       GroupTriggerConfig  `json:"group_trigger,omitempty"`
	ReasoningChannelID string              `json:"reasoning_channel_id"    env:"PICOCLAW_CHANNELS_WECOM_REASONING_CHANNEL_ID"`
}

type WeComAppConfig struct {
	Enabled            bool                `json:"enabled"                 env:"PICOCLAW_CHANNELS_WECOM_APP_ENABLED"`
	CorpID             string              `json:"corp_id"                 env:"PICOCLAW_CHANNELS_WECOM_APP_CORP_ID"`
	CorpSecret         string              `json:"corp_secret"             env:"PICOCLAW_CHANNELS_WECOM_APP_CORP_SECRET"`
	AgentID            int64               `json:"agent_id"                env:"PICOCLAW_CHANNELS_WECOM_APP_AGENT_ID"`
	Token              string              `json:"token"                   env:"PICOCLAW_CHANNELS_WECOM_APP_TOKEN"`
	EncodingAESKey     string              `json:"encoding_aes_key"        env:"PICOCLAW_CHANNELS_WECOM_APP_ENCODING_AES_KEY"`
	WebhookHost        string              `json:"webhook_host"            env:"PICOCLAW_CHANNELS_WECOM_APP_WEBHOOK_HOST"`
	WebhookPort        int                 `json:"webhook_port"            env:"PICOCLAW_CHANNELS_WECOM_APP_WEBHOOK_PORT"`
	WebhookPath        string              `json:"webhook_path"            env:"PICOCLAW_CHANNELS_WECOM_APP_WEBHOOK_PATH"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"              env:"PICOCLAW_CHANNELS_WECOM_APP_ALLOW_FROM"`
	ReplyTimeout       int                 `json:"reply_timeout"           env:"PICOCLAW_CHANNELS_WECOM_APP_REPLY_TIMEOUT"`
	GroupTrigger       GroupTriggerConfig  `json:"group_trigger,omitempty"`
	ReasoningChannelID string              `json:"reasoning_channel_id"    env:"PICOCLAW_CHANNELS_WECOM_APP_REASONING_CHANNEL_ID"`
}

type WeComAIBotConfig struct {
	Enabled            bool                `json:"enabled"              env:"PICOCLAW_CHANNELS_WECOM_AIBOT_ENABLED"`
	Token              string              `json:"token"                env:"PICOCLAW_CHANNELS_WECOM_AIBOT_TOKEN"`
	EncodingAESKey     string              `json:"encoding_aes_key"     env:"PICOCLAW_CHANNELS_WECOM_AIBOT_ENCODING_AES_KEY"`
	WebhookPath        string              `json:"webhook_path"         env:"PICOCLAW_CHANNELS_WECOM_AIBOT_WEBHOOK_PATH"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"           env:"PICOCLAW_CHANNELS_WECOM_AIBOT_ALLOW_FROM"`
	ReplyTimeout       int                 `json:"reply_timeout"        env:"PICOCLAW_CHANNELS_WECOM_AIBOT_REPLY_TIMEOUT"`
	MaxSteps           int                 `json:"max_steps"            env:"PICOCLAW_CHANNELS_WECOM_AIBOT_MAX_STEPS"`       // Maximum streaming steps
	WelcomeMessage     string              `json:"welcome_message"      env:"PICOCLAW_CHANNELS_WECOM_AIBOT_WELCOME_MESSAGE"` // Sent on enter_chat event; empty = no welcome
	ReasoningChannelID string              `json:"reasoning_channel_id" env:"PICOCLAW_CHANNELS_WECOM_AIBOT_REASONING_CHANNEL_ID"`
}

type PicoConfig struct {
	Enabled         bool                `json:"enabled"                     env:"PICOCLAW_CHANNELS_PICO_ENABLED"`
	Token           string              `json:"token"                       env:"PICOCLAW_CHANNELS_PICO_TOKEN"`
	AllowTokenQuery bool                `json:"allow_token_query,omitempty"`
	AllowOrigins    []string            `json:"allow_origins,omitempty"`
	PingInterval    int                 `json:"ping_interval,omitempty"`
	ReadTimeout     int                 `json:"read_timeout,omitempty"`
	WriteTimeout    int                 `json:"write_timeout,omitempty"`
	MaxConnections  int                 `json:"max_connections,omitempty"`
	AllowFrom       FlexibleStringSlice `json:"allow_from"                  env:"PICOCLAW_CHANNELS_PICO_ALLOW_FROM"`
	Placeholder     PlaceholderConfig   `json:"placeholder,omitempty"`
}

type HeartbeatConfig struct {
	Enabled  bool `json:"enabled"  env:"PICOCLAW_HEARTBEAT_ENABLED"`
	Interval int  `json:"interval" env:"PICOCLAW_HEARTBEAT_INTERVAL"` // minutes, min 5
}

type DevicesConfig struct {
	Enabled    bool `json:"enabled"     env:"PICOCLAW_DEVICES_ENABLED"`
	MonitorUSB bool `json:"monitor_usb" env:"PICOCLAW_DEVICES_MONITOR_USB"`
}

// ProvidersConfig is DEPRECATED. Use ModelList instead.
// Kept for backward compatibility - will be auto-migrated to ModelList on load.
type ProvidersConfig struct {
	Anthropic     DeprecatedProviderConfig `json:"anthropic,omitempty"`
	OpenAI        DeprecatedOpenAIProvider `json:"openai,omitempty"`
	LiteLLM       DeprecatedProviderConfig `json:"litellm,omitempty"`
	OpenRouter    DeprecatedProviderConfig `json:"openrouter,omitempty"`
	Groq          DeprecatedProviderConfig `json:"groq,omitempty"`
	Zhipu         DeprecatedProviderConfig `json:"zhipu,omitempty"`
	VLLM          DeprecatedProviderConfig `json:"vllm,omitempty"`
	Gemini        DeprecatedProviderConfig `json:"gemini,omitempty"`
	Nvidia        DeprecatedProviderConfig `json:"nvidia,omitempty"`
	Ollama        DeprecatedProviderConfig `json:"ollama,omitempty"`
	Moonshot      DeprecatedProviderConfig `json:"moonshot,omitempty"`
	ShengSuanYun  DeprecatedProviderConfig `json:"shengsuanyun,omitempty"`
	DeepSeek      DeprecatedProviderConfig `json:"deepseek,omitempty"`
	Cerebras      DeprecatedProviderConfig `json:"cerebras,omitempty"`
	VolcEngine    DeprecatedProviderConfig `json:"volcengine,omitempty"`
	GitHubCopilot DeprecatedProviderConfig `json:"github_copilot,omitempty"`
	Antigravity   DeprecatedProviderConfig `json:"antigravity,omitempty"`
	Qwen          DeprecatedProviderConfig `json:"qwen,omitempty"`
	Mistral       DeprecatedProviderConfig `json:"mistral,omitempty"`
	Bailian       DeprecatedProviderConfig `json:"bailian,omitempty"`
	MiniMaxPortal DeprecatedProviderConfig `json:"minimax_portal,omitempty"`
	KimiCoding    DeprecatedProviderConfig `json:"kimi_coding,omitempty"`
}

// DeprecatedProviderConfig is DEPRECATED. Use ModelConfig in ModelList instead.
type DeprecatedProviderConfig struct {
	APIKey         string `json:"api_key,omitempty"`      // DEPRECATED: Move to ModelConfig
	APIBase        string `json:"api_base,omitempty"`     // DEPRECATED: Move to ModelConfig  
	Proxy          string `json:"proxy,omitempty"`        // DEPRECATED: Move to ModelConfig
	RequestTimeout int    `json:"request_timeout,omitempty"` // DEPRECATED: Move to ModelConfig
	AuthMethod     string `json:"auth_method,omitempty"`  // DEPRECATED: Move to ModelConfig
	ConnectMode    string `json:"connect_mode,omitempty"` // DEPRECATED: Move to ModelConfig
}

// DeprecatedOpenAIProvider is DEPRECATED. Use ModelConfig in ModelList instead.
type DeprecatedOpenAIProvider struct {
	APIKey         string `json:"api_key,omitempty"`         // DEPRECATED: Move to ModelConfig
	APIBase        string `json:"api_base,omitempty"`        // DEPRECATED: Move to ModelConfig
	Proxy          string `json:"proxy,omitempty"`           // DEPRECATED: Move to ModelConfig
	RequestTimeout int    `json:"request_timeout,omitempty"` // DEPRECATED: Move to ModelConfig
	AuthMethod     string `json:"auth_method,omitempty"`     // DEPRECATED: Move to ModelConfig
	ConnectMode    string `json:"connect_mode,omitempty"`    // DEPRECATED: Move to ModelConfig
	WebSearch      bool   `json:"web_search,omitempty"`      // DEPRECATED: Move to ModelConfig
}

// IsEmpty checks if any deprecated provider config has values (for migration detection)
// DEPRECATED: ProvidersConfig is deprecated, use ModelList instead
func (p ProvidersConfig) IsEmpty() bool {
	return p.Anthropic.isEmpty() && p.OpenAI.isEmpty() && p.LiteLLM.isEmpty() &&
		p.OpenRouter.isEmpty() && p.Groq.isEmpty() && p.Zhipu.isEmpty() &&
		p.VLLM.isEmpty() && p.Gemini.isEmpty() && p.Nvidia.isEmpty() &&
		p.Ollama.isEmpty() && p.Moonshot.isEmpty() && p.ShengSuanYun.isEmpty() &&
		p.DeepSeek.isEmpty() && p.Cerebras.isEmpty() && p.VolcEngine.isEmpty() &&
		p.GitHubCopilot.isEmpty() && p.Antigravity.isEmpty() && p.Qwen.isEmpty() &&
		p.Mistral.isEmpty() && p.Bailian.isEmpty() && p.MiniMaxPortal.isEmpty() &&
		p.KimiCoding.isEmpty()
}

func (d DeprecatedProviderConfig) isEmpty() bool {
	return d.APIKey == "" && d.APIBase == "" && d.AuthMethod == ""
}

func (d DeprecatedOpenAIProvider) isEmpty() bool {
	return d.APIKey == "" && d.APIBase == "" && d.Proxy == "" && 
		d.RequestTimeout == 0 && d.AuthMethod == "" && d.ConnectMode == "" && !d.WebSearch
}

// MarshalJSON implements custom JSON marshaling for ProvidersConfig
// to omit the entire section when empty
func (p ProvidersConfig) MarshalJSON() ([]byte, error) {
	if p.IsEmpty() {
		return []byte("null"), nil
	}
	type Alias ProvidersConfig
	return json.Marshal((*Alias)(&p))
}

// ProviderConfig is DEPRECATED. Use ModelConfig in ModelList instead.
type ProviderConfig = DeprecatedProviderConfig

// OpenAIProviderConfig is DEPRECATED. Use ModelConfig in ModelList instead.
type OpenAIProviderConfig = DeprecatedOpenAIProvider

// ModelConfig represents a model-centric provider configuration.
// It allows adding new providers (especially OpenAI-compatible ones) via configuration only.
// Supports both explicit provider/model format and legacy protocol/model-identifier format.
//
// New format (preferred):
//   {
//     "provider": "kimi-coding",
//     "model": "kimi-for-coding",
//     "api_key": "...",
//     "api_base": "...",
//     "temperature": 0.7,
//     "top_p": 0.9,
//     "enable_thinking": true
//   }
//
// Legacy format (backward compatible):
//   {
//     "model_name": "kimi-for-coding",
//     "model": "kimi-coding/kimi-for-coding",
//     "api_key": "...",
//     "api_base": "..."
//   }
type ModelConfig struct {
	// Explicit provider field (new format, preferred)
	// Specifies the provider name (e.g., "kimi-coding", "moonshotai", "openai")
	Provider string `json:"provider,omitempty"`

	// Model identifier (new format) or full provider/model string (legacy format)
	// In new format: just the model name (e.g., "kimi-for-coding")
	// In legacy format: provider/model string (e.g., "kimi-coding/kimi-for-coding")
	Model string `json:"model"`

	// User-facing alias for the model (optional in new format, required in legacy)
	// If not provided in new format, will be auto-generated as "provider/model"
	ModelName string `json:"model_name,omitempty"`

	// HTTP-based providers
	APIBase string `json:"api_base,omitempty"` // API endpoint URL
	APIKey  string `json:"api_key"`            // API authentication key
	Proxy   string `json:"proxy,omitempty"`    // HTTP proxy URL

	// Special providers (CLI-based, OAuth, etc.)
	AuthMethod  string `json:"auth_method,omitempty"`  // Authentication method: oauth, token
	ConnectMode string `json:"connect_mode,omitempty"` // Connection mode: stdio, grpc
	Workspace   string `json:"workspace,omitempty"`    // Workspace path for CLI-based providers

	// Model parameters
	Temperature   *float64 `json:"temperature,omitempty"`   // Temperature for sampling (0.0 to 2.0)
	TopP          *float64 `json:"top_p,omitempty"`        // Top-p sampling (0.0 to 1.0)
	EnableThinking *bool   `json:"enable_thinking,omitempty"` // Enable thinking mode (if supported by provider)

	// Optional optimizations
	RPM            int    `json:"rpm,omitempty"`              // Requests per minute limit
	MaxTokensField string `json:"max_tokens_field,omitempty"` // Field name for max tokens (e.g., "max_completion_tokens")
	RequestTimeout int    `json:"request_timeout,omitempty"`
	MaxConcurrent  int    `json:"max_concurrent,omitempty"` // Max concurrent API requests (0 = unlimited, use for APIs with low limits like DashScope)
}

// Validate checks if the ModelConfig has all required fields.
func (c *ModelConfig) Validate() error {
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}

	// For new format (explicit provider), model_name is optional
	// For legacy format (no provider), model_name is required
	if c.Provider == "" && c.ModelName == "" {
		return fmt.Errorf("model_name is required when provider is not specified")
	}

	return nil
}

// GetEffectiveModelName returns the effective model name for this configuration.
// If ModelName is set, returns it. Otherwise, returns "provider/model".
func (c *ModelConfig) GetEffectiveModelName() string {
	if c.ModelName != "" {
		return c.ModelName
	}
	if c.Provider != "" {
		return c.Provider + "/" + c.Model
	}
	// Legacy format - return the full model string
	return c.Model
}

// GetEffectiveProvider returns the effective provider name.
// If Provider is set, returns it. Otherwise, extracts from Model string.
func (c *ModelConfig) GetEffectiveProvider() string {
	if c.Provider != "" {
		return c.Provider
	}
	// Extract provider from legacy model string (e.g., "kimi-coding/kimi-for-coding" -> "kimi-coding")
	if strings.Contains(c.Model, "/") {
		parts := strings.SplitN(c.Model, "/", 2)
		return parts[0]
	}
	// Default to "openai" for models without provider prefix
	return "openai"
}

// GetEffectiveModelID returns the effective model identifier.
// If Provider is set, returns Model. Otherwise, extracts model ID from Model string.
func (c *ModelConfig) GetEffectiveModelID() string {
	if c.Provider != "" {
		return c.Model
	}
	// Extract model ID from legacy model string (e.g., "kimi-coding/kimi-for-coding" -> "kimi-for-coding")
	if strings.Contains(c.Model, "/") {
		parts := strings.SplitN(c.Model, "/", 2)
		if len(parts) > 1 {
			return parts[1]
		}
	}
	// Return full model string for models without provider prefix
	return c.Model
}

type GatewayConfig struct {
	Host      string `json:"host" env:"PICOCLAW_GATEWAY_HOST"`
	Port      int    `json:"port" env:"PICOCLAW_GATEWAY_PORT"`
	UIEnabled bool   `json:"ui_enabled" env:"PICOCLAW_GATEWAY_UI_ENABLED"`
}

type BraveConfig struct {
	Enabled    bool   `json:"enabled"     env:"PICOCLAW_TOOLS_WEB_BRAVE_ENABLED"`
	APIKey     string `json:"api_key"     env:"PICOCLAW_TOOLS_WEB_BRAVE_API_KEY"`
	MaxResults int    `json:"max_results" env:"PICOCLAW_TOOLS_WEB_BRAVE_MAX_RESULTS"`
}

type TavilyConfig struct {
	Enabled    bool   `json:"enabled"     env:"PICOCLAW_TOOLS_WEB_TAVILY_ENABLED"`
	APIKey     string `json:"api_key"     env:"PICOCLAW_TOOLS_WEB_TAVILY_API_KEY"`
	BaseURL    string `json:"base_url"    env:"PICOCLAW_TOOLS_WEB_TAVILY_BASE_URL"`
	MaxResults int    `json:"max_results" env:"PICOCLAW_TOOLS_WEB_TAVILY_MAX_RESULTS"`
}

type DuckDuckGoConfig struct {
	Enabled    bool `json:"enabled"     env:"PICOCLAW_TOOLS_WEB_DUCKDUCKGO_ENABLED"`
	MaxResults int  `json:"max_results" env:"PICOCLAW_TOOLS_WEB_DUCKDUCKGO_MAX_RESULTS"`
}

type PerplexityConfig struct {
	Enabled    bool   `json:"enabled"     env:"PICOCLAW_TOOLS_WEB_PERPLEXITY_ENABLED"`
	APIKey     string `json:"api_key"     env:"PICOCLAW_TOOLS_WEB_PERPLEXITY_API_KEY"`
	MaxResults int    `json:"max_results" env:"PICOCLAW_TOOLS_WEB_PERPLEXITY_MAX_RESULTS"`
}

type GLMSearchConfig struct {
	Enabled bool   `json:"enabled"  env:"PICOCLAW_TOOLS_WEB_GLM_ENABLED"`
	APIKey  string `json:"api_key"  env:"PICOCLAW_TOOLS_WEB_GLM_API_KEY"`
	BaseURL string `json:"base_url" env:"PICOCLAW_TOOLS_WEB_GLM_BASE_URL"`
	// SearchEngine specifies the search backend: "search_std" (default),
	// "search_pro", "search_pro_sogou", or "search_pro_quark".
	SearchEngine string `json:"search_engine" env:"PICOCLAW_TOOLS_WEB_GLM_SEARCH_ENGINE"`
	MaxResults   int    `json:"max_results"   env:"PICOCLAW_TOOLS_WEB_GLM_MAX_RESULTS"`
}

type WebToolsConfig struct {
	Brave      BraveConfig      `json:"brave"`
	Tavily     TavilyConfig     `json:"tavily"`
	DuckDuckGo DuckDuckGoConfig `json:"duckduckgo"`
	Perplexity PerplexityConfig `json:"perplexity"`
	GLMSearch  GLMSearchConfig  `json:"glm_search"`
	// Proxy is an optional proxy URL for web tools (http/https/socks5/socks5h).
	// For authenticated proxies, prefer HTTP_PROXY/HTTPS_PROXY env vars instead of embedding credentials in config.
	Proxy           string `json:"proxy,omitempty"             env:"PICOCLAW_TOOLS_WEB_PROXY"`
	FetchLimitBytes int64  `json:"fetch_limit_bytes,omitempty" env:"PICOCLAW_TOOLS_WEB_FETCH_LIMIT_BYTES"`
}

type CronToolsConfig struct {
	ExecTimeoutMinutes int `json:"exec_timeout_minutes" env:"PICOCLAW_TOOLS_CRON_EXEC_TIMEOUT_MINUTES"` // 0 means no timeout
}

type ExecConfig struct {
	EnableDenyPatterns  bool     `json:"enable_deny_patterns"  env:"PICOCLAW_TOOLS_EXEC_ENABLE_DENY_PATTERNS"`
	CustomDenyPatterns  []string `json:"custom_deny_patterns"  env:"PICOCLAW_TOOLS_EXEC_CUSTOM_DENY_PATTERNS"`
	CustomAllowPatterns []string `json:"custom_allow_patterns" env:"PICOCLAW_TOOLS_EXEC_CUSTOM_ALLOW_PATTERNS"`
	// SafetyLevel controls which deny patterns are active:
	//   "strict"     — all patterns (legacy default, very restrictive)
	//   "balanced"   — critical + cautious patterns (recommended)
	//   "permissive" — critical only (blocks rm -rf, disk wipe, fork bomb, curl|bash)
	// Empty string defaults to "balanced".
	SafetyLevel string `json:"safety_level,omitempty" env:"PICOCLAW_TOOLS_EXEC_SAFETY_LEVEL"`
}

type MediaCleanupConfig struct {
	Enabled  bool `json:"enabled"          env:"PICOCLAW_MEDIA_CLEANUP_ENABLED"`
	MaxAge   int  `json:"max_age_minutes"  env:"PICOCLAW_MEDIA_CLEANUP_MAX_AGE"`
	Interval int  `json:"interval_minutes" env:"PICOCLAW_MEDIA_CLEANUP_INTERVAL"`
}

type ToolsConfig struct {
	AllowReadPaths  []string           `json:"allow_read_paths"  env:"PICOCLAW_TOOLS_ALLOW_READ_PATHS"`
	AllowWritePaths []string           `json:"allow_write_paths" env:"PICOCLAW_TOOLS_ALLOW_WRITE_PATHS"`
	Web             WebToolsConfig     `json:"web"`
	Cron            CronToolsConfig    `json:"cron"`
	Exec            ExecConfig         `json:"exec"`
	Skills          SkillsToolsConfig  `json:"skills"`
	MediaCleanup    MediaCleanupConfig `json:"media_cleanup"`
	MCP             MCPConfig          `json:"mcp"`
}

type SkillsToolsConfig struct {
	Registries            SkillsRegistriesConfig `json:"registries"`
	MaxConcurrentSearches int                    `json:"max_concurrent_searches" env:"PICOCLAW_SKILLS_MAX_CONCURRENT_SEARCHES"`
	SearchCache           SearchCacheConfig      `json:"search_cache"`
}

type SearchCacheConfig struct {
	MaxSize    int `json:"max_size"    env:"PICOCLAW_SKILLS_SEARCH_CACHE_MAX_SIZE"`
	TTLSeconds int `json:"ttl_seconds" env:"PICOCLAW_SKILLS_SEARCH_CACHE_TTL_SECONDS"`
}

type SkillsRegistriesConfig struct {
	ClawHub ClawHubRegistryConfig `json:"clawhub"`
}

type ClawHubRegistryConfig struct {
	Enabled         bool   `json:"enabled"           env:"PICOCLAW_SKILLS_REGISTRIES_CLAWHUB_ENABLED"`
	BaseURL         string `json:"base_url"          env:"PICOCLAW_SKILLS_REGISTRIES_CLAWHUB_BASE_URL"`
	AuthToken       string `json:"auth_token"        env:"PICOCLAW_SKILLS_REGISTRIES_CLAWHUB_AUTH_TOKEN"`
	SearchPath      string `json:"search_path"       env:"PICOCLAW_SKILLS_REGISTRIES_CLAWHUB_SEARCH_PATH"`
	SkillsPath      string `json:"skills_path"       env:"PICOCLAW_SKILLS_REGISTRIES_CLAWHUB_SKILLS_PATH"`
	DownloadPath    string `json:"download_path"     env:"PICOCLAW_SKILLS_REGISTRIES_CLAWHUB_DOWNLOAD_PATH"`
	Timeout         int    `json:"timeout"           env:"PICOCLAW_SKILLS_REGISTRIES_CLAWHUB_TIMEOUT"`
	MaxZipSize      int    `json:"max_zip_size"      env:"PICOCLAW_SKILLS_REGISTRIES_CLAWHUB_MAX_ZIP_SIZE"`
	MaxResponseSize int    `json:"max_response_size" env:"PICOCLAW_SKILLS_REGISTRIES_CLAWHUB_MAX_RESPONSE_SIZE"`
}

// MCPServerConfig defines configuration for a single MCP server
type MCPServerConfig struct {
	// Enabled indicates whether this MCP server is active
	Enabled bool `json:"enabled"`
	// Command is the executable to run (e.g., "npx", "python", "/path/to/server")
	Command string `json:"command"`
	// Args are the arguments to pass to the command
	Args []string `json:"args,omitempty"`
	// Env are environment variables to set for the server process (stdio only)
	Env map[string]string `json:"env,omitempty"`
	// EnvFile is the path to a file containing environment variables (stdio only)
	EnvFile string `json:"env_file,omitempty"`
	// Type is "stdio", "sse", or "http" (default: stdio if command is set, sse if url is set)
	Type string `json:"type,omitempty"`
	// URL is used for SSE/HTTP transport
	URL string `json:"url,omitempty"`
	// Headers are HTTP headers to send with requests (sse/http only)
	Headers map[string]string `json:"headers,omitempty"`
}

// MCPConfig defines configuration for all MCP servers
type MCPConfig struct {
	// Enabled globally enables/disables MCP integration
	Enabled bool `json:"enabled" env:"PICOCLAW_TOOLS_MCP_ENABLED"`
	// Servers is a map of server name to server configuration
	Servers map[string]MCPServerConfig `json:"servers,omitempty"`
}

// SubagentRoleConfig defines configuration for a subagent role type.
// Roles are predefined agent types with specific capabilities and behaviors.
type SubagentRoleConfig struct {
	// Model is the model identifier for this role (e.g., "gpt-4", "claude-sonnet")
	Model string `json:"model,omitempty" env:"PICOCLAW_SUBAGENT_ROLE_{{.Name}}_MODEL"`
	// Description explains what this role does
	Description string `json:"description,omitempty"`
	// SystemPromptAddon is appended to the base system prompt for this role
	SystemPromptAddon string `json:"system_prompt_addon,omitempty"`
	// MaxIterations limits tool iterations for this role (0 = use default)
	MaxIterations int `json:"max_iterations,omitempty" env:"PICOCLAW_SUBAGENT_ROLE_{{.Name}}_MAX_ITERATIONS"`
	// Temperature controls response randomness (nil = use default)
	Temperature *float64 `json:"temperature,omitempty" env:"PICOCLAW_SUBAGENT_ROLE_{{.Name}}_TEMPERATURE"`
	// MaxTokens limits response length (0 = use default)
	MaxTokens int `json:"max_tokens,omitempty" env:"PICOCLAW_SUBAGENT_ROLE_{{.Name}}_MAX_TOKENS"`
	// TimeoutSeconds is the default timeout for this role (0 = use default)
	TimeoutSeconds int `json:"timeout_seconds,omitempty" env:"PICOCLAW_SUBAGENT_ROLE_{{.Name}}_TIMEOUT"`
	// Extendable allows the task timeout to be extended (default: true)
	Extendable bool `json:"extendable,omitempty" env:"PICOCLAW_SUBAGENT_ROLE_{{.Name}}_EXTENDABLE"`
	// MaxExtensions is the maximum number of timeout extensions allowed (0 = unlimited, default: 3)
	MaxExtensions int `json:"max_extensions,omitempty" env:"PICOCLAW_SUBAGENT_ROLE_{{.Name}}_MAX_EXTENSIONS"`
	// AllowedTools restricts which tools this role can use (empty = all allowed)
	AllowedTools []string `json:"allowed_tools,omitempty"`
}

// RAGConfig defines configuration for Retrieval-Augmented Generation.
type RAGConfig struct {
	// Enabled turns RAG on/off
	Enabled bool `json:"enabled" env:"PICOCLAW_MEMORY_RAG_ENABLED"`
	// EmbeddingModel specifies the embedding model to use:
	// - "none" for keyword-only search (default, zero-config)
	// - "local" for local hash-based embeddings
	// - "openai" for OpenAI embeddings
	// - "http" for HTTP embedding service
	EmbeddingModel string `json:"embedding_model,omitempty" env:"PICOCLAW_MEMORY_RAG_EMBEDDING_MODEL"`
	// ChunkSize is the size of each text chunk in tokens
	ChunkSize int `json:"chunk_size,omitempty" env:"PICOCLAW_MEMORY_RAG_CHUNK_SIZE"`
	// Overlap is the number of overlapping tokens between chunks
	Overlap int `json:"overlap,omitempty" env:"PICOCLAW_MEMORY_RAG_OVERLAP"`
	// MaxResults is the maximum number of chunks to retrieve per query
	MaxResults int `json:"max_results,omitempty" env:"PICOCLAW_MEMORY_RAG_MAX_RESULTS"`
	// SimilarityThreshold is the minimum similarity score (0.0-1.0) for retrieval
	SimilarityThreshold float64 `json:"similarity_threshold,omitempty" env:"PICOCLAW_MEMORY_RAG_SIMILARITY_THRESHOLD"`
	// APIKey for external embedding providers (if not using local)
	APIKey string `json:"api_key,omitempty" env:"PICOCLAW_MEMORY_RAG_API_KEY"`
	// APIBase for external embedding providers
	APIBase string `json:"api_base,omitempty" env:"PICOCLAW_MEMORY_RAG_API_BASE"`
	// ModelPath is the path to local embedding model (for "local" type)
	ModelPath string `json:"model_path,omitempty" env:"PICOCLAW_MEMORY_RAG_MODEL_PATH"`
	// Dimension is the embedding dimension (384 for MiniLM, 1536 for OpenAI)
	Dimension int `json:"dimension,omitempty" env:"PICOCLAW_MEMORY_RAG_DIMENSION"`
	// VectorWeight is the weight for vector similarity in hybrid search (0.0-1.0)
	VectorWeight float64 `json:"vector_weight,omitempty" env:"PICOCLAW_MEMORY_RAG_VECTOR_WEIGHT"`
	// KeywordWeight is the weight for keyword/BM25 score in hybrid search (0.0-1.0)
	KeywordWeight float64 `json:"keyword_weight,omitempty" env:"PICOCLAW_MEMORY_RAG_KEYWORD_WEIGHT"`
}

// MemoryConfig defines configuration for the memory system.
type MemoryConfig struct {
	// Type is the storage backend: "sqlite", "postgres", "memory"
	Type string `json:"type,omitempty" env:"PICOCLAW_MEMORY_TYPE"`
	// Database is the connection string or file path
	Database string `json:"database,omitempty" env:"PICOCLAW_MEMORY_DATABASE"`
	// RAG contains RAG-specific configuration
	RAG RAGConfig `json:"rag,omitempty"`
	// ConceptRetentionDays is how long to keep concept memories
	ConceptRetentionDays int `json:"concept_retention_days,omitempty" env:"PICOCLAW_MEMORY_CONCEPT_RETENTION_DAYS"`
	// MaxContextMemories is the maximum memories to include in context
	MaxContextMemories int `json:"max_context_memories,omitempty" env:"PICOCLAW_MEMORY_MAX_CONTEXT_MEMORIES"`
	// AutoSave enables automatic saving of conversation memories
	AutoSave bool `json:"auto_save,omitempty" env:"PICOCLAW_MEMORY_AUTO_SAVE"`
}

// JobConfig defines configuration for job tracking.
type JobConfig struct {
	// Persistence is the storage backend: "sqlite", "postgres", "memory"
	Persistence string `json:"persistence,omitempty" env:"PICOCLAW_JOBS_PERSISTENCE"`
	// DatabasePath is the file path for SQLite or connection string
	DatabasePath string `json:"database_path,omitempty" env:"PICOCLAW_JOBS_DATABASE_PATH"`
	// DefaultTimeout is the default job timeout in seconds
	DefaultTimeout int `json:"default_timeout,omitempty" env:"PICOCLAW_JOBS_DEFAULT_TIMEOUT"`
	// MaxConcurrent limits concurrent jobs (0 = unlimited)
	MaxConcurrent int `json:"max_concurrent,omitempty" env:"PICOCLAW_JOBS_MAX_CONCURRENT"`
	// CleanupIntervalHours is how often to clean up old jobs
	CleanupIntervalHours int `json:"cleanup_interval_hours,omitempty" env:"PICOCLAW_JOBS_CLEANUP_INTERVAL_HOURS"`
	// RetentionDays is how long to keep completed jobs
	RetentionDays int `json:"retention_days,omitempty" env:"PICOCLAW_JOBS_RETENTION_DAYS"`
}

// WorkspaceConfig defines configuration for workspace management.
// All agents share the same workspace directory specified by Path.
type WorkspaceConfig struct {
	// Path is the shared workspace directory for all agents (default: ~/.picoclaw/workspace)
	Path string `json:"path,omitempty" env:"PICOCLAW_WORKSPACE_PATH"`
	// DEPRECATED: Shared is always true. All agents use the same workspace.
	Shared bool `json:"shared,omitempty" env:"PICOCLAW_WORKSPACE_SHARED"`
	// DEPRECATED: ProjectSubdirs is not used.
	ProjectSubdirs bool `json:"project_subdirs,omitempty" env:"PICOCLAW_WORKSPACE_PROJECT_SUBDIRS"`
	// DEPRECATED: MaxSizeMB is not enforced.
	MaxSizeMB int `json:"max_size_mb,omitempty" env:"PICOCLAW_WORKSPACE_MAX_SIZE_MB"`
	// DEPRECATED: AllowedPaths is not used. Use Path only.
	AllowedPaths []string `json:"allowed_paths,omitempty"`
	// DeniedPaths blocks specific paths (e.g., ~/.ssh, ~/.aws)
	DeniedPaths []string `json:"denied_paths,omitempty"`
}

// A2AConfig defines configuration for Agent-to-Agent (A2A) collaboration system
// with token optimization and multi-agent workspace management.
type A2AConfig struct {
	// Enabled enables the A2A multi-agent collaboration system
	Enabled bool `json:"enabled,omitempty" env:"PICOCLAW_A2A_ENABLED"`
	// TokenOptimization controls all 5 phases of token optimization
	TokenOptimization A2ATokenOptimizationConfig `json:"token_optimization,omitempty"`
	// Compression configures tool loop context compression settings
	Compression A2ACompressionConfig `json:"compression,omitempty"`
	// Messaging configures A2A message management settings
	Messaging A2AMessagingConfig `json:"messaging,omitempty"`
	// Orchestrator configures the A2A orchestrator behavior
	Orchestrator A2AOrchestratorConfig `json:"orchestrator,omitempty"`
}

// A2ATokenOptimizationConfig defines configuration for the 5-phase token optimization system.
type A2ATokenOptimizationConfig struct {
	// Phase1RedundantIdentity removes redundant agent identity from task prompts
	Phase1RedundantIdentity bool `json:"phase1_redundant_identity,omitempty"`
	// Phase2ContextCompression compresses tool loop context after N iterations
	Phase2ContextCompression bool `json:"phase2_context_compression,omitempty"`
	// Phase3MessageSummarization summarizes older A2A messages to reduce context
	Phase3MessageSummarization bool `json:"phase3_message_summarization,omitempty"`
	// Phase4ProjectSummary uses running project summary instead of full message history
	Phase4ProjectSummary bool `json:"phase4_project_summary,omitempty"`
	// Phase5LightweightPrompt uses minimal A2A system prompts without skills summary
	Phase5LightweightPrompt bool `json:"phase5_lightweight_prompt,omitempty"`
}

// A2ACompressionConfig defines configuration for tool loop context compression.
type A2ACompressionConfig struct {
	// FullIterationsKeep is the number of recent iterations to keep uncompressed
	FullIterationsKeep int `json:"full_iterations_keep,omitempty"`
	// CompressThresholdTokens triggers compression when context exceeds this size
	CompressThresholdTokens int `json:"compress_threshold_tokens,omitempty"`
	// MaxContextTokens is the maximum tokens before aggressive compression
	MaxContextTokens int `json:"max_context_tokens,omitempty"`
}

// A2AMessagingConfig defines configuration for A2A message management.
type A2AMessagingConfig struct {
	// RecentMessagesKeep is the number of recent messages to keep in full
	RecentMessagesKeep int `json:"recent_messages_keep,omitempty"`
	// SummarizeThreshold triggers message summarization after N messages
	SummarizeThreshold int `json:"summarize_threshold,omitempty"`
	// ArchiveThreshold archives old messages after N messages
	ArchiveThreshold int `json:"archive_threshold,omitempty"`
}

// A2AOrchestratorConfig defines configuration for the A2A orchestrator.
type A2AOrchestratorConfig struct {
	// MaxConcurrentTasksPerAgent is the maximum number of concurrent tasks per agent (default: 2)
	MaxConcurrentTasksPerAgent int `json:"max_concurrent_tasks_per_agent,omitempty" env:"PICOCLAW_A2A_MAX_CONCURRENT_TASKS"`
	// MaxConcurrentLLMCalls is the maximum number of concurrent LLM API calls globally (default: 3)
	MaxConcurrentLLMCalls int `json:"max_concurrent_llm_calls,omitempty" env:"PICOCLAW_A2A_MAX_LLM_CALLS"`
	// OutsourcePoolSize is the maximum number of outsource agents in the pool (default: 5)
	OutsourcePoolSize int `json:"outsource_pool_size,omitempty" env:"PICOCLAW_A2A_OUTSOURCE_POOL_SIZE"`
	// OutsourceAgentTTLMinutes is the TTL for outsource agents in minutes (default: 30)
	OutsourceAgentTTLMinutes int `json:"outsource_agent_ttl_minutes,omitempty" env:"PICOCLAW_A2A_OUTSOURCE_TTL_MINUTES"`
	// SharedContextMaxLogSize is the maximum number of message log entries (default: 1000)
	SharedContextMaxLogSize int `json:"shared_context_max_log_size,omitempty" env:"PICOCLAW_A2A_SHARED_LOG_SIZE"`
	// SharedContextMaxContext is the maximum number of context entries (default: 10000)
	SharedContextMaxContext int `json:"shared_context_max_context,omitempty" env:"PICOCLAW_A2A_SHARED_CONTEXT_SIZE"`
	// PersistencePath is the path for persisting A2A projects (default: ~/.picoclaw/a2a_projects)
	PersistencePath string `json:"persistence_path,omitempty" env:"PICOCLAW_A2A_PERSISTENCE_PATH"`
	// DefaultLanguage is the default language for agent introductions (default: en)
	DefaultLanguage string `json:"default_language,omitempty" env:"PICOCLAW_A2A_DEFAULT_LANGUAGE"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	// Pre-scan the JSON to check how many model_list entries the user provided.
	// Go's JSON decoder reuses existing slice backing-array elements rather than
	// zero-initializing them, so fields absent from the user's JSON (e.g. api_base)
	// would silently inherit values from the DefaultConfig template at the same
	// index position. We only reset cfg.ModelList when the user actually provides
	// entries; when count is 0 we keep DefaultConfig's built-in list as fallback.
	var tmp Config
	if err := json.Unmarshal(data, &tmp); err != nil {
		return nil, err
	}
	if len(tmp.ModelList) > 0 {
		cfg.ModelList = nil
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	// Migrate legacy channel config fields to new unified structures
	cfg.migrateChannelConfigs()

	// Auto-migrate: if only legacy providers config exists, convert to model_list
	if len(cfg.ModelList) == 0 && cfg.HasProvidersConfig() {
		cfg.ModelList = ConvertProvidersToModelList(cfg)
	}

	// Validate model_list for uniqueness and required fields
	if err := cfg.ValidateModelList(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) migrateChannelConfigs() {
	// Discord: mention_only -> group_trigger.mention_only
	if c.Channels.Discord.MentionOnly && !c.Channels.Discord.GroupTrigger.MentionOnly {
		c.Channels.Discord.GroupTrigger.MentionOnly = true
	}

	// OneBot: group_trigger_prefix -> group_trigger.prefixes
	if len(c.Channels.OneBot.GroupTriggerPrefix) > 0 &&
		len(c.Channels.OneBot.GroupTrigger.Prefixes) == 0 {
		c.Channels.OneBot.GroupTrigger.Prefixes = c.Channels.OneBot.GroupTriggerPrefix
	}
}

func SaveConfig(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Use unified atomic write utility with explicit sync for flash storage reliability.
	return fileutil.WriteFileAtomic(path, data, 0o600)
}

func (c *Config) WorkspacePath() string {
	// Use workspace.path from config as the single source of truth
	// All agents share the same workspace directory
	if c.Workspace.Path != "" {
		return expandHome(c.Workspace.Path)
	}
	// Fallback to legacy agents.defaults.workspace for backward compatibility
	return expandHome(c.Agents.Defaults.Workspace)
}

func (c *Config) GetAPIKey() string {
	if c.Providers.OpenRouter.APIKey != "" {
		return c.Providers.OpenRouter.APIKey
	}
	if c.Providers.Anthropic.APIKey != "" {
		return c.Providers.Anthropic.APIKey
	}
	if c.Providers.OpenAI.APIKey != "" {
		return c.Providers.OpenAI.APIKey
	}
	if c.Providers.Gemini.APIKey != "" {
		return c.Providers.Gemini.APIKey
	}
	if c.Providers.Zhipu.APIKey != "" {
		return c.Providers.Zhipu.APIKey
	}
	if c.Providers.Groq.APIKey != "" {
		return c.Providers.Groq.APIKey
	}
	if c.Providers.VLLM.APIKey != "" {
		return c.Providers.VLLM.APIKey
	}
	if c.Providers.ShengSuanYun.APIKey != "" {
		return c.Providers.ShengSuanYun.APIKey
	}
	if c.Providers.Cerebras.APIKey != "" {
		return c.Providers.Cerebras.APIKey
	}
	return ""
}

func (c *Config) GetAPIBase() string {
	if c.Providers.OpenRouter.APIKey != "" {
		if c.Providers.OpenRouter.APIBase != "" {
			return c.Providers.OpenRouter.APIBase
		}
		return "https://openrouter.ai/api/v1"
	}
	if c.Providers.Zhipu.APIKey != "" {
		return c.Providers.Zhipu.APIBase
	}
	if c.Providers.VLLM.APIKey != "" && c.Providers.VLLM.APIBase != "" {
		return c.Providers.VLLM.APIBase
	}
	return ""
}

func expandHome(path string) string {
	if path == "" {
		return path
	}
	if path[0] == '~' {
		home, _ := os.UserHomeDir()
		if len(path) > 1 && path[1] == '/' {
			return home + path[1:]
		}
		return home
	}
	return path
}

// GetModelConfig returns the ModelConfig for the given model name.
// If multiple configs exist with the same model_name, it uses round-robin
// selection for load balancing. Returns an error if the model is not found.
func (c *Config) GetModelConfig(modelName string) (*ModelConfig, error) {
	matches := c.findMatches(modelName)
	if len(matches) == 0 {
		return nil, fmt.Errorf("model %q not found in model_list or providers", modelName)
	}
	if len(matches) == 1 {
		return &matches[0], nil
	}

	// Multiple configs - use round-robin for load balancing
	idx := rrCounter.Add(1) % uint64(len(matches))
	return &matches[idx], nil
}

// findMatches finds all ModelConfig entries with the given model_name.
func (c *Config) findMatches(modelName string) []ModelConfig {
	var matches []ModelConfig

	// If modelName contains "/", treat it as provider/model format
	// Search by "model" field for legacy format, or by effective model name for new format
	if strings.Contains(modelName, "/") {
		for i := range c.ModelList {
			modelCfg := &c.ModelList[i]

			// Check legacy format: direct match on Model field
			if modelCfg.Model == modelName {
				matches = append(matches, *modelCfg)
				continue
			}

			// Check new format: match on effective model name (provider/model)
			if modelCfg.GetEffectiveModelName() == modelName {
				matches = append(matches, *modelCfg)
			}
		}
		return matches
	}

	// Otherwise, search by model_name (original behavior)
	// This handles both explicit ModelName and cases where modelName is just a model ID
	for i := range c.ModelList {
		modelCfg := &c.ModelList[i]

		// Check explicit ModelName
		if modelCfg.ModelName == modelName {
			matches = append(matches, *modelCfg)
			continue
		}

		// Check if modelName matches the model ID in new format
		// (e.g., searching for "kimi-for-coding" when config has provider="kimi-coding", model="kimi-for-coding")
		if modelCfg.Provider != "" && modelCfg.Model == modelName {
			matches = append(matches, *modelCfg)
			continue
		}

		// Check if modelName matches the effective model name without provider prefix
		// This handles edge cases where someone might reference just the model part
		effectiveModelID := modelCfg.GetEffectiveModelID()
		if effectiveModelID == modelName {
			matches = append(matches, *modelCfg)
		}
	}
	return matches
}

// HasProvidersConfig checks if any provider in the old providers config has configuration.
func (c *Config) HasProvidersConfig() bool {
	return !c.Providers.IsEmpty()
}

// ValidateModelList validates all ModelConfig entries in the model_list.
// It checks that each model config is valid.
// Note: Multiple entries with the same model_name are allowed for load balancing.
func (c *Config) ValidateModelList() error {
	for i := range c.ModelList {
		if err := c.ModelList[i].Validate(); err != nil {
			return fmt.Errorf("model_list[%d]: %w", i, err)
		}
	}
	return nil
}

// ValidateSubagentRoles validates all subagent role configurations.
func (c *Config) ValidateSubagentRoles() error {
	validRoles := []string{"planner", "researcher", "coder", "reviewer", "executor", "specialist", "debugger", "architect", "writer"}
	for roleName, role := range c.SubagentRoles {
		if roleName == "" {
			return fmt.Errorf("subagent role name cannot be empty")
		}
		// Validate temperature range if set
		if role.Temperature != nil && (*role.Temperature < 0 || *role.Temperature > 2) {
			return fmt.Errorf("subagent role %s: temperature must be between 0 and 2", roleName)
		}
		// Validate max iterations
		if role.MaxIterations < 0 {
			return fmt.Errorf("subagent role %s: max_iterations cannot be negative", roleName)
		}
		// Validate timeout
		if role.TimeoutSeconds < 0 {
			return fmt.Errorf("subagent role %s: timeout_seconds cannot be negative", roleName)
		}
		// Warn about unknown roles (but allow them)
		isKnown := false
		for _, known := range validRoles {
			if roleName == known {
				isKnown = true
				break
			}
		}
		if !isKnown {
			// This is a warning, not an error - custom roles are allowed
			continue
		}
	}
	return nil
}

// ValidateMemory validates the memory configuration.
func (c *Config) ValidateMemory() error {
	validTypes := []string{"sqlite", "postgres", "memory", ""}
	isValidType := false
	for _, t := range validTypes {
		if c.Memory.Type == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("memory.type must be one of: sqlite, postgres, memory")
	}

	// Validate RAG config
	if c.Memory.RAG.Enabled {
		validEmbeddingModels := []string{"local", "openai", "custom", "none", ""}
		isValidEmbedding := false
		for _, m := range validEmbeddingModels {
			if c.Memory.RAG.EmbeddingModel == m {
				isValidEmbedding = true
				break
			}
		}
		if !isValidEmbedding {
			return fmt.Errorf("memory.rag.embedding_model must be one of: local, openai, custom, none")
		}
		if c.Memory.RAG.ChunkSize < 0 {
			return fmt.Errorf("memory.rag.chunk_size cannot be negative")
		}
		if c.Memory.RAG.Overlap < 0 {
			return fmt.Errorf("memory.rag.overlap cannot be negative")
		}
		if c.Memory.RAG.SimilarityThreshold < 0 || c.Memory.RAG.SimilarityThreshold > 1 {
			return fmt.Errorf("memory.rag.similarity_threshold must be between 0 and 1")
		}
	}

	if c.Memory.ConceptRetentionDays < 0 {
		return fmt.Errorf("memory.concept_retention_days cannot be negative")
	}
	if c.Memory.MaxContextMemories < 0 {
		return fmt.Errorf("memory.max_context_memories cannot be negative")
	}

	return nil
}

// ValidateJobs validates the job configuration.
func (c *Config) ValidateJobs() error {
	validPersistence := []string{"sqlite", "postgres", "memory", ""}
	isValid := false
	for _, p := range validPersistence {
		if c.Jobs.Persistence == p {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("jobs.persistence must be one of: sqlite, postgres, memory")
	}

	if c.Jobs.DefaultTimeout < 0 {
		return fmt.Errorf("jobs.default_timeout cannot be negative")
	}
	if c.Jobs.MaxConcurrent < 0 {
		return fmt.Errorf("jobs.max_concurrent cannot be negative")
	}
	if c.Jobs.CleanupIntervalHours < 0 {
		return fmt.Errorf("jobs.cleanup_interval_hours cannot be negative")
	}
	if c.Jobs.RetentionDays < 0 {
		return fmt.Errorf("jobs.retention_days cannot be negative")
	}

	return nil
}

// ValidateWorkspace validates the workspace configuration.
// Note: Workspace config is deprecated. Use Agents.Defaults.Workspace instead.
func (c *Config) ValidateWorkspace() error {
	// Validation for deprecated Workspace config removed.
	// Per-agent workspace validation should be done at agent level.
	return nil
}

// Validate performs full configuration validation.
func (c *Config) Validate() error {
	if err := c.ValidateModelList(); err != nil {
		return err
	}
	if err := c.ValidateSubagentRoles(); err != nil {
		return err
	}
	if err := c.ValidateMemory(); err != nil {
		return err
	}
	if err := c.ValidateJobs(); err != nil {
		return err
	}
	if err := c.ValidateWorkspace(); err != nil {
		return err
	}
	return nil
}

// GetSubagentRole returns the configuration for a specific subagent role.
// Returns the role config and a boolean indicating if the role exists.
func (c *Config) GetSubagentRole(roleName string) (SubagentRoleConfig, bool) {
	role, exists := c.SubagentRoles[roleName]
	return role, exists
}

// GetMemoryDatabasePath returns the expanded path to the memory database.
func (c *Config) GetMemoryDatabasePath() string {
	if c.Memory.Database == "" {
		return ""
	}
	return expandHome(c.Memory.Database)
}

// GetJobsDatabasePath returns the expanded path to the jobs database.
func (c *Config) GetJobsDatabasePath() string {
	if c.Jobs.DatabasePath == "" {
		return ""
	}
	return expandHome(c.Jobs.DatabasePath)
}

// GetWorkspacePath returns the expanded workspace path.
// All agents share the same workspace from workspace.path config.
func (c *Config) GetWorkspacePath() string {
	return c.WorkspacePath()
}
