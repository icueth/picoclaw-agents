package providers

import (
	"fmt"
	"strings"

	"picoclaw/agent/pkg/auth"
	"picoclaw/agent/pkg/config"
)

const defaultAnthropicAPIBase = "https://api.anthropic.com/v1"

var getCredential = auth.GetCredential

type providerType int

const (
	providerTypeHTTPCompat providerType = iota
	providerTypeClaudeAuth
	providerTypeCodexAuth
	providerTypeCodexCLIToken
	providerTypeClaudeCLI
	providerTypeCodexCLI
	providerTypeGitHubCopilot
	providerTypeMiniMaxPortal
)

type providerSelection struct {
	providerType    providerType
	apiKey          string
	apiBase         string
	proxy           string
	model           string
	workspace       string
	connectMode     string
	enableWebSearch bool
}

func resolveProviderSelection(cfg *config.Config) (providerSelection, error) {
	model := cfg.Agents.Defaults.GetModelName()
	providerName := strings.ToLower(cfg.Agents.Defaults.Provider)
	lowerModel := strings.ToLower(model)

	sel := providerSelection{
		providerType: providerTypeHTTPCompat,
		model:        model,
	}

	// FIRST: Try to get configuration from ModelList (preferred method)
	// This is the new way to configure providers - all settings in one place
	if len(cfg.ModelList) > 0 {
		modelCfg, err := cfg.GetModelConfig(model)
		if err == nil && modelCfg != nil {
			// Get effective provider and model ID from the config
			protocol := modelCfg.GetEffectiveProvider()
			modelID := modelCfg.GetEffectiveModelID()

			sel.apiKey = modelCfg.APIKey
			sel.apiBase = modelCfg.APIBase
			sel.proxy = modelCfg.Proxy
			sel.model = modelID

			// Handle auth methods
			if modelCfg.AuthMethod == "oauth" || modelCfg.AuthMethod == "token" {
				switch protocol {
				case "anthropic":
					sel.providerType = providerTypeClaudeAuth
					if sel.apiBase == "" {
						sel.apiBase = defaultAnthropicAPIBase
					}
					return sel, nil
				case "openai":
					sel.providerType = providerTypeCodexAuth
					if sel.apiBase == "" {
						sel.apiBase = "https://api.openai.com/v1"
					}
					return sel, nil
				case "minimax-portal":
					sel.providerType = providerTypeMiniMaxPortal
					if sel.apiBase == "" {
						sel.apiBase = minimaxPortalBaseURLGlobal
					}
					return sel, nil
				}
			}

			// Set default API base if not specified
			if sel.apiBase == "" {
				sel.apiBase = getDefaultAPIBase(protocol)
			}

			// If we have API key or base, return immediately (ModelList takes precedence)
			if sel.apiKey != "" || sel.apiBase != "" {
				return sel, nil
			}
		}
	}

	// SECOND: Fallback to deprecated ProvidersConfig (legacy method)
	// DEPRECATED: This will be removed in a future version.
	// Please migrate to using model_list instead.
	if providerName != "" {
		switch providerName {
		case "groq":
			if cfg.Providers.Groq.APIKey != "" {
				sel.apiKey = cfg.Providers.Groq.APIKey
				sel.apiBase = cfg.Providers.Groq.APIBase
				sel.proxy = cfg.Providers.Groq.Proxy
				if sel.apiBase == "" {
					sel.apiBase = "https://api.groq.com/openai/v1"
				}
			}
		case "openai", "gpt":
			if cfg.Providers.OpenAI.APIKey != "" || cfg.Providers.OpenAI.AuthMethod != "" {
				sel.enableWebSearch = cfg.Providers.OpenAI.WebSearch
				if cfg.Providers.OpenAI.AuthMethod == "codex-cli" {
					sel.providerType = providerTypeCodexCLIToken
					return sel, nil
				}
				if cfg.Providers.OpenAI.AuthMethod == "oauth" || cfg.Providers.OpenAI.AuthMethod == "token" {
					sel.providerType = providerTypeCodexAuth
					return sel, nil
				}
				sel.apiKey = cfg.Providers.OpenAI.APIKey
				sel.apiBase = cfg.Providers.OpenAI.APIBase
				sel.proxy = cfg.Providers.OpenAI.Proxy
				if sel.apiBase == "" {
					sel.apiBase = "https://api.openai.com/v1"
				}
			}
		case "anthropic", "claude":
			if cfg.Providers.Anthropic.APIKey != "" || cfg.Providers.Anthropic.AuthMethod != "" {
				if cfg.Providers.Anthropic.AuthMethod == "oauth" || cfg.Providers.Anthropic.AuthMethod == "token" {
					sel.apiBase = cfg.Providers.Anthropic.APIBase
					if sel.apiBase == "" {
						sel.apiBase = defaultAnthropicAPIBase
					}
					sel.providerType = providerTypeClaudeAuth
					return sel, nil
				}
				sel.apiKey = cfg.Providers.Anthropic.APIKey
				sel.apiBase = cfg.Providers.Anthropic.APIBase
				sel.proxy = cfg.Providers.Anthropic.Proxy
				if sel.apiBase == "" {
					sel.apiBase = defaultAnthropicAPIBase
				}
			}
		case "openrouter":
			if cfg.Providers.OpenRouter.APIKey != "" {
				sel.apiKey = cfg.Providers.OpenRouter.APIKey
				sel.proxy = cfg.Providers.OpenRouter.Proxy
				if cfg.Providers.OpenRouter.APIBase != "" {
					sel.apiBase = cfg.Providers.OpenRouter.APIBase
				} else {
					sel.apiBase = "https://openrouter.ai/api/v1"
				}
			}
		case "litellm":
			if cfg.Providers.LiteLLM.APIKey != "" || cfg.Providers.LiteLLM.APIBase != "" {
				sel.apiKey = cfg.Providers.LiteLLM.APIKey
				sel.apiBase = cfg.Providers.LiteLLM.APIBase
				sel.proxy = cfg.Providers.LiteLLM.Proxy
				if sel.apiBase == "" {
					sel.apiBase = "http://localhost:4000/v1"
				}
			}
		case "zhipu", "glm":
			if cfg.Providers.Zhipu.APIKey != "" {
				sel.apiKey = cfg.Providers.Zhipu.APIKey
				sel.apiBase = cfg.Providers.Zhipu.APIBase
				sel.proxy = cfg.Providers.Zhipu.Proxy
				if sel.apiBase == "" {
					sel.apiBase = "https://open.bigmodel.cn/api/paas/v4"
				}
			}
		case "gemini", "google":
			if cfg.Providers.Gemini.APIKey != "" {
				sel.apiKey = cfg.Providers.Gemini.APIKey
				sel.apiBase = cfg.Providers.Gemini.APIBase
				sel.proxy = cfg.Providers.Gemini.Proxy
				if sel.apiBase == "" {
					sel.apiBase = "https://generativelanguage.googleapis.com/v1beta"
				}
			}
		case "vllm":
			if cfg.Providers.VLLM.APIBase != "" {
				sel.apiKey = cfg.Providers.VLLM.APIKey
				sel.apiBase = cfg.Providers.VLLM.APIBase
				sel.proxy = cfg.Providers.VLLM.Proxy
			}
		case "shengsuanyun":
			if cfg.Providers.ShengSuanYun.APIKey != "" {
				sel.apiKey = cfg.Providers.ShengSuanYun.APIKey
				sel.apiBase = cfg.Providers.ShengSuanYun.APIBase
				sel.proxy = cfg.Providers.ShengSuanYun.Proxy
				if sel.apiBase == "" {
					sel.apiBase = "https://router.shengsuanyun.com/api/v1"
				}
			}
		case "nvidia":
			if cfg.Providers.Nvidia.APIKey != "" {
				sel.apiKey = cfg.Providers.Nvidia.APIKey
				sel.apiBase = cfg.Providers.Nvidia.APIBase
				sel.proxy = cfg.Providers.Nvidia.Proxy
				if sel.apiBase == "" {
					sel.apiBase = "https://integrate.api.nvidia.com/v1"
				}
			}
		case "claude-cli", "claude-code", "claudecode":
			workspace := cfg.WorkspacePath()
			if workspace == "" {
				workspace = "."
			}
			sel.providerType = providerTypeClaudeCLI
			sel.workspace = workspace
			return sel, nil
		case "codex-cli", "codex-code":
			workspace := cfg.WorkspacePath()
			if workspace == "" {
				workspace = "."
			}
			sel.providerType = providerTypeCodexCLI
			sel.workspace = workspace
			return sel, nil
		case "deepseek":
			if cfg.Providers.DeepSeek.APIKey != "" {
				sel.apiKey = cfg.Providers.DeepSeek.APIKey
				sel.apiBase = cfg.Providers.DeepSeek.APIBase
				sel.proxy = cfg.Providers.DeepSeek.Proxy
				if sel.apiBase == "" {
					sel.apiBase = "https://api.deepseek.com/v1"
				}
				if model != "deepseek-chat" && model != "deepseek-reasoner" {
					sel.model = "deepseek-chat"
				}
			}
		case "mistral":
			if cfg.Providers.Mistral.APIKey != "" {
				sel.apiKey = cfg.Providers.Mistral.APIKey
				sel.apiBase = cfg.Providers.Mistral.APIBase
				sel.proxy = cfg.Providers.Mistral.Proxy
				if sel.apiBase == "" {
					sel.apiBase = "https://api.mistral.ai/v1"
				}
			}
		case "qwen":
			if cfg.Providers.Qwen.APIKey != "" || cfg.Providers.Qwen.APIBase != "" {
				sel.apiKey = cfg.Providers.Qwen.APIKey
				sel.apiBase = cfg.Providers.Qwen.APIBase
				sel.proxy = cfg.Providers.Qwen.Proxy
				if sel.apiBase == "" {
					sel.apiBase = "https://dashscope.aliyuncs.com/compatible-mode/v1"
				}
			}
		case "bailian", "dashscope":
			if cfg.Providers.Bailian.APIKey != "" {
				sel.apiKey = cfg.Providers.Bailian.APIKey
				sel.apiBase = cfg.Providers.Bailian.APIBase
				sel.proxy = cfg.Providers.Bailian.Proxy
				if sel.apiBase == "" {
					sel.apiBase = "https://coding-intl.dashscope.aliyuncs.com/v1"
				}
			}
		case "minimax-portal", "minimax":
			if cfg.Providers.MiniMaxPortal.AuthMethod == "oauth" {
				sel.apiBase = cfg.Providers.MiniMaxPortal.APIBase
				if sel.apiBase == "" {
					sel.apiBase = minimaxPortalBaseURLGlobal
				}
				sel.providerType = providerTypeMiniMaxPortal
				return sel, nil
			}
			if cfg.Providers.MiniMaxPortal.APIKey != "" {
				sel.apiKey = cfg.Providers.MiniMaxPortal.APIKey
				sel.apiBase = cfg.Providers.MiniMaxPortal.APIBase
				sel.proxy = cfg.Providers.MiniMaxPortal.Proxy
				if sel.apiBase == "" {
					sel.apiBase = minimaxPortalBaseURLGlobal
				}
			}
		case "kimi-coding", "kimi":
			if cfg.Providers.KimiCoding.APIKey != "" {
				sel.apiKey = cfg.Providers.KimiCoding.APIKey
				sel.apiBase = cfg.Providers.KimiCoding.APIBase
				sel.proxy = cfg.Providers.KimiCoding.Proxy
				if sel.apiBase == "" {
					sel.apiBase = "https://api.kimi.com/coding/v1"
				}
			}
		case "github_copilot", "copilot":
			sel.providerType = providerTypeGitHubCopilot
			if cfg.Providers.GitHubCopilot.APIBase != "" {
				sel.apiBase = cfg.Providers.GitHubCopilot.APIBase
			} else {
				sel.apiBase = "localhost:4321"
			}
			sel.connectMode = cfg.Providers.GitHubCopilot.ConnectMode
			return sel, nil
		}
	}

	// Fallback: infer provider from model and configured keys.
	if sel.apiKey == "" && sel.apiBase == "" {
		switch {
		case strings.HasPrefix(model, "kimi-coding/") && cfg.Providers.KimiCoding.APIKey != "":
			sel.apiKey = cfg.Providers.KimiCoding.APIKey
			sel.apiBase = cfg.Providers.KimiCoding.APIBase
			sel.proxy = cfg.Providers.KimiCoding.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://api.kimi.com/coding/v1"
			}
		case (strings.Contains(lowerModel, "kimi") || strings.Contains(lowerModel, "moonshot") || strings.HasPrefix(model, "moonshot/")) && cfg.Providers.Moonshot.APIKey != "":
			sel.apiKey = cfg.Providers.Moonshot.APIKey
			sel.apiBase = cfg.Providers.Moonshot.APIBase
			sel.proxy = cfg.Providers.Moonshot.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://api.moonshot.ai/v1"
			}
		case (strings.HasPrefix(model, "openrouter/") ||
			strings.HasPrefix(model, "anthropic/") ||
			strings.HasPrefix(model, "openai/") ||
			strings.HasPrefix(model, "meta-llama/") ||
			strings.HasPrefix(model, "deepseek/") ||
			strings.HasPrefix(model, "google/")) && cfg.Providers.OpenRouter.APIKey != "":
			sel.apiKey = cfg.Providers.OpenRouter.APIKey
			sel.proxy = cfg.Providers.OpenRouter.Proxy
			if cfg.Providers.OpenRouter.APIBase != "" {
				sel.apiBase = cfg.Providers.OpenRouter.APIBase
			} else {
				sel.apiBase = "https://openrouter.ai/api/v1"
			}
		case (strings.Contains(lowerModel, "claude") || strings.HasPrefix(model, "anthropic/")) &&
			(cfg.Providers.Anthropic.APIKey != "" || cfg.Providers.Anthropic.AuthMethod != ""):
			if cfg.Providers.Anthropic.AuthMethod == "oauth" || cfg.Providers.Anthropic.AuthMethod == "token" {
				sel.apiBase = cfg.Providers.Anthropic.APIBase
				if sel.apiBase == "" {
					sel.apiBase = defaultAnthropicAPIBase
				}
				sel.providerType = providerTypeClaudeAuth
				return sel, nil
			}
			sel.apiKey = cfg.Providers.Anthropic.APIKey
			sel.apiBase = cfg.Providers.Anthropic.APIBase
			sel.proxy = cfg.Providers.Anthropic.Proxy
			if sel.apiBase == "" {
				sel.apiBase = defaultAnthropicAPIBase
			}
		case (strings.Contains(lowerModel, "gpt") || strings.HasPrefix(model, "openai/")) &&
			(cfg.Providers.OpenAI.APIKey != "" || cfg.Providers.OpenAI.AuthMethod != ""):
			sel.enableWebSearch = cfg.Providers.OpenAI.WebSearch
			if cfg.Providers.OpenAI.AuthMethod == "codex-cli" {
				sel.providerType = providerTypeCodexCLIToken
				return sel, nil
			}
			if cfg.Providers.OpenAI.AuthMethod == "oauth" || cfg.Providers.OpenAI.AuthMethod == "token" {
				sel.providerType = providerTypeCodexAuth
				return sel, nil
			}
			sel.apiKey = cfg.Providers.OpenAI.APIKey
			sel.apiBase = cfg.Providers.OpenAI.APIBase
			sel.proxy = cfg.Providers.OpenAI.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://api.openai.com/v1"
			}
		case (strings.Contains(lowerModel, "gemini") || strings.HasPrefix(model, "google/")) && cfg.Providers.Gemini.APIKey != "":
			sel.apiKey = cfg.Providers.Gemini.APIKey
			sel.apiBase = cfg.Providers.Gemini.APIBase
			sel.proxy = cfg.Providers.Gemini.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://generativelanguage.googleapis.com/v1beta"
			}
		case (strings.Contains(lowerModel, "qwen") || strings.HasPrefix(model, "qwen/")) && cfg.Providers.Qwen.APIKey != "":
			sel.apiKey = cfg.Providers.Qwen.APIKey
			sel.apiBase = cfg.Providers.Qwen.APIBase
			sel.proxy = cfg.Providers.Qwen.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://dashscope.aliyuncs.com/compatible-mode/v1"
			}
		case (strings.Contains(lowerModel, "glm") || strings.Contains(lowerModel, "zhipu") || strings.Contains(lowerModel, "zai")) && cfg.Providers.Zhipu.APIKey != "":
			sel.apiKey = cfg.Providers.Zhipu.APIKey
			sel.apiBase = cfg.Providers.Zhipu.APIBase
			sel.proxy = cfg.Providers.Zhipu.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://open.bigmodel.cn/api/paas/v4"
			}
		case (strings.Contains(lowerModel, "groq") || strings.HasPrefix(model, "groq/")) && cfg.Providers.Groq.APIKey != "":
			sel.apiKey = cfg.Providers.Groq.APIKey
			sel.apiBase = cfg.Providers.Groq.APIBase
			sel.proxy = cfg.Providers.Groq.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://api.groq.com/openai/v1"
			}
		case (strings.Contains(lowerModel, "nvidia") || strings.HasPrefix(model, "nvidia/")) && cfg.Providers.Nvidia.APIKey != "":
			sel.apiKey = cfg.Providers.Nvidia.APIKey
			sel.apiBase = cfg.Providers.Nvidia.APIBase
			sel.proxy = cfg.Providers.Nvidia.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://integrate.api.nvidia.com/v1"
			}
		case (strings.Contains(lowerModel, "ollama") || strings.HasPrefix(model, "ollama/")) && cfg.Providers.Ollama.APIKey != "":
			sel.apiKey = cfg.Providers.Ollama.APIKey
			sel.apiBase = cfg.Providers.Ollama.APIBase
			sel.proxy = cfg.Providers.Ollama.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "http://localhost:11434/v1"
			}
		case (strings.Contains(lowerModel, "mistral") || strings.HasPrefix(model, "mistral/")) && cfg.Providers.Mistral.APIKey != "":
			sel.apiKey = cfg.Providers.Mistral.APIKey
			sel.apiBase = cfg.Providers.Mistral.APIBase
			sel.proxy = cfg.Providers.Mistral.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://api.mistral.ai/v1"
			}
		case (strings.Contains(lowerModel, "minimax") || strings.HasPrefix(model, "minimax-portal/")) &&
			(cfg.Providers.MiniMaxPortal.APIKey != "" || cfg.Providers.MiniMaxPortal.AuthMethod != ""):
			if cfg.Providers.MiniMaxPortal.AuthMethod == "oauth" {
				sel.apiBase = cfg.Providers.MiniMaxPortal.APIBase
				if sel.apiBase == "" {
					sel.apiBase = minimaxPortalBaseURLGlobal
				}
				sel.providerType = providerTypeMiniMaxPortal
				return sel, nil
			}
			sel.apiKey = cfg.Providers.MiniMaxPortal.APIKey
			sel.apiBase = cfg.Providers.MiniMaxPortal.APIBase
			sel.proxy = cfg.Providers.MiniMaxPortal.Proxy
			if sel.apiBase == "" {
				sel.apiBase = minimaxPortalBaseURLGlobal
			}
		case cfg.Providers.VLLM.APIBase != "":
			sel.apiKey = cfg.Providers.VLLM.APIKey
			sel.apiBase = cfg.Providers.VLLM.APIBase
			sel.proxy = cfg.Providers.VLLM.Proxy
		default:
			if cfg.Providers.OpenRouter.APIKey != "" {
				sel.apiKey = cfg.Providers.OpenRouter.APIKey
				sel.proxy = cfg.Providers.OpenRouter.Proxy
				if cfg.Providers.OpenRouter.APIBase != "" {
					sel.apiBase = cfg.Providers.OpenRouter.APIBase
				} else {
					sel.apiBase = "https://openrouter.ai/api/v1"
				}
			} else {
				return providerSelection{}, fmt.Errorf("no API key configured for model: %s", model)
			}
		}
	}

	if sel.providerType == providerTypeHTTPCompat {
		if sel.apiKey == "" && !strings.HasPrefix(model, "bedrock/") {
			return providerSelection{}, fmt.Errorf("no API key configured for provider (model: %s)", model)
		}
		if sel.apiBase == "" {
			return providerSelection{}, fmt.Errorf("no API base configured for provider (model: %s)", model)
		}
	}

	return sel, nil
}
