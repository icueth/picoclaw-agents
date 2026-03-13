// PicoClaw - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package providers

import (
	"fmt"
	"strings"

	"picoclaw/agent/pkg/config"
)

// createClaudeAuthProvider creates a Claude provider using OAuth credentials from auth store.
func createClaudeAuthProvider() (LLMProvider, error) {
	cred, err := getCredential("anthropic")
	if err != nil {
		return nil, fmt.Errorf("loading auth credentials: %w", err)
	}
	if cred == nil {
		return nil, fmt.Errorf("no credentials for anthropic. Run: picoclaw auth login --provider anthropic")
	}
	return NewClaudeProviderWithTokenSource(cred.AccessToken, createClaudeTokenSource()), nil
}

// createCodexAuthProvider creates a Codex provider using OAuth credentials from auth store.
func createCodexAuthProvider() (LLMProvider, error) {
	cred, err := getCredential("openai")
	if err != nil {
		return nil, fmt.Errorf("loading auth credentials: %w", err)
	}
	if cred == nil {
		return nil, fmt.Errorf("no credentials for openai. Run: picoclaw auth login --provider openai")
	}
	return NewCodexProviderWithTokenSource(cred.AccessToken, cred.AccountID, createCodexTokenSource()), nil
}

// ExtractProtocol extracts the protocol prefix and model identifier from a model string.
// If no prefix is specified, it defaults to "openai".
// Examples:
//   - "openai/gpt-4o" -> ("openai", "gpt-4o")
//   - "anthropic/claude-sonnet-4.6" -> ("anthropic", "claude-sonnet-4.6")
//   - "gpt-4o" -> ("openai", "gpt-4o")
func ExtractProtocol(model string) (protocol, modelID string) {
	model = strings.TrimSpace(model)

	protocol, modelID, found := strings.Cut(model, "/")
	if !found {
		return "openai", model
	}
	return protocol, modelID
}

// createHTTPProvider creates an HTTP provider from ModelConfig, applying concurrency limits if configured.
func createHTTPProvider(cfg *config.ModelConfig, apiBase string) *HTTPProvider {
	if cfg.MaxConcurrent > 0 {
		return NewHTTPProviderWithOptions(
			cfg.APIKey, apiBase, cfg.Proxy, cfg.MaxTokensField,
			cfg.RequestTimeout, cfg.MaxConcurrent,
		)
	}
	return NewHTTPProviderWithMaxTokensFieldAndRequestTimeout(
		cfg.APIKey, apiBase, cfg.Proxy, cfg.MaxTokensField, cfg.RequestTimeout,
	)
}

// CreateProviderFromConfig creates a provider based on the ModelConfig.
// It uses the explicit Provider field when available, otherwise falls back to extracting
// the protocol prefix from the Model field (legacy format).
// Supported protocols: openai, litellm, anthropic, minimax-portal, antigravity, claude-cli, codex-cli, github-copilot
// Returns the provider, the model ID (without protocol prefix), and any error.
func CreateProviderFromConfig(cfg *config.ModelConfig) (LLMProvider, string, error) {
	if cfg == nil {
		return nil, "", fmt.Errorf("config is nil")
	}

	if cfg.Model == "" {
		return nil, "", fmt.Errorf("model is required")
	}

	// Get effective provider and model ID from the config
	protocol := cfg.GetEffectiveProvider()
	modelID := cfg.GetEffectiveModelID()

	switch protocol {
	case "openai":
		// OpenAI with OAuth/token auth (Codex-style)
		if cfg.AuthMethod == "oauth" || cfg.AuthMethod == "token" {
			provider, err := createCodexAuthProvider()
			if err != nil {
				return nil, "", err
			}
			return provider, modelID, nil
		}
		// OpenAI with API key
		if cfg.APIKey == "" && cfg.APIBase == "" {
			return nil, "", fmt.Errorf("api_key or api_base is required for HTTP-based protocol %q", protocol)
		}
		apiBase := cfg.APIBase
		if apiBase == "" {
			apiBase = getDefaultAPIBase(protocol)
		}
		return createHTTPProvider(cfg, apiBase), modelID, nil

	case "litellm", "openrouter", "groq", "zhipu", "gemini", "nvidia",
		"ollama", "moonshot", "moonshotai", "shengsuanyun", "deepseek", "cerebras",
		"volcengine", "vllm", "qwen", "mistral", "bailian", "dashscope":
		// All other OpenAI-compatible HTTP providers
		if cfg.APIKey == "" && cfg.APIBase == "" {
			return nil, "", fmt.Errorf("no API key or base configured for model: %s", cfg.GetEffectiveModelName())
		}

		apiBase := cfg.APIBase
		if apiBase == "" {
			apiBase = getDefaultAPIBase(protocol)
		}
		return createHTTPProvider(cfg, apiBase), modelID, nil

	case "anthropic":
		if cfg.AuthMethod == "oauth" || cfg.AuthMethod == "token" {
			// Use OAuth credentials from auth store
			provider, err := createClaudeAuthProvider()
			if err != nil {
				return nil, "", err
			}
			return provider, modelID, nil
		}
		// Use API key with HTTP API
		apiBase := cfg.APIBase
		if apiBase == "" {
			apiBase = "https://api.anthropic.com/v1"
		}
		if cfg.APIKey == "" {
			return nil, "", fmt.Errorf("api_key is required for anthropic protocol (model: %s)", cfg.GetEffectiveModelName())
		}
		return createHTTPProvider(cfg, apiBase), modelID, nil

	case "minimax-portal":
		// MiniMax Portal: OAuth-authenticated access to MiniMax models via Anthropic API
		if cfg.AuthMethod == "oauth" {
			provider, err := NewMiniMaxPortalProvider()
			if err != nil {
				return nil, "", err
			}
			return provider, modelID, nil
		}
		// Fallback: API key with HTTP API (Anthropic compat)
		apiBase := cfg.APIBase
		if apiBase == "" {
			apiBase = minimaxPortalBaseURLGlobal
		}
		if cfg.APIKey == "" {
			return nil, "", fmt.Errorf("api_key is required for minimax-portal without oauth (model: %s)", cfg.GetEffectiveModelName())
		}
		return createHTTPProvider(cfg, apiBase), modelID, nil

	case "kimi-coding":
		// Kimi Coding API: Requires User-Agent: KimiCLI/1.11.0 header
		apiBase := cfg.APIBase
		if apiBase == "" {
			apiBase = "https://api.kimi.com/coding/v1"
		}
		if cfg.APIKey == "" {
			return nil, "", fmt.Errorf("api_key is required for kimi-coding (model: %s)", cfg.GetEffectiveModelName())
		}
		// Create provider with custom User-Agent header
		headers := map[string]string{
			"User-Agent": "KimiCLI/1.11.0",
		}
		// For Kimi, we need to pass the full model name including provider prefix
		// to maintain compatibility with their API expectations
		fullModelName := cfg.GetEffectiveModelName()
		return NewHTTPProviderWithHeaders(cfg.APIKey, apiBase, cfg.Proxy, headers), fullModelName, nil

	case "antigravity":
		return NewAntigravityProvider(), modelID, nil

	case "claude-cli", "claudecli":
		workspace := cfg.Workspace
		if workspace == "" {
			workspace = "."
		}
		return NewClaudeCliProvider(workspace), modelID, nil

	case "codex-cli", "codexcli":
		workspace := cfg.Workspace
		if workspace == "" {
			workspace = "."
		}
		return NewCodexCliProvider(workspace), modelID, nil

	case "github-copilot", "copilot":
		apiBase := cfg.APIBase
		if apiBase == "" {
			apiBase = "localhost:4321"
		}
		connectMode := cfg.ConnectMode
		if connectMode == "" {
			connectMode = "grpc"
		}
		provider, err := NewGitHubCopilotProvider(apiBase, connectMode, modelID)
		if err != nil {
			return nil, "", err
		}
		return provider, modelID, nil

	default:
		// Unknown protocol with api_base = custom OpenAI-compatible server.
		// Pass the FULL original model name (e.g. "Qwen/Qwen3.5-35B-A3B-FP8")
		// so the server receives the exact model identifier it expects.
		if cfg.APIBase != "" {
			fullModelName := cfg.GetEffectiveModelName()
			return createHTTPProvider(cfg, cfg.APIBase), fullModelName, nil
		}
		return nil, "", fmt.Errorf("unknown protocol %q in model %q (set api_base for custom servers)", protocol, cfg.GetEffectiveModelName())
	}
}

// getDefaultAPIBase returns the default API base URL for a given protocol.
func getDefaultAPIBase(protocol string) string {
	switch protocol {
	case "openai":
		return "https://api.openai.com/v1"
	case "openrouter":
		return "https://openrouter.ai/api/v1"
	case "litellm":
		return "http://localhost:4000/v1"
	case "groq":
		return "https://api.groq.com/openai/v1"
	case "zhipu":
		return "https://open.bigmodel.cn/api/paas/v4"
	case "gemini":
		return "https://generativelanguage.googleapis.com/v1beta"
	case "nvidia":
		return "https://integrate.api.nvidia.com/v1"
	case "ollama":
		return "http://localhost:11434/v1"
	case "moonshot", "moonshotai":
		return "https://api.moonshot.ai/v1"
	case "shengsuanyun":
		return "https://router.shengsuanyun.com/api/v1"
	case "deepseek":
		return "https://api.deepseek.com/v1"
	case "cerebras":
		return "https://api.cerebras.ai/v1"
	case "volcengine":
		return "https://ark.cn-beijing.volces.com/api/v3"
	case "qwen":
		return "https://dashscope.aliyuncs.com/compatible-mode/v1"
	case "vllm":
		return "http://localhost:8000/v1"
	case "mistral":
		return "https://api.mistral.ai/v1"
	case "bailian", "dashscope":
		return "https://coding-intl.dashscope.aliyuncs.com/v1"
	default:
		return ""
	}
}
