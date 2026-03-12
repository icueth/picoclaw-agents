package agent

import (
	"math"
	"strings"
	"unicode"
)

// AgentTier represents an agent with its capabilities for matching
type AgentTier struct {
	AgentID      string
	Capabilities []string
}

// MatchResult represents the result of a keyword match
type MatchResult struct {
	AgentID    string
	Confidence float64
	Hits       int
}

// KeywordScorer matches tasks to agents based on keyword analysis
type KeywordScorer struct {
	primaryKeywords   map[string][]string
	secondaryKeywords map[string][]string
}

// NewKeywordScorer creates a new keyword scorer with default keyword mappings
func NewKeywordScorer() *KeywordScorer {
	return &KeywordScorer{
		primaryKeywords: map[string][]string{
			"coder":      {"code", "debug", "programming", "function", "script", "error", "bug", "fix", "refactor"},
			"researcher": {"research", "find", "information", "analyze", "analysis", "search", "data"},
			"writer":     {"write", "writing", "article", "story", "content", "creative", "blog", "post"},
		},
		secondaryKeywords: map[string][]string{
			"coder":      {"python", "javascript", "java", "go", "rust", "c++", "html", "css", "sql", "api", "library", "framework"},
			"researcher": {"market", "trends", "news", "report", "study", "survey"},
			"writer":     {"essay", "document", "documentation", "copy", "narrative"},
		},
	}
}

// Match finds the best matching agent for a given task
func (k *KeywordScorer) Match(task string, agents []AgentTier) *MatchResult {
	taskLower := strings.ToLower(task)
	isThaiTask := IsThai(task)

	var bestMatch *MatchResult
	maxScore := 0.0

	for _, agent := range agents {
		score := 0.0
		hits := 0

		// Check primary keywords (higher weight)
		for _, keyword := range k.primaryKeywords[agent.AgentID] {
			if strings.Contains(taskLower, keyword) {
				score += 1.0
				hits++
			}
		}

		// Check secondary keywords (lower weight)
		for _, keyword := range k.secondaryKeywords[agent.AgentID] {
			if strings.Contains(taskLower, keyword) {
				score += 0.5
				hits++
			}
		}

		// Check capabilities
		for _, capability := range agent.Capabilities {
			capLower := strings.ToLower(capability)
			if strings.Contains(taskLower, capLower) {
				score += 0.8
				hits++
			}
		}

		// Normalize score based on task length
		if len(task) > 0 {
			score = score / math.Sqrt(float64(len(task))) * 10
		}

		// Boost for Thai language tasks (Thai keywords)
		if isThaiTask {
			thaiKeywords := map[string][]string{
				"coder":      {"โค้ด", "เขียน", "แก้บัก", "debug", "ฟังก์ชัน", "python", "javascript"},
				"researcher": {"หาข้อมูล", "วิเคราะห์", "research", "ข่าว", "study"},
				"writer":     {"เขียน", "บทความ", "เรื่องสั้น", "story", "content"},
			}
			for _, keyword := range thaiKeywords[agent.AgentID] {
				if strings.Contains(task, keyword) {
					score += 1.2
					hits++
				}
			}
		}

		if score > maxScore && hits > 0 {
			maxScore = score
			bestMatch = &MatchResult{
				AgentID:    agent.AgentID,
				Confidence: math.Min(score, 1.0),
				Hits:       hits,
			}
		}
	}

	return bestMatch
}

// IsThai checks if text contains Thai characters
func IsThai(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Thai, r) {
			return true
		}
	}
	return false
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
