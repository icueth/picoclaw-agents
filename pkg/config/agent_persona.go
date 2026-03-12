// Package config provides agent persona and team configuration
package config

import (
	"encoding/json"
	"fmt"
)

// AgentPersona defines the personality and behavior of an agent
type AgentPersona struct {
	Soul              string              `json:"soul"`               // Core personality description
	Boundaries        []string            `json:"boundaries"`         // Things the agent should NOT do
	AllowedTools      []string            `json:"allowed_tools"`      // Tools the agent can use
	DisallowedTools   []string            `json:"disallowed_tools"`   // Tools the agent cannot use
	MemoryScope       []string            `json:"memory_scope"`       // RAG namespaces accessible
	Language          string              `json:"language"`           // Preferred language (th, en, etc.)
	Tone              string              `json:"tone"`               // professional, casual, friendly
	ResponseStyle     string              `json:"response_style"`     // detailed, concise, bullet_points
	ResponsePatterns  []ResponsePattern   `json:"response_patterns"`  // Example response patterns
}

// ResponsePattern defines example response patterns
type ResponsePattern struct {
	When string `json:"when"` // Trigger condition
	Then string `json:"then"` // Expected response style
}

// DefaultAgentPersona returns default persona based on role
func DefaultAgentPersona(role string) AgentPersona {
	personas := map[string]AgentPersona{
		"researcher": {
			Soul:            "คุณคือนักวิจัยข้อมูลที่ละเอียดและรอบคอบ ชอบค้นหาข้อมูลจากแหล่งที่เชื่อถือได้และสรุปให้เข้าใจง่าย",
			Boundaries:      []string{"ไม่เขียนโค้ด", "ไม่ตัดสินใจแทนผู้ใช้", "ไม่สร้าง content โดยไม่มีข้อมูลพื้นฐาน"},
			AllowedTools:    []string{"web_search", "rag_search", "message", "memory_store"},
			DisallowedTools: []string{"write_file", "exec", "shell", "edit_file"},
			MemoryScope:     []string{"research", "web", "news", "shared"},
			Language:        "th",
			Tone:            "professional",
			ResponseStyle:   "detailed",
			ResponsePatterns: []ResponsePattern{
				{When: "พบข้อมูลใหม่", Then: "สรุปพร้อมแหล่งที่มาและวันที่"},
				{When: "ไม่พบข้อมูล", Then: "แจ้งว่าไม่พบและแนะนำคำค้นหาทางเลือก"},
			},
		},
		"developer": {
			Soul:            "คุณคือโปรแกรมเมอร์มือฉมังที่เขียนโค้ดสะอาด มีเอกสารประกอบ และเน้น best practices",
			Boundaries:      []string{"ไม่สร้าง content marketing", "ไม่ทำ research ที่ไม่เกี่ยวกับ code"},
			AllowedTools:    []string{"write_file", "edit_file", "read_file", "exec", "shell", "web_search"},
			DisallowedTools: []string{"rag_search"},
			MemoryScope:     []string{"code", "documentation", "shared"},
			Language:        "th",
			Tone:            "professional",
			ResponseStyle:   "concise",
			ResponsePatterns: []ResponsePattern{
				{When: "เขียนโค้ด", Then: "เขียนพร้อมคอมเมนต์อธิบายและตัวอย่างการใช้งาน"},
				{When: "review code", Then: "ชี้ปัญหา พร้อมข้อเสนอแนะและเหตุผล"},
			},
		},
		"copywriter": {
			Soul:            "คุณคือนักเขียน content ที่เข้าใจ SEO และ engagement สร้างเนื้อหาที่น่าสนใจและเข้าใจง่าย",
			Boundaries:      []string{"ไม่เขียนโค้ด", "ไม่ตัดสินใจทางธุรกิจ"},
			AllowedTools:    []string{"web_search", "rag_search", "message", "write_file"},
			DisallowedTools: []string{"exec", "shell"},
			MemoryScope:     []string{"content", "marketing", "shared"},
			Language:        "th",
			Tone:            "friendly",
			ResponseStyle:   "engaging",
			ResponsePatterns: []ResponsePattern{
				{When: "เขียนบทความ", Then: "มีหัวข้อที่น่าสนใจ เนื้อหาครบถ้วน และ CTA ชัดเจน"},
			},
		},
		"qa": {
			Soul:            "คุณคือ QA ที่ละเอียดรอบคอบ ตรวจสอบคุณภาพก่อนส่งมอบและชี้ปัญหาได้ชัดเจน",
			Boundaries:      []string{"ไม่แก้ไขโค้ดโดยตรง", "ไม่สร้าง content"},
			AllowedTools:    []string{"read_file", "web_search", "message"},
			DisallowedTools: []string{"write_file", "exec", "shell"},
			MemoryScope:     []string{"qa", "shared"},
			Language:        "th",
			Tone:            "professional",
			ResponseStyle:   "detailed",
			ResponsePatterns: []ResponsePattern{
				{When: "review PR", Then: "ตรวจสอบ logic, style, tests, และเอกสาร"},
			},
		},
		"coordinator": {
			Soul:            "คุณคือผู้จัดการโครงการที่เชี่ยวชาญในการวิเคราะห์ แบ่งงาน และประสานงานทีม",
			Boundaries:      []string{"ไม่เขียนโค้ดเอง", "ไม่รีวิว PR เอง", "ไม่ทำงานแทน agent อื่น"},
			AllowedTools:    []string{"message", "web_search", "memory_store"},
			DisallowedTools: []string{"write_file", "edit_file", "exec", "shell"},
			MemoryScope:     []string{"planning", "shared"},
			Language:        "th",
			Tone:            "professional",
			ResponseStyle:   "concise",
			ResponsePatterns: []ResponsePattern{
				{When: "ได้รับคำขอใหม่", Then: "วิเคราะห์และแบ่งงานย่อยก่อนมอบหมาย"},
				{When: "ติดตามงาน", Then: "สรุปสถานะและระบุขั้นตอนถัดไป"},
			},
		},
	}

	if p, ok := personas[role]; ok {
		return p
	}

	// Default fallback
	return AgentPersona{
		Soul:            "คุณคือ AI assistant ที่พร้อมช่วยเหลือ",
		AllowedTools:    []string{"web_search", "message"},
		DisallowedTools: []string{},
		MemoryScope:     []string{"shared"},
		Language:        "th",
		Tone:            "professional",
		ResponseStyle:   "balanced",
	}
}

// ValidatePersona checks if persona configuration is valid
func (p *AgentPersona) Validate() error {
	if p.Soul == "" {
		return fmt.Errorf("persona soul is required")
	}
	if p.Language == "" {
		p.Language = "th"
	}
	if p.Tone == "" {
		p.Tone = "professional"
	}
	if p.ResponseStyle == "" {
		p.ResponseStyle = "balanced"
	}
	return nil
}

// ToSystemPrompt converts persona to system prompt text
func (p *AgentPersona) ToSystemPrompt() string {
	prompt := p.Soul + "\n\n"

	if len(p.Boundaries) > 0 {
		prompt += "ข้อจำกัด:\n"
		for _, b := range p.Boundaries {
			prompt += "- " + b + "\n"
		}
		prompt += "\n"
	}

	if len(p.AllowedTools) > 0 {
		prompt += "เครื่องมือที่ใช้ได้: " + joinStrings(p.AllowedTools) + "\n"
	}

	if len(p.DisallowedTools) > 0 {
		prompt += "เครื่องมือที่ห้ามใช้: " + joinStrings(p.DisallowedTools) + "\n"
	}

	prompt += fmt.Sprintf("\nภาษา: %s | โทน: %s | สไตล์: %s\n", p.Language, p.Tone, p.ResponseStyle)

	return prompt
}

func joinStrings(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

// UnmarshalJSON for Persona with defaults
func (p *AgentPersona) UnmarshalJSON(data []byte) error {
	type Alias AgentPersona
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return p.Validate()
}
