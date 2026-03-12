package agent

import (
	"fmt"
	"strings"

	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/providers"
	"picoclaw/agent/pkg/routing"
)

// DefaultModel is the default model to use when no specific model is configured.
const DefaultModel = "kimi-k2.5"

// ModelResolver resolves models and providers for agents.
// It maps agent configurations to the appropriate LLM provider and model.
type ModelResolver struct {
	cfg             *config.Config
	defaultProvider providers.LLMProvider
	agentModels     map[string]string // agentID -> resolved model name
}

// NewModelResolver creates a new ModelResolver with the given config and default provider.
func NewModelResolver(cfg *config.Config, defaultProvider providers.LLMProvider) *ModelResolver {
	return &ModelResolver{
		cfg:             cfg,
		defaultProvider: defaultProvider,
		agentModels:     make(map[string]string),
	}
}

// GetProviderForAgent returns the appropriate LLM provider for a given agent ID.
// Currently returns the default provider, but can be extended to support
// per-agent provider selection based on model requirements.
func (r *ModelResolver) GetProviderForAgent(agentID string) (providers.LLMProvider, error) {
	if r.defaultProvider == nil {
		return nil, fmt.Errorf("no default provider configured")
	}

	// Normalize the agent ID
	normalizedID := routing.NormalizeAgentID(agentID)

	// TODO: In future phases, this can look up agent-specific provider
	// configurations based on the agent's model requirements
	_ = normalizedID

	return r.defaultProvider, nil
}

// ResolveModel resolves the model configuration for an agent.
// Returns the model name to use, defaulting to "kimi-k2.5" if not specified.
func (r *ModelResolver) ResolveModel(agentConfig *config.AgentConfig) string {
	if agentConfig == nil {
		return DefaultModel
	}

	// Check if we already resolved this agent
	if resolved, ok := r.agentModels[agentConfig.ID]; ok {
		return resolved
	}

	var model string

	// Try to get model from agent config
	if agentConfig.Model != nil && strings.TrimSpace(agentConfig.Model.Primary) != "" {
		model = strings.TrimSpace(agentConfig.Model.Primary)
	} else if r.cfg != nil {
		// Fall back to default model from config
		model = r.cfg.Agents.Defaults.GetModelName()
	}

	// If still empty, use default
	if model == "" {
		model = DefaultModel
	}

	// Cache the resolved model
	r.agentModels[agentConfig.ID] = model

	return model
}

// ResolveModelWithFallbacks resolves the model configuration including fallbacks.
// Returns the primary model and any fallback models.
func (r *ModelResolver) ResolveModelWithFallbacks(agentConfig *config.AgentConfig) (primary string, fallbacks []string) {
	primary = r.ResolveModel(agentConfig)

	if agentConfig == nil || agentConfig.Model == nil {
		return primary, nil
	}

	// Copy fallback list
	if len(agentConfig.Model.Fallbacks) > 0 {
		fallbacks = make([]string, len(agentConfig.Model.Fallbacks))
		copy(fallbacks, agentConfig.Model.Fallbacks)
	}

	return primary, fallbacks
}

// GetModelConfig returns the full ModelConfig for a given model name.
// Looks up the model in the config's model_list.
func (r *ModelResolver) GetModelConfig(modelName string) (*config.ModelConfig, error) {
	if r.cfg == nil {
		return nil, fmt.Errorf("no config available")
	}

	return r.cfg.GetModelConfig(modelName)
}

// RegisterAgentModel manually registers a resolved model for an agent.
// This can be used for dynamic model assignment.
func (r *ModelResolver) RegisterAgentModel(agentID, model string) {
	r.agentModels[agentID] = model
}

// GetResolvedModel returns the cached resolved model for an agent.
func (r *ModelResolver) GetResolvedModel(agentID string) (string, bool) {
	model, ok := r.agentModels[agentID]
	return model, ok
}

// ClearCache clears the resolved model cache.
func (r *ModelResolver) ClearCache() {
	r.agentModels = make(map[string]string)
}
