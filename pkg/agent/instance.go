package agent

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/providers"
	"picoclaw/agent/pkg/routing"
	"picoclaw/agent/pkg/session"
	"picoclaw/agent/pkg/tools"
)

// AgentInstance represents a fully configured agent with shared workspace,
// session manager, context builder, and tool registry.
type AgentInstance struct {
	ID                        string
	Name                      string
	Model                     string
	Fallbacks                 []string
	Workspace                 string // Shared workspace directory (all agents use the same)
	MaxIterations             int
	MaxTokens                 int
	Temperature               float64
	ContextWindow             int
	SummarizeMessageThreshold int
	SummarizeTokenPercent     int
	Provider                  providers.LLMProvider
	Sessions                  *session.SessionManager
	ContextBuilder            *ContextBuilder
	Tools                     *tools.ToolRegistry
	Subagents                 *config.SubagentsConfig
	SkillsFilter              []string
	Candidates                []providers.FallbackCandidate
	Config                    *config.AgentConfig // Original config for capability_prompts access
}

// NewAgentInstance creates an agent instance from config.
func NewAgentInstance(
	agentCfg *config.AgentConfig,
	defaults *config.AgentDefaults,
	cfg *config.Config,
	provider providers.LLMProvider,
) *AgentInstance {
	workspace := resolveAgentWorkspace(agentCfg, cfg)
	os.MkdirAll(workspace, 0o755)

	model := resolveAgentModel(agentCfg, defaults, cfg)
	fallbacks := resolveAgentFallbacks(agentCfg, defaults)

	restrict := defaults.RestrictToWorkspace
	readRestrict := restrict && !defaults.AllowReadOutsideWorkspace

	// Compile path whitelist patterns from config.
	allowReadPaths := compilePatterns(cfg.Tools.AllowReadPaths)
	allowWritePaths := compilePatterns(cfg.Tools.AllowWritePaths)

	toolsRegistry := tools.NewToolRegistry()
	toolsRegistry.Register(tools.NewReadFileTool(workspace, readRestrict, allowReadPaths))
	toolsRegistry.Register(tools.NewWriteFileTool(workspace, restrict, allowWritePaths))
	toolsRegistry.Register(tools.NewListDirTool(workspace, readRestrict, allowReadPaths))
	execTool, err := tools.NewExecToolWithConfig(workspace, restrict, cfg)
	if err != nil {
		log.Fatalf("Critical error: unable to initialize exec tool: %v", err)
	}
	toolsRegistry.Register(execTool)

	toolsRegistry.Register(tools.NewEditFileTool(workspace, restrict, allowWritePaths))
	toolsRegistry.Register(tools.NewAppendFileTool(workspace, restrict, allowWritePaths))

	// Determine agent ID and directory
	agentID := routing.DefaultAgentID
	agentName := ""
	var subagents *config.SubagentsConfig
	var skillsFilter []string
	
	if agentCfg != nil {
		agentID = routing.NormalizeAgentID(agentCfg.ID)
		agentName = agentCfg.Name
		subagents = agentCfg.Subagents
		skillsFilter = agentCfg.Skills
	}
	
	// Initialize per-agent workspace (auto-creates directories and MEMORY.md)
	agentWorkspace := NewAgentWorkspace(workspace, agentID)
	
	// Sessions are stored per-agent in workspace/agents/{agent_id}/sessions/
	sessionsManager := session.NewSessionManager(agentWorkspace.SessionDir)

	contextBuilder := NewContextBuilder(workspace)
	// DEPRECATED: agentDir no longer used for persona files
	// Persona files are loaded from workspace/agents/{department}/ or workspace root
	// SetAgentID also initializes per-agent memory stores via agentWorkspace
	contextBuilder.SetAgentID(agentID)
	if agentCfg != nil {
		contextBuilder.SetEmbeddedPrompt(agentCfg.Prompt)
		if agentCfg.Department != "" {
			contextBuilder.SetDepartment(agentCfg.Department)
		}
	}

	maxIter := defaults.MaxToolIterations
	if maxIter == 0 {
		maxIter = 20
	}

	maxTokens := defaults.MaxTokens
	if maxTokens == 0 {
		maxTokens = 8192
	}

	temperature := 0.7
	if defaults.Temperature != nil {
		temperature = *defaults.Temperature
	}

	summarizeMessageThreshold := defaults.SummarizeMessageThreshold
	if summarizeMessageThreshold == 0 {
		summarizeMessageThreshold = 20
	}

	summarizeTokenPercent := defaults.SummarizeTokenPercent
	if summarizeTokenPercent == 0 {
		summarizeTokenPercent = 75
	}

	// Resolve fallback candidates
	modelCfg := providers.ModelConfig{
		Primary:   model,
		Fallbacks: fallbacks,
	}
	resolveFromModelList := func(raw string) (string, bool) {
		ensureProtocol := func(model string) string {
			model = strings.TrimSpace(model)
			if model == "" {
				return ""
			}
			if strings.Contains(model, "/") {
				return model
			}
			return "openai/" + model
		}

		raw = strings.TrimSpace(raw)
		if raw == "" {
			return "", false
		}

		if cfg != nil {
			if mc, err := cfg.GetModelConfig(raw); err == nil && mc != nil && strings.TrimSpace(mc.Model) != "" {
				return ensureProtocol(mc.Model), true
			}

			for i := range cfg.ModelList {
				fullModel := strings.TrimSpace(cfg.ModelList[i].Model)
				if fullModel == "" {
					continue
				}
				if fullModel == raw {
					return ensureProtocol(fullModel), true
				}
				_, modelID := providers.ExtractProtocol(fullModel)
				if modelID == raw {
					return ensureProtocol(fullModel), true
				}
			}
		}

		return "", false
	}

	candidates := providers.ResolveCandidatesWithLookup(modelCfg, defaults.Provider, resolveFromModelList)

	// Use the resolved model name from candidates (the actual API model ID)
	// instead of the alias. This ensures all Chat() calls send the correct
	// model name (e.g., "Qwen/Qwen3.5-35B-A3B-FP8") instead of the
	// user-facing alias (e.g., "qwen3.5-custom-server").
	resolvedModel := model
	if len(candidates) > 0 {
		resolvedModel = candidates[0].Model
	}

	// Set the model name in context builder so the model knows its own identity
	logger.InfoCF("agent", "Setting model name in context builder", map[string]any{
		"model":          model,
		"resolved_model": resolvedModel,
		"agent_id":       agentID,
	})
	contextBuilder.SetModelName(resolvedModel)

	return &AgentInstance{
		ID:                        agentID,
		Name:                      agentName,
		Model:                     resolvedModel,
		Fallbacks:                 fallbacks,
		Workspace:                 workspace,
		MaxIterations:             maxIter,
		MaxTokens:                 maxTokens,
		Temperature:               temperature,
		ContextWindow:             maxTokens,
		SummarizeMessageThreshold: summarizeMessageThreshold,
		SummarizeTokenPercent:     summarizeTokenPercent,
		Provider:                  provider,
		Sessions:                  sessionsManager,
		ContextBuilder:            contextBuilder,
		Tools:                     toolsRegistry,
		Subagents:                 subagents,
		SkillsFilter:              skillsFilter,
		Candidates:                candidates,
		Config:                    agentCfg,
	}
}

// resolveAgentWorkspace determines the workspace directory for an agent.
// FORCED: All agents use the shared workspace from workspace.path config
func resolveAgentWorkspace(agentCfg *config.AgentConfig, cfg *config.Config) string {
	// Respect explicit override if present in agent config
	if agentCfg != nil && strings.TrimSpace(agentCfg.Workspace) != "" {
		return expandHome(strings.TrimSpace(agentCfg.Workspace))
	}
	// Use the centralized workspace.path from config
	return cfg.WorkspacePath()
}

// resolveAgentModel resolves the primary model for an agent.
func resolveAgentModel(agentCfg *config.AgentConfig, defaults *config.AgentDefaults, cfg *config.Config) string {
	if agentCfg != nil {
		if model := agentCfg.GetModel(); model != "" {
			return model
		}
		if agentCfg.Department != "" && cfg != nil {
			if model := cfg.GetDepartmentModel(agentCfg.Department); model != "" {
				return model
			}
		}
	}
	return defaults.GetModelName()
}

// resolveAgentFallbacks resolves the fallback models for an agent.
func resolveAgentFallbacks(agentCfg *config.AgentConfig, defaults *config.AgentDefaults) []string {
	if agentCfg != nil && agentCfg.Model != nil && agentCfg.Model.Fallbacks != nil {
		return agentCfg.Model.Fallbacks
	}
	return defaults.ModelFallbacks
}

func compilePatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			fmt.Printf("Warning: invalid path pattern %q: %v\n", p, err)
			continue
		}
		compiled = append(compiled, re)
	}
	return compiled
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
