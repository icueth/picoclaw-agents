package config

import "time"

// ModelCapability defines the capabilities and limitations of a specific LLM model
type ModelCapability struct {
	Name              string
	ContextWindow     int
	IsGoodAtCoding    bool
	IsGoodAtReasoning bool
	IsGoodAtCreative  bool
	IsGoodAtAnalysis  bool
	MaxTaskDuration   time.Duration // Maximum duration this model should handle tasks alone
	RecommendedRoles  []string      // Roles this model is suitable for
	Strengths         []string      // What this model excels at
	Weaknesses        []string      // What this model struggles with
	DelegateTasks     []string      // Task types that should always be delegated
}

// ModelCapabilitiesRegistry contains capability information for known models
var ModelCapabilitiesRegistry = map[string]ModelCapability{
	// 通义千问 (Qwen) - Alibaba
	"qwen3.5-plus": {
		Name:              "Qwen 3.5 Plus",
		ContextWindow:     32768,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   12 * time.Minute,
		RecommendedRoles:  []string{"planner", "coder", "reviewer", "researcher"},
		Strengths:         []string{"coding", "chinese language", "reasoning", "tool use"},
		Weaknesses:        []string{"very large contexts > 32k"},
		DelegateTasks:     []string{},
	},
	"qwen-plus": {
		Name:              "Qwen Plus",
		ContextWindow:     32768,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   10 * time.Minute,
		RecommendedRoles:  []string{"planner", "coder", "reviewer"},
		Strengths:         []string{"coding", "chinese language", "balanced performance"},
		Weaknesses:        []string{"very complex architecture"},
		DelegateTasks:     []string{},
	},

	// Kimi - Moonshot AI
	"kimi-k2.5": {
		Name:              "Kimi K2.5",
		ContextWindow:     256000,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtCreative:  true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   20 * time.Minute,
		RecommendedRoles:  []string{"planner", "coder", "reviewer", "architect", "researcher", "writer"},
		Strengths:         []string{"massive context", "excellent coding", "chinese language", "long document processing"},
		Weaknesses:        []string{},
		DelegateTasks:     []string{},
	},

	// MiniMax
	"minimax-m2.5": {
		Name:              "MiniMax M2.5",
		ContextWindow:     8192,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   10 * time.Minute,
		RecommendedRoles:  []string{"coder", "planner", "reviewer"},
		Strengths:         []string{"coding", "chinese language", "fast responses"},
		Weaknesses:        []string{"large contexts", "very complex reasoning"},
		DelegateTasks:     []string{},
	},

	// 智谱 AI (GLM)
	"glm-4.7": {
		Name:              "GLM-4.7",
		ContextWindow:     128000,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   15 * time.Minute,
		RecommendedRoles:  []string{"planner", "coder", "reviewer", "researcher"},
		Strengths:         []string{"coding", "chinese language", "large context", "tool use"},
		Weaknesses:        []string{},
		DelegateTasks:     []string{},
	},

	// Legacy Moonshot model
	"moonshot-v1-8k": {
		Name:              "Moonshot V1 8K",
		ContextWindow:     8192,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   10 * time.Minute,
		RecommendedRoles:  []string{"coder", "planner", "reviewer"},
		Strengths:         []string{"coding", "chinese language", "efficient"},
		Weaknesses:        []string{"limited context", "very complex tasks"},
		DelegateTasks:     []string{},
	},

	// Groq - Llama
	"llama-3.3-70b": {
		Name:              "Llama 3.3 70B (Groq)",
		ContextWindow:     128000,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   15 * time.Minute,
		RecommendedRoles:  []string{"planner", "coder", "reviewer", "researcher"},
		Strengths:         []string{"very fast inference", "large context", "good coding"},
		Weaknesses:        []string{},
		DelegateTasks:     []string{},
	},

	// OpenAI Models
	"gpt-4": {
		Name:              "GPT-4",
		ContextWindow:     8192,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   10 * time.Minute,
		RecommendedRoles:  []string{"planner", "coder", "reviewer", "architect"},
		Strengths:         []string{"complex reasoning", "structured output", "instruction following"},
		Weaknesses:        []string{"very large context windows"},
		DelegateTasks:     []string{}, // Can handle most tasks
	},
	"gpt-4-turbo": {
		Name:              "GPT-4 Turbo",
		ContextWindow:     128000,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   15 * time.Minute,
		RecommendedRoles:  []string{"planner", "coder", "reviewer", "architect", "researcher"},
		Strengths:         []string{"large context", "complex reasoning", "coding"},
		Weaknesses:        []string{},
		DelegateTasks:     []string{},
	},
	"gpt-3.5-turbo": {
		Name:              "GPT-3.5 Turbo",
		ContextWindow:     4096,
		IsGoodAtCoding:    false,
		IsGoodAtReasoning: false,
		IsGoodAtAnalysis:  false,
		MaxTaskDuration:   5 * time.Minute,
		RecommendedRoles:  []string{"general", "simple_tasks"},
		Strengths:         []string{"fast", "cheap", "simple queries"},
		Weaknesses:        []string{"complex coding", "deep reasoning", "large contexts"},
		DelegateTasks:     []string{"coding", "complex_analysis", "architecture", "debugging"},
	},

	// Anthropic Models
	"claude-opus": {
		Name:              "Claude Opus",
		ContextWindow:     200000,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtCreative:  true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   20 * time.Minute,
		RecommendedRoles:  []string{"planner", "coder", "reviewer", "architect", "researcher", "writer"},
		Strengths:         []string{"very large context", "excellent coding", "nuanced understanding"},
		Weaknesses:        []string{},
		DelegateTasks:     []string{},
	},
	"claude-sonnet": {
		Name:              "Claude Sonnet",
		ContextWindow:     200000,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   15 * time.Minute,
		RecommendedRoles:  []string{"coder", "reviewer", "researcher", "planner"},
		Strengths:         []string{"large context", "good coding", "balanced performance"},
		Weaknesses:        []string{"very complex creative tasks"},
		DelegateTasks:     []string{},
	},
	"claude-haiku": {
		Name:              "Claude Haiku",
		ContextWindow:     200000,
		IsGoodAtCoding:    false,
		IsGoodAtReasoning: false,
		IsGoodAtAnalysis:  false,
		MaxTaskDuration:   3 * time.Minute,
		RecommendedRoles:  []string{"simple_tasks", "quick_responses"},
		Strengths:         []string{"very fast", "cheap", "large context"},
		Weaknesses:        []string{"coding", "complex reasoning", "nuanced tasks"},
		DelegateTasks:     []string{"coding", "debugging", "architecture", "complex_analysis"},
	},

	// Google Models
	"gemini-pro": {
		Name:              "Gemini Pro",
		ContextWindow:     1000000,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   15 * time.Minute,
		RecommendedRoles:  []string{"planner", "coder", "researcher", "architect"},
		Strengths:         []string{"massive context", "multilingual", "good reasoning"},
		Weaknesses:        []string{"some coding edge cases"},
		DelegateTasks:     []string{},
	},

	// DeepSeek Models
	"deepseek-chat": {
		Name:              "DeepSeek Chat",
		ContextWindow:     64000,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   12 * time.Minute,
		RecommendedRoles:  []string{"coder", "planner", "reviewer"},
		Strengths:         []string{"coding", "reasoning", "large context"},
		Weaknesses:        []string{},
	},
	"deepseek-coder": {
		Name:              "DeepSeek Coder",
		ContextWindow:     64000,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   15 * time.Minute,
		RecommendedRoles:  []string{"coder", "reviewer", "architect"},
		Strengths:         []string{"excellent coding", "code completion", "debugging"},
		Weaknesses:        []string{},
		DelegateTasks:     []string{},
	},

	// Local/Other Models (generally weaker)
	"ollama-llama2": {
		Name:              "Llama 2 (Local)",
		ContextWindow:     4096,
		IsGoodAtCoding:    false,
		IsGoodAtReasoning: false,
		IsGoodAtAnalysis:  false,
		MaxTaskDuration:   5 * time.Minute,
		RecommendedRoles:  []string{"simple_tasks"},
		Strengths:         []string{"local", "private", "no API cost"},
		Weaknesses:        []string{"coding quality", "complex reasoning", "instruction following"},
		DelegateTasks:     []string{"coding", "debugging", "architecture", "complex_analysis", "planning"},
	},
	"ollama-codellama": {
		Name:              "CodeLlama (Local)",
		ContextWindow:     4096,
		IsGoodAtCoding:    true,
		IsGoodAtReasoning: false,
		IsGoodAtAnalysis:  false,
		MaxTaskDuration:   8 * time.Minute,
		RecommendedRoles:  []string{"coder"},
		Strengths:         []string{"local", "code completion", "private"},
		Weaknesses:        []string{"complex architecture", "reasoning beyond code"},
		DelegateTasks:     []string{"architecture", "complex_planning", "research"},
	},
}

// GetModelCapability returns capability info for a model name
// Falls back to a conservative default if model is unknown
func GetModelCapability(modelName string) ModelCapability {
	if capability, ok := ModelCapabilitiesRegistry[modelName]; ok {
		return capability
	}

	// Default conservative capability for unknown models
	return ModelCapability{
		Name:              modelName,
		ContextWindow:     4096,
		IsGoodAtCoding:    false,
		IsGoodAtReasoning: false,
		IsGoodAtAnalysis:  false,
		MaxTaskDuration:   5 * time.Minute,
		RecommendedRoles:  []string{"general"},
		Strengths:         []string{"unknown"},
		Weaknesses:        []string{"unknown capabilities - be conservative"},
		DelegateTasks:     []string{"coding", "debugging", "architecture", "complex_analysis", "planning"},
	}
}

// ShouldDelegateTask checks if a task type should be delegated for a given model
func ShouldDelegateTask(modelName string, taskType string) bool {
	capability := GetModelCapability(modelName)
	for _, delegateTask := range capability.DelegateTasks {
		if delegateTask == taskType {
			return true
		}
	}
	return false
}

// IsModelGoodAtRole checks if a model is suitable for a specific role
func IsModelGoodAtRole(modelName string, role string) bool {
	capability := GetModelCapability(modelName)
	for _, recommendedRole := range capability.RecommendedRoles {
		if recommendedRole == role {
			return true
		}
	}
	return false
}

// GetRecommendedRolesForTask returns recommended subagent roles for a task type
func GetRecommendedRolesForTask(taskType string) []string {
	switch taskType {
	case "coding", "programming", "debugging", "implementation":
		return []string{"coder"}
	case "architecture", "design", "planning":
		return []string{"architect", "planner"}
	case "research", "analysis", "investigation":
		return []string{"researcher"}
	case "review", "audit", "quality_check":
		return []string{"reviewer"}
	case "documentation", "writing":
		return []string{"writer"}
	default:
		return []string{"planner", "coder"}
	}
}
