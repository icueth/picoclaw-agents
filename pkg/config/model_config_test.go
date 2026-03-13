// PicoClaw - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package config

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

func TestGetModelConfig_Found(t *testing.T) {
	cfg := &Config{
		ModelList: []ModelConfig{
			{ModelName: "test-model", Model: "openai/gpt-4o", APIKey: "key1"},
			{ModelName: "other-model", Model: "anthropic/claude", APIKey: "key2"},
		},
	}

	result, err := cfg.GetModelConfig("test-model")
	if err != nil {
		t.Fatalf("GetModelConfig() error = %v", err)
	}
	if result.Model != "openai/gpt-4o" {
		t.Errorf("Model = %q, want %q", result.Model, "openai/gpt-4o")
	}
}

func TestGetModelConfig_NotFound(t *testing.T) {
	cfg := &Config{
		ModelList: []ModelConfig{
			{ModelName: "test-model", Model: "openai/gpt-4o", APIKey: "key1"},
		},
	}

	_, err := cfg.GetModelConfig("nonexistent")
	if err == nil {
		t.Fatal("GetModelConfig() expected error for nonexistent model")
	}
}

func TestGetModelConfig_EmptyList(t *testing.T) {
	cfg := &Config{
		ModelList: []ModelConfig{},
	}

	_, err := cfg.GetModelConfig("any-model")
	if err == nil {
		t.Fatal("GetModelConfig() expected error for empty model list")
	}
}

func TestGetModelConfig_RoundRobin(t *testing.T) {
	cfg := &Config{
		ModelList: []ModelConfig{
			{ModelName: "lb-model", Model: "openai/gpt-4o-1", APIKey: "key1"},
			{ModelName: "lb-model", Model: "openai/gpt-4o-2", APIKey: "key2"},
			{ModelName: "lb-model", Model: "openai/gpt-4o-3", APIKey: "key3"},
		},
	}

	// Test round-robin distribution
	results := make(map[string]int)
	for range 30 {
		result, err := cfg.GetModelConfig("lb-model")
		if err != nil {
			t.Fatalf("GetModelConfig() error = %v", err)
		}
		results[result.Model]++
	}

	// Each model should appear roughly 10 times (30 calls / 3 models)
	for model, count := range results {
		if count < 5 || count > 15 {
			t.Errorf("Model %s appeared %d times, expected ~10", model, count)
		}
	}
}

func TestGetModelConfig_Concurrent(t *testing.T) {
	cfg := &Config{
		ModelList: []ModelConfig{
			{ModelName: "concurrent-model", Model: "openai/gpt-4o-1", APIKey: "key1"},
			{ModelName: "concurrent-model", Model: "openai/gpt-4o-2", APIKey: "key2"},
		},
	}

	const goroutines = 100
	const iterations = 10

	var wg sync.WaitGroup
	errors := make(chan error, goroutines*iterations)

	for range goroutines {
		wg.Go(func() {
			for range iterations {
				_, err := cfg.GetModelConfig("concurrent-model")
				if err != nil {
					errors <- err
				}
			}
		})
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent GetModelConfig() error: %v", err)
	}
}

func TestAgentDefaults_GetModelName_BackwardCompat(t *testing.T) {
	tests := []struct {
		name     string
		defaults AgentDefaults
		wantName string
	}{
		{
			name:     "new model_name field only",
			defaults: AgentDefaults{ModelName: "new-model"},
			wantName: "new-model",
		},
		{
			name:     "old model field only",
			defaults: AgentDefaults{Model: "legacy-model"},
			wantName: "legacy-model",
		},
		{
			name:     "both fields - model_name takes precedence",
			defaults: AgentDefaults{ModelName: "new-model", Model: "old-model"},
			wantName: "new-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.defaults.GetModelName(); got != tt.wantName {
				t.Errorf("GetModelName() = %q, want %q", got, tt.wantName)
			}
		})
	}
}

func TestAgentDefaults_JSON_BackwardCompat(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantName string
	}{
		{
			name:     "new model_name field",
			json:     `{"model_name": "gpt4"}`,
			wantName: "gpt4",
		},
		{
			name:     "old model field",
			json:     `{"model": "gpt4"}`,
			wantName: "gpt4",
		},
		{
			name:     "both fields - model_name wins",
			json:     `{"model_name": "new", "model": "old"}`,
			wantName: "new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var defaults AgentDefaults
			if err := json.Unmarshal([]byte(tt.json), &defaults); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}
			if got := defaults.GetModelName(); got != tt.wantName {
				t.Errorf("GetModelName() = %q, want %q", got, tt.wantName)
			}
		})
	}
}

func TestFullConfig_JSON_BackwardCompat(t *testing.T) {
	// Test complete config with both old and new formats
	oldFormat := `{
		"agents": {
			"defaults": {
				"workspace": "~/.picoclaw/workspace",
				"model": "gpt4",
				"max_tokens": 4096
			}
		},
		"model_list": [
			{
				"model_name": "gpt4",
				"model": "openai/gpt-4o",
				"api_key": "test-key"
			}
		]
	}`

	newFormat := `{
		"agents": {
			"defaults": {
				"workspace": "~/.picoclaw/workspace",
				"model_name": "gpt4",
				"max_tokens": 4096
			}
		},
		"model_list": [
			{
				"model_name": "gpt4",
				"model": "openai/gpt-4o",
				"api_key": "test-key"
			}
		]
	}`

	for name, jsonStr := range map[string]string{
		"old format (model)":      oldFormat,
		"new format (model_name)": newFormat,
	} {
		t.Run(name, func(t *testing.T) {
			cfg := &Config{}
			if err := json.Unmarshal([]byte(jsonStr), cfg); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			// Check that GetModelName returns correct value
			if got := cfg.Agents.Defaults.GetModelName(); got != "gpt4" {
				t.Errorf("GetModelName() = %q, want %q", got, "gpt4")
			}

			// Check that GetModelConfig works
			modelCfg, err := cfg.GetModelConfig("gpt4")
			if err != nil {
				t.Fatalf("GetModelConfig error: %v", err)
			}
			if modelCfg.Model != "openai/gpt-4o" {
				t.Errorf("Model = %q, want %q", modelCfg.Model, "openai/gpt-4o")
			}
		})
	}
}

func TestModelConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ModelConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ModelConfig{
				ModelName: "test",
				Model:     "openai/gpt-4o",
			},
			wantErr: false,
		},
		{
			name: "missing model_name",
			config: ModelConfig{
				Model: "openai/gpt-4o",
			},
			wantErr: true,
		},
		{
			name: "missing model",
			config: ModelConfig{
				ModelName: "test",
			},
			wantErr: true,
		},
		{
			name:    "empty config",
			config:  ModelConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ValidateModelList(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string // partial error message to check
	}{
		{
			name: "valid list",
			config: &Config{
				ModelList: []ModelConfig{
					{ModelName: "test1", Model: "openai/gpt-4o"},
					{ModelName: "test2", Model: "anthropic/claude"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid entry",
			config: &Config{
				ModelList: []ModelConfig{
					{ModelName: "test1", Model: "openai/gpt-4o"},
					{ModelName: "", Model: "anthropic/claude"}, // missing model_name
				},
			},
			wantErr: true,
			errMsg:  "model_name is required",
		},
		{
			name: "empty list",
			config: &Config{
				ModelList: []ModelConfig{},
			},
			wantErr: false,
		},
		{
			// Load balancing: multiple entries with same model_name are allowed
			name: "duplicate model_name for load balancing",
			config: &Config{
				ModelList: []ModelConfig{
					{ModelName: "gpt-4", Model: "openai/gpt-4o", APIKey: "key1"},
					{ModelName: "gpt-4", Model: "openai/gpt-4-turbo", APIKey: "key2"},
				},
			},
			wantErr: false, // Changed: duplicates are allowed for load balancing
		},
		{
			// Load balancing: non-adjacent entries with same model_name are also allowed
			name: "duplicate model_name non-adjacent for load balancing",
			config: &Config{
				ModelList: []ModelConfig{
					{ModelName: "model-a", Model: "openai/gpt-4o"},
					{ModelName: "model-b", Model: "anthropic/claude"},
					{ModelName: "model-a", Model: "openai/gpt-4-turbo"},
				},
			},
			wantErr: false, // Changed: duplicates are allowed for load balancing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateModelList()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModelList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateModelList() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestModelConfig_RequestTimeoutParsing(t *testing.T) {
	jsonData := `{
		"model_name": "slow-local",
		"model": "openai/local-model",
		"api_base": "http://localhost:11434/v1",
		"request_timeout": 300
	}`

	var cfg ModelConfig
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if cfg.RequestTimeout != 300 {
		t.Fatalf("RequestTimeout = %d, want 300", cfg.RequestTimeout)
	}
}

func TestModelConfig_RequestTimeoutDefaultZeroValue(t *testing.T) {
	jsonData := `{
		"model_name": "default-timeout",
		"model": "openai/gpt-4o",
		"api_key": "test-key"
	}`

	var cfg ModelConfig
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if cfg.RequestTimeout != 0 {
		t.Fatalf("RequestTimeout = %d, want 0", cfg.RequestTimeout)
	}
}

func TestModelConfig_NewFormatWithProvider(t *testing.T) {
	jsonData := `{
		"provider": "kimi-coding",
		"model": "kimi-for-coding",
		"api_key": "test-key",
		"api_base": "https://api.kimi.com/coding/v1"
	}`

	var cfg ModelConfig
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if cfg.Provider != "kimi-coding" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "kimi-coding")
	}
	if cfg.Model != "kimi-for-coding" {
		t.Errorf("Model = %q, want %q", cfg.Model, "kimi-for-coding")
	}
	if cfg.APIKey != "test-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "test-key")
	}
	if cfg.APIBase != "https://api.kimi.com/coding/v1" {
		t.Errorf("APIBase = %q, want %q", cfg.APIBase, "https://api.kimi.com/coding/v1")
	}

	// Test effective methods
	if cfg.GetEffectiveModelName() != "kimi-coding/kimi-for-coding" {
		t.Errorf("GetEffectiveModelName() = %q, want %q", cfg.GetEffectiveModelName(), "kimi-coding/kimi-for-coding")
	}
	if cfg.GetEffectiveProvider() != "kimi-coding" {
		t.Errorf("GetEffectiveProvider() = %q, want %q", cfg.GetEffectiveProvider(), "kimi-coding")
	}
	if cfg.GetEffectiveModelID() != "kimi-for-coding" {
		t.Errorf("GetEffectiveModelID() = %q, want %q", cfg.GetEffectiveModelID(), "kimi-for-coding")
	}
}

func TestModelConfig_LegacyFormat(t *testing.T) {
	jsonData := `{
		"model_name": "kimi-for-coding",
		"model": "kimi-coding/kimi-for-coding",
		"api_key": "test-key"
	}`

	var cfg ModelConfig
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if cfg.Provider != "" {
		t.Errorf("Provider = %q, want empty", cfg.Provider)
	}
	if cfg.Model != "kimi-coding/kimi-for-coding" {
		t.Errorf("Model = %q, want %q", cfg.Model, "kimi-coding/kimi-for-coding")
	}

	// Test effective methods
	// model_name takes precedence, so GetEffectiveModelName() should return model_name
	if cfg.GetEffectiveModelName() != "kimi-for-coding" {
		t.Errorf("GetEffectiveModelName() = %q, want %q", cfg.GetEffectiveModelName(), "kimi-for-coding")
	}
	if cfg.GetEffectiveProvider() != "kimi-coding" {
		t.Errorf("GetEffectiveProvider() = %q, want %q", cfg.GetEffectiveProvider(), "kimi-coding")
	}
	if cfg.GetEffectiveModelID() != "kimi-for-coding" {
		t.Errorf("GetEffectiveModelID() = %q, want %q", cfg.GetEffectiveModelID(), "kimi-for-coding")
	}
}

func TestModelConfig_ValidateNewFormat(t *testing.T) {
	tests := []struct {
		name    string
		config  ModelConfig
		wantErr bool
	}{
		{
			name: "valid new format with provider",
			config: ModelConfig{
				Provider: "kimi-coding",
				Model:    "kimi-for-coding",
				APIKey:   "test-key",
			},
			wantErr: false,
		},
		{
			name: "valid new format with provider but no model_name",
			config: ModelConfig{
				Provider: "kimi-coding",
				Model:    "kimi-for-coding",
				APIKey:   "test-key",
			},
			wantErr: false, // model_name is optional in new format
		},
		{
			name: "invalid new format missing model",
			config: ModelConfig{
				Provider: "kimi-coding",
				APIKey:   "test-key",
			},
			wantErr: true,
		},
		{
			name: "invalid legacy format missing model_name",
			config: ModelConfig{
				Model: "kimi-coding/kimi-for-coding",
				APIKey: "test-key",
			},
			wantErr: true, // model_name required when provider not specified
		},
		{
			name: "valid legacy format with model_name",
			config: ModelConfig{
				ModelName: "kimi-for-coding",
				Model:     "kimi-coding/kimi-for-coding",
				APIKey:    "test-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDepartmentModelConfig_StringFormat(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected DepartmentModelConfig
	}{
		{
			name: "legacy provider/model format",
			json: `"kimi-coding/kimi-for-coding"`,
			expected: DepartmentModelConfig{
				Provider: "kimi-coding",
				Model:    "kimi-for-coding",
			},
		},
		{
			name: "just model name",
			json: `"kimi-for-coding"`,
			expected: DepartmentModelConfig{
				Provider: "",
				Model:    "kimi-for-coding",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config DepartmentModelConfig
			if err := json.Unmarshal([]byte(tt.json), &config); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if config.Provider != tt.expected.Provider {
				t.Errorf("Provider = %q, want %q", config.Provider, tt.expected.Provider)
			}
			if config.Model != tt.expected.Model {
				t.Errorf("Model = %q, want %q", config.Model, tt.expected.Model)
			}
			if config.GetEffectiveModelName() != tt.expected.GetEffectiveModelName() {
				t.Errorf("GetEffectiveModelName() = %q, want %q", config.GetEffectiveModelName(), tt.expected.GetEffectiveModelName())
			}
		})
	}
}

func TestDepartmentModelConfig_ObjectFormat(t *testing.T) {
	jsonData := `{
		"provider": "kimi-coding",
		"model": "kimi-for-coding"
	}`

	var config DepartmentModelConfig
	if err := json.Unmarshal([]byte(jsonData), &config); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if config.Provider != "kimi-coding" {
		t.Errorf("Provider = %q, want %q", config.Provider, "kimi-coding")
	}
	if config.Model != "kimi-for-coding" {
		t.Errorf("Model = %q, want %q", config.Model, "kimi-for-coding")
	}
	if config.GetEffectiveModelName() != "kimi-coding/kimi-for-coding" {
		t.Errorf("GetEffectiveModelName() = %q, want %q", config.GetEffectiveModelName(), "kimi-coding/kimi-for-coding")
	}
}

func TestDepartmentModelConfig_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		config   DepartmentModelConfig
		expected string
	}{
		{
			name: "with provider",
			config: DepartmentModelConfig{
				Provider: "kimi-coding",
				Model:    "kimi-for-coding",
			},
			expected: `{"provider":"kimi-coding","model":"kimi-for-coding"}`,
		},
		{
			name: "without provider",
			config: DepartmentModelConfig{
				Model: "kimi-for-coding",
			},
			expected: `"kimi-for-coding"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("Marshal() = %q, want %q", string(data), tt.expected)
			}
		})
	}
}

func TestAgentDefaults_NewFormat(t *testing.T) {
	tests := []struct {
		name     string
		defaults AgentDefaults
		wantName string
	}{
		{
			name: "new provider + model format",
			defaults: AgentDefaults{
				Provider: "kimi-coding",
				Model:    "kimi-for-coding",
			},
			wantName: "kimi-coding/kimi-for-coding",
		},
		{
			name: "legacy model field only",
			defaults: AgentDefaults{
				Model: "kimi-coding/kimi-for-coding",
			},
			wantName: "kimi-coding/kimi-for-coding",
		},
		{
			name: "just model name",
			defaults: AgentDefaults{
				Model: "kimi-for-coding",
			},
			wantName: "kimi-for-coding",
		},
		{
			name: "model_name takes precedence",
			defaults: AgentDefaults{
				ModelName: "custom-name",
				Provider:  "kimi-coding",
				Model:     "kimi-for-coding",
			},
			wantName: "custom-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.defaults.GetModelName(); got != tt.wantName {
				t.Errorf("GetModelName() = %q, want %q", got, tt.wantName)
			}
		})
	}
}

func TestAgentDefaults_JSONNewFormat(t *testing.T) {
	jsonData := `{
		"provider": "kimi-coding",
		"model": "kimi-for-coding"
	}`

	var defaults AgentDefaults
	if err := json.Unmarshal([]byte(jsonData), &defaults); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if defaults.Provider != "kimi-coding" {
		t.Errorf("Provider = %q, want %q", defaults.Provider, "kimi-coding")
	}
	if defaults.Model != "kimi-for-coding" {
		t.Errorf("Model = %q, want %q", defaults.Model, "kimi-for-coding")
	}
	if defaults.GetModelName() != "kimi-coding/kimi-for-coding" {
		t.Errorf("GetModelName() = %q, want %q", defaults.GetModelName(), "kimi-coding/kimi-for-coding")
	}
}

func TestConfig_GetModelConfig_NewFormat(t *testing.T) {
	cfg := &Config{
		ModelList: []ModelConfig{
			{
				Provider: "kimi-coding",
				Model:    "kimi-for-coding",
				APIKey:   "test-key",
			},
			{
				ModelName: "legacy-model",
				Model:     "anthropic/claude-sonnet",
				APIKey:    "test-key2",
			},
		},
	}

	// Test finding by effective model name (provider/model)
	result1, err := cfg.GetModelConfig("kimi-coding/kimi-for-coding")
	if err != nil {
		t.Fatalf("GetModelConfig() error = %v", err)
	}
	if result1.Provider != "kimi-coding" || result1.Model != "kimi-for-coding" {
		t.Errorf("Expected kimi-coding/kimi-for-coding, got provider=%q, model=%q", result1.Provider, result1.Model)
	}

	// Test finding by model name (when provider is set)
	result2, err := cfg.GetModelConfig("kimi-for-coding")
	if err != nil {
		t.Fatalf("GetModelConfig() error = %v", err)
	}
	if result2.Provider != "kimi-coding" || result2.Model != "kimi-for-coding" {
		t.Errorf("Expected kimi-coding/kimi-for-coding, got provider=%q, model=%q", result2.Provider, result2.Model)
	}

	// Test finding legacy model by model_name
	result3, err := cfg.GetModelConfig("legacy-model")
	if err != nil {
		t.Fatalf("GetModelConfig() error = %v", err)
	}
	if result3.ModelName != "legacy-model" || result3.Model != "anthropic/claude-sonnet" {
		t.Errorf("Expected legacy-model/anthropic/claude-sonnet, got model_name=%q, model=%q", result3.ModelName, result3.Model)
	}

	// Test finding legacy model by full model string
	result4, err := cfg.GetModelConfig("anthropic/claude-sonnet")
	if err != nil {
		t.Fatalf("GetModelConfig() error = %v", err)
	}
	if result4.Model != "anthropic/claude-sonnet" {
		t.Errorf("Expected anthropic/claude-sonnet, got model=%q", result4.Model)
	}
}

func TestConfig_GetDepartmentModel_NewFormat(t *testing.T) {
	cfg := &Config{
		Agents: AgentsConfig{
			Defaults: AgentDefaults{
				Provider: "default-provider",
				Model:    "default-model",
			},
			DepartmentModels: map[string]DepartmentModelConfig{
				"engineering": {
					Provider: "eng-provider",
					Model:    "eng-model",
				},
				"marketing": {
					Model: "marketing-model", // no provider
				},
			},
		},
	}

	// Test department with provider+model
	if model := cfg.GetDepartmentModel("engineering"); model != "eng-provider/eng-model" {
		t.Errorf("GetDepartmentModel(engineering) = %q, want %q", model, "eng-provider/eng-model")
	}

	// Test department with just model
	if model := cfg.GetDepartmentModel("marketing"); model != "marketing-model" {
		t.Errorf("GetDepartmentModel(marketing) = %q, want %q", model, "marketing-model")
	}

	// Test fallback to defaults
	if model := cfg.GetDepartmentModel("nonexistent"); model != "default-provider/default-model" {
		t.Errorf("GetDepartmentModel(nonexistent) = %q, want %q", model, "default-provider/default-model")
	}
}

func TestModelConfig_WithParameters(t *testing.T) {
	jsonData := `{
		"provider": "kimi-coding",
		"model": "kimi-for-coding",
		"api_key": "test-key",
		"temperature": 0.7,
		"top_p": 0.9,
		"enable_thinking": true
	}`

	var cfg ModelConfig
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if cfg.Provider != "kimi-coding" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "kimi-coding")
	}
	if cfg.Model != "kimi-for-coding" {
		t.Errorf("Model = %q, want %q", cfg.Model, "kimi-for-coding")
	}
	if cfg.APIKey != "test-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "test-key")
	}

	// Test parameters
	if cfg.Temperature == nil || *cfg.Temperature != 0.7 {
		t.Errorf("Temperature = %v, want %f", cfg.Temperature, 0.7)
	}
	if cfg.TopP == nil || *cfg.TopP != 0.9 {
		t.Errorf("TopP = %v, want %f", cfg.TopP, 0.9)
	}
	if cfg.EnableThinking == nil || !*cfg.EnableThinking {
		t.Errorf("EnableThinking = %v, want true", cfg.EnableThinking)
	}
}
