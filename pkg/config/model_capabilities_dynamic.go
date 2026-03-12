package config

import (
	"strings"
	"time"
)

// DynamicCapabilityManager โหลด capabilities จาก config ที่เปิดใช้งานจริง
type DynamicCapabilityManager struct {
	cfg               *Config
	capabilityMapping map[string]string // model_name → base_model
}

// NewDynamicCapabilityManager สร้าง manager จาก config
func NewDynamicCapabilityManager(cfg *Config) *DynamicCapabilityManager {
	dcm := &DynamicCapabilityManager{
		cfg:               cfg,
		capabilityMapping: make(map[string]string),
	}
	dcm.buildMapping()
	return dcm
}

// buildMapping สร้าง mapping จาก model_list
func (dcm *DynamicCapabilityManager) buildMapping() {
	for _, model := range dcm.cfg.ModelList {
		// หา base model จาก model field (e.g., "zhipu/glm-4.7" → "glm-4.7")
		baseModel := dcm.extractBaseModel(model.Model)

		// เก็บ mapping จาก model_name ที่ user ตั้ง → base model
		if model.ModelName != "" {
			dcm.capabilityMapping[model.ModelName] = baseModel
		}

		// เก็บ mapping จาก model field ด้วย
		if model.Model != "" {
			dcm.capabilityMapping[model.Model] = baseModel
		}

		// เก็บ mapping จาก base model ตัวเอง
		dcm.capabilityMapping[baseModel] = baseModel
	}
}

// extractBaseModel แยกชื่อ model จาก full path
func (dcm *DynamicCapabilityManager) extractBaseModel(modelPath string) string {
	// "zhipu/glm-4.7" → "glm-4.7"
	// "openai/gpt-4" → "gpt-4"
	// "qwen/qwen-plus" → "qwen-plus"
	if idx := strings.LastIndex(modelPath, "/"); idx != -1 {
		return modelPath[idx+1:]
	}
	return modelPath
}

// GetCapability หา capability สำหรับ model ที่ใช้งาน
func (dcm *DynamicCapabilityManager) GetCapability(modelName string) ModelCapability {
	// 1. หาใน mapping ก่อน
	if baseModel, ok := dcm.capabilityMapping[modelName]; ok {
		if capability, ok := ModelCapabilitiesRegistry[baseModel]; ok {
			return capability
		}
	}

	// 2. ลองหาโดยตรงใน registry
	if capability, ok := ModelCapabilitiesRegistry[modelName]; ok {
		return capability
	}

	// 3. หาจาก model_list ถ้ามี
	for _, model := range dcm.cfg.ModelList {
		if model.ModelName == modelName || model.Model == modelName {
			baseModel := dcm.extractBaseModel(model.Model)
			if capability, ok := ModelCapabilitiesRegistry[baseModel]; ok {
				return capability
			}
			// ถ้าไม่เจอใน registry สร้างจาก config
			return dcm.buildCapabilityFromModelConfig(model)
		}
	}

	// 4. Fallback ไป default
	return dcm.getDefaultCapability(modelName)
}

// buildCapabilityFromModelConfig สร้าง capability จาก ModelConfig
func (dcm *DynamicCapabilityManager) buildCapabilityFromModelConfig(model ModelConfig) ModelCapability {
	baseModel := dcm.extractBaseModel(model.Model)

	// ดึง context window จาก config ถ้ามี (ใช้ default 4096)
	contextWindow := 4096 // default

	// ดึง timeout จาก config
	maxDuration := 10 * time.Minute // default
	if model.RequestTimeout > 0 {
		maxDuration = time.Duration(model.RequestTimeout) * time.Second
	}

	return ModelCapability{
		Name:              model.ModelName,
		ContextWindow:     contextWindow,
		IsGoodAtCoding:    true, // สมมติว่า model ที่เปิดใช้งาน = ใช้ได้
		IsGoodAtReasoning: true,
		IsGoodAtAnalysis:  true,
		MaxTaskDuration:   maxDuration,
		RecommendedRoles:  []string{"planner", "coder", "reviewer", "researcher"},
		Strengths:         []string{"configured model", "active in model_list", "base: " + baseModel},
		Weaknesses:        []string{"capabilities inferred from config"},
		DelegateTasks:     []string{}, // ไม่ delegate ถ้าไม่รู้
	}
}

// getDefaultCapability สร้าง default capability จาก model config
func (dcm *DynamicCapabilityManager) getDefaultCapability(modelName string) ModelCapability {
	// หาใน model_list ก่อน
	for _, model := range dcm.cfg.ModelList {
		if model.ModelName == modelName {
			return dcm.buildCapabilityFromModelConfig(model)
		}
	}

	// ถ้าไม่เจอเลย ใช้ conservative default
	return ModelCapability{
		Name:              modelName,
		ContextWindow:     4096,
		IsGoodAtCoding:    false,
		IsGoodAtReasoning: false,
		IsGoodAtAnalysis:  false,
		MaxTaskDuration:   5 * time.Minute,
		RecommendedRoles:  []string{"general"},
		Strengths:         []string{"unknown"},
		Weaknesses:        []string{"unknown capabilities - use with caution"},
		DelegateTasks:     []string{"coding", "architecture", "complex_analysis"},
	}
}

// IsModelActive เช็คว่า model อยู่ใน model_list ที่เปิดใช้งานหรือไม่
func (dcm *DynamicCapabilityManager) IsModelActive(modelName string) bool {
	for _, model := range dcm.cfg.ModelList {
		if model.ModelName == modelName || model.Model == modelName {
			return true
		}
	}
	return false
}

// GetActiveModels คืนค่า list ของ models ที่เปิดใช้งาน
func (dcm *DynamicCapabilityManager) GetActiveModels() []string {
	models := make([]string, 0, len(dcm.cfg.ModelList))
	seen := make(map[string]bool)

	for _, model := range dcm.cfg.ModelList {
		name := model.ModelName
		if name == "" {
			name = model.Model
		}
		if !seen[name] {
			models = append(models, name)
			seen[name] = true
		}
	}

	return models
}

// GetCapabilityForSubagentRole หา capability สำหรับ subagent role
func (dcm *DynamicCapabilityManager) GetCapabilityForSubagentRole(roleName string) ModelCapability {
	// หาใน subagent_roles
	if role, ok := dcm.cfg.SubagentRoles[roleName]; ok {
		// ใช้ model จาก role config
		if role.Model != "" {
			return dcm.GetCapability(role.Model)
		}
	}

	// ถ้าไม่เจอ ใช้ default
	return dcm.getDefaultCapability(roleName)
}
